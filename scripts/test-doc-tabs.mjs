import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { runInNewContext } from 'node:vm';

// Use generated HTML for initial state; mock browser events for the real tab script.
const html = readFileSync(process.env.DOME_DOC_TABS_HTML || '.build/docs/next/quickstart/index.html', 'utf8');
const script = readFileSync(new URL('../docs/assets/tabs.js', import.meta.url), 'utf8');
function fixture() {
  const elements = new Map();
  const document = { activeElement: null, getElementById: id => elements.get(id) };
  class Element extends EventTarget {
    constructor(attrs) {
      super();
      this.attrs = new Map([...attrs.matchAll(/([\w-]+)="([^"]*)"/g)].map(match => [match[1], match[2]]));
      this.hidden = /\shidden(?:\s|$)/.test(attrs);
      this.tabIndex = Number(this.attrs.get('tabindex') || 0);
      elements.set(this.attrs.get('id'), this);
    }
    getAttribute(name) { return this.attrs.get(name); }
    setAttribute(name, value) { this.attrs.set(name, value); }
    focus() { document.activeElement = this; }
  }
  const groups = [];
  for (const group of html.matchAll(/<div class="command-tabs" data-tabs>([\s\S]*?)<\/div>\s*<noscript>/g)) {
    const members = [];
    for (const match of group[1].matchAll(/<(?:button|section)\s+([^>]*role="tab(?:panel)?"[^>]*)>/g)) members.push(new Element(match[1]));
    const tabs = members.filter(element => element.getAttribute('role') === 'tab');
    assert.equal(tabs.length, 2);
    groups.push({ tabs, panels: tabs.map(tab => elements.get(tab.getAttribute('aria-controls'))), querySelectorAll: () => tabs });
  }
  assert.equal(groups.length, 4);
  document.querySelectorAll = () => groups;
  runInNewContext(script, { document });
  return { document, groups };
}

test('each group defaults to its first tab and switches panels in both directions', () => {
  const { groups } = fixture();
  for (const { tabs, panels } of groups) {
    assert.equal(tabs[0].getAttribute('aria-selected'), 'true');
    assert.equal(panels[0].hidden, false);
    assert.equal(panels[1].hidden, true);
    tabs[1].dispatchEvent(new Event('click'));
    assert.deepEqual(panels.map(panel => panel.hidden), [true, false]);
    assert.deepEqual(tabs.map(tab => tab.getAttribute('aria-selected')), ['false', 'true']);
    assert.deepEqual(tabs.map(tab => tab.tabIndex), [-1, 0]);
    tabs[0].dispatchEvent(new Event('click'));
    assert.deepEqual(panels.map(panel => panel.hidden), [false, true]);
  }
});

test('keyboard navigation switches panels and focus without header controls', () => {
  const { document, groups } = fixture();
  for (const { tabs, panels } of groups) {
    for (const [index, key, next] of [[0, 'ArrowRight', 1], [1, 'Home', 0], [0, 'End', 1], [1, 'ArrowLeft', 0]]) {
      const event = new Event('keydown', { cancelable: true });
      event.key = key;
      tabs[index].dispatchEvent(event);
      assert.equal(event.defaultPrevented, true);
      assert.equal(document.activeElement, tabs[next]);
      assert.equal(panels[next].hidden, false);
      assert.equal(tabs[next].getAttribute('aria-selected'), 'true');
    }
  }
});

test('all quickstart tab groups have independent selection', () => {
  const { groups } = fixture();
  assert.equal(groups[0].tabs[0].getAttribute('id'), 'bootstrap-go-tab');
  assert.equal(groups[1].tabs[0].getAttribute('id'), 'workspace-cli-tab');
  assert.equal(groups[2].tabs[0].getAttribute('id'), 'contents-shell-tab');
  assert.equal(groups[3].tabs[0].getAttribute('id'), 'define-shell-tab');
  groups[0].tabs[1].dispatchEvent(new Event('click'));
  assert.deepEqual(groups[0].panels.map(panel => panel.hidden), [true, false]);
  assert.deepEqual(groups[1].panels.map(panel => panel.hidden), [false, true]);
  groups[1].tabs[1].dispatchEvent(new Event('click'));
  assert.deepEqual(groups[0].panels.map(panel => panel.hidden), [true, false]);
  assert.deepEqual(groups[1].panels.map(panel => panel.hidden), [true, false]);
  assert.deepEqual(groups[2].panels.map(panel => panel.hidden), [false, true]);
  groups[2].tabs[1].dispatchEvent(new Event('click'));
  assert.deepEqual(groups[2].panels.map(panel => panel.hidden), [true, false]);
  assert.deepEqual(groups[3].panels.map(panel => panel.hidden), [false, true]);
  groups[3].tabs[1].dispatchEvent(new Event('click'));
  assert.deepEqual(groups[3].panels.map(panel => panel.hidden), [true, false]);
  assert.deepEqual(groups[1].panels.map(panel => panel.hidden), [true, false]);
});
