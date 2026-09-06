# WebSocket 스트리밍 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tiingo WebSocket 5종(Crypto·Forex·IEX·Equity·BOATS)을 단일 `stream` 패키지로 구현해 `tiingo.Client.Stream` 으로 노출한다.

**Architecture:** 다섯 엔드포인트가 연결·인증·구독·하트비트·재연결·봉투 파싱을 공유하고 다른 건 `data` 배열 해석뿐이다. 그래서 기계장치(`conn.go`·`envelope.go`·`decode.go`)를 한 번만 만들고 엔드포인트별 파일은 배열 매핑만 담는다. 소비는 채널(`s.Messages()`), 종료 사유는 `s.Err()`.

**Tech Stack:** Go 1.25, `github.com/coder/websocket` v1.8.14(첫 비테스트 외부 의존성), `stretchr/testify`

**스펙:** `docs/superpowers/specs/2026-09-06-websocket-design.md`
**브랜치:** `feature/websocket` (이미 생성됨, main `eac0c22` 기준)

---

## 이 계획을 실행하기 전에 알아야 할 것

**검증된 엔드포인트는 Crypto 하나뿐이다.** Forex 는 조사 시점 FX 휴장, IEX·Equity 는 미국 장 마감,
BOATS 는 403(영구)이었다. 나머지 넷의 배열 매핑은 **문서 기반**이며, 그 사실을 타입 주석에 남기는
것이 구현의 일부다. 위치 배열은 매핑이 틀려도 컴파일·테스트가 통과하고 값만 조용히 바뀌므로,
엄격 디코딩(Task 3)이 유일한 방어선이다.

**가장 틀리기 쉬운 지점**: **Forex 문서의 인덱스 표가 틀렸다.**

표는 `Ask Price`를 6, `Ask Size`를 7 이라 적지만, 같은 문서의 예시가 그 반대임을 증명한다:

```
["Q","eurnok",ts, 5000000.0, 9.6764, 9.678135, 5000000.0, 9.67987]
                   bidSize   bidPrice  mid      askSize   askPrice
```

`idx6 = 5000000.0` 은 명백히 수량이고, `mid = (bid+ask)/2` 는 **`idx7` 을 매도호가로 놓을 때만**
맞는다(`(9.6764+9.67987)/2 = 9.678135`). 즉 **Forex 는 Crypto 와 같은 순서**다.

**표의 숫자가 아니라 예시의 산술을 따른다.** Equity 유동성·BOATS 호가는 같은 방식으로 검산했고
표와 예시가 일치했다 — Forex 만 어긋난다.

**참고할 형제 구현**: `/Users/user/src/workspace_moneyflow/toss-go/stream/` 이 같은
`coder/websocket` 으로 같은 문제(재연결·백오프·스텁 서버)를 이미 풀었다. `conn.go` 의 `backoff`,
`testserver_test.go` 의 스텁 서버가 특히 참고가 된다. **복사하지 말고 구조만 참고한다** —
토스는 주문이 LOSSLESS 라 백프레셔에 연결을 끊지만, 여기 시세는 LOSSY 라 오래된 것을 버린다.

---

## 파일 구조

| 파일 | 책임 |
| --- | --- |
| `stream/client.go` (신규) | 패키지 doc, `Client`, `New`, 5개 진입 메서드 |
| `stream/message.go` (신규) | `Message` 인터페이스 |
| `stream/envelope.go` (신규) | `messageType` I/H/A/E 분기 |
| `stream/decode.go` (신규) | 위치 배열 커서 `arr`(엄격 검증) |
| `stream/conn.go` (신규) | 연결·구독·재연결·백오프·채널 펌프 |
| `stream/crypto.go` (신규) | `CryptoTrade`/`CryptoQuote`/`CryptoOptions` |
| `stream/forex.go` (신규) | `ForexQuote`/`ForexOptions` |
| `stream/iex.go` (신규) | `IEXReferencePrice`/`IEXOptions` |
| `stream/equity.go` (신규) | `EquityReferencePrice`/`EquityLiquidity`/`EquityOptions` |
| `stream/boats.go` (신규) | `BOATSQuote`/`BOATSTrade`/`BOATSOptions` |
| `stream/testserver_test.go` (신규) | 실제 업그레이드를 하는 스텁 WS 서버 |
| `client.go` (수정) | `Stream *stream.Client` 필드 |
| `examples/stream/main.go` (신규) | 실행 예제(crypto) |
| `integration_test.go` (수정) | crypto 스트림 + BOATS 403 |
| `README.md` (수정) | WebSocket 절 신설 |
| `../CLAUDE.md` (수정) | 워크스페이스 문서 |

---

### Task 1: 의존성과 메시지 캡처

**Files:**
- Modify: `go.mod`, `go.sum`
- Create: `stream/testdata/crypto_live.jsonl`
- Create: `stream/testdata/{forex,iex,equity,boats}_doc.jsonl`

- [ ] **Step 1: `coder/websocket` 을 추가한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go get github.com/coder/websocket@v1.8.14
go mod tidy
grep -n "coder/websocket" go.mod
```

버전은 형제 프로젝트 `toss-go` 와 같은 v1.8.14 로 맞춘다. 다른 버전이 잡히면 명시적으로 지정한다.

- [ ] **Step 2: crypto 실측 메시지를 캡처한다**

crypto 는 24시간 거래되므로 언제든 받을 수 있다. `websocat` 이 설치돼 있다.

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
mkdir -p stream/testdata
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
{ printf '{"eventName":"subscribe","authorization":"%s","eventData":{"thresholdLevel":2,"tickers":["btcusd","ethusd"]}}\n' "$KEY"; sleep 20; } \
  | timeout 25 websocat wss://api.tiingo.com/crypto > stream/testdata/crypto_live.jsonl
wc -l stream/testdata/crypto_live.jsonl
```

**API 키를 출력하거나 보고서에 넣지 않는다.** 캡처 파일에 키가 들어가지 않는지 반드시 확인한다 —
구독 프레임은 클라이언트가 보내는 것이라 수신 로그에는 없어야 하지만 grep 으로 확인한다.

- [ ] **Step 3: 캡처가 필요한 성질을 갖췄는지 검증한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go && python3 - <<'PY'
import json
rows = [json.loads(l) for l in open('stream/testdata/crypto_live.jsonl') if l.strip()]
types = {}
for r in rows:
    types[r.get('messageType')] = types.get(r.get('messageType'), 0) + 1
data = [r['data'] for r in rows if r.get('messageType') == 'A']
kinds = {}
for d in data:
    kinds[d[0]] = kinds.get(d[0], 0) + 1
print('messageType 분포:', types)
print('data[0] 분포   :', kinds)
print('T 길이         :', sorted({len(d) for d in data if d[0] == 'T'}))
print('Q 길이         :', sorted({len(d) for d in data if d[0] == 'Q'}))
q = next((d for d in data if d[0] == 'Q'), None)
if q:
    bid, mid, ask = q[5], q[6], q[8]
    print('mid 검산       :', abs((bid + ask) / 2 - mid) < 1e-6, f'({bid} {mid} {ask})')
