# Tiingo Go SDK — Corporate Actions 카테고리 설계

- 작성일: 2026-09-05
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `feature/corporate-actions`)
- 토픽: Tiingo Corporate Actions — 여덟 번째이자 **의도적으로 미완성인** 카테고리
- 선행: v0.7.0 (EOD + Fundamentals + Crypto + Forex + IEX + Search + Equity Realtime)
- API 참조: `docs/api/rest/dividends.md`, `docs/api/rest/splits.md`

## 배경 / 목적

**남은 REST 카테고리가 권한 장벽으로 사실상 소진됐다.** 2026-09-05 실측:

| 카테고리 | 상태 |
|---|---|
| News | 403 `You do not have permission to access the News API` |
| Fund Fees | 403 `You do not have permission to access the Mutual Fund API` |
| BOATS | 403 `You do not have the BOATS entitlement` (유료 add-on) |
| Corporate Actions | **5개 중 4개 403**(본문 없는 빈 403), 1개만 200 |

무료 키로 되는 마지막 REST 엔드포인트가 `distribution-yield` 하나다. 이 카테고리는 그 하나만
구현하고, 나머지 넷은 권한이 열리면 채운다. **반쪽인 상태가 설계 의도이며 문서에 명시한다.**

배당수익률 시계열 자체는 쓸모가 분명하다 — 종목 분석에서 현재 수익률과 그 추이를 보여줄 수 있고,
`moneyflow` 통합 시 바로 쓰인다.

## 사전 조사 결과 (실호출 확인, 2026-09-05)

### 되는 것

| 호출 | 결과 |
|---|---|
| `GET /tiingo/corporate-actions/aapl/distribution-yield` | 200, **11,523건**(1980-12-12 ~ 2026-09-04) |
| `?startDate=2026-08-01` | 200, 25건 |
| `?startDate=2026-08-01&endDate=2026-08-15` | 200, 10건 |
| `?columns=trailingDiv1Y` | 200, **11,523건 — `date` 가 그대로 온다** |
| `/nosuchtickerxyz/distribution-yield` | **200 + `[]`** (404 아님) |

### 안 되는 것 (전부 403, 본문 `content-length: 0`)

- `GET /tiingo/corporate-actions/<ticker>/distributions`
- `GET /tiingo/corporate-actions/distributions?exDate=...`
- `GET /tiingo/corporate-actions/splits?exDate=...`
- (splits 의 티커별 판도 같은 계열로 본다)

News·Fund Fees·BOATS 의 403 은 사유가 담긴 본문을 주는데, 이쪽 403 은 본문이 비어 있다.
사유를 알 수 없으므로 "권한 없음" 이상은 단정하지 않는다.

### 문서와 다른 네 가지

1. **`startDate`/`endDate` 가 문서 표에 없는데 동작한다.** 표에는 `format`·`columns` 만 있다.
   필터가 실제로 걸리는 것을 건수로 확인했다(11,523 → 25 → 10). Search 의 `isin` 과 같은 패턴이다.
2. **`columns` 가 필드를 떨어뜨리지 못한다.** 문서는 "`columns=trailingDiv1Y` 면 그 필드 하나만
   반환" 이라 적지만 실제로는 `date` 가 함께 온다. EOD·IEX 의 `columns` 는 실제로 떨어뜨리므로
   같은 이름이 카테고리마다 다르게 동작하는 셈이다.
3. **`trailingDiv1Y` 는 JSON 숫자다.** 문서 표는 `string` 이라 적었다. 실제 `0.0033128106`.
4. **없는 티커가 200 + `[]`** 다. `equity`·`iex` 의 시세는 같은 상황에서 404 를 준다.

### 문서가 예고한 변화

> "we will continue to add new daily metrics, so the fields will change throughout time.
> We recommend you **do not** make parsing code that requires columns or fields to be in a
> particular order."

