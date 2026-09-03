import { test } from 'node:test';
import assert from 'node:assert';
import { slugFromHref, groupDir, buildNav, redactToken, escapeCell, renderTable, orderTabs, renderBlocks, renderPage, parseSourceRows, renderIndex } from './lib.mjs';

test('slugFromHref returns last path segment', () => {
  assert.strictEqual(slugFromHref('/documentation/corporate-actions/dividends'), 'dividends');
  assert.strictEqual(slugFromHref('/documentation/end-of-day'), 'end-of-day');
  assert.strictEqual(slugFromHref('/documentation/websockets/iex/'), 'iex');
  assert.strictEqual(slugFromHref('/documentation/end-of-day?x=1#top'), 'end-of-day');
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
  assert.strictEqual(redactToken('Token=ABCDEF0123456789ABCDEF0123456789ABCDEF01'), 'Token=<TOKEN>');
  assert.strictEqual(redactToken('Authorization: Token 0123456789abcdef0123456789abcdef01234567'), 'Authorization: Token <TOKEN>');
  assert.strictEqual(redactToken('{"eventName":"subscribe","authorization":"0123456789abcdef0123456789abcdef01234567"}'), '{"eventName":"subscribe","authorization":"<TOKEN>"}');
  assert.strictEqual(redactToken('{ "token": "0123456789abcdef0123456789abcdef01234567" }'), '{ "token": "<TOKEN>" }');
  assert.strictEqual(redactToken('token=<TOKEN>'), 'token=<TOKEN>');
  assert.strictEqual(redactToken("'Authorization' : 'Token Not logged-in or registered. Please login or register to see your API Token'"), "'Authorization' : 'Token <TOKEN>'");
  assert.strictEqual(redactToken('{"eventName":"subscribe","authorization":"Not logged-in or registered. Please login or register to see your API Token"}'), '{"eventName":"subscribe","authorization":"<TOKEN>"}');
  assert.strictEqual(redactToken("{'eventName':'subscribe','authorization':'0123456789abcdef0123456789abcdef01234567'}"), "{'eventName':'subscribe','authorization':'<TOKEN>'}");
  assert.strictEqual(redactToken('token abcdefghijklmnopqrstuvwxyz1234'), 'token <TOKEN>');
  assert.strictEqual(redactToken('Click here to see your API Token.'), 'Click here to see your API Token.');
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

test('renderBlocks skips empty headings and tolerates leading whitespace in boilerplate', () => {
  const md = renderBlocks([
    { type: 'heading', level: 2, text: '   ' },
    { type: 'para', md: '  Just remember, you will need your token in order to connect.' },
    { type: 'para', md: 'Kept.' },
  ]);
  assert.strictEqual(md, 'Kept.\n');
});

test('renderBlocks renders tab headings at level+1 when called with level 3 or before any heading', () => {
  const tabs = { type: 'tabs', tabs: [{ name: 'Request', blocks: [{ type: 'para', md: 'p' }] }] };
  assert.match(renderBlocks([tabs], 3), /^#### Request\n\np\n$/);
  assert.match(renderBlocks([tabs]), /^### Request\n\np\n$/);
});

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

test('renderIndex → parseSourceRows round-trips the source rows', () => {
  const nav = [{ group: '1. General', dir: 'general', pages: [{ title: '1.1 Overview', slug: 'overview', href: '/documentation/general/overview' }] }];
  const rows = { 'llms.txt': { updated: '2026-08-18', fetched: '2026-09-04' }, 'llms-full.txt': { updated: '2026-08-14', fetched: '2026-09-05' } };
  assert.deepStrictEqual(parseSourceRows(renderIndex(nav, rows)), rows);
  assert.match(renderIndex(nav, rows), /llms 원본 2개/);
});