PY
grep -c "authorization" stream/testdata/crypto_live.jsonl || echo "키 유출 없음"
```

Expected:

```
messageType 분포: {'I': 1, 'H': ..., 'A': ...}
data[0] 분포   : {'T': ..., 'Q': ...}
T 길이         : [6]
Q 길이         : [9]
mid 검산       : True (...)
키 유출 없음
```

- **`Q` 가 하나도 없으면** `thresholdLevel` 이 2 인지 확인하고 다시 캡처한다. 호가 없이는
  9원소 매핑을 검증할 수 없다.
- **`mid 검산` 이 False 면 BLOCKED 로 보고**한다. 인덱스 5·6·8 이 bid·mid·ask 라는 전제가
  깨진 것이고, 그러면 crypto 매핑 전체를 다시 확인해야 한다.
- 길이가 6·9 가 아니면 그대로 보고한다(스키마 변경 가능성).

- [ ] **Step 4: 나머지 넷의 문서 예시를 파일로 만든다**

이 넷은 실측할 수 없으므로 **문서의 예시 응답에서** 만들고 파일명에 `_doc` 을 붙여 구분한다.

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go/docs/api/websockets
grep -nA12 'Response:' forex.md | head -20
grep -nA12 'Response:' iex.md | head -20
grep -nA12 'Response:' equity-realtime-stock-data.md | head -24
grep -nA12 'Response:' boats.md | head -24
```

만들 파일: `forex_doc.jsonl`, `iex_doc.jsonl`, `equity_doc.jsonl`, `boats_doc.jsonl`.
각 파일은 해당 엔드포인트가 낼 수 있는 **모든 메시지 종류를 한 줄씩** 담는다
(equity 는 기준가 3원소와 유동성 8원소 두 줄, boats 는 Q 9원소와 T 10원소 두 줄).

문서에 예시 응답이 없어 배열 인덱스 표를 근거로 값을 구성했다면, **그 사실을 보고서에 적는다**
(JSONL 이라 파일 안에 주석을 넣을 수 없다).

- [ ] **Step 5: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add go.mod go.sum stream/testdata
git commit -m "test(stream): coder/websocket 의존성과 메시지 캡처 추가"
```

---

### Task 2: 봉투 파싱

**Files:**
- Create: `stream/envelope.go`
- Test: `stream/envelope_test.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`stream/envelope_test.go`:

```go
package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnvelope_Info(t *testing.T) {
	raw := `{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":49792869},"messageType":"I"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgInfo, e.MessageType)
	assert.Equal(t, 200, e.Response.Code)
	id, err := e.subscriptionID()
	require.NoError(t, err)
	assert.Equal(t, int64(49792869), id)
}

func TestParseEnvelope_Heartbeat(t *testing.T) {
	raw := `{"response":{"code":200,"message":"HeartBeat"},"messageType":"H"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgHeartbeat, e.MessageType)
	assert.Empty(t, e.Data, "H 에는 data 가 없다")
}

func TestParseEnvelope_Data(t *testing.T) {
	raw := `{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.0]}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgData, e.MessageType)
	assert.Equal(t, "crypto_data", e.Service)
	assert.NotEmpty(t, e.Data, "A 의 data 는 배열 원문으로 남는다")
}

// BOATS 구독 거부가 이 형태로 온다 — 실측.
func TestParseEnvelope_Error(t *testing.T) {
	raw := `{"response":{"code":403,"message":"Not authorized for service: boats"},"messageType":"E"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, msgError, e.MessageType)

	err = e.asError()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "boats")
}

// 구독 확인이 200 이 아니면 에러로 다룬다.
func TestEnvelope_InfoNonOK(t *testing.T) {
	raw := `{"response":{"code":401,"message":"Not authorized"},"messageType":"I"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	require.Error(t, e.asError())
}

// 정상 구독 확인은 에러가 아니다.
func TestEnvelope_InfoOKIsNotError(t *testing.T) {
	raw := `{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":1},"messageType":"I"}`
	e, err := parseEnvelope([]byte(raw))
	require.NoError(t, err)
	assert.NoError(t, e.asError())
}

func TestParseEnvelope_Malformed(t *testing.T) {
	_, err := parseEnvelope([]byte(`{"messageType":`))
	assert.Error(t, err)
}

// 모르는 messageType 은 에러가 아니라 무시 대상이다 — Tiingo 가 종류를 늘릴 수 있다.
func TestParseEnvelope_UnknownType(t *testing.T) {
	e, err := parseEnvelope([]byte(`{"messageType":"Z"}`))
	require.NoError(t, err)
	assert.Equal(t, messageType("Z"), e.MessageType)
	assert.False(t, e.MessageType.known())
}
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/...`
Expected: 컴파일 실패 — `undefined: parseEnvelope`

- [ ] **Step 3: 구현을 쓴다**

`stream/envelope.go`:

```go
package stream

import (
	"encoding/json"
	"fmt"
)

// messageType 은 봉투의 messageType 필드다. Tiingo 가 종류를 늘릴 수 있으므로
// 모르는 값은 에러가 아니라 무시 대상으로 다룬다.
type messageType string

const (
	msgInfo      messageType = "I" // 구독 확인. data 는 객체({"subscriptionId":...})
	msgHeartbeat messageType = "H" // 하트비트. data 없음
	msgData      messageType = "A" // 실제 데이터. data 는 위치 배열
	msgError     messageType = "E" // 에러. data 없음, response 에 사유
)

// known 은 SDK 가 다루는 종류인지 알려준다.
func (m messageType) known() bool {
	switch m {
	case msgInfo, msgHeartbeat, msgData, msgError:
		return true
	}
	return false
}

// envelope 는 모든 WebSocket 메시지의 겉껍질이다.
//
// data 의 타입이 messageType 마다 다르다 — I 는 객체, A 는 배열, H·E 는 없음.
// 그래서 RawMessage 로 받아 두고 종류에 따라 나중에 해석한다.
type envelope struct {
	Service     string          `json:"service"`     // 피드 식별자(crypto_data, cons 등). A 에만 있다
	MessageType messageType     `json:"messageType"` // I/H/A/E
	Data        json.RawMessage `json:"data"`        // 종류에 따라 객체·배열·없음
	Response    struct {
		Code    int    `json:"code"`    // HTTP 유사 상태코드
		Message string `json:"message"` // 사람이 읽는 사유
	} `json:"response"`
}

// parseEnvelope 는 프레임 한 줄을 봉투로 읽는다.
func parseEnvelope(b []byte) (*envelope, error) {
	var e envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("tiingo: malformed websocket frame: %w", err)
	}
	return &e, nil
}

// asError 는 봉투가 실패를 뜻하면 에러를, 아니면 nil 을 돌려준다.
// E 는 물론이고 200 이 아닌 구독 확인(I)도 실패로 다룬다.
func (e *envelope) asError() error {
	if e.MessageType != msgError && (e.Response.Code == 0 || e.Response.Code == 200) {
		return nil
	}
	return fmt.Errorf("tiingo: stream error (code %d): %s", e.Response.Code, e.Response.Message)
}

