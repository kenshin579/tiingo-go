// Tiingo End-of-Day 예제.
//
//	TIINGO_API_KEY=... go run ./examples/eod
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/eod"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	m, err := c.EOD.Meta(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (%s) — %s ~ %s\n", m.Name, m.ExchangeCode, m.StartDate, m.EndDate)

	p, err := c.EOD.LatestPrice(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("최신 종가 %s: %.2f (거래량 %d)\n", p.Date, p.Close, p.Volume)

	ps, err := c.EOD.HistoricalPrices(ctx, "AAPL", &eod.PriceOptions{
		StartDate:    time.Now().AddDate(0, -1, 0),
		ResampleFreq: eod.ResampleWeekly,
		Sort:         "-date",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("최근 1개월 주별 %d건\n", len(ps))
	for _, x := range ps {
		fmt.Printf("  %s  종가 %.2f  수정종가 %.2f\n", x.Date, x.Close, x.AdjClose)
	}
}
