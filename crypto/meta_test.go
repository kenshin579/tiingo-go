package crypto

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
	raw, err := os.ReadFile("testdata/meta_btcusd.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ms, err := c.Meta(context.Background(), "btcusd")
	require.NoError(t, err)
	require.Len(t, ms, 1)

	m := ms[0]
	assert.Equal(t, "btcusd", m.Ticker)
	assert.NotEmpty(t, m.Name)
	assert.Equal(t, "btc", m.BaseCurrency)
	assert.Equal(t, "usd", m.QuoteCurrency)

	assert.Equal(t, "/tiingo/crypto", lastReq().URL.Path)
	assert.Equal(t, "btcusd", lastReq().URL.Query().Get("tickers"))
}

func TestMeta_MultipleTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background(), "btcusd", "ethusd")
	require.NoError(t, err)
	assert.Equal(t, "btcusd,ethusd", lastReq().URL.Query().Get("tickers"))
}

func TestMeta_NoTickers_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background())
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "티커를 안 주면 전체 조회 — 쿼리 없음")
}

func TestMeta_EmptyTickerElement(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background(), "btcusd", "  ")
	assert.Error(t, err)
}

func TestMeta_Forbidden(t *testing.T) {
	c, _ := newStubClient(t, http.StatusForbidden, `{"detail":"nope"}`)
	_, err := c.Meta(context.Background(), "btcusd")
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}
