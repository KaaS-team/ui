// Builds ./cmd/catalogue, starts it on a free port and resolves to the catalogue URL. Shared by axe.mjs and screens.mjs.
// It runs the built binary rather than `go run`, so stop() kills the server itself and not only the go tool.
import { execFileSync, spawn } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

export function serve() {
  const bin = join(mkdtempSync(join(tmpdir(), 'ui-catalogue-')), 'catalogue');
  execFileSync('go', ['build', '-o', bin, './cmd/catalogue'], { stdio: 'inherit' });
  const child = spawn(bin, ['-addr', '127.0.0.1:0'], { stdio: ['ignore', 'pipe', 'inherit'] });
  const url = new Promise((resolve, reject) => {
    let out = '';
    child.stdout.on('data', (chunk) => {
      out += chunk;
      const line = out.split('\n')[0];
      if (out.includes('\n')) resolve(line.trim());
    });
    child.on('exit', (code) => reject(new Error(`catalogue server exited with ${code}`)));
  });
  return { url, stop: () => child.kill() };
}

// The four views every screen is checked in: two widths, light and dark.
export const views = [
  { name: 'light-1280', width: 1280, height: 800, colorScheme: 'light' },
  { name: 'dark-1280', width: 1280, height: 800, colorScheme: 'dark' },
  { name: 'light-390', width: 390, height: 844, colorScheme: 'light' },
  { name: 'dark-390', width: 390, height: 844, colorScheme: 'dark' },
];
