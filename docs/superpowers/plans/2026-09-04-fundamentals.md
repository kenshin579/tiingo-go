# Fundamentals 카테고리 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `tiingo-go` 에 Fundamentals 카테고리(지표 정의·회사 메타·재무제표·일별 지표) 4개 엔드포인트를 추가하고 v0.2.0 을 낸다.

**Architecture:** 기존 `eod` 와 동일한 카테고리 패키지 패턴 — `fundamentals.Client` 가 `internal/httpclient` 를 통해 호출하고 `types` 를 직접 import 한다(루트 import 는 순환). 응답의 `statementData` 는 타입 필드로 전개하지 않고 `[]DataPoint` 원문 + `Get`/`Map` 헬퍼 + 코드 상수로 다룬다. 시각이 의미 있는 필드를 위해 `types.Time` 을 신설한다.

**Tech Stack:** Go 1.25, testify, `httptest` 스텁, build tag `integration`.

**Spec:** `docs/superpowers/specs/2026-09-04-fundamentals-design.md`
**API 참조:** `docs/api/rest/fundamentals.md`

**실행 환경:** `TIINGO_API_KEY` 가 `~/.zshrc` 에 있다. **fixture 는 실호출로 저장한다**(EOD 때와 다름). 키를 읽을 때는 값을 출력하지 말 것:
```bash
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
```
Fundamentals 는 유료 add-on 이고 무료 플랜은 Dow 30 3년치다 — **fixture·통합 테스트 모두 AAPL 만 쓴다.**

---

## 파일 구조

| 파일 | 책임 |
|---|---|
| `types/time.go` | `Time` — 시각 보존 타임스탬프(RFC3339) |
| `types/time_test.go` | `Time` 단위 테스트 |
| `fundamentals/client.go` | `Client` 선언과 생성자 |
| `fundamentals/definitions.go` | `Definition` + `Definitions()` |
| `fundamentals/meta.go` | `Meta` + `Meta(ctx, tickers...)` |
| `fundamentals/statements.go` | `Statement`, `StatementData`, `DataPoint`, `StatementOptions`, `Get`/`Map` + `Statements()` |
| `fundamentals/daily.go` | `DailyMetric`, `DailyOptions` + `Daily()` |
| `fundamentals/codes.go` | dataCode 상수 85개(그룹별) |
| `fundamentals/testdata/*.json` | 실호출 fixture 4종 |
| `fundamentals/*_test.go` | 파싱·쿼리·헬퍼·상수 일치 테스트 |
| `client.go` (수정) | `Fundamentals *fundamentals.Client` 필드 |
| `integration_test.go` (수정) | Fundamentals 실호출 4건 |
| `examples/fundamentals/main.go` | 실행 가능한 예시 |
| `README.md` (수정) | 커버리지 표 4행 + 무료 플랜 각주 |

**주석 규약**: 짧은 설명은 필드 오른쪽 줄 끝, 길면 필드 위. 한국어. 타입·메서드에 doc 주석.

---

### Task 1: `types.Time`

**Files:** Create `types/time.go`, `types/time_test.go`

`statementLastUpdated` 같은 필드는 `2026-08-21T01:01:17.444Z` 로 오고 증분 동기화의 기준이라 시각을 버리면 못 쓴다. `Date` 는 날짜만 남기므로 별도 타입이 필요하다.

- [ ] **Step 1: 실패하는 테스트 작성** — `types/time_test.go`:

```go
package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"밀리초 UTC", `"2026-08-21T01:01:17.444Z"`, time.Date(2026, 8, 21, 1, 1, 17, 444000000, time.UTC)},
		{"초 단위", `"2026-08-21T01:01:17Z"`, time.Date(2026, 8, 21, 1, 1, 17, 0, time.UTC)},
		{"오프셋은 UTC 로 환산", `"2026-08-20T20:01:17-05:00"`, time.Date(2026, 8, 21, 1, 1, 17, 0, time.UTC)},
		{"null", `null`, time.Time{}},
		{"빈 문자열", `""`, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Time
			require.NoError(t, json.Unmarshal([]byte(tt.in), &v))
			assert.True(t, v.Time.Equal(tt.want), "got %v, want %v", v.Time, tt.want)
		})
	}
}

func TestTime_UnmarshalJSON_Invalid(t *testing.T) {
	var v Time
	assert.Error(t, json.Unmarshal([]byte(`"2026-08-21"`), &v), "날짜만 있는 값은 Time 이 아니다")
	assert.Error(t, json.Unmarshal([]byte(`12345`), &v))
}

func TestTime_Marshal(t *testing.T) {
	v := Time{Time: time.Date(2026, 8, 21, 1, 1, 17, 0, time.UTC)}
	b, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, `"2026-08-21T01:01:17Z"`, string(b))

	b, err = json.Marshal(Time{})
	require.NoError(t, err)
	assert.Equal(t, `null`, string(b))

	txt, err := v.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "2026-08-21T01:01:17Z", string(txt))

	txt, err = Time{}.MarshalText()
	require.NoError(t, err)
	assert.Empty(t, string(txt))
}

func TestTime_TextRoundTrip(t *testing.T) {
	var v Time
	require.NoError(t, v.UnmarshalText([]byte("2026-08-21T01:01:17.444Z")))
	assert.Equal(t, 444000000, v.Time.Nanosecond(), "시각을 버리지 않는다")
	assert.Equal(t, "2026-08-21T01:01:17.444Z", v.String())
	require.NoError(t, v.UnmarshalText([]byte("")))
	assert.True(t, v.IsZero())
	assert.Error(t, (&Time{}).UnmarshalText([]byte("nope")))
}
```

