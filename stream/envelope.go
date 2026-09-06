package stream

import (
	"encoding/json"
	"fmt"
)

// messageType 은 봉투의 messageType 필드다. Tiingo 가 종류를 늘릴 수 있으므로
// 모르는 값은 에러가 아니라 무시 대상으로 다룬다.
type messageType string

const (
	msgInfo      messageType = "I" // 구독 확인. data 는 객체({"subscriptionId":...})
	msgHeartbeat messageType = "H" // 하트비트. data 없음
	msgData      messageType = "A" // 실제 데이터. data 는 위치 배열
	msgError     messageType = "E" // 에러. data 없음, response 에 사유
)

// known 은 SDK 가 다루는 종류인지 알려준다.
func (m messageType) known() bool {
	switch m {
	case msgInfo, msgHeartbeat, msgData, msgError:
		return true
	}
	return false
}

// envelope 는 모든 WebSocket 메시지의 겉껍질이다.
//
// data 의 타입이 messageType 마다 다르다 — I 는 객체, A 는 배열, H·E 는 없음.
// 그래서 RawMessage 로 받아 두고 종류에 따라 나중에 해석한다.
type envelope struct {
	Service     string          `json:"service"`     // 피드 식별자(crypto_data, cons 등). A 에만 있다
	MessageType messageType     `json:"messageType"` // I/H/A/E
	Data        json.RawMessage `json:"data"`        // 종류에 따라 객체·배열·없음
	Response    struct {
		Code    int    `json:"code"`    // HTTP 유사 상태코드
		Message string `json:"message"` // 사람이 읽는 사유
	} `json:"response"`
}

// parseEnvelope 는 프레임 한 줄을 봉투로 읽는다.
func parseEnvelope(b []byte) (*envelope, error) {
	var e envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("tiingo: malformed websocket frame: %w", err)
	}
	return &e, nil
}

// asError 는 봉투가 실패를 뜻하면 에러를, 아니면 nil 을 돌려준다.
// E 는 물론이고 200 이 아닌 구독 확인(I)도 실패로 다룬다.
// 데이터 프레임(A)은 response 가 없어(code 0) 여기서 걸리지 않는다.
func (e *envelope) asError() error {
	if e.MessageType != msgError && (e.Response.Code == 0 || e.Response.Code == 200) {
		return nil
	}
	return fmt.Errorf("tiingo: stream error (code %d): %s", e.Response.Code, e.Response.Message)
}

// subscriptionID 는 구독 확인(I)에 담긴 id 를 꺼낸다.
func (e *envelope) subscriptionID() (int64, error) {
	var d struct {
		SubscriptionID int64 `json:"subscriptionId"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil {
		return 0, fmt.Errorf("tiingo: malformed subscription ack: %w", err)
	}
	return d.SubscriptionID, nil
}
