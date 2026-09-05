// Tiingo Crypto 예제.
//
//	TIINGO_API_KEY=... go run ./examples/crypto
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/crypto"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ms, err := c.Crypto.Meta(ctx, "btcusd", "ethusd")
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range ms {
		fmt.Printf("%-8s %s (%s/%s)\n", m.Ticker, m.Name, m.BaseCurrency, m.QuoteCurrency)
	}

	b, err := c.Crypto.TopOfBookFor(ctx, "btcusd", nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(b.TopOfBookData) > 0 {
		d := b.TopOfBookData[0]
		fmt.Printf("호가 %s: 매수 %.2f(%s) / 매도 %.2f(%s), 최종 %.2f\n",
			d.QuoteTimestamp, d.BidPrice, d.BidExchange, d.AskPrice, d.AskExchange, d.LastPrice)
	}

	s, err := c.Crypto.PricesFor(ctx, "btcusd", &crypto.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -7),
		ResampleFreq: crypto.Resample1Day,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("최근 일봉 %d건\n", len(s.PriceData))
	for _, p := range s.PriceData {
		fmt.Printf("  %s  종가 %.2f  거래량 %.4f\n", p.Date, p.Close, p.Volume)
	}
}
