package fundamentals

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// DailyMetric 은 주가 기반 일별 지표다. 재무제표와 달리 매일 갱신된다.
type DailyMetric struct {
	Date          types.Date `json:"date"`          // 지표 기준일
	MarketCap     float64    `json:"marketCap"`     // 시가총액
	EnterpriseVal float64    `json:"enterpriseVal"` // 기업가치(EV)
	PERatio       float64    `json:"peRatio"`       // 주가수익비율(P/E)
	PBRatio       float64    `json:"pbRatio"`       // 주가순자산비율(P/B)
	TrailingPEG1Y float64    `json:"trailingPEG1Y"` // 최근 1년 기준 PEG
}

// DailyOptions 는 Daily 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type DailyOptions struct {
	StartDate time.Time // 조회 시작일(이상, >=). zero 면 미전송
	EndDate   time.Time // 조회 종료일(이하, <=). zero 면 미전송
	Sort      string    // 정렬 컬럼. "date" 오름차순, "-date" 내림차순
}

// params 는 옵션을 쿼리 맵으로 바꾼다. zero 값은 넣지 않는다.
func (o *DailyOptions) params() map[string]string {
	if o == nil {
		return nil
	}
	q := map[string]string{}
	if !o.StartDate.IsZero() {
		q["startDate"] = o.StartDate.Format(types.DateLayout)
	}
	if !o.EndDate.IsZero() {
		q["endDate"] = o.EndDate.Format(types.DateLayout)
	}
	if o.Sort != "" {
		q["sort"] = o.Sort
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Daily 는 주가 기반 일별 지표를 받는다. GET /tiingo/fundamentals/<ticker>/daily
// ticker 자리에는 티커 또는 permaTicker(상장폐지·재사용 심볼용)를 넣는다.
func (c *Client) Daily(ctx context.Context, ticker string, opts *DailyOptions) ([]DailyMetric, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ds []DailyMetric
	path := "/tiingo/fundamentals/" + url.PathEscape(t) + "/daily"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ds); err != nil {
		return nil, err
	}
	return ds, nil
}
