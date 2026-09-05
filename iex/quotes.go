package iex

import (
	"context"
	"fmt"
	"strings"

	"github.com/kenshin579/tiingo-go/types"
)

// Quote 는 IEX 실시간 스냅샷이다. 호가·최종체결에 더해 당일 시가·고가·저가와 전일 종가까지 담는다.
// 장 마감 시간대에는 호가·체결 관련 9개 필드가 모두 null 로 오므로 포인터다 — nil 은 "값 없음"이고
// 0 과 구분된다.
type Quote struct {
	Ticker            string      `json:"ticker"`            // 티커. 요청과 무관하게 대문자로 온다
	Timestamp         types.Time  `json:"timestamp"`         // 스냅샷 기준 시각
	LastSaleTimestamp *types.Time `json:"lastSaleTimestamp"` // 마지막 체결 시각. 장 마감 시 nil
	QuoteTimestamp    *types.Time `json:"quoteTimestamp"`    // 호가 기준 시각. 장 마감 시 nil
	Open              float64     `json:"open"`              // 당일 시가
	High              float64     `json:"high"`              // 당일 고가
	Low               float64     `json:"low"`               // 당일 저가
	Mid               *float64    `json:"mid"`               // 중간가. 장 마감 시 nil
	TngoLast          float64     `json:"tngoLast"`          // Tiingo 기준 최종가. 장 마감 후에도 채워진다
	Last              *float64    `json:"last"`              // IEX 최종 체결가. 장 마감 시 nil
	// LastSize/BidSize/AskSize 는 *float64 다. Tiingo 가 프라이스 엔드포인트에서
	// 수량을 "38607.0" 처럼 소수점을 붙여 보내는 사례가 있어(prices.go 참고),
	// 이 스냅샷 필드들도 같은 계열이라 int64 로 두면 장중 응답에서 디코딩이 깨질 위험이 있다.
	LastSize  *float64 `json:"lastSize"`  // 마지막 체결 수량. 장 마감 시 nil
	BidSize   *float64 `json:"bidSize"`   // 매수 호가 수량. 장 마감 시 nil
	BidPrice  *float64 `json:"bidPrice"`  // 매수 호가. 장 마감 시 nil
	AskPrice  *float64 `json:"askPrice"`  // 매도 호가. 장 마감 시 nil
	AskSize   *float64 `json:"askSize"`   // 매도 호가 수량. 장 마감 시 nil
	Volume    int64    `json:"volume"`    // 당일 누적 거래량
	PrevClose float64  `json:"prevClose"` // 전일 종가
}

// joinTickers 는 티커 목록을 검증해 콤마로 합친다. 공백뿐인 원소나 빈 목록은 에러다.
func joinTickers(tickers []string) (string, error) {
	cleaned := make([]string, 0, len(tickers))
	for _, t := range tickers {
		t = strings.TrimSpace(t)
		if t == "" {
			return "", fmt.Errorf("tiingo: ticker must not be empty")
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return "", fmt.Errorf("tiingo: at least one ticker is required")
	}
	return strings.Join(cleaned, ","), nil
}

// Quotes 는 IEX 실시간 스냅샷을 받는다. GET /iex/
// 경로 끝 슬래시가 필요하다 — /iex 는 301 이다.
// 없는 티커는 에러가 아니라 응답에서 빠지고 순서도 요청과 다를 수 있으므로,
// 결과를 인덱스로 대응시키지 말고 Ticker 필드로 찾는다.
func (c *Client) Quotes(ctx context.Context, tickers []string) ([]Quote, error) {
	joined, err := joinTickers(tickers)
	if err != nil {
		return nil, err
	}
	var qs []Quote
	if err := c.http.GetJSON(ctx, "/iex/", map[string]string{"tickers": joined}, &qs); err != nil {
		return nil, err
	}
	return qs, nil
}
