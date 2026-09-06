# Tiingo Go SDK — WebSocket 스트리밍 설계

- 작성일: 2026-09-06
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/websocket`)
- 토픽: Tiingo WebSocket 5종 실시간 스트리밍 — REST 8개 카테고리에 이은 새 축
- 선행: v0.8.0 (REST 8 카테고리, 권한 장벽으로 소진)
- API 참조: `docs/api/websockets/{crypto,forex,iex,equity-realtime-stock-data,boats}.md`

## 배경 / 목적

REST 는 권한 장벽으로 끝났다(News·Fund Fees 403, BOATS 유료, Corporate Actions 4/5 403).
남은 건 WebSocket 이다. 그리고 **무료 키로 5개 중 4개가 열린다** — 실측으로 확인했다.

이건 REST 확장이 아니라 **새로운 축**이다. 요청-응답이 아니라 연결 수명·구독 상태·재연결·역압을
다뤄야 하고, 무엇보다 **데이터가 위치 기반 배열로 온다.**

## 사전 조사 결과 (실호출 확인, 2026-09-06)

### 접근 권한

| 엔드포인트 | 구독 결과 |
|---|---|
| `wss://api.tiingo.com/crypto` | **200** `{"subscriptionId":49797062}` |
| `wss://api.tiingo.com/fx` | **200** |
| `wss://api.tiingo.com/iex` | **200** (thresholdLevel 6) |
| `wss://api.tiingo.com/equity/intraday` | **200** (thresholdLevel 6) |
| `wss://api.tiingo.com/boats` | **403** `{"code":403,"message":"Not authorized for service: boats"}` |

문서가 "With Tiingo Free, Power, Commercial, and Redistribution plans all come with access to the
firehose" 라고 명시한다. IEX 는 거래소 협약 없이 thresholdLevel 6(Tiingo 자체 산출 기준가)을 쓴다 —
0/5 는 IEX Exchange 와 별도 협약이 필요하다(2025-02-01 정책 변경).

### 봉투(envelope) — 네 가지 messageType

```json
{"response":{"code":200,"message":"Success"},"data":{"subscriptionId":49792869},"messageType":"I"}
{"response":{"code":200,"message":"HeartBeat"},"messageType":"H"}
{"service":"crypto_data","messageType":"A","data":["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.0]}
{"response":{"code":403,"message":"Not authorized for service: boats"},"messageType":"E"}
```

**`data` 필드의 타입이 messageType 마다 다르다** — `I` 는 객체, `H` 는 없음, `A` 는 배열, `E` 는 없음.
단일 구조체로 받으면 `data` 를 `json.RawMessage` 로 두고 분기해야 한다.

### 데이터 배열 — 엔드포인트마다 다르고, 어떤 건 종류 문자가 없다

**Crypto** (`service: "crypto_data"`) — **실측 검증 완료**

```
["T", ticker, date, exchange, lastSize, lastPrice]                                   6개
["Q", ticker, date, exchange, bidSize, bidPrice, midPrice, askSize, askPrice]        9개
```

실측 예: `["Q","btcusd",ts,"gdax",6.517e-5,80046.4,80046.405,0.63361423,80046.41]`
→ `mid = (80046.4 + 80046.41)/2 = 80046.405` 로 검산 일치. 지수 표기(`6.517e-5`)도 섞여 온다.
thresholdLevel: `2` = 호가+체결, `5` = 체결만.

**Forex** — 라이브 미검증(조사 시점 FX 휴장)이나 **문서 예시로 산술 검증함**

```
["Q", ticker, date, bidSize, bidPrice, midPrice, askSize, askPrice]                  8개
```

**⚠ 문서의 인덱스 표가 틀렸다.** 표는 `Ask Size`를 7, `Ask Price`를 6 이라 적지만, **같은 문서의
예시 응답이 그 반대임을 증명한다**:

```
["Q","eurnok","2019-07-05T15:49:15.157000+00:00", 5000000.0, 9.6764, 9.678135, 5000000.0, 9.67987]
                                                   bidSize   bidPrice  mid      askSize   askPrice
```

