# Tiingo API 문서 카탈로그 설계

- 작성일: 2026-09-04
- 상태: 확정 (브레인스토밍 완료)
- 레포: `github.com/kenshin579/tiingo-go` (워크스페이스 `tiingo-go/`, branch `chore/api-docs`)
- 토픽: Tiingo API 문서를 SDK 개발의 1차 참조로 레포에 보관 — 공식 llms 원본 보관 + 웹 문서 23페이지 md 변환

## 배경 / 목적

`tiingo-go`(Tiingo API Go 클라이언트, fmp-go 와 같은 구조)를 만들기에 앞서 API 문서를 레포에 갖춘다.
fmp-go 는 Playwright 크롤러로 274 페이지를 md 로 변환했고, toss-go 는 공식 기계판독 원본(openapi/asyncapi)을
그대로 보관했다. Tiingo 는 두 상황이 섞여 있다.

- Tiingo 가 `llms.txt` / `llms-full.txt` 를 **공식 제공**한다. 엔드포인트·파라미터·정책은 전부 들어 있다.
- 그러나 `llms-full.txt` 는 웹 문서의 **축약본**이다. 일부 엔드포인트의 응답 필드가 표가 아니라 이름 나열이고
  (타입·설명 없음), 엔드포인트별 응답 예시 JSON 도 일부만 있다. SDK 구조체 타입을 정할 때 이 차이가 실제로 영향을 준다.

따라서 **둘 다 보관**한다. llms 원본은 Tiingo 가 유지하는 개념·정책의 source of truth, 크롤링 md 는 필드 타입·예시가
완전한 엔드포인트 레퍼런스다.

작업 순서:
- **0 (본 스펙)** — 문서 카탈로그. 코드(SDK) 없음, `go.mod` 없음.
- **1** — Tiingo Go SDK 기반 + 첫 카테고리(End-of-Day). 별도 스펙.
- **2** — 카테고리별 점진 확장, moneyflow 통합. 별도 스펙.

## 사전 조사 결과 (확정 사실, 2026-09-04 기준)

### 웹 문서 사이트

- `https://www.tiingo.com/documentation/...` 는 **Angular SPA**. 서버 HTML 에 본문이 없고 문서 모듈은 lazy chunk
  (`src_app_api_documentation_documentation_module_ts.*.js`, 직접 다운로드는 403)로 렌더된다. **반드시 headless
  브라우저로 렌더링**해야 읽을 수 있다. `sitemap.xml` 은 없다(404).
- 사이드바(`mat-sidenav` 안 `a.side-bar-link-container[href]`)가 문서 페이지를 전부 열거한다. 그룹 링크는
  `indent-level` 없음(`1. General` 등), leaf 페이지는 `.indent-level-1`. leaf 23개(3+11+5+1+3):

| 그룹 | URL (`/documentation/` 이하) |
|---|---|
| 1. General (3) | `general/overview`, `general/connecting`, `general/changelog` |
| 2. REST (11) | `end-of-day`, `news`, `crypto`, `forex`, `equity-realtime-stock-data`, `iex`, `boats`, `fundamentals`, `mutual-fund-and-etf-fees`, `corporate-actions/dividends`, `corporate-actions/splits` |
| 3. Websockets (5) | `websockets/crypto`, `websockets/forex`, `websockets/equity-realtime-stock-data`, `websockets/iex`, `websockets/boats` |
| 4. Utilities (1) | `utilities/search` |
| 5. Appendix (3) | `appendix/developers`, `appendix/integrations`, `appendix/symbology` |

