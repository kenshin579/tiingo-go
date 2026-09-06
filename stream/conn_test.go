package stream

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStream_SubscribeAndReceive(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, &CryptoOptions{Tickers: []string{"btcusd"}, Threshold: CryptoTradesAndQuotesLevel})
	require.NoError(t, err)
	defer s.Close()

	ts.push(`{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.0]}`)

	select {
	case m := <-s.Messages():
		tr, ok := m.(CryptoTrade)
		require.True(t, ok)
		assert.Equal(t, "btcusd", tr.Ticker)
	case <-time.After(3 * time.Second):
		t.Fatal("메시지가 오지 않았다")
	}
	assert.Equal(t, int64(0), s.Reconnects())

	// 구독 프레임에 인증과 옵션이 실렸는지 확인한다.
	var sub struct {
		EventName     string `json:"eventName"`
		Authorization string `json:"authorization"`
		EventData     struct {
			ThresholdLevel int      `json:"thresholdLevel"`
			Tickers        []string `json:"tickers"`
		} `json:"eventData"`
	}
	require.NotEmpty(t, ts.received())
	require.NoError(t, json.Unmarshal([]byte(ts.received()[0]), &sub))
	assert.Equal(t, "subscribe", sub.EventName)
	assert.Equal(t, "k", sub.Authorization)
	assert.Equal(t, 2, sub.EventData.ThresholdLevel)
	assert.Equal(t, []string{"btcusd"}, sub.EventData.Tickers)
}

// 티커를 비우면 tickers 키를 보내지 않는다 — 그래야 전 종목이 온다.
func TestStream_NoTickersOmitsKey(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return len(ts.received()) >= 1 }, 3*time.Second, 20*time.Millisecond)
	assert.NotContains(t, ts.received()[0], `"tickers"`)
	assert.Contains(t, ts.received()[0], `"thresholdLevel":5`, "nil 옵션은 체결만(5)")
}

// 구독 확인의 id 를 노출한다.
func TestStream_SubscriptionID(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return s.SubscriptionID() == 1 }, 3*time.Second, 20*time.Millisecond)
}

// I·H 는 채널로 나오지 않는다.
func TestStream_InfoAndHeartbeatNotEmitted(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	ts.push(`{"response":{"code":200,"message":"HeartBeat"},"messageType":"H"}`)
	ts.push(`{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54+00:00","x",1,2]}`)

	select {
	case m := <-s.Messages():
		_, ok := m.(CryptoTrade)
		assert.True(t, ok, "하트비트를 건너뛰고 데이터가 먼저 나온다")
	case <-time.After(3 * time.Second):
		t.Fatal("메시지가 오지 않았다")
	}
}

// E 는 Err() 로 흐르고 채널이 닫힌다. 재연결하지 않는다.
func TestStream_ErrorClosesStream(t *testing.T) {
	ts := newTestServer(t)
	ts.setAutoAck(false)
	c := New("k", WithBaseURL(ts.url), WithBackoff(10*time.Millisecond, 20*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.BOATS(ctx, nil)
	require.NoError(t, err)
	ts.push(`{"response":{"code":403,"message":"Not authorized for service: boats"},"messageType":"E"}`)

	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	require.Error(t, s.Err())
	assert.Contains(t, s.Err().Error(), "403")
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, ts.conns(), "권한 에러 뒤에는 재연결하지 않는다")
}

// 연결이 끊기면 자동으로 다시 붙고 구독을 재전송한다.
func TestStream_Reconnects(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithBackoff(10*time.Millisecond, 50*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, &CryptoOptions{Tickers: []string{"btcusd"}})
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	ts.dropConn()
	assert.Eventually(t, func() bool { return ts.conns() >= 2 }, 5*time.Second, 20*time.Millisecond)
	assert.Eventually(t, func() bool { return len(ts.received()) >= 2 }, 3*time.Second, 20*time.Millisecond)
	assert.Equal(t, ts.received()[0], ts.received()[1], "재전송된 구독 프레임은 처음과 같다")
	assert.Eventually(t, func() bool { return s.Reconnects() == 1 }, 3*time.Second, 20*time.Millisecond)
}

// 하트비트도 데이터도 없이 조용하면 끊긴 것으로 보고 다시 붙는다 — 문서상 서버는 30초마다 H 를 보낸다.
func TestStream_ReconnectsOnSilence(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url),
		WithReadTimeout(150*time.Millisecond),
		WithBackoff(10*time.Millisecond, 20*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	// 서버가 아무것도 보내지 않는다 — 클라이언트가 스스로 끊고 다시 붙어야 한다.
	assert.Eventually(t, func() bool { return ts.conns() >= 2 }, 3*time.Second, 20*time.Millisecond)
}

// 하트비트가 오는 동안은 끊지 않는다.
func TestStream_HeartbeatKeepsAlive(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithReadTimeout(200*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		ts.push(`{"response":{"code":200,"message":"HeartBeat"},"messageType":"H"}`)
	}
	assert.Equal(t, 1, ts.conns(), "하트비트가 데드라인을 갱신하므로 재연결하지 않는다")
}

// 재연결 후에도 데이터가 계속 흐른다.
func TestStream_ReceivesAfterReconnect(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithBackoff(10*time.Millisecond, 50*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	ts.dropConn()
	assert.Eventually(t, func() bool { return ts.conns() >= 2 }, 5*time.Second, 20*time.Millisecond)

	ts.push(`{"service":"crypto_data","messageType":"A","data":["T","ethusd","2026-09-06T03:55:54+00:00","x",1,2]}`)
	select {
	case m := <-s.Messages():
		tr, ok := m.(CryptoTrade)
		require.True(t, ok)
		assert.Equal(t, "ethusd", tr.Ticker)
	case <-time.After(3 * time.Second):
		t.Fatal("재연결 후 메시지가 오지 않았다")
	}
}

// 소비하지 않으면 오래된 메시지를 버리고 카운터가 오른다 — 시세는 LOSSY 다.
func TestStream_DropsWhenSlow(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithBuffer(4))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	for i := 0; i < 50; i++ {
		ts.push(`{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54+00:00","x",1,2]}`)
	}
	assert.Eventually(t, func() bool { return s.Dropped() > 0 }, 3*time.Second, 20*time.Millisecond)
	assert.LessOrEqual(t, len(s.Messages()), 4, "버퍼를 넘지 않는다")
}

// 버린 뒤에도 가장 최신 메시지가 채널에 남아 있다 — 오래된 것을 버리고 새 것을 넣는다.
func TestStream_DropsOldestKeepsNewest(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithBuffer(2))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	for _, tk := range []string{"aaa", "bbb", "ccc", "ddd", "eee"} {
		ts.push(`{"service":"crypto_data","messageType":"A","data":["T","` + tk + `","2026-09-06T03:55:54+00:00","x",1,2]}`)
	}
	assert.Eventually(t, func() bool { return s.Dropped() >= 3 }, 3*time.Second, 20*time.Millisecond)

	var got []string
	for i := 0; i < 2; i++ {
		select {
		case m := <-s.Messages():
			got = append(got, m.(CryptoTrade).Ticker)
		case <-time.After(time.Second):
			t.Fatal("버퍼에 메시지가 남아 있어야 한다")
		}
	}
	assert.Contains(t, got, "eee", "가장 최신 메시지는 살아남는다")
	assert.NotContains(t, got, "aaa", "가장 오래된 메시지는 버려진다")
}

