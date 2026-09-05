package search

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// searchPath 는 검색 엔드포인트 경로다. 검색어는 경로가 아니라 쿼리 파라미터로 보낸다 —
// 경로 방식(/search/<query>)도 같은 결과를 주지만, 검색어에 공백·슬래시가 들어오면 경로가 깨진다.
const searchPath = "/tiingo/utilities/search"

// Result 는 검색 결과 한 건이다.
//
// Ticker 는 유일하지 않다 — 같은 티커가 국가별로 여러 건 올 수 있고(예: AAPL 의 미국·캐나다 상장),
// 유일 키는 PermaTicker 다. 결과를 티커로 대응시키면 안 된다.
type Result struct {
	Ticker            string `json:"ticker"`            // 티커. 국가가 다르면 중복될 수 있다
	Name              string `json:"name"`              // 자산 이름
	PermaTicker       string `json:"permaTicker"`       // 영구 식별자. 검색 결과의 유일 키
	OpenFIGIComposite FIGI   `json:"openFIGIComposite"` // OpenFIGI 복합 식별자. 없으면 빈 문자열
	AssetType         string `json:"assetType"`         // 자산 종류(Stock, ETF, Mutual Fund)
	IsActive          bool   `json:"isActive"`          // 현재 거래 중인지
	CountryCode       string `json:"countryCode"`       // 상장 국가(예: US, CA). 문서 표에는 없으나 실제로 온다
}

// SearchOptions 는 검색의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
// Search 와 SearchByISIN 이 함께 쓴다 — 두 검색 모두 아래 네 파라미터를 받는다.
type SearchOptions struct {
	ExactTickerMatch bool // true 면 티커가 정확히 일치하는 것만
	IncludeDelisted  bool // true 면 상장폐지 종목도 포함한다(기본은 제외)
	// Limit 은 최대 건수다. 0 이하면 미전송이고 Tiingo 기본값 10 이 적용된다(무제한이 아니다).
	// 상한은 100 이라 그보다 큰 값을 줘도 100 건까지만 온다.
	Limit int
	// Columns 는 받을 컬럼을 지정한다. 지정하면 응답에서 나머지 키가 빠져 해당 필드는 zero 값이 된다.
	Columns []string
}

// apply 는 옵션을 쿼리 맵에 채워 넣는다. 수신자가 nil 이면 아무것도 하지 않는다.
func (o *SearchOptions) apply(q map[string]string) {
	if o == nil {
		return
	}
	if o.ExactTickerMatch {
		q["exactTickerMatch"] = "true" // 기본값(false)일 때는 보내지 않는다
	}
	if o.IncludeDelisted {
		q["includeDelisted"] = "true"
	}
	if o.Limit > 0 {
		q["limit"] = strconv.Itoa(o.Limit)
	}
	if len(o.Columns) > 0 {
		q["columns"] = strings.Join(o.Columns, ",")
	}
}

// searchBy 는 검색 키(query 또는 isin) 하나로 요청한다.
// key 는 쿼리 파라미터 이름이자 에러 메시지에 쓰이는 인자 이름이다 — 두 메서드의 인자명과 같게 유지한다.
func (c *Client) searchBy(ctx context.Context, key, value string, opts *SearchOptions) ([]Result, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil, fmt.Errorf("tiingo: %s must not be empty", key)
	}
	q := map[string]string{key: v}
	opts.apply(q)

	var rs []Result
	if err := c.http.GetJSON(ctx, searchPath, q, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

// Search 는 티커·이름으로 자산을 검색한다. GET /tiingo/utilities/search
//
// 일치하는 자산이 없으면 에러가 아니라 빈 슬라이스를 돌려준다.
// 결과의 Ticker 는 중복될 수 있으므로 PermaTicker 로 구분한다.
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) ([]Result, error) {
	return c.searchBy(ctx, "query", query, opts)
}

// SearchByISIN 은 ISIN 으로 자산을 검색한다. GET /tiingo/utilities/search
//
// isin 파라미터는 카탈로그 문서의 파라미터 표에 없지만 실제로 지원된다 — Tiingo 개요 문서가
// "ISIN and OpenFIGI search are supported via the Search utility" 라고 적고 있고,
// 파라미터를 하나도 주지 않았을 때의 400 메시지도 query 또는 isin 을 요구한다.
// SearchOptions 의 네 파라미터를 모두 함께 쓸 수 있다(실호출 확인).
func (c *Client) SearchByISIN(ctx context.Context, isin string, opts *SearchOptions) ([]Result, error) {
	return c.searchBy(ctx, "isin", isin, opts)
}
