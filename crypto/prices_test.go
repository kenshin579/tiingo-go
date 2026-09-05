package crypto

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

func TestPrices_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_btcusd_1day.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Prices(context.Background(), []string{"btcusd"}, &PriceOptions{
		StartDate:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample1Day,
		Exchanges:    []string{"GDAX", "BULLISH"},
	})
	require.NoError(t, err)
	require.Len(t, ss, 1)

	s := ss[0]
	assert.Equal(t, "btcusd", s.Ticker)
	assert.Equal(t, "btc", s.BaseCurrency)
	assert.Equal(t, "usd", s.QuoteCurrency)
	require.NotEmpty(t, s.PriceData)

	p := s.PriceData[0]
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.Greater(t, p.High, 0.0)
	assert.GreaterOrEqual(t, p.High, p.Low)
	assert.Greater(t, p.Volume, 0.0)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/crypto/prices", lastReq().URL.Path)
	assert.Equal(t, "btcusd", q.Get("tickers"))
	assert.Equal(t, "2026-09-01", q.Get("startDate"))
	assert.Equal(t, "2026-09-04", q.Get("endDate"))
	assert.Equal(t, "1day", q.Get("resampleFreq"))
	assert.Equal(t, "GDAX,BULLISH", q.Get("exchanges"))
}

// 1min fixture 로 types.Time 사용이 옳은지 고정한다 — Date 였다면 시각이 잘려 두 행이 같아진다.
func TestPrices_IntradayKeepsTime(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_btcusd_1min.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	s, err := c.PricesFor(context.Background(), "btcusd", &PriceOptions{ResampleFreq: Resample1Min})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(s.PriceData), 2)

	d0, d1 := s.PriceData[0].Date.Time, s.PriceData[1].Date.Time
	assert.Equal(t, time.Minute, d1.Sub(d0), "1분봉이면 두 행이 1분 차이여야 한다")
	assert.NotEqual(t, d0, d1)
}

func TestPrices_NilOptions_TickersOnly(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), []string{"btcusd"}, nil)
	require.NoError(t, err)
	q := lastReq().URL.Query()
	assert.Equal(t, "btcusd", q.Get("tickers"))
	assert.Empty(t, q.Get("startDate"))
	assert.Empty(t, q.Get("resampleFreq"))
	assert.Empty(t, q.Get("exchanges"))
}

func TestPrices_MultipleTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), []string{"btcusd", "ethusd"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "btcusd,ethusd", lastReq().URL.Query().Get("tickers"))
}

func TestPrices_EmptyTickers(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), nil, nil)
	assert.Error(t, err, "tickers 는 필수다")

	_, err = c.Prices(context.Background(), []string{" "}, nil)
	assert.Error(t, err)
}

func TestPricesFor_EmptyArray_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.PricesFor(context.Background(), "btcusd", nil)
	assert.True(t, errors.Is(err, httpclient.ErrNotFound), "빈 배열은 ErrNotFound")
}