Run `go test ./types/ -run TestTime` → FAIL (`undefined: Time`).

- [ ] **Step 2: `types/time.go` 구현**

```go
package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Time 은 시각까지 의미 있는 Tiingo 타임스탬프다(예: statementLastUpdated).
// 날짜만 필요한 필드는 Date 를 쓴다 — Date 는 시각을 버리고 날짜만 남긴다.
// null 이나 빈 문자열은 zero value 가 되며 IsZero() 로 구분한다.
type Time struct {
	time.Time
}

// parse 는 RFC3339 를 UTC 로 파싱한다. 빈 문자열은 zero value 다.
func (t *Time) parse(s string) error {
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("tiingo: unrecognized timestamp %q", s)
	}
	t.Time = parsed.UTC()
	return nil
}

// UnmarshalJSON 은 RFC3339 문자열을 받는다.
func (t *Time) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		t.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("tiingo: timestamp must be a JSON string: %w", err)
	}
	return t.parse(s)
}

// UnmarshalText 는 UnmarshalJSON 과 같은 형식을 받는다.
func (t *Time) UnmarshalText(b []byte) error { return t.parse(string(b)) }

// MarshalJSON 은 RFC3339 문자열로 쓴다. zero value 는 null.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.String())
}

// MarshalText 는 RFC3339 문자열로 쓴다. zero value 는 빈 문자열.
func (t Time) MarshalText() ([]byte, error) { return []byte(t.String()), nil }

// String 은 RFC3339 를 반환한다. zero value 는 빈 문자열.
func (t Time) String() string {
	if t.Time.IsZero() {
		return ""
	}
	return t.Time.Format(time.RFC3339Nano)
}
```

Run `go test ./types/ -v` → 기존 Date 테스트 + Time 4개 모두 PASS. `go vet ./...`, `gofmt -l .` 클린.

**주의**: `String()` 이 `RFC3339Nano` 를 쓰므로 나노초가 0 이면 `2026-08-21T01:01:17Z`, 있으면 `...17.444Z` 가 된다 — 위 테스트가 두 경우를 모두 고정한다.

- [ ] **Step 3: 커밋**

```bash
git add types/time.go types/time_test.go
git commit -m "feat(types): Time — 시각 보존 RFC3339 타임스탬프"
```

---

### Task 2: fixture 4종 저장 (실호출)

**Files:** Create `fundamentals/testdata/{definitions,meta_aapl,statements_aapl,daily_aapl}.json`

