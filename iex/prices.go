package iex

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// Price 는 인트라데이 구간 하나의 시세다.
// Volume 은 PriceOptions.Columns 에 "volume" 을 넣어야 응답에 포함된다 — 없으면 0 이다.
type Price struct {
	Date  types.Time `json:"date"`  // 구간 시작 시각
	Open  float64    `json:"open"`  // 시가
	High  float64    `json:"high"`  // 고가
	Low   float64    `json:"low"`   // 저가
	Close float64    `json:"close"` // 종가
	// Volume 은 IEX 거래량이다. columns 에 명시하지 않으면 응답에 없어 0 이다.
	// Tiingo 가 이 엔드포인트에서는 "38607.0" 처럼 소수점을 붙여 보내므로 int64 로는 디코딩이 실패한다
	// (스냅샷의 Volume 은 정수로 와서 int64 다).
	Volume float64 `json:"volume"`
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
	AfterHours   bool         // true 면 장 전후 데이터를 포함한다
	ForceFill    bool         // true 면 거래가 없던 구간도 직전 값으로 채운다
	// Columns 는 받을 컬럼을 지정한다. 비어 있으면 기본 컬럼(date/open/high/low/close)만 오고
	// Volume 은 0 이다 — 거래량이 필요하면 "volume" 을 반드시 포함해야 한다.
	Columns []string
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
	if o.AfterHours {
		q["afterHours"] = "true" // 기본값(false)일 때는 보내지 않는다
	}
	if o.ForceFill {
		q["forceFill"] = "true"
	}
	if len(o.Columns) > 0 {
		q["columns"] = strings.Join(o.Columns, ",")
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Prices 는 인트라데이 시세를 받는다. GET /iex/<ticker>/prices
// 티커는 하나만 받는다 — 복수를 경로에 넣으면 Tiingo 가 404 를 돌려준다.
// 휴장 구간이면 빈 슬라이스이고, 없는 티커면 404 라 APIError(StatusCode 404)가 온다.
func (c *Client) Prices(ctx context.Context, ticker string, opts *PriceOptions) ([]Price, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ps []Price
	if err := c.http.GetJSON(ctx, "/iex/"+url.PathEscape(t)+"/prices", opts.params(), &ps); err != nil {
		return nil, err
	}
	return ps, nil
}
