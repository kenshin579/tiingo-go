# Corporate Actions 카테고리 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tiingo Corporate Actions 중 무료 키로 접근 가능한 유일한 엔드포인트(배당수익률 시계열)를 `corporateactions` 서브패키지로 구현해 `tiingo.Client.CorporateActions` 로 노출한다.

**Architecture:** 앞선 일곱 카테고리와 동일한 구조다 — 패키지가 `internal/httpclient` 를 받아 sub-client 를 만들고 루트 `tiingo.Client` 가 필드로 들고 있다. 엔드포인트가 하나뿐이라 `search` 와 같은 단일 파일 구성이다. **이 카테고리는 5개 중 1개만 구현하는 의도적 반쪽**이며, 나머지 4개가 403 이라 빠졌다는 사실을 패키지 doc·README 에 명시하는 것이 구현만큼 중요하다.

**Tech Stack:** Go 1.25, `encoding/json`, `stretchr/testify`, `net/http/httptest`

**스펙:** `docs/superpowers/specs/2026-09-05-corporate-actions-design.md`
**브랜치:** `feature/corporate-actions` (이미 생성됨, main `62277be` 기준)

---

## Tiingo 시간당 요청 한도

이 계획은 실호출이 필요한 단계가 둘이다(Task 1 fixture, Task 5 통합 테스트). Tiingo 는 시간당
요청 한도가 있고 **정시에 리셋된다**. 429 응답은 이렇게 온다:

```json
{"detail":"Error: You have run over your hourly request allocation. Please upgrade at https://api.tiingo.com/pricing to have your limits increased."}
```

429 를 만나면 **정시까지 기다렸다 다시 시도**한다. 테스트를 약하게 고치거나 fixture 를 손으로
만들어 우회하지 않는다. 기다릴 수 없으면 BLOCKED 로 보고한다.

---

## 파일 구조

| 파일 | 책임 |
| --- | --- |
| `corporateactions/client.go` (신규) | 패키지 doc(**막힌 4개 설명 포함**), `Client`, `New` |
| `corporateactions/yield.go` (신규) | `Yield`, `YieldOptions`, `DistributionYield` |
| `corporateactions/testutil_test.go` (신규) | `newStubClient` 헬퍼 |
| `corporateactions/yield_test.go` (신규) | 단위 테스트 |
| `corporateactions/testdata/*.json` (신규) | 실호출 fixture 2종 |
| `client.go` (수정) | `CorporateActions *corporateactions.Client` 필드 + 배선 |
| `examples/corporateactions/main.go` (신규) | 실행 예제 |
| `integration_test.go` (수정) | 실 API 테스트 3건 |
| `README.md` (수정) | 커버리지 표 + 권한 상태 갱신 |
| `../CLAUDE.md` (수정) | 워크스페이스 문서 |

---

### Task 1: fixture 확보

**Files:**
- Create: `corporateactions/testdata/yield_aapl_recent.json`
- Create: `corporateactions/testdata/yield_aapl_early.json`

- [ ] **Step 1: 실 API 로 fixture 2종을 받는다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
H="Authorization: Token $KEY"
B="https://api.tiingo.com/tiingo/corporate-actions/aapl/distribution-yield"
mkdir -p corporateactions/testdata
curl -s -H "$H" "$B?startDate=2026-08-01"                      | python3 -m json.tool > corporateactions/testdata/yield_aapl_recent.json
curl -s -H "$H" "$B?startDate=1980-12-01&endDate=1981-01-15"   | python3 -m json.tool > corporateactions/testdata/yield_aapl_early.json
```

토큰은 `Authorization` 헤더로만 보낸다 — URL 쿼리에 넣지 않는다. **API 키를 출력하거나 보고서에
넣지 않는다.**

- [ ] **Step 2: fixture 가 필요한 성질을 갖췄는지 검증한다**

```bash
python3 - <<'PY'
import json
recent = json.load(open('corporateactions/testdata/yield_aapl_recent.json'))
early  = json.load(open('corporateactions/testdata/yield_aapl_early.json'))
for name, d in (('recent', recent), ('early', early)):
    if not isinstance(d, list):
        print(f'{name}: 배열이 아니다 →', d); raise SystemExit(1)
