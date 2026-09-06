# tiingo-go

[Tiingo](https://www.tiingo.com) 금융 데이터 API 의 Go 클라이언트 라이브러리.

## 설치

```bash
go get github.com/kenshin579/tiingo-go@latest
```

Go 1.25+. 런타임 의존성 없음(테스트만 testify).

## 사용

```go
c, err := tiingo.NewClientFromEnv() // TIINGO_API_KEY
if err != nil {
    log.Fatal(err)
}
ctx := context.Background()

// 자산 메타
m, _ := c.EOD.Meta(ctx, "AAPL")
fmt.Println(m.Name, m.ExchangeCode, m.StartDate, m.EndDate)

// 최신 종가
p, _ := c.EOD.LatestPrice(ctx, "AAPL")
fmt.Println(p.Date, p.Close, p.AdjClose)

// 기간 조회(주별 리샘플, 내림차순)
ps, _ := c.EOD.HistoricalPrices(ctx, "AAPL", &eod.PriceOptions{
    StartDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    ResampleFreq: eod.ResampleWeekly,
    Sort:         "-date",
})

// 재무제표(최신 기간부터). 지표는 코드로 조회한다 — 신규 지표가 계속 추가되기 때문
ss, _ := c.Fundamentals.Statements(ctx, "AAPL", &fundamentals.StatementOptions{Sort: "-date"})
rev, ok := ss[0].StatementData.Get(fundamentals.CodeRevenue)

// 암호화폐 일봉(단일 페어는 PricesFor 로 바로 받는다)
cs, _ := c.Crypto.PricesFor(ctx, "btcusd", &crypto.PriceOptions{ResampleFreq: crypto.Resample1Day})

// 통화쌍 호가(복수 조회 가능)
qs, _ := c.Forex.TopOfBook(ctx, []string{"eurusd", "usdjpy"})

// IEX 실시간 스냅샷(장 마감 시 호가 필드는 nil)
iqs, _ := c.IEX.Quotes(ctx, []string{"AAPL", "MSFT"})

// 자산 검색(티커·이름). 같은 티커가 국가별로 중복될 수 있어 PermaTicker 로 구분한다
rs, _ := c.Search.Search(ctx, "apple", &search.SearchOptions{Limit: 5})

// ISIN 으로 자산 하나를 지목한다(문서 파라미터 표에는 없지만 동작한다)
byISIN, _ := c.Search.SearchByISIN(ctx, "US0378331005", nil)

// 통합 피드 스냅샷(여러 거래소·ATS·OTC). 유동성 지표는 없을 수 있어 포인터다
es, _ := c.Equity.Snapshots(ctx, []string{"AAPL", "SPY"})

// 배당수익률 시계열. 옵션 없이 부르면 상장 이후 전 기간이 온다
ys, _ := c.CorporateActions.DistributionYield(ctx, "AAPL",
    &corporateactions.YieldOptions{StartDate: time.Now().AddDate(0, -1, 0)})
```

실행 가능한 예시: [`examples/eod`](examples/eod), [`examples/fundamentals`](examples/fundamentals), [`examples/crypto`](examples/crypto), [`examples/forex`](examples/forex), [`examples/iex`](examples/iex), [`examples/search`](examples/search), [`examples/equity`](examples/equity), [`examples/corporateactions`](examples/corporateactions), [`examples/stream`](examples/stream).

## 인증

[Tiingo 계정](https://www.tiingo.com/account/api/token)에서 발급한 토큰을 `TIINGO_API_KEY`
환경변수로 두거나 `tiingo.NewClient(apiKey)` 로 넘긴다. 토큰은 `Authorization: Token <key>`
헤더로 전송되며 URL 쿼리에 실리지 않는다.

## 커버리지

| 그룹 | 메서드 | 엔드포인트 |
| --- | --- | --- |
| End-of-Day | `EOD.Meta` | `GET /tiingo/daily/<ticker>` |
| End-of-Day | `EOD.LatestPrice` | `GET /tiingo/daily/<ticker>/prices` |
| End-of-Day | `EOD.HistoricalPrices` | `GET /tiingo/daily/<ticker>/prices` |
| Fundamentals\* | `Fundamentals.Definitions` | `GET /tiingo/fundamentals/definitions` |
| Fundamentals\* | `Fundamentals.Meta` | `GET /tiingo/fundamentals/meta` |
| Fundamentals\* | `Fundamentals.Statements` | `GET /tiingo/fundamentals/<ticker>/statements` |
| Fundamentals\* | `Fundamentals.Daily` | `GET /tiingo/fundamentals/<ticker>/daily` |
| Crypto | `Crypto.Meta` | `GET /tiingo/crypto` |
| Crypto | `Crypto.Prices` / `PricesFor` | `GET /tiingo/crypto/prices` |
| Crypto | `Crypto.TopOfBook` / `TopOfBookFor` | `GET /tiingo/crypto/top` |
| Forex | `Forex.TopOfBook` | `GET /tiingo/fx/top` |
| Forex | `Forex.Prices` | `GET /tiingo/fx/<tickers>/prices` |
| IEX | `IEX.Quotes` | `GET /iex/` |
| IEX | `IEX.Prices` | `GET /iex/<ticker>/prices` |
| Search | `Search.Search` | `GET /tiingo/utilities/search` |
| Search | `Search.SearchByISIN` | `GET /tiingo/utilities/search` |
| Equity Realtime | `Equity.Snapshots` | `GET /tiingo/equity/intraday/` |
| Equity Realtime | `Equity.AllSnapshots` | `GET /tiingo/equity/intraday/` |
| Equity Realtime | `Equity.Prices` | `GET /tiingo/equity/intraday/<ticker>/prices` |
| Corporate Actions\*\* | `CorporateActions.DistributionYield` | `GET /tiingo/corporate-actions/<ticker>/distribution-yield` |

\* Fundamentals 는 별도 구독(add-on)이다. 무료 플랜은 Dow 30 종목의 3년치만 제공하며, 권한 밖
종목은 `APIError`(400/403)로 돌아온다.

\*\* Corporate Actions 는 이 그룹 5개 중 1개만 구현돼 있다. 배당 내역(distributions, 티커별·배치)과
분할(splits, 티커별·배치)은 무료 키에서 403 이라 응답 형태를 확인할 수 없어 넣지 않았다.

나머지 REST 그룹은 이 계정 권한으로 접근이 막혀 있다(2026-09-05 실측) — News 와 Fund Fees 는
403 권한 없음, BOATS 는 유료 add-on, Corporate Actions 의 배당 내역·분할도 403 이다.
WebSocket 5종은 아래 절에 있다.

## WebSocket

실시간 스트리밍은 REST 와 별개 연결이며 `c.Stream` 으로 접근한다. 채널로 소비하고,
채널이 닫힌 뒤 `Err()` 로 종료 사유를 확인한다.

```go
s, _ := c.Stream.Crypto(ctx, &stream.CryptoOptions{
    Tickers:   []string{"btcusd"},
    Threshold: stream.CryptoTradesAndQuotes,
})
defer s.Close()

for msg := range s.Messages() {
    switch m := msg.(type) {
    case stream.CryptoTrade:
        fmt.Println(m.Ticker, m.Exchange, m.LastPrice)
    case stream.CryptoQuote:
        fmt.Println(m.Ticker, m.BidPrice, m.AskPrice)
    }
}
if err := s.Err(); err != nil { /* ctx 취소·서버 거부 등 */ }
```

| 피드 | 메서드 | 엔드포인트 | 배열 매핑 검증 |
| --- | --- | --- | --- |
| Crypto | `Stream.Crypto` | `wss://api.tiingo.com/crypto` | 실호출 |
| Forex | `Stream.Forex` | `wss://api.tiingo.com/fx` | 문서 예시 산술\* |
| IEX | `Stream.IEX` | `wss://api.tiingo.com/iex` | 문서 |
| Equity Realtime | `Stream.Equity` | `wss://api.tiingo.com/equity/intraday` | 문서 |
| BOATS | `Stream.BOATS` | `wss://api.tiingo.com/boats` | 불가(계정 권한 403) |

\* Tiingo 의 forex 문서는 인덱스 표가 틀렸다 — 표는 askPrice=6/askSize=7 이라 적지만 같은 문서의
예시에서 `mid=(bid+ask)/2` 검산은 askSize=6/askPrice=7 일 때만 맞는다. 코드는 예시를 따른다.

데이터가 필드명이 아니라 **배열 위치**로 오므로, 배열이 기대보다 짧거나 타입이 다르면 값을
조용히 바꾸는 대신 에러를 낸다. 기대보다 긴 배열은 통과시킨다(Tiingo 가 필드를 추가할 수 있다).
Crypto 외 네 피드의 매핑은 조사 시점에 FX·미국 주식 시장이 닫혀 있어 실호출로 확인하지
못했다. 각 타입 주석에 검증 여부를 적어 두었다.

연결이 끊기면 지수 backoff 로 자동 재연결하고 구독을 재전송한다. 서버가 30초마다 보내는
하트비트가 90초(`WithReadTimeout`) 동안 끊기면 죽은 연결로 보고 다시 붙는다. 재연결 횟수는
`s.Reconnects()` 로 확인한다 — 값이 오르면 그 사이 메시지가 빠졌을 수 있다.

소비가 느려 버퍼가 차면 가장 오래된 메시지를 버리며, 누락 건수는 `s.Dropped()` 로 확인한다.
전 종목을 받을 때(`Tickers` 를 비움)는 기본 버퍼 256 이 부족할 수 있으므로 `WithBuffer` 로 늘린다.

## 날짜 타입

Tiingo 는 같은 API 에서 두 가지 날짜 형식을 쓴다 — 가격은 `2019-01-02T00:00:00.000Z`,
메타는 `1980-12-12`. `tiingo.Date`(= `types.Date`)가 둘 다 받아 `time.Time` 으로 정규화하고
`YYYY-MM-DD` 로 직렬화한다. `time.Time` 을 임베드하므로 `IsZero()`, `Before()`, `Year()` 등을
그대로 쓸 수 있다. 다만 `database/sql` 에 직접 넘길 때는 `d.Time` 을 쓴다.

`types.Time`(루트 별칭 `tiingo.Time`)은 시각까지 보존한다. `statementLastUpdated` 처럼 갱신
시각이 의미 있는 필드, 그리고 암호화폐 시세처럼 인트라데이 값이 오는 필드에 쓰며, 직렬화는
RFC3339 다. 예를 들어 `resampleFreq=1min` 시세의 `date` 는 분 단위 시각이라 `Date` 로는 표현할 수 없다.

## null 필드

IEX 스냅샷은 장 마감 시간대에 호가·체결 관련 9개 필드가 `null` 로 온다. 값 타입으로 받으면 0 과
구분되지 않으므로 해당 필드는 포인터(`*float64`, `*types.Time`)이며 `nil` 은 "값 없음"이다.
없는 티커는 에러가 아니라 응답에서 빠지고 순서도 요청과 다르므로, 결과는 `Ticker` 필드로 찾는다.

검색 결과의 `OpenFIGIComposite` 는 값이 없을 때 Tiingo 가 `null` 과 문자열 `"nan"` 을 섞어 보내므로
둘 다 빈 문자열로 정규화된다. `r.OpenFIGIComposite != ""` 하나만 확인하면 된다.

Equity Realtime 스냅샷의 유동성 5개 필드(`LqSpread`, `LqBidPrice`, `LqBidSize`, `LqAskPrice`,
`LqAskSize`)도 통합 피드가 값을 내지 않으면 `null` 이라 포인터다 — 전 종목 조회 기준 44% 가
그렇다. 이름이 비슷한 `LqRefPrice` 는 늘 채워져 값 타입이다.

## 에러 처리

```go
p, err := c.EOD.LatestPrice(ctx, "NOSUCH")
if errors.Is(err, tiingo.ErrNotFound) {
    // 결과 없음
}
var apiErr *tiingo.APIError
if errors.As(err, &apiErr) {
    // apiErr.StatusCode: 401 토큰 오류, 403 권한/플랜, 404 없는 티커, 429 rate limit
}
```

## Rate limit

Tiingo 는 시간당·일당 요청 수와 월 대역폭으로 제한하며 분/초 단위 제한은 없다. 현재 사용량은
[API usage](https://www.tiingo.com/account/api/usage) 에서 확인한다. 이 라이브러리는 재시도나
백오프를 하지 않는다(429 는 `APIError` 로 그대로 전달).

## 테스트

```bash
go test ./...                                       # 단위 테스트
TIINGO_API_KEY=... go test -tags integration ./...  # 실호출 통합 테스트
go build -o /dev/null ./examples/eod                # 예제 빌드(레포 루트의 동명 디렉터리와 겹치면 -o 필요)
go build -o /dev/null ./examples/search             # eod/·search/·equity/·corporateactions/·stream/ 가 이에 해당한다
go build -o /dev/null ./examples/equity
go build -o /dev/null ./examples/corporateactions
go build -o /dev/null ./examples/stream
```

## 문서

- [`docs/api/README.md`](docs/api/README.md) — Tiingo 문서 사이트 23페이지를 변환한 md +
  Tiingo 공식 `llms.txt`/`llms-full.txt` 원본. 재생성은 `./scripts/fetch-docs.sh` 와
  `cd tools/gendocs && npm run gen`.
- 설계·계획: [`docs/superpowers/`](docs/superpowers/)

## 라이선스

[MIT](LICENSE)
