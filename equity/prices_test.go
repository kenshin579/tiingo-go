package equity

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
	raw, err := os.ReadFile("testdata/prices_aapl_1hour.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		StartDate:    time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample1Hour,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)

	p := ps[0]
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.GreaterOrEqual(t, p.High, p.Low)
	assert.Zero(t, p.Volume, "columns 를 주지 않으면 응답에 volume 이 없어 0 이다")

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/equity/intraday/aapl/prices", lastReq().URL.Path)
	assert.Equal(t, "2026-09-04", q.Get("startDate"))
	assert.Equal(t, "1hour", q.Get("resampleFreq"))
	assert.Empty(t, q.Get("endDate"), "zero 면 보내지 않는다")
	assert.Empty(t, q.Get("afterHours"), "false 면 보내지 않는다")
	assert.Empty(t, q.Get("forceFill"))
	assert.Empty(t, q.Get("columns"))
}

// columns 에 volume 을 넣으면 거래량이 온다.
func TestPrices_ColumnsVolume(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_aapl_volume.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		ResampleFreq: Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)
	assert.Greater(t, ps[0].Volume, 0.0, "columns 에 volume 을 넣으면 채워진다")
	assert.Equal(t, "open,high,low,close,volume", lastReq().URL.Query().Get("columns"))
}

// 소수점이 붙은 volume 을 실제로 디코딩하는지 못 박는다 — int64 였다면 여기서 에러가 난다.
// fixture 는 재포맷하면 .0 이 사라질 수 있어 리터럴을 인라인으로 둔다.
func TestPrices_FractionalVolume(t *testing.T) {
	body := `[{"date":"2026-09-04T14:00:00.000Z","open":323.64,"high":323.72,
	  "low":318.29,"close":318.44,"volume":448768.0}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ps, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.Len(t, ps, 1)
	assert.InDelta(t, 448768.0, ps[0].Volume, 0.001)
}

func TestPrices_AllOptions(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", &PriceOptions{
		StartDate:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample5Min,
		AfterHours:   true,
		ForceFill:    true,
		Columns:      []string{"open", "close"},
	})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "2026-09-01", q.Get("startDate"))
	assert.Equal(t, "2026-09-04", q.Get("endDate"))
	assert.Equal(t, "5min", q.Get("resampleFreq"))
	assert.Equal(t, "true", q.Get("afterHours"))
	assert.Equal(t, "true", q.Get("forceFill"))
	assert.Equal(t, "open,close", q.Get("columns"))
}

func TestPrices_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/equity/intraday/aapl/prices", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

// 앞뒤 공백은 잘라내고 경로에 넣는다.
// zero 값만 든 옵션은 nil 옵션과 똑같이 쿼리가 비어야 한다.
func TestPrices_ZeroOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "aapl", &PriceOptions{})
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "보낼 값이 없으면 쿼리를 만들지 않는다")
}

func TestPrices_TrimsTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "  aapl  ", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/equity/intraday/aapl/prices", lastReq().URL.Path)
}

func TestPrices_EmptyTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), "  ", nil)
	require.Error(t, err)
	assert.Nil(t, lastReq(), "왕복하기 전에 막는다")
}

// 휴장 구간은 에러가 아니라 빈 슬라이스다.
func TestPrices_EmptyResult(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	ps, err := c.Prices(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, ps)
}

// 없는 티커는 404 다 — 빈 슬라이스가 아니다.
func TestPrices_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Prices(context.Background(), "nosuchxyz", nil)
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
