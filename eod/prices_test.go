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

func TestHistoricalPrices_EmptyOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.HistoricalPrices(context.Background(), "aapl", &PriceOptions{Columns: []string{}})
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "값이 모두 zero 면 쿼리를 붙이지 않는다")
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
