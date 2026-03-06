/**
 * Orbis Router (Deterministic Router)
 *
 * Provides explicit component stack resolution without implicit reactivity.
 */

class RouteNode {
  constructor(segment) {
    this.segment = segment;
    this.isParam = segment.startsWith(":");
    this.paramName = this.isParam ? segment.slice(1) : null;
    this.component = null;
    this.loadComponent = null; // Phase 9.2 Lazy Loader
    this.guard = null; // Phase 9.2 Explicit synchronous abort mechanism
    this.prefetch = false;
    this.isLayout = false; // Strictly differentiates layout parents vs leaf matches
    // Static children map for O(1) segment lookups
    this.staticChildren = new Map();
    // Dynamic parameter child (only one allowed per level)
    this.dynamicChild = null;
  }
}

export class Router {
  constructor(container, options = {}) {
    this.container = container;
    this.mode = options.mode || "history"; // "history" or "hash"
    this.rootNode = new RouteNode("");
    this.currentStack = [];
    this.currentPath = null;
    this._events = {
      beforeNavigation: [],
      routeMatched: [],
      componentStackResolved: [],
      afterNavigation: [],
    };

    // Bind explicit navigation events deterministically based on mode
    if (this.mode === "history") {
      window.addEventListener("popstate", () => {
        this.resolve(
          window.location.pathname +
            window.location.search +
            window.location.hash,
        );
      });
    } else if (this.mode === "hash") {
      window.addEventListener("hashchange", () => {
        const hashPath = window.location.hash.slice(1) || "/";
        this.resolve(hashPath);
      });
    }
  }

  /**
   * Registers a sync listener preventing async hooks.
   */
  on(eventName, callback) {
    if (this._events[eventName]) {
      this._events[eventName].push(callback);
    }
  }

  /**
   * Dispatches explicit lifecycle payloads read-only.
   */
  _emit(eventName, context) {
    if (this._events[eventName]) {
      for (const cb of this._events[eventName]) {
        cb(Object.freeze(Object.assign({}, context))); // Strict info-only payload isolating mutation
      }
    }
  }

  /**
   * Registers a deterministic route tree.
   * @param {Array} routes - Tree of route configs { path, component, children }
   */
  registerRoutes(routes) {
    for (const route of routes) {
      this._insertRoute(this.rootNode, route.path, route);
    }
  }

  _insertRoute(parentNode, currentPath, routeDef) {
    // Normalize path: remove leading/trailing slashes, split
    const segments = currentPath.split("/").filter((s) => s.length > 0);

    let currentNode = parentNode;

    for (const segment of segments) {
      const isParam = segment.startsWith(":");

      if (isParam) {
        if (!currentNode.dynamicChild) {
          currentNode.dynamicChild = new RouteNode(segment);
        } else if (currentNode.dynamicChild.segment !== segment) {
          throw new Error(
            `Orbis Router: Conflicting dynamic parameters at segment '${segment}'.`,
          );
        }
        currentNode = currentNode.dynamicChild;
      } else {
        if (!currentNode.staticChildren.has(segment)) {
          currentNode.staticChildren.set(segment, new RouteNode(segment));
        }
        currentNode = currentNode.staticChildren.get(segment);
      }
    }

    // Validate Configuration
    if (routeDef.component && routeDef.loadComponent) {
      throw new Error(
        `Orbis Router: Route '${currentPath}' cannot define both 'component' and 'loadComponent'.`,
      );
    }

    // Assign hooks explicitly
    currentNode.component = routeDef.component;
    currentNode.loadComponent = routeDef.loadComponent;
    currentNode.guard = routeDef.guard;
    currentNode.prefetch = routeDef.prefetch;

    // Dispatch background module prefetch immediately disconnected from resolving threads
    if (
      currentNode.prefetch &&
      typeof currentNode.loadComponent === "function"
    ) {
      // Intentionally un-awaited detached Promise resolution warming native module cache
      currentNode
        .loadComponent()
        .catch((err) => console.warn(`Orbis Router Prefetch Failed:`, err));
    }

    // Process children recursively
    if (routeDef.children && routeDef.children.length > 0) {
      currentNode.isLayout = true;
      for (const child of routeDef.children) {
        this._insertRoute(currentNode, child.path, child);
      }
    }
  }

