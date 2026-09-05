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

func TestTopOfBook_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/top_btcusd.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ts, err := c.TopOfBook(context.Background(), []string{"btcusd"}, &TopOptions{
		Exchanges: []string{"GDAX"},
	})
	require.NoError(t, err)
	require.Len(t, ts, 1)

	b := ts[0]
	assert.Equal(t, "btcusd", b.Ticker)
	require.NotEmpty(t, b.TopOfBookData)

	d := b.TopOfBookData[0]
	assert.False(t, d.QuoteTimestamp.IsZero())
	assert.False(t, d.LastSaleTimestamp.IsZero())
	assert.Greater(t, d.BidPrice, 0.0)
	assert.Greater(t, d.AskPrice, 0.0)
	assert.Greater(t, d.LastPrice, 0.0)
	assert.NotEmpty(t, d.BidExchange)
	assert.NotEmpty(t, d.LastExchange)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/crypto/top", lastReq().URL.Path)
	assert.Equal(t, "btcusd", q.Get("tickers"))
	assert.Equal(t, "GDAX", q.Get("exchanges"))
	assert.Empty(t, q.Get("includeRawExchangeData"), "false 면 보내지 않는다")
}

func TestTopOfBook_MultiFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/top_multi.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ts, err := c.TopOfBook(context.Background(), []string{"btcusd", "ethusd"}, nil)
	require.NoError(t, err)
	require.Len(t, ts, 2)
	assert.Equal(t, "btcusd", ts[0].Ticker)
	assert.Equal(t, "ethusd", ts[1].Ticker)
	assert.Equal(t, "btcusd,ethusd", lastReq().URL.Query().Get("tickers"))
}

func TestTopOfBook_IncludeRawExchangeData(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.TopOfBook(context.Background(), []string{"btcusd"}, &TopOptions{IncludeRawExchangeData: true})
	require.NoError(t, err)
	assert.Equal(t, "true", lastReq().URL.Query().Get("includeRawExchangeData"))
}

func TestTopOfBook_EmptyTickers(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.TopOfBook(context.Background(), []string{}, nil)
	assert.Error(t, err)
}

func TestTopOfBookFor_EmptyArray_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.TopOfBookFor(context.Background(), "btcusd", nil)
	assert.True(t, errors.Is(err, httpclient.ErrNotFound))
}
