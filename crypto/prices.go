package crypto

import (
	"context"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/internal/httpclient"
	"github.com/kenshin579/tiingo-go/types"
)

// Price 는 리샘플 구간 하나의 시세다. 구간 길이는 요청의 resampleFreq 가 정한다.
type Price struct {
	Date           types.Time `json:"date"`           // 구간 시작 시각. 1day 면 자정, 1min 이면 분 단위
	Open           float64    `json:"open"`           // 시가
	High           float64    `json:"high"`           // 고가
	Low            float64    `json:"low"`            // 저가
	Close          float64    `json:"close"`          // 종가
	Volume         float64    `json:"volume"`         // 거래량(기준 통화 단위)
	VolumeNotional float64    `json:"volumeNotional"` // 거래대금(표시 통화 단위)
	TradesDone     float64    `json:"tradesDone"`     // 체결 건수
}

// PriceSeries 는 티커 하나의 시세 묶음이다. Tiingo 가 티커별로 감싸서 돌려준다.
type PriceSeries struct {
	Ticker        string  `json:"ticker"`        // 페어 티커
	BaseCurrency  string  `json:"baseCurrency"`  // 기준 통화
	QuoteCurrency string  `json:"quoteCurrency"` // 표시 통화
	PriceData     []Price `json:"priceData"`     // 시간순 시세. 요청 범위·주기에 따라 길이가 다르다
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
	Exchanges    []string     // 특정 거래소로 한정. 비어 있으면 전체
}

// apply 는 옵션을 쿼리 맵에 채워 넣는다. zero 값은 넣지 않는다.
func (o *PriceOptions) apply(q map[string]string) {
	if o == nil {
		return
	}
	if !o.StartDate.IsZero() {
		q["startDate"] = o.StartDate.Format(types.DateLayout)
	}
	if !o.EndDate.IsZero() {
		q["endDate"] = o.EndDate.Format(types.DateLayout)
	}
	if o.ResampleFreq != "" {
		q["resampleFreq"] = o.ResampleFreq
	}
	if len(o.Exchanges) > 0 {
		q["exchanges"] = strings.Join(o.Exchanges, ",")
	}
}

// Prices 는 티커별 시세 묶음을 받는다. GET /tiingo/crypto/prices
// tickers 는 하나 이상이어야 한다.
func (c *Client) Prices(ctx context.Context, tickers []string, opts *PriceOptions) ([]PriceSeries, error) {
	joined, err := joinTickers(tickers, true)
	if err != nil {
		return nil, err
	}
	q := map[string]string{"tickers": joined}
	opts.apply(q)

	var ss []PriceSeries
	if err := c.http.GetJSON(ctx, "/tiingo/crypto/prices", q, &ss); err != nil {
		return nil, err
	}
	return ss, nil
}

// PricesFor 는 티커 하나의 시세 묶음을 받는다. 결과가 없으면 ErrNotFound 를 반환한다.
func (c *Client) PricesFor(ctx context.Context, ticker string, opts *PriceOptions) (*PriceSeries, error) {
	ss, err := c.Prices(ctx, []string{ticker}, opts)
	if err != nil {
		return nil, err
	}
	if len(ss) == 0 {
		return nil, httpclient.ErrNotFound
	}
	return &ss[0], nil
}
