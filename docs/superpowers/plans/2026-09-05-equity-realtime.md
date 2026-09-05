# Equity Realtime 카테고리 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tiingo Equity Realtime(통합 주식 기준가·유동성 스냅샷 + 인트라데이 시세)을 `equity` 서브패키지로 구현해 `tiingo.Client.Equity` 로 노출한다.

**Architecture:** 앞선 여섯 카테고리와 동일한 구조다 — `equity` 패키지가 `internal/httpclient` 를 받아 sub-client 를 만들고, 루트 `tiingo.Client` 가 필드로 들고 있다. 구성은 `iex` 를 그대로 따라 스냅샷(`snapshots.go`)과 시세(`prices.go`)를 파일로 나눈다. 핵심 위험 두 가지는 (1) `volume` 이 정수·소수를 섞어 오므로 `float64` 여야 한다는 것과 (2) 유동성 5필드가 44% 확률로 null 이라 포인터여야 한다는 것이다.

**Tech Stack:** Go 1.25, `encoding/json`, `stretchr/testify`, `net/http/httptest`

**스펙:** `docs/superpowers/specs/2026-09-05-equity-realtime-design.md`
**브랜치:** `feature/equity-realtime` (이미 생성됨, main `41af454` 기준)

---

## 파일 구조

| 파일 | 책임 |
| --- | --- |
| `equity/client.go` (신규) | 패키지 doc, `Client`, `New` |
| `equity/snapshots.go` (신규) | `Snapshot`, `joinTickers`, `Snapshots`, `AllSnapshots` |
| `equity/prices.go` (신규) | `Price`, `ResampleFreq`, `PriceOptions`, `Prices` |
| `equity/testutil_test.go` (신규) | `newStubClient` 헬퍼 |
| `equity/snapshots_test.go` (신규) | 스냅샷 테스트 |
| `equity/prices_test.go` (신규) | 시세 테스트 |
| `equity/testdata/*.json` (신규) | 실호출 fixture 4종 |
| `client.go` (수정) | `Equity *equity.Client` 필드 + 배선 |
| `iex/prices.go` (수정) | 잘못된 주석 정정 |
| `examples/equity/main.go` (신규) | 실행 예제 |
| `integration_test.go` (수정) | 실 API 테스트 4건 |
| `README.md` (수정) | 커버리지 표·사용 예 |
| `../CLAUDE.md` (수정) | 워크스페이스 문서 |

---

### Task 1: fixture 확보

**Files:**
- Create: `equity/testdata/snapshots_aapl_spy.json`
- Create: `equity/testdata/snapshots_null_liquidity.json`
- Create: `equity/testdata/prices_aapl_1hour.json`
- Create: `equity/testdata/prices_aapl_volume.json`

fixture 를 먼저 받는다. 뒤 태스크의 테스트가 이 파일들의 성질에 의존하므로, 성질이 없는 fixture 는
테스트를 조용히 무력화한다.

