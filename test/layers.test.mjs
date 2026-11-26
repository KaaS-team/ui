import { test } from 'node:test';
import assert from 'node:assert/strict';
import { read, stripComments, layerFiles } from './css.mjs';

// Layer files reach type and spacing only through tokens. A literal font size or stack is a second scale that
// no theme or density change can reach.
const declarations = (prop) =>
  layerFiles().flatMap((f) =>
    stripComments(read(f))
      .split('\n')
      .map((line, i) => [f, i + 1, line])
      .flatMap(([file, n, line]) =>
        [...line.matchAll(new RegExp(`(?:^|[;{\\s])${prop}\\s*:\\s*([^;]+)`, 'g'))].map((m) => `${file}:${n}: ${m[1].trim()}`),
      ),
  );

test('every font-size is a --ui-type-* token or inherit', () => {
  const bad = declarations('font-size').filter((d) => !/: (?:var\(--ui-type-[\w-]+\)|inherit)$/.test(d));
  assert.deepEqual(bad, []);
});

test('every font-family is a --ui-font-* token or inherit', () => {
  const bad = declarations('font-family').filter((d) => !/: (?:var\(--ui-font-[\w-]+\)|inherit)$/.test(d));
  assert.deepEqual(bad, []);
});

test('the font shorthand is only ever inherit', () => {
  assert.deepEqual(declarations('font').filter((d) => !/: inherit$/.test(d)), []);
});

// The demo is the screenshot source and the reference markup; it shows the classes doing all the work.
test('css/demo.html has no inline style and no <style> block', () => {
  const html = read('css/demo.html');
  assert.doesNotMatch(html, /\sstyle\s*=/i);
  assert.doesNotMatch(html, /<style[\s>]/i);
});

// Every layer class the demo uses is defined in a layer, so the demo never shows an unstyled component.
test('every ui-* class in css/demo.html is defined in a layer file', () => {
  const css = layerFiles().map((f) => stripComments(read(f))).join('\n');
  const defined = new Set([...css.matchAll(/\.(ui-[\w-]+)/g)].map((m) => m[1]));
  const used = new Set(
    [...read('css/demo.html').matchAll(/class="([^"]+)"/g)].flatMap((m) => m[1].split(/\s+/)).filter((c) => c.startsWith('ui-')),
  );
  assert.deepEqual([...used].filter((c) => !defined.has(c)), []);
});
