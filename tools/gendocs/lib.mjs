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
