# Tiingo Go SDK 기반 + End-of-Day 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `github.com/kenshin579/tiingo-go` v0.1.0 — Tiingo API 의 확장 가능한 Go 클라이언트 기반과 첫 카테고리(End-of-Day: 메타 + 일별 시세)를 만든다.

**Architecture:** 단일 진입점 `tiingo.Client` 가 `internal/httpclient`(인증 헤더·GET·에러 매핑)를 쥐고 카테고리 서브클라이언트(`EOD *eod.Client`)를 필드로 노출한다. 순수 파싱 타입(`tiingo.Date`)과 도메인 패키지(`eod`)를 분리해 각각 독립 테스트한다. fmp-go 와 같은 구조라 이후 카테고리는 패키지만 추가하면 된다.

**Tech Stack:** Go 1.25, 표준 라이브러리 + `github.com/stretchr/testify` (테스트 전용), `httptest` 스텁, build tag `integration` 으로 분리된 실호출 테스트.

**Spec:** `docs/superpowers/specs/2026-09-04-sdk-foundation-eod-design.md`
**API 참조:** `docs/api/rest/end-of-day.md`, `docs/api/general/{overview,connecting}.md`

---

## 파일 구조

| 파일 | 책임 |
|---|---|
| `go.mod` | 모듈 선언, Go 1.25, testify |
| `internal/httpclient/client.go` | HTTP 계층 하나: `Authorization: Token` 주입, GET, JSON 디코딩, `APIError`/`ErrNotFound` |
| `internal/httpclient/client_test.go` | httptest 스텁 단위 테스트 |
| `types.go` | `Date` — Tiingo 의 두 날짜 형식 파싱·직렬화 (순수, 의존성 없음) |
| `types_test.go` | `Date` 단위 테스트 |
| `errors.go` | `APIError`, `ErrNotFound` 를 루트 패키지로 재노출 |
| `config.go` | functional option 3종 |
| `client.go` | `tiingo.Client` 조립, `NewClient` |
| `from_env.go` | `NewClientFromEnv` — `TIINGO_API_KEY` |
| `eod/client.go` | `eod.Client` 선언과 생성자 |
| `eod/meta.go` | `Meta` 타입 + `Meta()` |
| `eod/prices.go` | `Price`, `ResampleFreq`, `PriceOptions`, `HistoricalPrices()`, `LatestPrice()` |
| `eod/meta_test.go`, `eod/prices_test.go` | fixture 기반 파싱·쿼리 테스트 |
| `eod/testdata/*.json` | 실호출 fixture |
| `examples/eod/main.go` | 실행 가능한 사용 예시 |
| `integration_test.go` | build tag `integration`, 실호출 계약 검증 |
| `scripts/release.sh` | 릴리스 자동화 (fmp-go 와 동일 절차) |
| `README.md` | 설치·사용·커버리지·인증 안내로 교체 |

**주석 규약(스펙 결정 6)**: 짧은 설명은 필드 오른쪽 줄 끝 주석, 한 줄(약 120자)에 안 들어가면 필드 위 주석. 타입·메서드에는 doc 주석. 모든 주석은 한국어.

---

### Task 1: 모듈 초기화

**Files:**
- Create: `go.mod`

- [ ] **Step 1: 브랜치 확인**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && git branch --show-current`
Expected: `feature/sdk-foundation`

- [ ] **Step 2: 모듈 생성 + testify 추가**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go mod init github.com/kenshin579/tiingo-go
go get github.com/stretchr/testify@latest
go mod tidy
```
Expected: `go.mod` 에 `module github.com/kenshin579/tiingo-go`, `go 1.25`(또는 설치된 toolchain 버전), `require github.com/stretchr/testify v1.x`. `go.sum` 생성.

- [ ] **Step 3: 빌드 확인**

Run: `go build ./... && go vet ./...`
Expected: 출력 없음(패키지가 아직 없어도 성공).

- [ ] **Step 4: 커밋**

```bash
git add go.mod go.sum
git commit -m "chore: Go 모듈 초기화 (tiingo-go, testify)"
```

---

### Task 2: `tiingo.Date` — 두 날짜 형식 파싱

**Files:**
- Create: `types.go`, `types_test.go`

Tiingo 는 같은 API 안에서 날짜 형식이 둘이다. 가격 응답은 `"2019-01-02T00:00:00.000Z"`, 메타 응답은 `"1980-12-12"`. 요청 쿼리로 나갈 때는 `YYYY-MM-DD`.

- [ ] **Step 1: 실패하는 테스트 작성**

`types_test.go`:

```go
package tiingo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"RFC3339 밀리초 UTC", `"2019-01-02T00:00:00.000Z"`, time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"RFC3339 오프셋", `"2019-01-02T00:00:00+00:00"`, time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"날짜만", `"1980-12-12"`, time.Date(1980, 12, 12, 0, 0, 0, 0, time.UTC)},
		{"null", `null`, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			require.NoError(t, json.Unmarshal([]byte(tt.in), &d))
			assert.True(t, d.Time.Equal(tt.want), "got %v, want %v", d.Time, tt.want)
		})
	}
}

func TestDate_UnmarshalJSON_Invalid(t *testing.T) {
	var d Date
	assert.Error(t, json.Unmarshal([]byte(`"2019/01/02"`), &d))
	assert.Error(t, json.Unmarshal([]byte(`12345`), &d))
}

func TestDate_MarshalJSON(t *testing.T) {
	d := Date{Time: time.Date(2019, 1, 2, 15, 4, 5, 0, time.UTC)}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, `"2019-01-02"`, string(b))

	b, err = json.Marshal(Date{})
	require.NoError(t, err)
	assert.Equal(t, `null`, string(b))
}

func TestDate_String(t *testing.T) {
	assert.Equal(t, "2019-01-02", Date{Time: time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)}.String())
	assert.Equal(t, "", Date{}.String())
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./ -run TestDate`
Expected: FAIL — `undefined: Date`

