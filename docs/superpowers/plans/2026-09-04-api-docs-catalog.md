# Tiingo API 문서 카탈로그 구현 계획

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Tiingo 문서 사이트 23페이지를 Playwright 크롤러로 md 변환해 `docs/api/`에 두고, Tiingo 공식 `llms.txt`/`llms-full.txt` 원본을 함께 보관한다.

**Architecture:** `tools/gendocs/`(Node + Playwright)가 사이드바로 페이지를 열거하고, 각 페이지의 DOM을 브라우저 안에서 구조화된 블록 JSON으로 추출한 뒤(`gendocs.mjs`, I/O), 순수 함수(`lib.mjs`)가 블록을 md로 렌더한다. `scripts/fetch-docs.sh`(bash + curl)가 llms 원본 2개를 내려받고 `docs/api/README.md`의 원본 표를 갱신한다. 스펙: `docs/superpowers/specs/2026-09-04-api-docs-catalog-design.md`.

**Tech Stack:** Node 25 (`node --test`), `@playwright/test` ^1.60 (chromium), bash, curl. Go 코드 없음(`go.mod` 없음).

---

## 파일 구조

| 파일 | 책임 |
|---|---|
| `.gitignore` | `tools/gendocs/node_modules/`, `tools/gendocs/failures.log` |
| `README.md` | 레포 한 줄 소개 + 문서 카탈로그 안내 |
| `tools/gendocs/package.json` | Playwright 의존, `gen`/`test` 스크립트 |
| `tools/gendocs/lib.mjs` | 순수 함수: 사이드바 링크 → nav 구조, 블록 → md, 인덱스 md, 토큰 치환 |
| `tools/gendocs/lib.test.mjs` | `lib.mjs` 단위 테스트 (`node --test`) |
| `tools/gendocs/gendocs.mjs` | Playwright 실행, 열거, `extractPage`(브라우저 셀렉터는 여기에만), 파일 쓰기, 재시도, 인덱스 |
| `scripts/fetch-docs.sh` | llms 원본 2개 다운로드 + README 원본 표 갱신 |
| `docs/api/README.md` | 생성 인덱스 (gendocs가 쓰고, fetch-docs.sh가 원본 표 두 행만 갱신) |
| `docs/api/{general,rest,websockets,utilities,appendix}/*.md` | 생성된 23개 페이지 |
| `docs/api/llms.txt`, `docs/api/llms-full.txt` | Tiingo 공식 원본 |

### 블록 JSON 계약 (`extractPage` 출력 = `lib.mjs` 입력)

```js
// 페이지
{ title: '2.1 End-of-Day (EOD) Stock Price API Documentation', blocks: Block[] }
// Block
{ type: 'heading', level: 1|2|3, text: '2.1.2 End-of-Day Endpoint' }
{ type: 'para', md: 'Both raw prices and [CRSP](https://...) are available.' }
{ type: 'list', ordered: false, items: ['5,500+ Equities Covered.', ...] }
{ type: 'code', lang: '' | 'python' | 'json', text: '# Meta Data\nhttps://api.tiingo.com/tiingo/daily/<ticker>' }
{ type: 'table', header: ['Field Name','JSON Field','Data Type','Description'], rows: [['Date','date','date','The date...'], ...] }
{ type: 'tabs', tabs: [{ name: 'Response'|'Request'|'Examples', blocks: Block[] }] }
```

### nav 구조 (`buildNav` 출력)

```js
[
  { group: '1. General', dir: 'general',
    pages: [{ title: '1.1 Overview', slug: 'overview', href: '/documentation/general/overview' }, ...] },
  { group: '2. REST', dir: 'rest', pages: [...] },
  ...
]
```

---

### Task 1: 스캐폴딩 (.gitignore, package.json, Playwright 설치)

**Files:**
- Create: `.gitignore`
- Create: `tools/gendocs/package.json`

- [ ] **Step 1: 브랜치 확인**

Run: `cd /Users/user/src/workspace_moneyflow/tiingo-go && git branch --show-current`
Expected: `chore/api-docs`

- [ ] **Step 2: `.gitignore` 작성**

```gitignore
# tools/gendocs (Node 크롤러) 산출물
tools/gendocs/node_modules/
tools/gendocs/failures.log
```

- [ ] **Step 3: `tools/gendocs/package.json` 작성**

```json
{
  "name": "tiingo-gendocs",
  "private": true,
  "type": "module",
  "version": "0.0.0",
  "description": "Tiingo API docs → markdown catalog generator",
  "scripts": {
    "gen": "node gendocs.mjs",
    "test": "node --test"
  },
  "devDependencies": {
    "@playwright/test": "^1.60.0"
  }
}
```

- [ ] **Step 4: 의존성 + chromium 설치**

Run: `cd tools/gendocs && npm install && npx playwright install chromium`
Expected: `package-lock.json` 생성, `added N packages`, chromium 다운로드 완료 메시지. (이 맥에는 2026-09-04 기준 Playwright 브라우저가 없다.)

- [ ] **Step 5: 커밋**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add .gitignore tools/gendocs/package.json tools/gendocs/package-lock.json
git commit -m "chore: gendocs 크롤러 스캐폴딩 (Node + Playwright)"
```

---

### Task 2: `lib.mjs` — nav·슬러그·토큰·표 헬퍼

**Files:**
- Create: `tools/gendocs/lib.mjs`
- Create: `tools/gendocs/lib.test.mjs`

- [ ] **Step 1: 실패하는 테스트 작성**

`tools/gendocs/lib.test.mjs`:

```js
import { test } from 'node:test';
import assert from 'node:assert';
import { slugFromHref, groupDir, buildNav, redactToken, escapeCell, renderTable } from './lib.mjs';

test('slugFromHref returns last path segment', () => {
  assert.strictEqual(slugFromHref('/documentation/corporate-actions/dividends'), 'dividends');
  assert.strictEqual(slugFromHref('/documentation/end-of-day'), 'end-of-day');
  assert.strictEqual(slugFromHref('/documentation/websockets/iex/'), 'iex');
});

test('groupDir lowercases the group title without its number', () => {
  assert.strictEqual(groupDir('1. General'), 'general');
  assert.strictEqual(groupDir('2. REST'), 'rest');
  assert.strictEqual(groupDir('3. Websockets'), 'websockets');
  assert.strictEqual(groupDir('5. Appendix'), 'appendix');
});

test('buildNav groups leaf links under the preceding group link', () => {
  const links = [
    { href: '/documentation/general', title: '1. General', leaf: false },
    { href: '/documentation/general/overview', title: '1.1 Overview', leaf: true },
    { href: '/documentation/general/connecting', title: '1.2 Connecting', leaf: true },
    { href: '/documentation/end-of-day', title: '2. REST', leaf: false },
    { href: '/documentation/end-of-day', title: '2.1 End-of-Day', leaf: true },
    { href: '/documentation/corporate-actions/dividends', title: '2.10 Dividends', leaf: true },
    { href: '/documentation/corporate-actions/dividends', title: '2.10 Dividends', leaf: true }, // 중복
  ];
  const nav = buildNav(links);
  assert.strictEqual(nav.length, 2);
  assert.deepStrictEqual(nav[0], {
    group: '1. General', dir: 'general',
    pages: [
      { title: '1.1 Overview', slug: 'overview', href: '/documentation/general/overview' },
      { title: '1.2 Connecting', slug: 'connecting', href: '/documentation/general/connecting' },
    ],
  });
  assert.strictEqual(nav[1].dir, 'rest');
  assert.deepStrictEqual(nav[1].pages.map((p) => p.slug), ['end-of-day', 'dividends']);
});

test('buildNav ignores leaf links before any group', () => {
  assert.deepStrictEqual(buildNav([{ href: '/documentation/x', title: 'x', leaf: true }]), []);
});

test('redactToken replaces the logged-out sentence and real-looking tokens', () => {
  assert.strictEqual(
    redactToken('https://api.tiingo.com/tiingo/daily/aapl?token=Not logged-in or registered. Please login or register to see your API Token", headers=headers)'),
    'https://api.tiingo.com/tiingo/daily/aapl?token=<TOKEN>", headers=headers)',
  );
  assert.strictEqual(redactToken('?token=0123456789abcdef0123456789abcdef01234567&x=1'), '?token=<TOKEN>&x=1');
  assert.strictEqual(redactToken('no token here'), 'no token here');
});

test('escapeCell escapes pipes and turns newlines into <br>', () => {
  assert.strictEqual(escapeCell('a | b\nc\n\nd'), 'a \\| b<br>c<br>d');
  assert.strictEqual(escapeCell('  x  '), 'x');
});

test('renderTable renders header, separator, rows; empty when no rows', () => {
  const md = renderTable({ header: ['Field Name', 'JSON Field'], rows: [['Date', 'date'], ['Open|', 'open\nfloat']] });
  assert.strictEqual(md, [
    '| Field Name | JSON Field |',
    '| --- | --- |',
    '| Date | date |',
    '| Open\\| | open<br>float |',
  ].join('\n'));
  assert.strictEqual(renderTable({ header: [], rows: [] }), '');
  assert.strictEqual(renderTable({ header: ['a'], rows: [] }), '');
});
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `cd tools/gendocs && npm test`
Expected: FAIL — `Cannot find module './lib.mjs'`

