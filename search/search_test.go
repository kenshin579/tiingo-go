package search

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

func TestSearch_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/search_apple.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	rs, err := c.Search(context.Background(), "apple", nil)
	require.NoError(t, err)
	require.NotEmpty(t, rs)

	// 첫 행은 캐나다 상장 AAPL 이고 openFIGIComposite 가 null 이다 — 모든 태그를 한 번에 고정한다.
	assert.Equal(t, Result{
		Ticker:            "AAPL",
		Name:              "Apple Inc",
		PermaTicker:       "CA000000137372",
		OpenFIGIComposite: "",
		AssetType:         "Stock",
		IsActive:          true,
		CountryCode:       "CA",
	}, rs[0])

	assert.Equal(t, "/tiingo/utilities/search", lastReq().URL.Path)
	assert.Equal(t, "apple", lastReq().URL.Query().Get("query"))
}

// 같은 티커가 국가별로 여러 건 올 수 있다. SDK 가 임의로 합치거나 걸러내면 안 된다.
func TestSearch_DuplicateTickersPreserved(t *testing.T) {
	raw, err := os.ReadFile("testdata/search_apple.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	rs, err := c.Search(context.Background(), "apple", nil)
	require.NoError(t, err)

	var aapl []Result
	for _, r := range rs {
		if r.Ticker == "AAPL" {
			aapl = append(aapl, r)
		}
	}
	require.Len(t, aapl, 2, "미국·캐나다 AAPL 두 건이 그대로 온다")
	assert.NotEqual(t, aapl[0].PermaTicker, aapl[1].PermaTicker, "유일 키는 PermaTicker 다")
}

// null FIGI 가 빈 문자열로 정규화되는지 실제 응답으로 확인한다.
func TestSearch_NullFIGINormalized(t *testing.T) {
	raw, err := os.ReadFile("testdata/search_apple.json")
	require.NoError(t, err)

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	rs, err := c.Search(context.Background(), "apple", nil)
	require.NoError(t, err)

	var empty, filled int
	for _, r := range rs {
		if r.OpenFIGIComposite == "" {
			empty++
		} else {
			filled++
		}
	}
	assert.Positive(t, empty, "null 인 행이 빈 문자열이 된다")
	assert.Positive(t, filled, "정상 FIGI 도 함께 온다")
}

// 문자열 "nan" 이 섞인 실제 응답에서도 빈 문자열이어야 한다.
func TestSearch_NaNFIGINormalized(t *testing.T) {
	raw, err := os.ReadFile("testdata/search_microsoft.json")
	require.NoError(t, err)
	require.Contains(t, string(raw), `"nan"`, "fixture 에 nan 이 들어 있어야 의미가 있다")

	c, _ := newStubClient(t, http.StatusOK, string(raw))
	rs, err := c.Search(context.Background(), "microsoft", nil)
	require.NoError(t, err)
	for _, r := range rs {
		assert.NotEqual(t, FIGI("nan"), r.OpenFIGIComposite, "nan 이 그대로 남으면 안 된다")
	}
}

// columns 를 주면 응답에서 빠진 필드는 zero 값이 된다.
func TestSearch_Columns(t *testing.T) {
	raw, err := os.ReadFile("testdata/search_columns.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	rs, err := c.Search(context.Background(), "apple", &SearchOptions{
		Columns: []string{"ticker", "name"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, rs)
	assert.NotEmpty(t, rs[0].Ticker)
	assert.Empty(t, rs[0].PermaTicker, "columns 에 없는 필드는 zero 값이다")
	assert.Equal(t, "ticker,name", lastReq().URL.Query().Get("columns"))
}

func TestSearch_AllOptions(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Search(context.Background(), "apple", &SearchOptions{
		ExactTickerMatch: true,
		IncludeDelisted:  true,
		Limit:            25,
		Columns:          []string{"ticker", "name"},
	})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "apple", q.Get("query"))
	assert.Equal(t, "true", q.Get("exactTickerMatch"))
	assert.Equal(t, "true", q.Get("includeDelisted"))
	assert.Equal(t, "25", q.Get("limit"))
	assert.Equal(t, "ticker,name", q.Get("columns"))
}

// zero 값 옵션은 요청에서 빠진다.
func TestSearch_ZeroOptionsOmitted(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Search(context.Background(), "apple", &SearchOptions{})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "apple", q.Get("query"))
	assert.Empty(t, q.Get("exactTickerMatch"), "false 면 보내지 않는다")
	assert.Empty(t, q.Get("includeDelisted"))
	assert.Empty(t, q.Get("limit"), "0 이면 보내지 않는다 — Tiingo 기본값 10 을 쓴다")
	assert.Empty(t, q.Get("columns"))

	c2, lastReq2 := newStubClient(t, http.StatusOK, `[]`)
	_, err = c2.Search(context.Background(), "apple", &SearchOptions{Limit: -1})
	require.NoError(t, err)
	assert.Empty(t, lastReq2().URL.Query().Get("limit"), "음수도 미전송이다")
}

func TestSearch_NilOptions(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Search(context.Background(), "apple", nil)
	require.NoError(t, err)
	assert.Equal(t, "query=apple", lastReq().URL.RawQuery)
}

// 검색어에 공백·슬래시가 들어가도 쿼리 파라미터라 안전하다.
func TestSearch_EscapesQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Search(context.Background(), "johnson & johnson/co", nil)
	require.NoError(t, err)
	assert.Equal(t, "/tiingo/utilities/search", lastReq().URL.Path)
	assert.Equal(t, "johnson & johnson/co", lastReq().URL.Query().Get("query"))
}

func TestSearch_EmptyQuery(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Search(context.Background(), "   ", nil)
	assert.Error(t, err, "왕복하기 전에 막는다")
}

// 일치하는 자산이 없으면 200 + [] 다. 에러가 아니다.
func TestSearch_EmptyResult(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	rs, err := c.Search(context.Background(), "zzzznosuchassetxyz", nil)
	require.NoError(t, err)
	assert.Empty(t, rs)
}

func TestSearch_APIError(t *testing.T) {
	c, _ := newStubClient(t, http.StatusBadRequest, `{"detail":"Error: query format was not correct."}`)
	_, err := c.Search(context.Background(), "apple", nil)
	var apiErr *httpclient.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestSearchByISIN_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/search_isin.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	rs, err := c.SearchByISIN(context.Background(), "US0378331005", nil)
	require.NoError(t, err)
	require.Len(t, rs, 1, "ISIN 은 자산 하나를 정확히 지목한다")
	assert.Equal(t, "AAPL", rs[0].Ticker)
	assert.NotEmpty(t, rs[0].PermaTicker)

	q := lastReq().URL.Query()
	assert.Equal(t, "US0378331005", q.Get("isin"))
	assert.Empty(t, q.Get("query"), "isin 검색은 query 를 함께 보내지 않는다")
}

// ISIN 검색도 같은 옵션을 받는다.
func TestSearchByISIN_Options(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.SearchByISIN(context.Background(), "US0378331005", &SearchOptions{
		Limit:   1,
		Columns: []string{"ticker", "name"},
	})
	require.NoError(t, err)

	q := lastReq().URL.Query()
	assert.Equal(t, "US0378331005", q.Get("isin"))
	assert.Equal(t, "1", q.Get("limit"))
	assert.Equal(t, "ticker,name", q.Get("columns"))
}

// 에러 메시지가 어느 인자인지 가리키는지 고정한다 — searchBy 의 key 가 인자 이름 노릇을 한다.
func TestSearchByISIN_EmptyISIN(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.SearchByISIN(context.Background(), " ", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "isin")
}
