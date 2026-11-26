// Full-page screenshots of /_ui/catalogue in every view, for review: `just screens [dir]` (default reports/cloud/U2).
import { mkdirSync } from 'node:fs';
import { chromium } from 'playwright';
import { serve, views } from './serve.mjs';

const dir = process.argv[2] ?? 'reports/cloud/U2';
mkdirSync(dir, { recursive: true });
const server = serve();
try {
  const url = await server.url;
  const browser = await chromium.launch();
  for (const v of views) {
    const page = await browser.newPage({ viewport: { width: v.width, height: v.height }, colorScheme: v.colorScheme });
    await page.goto(url, { waitUntil: 'load' });
    const scroll = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    if (scroll > 0) console.warn(`${v.name}: page scrolls sideways by ${scroll}px`);
    await page.screenshot({ path: `${dir}/catalogue-${v.name}.png`, fullPage: true });
    await page.screenshot({ path: `${dir}/catalogue-${v.name}-fold.png` });
    console.log(`${dir}/catalogue-${v.name}.png`);
    await page.close();
  }
  await browser.close();
} finally {
  server.stop();
}
