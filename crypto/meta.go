package crypto

import "context"

// Meta 는 암호화폐 페어의 메타 정보다.
type Meta struct {
	Ticker        string `json:"ticker"`        // 페어 티커(예: btcusd)
	Name          string `json:"name"`          // 사람이 읽는 이름(예: Bitcoin (BTC/USD))
	BaseCurrency  string `json:"baseCurrency"`  // 기준 통화(예: btc)
	QuoteCurrency string `json:"quoteCurrency"` // 표시 통화(예: usd)
}

// Meta 는 페어 메타 정보를 받는다. GET /tiingo/crypto
// tickers 를 주지 않으면 지원하는 페어 전체를 받는다 — 대량 응답이므로 주의한다.
func (c *Client) Meta(ctx context.Context, tickers ...string) ([]Meta, error) {
	var params map[string]string
	if len(tickers) > 0 {
		joined, err := joinTickers(tickers, false)
		if err != nil {
			return nil, err
		}
		params = map[string]string{"tickers": joined}
	}
	var ms []Meta
	if err := c.http.GetJSON(ctx, "/tiingo/crypto", params, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}