// subscriptionID 는 구독 확인(I)에 담긴 id 를 꺼낸다.
func (e *envelope) subscriptionID() (int64, error) {
	var d struct {
		SubscriptionID int64 `json:"subscriptionId"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil {
		return 0, fmt.Errorf("tiingo: malformed subscription ack: %w", err)
	}
	return d.SubscriptionID, nil
}
```

- [ ] **Step 4: 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/... -v`
Expected: PASS — 8건

Run: `go build ./... && go vet ./... && gofmt -l stream/` — 무출력

- [ ] **Step 5: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add stream/envelope.go stream/envelope_test.go
git commit -m "feat(stream): 봉투 파싱(messageType I/H/A/E)"
```

---

### Task 3: 엄격한 위치 배열 디코더

**Files:**
- Create: `stream/decode.go`
- Test: `stream/decode_test.go`

이 태스크가 이 계획의 안전장치다. 검증 못 한 엔드포인트가 넷이므로, 매핑이 어긋났을 때
**조용히 틀린 값**이 아니라 **시끄러운 에러**가 나오게 만드는 것이 목적이다.

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`stream/decode_test.go`:

```go
package stream

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustArr(t *testing.T, raw string, want int) *arr {
	t.Helper()
	a, err := newArr(json.RawMessage(raw), want)
	require.NoError(t, err)
	return a
}

func TestArr_Extract(t *testing.T) {
	a := mustArr(t, `["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.0]`, 6)
	assert.Equal(t, "T", a.str(0))
	assert.Equal(t, "btcusd", a.str(1))
	assert.False(t, a.time(2).IsZero())
	assert.Equal(t, "bitfinex", a.str(3))
	assert.InDelta(t, 0.0002, a.f64(4), 1e-12)
	assert.InDelta(t, 80132.0, a.f64(5), 1e-9)
	require.NoError(t, a.err())
}

// 지수 표기도 받아야 한다 — 실측에서 6.517e-5 로 온다.
func TestArr_ScientificNotation(t *testing.T) {
	a := mustArr(t, `["Q","btcusd","2026-09-06T03:56:52.172908+00:00","gdax",6.517e-5,80046.4,80046.405,0.63361423,80046.41]`, 9)
	assert.InDelta(t, 6.517e-5, a.f64(4), 1e-12)
	require.NoError(t, a.err())
}

// 짧은 배열은 에러다 — 매핑이 밀리면 값이 조용히 바뀌므로 여기서 막는다.
func TestArr_TooShort(t *testing.T) {
	_, err := newArr(json.RawMessage(`["T","btcusd","2026-09-06T03:55:54+00:00","bitfinex",0.0002]`), 6)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "6")
	assert.Contains(t, err.Error(), "5")
}

// 긴 배열은 에러가 아니다 — Tiingo 가 필드를 추가할 수 있고, 뒤에 붙는 건 무시하면 된다.
func TestArr_LongerIsOK(t *testing.T) {
	a, err := newArr(json.RawMessage(`["T","btcusd","2026-09-06T03:55:54+00:00","bitfinex",0.0002,80132.0,"extra"]`), 6)
	require.NoError(t, err)
	assert.Equal(t, "btcusd", a.str(1))
	require.NoError(t, a.err())
}

// 타입이 다르면 에러다.
func TestArr_WrongType(t *testing.T) {
	a := mustArr(t, `["T","btcusd","2026-09-06T03:55:54+00:00","bitfinex","not-a-number",80132.0]`, 6)
	a.f64(4)
	require.Error(t, a.err())
	assert.Contains(t, a.err().Error(), "index 4")
}

// 첫 에러가 보존된다 — 뒤에서 또 실패해도 원인을 잃지 않는다.
func TestArr_FirstErrorWins(t *testing.T) {
	a := mustArr(t, `["T","btcusd","bad-time","bitfinex","x","y"]`, 6)
	a.time(2)
	first := a.err()
	require.Error(t, first)
	a.f64(4)
	assert.Equal(t, first, a.err(), "첫 에러가 유지된다")
}

func TestArr_NotAnArray(t *testing.T) {
	_, err := newArr(json.RawMessage(`{"a":1}`), 3)
	assert.Error(t, err)
}

// null 은 포인터 접근자에서만 허용한다.
func TestArr_NullHandling(t *testing.T) {
	a := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1,"AAPL",null]`, 5)
	assert.Nil(t, a.f64p(4))
	require.NoError(t, a.err())

	b := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1,"AAPL",null]`, 5)
	b.f64(4)
	require.Error(t, b.err(), "값 타입 접근자는 null 을 거부한다")
}

func TestArr_Int64(t *testing.T) {
	a := mustArr(t, `["Q","2026-09-06T03:55:54+00:00",1757165400000000000,"AAPL"]`, 4)
	assert.Equal(t, int64(1757165400000000000), a.i64(2))
	require.NoError(t, a.err())
}
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/... -run TestArr`
Expected: 컴파일 실패 — `undefined: newArr`

- [ ] **Step 3: 구현을 쓴다**

`stream/decode.go`:

```go
package stream

import (
	"encoding/json"
	"fmt"

	"github.com/kenshin579/tiingo-go/types"
)

// arr 은 위치 기반 data 배열을 읽는 커서다.
//
// Tiingo WebSocket 은 데이터를 필드명이 아니라 배열 위치로 보낸다. 위치 매핑이 틀리면
// 컴파일도 테스트도 통과하고 값만 조용히 바뀌므로, 길이와 타입을 여기서 확인해 시끄럽게
// 실패시키는 것이 유일한 방어선이다. 첫 에러를 붙들고 있다가 err() 로 한 번에 돌려준다.
type arr struct {
	raw   []json.RawMessage
	first error // 처음 발생한 에러. 이후 실패는 덮어쓰지 않는다
}

// newArr 은 배열 원문을 커서로 만든다. want 보다 짧으면 에러다.
//
// want 보다 긴 것은 허용한다 — Tiingo 가 REST 에서 "필드를 계속 추가하겠다" 고 밝혔고
// WebSocket 도 같을 수 있다. 뒤에 붙는 원소는 무시한다.
func newArr(raw json.RawMessage, want int) (*arr, error) {
	var xs []json.RawMessage
	if err := json.Unmarshal(raw, &xs); err != nil {
		return nil, fmt.Errorf("tiingo: data must be a JSON array: %w", err)
	}
	if len(xs) < want {
		return nil, fmt.Errorf("tiingo: data array too short: want at least %d elements, got %d", want, len(xs))
	}
	return &arr{raw: xs}, nil
}

// fail 은 첫 에러만 기록한다.
func (a *arr) fail(i int, err error) {
	if a.first == nil {
		a.first = fmt.Errorf("tiingo: data index %d: %w", i, err)
	}
}

// err 은 지금까지 발생한 첫 에러를 돌려준다.
func (a *arr) err() error { return a.first }

// str 은 문자열 원소를 읽는다.
func (a *arr) str(i int) string {
	var v string
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return ""
	}
	return v
}

// f64 는 실수 원소를 읽는다. null 은 에러다 — 결손을 허용하려면 f64p 를 쓴다.
func (a *arr) f64(i int) float64 {
	var v float64
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return 0
	}
	return v
}

// f64p 는 실수 원소를 읽되 null 을 nil 로 돌려준다.
func (a *arr) f64p(i int) *float64 {
	if string(a.raw[i]) == "null" {
		return nil
	}
	v := a.f64(i)
	if a.first != nil {
		return nil
	}
	return &v
}

// i64 는 정수 원소를 읽는다.
func (a *arr) i64(i int) int64 {
	var v int64
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return 0
	}
	return v
}

