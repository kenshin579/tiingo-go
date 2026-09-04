package tiingo

import (
	"errors"

	"github.com/kenshin579/tiingo-go/eod"
	"github.com/kenshin579/tiingo-go/internal/httpclient"
)

// Client 는 tiingo-go 라이브러리의 단일 진입점이다.
// 카테고리별 서브클라이언트를 필드로 노출한다.
type Client struct {
	http *httpclient.Client

	EOD *eod.Client // 일별 시세·메타(End-of-Day)
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
	return c, nil
}
