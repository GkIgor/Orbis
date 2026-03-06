import { describe, it, expect, beforeEach, vi } from "vitest";
import { Router } from "./router.js";

// Mock Components
class RootComponent {
  constructor() {
    this.name = "Root";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = {};
  }
  destroy() {}
  render() {}
}
class QueryComponent {
  constructor() {
    this.name = "QueryReader";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = {};
  }
  destroy() {}
  render() {}
}

const mockContainer = {
  createComponent: (ClassDef) => new ClassDef(),
};

describe("Orbis Router Extensions (Hash, Query, Events)", () => {
  let router;

  beforeEach(() => {
    global.window = {
      location: { pathname: "/", search: "", hash: "" },
      addEventListener: vi.fn(),
      history: { pushState: vi.fn() },
    };
    global.document = {
      querySelector: vi
        .fn()
        .mockReturnValue({ appendChild: vi.fn(), innerHTML: "" }),
    };
  });

  describe("Query Parameter Parsing", () => {
    it("extracts explicit queries without reactive properties", () => {
      router = new Router(mockContainer, { mode: "history" });
      router.registerRoutes([
        { path: "/dashboard", component: QueryComponent },
      ]);

      router._diffAndRender = vi.fn();

      const result = router.resolve("/dashboard?page=2&filter=active");

      expect(result.path).toBe("/dashboard"); // Stripped path explicitly for O(depth) matcher
      expect(result.query).toEqual({ page: "2", filter: "active" });
    });

    it("serializes Object queries strictly in navigate()", () => {
      router = new Router(mockContainer, { mode: "history" });
      router.resolve = vi.fn();

      router.navigate("/dashboard", { query: { sort: "desc", limit: 10 } });

      expect(global.window.history.pushState).toHaveBeenCalledWith(
        {},
        "",
        "/dashboard?sort=desc&limit=10",
      );
      expect(router.resolve).toHaveBeenCalledWith(
        "/dashboard?sort=desc&limit=10",
      );
    });
  });

  describe("Hash Routing Mode Support", () => {
    it("strips physical hash and acts transparently backward-compatible", () => {
      router = new Router(mockContainer, { mode: "hash" });
      router.registerRoutes([
        { path: "/dashboard/users", component: RootComponent },
      ]);

      router._diffAndRender = vi.fn();

      // Simulated native hash change
      const parsed = router._parseUrl("#/dashboard/users?page=1");

      expect(parsed.path).toBe("/dashboard/users"); // Explicit path mapping retained perfectly
      expect(parsed.query).toEqual({ page: "1" });
    });

    it("mutates physical location.hash rather than pushState under mode=hash", () => {
      router = new Router(mockContainer, { mode: "hash" });
      router.resolve = vi.fn(); // Mock away downstream effects

      router.navigate("/settings", { query: { dark: true } });

      expect(global.window.location.hash).toBe("/settings?dark=true");
      expect(global.window.history.pushState).not.toHaveBeenCalled();
    });
  });

  describe("Synchronous Deterministic Events", () => {
    it("emits sequentially and provides completely frozen read-only immutable payloads", () => {
      router = new Router(mockContainer);
      router.registerRoutes([{ path: "/dashboard", component: RootComponent }]);

      const sequence = [];
      const beforePayloads = [];

      router.on("beforeNavigation", (ctx) => {
        sequence.push("beforeNavigation");
        beforePayloads.push(ctx);
        // Attempt explicitly forbidden side effect -> object is strictly frozen natively
        expect(() => {
          ctx.path = "/hijack";
        }).toThrow();
      });
      router.on("routeMatched", () => sequence.push("routeMatched"));
      router.on("componentStackResolved", () =>
        sequence.push("componentStackResolved"),
      );
      router.on("afterNavigation", () => sequence.push("afterNavigation"));

      router.resolve("/dashboard?page=42");

      expect(sequence).toEqual([
        "beforeNavigation",
        "routeMatched",
        "componentStackResolved",
        "afterNavigation",
      ]);

      expect(beforePayloads[0].path).toBe("/dashboard");
      expect(beforePayloads[0].query).toEqual({ page: "42" });
    });
  });
});