// time 은 타임스탬프 원소를 읽는다.
func (a *arr) time(i int) types.Time {
	var v types.Time
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return types.Time{}
	}
	return v
}
```

`types.Time` 이 `UnmarshalJSON` 을 가졌는지 확인한다(`types/time.go`). 없으면 BLOCKED 로 보고한다.

- [ ] **Step 4: 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/... -v -run TestArr`
Expected: PASS — 9건

- [ ] **Step 5: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add stream/decode.go stream/decode_test.go
git commit -m "feat(stream): 위치 배열 엄격 디코더"
```

---

### Task 4: 엔드포인트별 메시지 타입

**Files:**
- Create: `stream/message.go`, `stream/crypto.go`, `stream/forex.go`, `stream/iex.go`, `stream/equity.go`, `stream/boats.go`
- Test: `stream/messages_test.go`

각 파일은 **타입 + 옵션 + 배열 → 타입 변환 함수**만 담는다. 연결은 Task 5 다.

**모든 타입의 doc 주석 첫 줄에 배열 매핑을, 둘째 줄에 실측/문서 추정 여부를 적는다.** 이건
장식이 아니라 산출물이다 — 나중에 시장이 열렸을 때 확인할 목록이 코드에 남아야 한다.

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`stream/messages_test.go`:

```go
package stream

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 실측 캡처로 crypto 매핑을 고정한다.
func TestDecodeCrypto_Trade(t *testing.T) {
	raw := json.RawMessage(`["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.00000000001]`)
	m, err := decodeCrypto(raw)
	require.NoError(t, err)

	tr, ok := m.(CryptoTrade)
	require.True(t, ok)
	assert.Equal(t, "btcusd", tr.Ticker)
	assert.Equal(t, "bitfinex", tr.Exchange)
	assert.InDelta(t, 0.0002, tr.LastSize, 1e-12)
	assert.InDelta(t, 80132.0, tr.LastPrice, 1e-6)
	assert.False(t, tr.Date.IsZero())
}

// 실측 호가. mid = (bid+ask)/2 로 매핑이 맞는지 검산한다.
func TestDecodeCrypto_Quote(t *testing.T) {
	raw := json.RawMessage(`["Q","btcusd","2026-09-06T03:56:52.172908+00:00","gdax",6.517e-5,80046.4,80046.405,0.63361423,80046.41]`)
	m, err := decodeCrypto(raw)
	require.NoError(t, err)

	q, ok := m.(CryptoQuote)
	require.True(t, ok)
	assert.Equal(t, "gdax", q.Exchange)
	assert.InDelta(t, 6.517e-5, q.BidSize, 1e-12)
	assert.InDelta(t, 80046.4, q.BidPrice, 1e-6)
	assert.InDelta(t, 0.63361423, q.AskSize, 1e-9)
	assert.InDelta(t, 80046.41, q.AskPrice, 1e-6)
	assert.InDelta(t, (q.BidPrice+q.AskPrice)/2, q.MidPrice, 1e-6, "매핑이 맞으면 mid 가 검산된다")
}

// Forex 문서의 인덱스 표는 askPrice=6 이라 적지만 문서 예시의 산술은 askPrice=7 임을 증명한다.
// 실제 문서 예시(eurnok)를 그대로 써서 못 박는다 — 값을 지어내면 검산이 무의미해진다.
func TestDecodeForex_Quote_DocIndexTableIsWrong(t *testing.T) {
	// ["Q", ticker, date, bidSize, bidPrice, midPrice, askSize, askPrice]
	raw := json.RawMessage(`["Q","eurnok","2019-07-05T15:49:15.157000+00:00",5000000.0,9.6764,9.678135,5000000.0,9.67987]`)
	m, err := decodeForex(raw)
	require.NoError(t, err)

	q, ok := m.(ForexQuote)
	require.True(t, ok)
	assert.Equal(t, "eurnok", q.Ticker)
	assert.InDelta(t, 5000000.0, q.BidSize, 1e-9)
	assert.InDelta(t, 9.6764, q.BidPrice, 1e-9)
	assert.InDelta(t, 5000000.0, q.AskSize, 1e-9, "인덱스 6 은 수량이다(표는 가격이라 적었다)")
	assert.InDelta(t, 9.67987, q.AskPrice, 1e-9, "인덱스 7 이 가격이다(표는 수량이라 적었다)")
	assert.InDelta(t, (q.BidPrice+q.AskPrice)/2, q.MidPrice, 1e-9,
		"매핑이 맞으면 mid 가 검산된다 — 이 검산이 표를 반증한 근거다")
}

// IEX 기준가는 종류 문자가 없고 3원소다.
func TestDecodeIEX_ReferencePrice(t *testing.T) {
	raw := json.RawMessage(`["2026-09-06T13:30:00.000000000+00:00","AAPL",320.08]`)
	m, err := decodeIEX(raw)
	require.NoError(t, err)

	p, ok := m.(IEXReferencePrice)
	require.True(t, ok)
	assert.Equal(t, "AAPL", p.Ticker)
	assert.InDelta(t, 320.08, p.ReferencePrice, 1e-9)
	assert.False(t, p.Date.IsZero())
}

// Equity 는 두 메시지 모두 종류 문자가 없어 길이로 구분한다.
func TestDecodeEquity_ByLength(t *testing.T) {
	ref := json.RawMessage(`["2026-09-06T13:30:00+00:00","AAPL",320.08]`)
	m, err := decodeEquity(ref)
	require.NoError(t, err)
	_, ok := m.(EquityReferencePrice)
	assert.True(t, ok, "3원소는 기준가다")

	lq := json.RawMessage(`["2026-09-06T13:30:00+00:00","AAPL",0.0005,100,320.015,320.08,320.175,200]`)
	m, err = decodeEquity(lq)
	require.NoError(t, err)
	l, ok := m.(EquityLiquidity)
	require.True(t, ok, "8원소는 유동성이다")
	assert.InDelta(t, 0.0005, l.Spread, 1e-9)
	assert.InDelta(t, 320.08, l.ReferencePrice, 1e-9)
	assert.InDelta(t, 320.015, l.BidPrice, 1e-9)
	assert.InDelta(t, 320.175, l.AskPrice, 1e-9)
}

// 어느 길이에도 맞지 않으면 에러다.
func TestDecodeEquity_UnknownLength(t *testing.T) {
	_, err := decodeEquity(json.RawMessage(`["2026-09-06T13:30:00+00:00","AAPL",1,2]`))
	assert.Error(t, err)
}

func TestDecodeBOATS_QuoteAndTrade(t *testing.T) {
	q := json.RawMessage(`["Q","2026-09-06T13:30:00+00:00",1757165400000000000,"AAPL",100,320.01,320.02,320.03,200]`)
	m, err := decodeBOATS(q)
	require.NoError(t, err)
	_, ok := m.(BOATSQuote)
	assert.True(t, ok)

	tr := json.RawMessage(`["T","2026-09-06T13:30:00+00:00",1757165400000000000,"AAPL",320.02,50,"@","","",""]`)
	m, err = decodeBOATS(tr)
	require.NoError(t, err)
	b, ok := m.(BOATSTrade)
	require.True(t, ok)
	assert.False(t, b.Break, "T 는 정상 체결")

	br := json.RawMessage(`["B","2026-09-06T13:30:00+00:00",1757165400000000000,"AAPL",320.02,50,"@","","",""]`)
	m, err = decodeBOATS(br)
	require.NoError(t, err)
	b2, ok := m.(BOATSTrade)
	require.True(t, ok)
	assert.True(t, b2.Break, "B 는 체결 취소")
}