- [ ] **Step 1: 실 API 로 fixture 4종을 받는다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
H="Authorization: Token $KEY"
B="https://api.tiingo.com/tiingo/equity/intraday"
mkdir -p equity/testdata
curl -s -H "$H" "$B/?tickers=aapl,spy"        | python3 -m json.tool > equity/testdata/snapshots_aapl_spy.json
curl -s -H "$H" "$B/?tickers=aapl,dcfc,pxs,ault" | python3 -m json.tool > equity/testdata/snapshots_null_liquidity.json
curl -s -H "$H" "$B/aapl/prices?startDate=2026-09-04&resampleFreq=1hour" | python3 -m json.tool > equity/testdata/prices_aapl_1hour.json
curl -s -H "$H" "$B/aapl/prices?startDate=2026-09-04&resampleFreq=1hour&columns=open,high,low,close,volume" | python3 -m json.tool > equity/testdata/prices_aapl_volume.json
```

토큰은 `Authorization` 헤더로만 보낸다 — URL 쿼리에 넣지 않는다(레포 규약이자 로그 유출 방지).
**API 키를 출력하거나 보고서에 넣지 않는다.**

- [ ] **Step 2: fixture 가 필요한 성질을 갖췄는지 검증한다**

```bash
python3 - <<'PY'
import json, re
snap = json.load(open('equity/testdata/snapshots_aapl_spy.json'))
nullq = json.load(open('equity/testdata/snapshots_null_liquidity.json'))
p1 = json.load(open('equity/testdata/prices_aapl_1hour.json'))
p2 = json.load(open('equity/testdata/prices_aapl_volume.json'))
LQ = ['lqSpread','lqBidPrice','lqBidSize','lqAskPrice','lqAskSize']
print('스냅샷 건수      :', len(snap))
print('스냅샷 필드수    :', len(snap[0]) if snap else 0)
print('스냅샷 null 없음 :', all(v is not None for r in snap for v in r.values()))
print('누락 티커 있음   :', len(nullq) < 4, '→', sorted(r['ticker'] for r in nullq))
print('Lq 전부 null 행  :', any(all(r[k] is None for k in LQ) for r in nullq))
print('시세 건수        :', len(p1), '| 키:', sorted(p1[0]) if p1 else '-')
print('volume 키 없음   :', 'volume' not in (p1[0] if p1 else {}))
print('volume 소수점    :', bool(re.search(r'"volume":\s*[0-9]+\.[0-9]', open('equity/testdata/prices_aapl_volume.json').read())))
PY
```

Expected:
```
스냅샷 건수      : 2
스냅샷 필드수    : 14
스냅샷 null 없음 : True
누락 티커 있음   : True → ['AAPL', 'AULT', 'PXS']
Lq 전부 null 행  : True
시세 건수        : 6
volume 키 없음   : True
volume 소수점    : True
```

**모든 불리언이 True 여야 한다.** 실시간 데이터라 값은 달라질 수 있지만 성질은 유지돼야 한다.

- `Lq 전부 null 행` 이 False 면 다른 저유동성 티커를 찾는다(`pxs`, `ault`, `dcfc`, `bfri` 등을
  조합해 `?tickers=` 로 시도). 이 fixture 는 포인터 설계를 검증하는 유일한 근거라 반드시 확보한다.
- `volume 소수점` 이 False 면 `startDate` 를 다른 거래일로 바꿔 다시 받는다. 정수로만 온 fixture 는
  `float64` 결정의 근거가 되지 못한다. 세 번 시도해도 안 나오면 **BLOCKED 로 보고**한다 —
  설계 전제가 흔들리는 것이므로 임의로 진행하지 않는다.
- `누락 티커 있음` 이 False 면(4건 다 옴) 그대로 두되 보고에 적는다. 이 fixture 의 주 목적은
  Lq null 이고, 누락은 부수적이다.

fixture 를 손으로 편집해 성질을 만들어내지 않는다. 실제 응답이어야 한다.

- [ ] **Step 3: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add equity/testdata
git commit -m "test(equity): 실호출 fixture 4종 추가"
```

---

### Task 2: 패키지 뼈대와 스냅샷

**Files:**
- Create: `equity/client.go`
- Create: `equity/snapshots.go`
- Create: `equity/testutil_test.go`
- Test: `equity/snapshots_test.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`equity/testutil_test.go`:

```go
package equity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
)

// newStubClient 는 고정 응답을 돌려주는 서버에 붙은 Client 와, 마지막 요청을 꺼내는 함수를 준다.
func newStubClient(t *testing.T, status int, body string) (*Client, func() *http.Request) {
	t.Helper()
	var last *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last = r.Clone(r.Context())
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c := New(httpclient.New(httpclient.Config{APIKey: "k", BaseURL: srv.URL}))
	return c, func() *http.Request { return last }
}
```

`equity/snapshots_test.go`:

```go
package equity

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshots_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_aapl_spy.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "spy"})
	require.NoError(t, err)
	require.Len(t, ss, 2)

	byTicker := map[string]Snapshot{}
	for _, s := range ss {
		byTicker[s.Ticker] = s
	}
	s, ok := byTicker["AAPL"]
	require.True(t, ok, "응답은 대문자 티커로 온다")
	assert.False(t, s.Timestamp.IsZero())
	assert.Greater(t, s.Open, 0.0)
	assert.Greater(t, s.High, 0.0)
	assert.Greater(t, s.Low, 0.0)
	assert.Greater(t, s.TngoLast, 0.0)
	assert.Greater(t, s.PrevClose, 0.0)
	assert.Greater(t, s.LqRefPrice, 0.0)
	assert.Greater(t, s.Volume, 0.0)

	assert.Equal(t, "/tiingo/equity/intraday/", lastReq().URL.Path, "끝 슬래시가 있어야 301 을 피한다")
	assert.Equal(t, "aapl,spy", lastReq().URL.Query().Get("tickers"))
}

