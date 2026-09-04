package fundamentals

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

func TestMeta_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/meta_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ms, err := c.Meta(context.Background(), "aapl")
	require.NoError(t, err)
	require.Len(t, ms, 1)

	m := ms[0]
	assert.Equal(t, "aapl", m.Ticker)
	assert.Equal(t, "Apple Inc", m.Name)
	assert.Equal(t, "US000000000038", m.PermaTicker)
	assert.True(t, m.IsActive)
	assert.False(t, m.IsADR)
	assert.Equal(t, "Technology", m.Sector)
	assert.Equal(t, 3571, m.SICCode)
	assert.Equal(t, "usd", m.ReportingCurrency)
	assert.NotEmpty(t, m.DataProviderPermaTicker, "문서 표에 없지만 실제로 오는 필드")
	assert.False(t, m.StatementLastUpdated.IsZero())
	assert.NotZero(t, m.StatementLastUpdated.Hour()+m.StatementLastUpdated.Minute()+m.StatementLastUpdated.Second(),
		"types.Time 이므로 시각이 보존돼야 한다")

	assert.Equal(t, "/tiingo/fundamentals/meta", lastReq().URL.Path)
	assert.Equal(t, "aapl", lastReq().URL.Query().Get("tickers"))
}

func TestMeta_MultipleTickers(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background(), "aapl", "msft")
	require.NoError(t, err)
	assert.Equal(t, "aapl,msft", lastReq().URL.Query().Get("tickers"))
}

func TestMeta_NoTickers_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background())
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery, "티커를 안 주면 쿼리 없이 전체 조회")
}

func TestMeta_EmptyTickerElement(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Meta(context.Background(), "aapl", "  ")
	assert.Error(t, err, "빈 티커가 섞이면 에러")
}

func TestMeta_Forbidden(t *testing.T) {
	c, _ := newStubClient(t, http.StatusForbidden, `{"detail":"You do not have permission"}`)
	_, err := c.Meta(context.Background(), "aapl")
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}
