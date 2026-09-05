package equity

import (
	"context"
	"fmt"
	"strings"

	"github.com/kenshin579/tiingo-go/types"
)

// intradayPath 는 이 카테고리의 공통 경로다. 스냅샷은 이 경로를 그대로 쓰고,
// 시세는 뒤에 <ticker>/prices 를 붙인다. 끝 슬래시가 필요하다 — 없으면 Tiingo 가 301 을 돌려준다.
const intradayPath = "/tiingo/equity/intraday/"

// Snapshot 은 통합 피드 기준가·유동성 스냅샷이다.
//
// LqSpread·LqBidPrice·LqBidSize·LqAskPrice·LqAskSize 다섯 개는 통합 피드가 값을 내지 않으면
// null 이라 포인터다 — 전 종목 조회 18,654건 중 8,185건(44%)이 그렇다. nil 은 "값 없음"이고
// 0 과 구분된다. 이름이 비슷한 LqRefPrice 는 늘 채워져 값 타입이다.
type Snapshot struct {
	Ticker     string     `json:"ticker"`     // 티커. 요청과 무관하게 대문자로 온다
	Timestamp  types.Time `json:"timestamp"`  // 마지막 갱신 시각
	Open       float64    `json:"open"`       // 당일 시가
	High       float64    `json:"high"`       // 당일 고가
	Low        float64    `json:"low"`        // 당일 저가
	TngoLast   float64    `json:"tngoLast"`   // Tiingo 기준 최종가. 스프레드가 넓지 않으면 중간가를 쓴다
	PrevClose  float64    `json:"prevClose"`  // 전일 종가
	LqRefPrice float64    `json:"lqRefPrice"` // 유동성 기준가. TngoLast 와 같은 값이다
	// Volume 은 통합 장중 거래량이다. 문서는 int64 라 하지만 행에 따라 1778528.0 처럼 소수점이
	// 붙어 오므로(같은 응답에서 972327 처럼 정수로도 온다) int64 로는 디코딩이 실패한다.
	Volume float64 `json:"volume"`
	// 아래 5개는 유동성 위험 지표다. 호가 그 자체가 아니라 지표의 구성요소이며,
	// WebSocket 의 thresholdLevel 4 메시지와 같은 값이다.
	LqSpread   *float64 `json:"lqSpread"`   // 상대 스프레드. 0.04 면 4%
	LqBidPrice *float64 `json:"lqBidPrice"` // 매수 측 가격
	LqBidSize  *float64 `json:"lqBidSize"`  // 매수 측 수량(주). Volume 과 같은 계열이라 float64
	LqAskPrice *float64 `json:"lqAskPrice"` // 매도 측 가격
	LqAskSize  *float64 `json:"lqAskSize"`  // 매도 측 수량(주)
}

// joinTickers 는 티커 목록을 검증해 콤마로 합친다. 공백뿐인 원소나 빈 목록은 에러다.
func joinTickers(tickers []string) (string, error) {
	cleaned := make([]string, 0, len(tickers))
	for _, t := range tickers {
		t = strings.TrimSpace(t)
		if t == "" {
			return "", fmt.Errorf("tiingo: ticker must not be empty")
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return "", fmt.Errorf("tiingo: at least one ticker is required")
	}
	return strings.Join(cleaned, ","), nil
}

// Snapshots 는 지정한 티커의 스냅샷을 받는다. GET /tiingo/equity/intraday/
//
// 없는 티커는 에러가 아니라 응답에서 빠지고 순서도 요청과 다를 수 있으므로,
// 결과를 인덱스로 대응시키지 말고 Ticker 필드로 찾는다.
// 비활성 종목은 Timestamp 가 한참 지난 값으로 오므로(실측 1년 이상) 최신성이 중요하면 확인한다.
// 전 종목을 받으려면 AllSnapshots 를 쓴다 — 빈 목록은 에러다.
func (c *Client) Snapshots(ctx context.Context, tickers []string) ([]Snapshot, error) {
	joined, err := joinTickers(tickers)
	if err != nil {
		return nil, err
	}
	var ss []Snapshot
	if err := c.http.GetJSON(ctx, intradayPath, map[string]string{"tickers": joined}, &ss); err != nil {
		return nil, err
	}
	return ss, nil
}

// AllSnapshots 는 지원하는 전 종목의 스냅샷을 받는다. GET /tiingo/equity/intraday/
//
// 응답이 크다 — 실측 18,654건 약 4.85MB 로, 받는 데 3초 남짓 걸린다.
// 전 종목 스캔이 목적이 아니라면 Snapshots 를 쓴다.
// 상장폐지·비활성 종목도 섞여 있어 Timestamp 가 한 달 이상 지난 행이 있다.
// 기본 HTTP 타임아웃은 30초이고 본문 수신까지 포함하므로, 느린 회선에서는
// tiingo.WithTimeout 으로 늘린다.
func (c *Client) AllSnapshots(ctx context.Context) ([]Snapshot, error) {
	var ss []Snapshot
	if err := c.http.GetJSON(ctx, intradayPath, nil, &ss); err != nil {
		return nil, err
	}
	return ss, nil
}
