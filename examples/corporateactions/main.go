// Tiingo Corporate Actions 예제.
//
//	TIINGO_API_KEY=... go run ./examples/corporateactions
//
// 옵션 없이 부르면 상장 이후 전 기간(AAPL 은 1만 건 이상)이 오므로 기간을 좁혀 쓴다.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/corporateactions"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	ys, err := c.CorporateActions.DistributionYield(ctx, "AAPL", &corporateactions.YieldOptions{
		StartDate: time.Now().AddDate(0, -1, 0),
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(ys) == 0 {
		fmt.Println("수익률 데이터 없음")
		return
	}
	fmt.Printf("AAPL 최근 1개월 배당수익률 %d건 (마지막 5건)\n", len(ys))
	for _, y := range ys[max(0, len(ys)-5):] {
		fmt.Printf("  %s  %.4f%%\n", y.Date, y.TrailingDiv1Y*100)
	}
}
