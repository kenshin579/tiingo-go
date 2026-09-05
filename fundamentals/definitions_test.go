package fundamentals

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefinitions_Fixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/definitions.json")
	require.NoError(t, err)

	c, lastReq := newStubClient(t, http.StatusOK, string(raw))
	ds, err := c.Definitions(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, ds)

	byCode := map[string]Definition{}
	for _, d := range ds {
		byCode[d.DataCode] = d
	}
	rev, ok := byCode["revenue"]
	require.True(t, ok, "revenue 정의가 있어야 한다")
	assert.NotEmpty(t, rev.Name)
	assert.Equal(t, "incomeStatement", rev.StatementType)

	counts := map[string]int{}
	for _, d := range ds {
		counts[d.StatementType]++
	}
	for _, st := range []string{"balanceSheet", "incomeStatement", "cashFlow", "overview"} {
		assert.NotZero(t, counts[st], "%s 지표가 있어야 한다", st)
	}

	assert.Equal(t, "/tiingo/fundamentals/definitions", lastReq().URL.Path)
	assert.Empty(t, lastReq().URL.RawQuery)
}