// volume 은 문서상 int64 지만 1778528.0 처럼 소수점이 붙어 온다 — int64 였다면 디코딩이 실패한다.
func TestSnapshots_FractionalVolume(t *testing.T) {
	body := `[{"ticker":"AAPL","timestamp":"2026-09-04T19:59:34.935723167-04:00",
	  "open":328.29,"high":328.915,"low":317.845,"tngoLast":320.08,"prevClose":328.29,
	  "volume":1778528.0,"lqRefPrice":320.08,"lqSpread":0.0005,
	  "lqBidPrice":320.015,"lqBidSize":100,"lqAskPrice":320.175,"lqAskSize":100}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ss, err := c.Snapshots(context.Background(), []string{"aapl"})
	require.NoError(t, err)
	require.Len(t, ss, 1)
	assert.InDelta(t, 1778528.0, ss[0].Volume, 0.001)
}

// 같은 엔드포인트가 소수점 없는 정수도 보낸다. 둘 다 받아야 한다.
func TestSnapshots_IntegerVolume(t *testing.T) {
	body := `[{"ticker":"000425","timestamp":"2026-07-29T20:00:00+00:00",
	  "open":8.76,"high":9.09,"low":8.72,"tngoLast":9.0,"prevClose":8.76,
	  "volume":147928660,"lqRefPrice":9.0,"lqSpread":null,
	  "lqBidPrice":null,"lqBidSize":null,"lqAskPrice":null,"lqAskSize":null}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ss, err := c.Snapshots(context.Background(), []string{"000425"})
	require.NoError(t, err)
	require.Len(t, ss, 1)
	assert.InDelta(t, 147928660.0, ss[0].Volume, 0.001)
}

// 유동성 5개 필드는 통합 피드가 값을 안 내면 null 이다 — 값 타입이었다면 0 으로 뭉개진다.
func TestSnapshots_NullLiquidityIsNil(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_null_liquidity.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "dcfc", "pxs", "ault"})
	require.NoError(t, err)

	var nilCount int
	for _, s := range ss {
		if s.LqSpread == nil {
			nilCount++
			assert.Nil(t, s.LqBidPrice, "유동성 필드는 함께 비어 온다")
			assert.Nil(t, s.LqBidSize)
			assert.Nil(t, s.LqAskPrice)
			assert.Nil(t, s.LqAskSize)
			assert.Greater(t, s.TngoLast, 0.0, "유동성이 비어도 가격 필드는 채워진다")
		}
	}
	assert.Positive(t, nilCount, "fixture 에 Lq 가 null 인 행이 있어야 의미가 있다")
}

// 장중처럼 값이 들어온 경우도 확인한다.
func TestSnapshots_NonNullLiquidity(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_aapl_spy.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "spy"})
	require.NoError(t, err)
	require.NotEmpty(t, ss)

	s := ss[0]
	require.NotNil(t, s.LqSpread)
	assert.Greater(t, *s.LqSpread, 0.0)
	require.NotNil(t, s.LqBidPrice)
	assert.Greater(t, *s.LqBidPrice, 0.0)
	require.NotNil(t, s.LqBidSize)
	assert.Greater(t, *s.LqBidSize, 0.0)
	require.NotNil(t, s.LqAskPrice)
	require.NotNil(t, s.LqAskSize)
}

// 없는 티커는 에러가 아니라 응답에서 빠진다. 순서도 요청과 다를 수 있다.
func TestSnapshots_UnknownTickerOmitted(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_null_liquidity.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "dcfc", "pxs", "ault"})
	require.NoError(t, err)
	assert.Less(t, len(ss), 4, "없는 티커는 응답에서 빠져 요청보다 짧다")

	seen := map[string]bool{}
	for _, s := range ss {
		seen[s.Ticker] = true
	}
	assert.True(t, seen["AAPL"], "순서는 보장되지 않으므로 Ticker 로 찾는다")
	assert.Equal(t, "aapl,dcfc,pxs,ault", lastReq().URL.Query().Get("tickers"))
}

func TestSnapshots_EmptyTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Snapshots(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, lastReq(), "왕복하기 전에 막는다")

	_, err = c.Snapshots(context.Background(), []string{" "})
	assert.Error(t, err, "공백뿐인 원소도 에러")
}

// AllSnapshots 는 tickers 를 아예 보내지 않는다 — 그래야 전 종목이 온다.
func TestAllSnapshots_NoTickersParam(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_aapl_spy.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.AllSnapshots(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ss)

	assert.Equal(t, "/tiingo/equity/intraday/", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery, "쿼리가 비어야 전 종목이 온다")
}

