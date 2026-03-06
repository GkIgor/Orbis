import { JSDOM } from "jsdom";
import assert from "assert/strict";
import fs from "fs";

// Setup JSDOM environment so the runtime files have a DOM to operate on.
const dom = new JSDOM(`<!DOCTYPE html><html><body></body></html>`, {
  url: "http://localhost/",
});
global.window = dom.window;
global.document = window.document;
global.HTMLElement = window.HTMLElement;
global.customElements = window.customElements;

// Now we can import the runtime
import { OrbisComponent, registerComponent, mount } from "./index.js";

// --- Test 1: Component Registration ---
console.log("Running Test 1: Component Registration");
class TestElement extends OrbisComponent {
  _compileRender(ctx, root) {
    const text = document.createTextNode("Hello Orbis");
    root.appendChild(text);
  }
}
registerComponent("test-element", TestElement);
assert.ok(
  customElements.get("test-element"),
  "test-element should be registered",
);

// --- Test 2: Component Mounting ---
console.log("Running Test 2: Component Mounting");
const target = document.createElement("div");
const instance = mount("test-element", target);
assert.strictEqual(
  target.children[0],
  instance,
  "Component should be appended to target",
);
assert.strictEqual(
  instance.shadowRoot.innerHTML,
  "Hello Orbis",
  "render() should have executed synchronously during mount",
);

// --- Test 3: Render Lifecycle Execution Order ---
console.log("Running Test 3: Render Lifecycle Order");
const events = [];
class LifecycleTest extends OrbisComponent {
  beforeRender() {
    events.push("beforeRender");
  }
  _compileRender(ctx, root) {
    events.push("compileRender");
  }
  afterRender() {
    events.push("afterRender");
  }
}
registerComponent("lifecycle-test", LifecycleTest);
const lcInstance = document.createElement("lifecycle-test");
lcInstance.render();
assert.deepStrictEqual(
  events,
  ["beforeRender", "compileRender", "afterRender"],
  "Lifecycle events must execute in explicit order: beforeRender -> render DOM -> afterRender",
);

// --- Test 4: Shadow DOM Creation ---
console.log("Running Test 4: Shadow DOM Creation");
class ShadowTest extends OrbisComponent {}
registerComponent("shadow-test", ShadowTest);
const shadowInstance = document.createElement("shadow-test");
assert.ok(shadowInstance.shadowRoot, "ShadowRoot must be created");
assert.strictEqual(
  shadowInstance.shadowRoot.mode,
  "open",
  "ShadowRoot must be open",
);

// --- Test 5: DOM Clearing Before Render ---
console.log("Running Test 5: DOM Clearing Before Render");
let renderCount = 0;
class ClearingTest extends OrbisComponent {
  _compileRender(ctx, root) {
    renderCount++;
    const p = document.createElement("p");
    p.textContent = `Render ${renderCount}`;
    root.appendChild(p);
  }
}
registerComponent("clearing-test", ClearingTest);
const clearingInstance = document.createElement("clearing-test");
clearingInstance.render(); // Render 1
assert.strictEqual(clearingInstance.shadowRoot.innerHTML, "<p>Render 1</p>");
clearingInstance.render(); // Render 2
assert.strictEqual(
  clearingInstance.shadowRoot.innerHTML,
  "<p>Render 2</p>",
  "ShadowRoot should be cleared before second render",
);

// --- Test 6: Event Listener Attachment Compatibility ---
console.log("Running Test 6: Event Listener Compatibility");
let clicked = false;
class EventTest extends OrbisComponent {
  onClick() {
    clicked = true;
  }
  _compileRender(ctx, root) {
    const btn = document.createElement("button");
    btn.addEventListener("click", () => ctx.onClick());
    root.appendChild(btn);
  }
}
registerComponent("event-test", EventTest);
const eventInstance = document.createElement("event-test");
eventInstance.render();
const btn = eventInstance.shadowRoot.querySelector("button");
btn.click(); // Trigger click event
assert.strictEqual(
  clicked,
  true,
  "Bound event listener should trigger context method",
);

console.log(
  "\n✅ All 6 Orbis runtime tests passed successfully! Determinism verified.",
);
