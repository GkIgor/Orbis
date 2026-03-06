/**
 * OrbisComponent is the base class for all compiled Orbis components.
 *
 * Responsibilities:
 * - Attach an open ShadowRoot
 * - Provide lifecycle hooks (beforeRender, afterRender)
 * - Provide a deterministic render() method
 *
 * Rules:
 * - No implicit rendering or reactivity
 * - No template string parsing
 * - No dependency tracking
 * - Executed strictly according to RFC 0001
 */
export class OrbisComponent extends HTMLElement {
  constructor() {
    super();
    this.shadowRoot = this.attachShadow({ mode: "open" });
    this.state = {};
    this.container = null; // Injected by DI later
    this.route = { params: {} }; // Injected by Router later

    this.__children = [];
    this.__listeners = [];
    this.__mounted = false;
  }

  /**
   * Executed immediately before the DOM subtree is cleared and rebuilt.
   * Can be overridden by subclasses.
   */
  beforeRender() {}

  /**
   * Executed immediately after the DOM subtree has been complete rebuilt.
   * Can be overridden by subclasses.
   */
  afterRender() {}

  /**
   * The primary entry point for a render cycle.
   * This method:
   * 1. Calls beforeRender()
   * 2. Clears the shadowRoot
   * 3. Executes the compiled DOM instructions
   * 4. Calls afterRender()
   *
   * Subclasses MUST NOT override this method if they are expected to be compiled.
   * Subclasses MUST implement `_compileRender(ctx, root)` which contains the compiled DOM code.
   */
  render() {
    // Step 1: beforeRender hook
    this.beforeRender();

    // Step 2: Clear previous subtree
    this.shadowRoot.innerHTML = "";

    // Step 3 & 4: Execute compiled DOM construction and attachment
    if (typeof this._compileRender === "function") {
      this._compileRender(this, this.shadowRoot);
    } else {
      console.warn(
        `[Orbis] Component <${this.localName}> requires a _compileRender method.`,
      );
    }

    // Step 5: afterRender hook
    this.afterRender();
  }

  /**
   * Determines partial state update explicitly.
   * Merges into this.state.
   * Does NOT automatically trigger a rerender.
   * @param {Object} partialState
   */
  setState(partialState) {
    if (partialState && typeof partialState === "object") {
      Object.assign(this.state, partialState);
    }
  }

  /**
   * Registers a child component instance to this parent component.
   * Ensures the child is properly destroyed when the parent is re-rendered or destroyed.
   * @param {OrbisComponent} child
   */
  registerChild(child) {
    this.__children.push(child);
  }

  /**
   * Reursively destroys all tracked child components and clears the registry.
   */
  destroyChildren() {
    for (const child of this.__children) {
      if (typeof child.destroy === "function") {
        child.destroy();
      }
    }
    this.__children = [];
  }

  /**
   * Safely attaches an event listener and tracks it for deterministic cleanup.
   * Compiled templates MUST use this instead of element.addEventListener.
   * @param {HTMLElement} element
   * @param {string} type
   * @param {Function} handler
   */
  addListener(element, type, handler) {
    element.addEventListener(type, handler);
    this.__listeners.push({ element, type, handler });
  }

  /**
   * Removes all tracked event listeners to prevent closure garbage collection leaks.
   */
  removeEventListeners() {
    for (const l of this.__listeners) {
      l.element.removeEventListener(l.type, l.handler);
    }
    this.__listeners = [];
  }

  /**
   * Lifecycle hook executed before destruction begins.
   */
  beforeDestroy() {}

  /**
   * Lifecycle hook executed after destruction completes.
   */
  afterDestroy() {}

  /**
   * The formal destruction pipeline.
   * Severes logical references, DOM references, and closures to prevent leaks.
   */
  destroy() {
    this.beforeDestroy();
    this.destroyChildren();
    this.removeEventListeners();

    if (this.shadowRoot) {
      this.shadowRoot.innerHTML = "";
    }

    this.afterDestroy();
  }

  /**
   * Explicit manual trigger for a new render cycle.
   * Following RFC 0001, this destroys the old tracked children and rebuilds the DOM.
   */
  rerender() {
    // Destroy existing children and listeners before clearing DOM to prevent dangling closures
    this.destroyChildren();
    this.removeEventListeners();

    // The base render() method already calls shadowRoot.innerHTML = ""
    this.render();
  }
}
