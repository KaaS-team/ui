import { test } from 'node:test';
import assert from 'node:assert/strict';
import { existsSync } from 'node:fs';
import { read, stripComments, walk } from './css.mjs';

// A colour written outside tokens.css skips dark mode and drifts from the scale; kaas.team ended up with a dozen
// greys before its token test ran. colour-allow.txt holds justified exceptions as `path  reason`, and an entry for a
// file that is now clean fails, so the list only shrinks. The idea follows the ALLOW ratchet in ai-bootcamp's
// check-structure.ts; no code is copied.
const TOKENS = 'css/tokens.css';
const SCANNED = /\.(css|tmpl|html|svg|go|js|mjs|ts|tsx)$/;
const SKIP = new Set(['.git', 'node_modules', 'scripts', 'test']);
export const COLOUR = /#[0-9a-fA-F]{3,8}\b|\b(?:rgba?|hsla?|hwb|oklch|oklab|lab|lch)\(/;

export function parseAllow(text) {
  const lines = text
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l && !l.startsWith('#'));
  const ceilingLine = lines.find((l) => l.startsWith('ceiling:'));
  const ceiling = ceilingLine === undefined ? NaN : Number(ceilingLine.slice('ceiling:'.length).trim());
  const entries = new Map(
    lines
      .filter((l) => l !== ceilingLine)
      .map((l) => {
        const [path, ...why] = l.split(/\s+/);
        return [path, why.join(' ')];
      }),
  );
  return { ceiling, entries };
}

const { ceiling, entries: allow } = parseAllow(read('colour-allow.txt'));

// CSS and HTML comments are prose; Go and script comments are kept, since a colour there is usually real.
const strip = (path, src) =>
  /\.(css|html|svg|tmpl)$/.test(path) ? stripComments(src).replace(/<!--[\s\S]*?-->/g, '') : src;

const raw = (path) =>
  strip(path, read(path))
    .split('\n')
    .map((line, i) => [i + 1, line])
    .filter(([, line]) => COLOUR.test(line));

const files = walk('.', SKIP).filter((f) => SCANNED.test(f) && f !== TOKENS);

test('no raw colour outside css/tokens.css', () => {
  const bad = files
    .filter((f) => !allow.has(f))
    .flatMap((f) => raw(f).map(([n, line]) => `${f}:${n}: ${line.trim()}`));
  assert.deepEqual(bad, []);
});

test('every colour-allow.txt entry exists, has a reason and still needs its exception', () => {
  const stale = [...allow].filter(([path, why]) => !why || !existsSync(path) || raw(path).length === 0);
  assert.deepEqual(stale.map(([p]) => p), []);
});

test('colour-allow.txt ceiling equals its entry count, so the list only grows in a reviewed diff', () => {
  assert.ok(Number.isInteger(ceiling), 'colour-allow.txt needs a `ceiling: N` line');
  assert.equal(allow.size, ceiling);
});

test('parseAllow reads the ceiling and entries apart', () => {
  const { ceiling: c, entries } = parseAllow('# x\nceiling: 1\ncss/demo.html  sample swatch\n');
  assert.equal(c, 1);
  assert.deepEqual([...entries], [['css/demo.html', 'sample swatch']]);
});

test('the colour pattern catches every CSS colour function and hex form', () => {
  for (const s of ['#fff', '#a1b2c3', '#a1b2c3d4', 'rgb(0 0 0)', 'rgba(0,0,0,.1)', 'hsl(0 0% 0%)', 'oklch(0.5 0.1 20)', 'hwb(0 0% 0%)']) {
    assert.match(s, COLOUR, s);
  }
  for (const s of ['var(--ui-text)', 'color-mix(in srgb, var(--ui-text), transparent)', 'currentColor', 'transparent']) {
    assert.doesNotMatch(s, COLOUR, s);
  }
});
