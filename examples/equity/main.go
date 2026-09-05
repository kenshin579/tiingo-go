// Tiingo Equity Realtime 예제.
//
//	TIINGO_API_KEY=... go run ./examples/equity
//
// 통합 피드라 유동성 지표(LqSpread 등)가 함께 온다. 값이 없는 종목은 nil 이다.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/equity"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ss, err := c.Equity.Snapshots(ctx, []string{"AAPL", "SPY", "AULT"})
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range ss {
		fmt.Printf("%-6s %s  기준가 %.4f (전일 %.4f)  거래량 %.0f\n",
			s.Ticker, s.Timestamp, s.TngoLast, s.PrevClose, s.Volume)
		if s.LqSpread != nil && s.LqBidPrice != nil && s.LqAskPrice != nil {
			fmt.Printf("       유동성 %.4f / %.4f  스프레드 %.4f%%\n",
				*s.LqBidPrice, *s.LqAskPrice, *s.LqSpread*100)
		} else {
			fmt.Println("       유동성 지표 없음 — 통합 피드가 값을 내지 않았다")
		}
	}

	ps, err := c.Equity.Prices(ctx, "AAPL", &equity.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: equity.Resample1Hour,
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
