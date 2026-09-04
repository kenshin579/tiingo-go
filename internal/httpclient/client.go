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
