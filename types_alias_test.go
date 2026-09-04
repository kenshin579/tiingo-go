package tiingo

import (
	"encoding/json"
	"testing"

	"github.com/kenshin579/tiingo-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateAlias(t *testing.T) {
	var d Date
	require.NoError(t, json.Unmarshal([]byte(`"2019-01-02T00:00:00.000Z"`), &d))
	assert.Equal(t, "2019-01-02", d.String())
	assert.Equal(t, DateLayout, types.DateLayout)

	var _ types.Date = d // 별칭이므로 같은 타입
}