func TestSnapshots_APIError(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Snapshots(context.Background(), []string{"aapl"})
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./equity/...`
Expected: 컴파일 실패 — `undefined: Client`, `undefined: Snapshot`

- [ ] **Step 3: 구현을 쓴다**

`equity/client.go`:

```go
// Package equity 는 Tiingo Equity Realtime(통합 주식 기준가·유동성 스냅샷·인트라데이 시세)
// API sub-client 다. tiingo.Client.Equity 로 접근한다.
//
// IEX 가 단일 거래소 피드인 것과 달리 여러 거래소·ATS·OTC 를 합친 통합 피드다.
// 스냅샷의 유동성 5개 필드(Lq*)는 통합 피드가 값을 내지 않으면 null 이라 포인터다 —
// 전 종목 조회 기준 44% 가 그렇다.
// 없는 티커는 스냅샷에서 에러가 아니라 응답에서 빠지므로 결과 길이가 요청보다 짧을 수 있다.
package equity

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Equity Realtime sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
```

`equity/snapshots.go`:

```go
package equity

import (
	"context"
	"fmt"
	"strings"

	"github.com/kenshin579/tiingo-go/types"
)

// snapshotPath 는 스냅샷 경로다. 끝 슬래시가 필요하다 — 없으면 Tiingo 가 301 을 돌려준다.
const snapshotPath = "/tiingo/equity/intraday/"

// Snapshot 은 통합 피드 기준가·유동성 스냅샷이다.
//
// 유동성 5개 필드(Lq*)는 통합 피드가 값을 내지 않으면 null 이라 포인터다 — 전 종목 조회
// 18,654건 중 8,185건(44%)이 그렇다. nil 은 "값 없음"이고 0 과 구분된다.
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

// joinTickers 는 티커 목록을 검증해 콤마로 합친다. 공백뿐인 원소나 빈 목록은 에러다.
func joinTickers(tickers []string) (string, error) {
	cleaned := make([]string, 0, len(tickers))
	for _, t := range tickers {
		t = strings.TrimSpace(t)
		if t == "" {
			return "", fmt.Errorf("tiingo: ticker must not be empty")
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return "", fmt.Errorf("tiingo: at least one ticker is required")
	}
	return strings.Join(cleaned, ","), nil
}

// Snapshots 는 지정한 티커의 스냅샷을 받는다. GET /tiingo/equity/intraday/
//
// 없는 티커는 에러가 아니라 응답에서 빠지고 순서도 요청과 다를 수 있으므로,
// 결과를 인덱스로 대응시키지 말고 Ticker 필드로 찾는다.
// 전 종목을 받으려면 AllSnapshots 를 쓴다 — 빈 목록은 에러다.
func (c *Client) Snapshots(ctx context.Context, tickers []string) ([]Snapshot, error) {
	joined, err := joinTickers(tickers)
	if err != nil {
		return nil, err
	}
	var ss []Snapshot
	if err := c.http.GetJSON(ctx, snapshotPath, map[string]string{"tickers": joined}, &ss); err != nil {
		return nil, err
	}
	return ss, nil
}

// AllSnapshots 는 지원하는 전 종목의 스냅샷을 받는다. GET /tiingo/equity/intraday/
//
// 응답이 크다 — 실측 18,654건 약 4.85MB 로, 받는 데 3초 남짓 걸린다.
// 전 종목 스캔이 목적이 아니라면 Snapshots 를 쓴다.
// 상장폐지·비활성 종목도 섞여 있어 Timestamp 가 한 달 이상 지난 행이 있다.
func (c *Client) AllSnapshots(ctx context.Context) ([]Snapshot, error) {
	var ss []Snapshot
	if err := c.http.GetJSON(ctx, snapshotPath, nil, &ss); err != nil {
		return nil, err
	}
	return ss, nil
}
```

`GetJSON` 의 시그니처(`internal/httpclient/client.go`)가 위 호출과 맞는지 먼저 확인한다.
다르면 설계를 바꾸지 말고 **BLOCKED 로 실제 시그니처를 보고**한다.

- [ ] **Step 4: 테스트 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./equity/... -v`
Expected: PASS — `TestSnapshots_*` 7개 + `TestAllSnapshots_NoTickersParam`

Run: `go build ./... && go vet ./... && gofmt -l equity/` — 모두 무출력

- [ ] **Step 5: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add equity/client.go equity/snapshots.go equity/snapshots_test.go equity/testutil_test.go
git commit -m "feat(equity): 통합 피드 스냅샷 추가"
```

---

### Task 3: 인트라데이 시세

**Files:**
- Create: `equity/prices.go`
- Test: `equity/prices_test.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`equity/prices_test.go`:

```go
package equity

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrices_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl_1hour.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		StartDate:    time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample1Hour,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)

	p := ps[0]
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.GreaterOrEqual(t, p.High, p.Low)
	assert.Zero(t, p.Volume, "columns 를 주지 않으면 응답에 volume 이 없어 0 이다")

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/equity/intraday/aapl/prices", lastReq().URL.Path)
	assert.Equal(t, "2026-09-04", q.Get("startDate"))
	assert.Equal(t, "1hour", q.Get("resampleFreq"))
	assert.Empty(t, q.Get("endDate"), "zero 면 보내지 않는다")
	assert.Empty(t, q.Get("afterHours"), "false 면 보내지 않는다")
	assert.Empty(t, q.Get("forceFill"))
	assert.Empty(t, q.Get("columns"))
}

// columns 에 volume 을 넣으면 거래량이 온다. 448768.0 처럼 소수점이 붙는다.
func TestPrices_ColumnsVolume(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl_volume.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		ResampleFreq: Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)
	assert.Greater(t, ps[0].Volume, 0.0, "columns 에 volume 을 넣으면 채워진다")
	assert.Equal(t, "open,high,low,close,volume", lastReq().URL.Query().Get("columns"))
}

// 소수점이 붙은 volume 을 실제로 디코딩하는지 못 박는다 — int64 였다면 여기서 에러가 난다.
func TestPrices_FractionalVolume(t *testing.T) {
	body := `[{"date":"2026-09-04T14:00:00.000Z","open":323.64,"high":323.72,
	  "low":318.29,"close":318.44,"volume":448768.0}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ps, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.Len(t, ps, 1)
	assert.InDelta(t, 448768.0, ps[0].Volume, 0.001)
}

func TestPrices_AllOptions(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		StartDate:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample5Min,
		AfterHours:   true,
		ForceFill:    true,
		Columns:      []string{"open", "close"},
	})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "2026-09-01", q.Get("startDate"))
	assert.Equal(t, "2026-09-04", q.Get("endDate"))
	assert.Equal(t, "5min", q.Get("resampleFreq"))
	assert.Equal(t, "true", q.Get("afterHours"))
	assert.Equal(t, "true", q.Get("forceFill"))
	assert.Equal(t, "open,close", q.Get("columns"))
}

