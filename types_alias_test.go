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

func TestTimeAlias(t *testing.T) {
	var v Time
	require.NoError(t, json.Unmarshal([]byte(`"2026-08-21T01:01:17.444Z"`), &v))
	assert.Equal(t, 17, v.Second(), "시각이 보존된다")
	assert.Equal(t, "2026-08-21T01:01:17.444Z", v.String())

	var _ types.Time = v // 별칭이므로 같은 타입
}
