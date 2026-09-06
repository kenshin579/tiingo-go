package tiingo

import (
	"net/http"
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
	assert.NotNil(t, c.Fundamentals)
	assert.NotNil(t, c.Crypto)
	assert.NotNil(t, c.Forex)
	assert.NotNil(t, c.IEX)
	assert.NotNil(t, c.Search)
	assert.NotNil(t, c.Equity)
	assert.NotNil(t, c.CorporateActions)
	assert.NotNil(t, c.Stream)

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

// 루트의 WithHTTPClient 가 스트림 클라이언트에도 전달된다.
func TestNewClient_StreamInheritsHTTPClient(t *testing.T) {
	hc := &http.Client{}
	c, err := NewClient("key", WithHTTPClient(hc))
	require.NoError(t, err)
	assert.Same(t, hc, c.Stream.HTTPClient(), "같은 *http.Client 여야 한다")
}
