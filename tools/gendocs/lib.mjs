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