// 모르는 종류 문자는 에러가 아니라 nil 이다 — 무시하고 다음 메시지로 간다.
func TestDecode_UnknownKindIsIgnored(t *testing.T) {
	m, err := decodeCrypto(json.RawMessage(`["Z","btcusd","2026-09-06T03:55:54+00:00","x",1,2]`))
	require.NoError(t, err)
	assert.Nil(t, m)
}

// 짧은 배열은 에러로 올라온다.
func TestDecode_ShortArrayErrors(t *testing.T) {
	_, err := decodeCrypto(json.RawMessage(`["T","btcusd"]`))
	assert.Error(t, err)
}
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/...`
Expected: 컴파일 실패 — `undefined: decodeCrypto` 등

- [ ] **Step 3: 구현을 쓴다**

`stream/message.go`:

```go
package stream

// Message 는 스트림이 내보내는 데이터 메시지다. 타입 스위치로 구분한다.
//
// 봉투의 구독 확인(I)·하트비트(H)는 내부에서 소화되며 채널로 나오지 않는다.
// 에러(E)는 Stream.Err() 로 흐른다.
type Message interface{ isMessage() }
```

`stream/crypto.go`:

```go
package stream

import (
	"encoding/json"
	"fmt"

	"github.com/kenshin579/tiingo-go/types"
)

// CryptoThreshold 는 crypto 피드의 구독 수준이다.
type CryptoThreshold int

const (
	CryptoTradesAndQuotes CryptoThreshold = 2 // 호가 + 체결
	CryptoTradesOnly      CryptoThreshold = 5 // 체결만
)

// CryptoOptions 는 crypto 스트림 구독 인자다.
type CryptoOptions struct {
	Tickers   []string        // 구독할 페어(btcusd 등). 비우면 전체
	Threshold CryptoThreshold // 0 이면 CryptoTradesOnly(5)
}

// CryptoTrade 는 암호화폐 체결이다.
//
// 배열: ["T", ticker, date, exchange, lastSize, lastPrice] (6개)
// 이 매핑은 실호출로 검증했다(2026-09-06).
type CryptoTrade struct {
	Ticker    string     // 페어. 소문자로 온다
	Date      types.Time // 체결 시각
	Exchange  string     // 체결 거래소(bitfinex, gdax, bitstamp 등)
	LastSize  float64    // 기준 통화 기준 체결 수량
	LastPrice float64    // 체결가
}

func (CryptoTrade) isMessage() {}

// CryptoQuote 는 암호화폐 호가다. Threshold 가 CryptoTradesAndQuotes 일 때만 온다.
//
// 배열: ["Q", ticker, date, exchange, bidSize, bidPrice, midPrice, askSize, askPrice] (9개)
// 이 매핑은 실호출로 검증했다(2026-09-06, mid=(bid+ask)/2 검산 일치).
type CryptoQuote struct {
	Ticker   string     // 페어
	Date     types.Time // 호가 시각
	Exchange string     // 거래소
	BidSize  float64    // 매수 수량
	BidPrice float64    // 매수 호가
	MidPrice float64    // 중간가. (BidPrice+AskPrice)/2
	AskSize  float64    // 매도 수량
	AskPrice float64    // 매도 호가
}

func (CryptoQuote) isMessage() {}

// decodeCrypto 는 crypto data 배열을 메시지로 바꾼다.
// 모르는 종류 문자면 (nil, nil) 을 돌려준다 — 무시하고 넘어가라는 뜻이다.
func decodeCrypto(raw json.RawMessage) (Message, error) {
	kind, err := arrKind(raw)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "T":
		a, err := newArr(raw, 6)
		if err != nil {
			return nil, err
		}
		m := CryptoTrade{
			Ticker:    a.str(1),
			Date:      a.time(2),
			Exchange:  a.str(3),
			LastSize:  a.f64(4),
			LastPrice: a.f64(5),
		}
		return m, a.err()
	case "Q":
		a, err := newArr(raw, 9)
		if err != nil {
			return nil, err
		}
		m := CryptoQuote{
			Ticker:   a.str(1),
			Date:     a.time(2),
			Exchange: a.str(3),
			BidSize:  a.f64(4),
			BidPrice: a.f64(5),
			MidPrice: a.f64(6),
			AskSize:  a.f64(7),
			AskPrice: a.f64(8),
		}
		return m, a.err()
	}
	return nil, nil
}

// arrKind 는 배열의 첫 원소(종류 문자)를 읽는다.
// 종류 문자가 없는 피드(IEX·Equity)에서는 쓰지 않는다.
func arrKind(raw json.RawMessage) (string, error) {
	var xs []json.RawMessage
	if err := json.Unmarshal(raw, &xs); err != nil {
		return "", fmt.Errorf("tiingo: data must be a JSON array: %w", err)
	}
	if len(xs) == 0 {
		return "", fmt.Errorf("tiingo: data array is empty")
	}
	var k string
	if err := json.Unmarshal(xs[0], &k); err != nil {
		return "", fmt.Errorf("tiingo: data index 0 (kind): %w", err)
	}
	return k, nil
}
```

`stream/forex.go`, `stream/iex.go`, `stream/equity.go`, `stream/boats.go` 도 같은 형태로 만든다.
각 배열 매핑은 아래 표를 따른다. **이 표는 문서의 인덱스 숫자가 아니라 문서 예시의 산술로
검증한 결과다** — Forex 는 문서 인덱스 표가 틀렸으므로 아래 표를 그대로 쓴다.

| 파일 | 타입 | 배열 |
|---|---|---|
| `forex.go` | `ForexQuote` | `["Q", ticker, date, bidSize, bidPrice, midPrice, askSize, askPrice]` (8) |
| `iex.go` | `IEXReferencePrice` | `[date, ticker, refPrice]` (3, 종류 문자 없음) |
| `equity.go` | `EquityReferencePrice` | `[date, ticker, refPrice]` (3) |
| `equity.go` | `EquityLiquidity` | `[date, ticker, spread, bidSize, bidPrice, refPrice, askPrice, askSize]` (8) |
| `boats.go` | `BOATSQuote` | `["Q", date, nanos, ticker, bidSize, bidPrice, midPrice, askPrice, askSize]` (9) |
| `boats.go` | `BOATSTrade` | `["T"\|"B", date, nanos, ticker, lastPrice, lastSize, sc1, sc2, sc3, sc4]` (10) |

세부 규칙:

- `IEXReferencePrice`·`Equity*`·`BOATS*` 의 doc 주석 둘째 줄은 **"이 매핑은 문서 기반이며
  실호출로 검증하지 못했다"** 로 적고, 왜인지(시장 시간 / 계정 권한 403) 덧붙인다.
- `ForexQuote` 는 다르게 적는다 — **문서 인덱스 표가 틀렸고 예시의 산술로 바로잡았다**는 사실과,
  Crypto 와 같은 순서라는 점을 주석에 남긴다. 라이브 검증은 여전히 남아 있다.
- `IEX`·`Equity` 는 종류 문자가 없으므로 `arrKind` 를 쓰지 않는다. `decodeEquity` 는 **배열 길이**로
  3 → 기준가, 8 → 유동성으로 가르고, 둘 다 아니면 에러다. `decodeIEX` 는 3원소만 받는다.
- `BOATSTrade.Break` 는 종류 문자가 `"B"` 일 때 true 다. 판매 조건 4개는 `[4]string` 이 아니라
  `SaleConditions []string` 으로 두고 빈 문자열도 그대로 담는다(문서상 미사용 자리는 `""`).
- `BOATSQuote.MidPrice` 는 한쪽 호가가 없으면 null 이라고 문서가 명시하므로 `*float64` 다.
  나머지 BOATS 필드는 값 타입으로 둔다.
- 옵션 타입도 각각 만든다(`ForexOptions`·`IEXOptions`·`EquityOptions`·`BOATSOptions`).
  Threshold 상수: `ForexTradesOnly = 5`, `IEXReferencePriceLevel = 6`,
  `EquityReferencePriceLevel = 6`, `EquityLiquidityLevel = 4`, `BOATSLevel = 6`.

- [ ] **Step 4: 통과를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/... -v`
Expected: PASS — Task 2·3 것 포함 전부