  /**
   * Pushes a new history state and explicitly resolves the route stack natively handling options.
   * @param {string} path - The destination root URL
   * @param {Object} options - Navigation context options (e.g. { query: { param: 'val' }})
   */
  async navigate(path, options = {}) {
    let fullUrl = path;

    // Explicitly deterministic string serialization
    if (options.query && Object.keys(options.query).length > 0) {
      const qParams = new URLSearchParams();
      for (const [k, v] of Object.entries(options.query)) {
        qParams.append(k, String(v));
      }
      fullUrl += "?" + qParams.toString();
    }

    if (this.currentPath === fullUrl) return;

    if (this.mode === "history") {
      window.history.pushState({}, "", fullUrl);
    } else if (this.mode === "hash") {
      window.location.hash = fullUrl;
      return; // hashchange listener triggers resolve recursively
    }

    await this.resolve(fullUrl);
  }

  /**
   * Helper extracting hash/query deterministically
   * @param {string} url - Target href
   * @returns {Object} - Physical parsed primitives
   */
  _parseUrl(url) {
    let rawPath = url;
    let hash = "";

    if (this.mode === "hash" && rawPath.startsWith("#")) {
      rawPath = rawPath.slice(1);
    }

    const hashIndex = rawPath.indexOf("#");
    if (hashIndex > -1) {
      // In hash routing, the prefix "#/path" comes in directly so we don't break string offsets
      if (this.mode === "history") {
        hash = rawPath.slice(hashIndex + 1);
        rawPath = rawPath.slice(0, hashIndex);
      }
    }

    const query = {};
    const queryIndex = rawPath.indexOf("?");
    let path = rawPath;

    if (queryIndex > -1) {
      path = rawPath.slice(0, queryIndex);
      const qString = rawPath.slice(queryIndex + 1);
      const searchParams = new URLSearchParams(qString);
      for (const [k, v] of searchParams.entries()) {
        query[k] = v;
      }
    }

    return { path, query, hash };
  }

  /**
   * Resolves a URL path into a Stack of Components + extracted Params, Query, and Hash.
   * Dispatches deterministic lifecycle events strictly synchronous.
   * Runs guard pipeline and dynamic imports explicitly before modifying DOM.
   * @param {string} fullPath - URL pathname including query string
   */
  async resolve(fullPath) {
    const parsed = this._parseUrl(fullPath);
    const segments = parsed.path.split("/").filter((s) => s.length > 0);

    const result = {
      path: parsed.path,
      componentStack: [],
      params: {},
      query: parsed.query,
      hash: parsed.hash,
      from: this.currentPath,
    };

    // 1. SYNC EMIT: beforeNavigation
    this._emit("beforeNavigation", result);

    let currentNode = this.rootNode;
    const matchedNodes = [];

    // If path is purely root "/"
    if (segments.length === 0) {
      if (currentNode.component || currentNode.loadComponent) {
        matchedNodes.push(currentNode);
      }
    } else {
      // Resolve O(depth) explicit tree path
      for (let i = 0; i < segments.length; i++) {
        const segment = segments[i];
        const isLastSegment = i === segments.length - 1;

        if (currentNode.staticChildren.has(segment)) {
          currentNode = currentNode.staticChildren.get(segment);
        } else if (currentNode.dynamicChild) {
          currentNode = currentNode.dynamicChild;
          result.params[currentNode.paramName] = segment;
        } else {
          throw new Error(
            `Orbis Router: No route matched for path '${parsed.path}'.`,
          );
        }

        if (
          (currentNode.component || currentNode.loadComponent) &&
          (currentNode.isLayout || isLastSegment)
        ) {
          matchedNodes.push(currentNode);
        }
      }
    }

    // 2. RUN GUARDS (Synchronous Pipeline)
    for (const node of matchedNodes) {
      if (typeof node.guard === "function") {
        // Freeze context explicitly so guards cannot mutate layout logic
        const guardResult = node.guard(
          Object.freeze(Object.assign({}, result)),
        );
        if (guardResult === false) {
          return; // Abort resolution silently without altering the preexisting DOM explicitly
        }
      }
    }

    // 3. LOAD LAZY COMPONENTS (Async Pipeline)
    for (const node of matchedNodes) {
      if (!node.component && typeof node.loadComponent === "function") {
        const module = await node.loadComponent();
        // Cache explicitly to avoid network overhead natively returning visits
        node.component = module.default || module;
      }

      if (node.component) {
        result.componentStack.push(node.component);
      } else {
        throw new Error(
          `Orbis Router: Failed to resolve component dynamically for node '${node.segment}'.`,
        );
      }
    }

    // Pipeline completed sequentially
    this.currentPath = parsed.path;

    // 4. SYNC EMIT: routeMatched
    this._emit("routeMatched", result);

    // 5. DOM MUTATION
    this._diffAndRender(result);

    // 6. SYNC EMIT: afterNavigation
    this._emit("afterNavigation", result);

    return result;
  }

