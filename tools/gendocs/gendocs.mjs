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
// .permanent 는 사이드바 하단의 전역 내비(Home/Documentation/Products…)라 제외한다.
async function enumerate(page) {
  await page.goto(ENTRY, { waitUntil: 'networkidle', timeout: 60000 });
  await page.waitForSelector('mat-sidenav a.side-bar-link-container:not(.permanent)', { timeout: 30000 });
  const links = await page.$$eval('mat-sidenav a.side-bar-link-container:not(.permanent)', (as) =>
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
