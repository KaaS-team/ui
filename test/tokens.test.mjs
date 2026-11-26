import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync } from 'node:fs';
import { read, stripComments, tokenBlocks, tokenDecls, tokensJSON } from './css.mjs';

// tokens.json is the list consumers code against (the Go side and the QA dashboard). A token added to the
// CSS but not the list, or listed but never declared, is a name some consumer reads as empty.
const { light, dark, forced } = tokenBlocks();
const groups = Object.values(tokensJSON().groups);
const listed = groups.flatMap((g) => g.names);

test('tokens.css and tokens.json list exactly the same names', () => {
  assert.deepEqual([...new Set(light)].sort(), [...new Set(listed)].sort());
});

test('no token is declared or listed twice', () => {
  assert.deepEqual(light.filter((n, i) => light.indexOf(n) !== i), []);
  assert.deepEqual(listed.filter((n, i) => listed.indexOf(n) !== i), []);
});

test('every token is named --ui-*', () => {
  assert.deepEqual(light.filter((n) => !n.startsWith('--ui-')), []);
});

test('dark mode redefines exactly the groups marked dark', () => {
  const want = groups.filter((g) => g.dark).flatMap((g) => g.names);
  assert.deepEqual([...dark].sort(), [...want].sort());
});

// data-theme="dark" forces dark on a light OS. Its block is a copy of the media block; a copy that drifts
// gives one product two dark themes depending on how the user picked dark.
test("[data-theme='dark'] repeats the dark block exactly", () => {
  const d = tokenDecls();
  assert.deepEqual(forced, dark);
  assert.deepEqual(new Map(d.forced), new Map(d.dark));
});

// Without these the browser draws scrollbars and form controls in the OS scheme, not the forced one.
test('each data-theme sets its own color-scheme', () => {
  const src = stripComments(read('css/tokens.css'));
  for (const t of ['light', 'dark']) {
    assert.match(src, new RegExp(`:root\\[data-theme='${t}'\\]\\s*\\{[^}]*color-scheme:\\s*${t};`));
  }
});

// A var() of a name that does not exist falls back silently to the initial value, so the page just looks wrong.
// A layer may keep private variables for one component (a button's colours per variant); they are named --_*,
// never --ui-*, and must be declared in the same file that reads them.
test('every var() in css/ is a declared token or a --_* local of the same file', () => {
  const known = new Set(light);
  const bad = readdirSync('css')
    .filter((f) => f.endsWith('.css'))
    .flatMap((f) => {
      const src = stripComments(read(`css/${f}`));
      const local = new Set([...src.matchAll(/(--_[\w-]+)\s*:/g)].map((m) => m[1]));
      return [...src.matchAll(/var\((--[\w-]+)/g)]
        .map((m) => m[1])
        .filter((n) => !known.has(n) && !local.has(n))
        .map((n) => `css/${f}: ${n}`);
    });
  assert.deepEqual(bad, []);
});

// Only tokens.css declares --ui-* names; a layer that declares one shadows the token for its subtree.
test('no layer file declares a --ui-* property', () => {
  const bad = readdirSync('css')
    .filter((f) => f.endsWith('.css') && f !== 'tokens.css')
    .flatMap((f) => [...stripComments(read(`css/${f}`)).matchAll(/(--ui-[\w-]+)\s*:/g)].map((m) => `css/${f}: ${m[1]}`));
  assert.deepEqual(bad, []);
});

// ui.css is what consumers import; a layer file it skips never reaches a page.
test('ui.css imports tokens, base, then every layer file', () => {
  const imports = [...read('css/ui.css').matchAll(/@import url\('([\w.-]+)'\)/g)].map((m) => m[1]);
  const layers = readdirSync('css').filter((f) => f.endsWith('.css') && f !== 'ui.css');
  assert.deepEqual(imports.slice(0, 2), ['tokens.css', 'base.css']);
  assert.deepEqual([...imports].sort(), [...layers].sort());
});