  /**
   * Diffs the incoming component stack against the current DOM.
   * Strictly tears down stale trees and explicitly mounts new branches.
   */
  _diffAndRender(routeResult) {
    const newStack = routeResult.componentStack;
    const params = routeResult.params;
    const query = routeResult.query;
    const hash = routeResult.hash;
    const path = routeResult.path;

    // 3. SYNC EMIT: componentStackResolved
    this._emit("componentStackResolved", routeResult);

    let divergenceIndex = 0;

    // 1. Identify divergence where classes no longer match
    while (
      divergenceIndex < this.currentStack.length &&
      divergenceIndex < newStack.length &&
      this.currentStack[divergenceIndex].constructor ===
        newStack[divergenceIndex]
    ) {
      divergenceIndex++;
    }

    // 2. Recursive destroy on stale instances
    for (let i = this.currentStack.length - 1; i >= divergenceIndex; i--) {
      const staleInstance = this.currentStack[i];
      // Note: destroy() implementation handled inherently by Phase 6.5 memory cleanup natively on OrbisComponent
      staleInstance.destroy();
    }

    // Trim current physical stack array cleanly
    this.currentStack = this.currentStack.slice(0, divergenceIndex);

    // 3. Construct and Mount novel nested components
    let parentNode = null;

    // Only resolve a parent physical container if we possess net new children to securely append.
    if (divergenceIndex < newStack.length) {
      // If we kept portions of the stack, the mounting parent is the slot inside the last preserved component
      if (divergenceIndex > 0) {
        const highestPreservedComponent =
          this.currentStack[divergenceIndex - 1];
        parentNode =
          highestPreservedComponent.shadowRoot.querySelector("router-slot");

        if (!parentNode) {
          throw new Error(
            `Orbis Router: Expected <router-slot> in ${highestPreservedComponent.constructor.name}`,
          );
        }
      } else {
        // We are rebuilding from root body/app
        parentNode = document.querySelector("#app");
        if (parentNode) {
          parentNode.innerHTML = ""; // Clear root
        }
      }
    }

    for (let i = divergenceIndex; i < newStack.length; i++) {
      const ComponentClass = newStack[i];

      // Inject DI container dependencies securely
      const instance = this.container.createComponent(ComponentClass);

      // Inject deterministic parameter payload thoroughly
      instance.route = {
        path,
        params: Object.assign({}, params),
        query: Object.assign({}, query),
        hash,
      };

      // Mount physically
      if (parentNode) {
        parentNode.appendChild(instance);
      }

      // Explicit initial deterministic render
      instance.render();

      this.currentStack.push(instance);

      // Subsequent nested children expect a <router-slot> natively in the layout we just rendered
      if (i < newStack.length - 1 && parentNode) {
        parentNode = instance.shadowRoot.querySelector("router-slot");
        if (!parentNode) {
          throw new Error(
            `Orbis Router: Component ${ComponentClass.name} must declare a <router-slot> to host child routes.`,
          );
        }
      }
    }

    return routeResult;
  }
}
