package search

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tiingo 는 값이 없을 때 null 과 문자열 "nan" 두 가지를 보낸다. 둘 다 빈 문자열이어야 한다.
func TestFIGI_Unmarshal(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want FIGI
	}{
		{"정상 값", `"BBG000B9XRY4"`, "BBG000B9XRY4"},
		{"null 은 빈 문자열", `null`, ""},
		{"nan 은 빈 문자열", `"nan"`, ""},
		{"NaN 도 빈 문자열", `"NaN"`, ""},
		{"빈 문자열 그대로", `""`, ""},
		{"공백은 잘라낸다", `"  BBG000B9XRY4  "`, "BBG000B9XRY4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FIGI
			require.NoError(t, json.Unmarshal([]byte(tt.raw), &f))
			assert.Equal(t, tt.want, f)
		})
	}
}

// 숫자·객체처럼 문자열이 아닌 값은 에러여야 한다 — 조용히 삼키면 스키마 변경을 놓친다.
func TestFIGI_UnmarshalRejectsNonString(t *testing.T) {
	var f FIGI
	assert.Error(t, json.Unmarshal([]byte(`123`), &f))
}

// 에러 메시지는 tiingo: 접두어를 달아 다른 패키지와 맞춘다.
func TestFIGI_UnmarshalErrorIsWrapped(t *testing.T) {
	var f FIGI
	err := json.Unmarshal([]byte(`123`), &f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tiingo: figi must be a JSON string")
}

// 구조체 필드로 쓰일 때도 동작해야 한다.
func TestFIGI_InStruct(t *testing.T) {
	var v struct {
		F FIGI `json:"f"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"f":"nan"}`), &v))
	assert.Equal(t, FIGI(""), v.F)
	assert.True(t, v.F == "", "기저 타입이 string 이라 문자열과 직접 비교된다")
}