- 본문 컨테이너는 `tiingo-api-canvas`. 구조는 페이지마다 균일하다.
  - `h1/h2/h3` — 번호 붙은 섹션 제목 (예: `2.1.2 End-of-Day Endpoint`, `3.4.3 Top-of-Book & Last Trade`).
  - 탭 밖 `pre` — 엔드포인트 URL 블록 (`# Meta Data\nhttps://api.tiingo.com/tiingo/daily/<ticker>`).
  - `mat-tab-group` — 엔드포인트마다 하나, 탭 라벨(`.mat-tab-label`)은 `Response` / `Request` / `Examples`.
    탭 본문은 클릭 후 `.mat-tab-body-active` 에서 읽는다.
  - 필드 표는 `<table>` 이 아니라 `tiingo-doc-table[tabletype=response|request]` 안의 div —
    `.header-row .header-cell`(헤더), `.parameter-row .parameter-cell`(행). 셀 안에 `pre`(JSON 필드명)와
    줄바꿈(목록이 들어간 설명)이 있을 수 있다.
    - Response 헤더: Field Name / JSON Field / Data Type / Description
    - Request 헤더: Field Name / Parameter / JSON Field / Data Type / Required / Description
  - Examples 탭은 언어 서브탭(Python / Node / PHP) + 코드 `pre` + `Response:` 아래 예시 JSON `pre`.
    비로그인 상태에서는 코드 안에 `token=Not logged-in or registered. Please login or register to see your API Token`
    문자열이 들어간다. **로그인 상태면 실제 토큰이 렌더된다.**
- Overview / Changelog / Appendix 페이지는 탭 없이 헤딩 + 문단 + 코드블록만 있다.
- 자동 요청 차단 없음. curl 기본 UA 로도 200.

### 공식 llms 원본

| URL | 크기 | 내용 |
|---|---|---|
| `https://www.tiingo.com/llms.txt` | 9KB, 85줄 | 제품 라우팅 노트 + 문서·제품 페이지 링크 목록. `Last updated: 2026-08-18`. |
| `https://www.tiingo.com/llms-full.txt` | 88KB, 1,525줄 | "complete machine-readable corpus". 섹션마다 `title / description / last_updated` 프론트매터 구분선. REST 약 45 + WS 5 엔드포인트 URL, 파라미터 표, (일부) 응답 필드 표, (일부) 예시 JSON, 인증, rate limit, 플랜 제한, 심볼 규칙, 공통 에러, 미제공 기능 목록. |

웹 vs llms-full 대조(실제 렌더링으로 확인):
- End-of-Day: llms-full 이 웹의 파라미터·응답 표·예시 JSON 을 모두 포함하고, 증분 동기화 패턴·`permaTicker` 언급 등
  웹에 없는 설명이 더 있다.
- Fundamentals: 웹은 meta 응답 16개 필드, daily 응답 6개 필드에 타입(`datetime`, `int32`, `boolean` …)과 설명이
  있으나 llms-full 은 필드명만 나열. 웹의 `2.8.6 Additional Information & FAQ` 섹션은 llms-full 에 없다.
- 언어별 코드 샘플(Python/Node/PHP)은 llms-full 에 없다. SDK 에 가치가 없으므로 Python 1개만 보관한다.

## 결정 사항 (브레인스토밍)

1. **하이브리드**: llms 원본 2개를 그대로 보관 + 웹 문서 23페이지를 크롤러로 md 변환. (검토했던 대안: llms-full 만
   보관 → 필드 타입 누락으로 기각. 웹 크롤링만 → Tiingo 가 유지하는 정책·개념 원본을 버릴 이유가 없어 기각.)
2. **md 단위 = 웹 페이지 1개 = md 1개 (23 파일)**. (대안: 엔드포인트별 1파일(fmp-go 식, 약 50개) → Tiingo 는 한
   페이지에 엔드포인트 2~5개와 개요·FAQ 가 묶여 있어 쪼개면 문맥이 끊김. 파일 안에서 `##` 섹션으로 나누면 참조
   정확도는 충분.)
3. **크롤러 = Node + Playwright, `tools/gendocs/`** (fmp-go 와 동일 구조). (대안: Go + chromedp → 문서 생성 전용
   의존성이 라이브러리 go.mod 에 섞이거나 별도 모듈 필요. 1회성 DevTools 추출 → 갱신 절차 없음. 둘 다 기각.)
4. **크롤링 범위 = 23페이지 전부** (changelog / integrations / developers 포함). 제외 규칙이 코드를 늘리고, changelog 는
   필드 추가·폐기 이력이라 SDK 유지보수에 유용.
