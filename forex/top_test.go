package forex

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
	raw, err := os.ReadFile("testdata/top_eurusd.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	qs, err := c.TopOfBook(context.Background(), []string{"eurusd"})
	require.NoError(t, err)
	require.Len(t, qs, 1)

	q := qs[0]
	assert.Equal(t, "eurusd", q.Ticker)
	assert.False(t, q.QuoteTimestamp.IsZero())
	assert.Greater(t, q.BidPrice, 0.0)
	assert.Greater(t, q.AskPrice, 0.0)
	assert.Greater(t, q.MidPrice, 0.0)
	assert.GreaterOrEqual(t, q.AskPrice, q.BidPrice)
	assert.Greater(t, q.BidSize, 0.0)

	assert.Equal(t, "/tiingo/fx/top", lastReq().URL.Path)
	assert.Equal(t, "eurusd", lastReq().URL.Query().Get("tickers"))
}

func TestTopOfBook_MultiFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/top_multi.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	qs, err := c.TopOfBook(context.Background(), []string{"eurusd", "usdjpy"})
	require.NoError(t, err)
	require.Len(t, qs, 2)
	assert.Equal(t, "eurusd", qs[0].Ticker)
	assert.Equal(t, "usdjpy", qs[1].Ticker)
	assert.Equal(t, "eurusd,usdjpy", lastReq().URL.Query().Get("tickers"))
}

func TestTopOfBook_EmptyTickers(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.TopOfBook(context.Background(), nil)
	assert.Error(t, err, "tickers 는 필수다")

	_, err = c.TopOfBook(context.Background(), []string{" "})
	assert.Error(t, err)
}

func TestTopOfBook_BadRequest(t *testing.T) {
	c, _ := newStubClient(t, http.StatusBadRequest, `{"detail":"Error: Please pass a \"tickers\" parameter."}`)
	_, err := c.TopOfBook(context.Background(), []string{"eurusd"})
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "tickers")
}
