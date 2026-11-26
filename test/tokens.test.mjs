import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync } from 'node:fs';
import { read, stripComments, tokenBlocks, tokensJSON } from './css.mjs';

// tokens.json is the list consumers code against (the Go side and the QA dashboard). A token added to the
// CSS but not the list, or listed but never declared, is a name some consumer reads as empty.
const { light, dark } = tokenBlocks();
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

// A var() of a name that does not exist falls back silently to the initial value, so the page just looks wrong.
test('every var(--ui-*) in css/ is a declared token', () => {
  const known = new Set(light);
  const bad = readdirSync('css')
    .filter((f) => f.endsWith('.css'))
    .flatMap((f) =>
      [...stripComments(read(`css/${f}`)).matchAll(/var\((--[\w-]+)/g)]
        .map((m) => m[1])
        .filter((n) => !known.has(n))
        .map((n) => `css/${f}: ${n}`),
    );
  assert.deepEqual(bad, []);
});

// ui.css is what consumers import; a layer file it skips never reaches a page.
test('ui.css imports tokens, base, then every layer file', () => {
  const imports = [...read('css/ui.css').matchAll(/@import url\('([\w.-]+)'\)/g)].map((m) => m[1]);
  const layers = readdirSync('css').filter((f) => f.endsWith('.css') && f !== 'ui.css');
  assert.deepEqual(imports.slice(0, 2), ['tokens.css', 'base.css']);
  assert.deepEqual([...imports].sort(), [...layers].sort());
});
