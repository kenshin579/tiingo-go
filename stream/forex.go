package stream

import (
	"encoding/json"

	"github.com/kenshin579/tiingo-go/types"
)

// ForexThreshold 는 forex 피드의 구독 수준이다.
type ForexThreshold int

const (
	ForexTopOfBookLevel ForexThreshold = 5 // 호가(top-of-book) 전부. 이 피드에는 체결 메시지가 없다
)

// ForexOptions 는 forex 스트림 구독 인자다.
type ForexOptions struct {
	Tickers   []string       // 구독할 통화쌍(eurusd 등). 비우면 전체
	Threshold ForexThreshold // 0 이면 ForexTopOfBookLevel(5)
}

// ForexQuote 는 외환 호가다. Forex 피드는 호가만 보내고 체결은 없다.
//
// 배열: ["Q", ticker, date, bidSize, bidPrice, midPrice, askSize, askPrice] (8개)
//
// 이 매핑은 실호출이 아니라 문서 예시(eurnok)의 산술로 검증했다 —
// mid = (bidPrice+askPrice)/2 = (9.6764+9.67987)/2 = 9.678135 가 인덱스 7 을 가격으로 볼 때만 성립한다.
// Tiingo 문서의 인덱스 표는 askPrice=6/askSize=7 이라 적어 놓았지만 틀렸다: 인덱스 6 의 값
// 5000000.0 은 가격이 아니라 수량이다. 실호출 검증은 아직이다 — 캡처 당시 FX 장이 닫혀 있었다.
type ForexQuote struct {
	Ticker   string     // 통화쌍. 소문자로 온다
	Date     types.Time // 호가 시각
	BidSize  float64    // 매수 수량
	BidPrice float64    // 매수 호가
	// MidPrice 는 중간가. 문서 예시에 값이 있고 null 명시가 없어 값 타입이다(BOATS 와 다름).
	// 라이브 미검증.
	MidPrice float64
	AskSize  float64 // 매도 수량
	AskPrice float64 // 매도 호가
}

func (ForexQuote) isMessage() {}

// decodeForex 는 forex data 배열을 메시지로 바꾼다.
// 모르는 종류 문자면 (nil, nil) 을 돌려준다.
func decodeForex(raw json.RawMessage) (Message, error) {
	kind, err := arrKind(raw)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "Q":
		a, err := newArr(raw, 8)
		if err != nil {
			return nil, err
		}
		m := ForexQuote{
			Ticker:   a.str(1),
			Date:     a.time(2),
			BidSize:  a.f64(3),
			BidPrice: a.f64(4),
			MidPrice: a.f64(5),
			AskSize:  a.f64(6),
			AskPrice: a.f64(7),
		}
		return m, a.err()
	}
	return nil, nil
}
