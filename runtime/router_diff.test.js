import { describe, it, expect, beforeEach, vi } from "vitest";
import { Router } from "./router.js";

class LayoutComponent {
  constructor() {
    this.name = "Layout";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = { params: {} };
  }
  destroy() {
    this.destroyed = true;
  }
  render() {
    this.rendered = true;
  }
}

class ParentComponent {
  constructor() {
    this.name = "Parent";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = { params: {} };
  }
  destroy() {
    this.destroyed = true;
  }
  render() {
    this.rendered = true;
  }
}

class ChildComponentA {
  constructor() {
    this.name = "ChildA";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = { params: {} };
  }
  destroy() {
    this.destroyed = true;
  }
  render() {
    this.rendered = true;
  }
}

class ChildComponentB {
  constructor() {
    this.name = "ChildB";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = { params: {} };
  }
  destroy() {
    this.destroyed = true;
  }
  render() {
    this.rendered = true;
  }
}

const mockContainer = {
  createComponent: (ClassDef) => new ClassDef(),
};

describe("Orbis Router Component Diffing", () => {
  let router;
  let rootApp;

  beforeEach(() => {
    rootApp = { appendChild: vi.fn(), innerHTML: "" };
    global.document = { querySelector: vi.fn().mockReturnValue(rootApp) };
    global.window = {
      addEventListener: vi.fn(),
      history: { pushState: vi.fn() },
      location: { pathname: "/" },
    };

    router = new Router(mockContainer);
  });

  it("mounts sequentially searching for router-slots", () => {
    const parentSlot = { appendChild: vi.fn() };
    const layoutSlot = { appendChild: vi.fn() };

    // Setup nested querySelectors
    const instances = [];
    mockContainer.createComponent = (ClassDef) => {
      const inst = new ClassDef();
      if (inst.name === "Layout")
        inst.shadowRoot.querySelector.mockReturnValue(layoutSlot);
      if (inst.name === "Parent")
        inst.shadowRoot.querySelector.mockReturnValue(parentSlot);
      instances.push(inst);
      return inst;
    };

    router._diffAndRender({
      componentStack: [LayoutComponent, ParentComponent, ChildComponentA],
      params: { id: "42" }, // Pass along explicit parameters
    });

    expect(instances.length).toBe(3);

    // Layout mounted to root
    expect(rootApp.appendChild).toHaveBeenCalledWith(instances[0]);
    // Parent mounted to Layout's slot
    expect(layoutSlot.appendChild).toHaveBeenCalledWith(instances[1]);
    // ChildA mounted to Parent's slot
    expect(parentSlot.appendChild).toHaveBeenCalledWith(instances[2]);

    // All should be rendered and implicitly injected with state context
    expect(instances[0].rendered).toBe(true);
    expect(instances[2].rendered).toBe(true);
    // Explicit parameters securely loaded
    expect(instances[2].route.params.id).toBe("42");
  });

  it("retains matching stack nodes and ONLY explicitly rebuilds diverging tails without virtual DOM", () => {
    const layoutInst = new LayoutComponent();
    const parentInst = new ParentComponent();
    const childAInst = new ChildComponentA();
    childAInst.destroy = vi.fn();
    layoutInst.destroy = vi.fn();

    // Fabricate an existing deterministic browser DOM
    router.currentStack = [layoutInst, parentInst, childAInst];

    // Navigate to ChildB
    const parentSlot = { appendChild: vi.fn() };
    parentInst.shadowRoot.querySelector.mockReturnValue(parentSlot);

    let instantiatedChildB = null;
    mockContainer.createComponent = (ClassDef) => {
      instantiatedChildB = new ClassDef();
      return instantiatedChildB;
    };

    router._diffAndRender({
      componentStack: [LayoutComponent, ParentComponent, ChildComponentB],
      params: {},
    });

    // Validations:

    // 1. Never destroys or rebuilds matching components
    expect(layoutInst.destroy).not.toHaveBeenCalled();
    expect(childAInst.destroy).toHaveBeenCalled(); // Only the diverging node explicitly teardown

    // 2. Extracts router-slot physically native off the pre-existing DOM instance
    expect(parentInst.shadowRoot.querySelector).toHaveBeenCalledWith(
      "router-slot",
    );

    // 3. Mounts directly to slot securely
    expect(parentSlot.appendChild).toHaveBeenCalledWith(instantiatedChildB);
    expect(instantiatedChildB.rendered).toBe(true);

    // 4. Physical array tracking trimmed
    expect(router.currentStack).toEqual([
      layoutInst,
      parentInst,
      instantiatedChildB,
    ]);
  });
});
