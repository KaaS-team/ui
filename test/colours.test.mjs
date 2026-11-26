import { test } from 'node:test';
import assert from 'node:assert/strict';
import { existsSync } from 'node:fs';
import { read, stripComments, walk } from './css.mjs';

// A colour written outside tokens.css skips dark mode and drifts from the scale; kaas.team ended up with a dozen
// greys before its token test ran. colour-allow.txt holds justified exceptions as `path  reason`, and an entry for a
// file that is now clean fails, so the list only shrinks.
const TOKENS = 'css/tokens.css';
const SCANNED = /\.(css|tmpl|html|go)$/;
const SKIP = new Set(['.git', 'node_modules', 'scripts', 'test']);
const COLOUR = /#[0-9a-fA-F]{3,8}\b|\b(?:rgba?|hsla?|oklch|oklab|lab|lch)\(/;

const allow = new Map(
  read('colour-allow.txt')
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l && !l.startsWith('#'))
    .map((l) => {
      const [path, ...why] = l.split(/\s+/);
      return [path, why.join(' ')];
    }),
);

const raw = (path) => {
  const src = path.endsWith('.go') ? read(path) : stripComments(read(path));
  return src
    .split('\n')
    .map((line, i) => [i + 1, line])
    .filter(([, line]) => COLOUR.test(line));
};

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
