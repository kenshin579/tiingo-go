# Tiingo Go SDK — Forex 카테고리 설계

- 작성일: 2026-09-05
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/forex`)
- 토픽: Tiingo Forex(통화쌍 호가·시세) Go 클라이언트 — 네 번째 카테고리
- 선행: v0.3.0 (기반 + End-of-Day + Fundamentals + Crypto). 앞선 세 설계의 구조·규약을 따른다.
- API 참조: `docs/api/rest/forex.md`

## 배경 / 목적

카테고리 단위 점진 확장의 네 번째. Crypto 바로 다음으로 고른 이유는 두 API 가 겉보기에 비슷해
(호가 + 리샘플 시세) 방금 만든 패턴을 재사용할 수 있다고 봤기 때문이다. 실제로 조사해 보니 **응답
구조가 오히려 더 단순**해서 Crypto 의 중첩 대응이 필요 없다(아래 «Crypto 와의 차이»).

## 사전 조사 결과 (실호출 확인, 2026-09-05)

| 엔드포인트 | 결과 |
|---|---|
| `GET /tiingo/fx/top?tickers=` | 200. **평탄한 배열** — `{ticker, quoteTimestamp, bidPrice, bidSize, askPrice, askSize, midPrice}` |
| `GET /tiingo/fx/<ticker>/top` | 200. 위와 **완전히 같은 응답**(경로만 다름) |
| `GET /tiingo/fx/<ticker>/prices` | 200. **평탄한 배열** — `{date, ticker, open, high, low, close}` |

### Crypto 와의 차이 (설계를 가르는 지점)

1. **응답이 평탄하다.** Crypto 는 티커별로 한 겹 감싸(`priceData[]`, `topOfBookData[]`) `PricesFor`
   같은 편의 메서드가 필요했지만, Forex 는 각 행에 `ticker` 가 붙은 평탄한 배열이라 덧씌울 것이 없다.
2. **복수 티커를 URL 경로에 콤마로 넣는다.** `/fx/eurusd,usdjpy/prices` 가 200 이고 응답에 두 티커가
   섞여 온다(실측 4행, `{'eurusd','usdjpy'}`). Crypto 는 쿼리 파라미터였다.
3. **`top` 은 `tickers` 가 사실상 필수**다. 없이 호출하면 400 `{"detail":"Error: Please pass a
   \"tickers\" parameter."}`. Crypto 의 메타가 전체를 돌려준 것과 다르다. 카탈로그 표는 선택(req=N)
   이라 적지만 실제는 아니다.
4. **거래량이 없다.** 시세 필드는 `open/high/low/close` 뿐이다. FX 는 중앙 거래소가 없어 Tiingo 가
   거래량을 제공하지 않는다(Crypto 는 `volume`, `volumeNotional`, `tradesDone` 이 있었다).
5. **호가에 체결 정보가 없다.** `midPrice` 가 있고 `lastPrice`/`lastExchange` 는 없다. 순수 호가
   스냅샷이므로 Crypto 의 `TopOfBookData` 와 같은 이름을 쓰면 혼동된다.

### 그 밖의 실측 사실

- **`date` 는 시각이다.** `resampleFreq=1min` 으로 평일(2026-09-03 목)을 조회하면 1425행이 오고
  `2026-09-03T00:00:00.000Z`, `...T00:01:00.000Z` 처럼 분 단위다 → `types.Time` 을 쓴다.
- **주말은 빈 배열이다.** 2026-09-05(토)·09-06(일) 조회는 `[]`. FX 시장이 닫혀 있기 때문이다.
- **없는 티커도 200 + 빈 배열이다.** `nosuchpair` 조회가 에러가 아니라 `[]` 다. 따라서 호출자는
  "없는 통화쌍"과 "시장 휴장"을 응답만으로 구분할 수 없다 — 이 사실을 doc 주석에 명시한다.
- 호가 타임스탬프는 `2026-09-05T07:15:03.205000+00:00`(마이크로초, `+00:00` 오프셋).

## 결정 사항 (브레인스토밍)

1. **응답 구조를 그대로 유지하고 편의 메서드를 두지 않는다**(선택지 A). 메서드는 둘 뿐이다.
   (기각: `GroupByTicker` 헬퍼 — API 표면이 늘고 정렬·순서 보장 질문이 따라온다. Crypto 처럼
   티커별 구조로 재조립 — 실제 응답과 달라져 문서 대조가 어렵고 변환 비용만 생긴다.)
2. **`top` 은 배치 경로(`/fx/top?tickers=`)만 쓴다.** 단일·복수 응답이 동일함을 실측했으므로 코드
   경로를 하나로 둔다. (기각: `/fx/<ticker>/top` 도 `TopOfBookFor` 로 노출 — 반환 타입까지 같아
   존재 이유가 없다.)
3. **호가 타입 이름은 `Quote`.** Crypto 의 `TopOfBookData` 와 필드가 다르므로 같은 이름을 피한다.
   메서드 이름은 엔드포인트에 맞춰 `TopOfBook` 을 쓴다.
4. **`Price.Date`·`Quote.QuoteTimestamp` 는 `types.Time`**. 실측대로 시각이 의미 있다.
5. **`ResampleFreq` 는 Crypto 와 같은 방식** — `string` 별칭 + 자주 쓰는 상수. 문서가 임의의
   `<숫자><단위>` 조합을 허용한다.
6. **필드 주석 규약 유지**, fixture 는 실호출.

## 아키텍처

```
tiingo-go/
├── client.go             # (수정) Forex *forex.Client 필드 추가
├── forex/
│   ├── client.go         # forex.New(hc) *Client
│   ├── top.go            # Quote + TopOfBook(ctx, tickers)
│   ├── prices.go         # Price, PriceOptions, ResampleFreq + Prices(ctx, tickers, opts)
│   ├── *_test.go
│   └── testdata/*.json   # 실호출 fixture 4종
├── examples/forex/main.go
└── integration_test.go   # (수정) Forex 4건 추가
```

`crypto` 와 동일하게 `internal/httpclient` 로 호출하고 `types` 를 직접 import 한다. 티커 검증·결합은
`crypto/tickers.go` 와 같은 로직이 필요하지만 **패키지가 다르므로 각자 둔다**(내부 헬퍼를 공유하려면
`internal` 패키지로 빼야 하는데, 함수 하나를 위해 그럴 값이 없다 — 후속으로 카테고리가 더 늘면 재검토).

## 데이터 모델

```go
// Quote 는 통화쌍 호가 스냅샷이다. 체결 정보 없이 매수·매도·중간가만 있다.
type Quote struct {
	Ticker         string     `json:"ticker"`         // 통화쌍(예: eurusd)
	QuoteTimestamp types.Time `json:"quoteTimestamp"` // 호가 기준 시각
	BidPrice       float64    `json:"bidPrice"`       // 매수 호가
	BidSize        float64    `json:"bidSize"`        // 매수 호가 수량
	AskPrice       float64    `json:"askPrice"`       // 매도 호가
	AskSize        float64    `json:"askSize"`        // 매도 호가 수량
	MidPrice       float64    `json:"midPrice"`       // 중간가
}

// Price 는 리샘플 구간 하나의 시세다. FX 는 중앙 거래소가 없어 거래량이 제공되지 않는다.
type Price struct {
	Date   types.Time `json:"date"`   // 구간 시작 시각. 1day 면 자정, 1min 이면 분 단위
	Ticker string     `json:"ticker"` // 통화쌍. 복수 조회 시 행마다 다르다
	Open   float64    `json:"open"`   // 시가
	High   float64    `json:"high"`   // 고가
	Low    float64    `json:"low"`    // 저가
	Close  float64    `json:"close"`  // 종가
}
```

## 메서드와 옵션

```go
// TopOfBook 은 통화쌍 호가를 받는다. GET /tiingo/fx/top
// tickers 는 하나 이상이어야 한다 — 없이 호출하면 Tiingo 가 400 을 돌려준다.
func (c *Client) TopOfBook(ctx context.Context, tickers []string) ([]Quote, error)

// Prices 는 리샘플 시세를 받는다. GET /tiingo/fx/<tickers>/prices
// 복수 티커는 경로에 콤마로 들어가고, 응답의 각 행에 Ticker 가 붙는다.
// 시장이 닫혀 있거나(주말) 없는 통화쌍이면 에러가 아니라 빈 슬라이스를 돌려준다 — 둘은 구분되지 않는다.
func (c *Client) Prices(ctx context.Context, tickers []string, opts *PriceOptions) ([]Price, error)
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
}
```

- 티커는 `strings.TrimSpace` 후 빈 값이면 에러, 슬라이스가 비어도 에러(두 엔드포인트 모두 필수).
- `Prices` 는 티커를 콤마로 합쳐 **경로**에 넣고 `url.PathEscape` 한다. 콤마는 경로에서 이스케이프되지
  않아야 하므로 각 티커를 개별 escape 한 뒤 콤마로 join 한다.
- `TopOfBook` 은 티커를 콤마로 합쳐 **쿼리** `tickers` 에 넣는다.
- `StartDate`/`EndDate` 는 `YYYY-MM-DD`.

## 테스트

- **fixture 4종**(실호출): `top_eurusd.json`, `top_multi.json`(eurusd,usdjpy),
  `prices_eurusd_1day.json`, `prices_eurusd_1min.json`(평일 2026-09-03, 1425행).
- **단위 테스트**: 두 메서드의 파싱, 경로·쿼리 생성(콤마 결합, 개별 escape, zero 값 생략, 날짜 형식),
  빈 티커·빈 슬라이스 에러, 400 → `APIError`, 빈 배열이 에러가 아님(주말·없는 티커).
- **시각 보존 테스트**: 1min fixture 의 첫 두 행이 1분 차이임을 확인한다(`types.Time` 근거 고정).
- **integration**: EURUSD·USDJPY 로 두 메서드 실호출 + 주말·없는 티커의 빈 배열 확인. 기존 13건에 추가.
  주말에 돌려도 실패하지 않도록 시세 테스트는 **최근 7일 범위**로 조회해 빈 응답을 허용한다.

## 릴리스

머지 후 **v0.4.0**. README 커버리지 표에 Forex 2행 추가.

## 범위 밖 (후속)

- Dividends/Splits — 권한 403 으로 보류 중(crypto 스펙 «전환 경위» 참조).
- 나머지 REST(News, Equity Realtime, IEX, BOATS, Fund Fees, Search)와 WebSocket 5종.
- CSV, 재시도, moneyflow 통합.
- 티커 검증 헬퍼의 패키지 간 공유 — 카테고리가 더 늘면 `internal` 로 추출 재검토.

## 주의

- 카탈로그 표가 `top` 의 `tickers` 를 선택(req=N)이라 적지만 **실제로는 필수**다(400). 문서보다 실측을
  따른다.
- 빈 배열이 정상 응답이라 `ErrNotFound` 를 쓰지 않는다. Crypto 의 `PricesFor` 는 중첩 구조에서 첫
  요소를 꺼내는 성격이라 `ErrNotFound` 가 맞았지만, 여기서는 빈 결과가 곧 "그 기간에 데이터 없음"이다.
