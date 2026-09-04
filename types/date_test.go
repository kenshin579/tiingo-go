package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"RFC3339 밀리초 UTC", `"2019-01-02T00:00:00.000Z"`, time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"RFC3339 오프셋", `"2019-01-02T00:00:00+00:00"`, time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"날짜만", `"1980-12-12"`, time.Date(1980, 12, 12, 0, 0, 0, 0, time.UTC)},
		{"null", `null`, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			require.NoError(t, json.Unmarshal([]byte(tt.in), &d))
			assert.True(t, d.Time.Equal(tt.want), "got %v, want %v", d.Time, tt.want)
		})
	}
}

func TestDate_UnmarshalJSON_Invalid(t *testing.T) {
	var d Date
	assert.Error(t, json.Unmarshal([]byte(`"2019/01/02"`), &d))
	assert.Error(t, json.Unmarshal([]byte(`12345`), &d))
}

func TestDate_MarshalJSON(t *testing.T) {
	d := Date{Time: time.Date(2019, 1, 2, 15, 4, 5, 0, time.UTC)}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, `"2019-01-02"`, string(b))

	b, err = json.Marshal(Date{})
	require.NoError(t, err)
	assert.Equal(t, `null`, string(b))
}

func TestDate_String(t *testing.T) {
	assert.Equal(t, "2019-01-02", Date{Time: time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)}.String())
	assert.Equal(t, "", Date{}.String())
}
