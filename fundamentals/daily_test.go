package fundamentals

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaily_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/daily_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ds, err := c.Daily(context.Background(), "aapl", &DailyOptions{
		StartDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Sort:      "-date",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ds)

	d := ds[0]
	assert.False(t, d.Date.IsZero())
	assert.Greater(t, d.MarketCap, 0.0)
	assert.Greater(t, d.EnterpriseVal, 0.0)
	assert.Greater(t, d.PERatio, 0.0)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/fundamentals/aapl/daily", lastReq().URL.Path)
	assert.Equal(t, "2026-09-01", q.Get("startDate"))
	assert.Equal(t, "-date", q.Get("sort"))
	assert.Empty(t, q.Get("endDate"))
}

func TestDaily_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Daily(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestDaily_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Daily(context.Background(), "", nil)
	assert.Error(t, err)
}
