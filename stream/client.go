// Package stream 은 Tiingo WebSocket 실시간 피드 클라이언트다.
// 피드는 다섯 가지다 — Crypto(/crypto), Forex(/fx), IEX(/iex), Equity 통합(/equity/intraday), BOATS(/boats).
//
// 사용법:
//
//	c := stream.New(apiKey)
//	s, err := c.Crypto(ctx, &stream.CryptoOptions{Tickers: []string{"btcusd"}})
//	defer s.Close()
//	for m := range s.Messages() {
//		switch v := m.(type) {
//		case stream.CryptoTrade:  ...
//		case stream.CryptoQuote:  ...
//		}
//	}
//	if err := s.Err(); err != nil { ... } // 채널이 닫힌 뒤 원인
//
// 동작 원칙:
//
//   - 메시지는 채널로 소비한다. 구독 확인(I)·하트비트(H)는 내부에서 소화되고 채널로 나오지 않는다.
//     확인에 담긴 id 는 Stream.SubscriptionID 로 본다.
//   - 소비가 느리면 가장 오래된 버퍼 메시지를 버리고 새 것을 넣는다(기본 버퍼 256, Stream.Dropped 로 집계).
//     시세는 본질적으로 손실을 허용하는 데이터이고, 읽기를 막으면 소켓이 정체되어 서버가 연결을 끊는다.
//     Tickers 를 비워 전 종목을 받을 때는 256 이 부족할 수 있으므로 WithBuffer 로 늘린다.
//   - 전송 오류(읽기 실패·연결 끊김)는 지수 백오프(+jitter)로 자동 재연결하고 구독 프레임을 다시 보낸다.
//     서버가 E 프레임으로 거부하면(권한 없음 등) 재연결하지 않고 스트림을 끝낸다 — 다시 붙어도 같은 결과다.
//     하트비트조차 없이 WithReadTimeout(기본 90s) 동안 조용하면 끊긴 것으로 보고 역시 다시 붙는다.
//   - 데이터 프레임(A) 한 건이 깨졌으면 그 건만 건너뛴다. 스트림은 계속된다.
//   - 진입 메서드는 연결을 기다리지 않고 즉시 돌아온다. 연결 실패는 채널이 닫힌 뒤 Stream.Err 로 드러난다.
//
// 인증은 헤더나 쿼리가 아니라 구독 프레임 안의 authorization 필드로 보낸다.
//
// 배열 매핑 검증 수준(2026-09-06 기준):
//
//   - Crypto: 실호출로 검증(체결 11건·호가 18건).
//   - Forex: 문서 예시의 산술로 검증. Tiingo 문서의 인덱스 표는 askPrice/askSize 순서가 틀렸고,
//     예시 값의 mid=(bid+ask)/2 검산이 맞는 순서를 따른다. 캡처 당시 FX 장이 닫혀 실호출 미검증.
//   - IEX·Equity: 문서만. 캡처 당시 미국 장이 닫혀 있었다.
//   - BOATS: 문서만. 계정에 엔타이틀먼트가 없어(403) 실호출이 불가능하다.
package stream

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// DefaultBaseURL 은 Tiingo WebSocket 엔드포인트의 공통 접두다. 피드별 경로가 뒤에 붙는다.
const DefaultBaseURL = "wss://api.tiingo.com"

const (
	defaultBuffer     = 256              // 기본 채널 버퍼
	defaultBackoffMin = time.Second      // 재연결 백오프 시작
	defaultBackoffMax = 30 * time.Second // 재연결 백오프 상한
	// defaultReadTimeout 은 프레임이 하나도 오지 않아도 연결이 살아 있다고 볼 최대 시간이다.
	// 문서상 서버는 30초마다 하트비트(H)를 보내므로 그 3배로 잡는다. 이 시간 동안 조용하면
	// 네트워크가 소리 없이 죽은 것으로 보고 다시 붙는다.
	defaultReadTimeout = 90 * time.Second
)

// Client 는 WebSocket 피드의 진입점이다. New 로 만든다.
type Client struct {
	apiKey      string
	baseURL     string
	buffer      int
	backoffMin  time.Duration
	backoffMax  time.Duration
	readTimeout time.Duration // 프레임 없이 버틸 최대 시간. 넘기면 재연결
	httpClient  *http.Client  // nil 이면 http.DefaultClient
}

// Option 은 New 의 설정 함수다.
type Option func(*Client)

// WithBaseURL 은 엔드포인트 접두를 바꾼다(테스트·프록시용). 끝 슬래시 없이 "ws(s)://host" 형태.
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = u } }

// WithBuffer 는 메시지 채널 버퍼 크기를 정한다(기본 256). 0 이하는 무시한다.
func WithBuffer(n int) Option {
	return func(c *Client) {
		if n > 0 {
			c.buffer = n
		}
	}
}