`idx6 = 5000000.0` 은 명백히 수량이고, `mid = (bidPrice + askPrice)/2` 검산은 **`idx7` 을
매도호가로 놓을 때만** 맞는다(`(9.6764 + 9.67987)/2 = 9.678135`). 두 번째 예시(gbpaud)도 같다.

즉 **Forex 는 Crypto 와 같은 순서**(bidSize, bidPrice, mid, askSize, askPrice)다.
문서 표의 행 출력 순서가 오히려 옳고 인덱스 숫자가 뒤바뀌어 인쇄됐다.
**표의 숫자가 아니라 예시의 산술을 따른다.**

같은 방식으로 Equity 유동성·BOATS 호가도 검산했고 그쪽은 표와 예시가 일치한다.

**IEX** — 미검증(조사 시점 미국 장 마감)

```
thresholdLevel 6 (기준가):  [date, ticker, refPrice]                                 3개, 종류 문자 없음
thresholdLevel 0/5 (TOPS):  [type, date, nanoseconds, ticker, bidSize, bidPrice,
                             midPrice, askPrice, askSize, lastPrice, lastSize]      11개
```

TOPS 는 IEX Exchange 협약이 필요해 이 계정으로는 받을 수 없다. **기준가만 구현한다.**

**Equity Realtime** (`service: "cons"`) — 미검증(조사 시점 미국 장 마감)

```
thresholdLevel 6 (기준가):   [date, ticker, refPrice]                                3개
thresholdLevel 4 (유동성):   [date, ticker, lqSpread, lqBidSize, lqBidPrice,
                              refPrice, lqAskPrice, lqAskSize]                       8개
```

**둘 다 종류 문자가 없다.** 구분은 구독한 thresholdLevel 과 배열 길이(3 vs 8)로 한다.

**BOATS** — **검증 불가**(403, 영구)

```
["Q", date, nanoseconds, ticker, bidSize, bidPrice, midPrice, askPrice, askSize]     9개
["T"|"B", date, nanoseconds, ticker, lastPrice, lastSize, sc1, sc2, sc3, sc4]       10개
```

`"B"` 는 체결 취소(trade break). 판매 조건 4개는 문자이며 미사용 자리는 빈 문자열이다.

## 결정 사항 (브레인스토밍)

1. **범위는 5개 전부**(사용자 결정). BOATS 는 403 이라 실행 검증이 불가능하고, Forex·IEX·Equity 는
   조사 시점 시장이 닫혀 있어 문서 기반이다. 위험을 **엄격 디코딩**으로 상쇄한다(아래 5번).
   (제안했던 축소안 — Crypto 만 먼저 — 은 사용자가 기각했다.)
2. **소비는 채널**(사용자 결정). `s.Messages() <-chan Message` + 종료 후 `s.Err()`.
   (기각: 콜백 — 제어가 뒤집히고 배압을 다루기 어렵다. 이터레이터 — 여러 스트림 합성이 어렵다.)
   근거: 실제 용도가 "받아서 다른 신호와 함께 처리" 라 `select` 가 되는 게 중요하다.
3. **단일 `stream` 패키지.** 다섯 엔드포인트가 연결·인증·구독·하트비트·재연결·봉투 파싱을 전부
   공유하고 다른 건 배열 해석뿐이다. 카테고리별로 나누면 그 기계장치를 다섯 번 복제하게 된다.
   (이 레포가 복제를 택해 온 건 작은 구조체였지 이런 로직이 아니다.)
   루트 노출은 `c.Stream` — 기존 `c.Crypto`(REST)에 붙이면 REST 패키지가 WebSocket 의존성을 끈다.
4. **`I`·`H` 는 내부에서 소화하고 채널로 내보내지 않는다.** 사용자는 실제 데이터만 받는다.
   `E` 는 `Err()` 로 흐른다. 구독 확인의 `subscriptionId` 는 `s.SubscriptionID()` 로 노출한다.
5. **위치 배열을 엄격하게 읽는다.** 길이와 각 원소 타입을 확인하고 어긋나면 에러를 낸다.
   검증 못 한 엔드포인트가 넷이라 이게 유일한 안전장치다 — 잘못된 위치 매핑은 컴파일도 테스트도
   통과하고 값만 조용히 틀린다. 각 매핑 주석에 **실측인지 문서 추정인지** 표시한다.
