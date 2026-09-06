package corporateactions

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// Yield 는 하루치 배당수익률이다.
//
// Tiingo 가 이 엔드포인트에 지표를 계속 추가하겠다고 문서에 밝혔으므로 필드가 늘 수 있다.
// 모르는 필드는 무시되니 디코딩은 깨지지 않고, 새 지표가 필요해지면 그때 필드를 더하면 된다.
type Yield struct {
	Date types.Date `json:"date"` // 기준일. 응답은 늘 자정이라 날짜만 의미가 있다
	// TrailingDiv1Y 는 최근 1년 분배금 기준 수익률이다. 0.0033 이면 0.33% 다.
	// 문서 표는 string 이라 적지만 실제로는 JSON 숫자로 온다.
	// 배당을 시작하기 전 구간은 0 이며, 이는 결손이 아니라 정상 값이다.
	TrailingDiv1Y float64 `json:"trailingDiv1Y"`
}

// YieldOptions 는 DistributionYield 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
//
// 두 필드 모두 Tiingo 문서의 파라미터 표에는 없지만 실제로 동작한다(실호출 확인).
// 지정하지 않으면 상장 이후 전 기간이 오므로 응답이 매우 클 수 있다 — AAPL 이 11,523건이다.
type YieldOptions struct {
	StartDate time.Time // 조회 시작(>=). zero 면 미전송
	EndDate   time.Time // 조회 종료(<=). zero 면 미전송
}

// params 는 옵션을 쿼리 맵으로 바꾼다. 넣을 값이 없으면 nil 이다.
func (o *YieldOptions) params() map[string]string {
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
	if len(q) == 0 {
		return nil
	}
	return q
}

// DistributionYield 는 배당수익률 시계열을 받는다.
// GET /tiingo/corporate-actions/<ticker>/distribution-yield
//
// 옵션 없이 부르면 상장 이후 전 기간이 온다(AAPL 실측 11,523건, 1980년부터).
// 기간을 좁히려면 YieldOptions 를 쓴다. 기본 HTTP 타임아웃은 30초이고 본문 수신까지 포함하므로,
// 느린 회선에서 전 기간을 받는다면 tiingo.WithTimeout 으로 늘린다.
// 없는 티커는 404 가 아니라 빈 슬라이스다 — 이 점이 equity·iex 의 시세와 다르다.
func (c *Client) DistributionYield(ctx context.Context, ticker string, opts *YieldOptions) ([]Yield, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ys []Yield
	path := "/tiingo/corporate-actions/" + url.PathEscape(t) + "/distribution-yield"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ys); err != nil {
		return nil, err
	}
	return ys, nil
}