5. **페이지 열거는 사이드바를 읽는다** (하드코딩 목록 없음). Tiingo 가 페이지를 추가하면 자동 반영. 23개가 아니면
   경고만 하고 진행.
6. **llms-full.txt 를 섹션별 md 로 분할하지 않는다.** 웹 크롤링 md 가 그 역할을 하므로 원본 그대로만 둔다.
7. `go.mod` 는 이번에 만들지 않는다. SDK 스펙에서 생성.

## 파일 구성

```
tiingo-go/
├── README.md                     # 한 줄 소개 + 문서 카탈로그 안내 (SDK 내용은 이후 스펙에서)
├── .gitignore                    # tools/gendocs/node_modules, failures.log
├── docs/
│   ├── api/
│   │   ├── README.md             # 생성 인덱스: llms 원본 안내 + 그룹별 페이지 목록
│   │   ├── llms.txt              # Tiingo 공식 원본 (fetch-docs.sh)
│   │   ├── llms-full.txt         # Tiingo 공식 원본 (fetch-docs.sh)
│   │   ├── general/{overview,connecting,changelog}.md
│   │   ├── rest/{end-of-day,news,crypto,forex,equity-realtime-stock-data,iex,boats,
│   │   │         fundamentals,mutual-fund-and-etf-fees,dividends,splits}.md
│   │   ├── websockets/{crypto,forex,equity-realtime-stock-data,iex,boats}.md
│   │   ├── utilities/search.md
│   │   └── appendix/{developers,integrations,symbology}.md
│   └── superpowers/{specs,plans}/
├── scripts/
│   └── fetch-docs.sh             # llms.txt + llms-full.txt 재다운로드
└── tools/gendocs/
    ├── package.json              # @playwright/test, scripts: gen / test
    ├── gendocs.mjs               # 열거 + 렌더링 + 파일 쓰기 (I/O)
    ├── lib.mjs                   # 순수 함수: 추출 JSON → md (테스트 대상)
    ├── lib.test.mjs              # node --test
    └── failures.log              # 런타임 산출물, 커밋하지 않음
```

- 디렉터리 = 사이드바 그룹명 소문자(`general`, `rest`, `websockets`, `utilities`, `appendix`).
- 파일명 = URL **마지막 세그먼트**. `corporate-actions/dividends` → `rest/dividends.md`. 23개 안에서 충돌 없음(확인).

## 크롤러 설계 (`tools/gendocs/`)

### 흐름 (`gendocs.mjs`)

1. **열거**: `https://www.tiingo.com/documentation/general/overview` 를 열고 사이드바에서
   `a.side-bar-link-container[href]` 를 문서 순서로 읽는다. `.indent-level-1` 이 없는 링크는 그룹(제목 `1. General`
   → 디렉터리 `general`), 있는 링크는 직전 그룹에 속하는 leaf 페이지. 기대 23개, 다르면 콘솔 경고.
2. **페이지 추출** (`extractPage(page)` — 브라우저 컨텍스트에서 실행, 셀렉터는 이 함수에만 존재):
   `page.goto(url, {waitUntil:'networkidle'})` + 1초 대기 후 `tiingo-api-canvas` 안을 문서 순서로 걸어
   블록 배열을 만든다.
   - `h1/h2/h3` → `{type:'heading', level, text}`
   - `p` → `{type:'para', md}` (링크는 `[text](절대URL)`, `code` 는 백틱)
   - `ul/ol` → `{type:'list', ordered, items[]}`
   - 탭 밖 `pre` → `{type:'code', text}`
   - `mat-tab-group` → `{type:'tabs', tabs:[{name, blocks}]}`. 라벨마다 `.mat-tab-label` 클릭 → 300ms 대기 →
     `.mat-tab-body-active` 를 같은 규칙으로 걷는다. 추가로:
     - `tiingo-doc-table` → `{type:'table', header[], rows[][]}` (셀 텍스트는 `innerText`, 줄바꿈 보존)
     - Examples 탭: 언어 서브탭 중 **Python 만** 클릭해 코드 `pre` → `{type:'code', lang:'python'}`,
       `Response:` 다음 `pre` → `{type:'code', lang:'json'}`.