6. **재연결은 자동.** 지수 backoff 로 다시 붙고 구독을 재전송한다. 네트워크가 한 번 끊겼다고
   스트림이 조용히 죽으면 시세 소비자에게 치명적이다. `toss-go` 의 `stream.Stream` 과 같은 선택.
7. **의존성은 `coder/websocket`**, `toss-go` 와 같은 버전(v1.8.14). 이 라이브러리의 첫 비테스트
   외부 의존성이다. `gorilla/websocket` 은 유지보수가 멈춘 시기가 있었고, `coder/websocket` 은
   표준 `context` 친화적이며 형제 프로젝트에서 이미 쓰고 있다.

## 아키텍처

```
tiingo-go/
├── client.go                  # (수정) Stream *stream.Client 필드
├── stream/
│   ├── client.go              # 패키지 doc, Client, New(apiKey, opts), 5개 진입 메서드
│   ├── conn.go                # 연결·구독·하트비트·재연결 기계장치 (공유)
│   ├── envelope.go            # messageType I/H/A/E 분기, RawMessage 처리
│   ├── decode.go              # 위치 배열 → 값 추출 헬퍼(엄격 검증)
│   ├── crypto.go              # CryptoTrade, CryptoQuote, CryptoOptions
│   ├── forex.go               # ForexQuote, ForexOptions
│   ├── iex.go                 # IEXReferencePrice, IEXOptions
│   ├── equity.go              # EquityReferencePrice, EquityLiquidity, EquityOptions
│   ├── boats.go               # BoatsQuote, BoatsTrade, BoatsOptions (검증 불가)
│   ├── *_test.go
│   └── testdata/*.jsonl       # 실측 메시지 캡처(crypto 만 실측, 나머지는 문서 예시)
├── examples/stream/main.go
└── integration_test.go        # (수정) crypto 스트림 1건 — 24시간 거래라 항상 가능
```

## 공개 API

```go
// Stream 은 열린 WebSocket 구독이다. Messages 를 다 읽은 뒤 Err 로 종료 사유를 확인한다.
type Stream struct{ ... }

func (s *Stream) Messages() <-chan Message // 연결이 끝나면 닫힌다
func (s *Stream) Err() error               // 채널이 닫힌 뒤 호출한다. 정상 종료면 nil
func (s *Stream) SubscriptionID() int64    // 구독 확인 메시지의 id
func (s *Stream) Close() error             // 명시적 종료. ctx 취소로도 끝난다

// Message 는 스트림이 내보내는 데이터 메시지다. 타입 스위치로 구분한다.
type Message interface{ isMessage() }

func (c *Client) Crypto(ctx context.Context, opts *CryptoOptions) (*Stream, error)
func (c *Client) Forex(ctx context.Context, opts *ForexOptions) (*Stream, error)
func (c *Client) IEX(ctx context.Context, opts *IEXOptions) (*Stream, error)
func (c *Client) Equity(ctx context.Context, opts *EquityOptions) (*Stream, error)
func (c *Client) BOATS(ctx context.Context, opts *BOATSOptions) (*Stream, error)
```

`Messages()` 채널은 버퍼를 갖는다(기본 256). 소비자가 느리면 채널이 차고, 그때는 **가장 오래된
메시지를 버리고 카운터를 올린다** — 시세 스트림에서 블로킹은 연결 전체를 정체시킨다.
누락 건수는 `s.Dropped()` 로 확인한다.

## 데이터 모델 (발췌)

```go
// CryptoTrade 는 암호화폐 체결이다. thresholdLevel 2·5 에서 온다.
// 배열 매핑은 실측으로 확인했다: ["T", ticker, date, exchange, lastSize, lastPrice]
type CryptoTrade struct {
	Ticker    string     // 페어(btcusd 등)
	Date      types.Time // 체결 시각
	Exchange  string     // 체결 거래소(bitfinex, gdax 등)
	LastSize  float64    // 기준 통화 기준 체결 수량
	LastPrice float64    // 체결가
}

// CryptoQuote 는 암호화폐 호가다. thresholdLevel 2 에서만 온다.
// 배열 매핑은 실측으로 확인했다(mid=(bid+ask)/2 검산 일치).
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

// ForexQuote 는 통화쌍 호가다.
// 배열 매핑은 문서 예시의 산술로 검증했다(라이브 미검증) —
// 문서의 인덱스 표는 askPrice=6 이라 적지만 예시가 askPrice=7 임을 증명한다.
type ForexQuote struct { ... }
```