print('recent 건수     :', len(recent))
print('recent 키       :', sorted(recent[0]) if recent else '-')
print('recent 값 양수  :', all(r['trailingDiv1Y'] > 0 for r in recent))
print('early 건수      :', len(early))
print('early 0.0 행    :', sum(1 for r in early if r['trailingDiv1Y'] == 0.0), '/', len(early))
print('숫자 타입       :', all(isinstance(r['trailingDiv1Y'], (int, float)) for r in recent + early))
PY
```

Expected:
```
recent 건수     : 25
recent 키       : ['date', 'trailingDiv1Y']
recent 값 양수  : True
early 건수      : 20
early 0.0 행    : 20 / 20
숫자 타입       : True
```

**성질 판정 기준**(건수는 달라질 수 있다):

- `recent` 는 비어 있지 않고 `trailingDiv1Y` 가 모두 양수여야 한다.
- `early` 에 **`trailingDiv1Y == 0.0` 인 행이 하나 이상** 있어야 한다. 이 fixture 의 존재 이유가
  "0.0 은 결손이 아니라 정상 값(배당 시작 전)" 을 고정하는 것이다. 하나도 없으면 다른 초기
  구간으로 다시 받는다(AAPL 은 1980-12-12 부터 데이터가 있고 초기 구간이 0.0 이다).
- `숫자 타입` 이 False 면 응답 스키마가 바뀐 것이므로 **BLOCKED 로 보고**한다. 문서는 이 필드를
  `string` 이라 적지만 실제로는 숫자로 오며, 그것이 `float64` 설계의 근거다.
- 배열이 아니면 429 일 가능성이 높다 — 위 "시간당 요청 한도" 절을 따른다.

fixture 를 손으로 편집해 성질을 만들지 않는다.

- [ ] **Step 3: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add corporateactions/testdata
git commit -m "test(corporateactions): 실호출 fixture 2종 추가"
```

---

### Task 2: 패키지와 배당수익률 조회

**Files:**
- Create: `corporateactions/client.go`
- Create: `corporateactions/yield.go`
- Create: `corporateactions/testutil_test.go`
- Test: `corporateactions/yield_test.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`corporateactions/testutil_test.go`:

```go
package corporateactions

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

`corporateactions/yield_test.go`:

```go
package corporateactions

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

func TestDistributionYield_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/yield_aapl_recent.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ys, err := c.DistributionYield(context.Background(), "aapl", &YieldOptions{
		StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.NotEmpty(t, ys)

	y := ys[0]
	assert.False(t, y.Date.IsZero())
	assert.Greater(t, y.TrailingDiv1Y, 0.0)

	assert.Equal(t, "/tiingo/corporate-actions/aapl/distribution-yield", lastReq().URL.Path)
	assert.Equal(t, "2026-08-01", lastReq().URL.Query().Get("startDate"))
	assert.Empty(t, lastReq().URL.Query().Get("endDate"), "zero 면 보내지 않는다")
}

// 응답은 T00:00:00.000Z 로 오지만 의미는 날짜다. types.Date 가 시각을 잘라내는지 확인한다.
func TestDistributionYield_DateIsDateOnly(t *testing.T) {
	body := `[{"date":"2026-09-04T00:00:00.000Z","trailingDiv1Y":0.0033128106}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ys, err := c.DistributionYield(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.Len(t, ys, 1)
	assert.Equal(t, "2026-09-04", ys[0].Date.String())
	assert.InDelta(t, 0.0033128106, ys[0].TrailingDiv1Y, 1e-12)
}

// 배당 시작 전 구간은 0.0 이 정상 값이다 — 결손이 아니므로 포인터로 두지 않는다.
func TestDistributionYield_ZeroIsValid(t *testing.T) {
	raw, err := os.ReadFile("testdata/yield_aapl_early.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ys, err := c.DistributionYield(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.NotEmpty(t, ys)

	var zeros int
	for _, y := range ys {
		if y.TrailingDiv1Y == 0.0 {
			zeros++
			assert.False(t, y.Date.IsZero(), "값이 0 이어도 날짜는 있다")
		}
	}
	assert.Positive(t, zeros, "fixture 에 0.0 인 행이 있어야 의미가 있다")
}

func TestDistributionYield_BothDates(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "aapl", &YieldOptions{
		StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "2026-08-01", q.Get("startDate"))
	assert.Equal(t, "2026-08-15", q.Get("endDate"))
}

func TestDistributionYield_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/corporate-actions/aapl/distribution-yield", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

// zero 값만 든 옵션은 nil 옵션과 똑같이 쿼리가 비어야 한다.
func TestDistributionYield_ZeroOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "aapl", &YieldOptions{})
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "보낼 값이 없으면 쿼리를 만들지 않는다")
}

