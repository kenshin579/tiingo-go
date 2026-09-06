package stream

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnvelope_Info(t *testing.T) {
	raw := `{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":49792869},"messageType":"I"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgInfo, e.MessageType)
	assert.Equal(t, 200, e.Response.Code)
	id, err := e.subscriptionID()
	require.NoError(t, err)
	assert.Equal(t, int64(49792869), id)
}

func TestParseEnvelope_Heartbeat(t *testing.T) {
	raw := `{"response":{"code":200,"message":"HeartBeat"},"messageType":"H"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgHeartbeat, e.MessageType)
	assert.Empty(t, e.Data, "H 에는 data 가 없다")
}

func TestParseEnvelope_Data(t *testing.T) {
	raw := `{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.0]}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgData, e.MessageType)
	assert.Equal(t, "crypto_data", e.Service)
	assert.NotEmpty(t, e.Data, "A 의 data 는 배열 원문으로 남는다")
}

// BOATS 구독 거부가 이 형태로 온다 — 실측.
func TestParseEnvelope_Error(t *testing.T) {
	raw := `{"response":{"code":403,"message":"Not authorized for service: boats"},"messageType":"E"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgError, e.MessageType)

	err = e.asError()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "boats")
}

// 구독 확인이 200 이 아니면 에러로 다룬다.
func TestEnvelope_InfoNonOK(t *testing.T) {
	raw := `{"response":{"code":401,"message":"Not authorized"},"messageType":"I"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	require.Error(t, e.asError())
}

// 정상 구독 확인은 에러가 아니다.
func TestEnvelope_InfoOKIsNotError(t *testing.T) {
	raw := `{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":1},"messageType":"I"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.NoError(t, e.asError())
}

// 데이터 프레임은 response 가 없으므로 에러가 아니다.
func TestEnvelope_DataIsNotError(t *testing.T) {
	raw := `{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54+00:00","x",1,2]}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.NoError(t, e.asError())
}

func TestParseEnvelope_Malformed(t *testing.T) {
	_, err := parseEnvelope([]byte(`{"messageType":`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tiingo:")
}

// 모르는 messageType 은 에러가 아니라 무시 대상이다 — Tiingo 가 종류를 늘릴 수 있다.
func TestParseEnvelope_UnknownType(t *testing.T) {
	e, err := parseEnvelope([]byte(`{"messageType":"Z"}`))
	require.NoError(t, err)
	assert.Equal(t, messageType("Z"), e.MessageType)
	assert.False(t, e.MessageType.known())
}

// 실제 캡처 파일의 모든 줄이 파싱돼야 한다.
func TestParseEnvelope_LiveCapture(t *testing.T) {
	raw, err := os.ReadFile("testdata/crypto_live.jsonl")
	require.NoError(t, err)

	counts := map[messageType]int{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		e, err := parseEnvelope([]byte(line))
		require.NoError(t, err, "실측 캡처는 전부 파싱돼야 한다: %s", line)
		counts[e.MessageType]++
	}
	assert.Positive(t, counts[msgData], "A 가 있어야 한다")
	assert.Positive(t, counts[msgInfo], "I 가 있어야 한다")
}
