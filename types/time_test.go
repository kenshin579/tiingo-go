package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{"밀리초 UTC", `"2026-08-21T01:01:17.444Z"`, time.Date(2026, 8, 21, 1, 1, 17, 444000000, time.UTC)},
		{"초 단위", `"2026-08-21T01:01:17Z"`, time.Date(2026, 8, 21, 1, 1, 17, 0, time.UTC)},
		{"오프셋은 UTC 로 환산", `"2026-08-20T20:01:17-05:00"`, time.Date(2026, 8, 21, 1, 1, 17, 0, time.UTC)},
		{"null", `null`, time.Time{}},
		{"빈 문자열", `""`, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Time
			require.NoError(t, json.Unmarshal([]byte(tt.in), &v))
			assert.True(t, v.Time.Equal(tt.want), "got %v, want %v", v.Time, tt.want)
		})
	}
}

func TestTime_UnmarshalJSON_Invalid(t *testing.T) {
	var v Time
	assert.Error(t, json.Unmarshal([]byte(`"2026-08-21"`), &v), "날짜만 있는 값은 Time 이 아니다")
	assert.Error(t, json.Unmarshal([]byte(`12345`), &v))
}

func TestTime_Marshal(t *testing.T) {
	v := Time{Time: time.Date(2026, 8, 21, 1, 1, 17, 0, time.UTC)}
	b, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Equal(t, `"2026-08-21T01:01:17Z"`, string(b))

	b, err = json.Marshal(Time{})
	require.NoError(t, err)
	assert.Equal(t, `null`, string(b))

	txt, err := v.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "2026-08-21T01:01:17Z", string(txt))

	txt, err = Time{}.MarshalText()
	require.NoError(t, err)
	assert.Empty(t, string(txt))
}

func TestTime_TextRoundTrip(t *testing.T) {
	var v Time
	require.NoError(t, v.UnmarshalText([]byte("2026-08-21T01:01:17.444Z")))
	assert.Equal(t, 444000000, v.Time.Nanosecond(), "시각을 버리지 않는다")
	assert.Equal(t, "2026-08-21T01:01:17.444Z", v.String())
	require.NoError(t, v.UnmarshalText([]byte("")))
	assert.True(t, v.IsZero())
	assert.Error(t, (&Time{}).UnmarshalText([]byte("nope")))
}