// 앞뒤 공백은 잘라내고 경로에 넣는다.
func TestDistributionYield_TrimsTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "  aapl  ", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/corporate-actions/aapl/distribution-yield", lastReq().URL.Path)
}

func TestDistributionYield_EmptyTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "  ", nil)
	require.Error(t, err)
	assert.Nil(t, lastReq(), "왕복하기 전에 막는다")
}

// 없는 티커는 200 + [] 다 — equity·iex 의 시세가 404 를 주는 것과 다르다.
func TestDistributionYield_EmptyResult(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	ys, err := c.DistributionYield(context.Background(), "nosuchxyz", nil)
	require.NoError(t, err)
	assert.Empty(t, ys)
}

// 같은 API 그룹의 다른 엔드포인트가 403 이므로 이 경로도 언젠가 막힐 수 있다.
func TestDistributionYield_Forbidden(t *testing.T) {
	c, _ := newStubClient(t, http.StatusForbidden, ``)
	_, err := c.DistributionYield(context.Background(), "aapl", nil)
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./corporateactions/...`
Expected: 컴파일 실패 — `undefined: Client`, `undefined: Yield`, `undefined: YieldOptions`

- [ ] **Step 3: 구현을 쓴다**

`corporateactions/client.go`:

```go
// Package corporateactions 는 Tiingo Corporate Actions API sub-client 다.
// tiingo.Client.CorporateActions 로 접근한다.
//
// 지금은 배당수익률(DistributionYield) 하나만 구현돼 있다. 같은 API 그룹의 배당 내역
// (distributions, 티커별·배치)과 분할(splits)은 이 라이브러리를 만든 계정 권한으로 403 이라
// 응답 형태를 확인할 수 없어 넣지 않았다 — 실호출로 검증하지 않은 모델은 두지 않는다는 원칙이다.
// 권한이 열리면 이 패키지에 추가한다.
package corporateactions

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Corporate Actions sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
```

`corporateactions/yield.go`:

```go
package corporateactions

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// Yield 는 하루치 배당수익률이다.
//
// Tiingo 가 이 엔드포인트에 지표를 계속 추가하겠다고 문서에 밝혔으므로 필드가 늘 수 있다.
// 모르는 필드는 무시되니 디코딩은 깨지지 않고, 새 지표가 필요해지면 그때 필드를 더하면 된다.
type Yield struct {
	Date types.Date `json:"date"` // 기준일. 응답은 늘 자정이라 날짜만 의미가 있다
	// TrailingDiv1Y 는 최근 1년 분배금 기준 수익률이다. 0.0033 이면 0.33% 다.
	// 문서 표는 string 이라 적지만 실제로는 JSON 숫자로 온다.
	// 배당을 시작하기 전 구간은 0 이며, 이는 결손이 아니라 정상 값이다.
	TrailingDiv1Y float64 `json:"trailingDiv1Y"`
}

// YieldOptions 는 DistributionYield 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
//
// 두 필드 모두 Tiingo 문서의 파라미터 표에는 없지만 실제로 동작한다(실호출 확인).
// 지정하지 않으면 상장 이후 전 기간이 오므로 응답이 매우 클 수 있다 — AAPL 이 11,523건이다.
type YieldOptions struct {
	StartDate time.Time // 조회 시작(>=). zero 면 미전송
	EndDate   time.Time // 조회 종료(<=). zero 면 미전송
}

// params 는 옵션을 쿼리 맵으로 바꾼다. 넣을 값이 없으면 nil 이다.
func (o *YieldOptions) params() map[string]string {
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
	if len(q) == 0 {
		return nil
	}
	return q
}

// DistributionYield 는 배당수익률 시계열을 받는다.
// GET /tiingo/corporate-actions/<ticker>/distribution-yield
//
// 옵션 없이 부르면 상장 이후 전 기간이 온다(AAPL 실측 11,523건, 1980년부터).
// 기간을 좁히려면 YieldOptions 를 쓴다.
// 없는 티커는 404 가 아니라 빈 슬라이스다 — 이 점이 equity·iex 의 시세와 다르다.
func (c *Client) DistributionYield(ctx context.Context, ticker string, opts *YieldOptions) ([]Yield, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ys []Yield
	path := "/tiingo/corporate-actions/" + url.PathEscape(t) + "/distribution-yield"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ys); err != nil {
		return nil, err
	}
	return ys, nil
}
```

`types.DateLayout` 과 `types.Date` 가 존재하는지 확인한다(`eod/prices.go` 가 둘 다 쓴다).
`GetJSON` 시그니처가 위 호출과 다르면 설계를 바꾸지 말고 **BLOCKED 로 보고**한다.

- [ ] **Step 4: 테스트 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./corporateactions/... -v`
Expected: PASS — 10건

