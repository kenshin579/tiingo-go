package iex

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrices_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl_5min.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		StartDate:    time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample5Min,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)

	p := ps[0]
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.GreaterOrEqual(t, p.High, p.Low)
	assert.Zero(t, p.Volume, "columns 를 주지 않으면 응답에 volume 이 없어 0 이다")

	q := lastReq().URL.Query()
	assert.Equal(t, "/iex/aapl/prices", lastReq().URL.Path)
	assert.Equal(t, "2026-09-03", q.Get("startDate"))
	assert.Equal(t, "2026-09-03", q.Get("endDate"))
	assert.Equal(t, "5min", q.Get("resampleFreq"))
	assert.Empty(t, q.Get("columns"))
	assert.Empty(t, q.Get("afterHours"), "false 면 보내지 않는다")
	assert.Empty(t, q.Get("forceFill"))
}

// 5분봉이면 연속한 두 행이 5분 차이여야 한다 — types.Time 사용 근거를 고정한다.
func TestPrices_IntradayKeepsTime(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl_5min.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ps), 2)
	assert.Equal(t, 5*time.Minute, ps[1].Date.Time.Sub(ps[0].Date.Time))
}

func TestPrices_ColumnsVolume(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl_volume.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		ResampleFreq: Resample5Min,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)
	assert.Greater(t, ps[0].Volume, 0.0, "columns 에 volume 을 넣으면 채워진다")
	assert.Equal(t, "open,high,low,close,volume", lastReq().URL.Query().Get("columns"))
}

func TestPrices_AfterHoursForceFill(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", &PriceOptions{AfterHours: true, ForceFill: true})
	require.NoError(t, err)
	q := lastReq().URL.Query()
	assert.Equal(t, "true", q.Get("afterHours"))
	assert.Equal(t, "true", q.Get("forceFill"))
}

func TestPrices_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Equal(t, "/iex/aapl/prices", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestPrices_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "  ", nil)
	assert.Error(t, err)
}

// 없는 티커·휴장은 에러가 아니라 빈 슬라이스다.
func TestPrices_EmptyArray_NotAnError(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	ps, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, ps)
}
