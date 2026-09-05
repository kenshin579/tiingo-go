# Tiingo Go SDK — IEX 카테고리 설계

- 작성일: 2026-09-05
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/iex`)
- 토픽: Tiingo IEX(미국 주식 실시간 스냅샷·인트라데이 시세) Go 클라이언트 — 다섯 번째 카테고리
- 선행: v0.4.0 (EOD + Fundamentals + Crypto + Forex). 앞선 설계들의 구조·규약을 따른다.
- API 참조: `docs/api/rest/iex.md`

## 배경 / 목적

카테고리 단위 점진 확장의 다섯 번째. IEX 는 Starter 플랜에 포함돼 실호출 검증이 확실하고, 미국 주식
실시간 스냅샷이라 moneyflow 의 종목 화면에 바로 쓸모가 있다.

**이 카테고리는 앞선 넷과 결정적으로 다른 점이 하나 있다 — null 필드**. 응답 17개 필드 중 9개가
장 마감 시간대에 `null` 로 온다. 값 타입으로 받으면 "호가 없음"이 "가격 0"으로 조용히 바뀐다.

## 사전 조사 결과 (실호출 확인, 2026-09-05 토요일 = 장 마감)

| 엔드포인트 | 결과 |
|---|---|
| `GET /iex?tickers=` | **301** (본문 없음) — 끝 슬래시가 없으면 리다이렉트 |
| `GET /iex/?tickers=aapl,msft` | 200. 평탄한 배열 2건 |
| `GET /iex/<ticker>` | 200. 위와 같은 형태 1건 |
| `GET /iex/<ticker>/prices` | 200. 평탄한 배열 `{date, open, high, low, close}` |
| `GET /iex/aapl,msft/prices` | **404** `{"detail":"Not found."}` — 시세는 복수 티커 불가 |

### null 필드 (설계를 가르는 지점)

AAPL·MSFT·BRK-A(고가)·SPG-P-J(우선주)·VFINX(뮤추얼펀드) 다섯을 조회한 결과, **모든 종목에서
정확히 같은 9개 필드가 null** 이다:

```
askPrice, askSize, bidPrice, bidSize, last, lastSaleTimestamp, lastSize, mid, quoteTimestamp
```

나머지 8개(`ticker, timestamp, open, high, low, tngoLast, volume, prevClose`)는 전부 값이 있다.
장 마감이라 호가·체결 정보가 없기 때문이며, 장중에는 채워질 것으로 본다(아래 «위험» 참조).

### 그 밖의 실측 사실

- **없는 티커는 응답에서 조용히 빠진다.** 5개를 요청했는데 4개만 왔다(`dcfc` 누락). 단독으로 조회해도
  에러가 아니라 `[]` 다(배치·티커 경로 모두 200). 따라서 결과 길이가 요청보다 짧을 수 있고, 호출자는
  이를 감안해야 한다.
- **응답의 티커는 대문자**다. `?tickers=aapl` 로 요청해도 `"ticker":"AAPL"` 로 온다.
- **응답 순서가 요청 순서와 다르다.** `tickers=aapl,msft,dcfc` 로 요청하면 `['MSFT','AAPL']` 로 온다
  (실측). 없는 티커 누락과 겹치면 인덱스 대응이 완전히 깨지므로, 호출자는 반드시 `Ticker` 필드로
  찾아야 한다 — doc 주석에 명시한다.
- **시세에 거래량이 기본으로 없다.** 문서: "This value will only be exposed if explicitly passed to the
  `columns` request parameter." 실측으로도 5분봉 156행 어디에도 `volume` 키가 없었고,
  `columns=open,high,low,close,volume` 을 주면 `volume` 이 채워졌다(예: 223412.0).
- **`afterHours=true` 는 행이 늘어난다**(실측 1시간봉 12행 → 19행). 장 전후 데이터가 포함된다.
- `forceFill=true` 는 같은 12행(거래 없는 구간이 없었던 듯).
- 시세의 `date` 는 `2026-09-03T13:30:00.000Z` 처럼 시각이다 → `types.Time`.
- **`volume` 의 JSON 표기가 엔드포인트마다 다르다.** 시세는 `"volume": 38607.0`(소수점 있음),
  스냅샷은 `"volume": 39606884`(정수). Go 의 `encoding/json` 은 소수점 있는 수를 `int64` 에 넣지 못하고
  **에러**를 낸다(`cannot unmarshal number 38607.0 into ... int64`, 실측). 따라서 **시세의 Volume 은
  `float64`**, 스냅샷의 Volume 은 `int64` 로 둔다. 문서 표는 둘 다 int64 라 적지만 실제와 다르다.
- 같은 이유로 스냅샷의 null 인 수량 필드(`lastSize`, `bidSize`, `askSize`)는 **`*float64`** 로 둔다.
  장 마감이라 실제 표기를 못 봤는데, 같은 API 가 수량에 소수점을 붙이는 전례가 있으므로 `*int64` 면
  장중에 디코딩이 하드 실패한다. `*float64` 는 `100` 과 `100.0` 을 모두 받는다.

## 결정 사항 (브레인스토밍)

1. **null 가능 필드 9개를 포인터로 둔다**(선택지 A). nil 이 "값 없음"이고 0 과 구분된다.
   (기각: 전부 값 타입 — 장 마감이 하루의 절반 이상이라 null 이 예외가 아니라 일상이고, 시각 필드가
   zero time 이 되면 의미가 모호하다. 값 타입 + `HasQuote()` 판별 메서드 — 판별 규칙을 SDK 가 임의로
   정하게 된다.) 선례: Fundamentals 의 `*types.Date`.
2. **메서드 2개**: `Quotes(ctx, tickers []string)` 와 `Prices(ctx, ticker string, opts)`.
   시세는 복수 티커가 404 이므로 **슬라이스가 아니라 `string`** 으로 받는다 — 시그니처가 제약을 말한다.
3. **이름은 `Quotes`/`Quote`**. 문서 제목이 "Current Top-of-Book & Last Price" 이고 실제 필드가 호가를
   넘어 당일 OHLC·거래량·전일종가까지 담으므로, Crypto·Forex 의 순수 호가(`TopOfBook`)와 구분한다.
4. **`columns` 를 그대로 노출한다.** `IncludeVolume bool` 같은 특수 플래그로 감싸지 않는다 — 다른 컬럼
   조합을 막고 Tiingo 가 컬럼을 늘리면 따라가야 한다. API 사양을 그대로 전달하는 원칙을 지킨다.
5. **배치 경로는 `/iex/`(끝 슬래시 포함)** 를 쓴다. 단일 티커도 이 경로로 처리해 코드 경로를 하나로 둔다.
6. 필드 주석 규약 유지, fixture 는 실호출.

## 아키텍처

```
tiingo-go/
├── client.go             # (수정) IEX *iex.Client 필드 추가
├── iex/
│   ├── client.go         # iex.New(hc) *Client
│   ├── quotes.go         # Quote + Quotes(ctx, tickers)
│   ├── prices.go         # Price, PriceOptions, ResampleFreq + Prices(ctx, ticker, opts)
│   ├── *_test.go
│   └── testdata/*.json   # 실호출 fixture 4종
├── examples/iex/main.go
└── integration_test.go   # (수정) IEX 4건 추가
```

`forex` 와 동일하게 `internal/httpclient` 로 호출하고 `types` 를 직접 import 한다. 티커 검증·결합
헬퍼는 `forex` 와 같은 형태가 필요하지만 패키지가 다르므로 각자 둔다(카테고리가 더 늘면 `internal`
추출을 재검토 — Forex 스펙에도 같은 메모).

## 데이터 모델

```go
// Quote 는 IEX 실시간 스냅샷이다. 호가·최종체결에 더해 당일 시가·고가·저가와 전일 종가까지 담는다.
// 장 마감 시간대에는 호가·체결 관련 9개 필드가 모두 null 로 오므로 포인터다 — nil 은 "값 없음"이고
// 0 과 구분된다.
type Quote struct {
	Ticker            string      `json:"ticker"`            // 티커. 요청과 무관하게 대문자로 온다
	Timestamp         types.Time  `json:"timestamp"`         // 스냅샷 기준 시각
	LastSaleTimestamp *types.Time `json:"lastSaleTimestamp"` // 마지막 체결 시각. 장 마감 시 nil
	QuoteTimestamp    *types.Time `json:"quoteTimestamp"`    // 호가 기준 시각. 장 마감 시 nil
	Open              float64     `json:"open"`              // 당일 시가
	High              float64     `json:"high"`              // 당일 고가
	Low               float64     `json:"low"`               // 당일 저가
	Mid               *float64    `json:"mid"`               // 중간가. 장 마감 시 nil
	TngoLast          float64     `json:"tngoLast"`          // Tiingo 기준 최종가. 장 마감 후에도 채워진다
	Last              *float64    `json:"last"`              // IEX 최종 체결가. 장 마감 시 nil
	LastSize          *float64    `json:"lastSize"`          // 마지막 체결 수량. 장 마감 시 nil
	BidSize           *float64    `json:"bidSize"`           // 매수 호가 수량. 장 마감 시 nil
	BidPrice          *float64    `json:"bidPrice"`          // 매수 호가. 장 마감 시 nil
	AskPrice          *float64    `json:"askPrice"`          // 매도 호가. 장 마감 시 nil
	AskSize           *float64    `json:"askSize"`           // 매도 호가 수량. 장 마감 시 nil
	Volume            int64       `json:"volume"`            // 당일 누적 거래량
	PrevClose         float64     `json:"prevClose"`         // 전일 종가
}

// Price 는 인트라데이 구간 하나의 시세다.
// Volume 은 PriceOptions.Columns 에 "volume" 을 넣어야 응답에 포함된다 — 없으면 0 이다.
type Price struct {
	Date   types.Time `json:"date"`   // 구간 시작 시각
	Open   float64    `json:"open"`   // 시가
	High   float64    `json:"high"`   // 고가
	Low    float64    `json:"low"`    // 저가
	Close  float64    `json:"close"`  // 종가
	Volume float64    `json:"volume"` // IEX 거래량. columns 미지정 시 0. Tiingo 가 소수점을 붙여 보내 float64 다
}
```

## 메서드와 옵션

```go
// Quotes 는 IEX 실시간 스냅샷을 받는다. GET /iex/
// 경로 끝 슬래시가 필요하다 — /iex 는 301 이다.
// 없는 티커는 에러가 아니라 응답에서 빠지므로 결과 길이가 요청보다 짧을 수 있다.
func (c *Client) Quotes(ctx context.Context, tickers []string) ([]Quote, error)

// Prices 는 인트라데이 시세를 받는다. GET /iex/<ticker>/prices
// 티커는 하나만 받는다 — 복수를 경로에 넣으면 Tiingo 가 404 를 돌려준다.
func (c *Client) Prices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error)
```

```go
// ResampleFreq 는 시세 리샘플 주기다. "5min", "4hour" 처럼 <숫자><단위> 형식이면 무엇이든 되므로
// 아래 상수는 자주 쓰는 값일 뿐 목록이 닫혀 있지 않다.
type ResampleFreq = string

const (
	Resample1Min  ResampleFreq = "1min"  // 1분봉
	Resample5Min  ResampleFreq = "5min"  // 5분봉
	Resample1Hour ResampleFreq = "1hour" // 1시간봉
	Resample1Day  ResampleFreq = "1day"  // 일봉
)

// PriceOptions 는 Prices 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작. zero 면 미전송
	EndDate      time.Time    // 조회 종료. zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플 주기. 빈 값이면 Tiingo 기본값
	AfterHours   bool         // true 면 장 전후 데이터를 포함한다(실측: 1시간봉 12행 → 19행)
	ForceFill    bool         // true 면 거래가 없던 구간도 직전 값으로 채운다
	// Columns 는 받을 컬럼을 지정한다. 비어 있으면 Tiingo 기본 컬럼(date/open/high/low/close)만 오고
	// Volume 은 0 이다 — 거래량이 필요하면 "volume" 을 반드시 포함해야 한다.
	Columns []string
}
```

- 티커는 `strings.TrimSpace` 후 빈 값이면 에러. `Quotes` 는 슬라이스가 비어도 에러.
- `Quotes` 는 티커를 콤마로 합쳐 쿼리 `tickers` 에 넣는다.
- `Prices` 는 티커를 `url.PathEscape` 해 경로에 넣는다(복수 불가이므로 결합 로직 없음).
- `AfterHours`/`ForceFill` 은 true 일 때만 전송. `Columns` 는 콤마 결합.
- `StartDate`/`EndDate` 는 `YYYY-MM-DD`.

## 테스트

- **fixture 4종**(실호출): `quotes_aapl.json`(단일), `quotes_multi.json`(aapl,msft + 없는 티커
  `dcfc` 를 섞어 **누락**을 고정), `prices_aapl_5min.json`(기본 — volume 없음),
  `prices_aapl_volume.json`(`columns=open,high,low,close,volume`).
- **단위 테스트**: 두 메서드의 파싱, **null 필드가 nil 로 디코딩되는지**(9개 각각 확인),
  `Columns` 없이는 `Volume` 이 0 이고 있으면 채워지는지, 요청 경로가 `/iex/`(끝 슬래시)인지,
  쿼리 생성(zero 값 생략, `afterHours`/`forceFill` 은 true 일 때만, 콤마 결합, 날짜 형식),
  빈 티커·빈 슬라이스 에러, 빈 배열이 에러가 아님, 404 → `APIError`.
- **integration**: AAPL·MSFT 로 두 메서드 실호출. 장 마감/개장 어느 때 돌려도 깨지지 않도록
  **null 필드는 nil 여부를 단정하지 않고**, 값이 있으면 양수인지만 확인한다. 시세는 최근 7일 범위로
  조회하고 빈 응답을 허용한다(주말·휴장 대비).

## 릴리스

머지 후 **v0.5.0**. README 커버리지 표에 IEX 2행 추가.

## 범위 밖 (후속)

- Dividends/Splits — 403 보류 중.
- 나머지 REST(News, Equity Realtime, BOATS, Fund Fees, Search)와 WebSocket 5종.
- CSV, 재시도, moneyflow 통합.
- 티커 헬퍼의 패키지 간 공유(`internal` 추출) — 카테고리가 더 늘면 재검토.

## 위험 / 주의

- **장중 응답을 아직 못 봤다.** 조사는 전부 토요일(장 마감)에 했다. 장중에는 9개 null 필드가 채워질
  것이고, 반대로 **개장 직후 첫 체결 전에는 `open`/`high`/`low` 가 null 일 가능성**이 있다. 그 경우
  값 타입인 이 필드들은 디코딩이 실패하지 않고 조용히 0 이 된다. 장중에 한 번 실호출해 확인하고,
  null 이 확인되면 해당 필드도 포인터로 바꾼다(공개 API 변경이므로 v0.5.x 안에서 처리).
- 없는 티커가 조용히 빠지므로 `len(quotes) < len(tickers)` 가 정상이고, **순서도 요청과 다르다**.
  인덱스로 대응시키면 안 되고 `Ticker` 필드로 찾아야 한다 — doc 주석에 명시한다.
- `/iex`(슬래시 없음)는 301 이다. Go 의 `http.Client` 는 리다이렉트를 따라가지만, 경로를 `/iex/` 로
  고정해 불필요한 왕복을 없앤다.
