package eod

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/kenshin579/tiingo-go/types"
)

// Meta 는 자산의 메타 정보다(EOD 메타 엔드포인트 응답).
type Meta struct {
	Ticker       string      `json:"ticker"`       // 자산의 티커. 주식 클래스는 마침표 대신 대시(예: BRK-A)
	Name         string      `json:"name"`         // 자산의 전체 이름
	ExchangeCode string      `json:"exchangeCode"` // 상장 거래소 식별자(예: NASDAQ)
	Description  string      `json:"description"`  // 자산에 대한 장문 설명
	StartDate    *types.Date `json:"startDate"`    // 가격 데이터가 있는 가장 이른 날짜. nil 이거나 IsZero() 면 가격 데이터 없음
	EndDate      *types.Date `json:"endDate"`      // 가격 데이터가 있는 가장 늦은 날짜. nil 이거나 IsZero() 면 가격 데이터 없음
}

// Meta 는 자산의 메타 정보를 조회한다. GET /tiingo/daily/<ticker>
func (c *Client) Meta(ctx context.Context, ticker string) (*Meta, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var m Meta
	if err := c.http.GetJSON(ctx, "/tiingo/daily/"+url.PathEscape(t), nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