- [ ] **Step 3: `types.go` 구현**

```go
// Package tiingo 는 Tiingo API 의 Go 클라이언트다.
package tiingo

import (
	"encoding/json"
	"fmt"
	"time"
)

// DateLayout 은 Tiingo 요청 쿼리와 Date 직렬화에 쓰는 날짜 형식.
const DateLayout = "2006-01-02"

// Date 는 Tiingo 의 날짜 값이다. 응답에서 두 형식(RFC3339 타임스탬프, YYYY-MM-DD)이 모두
// 오기 때문에 둘 다 받아 time.Time 으로 정규화하고, 직렬화할 때는 YYYY-MM-DD 로 쓴다.
// null 이나 빈 문자열은 zero value 가 되며 IsZero() 로 구분한다.
type Date struct {
	time.Time
}

// UnmarshalJSON 은 RFC3339 타임스탬프와 YYYY-MM-DD 를 모두 받는다.
func (d *Date) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		d.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("tiingo: date must be a JSON string: %w", err)
	}
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.Time = t.UTC()
		return nil
	}
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return fmt.Errorf("tiingo: unrecognized date %q", s)
	}
	d.Time = t
	return nil
}

// MarshalJSON 은 YYYY-MM-DD 문자열로 쓴다. zero value 는 null.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(DateLayout))
}

// String 은 YYYY-MM-DD 를 반환한다. zero value 는 빈 문자열.
func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format(DateLayout)
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./ -run TestDate -v`
Expected: 4개 테스트 함수 모두 PASS.

- [ ] **Step 5: 커밋**

```bash
git add types.go types_test.go
git commit -m "feat: tiingo.Date — 응답의 두 날짜 형식 파싱, YYYY-MM-DD 직렬화"
```

---

### Task 3: `internal/httpclient` — 인증·GET·에러 매핑

**Files:**
- Create: `internal/httpclient/client.go`, `internal/httpclient/client_test.go`

- [ ] **Step 1: 실패하는 테스트 작성**

`internal/httpclient/client_test.go`:

```go
package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJSON_인증헤더와_쿼리(t *testing.T) {
	var gotAuth, gotAccept, gotQuery, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotQuery = r.URL.RawQuery
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ticker":"AAPL"}`))
	}))
	defer srv.Close()

	c := New(Config{APIKey: "secret", BaseURL: srv.URL})
	var out struct {
		Ticker string `json:"ticker"`
	}
	require.NoError(t, c.GetJSON(context.Background(), "/tiingo/daily/aapl", map[string]string{"startDate": "2019-01-02"}, &out))

	assert.Equal(t, "Token secret", gotAuth)
	assert.Equal(t, "application/json", gotAccept)
	assert.Equal(t, "/tiingo/daily/aapl", gotPath)
	assert.Equal(t, "startDate=2019-01-02", gotQuery, "토큰은 쿼리에 실리면 안 된다")
	assert.Equal(t, "AAPL", out.Ticker)
}

func TestGetJSON_에러응답_매핑(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"detail 키", http.StatusNotFound, `{"detail":"Not found."}`, "Not found."},
		{"error 키", http.StatusForbidden, `{"error":"You do not have permission"}`, "You do not have permission"},
		{"message 키", http.StatusUnauthorized, `{"message":"Invalid token"}`, "Invalid token"},
		{"JSON 아님", http.StatusInternalServerError, `boom`, "500 Internal Server Error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := New(Config{APIKey: "k", BaseURL: srv.URL})
			var out map[string]any
			err := c.GetJSON(context.Background(), "/x", nil, &out)

			var apiErr *APIError
			require.True(t, errors.As(err, &apiErr), "want *APIError, got %T", err)
			assert.Equal(t, tt.status, apiErr.StatusCode)
			assert.Contains(t, apiErr.Message, tt.want)
		})
	}
}

func TestGetJSON_디코딩실패(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	c := New(Config{APIKey: "k", BaseURL: srv.URL})
	var out map[string]any
	assert.Error(t, c.GetJSON(context.Background(), "/x", nil, &out))
}

func TestGetJSON_컨텍스트_취소(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	c := New(Config{APIKey: "k", BaseURL: srv.URL, Timeout: 20 * time.Millisecond})
	var out map[string]any
	assert.Error(t, c.GetJSON(context.Background(), "/x", nil, &out))
}

func TestGetRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("date,close\n2019-01-02,157.92\n"))
	}))
	defer srv.Close()

	c := New(Config{APIKey: "k", BaseURL: srv.URL})
	b, err := c.GetRaw(context.Background(), "/x", map[string]string{"format": "csv"})
	require.NoError(t, err)
	assert.Contains(t, string(b), "2019-01-02,157.92")
}