Run: `go build ./... && go vet ./... && gofmt -l corporateactions/` — 모두 무출력
Run: `go test ./corporateactions/... -cover` — 100% 여야 한다

- [ ] **Step 5: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add corporateactions/client.go corporateactions/yield.go corporateactions/yield_test.go corporateactions/testutil_test.go
git commit -m "feat(corporateactions): 배당수익률 시계열 추가"
```

---

### Task 3: 루트 클라이언트 배선과 예제

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`
- Create: `examples/corporateactions/main.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`client_test.go` 의 `TestNewClient` 안, `assert.NotNil(t, c.Equity)` 다음 줄에 추가한다:

```go
	assert.NotNil(t, c.CorporateActions)
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test . -run TestNewClient`
Expected: 컴파일 실패 — `c.CorporateActions undefined`

- [ ] **Step 3: 배선한다**

`client.go` 를 세 군데 고친다. 기존 정렬을 유지하고 `gofmt -w client.go` 로 마무리한다.

import 블록(경로 알파벳 순서상 `crypto` 앞):

```go
	"github.com/kenshin579/tiingo-go/corporateactions"
```

`Client` 구조체에서 `Equity` 다음 줄:

```go
	CorporateActions *corporateactions.Client // 배당수익률(배당 내역·분할은 권한 403 이라 미구현)
```

`NewClient` 에서 `c.Equity = equity.New(hc)` 다음 줄:

```go
	c.CorporateActions = corporateactions.New(hc)
```

**주의**: 이 필드 이름이 기존 필드들보다 길어 gofmt 가 구조체 전체의 주석 정렬을 다시 맞춘다.
`git diff` 에 다른 필드 줄이 함께 뜨는 것은 정상이다.

- [ ] **Step 4: 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test . -v -run TestNewClient`
Expected: PASS

- [ ] **Step 5: 예제를 만든다**

`examples/corporateactions/main.go`. `examples/search/main.go` 를 먼저 읽고 구조와 어투를 맞춘다.

```go
// Tiingo Corporate Actions 예제.
//
//	TIINGO_API_KEY=... go run ./examples/corporateactions
//
// 옵션 없이 부르면 상장 이후 전 기간(AAPL 은 1만 건 이상)이 오므로 기간을 좁혀 쓴다.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/corporateactions"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ys, err := c.CorporateActions.DistributionYield(ctx, "AAPL", &corporateactions.YieldOptions{
		StartDate: time.Now().AddDate(0, -1, 0),
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(ys) == 0 {
		fmt.Println("수익률 데이터 없음")
		return
	}
	fmt.Printf("AAPL 최근 1개월 배당수익률 %d건 (마지막 5건)\n", len(ys))
	for _, y := range ys[max(0, len(ys)-5):] {
		fmt.Printf("  %s  %.4f%%\n", y.Date, y.TrailingDiv1Y*100)
	}
}
```

- [ ] **Step 6: 빌드하고 실행한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && (gofmt -l . | grep -v '^$' || echo "포맷 OK")
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go run ./examples/corporateactions
```

Expected: 최근 1개월치 건수와 마지막 5건. 429 가 나면 정시까지 기다렸다 다시 시도하고,
그래도 안 되면 코드가 컴파일되고 `go test ./...` 가 통과하는 것만 확인한 뒤 커밋하되
**실행을 못 했다고 보고서에 분명히 적는다.**

**API 키를 출력하거나 보고서에 넣지 않는다.** 프로그램 출력은 넣어도 된다.

- [ ] **Step 7: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add client.go client_test.go examples/corporateactions/main.go
git commit -m "feat(corporateactions): 루트 클라이언트 배선과 예제 추가"
```

