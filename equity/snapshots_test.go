package equity

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

func TestSnapshots_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_aapl_spy.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "spy"})
	require.NoError(t, err)
	require.Len(t, ss, 2)

	byTicker := map[string]Snapshot{}
	for _, s := range ss {
		byTicker[s.Ticker] = s
	}
	s, ok := byTicker["AAPL"]
	require.True(t, ok, "응답은 대문자 티커로 온다")
	assert.False(t, s.Timestamp.IsZero())
	assert.Greater(t, s.Open, 0.0)
	assert.Greater(t, s.High, 0.0)
	assert.Greater(t, s.Low, 0.0)
	assert.Greater(t, s.TngoLast, 0.0)
	assert.Greater(t, s.PrevClose, 0.0)
	assert.Greater(t, s.LqRefPrice, 0.0)
	assert.Greater(t, s.Volume, 0.0)

	assert.Equal(t, "/tiingo/equity/intraday/", lastReq().URL.Path, "끝 슬래시가 있어야 301 을 피한다")
	assert.Equal(t, "aapl,spy", lastReq().URL.Query().Get("tickers"))
}

// volume 은 문서상 int64 지만 1778528.0 처럼 소수점이 붙어 온다 — int64 였다면 디코딩이 실패한다.
// fixture 는 재포맷하면 .0 이 사라질 수 있어 리터럴을 인라인으로 둔다.
func TestSnapshots_FractionalVolume(t *testing.T) {
	body := `[{"ticker":"AAPL","timestamp":"2026-09-04T19:59:34.935723167-04:00",
	  "open":328.29,"high":328.915,"low":317.845,"tngoLast":320.08,"prevClose":328.29,
	  "volume":1778528.0,"lqRefPrice":320.08,"lqSpread":0.0005,
	  "lqBidPrice":320.015,"lqBidSize":100,"lqAskPrice":320.175,"lqAskSize":100}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ss, err := c.Snapshots(context.Background(), []string{"aapl"})
	require.NoError(t, err)
	require.Len(t, ss, 1)
	assert.InDelta(t, 1778528.0, ss[0].Volume, 0.001)
}

// 같은 엔드포인트가 소수점 없는 정수도 보낸다(fixture 의 AULT 는 972327, 전 종목 조회의
// 000425 는 147928660). 둘 다 받아야 한다.
func TestSnapshots_IntegerVolume(t *testing.T) {
	body := `[{"ticker":"000425","timestamp":"2026-07-29T20:00:00+00:00",
	  "open":8.76,"high":9.09,"low":8.72,"tngoLast":9.0,"prevClose":8.76,
	  "volume":147928660,"lqRefPrice":9.0,"lqSpread":null,
	  "lqBidPrice":null,"lqBidSize":null,"lqAskPrice":null,"lqAskSize":null}]`

	c, _ := newStubClient(t, http.StatusOK, body)
	ss, err := c.Snapshots(context.Background(), []string{"000425"})
	require.NoError(t, err)
	require.Len(t, ss, 1)
	assert.InDelta(t, 147928660.0, ss[0].Volume, 0.001)
}

// 유동성 5개 필드는 통합 피드가 값을 안 내면 null 이다 — 값 타입이었다면 0 으로 뭉개진다.
func TestSnapshots_NullLiquidityIsNil(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_null_liquidity.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "dcfc", "pxs", "ault"})
	require.NoError(t, err)

	var nilCount int
	for _, s := range ss {
		if s.LqSpread == nil {
			nilCount++
			assert.Nil(t, s.LqBidPrice, "유동성 필드는 함께 비어 온다")
			assert.Nil(t, s.LqBidSize)
			assert.Nil(t, s.LqAskPrice)
			assert.Nil(t, s.LqAskSize)
			assert.Greater(t, s.TngoLast, 0.0, "유동성이 비어도 가격 필드는 채워진다")
			assert.Greater(t, s.Volume, 0.0, "정수로 온 volume 도 디코딩된다")
		}
	}
	assert.Positive(t, nilCount, "fixture 에 Lq 가 null 인 행이 있어야 의미가 있다")
}

// 장중처럼 값이 들어온 경우도 확인한다.
func TestSnapshots_NonNullLiquidity(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_aapl_spy.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "spy"})
	require.NoError(t, err)
	require.NotEmpty(t, ss)

	// 순서가 보장되지 않는 응답이므로 인덱스가 아니라 Ticker 로 찾는다.
	byTicker := map[string]Snapshot{}
	for _, r := range ss {
		byTicker[r.Ticker] = r
	}
	s, ok := byTicker["AAPL"]
	require.True(t, ok)
	require.NotNil(t, s.LqSpread)
	assert.Greater(t, *s.LqSpread, 0.0)
	require.NotNil(t, s.LqBidPrice)
	assert.Greater(t, *s.LqBidPrice, 0.0)
	require.NotNil(t, s.LqBidSize)
	assert.Greater(t, *s.LqBidSize, 0.0)
	require.NotNil(t, s.LqAskPrice)
	require.NotNil(t, s.LqAskSize)
}

// 없는 티커는 에러가 아니라 응답에서 빠진다. 순서도 요청과 다를 수 있다.
func TestSnapshots_UnknownTickerOmitted(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_null_liquidity.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Snapshots(context.Background(), []string{"aapl", "dcfc", "pxs", "ault"})
	require.NoError(t, err)
	assert.Less(t, len(ss), 4, "fixture 자체가 요청 4개보다 짧다 — 아래 Ticker 조회가 실제 단정이다")

	seen := map[string]bool{}
	for _, s := range ss {
		seen[s.Ticker] = true
	}
	assert.True(t, seen["AAPL"], "순서는 보장되지 않으므로 Ticker 로 찾는다")
	assert.Equal(t, "aapl,dcfc,pxs,ault", lastReq().URL.Query().Get("tickers"))
}

func TestSnapshots_EmptyTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Snapshots(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, lastReq(), "왕복하기 전에 막는다")

	_, err = c.Snapshots(context.Background(), []string{" "})
	assert.Error(t, err, "공백뿐인 원소도 에러")
}

// 앞뒤 공백은 잘라내고 보낸다 — 그대로 보내면 Tiingo 가 빈 배열을 돌려준다.
func TestSnapshots_TrimsTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Snapshots(context.Background(), []string{" aapl ", "spy"})
	require.NoError(t, err)
	assert.Equal(t, "aapl,spy", lastReq().URL.Query().Get("tickers"))
}

// AllSnapshots 는 tickers 를 아예 보내지 않는다 — 그래야 전 종목이 온다.
func TestAllSnapshots_NoTickersParam(t *testing.T) {
	raw, err := os.ReadFile("testdata/snapshots_aapl_spy.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.AllSnapshots(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ss)

	assert.Equal(t, "/tiingo/equity/intraday/", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery, "쿼리가 비어야 전 종목이 온다")
}

func TestSnapshots_APIError(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Snapshots(context.Background(), []string{"aapl"})
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
