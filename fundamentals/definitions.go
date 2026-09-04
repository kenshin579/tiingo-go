package fundamentals

import "context"

// Definition 은 지표 하나의 정의다. Definitions 로 전체 목록을 받는다.
type Definition struct {
	DataCode      string `json:"dataCode"`      // 지표 식별자(예: peRatio). codes.go 상수와 같은 값
	Name          string `json:"name"`          // 사람이 읽는 이름(예: Revenue Per Share)
	Description   string `json:"description"`   // 지표 설명
	StatementType string `json:"statementType"` // 소속 묶음: balanceSheet / incomeStatement / cashFlow / overview
	Units         string `json:"units"`         // 단위. "$" 금액, "%" 비율, 빈 값이면 무차원(배수 등)
}

// Definitions 는 사용 가능한 지표 정의 전체를 받는다. GET /tiingo/fundamentals/definitions
// 지표가 추가되면 목록도 늘어난다(2026-09-04 기준 85개).
func (c *Client) Definitions(ctx context.Context) ([]Definition, error) {
	var ds []Definition
	if err := c.http.GetJSON(ctx, "/tiingo/fundamentals/definitions", nil, &ds); err != nil {
		return nil, err
	}
	return ds, nil
}