3. **렌더** (`lib.mjs`, 순수 함수): 블록 배열 + 메타(제목, 출처 URL, 생성일) → md 문자열.
4. **쓰기**: `docs/api/<group>/<slug>.md`, 디렉터리 자동 생성, 덮어쓰기.
5. **인덱스**: 전 페이지 처리 후 `docs/api/README.md` 생성.
6. **안정성**: 페이지 간 500ms 대기, 실패 시 1회 재시도, 최종 실패는 `failures.log` 에 URL 기록 후 계속. 성공/실패
   카운트 콘솔 요약. 전 과정 멱등.
7. **토큰 방어**: 비로그인 컨텍스트(새 브라우저 컨텍스트, 쿠키 없음)로 실행하고, 렌더 단계에서 `token=` 뒤 문자열을
   무조건 `<TOKEN>` 으로 치환한다(이중 방어).

실행: `cd tools/gendocs && npm install && npx playwright install chromium && npm run gen`.

### `lib.mjs` 렌더 규칙

- 제목: 페이지 `h1` 텍스트(없으면 사이드바 제목). 섹션 헤딩은 원문 번호 그대로(웹과 대조 편의). `h1` 은 페이지 제목으로
  쓰고 본문 헤딩 레벨은 `h2`→`##`, `h3`→`###`.
- 탭 순서는 웹(Response→Request→Examples)이 아니라 **Request → Response → Example** (SDK 구현 시 읽는 순서).
  탭 헤딩은 직전 섹션 헤딩보다 한 단계 아래(`###` 또는 `####`).
- 표: md 표. 셀 안 줄바꿈은 `<br>`, `|` 는 `\|`. 헤더는 웹 헤더 텍스트 그대로(`Field Name | JSON Field | ...`).
- 표가 없는 탭은 문단만 쓴다. 블록이 하나도 없는 탭은 헤딩까지 생략.
- 코드블록: 엔드포인트 URL 블록은 언어 없음, Python 예시는 `python`, 응답 예시는 `json`.
- 상단 인용: `> 출처: <URL> · 생성: YYYY-MM-DD (tools/gendocs)`.

### md 템플릿 (예: `rest/end-of-day.md`)

```markdown
# 2.1 REST - End-of-Day Prices

> 출처: https://www.tiingo.com/documentation/end-of-day · 생성: 2026-09-04 (tools/gendocs)

## REST Endpoints

​```
# Meta Data
https://api.tiingo.com/tiingo/daily/<ticker>
# Latest Price
https://api.tiingo.com/tiingo/daily/<ticker>/prices
​```

## 2.1.1 Overview

Tiingo's End-of-Day prices use a proprietary error checking framework ...

## 2.1.2 End-of-Day Endpoint

To request price data for a stock, use the following REST endpoints.

​```
# Latest Price Information
https://api.tiingo.com/tiingo/daily/<ticker>/prices
​```

### Request

| Field Name | Parameter | JSON Field | Data Type | Required | Description |
|---|---|---|---|---|---|
| Ticker | URL | N/A | string | Y | Ticker related to the asset. |
| Start Date | GET | startDate | date | N | If startDate or endDate is not null, ... |

### Response

| Field Name | JSON Field | Data Type | Description |
|---|---|---|---|
| Date | date | date | The date this data pertains to. |
| Open | open | float | The opening price for the asset on the given date. |

### Example

​```python
import requests
headers = { 'Content-Type': 'application/json' }
requestResponse = requests.get("https://api.tiingo.com/tiingo/daily/aapl/prices?startDate=2019-01-02&token=<TOKEN>", headers=headers)
print(requestResponse.json())
​```

​```json
[
  { "date": "2019-01-02T00:00:00.000Z", "close": 157.92, "high": 158.85, ... }
]
​```
```

### `docs/api/README.md` (생성)

