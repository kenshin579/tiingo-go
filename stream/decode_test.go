package stream

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustArr(t *testing.T, raw string, want int) *arr {
	t.Helper()
	a, err := newArr(json.RawMessage(raw), want)
	require.NoError(t, err)
	return a
}

func TestArr_Extract(t *testing.T) {
	a := mustArr(t, `["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.0]`, 6)
	assert.Equal(t, "T", a.str(0))
	assert.Equal(t, "btcusd", a.str(1))
	assert.False(t, a.time(2).IsZero())
	assert.Equal(t, "bitfinex", a.str(3))
	assert.InDelta(t, 0.0002, a.f64(4), 1e-12)
	assert.InDelta(t, 80132.0, a.f64(5), 1e-9)
	require.NoError(t, a.err())
}

// 지수 표기도 받아야 한다 — 실측에서 6.517e-5 로 온다.
func TestArr_ScientificNotation(t *testing.T) {
	a := mustArr(t, `["Q","btcusd","2026-09-06T03:56:52.172908+00:00","gdax",6.517e-5,80046.4,80046.405,0.63361423,80046.41]`, 9)
	assert.InDelta(t, 6.517e-5, a.f64(4), 1e-12)
	require.NoError(t, a.err())
}

// 짧은 배열은 에러다 — 매핑이 밀리면 값이 조용히 바뀌므로 여기서 막는다.
func TestArr_TooShort(t *testing.T) {
	_, err := newArr(json.RawMessage(`["T","btcusd","2026-09-06T03:55:54+00:00","bitfinex",0.0002]`), 6)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "6")
	assert.Contains(t, err.Error(), "5")
}

// 긴 배열은 에러가 아니다 — Tiingo 가 필드를 추가할 수 있고, 뒤에 붙는 건 무시하면 된다.
func TestArr_LongerIsOK(t *testing.T) {
	a, err := newArr(json.RawMessage(`["T","btcusd","2026-09-06T03:55:54+00:00","bitfinex",0.0002,80132.0,"extra"]`), 6)
	require.NoError(t, err)
	assert.Equal(t, "btcusd", a.str(1))
	require.NoError(t, a.err())
}

// 타입이 다르면 에러다.
func TestArr_WrongType(t *testing.T) {
	a := mustArr(t, `["T","btcusd","2026-09-06T03:55:54+00:00","bitfinex","not-a-number",80132.0]`, 6)
	a.f64(4)
	require.Error(t, a.err())
	assert.Contains(t, a.err().Error(), "index 4")
}

// 첫 에러가 보존된다 — 뒤에서 또 실패해도 원인을 잃지 않는다.
func TestArr_FirstErrorWins(t *testing.T) {
	a := mustArr(t, `["T","btcusd","bad-time","bitfinex","x","y"]`, 6)
	a.time(2)
	first := a.err()
	require.Error(t, first)
	a.f64(4)
	assert.Equal(t, first, a.err(), "첫 에러가 유지된다")
}

func TestArr_NotAnArray(t *testing.T) {
	_, err := newArr(json.RawMessage(`{"a":1}`), 3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tiingo:")
}

// null 은 포인터 접근자에서만 허용한다.
func TestArr_NullHandling(t *testing.T) {
	a := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1,"AAPL",null]`, 5)
	assert.Nil(t, a.f64p(4))
	require.NoError(t, a.err())

	b := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1,"AAPL",null]`, 5)
	b.f64(4)
	require.Error(t, b.err(), "값 타입 접근자는 null 을 거부한다")
}

// 포인터 접근자도 값이 있으면 그대로 읽는다.
func TestArr_PointerWithValue(t *testing.T) {
	a := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1,"AAPL",81.585]`, 5)
	p := a.f64p(4)
	require.NotNil(t, p)
	assert.InDelta(t, 81.585, *p, 1e-9)
	require.NoError(t, a.err())
}

func TestArr_Int64(t *testing.T) {
	a := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1757165400000000000,"AAPL"]`, 4)
	assert.Equal(t, int64(1757165400000000000), a.i64(2))
	require.NoError(t, a.err())
}

// 실측 캡처의 모든 데이터 배열이 길이 검증을 통과해야 한다.
func TestArr_LiveCaptureLengths(t *testing.T) {
	raw, err := os.ReadFile("testdata/crypto_live.jsonl")
	require.NoError(t, err)

	var trades, quotes int
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		e, err := parseEnvelope([]byte(line))
		require.NoError(t, err)
		if e.MessageType != msgData {
			continue
		}
		kind := mustArr(t, string(e.Data), 1).str(0)
		switch kind {
		case "T":
			a, err := newArr(e.Data, 6)
			require.NoError(t, err, "체결은 6원소 이상이다")
			require.NoError(t, a.err())
			trades++
		case "Q":
			a, err := newArr(e.Data, 9)
			require.NoError(t, err, "호가는 9원소 이상이다")
			// mid = (bid+ask)/2 — 인덱스 매핑이 맞다는 증거다.
			bid, mid, ask := a.f64(5), a.f64(6), a.f64(8)
			require.NoError(t, a.err())
			assert.InDelta(t, (bid+ask)/2, mid, 1e-6)
			quotes++
		}
	}
	assert.Positive(t, trades)
	assert.Positive(t, quotes)
}
