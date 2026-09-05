# Tiingo Go SDK — Fundamentals 카테고리 설계

- 작성일: 2026-09-04
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/fundamentals`)
- 토픽: Tiingo Fundamentals(재무제표·일별 지표·지표 정의·회사 메타) Go 클라이언트 — 두 번째 카테고리
- 선행: v0.1.1 (기반 + End-of-Day). 설계 `2026-09-04-sdk-foundation-eod-design.md` 의 구조·규약을 그대로 따른다.
- API 참조: `docs/api/rest/fundamentals.md`, `docs/api/llms-full.txt`

## 배경 / 목적

`tiingo-go` 는 카테고리 단위로 점진 확장한다(fmp-go 패턴). v0.1.1 의 End-of-Day 에 이어 두 번째
카테고리로 **Fundamentals** 를 구현한다. moneyflow 의 US 재무 탭이 현재 FMP 무료 250콜/일 예산
게이트에 묶여 있어, 대체·보강 소스로서 실사용처가 분명하다.

## 사전 조사 결과 (실호출 확인, 2026-09-04)

`TIINGO_API_KEY` 로 네 엔드포인트를 모두 호출해 확인한 사실이다. 문서만으로는 알 수 없던 부분이
있어 설계 근거로 삼는다.

| 엔드포인트 | 결과 |
|---|---|
| `GET /tiingo/fundamentals/definitions` | 200. **85개** 지표 — balanceSheet 26, overview 23, incomeStatement 22, cashFlow 14. `units` 는 `$` 64, `%` 8, 빈 값 5, `null` 8 |
| `GET /tiingo/fundamentals/meta?tickers=aapl` | 200. 문서 표의 16필드 + **`dataProviderPermaTicker`(문서에 없음)** = 17필드 |
| `GET /tiingo/fundamentals/<ticker>/daily` | 200. `date`(RFC3339), `marketCap`, `enterpriseVal`, `peRatio`, `pbRatio`, `trailingPEG1Y` |
| `GET /tiingo/fundamentals/<ticker>/statements` | 200. `date`(**YYYY-MM-DD 평문**), `year`, `quarter`, `statementData{balanceSheet, incomeStatement, cashFlow, overview}` |

- **`statementData` 는 기간마다 지표 수가 다르다.** 2026 Q3 실측: balanceSheet 26, incomeStatement 22,
  cashFlow 14, overview 14. 각 원소는 `{dataCode, value}` 쌍이다.
- **`asReported=true` 는 건수와 날짜가 다르다.** 같은 `startDate=2026-01-01` 로 false 는 2건(첫 date
  `2026-06-27`, 회계기간 종료일), true 는 3건(첫 date `2026-07-31`, SEC 공개일).
- **`dataCode` 는 모두 `^[A-Za-z][A-Za-z0-9]*$`** 형태다(비표준 문자 없음) — Go 상수명으로 기계 변환이 안전하다.
- `statementLastUpdated` / `dailyLastUpdated` 는 `2026-08-21T01:01:17.444Z` 처럼 **시각까지 의미 있는 값**이다.
- 문서가 명시한다: "새 지표가 계속 추가되니 코드가 이를 감당해야 한다"(`docs/api/rest/fundamentals.md` 2.8.3).
- Fundamentals 는 **유료 add-on**이며 무료 플랜은 Dow 30 3년치다. 위 호출이 200 인 것은 AAPL 이
  Dow 30 이기 때문이다. 권한 밖 종목은 400/403 이 온다.

## 결정 사항 (브레인스토밍)

1. **`statementData` 는 원문 그대로 + 조회 헬퍼 + 코드 상수** (선택지 C).
   `[]DataPoint{DataCode, Value}` 를 그대로 두고 `Get`/`Map` 헬퍼와 `codes.go` 상수를 제공한다.
   (기각: 85개 지표를 타입 필드로 전개 — 문서가 경고한 지표 추가마다 라이브러리 수정이 필요하고,
   없는 지표를 0 과 구분할 수 없다. 헬퍼 없는 원문 그대로 — 사용자가 매번 문서를 뒤져야 한다.)
2. **메서드는 엔드포인트와 1:1 인 4개**, `Meta` 는 가변 인자. (기각: `LatestDaily` 같은 편의 래퍼 —
   기간 조회가 기본 용도라 필요가 확인된 뒤 추가. `Meta`/`MetaAll` 분리 — 복수 조회에 또 다른
   메서드가 필요해진다.)
3. **`types.Time` 신설.** `statementLastUpdated` 계열은 증분 동기화의 기준이라 시각을 버리면 못 쓴다.
   `Date` 의 doc 주석이 이미 "시각이 의미 있는 필드는 `time.Time` 을 쓴다"고 안내하는 것과 일치한다.
4. **필드 주석 규약 유지**: 짧은 설명은 필드 오른쪽 줄 끝, 길면 필드 위. 한국어. 타입·메서드에 doc 주석.
5. **`dataProviderPermaTicker` 를 포함한다.** 문서 표에는 없지만 실제로 오고, EOD 때와 달리 실호출로
   확인했다.
6. **무료 플랜 제약은 코드가 아니라 문서로 다룬다.** 권한 오류는 `APIError` 상태 코드 그대로 전달한다.

## 아키텍처

```
tiingo-go/
├── client.go                     # (수정) Fundamentals *fundamentals.Client 필드 추가
├── types/
│   ├── date.go                   # 기존
│   └── time.go                   # (신규) Time — 시각 보존 타임스탬프
├── fundamentals/
│   ├── client.go                 # fundamentals.New(hc) *Client
│   ├── definitions.go            # Definition + Definitions()
│   ├── meta.go                   # Meta + Meta(ctx, tickers...)
│   ├── statements.go             # Statement, StatementData, DataPoint, StatementOptions + Statements()
│   ├── daily.go                  # DailyMetric, DailyOptions + Daily()
│   ├── codes.go                  # dataCode 상수 85개(그룹별)
│   ├── *_test.go
│   └── testdata/*.json           # 실호출 fixture 4종
├── examples/fundamentals/main.go # 실행 가능한 예시
└── integration_test.go           # (수정) Fundamentals 4건 추가
```

`eod` 와 동일하게 `internal/httpclient` 를 통해 호출하고, `types` 를 직접 import 한다(루트 import 는 순환).

## 데이터 모델

```go
// DataPoint 는 지표 하나의 코드와 값이다.
type DataPoint struct {
	DataCode string  `json:"dataCode"` // 지표 식별자(예: revenue). codes.go 의 상수 참고
	Value    float64 `json:"value"`    // 지표 값
}

// StatementData 는 한 기간의 재무 데이터를 네 묶음으로 나눈 것이다.
// 지표 집합은 기간마다 다르고 Tiingo 가 계속 추가하므로 코드→값 조회로 접근한다.
type StatementData struct {
	BalanceSheet    []DataPoint `json:"balanceSheet"`    // 재무상태표 지표(26종 기준)
	IncomeStatement []DataPoint `json:"incomeStatement"` // 손익계산서 지표(22종 기준)
	CashFlow        []DataPoint `json:"cashFlow"`        // 현금흐름표 지표(14종 기준)
	Overview        []DataPoint `json:"overview"`        // 여러 재무제표를 조합한 비율·지표(23종 기준)
}

// Get 은 네 묶음 전체에서 코드를 찾는다. 없으면 ok 가 false 다(값 0 과 구분된다).
func (s StatementData) Get(code string) (value float64, ok bool)

// Map 은 코드→값 맵을 만든다. 여러 지표를 반복 조회할 때 쓴다.
func (s StatementData) Map() map[string]float64

// Statement 는 한 회계기간의 재무제표다.
type Statement struct {
	Date          types.Date    `json:"date"`    // asReported=false 면 회계기간 종료일, true 면 SEC 공개일
	Year          int           `json:"year"`    // 회계연도
	Quarter       int           `json:"quarter"` // 0 이면 연간 보고서, 1~4 는 해당 분기
	StatementData StatementData `json:"statementData"`
}

// Definition 은 지표 하나의 정의다. Definitions 로 전체 목록을 받는다.
type Definition struct {
	DataCode      string `json:"dataCode"`      // 지표 식별자(예: peRatio). codes.go 상수와 같은 값
	Name          string `json:"name"`          // 사람이 읽는 이름(예: Revenue Per Share)
	Description   string `json:"description"`   // 지표 설명
	StatementType string `json:"statementType"` // 소속 묶음: balanceSheet / incomeStatement / cashFlow / overview
	Units         string `json:"units"`         // 단위. "$" 금액, "%" 비율, 빈 값이면 무차원(배수 등)
}

// Meta 는 펀더멘털 커버리지에 포함된 회사의 메타 정보다.
// 섹터·산업·소재지는 Fundamentals 권한이 있어야 채워진다(EOD 메타에는 없다).
type Meta struct {
	PermaTicker             string     `json:"permaTicker"`             // 티커 변경·재사용에도 안정적인 영구 식별자
	Ticker                  string     `json:"ticker"`                  // 티커(소문자로 온다)
	Name                    string     `json:"name"`                    // 회사명
	IsActive                bool       `json:"isActive"`                // 현재 상장·거래 중인지
	IsADR                   bool       `json:"isADR"`                   // 미국 예탁증서(ADR) 여부
	Sector                  string     `json:"sector"`                  // 섹터(예: Technology)
	Industry                string     `json:"industry"`                // 산업(예: Consumer Electronics)
	SICCode                 int        `json:"sicCode"`                 // SEC 표준산업분류 코드
	SICSector               string     `json:"sicSector"`               // SIC 기준 섹터
	SICIndustry             string     `json:"sicIndustry"`             // SIC 기준 산업
	ReportingCurrency       string     `json:"reportingCurrency"`       // 보고 통화(예: usd)
	Location                string     `json:"location"`                // 소재지(예: California, USA)
	CompanyWebsite          string     `json:"companyWebsite"`          // 회사 홈페이지
	SECFilingWebsite        string     `json:"secFilingWebsite"`        // 이 회사의 SEC EDGAR 공시 목록 URL
	StatementLastUpdated    types.Time `json:"statementLastUpdated"`    // 재무제표가 마지막으로 갱신된 시각
	DailyLastUpdated        types.Time `json:"dailyLastUpdated"`        // 일별 지표가 마지막으로 갱신된 시각
	DataProviderPermaTicker string     `json:"dataProviderPermaTicker"` // 원천 제공자 쪽 식별자(문서 표에는 없으나 실제로 온다)
}

// DailyMetric 은 주가 기반 일별 지표다. 재무제표와 달리 매일 갱신된다.
type DailyMetric struct {
	Date          types.Date `json:"date"`          // 지표 기준일
	MarketCap     float64    `json:"marketCap"`     // 시가총액
	EnterpriseVal float64    `json:"enterpriseVal"` // 기업가치(EV)
	PERatio       float64    `json:"peRatio"`       // 주가수익비율(P/E)
	PBRatio       float64    `json:"pbRatio"`       // 주가순자산비율(P/B)
	TrailingPEG1Y float64    `json:"trailingPEG1Y"` // 최근 1년 기준 PEG
}
```

`Get` 이 네 묶음을 모두 뒤지는 이유: 사용자가 `revenue` 가 손익계산서인지 개요인지 외울 필요가 없게
한다. 묶음을 특정하려면 `s.IncomeStatement` 를 직접 순회하면 된다.

## 메서드와 옵션

```go
// Definitions 는 사용 가능한 지표 정의 전체를 받는다. GET /tiingo/fundamentals/definitions
// 지표가 추가되면 목록도 늘어난다(2026-09-04 기준 85개).
func (c *Client) Definitions(ctx context.Context) ([]Definition, error)

// Meta 는 펀더멘털 커버리지 회사의 메타 정보를 받는다. GET /tiingo/fundamentals/meta
// tickers 를 주지 않으면 커버리지 전체(수천 건)를 받는다 — 대량 응답이므로 주의한다.
func (c *Client) Meta(ctx context.Context, tickers ...string) ([]Meta, error)

// Statements 는 분기·연간 재무제표를 받는다. GET /tiingo/fundamentals/<ticker>/statements
func (c *Client) Statements(ctx context.Context, ticker string, opts *StatementOptions) ([]Statement, error)

// Daily 는 주가 기반 일별 지표를 받는다. GET /tiingo/fundamentals/<ticker>/daily
func (c *Client) Daily(ctx context.Context, ticker string, opts *DailyOptions) ([]DailyMetric, error)
```

```go
// StatementOptions 는 Statements 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type StatementOptions struct {
	StartDate time.Time // 조회 시작일(이상, >=). zero 면 미전송
	EndDate   time.Time // 조회 종료일(이하, <=). zero 면 미전송
	Sort      string    // 정렬 컬럼. Statements 는 "date" / "-date" 만 지원한다
	// AsReported 가 true 면 SEC 공개 시점의 원본 수치를 받는다(date 는 공개일).
	// false(기본)면 수정·재작성을 반영한 최신 수치이고 date 는 회계기간 종료일이다.
	// 같은 기간을 조회해도 두 모드의 건수와 날짜가 다르다(실측: 2건 vs 3건).
	AsReported bool
}

// DailyOptions 는 Daily 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type DailyOptions struct {
	StartDate time.Time // 조회 시작일(이상, >=). zero 면 미전송
	EndDate   time.Time // 조회 종료일(이하, <=). zero 면 미전송
	Sort      string    // 정렬 컬럼. "date" 오름차순, "-date" 내림차순
}
```

- `AsReported` 는 false 일 때 파라미터를 보내지 않는다(Tiingo 기본값과 같음). true 일 때만 `asReported=true`.
- `Meta` 의 가변 인자는 하나 이상이면 `tickers=aapl,msft` 로 합친다. 빈 문자열이 섞이면 에러.
- 티커는 `strings.TrimSpace` 후 빈 값이면 에러, URL 경로에는 `url.PathEscape`.
- 옵션 포인터가 nil 이면 쿼리를 붙이지 않는다(EOD `PriceOptions` 와 동일).

## `types.Time`

```go
// Time 은 시각까지 의미 있는 Tiingo 타임스탬프다(예: statementLastUpdated).
// 날짜만 필요한 필드는 Date 를 쓴다.
type Time struct {
	time.Time
}
```

- `UnmarshalJSON`/`UnmarshalText`: RFC3339 를 UTC 로 파싱해 **시각을 보존**한다. `null`·빈 문자열은
  zero value. 그 외는 에러.
- `MarshalJSON`/`MarshalText`/`String`: RFC3339(`2006-01-02T15:04:05Z07:00`). zero value 는 JSON `null`,
  텍스트는 빈 문자열 — `Date` 와 같은 규칙.
- `Date` 와의 유일한 차이는 시각을 버리지 않는 것이다.

## codes.go

85개 상수를 `statementType` 그룹별로 둔다. 값은 `dataCode`, 주석은 definitions 의 `name` 과 단위다.

```go
// 손익계산서(incomeStatement)
const (
	CodeRevenue   = "revenue"  // Revenue — 매출($)
	CodeEBITDA    = "ebitda"   // EBITDA($)
	// ...
)
```

- 상수는 definitions 실호출 결과로 일회성 생성 후 커밋한다. `dataCode` 가 모두 영숫자라 상수명 변환이
  안전하다(`Code` + UpperCamel).
- **상수가 없어도 API 호출은 막히지 않는다.** `Get`/`Map` 은 임의 문자열을 받으므로, 지표가 추가돼도
  라이브러리는 계속 동작하고 상수는 다음 릴리스에서 따라가면 된다. 이 점을 파일 상단 주석에 적는다.
- fixture 의 `dataCode` 전부가 상수로 존재하는지 검사하는 테스트를 둔다(지표 추가 시 누락 감지).

## 무료 플랜 제약

Fundamentals 는 별도 구독이며 무료 플랜은 Dow 30 3년치다. 권한 밖 종목은 400/403 이 오고, 라이브러리는
`APIError` 로 상태 코드를 그대로 전달한다. 코드로 우회하지 않는다. 이 사실을 패키지 doc 주석과 README
커버리지 표 각주에 적고, 통합 테스트는 AAPL 만 쓴다.

## 테스트

- **fixture 4종**을 실호출로 저장한다(definitions 85건, meta AAPL, statements AAPL, daily AAPL).
  EOD 때와 달리 문서 예시가 아니라 실제 응답이다.
- **단위 테스트**: 네 메서드의 파싱, `Get`(존재·부재·값 0 구분), `Map`, 옵션 쿼리 생성(zero 값 생략,
  `asReported` 는 true 일 때만, 날짜 형식, `tickers` 콤마 결합), 빈 티커·빈 tickers 원소 에러,
  403 → `APIError`.
- **codes 일치 테스트**: definitions fixture 의 `dataCode` 85개가 모두 상수에 있는지.
- **`types.Time`**: RFC3339 파싱(시각 보존), `null`/빈 문자열, 잘못된 형식 에러, 직렬화 왕복.
- **integration**(`//go:build integration`): AAPL 로 네 엔드포인트 실호출. 기존 EOD 4건에 이어 추가.

## 릴리스

머지 후 **v0.2.0** (새 카테고리 추가 — API 확장이므로 마이너 증가). README 커버리지 표에 Fundamentals
4행 추가, 예제 링크 추가.

## 범위 밖 (후속)

- CSV 응답(`format=csv`), 재시도·백오프.
- `permaTicker` 전용 메서드 — 현재 시그니처가 티커 자리에 그대로 받으므로 불필요.
- 나머지 REST 카테고리(News, Crypto, Forex, IEX, BOATS, Fund Fees, Dividends, Splits, Search)와 WebSocket.
- moneyflow 통합.

## 주의

- `statements` 의 `date` 는 평문 `YYYY-MM-DD`, `daily` 의 `date` 는 RFC3339 다. `types.Date` 가 두 형식을
  모두 받으므로 그대로 쓴다(v0.1.1 에서 오프셋 타임스탬프의 하루 밀림도 수정됨).
- `statementData` 의 네 묶음은 **키가 없을 수도 있다**고 가정한다(권한·기간에 따라). nil 슬라이스로
  디코딩되며 `Get`/`Map` 은 nil 을 안전하게 건너뛴다.
- 문서 표에 없는 `dataProviderPermaTicker` 처럼, 앞으로도 실호출과 문서가 어긋날 수 있다. fixture 는
  실호출로 갱신하는 것을 원칙으로 한다.