1. 제목 + 한 줄 설명(웹 문서 23페이지 자동 변환 + Tiingo 공식 llms 원본 보관).
2. 원본 표: 파일 / 출처 URL / `Last updated`(llms.txt 의 값) / 가져온 날짜 — `fetch-docs.sh` 가 갱신.
3. 재생성 방법 두 줄(`npm run gen`, `scripts/fetch-docs.sh`).
4. 그룹별 페이지 목록: `## 1. General` 아래 `- [1.1 Overview](general/overview.md)` 형식.

## `scripts/fetch-docs.sh`

- bash, `set -euo pipefail`. 의존: `curl` 만.
- 레포 루트를 스크립트 위치에서 계산(`$(dirname "$0")/..`).
- `llms.txt`, `llms-full.txt` 를 브라우저 UA 헤더로 `curl -fsSL` 해 임시 파일에 받고, 둘 다 성공했을 때만
  `docs/api/` 로 이동(부분 갱신 방지). 실패 시 즉시 중단.
- `llms.txt` 의 `Last updated: YYYY-MM-DD` 를 읽어 콘솔에 출력하고 `docs/api/README.md` 원본 표의 두 행을 `sed` 로
  갱신(`Last updated`, 가져온 날짜).
- 멱등: 원본이 안 바뀌면 재실행해도 diff 없음.

## 에러 / 부분 처리

- 페이지 구조가 예상과 다르면(캔버스 없음, 탭 라벨 없음) 있는 블록만으로 md 를 만들고 콘솔에 경고. 완전 실패는
  `failures.log` 에 기록 후 계속 — 전체 작업을 중단시키지 않는다.
- 사이드바 leaf 가 23개가 아니면 경고만 하고 발견된 만큼 처리.
- 재실행으로 보강 가능. 성공/실패 카운트를 콘솔에 요약.

## 검증

- `find docs/api -name '*.md' -not -name README.md | wc -l` = 23, `failures.log` 비어 있음.
- `cd tools/gendocs && npm test` 통과. `lib.test.mjs` 는 fixture 블록 배열로 다음을 검증: 표 렌더(셀 줄바꿈 `<br>`,
  `|` 이스케이프), 탭 재정렬(Request→Response→Example), `token=` 치환, 빈 탭 생략, 헤딩 레벨 매핑.
- 육안 검수 3개: `rest/end-of-day.md`(표·예시 JSON), `websockets/iex.md`(h3 중첩 + 표 없는 탭),
  `general/overview.md`(탭 없는 산문 페이지).
- `docs/api/README.md` 의 상대 링크 23개가 모두 실제 파일을 가리키는지 스크립트로 확인.
- 크롤러 재실행 후 `git diff --stat` 이 생성 날짜 줄 외에 비어 있는지 확인(멱등).
- 생성된 md 어디에도 `Not logged-in` 문자열과 실제 토큰 패턴이 없는지 `grep`.

## 범위 밖 / 후속

- SDK 코드, `go.mod`, 실호출 fixture, moneyflow 통합 — 다음 스펙.
- 크롤러 CI 자동화 / 스케줄 — 수동 재실행으로 충분.
- llms-full.txt 섹션 분할 — 하지 않음(결정 6).
- Node / PHP 코드 샘플 보관 — 하지 않음.

## 위험 / 주의

- Angular Material 셀렉터(`mat-tab-group`, `.mat-tab-label`, `tiingo-doc-table`, `.parameter-cell`)는 Tiingo 배포로
  바뀔 수 있다. 셀렉터는 `extractPage()` 한 함수에만 두어 수정 지점을 한 곳으로 한다.
- 로그인 상태로 크롤링하면 예시 코드에 실제 API 토큰이 렌더된다. 비로그인 컨텍스트 + `<TOKEN>` 치환으로 이중 방어하고,
  검증 단계에서 `grep` 으로 확인한다.
- Playwright chromium 설치 필요(`npx playwright install chromium`). 이 맥에는 2026-09-04 기준 설치돼 있지 않다.
- 23페이지 렌더링은 1~2분. 500ms 대기로 사이트에 부하를 주지 않는다.
