// Tiingo Search 예제.
//
//	TIINGO_API_KEY=... go run ./examples/search
//
// 같은 티커가 국가별로 중복될 수 있어 국가 코드를 함께 찍는다.
package main

import (
	"context"
	"fmt"
	"log"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/search"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	rs, err := c.Search.Search(ctx, "apple", &search.SearchOptions{Limit: 5})
	if err != nil {
		log.Fatal(err)
	}
	if len(rs) == 0 {
		fmt.Println("검색 결과 없음")
		return
	}
	fmt.Printf("\"apple\" 검색 결과 %d건\n", len(rs))
	for _, r := range rs {
		figi := string(r.OpenFIGIComposite)
		if figi == "" {
			figi = "-" // null 이거나 "nan" 이면 빈 문자열로 정규화된다
		}
		fmt.Printf("  %-8s %-2s %-10s %-14s %s\n", r.Ticker, r.CountryCode, r.AssetType, figi, r.Name)
	}

	// ISIN 은 자산 하나를 정확히 지목한다.
	byISIN, err := c.Search.SearchByISIN(ctx, "US0378331005", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range byISIN {
		fmt.Printf("ISIN US0378331005 → %s (%s) %s\n", r.Ticker, r.CountryCode, r.Name)
	}
}
