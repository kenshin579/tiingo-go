// Tiingo Forex 예제.
//
//	TIINGO_API_KEY=... go run ./examples/forex
//
// FX 는 주말에 시세가 비어 있다.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/forex"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	qs, err := c.Forex.TopOfBook(ctx, []string{"eurusd", "usdjpy", "gbpusd"})
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range qs {
		fmt.Printf("%-8s %s  매수 %.5f / 매도 %.5f (중간 %.5f)\n",
			q.Ticker, q.QuoteTimestamp, q.BidPrice, q.AskPrice, q.MidPrice)
	}

	ps, err := c.Forex.Prices(ctx, []string{"eurusd"}, &forex.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -7),
		ResampleFreq: forex.Resample1Day,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(ps) == 0 {
		fmt.Println("시세 없음 — 조회 구간이 전부 휴장일일 수 있다")
		return
	}
	fmt.Printf("최근 일봉 %d건\n", len(ps))
	for _, p := range ps {
		fmt.Printf("  %s  %s  종가 %.5f\n", p.Date, p.Ticker, p.Close)
	}
}
