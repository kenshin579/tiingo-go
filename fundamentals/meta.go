package fundamentals

import (
	"context"
	"fmt"
	"strings"

	"github.com/kenshin579/tiingo-go/types"
)

// Meta 는 펀더멘털 커버리지에 포함된 회사의 메타 정보다.
// 섹터·산업·소재지는 Fundamentals 권한이 있어야 채워진다(EOD 메타에는 없다).
type Meta struct {
	PermaTicker             string     `json:"permaTicker"`             // 티커 변경·재사용에도 안정적인 영구 식별자
	Ticker                  string     `json:"ticker"`                  // 티커(소문자로 온다)
	Name                    string     `json:"name"`                    // 회사명
	IsActive                bool       `json:"isActive"`                // 현재 상장·거래 중인지
	IsADR                   bool       `json:"isADR"`                   // 미국 예탁증서(ADR) 여부
	Sector                  string     `json:"sector"`                  // 섹터(예: Technology)
	Industry                string     `json:"industry"`                // 산업(예: Consumer Electronics)
	SICCode                 int        `json:"sicCode"`                 // SEC 표준산업분류 코드
	SICSector               string     `json:"sicSector"`               // SIC 기준 섹터
	SICIndustry             string     `json:"sicIndustry"`             // SIC 기준 산업
	ReportingCurrency       string     `json:"reportingCurrency"`       // 보고 통화(예: usd)
	Location                string     `json:"location"`                // 소재지(예: California, USA)
	CompanyWebsite          string     `json:"companyWebsite"`          // 회사 홈페이지
	SECFilingWebsite        string     `json:"secFilingWebsite"`        // 이 회사의 SEC EDGAR 공시 목록 URL
	StatementLastUpdated    types.Time `json:"statementLastUpdated"`    // 재무제표가 마지막으로 갱신된 시각
	DailyLastUpdated        types.Time `json:"dailyLastUpdated"`        // 일별 지표가 마지막으로 갱신된 시각
	DataProviderPermaTicker string     `json:"dataProviderPermaTicker"` // 원천 제공자 쪽 식별자(문서 표에는 없으나 실제로 온다)
}

// Meta 는 펀더멘털 커버리지 회사의 메타 정보를 받는다. GET /tiingo/fundamentals/meta
// tickers 를 주지 않으면 커버리지 전체(수천 건)를 받는다 — 대량 응답이므로 주의한다.
func (c *Client) Meta(ctx context.Context, tickers ...string) ([]Meta, error) {
	var params map[string]string
	if len(tickers) > 0 {
		cleaned := make([]string, 0, len(tickers))
		for _, t := range tickers {
			t = strings.TrimSpace(t)
			if t == "" {
				return nil, fmt.Errorf("tiingo: ticker must not be empty")
			}
			cleaned = append(cleaned, t)
		}
		params = map[string]string{"tickers": strings.Join(cleaned, ",")}
	}
	var ms []Meta
	if err := c.http.GetJSON(ctx, "/tiingo/fundamentals/meta", params, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}
