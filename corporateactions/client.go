// Package corporateactions 는 Tiingo Corporate Actions API sub-client 다.
// tiingo.Client.CorporateActions 로 접근한다.
//
// 지금은 배당수익률(DistributionYield) 하나만 구현돼 있다. 같은 API 그룹의 배당 내역
// (distributions, 티커별·배치)과 분할(splits)은 이 라이브러리를 만든 계정 권한으로 403 이라
// 응답 형태를 확인할 수 없어 넣지 않았다 — 실호출로 검증하지 않은 모델은 두지 않는다는 원칙이다.
// 권한이 열리면 이 패키지에 추가한다.
package corporateactions

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// Client 는 Corporate Actions sub-client.
type Client struct {
	http *httpclient.Client
}

// New 는 internal 용도 — 루트 tiingo.NewClient 가 호출한다.
func New(http *httpclient.Client) *Client { return &Client{http: http} }
