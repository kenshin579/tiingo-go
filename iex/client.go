// Package iex 는 Tiingo IEX(미국 주식 실시간 스냅샷·인트라데이 시세) API sub-client 다.
// tiingo.Client.IEX 로 접근한다.
//
// 스냅샷은 장 마감 시간대에 호가·체결 관련 필드가 모두 null 로 오므로 해당 필드는 포인터다.
// 없는 티커는 에러가 아니라 응답에서 빠지므로 결과 길이가 요청보다 짧을 수 있다.
package iex

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 IEX sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
