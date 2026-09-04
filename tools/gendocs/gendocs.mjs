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
      href: a.pathname,
      title: (a.querySelector('.side-bar-link-title')?.textContent || a.textContent).trim().replace(/\s+/g, ' '),
      leaf: a.classList.contains('indent-level-1'),
    })).filter((l) => l.href && l.href.startsWith('/documentation/')),
  );
  return buildNav(links);
}

// extractPage: 현재 페이지의 tiingo-api-canvas 를 문서 순서로 걸어 블록 배열을 만든다(브라우저 컨텍스트).
// 셀렉터(2026-09-04 확인): tiingo-api-canvas, h1-h3, p, ul/ol, pre, mat-tab-group > .mat-tab-label /
// .mat-tab-body-active, tiingo-doc-table .header-row .header-cell / .parameter-row .parameter-cell
async function extractPage(page) {
  return await page.evaluate(async () => {
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
    const canvas = document.querySelector('tiingo-api-canvas');
    const title = document.title.replace(/\s*\|\s*Tiingo\s*$/, '').trim();
    if (!canvas) return { title, blocks: [], warn: 'no tiingo-api-canvas' };

    // 비로그인 상태에서 토큰 자리에 렌더되는 문구(bare pre — token= 패턴이 아니라 렌더러의 redactToken 이 못 잡는다)
    const NOT_LOGGED_IN = 'Not logged-in or registered. Please login or register to see your API Token';
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
        // 섹션마다 반복되는 브레드크럼("2.1 REST - End-of-Day Prices")은 본문이 아니므로 건너뛴다.
        if (el.classList.contains('documentation-breadcrumb-top-header')) continue;
        if (/^h[1-3]$/.test(tag)) {
          const text = clean(el.innerText);
          if (text) blocks.push({ type: 'heading', level: Number(tag[1]), text });
        } else if (tag === 'p' || (tag === 'div' && !hasBlockChild(el) && clean(el.innerText))) {
          const md = clean(inline(el));
          if (md) blocks.push({ type: 'para', md });
        } else if (tag === 'ul' || tag === 'ol') {
          blocks.push({ type: 'list', ordered: tag === 'ol', items: [...el.querySelectorAll(':scope > li')].map((li) => clean(inline(li))).filter(Boolean) });
        } else if (tag === 'pre') {
          let text = el.innerText.replace(/\s+$/, '');
          if (!text.trim()) continue;
          if (text.trim() === NOT_LOGGED_IN) text = '<TOKEN>';
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

async function main() {
  const browser = await chromium.launch();
  try {
    // 새 컨텍스트 = 쿠키 없음(비로그인). 예시 코드에 실제 토큰이 렌더되지 않도록 한다.
    const ctx = await browser.newContext({ userAgent: UA, viewport: { width: 1440, height: 900 } });
    const page = await ctx.newPage();

    const nav = await enumerate(page);
    const total = nav.reduce((n, g) => n + g.pages.length, 0);
    if (total === 0) throw new Error('사이드바 열거 결과 0개 — 셀렉터(mat-sidenav a.side-bar-link-container)를 확인하세요');
    for (const g of nav) {
      if (!g.dir) throw new Error(`빈 그룹 디렉터리명: ${JSON.stringify(g.group)}`);
      for (const p of g.pages) if (!p.slug) throw new Error(`빈 slug: ${p.href}`);
    }
    console.log(`enumerated ${total} pages in ${nav.length} groups`);
    if (total !== EXPECTED_PAGES) console.warn(`  WARN: expected ${EXPECTED_PAGES} pages (site changed?)`);
    if (process.env.ENUM_ONLY) {
      for (const g of nav) { console.log(`${g.group} → ${g.dir}/`); for (const p of g.pages) console.log(`  ${p.title}  ${p.href}  → ${p.slug}.md`); }
      return;
    }
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
  } finally {
    await browser.close();
  }
}

main().catch((e) => { console.error(e); process.exit(1); });
