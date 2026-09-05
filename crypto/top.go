package crypto

import (
	"context"
	"strings"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
	"github.com/kenshin579/tiingo-go/types"
)

// TopOfBookData 는 호가창 최상단 한 시점의 스냅샷이다.
type TopOfBookData struct {
	QuoteTimestamp    types.Time `json:"quoteTimestamp"`    // 호가 기준 시각
	LastSaleTimestamp types.Time `json:"lastSaleTimestamp"` // 마지막 체결 시각
	BidSize           float64    `json:"bidSize"`           // 매수 호가 수량
	BidPrice          float64    `json:"bidPrice"`          // 매수 호가
	AskSize           float64    `json:"askSize"`           // 매도 호가 수량
	AskPrice          float64    `json:"askPrice"`          // 매도 호가
	LastSize          float64    `json:"lastSize"`          // 마지막 체결 수량
	LastSizeNotional  float64    `json:"lastSizeNotional"`  // 마지막 체결 대금(표시 통화)
	LastPrice         float64    `json:"lastPrice"`         // 마지막 체결가
	BidExchange       string     `json:"bidExchange"`       // 최우선 매수 호가를 낸 거래소
	AskExchange       string     `json:"askExchange"`       // 최우선 매도 호가를 낸 거래소
	LastExchange      string     `json:"lastExchange"`      // 마지막 체결이 일어난 거래소
}

// TopOfBook 은 티커 하나의 호가 묶음이다. Tiingo 가 티커별로 감싸서 돌려준다.
type TopOfBook struct {
	Ticker        string          `json:"ticker"`        // 페어 티커
	BaseCurrency  string          `json:"baseCurrency"`  // 기준 통화
	QuoteCurrency string          `json:"quoteCurrency"` // 표시 통화
	TopOfBookData []TopOfBookData `json:"topOfBookData"` // 호가 스냅샷. 보통 1건
}

// TopOptions 는 TopOfBook 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type TopOptions struct {
	Exchanges []string // 특정 거래소로 한정. 비어 있으면 전체
	// IncludeRawExchangeData 가 true 면 집계 전 거래소별 원본 호가도 함께 온다.
	// 주의: 2026-09-05 기준 Tiingo 가 이 값에 HTTP 500 을 돌려준다(서버 쪽 문제).
	IncludeRawExchangeData bool
}

// apply 는 옵션을 쿼리 맵에 채워 넣는다. zero 값은 넣지 않는다.
func (o *TopOptions) apply(q map[string]string) {
	if o == nil {
		return
	}
	if len(o.Exchanges) > 0 {
		q["exchanges"] = strings.Join(o.Exchanges, ",")
	}
	if o.IncludeRawExchangeData {
		q["includeRawExchangeData"] = "true" // 기본값(false)일 때는 보내지 않는다
	}
}

// TopOfBook 은 티커별 호가 묶음을 받는다. GET /tiingo/crypto/top
// tickers 는 하나 이상이어야 한다.
func (c *Client) TopOfBook(ctx context.Context, tickers []string, opts *TopOptions) ([]TopOfBook, error) {
	joined, err := joinTickers(tickers, true)
	if err != nil {
		return nil, err
	}
	q := map[string]string{"tickers": joined}
	opts.apply(q)

	var ts []TopOfBook
	if err := c.http.GetJSON(ctx, "/tiingo/crypto/top", q, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// TopOfBookFor 는 티커 하나의 호가 묶음을 받는다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) TopOfBookFor(ctx context.Context, ticker string, opts *TopOptions) (*TopOfBook, error) {
	ts, err := c.TopOfBook(ctx, []string{ticker}, opts)
	if err != nil {
		return nil, err
	}
	if len(ts) == 0 {
		return nil, httpclient.ErrNotFound
	}
	return &ts[0], nil
}
