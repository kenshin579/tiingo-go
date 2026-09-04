package eod

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
	m, err := c.Meta(context.Background(), "aapl")
	require.NoError(t, err)

	assert.Equal(t, "AAPL", m.Ticker)
	assert.Equal(t, "Apple Inc", m.Name)
	assert.Equal(t, "NASDAQ", m.ExchangeCode)
	assert.NotEmpty(t, m.Description)
	require.NotNil(t, m.StartDate)
	assert.Equal(t, "1980-12-12", m.StartDate.String())
	require.NotNil(t, m.EndDate)
	assert.False(t, m.EndDate.IsZero())

	assert.Equal(t, "/tiingo/daily/aapl", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestMeta_NullDates(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `{"ticker":"XYZ","name":"No Data","exchangeCode":"NYSE","description":"","startDate":null,"endDate":null}`)
	m, err := c.Meta(context.Background(), "xyz")
	require.NoError(t, err)
	assert.Nil(t, m.StartDate, "startDate null 이면 nil")
	assert.Nil(t, m.EndDate)
}

func TestMeta_EmptyStringDates(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `{"ticker":"XYZ","name":"No Data","exchangeCode":"NYSE","description":"","startDate":"","endDate":""}`)
	m, err := c.Meta(context.Background(), "xyz")
	require.NoError(t, err)
	require.NotNil(t, m.StartDate, "빈 문자열은 nil 이 아니라 zero Date")
	assert.True(t, m.StartDate.IsZero())
	assert.True(t, m.EndDate.IsZero())
}

func TestMeta_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `{}`)
	_, err := c.Meta(context.Background(), "  ")
	assert.Error(t, err)
}

func TestMeta_NotFound(t *testing.T) {
	c, _ := newStubClient(t, http.StatusNotFound, `{"detail":"Not found."}`)
	_, err := c.Meta(context.Background(), "nosuch")
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
