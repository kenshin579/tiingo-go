# Tiingo Go SDK — Search 카테고리 설계

- 작성일: 2026-09-05
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/search`)
- 토픽: Tiingo Search(티커·이름·ISIN 자산 검색) Go 클라이언트 — 여섯 번째 카테고리
- 선행: v0.5.0 (EOD + Fundamentals + Crypto + Forex + IEX). 앞선 설계들의 구조·규약을 따른다.
- API 참조: `docs/api/utilities/search.md`, `docs/api/llms-full.txt`

## 배경 / 목적

카테고리 단위 점진 확장의 여섯 번째이자 가장 작은 것(엔드포인트 1개). 다른 카테고리의 진입점 역할을
한다 — 사용자가 아는 이름·ISIN 으로 티커를 찾아 EOD·Fundamentals·IEX 를 호출하는 흐름이다.

## 사전 조사 결과 (실호출 확인, 2026-09-05)

| 호출 | 결과 |
|---|---|
| `GET /tiingo/utilities/search?query=apple` | 200, 10건 |
| `GET /tiingo/utilities/search/apple` (경로 방식) | 200, **위와 동일한 결과** |
| `?query=zzzznosuchassetxyz` | 200 + `[]` (에러 아님) |
| 파라미터 없음 | **400** `{"detail":"Error: query format was not correct. \"query\" or \"isin\" must be a string."}` |
| `?isin=US0378331005` | **200, 1건**(AAPL) — 문서 표에 없는 파라미터인데 동작한다 |

### 응답 필드가 문서와 다르다

문서 표는 `ticker, name, assetType, isActive, permaTicker, openFIGI` 6개라 적지만, 실제 응답은 7개다:

```
name, ticker, permaTicker, openFIGIComposite, assetType, isActive, countryCode
```

- 문서의 `openFIGI` → 실제 **`openFIGIComposite`**.
- 문서에 없는 **`countryCode`** 가 온다(`US`, `CA` 등).

### 티커가 유일하지 않다

`query=apple` 결과에 `AAPL` 이 **두 건** 온다:

| ticker | permaTicker | countryCode | openFIGIComposite |
|---|---|---|---|
| AAPL | `CA000000137372` | CA | `null` |
| AAPL | `US000000000038` | US | `BBG000B9XRY4` |

같은 티커가 국가별로 존재하므로 **유일 키는 `permaTicker`** 다. 호출자가 티커로 결과를 대응시키면 안 된다.

### `openFIGIComposite` 의 빈 값이 두 가지다

- 캐나다 상장 등에서는 `null`.
- **MSFX 는 문자열 `"nan"`** 으로 온다(실측). FIGI 형식(12자 영숫자)이 아니며 명백한 데이터 오류다.

### 파라미터 실측

| 파라미터 | 확인 결과 |
|---|---|
| `exactTickerMatch=true` | `query=aapl` → 2건(미국·캐나다 AAPL) |
| `includeDelisted=true` | 상장폐지 포함, 10건 |
| `limit=3` | 3건. **`limit=1000` 은 100건으로 잘린다**(상한 100) |
| `columns=ticker,name` | 응답 키가 실제로 `name, ticker` 둘만 남는다 |
| `isin=<ISIN>` | 동작(위 표). `llms-full.txt`: "ISIN and OpenFIGI search are supported via the Search utility" |

## 결정 사항 (브레인스토밍)

1. **`openFIGIComposite` 은 `string` + 정규화**(선택지 B). 커스텀 `UnmarshalJSON` 으로 `null` 과
   `"nan"` 을 **모두 빈 문자열**로 만든다. 사용자는 `if figi != ""` 하나로 판별한다.
   (기각: 원문 그대로 — `"nan"` 을 유효한 FIGI 로 오해할 수 있다. `*string` — 두 빈 값이 다르게
   표현돼 판별이 더 복잡해진다.) 이 라이브러리는 응답을 그대로 전달하는 원칙을 지켜 왔지만, 그건
   **의미 있는 값**에 대한 것이고 `"nan"` 은 그 자체가 결손 표현이다. 원문이 `"nan"` 일 수 있다는
   사실은 필드 주석에 남긴다. 선례: Fundamentals 의 빈 문자열 날짜 → zero value.
2. **쿼리 파라미터 방식(`?query=`)을 쓴다.** 경로 방식과 결과가 같은데, 검색어에 공백·슬래시가
   들어오면 경로 구조가 깨질 위험이 있다. 쿼리 파라미터는 인코딩이 안전하다.
3. **메서드 2개로 나눈다**: `Search(ctx, query, opts)` 와 `SearchByISIN(ctx, isin, opts)`.
   Tiingo 가 `query` 또는 `isin` 중 하나를 요구하므로(둘 다 없으면 400), 시그니처로 구분하는 편이
   하나의 옵션 구조체에 둘을 넣는 것보다 명확하다.
4. **타입 이름은 `Result`**. 패키지가 `search` 라 `search.Result` 가 자연스럽다.
5. **`Limit` 은 `int`, 0 이면 미전송**. Tiingo 기본값이 10 이므로 0 을 "무제한"으로 오해하지 않도록
   주석에 적는다. 상한이 100 이라는 사실도 적는다.
6. 필드 주석 규약 유지, fixture 는 실호출.

## 아키텍처

```
tiingo-go/
├── client.go             # (수정) Search *search.Client 필드 추가
├── search/
│   ├── client.go         # search.New(hc) *Client
│   ├── search.go         # Result, FIGI, SearchOptions + Search / SearchByISIN
│   ├── *_test.go
│   └── testdata/*.json   # 실호출 fixture 4종
├── examples/search/main.go
└── integration_test.go   # (수정) Search 4건 추가
```

엔드포인트가 하나뿐이라 파일도 하나로 둔다.

## 데이터 모델

```go
// Result 는 검색 결과 한 건이다.
// Ticker 는 유일하지 않다 — 같은 티커가 국가별로 여러 건 올 수 있고(예: AAPL 의 미국·캐나다 상장),
// 유일 키는 PermaTicker 다.
type Result struct {
	Ticker            string `json:"ticker"`            // 티커. 국가가 다르면 중복될 수 있다
	Name              string `json:"name"`              // 자산 이름
	PermaTicker       string `json:"permaTicker"`       // 영구 식별자. 검색 결과의 유일 키
	OpenFIGIComposite FIGI   `json:"openFIGIComposite"` // OpenFIGI 복합 식별자. 없으면 빈 문자열
	AssetType         string `json:"assetType"`         // 자산 종류(Stock, ETF, Mutual Fund)
	IsActive          bool   `json:"isActive"`          // 현재 거래 중인지
	CountryCode       string `json:"countryCode"`       // 상장 국가(예: US, CA). 문서 표에는 없으나 실제로 온다
}

// FIGI 는 OpenFIGI 식별자다. Tiingo 는 값이 없을 때 null 과 문자열 "nan" 두 가지로 보내는데
// 둘 다 결손을 뜻하므로 빈 문자열로 정규화한다 — 호출자는 `figi != ""` 하나만 확인하면 된다.
type FIGI string

func (f *FIGI) UnmarshalJSON(b []byte) error // null·"nan" → "", 그 외 문자열은 그대로
```

`FIGI` 를 별도 타입으로 둔 이유: 정규화를 한 곳에 모으고, 필드 타입만 봐도 특별한 처리가 있음을
알 수 있게 하기 위함이다. 기저 타입이 `string` 이라 비교·출력은 문자열과 동일하게 쓴다.

## 메서드와 옵션

```go
// Search 는 티커·이름으로 자산을 검색한다. GET /tiingo/utilities/search
// 일치하는 자산이 없으면 에러가 아니라 빈 슬라이스를 돌려준다.
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) ([]Result, error)

// SearchByISIN 은 ISIN 으로 자산을 검색한다. GET /tiingo/utilities/search
// 카탈로그 표에는 없는 파라미터지만 실제로 지원된다(llms-full.txt 가 명시하고 실호출로 확인).
func (c *Client) SearchByISIN(ctx context.Context, isin string, opts *SearchOptions) ([]Result, error)

// SearchOptions 는 검색의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type SearchOptions struct {
	ExactTickerMatch bool // true 면 티커가 정확히 일치하는 것만
	IncludeDelisted  bool // true 면 상장폐지 종목도 포함한다(기본은 제외)
	// Limit 은 최대 건수다. 0 이면 미전송이고 Tiingo 기본값 10 이 적용된다(무제한이 아니다).
	// 상한은 100 이라 그보다 큰 값을 줘도 100 건까지만 온다.
	Limit int
	// Columns 는 받을 컬럼을 지정한다. 지정하면 응답에서 나머지 키가 빠져 해당 필드는 zero 값이 된다.
	Columns []string
}
```

- `query`/`isin` 은 `strings.TrimSpace` 후 빈 값이면 에러(Tiingo 도 400 을 주지만 왕복을 아낀다).
- `ExactTickerMatch`/`IncludeDelisted` 는 true 일 때만 전송, `Limit` 은 0 초과일 때만, `Columns` 는
  콤마 결합.
- 두 메서드는 같은 옵션 구조체를 공유한다. `isin` 과 네 옵션의 조합을 실호출로 모두 확인했다 —
  `limit`·`includeDelisted`·`exactTickerMatch` 는 200 이고, `columns` 는 지정한 키만 남는다.
  옵션 구조체를 나눌 이유가 없다.

## 테스트

- **fixture 4종**(실호출): `search_apple.json`(10건, AAPL 중복·`openFIGIComposite` null 포함),
  `search_microsoft.json`(**`"nan"` 이 들어간 MSFX 포함** — 정규화 테스트용), `search_isin.json`(1건),
  `search_columns.json`(`columns=ticker,name` — 필드 누락 시 zero 값 확인).
- **단위 테스트**: 두 메서드의 파싱, **`FIGI` 정규화**(null → `""`, `"nan"` → `""`, 정상 값은 그대로),
  티커 중복이 그대로 유지되는지(SDK 가 임의로 합치지 않음), 쿼리 생성(zero 값 생략, true 일 때만,
  `Limit` 0 미전송, `Columns` 콤마 결합), 빈 query/isin 에러, 빈 배열이 에러가 아님, 400 → `APIError`.
- **integration**: `apple` 검색, ISIN 검색, `exactTickerMatch`, 없는 자산(빈 배열) 4건.

## 릴리스

머지 후 **v0.6.0**. README 커버리지 표에 Search 2행 추가.

## 범위 밖 (후속)

- Dividends/Splits — 403 보류 중.
- 나머지 REST(News, Equity Realtime, BOATS, Fund Fees)와 WebSocket 5종.
- CSV, 재시도, moneyflow 통합.

## 주의

- 문서 표가 응답 필드를 두 군데 틀리게 적는다(`openFIGI` vs 실제 `openFIGIComposite`, `countryCode`
  누락). 실측을 따른다.
- `isin` 파라미터는 카탈로그 표에 없다. `llms-full.txt` 의 서술과 400 에러 메시지, 실호출 세 가지로
  뒷받침된다.
- `limit` 상한 100 은 문서에 없다("Defaults to 10 and can be set to a..." 로 잘려 있음). 실측값이다.
