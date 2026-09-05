// Package crypto 는 Tiingo Crypto(암호화폐 페어 메타·시세·호가) API sub-client 다.
// tiingo.Client.Crypto 로 접근한다.
//
// prices 와 top 응답은 티커별로 한 겹 감싸여 있다([]PriceSeries, []TopOfBook).
// 단일 티커만 쓸 때는 PricesFor / TopOfBookFor 를 쓰면 첫 요소를 바로 받는다.
package crypto

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Crypto sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