---

### Task 4: 통합 테스트

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: 실 API 테스트를 추가한다**

파일 끝에 붙인다. import 블록에 `"github.com/kenshin579/tiingo-go/corporateactions"` 를
순서에 맞게 추가한다.

```go
func TestIntegration_DistributionYield(t *testing.T) {
	c := newClient(t)
	ys, err := c.CorporateActions.DistributionYield(context.Background(), "AAPL",
		&corporateactions.YieldOptions{StartDate: time.Now().AddDate(0, -1, 0)})
	require.NoError(t, err)
	require.NotEmpty(t, ys)
	for _, y := range ys {
		assert.False(t, y.Date.IsZero())
		assert.GreaterOrEqual(t, y.TrailingDiv1Y, 0.0, "수익률은 음수가 될 수 없다")
	}
}

// 기간을 좁히면 실제로 줄어든다 — 문서 표에 없는 파라미터라 실 API 로 확인해 둔다.
func TestIntegration_DistributionYieldDateRange(t *testing.T) {
	c := newClient(t)
	narrow, err := c.CorporateActions.DistributionYield(context.Background(), "AAPL",
		&corporateactions.YieldOptions{
			StartDate: time.Now().AddDate(0, 0, -10),
			EndDate:   time.Now().AddDate(0, 0, -5),
		})
	require.NoError(t, err)
	assert.Less(t, len(narrow), 10, "닷새 구간이라 거래일 수만큼만 온다")
}

// 없는 티커는 404 가 아니라 빈 슬라이스다.
func TestIntegration_DistributionYieldUnknownTicker(t *testing.T) {
	c := newClient(t)
	ys, err := c.CorporateActions.DistributionYield(context.Background(), "NOSUCHTICKERXYZ", nil)
	require.NoError(t, err)
	assert.Empty(t, ys)
}
```

옵션 없는 전체 조회(11,523건)는 통합 테스트에 넣지 않는다 — 매 실행마다 대량 응답을 받을 이유가
없고 시간당 한도만 축낸다. 그 동작은 예제와 doc 주석으로 알린다.

- [ ] **Step 2: 실 API 로 실행한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration -run TestIntegration_DistributionYield -v .
```

Expected: 3건 PASS. 429 면 정시까지 기다렸다 재시도한다.

- [ ] **Step 3: 전체 통합 테스트를 돌린다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration . 2>&1 | tail -5
```

Expected: ok — 32건(기존 29 + 3). 한도에 걸리면 Step 2 결과가 핵심이므로 그대로 보고한다.

- [ ] **Step 4: 비통합 빌드가 멀쩡한지 본다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && go test ./... && (gofmt -l . | grep -v '^$' || echo "포맷 OK")
```

- [ ] **Step 5: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add integration_test.go
git commit -m "test(corporateactions): 실 API 통합 테스트 3건 추가"
```

---

### Task 5: 문서 갱신

**Files:**
- Modify: `README.md`
- Modify: `/Users/user/src/workspace_moneyflow/CLAUDE.md`

이 카테고리는 반쪽이므로 **무엇이 왜 빠졌는지** 적는 것이 핵심이다.

- [ ] **Step 1: README 사용 예에 추가한다**

`## 사용` 코드 블록의 Equity 줄 다음에 넣는다:

```go
// 배당수익률 시계열. 옵션 없이 부르면 상장 이후 전 기간이 온다
ys, _ := c.CorporateActions.DistributionYield(ctx, "AAPL",
    &corporateactions.YieldOptions{StartDate: time.Now().AddDate(0, -1, 0)})
```

"실행 가능한 예시" 줄 끝에 `, [`examples/corporateactions`](examples/corporateactions)` 를 더한다.

- [ ] **Step 2: 커버리지 표에 1행을 더한다**

Equity 행 다음에 넣는다:

