package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// readLimit 은 프레임 하나의 최대 크기다. 전 종목 구독 시 프레임이 커질 수 있어 넉넉히 잡는다.
const readLimit = 16 << 20

// Stream 은 열린 피드 하나다. 진입 메서드(Client.Crypto 등)가 만든다.
//
// 메시지는 Messages 로 읽고, 채널이 닫히면 Err 로 원인을 본다. 연결·재연결·구독 재전송은
// 내부 goroutine 이 맡는다. Close 로 끝낸다.
type Stream struct {
	msgs        chan Message       // 소비자에게 나가는 메시지. run goroutine 만 닫는다
	done        chan struct{}      // run goroutine 이 끝나면 닫힌다
	closed      chan struct{}      // Close 가 불리면 닫힌다. ctx 취소와 구분하는 근거
	cancel      context.CancelFunc // 내부 ctx 취소
	readTimeout time.Duration      // 프레임 없이 버틸 최대 시간(Client.readTimeout)
	everAcked   bool               // 구독 확인을 한 번이라도 받았는지. run goroutine 만 만진다

	closeOnce  sync.Once
	mu         sync.Mutex
	err        error        // 최종 에러. msgs 가 닫히기 전에 확정된다
	subID      atomic.Int64 // 구독 확인(I)의 subscriptionId
	dropped    atomic.Int64 // 느린 소비자 때문에 버린 메시지 수
	reconnects atomic.Int64 // 첫 연결 이후 다시 붙은 횟수(구독 확인 기준)
}

// Messages 는 데이터 메시지 채널이다. 스트림이 끝나면 닫힌다.
// 타입 스위치로 피드별 메시지(CryptoTrade, ForexQuote 등)를 구분한다.
func (s *Stream) Messages() <-chan Message { return s.msgs }

// Err 은 스트림이 끝난 원인이다. Messages 채널이 닫힌 뒤에 읽어야 의미가 있다 —
// 그 전에는 nil 이다. 깨끗한 Close 는 nil, ctx 취소는 ctx 에러, 서버 거부(E)는 그 에러다.
func (s *Stream) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// SubscriptionID 는 서버가 구독 확인(I)에 담아 준 id 다. 아직 확인 전이면 0.
// 재연결하면 새 id 로 바뀐다.
// 재연결 중에는 직전 연결의 id 가 새 ack 가 올 때까지 남아 있다 — 최신 여부는 Reconnects 와 함께 본다.
func (s *Stream) SubscriptionID() int64 { return s.subID.Load() }

// Reconnects 는 첫 연결 이후 다시 붙은 횟수다. 재연결 사이에는 메시지가 빠질 수 있으므로 값이 오르면 REST 로 재동기화를 고려한다.
func (s *Stream) Reconnects() int64 { return s.reconnects.Load() }

// Dropped 는 소비가 느려 버린 메시지의 누적 수다.
func (s *Stream) Dropped() int64 { return s.dropped.Load() }

// Close 는 스트림을 끝내고 goroutine 이 정리될 때까지 기다린다. 여러 번 불러도 안전하다.
// 아직 흐르고 있던 스트림을 이렇게 끝내면 Err 은 nil 이다. 이미 다른 이유로 끝난 뒤라면 그 에러가 남는다.
// 항상 nil 을 돌려준다.
func (s *Stream) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
		s.cancel()
	})
	<-s.done
	return nil
}

// isClosed 는 Close 가 불렸는지 알려준다.
func (s *Stream) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

// finish 는 최종 에러를 확정한다. run goroutine 이 채널을 닫기 직전에 한 번 부른다.
func (s *Stream) finish(err error) {
	s.mu.Lock()
	s.err = err
	s.mu.Unlock()
}

