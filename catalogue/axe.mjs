// The axe check of /_ui/catalogue, run by `just axe` (part of `just check`). Every partial in every state is on
// that page, so zero violations there means zero for each partial. Each view is checked: light and dark, 1280 and 390.
import { createRequire } from 'node:module';
import { chromium } from 'playwright';
import { serve, views } from './serve.mjs';

const axeSource = createRequire(import.meta.url)('axe-core').source;
const server = serve();
let failed = 0;
try {
  const url = await server.url;
  const browser = await chromium.launch();
  for (const v of views) {
    const page = await browser.newPage({ viewport: { width: v.width, height: v.height }, colorScheme: v.colorScheme });
    await page.goto(url, { waitUntil: 'load' });
    await page.addScriptTag({ content: axeSource });
    const result = await page.evaluate(() => window.axe.run(document, { resultTypes: ['violations'] }));
    for (const r of result.violations) {
      failed++;
      console.error(`${v.name}: ${r.id} (${r.impact}): ${r.help}`);
      for (const n of r.nodes.slice(0, 5)) console.error(`  ${n.target.join(' ')}: ${n.failureSummary.split('\n').slice(1).join(' ').trim()}`);
    }
    console.log(`${v.name}: ${result.violations.length} violation(s), ${result.passes?.length ?? 0} rules checked`);
    await page.close();
  }
  await browser.close();
} finally {
  server.stop();
}
process.exit(failed ? 1 : 0);
