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
