import { describe, it, expect, beforeEach, vi } from "vitest";
import { Router } from "./router.js";

// Mocks
class ProfileComponent {
  constructor() {
    this.name = "Profile";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = {};
    this.routeData = {};
  }
  destroy() {}
  render() {}
}

const mockContainer = {
  createComponent: (ClassDef) => new ClassDef(),
};

describe("Orbis Router Phase 9.3 (Resolvers, Scroll, DevTools)", () => {
  let router;

  beforeEach(() => {
    global.window = {
      location: { pathname: "/", search: "", hash: "" },
      addEventListener: vi.fn(),
      history: { pushState: vi.fn(), replaceState: vi.fn() },
      scrollTo: vi.fn(),
      __ORBIS_DEVTOOLS__: vi.fn(),
    };
    global.document = {
      querySelector: vi
        .fn()
        .mockReturnValue({ appendChild: vi.fn(), innerHTML: "" }),
    };
  });

  describe("Route Data Resolvers", () => {
    it("strictly `await`s asynchronous fetch tasks gathering deterministic keys natively into `.routeData`", async () => {
      router = new Router(mockContainer);
      const mockFetch = vi.fn().mockResolvedValue({ id: 99, name: "Igor" });

      router.registerRoutes([
        {
          path: "/profile/:id",
          component: ProfileComponent,
          resolve: { user: mockFetch },
        },
      ]);

      router._diffAndRender = vi.fn();

      const result = await router.resolve("/profile/99");

      expect(mockFetch).toHaveBeenCalledTimes(1);

      // Proves explicit context delivery securely bound
      const ctx = mockFetch.mock.calls[0][0];
      expect(ctx.params.id).toBe("99");
      expect(Object.isFrozen(ctx)).toBe(true);

      // Merged payload identically maps natively
      expect(result.routeData.user).toEqual({ id: 99, name: "Igor" });
    });

    it("halts navigation completely firing `navigationCancelled` implicitly when an endpoint natively throws", async () => {
      router = new Router(mockContainer);
      const crashFetch = vi.fn().mockRejectedValue(new Error("Network Error"));

      router.registerRoutes([
        {
          path: "/crash",
          component: ProfileComponent,
          resolve: { user: crashFetch },
        },
      ]);

      const sequence = [];
      router.on("navigationCancelled", () => sequence.push("cancelled"));
      router.on("afterNavigation", () => sequence.push("done"));

      router._diffAndRender = vi.fn();

      const result = await router.resolve("/crash");

      expect(router._diffAndRender).not.toHaveBeenCalled();
      expect(sequence).toEqual(["cancelled"]);
      expect(result).toBeUndefined(); // Layout unmutated seamlessly
    });

    it("merges hierarchical data strictly replacing overlapping keys logically deterministically", async () => {
      router = new Router(mockContainer);
      router.registerRoutes([
        {
          path: "/api",
          component: ProfileComponent,
          resolve: { base: () => "API_URL" },
          children: [
            {
              path: "users",
              component: class {},
              resolve: { userList: () => ["User1", "User2"] },
            },
          ],
        },
      ]);

      router._diffAndRender = vi.fn();

      const result = await router.resolve("/api/users");
      expect(result.routeData).toEqual({
        base: "API_URL",
        userList: ["User1", "User2"],
      });
    });
  });

  describe("DevTools Emitter Isolation", () => {
    it("broadcasts immutable diagnostic events unconditionally safely throughout `resolve` orchestrations", async () => {
      router = new Router(mockContainer);
      router.registerRoutes([{ path: "/fast", component: class {} }]);
      router._diffAndRender = vi.fn();

      await router.resolve("/fast");

      expect(global.window.__ORBIS_DEVTOOLS__).toHaveBeenCalledWith(
        "routeResolved",
        expect.anything(),
      );
      expect(global.window.__ORBIS_DEVTOOLS__).toHaveBeenCalledWith(
        "guardsExecuted",
        expect.anything(),
      );
      expect(global.window.__ORBIS_DEVTOOLS__).toHaveBeenCalledWith(
        "componentRendered",
        expect.anything(),
      );

      const payload = global.window.__ORBIS_DEVTOOLS__.mock.calls[0][1];
      expect(Object.isFrozen(payload)).toBe(true);
    });
  });

  describe("Explicit Scroll Restoration", () => {
    it("restores `(0, 0)` explicitly post-render unconditionally when `scrollRestoration: 'top'` defined", async () => {
      router = new Router(mockContainer, { scrollRestoration: "top" });
      router.registerRoutes([{ path: "/home", component: class {} }]);
      router._diffAndRender = vi.fn();

      await router.resolve("/home");

      expect(global.window.scrollTo).toHaveBeenCalledWith(0, 0);
    });

    it("does absolutely nothing magically implicitly when config defaults to 'none'", async () => {
      router = new Router(mockContainer); // Default is "none"
      router.registerRoutes([{ path: "/home", component: class {} }]);
      router._diffAndRender = vi.fn();

      await router.resolve("/home");

      expect(global.window.scrollTo).not.toHaveBeenCalled();
    });
  });
});
