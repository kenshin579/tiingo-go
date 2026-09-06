package stream

import (
	"encoding/json"
	"fmt"

	"github.com/kenshin579/tiingo-go/types"
)

// EquityThreshold 는 equity 통합 피드의 구독 수준이다.
type EquityThreshold int

const (
	EquityLiquidityLevel      EquityThreshold = 4 // 기준가 + 유동성
	EquityReferencePriceLevel EquityThreshold = 6 // 기준가만
)

// EquityOptions 는 equity 통합 피드 구독 인자다.
type EquityOptions struct {
	Tickers   []string        // 구독할 티커. 비우면 전체
	Threshold EquityThreshold // 0 이면 EquityReferencePriceLevel(6)
}

// EquityReferencePrice 는 equity 통합 피드 기준가다. 종류 문자가 없다.
//
// 배열: [date, ticker, refPrice] (3개)
// 문서 기반이며 실호출로 검증하지 못했다 — 캡처 당시 미국 장이 닫혀 있었다.
type EquityReferencePrice struct {
	Date           types.Time // 갱신 시각
	Ticker         string     // 티커. 소문자로 온다
	ReferencePrice float64    // 기준가
}

func (EquityReferencePrice) isMessage() {}

// EquityLiquidity 는 equity 통합 피드 유동성 스냅샷이다. 종류 문자가 없다.
//
// 배열: [date, ticker, spread, bidSize, bidPrice, refPrice, askPrice, askSize] (8개)
// 문서 기반이며 실호출로 검증하지 못했다(미국 장 닫힘). 다만 문서 예시의 값 모양은 표와
// 일치한다 — 인덱스 4/5/6(162.30/162.33/162.36)은 나란한 가격대이고 3/7(둘 다 100)은
// 수량이다. IEX 계열은 매도 수량보다 매도가가 먼저 온다(Crypto·Forex 와 반대 순서).
type EquityLiquidity struct {
	Date           types.Time // 갱신 시각
	Ticker         string     // 티커
	Spread         float64    // 상대 스프레드. 0.0011 이면 0.11%
	BidSize        float64    // 매수 수량
	BidPrice       float64    // 매수 호가
	ReferencePrice float64    // 기준가
	AskPrice       float64    // 매도 호가
	AskSize        float64    // 매도 수량
}

func (EquityLiquidity) isMessage() {}

// decodeEquity 는 equity data 배열을 메시지로 바꾼다. 종류 문자가 없어 길이로 구분한다 —
// 3 이면 기준가, 8 이면 유동성. 그 외 길이는 에러다: 종류 문자가 없는 상태에서 4~7 이나
// 9 이상은 기준가가 늘어난 것인지 유동성이 늘어난 것인지 판단할 근거가 없다.
//
// 남는 위험이 하나 있다: 기준가 메시지가 정확히 8원소로 넓어지면 유동성으로 잘못 읽히고,
// 종류 문자가 없어 이를 잡을 방법이 없다. Tiingo 가 이 피드의 배열을 넓히면 이 함수를 먼저 본다.
func decodeEquity(raw json.RawMessage) (Message, error) {
	var xs []json.RawMessage
	if err := json.Unmarshal(raw, &xs); err != nil {
		return nil, fmt.Errorf("tiingo: data must be a JSON array: %w", err)
	}
	switch len(xs) {
	case 3:
		a := arrFrom(xs)
		m := EquityReferencePrice{
			Date:           a.time(0),
			Ticker:         a.str(1),
			ReferencePrice: a.f64(2),
		}
		return m, a.err()
	case 8:
		a := arrFrom(xs)
		m := EquityLiquidity{
			Date:           a.time(0),
			Ticker:         a.str(1),
			Spread:         a.f64(2),
			BidSize:        a.f64(3),
			BidPrice:       a.f64(4),
			ReferencePrice: a.f64(5),
			AskPrice:       a.f64(6),
			AskSize:        a.f64(7),
		}
		return m, a.err()
	}
	return nil, fmt.Errorf("tiingo: equity data array has unexpected length %d", len(xs))
}
