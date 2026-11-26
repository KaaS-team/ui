// Shared readers for the CSS tests. No dependencies: the files are small and the grammar we need is narrow.
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

export const read = (path) => readFileSync(path, 'utf8');
export const stripComments = (css) => css.replace(/\/\*[\s\S]*?\*\//g, '');

// The body of the first rule whose selector is exactly `selector`, or '' when there is none.
// Token blocks hold only declarations, so the body ends at the first closing brace.
function ruleBody(src, selector) {
  const at = src.indexOf(`${selector} {`);
  if (at < 0) return '';
  const open = src.indexOf('{', at);
  return src.slice(open + 1, src.indexOf('}', open));
}

// Custom property declarations in a block body, in source order, as [name, value] pairs.
const decls = (body) => [...body.matchAll(/(--[\w-]+)\s*:\s*([^;]+);/g)].map((m) => [m[1], m[2].trim()]);

// The three token blocks in tokens.css, each as [name, value] pairs:
//   light:  :root, the defaults
//   dark:   :root:not([data-theme='light']) inside @media (prefers-color-scheme: dark)
//   forced: :root[data-theme='dark'], the manual override
export function tokenDecls(css = read('css/tokens.css')) {
  const src = stripComments(css);
  const mediaAt = src.indexOf('@media (prefers-color-scheme: dark)');
  const media = mediaAt < 0 ? '' : src.slice(mediaAt);
  return {
    light: decls(ruleBody(src, ':root')),
    dark: decls(ruleBody(media, ":root:not([data-theme='light'])")),
    forced: decls(ruleBody(src, ":root[data-theme='dark']")),
  };
}

// The same blocks as name lists.
export function tokenBlocks(css) {
  const d = tokenDecls(css);
  const names = (pairs) => pairs.map(([n]) => n);
  return { light: names(d.light), dark: names(d.dark), forced: names(d.forced) };
}

// Resolved values per theme: dark is light with the dark block laid over it.
export function themes(css) {
  const d = tokenDecls(css);
  const light = new Map(d.light);
  return { light, dark: new Map([...light, ...d.dark]) };
}

export const tokensJSON = () => JSON.parse(read('tokens.json'));

// Layer files: every css/*.css except tokens.css, which holds values rather than rules.
export const layerFiles = () =>
  readdirSync('css')
    .filter((f) => f.endsWith('.css') && f !== 'tokens.css')
    .map((f) => `css/${f}`);

export function walk(dir, skip = new Set(['.git', 'node_modules'])) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
    if (skip.has(e.name)) return [];
    const p = dir === '.' ? e.name : join(dir, e.name);
    return e.isDirectory() ? walk(p, skip) : [p];
  });
}
