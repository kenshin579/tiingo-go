package forex

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
	raw, err := os.ReadFile("testdata/prices_eurusd_1day.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), []string{"eurusd"}, &PriceOptions{
		StartDate:    time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
		ResampleFreq: Resample1Day,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)

	p := ps[0]
	assert.Equal(t, "eurusd", p.Ticker)
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.GreaterOrEqual(t, p.High, p.Low)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/fx/eurusd/prices", lastReq().URL.Path)
	assert.Equal(t, "2026-09-01", q.Get("startDate"))
	assert.Equal(t, "2026-09-04", q.Get("endDate"))
	assert.Equal(t, "1day", q.Get("resampleFreq"))
}

// 1min fixture 로 types.Time 사용이 옳은지 고정한다 — Date 였다면 시각이 잘려 두 행이 같아진다.
func TestPrices_IntradayKeepsTime(t *testing.T) {
	raw, err := os.ReadFile("testdata/prices_eurusd_1min.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ps, err := c.Prices(context.Background(), []string{"eurusd"}, &PriceOptions{ResampleFreq: Resample1Min})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ps), 2)

	d0, d1 := ps[0].Date.Time, ps[1].Date.Time
	assert.Equal(t, time.Minute, d1.Sub(d0), "1분봉이면 두 행이 1분 차이여야 한다")
}

func TestPrices_MultipleTickers_PathJoin(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), []string{"eurusd", "usdjpy"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/fx/eurusd,usdjpy/prices", lastReq().URL.Path, "복수 티커는 경로에 콤마로")
	assert.Empty(t, lastReq().URL.RawQuery, "옵션이 없으면 쿼리도 없다")
}

func TestPrices_EmptyTickers(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Prices(context.Background(), nil, nil)
	assert.Error(t, err)

	_, err = c.Prices(context.Background(), []string{"  "}, nil)
	assert.Error(t, err)
}

// 주말·없는 통화쌍은 에러가 아니라 빈 슬라이스다.
func TestPrices_EmptyArray_NotAnError(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	ps, err := c.Prices(context.Background(), []string{"eurusd"}, nil)
	require.NoError(t, err)
	assert.Empty(t, ps)
}
