import { test } from 'node:test';
import assert from 'node:assert/strict';
import { themes, tokensJSON } from './css.mjs';

// WCAG 2.2 contrast for every pair the layers draw, in both themes. Each pair is its own test named with
// its ratio, so `just tokens` lists them all. AA: 4.5 for text (1.4.3), 3 for borders, rings and markers
// that identify a control or state (1.4.11). A token retune that drops a pair below its floor fails here.
const TEXT = 4.5;
const UI = 3;
const STATUSES = ['pass', 'wait', 'fail', 'info'];
const SURFACES = ['--ui-bg', '--ui-surface', '--ui-raised', '--ui-hover'];

const pairs = [
  // body and secondary text on every surface a page, panel, table header, log or hovered row uses
  ...SURFACES.flatMap((s) => [
    ['--ui-text', s, TEXT],
    ['--ui-muted', s, TEXT],
  ]),
  // filled buttons: primary at rest and hovered, danger
  ['--ui-on-accent', '--ui-accent', TEXT],
  ['--ui-on-accent', '--ui-accent-hover', TEXT],
  ['--ui-on-accent', '--ui-fail', TEXT],
  ...STATUSES.flatMap((s) => [
    // chips and attention titles: status text on its own tint
    [`--ui-${s}`, `--ui-${s}-bg`, TEXT],
    // attention body text sits on the tint too
    ['--ui-text', `--ui-${s}-bg`, TEXT],
    // log lines, stat deltas and timeline text: status text on panels and the log background
    [`--ui-${s}`, '--ui-surface', TEXT],
    [`--ui-${s}`, '--ui-raised', TEXT],
    [`--ui-${s}`, '--ui-bg', TEXT],
  ]),
  // control boundaries: input, select, checkbox and toggle borders
  ['--ui-border-strong', '--ui-surface', UI],
  ['--ui-border-strong', '--ui-raised', UI],
  // the focus ring on every surface it can land on
  ...SURFACES.map((s) => ['--ui-focus', s, UI]),
  // a primary button and a checked toggle read as filled against the panel
  ['--ui-accent', '--ui-surface', UI],
  ['--ui-accent', '--ui-raised', UI],
];

function luminance(hex) {
  const n = parseInt(hex.slice(1), 16);
  const [r, g, b] = [n >> 16, (n >> 8) & 255, n & 255].map((v) => {
    const c = v / 255;
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

export function contrast(a, b) {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

const { groups } = tokensJSON();
const colours = [...groups.colour.names, ...groups.status.names];
const resolved = themes();

// The ratio math reads #rrggbb only; a colour token in another form would be skipped, so it fails instead.
test('every colour and status token is #rrggbb in both themes', () => {
  const bad = Object.entries(resolved).flatMap(([theme, values]) =>
    colours.filter((n) => !/^#[0-9a-f]{6}$/i.test(values.get(n) ?? '')).map((n) => `${theme} ${n}: ${values.get(n)}`),
  );
  assert.deepEqual(bad, []);
});

test('contrast pairs name only colour and status tokens', () => {
  assert.deepEqual(pairs.flat().filter((n) => typeof n === 'string' && !colours.includes(n)), []);
});

for (const [theme, values] of Object.entries(resolved)) {
  for (const [fg, bg, floor] of pairs) {
    const ratio = contrast(values.get(fg), values.get(bg));
    const kind = floor === TEXT ? 'text' : 'ui';
    test(`${theme}: ${fg} on ${bg} ${ratio.toFixed(2)}:1 (${kind}, AA ${floor})`, () => {
      assert.ok(ratio >= floor, `${ratio.toFixed(2)} is below ${floor}`);
    });
  }
}
