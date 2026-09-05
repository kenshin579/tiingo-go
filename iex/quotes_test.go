package iex

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

func TestQuotes_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/quotes_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	qs, err := c.Quotes(context.Background(), []string{"aapl"})
	require.NoError(t, err)
	require.Len(t, qs, 1)

	q := qs[0]
	assert.Equal(t, "AAPL", q.Ticker, "응답은 대문자로 온다")
	assert.False(t, q.Timestamp.IsZero())
	assert.Greater(t, q.Open, 0.0)
	assert.Greater(t, q.High, 0.0)
	assert.Greater(t, q.Low, 0.0)
	assert.Greater(t, q.TngoLast, 0.0)
	assert.Greater(t, q.PrevClose, 0.0)
	assert.Greater(t, q.Volume, int64(0))

	assert.Equal(t, "/iex/", lastReq().URL.Path, "끝 슬래시가 있어야 301 을 피한다")
	assert.Equal(t, "aapl", lastReq().URL.Query().Get("tickers"))
}

// 장 마감 fixture 라 호가·체결 9개 필드가 nil 이어야 한다 — 값 타입이었다면 0 으로 뭉개진다.
func TestQuotes_NullFieldsAreNil(t *testing.T) {
	raw, err := os.ReadFile("testdata/quotes_aapl.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	qs, err := c.Quotes(context.Background(), []string{"aapl"})
	require.NoError(t, err)
	require.Len(t, qs, 1)

	q := qs[0]
	assert.Nil(t, q.LastSaleTimestamp)
	assert.Nil(t, q.QuoteTimestamp)
	assert.Nil(t, q.Mid)
	assert.Nil(t, q.Last)
	assert.Nil(t, q.LastSize)
	assert.Nil(t, q.BidSize)
	assert.Nil(t, q.BidPrice)
	assert.Nil(t, q.AskPrice)
	assert.Nil(t, q.AskSize)
}

// 장중 응답을 흉내 내 포인터에 값이 들어오는지 확인한다.
func TestQuotes_NonNullFields(t *testing.T) {
	body := `[{"ticker":"AAPL","timestamp":"2026-09-03T15:59:59.999999+00:00",
	  "lastSaleTimestamp":"2026-09-03T15:59:58.000000+00:00","quoteTimestamp":"2026-09-03T15:59:59.500000+00:00",
	  "open":324.97,"high":330.8,"low":324.3,"mid":327.5,"tngoLast":327.6,"last":327.55,
	  "lastSize":100,"bidSize":200,"bidPrice":327.4,"askPrice":327.6,"askSize":300,
	  "volume":39606884,"prevClose":328.21}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	qs, err := c.Quotes(context.Background(), []string{"aapl"})
	require.NoError(t, err)
	require.Len(t, qs, 1)

	q := qs[0]
	require.NotNil(t, q.BidPrice)
	assert.InDelta(t, 327.4, *q.BidPrice, 0.001)
	require.NotNil(t, q.AskSize)
	assert.Equal(t, int64(300), *q.AskSize)
	require.NotNil(t, q.QuoteTimestamp)
	assert.False(t, q.QuoteTimestamp.IsZero())
	require.NotNil(t, q.Last)
	assert.InDelta(t, 327.55, *q.Last, 0.001)
}

// 없는 티커는 에러가 아니라 응답에서 빠진다. 순서도 요청과 다를 수 있다.
func TestQuotes_UnknownTickerOmitted(t *testing.T) {
	raw, err := os.ReadFile("testdata/quotes_multi.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	qs, err := c.Quotes(context.Background(), []string{"aapl", "msft", "dcfc"})
	require.NoError(t, err)
	assert.Len(t, qs, 2, "없는 티커는 응답에서 빠져 요청보다 짧다")

	seen := map[string]bool{}
	for _, q := range qs {
		seen[q.Ticker] = true
	}
	assert.True(t, seen["AAPL"] && seen["MSFT"], "순서는 보장되지 않으므로 Ticker 로 찾는다")
	assert.Equal(t, "aapl,msft,dcfc", lastReq().URL.Query().Get("tickers"))
}

func TestQuotes_EmptyTickers(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Quotes(context.Background(), nil)
	assert.Error(t, err)

	_, err = c.Quotes(context.Background(), []string{" "})
	assert.Error(t, err)
}

func TestQuotes_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Quotes(context.Background(), []string{"aapl"})
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
