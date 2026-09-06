package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/coder/websocket"
)

// subscribeAck 는 Tiingo 가 구독 프레임에 돌려주는 확인(I) 프레임이다. 실서버 캡처 그대로.
const subscribeAck = `{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":1},"messageType":"I"}`

// testServer 는 실제 웹소켓 업그레이드를 하는 스텁이다.
// 클라이언트가 보낸 텍스트 프레임을 기록하고, 테스트가 원하는 프레임을 밀어 넣을 수 있다.
//
// 기본적으로 구독 프레임("eventName":"subscribe")에는 실서버와 같이 즉시 I 확인을 돌려준다.
// 거부(E) 등 응답을 직접 통제해야 하는 테스트는 setAutoAck(false) 로 끄고 push 로 보낸다.
type testServer struct {
	srv *httptest.Server
	url string // ws://127.0.0.1:PORT

	mu      sync.Mutex
	frames  []string      // 클라이언트 → 서버 텍스트 프레임(구독 프레임)
	paths   []string      // 연결별 요청 경로(진입 메서드 경로 검증용)
	conn    int           // 총 연결 수(재연결 검증용)
	autoAck bool          // 구독 프레임에 자동으로 I 를 돌려줄지
	pushCh  chan string   // 서버 → 클라이언트로 밀어 넣을 프레임. 연결이 없어도 큐에 쌓인다
	closeCh chan struct{} // 현재 연결을 강제 종료하라는 신호
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	ts := &testServer{pushCh: make(chan string, 64), closeCh: make(chan struct{}, 8), autoAck: true}
	ts.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		ts.mu.Lock()
		ts.conn++
		ts.paths = append(ts.paths, r.URL.Path)
		ts.mu.Unlock()

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		go func() { // 읽기 루프: 수신 프레임 기록 + 구독에 자동 확인
			for {
				_, data, err := c.Read(ctx)
				if err != nil {
					cancel()
					return
				}
				s := string(data)
				ts.mu.Lock()
				ts.frames = append(ts.frames, s)
				autoAck := ts.autoAck
				ts.mu.Unlock()
				if autoAck && strings.Contains(s, `"eventName":"subscribe"`) {
					_ = c.Write(ctx, websocket.MessageText, []byte(subscribeAck))
				}
			}
		}()
		for {
			select {
			case <-ctx.Done():
				_ = c.CloseNow()
				return
			case <-ts.closeCh:
				_ = c.CloseNow()
				return
			case msg := <-ts.pushCh:
				if err := c.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
					return
				}
			}
		}
	}))
	ts.url = "ws" + strings.TrimPrefix(ts.srv.URL, "http")
	t.Cleanup(ts.srv.Close)
	return ts
}

// push 는 서버 → 클라이언트 텍스트 프레임을 보낸다. 연결이 아직 없으면 큐에 남아 있다가
// 다음 연결에서 나간다(버퍼 64).
func (ts *testServer) push(frame string) { ts.pushCh <- frame }

// dropConn 은 현재 연결을 강제로 끊는다(재연결 검증용).
func (ts *testServer) dropConn() { ts.closeCh <- struct{}{} }

// conns 는 지금까지 받아들인 총 연결 수다.
func (ts *testServer) conns() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.conn
}

// received 는 클라이언트가 보낸 텍스트 프레임의 복사본이다(순서 보존).
func (ts *testServer) received() []string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]string(nil), ts.frames...)
}

// lastPath 는 가장 최근 연결의 요청 경로다. 연결이 없으면 빈 문자열.
func (ts *testServer) lastPath() string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.paths) == 0 {
		return ""
	}
	return ts.paths[len(ts.paths)-1]
}

// setAutoAck 는 구독 프레임에 대한 자동 I 확인을 켜고 끈다.
func (ts *testServer) setAutoAck(v bool) {
	ts.mu.Lock()
	ts.autoAck = v
	ts.mu.Unlock()
}