func TestNew_기본값(t *testing.T) {
	c := New(Config{APIKey: "k"})
	assert.Equal(t, DefaultBaseURL, c.baseURL)
	assert.Equal(t, 30*time.Second, c.http.Timeout)
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./internal/httpclient/`
Expected: FAIL — `undefined: New` 등 (패키지 미존재).

- [ ] **Step 3: `internal/httpclient/client.go` 구현**

```go
// Package httpclient 는 Tiingo REST 호출의 단일 GET 통로다.
// Authorization 헤더 주입, JSON 디코딩, 에러 매핑(HTTP 상태 + Tiingo 에러 바디)을 담당한다.
package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultBaseURL 은 Tiingo REST API 베이스 URL.
const DefaultBaseURL = "https://api.tiingo.com"

// DefaultTimeout 은 HTTP 클라이언트 기본 타임아웃.
const DefaultTimeout = 30 * time.Second

// ErrNotFound 는 조회 결과가 없을 때(빈 배열 등) 서비스 계층이 반환하는 sentinel.
var ErrNotFound = errors.New("tiingo: not found")

// APIError 는 Tiingo 에러 응답이다. errors.As 로 StatusCode/Message 에 접근한다.
// 401 은 토큰 오류, 403 은 권한·플랜 부족, 404 는 없는 티커, 429 는 rate limit 이다.
type APIError struct {
	StatusCode int    // HTTP 상태코드
	Message    string // Tiingo 에러 메시지 또는 상태 텍스트
}

func (e *APIError) Error() string {
	return fmt.Sprintf("tiingo: api error (status %d): %s", e.StatusCode, e.Message)
}

// Config 는 Client 생성 인자.
type Config struct {
	APIKey     string
	BaseURL    string        // 빈 값이면 DefaultBaseURL
	Timeout    time.Duration // 0이면 DefaultTimeout
	HTTPClient *http.Client  // nil이면 Timeout 적용 기본 클라이언트
}

// Client 는 Tiingo HTTP 계층.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// New 는 Config 로 Client 를 만든다.
func New(cfg Config) *Client {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}
		hc = &http.Client{Timeout: timeout}
	}
	return &Client{apiKey: cfg.APIKey, baseURL: base, http: hc}
}

