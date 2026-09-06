package stream

import (
	"encoding/json"

	"github.com/kenshin579/tiingo-go/types"
)

// IEXThreshold 는 IEX 피드의 구독 수준이다.
type IEXThreshold int

const (
	IEXReferencePriceLevel IEXThreshold = 6 // 기준가만(IEX 는 이 레벨 하나뿐)
)

// IEXOptions 는 IEX 스트림 구독 인자다.
type IEXOptions struct {
	Tickers   []string     // 구독할 티커. 비우면 전체
	Threshold IEXThreshold // 0 이면 IEXReferencePriceLevel(6)
}

// IEXReferencePrice 는 IEX 기준가다. 종류 문자가 없다.
//
// 배열: [date, ticker, refPrice] (3개)
// 문서 기반이며 실호출로 검증하지 못했다 — 캡처 당시 미국 장이 닫혀 있었다.
type IEXReferencePrice struct {
	Date           types.Time // 갱신 시각
	Ticker         string     // 티커. 대문자로 온다
	ReferencePrice float64    // 기준가
}

func (IEXReferencePrice) isMessage() {}

// decodeIEX 는 IEX data 배열을 메시지로 바꾼다. 종류 문자가 없어 길이만 확인한다.
func decodeIEX(raw json.RawMessage) (Message, error) {
	a, err := newArr(raw, 3)
	if err != nil {
		return nil, err
	}
	m := IEXReferencePrice{
		Date:           a.time(0),
		Ticker:         a.str(1),
		ReferencePrice: a.f64(2),
	}
	return m, a.err()
}