func TestPrices_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/equity/intraday/aapl/prices", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestPrices_EmptyTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "  ", nil)
	require.Error(t, err)
	assert.Nil(t, lastReq(), "왕복하기 전에 막는다")
}

// 휴장 구간은 에러가 아니라 빈 슬라이스다.
func TestPrices_EmptyResult(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	ps, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, ps)
}

// 없는 티커는 404 다 — 빈 슬라이스가 아니다.
func TestPrices_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Prices(context.Background(), "nosuchxyz", nil)
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./equity/... -run TestPrices`
Expected: 컴파일 실패 — `undefined: PriceOptions`, `c.Prices undefined`

- [ ] **Step 3: 구현을 쓴다**

`equity/prices.go`:

```go
package equity

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// Price 는 인트라데이 구간 하나의 시세다.
// Volume 은 PriceOptions.Columns 에 "volume" 을 넣어야 응답에 포함된다 — 없으면 0 이다.
type Price struct {
	Date  types.Time `json:"date"`  // 구간 시작 시각
	Open  float64    `json:"open"`  // 시가
	High  float64    `json:"high"`  // 고가
	Low   float64    `json:"low"`   // 저가
	Close float64    `json:"close"` // 종가
	// Volume 은 통합 거래량이다. columns 에 명시하지 않으면 응답에 없어 0 이다.
	// 스냅샷과 마찬가지로 448768.0 처럼 소수점이 붙어 int64 로는 디코딩이 실패한다.
	Volume float64 `json:"volume"`
}

// ResampleFreq 는 시세 리샘플 주기다. "<숫자><단위>" 형식이면 무엇이든 되므로 아래 상수는
// 자주 쓰는 값일 뿐 목록이 닫혀 있지 않다. 최소 단위는 1min 이고, 미지정 시 Tiingo 기본값은 5min 이다.
type ResampleFreq = string

const (
	Resample1Min  ResampleFreq = "1min"  // 1분봉
	Resample5Min  ResampleFreq = "5min"  // 5분봉(Tiingo 기본값)
	Resample15Min ResampleFreq = "15min" // 15분봉
	Resample1Hour ResampleFreq = "1hour" // 1시간봉
	Resample4Hour ResampleFreq = "4hour" // 4시간봉
)

// PriceOptions 는 Prices 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작. zero 면 미전송
	EndDate      time.Time    // 조회 종료. zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플 주기. 빈 값이면 Tiingo 기본값 5min
	AfterHours   bool         // true 면 장 전후 데이터를 포함한다
	ForceFill    bool         // true 면 거래가 없던 구간도 직전 값으로 채운다
	// Columns 는 받을 컬럼을 지정한다. 비어 있으면 기본 컬럼(date/open/high/low/close)만 오고
	// Volume 은 0 이다 — 거래량이 필요하면 "volume" 을 반드시 포함해야 한다.
	Columns []string
}

// params 는 옵션을 쿼리 맵으로 바꾼다. 넣을 값이 없으면 nil 이다.
func (o *PriceOptions) params() map[string]string {
	if o == nil {
		return nil
	}
	q := map[string]string{}
	if !o.StartDate.IsZero() {
		q["startDate"] = o.StartDate.Format(types.DateLayout)
	}
	if !o.EndDate.IsZero() {
		q["endDate"] = o.EndDate.Format(types.DateLayout)
	}
	if o.ResampleFreq != "" {
		q["resampleFreq"] = o.ResampleFreq
	}
	if o.AfterHours {
		q["afterHours"] = "true" // 기본값(false)일 때는 보내지 않는다
	}
	if o.ForceFill {
		q["forceFill"] = "true"
	}
	if len(o.Columns) > 0 {
		q["columns"] = strings.Join(o.Columns, ",")
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Prices 는 인트라데이 시세를 받는다. GET /tiingo/equity/intraday/<ticker>/prices
//
// 티커는 하나만 받는다. 휴장 구간이면 빈 슬라이스이고, 없는 티커면 404 라
// APIError(StatusCode 404)가 온다 — ErrNotFound 가 아니다.
func (c *Client) Prices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ps []Price
	path := "/tiingo/equity/intraday/" + url.PathEscape(t) + "/prices"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ps); err != nil {
		return nil, err
	}
	return ps, nil
}
```

`types.DateLayout` 이 존재하는지 확인한다(`iex/prices.go` 가 같은 상수를 쓴다).

- [ ] **Step 4: 테스트 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./equity/... -v`
Expected: PASS — 스냅샷 8개 + 시세 8개

