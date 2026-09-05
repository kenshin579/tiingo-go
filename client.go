package tiingo

import (
	"errors"

	"github.com/kenshin579/tiingo-go/crypto"
	"github.com/kenshin579/tiingo-go/eod"
	"github.com/kenshin579/tiingo-go/equity"
	"github.com/kenshin579/tiingo-go/forex"
	"github.com/kenshin579/tiingo-go/fundamentals"
	"github.com/kenshin579/tiingo-go/iex"
	"github.com/kenshin579/tiingo-go/internal/httpclient"
	"github.com/kenshin579/tiingo-go/search"
)

// Client 는 tiingo-go 라이브러리의 단일 진입점이다.
// 카테고리별 서브클라이언트를 필드로 노출한다.
type Client struct {
	http *httpclient.Client

	EOD          *eod.Client          // 일별 시세·메타(End-of-Day)
	Fundamentals *fundamentals.Client // 재무제표·일별 지표(Fundamentals, 별도 구독)
	Crypto       *crypto.Client       // 암호화폐 페어 메타·시세·호가
	Forex        *forex.Client        // 통화쌍 호가·시세
	IEX          *iex.Client          // 미국 주식 실시간 스냅샷·인트라데이 시세
	Search       *search.Client       // 티커·이름·ISIN 자산 검색
	Equity       *equity.Client       // 통합 피드 기준가·유동성 스냅샷, 인트라데이 시세
}

// NewClient 는 API 키로 Client 를 만든다. 키가 비어 있으면 에러다.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("tiingo: apiKey is required")
	}
	var cfg clientOptions
	for _, opt := range opts {
		opt(&cfg)
	}
	hc := httpclient.New(httpclient.Config{
		APIKey:     apiKey,
		BaseURL:    cfg.baseURL,
		Timeout:    cfg.timeout,
		HTTPClient: cfg.httpClient,
	})
	c := &Client{http: hc}
	c.EOD = eod.New(hc)
	c.Fundamentals = fundamentals.New(hc)
	c.Crypto = crypto.New(hc)
	c.Forex = forex.New(hc)
	c.IEX = iex.New(hc)
	c.Search = search.New(hc)
	c.Equity = equity.New(hc)
	return c, nil
}
