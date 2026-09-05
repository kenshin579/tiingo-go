package forex

import (
	"context"

	"github.com/kenshin579/tiingo-go/types"
)

// Quote 는 통화쌍 호가 스냅샷이다. 체결 정보 없이 매수·매도·중간가만 있다.
type Quote struct {
	Ticker         string     `json:"ticker"`         // 통화쌍(예: eurusd)
	QuoteTimestamp types.Time `json:"quoteTimestamp"` // 호가 기준 시각
	BidPrice       float64    `json:"bidPrice"`       // 매수 호가
	BidSize        float64    `json:"bidSize"`        // 매수 호가 수량
	AskPrice       float64    `json:"askPrice"`       // 매도 호가
	AskSize        float64    `json:"askSize"`        // 매도 호가 수량
	MidPrice       float64    `json:"midPrice"`       // 중간가
}

// TopOfBook 은 통화쌍 호가를 받는다. GET /tiingo/fx/top
// tickers 는 하나 이상이어야 한다 — 없이 호출하면 Tiingo 가 400 을 돌려준다.
func (c *Client) TopOfBook(ctx context.Context, tickers []string) ([]Quote, error) {
	joined, err := joinTickers(tickers)
	if err != nil {
		return nil, err
	}
	var qs []Quote
	if err := c.http.GetJSON(ctx, "/tiingo/fx/top", map[string]string{"tickers": joined}, &qs); err != nil {
		return nil, err
	}
	return qs, nil
}