Run: `go build ./... && go vet ./... && gofmt -l equity/` — 모두 무출력
Run: `go test ./equity/... -cover` — 100% 에 가까워야 한다

- [ ] **Step 5: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add equity/prices.go equity/prices_test.go
git commit -m "feat(equity): 인트라데이 시세 추가"
```

---

### Task 4: 루트 클라이언트 배선과 예제

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`
- Create: `examples/equity/main.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`client_test.go` 의 `TestNewClient` 안, `assert.NotNil(t, c.Search)` 다음 줄에 추가한다:

```go
	assert.NotNil(t, c.Equity)
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test . -run TestNewClient`
Expected: 컴파일 실패 — `c.Equity undefined`

- [ ] **Step 3: 배선한다**

`client.go` 를 세 군데 고친다. 기존 정렬을 유지하고 gofmt 로 마무리한다.

import 블록에 추가(경로 알파벳 순서상 `crypto` 와 `eod` 사이):

```go
	"github.com/kenshin579/tiingo-go/equity"
```

`Client` 구조체에서 `Search` 다음 줄:

```go
	Equity       *equity.Client       // 통합 피드 기준가·유동성 스냅샷, 인트라데이 시세
```

`NewClient` 에서 `c.Search = search.New(hc)` 다음 줄:

```go
	c.Equity = equity.New(hc)
```

- [ ] **Step 4: 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test . -v -run TestNewClient`
Expected: PASS

- [ ] **Step 5: 예제를 만든다**

`examples/equity/main.go`. `examples/iex/main.go` 를 먼저 읽고 구조와 어투를 맞춘다.

```go
// Tiingo Equity Realtime 예제.
//
//	TIINGO_API_KEY=... go run ./examples/equity
//
// 통합 피드라 유동성 지표(Lq*)가 함께 온다. 값이 없는 종목은 nil 이다.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/equity"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ss, err := c.Equity.Snapshots(ctx, []string{"AAPL", "SPY", "AULT"})
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range ss {
		fmt.Printf("%-6s %s  기준가 %.2f (전일 %.2f)  거래량 %.0f\n",
			s.Ticker, s.Timestamp, s.TngoLast, s.PrevClose, s.Volume)
		if s.LqSpread != nil && s.LqBidPrice != nil && s.LqAskPrice != nil {
			fmt.Printf("       유동성 %.2f / %.2f  스프레드 %.4f%%\n",
				*s.LqBidPrice, *s.LqAskPrice, *s.LqSpread*100)
		} else {
			fmt.Println("       유동성 지표 없음 — 통합 피드가 값을 내지 않았다")
		}
	}

	ps, err := c.Equity.Prices(ctx, "AAPL", &equity.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: equity.Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(ps) == 0 {
		fmt.Println("시세 없음 — 조회 구간이 전부 휴장일일 수 있다")
		return
	}
	fmt.Printf("최근 1시간봉 %d건 (마지막 5건)\n", len(ps))
	for _, p := range ps[max(0, len(ps)-5):] {
		fmt.Printf("  %s  종가 %.2f  거래량 %.0f\n", p.Date, p.Close, p.Volume)
	}
}
```

- [ ] **Step 6: 빌드하고 실행한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && gofmt -l . | grep -v '^$' || echo "포맷 OK"
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go run ./examples/equity
```

Expected: 스냅샷 3건(AULT 는 "유동성 지표 없음" 이 나올 가능성이 높다)과 1시간봉 요약.
**API 키를 출력하거나 보고서에 넣지 않는다.** 프로그램 출력은 보고서에 넣어도 된다.

- [ ] **Step 7: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add client.go client_test.go examples/equity/main.go
git commit -m "feat(equity): 루트 클라이언트 배선과 예제 추가"
```

---

### Task 5: iex 주석 정정

**Files:**
- Modify: `iex/prices.go`
- Test: `iex/prices_test.go`

`iex/prices.go:82` 가 "휴장이거나 없는 티커면 에러가 아니라 빈 슬라이스를 돌려준다" 라고 적었지만,
실호출 확인 결과 **없는 티커는 404** 다. 빈 슬라이스는 휴장 구간일 때만이다. 주석만 고치면 같은
오류가 다시 생길 수 있으므로 테스트로 못 박는다.