// emit 은 메시지를 채널에 넣는다. 절대 막히지 않는다 — 버퍼가 차 있으면 가장 오래된 것을
// 하나 빼서 버리고(dropped 증가) 새 것을 넣는다. run goroutine 에서만 부르므로 생산자끼리는
// 경쟁하지 않는다. 그 사이 소비자가 채널을 비웠으면 두 번째 넣기는 그냥 성공한다.
func (s *Stream) emit(m Message) {
	select {
	case s.msgs <- m:
		return
	default:
	}
	// 소비자의 수신과 같은 채널에서 경쟁할 수 있다 — 그 순간 소비자가 자리를 비웠다면 하나를
	// 불필요하게 더 버릴 수 있다. 상한(버퍼 크기)은 지켜지고 emit 은 절대 블로킹하지 않는다.
	select {
	case <-s.msgs: // 가장 오래된 것을 버린다
		s.dropped.Add(1)
	default:
	}
	select {
	case s.msgs <- m:
	default: // 버퍼 0 처럼 넣을 자리가 아예 없으면 새 것을 버린다. 정상 설정에서는 오지 않는다
		s.dropped.Add(1)
	}
}

// open 은 스트림을 만들고 run goroutine 을 띄운 뒤 곧바로 돌아온다.
// 동기적으로 실패하는 것은 잘못된 base URL 뿐이다. 연결 실패는 Err 로 흐른다.
func (c *Client) open(ctx context.Context, path string, ev eventData, decode decodeFunc) (*Stream, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("tiingo: invalid stream base URL %q: %w", c.baseURL, err)
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return nil, fmt.Errorf("tiingo: stream base URL must be ws:// or wss://, got %q", c.baseURL)
	}
	frame, err := json.Marshal(subscribeFrame{EventName: "subscribe", Authorization: c.apiKey, EventData: ev})
	if err != nil {
		return nil, fmt.Errorf("tiingo: marshal subscribe frame: %w", err)
	}

	ictx, cancel := context.WithCancel(ctx)
	s := &Stream{
		msgs:        make(chan Message, c.buffer),
		done:        make(chan struct{}),
		closed:      make(chan struct{}),
		cancel:      cancel,
		readTimeout: c.readTimeout,
	}
	go c.run(ictx, s, u.String(), frame, decode)
	return s, nil
}

// run 은 연결·구독·읽기·재연결을 반복하는 goroutine 본체다. 끝날 때 채널을 닫는다.
//
// 백오프 카운터는 구독 확인(I)이 온 뒤에만 0 으로 되돌린다. TCP 는 받아 주지만 곧바로 끊어
// 버리는 서버에 대해 연결 성공만으로 카운터를 되돌리면 백오프가 무력해지기 때문이다.
func (c *Client) run(ctx context.Context, s *Stream, wsURL string, frame []byte, decode decodeFunc) {
	defer close(s.done)
	defer close(s.msgs)
	defer s.cancel()

	attempt := 0
	var lastErr error // 마지막 전송 오류. ctx 로 끝날 때 원인을 함께 남긴다
	for {
		attempt++
		if !sleepCtx(ctx, backoff(attempt, c.backoffMin, c.backoffMax)) {
			s.finish(s.exitErr(ctx, lastErr))
			return
		}

		conn, err := c.dial(ctx, wsURL)
		if err != nil {
			if ctx.Err() != nil {
				s.finish(s.exitErr(ctx, lastErr))
				return
			}
			lastErr = err
			continue
		}
		if err := conn.Write(ctx, websocket.MessageText, frame); err != nil {
			_ = conn.CloseNow()
			if ctx.Err() != nil {
				s.finish(s.exitErr(ctx, lastErr))
				return
			}
			lastErr = fmt.Errorf("tiingo: send subscribe frame: %w", err)
			continue
		}

		fatal, acked, err := s.readLoop(ctx, conn, decode)
		_ = conn.CloseNow()
		if acked {
			attempt = 0
		}
		if fatal {
			s.finish(err)
			return
		}
		lastErr = err
	}
}

// exitErr 은 ctx 로 끝날 때의 최종 에러다. Close 로 끝났으면 nil, 아니면 ctx 에러이고
// 재연결 중이었다면 마지막 전송 오류를 덧붙인다(errors.Is 로 ctx 에러 판별은 그대로 된다).
func (s *Stream) exitErr(ctx context.Context, lastErr error) error {
	if s.isClosed() {
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("tiingo: stream stopped while reconnecting: %w (last error: %v)", ctx.Err(), lastErr)
	}
	return ctx.Err()
}

