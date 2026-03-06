import { describe, it, expect, beforeEach, vi } from "vitest";
import { Router } from "./router.js";

// Mock Components
class DashboardComponent {
  constructor() {
    this.name = "Dashboard";
    this.shadowRoot = { querySelector: vi.fn() };
  }
  destroy() {}
  render() {}
}
class UsersComponent {
  constructor() {
    this.name = "Users";
    this.shadowRoot = { querySelector: vi.fn() };
  }
  destroy() {}
  render() {}
}
class UserDetailsComponent {
  constructor() {
    this.name = "UserDetails";
    this.shadowRoot = { querySelector: vi.fn() };
  }
  destroy() {}
  render() {}
}
class SettingsComponent {
  constructor() {
    this.name = "Settings";
    this.shadowRoot = { querySelector: vi.fn() };
  }
  destroy() {}
  render() {}
}

const mockContainer = {
  createComponent: (ClassDef) => new ClassDef(),
};

describe("Orbis Router Determinism", () => {
  let router;

  beforeEach(() => {
    // Stub out document / window environment for headless resolving tests
    global.window = {
      location: { pathname: "/" },
      addEventListener: vi.fn(),
      history: { pushState: vi.fn() },
    };
    global.document = {
      querySelector: vi
        .fn()
        .mockReturnValue({ appendChild: vi.fn(), innerHTML: "" }),
    };

    router = new Router(mockContainer);

    router.registerRoutes([
      {
        path: "/dashboard",
        component: DashboardComponent,
        children: [
          { path: "users", component: UsersComponent },
          { path: "users/:id", component: UserDetailsComponent },
          { path: "settings", component: SettingsComponent },
        ],
      },
    ]);
  });

  it("builds the route tree precisely", () => {
    // Validate depth and explicit parameters
    const root = router.rootNode;
    expect(root.staticChildren.has("dashboard")).toBe(true);

    const dashboard = root.staticChildren.get("dashboard");
    expect(dashboard.component).toBe(DashboardComponent);
    expect(dashboard.staticChildren.has("users")).toBe(true);
    expect(dashboard.staticChildren.has("settings")).toBe(true);

    const users = dashboard.staticChildren.get("users");
    expect(users.component).toBe(UsersComponent);
    expect(users.dynamicChild).not.toBeNull();
    expect(users.dynamicChild.isParam).toBe(true);
    expect(users.dynamicChild.paramName).toBe("id");
    expect(users.dynamicChild.component).toBe(UserDetailsComponent);
  });

  it("resolves exact static nested routes", () => {
    // Stub diffAndRender as we just want the pure resolve tree
    router._diffAndRender = vi.fn();

    const result = router.resolve("/dashboard/users");
    expect(result.componentStack).toEqual([DashboardComponent, UsersComponent]);
    expect(result.params).toEqual({});
  });

  it("extracts dynamic parameters explicitly", () => {
    router._diffAndRender = vi.fn();

    const result = router.resolve("/dashboard/users/42");
    expect(result.componentStack).toEqual([
      DashboardComponent,
      UserDetailsComponent,
    ]);
    expect(result.params).toEqual({ id: "42" });
  });

  it("prioritizes static segments over dynamic variables strictly", () => {
    router._diffAndRender = vi.fn();

    // We add a static route alongside the dynamic one
    router.registerRoutes([
      {
        path: "/dashboard/users/new",
        component: SettingsComponent, // Alias just to prove static hits
      },
    ]);

    const result = router.resolve("/dashboard/users/new");
    expect(result.componentStack).toEqual([
      DashboardComponent,
      SettingsComponent,
    ]);
    expect(result.params).toEqual({});
  });

  it("throws deterministic 404 when paths cannot be satisfied", () => {
    expect(() => {
      router.resolve("/dashboard/invalid");
    }).toThrow(/No route matched/);
  });
});
