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
  for (const match of html.matchAll(/<(?:button|section)\s+([^>]*role="tab(?:panel)?"[^>]*)>/g)) new Element(match[1]);
  const tabs = [...elements.values()].filter(element => element.getAttribute('role') === 'tab');
  assert.equal(tabs.length, 2);
  document.querySelectorAll = () => [{ querySelectorAll: () => tabs }];
  runInNewContext(script, { document });
  return { document, tabs, panels: tabs.map(tab => elements.get(tab.getAttribute('aria-controls'))) };
}

test('Go is selected initially; clicks switch visible panels and selected indicators in both directions', () => {
  const { tabs, panels } = fixture();
  assert.equal(tabs[0].getAttribute('aria-selected'), 'true');
  assert.equal(panels[0].hidden, false);
  assert.equal(panels[1].hidden, true);
  tabs[1].dispatchEvent(new Event('click'));
  assert.deepEqual(panels.map(panel => panel.hidden), [true, false]);
  assert.deepEqual(tabs.map(tab => tab.getAttribute('aria-selected')), ['false', 'true']);
  assert.deepEqual(tabs.map(tab => tab.tabIndex), [-1, 0]);
  tabs[0].dispatchEvent(new Event('click'));
  assert.deepEqual(panels.map(panel => panel.hidden), [false, true]);
});

test('keyboard navigation switches panels and focus without header controls', () => {
  const { document, tabs, panels } = fixture();
  for (const [index, key, next] of [[0, 'ArrowRight', 1], [1, 'Home', 0], [0, 'End', 1], [1, 'ArrowLeft', 0]]) {
    const event = new Event('keydown', { cancelable: true });
    event.key = key;
    tabs[index].dispatchEvent(event);
    assert.equal(event.defaultPrevented, true);
    assert.equal(document.activeElement, tabs[next]);
    assert.equal(panels[next].hidden, false);
    assert.equal(tabs[next].getAttribute('aria-selected'), 'true');
  }
});
