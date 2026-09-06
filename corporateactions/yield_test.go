package corporateactions

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

func TestDistributionYield_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/yield_aapl_recent.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ys, err := c.DistributionYield(context.Background(), "aapl", &YieldOptions{
		StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.NotEmpty(t, ys)

	y := ys[0]
	assert.False(t, y.Date.IsZero())
	assert.Greater(t, y.TrailingDiv1Y, 0.0)

	assert.Equal(t, "/tiingo/corporate-actions/aapl/distribution-yield", lastReq().URL.Path)
	assert.Equal(t, "2026-08-01", lastReq().URL.Query().Get("startDate"))
	assert.Empty(t, lastReq().URL.Query().Get("endDate"), "zero 면 보내지 않는다")
}

// 응답은 T00:00:00.000Z 로 오지만 의미는 날짜다. types.Date 가 시각을 잘라내는지 확인한다.
func TestDistributionYield_DateIsDateOnly(t *testing.T) {
	body := `[{"date":"2026-09-04T00:00:00.000Z","trailingDiv1Y":0.0033128106}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ys, err := c.DistributionYield(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.Len(t, ys, 1)
	assert.Equal(t, "2026-09-04", ys[0].Date.String())
	assert.InDelta(t, 0.0033128106, ys[0].TrailingDiv1Y, 1e-12)
}

// 배당 시작 전 구간은 0.0 이 정상 값이다 — 결손이 아니므로 포인터로 두지 않는다.
func TestDistributionYield_ZeroIsValid(t *testing.T) {
	raw, err := os.ReadFile("testdata/yield_aapl_early.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ys, err := c.DistributionYield(context.Background(), "aapl", nil)
	require.NoError(t, err)
	require.NotEmpty(t, ys)

	var zeros int
	for _, y := range ys {
		if y.TrailingDiv1Y == 0.0 {
			zeros++
			assert.False(t, y.Date.IsZero(), "값이 0 이어도 날짜는 있다")
		}
	}
	assert.Positive(t, zeros, "fixture 에 0.0 인 행이 있어야 의미가 있다")
}

func TestDistributionYield_BothDates(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "aapl", &YieldOptions{
		StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "2026-08-01", q.Get("startDate"))
	assert.Equal(t, "2026-08-15", q.Get("endDate"))
}

func TestDistributionYield_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/corporate-actions/aapl/distribution-yield", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

// zero 값만 든 옵션은 nil 옵션과 똑같이 쿼리가 비어야 한다.
func TestDistributionYield_ZeroOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "aapl", &YieldOptions{})
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "보낼 값이 없으면 쿼리를 만들지 않는다")
}

// 앞뒤 공백은 잘라내고 경로에 넣는다.
func TestDistributionYield_TrimsTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "  aapl  ", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/corporate-actions/aapl/distribution-yield", lastReq().URL.Path)
}

func TestDistributionYield_EmptyTicker(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.DistributionYield(context.Background(), "  ", nil)
	require.Error(t, err)
	assert.Nil(t, lastReq(), "왕복하기 전에 막는다")
}

// 없는 티커는 200 + [] 다 — equity·iex 의 시세가 404 를 주는 것과 다르다.
func TestDistributionYield_EmptyResult(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	ys, err := c.DistributionYield(context.Background(), "nosuchxyz", nil)
	require.NoError(t, err)
	assert.Empty(t, ys)
}

// 같은 API 그룹의 다른 엔드포인트가 403 이므로 이 경로도 언젠가 막힐 수 있다.
func TestDistributionYield_Forbidden(t *testing.T) {
	c, _ := newStubClient(t, http.StatusForbidden, ``)
	_, err := c.DistributionYield(context.Background(), "aapl", nil)
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}
