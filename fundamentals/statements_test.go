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

func TestStatements_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/statements_aapl.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ss, err := c.Statements(context.Background(), "aapl", &StatementOptions{
		StartDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Sort:       "-date",
		AsReported: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ss)

	s := ss[0]
	assert.False(t, s.Date.IsZero())
	assert.NotZero(t, s.Year)
	assert.NotEmpty(t, s.StatementData.BalanceSheet)
	assert.NotEmpty(t, s.StatementData.IncomeStatement)
	assert.NotEmpty(t, s.StatementData.CashFlow)
	assert.NotEmpty(t, s.StatementData.Overview)

	rev, ok := s.StatementData.Get(CodeRevenue)
	require.True(t, ok, "revenue 가 있어야 한다")
	assert.Greater(t, rev, 0.0)

	q := lastReq().URL.Query()
	assert.Equal(t, "/tiingo/fundamentals/aapl/statements", lastReq().URL.Path)
	assert.Equal(t, "2026-01-01", q.Get("startDate"))
	assert.Equal(t, "2026-12-31", q.Get("endDate"))
	assert.Equal(t, "-date", q.Get("sort"))
	assert.Equal(t, "true", q.Get("asReported"))
}

func TestStatements_NilOptions_NoQuery(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Statements(context.Background(), "aapl", nil)
	require.NoError(t, err)
	assert.Empty(t, lastReq().URL.RawQuery)
}

func TestStatements_AsReportedFalse_Omitted(t *testing.T) {
	c, lastReq := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Statements(context.Background(), "aapl", &StatementOptions{Sort: "date"})
	require.NoError(t, err)
	q := lastReq().URL.Query()
	assert.Equal(t, "date", q.Get("sort"))
	assert.Empty(t, q.Get("asReported"), "기본값이면 파라미터를 보내지 않는다")
	assert.Empty(t, q.Get("startDate"))
}

func TestStatements_EmptyTicker(t *testing.T) {
	c, _ := newStubClient(t, http.StatusOK, `[]`)
	_, err := c.Statements(context.Background(), "  ", nil)
	assert.Error(t, err)
}

func TestStatementData_Get(t *testing.T) {
	sd := StatementData{
		IncomeStatement: []DataPoint{{DataCode: "revenue", Value: 100}},
		Overview:        []DataPoint{{DataCode: "peRatio", Value: 0}},
	}
	v, ok := sd.Get("revenue")
	assert.True(t, ok)
	assert.Equal(t, 100.0, v)

	v, ok = sd.Get("peRatio")
	assert.True(t, ok, "값이 0 이어도 존재하면 true")
	assert.Equal(t, 0.0, v)

	_, ok = sd.Get("nosuchcode")
	assert.False(t, ok, "없는 코드는 false — 값 0 과 구분된다")

	var empty StatementData
	_, ok = empty.Get("revenue")
	assert.False(t, ok, "nil 슬라이스도 안전하다")
}

func TestStatementData_Map(t *testing.T) {
	sd := StatementData{
		BalanceSheet:    []DataPoint{{DataCode: "totalAssets", Value: 1}},
		IncomeStatement: []DataPoint{{DataCode: "revenue", Value: 2}},
		CashFlow:        []DataPoint{{DataCode: "capex", Value: 3}},
		Overview:        []DataPoint{{DataCode: "bvps", Value: 4}},
	}
	m := sd.Map()
	assert.Len(t, m, 4)
	assert.Equal(t, 2.0, m["revenue"])
	assert.Empty(t, StatementData{}.Map())
}