- [ ] **Step 3: `tools/gendocs/lib.mjs` 구현**

```js
// 순수 변환 로직 — Playwright/네트워크/파일시스템 의존 없음 (단위테스트 대상).

// slugFromHref: '/documentation/corporate-actions/dividends' → 'dividends'
export function slugFromHref(href) {
  return String(href).replace(/\/+$/, '').split('/').pop();
}

// groupDir: 사이드바 그룹 제목 '2. REST' → 디렉터리 'rest'
export function groupDir(title) {
  return String(title).replace(/^\d+\.\s*/, '').trim().toLowerCase().replace(/\s+/g, '-');
}

// buildNav: 사이드바 링크 배열(문서 순서, {href,title,leaf}) → [{group, dir, pages:[{title,slug,href}]}]
// leaf 가 아닌 링크가 그룹을 열고, 이어지는 leaf 링크가 그 그룹에 속한다. href 중복은 첫 것만.
export function buildNav(links) {
  const nav = [];
  const seen = new Set();
  for (const l of links) {
    if (!l.leaf) {
      nav.push({ group: l.title, dir: groupDir(l.title), pages: [] });
      continue;
    }
    const cur = nav[nav.length - 1];
    if (!cur || seen.has(l.href)) continue;
    seen.add(l.href);
    cur.pages.push({ title: l.title, slug: slugFromHref(l.href), href: l.href });
  }
  return nav;
}

// redactToken: 예시 코드의 token= 값을 <TOKEN> 으로. 비로그인 안내 문장(공백 포함)과
// 실제 토큰처럼 보이는 영숫자 20자 이상을 모두 치환한다.
export function redactToken(text) {
  return String(text)
    .replace(/token=Not logged-in[^"'\n&]*/g, 'token=<TOKEN>')
    .replace(/token=[A-Za-z0-9]{20,}/g, 'token=<TOKEN>');
}

// escapeCell: md 표 셀용 — 앞뒤 공백 제거, '|' 이스케이프, 줄바꿈(연속 포함) → <br>
export function escapeCell(text) {
  return String(text).trim().replace(/\|/g, '\\|').replace(/\s*\n\s*(\n\s*)*/g, '<br>');
}

// renderTable: {header, rows} → md 표. 헤더나 행이 없으면 빈 문자열.
export function renderTable({ header, rows }) {
  if (!Array.isArray(header) || header.length === 0 || !Array.isArray(rows) || rows.length === 0) return '';
  const line = (cells) => `| ${cells.map(escapeCell).join(' | ')} |`;
  return [line(header), `| ${header.map(() => '---').join(' | ')} |`, ...rows.map(line)].join('\n');
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `cd tools/gendocs && npm test`
Expected: `# pass 7`, `# fail 0`

- [ ] **Step 5: 커밋**

```bash
git add tools/gendocs/lib.mjs tools/gendocs/lib.test.mjs
git commit -m "feat(gendocs): nav/슬러그/토큰 치환/표 렌더 헬퍼"
```

---

### Task 3: `lib.mjs` — 블록 → md 렌더 (`renderBlocks`, `renderPage`) + Task 2 리뷰 반영

**Files:**
- Modify: `tools/gendocs/lib.mjs` (Task 2 함수 3개 수정 + 끝에 추가)
- Modify: `tools/gendocs/lib.test.mjs` (끝에 추가)

**Task 2 코드 리뷰 반영 사항** (리뷰어 지적: `redactToken` 이 `?token=` 쿼리 형태만 잡음 — 스펙의 "무조건 치환" 보다 좁다):
- `redactToken` 을 대소문자 무시로 바꾸고, 헤더 형태(`Authorization: Token <hex>`)와 WebSocket JSON 형태(`"authorization": "<token>"`, `"token": "<token>"`)도 치환. 캡처 그룹으로 원문의 `token=`/`Token ` 표기는 보존.
- `renderPage` 최종 문자열에 `redactToken` 을 한 번 더 적용(문단 안의 토큰도 잡히도록).
- `escapeCell` 의 죽은 하위 패턴 `(\n\s*)*` 제거(`/\s*\n\s*/g` 와 동일 동작).
- `slugFromHref` 가 `?query`/`#hash` 를 떼어내도록(파일 경로 안전).

- [ ] **Step 0: Task 2 함수 수정 — 실패하는 테스트 먼저**

`tools/gendocs/lib.test.mjs` 의 `slugFromHref` 테스트와 `redactToken` 테스트에 아래 assertion 을 추가한다.

```js
// slugFromHref 테스트 안에 추가
  assert.strictEqual(slugFromHref('/documentation/end-of-day?x=1#top'), 'end-of-day');

// redactToken 테스트 안에 추가
  assert.strictEqual(redactToken('Token=ABCDEF0123456789ABCDEF0123456789ABCDEF01'), 'Token=<TOKEN>');
  assert.strictEqual(redactToken('Authorization: Token 0123456789abcdef0123456789abcdef01234567'), 'Authorization: Token <TOKEN>');
  assert.strictEqual(redactToken('{"eventName":"subscribe","authorization":"0123456789abcdef0123456789abcdef01234567"}'), '{"eventName":"subscribe","authorization":"<TOKEN>"}');
  assert.strictEqual(redactToken('{ "token": "0123456789abcdef0123456789abcdef01234567" }'), '{ "token": "<TOKEN>" }');
  assert.strictEqual(redactToken('token=<TOKEN>'), 'token=<TOKEN>');
```

Run: `cd tools/gendocs && npm test` → Expected: FAIL (slugFromHref 1개, redactToken 여러 개 assertion 실패)

`tools/gendocs/lib.mjs` 의 세 함수를 아래로 교체한다.

```js
// slugFromHref: '/documentation/corporate-actions/dividends' → 'dividends' (?query/#hash 제거)
export function slugFromHref(href) {
  return String(href).split(/[?#]/)[0].replace(/\/+$/, '').split('/').pop();
}

// redactToken: 예시 코드의 토큰 값을 <TOKEN> 으로. 비로그인 안내 문장(공백 포함), 쿼리 `token=`,
// 헤더 `Authorization: Token <hex>`, JSON `"token"|"authorization": "<...>"` 형태를 대소문자 무시로
// 치환한다(원문 표기는 캡처로 보존). 실제 Tiingo 토큰은 40자 hex 이므로 영숫자 20자 이상을 값으로 본다.
export function redactToken(text) {
  return String(text)
    .replace(/(token=)Not logged-in[^"'\n&]*/gi, '$1<TOKEN>')
    .replace(/(token=)[A-Za-z0-9]{20,}/gi, '$1<TOKEN>')
    .replace(/(Token\s+)[A-Za-z0-9]{20,}/g, '$1<TOKEN>')
    .replace(/("(?:token|authorization)"\s*:\s*")[A-Za-z0-9]{20,}(")/gi, '$1<TOKEN>$2');
}

// escapeCell: md 표 셀용 — 앞뒤 공백 제거, '|' 이스케이프, 줄바꿈(연속 포함) → <br>
export function escapeCell(text) {
  return String(text).trim().replace(/\|/g, '\\|').replace(/\s*\n\s*/g, '<br>');
}
```

Run: `cd tools/gendocs && npm test` → Expected: `# pass 7`, `# fail 0`

```bash
git add tools/gendocs/lib.mjs tools/gendocs/lib.test.mjs
git commit -m "fix(gendocs): redactToken 헤더/JSON/대소문자 형태 확장, slug 쿼리 제거 (Task 2 리뷰 반영)"
```

- [ ] **Step 1: 실패하는 테스트 추가**

`tools/gendocs/lib.test.mjs` 상단 import 를 아래로 바꾸고, 파일 끝에 테스트를 추가한다.

```js
import { slugFromHref, groupDir, buildNav, redactToken, escapeCell, renderTable, orderTabs, renderBlocks, renderPage } from './lib.mjs';
```

```js
test('orderTabs puts Request, Response, Example first and renames Examples', () => {
  const tabs = [{ name: 'Response', blocks: [] }, { name: 'Request', blocks: [] }, { name: 'Examples', blocks: [] }, { name: 'Other', blocks: [] }];
  assert.deepStrictEqual(orderTabs(tabs).map((t) => t.name), ['Request', 'Response', 'Example', 'Other']);
});

test('renderBlocks maps headings (h1/h2→##, h3→###), lists, code langs', () => {
  const md = renderBlocks([
    { type: 'heading', level: 1, text: '3.4.1 Overview' },
    { type: 'para', md: 'Intro **bold** [link](https://x).' },
    { type: 'list', ordered: false, items: ['one', 'two'] },
    { type: 'list', ordered: true, items: ['first'] },
    { type: 'heading', level: 3, text: 'Reference Price Update Messages' },
    { type: 'code', lang: '', text: '# Meta Data\nhttps://api.tiingo.com/tiingo/daily/<ticker>' },
    { type: 'code', lang: 'json', text: '[{"a":1}]' },
  ]);
  assert.strictEqual(md, [
    '## 3.4.1 Overview', '',
    'Intro **bold** [link](https://x).', '',
    '- one', '- two', '',
    '1. first', '',
    '### Reference Price Update Messages', '',
    '```', '# Meta Data', 'https://api.tiingo.com/tiingo/daily/<ticker>', '```', '',
    '```json', '[{"a":1}]', '```', '',
  ].join('\n'));
});

