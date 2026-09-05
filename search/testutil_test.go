package search

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
