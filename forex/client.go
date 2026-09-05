// Package forex 는 Tiingo Forex(통화쌍 호가·시세) API sub-client 다.
// tiingo.Client.Forex 로 접근한다.
//
// 응답은 평탄한 배열이다 — Crypto 와 달리 티커별로 감싸지 않으며, 시세는 행마다 Ticker 가 붙는다.
// 시장이 닫혀 있거나(주말) 없는 통화쌍이면 에러가 아니라 빈 슬라이스가 온다.
package forex

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Forex sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
