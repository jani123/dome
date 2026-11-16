// Explicit integration check: real Hugo server, filesystem edits, and HTTP.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, cp, readFile, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { spawn } from 'node:child_process';
import { createServer } from 'node:net';
import { setTimeout as delay } from 'node:timers/promises';

async function waitForPage(url, marker) {
  for (let attempt = 0; attempt < 200; attempt++) {
    try {
      const response = await fetch(url);
      const html = await response.text();
      if (response.ok && html.includes(marker)) return html;
    } catch {}
    await delay(50);
  }
  throw new Error(`Preview failed to serve update: ${marker}`);
}
function checkPanels(html) {
  assert.doesNotMatch(html, /\{\{(?:&lt;|<)/);
  assert.equal((html.match(/class="command-tabs" data-tabs/g) || []).length, 4);
  assert.ok(html.indexOf('id="bootstrap"') < html.indexOf('id="bootstrap-go-tab"'));
  const openingHeading = html.indexOf('id="2-open-the-development-container-optional"');
  assert.ok(openingHeading >= 0 && openingHeading < html.indexOf('id="workspace-cli-tab"'));
}

test('Hugo preview renders shortcode panels correctly after source edits', { timeout: 30000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), 'dome-docs-preview-'));
  const probe = createServer();
  await new Promise(resolve => probe.listen(0, '127.0.0.1', resolve));
  const port = probe.address().port;
  await new Promise(resolve => probe.close(resolve));
  let server;
  try {
    await cp(resolve('docs'), join(directory, 'docs'), { recursive: true });
    server = spawn(resolve('.build/tools/hugo'), ['server', '--source', join(directory, 'docs'), '--destination', join(directory, 'public'), '--bind', '127.0.0.1', '--port', String(port), '--baseURL', `http://127.0.0.1:${port}/`, '--disableLiveReload', '--disableFastRender', '--panicOnWarning'], { stdio: 'ignore' });
    const url = `http://127.0.0.1:${port}/quickstart/`;
    checkPanels(await waitForPage(url, 'workspace-cli-tab'));
    const page = join(directory, 'docs/content/quickstart.md');
    const original = await readFile(page, 'utf8');
    for (const marker of ['Preview update one', 'Preview update two with a longer introduction']) {
      await writeFile(page, original.replace('## Prerequisites', `${marker}\n\n## Prerequisites`));
      checkPanels(await waitForPage(url, marker));
    }
  } finally {
    if (server && server.exitCode === null) {
      const stopped = new Promise(resolve => server.once('exit', resolve));
      server.kill();
      await stopped;
    }
    await rm(directory, { recursive: true, force: true });
  }
});
