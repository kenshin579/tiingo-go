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
