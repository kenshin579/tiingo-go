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
```

실행 가능한 예시: [`examples/eod`](examples/eod), [`examples/fundamentals`](examples/fundamentals), [`examples/crypto`](examples/crypto), [`examples/forex`](examples/forex).

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

\* Fundamentals 는 별도 구독(add-on)이다. 무료 플랜은 Dow 30 종목의 3년치만 제공하며, 권한 밖
종목은 `APIError`(400/403)로 돌아온다.

나머지 REST 그룹(News, Equity Realtime, IEX, BOATS, Fund Fees, Search, Dividends/Splits 등)과
WebSocket 은 순차 추가 예정.

## 날짜 타입

Tiingo 는 같은 API 에서 두 가지 날짜 형식을 쓴다 — 가격은 `2019-01-02T00:00:00.000Z`,
메타는 `1980-12-12`. `tiingo.Date`(= `types.Date`)가 둘 다 받아 `time.Time` 으로 정규화하고
`YYYY-MM-DD` 로 직렬화한다. `time.Time` 을 임베드하므로 `IsZero()`, `Before()`, `Year()` 등을
그대로 쓸 수 있다. 다만 `database/sql` 에 직접 넘길 때는 `d.Time` 을 쓴다.

`types.Time`(루트 별칭 `tiingo.Time`)은 시각까지 보존한다. `statementLastUpdated` 처럼 갱신
시각이 의미 있는 필드, 그리고 암호화폐 시세처럼 인트라데이 값이 오는 필드에 쓰며, 직렬화는
RFC3339 다. 예를 들어 `resampleFreq=1min` 시세의 `date` 는 분 단위 시각이라 `Date` 로는 표현할 수 없다.

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
go build -o /dev/null ./examples/eod                # 예제 빌드(레포 루트의 eod/ 와 이름이 겹쳐 -o 필요)
```

## 문서

- [`docs/api/README.md`](docs/api/README.md) — Tiingo 문서 사이트 23페이지를 변환한 md +
  Tiingo 공식 `llms.txt`/`llms-full.txt` 원본. 재생성은 `./scripts/fetch-docs.sh` 와
  `cd tools/gendocs && npm run gen`.
- 설계·계획: [`docs/superpowers/`](docs/superpowers/)

## 라이선스

[MIT](LICENSE)
