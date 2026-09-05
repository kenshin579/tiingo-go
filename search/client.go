// Package search 는 Tiingo Search(자산 검색) 유틸리티 API sub-client 다.
// tiingo.Client.Search 로 접근한다.
//
// 티커·이름으로 찾으려면 Search, ISIN 으로 찾으려면 SearchByISIN 을 쓴다.
// 일치하는 자산이 없으면 에러가 아니라 빈 슬라이스가 온다.
package search

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Search sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