Run: `go build ./... && go vet ./... && gofmt -l stream/` — 무출력

- [ ] **Step 5: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add stream/message.go stream/crypto.go stream/forex.go stream/iex.go stream/equity.go stream/boats.go stream/messages_test.go
git commit -m "feat(stream): 엔드포인트 5종 메시지 타입과 배열 매핑"
```

---

### Task 5: 연결·구독·재연결

**Files:**
- Create: `stream/conn.go`, `stream/client.go`
- Test: `stream/testserver_test.go`, `stream/conn_test.go`

- [ ] **Step 1: 스텁 서버를 만든다**

`/Users/user/src/workspace_moneyflow/toss-go/stream/testserver_test.go` 를 먼저 읽고 구조를 참고한다
(복사하지 말고 이 API 에 맞게 새로 쓴다 — 토스는 선언형 구독이라 ack 흐름이 다르다).

`stream/testserver_test.go` 가 갖춰야 할 것:

- 실제 `websocket.Accept` 로 업그레이드한다.
- 클라이언트가 보낸 텍스트 프레임을 기록하고 `received() []string` 으로 노출한다.
- 구독 프레임을 받으면 자동으로
  `{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":1},"messageType":"I"}`
  를 돌려준다. `setAutoAck(false)` 로 끌 수 있어야 한다.
- 테스트가 임의 프레임을 밀어 넣는 `push(string)` 을 제공한다.
- 현재 연결을 강제로 끊는 `dropConn()` 을 제공한다(재연결 검증용).
- 총 연결 수 `conns() int` 를 노출한다.
- 모든 상태 접근은 뮤텍스로 보호한다 — `-race` 로 돌린다.

- [ ] **Step 2: 실패하는 테스트를 쓴다**

`stream/conn_test.go`:

```go
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

	s, err := c.Crypto(ctx, &CryptoOptions{Tickers: []string{"btcusd"}, Threshold: CryptoTradesAndQuotes})
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

// 구독 확인의 id 를 노출한다.
func TestStream_SubscriptionID(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return s.SubscriptionID() != 0 }, 3*time.Second, 20*time.Millisecond)
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

// E 는 Err() 로 흐르고 채널이 닫힌다.
func TestStream_ErrorClosesStream(t *testing.T) {
	ts := newTestServer(t)
	ts.setAutoAck(false)
	c := New("k", WithBaseURL(ts.url))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.BOATS(ctx, nil)
	require.NoError(t, err)
	ts.push(`{"response":{"code":403,"message":"Not authorized for service: boats"},"messageType":"E"}`)

	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	require.Error(t, s.Err())
	assert.Contains(t, s.Err().Error(), "403")
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

	for i := 0; i < 50; i++ {
		ts.push(`{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54+00:00","x",1,2]}`)
	}
	assert.Eventually(t, func() bool { return s.Dropped() > 0 }, 3*time.Second, 20*time.Millisecond)
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
	cancel()

	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	assert.ErrorIs(t, s.Err(), context.Canceled)
}

// Close 후에는 재연결하지 않는다.
func TestStream_CloseStopsReconnect(t *testing.T) {
	ts := newTestServer(t)
	c := New("k", WithBaseURL(ts.url), WithBackoff(10*time.Millisecond, 20*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := c.Crypto(ctx, nil)
	require.NoError(t, err)
	assert.Eventually(t, func() bool { return ts.conns() == 1 }, 3*time.Second, 20*time.Millisecond)

	require.NoError(t, s.Close())
	before := ts.conns()
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, before, ts.conns(), "Close 후에는 다시 붙지 않는다")
}
```

- [ ] **Step 3: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test ./stream/...`
Expected: 컴파일 실패 — `undefined: New`, `undefined: WithBaseURL` 등

- [ ] **Step 4: 구현한다**

`stream/client.go` 가 담을 것:

- 패키지 doc. **어느 엔드포인트가 검증됐고 어느 것이 문서 기반인지 반드시 적는다.**
  BOATS 가 계정 권한 403 이라는 사실도 적는다.
- `Client` 구조체: apiKey, baseURL, 버퍼 크기, backoff min/max, `*http.Client`.
- `New(apiKey string, opts ...Option) *Client`
- 옵션: `WithBaseURL(string)`(테스트용), `WithBuffer(int)`, `WithBackoff(min, max time.Duration)`,
  `WithHTTPClient(*http.Client)`.
- 기본값: 버퍼 256, backoff min 1초 / max 30초.
- 5개 진입 메서드. 각각 경로·thresholdLevel 기본값·decode 함수를 묶어 공통 `open` 을 호출한다:

```go
func (c *Client) Crypto(ctx context.Context, opts *CryptoOptions) (*Stream, error) {
	th := CryptoTradesOnly
	var tickers []string
	if opts != nil {
		if opts.Threshold != 0 {
			th = opts.Threshold
		}
		tickers = opts.Tickers
	}
	return c.open(ctx, "/crypto", eventData{ThresholdLevel: int(th), Tickers: tickers}, decodeCrypto)
}
```

`stream/conn.go` 가 담을 것:

- `Stream` 구조체와 `Messages()`/`Err()`/`Close()`/`SubscriptionID()`/`Dropped()`.
  `SubscriptionID`·`Dropped`·`Err` 은 다른 고루틴이 쓰므로 뮤텍스나 atomic 으로 보호한다.
- `open` — 고루틴을 띄우고 즉시 `*Stream` 을 돌려준다. **첫 연결을 기다리지 않는다**
  (기다리면 호출이 블로킹돼 사용도 테스트도 불편하다). 연결 실패는 `Err()` 로 흐른다.
- 읽기 루프: 프레임 → `parseEnvelope` → 종류별 분기
  - `I`: `subscriptionID` 저장. `asError()` 가 에러면 스트림 종료.
  - `H`: 무시.
  - `A`: decode 함수 호출. `(nil, nil)` 이면 무시. **에러면 스트림을 끝내지 말고 건너뛴다**
    (한 메시지가 깨졌다고 연결 전체를 죽일 이유가 없다).
  - `E`: `asError()` 를 최종 에러로 두고 종료. **재연결하지 않는다** — 권한 문제는 재시도해도
    같은 결과다.
  - 모르는 타입: 무시.
- 채널 전송은 **논블로킹**이다. 가득 차면 가장 오래된 것을 하나 버리고 새 것을 넣는다.
  버린 수는 atomic 으로 센다.
- 재연결: 읽기 루프가 (E 가 아닌) 에러로 끝나면 backoff 후 다시 연결하고 구독 프레임을
  재전송한다. `ctx` 가 끝났거나 `Close()` 가 불렸으면 재연결하지 않는다.
- `backoff(attempt, min, max)`: 지수 + ±20% jitter. 첫 시도는 즉시.
  `toss-go/stream/conn.go` 의 것과 같은 형태로 만들되 직접 작성한다.
- 인증은 **구독 프레임의 `authorization` 필드**로 보낸다. URL 쿼리나 헤더가 아니다(실측 확인).
- 종료 시 `Messages()` 채널을 닫는다. 소비자가 `for range` 로 빠져나올 수 있어야 한다.

- [ ] **Step 5: 통과를 확인한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go test ./stream/... -v
go test ./stream/... -race          # 고루틴이 있으므로 race 검사가 필수다
go build ./... && go vet ./... && gofmt -l stream/
```

Expected: 전부 PASS, race 없음.

- [ ] **Step 6: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add stream/client.go stream/conn.go stream/conn_test.go stream/testserver_test.go
git commit -m "feat(stream): 연결·구독·재연결과 채널 소비"
```

---

### Task 6: 루트 배선과 예제

**Files:**
- Modify: `client.go`, `client_test.go`
- Create: `examples/stream/main.go`

- [ ] **Step 1: 실패하는 테스트를 쓴다**

`client_test.go` 의 `TestNewClient` 안, `c.CorporateActions` 단정 다음 줄에:

```go
	assert.NotNil(t, c.Stream)
```

- [ ] **Step 2: 실패를 확인한다**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && go test . -run TestNewClient`
Expected: 컴파일 실패 — `c.Stream undefined`

- [ ] **Step 3: 배선한다**

`client.go` 에 import·필드·생성을 더한다. **`stream.New` 는 `*httpclient.Client` 가 아니라
apiKey 를 받는다**(WebSocket 은 HTTP 클라이언트를 공유하지 않는다).

구조체 필드(`CorporateActions` 다음 줄):

```go
	Stream *stream.Client // 실시간 WebSocket 스트리밍(REST 와 별개 연결)
```

`NewClient` 안(`c.CorporateActions = ...` 다음 줄):

```go
	c.Stream = stream.New(apiKey)
```

`gofmt -w client.go` 로 정렬을 맞춘다. 필드 이름 길이가 바뀌면 gofmt 가 블록 전체의 주석 정렬을
다시 맞추므로 diff 에 다른 줄이 함께 뜨는 것은 정상이다.

- [ ] **Step 4: 예제를 만든다**

`examples/stream/main.go` — crypto 를 쓴다(24시간 거래라 언제 돌려도 데이터가 온다):

```go
// Tiingo WebSocket 스트리밍 예제.
//
//	TIINGO_API_KEY=... go run ./examples/stream
//
// crypto 는 24시간 거래되므로 언제 돌려도 메시지가 온다.
// 주식·FX 는 시장 시간에만 데이터가 흐른다.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/stream"
)

