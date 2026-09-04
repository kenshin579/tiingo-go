//go:build integration

package tiingo_test

import (
	"context"
	"os"
	"testing"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/eod"
	"github.com/kenshin579/tiingo-go/fundamentals"
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

func TestIntegration_FundamentalsDefinitions(t *testing.T) {
	c := newClient(t)
	ds, err := c.Fundamentals.Definitions(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ds)

	known := map[string]bool{}
	for _, code := range fundamentals.AllCodes {
		known[code] = true
	}
	var missing []string
	for _, d := range ds {
		if !known[d.DataCode] {
			missing = append(missing, d.DataCode)
		}
	}
	assert.Emptyf(t, missing, "codes.go 갱신 필요 — 새 dataCode: %v", missing)
}

func TestIntegration_FundamentalsMeta(t *testing.T) {
	c := newClient(t)
	ms, err := c.Fundamentals.Meta(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, "Apple Inc", ms[0].Name)
	assert.False(t, ms[0].StatementLastUpdated.IsZero())
}

func TestIntegration_FundamentalsStatements(t *testing.T) {
	c := newClient(t)
	ss, err := c.Fundamentals.Statements(context.Background(), "AAPL", &fundamentals.StatementOptions{
		StartDate: time.Now().AddDate(-1, 0, 0),
		Sort:      "-date",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ss)
	rev, ok := ss[0].StatementData.Get(fundamentals.CodeRevenue)
	assert.True(t, ok)
	assert.Greater(t, rev, 0.0)
}

func TestIntegration_FundamentalsDaily(t *testing.T) {
	c := newClient(t)
	ds, err := c.Fundamentals.Daily(context.Background(), "AAPL", &fundamentals.DailyOptions{
		StartDate: time.Now().AddDate(0, 0, -14),
	})
	require.NoError(t, err)
	require.NotEmpty(t, ds)
	assert.Greater(t, ds[0].MarketCap, 0.0)
}