응답 필드가 늘어날 것이라고 명시한다. Go 의 `encoding/json` 은 순서와 무관하고 모르는 필드를
무시하므로 구조체는 그대로 견딘다. 새 필드를 쓰려면 그때 필드를 추가하면 된다.

## 결정 사항 (브레인스토밍)

1. **패키지는 `corporateactions`, 루트 필드는 `CorporateActions`, 메서드는 `DistributionYield`.**
   (기각: `dividends` 패키지 + `Dividends.Yield` — 문서 페이지 제목과 맞고 짧지만, 막혀 있는 4개
   중에 **splits** 가 있다. 권한이 열렸을 때 분할을 배당 패키지에 넣는 건 명백히 틀린 모양이다.
   이름은 되돌리는 비용이 크고, 이 카테고리는 애초에 반쪽이라 나머지가 들어올 자리를 이름이
   막지 않아야 한다. URL 구조(`/tiingo/corporate-actions/`)와도 일치한다.)
2. **`StartDate`/`EndDate` 를 노출한다.** 문서 표에 없지만 실측으로 확인했다. 전체가 11,523건이라
   기간 제한 수단이 없으면 쓰기 어렵다. Search 의 `SearchByISIN` 과 같은 선례다.
3. **`Columns` 는 노출하지 않는다.** 실제로 필드를 줄이지 못하고(위 2번), 문서상 용도인 "필드 순서
   고정" 은 CSV 용인데 이 라이브러리는 CSV 를 지원하지 않는다. Go 디코딩은 순서와 무관하다.
   노출하면 "이걸 주면 필드가 줄겠지" 라는 잘못된 기대만 만든다.
4. **`Date` 는 `types.Date`.** 응답이 항상 `T00:00:00.000Z` 이고 의미상 날짜다.
   `eod.Price`·`fundamentals.Daily` 와 같다.
5. **막힌 4개는 코드로 만들지 않는다.** 스텁이나 "권한 없음" 에러를 던지는 메서드를 두지 않는다 —
   응답 형태를 실측할 수 없어 모델을 지어내야 하고, 이 레포는 지금까지 여덟 카테고리 전부
   실호출로 문서 오류를 찾아냈다. 검증 없는 모델은 규약 위반이다. 대신 패키지 doc 과 README 에
   무엇이 왜 빠졌는지 적는다.
6. 필드 주석 규약 유지, fixture 는 실호출.

## 아키텍처

```
tiingo-go/
├── client.go                       # (수정) CorporateActions *corporateactions.Client 추가
├── corporateactions/
│   ├── client.go                   # 패키지 doc(막힌 4개 설명 포함), Client, New(hc)
│   ├── yield.go                    # Yield, YieldOptions, DistributionYield
│   ├── *_test.go
│   └── testdata/*.json             # 실호출 fixture 2종
├── examples/corporateactions/main.go
└── integration_test.go             # (수정) 3건 추가
```

엔드포인트가 하나뿐이라 파일도 하나다(`search` 와 같은 구성).

## 데이터 모델

```go
// Yield 는 하루치 배당수익률이다.
//
// Tiingo 가 이 엔드포인트에 지표를 계속 추가하겠다고 문서에 밝혔으므로 필드가 늘 수 있다.
// 모르는 필드는 무시되니 디코딩은 깨지지 않고, 새 지표가 필요해지면 그때 필드를 더하면 된다.
type Yield struct {
	Date types.Date `json:"date"` // 기준일. 응답은 늘 자정이라 날짜만 의미가 있다
	// TrailingDiv1Y 는 최근 1년 분배금 기준 수익률이다. 0.0033 이면 0.33% 다.
	// 문서 표는 string 이라 적지만 실제로는 JSON 숫자로 온다.
	TrailingDiv1Y float64 `json:"trailingDiv1Y"`
}

// YieldOptions 는 DistributionYield 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
//
// 두 필드 모두 Tiingo 문서의 파라미터 표에는 없지만 실제로 동작한다(실호출 확인).
// 지정하지 않으면 상장 이후 전 기간이 오므로 응답이 매우 클 수 있다 — AAPL 이 11,523건이다.
type YieldOptions struct {
	StartDate time.Time // 조회 시작(>=). zero 면 미전송
	EndDate   time.Time // 조회 종료(<=). zero 면 미전송
}
```

