// Package eod 는 Tiingo End-of-Day(일별 시세·메타) API sub-client 다.
// tiingo.Client.EOD 로 접근한다.
package eod

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 End-of-Day sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