전 타입에 이 규약을 적용한다: **doc 주석 첫 줄에 배열 매핑을, 둘째 줄에 실측/문서 추정 여부를 적는다.**

## 엄격 디코딩

```go
// arr 은 위치 배열 하나를 읽는 커서다. 길이·타입이 어긋나면 에러를 모아 두었다가 한 번에 돌려준다.
// 위치 매핑이 틀리면 값만 조용히 바뀌므로, 여기서 시끄럽게 실패하는 것이 유일한 방어선이다.
type arr struct { raw []json.RawMessage; err error }

func (a *arr) str(i int) string
func (a *arr) f64(i int) float64
func (a *arr) i64(i int) int64
func (a *arr) time(i int) types.Time
func (a *arr) err() error
```

- 배열 길이가 기대와 다르면 에러(`tiingo: crypto trade expects 6 elements, got 5`).
- 원소 타입이 다르면 에러. `null` 은 포인터 필드에서만 허용한다.
- **길이가 기대보다 길면 에러가 아니다** — Tiingo 가 REST 에서 "필드를 계속 추가하겠다" 고 밝혔고
  WebSocket 도 같을 수 있다. 뒤에 붙는 건 무시한다. 짧은 것만 에러다.

## 테스트

- **단위**: `testdata/*.jsonl` 의 메시지를 넣어 각 타입으로 디코딩되는지. 봉투 4종 분기.
  **짧은 배열·타입 불일치가 에러가 되는지**(엄격 디코딩의 존재 이유). 긴 배열은 통과하는지.
  crypto 의 `mid = (bid+ask)/2` 검산. 재연결(스텁 서버를 끊고 재구독 프레임이 오는지).
  느린 소비자에서 오래된 메시지가 버려지고 `Dropped()` 가 오르는지.
- **fixture**: crypto 는 실측 캡처(체결·호가 각각 여러 건). 나머지 넷은 **문서 예시에서 만들고
  파일명·주석에 `_doc` 를 붙여 실측이 아님을 표시**한다.
- **integration**: crypto 스트림만. 24시간 거래라 언제 돌려도 메시지가 온다. 10초 안에 1건 이상
  받고 필드가 채워지는지 확인. 나머지 넷은 시장 시간에 묶여 있어 통합 테스트에 넣지 않는다.

## 릴리스

머지 후 **v0.9.0**. README 에 WebSocket 절을 새로 만들고, 커버리지 표와 별개로 다룬다
(REST 표에 섞으면 요청-응답과 스트리밍이 같은 것처럼 읽힌다).

## 범위 밖 (후속)

- IEX thresholdLevel 0/5(TOPS) — IEX Exchange 협약 필요.
- 스트림 중간 구독 변경(티커 추가·제거) — 우선 고정 구독으로 시작한다.
- moneyflow 통합, REST 재시도·rate limit 백오프.

## 주의

- **검증된 건 Crypto 뿐이다.** Forex 는 일요일 21시 UTC, IEX·Equity 는 미국 장중에 확인 가능하고
  BOATS 는 영영 불가능하다. 이 사실을 패키지 doc·README·각 타입 주석에 남겨,
  나중에 확인할 목록이 코드에 남게 한다.
- **Forex 문서의 인덱스 표가 틀렸다.** 표는 askPrice=6/askSize=7 이라 적지만 예시의 산술은
  askSize=6/askPrice=7 임을 증명한다. Crypto 와 같은 순서다. 표 숫자를 그대로 믿으면
  매도 가격과 수량이 뒤바뀐 채 컴파일·테스트가 통과한다.
- Equity 의 두 메시지는 종류 문자가 없어 **길이로 구분**한다(3 vs 8). 구독 threshold 와 함께 본다.
