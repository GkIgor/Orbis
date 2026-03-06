import { describe, it, expect, beforeEach, vi } from "vitest";
import { Router } from "./router.js";

// Mocks
class DashboardComponent {
  constructor() {
    this.name = "Dashboard";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = {};
  }
  destroy() {}
  render() {}
}
class SecuredUsersComponent {
  constructor() {
    this.name = "SecuredUsers";
    this.shadowRoot = { querySelector: vi.fn() };
    this.route = {};
  }
  destroy() {}
  render() {}
}

const mockContainer = {
  createComponent: (ClassDef) => new ClassDef(),
};

describe("Orbis Router Phase 9.2 (Guards & Lazy Loading)", () => {
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

    router = new Router(mockContainer);
  });

  describe("Route Guards", () => {
    it("permits synchronous access entirely without DOM mutations natively when returning true", async () => {
      const guardSpy = vi.fn().mockReturnValue(true);
      router.registerRoutes([
        { path: "/users", component: SecuredUsersComponent, guard: guardSpy },
      ]);

      router._diffAndRender = vi.fn();

      const result = await router.resolve("/users");

      expect(guardSpy).toHaveBeenCalledTimes(1);
      // Ensure context is perfectly frozen and complete
      const ctx = guardSpy.mock.calls[0][0];
      expect(Object.isFrozen(ctx)).toBe(true);
      expect(ctx.path).toEqual("/users");
      expect(router._diffAndRender).toHaveBeenCalledTimes(1);
    });

    it("immediately strictly aborts the DOM mutation pipeline statically when a guard returns false", async () => {
      const preGuardSpy = vi.fn().mockReturnValue(true);
      const blockingGuardSpy = vi.fn().mockReturnValue(false);
      const postGuardSpy = vi.fn().mockReturnValue(true);

      router.registerRoutes([
        {
          path: "/dashboard",
          component: DashboardComponent,
          guard: preGuardSpy,
          children: [
            {
              path: "users",
              component: SecuredUsersComponent,
              guard: blockingGuardSpy,
              children: [
                { path: "secret", component: class {}, guard: postGuardSpy },
              ],
            },
          ],
        },
      ]);

      router._diffAndRender = vi.fn();

      const result = await router.resolve("/dashboard/users/secret");

      // Proves sequential guard inheritance order executing cleanly:
      expect(preGuardSpy).toHaveBeenCalledTimes(1);
      expect(blockingGuardSpy).toHaveBeenCalledTimes(1);
      // Proves hard stop upon failure without mutating states
      expect(postGuardSpy).not.toHaveBeenCalled();
      expect(router._diffAndRender).not.toHaveBeenCalled();

      // Result is implicitly undefined due to early abort
      expect(result).toBeUndefined();
    });
  });

  describe("Lazy Component Loading", () => {
    it("throws upfront gracefully if statically invalid node signature parses inherently", () => {
      expect(() => {
        router.registerRoutes([
          {
            path: "/clash",
            component: DashboardComponent,
            loadComponent: () => {},
          },
        ]);
      }).toThrow(/cannot define both/);
    });

    it("dynamically halts execution to `await` Promise hooks internally resolving components deterministically", async () => {
      let resolveLoader;
      const loaderPromise = new Promise((r) => (resolveLoader = r));
      const loaderFn = vi.fn().mockReturnValue(loaderPromise);

      router.registerRoutes([{ path: "/async", loadComponent: loaderFn }]);

      router._diffAndRender = vi.fn();

      // We initiate navigation
      const resolveContext = router.resolve("/async");

      expect(loaderFn).toHaveBeenCalledTimes(1);

      // Router hangs sequentially before DOM hooks
      expect(router._diffAndRender).not.toHaveBeenCalled();

      // Loader finishes resolving simulating an imported ESM file chunk
      resolveLoader({ default: SecuredUsersComponent });

      const result = await resolveContext;

      expect(router._diffAndRender).toHaveBeenCalledTimes(1);
      expect(result.componentStack).toEqual([SecuredUsersComponent]);
    });

    it("caches explicitly resolved components sequentially stopping network redundance without reactive observation hooks", async () => {
      const loaderFn = vi.fn().mockResolvedValue(SecuredUsersComponent); // Simulate module resolution primitive
      router.registerRoutes([{ path: "/fast", loadComponent: loaderFn }]);

      await router.resolve("/fast");

      expect(loaderFn).toHaveBeenCalledTimes(1);

      await router.resolve("/fast");

      // Caching ensures secondary visits don't re-trigger async module loading
      expect(loaderFn).toHaveBeenCalledTimes(1);
    });

    it("fires un-awaited background `prefetch` cache requests when inserted cleanly", () => {
      const loaderFn = vi.fn().mockResolvedValue(DashboardComponent);

      router.registerRoutes([
        { path: "/warm", loadComponent: loaderFn, prefetch: true },
      ]);

      // Even without `resolve()` physically navigating to "/warm", the loader spins background implicitly ensuring rapid responses
      expect(loaderFn).toHaveBeenCalledTimes(1);
    });
  });
});