test('renderBlocks renders tabs one level below the current heading, reordered, empty tabs omitted, token redacted', () => {
  const md = renderBlocks([
    { type: 'heading', level: 2, text: '2.1.2 End-of-Day Endpoint' },
    { type: 'tabs', tabs: [
      { name: 'Response', blocks: [{ type: 'table', header: ['Field Name', 'JSON Field'], rows: [['Date', 'date']] }] },
      { name: 'Request', blocks: [{ type: 'para', md: 'params' }] },
      { name: 'Examples', blocks: [
        { type: 'code', lang: 'python', text: 'requests.get("https://api.tiingo.com/x?token=Not logged-in or registered. Please login or register to see your API Token")' },
        { type: 'code', lang: 'json', text: '[]' },
      ] },
    ] },
  ]);
  const i = (s) => md.indexOf(s);
  assert.ok(i('### Request') > -1 && i('### Response') > i('### Request') && i('### Example') > i('### Response'));
  assert.match(md, /\| Field Name \| JSON Field \|/);
  assert.match(md, /```python\nrequests\.get\("https:\/\/api\.tiingo\.com\/x\?token=<TOKEN>"\)\n```/);
  assert.doesNotMatch(md, /Not logged-in/);
});

test('renderBlocks omits tabs whose blocks render to nothing', () => {
  const md = renderBlocks([
    { type: 'heading', level: 2, text: 'S' },
    { type: 'tabs', tabs: [{ name: 'Response', blocks: [] }, { name: 'Request', blocks: [{ type: 'para', md: 'p' }] }] },
  ]);
  assert.doesNotMatch(md, /### Response/);
  assert.match(md, /### Request\n\np/);
});

test('renderBlocks drops the "Just remember, you will need your token" boilerplate paragraph', () => {
  const md = renderBlocks([
    { type: 'para', md: 'Just remember, you will need your token in order to connect. Keep it safe.' },
    { type: 'para', md: 'Click here to see your API Token.' },
    { type: 'para', md: 'Real content.' },
  ]);
  assert.strictEqual(md, 'Real content.\n');
});

test('renderPage adds title, source line, and an Endpoints section for leading code blocks', () => {
  const md = renderPage({
    title: 'End-of-Day (EOD) Stock Price API Documentation',
    sourceUrl: 'https://www.tiingo.com/documentation/end-of-day',
    generatedAt: '2026-09-04',
    blocks: [
      { type: 'para', md: 'REST Endpoints' },
      { type: 'code', lang: '', text: '# Meta Data\nhttps://api.tiingo.com/tiingo/daily/<ticker>' },
      { type: 'heading', level: 2, text: '2.1.1 Overview' },
      { type: 'para', md: 'Body.' },
    ],
  });
  assert.strictEqual(md, [
    '# End-of-Day (EOD) Stock Price API Documentation', '',
    '> 출처: https://www.tiingo.com/documentation/end-of-day · 생성: 2026-09-04 (tools/gendocs)', '',
    '## Endpoints', '',
    'REST Endpoints', '',
    '```', '# Meta Data', 'https://api.tiingo.com/tiingo/daily/<ticker>', '```', '',
    '## 2.1.1 Overview', '',
    'Body.', '',
  ].join('\n'));
});

test('renderPage without leading code blocks renders leading paragraphs without an Endpoints section', () => {
  const md = renderPage({ title: 'T', sourceUrl: 'u', generatedAt: 'd', blocks: [{ type: 'para', md: 'Lead.' }, { type: 'heading', level: 2, text: 'H' }] });
  assert.doesNotMatch(md, /## Endpoints/);
  assert.match(md, /^# T\n\n> 출처: u · 생성: d \(tools\/gendocs\)\n\nLead\.\n\n## H\n$/);
});

test('renderPage redacts tokens that appear outside code blocks', () => {
  const md = renderPage({ title: 'T', sourceUrl: 'u', generatedAt: 'd', blocks: [
    { type: 'para', md: 'Use ?token=0123456789abcdef0123456789abcdef01234567 in the URL.' },
    { type: 'table', header: ['a'], rows: [['Authorization: Token 0123456789abcdef0123456789abcdef01234567']] },
  ] });
  assert.doesNotMatch(md, /0123456789abcdef/);
  assert.match(md, /\?token=<TOKEN> in the URL/);
  assert.match(md, /Authorization: Token <TOKEN>/);
});
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `cd tools/gendocs && npm test`
Expected: FAIL — `orderTabs`/`renderBlocks`/`renderPage` export 없음 (SyntaxError: The requested module does not provide an export named 'orderTabs')

- [ ] **Step 3: `lib.mjs` 끝에 구현 추가**

```js
// ---- 블록 → md ----

const TAB_ORDER = { Request: 0, Response: 1, Example: 2 };

// orderTabs: 웹의 Response→Request→Examples 를 Request→Response→Example 순으로. 'Examples' 는 'Example' 로 개명.
export function orderTabs(tabs) {
  const renamed = tabs.map((t) => ({ ...t, name: t.name === 'Examples' ? 'Example' : t.name }));
  return renamed
    .map((t, i) => ({ t, i, k: t.name in TAB_ORDER ? TAB_ORDER[t.name] : 100 + i }))
    .sort((a, b) => a.k - b.k)
    .map((x) => x.t);
}

// isBoilerplate: 모든 페이지에 반복되는 토큰 안내 문단
function isBoilerplate(md) {
  return /^Just remember, you will need your token/.test(md) || /^Click here to see your API Token/.test(md);
}

// renderBlocks: 블록 배열 → md. level 은 현재 섹션 헤딩 깊이(기본 2). h1/h2 → ##, h3 → ###.
// 탭 헤딩은 현재 섹션보다 한 단계 아래(최대 ####).
export function renderBlocks(blocks, level = 2) {
  const out = [];
  let cur = level;
  for (const b of blocks) {
    switch (b.type) {
      case 'heading':
        cur = b.level <= 2 ? 2 : 3;
        out.push(`${'#'.repeat(cur)} ${b.text}`, '');
        break;
      case 'para':
        if (b.md && !isBoilerplate(b.md)) out.push(b.md, '');
        break;
      case 'list':
        if (b.items.length) out.push(...b.items.map((it, i) => `${b.ordered ? `${i + 1}.` : '-'} ${it}`), '');
        break;
      case 'code':
        out.push('```' + (b.lang || ''), redactToken(b.text), '```', '');
        break;
      case 'table': {
        const t = renderTable(b);
        if (t) out.push(t, '');
        break;
      }
      case 'tabs': {
        const h = '#'.repeat(Math.min(cur + 1, 4));
        for (const tab of orderTabs(b.tabs)) {
          const inner = renderBlocks(tab.blocks, cur);
          if (!inner.trim()) continue;
          out.push(`${h} ${tab.name}`, '', inner);
        }
        break;
      }
      default:
        break;
    }
  }
  return out.join('\n').replace(/\n{3,}/g, '\n\n');
}

// renderPage: 페이지 md. 첫 헤딩 앞의 블록(엔드포인트 URL 코드블록 등)은 '## Endpoints' 아래에 둔다
// (코드블록이 하나도 없으면 헤딩 없이 그대로).
export function renderPage({ title, sourceUrl, generatedAt, blocks }) {
  const firstHeading = blocks.findIndex((b) => b.type === 'heading');
  const lead = firstHeading === -1 ? blocks : blocks.slice(0, firstHeading);
  const rest = firstHeading === -1 ? [] : blocks.slice(firstHeading);
  const parts = [`# ${title}`, '', `> 출처: ${sourceUrl} · 생성: ${generatedAt} (tools/gendocs)`, ''];
  if (lead.length) {
    if (lead.some((b) => b.type === 'code')) parts.push('## Endpoints', '');
    parts.push(renderBlocks(lead));
  }
  if (rest.length) parts.push(renderBlocks(rest));
  // 최종 문자열에 한 번 더 치환 — 코드블록 밖(문단·표)에 토큰이 렌더된 경우도 잡는다.
  return redactToken(parts.join('\n').replace(/\n{3,}/g, '\n\n').trimEnd() + '\n');
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `cd tools/gendocs && npm test`
Expected: `# pass 15`, `# fail 0` (Task 3 리뷰 반영 커밋 후에는 17)

- [ ] **Step 5: 커밋**

```bash
git add tools/gendocs/lib.mjs tools/gendocs/lib.test.mjs
git commit -m "feat(gendocs): 블록 → md 렌더 (헤딩/탭 재정렬/표/코드/보일러플레이트 제거)"
```

- [ ] **Step 6: Task 3 코드 리뷰 반영** (리뷰어 지적: 비로그인 안내 문장이 헤더 `Token Not logged-in…`/따옴표 필드 `"authorization":"Not logged-in…"` 형태에서 치환되지 않음, Python dict 작은따옴표 `'authorization':'…'` 미대응, `Token\s+` 에 `i` 없음, 보일러플레이트 앞 공백, 빈 헤딩 `## `)

`redactToken` 을 공용 값 패턴 `TOKEN_VALUE = (?:Not logged-in[^"'\n&]*|[A-Za-z0-9]{20,})` 로 세 형태(`token=`, `token[ \t]+`, `["'](token|authorization)["']\s*:\s*(["'])…\2`) 모두 대소문자 무시로 통일. `isBoilerplate` 는 `^\s*` 허용. `renderBlocks` 의 heading 은 빈 텍스트면 건너뜀. 테스트: redactToken assertion 5개 추가 + `renderBlocks` 테스트 2개(빈 헤딩/앞 공백 보일러플레이트, level 3 및 헤딩 전 탭 → `####`/`###`) 추가 → `# pass 17`.

```bash
git commit -m "fix(gendocs): redactToken 헤더/작은따옴표 필드의 비로그인 문구 치환, 빈 헤딩 가드 (Task 3 리뷰 반영)"
```

---

### Task 4: `lib.mjs` — 인덱스 렌더 (`renderIndex`, `parseSourceRows`)

**Files:**
- Modify: `tools/gendocs/lib.mjs` (끝에 추가)
- Modify: `tools/gendocs/lib.test.mjs` (끝에 추가)

- [ ] **Step 1: 실패하는 테스트 추가**

import 줄에 `parseSourceRows, renderIndex` 를 추가하고 파일 끝에:

```js
test('parseSourceRows reads the llms rows from an existing README, defaults to "-"', () => {
  const readme = [
    '| 파일 | 출처 | Last updated | 가져온 날짜 |',
    '| --- | --- | --- | --- |',
    '| `llms.txt` | https://www.tiingo.com/llms.txt | 2026-08-18 | 2026-09-04 |',
    '| `llms-full.txt` | https://www.tiingo.com/llms-full.txt | 2026-08-14 | 2026-09-04 |',
  ].join('\n');
  assert.deepStrictEqual(parseSourceRows(readme), {
    'llms.txt': { updated: '2026-08-18', fetched: '2026-09-04' },
    'llms-full.txt': { updated: '2026-08-14', fetched: '2026-09-04' },
  });
  assert.deepStrictEqual(parseSourceRows(''), {
    'llms.txt': { updated: '-', fetched: '-' },
    'llms-full.txt': { updated: '-', fetched: '-' },
  });
});

test('renderIndex lists groups and pages in nav order with source rows', () => {
  const nav = [
    { group: '1. General', dir: 'general', pages: [{ title: '1.1 Overview', slug: 'overview', href: '/documentation/general/overview' }] },
    { group: '2. REST', dir: 'rest', pages: [{ title: '2.1 End-of-Day', slug: 'end-of-day', href: '/documentation/end-of-day' }, { title: '2.10 Dividends', slug: 'dividends', href: '/documentation/corporate-actions/dividends' }] },
  ];
  const md = renderIndex(nav, { 'llms.txt': { updated: '2026-08-18', fetched: '2026-09-04' }, 'llms-full.txt': { updated: '-', fetched: '-' } });
  assert.match(md, /^# Tiingo API 문서 카탈로그/);
  assert.match(md, /총 3개 페이지/);
  assert.match(md, /\| `llms\.txt` \| https:\/\/www\.tiingo\.com\/llms\.txt \| 2026-08-18 \| 2026-09-04 \|/);
  assert.match(md, /\| `llms-full\.txt` \| https:\/\/www\.tiingo\.com\/llms-full\.txt \| - \| - \|/);
  assert.match(md, /### 1\. General\n\n- \[1\.1 Overview\]\(general\/overview\.md\)/);
  assert.match(md, /### 2\. REST\n\n- \[2\.1 End-of-Day\]\(rest\/end-of-day\.md\)\n- \[2\.10 Dividends\]\(rest\/dividends\.md\)/);
});
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `cd tools/gendocs && npm test`
Expected: FAIL — `parseSourceRows` export 없음

- [ ] **Step 3: `lib.mjs` 끝에 구현 추가**

```js
// ---- 인덱스 (docs/api/README.md) ----

export const SOURCE_FILES = ['llms.txt', 'llms-full.txt'];
const SOURCE_URL = (name) => `https://www.tiingo.com/${name}`;

// parseSourceRows: 기존 README 의 원본 표 두 행에서 Last updated / 가져온 날짜를 읽는다.
// fetch-docs.sh 가 sed 로 갱신하는 값이라, gendocs 가 README 를 다시 쓸 때 보존해야 한다.
export function parseSourceRows(readme) {
  const rows = {};
  for (const name of SOURCE_FILES) {
    const re = new RegExp(`^\\| \`${name.replace('.', '\\.')}\` \\| [^|]* \\| ([^|]*) \\| ([^|]*) \\|`, 'm');
    const m = String(readme || '').match(re);
    rows[name] = m ? { updated: m[1].trim(), fetched: m[2].trim() } : { updated: '-', fetched: '-' };
  }
  return rows;
}

// renderIndex: nav + 원본 표 → docs/api/README.md
export function renderIndex(nav, sourceRows) {
  const total = nav.reduce((n, g) => n + g.pages.length, 0);
  const lines = [
    '# Tiingo API 문서 카탈로그',
    '',
    `Tiingo 문서 사이트(https://www.tiingo.com/documentation) 를 자동 변환한 md 총 ${total}개 페이지(tools/gendocs)와`,
    'Tiingo 가 공식 제공하는 llms 원본 2개를 보관합니다. `tiingo-go` SDK 개발의 1차 참조입니다.',
    '',
    '- 페이지 md: 웹 문서를 페이지 단위로 변환. 엔드포인트별 Request/Response 필드 표(타입·설명)와 Python 예시 + 응답 예시 JSON 포함.',
    '- llms 원본: Tiingo 가 유지하는 개념·정책·플랜 제한·심볼 규칙의 source of truth. 일부 응답 필드는 타입 없이 이름만 나열돼 있어 페이지 md 로 보완.',
    '',
    '## 공식 원본 (Tiingo 제공)',
    '',
    '| 파일 | 출처 | Last updated | 가져온 날짜 |',
    '| --- | --- | --- | --- |',
    ...SOURCE_FILES.map((n) => `| \`${n}\` | ${SOURCE_URL(n)} | ${sourceRows[n].updated} | ${sourceRows[n].fetched} |`),
    '',
    '## 재생성',
    '',
    '```bash',
    './scripts/fetch-docs.sh                      # llms.txt / llms-full.txt 갱신 (curl)',
    'cd tools/gendocs && npm install && npx playwright install chromium && npm run gen   # 페이지 md 재생성',
    '```',
    '',
    '## 페이지',
    '',
  ];
  for (const g of nav) {
    lines.push(`### ${g.group}`, '');
    for (const p of g.pages) lines.push(`- [${p.title}](${g.dir}/${p.slug}.md)`);
    lines.push('');
  }
  return lines.join('\n');
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `cd tools/gendocs && npm test`
Expected: `# pass 19`, `# fail 0`

- [ ] **Step 5: 커밋**

```bash
git add tools/gendocs/lib.mjs tools/gendocs/lib.test.mjs
git commit -m "feat(gendocs): docs/api/README.md 인덱스 렌더 + 원본 표 보존"
```

---

### Task 5: `gendocs.mjs` — 사이드바 열거 (ENUM_ONLY 스모크)

**Files:**
- Create: `tools/gendocs/gendocs.mjs`

- [ ] **Step 0: Task 4 코드 리뷰 반영** (사소: `parseSourceRows` 의 정규식 이스케이프가 첫 `.` 만 처리, README 문구 "llms 원본 2개" 하드코딩, JS↔bash 행 포맷 계약을 잠그는 라운드트립 테스트 부재)

`lib.test.mjs` 끝에 `renderIndex → parseSourceRows` 라운드트립 테스트 1개 추가(`assert.deepStrictEqual(parseSourceRows(renderIndex(nav, rows)), rows)`), `lib.mjs` 의 `parseSourceRows` 이스케이프를 `name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')` 로, `renderIndex` 의 "2개" 를 `${SOURCE_FILES.length}개` 로. `npm test` → `# pass 20`.

```bash
git commit -m "test(gendocs): 인덱스 원본 표 라운드트립 테스트, 정규식 이스케이프 일반화 (Task 4 리뷰 반영)"
```

- [ ] **Step 1: `gendocs.mjs` 작성 (열거 부분까지)**

```js
// Tiingo API docs 카탈로그 생성기.
// 사용: node gendocs.mjs                 (전체 23페이지)
//       ENUM_ONLY=1 node gendocs.mjs     (사이드바 열거 결과만 출력)
//       LIMIT=3 node gendocs.mjs         (앞 3페이지만 — 검증용)
//       ONLY=end-of-day node gendocs.mjs (slug 하나만)
//       RETRY=1 node gendocs.mjs         (failures.log 의 실패 href 만 재처리)
import { chromium } from '@playwright/test';
import { writeFile, mkdir, readFile, access } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { buildNav, renderPage, renderIndex, parseSourceRows } from './lib.mjs';

const UA = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36';
const BASE = 'https://www.tiingo.com';
const ENTRY = `${BASE}/documentation/general/overview`;
const HERE = path.dirname(fileURLToPath(import.meta.url));
const OUT_ROOT = path.resolve(HERE, '../../docs/api'); // tools/gendocs → repo/docs/api
const FAILURES = path.join(HERE, 'failures.log');
const DELAY_MS = 500;
const EXPECTED_PAGES = 23;

// enumerate: 사이드바(mat-sidenav)의 링크를 문서 순서로 읽어 nav 구조를 만든다.
async function enumerate(page) {
  await page.goto(ENTRY, { waitUntil: 'networkidle', timeout: 60000 });
  await page.waitForSelector('mat-sidenav a.side-bar-link-container', { timeout: 30000 });
  const links = await page.$$eval('mat-sidenav a.side-bar-link-container', (as) =>
    as.map((a) => ({
      href: a.getAttribute('href'),
      title: (a.querySelector('.side-bar-link-title')?.innerText || a.innerText).trim().replace(/\s+/g, ' '),
      leaf: a.classList.contains('indent-level-1'),
    })).filter((l) => l.href && l.href.startsWith('/documentation')),
  );
  return buildNav(links);
}

async function main() {
  const browser = await chromium.launch();
  try {
    // 새 컨텍스트 = 쿠키 없음(비로그인). 예시 코드에 실제 토큰이 렌더되지 않도록 한다.
    const ctx = await browser.newContext({ userAgent: UA, viewport: { width: 1440, height: 900 } });
    const page = await ctx.newPage();

    const nav = await enumerate(page);
    const total = nav.reduce((n, g) => n + g.pages.length, 0);
    console.log(`enumerated ${total} pages in ${nav.length} groups`);
    if (total !== EXPECTED_PAGES) console.warn(`  WARN: expected ${EXPECTED_PAGES} pages (site changed?)`);
    if (process.env.ENUM_ONLY) {
      for (const g of nav) { console.log(`${g.group} → ${g.dir}/`); for (const p of g.pages) console.log(`  ${p.title}  ${p.href}  → ${p.slug}.md`); }
      return;
    }
    // (Task 6 에서 페이지 처리 추가)
  } finally {
    await browser.close();
  }
}

main().catch((e) => { console.error(e); process.exit(1); });
```

- [ ] **Step 2: 열거 스모크 실행**

Run: `cd tools/gendocs && ENUM_ONLY=1 npm run gen`
Expected 출력(순서 그대로):

```
enumerated 23 pages in 5 groups
1. General → general/
  1.1 Overview  /documentation/general/overview  → overview.md
  1.2 Connecting  /documentation/general/connecting  → connecting.md
  1.3 Changelog  /documentation/general/changelog  → changelog.md
2. REST → rest/
  2.1 End-of-Day  /documentation/end-of-day  → end-of-day.md
  ... (news, crypto, forex, equity-realtime-stock-data, iex, boats, fundamentals, mutual-fund-and-etf-fees, dividends, splits)
3. Websockets → websockets/
  ... (crypto, forex, equity-realtime-stock-data, iex, boats)
4. Utilities → utilities/
  4.1 Search  /documentation/utilities/search  → search.md
5. Appendix → appendix/
  ... (developers, integrations, symbology)
```

23개가 아니거나 그룹 제목이 다르면(예: `3. Websockets` 가 `3. WebSockets`), `groupDir` 는 소문자화하므로 디렉터리는 영향 없다. leaf 판정(`indent-level-1`)이 어긋나 0개면 사이드바 셀렉터를 브라우저 DevTools 로 다시 확인한다(2026-09-04 확인값: `a.side-bar-link-container`, 그룹 링크는 indent 클래스 없음).

- [ ] **Step 3: 커밋**

```bash
git add tools/gendocs/gendocs.mjs
git commit -m "feat(gendocs): 사이드바 열거 + ENUM_ONLY 스모크"
```

**Task 5 코드 리뷰 반영**: 열거 결과 0개면 `throw`(README 빈 목록 덮어쓰기 방지), `a.pathname` + `startsWith('/documentation/')` 필터, `textContent` 제목, 빈 dir/slug 불변식 검사. 커밋 `fix(gendocs): 열거 0개면 실패, pathname 기반 필터, textContent 제목 (Task 5 리뷰 반영)`.

**구현 시 확인된 이탈(승인)**: 실제 leaf 는 23개(스펙의 24 는 합계 오기) → `EXPECTED_PAGES = 23`. 사이드바 하단 전역 링크(`Home`, `Documentation`(`/documentation`), `Products` …)는 `side-bar-link-container permanent` 클래스라 `Documentation → documentation/` 빈 그룹이 생김 → 셀렉터를 `mat-sidenav a.side-bar-link-container:not(.permanent)` 로.

---

### Task 6: `gendocs.mjs` — 페이지 추출·렌더·쓰기·재시도·인덱스

**Files:**
- Modify: `tools/gendocs/gendocs.mjs`

- [ ] **Step 1: `extractPage` 추가** (`enumerate` 함수 아래에)

브라우저 안에서 실행된다. **Tiingo DOM 셀렉터는 이 함수에만 둔다.**

```js
// extractPage: 현재 페이지의 tiingo-api-canvas 를 문서 순서로 걸어 블록 배열을 만든다(브라우저 컨텍스트).
// 셀렉터(2026-09-04 확인): tiingo-api-canvas, h1-h3, p, ul/ol, pre, mat-tab-group > .mat-tab-label /
// .mat-tab-body-active, tiingo-doc-table .header-row .header-cell / .parameter-row .parameter-cell
async function extractPage(page) {
  return await page.evaluate(async () => {
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
    const canvas = document.querySelector('tiingo-api-canvas');
    const title = document.title.replace(/\s*\|\s*Tiingo\s*$/, '').trim();
    if (!canvas) return { title, blocks: [], warn: 'no tiingo-api-canvas' };

    const BLOCK_TAGS = new Set(['div', 'p', 'ul', 'ol', 'pre', 'h1', 'h2', 'h3', 'h4', 'table', 'mat-tab-group', 'tiingo-doc-table', 'section', 'article']);
    const clean = (s) => s.replace(/[ \t\u00a0]+/g, ' ').replace(/ *\n */g, '\n').trim();

    // inline: 요소 안의 텍스트를 md 인라인으로 (링크는 절대 URL, code 는 백틱, strong 은 **)
    const inline = (node) => {
      let out = '';
      for (const n of node.childNodes) {
        if (n.nodeType === Node.TEXT_NODE) { out += n.textContent; continue; }
        if (n.nodeType !== Node.ELEMENT_NODE) continue;
        const tag = n.tagName.toLowerCase();
        if (tag === 'a' && n.getAttribute('href')) out += `[${clean(inline(n))}](${n.href})`;
        else if (tag === 'code') out += '`' + n.textContent.trim() + '`';
        else if (tag === 'strong' || tag === 'b') out += `**${clean(inline(n))}**`;
        else if (tag === 'br') out += '\n';
        else out += inline(n);
      }
      return out;
    };

    const table = (el) => ({
      type: 'table',
      header: [...el.querySelectorAll('.header-row .header-cell')].map((c) => c.innerText.trim()),
      rows: [...el.querySelectorAll('.parameter-row')].map((r) => [...r.querySelectorAll('.parameter-cell')].map((c) => c.innerText.trim())),
    });

    const hasBlockChild = (el) => [...el.querySelectorAll('*')].some((c) => BLOCK_TAGS.has(c.tagName.toLowerCase()));

    // walk: root 의 자식을 문서 순서로 순회. lang 은 언어 탭 안이면 'python'.
    const walk = async (root, blocks, lang) => {
      for (const el of root.children) {
        const tag = el.tagName.toLowerCase();
        if (/^h[1-3]$/.test(tag)) {
          const text = clean(el.innerText);
          if (text) blocks.push({ type: 'heading', level: Number(tag[1]), text });
        } else if (tag === 'p' || (tag === 'div' && !hasBlockChild(el) && clean(el.innerText))) {
          const md = clean(inline(el));
          if (md) blocks.push({ type: 'para', md });
        } else if (tag === 'ul' || tag === 'ol') {
          blocks.push({ type: 'list', ordered: tag === 'ol', items: [...el.querySelectorAll(':scope > li')].map((li) => clean(inline(li))).filter(Boolean) });
        } else if (tag === 'pre') {
          const text = el.innerText.replace(/\s+$/, '');
          if (!text.trim()) continue;
          const isJson = /^[\[{]/.test(text.trim());
          blocks.push({ type: 'code', lang: isJson ? 'json' : lang, text });
        } else if (tag === 'tiingo-doc-table') {
          blocks.push(table(el));
        } else if (tag === 'mat-tab-group') {
          await tabs(el, blocks);
        } else if (el.children.length) {
          await walk(el, blocks, lang);
        }
      }
    };

    // tabs: 탭 라벨을 클릭해 활성 본문을 읽는다. 라벨에 'Python' 이 있으면 언어 탭 그룹 → Python 만.
    // 중첩 그룹(Examples 안의 언어 탭)을 구분하기 위해 closest('mat-tab-group') 이 자기 자신인 것만 취한다.
    const tabs = async (group, blocks) => {
      const own = (sel) => [...group.querySelectorAll(sel)].filter((e) => e.closest('mat-tab-group') === group);
      const labels = own('.mat-tab-label');
      const names = labels.map((l) => l.innerText.trim());
      const isLang = names.includes('Python');
      const out = [];
      for (const label of labels) {
        const name = label.innerText.trim();
        if (isLang && name !== 'Python') continue;
        label.click();
        await sleep(300);
        const body = own('.mat-tab-body-active')[0];
        const inner = [];
        if (body) await walk(body, inner, isLang ? 'python' : '');
        if (isLang) blocks.push(...inner);
        else out.push({ name, blocks: inner });
      }
      if (!isLang) blocks.push({ type: 'tabs', tabs: out });
    };

    const blocks = [];
    await walk(canvas, blocks, '');
    return { title, blocks };
  });
}
```

- [ ] **Step 2: `main` 의 `// (Task 6 에서 페이지 처리 추가)` 자리를 아래로 교체**

```js
    // 처리 대상 선택
    let targets = nav.flatMap((g) => g.pages.map((p) => ({ ...p, dir: g.dir })));
    if (process.env.RETRY) {
      const raw = await readFile(FAILURES, 'utf8').catch(() => '');
      const failed = new Set(raw.split('\n').map((l) => l.trim()).filter(Boolean));
      targets = targets.filter((t) => failed.has(t.href));
      console.log(`retry mode: ${targets.length} pages from failures.log`);
    }
    if (process.env.ONLY) targets = targets.filter((t) => t.slug === process.env.ONLY);
    if (process.env.LIMIT) {
      const limit = parseInt(process.env.LIMIT, 10);
      if (Number.isNaN(limit)) throw new Error(`invalid LIMIT: ${process.env.LIMIT}`);
      targets = targets.slice(0, limit);
    }

    // 로컬 날짜(KST) — fetch-docs.sh 의 `date +%Y-%m-%d` 와 같은 기준
    const d = new Date();
    const generatedAt = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
    const failures = [];
    let ok = 0;
    for (const t of targets) {
      const url = BASE + t.href;
      let result = null;
      for (let attempt = 0; attempt < 2 && !result; attempt++) {
        if (attempt > 0) await page.waitForTimeout(3000);
        try {
          await page.goto(url, { waitUntil: 'networkidle', timeout: 60000 });
          await page.waitForSelector('tiingo-api-canvas', { timeout: 30000 });
          await page.waitForTimeout(1000);
          const r = await extractPage(page);
          if (r.warn) console.warn(`  WARN ${t.href}: ${r.warn}`);
          if (r.blocks.length) result = r;
          else console.warn(`  empty blocks: ${t.href} (attempt ${attempt + 1})`);
        } catch (e) {
          console.warn(`  attempt ${attempt + 1} failed: ${t.href} (${e.message})`);
        }
      }
      if (!result) { failures.push(t.href); continue; }
      const dir = path.join(OUT_ROOT, t.dir);
      await mkdir(dir, { recursive: true });
      await writeFile(path.join(dir, `${t.slug}.md`), renderPage({ title: result.title, sourceUrl: url, generatedAt, blocks: result.blocks }));
      ok++;
      console.log(`  wrote ${t.dir}/${t.slug}.md (${result.blocks.length} blocks)`);
      await page.waitForTimeout(DELAY_MS);
    }

    // 인덱스: nav 순서대로, 디스크에 실제 존재하는 파일만. 원본 표 두 행은 기존 README 에서 보존.
    const exists = async (p) => access(p).then(() => true, () => false);
    const present = [];
    for (const g of nav) {
      const pages = [];
      for (const p of g.pages) if (await exists(path.join(OUT_ROOT, g.dir, `${p.slug}.md`))) pages.push(p);
      if (pages.length) present.push({ ...g, pages });
    }
    if (present.length) {
      const oldReadme = await readFile(path.join(OUT_ROOT, 'README.md'), 'utf8').catch(() => '');
      await mkdir(OUT_ROOT, { recursive: true });
      await writeFile(path.join(OUT_ROOT, 'README.md'), renderIndex(present, parseSourceRows(oldReadme)));
    } else {
      console.warn('  생성된 페이지가 없어 README.md 를 갱신하지 않음');
    }
    await writeFile(FAILURES, failures.join('\n'));
    console.log(`done: ${ok} ok, ${failures.length} failed (failures.log); index ${present.reduce((n, g) => n + g.pages.length, 0)} pages`);
```

- [ ] **Step 3: 한 페이지로 실행해 결과 검수**

Run: `cd tools/gendocs && ONLY=end-of-day npm run gen && sed -n '1,80p' ../../docs/api/rest/end-of-day.md`
Expected: `wrote rest/end-of-day.md (N blocks)`, `done: 1 ok, 0 failed`. 파일에 다음이 모두 있어야 한다.
- `# End-of-Day (EOD) Stock Price API Documentation` 와 `> 출처: https://www.tiingo.com/documentation/end-of-day · 생성: 2026-..`
- `## Endpoints` 아래 코드블록에 `https://api.tiingo.com/tiingo/daily/<ticker>`
- `## 2.1.1 Overview`, `## 2.1.2 End-of-Day Endpoint`, `## 2.1.3 Meta Endpoint`
- `### Request` 표에 `| Resample Freq | GET | resampleFreq | string | N | ...daily: Values returned...<br>weekly: ... |` (셀 내 줄바꿈이 `<br>`)
- `### Response` 표 13행(`date` … `splitFactor`)
- `### Example` 에 ```` ```python ```` 블록(`token=<TOKEN>`)과 ```` ```json ```` 블록(`"date":"2019-01-02T00:00:00.000Z"`)
- "Just remember, you will need your token" 문단 없음

문제가 있으면 원인별로: 표가 비었으면 `tiingo-doc-table` 셀 클래스, 탭이 비었으면 `.mat-tab-body-active` 가 클릭 300ms 안에 안 바뀐 것(→ `sleep(300)` 을 600 으로), 산문이 빠졌으면 `p` 가 아닌 컨테이너(→ `hasBlockChild` 규칙 확인). 셀렉터 수정은 `extractPage` 안에서만.

- [ ] **Step 4: 탭 없는 페이지와 h3 중첩 페이지도 검수**

Run: `cd tools/gendocs && ONLY=overview npm run gen && ONLY=iex npm run gen`
(`ONLY=iex` 는 `rest/iex.md` 와 `websockets/iex.md` 둘 다 처리한다 — slug 가 같음.)
Expected:
- `docs/api/general/overview.md`: `## Endpoints` 없음, 인증·rate limit·포맷 문단과 헤딩이 있음, 빈 헤딩 없음.
- `docs/api/websockets/iex.md`: `## 3.4.1 Overview`(h1 이지만 `##`), `## 3.4.2 Reference Price (Derived Data Calculation)`, `### Reference Price Update Messages`(h3), 탭 헤딩 `### Request` / `### Response`, 탭 안의 h3(`Reference Price Update Messages`)는 Task 7 Step 0 이후 `####`, `wss://api.tiingo.com/iex` 코드블록.

- [ ] **Step 5: 단위 테스트 재실행 + 커밋**

Run: `cd tools/gendocs && npm test`
Expected: `# pass 20`, `# fail 0`

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
git add tools/gendocs/gendocs.mjs
git commit -m "feat(gendocs): 페이지 추출(탭·표·예시) + 렌더·쓰기·재시도·인덱스"
```

(생성된 `docs/api/*.md` 는 아직 커밋하지 않는다 — Task 7 에서 전체 생성 후 한 번에.)

**구현 시 확인된 이탈(승인)**: (1) 섹션마다 반복되는 브레드크럼 `div.documentation-breadcrumb-top-header`("2.1 REST - End-of-Day Prices")가 leaf div 규칙에 문단으로 잡혀 `walk` 에서 건너뜀. (2) `clean` 의 NBSP 를 `\u00a0` 로 이스케이프. (3) `general/overview` 1.1.2 의 "Your API Token is:" 아래 맨 `pre` 가 비로그인 안내문 전체라 `token=` 패턴이 아님 → `pre` 텍스트가 안내문과 정확히 같으면 `<TOKEN>` 으로 치환(상수 `NOT_LOGGED_IN`). **확인된 사이트 사실**: `websockets/iex` 의 h3("Reference Price Update Messages", `"eventData" Request Parameters`)는 탭 **안**에 있어 탭 헤딩과 같은 `###` 로 렌더됨 → Task 7 Step 0 에서 `renderBlocks` 가 탭 안 헤딩을 탭 헤딩+1 단계로 내리도록 수정. `general/overview` 1.1.1 의 "You can check them out here:" 뒤 목록은 비로그인 페이지에 렌더되지 않음(사이트 사실, 크롤러 손실 아님).

---

### Task 7: 전체 23페이지 생성 + 검증 + 커밋

**Files:**
- Create: `docs/api/README.md`, `docs/api/{general,rest,websockets,utilities,appendix}/*.md` (23개)

- [ ] **Step 0: Task 6 코드 리뷰 반영 + 탭 안 헤딩 단계 수정** (한 커밋)

리뷰 지적: (a) `walk` 가 `h1~h3` 만 수집해 `h4` 가 조용히 버려짐 — `rest/fundamentals` 2.8.6 FAQ 질문 7개 소실 확인. (b) 탭 본문 안 h3 가 탭 헤딩과 같은 단계로 렌더. (c) 버려진 요소에 대한 관측성 없음. (d) `ONLY`/`LIMIT` 부분 실행이 `failures.log` 를 덮어씀. (e) `document.title` 이 비면 제목 폴백 없음. (f) `NOT_LOGGED_IN` 정확 일치 대신 `startsWith`.

`lib.test.mjs` 끝에 추가(먼저 실패 확인):

```js
test('renderBlocks renders headings inside a tab one level below the tab heading', () => {
  const md = renderBlocks([
    { type: 'heading', level: 2, text: '3.4.2 Reference Price' },
    { type: 'tabs', tabs: [{ name: 'Response', blocks: [
      { type: 'para', md: 'top-level fields' },
      { type: 'heading', level: 3, text: 'Reference Price Update Messages' },
      { type: 'para', md: 'update fields' },
    ] }] },
    { type: 'heading', level: 2, text: '3.4.3 Next' },
  ]);
  assert.match(md, /### Response\n\ntop-level fields\n\n#### Reference Price Update Messages\n\nupdate fields\n\n## 3\.4\.3 Next/);
});

test('renderBlocks keeps h4-h6 depth (capped at 6)', () => {
  const md = renderBlocks([
    { type: 'heading', level: 2, text: '2.8.6 FAQ' },
    { type: 'heading', level: 4, text: 'Do you use XBRL?' },
    { type: 'para', md: 'Yes.' },
    { type: 'heading', level: 6, text: 'deep' },
  ]);
  assert.strictEqual(md, '## 2.8.6 FAQ\n\n#### Do you use XBRL?\n\nYes.\n\n###### deep\n');
});
```

`lib.mjs` `renderBlocks`:

```js
// renderBlocks: 블록 배열 → md. level 은 현재 섹션 헤딩 깊이(기본 2). h1/h2 → ##, h3~h6 → 같은 깊이(최대 6).
// headingFloor 는 헤딩의 최소 깊이 — 탭 본문 안에서는 탭 헤딩+1 로 내려 계층을 유지한다.
// 탭 헤딩은 현재 섹션보다 한 단계 아래(최대 5).
export function renderBlocks(blocks, level = 2, headingFloor = 2) {
  const out = [];
  let cur = level;
  for (const b of blocks) {
    switch (b.type) {
      case 'heading':
        if (!b.text || !b.text.trim()) break;
        cur = Math.min(Math.max(b.level <= 2 ? 2 : b.level, headingFloor), 6);
        out.push(`${'#'.repeat(cur)} ${b.text.trim()}`, '');
        break;
      // ... para / list / code / table 분기는 그대로 ...
      case 'tabs': {
        const depth = Math.min(cur + 1, 5);
        const h = '#'.repeat(depth);
        for (const tab of orderTabs(b.tabs)) {
          const inner = renderBlocks(tab.blocks, depth, Math.min(depth + 1, 6));
          if (!inner.trim()) continue;
          out.push(`${h} ${tab.name}`, '', inner);
        }
        break;
      }
```

`gendocs.mjs`:
- `extractPage`: 헤딩 정규식 `/^h[1-6]$/`, `BLOCK_TAGS` 에 `h5`, `h6` 추가. `dropped` 배열을 두고 `walk` 의 마지막 분기(자식 없음)에서 텍스트가 있으면 `dropped.push(`<${tag}> ${clean(el.innerText).slice(0, 60)}`)`, `tabs` 에서 라벨이 0개면 `dropped.push('<mat-tab-group> 라벨 없음')`. 반환 `{ title, blocks, dropped }`. `NOT_LOGGED_IN` 비교는 `text.trim().startsWith('Not logged-in')`. 캔버스 셀렉터는 상수 `CANVAS = 'tiingo-api-canvas'` 로 빼서 `page.evaluate(fn, CANVAS)` 인자와 `waitForSelector(CANVAS)` 양쪽에 사용.
- `main`: `r.dropped.length` 면 `console.warn(`  WARN ${t.href}: dropped ${n} — ${list.join(' | ')}`)`. `renderPage({ title: result.title || t.title, ... })`. `failures.log` 는 `ONLY`/`LIMIT` 가 없을 때만 다시 쓴다(전체/RETRY 실행 결과만 기록).

`npm test` → `# pass 22`, `# fail 0`. `ONLY=fundamentals npm run gen` 후 `grep -c '^#### ' docs/api/rest/fundamentals.md` = 7 이상(FAQ 질문), `ONLY=iex` 후 `websockets/iex.md` 에 `#### Reference Price Update Messages`.

```bash
git add tools/gendocs/lib.mjs tools/gendocs/lib.test.mjs tools/gendocs/gendocs.mjs
git commit -m "fix(gendocs): h4~h6 헤딩 수집·계층 렌더, 탭 안 헤딩 단계, 드롭 요소 경고, 제목 폴백 (Task 6 리뷰 반영)"
```

- [ ] **Step 1: 전체 실행**

Run: `cd tools/gendocs && npm run gen`
Expected: `enumerated 23 pages in 5 groups` … `done: 23 ok, 0 failed (failures.log); index 23 pages`. 1~2분 소요.
실패가 있으면 `RETRY=1 npm run gen` 으로 재시도.

- [ ] **Step 2: 개수·실패·토큰 검증**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
find docs/api -name '*.md' -not -name README.md | wc -l          # 23
wc -c tools/gendocs/failures.log                                  # 0
grep -rl "Not logged-in" docs/api || echo "no leaked login text"  # no leaked login text
grep -rE "token=[A-Za-z0-9]{20,}" docs/api || echo "no token"     # no token
grep -rL "^## " docs/api/*/*.md || echo "all pages have sections" # all pages have sections
```

- [ ] **Step 3: 인덱스 링크 검증**

```bash
grep -oE '\]\([a-z]+/[a-z0-9-]+\.md\)' docs/api/README.md | tr -d '()]' | sort > /tmp/links.txt
find docs/api -name '*.md' -not -name README.md | sed 's#^docs/api/##' | sort > /tmp/files.txt
diff /tmp/links.txt /tmp/files.txt && echo "index OK"
```
Expected: `index OK` (링크 23개 = 파일 23개, 차이 없음). `/tmp` 대신 스크래치 디렉터리를 써도 된다.

- [ ] **Step 4: 육안 검수 3개**

Run: `sed -n '1,60p' docs/api/rest/fundamentals.md`
Expected: `2.8.5 Meta Data` 섹션의 `### Response` 표에 `permaTicker` … `dailyLastUpdated` 16행과 타입(`datetime`, `int32`, `boolean`)이 있다. `2.8.6 Additional Information & FAQ` 섹션이 있다.

Run: `grep -c '^#' docs/api/general/changelog.md docs/api/appendix/symbology.md; grep -c '^#### ' docs/api/rest/fundamentals.md`
Expected: 앞 둘은 각각 1 이상(빈 페이지 아님), fundamentals 는 7 이상(FAQ h4).

- [ ] **Step 5: 멱등 확인**

Run: `cd tools/gendocs && LIMIT=2 npm run gen && cd ../.. && git status --short docs/api | head`
Expected: 재생성된 2개 파일은 `생성:` 날짜가 같은 날이면 diff 없음(`git status` 에 변경 없음). 날짜가 바뀐 경우에만 그 줄 1개 diff.

- [ ] **Step 6: 커밋**

```bash
git add docs/api
git commit -m "docs: Tiingo 문서 23페이지 md 카탈로그 생성 (tools/gendocs)"
```

**구현 시 확인된 사항(승인)**: Step 2 의 `grep "Not logged-in"` 이 `general/connecting.md` 의 WebSocket Python 예시(`'authToken': 'Not logged-in…'`)를 잡음 → `redactToken` 따옴표 필드 키를 `(?:[a-z_]*token|authorization)` 로 확장(테스트 1개 추가, `# pass 23`, 커밋 `fix(gendocs): redactToken 이 authToken 따옴표 필드도 가리도록`). 전체 크롤에서 `dropped` 경고는 Equity Realtime/BOATS 4페이지의 `<span> BETA` 배지뿐. 사이트 원문 그대로인 특이점: `connecting.md` 의 빈 `## 1.2 Connecting` 섹션, `changelog.md` 의 날짜 `##` 바로 아래 `###`, `symbology.md` 의 `{SYMBOL}-{SHARE CLASS}` 템플릿이 `isJson` 휴리스틱으로 ```` ```json ```` 라벨(코스메틱).

---

### Task 8: `scripts/fetch-docs.sh` + llms 원본 보관

**Files:**
- Create: `scripts/fetch-docs.sh`
- Create: `docs/api/llms.txt`, `docs/api/llms-full.txt`
- Modify: `docs/api/README.md` (원본 표 두 행 — 스크립트가 갱신)

- [ ] **Step 1: 스크립트 작성**

```bash
#!/usr/bin/env bash
# fetch-docs.sh — Tiingo 공식 llms 원본 2개를 docs/api/ 로 가져온다.
#
#   ./scripts/fetch-docs.sh
#
# 동작:
#   1. https://www.tiingo.com/llms.txt, llms-full.txt 를 임시 디렉토리에 다운로드
#      — 하나라도 실패하면 아무것도 반영하지 않고 중단
#   2. 내용이 바뀐 파일만 docs/api/ 로 교체
#   3. 교체된 파일에 대해 docs/api/README.md 의 "공식 원본" 표 행(Last updated·가져온 날짜) 갱신
#      (Last updated 는 파일 안의 'Last updated: YYYY-MM-DD' 첫 등장값, 없으면 '-')
#
# 의존: curl
# TIINGO_DOCS_BASE 환경변수로 원본 base URL 을 바꿀 수 있다(테스트용).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/docs/api"
README="$OUT/README.md"
BASE="${TIINGO_DOCS_BASE:-https://www.tiingo.com}"
UA="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"
FILES=(llms.txt llms-full.txt)

if ! command -v curl >/dev/null 2>&1; then
  echo "error: 'curl' 가 필요합니다" >&2
  exit 1
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# 1) 전부 다운로드 (하나라도 실패하면 set -e 로 여기서 중단 → docs/api 미변경)
for f in "${FILES[@]}"; do
  echo "GET $BASE/$f"
  curl -fsSL -A "$UA" "$BASE/$f" -o "$TMP/$f"
  if [[ ! -s "$TMP/$f" ]]; then
    echo "error: $f 가 비어 있음" >&2
    exit 1
  fi
done

TODAY="$(date +%Y-%m-%d)"
mkdir -p "$OUT"

# last_updated <file> — 'Last updated: YYYY-MM-DD'(llms.txt) 또는 'last_updated: "YYYY-MM-DD"'(llms-full.txt
# 프론트매터) 첫 등장값의 날짜, 없으면 빈 문자열(호출부에서 '-')
last_updated() {
  grep -m1 -oiE 'last[ _]updated:? *"?[0-9]{4}-[0-9]{2}-[0-9]{2}' "$1" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' || true
}

# update_readme_row <name> <updated> — README 원본 표에서 해당 행의 Last updated·가져온 날짜를 바꾼다
update_readme_row() {
  local name="$1" updated="$2"
  if [[ ! -f "$README" ]]; then
    echo "  warn: $README 가 없어 원본 표를 갱신하지 못함 (tools/gendocs 로 먼저 생성)" >&2
    return 0
  fi
  local esc="${name//./\\.}"
  sed -E "s#^(\| \`$esc\` \| [^|]* \|) [^|]* \| [^|]* \|#\1 $updated | $TODAY |#" "$README" > "$TMP/README.md"
  mv "$TMP/README.md" "$README"
  if ! grep -qF "| \`$name\` | $BASE/$name | $updated | $TODAY |" "$README"; then
    echo "  warn: README 원본 표에서 $name 행을 찾지 못해 갱신되지 않음" >&2
  fi
}

# 2), 3)
for f in "${FILES[@]}"; do
  if [[ -f "$OUT/$f" ]] && cmp -s "$TMP/$f" "$OUT/$f"; then
    echo "  $f: 변경 없음"
    continue
  fi
  mv "$TMP/$f" "$OUT/$f"
  upd="$(last_updated "$OUT/$f")"
  [[ -n "$upd" ]] || upd="-"
  update_readme_row "$f" "$upd"
  echo "  $f: 갱신 (Last updated=$upd)"
done

echo "done → $OUT"
```

- [ ] **Step 2: 실행 권한 + 실행**

Run: `chmod +x scripts/fetch-docs.sh && ./scripts/fetch-docs.sh`
Expected:
```
GET https://www.tiingo.com/llms.txt
GET https://www.tiingo.com/llms-full.txt
  llms.txt: 갱신 (Last updated=2026-08-18)
  llms-full.txt: 갱신 (Last updated=2026-08-14)
done → .../docs/api
```
`update_readme_row` 의 `grep -qF` 검증은 `$BASE/$name` 이 README 의 출처 열(`https://www.tiingo.com/llms.txt`)과 같을 때만 통과한다. 기본 BASE 로 실행하면 일치한다.

- [ ] **Step 3: 결과 확인**

```bash
wc -l docs/api/llms.txt docs/api/llms-full.txt     # 약 85 / 1525
grep -n 'llms' docs/api/README.md                   # 두 행에 Last updated 와 오늘 날짜
./scripts/fetch-docs.sh                             # 두 번째 실행: "변경 없음" 2줄, README diff 없음
git status --short
```

- [ ] **Step 4: gendocs 재실행 시 원본 표가 보존되는지 확인**

Run: `cd tools/gendocs && LIMIT=1 npm run gen && cd ../.. && grep -n 'llms' docs/api/README.md`
Expected: 두 행의 Last updated·가져온 날짜가 Step 2 값 그대로(`-` 로 되돌아가지 않음).

- [ ] **Step 5: 커밋**

```bash
git add scripts/fetch-docs.sh docs/api/llms.txt docs/api/llms-full.txt docs/api/README.md
git commit -m "docs: Tiingo 공식 llms.txt/llms-full.txt 보관 + fetch-docs.sh"
```

---

### Task 9: 루트 README + 최종 검증 + PR

**Files:**
- Modify: `README.md`

- [ ] **Step 1: `README.md` 작성**

```markdown
# tiingo-go

[Tiingo](https://www.tiingo.com) 금융 데이터 API(주식 EOD/실시간 시세, 뉴스, 펀더멘털, 암호화폐, 외환)의 Go 클라이언트 라이브러리. **구현 예정** — 현재는 API 문서 카탈로그 단계입니다.

## 문서

- [`docs/api/README.md`](docs/api/README.md) — Tiingo 문서 사이트 23페이지를 변환한 md + Tiingo 공식 `llms.txt`/`llms-full.txt` 원본.
- 재생성: `./scripts/fetch-docs.sh`(llms 원본), `cd tools/gendocs && npm install && npx playwright install chromium && npm run gen`(페이지 md).
- 설계: [`docs/superpowers/specs/`](docs/superpowers/specs/)
```

- [ ] **Step 2: 최종 검증 일괄 실행**

```bash
cd /Users/user/src/workspace_moneyflow/tiingo-go
(cd tools/gendocs && npm test)                                    # pass 23, fail 0
find docs/api -name '*.md' -not -name README.md | wc -l          # 23
test ! -s tools/gendocs/failures.log && echo "no failures"
grep -rE "Not logged-in|token=[A-Za-z0-9]{20,}" docs/api || echo "no token leak"
git status --short                                                # README.md 만 변경
file -I README.md docs/api/README.md docs/api/rest/end-of-day.md  # 모두 charset=utf-8
```

- [ ] **Step 3: 커밋 + 푸시 + PR**

```bash
git add README.md
git commit -m "docs: README — 문서 카탈로그 안내"
git push -u origin chore/api-docs
gh pr create --title "docs: Tiingo API 문서 카탈로그 (웹 23페이지 md + 공식 llms 원본)" --body "$(cat <<'EOF'
## Summary
- Tiingo 문서 사이트(Angular SPA) 23페이지를 Playwright 크롤러(`tools/gendocs`)로 페이지 단위 md 변환 → `docs/api/{general,rest,websockets,utilities,appendix}/`
- 엔드포인트별 Request/Response 필드 표(타입·설명) + Python 예시 + 응답 예시 JSON, `token=` 은 `<TOKEN>` 치환
- Tiingo 공식 `llms.txt` / `llms-full.txt` 원본 보관 + `scripts/fetch-docs.sh`
- 설계: `docs/superpowers/specs/2026-09-04-api-docs-catalog-design.md`, 계획: `docs/superpowers/plans/2026-09-04-api-docs-catalog.md`
- Go 코드 없음(`go.mod` 는 SDK 스펙에서)

## Test plan
- [x] `cd tools/gendocs && npm test` — 23 pass
- [x] `npm run gen` — 23 ok / 0 failed, `docs/api/README.md` 링크 23개 = 파일 23개
- [x] `grep` 으로 토큰·로그인 문구 유출 없음
- [x] `./scripts/fetch-docs.sh` 2회 실행 — 2회째 "변경 없음", README 원본 표 보존
- [x] 육안: `rest/end-of-day.md`, `rest/fundamentals.md`(meta 16필드 타입), `websockets/iex.md`(h3 + #### 탭), `general/overview.md`

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

리뷰어는 지정하지 않는다(사용자가 직접 지정).

---

## 자체 검토 (스펙 대조)

| 스펙 항목 | 태스크 |
|---|---|
| 사이드바 열거, 23개 아니면 경고 | Task 5 |
| `extractPage` 셀렉터 한 곳, h1-h3/p/list/pre/tabs/table/Examples(Python만, json) | Task 6 |
| 비로그인 컨텍스트 + `token=` 치환 이중 방어 | Task 5(newContext), Task 2(redactToken), Task 3(renderBlocks) |
| 렌더: 제목, 출처 줄, `##`/`###` 매핑, 탭 Request→Response→Example, 빈 탭 생략, 표 `<br>`/`\|`, 코드 언어 | Task 3 |
| `docs/api/README.md` 인덱스 (원본 표, 재생성, 그룹별 목록) + 원본 표 보존 | Task 4, Task 6 |
| 500ms 대기, 1회 재시도, failures.log, LIMIT/RETRY | Task 6 |
| `fetch-docs.sh` (curl, 임시→이동, Last updated, README sed, 멱등) | Task 8 |
| 검증(23개, failures 비움, 링크, 토큰 grep, 육안 3개, 멱등) | Task 7, Task 9 |
| `.gitignore`, 루트 README, PR | Task 1, Task 9 |

스펙과의 사소한 차이: 첫 헤딩 앞 엔드포인트 블록의 헤딩을 `## REST Endpoints` 가 아니라 `## Endpoints` 로 둔다(WebSocket 페이지에도 같은 규칙을 쓰기 위해).
