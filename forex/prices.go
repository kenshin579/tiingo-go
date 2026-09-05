package forex

import (
	"context"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// Price 는 리샘플 구간 하나의 시세다. FX 는 중앙 거래소가 없어 거래량이 제공되지 않는다.
type Price struct {
	Date   types.Time `json:"date"`   // 구간 시작 시각. 1day 면 자정, 1min 이면 분 단위
	Ticker string     `json:"ticker"` // 통화쌍. 복수 조회 시 행마다 다르다
	Open   float64    `json:"open"`   // 시가
	High   float64    `json:"high"`   // 고가
	Low    float64    `json:"low"`    // 저가
	Close  float64    `json:"close"`  // 종가
}

// ResampleFreq 는 시세 리샘플 주기다. "5min", "4hour" 처럼 <숫자><단위> 형식이면 무엇이든 되므로
// 아래 상수는 자주 쓰는 값일 뿐 목록이 닫혀 있지 않다.
type ResampleFreq = string

const (
	Resample1Min  ResampleFreq = "1min"  // 1분봉
	Resample5Min  ResampleFreq = "5min"  // 5분봉
	Resample1Hour ResampleFreq = "1hour" // 1시간봉
	Resample1Day  ResampleFreq = "1day"  // 일봉
)

// PriceOptions 는 Prices 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type PriceOptions struct {
	StartDate    time.Time    // 조회 시작. zero 면 미전송
	EndDate      time.Time    // 조회 종료. zero 면 미전송
	ResampleFreq ResampleFreq // 리샘플 주기. 빈 값이면 Tiingo 기본값
}

// params 는 옵션을 쿼리 맵으로 바꾼다. 넣을 값이 없으면 nil 이다.
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
		q["resampleFreq"] = o.ResampleFreq
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Prices 는 리샘플 시세를 받는다. GET /tiingo/fx/<tickers>/prices
// 복수 티커는 경로에 콤마로 들어가고, 응답의 각 행에 Ticker 가 붙는다.
// 시장이 닫혀 있거나(주말) 없는 통화쌍이면 에러가 아니라 빈 슬라이스를 돌려준다 — 둘은 구분되지 않는다.
func (c *Client) Prices(ctx context.Context, tickers []string, opts *PriceOptions) ([]Price, error) {
	joined, err := pathTickers(tickers)
	if err != nil {
		return nil, err
	}
	var ps []Price
	if err := c.http.GetJSON(ctx, "/tiingo/fx/"+joined+"/prices", opts.params(), &ps); err != nil {
		return nil, err
	}
	return ps, nil
}
