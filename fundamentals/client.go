// Package fundamentals 는 Tiingo Fundamentals(재무제표·일별 지표·지표 정의·회사 메타) API
// sub-client 다. tiingo.Client.Fundamentals 로 접근한다.
//
// Fundamentals 는 별도 구독(add-on)이다. 무료 플랜은 Dow 30 종목의 3년치만 제공하며,
// 권한 밖 종목은 400/403 이 오고 APIError 로 그대로 전달된다.
package fundamentals

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Fundamentals sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
