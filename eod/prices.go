package eod

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
	"github.com/kenshin579/tiingo-go/types"
)

// Price 는 일별 시세 한 건이다. 뮤추얼 펀드는 OHLC 에 그날의 NAV 가 들어간다.
type Price struct {
	Date      types.Date `json:"date"`      // 이 데이터가 해당하는 날짜
	Open      float64    `json:"open"`      // 시가
	High      float64    `json:"high"`      // 고가
	Low       float64    `json:"low"`       // 저가
	Close     float64    `json:"close"`     // 종가
	Volume    int64      `json:"volume"`    // 거래량(주)
	AdjOpen   float64    `json:"adjOpen"`   // 수정 시가
	AdjHigh   float64    `json:"adjHigh"`   // 수정 고가
	AdjLow    float64    `json:"adjLow"`    // 수정 저가
	AdjClose  float64    `json:"adjClose"`  // 수정 종가. CRSP 방법론, 반올림하지 않음
	AdjVolume int64      `json:"adjVolume"` // 수정 거래량(분할 계수 반영)
	DivCash   float64    `json:"divCash"`   // 해당일(배당락일) 지급 배당금
	// SplitFactor 는 분할·역분할·분배 시 가격 조정에 쓰이는 계수다. 1 이 아니면 그날 이벤트가
	// 있었다는 뜻이므로, 증분 동기화 중이라면 해당 종목의 전체 이력을 다시 받아야 한다.
	SplitFactor float64 `json:"splitFactor"`
}

// ResampleFreq 는 가격 리샘플링 주기다.
// 주의: daily 만 휴일 달력을 반영하고 나머지는 표준 영업일(월~금) 기준이다.
// 기간 중간 날짜를 주면 전체 기간이 잡히도록 startDate 는 뒤로, endDate 는 앞으로 조정된다
// (예: weekly 에 수요일 startDate → 월요일로 롤백, 수요일 endDate → 금요일로 롤포워드).
type ResampleFreq string

const (
	ResampleDaily    ResampleFreq = "daily"    // 일별. 휴일 달력 반영
	ResampleWeekly   ResampleFreq = "weekly"   // 주별. 금요일 마감
	ResampleMonthly  ResampleFreq = "monthly"  // 월별. 각 월 마지막 영업일 마감
	ResampleAnnually ResampleFreq = "annually" // 연별. 각 해 마지막 영업일 마감
)

// PriceOptions 는 HistoricalPrices 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작일(이상, >=). zero 면 미전송
	EndDate      time.Time    // 조회 종료일(이하, <=). zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플링 주기. 빈 값이면 미전송(일별)
	Sort         string       // 정렬 컬럼. "date" 오름차순, "-date" 내림차순
	Columns      []string     // 돌려받을 컬럼만 지정. 비어 있으면 전체
}

// params 는 옵션을 쿼리 맵으로 바꾼다. zero 값은 넣지 않는다.
func (o *PriceOptions) params() map[string]string {
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
	if o.ResampleFreq != "" {
		q["resampleFreq"] = string(o.ResampleFreq)
	}
	if o.Sort != "" {
		q["sort"] = o.Sort
	}
	if len(o.Columns) > 0 {
		q["columns"] = strings.Join(o.Columns, ",")
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// HistoricalPrices 는 일별 시세를 조회한다. GET /tiingo/daily/<ticker>/prices
// opts 가 nil 이거나 날짜가 없으면 Tiingo 는 최신 1건만 반환한다.
func (c *Client) HistoricalPrices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ps []Price
	path := "/tiingo/daily/" + url.PathEscape(t) + "/prices"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ps); err != nil {
		return nil, err
	}
	return ps, nil
}

// LatestPrice 는 가장 최근 일별 시세 1건을 조회한다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) LatestPrice(ctx context.Context, ticker string) (*Price, error) {
	ps, err := c.HistoricalPrices(ctx, ticker, nil)
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, httpclient.ErrNotFound
	}
	return &ps[0], nil
}
