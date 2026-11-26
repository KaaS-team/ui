// Shared readers for the CSS tests. No dependencies: the files are small and the grammar we need is narrow.
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

export const read = (path) => readFileSync(path, 'utf8');
export const stripComments = (css) => css.replace(/\/\*[\s\S]*?\*\//g, '');

// Declared custom properties: the light :root block and the dark block, as name sets.
export function tokenBlocks(css = read('css/tokens.css')) {
  const src = stripComments(css);
  const darkAt = src.indexOf('@media (prefers-color-scheme: dark)');
  const light = darkAt < 0 ? src : src.slice(0, darkAt);
  const dark = darkAt < 0 ? '' : src.slice(darkAt);
  const names = (s) => [...s.matchAll(/(--[\w-]+)\s*:/g)].map((m) => m[1]);
  return { light: names(light), dark: names(dark) };
}

export const tokensJSON = () => JSON.parse(read('tokens.json'));

export function walk(dir, skip = new Set(['.git', 'node_modules'])) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
    if (skip.has(e.name)) return [];
    const p = dir === '.' ? e.name : join(dir, e.name);
    return e.isDirectory() ? walk(p, skip) : [p];
  });
}
