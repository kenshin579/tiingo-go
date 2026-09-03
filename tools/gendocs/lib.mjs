// 순수 변환 로직 — Playwright/네트워크/파일시스템 의존 없음 (단위테스트 대상).

// slugFromHref: '/documentation/corporate-actions/dividends' → 'dividends' (?query/#hash 제거)
export function slugFromHref(href) {
  return String(href).split(/[?#]/)[0].replace(/\/+$/, '').split('/').pop();
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

// 토큰 값으로 볼 문자열: 비로그인 안내 문장(공백 포함, 따옴표/줄바꿈/& 앞까지) 또는 영숫자 20자 이상
// (실제 Tiingo 토큰은 40자 hex).
const TOKEN_VALUE = `(?:Not logged-in[^"'\\n&]*|[A-Za-z0-9]{20,})`;

// redactToken: 예시 코드의 토큰 값을 <TOKEN> 으로. 쿼리 `token=…`, 헤더 `Authorization: Token …`,
// 따옴표 필드 `"token"|"authorization": "…"`('…' 도) 세 형태를 대소문자 무시로 치환한다.
// 원문 표기(token=/Token /따옴표)는 캡처로 보존한다.
export function redactToken(text) {
  return String(text)
    .replace(new RegExp(`(token=)${TOKEN_VALUE}`, 'gi'), '$1<TOKEN>')
    .replace(new RegExp(`(token[ \\t]+)${TOKEN_VALUE}`, 'gi'), '$1<TOKEN>')
    .replace(new RegExp(`(["'](?:token|authorization)["']\\s*:\\s*)(["'])${TOKEN_VALUE}\\2`, 'gi'), '$1$2<TOKEN>$2');
}

// escapeCell: md 표 셀용 — 앞뒤 공백 제거, '|' 이스케이프, 줄바꿈(연속 포함) → <br>
export function escapeCell(text) {
  return String(text).trim().replace(/\|/g, '\\|').replace(/\s*\n\s*/g, '<br>');
}

// renderTable: {header, rows} → md 표. 헤더나 행이 없으면 빈 문자열.
export function renderTable({ header, rows }) {
  if (!Array.isArray(header) || header.length === 0 || !Array.isArray(rows) || rows.length === 0) return '';
  const line = (cells) => `| ${cells.map(escapeCell).join(' | ')} |`;
  return [line(header), `| ${header.map(() => '---').join(' | ')} |`, ...rows.map(line)].join('\n');
}

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

// isBoilerplate: 모든 페이지에 반복되는 토큰 안내 문단(앞 공백 허용)
function isBoilerplate(md) {
  return /^\s*(Just remember, you will need your token|Click here to see your API Token)/.test(md);
}

// renderBlocks: 블록 배열 → md. level 은 현재 섹션 헤딩 깊이(기본 2). h1/h2 → ##, h3 → ###.
// 탭 헤딩은 현재 섹션보다 한 단계 아래(최대 ####).
export function renderBlocks(blocks, level = 2) {
  const out = [];
  let cur = level;
  for (const b of blocks) {
    switch (b.type) {
      case 'heading':
        if (!b.text || !b.text.trim()) break;
        cur = b.level <= 2 ? 2 : 3;
        out.push(`${'#'.repeat(cur)} ${b.text.trim()}`, '');
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