- [ ] **Step 1: 네 응답 저장**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
mkdir -p fundamentals/testdata
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
B=https://api.tiingo.com/tiingo/fundamentals
curl -fsS -H "Authorization: Token $KEY" "$B/definitions"                              | python3 -m json.tool > fundamentals/testdata/definitions.json
curl -fsS -H "Authorization: Token $KEY" "$B/meta?tickers=aapl"                         | python3 -m json.tool > fundamentals/testdata/meta_aapl.json
curl -fsS -H "Authorization: Token $KEY" "$B/aapl/statements?startDate=2026-01-01"      | python3 -m json.tool > fundamentals/testdata/statements_aapl.json
curl -fsS -H "Authorization: Token $KEY" "$B/aapl/daily?startDate=2026-09-01"           | python3 -m json.tool > fundamentals/testdata/daily_aapl.json
```

- [ ] **Step 2: 검증**

```bash
for f in fundamentals/testdata/*.json; do python3 -m json.tool "$f" > /dev/null && echo "ok $f"; done
python3 -c "
import json
d=json.load(open('fundamentals/testdata/definitions.json')); print('definitions:',len(d))
m=json.load(open('fundamentals/testdata/meta_aapl.json')); print('meta fields:',len(m[0]),sorted(m[0])[:4])
s=json.load(open('fundamentals/testdata/statements_aapl.json')); print('statements:',len(s),'first q:',s[0]['quarter'],'groups:',list(s[0]['statementData']))
y=json.load(open('fundamentals/testdata/daily_aapl.json')); print('daily:',len(y),sorted(y[0]))
"
grep -rl "$(printf 'Token ')" fundamentals/testdata || echo "no token leaked"
```
Expected: 네 파일 모두 valid, definitions 85 내외, meta 17필드, statements 는 `statementData` 4그룹, daily 는 6필드. 파일에 토큰 문자열이 없어야 한다(응답에는 원래 없지만 확인).

- [ ] **Step 3: 커밋**

```bash
git add fundamentals/testdata/
git commit -m "test(fundamentals): 실호출 fixture 4종"
```

---

### Task 3: 패키지 골격 + `types.Time` 이 쓰이는 Meta

**Files:** Create `fundamentals/client.go`, `fundamentals/meta.go`, `fundamentals/testutil_test.go`, `fundamentals/meta_test.go`; Modify `client.go`

- [ ] **Step 1: `fundamentals/client.go`**

```go
// Package fundamentals 는 Tiingo Fundamentals(재무제표·일별 지표·지표 정의·회사 메타) API
// sub-client 다. tiingo.Client.Fundamentals 로 접근한다.
//
// Fundamentals 는 별도 구독(add-on)이다. 무료 플랜은 Dow 30 종목의 3년치만 제공하며,
// 권한 밖 종목은 400/403 이 오고 APIError 로 그대로 전달된다.
package fundamentals

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Fundamentals sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
```

- [ ] **Step 2: 루트 `client.go` 에 필드 추가**

`import` 에 `"github.com/kenshin579/tiingo-go/fundamentals"` 를 넣고:

```go
type Client struct {
	http *httpclient.Client

	EOD          *eod.Client          // 일별 시세·메타(End-of-Day)
	Fundamentals *fundamentals.Client // 재무제표·일별 지표(Fundamentals, 별도 구독)
}
```
`NewClient` 안에서 `c.Fundamentals = fundamentals.New(hc)` 를 `c.EOD = ...` 다음 줄에 추가.

- [ ] **Step 3: 테스트 헬퍼** — `fundamentals/testutil_test.go` (eod 의 것과 같은 모양):

```go
package fundamentals

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
)

// newStubClient 는 고정 응답을 돌려주는 테스트 서버에 붙은 Client 와,
// 서버가 받은 마지막 요청을 읽는 함수를 돌려준다.
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

- [ ] **Step 4: 실패하는 테스트** — `fundamentals/meta_test.go`:

```go
package fundamentals

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

func TestMeta_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/meta_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ms, err := c.Meta(context.Background(), "aapl")
	require.NoError(t, err)
	require.Len(t, ms, 1)

	m := ms[0]
	assert.Equal(t, "aapl", m.Ticker)
	assert.Equal(t, "Apple Inc", m.Name)
	assert.Equal(t, "US000000000038", m.PermaTicker)
	assert.True(t, m.IsActive)
	assert.False(t, m.IsADR)
	assert.Equal(t, "Technology", m.Sector)
	assert.Equal(t, 3571, m.SICCode)
	assert.Equal(t, "usd", m.ReportingCurrency)
	assert.NotEmpty(t, m.DataProviderPermaTicker, "문서 표에 없지만 실제로 오는 필드")
	assert.False(t, m.StatementLastUpdated.IsZero())
	assert.NotZero(t, m.StatementLastUpdated.Hour()+m.StatementLastUpdated.Minute()+m.StatementLastUpdated.Second(),
		"types.Time 이므로 시각이 보존돼야 한다")

	assert.Equal(t, "/tiingo/fundamentals/meta", lastReq().URL.Path)
	assert.Equal(t, "aapl", lastReq().URL.Query().Get("tickers"))
}

func TestMeta_MultipleTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background(), "aapl", "msft")
	require.NoError(t, err)
	assert.Equal(t, "aapl,msft", lastReq().URL.Query().Get("tickers"))
}

func TestMeta_NoTickers_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background())
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "티커를 안 주면 쿼리 없이 전체 조회")
}

func TestMeta_EmptyTickerElement(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background(), "aapl", "  ")
	assert.Error(t, err, "빈 티커가 섞이면 에러")
}

func TestMeta_Forbidden(t *testing.T) {
	c, _ := newStubClient(t, http.StatusForbidden, `{"detail":"You do not have permission"}`)
	_, err := c.Meta(context.Background(), "aapl")
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}
```

Run `go test ./fundamentals/ -run TestMeta` → FAIL (`c.Meta undefined`).

- [ ] **Step 5: `fundamentals/meta.go` 구현**

```go
package fundamentals

import (
	"context"
	"fmt"
	"strings"

	"github.com/kenshin579/tiingo-go/types"
)

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

// Meta 는 펀더멘털 커버리지 회사의 메타 정보를 받는다. GET /tiingo/fundamentals/meta
// tickers 를 주지 않으면 커버리지 전체(수천 건)를 받는다 — 대량 응답이므로 주의한다.
func (c *Client) Meta(ctx context.Context, tickers ...string) ([]Meta, error) {
	var params map[string]string
	if len(tickers) > 0 {
		cleaned := make([]string, 0, len(tickers))
		for _, t := range tickers {
			t = strings.TrimSpace(t)
			if t == "" {
				return nil, fmt.Errorf("tiingo: ticker must not be empty")
			}
			cleaned = append(cleaned, t)
		}
		params = map[string]string{"tickers": strings.Join(cleaned, ",")}
	}
	var ms []Meta
	if err := c.http.GetJSON(ctx, "/tiingo/fundamentals/meta", params, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}
```

- [ ] **Step 6: 검증 + 커밋**

Run `go build ./... && go test ./... && go vet ./... && gofmt -l .`
Expected: meta 테스트 5개 PASS, 전체 클린.

```bash
git add fundamentals/client.go fundamentals/meta.go fundamentals/testutil_test.go fundamentals/meta_test.go client.go
git commit -m "feat(fundamentals): 패키지 골격 + 회사 메타 (Meta)"
```

---

### Task 3.5: Task 1 리뷰 반영 (완료)

리뷰 지적 2건, 둘 다 계획의 범위 누락이었다.

1. (Important) 루트에 `Date` 별칭만 있고 `Time` 별칭이 없어 비대칭 — `Meta` 필드가 `types.Time` 이라 사용자가 타입 이름을 쓰려면 `types` 를 따로 import 해야 했다. `types.go` 에 `type Time = types.Time` 추가, `types_alias_test.go` 에 `TestTimeAlias` 추가.
2. (Important) `types/date.go` 주석이 `Time` 도입 전 내용("시각이 의미 있는 필드는 `time.Time` 을 그대로 쓴다") → "`Time` 을 쓴다"로 갱신. `types/time.go` 의 `String()` 주석에 RFC3339Nano 의 뒤쪽 0 절삭 동작을 한 줄 추가.

커밋 `fix(types): 루트 tiingo.Time 별칭 추가, Date 주석 갱신 (Task 1 리뷰 반영)`.

**Task 3 구현 시 유념(리뷰 Minor)**: `Meta` 를 티커 없이 부르면 수천 건을 받는데, `encoding/json` 은 슬라이스 원소 하나라도 `UnmarshalJSON` 이 실패하면 배치 전체가 실패한다. `null`·빈 문자열은 안전하지만 Tiingo 가 타임스탬프 포맷을 바꾸면 전량 실패할 수 있다.

---

### Task 4: 지표 정의 + 코드 상수

**Files:** Create `fundamentals/definitions.go`, `fundamentals/definitions_test.go`, `fundamentals/codes.go`, `fundamentals/codes_test.go`

- [ ] **Step 1: 실패하는 테스트** — `fundamentals/definitions_test.go`:

```go
package fundamentals

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefinitions_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/definitions.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ds, err := c.Definitions(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, ds)

	byCode := map[string]Definition{}
	for _, d := range ds {
		byCode[d.DataCode] = d
	}
	rev, ok := byCode["revenue"]
	require.True(t, ok, "revenue 정의가 있어야 한다")
	assert.NotEmpty(t, rev.Name)
	assert.Equal(t, "incomeStatement", rev.StatementType)

	types := map[string]int{}
	for _, d := range ds {
		types[d.StatementType]++
	}
	for _, st := range []string{"balanceSheet", "incomeStatement", "cashFlow", "overview"} {
		assert.NotZero(t, types[st], "%s 지표가 있어야 한다", st)
	}

	assert.Equal(t, "/tiingo/fundamentals/definitions", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}
```

`fundamentals/codes_test.go`:

```go
package fundamentals

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// fixture 의 dataCode 가 전부 상수로 존재하는지 확인한다.
// Tiingo 가 지표를 추가하면 fixture 갱신 시 이 테스트가 상수 누락을 잡는다.
func TestCodes_CoverFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/definitions.json")
	require.NoError(t, err)
	var ds []Definition
	require.NoError(t, json.Unmarshal(raw, &ds))
	require.NotEmpty(t, ds)

	known := map[string]bool{}
	for _, c := range AllCodes {
		known[c] = true
	}
	var missing []string
	for _, d := range ds {
		if !known[d.DataCode] {
			missing = append(missing, d.DataCode)
		}
	}
	require.Emptyf(t, missing, "codes.go 에 없는 dataCode: %v", missing)
	require.Len(t, AllCodes, len(ds), "상수 개수와 정의 개수가 같아야 한다")
}
```

Run `go test ./fundamentals/ -run "TestDefinitions|TestCodes"` → FAIL (`c.Definitions undefined`, `AllCodes undefined`).

- [ ] **Step 2: `fundamentals/definitions.go`**

```go
package fundamentals

import "context"

// Definition 은 지표 하나의 정의다. Definitions 로 전체 목록을 받는다.
type Definition struct {
	DataCode      string `json:"dataCode"`      // 지표 식별자(예: peRatio). codes.go 상수와 같은 값
	Name          string `json:"name"`          // 사람이 읽는 이름(예: Revenue Per Share)
	Description   string `json:"description"`   // 지표 설명
	StatementType string `json:"statementType"` // 소속 묶음: balanceSheet / incomeStatement / cashFlow / overview
	Units         string `json:"units"`         // 단위. "$" 금액, "%" 비율, 빈 값이면 무차원(배수 등)
}

// Definitions 는 사용 가능한 지표 정의 전체를 받는다. GET /tiingo/fundamentals/definitions
// 지표가 추가되면 목록도 늘어난다(2026-09-04 기준 85개).
func (c *Client) Definitions(ctx context.Context) ([]Definition, error) {
	var ds []Definition
	if err := c.http.GetJSON(ctx, "/tiingo/fundamentals/definitions", nil, &ds); err != nil {
		return nil, err
	}
	return ds, nil
}
```

- [ ] **Step 3: `fundamentals/codes.go` 생성**

fixture 에서 기계 생성한다(상수명 = `Code` + dataCode 의 첫 글자 대문자; 85개 모두 충돌 없음을 확인함):

```bash
python3 - <<'PY' > fundamentals/codes.go
import json, collections
d = json.load(open('fundamentals/testdata/definitions.json'))
def const(c): return 'Code' + c[0].upper() + c[1:]
g = collections.defaultdict(list)
for x in d: g[x['statementType']].append(x)
titles = {'incomeStatement':'손익계산서', 'balanceSheet':'재무상태표', 'cashFlow':'현금흐름표', 'overview':'개요·비율'}
out = []
out.append('package fundamentals')
out.append('')
out.append('// 이 파일은 /tiingo/fundamentals/definitions 응답에서 생성했다(2026-09-04, %d개).' % len(d))
out.append('// Tiingo 는 지표를 계속 추가하므로 목록이 최신이 아닐 수 있다. 상수가 없어도')
out.append('// StatementData.Get 은 임의 문자열을 받으므로 동작에는 지장이 없다.')
out.append('')
for st in ['incomeStatement','balanceSheet','cashFlow','overview']:
    items = sorted(g[st], key=lambda y: y['dataCode'])
    out.append('// %s(%s) 지표 %d개' % (titles[st], st, len(items)))
    out.append('const (')
    for x in items:
        u = x.get('units') or ''
        suffix = ' (%s)' % u if u else ''
        out.append('\t%s = "%s" // %s%s' % (const(x['dataCode']), x['dataCode'], x['name'], suffix))
    out.append(')')
    out.append('')
out.append('// AllCodes 는 위 상수 전체다. 정의 목록과의 동기화 검사에 쓴다.')
out.append('var AllCodes = []string{')
for st in ['incomeStatement','balanceSheet','cashFlow','overview']:
    for x in sorted(g[st], key=lambda y: y['dataCode']):
        out.append('\t%s,' % const(x['dataCode']))
out.append('}')
print('\n'.join(out))
PY
gofmt -w fundamentals/codes.go
head -20 fundamentals/codes.go
grep -c '^\t[A-Z]' fundamentals/codes.go   # 상수 + AllCodes 항목
```
Expected: 파일 상단에 생성 출처 주석, 그룹별 `const` 블록 4개, `AllCodes` 85개. `gofmt` 후 정렬된 주석.

- [ ] **Step 4: 검증 + 커밋**

Run `go build ./... && go test ./fundamentals/ -v && go vet ./... && gofmt -l .`
Expected: definitions 1개 + codes 1개 테스트 PASS(기존 meta 5개 포함).

```bash
git add fundamentals/definitions.go fundamentals/definitions_test.go fundamentals/codes.go fundamentals/codes_test.go
git commit -m "feat(fundamentals): 지표 정의 (Definitions) + dataCode 상수 85개"
```

---

### Task 5: 재무제표

**Files:** Create `fundamentals/statements.go`, `fundamentals/statements_test.go`

- [ ] **Step 1: 실패하는 테스트** — `fundamentals/statements_test.go`:

```go
package fundamentals

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatements_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/statements_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Statements(context.Background(), "aapl", &StatementOptions{
		StartDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Sort:       "-date",
		AsReported: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ss)

	s := ss[0]
	assert.False(t, s.Date.IsZero())
	assert.NotZero(t, s.Year)
	assert.NotEmpty(t, s.StatementData.BalanceSheet)
	assert.NotEmpty(t, s.StatementData.IncomeStatement)
	assert.NotEmpty(t, s.StatementData.CashFlow)
	assert.NotEmpty(t, s.StatementData.Overview)

	rev, ok := s.StatementData.Get(CodeRevenue)
	require.True(t, ok, "revenue 가 있어야 한다")
	assert.Greater(t, rev, 0.0)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/fundamentals/aapl/statements", lastReq().URL.Path)
	assert.Equal(t, "2026-01-01", q.Get("startDate"))
	assert.Equal(t, "2026-12-31", q.Get("endDate"))
	assert.Equal(t, "-date", q.Get("sort"))
	assert.Equal(t, "true", q.Get("asReported"))
}

func TestStatements_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Statements(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestStatements_AsReportedFalse_Omitted(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Statements(context.Background(), "aapl", &StatementOptions{Sort: "date"})
	require.NoError(t, err)
	q := lastReq().URL.Query()
	assert.Equal(t, "date", q.Get("sort"))
	assert.Empty(t, q.Get("asReported"), "기본값이면 파라미터를 보내지 않는다")
	assert.Empty(t, q.Get("startDate"))
}

func TestStatements_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Statements(context.Background(), "  ", nil)
	assert.Error(t, err)
}

func TestStatementData_Get(t *testing.T) {
	sd := StatementData{
		IncomeStatement: []DataPoint{{DataCode: "revenue", Value: 100}},
		Overview:        []DataPoint{{DataCode: "peRatio", Value: 0}},
	}
	v, ok := sd.Get("revenue")
	assert.True(t, ok)
	assert.Equal(t, 100.0, v)

	v, ok = sd.Get("peRatio")
	assert.True(t, ok, "값이 0 이어도 존재하면 true")
	assert.Equal(t, 0.0, v)

	_, ok = sd.Get("nosuchcode")
	assert.False(t, ok, "없는 코드는 false — 값 0 과 구분된다")

	var empty StatementData
	_, ok = empty.Get("revenue")
	assert.False(t, ok, "nil 슬라이스도 안전하다")
}

func TestStatementData_Map(t *testing.T) {
	sd := StatementData{
		BalanceSheet:    []DataPoint{{DataCode: "totalAssets", Value: 1}},
		IncomeStatement: []DataPoint{{DataCode: "revenue", Value: 2}},
		CashFlow:        []DataPoint{{DataCode: "capex", Value: 3}},
		Overview:        []DataPoint{{DataCode: "bvps", Value: 4}},
	}
	m := sd.Map()
	assert.Len(t, m, 4)
	assert.Equal(t, 2.0, m["revenue"])
	assert.Empty(t, StatementData{}.Map())
}
```

Run `go test ./fundamentals/ -run "TestStatement"` → FAIL (`undefined: StatementOptions`).

- [ ] **Step 2: `fundamentals/statements.go`**

```go
package fundamentals

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// DataPoint 는 지표 하나의 코드와 값이다.
type DataPoint struct {
	DataCode string  `json:"dataCode"` // 지표 식별자(예: revenue). codes.go 의 상수 참고
	Value    float64 `json:"value"`    // 지표 값
}

// StatementData 는 한 기간의 재무 데이터를 네 묶음으로 나눈 것이다.
// 지표 집합은 기간마다 다르고 Tiingo 가 계속 추가하므로 코드→값 조회로 접근한다.
type StatementData struct {
	BalanceSheet    []DataPoint `json:"balanceSheet"`    // 재무상태표 지표
	IncomeStatement []DataPoint `json:"incomeStatement"` // 손익계산서 지표
	CashFlow        []DataPoint `json:"cashFlow"`        // 현금흐름표 지표
	Overview        []DataPoint `json:"overview"`        // 여러 재무제표를 조합한 비율·지표
}

// groups 는 네 묶음을 순회 순서대로 돌려준다.
func (s StatementData) groups() [][]DataPoint {
	return [][]DataPoint{s.IncomeStatement, s.BalanceSheet, s.CashFlow, s.Overview}
}

// Get 은 네 묶음 전체에서 코드를 찾는다. 없으면 ok 가 false 다(값 0 과 구분된다).
// 어느 묶음에 있는지 몰라도 되도록 전체를 훑는다 — 묶음을 특정하려면 필드를 직접 순회한다.
func (s StatementData) Get(code string) (float64, bool) {
	for _, g := range s.groups() {
		for _, dp := range g {
			if dp.DataCode == code {
				return dp.Value, true
			}
		}
	}
	return 0, false
}

// Map 은 코드→값 맵을 만든다. 여러 지표를 반복 조회할 때 쓴다.
func (s StatementData) Map() map[string]float64 {
	m := map[string]float64{}
	for _, g := range s.groups() {
		for _, dp := range g {
			m[dp.DataCode] = dp.Value
		}
	}
	return m
}

// Statement 는 한 회계기간의 재무제표다.
type Statement struct {
	Date          types.Date    `json:"date"`    // asReported=false 면 회계기간 종료일, true 면 SEC 공개일
	Year          int           `json:"year"`    // 회계연도
	Quarter       int           `json:"quarter"` // 0 이면 연간 보고서, 1~4 는 해당 분기
	StatementData StatementData `json:"statementData"`
}

// StatementOptions 는 Statements 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type StatementOptions struct {
	StartDate time.Time // 조회 시작일(이상, >=). zero 면 미전송
	EndDate   time.Time // 조회 종료일(이하, <=). zero 면 미전송
	Sort      string    // 정렬 컬럼. Statements 는 "date" / "-date" 만 지원한다
	// AsReported 가 true 면 SEC 공개 시점의 원본 수치를 받는다(date 는 공개일).
	// false(기본)면 수정·재작성을 반영한 최신 수치이고 date 는 회계기간 종료일이다.
	// 같은 기간을 조회해도 두 모드의 건수와 날짜가 다르다.
	AsReported bool
}

// params 는 옵션을 쿼리 맵으로 바꾼다. zero 값은 넣지 않는다.
func (o *StatementOptions) params() map[string]string {
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
	if o.Sort != "" {
		q["sort"] = o.Sort
	}
	if o.AsReported {
		q["asReported"] = "true" // 기본값(false)일 때는 보내지 않는다
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Statements 는 분기·연간 재무제표를 받는다. GET /tiingo/fundamentals/<ticker>/statements
// ticker 자리에는 티커 또는 permaTicker(상장폐지·재사용 심볼용)를 넣는다.
func (c *Client) Statements(ctx context.Context, ticker string, opts *StatementOptions) ([]Statement, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ss []Statement
	path := "/tiingo/fundamentals/" + url.PathEscape(t) + "/statements"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ss); err != nil {
		return nil, err
	}
	return ss, nil
}
```

- [ ] **Step 3: 검증 + 커밋**

Run `go test ./fundamentals/ -v && go vet ./... && gofmt -l .`

```bash
git add fundamentals/statements.go fundamentals/statements_test.go
git commit -m "feat(fundamentals): 재무제표 (Statements) + StatementData 조회 헬퍼"
```

---

### Task 6: 일별 지표

**Files:** Create `fundamentals/daily.go`, `fundamentals/daily_test.go`

- [ ] **Step 1: 실패하는 테스트** — `fundamentals/daily_test.go`:

```go
package fundamentals

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaily_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/daily_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ds, err := c.Daily(context.Background(), "aapl", &DailyOptions{
		StartDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Sort:      "-date",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ds)

	d := ds[0]
	assert.False(t, d.Date.IsZero())
	assert.Greater(t, d.MarketCap, 0.0)
	assert.Greater(t, d.EnterpriseVal, 0.0)
	assert.Greater(t, d.PERatio, 0.0)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/fundamentals/aapl/daily", lastReq().URL.Path)
	assert.Equal(t, "2026-09-01", q.Get("startDate"))
	assert.Equal(t, "-date", q.Get("sort"))
	assert.Empty(t, q.Get("endDate"))
}

func TestDaily_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Daily(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestDaily_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Daily(context.Background(), "", nil)
	assert.Error(t, err)
}
```

Run `go test ./fundamentals/ -run TestDaily` → FAIL (`undefined: DailyOptions`).

- [ ] **Step 2: `fundamentals/daily.go`**

```go
package fundamentals

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// DailyMetric 은 주가 기반 일별 지표다. 재무제표와 달리 매일 갱신된다.
type DailyMetric struct {
	Date          types.Date `json:"date"`          // 지표 기준일
	MarketCap     float64    `json:"marketCap"`     // 시가총액
	EnterpriseVal float64    `json:"enterpriseVal"` // 기업가치(EV)
	PERatio       float64    `json:"peRatio"`       // 주가수익비율(P/E)
	PBRatio       float64    `json:"pbRatio"`       // 주가순자산비율(P/B)
	TrailingPEG1Y float64    `json:"trailingPEG1Y"` // 최근 1년 기준 PEG
}

// DailyOptions 는 Daily 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type DailyOptions struct {
	StartDate time.Time // 조회 시작일(이상, >=). zero 면 미전송
	EndDate   time.Time // 조회 종료일(이하, <=). zero 면 미전송
	Sort      string    // 정렬 컬럼. "date" 오름차순, "-date" 내림차순
}

// params 는 옵션을 쿼리 맵으로 바꾼다. zero 값은 넣지 않는다.
func (o *DailyOptions) params() map[string]string {
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
	if o.Sort != "" {
		q["sort"] = o.Sort
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Daily 는 주가 기반 일별 지표를 받는다. GET /tiingo/fundamentals/<ticker>/daily
// ticker 자리에는 티커 또는 permaTicker(상장폐지·재사용 심볼용)를 넣는다.
func (c *Client) Daily(ctx context.Context, ticker string, opts *DailyOptions) ([]DailyMetric, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ds []DailyMetric
	path := "/tiingo/fundamentals/" + url.PathEscape(t) + "/daily"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ds); err != nil {
		return nil, err
	}
	return ds, nil
}
```

- [ ] **Step 3: 검증 + 커밋**

Run `go test ./... -race && go vet ./... && gofmt -l .`

```bash
git add fundamentals/daily.go fundamentals/daily_test.go
git commit -m "feat(fundamentals): 일별 지표 (Daily)"
```

---

### Task 7: 예제 + 통합 테스트

**Files:** Create `examples/fundamentals/main.go`; Modify `integration_test.go`

- [ ] **Step 1: `examples/fundamentals/main.go`**

```go
// Tiingo Fundamentals 예제.
//
//	TIINGO_API_KEY=... go run ./examples/fundamentals
//
// Fundamentals 는 별도 구독이다. 무료 플랜은 Dow 30 종목의 3년치만 조회된다.
package main

import (
	"context"
	"fmt"
	"log"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/fundamentals"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ms, err := c.Fundamentals.Meta(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	if len(ms) == 0 {
		log.Fatal("메타 없음")
	}
	m := ms[0]
	fmt.Printf("%s — %s / %s (%s)\n", m.Name, m.Sector, m.Industry, m.Location)
	fmt.Printf("  재무제표 갱신 %s, 일별 지표 갱신 %s\n", m.StatementLastUpdated, m.DailyLastUpdated)

	ss, err := c.Fundamentals.Statements(ctx, "AAPL", &fundamentals.StatementOptions{Sort: "-date"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("재무제표 %d기간\n", len(ss))
	for _, s := range ss[:min(3, len(ss))] {
		rev, _ := s.StatementData.Get(fundamentals.CodeRevenue)
		ni, _ := s.StatementData.Get(fundamentals.CodeNetinc)
		label := fmt.Sprintf("%dQ%d", s.Year, s.Quarter)
		if s.Quarter == 0 {
			label = fmt.Sprintf("%d 연간", s.Year)
		}
		fmt.Printf("  %-10s %s  매출 %.0f  순이익 %.0f\n", label, s.Date, rev, ni)
	}

	ds, err := c.Fundamentals.Daily(ctx, "AAPL", &fundamentals.DailyOptions{Sort: "-date"})
	if err != nil {
		log.Fatal(err)
	}
	if len(ds) > 0 {
		d := ds[0]
		fmt.Printf("일별 지표 %s: 시총 %.0f  P/E %.2f  P/B %.2f\n", d.Date, d.MarketCap, d.PERatio, d.PBRatio)
	}
}
```

**주의**: `CodeNetinc` 의 정확한 상수명은 Task 4 에서 생성된 `codes.go` 를 보고 맞춘다(dataCode 가 `netinc` 이면 `CodeNetinc`). `min` 은 Go 1.21+ 내장 함수다.

- [ ] **Step 2: `integration_test.go` 에 4건 추가**

기존 EOD 테스트 아래에 붙인다(파일 상단 import 에 `"github.com/kenshin579/tiingo-go/fundamentals"` 추가):

```go
func TestIntegration_FundamentalsDefinitions(t *testing.T) {
	c := newClient(t)
	ds, err := c.Fundamentals.Definitions(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ds)

	known := map[string]bool{}
	for _, code := range fundamentals.AllCodes {
		known[code] = true
	}
	var missing []string
	for _, d := range ds {
		if !known[d.DataCode] {
			missing = append(missing, d.DataCode)
		}
	}
	assert.Emptyf(t, missing, "codes.go 갱신 필요 — 새 dataCode: %v", missing)
}

func TestIntegration_FundamentalsMeta(t *testing.T) {
	c := newClient(t)
	ms, err := c.Fundamentals.Meta(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, "Apple Inc", ms[0].Name)
	assert.False(t, ms[0].StatementLastUpdated.IsZero())
}

func TestIntegration_FundamentalsStatements(t *testing.T) {
	c := newClient(t)
	ss, err := c.Fundamentals.Statements(context.Background(), "AAPL", &fundamentals.StatementOptions{
		StartDate: time.Now().AddDate(-1, 0, 0),
		Sort:      "-date",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ss)
	rev, ok := ss[0].StatementData.Get(fundamentals.CodeRevenue)
	assert.True(t, ok)
	assert.Greater(t, rev, 0.0)
}

func TestIntegration_FundamentalsDaily(t *testing.T) {
	c := newClient(t)
	ds, err := c.Fundamentals.Daily(context.Background(), "AAPL", &fundamentals.DailyOptions{
		StartDate: time.Now().AddDate(0, 0, -14),
	})
	require.NoError(t, err)
	require.NotEmpty(t, ds)
	assert.Greater(t, ds[0].MarketCap, 0.0)
}
```

- [ ] **Step 3: 실행 검증**

```bash
go vet ./... && go vet -tags integration ./...
go build -o /dev/null ./examples/fundamentals
go test ./... -count=1
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration ./... -run TestIntegration -v 2>&1 | grep -E "^(--- |ok|FAIL)"
TIINGO_API_KEY="$KEY" go run ./examples/fundamentals
```
Expected: 통합 테스트 8건(EOD 4 + Fundamentals 4) 전부 PASS, 예제가 회사·재무제표·일별 지표를 출력. `TestIntegration_FundamentalsDefinitions` 가 실패하면 Tiingo 가 지표를 추가한 것이므로 fixture 와 `codes.go` 를 재생성한다(Task 2·4 의 명령 재실행).

- [ ] **Step 4: 커밋**

```bash
git add examples/fundamentals/main.go integration_test.go
git commit -m "feat(fundamentals): 예제 + integration 테스트"
```

---

### Task 8: README + 최종 검증 + PR

**Files:** Modify `README.md`

- [ ] **Step 1: README 갱신**

커버리지 표에 4행 추가:

```markdown
| Fundamentals* | `Fundamentals.Definitions` | `GET /tiingo/fundamentals/definitions` |
| Fundamentals* | `Fundamentals.Meta` | `GET /tiingo/fundamentals/meta` |
| Fundamentals* | `Fundamentals.Statements` | `GET /tiingo/fundamentals/<ticker>/statements` |
| Fundamentals* | `Fundamentals.Daily` | `GET /tiingo/fundamentals/<ticker>/daily` |

\* Fundamentals 는 별도 구독(add-on)이다. 무료 플랜은 Dow 30 종목의 3년치만 제공하며, 권한 밖
종목은 `APIError`(400/403)로 돌아온다.
```

"나머지 REST 그룹" 문장에서 Fundamentals 를 뺀다. 사용 예시 절에 재무제표 조회를 추가:

```go
// 재무제표(최신 기간부터)
ss, _ := c.Fundamentals.Statements(ctx, "AAPL", &fundamentals.StatementOptions{Sort: "-date"})
rev, ok := ss[0].StatementData.Get(fundamentals.CodeRevenue) // 지표는 코드로 조회한다
```

"날짜 타입" 절에 한 문단 추가: `types.Time` 은 시각까지 보존하며 `statementLastUpdated` 처럼 갱신
시각이 의미 있는 필드에 쓴다.

- [ ] **Step 2: 최종 검증**

```bash
go build ./... && go vet ./... && go vet -tags integration ./...
go test ./... -count=1 && go test ./... -race -count=1
gofmt -l . | grep -v '^docs/' || echo "gofmt clean"
go doc github.com/kenshin579/tiingo-go/fundamentals | head -30
git status --short
```
Expected: 모두 통과, `go doc` 에 `Client`, `Definition`, `Definitions`, `Meta`, `Statement`, `StatementData`, `DataPoint`, `StatementOptions`, `DailyMetric`, `DailyOptions`, `AllCodes` 와 상수가 주석과 함께 보인다.

- [ ] **Step 3: 커밋 + PR**

```bash
git add README.md
git commit -m "docs: README — Fundamentals 커버리지·예시 추가"
git push -u origin feature/fundamentals
gh pr create -R kenshin579/tiingo-go --title "feat(fundamentals): Fundamentals 카테고리 (v0.2.0)" --body "$(cat <<'EOF'
## Summary
- `Fundamentals` 서브클라이언트 추가 — `Definitions` / `Meta` / `Statements` / `Daily` 4개 엔드포인트
- `statementData` 는 응답 그대로 `[]DataPoint` 로 두고 `Get`/`Map` 헬퍼 + `dataCode` 상수 85개(`codes.go`) 제공. Tiingo 가 지표를 계속 추가하므로 타입 필드로 전개하지 않음 — 상수가 없어도 조회는 동작
- `types.Time` 신설 — `statementLastUpdated` 처럼 시각이 의미 있는 필드용(`Date` 는 날짜만 남긴다)
- fixture 4종은 **실호출 응답**. 문서 표에 없는 `dataProviderPermaTicker` 도 실제로 와서 포함
- 설계: `docs/superpowers/specs/2026-09-04-fundamentals-design.md`

## Test plan
- [x] `go test ./... -race` — fundamentals 파싱·쿼리 생성·`Get`/`Map`·상수 일치, `types.Time` 파싱·직렬화
- [x] `go vet ./...`, `go vet -tags integration ./...`, `gofmt` clean
- [x] `TIINGO_API_KEY=... go test -tags integration ./...` — 통합 8건(EOD 4 + Fundamentals 4) PASS
- [x] `go run ./examples/fundamentals` 동작 확인

## Note
Fundamentals 는 별도 구독이다. 무료 플랜은 Dow 30 3년치만 조회되며 권한 밖 종목은 `APIError`(400/403). 테스트·예제는 AAPL 만 쓴다.

머지 후 `./scripts/release.sh v0.2.0`.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

리뷰어는 지정하지 않는다(사용자가 직접 지정).

---

## 자체 검토 (스펙 대조)

| 스펙 항목 | 태스크 |
|---|---|
| 결정 1 (원문 + 헬퍼 + 상수) | Task 5(`Get`/`Map`), Task 4(`codes.go`) |
| 결정 2 (메서드 4개, Meta 가변 인자) | Task 3~6 |
| 결정 3 (`types.Time` 신설) | Task 1 |
| 결정 4 (주석 규약) | 전 태스크의 구조체 |
| 결정 5 (`dataProviderPermaTicker` 포함) | Task 3 |
| 결정 6 (무료 플랜은 문서로) | Task 3(패키지 doc), Task 8(README 각주) |
| 데이터 모델 6종 | Task 3(Meta), 4(Definition), 5(Statement·StatementData·DataPoint), 6(DailyMetric) |
| 옵션 zero 값 생략, `asReported` 는 true 일 때만 | Task 5, 6 의 `params()` + 전용 테스트 |
| fixture 실호출 | Task 2 |
| codes 일치 테스트 | Task 4(단위), Task 7(통합 — 실 API 와 대조) |
| integration 4건 | Task 7 |
| README·v0.2.0 | Task 8 |

빠진 스펙 항목 없음. 이름 일관성: `fundamentals.New`/`Client`/`Definition`/`Definitions`/`Meta`/`Statement`/`StatementData`/`DataPoint`/`StatementOptions`/`DailyMetric`/`DailyOptions`/`AllCodes`/`Code*`, `types.Time`/`types.Date`/`types.DateLayout` — 정의 태스크와 사용 태스크의 표기가 일치한다.