// WithBackoff 는 재연결 대기 시간의 시작값과 상한을 정한다(기본 1s / 30s).
func WithBackoff(min, max time.Duration) Option {
	return func(c *Client) {
		c.backoffMin = min
		c.backoffMax = max
	}
}

// WithReadTimeout 은 프레임(하트비트 포함)이 하나도 오지 않아도 연결을 살아 있다고 볼 최대
// 시간이다(기본 90s — 하트비트 주기 30s 의 3배). 넘기면 끊긴 것으로 보고 다시 붙는다.
// 0 이하는 무시한다.
func WithReadTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.readTimeout = d
		}
	}
}

// WithHTTPClient 는 핸드셰이크에 쓸 HTTP 클라이언트를 지정한다(프록시·타임아웃 등).
func WithHTTPClient(hc *http.Client) Option { return func(c *Client) { c.httpClient = hc } }

// New 는 WebSocket 클라이언트를 만든다. 연결은 각 진입 메서드에서 이뤄진다.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:      apiKey,
		baseURL:     DefaultBaseURL,
		buffer:      defaultBuffer,
		backoffMin:  defaultBackoffMin,
		backoffMax:  defaultBackoffMax,
		readTimeout: defaultReadTimeout,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// eventData 는 구독 프레임의 eventData 다.
// tickers 는 omitempty 라 비우면 키 자체가 빠지고, 그러면 서버가 전 종목을 보낸다.
type eventData struct {
	ThresholdLevel int      `json:"thresholdLevel"`
	Tickers        []string `json:"tickers,omitempty"`
}

// subscribeFrame 은 연결마다 한 번 보내는 구독 프레임이다. 인증 토큰이 여기에 실린다.
type subscribeFrame struct {
	EventName     string    `json:"eventName"`
	Authorization string    `json:"authorization"`
	EventData     eventData `json:"eventData"`
}

// decodeFunc 는 피드별 data 배열 디코더다. (nil, nil) 은 "모르는 종류, 무시".
type decodeFunc func(json.RawMessage) (Message, error)

// Crypto 는 암호화폐 피드를 연다. opts 가 nil 이면 전 종목·체결만(5).
func (c *Client) Crypto(ctx context.Context, opts *CryptoOptions) (*Stream, error) {
	if opts == nil {
		opts = &CryptoOptions{}
	}
	th := opts.Threshold
	if th == 0 {
		th = CryptoTradesOnly
	}
	return c.open(ctx, "/crypto", eventData{ThresholdLevel: int(th), Tickers: opts.Tickers}, decodeCrypto)
}

// Forex 는 외환 피드를 연다. opts 가 nil 이면 전 통화쌍·체결만(5).
func (c *Client) Forex(ctx context.Context, opts *ForexOptions) (*Stream, error) {
	if opts == nil {
		opts = &ForexOptions{}
	}
	th := opts.Threshold
	if th == 0 {
		th = ForexTradesOnly
	}
	return c.open(ctx, "/fx", eventData{ThresholdLevel: int(th), Tickers: opts.Tickers}, decodeForex)
}

// IEX 는 IEX 피드를 연다. opts 가 nil 이면 전 종목·기준가(6).
func (c *Client) IEX(ctx context.Context, opts *IEXOptions) (*Stream, error) {
	if opts == nil {
		opts = &IEXOptions{}
	}
	th := opts.Threshold
	if th == 0 {
		th = IEXReferencePriceLevel
	}
	return c.open(ctx, "/iex", eventData{ThresholdLevel: int(th), Tickers: opts.Tickers}, decodeIEX)
}

// Equity 는 equity 통합 피드를 연다. opts 가 nil 이면 전 종목·기준가(6).
func (c *Client) Equity(ctx context.Context, opts *EquityOptions) (*Stream, error) {
	if opts == nil {
		opts = &EquityOptions{}
	}
	th := opts.Threshold
	if th == 0 {
		th = EquityReferencePriceLevel
	}
	return c.open(ctx, "/equity/intraday", eventData{ThresholdLevel: int(th), Tickers: opts.Tickers}, decodeEquity)
}

// BOATS 는 BOATS 피드를 연다. opts 가 nil 이면 전 종목·호가+체결(6).
// 엔타이틀먼트가 없는 계정은 서버가 E(403) 로 거부하고, 그러면 스트림은 Err 에 그 에러를 담고 끝난다.
func (c *Client) BOATS(ctx context.Context, opts *BOATSOptions) (*Stream, error) {
	if opts == nil {
		opts = &BOATSOptions{}
	}
	th := opts.Threshold
	if th == 0 {
		th = BOATSTopOfBookAndTrades
	}
	return c.open(ctx, "/boats", eventData{ThresholdLevel: int(th), Tickers: opts.Tickers}, decodeBOATS)
}