func main() {
	c, err := tiingo.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := c.Stream.Crypto(ctx, &stream.CryptoOptions{
		Tickers:   []string{"btcusd", "ethusd"},
		Threshold: stream.CryptoTradesAndQuotes,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	fmt.Println("15초간 수신합니다...")
	var trades, quotes int
	for msg := range s.Messages() {
		switch m := msg.(type) {
		case stream.CryptoTrade:
			trades++
			if trades <= 3 {
				fmt.Printf("  체결 %-7s %-9s %.8f @ %.2f\n", m.Ticker, m.Exchange, m.LastSize, m.LastPrice)
			}
		case stream.CryptoQuote:
			quotes++
			if quotes <= 3 {
				fmt.Printf("  호가 %-7s %-9s %.2f / %.2f (중간 %.2f)\n",
					m.Ticker, m.Exchange, m.BidPrice, m.AskPrice, m.MidPrice)
			}
		}
	}
	fmt.Printf("체결 %d건, 호가 %d건 (누락 %d건)\n", trades, quotes, s.Dropped())
	if err := s.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
}
```

- [ ] **Step 5: 빌드하고 실행한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && (gofmt -l . | grep -v '^$' || echo "포맷 OK")
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go run ./examples/stream
```

Expected: 15초 동안 체결·호가가 각각 여러 건. **누락 0건**이어야 정상이다(콘솔 출력만 하는데
버려질 이유가 없다). **API 키를 출력하거나 보고서에 넣지 않는다.**

`go test ./...` 도 통과해야 한다.

- [ ] **Step 6: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add client.go client_test.go examples/stream/main.go
git commit -m "feat(stream): 루트 클라이언트 배선과 예제 추가"
```

---

### Task 7: 통합 테스트

**Files:**
- Modify: `integration_test.go`

- [ ] **Step 1: 실 API 테스트를 추가한다**

crypto 와 BOATS(403 경로)만 넣는다 — Forex·IEX·Equity 는 시장 시간에 묶여 있어 CI 에서
불안정해진다. import 에 `"github.com/kenshin579/tiingo-go/stream"` 을 순서에 맞게 추가한다.

```go
// crypto 는 24시간 거래되므로 언제 돌려도 메시지가 온다.
func TestIntegration_StreamCrypto(t *testing.T) {
	c := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s, err := c.Stream.Crypto(ctx, &stream.CryptoOptions{
		Tickers:   []string{"btcusd", "ethusd"},
		Threshold: stream.CryptoTradesAndQuotes,
	})
	require.NoError(t, err)
	defer s.Close()

	var got int
	for msg := range s.Messages() {
		switch m := msg.(type) {
		case stream.CryptoTrade:
			assert.NotEmpty(t, m.Ticker)
			assert.NotEmpty(t, m.Exchange)
			assert.Greater(t, m.LastPrice, 0.0)
			assert.False(t, m.Date.IsZero())
		case stream.CryptoQuote:
			assert.NotEmpty(t, m.Ticker)
			assert.Greater(t, m.BidPrice, 0.0)
			assert.Greater(t, m.AskPrice, 0.0)
			assert.InDelta(t, (m.BidPrice+m.AskPrice)/2, m.MidPrice, 1e-6,
				"매핑이 맞으면 mid 가 검산된다")
		}
		got++
		if got >= 5 {
			break
		}
	}
	assert.GreaterOrEqual(t, got, 1, "20초 안에 최소 한 건은 온다")
}

// 구독 확인 id 가 온다 — 인증이 실제로 통했다는 증거다.
func TestIntegration_StreamSubscriptionID(t *testing.T) {
	c := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := c.Stream.Crypto(ctx, &stream.CryptoOptions{Tickers: []string{"btcusd"}})
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return s.SubscriptionID() != 0 },
		10*time.Second, 100*time.Millisecond)
}

