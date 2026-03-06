/**
 * The deterministic Dependency Injection Container for Orbis.
 * Designed strictly according to RFC 0002.
 *
 * Rules:
 * - Deterministic, synchronous resolution
 * - No reflection or decorator parsing at runtime
 * - Stateful and Stateless providers are cached as strict Singletons
 * - Components are instantiated per mount
 */
export class Container {
  constructor() {
    // Map<ProviderClass, ProviderClass>
    this.providers = new Map();

    // Map<ProviderClass, Instance> - cache for Stateful and Stateless singletons
    this.instances = new Map();
  }

  /**
   * Registers a provider class in the container.
   * Expected to be called during application bootstrap by compiled generation.
   * @param {class} providerClass
   */
  register(providerClass) {
    if (!providerClass) {
      throw new Error("[Orbis DI] Cannot register an undefined provider.");
    }

    if (this.providers.has(providerClass)) {
      console.warn(
        `[Orbis DI] Provider ${providerClass.name} is already registered. Skipping.`,
      );
      return;
    }

    this.providers.set(providerClass, providerClass);
  }

  /**
   * Resolves a dependency token (Class) into a Singleton instance.
   * Implements implicit lazy-instantiation on first resolve.
   * @param {class} tokenClass
   * @returns {object} The singleton instance
   */
  resolve(tokenClass) {
    // 1. Check if we already have a cached instance (Singleton rule)
    if (this.instances.has(tokenClass)) {
      return this.instances.get(tokenClass);
    }

    // 2. Validate token is registered
    if (!this.providers.has(tokenClass)) {
      throw new Error(
        `[Orbis DI] Token ${tokenClass.name} not registered. Did you forget to call container.register()?`,
      );
    }

    // 3. Instantiate synchronously
    // Note: Orbis states do not have their own injected dependencies by design (States are terminal nodes).
    const instance = new tokenClass();

    // 4. Cache and return
    this.instances.set(tokenClass, instance);
    return instance;
  }

  /**
   * Factory function to create an Orbis component with constructor-injected dependencies.
   *
   * @param {class} ComponentClass - The component to document.createElement
   * @param {Array<class>} dependencies - Ordered list of provider tokens required by the constructor
   * @returns {HTMLElement} The bootstrapped component instance
   */
  createComponent(ComponentClass, dependencies = []) {
    // Technically document.createElement requires the registered string tag, not the class directly.
    // However, Orbis compilation manages customElement.define(tag, class).
    // We look up the registered tag name via private standard or constructor name wrapping.

    // Because customElements native API strictly requires elements to be created via
    // document.createElement or new Class() *after* define(), we return a constructed instance.

    // Resolve all dependencies explicitly and synchronously
    const resolvedArgs = [];
    for (let i = 0; i < dependencies.length; i++) {
      resolvedArgs.push(this.resolve(dependencies[i]));
    }

    // Instantiate
    return new ComponentClass(...resolvedArgs);
  }
}
