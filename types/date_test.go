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
		{"빈 문자열", `""`, time.Time{}},
		{"RFC3339 음수 오프셋", `"2019-01-02T19:00:00-05:00"`, time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)}, // 표기된 벽시계 날짜를 취한다(UTC 환산 시 하루 밀림)
		{"RFC3339 양수 오프셋", `"2019-01-02T02:00:00+09:00"`, time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC)}, // 반대 방향도 마찬가지
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

func TestDate_TextRoundTrip(t *testing.T) {
	var d Date
	require.NoError(t, d.UnmarshalText([]byte("1980-12-12")))
	assert.Equal(t, "1980-12-12", d.String())
	require.NoError(t, d.UnmarshalText([]byte("2019-01-02T00:00:00.000Z")))
	b, err := d.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "2019-01-02", string(b))

	m, err := json.Marshal(map[Date]int{d: 1})
	require.NoError(t, err)
	assert.JSONEq(t, `{"2019-01-02":1}`, string(m), "map 키도 YYYY-MM-DD 여야 한다")

	assert.Error(t, (&Date{}).UnmarshalText([]byte("2019/01/02")))
}

func TestDate_StructFields(t *testing.T) {
	var s struct {
		A *Date `json:"a"`
		B Date  `json:"b"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"a":null,"b":"2019-01-02T00:00:00.000Z"}`), &s))
	assert.Nil(t, s.A, "null 은 nil 포인터")
	assert.Equal(t, "2019-01-02", s.B.String())
}
