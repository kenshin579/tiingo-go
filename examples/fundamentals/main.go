// Tiingo Fundamentals 예제.
//
//	TIINGO_API_KEY=... go run ./examples/fundamentals
//
// Fundamentals 는 별도 구독이다. 무료 플랜은 Dow 30 종목의 3년치만 조회된다.
package main

import (
	"context"
	"fmt"
	"log"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/fundamentals"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ms, err := c.Fundamentals.Meta(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	if len(ms) == 0 {
		log.Fatal("메타 없음")
	}
	m := ms[0]
	fmt.Printf("%s — %s / %s (%s)\n", m.Name, m.Sector, m.Industry, m.Location)
	fmt.Printf("  재무제표 갱신 %s, 일별 지표 갱신 %s\n", m.StatementLastUpdated, m.DailyLastUpdated)

	ss, err := c.Fundamentals.Statements(ctx, "AAPL", &fundamentals.StatementOptions{Sort: "-date"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("재무제표 %d기간\n", len(ss))
	for _, s := range ss[:min(3, len(ss))] {
		rev, _ := s.StatementData.Get(fundamentals.CodeRevenue)
		ni, _ := s.StatementData.Get(fundamentals.CodeNetinc)
		label := fmt.Sprintf("%dQ%d", s.Year, s.Quarter)
		if s.Quarter == 0 {
			label = fmt.Sprintf("%d 연간", s.Year)
		}
		fmt.Printf("  %-10s %s  매출 %.0f  순이익 %.0f\n", label, s.Date, rev, ni)
	}

	ds, err := c.Fundamentals.Daily(ctx, "AAPL", &fundamentals.DailyOptions{Sort: "-date"})
	if err != nil {
		log.Fatal(err)
	}
	if len(ds) > 0 {
		d := ds[0]
		fmt.Printf("일별 지표 %s: 시총 %.0f  P/E %.2f  P/B %.2f\n", d.Date, d.MarketCap, d.PERatio, d.PBRatio)
	}
}