```markdown
| Corporate Actions\*\* | `CorporateActions.DistributionYield` | `GET /tiingo/corporate-actions/<ticker>/distribution-yield` |
```

표 아래 Fundamentals 각주 다음에 각주를 하나 더한다:

```markdown
\*\* Corporate Actions 는 이 그룹의 5개 중 1개만 구현돼 있다. 배당 내역(distributions)과
분할(splits)은 무료 키에서 403 이라 응답을 확인할 수 없어 넣지 않았다.
```

- [ ] **Step 3: 미구현 목록 문장을 실측으로 갱신한다**

실제 문장을 읽고 정확히 고친다 — 기억으로 다시 쓰지 않는다. 현재는 대략
`나머지 REST 그룹(News — 별도 플랜 필요, BOATS — 별도 등록 필요, Fund Fees, Dividends/Splits)과
WebSocket 은 순차 추가 예정.` 이다. 2026-09-05 실측 결과를 반영해 아래로 바꾼다:

```markdown
나머지 REST 그룹은 무료 키로 접근이 막혀 있다 — News 와 Fund Fees 는 권한 없음(403),
BOATS 는 유료 add-on, Corporate Actions 의 배당 내역·분할도 403 이다. WebSocket 은 순차 추가 예정.
```

- [ ] **Step 4: `## 테스트` 절의 예제 빌드 목록에 더한다**

`go build -o /dev/null ./examples/equity` 다음 줄에:

```bash
go build -o /dev/null ./examples/corporateactions
```

이름 충돌 열거 주석이 있으면 그대로 둔다 — `corporateactions/` 는 예제 이름과 같은 디렉터리가
레포 루트에 있으므로 이 줄도 `-o` 가 필요하다.

- [ ] **Step 5: 워크스페이스 CLAUDE.md 를 갱신한다**

`/Users/user/src/workspace_moneyflow/CLAUDE.md` 의 `### tiingo-go` 절을 고친다.
**이 파일은 git 저장소가 아니므로 커밋하지 않는다.**

`**Module**:` 문단의 버전 문장을 바꾼다. 현재는 `**v0.7.0 (2026-09-05)**: 일곱 카테고리 — ...` 로
시작한다. 그 문장 끝에 여덟 번째를 더한다:

```
**v0.8.0 (2026-09-05)**: 여덟 카테고리 — 앞의 일곱에 Corporate Actions(`DistributionYield` — 이 그룹 5개 중 1개만, 나머지는 403) 추가.
```

(앞 일곱 개의 열거는 기존 문장을 그대로 살리고 버전·개수·마지막 항목만 고친다.)

`**다음**:` 문단을 바꾼다:

```
**다음**: REST 는 권한 장벽으로 소진 — News·Fund Fees 403, BOATS 유료 add-on, Corporate Actions 잔여 4개 403. 남은 것은 WebSocket 5종(무료 키 가능 여부 미확인)과 moneyflow 통합.
```

bash 블록의 예제 빌드 주석에 `corporateactions/` 를 더한다.

- [ ] **Step 6: 최종 검증**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && go test ./... && (gofmt -l . | grep -v '^$' || echo "포맷 OK")
go build -o /dev/null ./examples/corporateactions && echo "예제 빌드 OK"
file -I README.md /Users/user/src/workspace_moneyflow/CLAUDE.md
grep -n "CorporateActions\|Corporate Actions" README.md
```

두 파일 모두 `charset=utf-8` 이어야 한다.

- [ ] **Step 7: 커밋한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add README.md
git commit -m "docs(corporateactions): README 커버리지와 권한 상태 갱신"
```

---

## 완료 후

`superpowers:finishing-a-development-branch` 로 PR 을 만든다. 머지 후 v0.8.0 릴리스:

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go && ./scripts/release.sh v0.8.0
```

릴리스 뒤 신규 모듈에서 소비를 확인한다:

```bash
cd $(mktemp -d) && go mod init probe && GOPROXY=https://proxy.golang.org go get github.com/kenshin579/tiingo-go@v0.8.0
```

## 범위 밖

- 같은 그룹의 403 4개(배당 내역 3, 분할 1) — 권한이 열리면
- News(403), Fund Fees(403), BOATS(유료 add-on)
- WebSocket 5종, CSV, 재시도, rate limit 백오프, moneyflow 통합