// BOATS 는 계정 권한이 없어 403 이 온다. 그 경로가 에러로 잘 흐르는지 확인한다.
func TestIntegration_StreamBOATSForbidden(t *testing.T) {
	c := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := c.Stream.BOATS(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	require.Error(t, s.Err(), "권한이 없으므로 에러로 끝난다")
	assert.Contains(t, s.Err().Error(), "403")
}
```

**`TestIntegration_StreamBOATSForbidden` 이 실패하면 그것은 좋은 소식일 수 있다** —
계정에 BOATS 권한이 생겼다는 뜻이므로, 실패 내용을 그대로 보고한다. 테스트를 약화하지 않는다.

- [ ] **Step 2: 실 API 로 실행한다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration -run TestIntegration_Stream -v .
```

Expected: 3건 PASS.

- [ ] **Step 3: 전체 통합 테스트를 돌린다**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
KEY=$(grep -m1 -E '^[[:space:]]*export[[:space:]]+TIINGO_API_KEY=' ~/.zshrc | sed -E 's/^[^=]*=//; s/^["'"'"']//; s/["'"'"']$//')
TIINGO_API_KEY="$KEY" go test -tags integration . 2>&1 | tail -5
```

Expected: ok — 35건(기존 32 + 3). REST 쪽이 429 로 막히면 정시까지 기다렸다 다시 돌린다.

- [ ] **Step 4: 비통합 빌드 확인**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && go test ./... -race && (gofmt -l . | grep -v '^$' || echo "포맷 OK")
```

- [ ] **Step 5: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add integration_test.go
git commit -m "test(stream): 실 API 통합 테스트 3건 추가"
```

---

### Task 8: 문서 갱신

**Files:**
- Modify: `README.md`, `/Users/user/src/workspace_moneyflow/CLAUDE.md`

- [ ] **Step 1: README 에 WebSocket 절을 신설한다**

REST 커버리지 표에 섞지 않는다 — 요청-응답과 스트리밍이 같은 것처럼 읽힌다.
`## 커버리지` 절 다음, `## 날짜 타입` 앞에 새 절을 넣는다:

````markdown
## WebSocket

실시간 스트리밍은 REST 와 별개 연결이며 `c.Stream` 으로 접근한다. 채널로 소비하고,
채널이 닫힌 뒤 `Err()` 로 종료 사유를 확인한다.

```go
s, _ := c.Stream.Crypto(ctx, &stream.CryptoOptions{
    Tickers:   []string{"btcusd"},
    Threshold: stream.CryptoTradesAndQuotes,
})
defer s.Close()

for msg := range s.Messages() {
    switch m := msg.(type) {
    case stream.CryptoTrade:
        fmt.Println(m.Ticker, m.Exchange, m.LastPrice)
    case stream.CryptoQuote:
        fmt.Println(m.Ticker, m.BidPrice, m.AskPrice)
    }
}
```

| 피드 | 메서드 | 엔드포인트 | 배열 매핑 검증 |
| --- | --- | --- | --- |
| Crypto | `Stream.Crypto` | `wss://api.tiingo.com/crypto` | 실호출 |
| Forex | `Stream.Forex` | `wss://api.tiingo.com/fx` | 문서 |
| IEX | `Stream.IEX` | `wss://api.tiingo.com/iex` | 문서 |
| Equity Realtime | `Stream.Equity` | `wss://api.tiingo.com/equity/intraday` | 문서 |
| BOATS | `Stream.BOATS` | `wss://api.tiingo.com/boats` | 불가(계정 권한 403) |

데이터가 필드명이 아니라 **배열 위치**로 오므로, 배열이 기대보다 짧거나 타입이 다르면 값을
조용히 바꾸는 대신 에러를 낸다. 기대보다 긴 배열은 통과시킨다(Tiingo 가 필드를 추가할 수 있다).

Crypto 외 네 피드의 배열 매핑은 문서 기반이다 — 조사 시점에 FX·미국 주식 시장이 닫혀 있었고
BOATS 는 계정 권한이 없다. 각 타입 주석에 검증 여부를 적어 두었다.

연결이 끊기면 지수 backoff 로 자동 재연결하고 구독을 재전송한다. 소비가 느려 버퍼가 차면
가장 오래된 메시지를 버리며, 누락 건수는 `s.Dropped()` 로 확인한다.
````

- [ ] **Step 2: 나머지 README 손질**

- "실행 가능한 예시" 줄에 `` , [`examples/stream`](examples/stream) `` 추가.
- `## 테스트` 의 예제 빌드 목록에 `go build -o /dev/null ./examples/stream` 추가하고,
  이름 충돌 열거 주석에 `stream` 을 더한다(루트에 `stream/` 이 생겼다).
- 커버리지 표 아래 미구현 문장을 고친다. 실제 문장을 읽고 정확히 편집한다 — 현재는
  "남은 것은 WebSocket 이다" 로 끝나는데, 이제 구현됐으므로 다음으로 바꾼다:

```markdown
나머지 REST 그룹은 이 계정 권한으로 접근이 막혀 있다(2026-09-05 실측) — News 와 Fund Fees 는
403 권한 없음, BOATS 는 유료 add-on, Corporate Actions 의 배당 내역·분할도 403 이다.
WebSocket 5종은 아래 절에 있다.
```

- [ ] **Step 3: 워크스페이스 CLAUDE.md 를 갱신한다**

**git 저장소가 아니므로 커밋하지 않는다.**

- `**Module**:` 문단의 버전 문장을 `**v0.9.0 (2026-09-06)**` 으로 바꾸고, 기존 여덟 카테고리
  열거를 유지한 채 뒤에 WebSocket 문장을 붙인다:
  `실시간은 c.Stream → WebSocket 5종(Crypto·Forex·IEX·Equity·BOATS, 채널 소비 + 자동 재연결;
  배열 매핑 실측 검증은 Crypto 만, BOATS 는 계정 권한 403).`
- bash 블록의 예제 빌드 주석 충돌 목록에 `stream/` 을 더한다.
- `**다음**:` 을 다음으로 바꾼다:
  `**다음**: moneyflow 통합. Tiingo 쪽은 REST·WebSocket 모두 계정 권한 한도까지 구현됐다.
  Forex·IEX·Equity 스트림 배열 매핑은 시장이 열렸을 때 실측 확인이 남아 있다.`

- [ ] **Step 4: 최종 검증**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
go build ./... && go vet ./... && go test ./... -race && (gofmt -l . | grep -v '^$' || echo "포맷 OK")
go build -o /dev/null ./examples/stream && echo "예제 빌드 OK"
file -I README.md /Users/user/src/workspace_moneyflow/CLAUDE.md
grep -n "WebSocket" README.md | head
```

두 파일 모두 `charset=utf-8` 이어야 한다.

- [ ] **Step 5: 커밋한다**

```
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add README.md
git commit -m "docs(stream): README WebSocket 절 신설"
```

---

## 완료 후

`superpowers:finishing-a-development-branch` 로 PR 을 만든다. 머지 후 v0.9.0 릴리스:

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go && ./scripts/release.sh v0.9.0
```

릴리스 뒤 신규 모듈에서 소비를 확인한다:

```bash
cd $(mktemp -d) && go mod init probe && GOPROXY=https://proxy.golang.org go get github.com/kenshin579/tiingo-go@v0.9.0
```

## 범위 밖

- IEX thresholdLevel 0/5(TOPS) — IEX Exchange 협약 필요
- 스트림 중간 구독 변경(티커 추가·제거) — 고정 구독으로 시작한다
- Forex·IEX·Equity 배열 매핑의 실측 확인 — 시장이 열렸을 때
- moneyflow 통합, REST 재시도·rate limit 백오프