// errorEnvelope 는 Tiingo 에러 바디 형태. 엔드포인트마다 키가 달라 셋 다 받는다.
type errorEnvelope struct {
	Detail  string `json:"detail"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (e errorEnvelope) message() string {
	switch {
	case e.Detail != "":
		return e.Detail
	case e.Error != "":
		return e.Error
	default:
		return e.Message
	}
}

// get 은 인증 헤더를 붙여 GET 후 응답 바디를 반환한다.
// 비-200 은 *APIError 로 매핑한다.
func (c *Client) get(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("tiingo: bad url %s: %w", path, err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	// 토큰은 쿼리(?token=)가 아니라 헤더로 보낸다 — URL·서버 로그에 키가 남지 않도록.
	req.Header.Set("Authorization", "Token "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tiingo: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tiingo: read %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		msg := resp.Status
		var env errorEnvelope
		if json.Unmarshal(body, &env) == nil && env.message() != "" {
			msg = env.message()
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}
	return body, nil
}

// GetJSON 은 GET 후 응답을 out 으로 디코딩한다. 비-200 은 *APIError.
func (c *Client) GetJSON(ctx context.Context, path string, params map[string]string, out any) error {
	body, err := c.get(ctx, path, params)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("tiingo: decode %s: %w", path, err)
	}
	return nil
}

// GetRaw 는 GET 후 응답 바디를 원시 바이트로 반환한다(CSV 등 비-JSON 용).
func (c *Client) GetRaw(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	return c.get(ctx, path, params)
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./internal/httpclient/ -v`
Expected: 6개 테스트 함수(하위 케이스 포함) 모두 PASS.

- [ ] **Step 5: 커밋**

```bash
git add internal/httpclient/
git commit -m "feat(httpclient): Authorization 헤더 인증, GET+JSON, 에러 매핑"
```

---

### Task 4: 루트 클라이언트 조립

**Files:**
- Create: `errors.go`, `config.go`, `client.go`, `from_env.go`, `client_test.go`

- [ ] **Step 1: 실패하는 테스트 작성**

`client_test.go`:

```go
package tiingo

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient("key")
	require.NoError(t, err)
	assert.NotNil(t, c.EOD)

	_, err = NewClient("")
	assert.Error(t, err, "빈 키는 에러")
}

func TestNewClient_Options(t *testing.T) {
	c, err := NewClient("key", WithBaseURL("http://127.0.0.1:1"), WithTimeout(5*time.Second))
	require.NoError(t, err)
	assert.NotNil(t, c.EOD)
}

func TestNewClientFromEnv(t *testing.T) {
	t.Setenv("TIINGO_API_KEY", "envkey")
	c, err := NewClientFromEnv()
	require.NoError(t, err)
	assert.NotNil(t, c.EOD)

	require.NoError(t, os.Unsetenv("TIINGO_API_KEY"))
	_, err = NewClientFromEnv()
	assert.Error(t, err, "환경변수 없으면 에러")
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./ -run TestNewClient`
Expected: FAIL — `undefined: NewClient`

- [ ] **Step 3: 네 파일 구현**

`errors.go`:

```go
package tiingo

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// APIError 는 Tiingo 에러 응답이다. errors.As 로 StatusCode/Message 에 접근한다.
type APIError = httpclient.APIError

// ErrNotFound 는 조회 결과가 없을 때 서비스 계층이 반환한다.
var ErrNotFound = httpclient.ErrNotFound
```

`config.go`:

```go
package tiingo

import (
	"net/http"
	"time"
)

type clientOptions struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// Option 은 Client 생성 옵션(functional option).
type Option func(*clientOptions)

// WithBaseURL 은 API 베이스 URL 을 바꾼다(테스트/프록시용).
func WithBaseURL(u string) Option { return func(o *clientOptions) { o.baseURL = u } }

// WithTimeout 은 HTTP 타임아웃을 지정한다(기본 30s).
func WithTimeout(d time.Duration) Option { return func(o *clientOptions) { o.timeout = d } }

// WithHTTPClient 는 사용자 정의 *http.Client 를 주입한다(설정 시 WithTimeout 은 무시된다).
func WithHTTPClient(c *http.Client) Option { return func(o *clientOptions) { o.httpClient = c } }
```

`client.go`:

```go
package tiingo

import (
	"errors"

	"github.com/kenshin579/tiingo-go/eod"
	"github.com/kenshin579/tiingo-go/internal/httpclient"
)

// Client 는 tiingo-go 라이브러리의 단일 진입점이다.
// 카테고리별 서브클라이언트를 필드로 노출한다.
type Client struct {
	http *httpclient.Client

	EOD *eod.Client // 일별 시세·메타(End-of-Day)
}

// NewClient 는 API 키로 Client 를 만든다. 키가 비어 있으면 에러다.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("tiingo: apiKey is required")
	}
	var cfg clientOptions
	for _, opt := range opts {
		opt(&cfg)
	}
	hc := httpclient.New(httpclient.Config{
		APIKey:     apiKey,
		BaseURL:    cfg.baseURL,
		Timeout:    cfg.timeout,
		HTTPClient: cfg.httpClient,
	})
	c := &Client{http: hc}
	c.EOD = eod.New(hc)
	return c, nil
}
```

`from_env.go`:

```go
package tiingo

import (
	"errors"
	"os"
)

// APIKeyEnv 는 API 키를 읽는 환경변수 이름.
const APIKeyEnv = "TIINGO_API_KEY"

// NewClientFromEnv 는 TIINGO_API_KEY 환경변수로 Client 를 만든다.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	key := os.Getenv(APIKeyEnv)
	if key == "" {
		return nil, errors.New("tiingo: " + APIKeyEnv + " is not set")
	}
	return NewClient(key, opts...)
}
```

Task 5 의 `eod` 패키지가 아직 없으므로 이 단계에서는 컴파일되지 않는다. **Task 5 를 먼저 만들 수는 없다**(`eod.New` 시그니처가 여기서 정해짐). 따라서 이 태스크에서 `eod/client.go` 의 최소 골격도 함께 만든다:

`eod/client.go`:

```go
// Package eod 는 Tiingo End-of-Day(일별 시세·메타) API sub-client 다.
// tiingo.Client.EOD 로 접근한다.
package eod

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 End-of-Day sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go build ./... && go test ./ -run TestNewClient -v`
Expected: 3개 테스트 PASS.

- [ ] **Step 5: 커밋**

```bash
git add errors.go config.go client.go from_env.go client_test.go eod/client.go
git commit -m "feat: tiingo.Client 진입점, 옵션, 환경변수 생성자"
```

---

### Task 5: EOD 메타 엔드포인트

**Files:**
- Create: `eod/meta.go`, `eod/meta_test.go`, `eod/testdata/meta_aapl.json`
- Create: `eod/testutil_test.go` (테스트용 스텁 클라이언트 헬퍼)

참조: `docs/api/rest/end-of-day.md` 의 `2.1.3 Meta Endpoint`. 응답은 **단일 객체**.

- [ ] **Step 1: fixture 저장**

실제 API 로 받아 저장한다(키가 없으면 아래 문서 예시로 대체 — 단, 그 경우 Step 5 에서 실호출 확인 필요).

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
mkdir -p eod/testdata
curl -fsS -H "Authorization: Token $TIINGO_API_KEY" \
  "https://api.tiingo.com/tiingo/daily/aapl" | python3 -m json.tool > eod/testdata/meta_aapl.json
cat eod/testdata/meta_aapl.json
```
Expected: `ticker`, `name`, `exchangeCode`, `description`, `startDate`, `endDate` 를 포함한 객체.
**중요**: 응답에 `permaTicker`, `openFigiComposite`, `assetType`, `isActive` 등 문서 표에 없는 필드가 있으면 그대로 두고 Step 3 의 구조체에 필드를 **추가**한다(스펙 «데이터 모델» 항목). 실제 키 이름을 보고 정확히 반영할 것.

`TIINGO_API_KEY` 가 없으면 다음 내용으로 fixture 를 만들고(문서 예시), Step 3 의 구조체는 6필드만 둔다:

```json
{
    "ticker": "AAPL",
    "name": "Apple Inc",
    "exchangeCode": "NASDAQ",
    "startDate": "1980-12-12",
    "endDate": "2019-01-25",
    "description": "Apple Inc. (Apple) designs, manufactures and markets mobile communication and media devices."
}
```

- [ ] **Step 2: 테스트 헬퍼 + 실패하는 테스트 작성**

`eod/testutil_test.go`:

```go
package eod

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
)

// newStubClient 는 고정 응답을 돌려주는 테스트 서버에 붙은 Client 와,
// 서버가 받은 마지막 요청 URL 을 읽는 함수를 돌려준다.
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

`eod/meta_test.go`:

```go
package eod

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
	m, err := c.Meta(context.Background(), "aapl")
	require.NoError(t, err)

	assert.Equal(t, "AAPL", m.Ticker)
	assert.Equal(t, "Apple Inc", m.Name)
	assert.Equal(t, "NASDAQ", m.ExchangeCode)
	assert.NotEmpty(t, m.Description)
	require.NotNil(t, m.StartDate)
	assert.Equal(t, "1980-12-12", m.StartDate.String())
	require.NotNil(t, m.EndDate)
	assert.False(t, m.EndDate.IsZero())

	assert.Equal(t, "/tiingo/daily/aapl", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestMeta_NullDates(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `{"ticker":"XYZ","name":"No Data","exchangeCode":"NYSE","description":"","startDate":null,"endDate":null}`)
	m, err := c.Meta(context.Background(), "xyz")
	require.NoError(t, err)
	assert.Nil(t, m.StartDate, "startDate null 이면 nil")
	assert.Nil(t, m.EndDate)
}

func TestMeta_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `{}`)
	_, err := c.Meta(context.Background(), "  ")
	assert.Error(t, err)
}

func TestMeta_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Meta(context.Background(), "nosuch")
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
```

- [ ] **Step 3: 테스트 실패 확인 후 `eod/meta.go` 구현**

Run: `go test ./eod/ -run TestMeta` → FAIL (`c.Meta undefined`)

```go
package eod

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	tiingo "github.com/kenshin579/tiingo-go"
)

// Meta 는 자산의 메타 정보다(EOD 메타 엔드포인트 응답).
type Meta struct {
	Ticker       string       `json:"ticker"`       // 자산의 티커. 주식 클래스는 마침표 대신 대시(예: BRK-A)
	Name         string       `json:"name"`         // 자산의 전체 이름
	ExchangeCode string       `json:"exchangeCode"` // 상장 거래소 식별자(예: NASDAQ)
	Description  string       `json:"description"`  // 자산에 대한 장문 설명
	StartDate    *tiingo.Date `json:"startDate"`    // 가격 데이터가 있는 가장 이른 날짜. nil 이면 가격 데이터 없음
	EndDate      *tiingo.Date `json:"endDate"`      // 가격 데이터가 있는 가장 늦은 날짜. nil 이면 가격 데이터 없음
}

// Meta 는 자산의 메타 정보를 조회한다. GET /tiingo/daily/<ticker>
func (c *Client) Meta(ctx context.Context, ticker string) (*Meta, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var m Meta
	if err := c.http.GetJSON(ctx, "/tiingo/daily/"+url.PathEscape(t), nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
```

**Step 1 에서 fixture 에 추가 필드가 보였다면** 여기에 같은 규약으로 추가한다. 예:

```go
	PermaTicker string `json:"permaTicker"` // 티커 변경·재사용에도 안정적인 영구 식별자
	IsActive    bool   `json:"isActive"`    // 현재 거래 활성 여부
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./eod/ -run TestMeta -v`
Expected: 4개 테스트 PASS. (fixture 에 필드를 추가했다면 `TestMeta_Fixture` 에 해당 필드 assertion 도 한 줄 추가한다.)

- [ ] **Step 5: 커밋**

```bash
git add eod/meta.go eod/meta_test.go eod/testutil_test.go eod/testdata/meta_aapl.json
git commit -m "feat(eod): 메타 엔드포인트 (Meta)"
```

---

### Task 6: EOD 가격 엔드포인트

**Files:**
- Create: `eod/prices.go`, `eod/prices_test.go`, `eod/testdata/prices_aapl.json`

참조: `docs/api/rest/end-of-day.md` 의 `2.1.2 End-of-Day Endpoint`. 응답은 **배열**.

- [ ] **Step 1: fixture 저장**

```bash
curl -fsS -H "Authorization: Token $TIINGO_API_KEY" \
  "https://api.tiingo.com/tiingo/daily/aapl/prices?startDate=2019-01-02&endDate=2019-01-07" \
  | python3 -m json.tool > eod/testdata/prices_aapl.json
head -20 eod/testdata/prices_aapl.json
```
키가 없으면 문서 예시(`docs/api/rest/end-of-day.md` 의 `### Example` JSON 4건)를 그대로 저장한다.
Expected: `date`, `open`, `high`, `low`, `close`, `volume`, `adj*`, `divCash`, `splitFactor` 를 가진 객체 배열.

- [ ] **Step 2: 실패하는 테스트 작성**

`eod/prices_test.go`:

```go
package eod

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

func TestHistoricalPrices_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.HistoricalPrices(context.Background(), "aapl", &PriceOptions{
		StartDate:    time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2019, 1, 7, 0, 0, 0, 0, time.UTC),
		ResampleFreq: ResampleDaily,
		Sort:         "-date",
		Columns:      []string{"date", "close"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)

	p := ps[0]
	assert.Equal(t, "2019-01-02", p.Date.String())
	assert.InDelta(t, 157.92, p.Close, 0.001)
	assert.InDelta(t, 154.89, p.Open, 0.001)
	assert.Equal(t, int64(37039737), p.Volume)
	assert.InDelta(t, 1.0, p.SplitFactor, 0.001)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/daily/aapl/prices", lastReq().URL.Path)
	assert.Equal(t, "2019-01-02", q.Get("startDate"))
	assert.Equal(t, "2019-01-07", q.Get("endDate"))
	assert.Equal(t, "daily", q.Get("resampleFreq"))
	assert.Equal(t, "-date", q.Get("sort"))
	assert.Equal(t, "date,close", q.Get("columns"))
}

func TestHistoricalPrices_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.HistoricalPrices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "옵션이 없으면 쿼리도 없어야 한다")
}

func TestHistoricalPrices_ZeroFieldsOmitted(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.HistoricalPrices(context.Background(), "aapl", &PriceOptions{Sort: "date"})
	require.NoError(t, err)
	q := lastReq().URL.Query()
	assert.Equal(t, "date", q.Get("sort"))
	assert.Empty(t, q.Get("startDate"))
	assert.Empty(t, q.Get("endDate"))
	assert.Empty(t, q.Get("resampleFreq"))
	assert.Empty(t, q.Get("columns"))
}

func TestHistoricalPrices_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.HistoricalPrices(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestLatestPrice(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[{"date":"2019-01-02T00:00:00.000Z","close":157.92,"volume":37039737,"splitFactor":1.0}]`)
	p, err := c.LatestPrice(context.Background(), "aapl")
	require.NoError(t, err)
	assert.Equal(t, "2019-01-02", p.Date.String())
	assert.InDelta(t, 157.92, p.Close, 0.001)
}

func TestLatestPrice_EmptyArray_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.LatestPrice(context.Background(), "aapl")
	assert.True(t, errors.Is(err, httpclient.ErrNotFound), "빈 배열은 ErrNotFound")
}
```

- [ ] **Step 3: 테스트 실패 확인 후 `eod/prices.go` 구현**

Run: `go test ./eod/ -run "TestHistorical|TestLatest"` → FAIL (`undefined: PriceOptions`)

```go
package eod

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/internal/httpclient"
)

// Price 는 일별 시세 한 건이다. 뮤추얼 펀드는 OHLC 에 그날의 NAV 가 들어간다.
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
	// SplitFactor 는 분할·역분할·분배 시 가격 조정에 쓰이는 계수다. 1 이 아니면 그날 이벤트가
	// 있었다는 뜻이므로, 증분 동기화 중이라면 해당 종목의 전체 이력을 다시 받아야 한다.
	SplitFactor float64 `json:"splitFactor"`
}

// ResampleFreq 는 가격 리샘플링 주기다.
// 주의: daily 만 휴일 달력을 반영하고 나머지는 표준 영업일(월~금) 기준이다.
// 기간 중간 날짜를 주면 전체 기간이 잡히도록 startDate 는 뒤로, endDate 는 앞으로 조정된다
// (예: weekly 에 수요일 startDate → 월요일로 롤백, 수요일 endDate → 금요일로 롤포워드).
type ResampleFreq string

const (
	ResampleDaily    ResampleFreq = "daily"    // 일별. 휴일 달력 반영
	ResampleWeekly   ResampleFreq = "weekly"   // 주별. 금요일 마감
	ResampleMonthly  ResampleFreq = "monthly"  // 월별. 각 월 마지막 영업일 마감
	ResampleAnnually ResampleFreq = "annually" // 연별. 각 해 마지막 영업일 마감
)

// PriceOptions 는 HistoricalPrices 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작일(이상, >=). zero 면 미전송
	EndDate      time.Time    // 조회 종료일(이하, <=). zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플링 주기. 빈 값이면 미전송(일별)
	Sort         string       // 정렬 컬럼. "date" 오름차순, "-date" 내림차순
	Columns      []string     // 돌려받을 컬럼만 지정. 비어 있으면 전체
}

