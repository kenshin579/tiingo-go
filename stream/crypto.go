package stream

import (
	"encoding/json"
	"fmt"

	"github.com/kenshin579/tiingo-go/types"
)

// CryptoThreshold 는 crypto 피드의 구독 수준이다.
type CryptoThreshold int

const (
	CryptoTradesAndQuotes CryptoThreshold = 2 // 호가 + 체결
	CryptoTradesOnly      CryptoThreshold = 5 // 체결만
)

// CryptoOptions 는 crypto 스트림 구독 인자다.
type CryptoOptions struct {
	Tickers   []string        // 구독할 페어(btcusd 등). 비우면 전체
	Threshold CryptoThreshold // 0 이면 CryptoTradesOnly(5)
}

// CryptoTrade 는 암호화폐 체결이다.
//
// 배열: ["T", ticker, date, exchange, lastSize, lastPrice] (6개)
// 이 매핑은 실호출로 검증했다(2026-09-06, 체결 11건).
type CryptoTrade struct {
	Ticker    string     // 페어. 소문자로 온다
	Date      types.Time // 체결 시각
	Exchange  string     // 체결 거래소(bitfinex, gdax, bitstamp 등)
	LastSize  float64    // 기준 통화 기준 체결 수량
	LastPrice float64    // 체결가
}

func (CryptoTrade) isMessage() {}

// CryptoQuote 는 암호화폐 호가다. Threshold 가 CryptoTradesAndQuotes 일 때만 온다.
//
// 배열: ["Q", ticker, date, exchange, bidSize, bidPrice, midPrice, askSize, askPrice] (9개)
// 이 매핑은 실호출로 검증했다(2026-09-06, 호가 18건, mid=(bid+ask)/2 검산 일치).
type CryptoQuote struct {
	Ticker   string     // 페어
	Date     types.Time // 호가 시각
	Exchange string     // 거래소
	BidSize  float64    // 매수 수량
	BidPrice float64    // 매수 호가
	// MidPrice 는 중간가. (BidPrice+AskPrice)/2. 실측 18건 모두 값이 있었고 문서도 null 을
	// 명시하지 않아 값 타입이다(BOATS 와 다름).
	MidPrice float64
	AskSize  float64 // 매도 수량
	AskPrice float64 // 매도 호가
}

func (CryptoQuote) isMessage() {}

// decodeCrypto 는 crypto data 배열을 메시지로 바꾼다.
// 모르는 종류 문자면 (nil, nil) 을 돌려준다 — 무시하고 넘어가라는 뜻이다.
func decodeCrypto(raw json.RawMessage) (Message, error) {
	kind, err := arrKind(raw)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "T":
		a, err := newArr(raw, 6)
		if err != nil {
			return nil, err
		}
		m := CryptoTrade{
			Ticker:    a.str(1),
			Date:      a.time(2),
			Exchange:  a.str(3),
			LastSize:  a.f64(4),
			LastPrice: a.f64(5),
		}
		return m, a.err()
	case "Q":
		a, err := newArr(raw, 9)
		if err != nil {
			return nil, err
		}
		m := CryptoQuote{
			Ticker:   a.str(1),
			Date:     a.time(2),
			Exchange: a.str(3),
			BidSize:  a.f64(4),
			BidPrice: a.f64(5),
			MidPrice: a.f64(6),
			AskSize:  a.f64(7),
			AskPrice: a.f64(8),
		}
		return m, a.err()
	}
	return nil, nil
}

// arrKind 는 배열의 첫 원소(종류 문자)를 읽는다.
// 종류 문자가 없는 피드(IEX·Equity)에서는 쓰지 않는다.
func arrKind(raw json.RawMessage) (string, error) {
	var xs []json.RawMessage
	if err := json.Unmarshal(raw, &xs); err != nil {
		return "", fmt.Errorf("tiingo: data must be a JSON array: %w", err)
	}
	if len(xs) == 0 {
		return "", fmt.Errorf("tiingo: data array is empty")
	}
	var k string
	if err := json.Unmarshal(xs[0], &k); err != nil {
		return "", fmt.Errorf("tiingo: data index 0 (kind): %w", err)
	}
	return k, nil
}
