# Tiingo Go SDK — 기반 + End-of-Day 설계

- 작성일: 2026-09-04
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/sdk-foundation`)
- 토픽: Tiingo API 의 Go 클라이언트 라이브러리 — 확장 가능한 기반 + 첫 카테고리(End-of-Day)
- 선행: 1단계 문서 카탈로그 (PR #1, `docs/api/` 23페이지 + 공식 llms 원본).
  본 스펙은 `docs/api/rest/end-of-day.md`, `docs/api/general/{overview,connecting}.md`,
  `docs/api/appendix/symbology.md`, `docs/api/llms-full.txt` 를 1차 참조로 쓴다.

## 배경 / 목적

Tiingo(주식 EOD/실시간 시세, 뉴스, 펀더멘털, 암호화폐, 외환) API 를 워크스페이스의 다른
프로젝트에서 재사용할 수 있도록 **독립 Go 라이브러리**로 만든다. opendart-go / fmp-go /
ecos-go 와 같은 계열이다.

장기 목표는 Tiingo REST 11 그룹 + WebSocket 5 종 전체 커버리지이며, **카테고리 단위 점진
확장**(opendart-go / fmp-go 패턴)으로 간다. 본 스펙의 범위는 **기반 + End-of-Day 카테고리**
= v0.1.0 이다.

첫 카테고리로 End-of-Day 를 고른 이유: 구조가 단순해 httpclient·에러 매핑·날짜/정렬
파라미터 패턴을 확정하기 좋고, 무료 플랜으로 실호출 fixture 와 integration test 가 가능하며,
moneyflow 의 US 일별 시세(현재 FMP 무료 250콜/일 예산 게이트에 묶여 있음)를 보강할 수 있다.

## 결정 사항 (브레인스토밍)

1. **빌드 전략**: 수작업 점진 확장(fmp-go 패턴). 카탈로그 md 의 필드 표에서 코드를 생성하는
   방식은 기각 — 표가 규칙적이긴 하나 Tiingo 측 문서 오류 4건(아래 «주의»)을 그대로 코드에
   옮기게 되고 생성기 유지 비용이 든다.
2. **첫 카테고리 = End-of-Day**. (검토·보류: Fundamentals — 유료 add-on 이라 무료 키로는
   Dow 30 만 되고 `statementData` 중첩 구조가 첫 카테고리로 복잡. News — moneyflow 활용처
   없음.)
3. **API 표면 = 메서드 3개**: `Meta` / `LatestPrice` / `HistoricalPrices`. 문서가 "Latest
   Price" 와 "Historical Prices" 를 별개 용례로 제시하고, 소비처의 두 갈래(현재가 1건 /
   기간 백필)와 일치한다. 실제 엔드포인트는 하나이므로 `LatestPrice` 는 `HistoricalPrices`
   위의 얇은 래퍼다. (대안: 메서드 2개 + 옵션 구조체 / functional option — 후자는 옵션이
   6개뿐이라 구조체 대비 이득이 적어 기각.)
4. **인증은 `Authorization: Token <key>` 헤더**. `?token=` 쿼리도 동작하지만 헤더 방식이
   URL·서버 로그·에러 메시지에 키를 남기지 않는다.
5. **외부 의존성 허용**. 테스트는 fmp-go 와 동일하게 `testify` 사용.
6. **필드 주석 규약**: 짧은 설명은 **필드 오른쪽 줄 끝 주석**, 한 줄(약 120자)에 안 들어가면
   필드 위 주석. 카탈로그 표의 DESCRIPTION 을 한국어로 옮기고, 문서에 없지만 중요한 맥락
   (증분 동기화 규칙 등)은 한 줄 덧붙인다. 타입·메서드에도 doc 주석을 단다.
7. **v1 은 JSON 만**. `format=csv` 는 후속(호출부가 `GetRaw` 로 열어둘 수는 있음).

## 아키텍처

```
tiingo-go/
├── go.mod                    # module github.com/kenshin579/tiingo-go, go 1.25
├── client.go                 # tiingo.Client — 단일 진입점, EOD 서브클라이언트 보유
├── config.go                 # Option (WithBaseURL / WithTimeout / WithHTTPClient)
├── from_env.go               # NewClientFromEnv() — TIINGO_API_KEY
├── errors.go                 # APIError, ErrNotFound (httpclient 타입 별칭)
├── types.go                  # Date — Tiingo 의 두 날짜 형식 파싱
├── types_test.go
├── eod/
│   ├── client.go             # eod.New(hc) *Client
│   ├── meta.go               # Meta 구조체 + Meta(ctx, ticker)
│   ├── prices.go             # Price, PriceOptions, ResampleFreq, HistoricalPrices, LatestPrice
│   ├── meta_test.go, prices_test.go
│   └── testdata/*.json       # 실호출 fixture
├── internal/httpclient/
│   ├── client.go             # Authorization 헤더, GET+JSON, 에러 매핑
│   └── client_test.go        # httptest 스텁
├── examples/eod/main.go      # 실행 가능한 예시
├── integration_test.go       # build tag integration, TIINGO_API_KEY 있을 때만
├── scripts/release.sh        # 태그 + GitHub Release (fmp-go 와 동일 절차)
└── (기존) docs/, tools/gendocs/, scripts/fetch-docs.sh — 변경 없음
```

- **`tiingo.Client`**: 단일 진입점. 내부에 `*httpclient.Client` 를 두고 서비스 서브클라이언트를
  필드로 노출. v1 은 `EOD *eod.Client` 하나. 이후 `News`/`Fundamentals`/`Crypto`/… 누적.
  ```go
  type Client struct {
      http *httpclient.Client
      EOD  *eod.Client // 일별 시세·메타(End-of-Day)
  }
  func NewClient(apiKey string, opts ...Option) (*Client, error)
  func NewClientFromEnv(opts ...Option) (*Client, error) // TIINGO_API_KEY
  ```
- **`internal/httpclient`**: baseURL 기본 `https://api.tiingo.com`, timeout 기본 30s, 재시도
  없음(v1). 모든 요청에 `Authorization: Token <key>` 와 `Accept: application/json` 헤더 주입.
  `GetJSON(ctx, path, params, out)` / `GetRaw(ctx, path, params)`(비-JSON 후속용).
- **`eod` 패키지**: 도메인 경계. `eod.New(hc)` → `*eod.Client`.

## 기반 계층 상세

### 인증

- `tiingo.NewClient(apiKey, opts...)` — apiKey 빈 값이면 에러.
- `tiingo.NewClientFromEnv(opts...)` — `TIINGO_API_KEY` 미설정 시 에러.
- 요청 헤더: `Authorization: Token <apiKey>`. 쿼리에 토큰을 넣지 않는다.

### 에러

```go
// APIError 는 Tiingo 에러 응답(비-200 또는 에러 바디).
type APIError struct {
    StatusCode int    // HTTP 상태코드 — 401 키 오류, 403 권한/플랜, 404 없는 티커, 429 rate limit
    Message    string // Tiingo 에러 메시지 또는 상태 텍스트
}
var ErrNotFound = errors.New("tiingo: not found") // 조회 결과 없음(빈 배열 등)
```

- 비-200 이면 바디에서 `detail` / `error` / `message` 키를 순서대로 찾아 `Message` 로 쓰고,
  없으면 상태 텍스트. 200 이어도 빈 배열이면 서비스 계층이 `ErrNotFound`.
- 네트워크/디코딩 에러는 `%w` 로 래핑.

### `tiingo.Date`

Tiingo 는 같은 API 안에서 날짜 형식이 둘이다 — 가격은 `2019-01-02T00:00:00.000Z`,
메타는 `1980-12-12`. 호출부마다 파싱이 흩어지지 않도록 전용 타입을 둔다.

```go
// Date 는 Tiingo 의 날짜 값. 응답의 두 형식(RFC3339 타임스탬프, YYYY-MM-DD)을 모두 받아
// time.Time 으로 정규화하고, 쿼리로 나갈 때는 YYYY-MM-DD 로 직렬화한다.
type Date struct{ time.Time }

func (d *Date) UnmarshalJSON(b []byte) error // null → zero value, 두 형식 파싱, 그 외 에러
func (d Date) MarshalJSON() ([]byte, error)  // "YYYY-MM-DD"
func (d Date) String() string                // YYYY-MM-DD
```

## End-of-Day 데이터 모델

엔드포인트(카탈로그 `docs/api/rest/end-of-day.md`):
- 메타: `GET /tiingo/daily/<ticker>` — 단일 객체 응답
- 가격: `GET /tiingo/daily/<ticker>/prices` — 배열 응답. 파라미터 없으면 최신 1건.

```go
// Meta 는 자산의 메타 정보(EOD 메타 엔드포인트 응답).
type Meta struct {
	Ticker       string       `json:"ticker"`       // 자산의 티커. 주식 클래스는 마침표 대신 대시(예: BRK-A)
	Name         string       `json:"name"`         // 자산의 전체 이름
	ExchangeCode string       `json:"exchangeCode"` // 상장 거래소 식별자(예: NASDAQ)
	Description  string       `json:"description"`  // 자산에 대한 장문 설명
	StartDate    *tiingo.Date `json:"startDate"`    // 가격 데이터가 있는 가장 이른 날짜. nil 이면 가격 데이터 없음
	EndDate      *tiingo.Date `json:"endDate"`      // 가격 데이터가 있는 가장 늦은 날짜. nil 이면 가격 데이터 없음
}

// Price 는 일별 시세 한 건. 뮤추얼 펀드는 OHLC 에 그날의 NAV 가 들어간다.
type Price struct {
	Date      tiingo.Date `json:"date"`      // 이 데이터가 해당하는 날짜
	Open      float64     `json:"open"`      // 시가
	High      float64     `json:"high"`      // 고가
	Low       float64     `json:"low"`       // 저가
	Close     float64     `json:"close"`     // 종가
	Volume    int64       `json:"volume"`    // 거래량(주)
	AdjOpen   float64     `json:"adjOpen"`   // 수정 시가
	AdjHigh   float64     `json:"adjHigh"`   // 수정 고가
	AdjLow    float64     `json:"adjLow"`    // 수정 저가
	AdjClose  float64     `json:"adjClose"`  // 수정 종가. CRSP 방법론, 반올림하지 않음
	AdjVolume int64       `json:"adjVolume"` // 수정 거래량(분할 계수 반영)
	DivCash   float64     `json:"divCash"`   // 해당일(배당락일) 지급 배당금
	// SplitFactor 는 분할·역분할·분배 시 가격 조정에 쓰이는 계수. 1 이 아니면 그날 이벤트가
	// 있었다는 뜻이므로, 증분 동기화 중이라면 해당 종목의 전체 이력을 다시 받아야 한다.
	SplitFactor float64 `json:"splitFactor"`
}
```

- 메타 응답은 표에 6필드지만 `llms-full.txt` 는 `permaTicker`(티커 변경·재사용에도 안정적인
  영구 식별자), OpenFIGI 복합 식별자, 자산 종류, `isActive` 도 노출된다고 적는다. **구현 시
  실호출 fixture 로 확인해 존재하면 필드로 추가**하고 같은 규약으로 주석을 단다.
- 가격 13필드는 문서 표와 예시 JSON 이 일치하므로 그대로 매핑한다.

### 요청 옵션

```go
// ResampleFreq 는 가격 리샘플링 주기.
// 주의: daily 만 휴일 달력을 반영하고, 나머지는 표준 영업일(월~금) 기준이다.
// 기간 중간 날짜를 주면 전체 기간이 잡히도록 startDate 는 뒤로, endDate 는 앞으로 조정된다
// (예: weekly 에 수요일 startDate → 월요일로 롤백, 수요일 endDate → 금요일로 롤포워드).
type ResampleFreq string

const (
	ResampleDaily    ResampleFreq = "daily"    // 일별. 휴일 달력 반영
	ResampleWeekly   ResampleFreq = "weekly"   // 주별. 금요일 마감
	ResampleMonthly  ResampleFreq = "monthly"  // 월별. 각 월 마지막 영업일 마감
	ResampleAnnually ResampleFreq = "annually" // 연별. 각 해 마지막 영업일 마감
)

// PriceOptions 는 HistoricalPrices 의 선택 파라미터. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작일(이상, >=). zero 면 미전송
	EndDate      time.Time    // 조회 종료일(이하, <=). zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플링 주기. 빈 값이면 미전송(일별)
	Sort         string       // 정렬 컬럼. "date" 오름차순, "-date" 내림차순
	Columns      []string     // 돌려받을 컬럼만 지정. 빈 슬라이스면 전체
}
```

`StartDate`/`EndDate` 는 `YYYY-MM-DD` 로 직렬화하고, `Columns` 는 콤마로 결합한다.

### 메서드

```go
// Meta 는 자산의 메타 정보를 조회한다. GET /tiingo/daily/<ticker>
func (c *Client) Meta(ctx context.Context, ticker string) (*Meta, error)

// HistoricalPrices 는 일별 시세를 조회한다. GET /tiingo/daily/<ticker>/prices
// opts 가 nil 이거나 날짜가 없으면 Tiingo 는 최신 1건만 반환한다.
func (c *Client) HistoricalPrices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error)

// LatestPrice 는 가장 최근 일별 시세 1건을 조회한다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) LatestPrice(ctx context.Context, ticker string) (*Price, error)
```

- `ticker` 는 빈 값이면 에러. URL 경로에 넣기 전에 `url.PathEscape`.
- `LatestPrice` 는 `HistoricalPrices(ctx, ticker, nil)` 결과의 첫 요소를 돌려주고, 빈 배열이면
  `ErrNotFound`.

## 테스트

- **`internal/httpclient`**: `httptest.Server` 스텁 — `Authorization: Token` 헤더 주입 확인,
  JSON 디코딩, 비-200 → `APIError`(상태코드·메시지 추출), 에러 바디 키 3종, 타임아웃.
- **`tiingo.Date`**: 두 입력 형식, `null` → zero value, 잘못된 형식 → 에러, 쿼리 직렬화.
- **`eod`**: 저장한 실호출 fixture 로 `Meta`·`Price` 파싱, `LatestPrice` 의 배열→단일 변환과
  빈 배열 → `ErrNotFound`, `PriceOptions` 가 만드는 쿼리(zero 값 미전송, 날짜 형식,
  `Columns` 콤마 결합), 빈 ticker 에러.
- **`integration_test.go`** (`//go:build integration`): `TIINGO_API_KEY` 가 있을 때만 실제
  AAPL 메타·가격 조회로 계약 검증. 기본 `go test ./...` 에서 제외(키 의존·rate limit 회피).

## 릴리스 & 문서

- `scripts/release.sh` — fmp-go / opendart-go 와 동일 절차: main/clean 검증 →
  `go build/vet/test` → 모듈 zip 검증 → `git tag vX.Y.Z` push → `gh release create
  --generate-notes`.
- 루트 `README.md` — "구현 예정" 문구를 실제 사용법으로 교체: 설치
  (`go get github.com/kenshin579/tiingo-go@v0.1.0`), 사용 예시(`NewClientFromEnv` →
  `EOD.LatestPrice` / `HistoricalPrices`), 커버리지 표(v1: End-of-Day / Meta·Prices),
  인증(`TIINGO_API_KEY`)과 rate limit(시간·일 요청 수, 월 대역폭 — 분/초 단위 제한 없음) 안내,
  문서 카탈로그 링크.
- `examples/eod/main.go` — 실행 가능한 예시.
- 완료 후 `main` 머지 + **v0.1.0** 태그 릴리스.

## 범위 밖 (후속)

- 나머지 REST 카테고리(News, Crypto, Forex, Equity Realtime, IEX, BOATS, Fundamentals,
  Fund Fees, Dividends, Splits, Search)와 WebSocket 5종 — 각각 별도 스펙.
- CSV(`format=csv`) 응답, 재시도·rate limit 백오프, `supported_tickers.zip` 다운로드 헬퍼.
- moneyflow 통합 — v0.1.0 릴리스 후 별도 작업.

## 주의 (카탈로그에서 확인된 Tiingo 측 문서 오류)

1단계 카탈로그 코드 리뷰에서 확인된 사항. EOD 범위는 아니지만 후속 카테고리에서 그대로
믿으면 안 된다(계획 문서 `2026-09-04-api-docs-catalog.md` Task 7 말미에도 기록).

- `websockets/forex` Response 표의 Ask Size(index 7)/Ask Price(index 6)는 예시 페이로드 및
  다른 WS 페이지(size 6, price 7)와 반대 — 예시를 따를 것.
- `rest/news` bulk download 의 "End Date" JSON 필드가 `startDate` 로 오기(예시는 `endDate`).
- `rest/dividends` `trailingDiv1Y` 가 `string` 으로 표기되나 예시는 float.
- 예시 JSON 문법 오류 7곳(`rest/iex`, `rest/mutual-fund-and-etf-fees`, `rest/fundamentals`,
  `websockets/iex` 2곳, `websockets/boats`, `utilities/search`) — fixture 로 그대로 쓰지 말 것.

EOD 페이지 자체는 표·예시·`llms-full.txt` 가 서로 일치해 알려진 오류가 없다.
