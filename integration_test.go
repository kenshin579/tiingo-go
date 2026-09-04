//go:build integration

package tiingo_test

import (
	"context"
	"os"
	"testing"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/eod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 실행: TIINGO_API_KEY=... go test -tags integration ./...
func newClient(t *testing.T) *tiingo.Client {
	t.Helper()
	if os.Getenv(tiingo.APIKeyEnv) == "" {
		t.Skip(tiingo.APIKeyEnv + " not set")
	}
	c, err := tiingo.NewClientFromEnv()
	require.NoError(t, err)
	return c
}

func TestIntegration_Meta(t *testing.T) {
	c := newClient(t)
	m, err := c.EOD.Meta(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", m.Ticker)
	assert.NotEmpty(t, m.Name)
	assert.Equal(t, "NASDAQ", m.ExchangeCode)
	require.NotNil(t, m.StartDate)
	assert.Equal(t, "1980-12-12", m.StartDate.String())
}

func TestIntegration_LatestPrice(t *testing.T) {
	c := newClient(t)
	p, err := c.EOD.LatestPrice(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.Greater(t, p.AdjClose, 0.0)
}

func TestIntegration_HistoricalPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.EOD.HistoricalPrices(context.Background(), "AAPL", &eod.PriceOptions{
		StartDate:    time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2019, 1, 7, 0, 0, 0, 0, time.UTC),
		ResampleFreq: eod.ResampleDaily,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)
	assert.Equal(t, "2019-01-02", ps[0].Date.String())
	assert.InDelta(t, 157.92, ps[0].Close, 0.01)
}

func TestIntegration_UnknownTicker(t *testing.T) {
	c := newClient(t)
	_, err := c.EOD.Meta(context.Background(), "NOSUCHTICKERXYZ")
	assert.Error(t, err)
}