## 메서드

```go
// DistributionYield 는 배당수익률 시계열을 받는다.
// GET /tiingo/corporate-actions/<ticker>/distribution-yield
//
// 옵션 없이 부르면 상장 이후 전 기간이 온다(AAPL 실측 11,523건, 1980년부터).
// 기간을 좁히려면 YieldOptions 를 쓴다.
// 없는 티커는 404 가 아니라 빈 슬라이스다 — 이 점이 equity·iex 의 시세와 다르다.
func (c *Client) DistributionYield(ctx context.Context, ticker string, opts *YieldOptions) ([]Yield, error)
```

- 티커는 `strings.TrimSpace` 후 빈 값이면 에러(왕복을 아낀다). 경로에 `url.PathEscape` 로 넣는다.
- 옵션은 `params() map[string]string` 으로 만들고 넣을 값이 없으면 nil 이다(`iex`·`equity` 규약).

## 패키지 doc 에 담을 것

막혀 있는 4개를 여기 적는다. 사용자가 "왜 배당 내역은 없나" 를 소스에서 바로 알 수 있어야 한다.

```go
// Package corporateactions 는 Tiingo Corporate Actions API sub-client 다.
// tiingo.Client.CorporateActions 로 접근한다.
//
// 지금은 배당수익률(DistributionYield) 하나만 구현돼 있다. 같은 API 그룹의 배당 내역
// (distributions, 티커별·배치)과 분할(splits)은 이 라이브러리를 만든 계정 권한으로 403 이라
// 응답 형태를 확인할 수 없어 넣지 않았다 — 실호출로 검증하지 않은 모델은 두지 않는다는 원칙이다.
// 권한이 열리면 이 패키지에 추가한다.
package corporateactions
```

## 테스트

- **fixture 2종**(실호출): `yield_aapl_recent.json`(`startDate` 로 좁힌 25건 남짓),
  `yield_aapl_early.json`(**초기 구간 — `trailingDiv1Y` 가 `0.0` 인 행 포함**, 값 0 이 정상임을 고정).
- **단위 테스트**: 파싱(`Date` 가 날짜로, `TrailingDiv1Y` 가 숫자로), `0.0` 이 정상 값임,
  쿼리 생성(zero 값 생략, 둘 다 지정, nil 옵션이면 쿼리 없음), 빈 티커가 **왕복 전에** 막히는지
  (`lastReq()` nil), 경로 이스케이프, 없는 티커의 빈 배열이 에러가 아님, 403 → `APIError`.
- **integration**: `AAPL` 기간 조회, 없는 티커가 빈 슬라이스, 옵션 없이 부르면 대량으로 오는지
  (건수만 확인 — 11,523건이라 한 번만 돌린다).

## 릴리스

머지 후 **v0.8.0**. README 커버리지 표에 1행 추가하고, 미구현 목록에서 Dividends/Splits 의
상태를 실측으로 갱신한다.

## 범위 밖 (후속)

- 같은 그룹의 403 4개 — 권한이 열리면.
- News(403), Fund Fees(403), BOATS(유료 add-on).
- WebSocket 5종 — 무료 키로 되는지 미확인.
- moneyflow 통합, CSV, 재시도, rate limit 백오프.

## 주의

- Tiingo 는 **시간당 요청 한도**가 있다. 옵션 없는 전체 조회는 11,523건이라 통합 테스트에서
  한 번만 돌린다. 429 가 나면 정시에 리셋된다.
- 이 카테고리는 **의도적으로 반쪽**이다. 커버리지 표와 패키지 doc 이 그 사실을 말하지 않으면
  버그로 오인된다.