// params 는 옵션을 쿼리 맵으로 바꾼다. zero 값은 넣지 않는다.
func (o *PriceOptions) params() map[string]string {
	if o == nil {
		return nil
	}
	q := map[string]string{}
	if !o.StartDate.IsZero() {
		q["startDate"] = o.StartDate.Format(tiingo.DateLayout)
	}
	if !o.EndDate.IsZero() {
		q["endDate"] = o.EndDate.Format(tiingo.DateLayout)
	}
	if o.ResampleFreq != "" {
		q["resampleFreq"] = string(o.ResampleFreq)
	}
	if o.Sort != "" {
		q["sort"] = o.Sort
	}
	if len(o.Columns) > 0 {
		q["columns"] = strings.Join(o.Columns, ",")
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// HistoricalPrices 는 일별 시세를 조회한다. GET /tiingo/daily/<ticker>/prices
// opts 가 nil 이거나 날짜가 없으면 Tiingo 는 최신 1건만 반환한다.
func (c *Client) HistoricalPrices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ps []Price
	path := "/tiingo/daily/" + url.PathEscape(t) + "/prices"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// LatestPrice 는 가장 최근 일별 시세 1건을 조회한다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) LatestPrice(ctx context.Context, ticker string) (*Price, error) {
	ps, err := c.HistoricalPrices(ctx, ticker, nil)
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, httpclient.ErrNotFound
	}
	return &ps[0], nil
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./... -v 2>&1 | tail -30`
Expected: 모든 패키지 PASS. `go vet ./...` 도 통과.

- [ ] **Step 5: 커밋**

```bash
git add eod/prices.go eod/prices_test.go eod/testdata/prices_aapl.json
git commit -m "feat(eod): 가격 엔드포인트 (HistoricalPrices, LatestPrice)"
```

---

### Task 7: 예제 + 통합 테스트

**Files:**
- Create: `examples/eod/main.go`, `integration_test.go`

- [ ] **Step 1: `examples/eod/main.go` 작성**

```go
// Tiingo End-of-Day 예제.
//
//	TIINGO_API_KEY=... go run ./examples/eod
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/eod"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	m, err := c.EOD.Meta(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (%s) — %s ~ %s\n", m.Name, m.ExchangeCode, m.StartDate, m.EndDate)

	p, err := c.EOD.LatestPrice(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("최신 종가 %s: %.2f (거래량 %d)\n", p.Date, p.Close, p.Volume)

	ps, err := c.EOD.HistoricalPrices(ctx, "AAPL", &eod.PriceOptions{
		StartDate:    time.Now().AddDate(0, -1, 0),
		ResampleFreq: eod.ResampleWeekly,
		Sort:         "-date",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("최근 1개월 주별 %d건\n", len(ps))
	for _, x := range ps {
		fmt.Printf("  %s  종가 %.2f  수정종가 %.2f\n", x.Date, x.Close, x.AdjClose)
	}
}
```

- [ ] **Step 2: `integration_test.go` 작성**

```go
//go:build integration

package tiingo_test

import (
	"context"
	"os"
	"testing"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/eod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 실행: TIINGO_API_KEY=... go test -tags integration ./...
func newClient(t *testing.T) *tiingo.Client {
	t.Helper()
	if os.Getenv(tiingo.APIKeyEnv) == "" {
		t.Skip(tiingo.APIKeyEnv + " not set")
	}
	c, err := tiingo.NewClientFromEnv()
	require.NoError(t, err)
	return c
}

func TestIntegration_Meta(t *testing.T) {
	c := newClient(t)
	m, err := c.EOD.Meta(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", m.Ticker)
	assert.NotEmpty(t, m.Name)
	assert.Equal(t, "NASDAQ", m.ExchangeCode)
	require.NotNil(t, m.StartDate)
	assert.Equal(t, "1980-12-12", m.StartDate.String())
}

func TestIntegration_LatestPrice(t *testing.T) {
	c := newClient(t)
	p, err := c.EOD.LatestPrice(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.Greater(t, p.AdjClose, 0.0)
}

func TestIntegration_HistoricalPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.EOD.HistoricalPrices(context.Background(), "AAPL", &eod.PriceOptions{
		StartDate:    time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2019, 1, 7, 0, 0, 0, 0, time.UTC),
		ResampleFreq: eod.ResampleDaily,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)
	assert.Equal(t, "2019-01-02", ps[0].Date.String())
	assert.InDelta(t, 157.92, ps[0].Close, 0.01)
}

func TestIntegration_UnknownTicker(t *testing.T) {
	c := newClient(t)
	_, err := c.EOD.Meta(context.Background(), "NOSUCHTICKERXYZ")
	assert.Error(t, err)
}
```

- [ ] **Step 3: 컴파일 + 실행 확인**

```bash
go vet ./...
go vet -tags integration ./...
go build ./examples/...
go test ./...                                   # integration 제외, 전부 PASS
TIINGO_API_KEY=$TIINGO_API_KEY go test -tags integration ./... -run TestIntegration -v
TIINGO_API_KEY=$TIINGO_API_KEY go run ./examples/eod
```
Expected: 키가 있으면 integration 4개 PASS 와 예제 출력(회사명·최신 종가·주별 목록). 키가 없으면 integration 은 SKIP 되고 나머지는 PASS — 그 경우 보고에 명시한다.

- [ ] **Step 4: 커밋**

```bash
git add examples/eod/main.go integration_test.go
git commit -m "feat: EOD 예제 + integration 테스트"
```

---

### Task 8: README + 릴리스 스크립트

**Files:**
- Modify: `README.md`
- Create: `scripts/release.sh`

- [ ] **Step 1: `README.md` 교체**

```markdown
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
```

실행 가능한 예시: [`examples/eod`](examples/eod).

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

나머지 REST 그룹(News, Fundamentals, Crypto, Forex, IEX, BOATS 등)과 WebSocket 은 순차 추가 예정.

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
go test ./...                                   # 단위 테스트
TIINGO_API_KEY=... go test -tags integration ./...  # 실호출 통합 테스트
```

## 문서

- [`docs/api/README.md`](docs/api/README.md) — Tiingo 문서 사이트 23페이지를 변환한 md +
  Tiingo 공식 `llms.txt`/`llms-full.txt` 원본. 재생성은 `./scripts/fetch-docs.sh` 와
  `cd tools/gendocs && npm run gen`.
- 설계·계획: [`docs/superpowers/`](docs/superpowers/)
```

- [ ] **Step 2: `scripts/release.sh` 작성**

`/Users/user/src/workspace_moneyflow/fmp-go/scripts/release.sh` 를 복사한 뒤 프로젝트 이름만 바꾼다.

```bash
cp /Users/user/src/workspace_moneyflow/fmp-go/scripts/release.sh scripts/release.sh
sed -i '' 's/fmp-go 릴리스 자동화/tiingo-go 릴리스 자동화/' scripts/release.sh
chmod +x scripts/release.sh
bash -n scripts/release.sh && echo "syntax OK"
grep -n "fmp" scripts/release.sh || echo "no fmp references left"
```
Expected: `syntax OK`. `fmp` 참조가 남아 있으면(경로·모듈명 등) 모두 `tiingo` 로 바꾸고 다시 확인한다. 스크립트는 main 브랜치·클린 트리·origin 동기화를 검증하고 build/vet/test → 모듈 zip 검증 → 태그 → `gh release create --generate-notes` 를 수행한다.

- [ ] **Step 3: 검증**

```bash
go build ./... && go vet ./... && go test ./...
gofmt -l . | grep -v '^docs/' || echo "gofmt clean"
file -I README.md scripts/release.sh
```
Expected: 테스트 전부 PASS, `gofmt clean`, 두 파일 charset=utf-8.

- [ ] **Step 4: 커밋**

```bash
git add README.md scripts/release.sh
git commit -m "docs: README 사용법 교체 + 릴리스 스크립트"
```

---

### Task 9: 최종 검증 + PR

- [ ] **Step 1: 전체 검증**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./...
go vet ./...
go vet -tags integration ./...
go test ./... -count=1
gofmt -l . | grep -v '^docs/' || echo "gofmt clean"
go test ./... -race -count=1
git status --short                              # 클린
git log --oneline main..HEAD | wc -l
```
Expected: 모두 성공, 워킹 트리 클린.

- [ ] **Step 2: 공개 API 확인**

```bash
go doc github.com/kenshin579/tiingo-go
go doc github.com/kenshin579/tiingo-go/eod
```
Expected: `Client`, `NewClient`, `NewClientFromEnv`, `Option` 3종, `Date`, `APIError`, `ErrNotFound`, `APIKeyEnv` 와 `eod` 의 `Client`, `Meta`, `Price`, `PriceOptions`, `ResampleFreq` 상수 4개, 메서드 3개가 주석과 함께 보인다. 주석이 비어 있는 공개 심볼이 있으면 채운다.

- [ ] **Step 3: 푸시 + PR**

```bash
git push -u origin feature/sdk-foundation
gh pr create --title "feat: Tiingo Go SDK 기반 + End-of-Day (v0.1.0)" --body "$(cat <<'EOF'
## Summary
- `tiingo.Client` 단일 진입점 + 카테고리 서브클라이언트 구조(fmp-go 패턴). v1 은 `EOD` 하나
- 인증은 `Authorization: Token <key>` 헤더(`TIINGO_API_KEY`) — 토큰이 URL 에 실리지 않음
- `internal/httpclient`: GET + JSON 디코딩, `APIError`(401/403/404/429 구분) / `ErrNotFound` 매핑
- `tiingo.Date`: 응답의 두 날짜 형식(RFC3339, YYYY-MM-DD)을 모두 파싱, 쿼리는 YYYY-MM-DD
- `eod`: `Meta` / `LatestPrice` / `HistoricalPrices`(startDate·endDate·resampleFreq·sort·columns)
- 설계: `docs/superpowers/specs/2026-09-04-sdk-foundation-eod-design.md`

## Test plan
- [x] `go test ./...` — httpclient(인증 헤더·에러 매핑·타임아웃), Date(두 형식·null·직렬화), eod(fixture 파싱·쿼리 생성·ErrNotFound)
- [x] `go vet ./...`, `go vet -tags integration ./...`, `gofmt` clean, `-race` 통과
- [x] `TIINGO_API_KEY=... go test -tags integration ./...` — AAPL 메타·최신가·기간 조회 실호출
- [x] `go run ./examples/eod` 동작 확인

## Follow-up
- 나머지 REST 카테고리와 WebSocket 은 별도 스펙. CSV·재시도·moneyflow 통합도 후속
- 머지 후 `./scripts/release.sh v0.1.0`

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

리뷰어는 지정하지 않는다(사용자가 직접 지정). PR #1(문서 카탈로그)이 아직 열려 있으면 base 가 `main` 이므로 카탈로그 커밋도 함께 보인다 — PR #1 을 먼저 머지하면 정리된다.

---

## 자체 검토 (스펙 대조)

| 스펙 항목 | 태스크 |
|---|---|
| 결정 1·2 (수작업 확장, 첫 카테고리 EOD) | Task 5, 6 |
| 결정 3 (메서드 3개) | Task 5(`Meta`), Task 6(`LatestPrice`/`HistoricalPrices`) |
| 결정 4 (Authorization 헤더) | Task 3 (구현 + 테스트로 쿼리 미포함 확인) |
| 결정 5 (testify) | Task 1 |
| 결정 6 (필드 주석 규약) | Task 5, 6 의 구조체 + Task 9 Step 2 (`go doc` 확인) |
| 결정 7 (JSON 만, GetRaw 는 열어둠) | Task 3 (`GetRaw` + 테스트) |
| `tiingo.Date` 두 형식 | Task 2 |
| `APIError`/`ErrNotFound` | Task 3(정의), Task 4(재노출), Task 6(빈 배열) |
| `PriceOptions` zero 값 생략 | Task 6 (`TestHistoricalPrices_ZeroFieldsOmitted`) |
| 메타 추가 필드(permaTicker 등) 확인 | Task 5 Step 1·3 |
| integration build tag | Task 7 |
| README·릴리스 스크립트 | Task 8 |
| v0.1.0 릴리스 | Task 9 Step 3 (follow-up 으로 명시) |

빠진 스펙 항목 없음. 태스크 간 이름 일관성 확인: `httpclient.New`/`Config`/`GetJSON`/`GetRaw`/`APIError`/`ErrNotFound`, `eod.New`/`Client`/`Meta`/`Price`/`PriceOptions`/`ResampleFreq`, `tiingo.Date`/`DateLayout`/`APIKeyEnv` — 정의 태스크와 사용 태스크의 표기가 일치한다.
