# Tiingo Go SDK — Equity Realtime 카테고리 설계

- 작성일: 2026-09-05
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/equity-realtime`)
- 토픽: Tiingo Equity Realtime(통합 주식 기준가·유동성 스냅샷 + 인트라데이 시세) — 일곱 번째 카테고리
- 선행: v0.6.0 (EOD + Fundamentals + Crypto + Forex + IEX + Search)
- API 참조: `docs/api/rest/equity-realtime-stock-data.md`

## 배경 / 목적

IEX 는 단일 거래소 피드다. Equity Realtime 은 **여러 거래소·ATS·OTC 를 합친 통합 피드**로,
같은 종목에 대해 더 넓은 시장을 반영한 기준가와 유동성 지표를 준다. 문서가 스스로를 "an alternative
to the IEX Endpoints" 라고 소개한다. 구조가 IEX 와 거의 같아 `iex` 패키지가 그대로 본이 된다.

## 사전 조사 결과 (실호출 확인, 2026-09-05 토요일 = 장 마감)

### 스냅샷 `GET /tiingo/equity/intraday/`

| 호출 | 결과 |
|---|---|
| `?tickers=aapl,spy` | 200, 2건, 14필드 |
| `/tiingo/equity/intraday` (끝 슬래시 없음) | **301** |
| `/tiingo/equity/intraday/aapl` (경로형) | 200, 배열 1건 — 쿼리형과 같은 결과 |
| `?tickers=aapl,NOSUCHXYZ` | 200, **AAPL 만**. 없는 티커는 빠진다 |
| `?tickers=aapl,dcfc,pxs,ault` | 200, **`[PXS, AAPL, AULT]`** — DCFC 누락, 순서도 요청과 다름 |
| tickers 없이 | 200, **18,654건 / 4.85MB / 3.3초** |

**응답 14필드**: `ticker timestamp open high low tngoLast prevClose volume lqRefPrice lqSpread
lqBidPrice lqBidSize lqAskPrice lqAskSize`

### 시세 `GET /tiingo/equity/intraday/<ticker>/prices`

| 호출 | 결과 |
|---|---|
| `?startDate=2026-09-04&resampleFreq=1hour` | 200, 6건. 키는 `date open high low close` |
| `&columns=open,high,low,close,volume` | 200, `volume` 포함 |
| `&afterHours=true` / `&forceFill=true` | 각각 200 |
| 없는 티커 | **404** `{"detail":"Not found."}` |
| 주말 구간(`startDate=endDate=2026-09-05`) | 200 + `[]` |

### 결정을 좌우한 실측 세 가지

**1. `volume` 이 정수와 소수를 섞어 온다.** 문서는 `int64` 라고 적지만:

- AAPL 스냅샷 → `1778528.0` (소수점 있음)
- 전체 조회의 `000425` → `147928660` (소수점 없음)
- 시세 → `448768.0`

**같은 엔드포인트가 행마다 다르다.** `int64` 로 두면 `.0` 이 붙은 행에서 디코딩이 실패한다.
IEX 에서 이미 겪은 함정이고(`iex/prices.go` 의 `Volume float64`), 여기서는 스냅샷까지 해당된다.

**2. 유동성 5개 필드가 흔히 null 이다.** AULT 는 `lqSpread lqBidPrice lqBidSize lqAskPrice lqAskSize`
가 전부 null 이었고, 전체 조회 18,654건 중 **8,185건(44%)** 이 `lqSpread` 가 null 이다. 예외가 아니라
절반 가까이다. 값 타입으로 두면 JSON null 이 **에러 없이 조용히 0** 이 된다(IEX 설계에서 실측한 사실).

**3. 나머지 9필드는 얇은 종목에서도 채워진다.** PXS·AULT 같은 저유동성 종목에서도 `ticker timestamp
open high low tngoLast prevClose volume lqRefPrice` 는 값이 있었다. 값 타입으로 둔다.

## 결정 사항 (브레인스토밍)

1. **메서드 이름은 내용을 따른다** — `Snapshots` / `AllSnapshots` / `Prices`.
   (기각: IEX 와 대칭을 맞춰 `Quotes` — 이 스냅샷은 호가가 아니라 통합 피드 기준가이고, 문서도
   "Reference Price & Liquidity Snapshot" 이라 부른다. `lqBidPrice` 는 문서가 명시하듯 호가 자체가
   아니라 유동성 위험 지표의 구성요소다. `Quotes` 라 부르면 `iex.Quote.BidPrice` 와 같은 것으로
   읽힌다.)
2. **전체 조회는 별도 메서드** `AllSnapshots(ctx)`.
   (기각: 티커를 선택 인자로 만들어 `Snapshots(ctx, nil)` 이 전체를 뜻하게 하기 — 4.85MB 다운로드가
   빈 슬라이스 하나로 조용히 일어난다. 지금 `iex.Quotes` 는 빈 목록을 에러로 막으므로 같은 모양의
   호출이 여기서만 5MB 가 되면 일관성이 깨진다. 이름 자체가 경고가 되는 편이 낫다.)
   (기각: 노출하지 않기 — 문서에 있는 기능이고, 한 번에 전 종목 스캔이 필요한 용도가 실재한다.)
3. **`Volume` 은 `float64`**, 유동성 5필드는 포인터, 수량 2개는 `*float64`.
   `LqBidSize`/`LqAskSize` 는 관측상 정수만 왔지만, **같은 응답의 `volume` 이 형식을 섞는 것을
   확인했으므로** 수량 계열도 `*float64` 로 둔다. IEX 에서 같은 판단을 했고 옳았다.
4. **`Price`/`PriceOptions`/`ResampleFreq` 는 `iex` 것과 사실상 같지만 복제한다.**
   `crypto`·`forex`·`iex` 가 이미 각자 정의한다. 카테고리 간 import 를 만들면 한쪽 API 변경이
   다른 쪽을 끌고 간다.
5. 필드 주석 규약 유지(오른쪽 줄 끝, 길면 위), fixture 는 실호출.

## 아키텍처

```
tiingo-go/
├── client.go              # (수정) Equity *equity.Client 필드 추가
├── equity/
│   ├── client.go          # 패키지 doc, Client, New(hc)
│   ├── snapshots.go       # Snapshot, Snapshots, AllSnapshots
│   ├── prices.go          # Price, ResampleFreq, PriceOptions, Prices
│   ├── *_test.go
│   └── testdata/*.json    # 실호출 fixture 4종
├── examples/equity/main.go
├── iex/prices.go          # (수정) 잘못된 주석 정정
└── integration_test.go    # (수정) Equity 4건 추가
```

스냅샷과 시세는 응답 모양도 옵션도 완전히 다르므로 파일을 나눈다(`iex` 와 같은 구성).

## 데이터 모델

```go
// Snapshot 은 통합 피드 기준가·유동성 스냅샷이다.
//
// 유동성 5개 필드(Lq*)는 통합 피드가 값을 내지 않으면 null 이라 포인터다 — 전체 조회 18,654건 중
// 8,185건(44%)이 그렇다. nil 은 "값 없음"이고 0 과 구분된다.
type Snapshot struct {
	Ticker     string     `json:"ticker"`     // 티커. 요청과 무관하게 대문자로 온다
	Timestamp  types.Time `json:"timestamp"`  // 마지막 갱신 시각
	Open       float64    `json:"open"`       // 당일 시가
	High       float64    `json:"high"`       // 당일 고가
	Low        float64    `json:"low"`        // 당일 저가
	TngoLast   float64    `json:"tngoLast"`   // Tiingo 기준 최종가. 스프레드가 넓지 않으면 중간가를 쓴다
	PrevClose  float64    `json:"prevClose"`  // 전일 종가
	LqRefPrice float64    `json:"lqRefPrice"` // 유동성 기준가. TngoLast 와 같은 값이다
	// Volume 은 통합 장중 거래량이다. 문서는 int64 라 하지만 행에 따라 1778528.0 처럼 소수점이
	// 붙어 오므로(같은 엔드포인트에서 147928660 처럼 정수로도 온다) int64 로는 디코딩이 실패한다.
	Volume float64 `json:"volume"`
	// 아래 5개는 유동성 위험 지표다. 호가 그 자체가 아니라 지표의 구성요소이며,
	// WebSocket 의 thresholdLevel 4 메시지와 같은 값이다.
	LqSpread   *float64 `json:"lqSpread"`   // 상대 스프레드. 0.04 면 4%
	LqBidPrice *float64 `json:"lqBidPrice"` // 매수 측 가격
	LqBidSize  *float64 `json:"lqBidSize"`  // 매수 측 수량(주). Volume 과 같은 계열이라 float64
	LqAskPrice *float64 `json:"lqAskPrice"` // 매도 측 가격
	LqAskSize  *float64 `json:"lqAskSize"`  // 매도 측 수량(주)
}

// Price 는 인트라데이 구간 하나의 시세다.
type Price struct {
	Date  types.Time `json:"date"`  // 구간 시작 시각
	Open  float64    `json:"open"`  // 시가
	High  float64    `json:"high"`  // 고가
	Low   float64    `json:"low"`   // 저가
	Close float64    `json:"close"` // 종가
	// Volume 은 통합 거래량이다. columns 에 "volume" 을 명시하지 않으면 응답에 없어 0 이다.
	// 스냅샷과 마찬가지로 448768.0 처럼 소수점이 붙어 int64 로는 디코딩이 실패한다.
	Volume float64 `json:"volume"`
}

// ResampleFreq 는 시세 리샘플 주기다. "<숫자><단위>" 형식이면 무엇이든 되므로 아래 상수는
// 자주 쓰는 값일 뿐 목록이 닫혀 있지 않다. 최소 단위는 1min, 미지정 시 Tiingo 기본값은 5min.
type ResampleFreq = string

const (
	Resample1Min  ResampleFreq = "1min"
	Resample5Min  ResampleFreq = "5min"
	Resample15Min ResampleFreq = "15min"
	Resample1Hour ResampleFreq = "1hour"
	Resample4Hour ResampleFreq = "4hour"
)

// PriceOptions 는 Prices 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작. zero 면 미전송
	EndDate      time.Time    // 조회 종료. zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플 주기. 빈 값이면 Tiingo 기본값 5min
	AfterHours   bool         // true 면 장 전후 데이터를 포함한다
	ForceFill    bool         // true 면 거래가 없던 구간도 직전 값으로 채운다
	// Columns 는 받을 컬럼을 지정한다. 비어 있으면 date/open/high/low/close 만 오고
	// Volume 은 0 이다 — 거래량이 필요하면 "volume" 을 반드시 포함해야 한다.
	Columns []string
}
```

## 메서드

```go
// Snapshots 는 지정한 티커의 스냅샷을 받는다. GET /tiingo/equity/intraday/
// 없는 티커는 에러가 아니라 응답에서 빠지고 순서도 요청과 다를 수 있으므로,
// 결과를 인덱스로 대응시키지 말고 Ticker 필드로 찾는다.
func (c *Client) Snapshots(ctx context.Context, tickers []string) ([]Snapshot, error)

// AllSnapshots 는 지원하는 전 종목의 스냅샷을 받는다. GET /tiingo/equity/intraday/
//
// 응답이 크다 — 실측 18,654건 약 4.85MB, 3초 남짓 걸린다. 전 종목 스캔이 아니라면 Snapshots 를 쓴다.
// 상장폐지·비활성 종목도 섞여 있어 Timestamp 가 한 달 이상 지난 행이 있다.
func (c *Client) AllSnapshots(ctx context.Context) ([]Snapshot, error)

// Prices 는 인트라데이 시세를 받는다. GET /tiingo/equity/intraday/<ticker>/prices
// 티커는 하나만 받는다. 없는 티커면 404(ErrNotFound), 휴장 구간이면 빈 슬라이스다.
func (c *Client) Prices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error)
```

- `Snapshots` 는 빈 목록·공백 원소를 에러로 막는다(`iex.joinTickers` 와 동일). 전체 조회는
  `AllSnapshots` 로만 간다.
- 경로 끝 슬래시가 필요하다 — `/tiingo/equity/intraday` 는 301 이다.

## 곁다리 수정

`iex/prices.go:82` 의 주석이 틀렸다. "휴장이거나 없는 티커면 에러가 아니라 빈 슬라이스를 돌려준다"
라고 적혀 있으나, 실호출 확인 결과 **없는 티커는 404** 이고 빈 슬라이스는 휴장 구간일 때만이다.
Equity 도 같은 동작이라 두 곳의 주석을 같은 문장으로 맞춘다.

## 테스트

- **fixture 4종**(실호출): `snapshots_aapl_spy.json`(2건, null 없음),
  `snapshots_null_liquidity.json`(**AULT 포함 — Lq* 5개가 null**),
  `prices_aapl_1hour.json`(columns 없이 — volume 0 확인),
  `prices_aapl_volume.json`(**`448768.0` 소수점 volume**).
- **단위 테스트**: 스냅샷 파싱, **Lq* null → nil**(0 이 아님), **소수점 volume 디코딩**,
  티커 조인·빈 목록 에러, 없는 티커가 빠져도 에러가 아님, `AllSnapshots` 가 tickers 를 안 보냄,
  쿼리 생성(zero 값 생략, true 일 때만, Columns 콤마 결합), 404 → `ErrNotFound`, 빈 배열이 에러가 아님.
- **integration**: `Snapshots`(AAPL·SPY), `AllSnapshots`(건수만 확인 — 매번 5MB 를 받으므로 한 건만),
  `Prices`, 없는 티커 404. 장 마감·주말에도 깨지지 않게 nil 여부는 단정하지 않는다.

## 릴리스

머지 후 **v0.7.0**. README 커버리지 표에 Equity 3행 추가.

## 범위 밖 (후속)

- News(403 권한 없음 확인), BOATS(별도 등록 필요), Fund Fees, Dividends/Splits(403 보류).
- WebSocket 5종 — 이 카테고리도 WebSocket 판이 있으나 스트리밍은 별도 설계가 필요하다.
- CSV(`format=csv`), 재시도, moneyflow 통합.

## 주의

- 조사 시점이 **토요일 장 마감**이라 스냅샷은 직전 거래일(금요일 20:00 ET) 값이다. 장중에 없던
  필드가 null 로 바뀌거나 그 반대일 가능성은 배제하지 못한다. 값 타입으로 둔 9필드는 저유동성
  종목에서도 채워지는 것을 확인했으나, 장중 재확인이 남는다.
- `format=csv` 는 이번 범위 밖이다. 이 라이브러리는 아직 어느 카테고리도 CSV 를 지원하지 않는다.
- 전체 조회에는 티커가 `000425` 같은 숫자 문자열인 행도 있다.
