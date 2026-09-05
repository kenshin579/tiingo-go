// Package equity 는 Tiingo Equity Realtime(통합 주식 기준가·유동성 스냅샷·인트라데이 시세)
// API sub-client 다. tiingo.Client.Equity 로 접근한다.
//
// IEX 가 단일 거래소 피드인 것과 달리 여러 거래소·ATS·OTC 를 합친 통합 피드다.
// 스냅샷의 유동성 지표 다섯 개(LqSpread·LqBidPrice·LqBidSize·LqAskPrice·LqAskSize)는
// 통합 피드가 값을 내지 않으면 null 이라 포인터다 — 전 종목 조회 기준 44% 가 그렇다.
// 없는 티커는 스냅샷에서 에러가 아니라 응답에서 빠지므로 결과 길이가 요청보다 짧을 수 있다.
package equity

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Equity Realtime sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
