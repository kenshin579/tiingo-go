# Tiingo Go SDK — Crypto 카테고리 설계

- 작성일: 2026-09-05
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/crypto`)
- 토픽: Tiingo Crypto(암호화폐 페어 메타·호가·시세) Go 클라이언트 — 세 번째 카테고리
- 선행: v0.2.0 (기반 + End-of-Day + Fundamentals). 설계
  `2026-09-04-sdk-foundation-eod-design.md` 와 `2026-09-04-fundamentals-design.md` 의 구조·규약을 따른다.
- API 참조: `docs/api/rest/crypto.md`

## 배경 / 목적

카테고리 단위 점진 확장의 세 번째. 원래 Dividends + Splits 를 하려 했으나 **이 계정으로 5개
엔드포인트 중 4개가 403** 이라 실호출 검증이 불가능해 보류하고, Starter 플랜에 포함돼 검증이
확실한 Crypto 로 전환했다(아래 «전환 경위»).

## 사전 조사 결과 (실호출 확인, 2026-09-05)

### 전환 경위 — Dividends/Splits 보류

| 엔드포인트 | 결과 |
|---|---|
| `GET /tiingo/corporate-actions/<ticker>/distribution-yield` | 200 |
| `GET /tiingo/corporate-actions/<ticker>/distributions` | **403** |
| `GET /tiingo/corporate-actions/<ticker>/splits` | **403** |
| `GET /tiingo/corporate-actions/distributions` (배치) | **403** |
| `GET /tiingo/corporate-actions/splits` (배치) | **403** |

- 인증 방식 문제가 아니다: `Authorization: Token` 헤더와 `?token=` 쿼리 **둘 다 403**, 같은 키로
  EOD 와 `distribution-yield` 는 200. 403 본문은 비어 있다(`content-length: 0`).
- 카탈로그 문서는 "End-of-Day 권한이 있는 모든 고객에게 제공"이라고 적지만 실제로는 막혀 있다.
  두 페이지 모두 "베타로 노출 중"이라고 밝히므로 베타 게이팅 또는 별도 활성화가 필요한 것으로 보인다.
- **판단**: 이 프로젝트의 검증 원칙(실호출 fixture + 통합 테스트로 계약 확인)을 지킬 수 없으므로
  권한이 열린 뒤로 미룬다. 문서 예시만으로 만들면 재작업 위험이 크다 — Fundamentals 의 메타 응답에
  문서 표에 없는 `dataProviderPermaTicker` 가 있었던 전례가 근거다.

### Crypto — 세 엔드포인트 모두 200

| 엔드포인트 | 응답 |
|---|---|
| `GET /tiingo/crypto?tickers=` | **평탄한 배열**. `{ticker, name, baseCurrency, quoteCurrency}` |
| `GET /tiingo/crypto/top?tickers=` | 티커별 객체 + 중첩 `topOfBookData[]`(12필드) |
| `GET /tiingo/crypto/prices?tickers=` | 티커별 객체 + 중첩 `priceData[]`(8필드) |

- **중첩 구조**: `top` 과 `prices` 는 바깥이 티커 메타(`ticker`, `baseCurrency`, `quoteCurrency`),
  안이 배열이다. 복수 티커를 주면 배열 원소가 티커 수만큼 온다(실측: `tickers=btcusd,ethusd` → 2개,
  각각 `topOfBookData` 1건). `meta` 만 평탄하다.
- **`date` 는 날짜가 아니라 시각이다**: `resampleFreq=1day` 면 `2026-09-01T00:00:00+00:00`,
  `resampleFreq=1min` 이면 `2026-09-05T00:01:00+00:00` 처럼 분 단위로 온다(실측 126행).
  따라서 `types.Date`(날짜만 남김)가 아니라 **`types.Time`** 을 써야 한다.
- 타임스탬프 형식은 `+00:00` 오프셋과 6자리 소수(`2026-09-05T01:59:29.438000+00:00`). `types.Time`
  의 RFC3339 파싱이 그대로 처리한다.
- `exchanges` 파라미터가 실제로 동작한다(실측: `exchanges=GDAX` → bid/ask 거래소가 모두 GDAX).
- 파라미터(카탈로그 표): `prices` 는 `tickers`(필수), `exchanges`, `startDate`, `endDate`,
  `resampleFreq`. `top` 은 `tickers`(필수), `exchanges`, `includeRawExchangeData`. `meta` 는
  `tickers`(선택), `format`.

## 결정 사항 (브레인스토밍)

1. **중첩 구조를 그대로 유지하고 단일 티커 편의 메서드를 얹는다**(선택지 C).
   `Prices`/`TopOfBook` 은 `[]PriceSeries`/`[]TopOfBook` 을 돌려주고, `PricesFor`/`TopOfBookFor` 가
   첫 요소를 돌려준다(없으면 `ErrNotFound`). EOD 의 `LatestPrice`→`HistoricalPrices` 와 같은 패턴.
   (기각: 평탄화 — 응답 구조를 감추고 티커 메타가 매 행에 중복된다. 편의 메서드 없음 — 단일 티커
   조회가 대부분인데 매번 `[0]` 접근이 필요하다.)
2. **메서드 5개, 필수/선택을 시그니처로 구분**. `tickers` 가 필수인 `Prices`/`TopOfBook` 은 슬라이스로
   받아 빈 값을 에러로 막고, 선택인 `Meta` 만 가변 인자다. (기각: 전부 가변 인자 — 옵션이 앞에 오는
   어색한 순서가 되고 필수 인자가 문법적으로 생략 가능해진다.)
3. **`Price.Date` 와 호가 타임스탬프는 `types.Time`**. 실측대로 시각이 의미 있다.
4. **`ResampleFreq` 는 `string` 별칭 + 상수 4개**. 문서가 임의의 `<숫자><단위>` 조합을 허용하므로
   EOD 처럼 값을 닫지 않는다. 자주 쓰는 값만 상수로 제공한다.
5. **필드 주석 규약 유지**: 짧은 설명은 필드 오른쪽 줄 끝, 길면 필드 위. 한국어.
6. **fixture 는 실호출**. Crypto 는 Starter 플랜 포함이라 다섯 종 모두 실제 응답으로 저장한다.

## 아키텍처

```
tiingo-go/
├── client.go                 # (수정) Crypto *crypto.Client 필드 추가
├── crypto/
│   ├── client.go             # crypto.New(hc) *Client
│   ├── meta.go               # Meta + Meta(ctx, tickers...)
│   ├── prices.go             # PriceSeries, Price, PriceOptions, ResampleFreq + Prices/PricesFor
│   ├── top.go                # TopOfBook, TopOfBookData, TopOptions + TopOfBook/TopOfBookFor
│   ├── *_test.go
│   └── testdata/*.json       # 실호출 fixture 5종
├── examples/crypto/main.go
└── integration_test.go       # (수정) Crypto 5건 추가
```

`eod`·`fundamentals` 와 동일하게 `internal/httpclient` 로 호출하고 `types` 를 직접 import 한다.

## 데이터 모델

```go
// Meta 는 암호화폐 페어의 메타 정보다.
type Meta struct {
	Ticker        string `json:"ticker"`        // 페어 티커(예: btcusd)
	Name          string `json:"name"`          // 사람이 읽는 이름(예: Bitcoin (BTC/USD))
	BaseCurrency  string `json:"baseCurrency"`  // 기준 통화(예: btc)
	QuoteCurrency string `json:"quoteCurrency"` // 표시 통화(예: usd)
}

// Price 는 리샘플 구간 하나의 시세다. 구간 길이는 요청의 resampleFreq 가 정한다.
type Price struct {
	Date           types.Time `json:"date"`           // 구간 시작 시각. 1day 면 자정, 1min 이면 분 단위
	Open           float64    `json:"open"`           // 시가
	High           float64    `json:"high"`           // 고가
	Low            float64    `json:"low"`            // 저가
	Close          float64    `json:"close"`          // 종가
	Volume         float64    `json:"volume"`         // 거래량(기준 통화 단위)
	VolumeNotional float64    `json:"volumeNotional"` // 거래대금(표시 통화 단위)
	TradesDone     float64    `json:"tradesDone"`     // 체결 건수
}

// PriceSeries 는 티커 하나의 시세 묶음이다. Tiingo 가 티커별로 감싸서 돌려준다.
type PriceSeries struct {
	Ticker        string  `json:"ticker"`        // 페어 티커
	BaseCurrency  string  `json:"baseCurrency"`  // 기준 통화
	QuoteCurrency string  `json:"quoteCurrency"` // 표시 통화
	PriceData     []Price `json:"priceData"`     // 시간순 시세. 요청 범위·주기에 따라 길이가 다르다
}

// TopOfBookData 는 호가창 최상단 한 시점의 스냅샷이다.
type TopOfBookData struct {
	QuoteTimestamp    types.Time `json:"quoteTimestamp"`    // 호가 기준 시각
	LastSaleTimestamp types.Time `json:"lastSaleTimestamp"` // 마지막 체결 시각
	BidSize           float64    `json:"bidSize"`           // 매수 호가 수량
	BidPrice          float64    `json:"bidPrice"`          // 매수 호가
	AskSize           float64    `json:"askSize"`           // 매도 호가 수량
	AskPrice          float64    `json:"askPrice"`          // 매도 호가
	LastSize          float64    `json:"lastSize"`          // 마지막 체결 수량
	LastSizeNotional  float64    `json:"lastSizeNotional"`  // 마지막 체결 대금(표시 통화)
	LastPrice         float64    `json:"lastPrice"`         // 마지막 체결가
	BidExchange       string     `json:"bidExchange"`       // 최우선 매수 호가를 낸 거래소
	AskExchange       string     `json:"askExchange"`       // 최우선 매도 호가를 낸 거래소
	LastExchange      string     `json:"lastExchange"`      // 마지막 체결이 일어난 거래소
}

// TopOfBook 은 티커 하나의 호가 묶음이다. Tiingo 가 티커별로 감싸서 돌려준다.
type TopOfBook struct {
	Ticker        string          `json:"ticker"`        // 페어 티커
	BaseCurrency  string          `json:"baseCurrency"`  // 기준 통화
	QuoteCurrency string          `json:"quoteCurrency"` // 표시 통화
	TopOfBookData []TopOfBookData `json:"topOfBookData"` // 호가 스냅샷. 보통 1건
}
```

## 메서드와 옵션

```go
// Meta 는 페어 메타 정보를 받는다. GET /tiingo/crypto
// tickers 를 주지 않으면 지원하는 페어 전체를 받는다 — 대량 응답이므로 주의한다.
func (c *Client) Meta(ctx context.Context, tickers ...string) ([]Meta, error)

// Prices 는 티커별 시세 묶음을 받는다. GET /tiingo/crypto/prices
func (c *Client) Prices(ctx context.Context, tickers []string, opts *PriceOptions) ([]PriceSeries, error)

// PricesFor 는 티커 하나의 시세 묶음을 받는다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) PricesFor(ctx context.Context, ticker string, opts *PriceOptions) (*PriceSeries, error)

// TopOfBook 은 티커별 호가 묶음을 받는다. GET /tiingo/crypto/top
func (c *Client) TopOfBook(ctx context.Context, tickers []string, opts *TopOptions) ([]TopOfBook, error)

// TopOfBookFor 는 티커 하나의 호가 묶음을 받는다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) TopOfBookFor(ctx context.Context, ticker string, opts *TopOptions) (*TopOfBook, error)
```

```go
// ResampleFreq 는 시세 리샘플 주기다. "5min", "4hour" 처럼 <숫자><단위> 형식이면 무엇이든 된다.
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
	Exchanges    []string     // 특정 거래소로 한정. 비어 있으면 전체
}

// TopOptions 는 TopOfBook 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type TopOptions struct {
	Exchanges []string // 특정 거래소로 한정. 비어 있으면 전체
	// IncludeRawExchangeData 가 true 면 집계 전 거래소별 원본 호가도 함께 온다.
	IncludeRawExchangeData bool
}
```

- `tickers` 는 콤마로 합쳐 보낸다. 원소가 빈 문자열이면 에러, 슬라이스 자체가 비어도 에러
  (`Prices`/`TopOfBook` 은 필수 파라미터).
- `Exchanges` 도 콤마 결합. `IncludeRawExchangeData` 는 true 일 때만 전송.
- `StartDate`/`EndDate` 는 `YYYY-MM-DD` 로 직렬화한다. 카탈로그 표는 datetime 이라 적지만 실호출로
  `startDate=2026-09-01&endDate=2026-09-03` 이 200 임을 확인했다(EOD·Fundamentals 와 같은 형식).
- 옵션 포인터가 nil 이면 필수 `tickers` 만 보낸다.

## 테스트

- **fixture 5종**(실호출): `meta_btcusd.json`, `prices_btcusd_1day.json`, `prices_btcusd_1min.json`
  (시각 보존 확인용), `top_btcusd.json`, `top_multi.json`(btcusd,ethusd 2건).
- **단위 테스트**: 다섯 메서드의 파싱, 중첩 배열 디코딩, `PricesFor`/`TopOfBookFor` 의 첫 요소 반환과
  빈 배열 → `ErrNotFound`, 옵션 쿼리 생성(zero 값 생략, `tickers`·`exchanges` 콤마 결합,
  `includeRawExchangeData` 는 true 일 때만, 날짜 형식), 빈 티커·빈 슬라이스 에러, 403 → `APIError`.
- **시각 보존 테스트**: 1min fixture 의 첫 두 행이 1분 차이임을 확인해 `types.Time` 사용이 옳음을 고정한다.
- **integration**(`//go:build integration`): BTC/USD·ETH/USD 로 다섯 메서드 실호출. 기존 8건에 이어 추가.

## 릴리스

머지 후 **v0.3.0**. README 커버리지 표에 Crypto 5행 추가, 예제 링크 추가.

## 범위 밖 (후속)

- Dividends/Splits — 권한이 열리면 재개(본 스펙 «전환 경위» 참조).
- Forex — Crypto 와 응답 구조가 매우 비슷하므로 이 카테고리 다음이 자연스럽다.
- 나머지 REST(News, Equity Realtime, IEX, BOATS, Fund Fees, Search)와 WebSocket 5종.
- CSV(`format=csv`), 재시도·백오프, moneyflow 통합.

## 주의

- **`includeRawExchangeData=true` 는 현재 Tiingo 가 HTTP 500 을 돌려준다**(2026-09-05 실측, 본문은
  HTML `text/html` 27바이트). `false` 는 200. 서버 쪽 문제로 보이므로 옵션은 문서대로 지원하되
  **통합 테스트가 이 값에 의존하지 않게** 하고, 단위 테스트로 쿼리 전송만 검증한다. 추후 서버가
  고쳐지면 응답을 실측해 추가 필드를 담는다(기본 경로 구조는 바꾸지 않는다).
- `Meta` 를 티커 없이 부르면 지원 페어 전체가 온다. `encoding/json` 은 슬라이스 원소 하나라도
  파싱에 실패하면 전체가 실패하므로, 대량 응답에서 예상 밖 값이 오면 전량 실패한다(Fundamentals
  리뷰에서 지적된 것과 같은 성질).
