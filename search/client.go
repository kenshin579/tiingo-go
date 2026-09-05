// Package search 는 Tiingo Search(자산 검색) 유틸리티 API sub-client 다.
// tiingo.Client.Search 로 접근한다.
//
// 티커·이름으로 찾으려면 Search, ISIN 으로 찾으려면 SearchByISIN 을 쓴다.
// 일치하는 자산이 없으면 에러가 아니라 빈 슬라이스가 온다.
//
// openFIGIComposite 는 값이 없을 때 Tiingo 가 null 과 문자열 "nan" 을 섞어 보내므로
// 둘 다 빈 문자열로 정규화한다(FIGI 참고) — 이 라이브러리에서 원문을 가공하는 유일한 곳이다.
package search

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Search sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