// 깨진 데이터 한 건 때문에 스트림 전체가 죽지 않는다.
func TestStream_BadFrameDoesNotKillStream(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	ts.push(`{"service":"crypto_data","messageType":"A","data":["T","btcusd"]}`) // 너무 짧다
	ts.push(`not json at all`)
	ts.push(`{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54+00:00","x",1,2]}`)

	select {
	case m := <-s.Messages():
		_, ok := m.(CryptoTrade)
		assert.True(t, ok, "깨진 건은 건너뛰고 정상 건이 온다")
	case <-time.After(3 * time.Second):
		t.Fatal("메시지가 오지 않았다")
	}
}

// ctx 취소로 끝나면 Err 은 ctx 에러다.
func TestStream_ContextCancel(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithCancel(context.Background())

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)
	cancel()

	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	assert.ErrorIs(t, s.Err(), context.Canceled)
}

// Close 는 깨끗하게 끝내고 재연결하지 않는다. Err 은 nil 이다.
func TestStream_CloseStopsReconnect(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithBackoff(10*time.Millisecond, 20*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)

	require.NoError(t, s.Close())
	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	assert.NoError(t, s.Err(), "명시적 Close 는 에러가 아니다")

	before := ts.conns()
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, before, ts.conns(), "Close 후에는 다시 붙지 않는다")
}

// Close 를 두 번 불러도 안전하다.
func TestStream_CloseIdempotent(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
}

// 다섯 진입 메서드가 각각 올바른 경로와 기본 threshold 를 쓴다.
func TestClient_EntryPoints(t *testing.T) {
	cases := []struct {
		name  string
		open  func(*Client, context.Context) (*Stream, error)
		path  string
		level string
	}{
		{"crypto", func(c *Client, ctx context.Context) (*Stream, error) { return c.Crypto(ctx, nil) }, "/crypto", `"thresholdLevel":5`},
		{"forex", func(c *Client, ctx context.Context) (*Stream, error) { return c.Forex(ctx, nil) }, "/fx", `"thresholdLevel":5`},
		{"iex", func(c *Client, ctx context.Context) (*Stream, error) { return c.IEX(ctx, nil) }, "/iex", `"thresholdLevel":6`},
		{"equity", func(c *Client, ctx context.Context) (*Stream, error) { return c.Equity(ctx, nil) }, "/equity/intraday", `"thresholdLevel":6`},
		{"boats", func(c *Client, ctx context.Context) (*Stream, error) { return c.BOATS(ctx, nil) }, "/boats", `"thresholdLevel":3`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			c := New("k", WithBaseURL(ts.url))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s, err := tc.open(c, ctx)
			require.NoError(t, err)
			defer s.Close()

			assert.Eventually(t, func() bool { return len(ts.received()) >= 1 }, 3*time.Second, 20*time.Millisecond)
			assert.Equal(t, tc.path, ts.lastPath(), "경로")
			assert.Contains(t, ts.received()[0], tc.level, "기본 threshold")
		})
	}
}

func TestBackoff(t *testing.T) {
	assert.Equal(t, time.Duration(0), backoff(0, time.Second, 10*time.Second), "첫 시도는 즉시")
	assert.Equal(t, time.Duration(0), backoff(1, time.Second, 10*time.Second))
	d2 := backoff(2, time.Second, 10*time.Second)
	assert.InDelta(t, float64(time.Second), float64(d2), float64(300*time.Millisecond), "2회차는 min 근처(±20%)")
	d3 := backoff(3, time.Second, 10*time.Second)
	assert.InDelta(t, float64(2*time.Second), float64(d3), float64(500*time.Millisecond))
	d9 := backoff(9, time.Second, 10*time.Second)
	assert.LessOrEqual(t, d9, 12*time.Second, "max 를 넘지 않는다(jitter 포함)")
}