- [ ] **Step 1: 동작을 고정하는 테스트를 쓴다**

`iex/prices_test.go` 끝에 추가한다. 파일에 `errors` 와 `internal/httpclient` import 가
없으면 넣는다(`iex/quotes_test.go` 가 이미 같은 형태로 쓰고 있으니 참고).

```go
// 없는 티커는 404 다 — 빈 슬라이스가 아니다. 주석이 틀렸던 부분이라 못 박는다.
func TestPrices_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Prices(context.Background(), "nosuchxyz", nil)
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
```

- [ ] **Step 2: 실행한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./iex/... -run TestPrices_NotFound -v`
Expected: PASS (구현은 이미 그렇게 동작한다. 실패하면 `httpclient` 가 404 를 다르게 다루는 것이니 보고한다)

- [ ] **Step 3: 주석을 고친다**

`iex/prices.go` 의 `Prices` doc 주석에서 아래 줄을

```go
// 휴장이거나 없는 티커면 에러가 아니라 빈 슬라이스를 돌려준다.
```

다음으로 바꾼다:

```go
// 휴장 구간이면 빈 슬라이스이고, 없는 티커면 404 라 APIError(StatusCode 404)가 온다.
```

- [ ] **Step 4: 확인하고 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go test ./iex/... && go vet ./... && gofmt -l iex/
git add iex/prices.go iex/prices_test.go
git commit -m "fix(iex): 없는 티커가 404 라는 사실을 주석·테스트로 정정"
```

---

### Task 6: 통합 테스트

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: 실 API 테스트를 추가한다**

파일 끝에 붙인다. import 블록에 `"github.com/kenshin579/tiingo-go/equity"` 를 순서에 맞게 추가한다.

```go
// 장 마감·주말에도 깨지지 않도록 Lq* 의 nil 여부는 단정하지 않는다.
func TestIntegration_EquitySnapshots(t *testing.T) {
	c := newClient(t)
	ss, err := c.Equity.Snapshots(context.Background(), []string{"AAPL", "SPY"})
	require.NoError(t, err)
	require.Len(t, ss, 2)
	for _, s := range ss {
		assert.NotEmpty(t, s.Ticker)
		assert.False(t, s.Timestamp.IsZero())
		assert.Greater(t, s.TngoLast, 0.0)
		assert.Greater(t, s.PrevClose, 0.0)
		assert.Greater(t, s.LqRefPrice, 0.0)
		if s.LqSpread != nil {
			assert.Greater(t, *s.LqSpread, 0.0, "유동성 지표가 있으면 양수여야 한다")
		}
	}
}

// 없는 티커는 응답에서 빠진다.
func TestIntegration_EquityUnknownTickerOmitted(t *testing.T) {
	c := newClient(t)
	ss, err := c.Equity.Snapshots(context.Background(), []string{"AAPL", "NOSUCHTICKERXYZ"})
	require.NoError(t, err)
	assert.Len(t, ss, 1, "없는 티커는 빠지고 AAPL 만 온다")
}

// 전 종목 조회는 약 5MB 라 한 건만 돌린다.
func TestIntegration_EquityAllSnapshots(t *testing.T) {
	c := newClient(t)
	ss, err := c.Equity.AllSnapshots(context.Background())
	require.NoError(t, err)
	assert.Greater(t, len(ss), 1000, "전 종목이라 수천 건이 온다")

	var withLq int
	for _, s := range ss {
		if s.LqSpread != nil {
			withLq++
		}
	}
	assert.Positive(t, withLq, "일부는 유동성 지표가 채워져 있다")
	assert.Less(t, withLq, len(ss), "일부는 비어 있다 — 포인터로 둔 이유다")
}

func TestIntegration_EquityPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.Equity.Prices(context.Background(), "AAPL", &equity.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: equity.Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	require.NoError(t, err)
	for _, p := range ps {
		assert.False(t, p.Date.IsZero())
		assert.Greater(t, p.Close, 0.0)
	}
}
```

- [ ] **Step 2: 실 API 로 실행한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration -run TestIntegration_Equity -v .
```

Expected: 4건 PASS

**Tiingo 는 시간당 요청 한도가 있다.** 429 가 나오면 한도 초과이니 정시까지 기다렸다 다시 돌린다.
테스트를 약하게 고쳐 통과시키지 않는다.

- [ ] **Step 3: 전체 통합 테스트를 돌린다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration . 2>&1 | tail -5
```

Expected: ok — 29건(기존 25 + Equity 4)

- [ ] **Step 4: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add integration_test.go
git commit -m "test(equity): 실 API 통합 테스트 4건 추가"
```

---

### Task 7: 문서 갱신

**Files:**
- Modify: `README.md`
- Modify: `/Users/user/src/workspace_moneyflow/CLAUDE.md`

- [ ] **Step 1: README 사용 예에 추가한다**

`## 사용` 코드 블록에서 IEX 줄 다음에 넣는다:

```go
// 통합 피드 스냅샷(여러 거래소·ATS·OTC). 유동성 지표는 없을 수 있어 포인터다
es, _ := c.Equity.Snapshots(ctx, []string{"AAPL", "SPY"})
```

"실행 가능한 예시" 줄 끝에 `, [`examples/equity`](examples/equity)` 를 더한다.

- [ ] **Step 2: 커버리지 표에 3행을 더한다**

Search 행 다음에 넣는다:

```markdown
| Equity Realtime | `Equity.Snapshots` | `GET /tiingo/equity/intraday/` |
| Equity Realtime | `Equity.AllSnapshots` | `GET /tiingo/equity/intraday/` |
| Equity Realtime | `Equity.Prices` | `GET /tiingo/equity/intraday/<ticker>/prices` |
```

표 아래 미구현 목록 문장에서 `Equity Realtime, ` 을 지운다. 실제 문장을 읽고 정확히 고친다 —
기억으로 전체 문장을 다시 쓰지 않는다. `News` 옆에는 권한 상태를 적는다:

```markdown
나머지 REST 그룹(News — 별도 플랜 필요, BOATS — 별도 등록 필요, Fund Fees, Dividends/Splits)과
WebSocket 은 순차 추가 예정.
```

- [ ] **Step 3: `## null 필드` 절에 한 문단을 더한다**

```markdown
Equity Realtime 스냅샷의 유동성 5개 필드(`LqSpread`, `LqBidPrice`, `LqBidSize`, `LqAskPrice`,
`LqAskSize`)는 통합 피드가 값을 내지 않으면 `null` 이라 포인터다 — 전 종목 조회 기준 44% 가
그렇다. 나머지 가격·거래량 필드는 저유동성 종목에서도 채워진다.
```

- [ ] **Step 4: `## 테스트` 절의 예제 빌드 줄에 equity 를 더한다**

`go build -o /dev/null ./examples/search` 다음 줄에:

```bash
go build -o /dev/null ./examples/equity
```

- [ ] **Step 5: 워크스페이스 CLAUDE.md 를 갱신한다**

`/Users/user/src/workspace_moneyflow/CLAUDE.md` 의 `### tiingo-go` 절을 고친다.
이 파일은 git 저장소가 아니므로 커밋하지 않는다(시도하면 실패한다 — 정상이다).

`**Module**:` 문단의 버전 문장을 아래로 바꾼다:

```
**v0.7.0 (2026-09-05)**: 일곱 카테고리 — EOD(`Meta`/`LatestPrice`/`HistoricalPrices`), Fundamentals(`Definitions`/`Meta`/`Statements`/`Daily`, 별도 구독), Crypto(`Meta`/`Prices`/`TopOfBook`), Forex(`TopOfBook`/`Prices`), IEX(`Quotes`/`Prices`), Search(`Search`/`SearchByISIN`), Equity Realtime(`Snapshots`/`AllSnapshots`/`Prices`).
```

`**다음**:` 문단을 아래로 바꾼다:

```
**다음**: 남은 REST 카테고리(News — 403 권한 없음, BOATS — 별도 등록 필요, Fund Fees, Dividends/Splits — 5개 중 4개 403)와 WebSocket 5종. 실측상 무료 키로 바로 되는 건 Fund Fees 뿐이다.
```

예제 빌드 주석 줄의 `eod/·search/` 를 `eod/·search/·equity/` 로 고친다.

- [ ] **Step 6: 최종 검증**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && go test ./... && gofmt -l . | grep -v '^$' || echo "포맷 OK"
file -I README.md /Users/user/src/workspace_moneyflow/CLAUDE.md
grep -n "Equity" README.md
```

두 파일 모두 `charset=utf-8` 이어야 한다 — 이 레포는 한글 인코딩이 깨진 적이 있으므로 확인한다.

- [ ] **Step 7: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add README.md
git commit -m "docs(equity): README 커버리지와 사용법 추가"
```

---

## 완료 후

`superpowers:finishing-a-development-branch` 로 PR 을 만든다. 머지 후 v0.7.0 릴리스:

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go && ./scripts/release.sh v0.7.0
```

릴리스 뒤 신규 모듈에서 소비를 확인한다:

```bash
cd $(mktemp -d) && go mod init probe && GOPROXY=https://proxy.golang.org go get github.com/kenshin579/tiingo-go@v0.7.0
```

## 범위 밖

- News(403), BOATS(등록 필요), Fund Fees, Dividends/Splits(403)
- WebSocket 5종 — 이 카테고리에도 WebSocket 판이 있으나 스트리밍은 별도 설계가 필요하다
- CSV(`format=csv`), 재시도, rate limit 백오프, moneyflow 통합
