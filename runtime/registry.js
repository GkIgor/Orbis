/**
 * registerComponent wraps the native customElements API to ensure
 * unified registration of Orbis components.
 *
 * @param {string} name - Data tag name, must contain a hyphen (e.g., 'app-root')
 * @param {typeof HTMLElement} componentClass - The component class extending OrbisComponent
 */
export function registerComponent(name, componentClass) {
  if (!name.includes("-")) {
    throw new Error(
      `[Orbis] Component name "${name}" must contain a hyphen to be a valid Custom Element.`,
    );
  }

  if (customElements.get(name)) {
    console.warn(
      `[Orbis] Component "${name}" is already registered. Skipping.`,
    );
    return;
  }

  customElements.define(name, componentClass);
}
