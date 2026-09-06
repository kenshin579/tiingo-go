// Tiingo WebSocket 스트리밍 예제.
//
//	TIINGO_API_KEY=... go run ./examples/stream
//
// crypto 는 24시간 거래되므로 언제 돌려도 메시지가 온다.
// 주식·FX 는 시장 시간에만 데이터가 흐른다.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/stream"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := c.Stream.Crypto(ctx, &stream.CryptoOptions{
		Tickers:   []string{"btcusd", "ethusd"},
		Threshold: stream.CryptoTradesAndQuotesLevel,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	fmt.Println("15초간 수신합니다...")
	var trades, quotes int
	for msg := range s.Messages() {
		switch m := msg.(type) {
		case stream.CryptoTrade:
			trades++
			if trades <= 3 {
				fmt.Printf("  체결 %-7s %-9s %.8f @ %.2f\n", m.Ticker, m.Exchange, m.LastSize, m.LastPrice)
			}
		case stream.CryptoQuote:
			quotes++
			if quotes <= 3 {
				fmt.Printf("  호가 %-7s %-9s %.2f / %.2f (중간 %.2f)\n",
					m.Ticker, m.Exchange, m.BidPrice, m.AskPrice, m.MidPrice)
			}
		}
	}
	fmt.Printf("체결 %d건, 호가 %d건 (구독 id %d, 재연결 %d회, 누락 %d건)\n",
		trades, quotes, s.SubscriptionID(), s.Reconnects(), s.Dropped())
	if err := s.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
}
