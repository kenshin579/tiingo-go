// Tiingo IEX 예제.
//
//	TIINGO_API_KEY=... go run ./examples/iex
//
// 장 마감 시간대에는 호가·체결 필드가 nil 이다.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/iex"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	qs, err := c.IEX.Quotes(ctx, []string{"AAPL", "MSFT"})
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range qs {
		fmt.Printf("%-6s %s  최종 %.2f (전일 %.2f)  거래량 %d\n",
			q.Ticker, q.Timestamp, q.TngoLast, q.PrevClose, q.Volume)
		if q.BidPrice != nil && q.AskPrice != nil {
			fmt.Printf("       호가 %.2f / %.2f\n", *q.BidPrice, *q.AskPrice)
		} else {
			fmt.Println("       호가 없음 — 장 마감")
		}
	}

	ps, err := c.IEX.Prices(ctx, "AAPL", &iex.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: iex.Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(ps) == 0 {
		fmt.Println("시세 없음 — 조회 구간이 전부 휴장일일 수 있다")
		return
	}
	fmt.Printf("최근 1시간봉 %d건 (마지막 5건)\n", len(ps))
	for _, p := range ps[max(0, len(ps)-5):] {
		fmt.Printf("  %s  종가 %.2f  거래량 %.0f\n", p.Date, p.Close, p.Volume)
	}
}