// dial 은 핸드셰이크를 수행한다. 인증은 핸드셰이크가 아니라 구독 프레임에서 한다.
func (c *Client) dial(ctx context.Context, wsURL string) (*websocket.Conn, error) {
	conn, resp, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPClient: c.httpClient})
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("tiingo: websocket dial %s (HTTP %d): %w", wsURL, resp.StatusCode, err)
		}
		return nil, fmt.Errorf("tiingo: websocket dial %s: %w", wsURL, err)
	}
	conn.SetReadLimit(readLimit)
	return conn, nil
}

// readLoop 는 연결 하나에서 프레임을 읽어 처리한다.
//
// 읽기마다 readTimeout 데드라인을 새로 건다. 서버는 데이터가 없어도 30초마다 하트비트(H)를
// 보내므로, 그 시간 동안 아무 프레임도 없으면 네트워크가 소리 없이 죽은 것이다(RST 없는
// NAT/방화벽 타임아웃). 데드라인 없이 Read 하면 OS TCP 타임아웃까지 수 분을 멈춰 있게 된다.
//
// 반환값 fatal 이 true 면 스트림을 끝내야 한다(err 이 최종 에러 — Close 면 nil, ctx 취소면
// ctx 에러, 서버 거부면 그 에러). false 면 전송 오류이니 재연결한다.
// acked 는 이 연결에서 구독 확인(I)을 받았는지다 — 백오프 카운터를 되돌리는 근거.
func (s *Stream) readLoop(ctx context.Context, conn *websocket.Conn, decode decodeFunc) (fatal, acked bool, err error) {
	for {
		rctx, cancel := context.WithTimeout(ctx, s.readTimeout)
		_, data, err := conn.Read(rctx)
		deadline := rctx.Err() == context.DeadlineExceeded
		cancel()
		if err != nil {
			if s.isClosed() {
				return true, acked, nil
			}
			if ctx.Err() != nil {
				return true, acked, ctx.Err()
			}
			if deadline {
				return false, acked, fmt.Errorf("tiingo: no frame for %s, reconnecting", s.readTimeout)
			}
			return false, acked, fmt.Errorf("tiingo: websocket read: %w", err)
		}
		e, err := parseEnvelope(data)
		if err != nil {
			continue // 깨진 프레임 한 건은 건너뛴다
		}
		switch e.MessageType {
		case msgInfo:
			if err := e.asError(); err != nil {
				return true, acked, err // 200 이 아닌 구독 확인은 거부다
			}
			if id, err := e.subscriptionID(); err == nil {
				s.subID.Store(id)
			}
			if !acked { // 같은 연결에서 확인이 두 번 와도 한 번만 센다
				if s.everAcked {
					s.reconnects.Add(1) // 확인이 온 순간 센다 — 연결이 끝난 뒤가 아니라
				}
				s.everAcked = true
			}
			acked = true
		case msgHeartbeat:
			// 내부에서 소화한다
		case msgData:
			m, err := decode(e.Data)
			if err != nil || m == nil {
				continue // 디코드 실패·모르는 종류는 건너뛴다
			}
			s.emit(m)
		case msgError:
			return true, acked, e.asError() // 권한 없음 등. 다시 붙어도 같으니 재연결하지 않는다
		default:
			// 모르는 종류는 무시한다
		}
	}
}

// sleepCtx 는 d 만큼 기다린다. ctx 가 먼저 끝나면 false.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// backoff 는 지수 백오프 대기 시간을 계산한다(±20% jitter). 첫 시도(attempt<=1)는 곧바로 한다 —
// 대부분의 끊김은 일시적이라 즉시 한 번 시도해 보는 편이 빠르게 복구된다. 두 번째 시도부터
// min, 2·min, 4·min, … 으로 늘리고 max 에서 멈춘다.
func backoff(attempt int, min, max time.Duration) time.Duration {
	if attempt <= 1 {
		return 0
	}
	d := min
	for i := 2; i < attempt && d < max; i++ {
		d *= 2
	}
	if d > max {
		d = max
	}
	jitter := 1 + (rand.Float64()*0.4 - 0.2) // 재시도 분산용. 암호학적 강도 불필요
	return time.Duration(float64(d) * jitter)
}
