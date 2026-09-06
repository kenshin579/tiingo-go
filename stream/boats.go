package stream

import (
	"encoding/json"

	"github.com/kenshin579/tiingo-go/types"
)

// BOATSThreshold 는 BOATS 피드의 구독 수준이다.
type BOATSThreshold int

const (
	BOATSTopOfBookAndTradesLevel BOATSThreshold = 3 // BOATS 가 제공하는 유일한 공개 레벨(호가+체결). 문서: thresholdLevel 3
)

// BOATSOptions 는 BOATS 스트림 구독 인자다.
type BOATSOptions struct {
	Tickers   []string       // 구독할 티커. 비우면 전체
	Threshold BOATSThreshold // 0 이면 BOATSTopOfBookAndTradesLevel(6)
}

// BOATSQuote 는 BOATS 호가다.
//
// 배열: ["Q", date, nanos, ticker, bidSize, bidPrice, midPrice, askPrice, askSize] (9개)
//
// 계정에 BOATS 엔타이틀먼트가 없어(403) 실호출 검증이 불가능하다. 문서 예시로만
// 검증했다 — mid = (81.58+81.59)/2 = 81.585 가 이 인덱스 순서에서 성립한다. IEX 계열과
// 같이 매도가가 매도 수량보다 먼저 온다(askPrice=7, askSize=8).
type BOATSQuote struct {
	Date        types.Time // 문자열 타임스탬프
	Nanoseconds int64      // 유닉스 나노초 타임스탬프(중복 정보, 문서 그대로 보존)
	Ticker      string     // 티커. 소문자로 온다
	BidSize     float64    // 매수 수량
	BidPrice    float64    // 매수 호가
	MidPrice    *float64   // 중간가. 한쪽 호가가 없으면 null(문서 명시)
	AskPrice    float64    // 매도 호가
	AskSize     float64    // 매도 수량
}

func (BOATSQuote) isMessage() {}

// BOATSTrade 는 BOATS 체결이다. 종류 문자가 "T"(정상 체결)와 "B"(체결 취소) 둘 다 이 타입으로 온다.
//
// 배열: [kind, date, nanos, ticker, lastPrice, lastSize, cond1..cond4] (10개)
// kind 는 "T" 또는 "B". 판매 조건 4개는 빈 문자열일 수 있다.
//
// 계정에 BOATS 엔타이틀먼트가 없어(403) 실호출 검증이 불가능하다. 문서 예시로만 확인했다.
type BOATSTrade struct {
	Break          bool       // true 면 체결 취소(kind=="B")
	Date           types.Time // 문자열 타임스탬프
	Nanoseconds    int64      // 유닉스 나노초 타임스탬프
	Ticker         string     // 티커
	LastPrice      float64    // 체결가
	LastSize       float64    // 체결 수량
	SaleConditions [4]string  // 판매 조건 4개(인덱스 6~9, 고정 슬롯). 미사용 칸은 빈 문자열
}

func (BOATSTrade) isMessage() {}

// decodeBOATS 는 BOATS data 배열을 메시지로 바꾼다.
// 모르는 종류 문자면 (nil, nil) 을 돌려준다.
func decodeBOATS(raw json.RawMessage) (Message, error) {
	kind, err := arrKind(raw)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "Q":
		a, err := newArr(raw, 9)
		if err != nil {
			return nil, err
		}
		m := BOATSQuote{
			Date:        a.time(1),
			Nanoseconds: a.i64(2),
			Ticker:      a.str(3),
			BidSize:     a.f64(4),
			BidPrice:    a.f64(5),
			MidPrice:    a.f64p(6),
			AskPrice:    a.f64(7),
			AskSize:     a.f64(8),
		}
		return m, a.err()
	case "T", "B":
		a, err := newArr(raw, 10)
		if err != nil {
			return nil, err
		}
		m := BOATSTrade{
			Break:       kind == "B",
			Date:        a.time(1),
			Nanoseconds: a.i64(2),
			Ticker:      a.str(3),
			LastPrice:   a.f64(4),
			LastSize:    a.f64(5),
			SaleConditions: [4]string{
				a.str(6), a.str(7), a.str(8), a.str(9),
			},
		}
		return m, a.err()
	}
	return nil, nil
}
