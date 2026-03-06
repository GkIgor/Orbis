/**
 * mount attaches an Orbis root component to a target DOM node and triggers
 * its initial deterministic render cycle.
 *
 * @param {string} componentName - The registered tag name of the component
 * @param {HTMLElement} target - The DOM node where the component should be injected
 * @returns {HTMLElement} The created component instance
 */
export function mount(componentName, target) {
  if (!(target instanceof HTMLElement)) {
    throw new Error("[Orbis] Mount target must be an HTMLElement.");
  }

  // Attempt to register if it exists but hasn't been defined?
  // Usually registration is a separate manual step in Orbis.
  if (!customElements.get(componentName)) {
    throw new Error(`[Orbis] Component "${componentName}" is not registered.`);
  }

  // Create the component
  const instance = document.createElement(componentName);

  // Append to target
  target.appendChild(instance);

  // Trigger initial render (synchronously and explicitly)
  if (typeof instance.render === "function") {
    instance.render();
  } else {
    throw new Error(
      `[Orbis] Component "${componentName}" is not a valid OrbisComponent (missing render method).`,
    );
  }

  return instance;
}
