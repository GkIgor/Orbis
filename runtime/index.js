export { OrbisComponent } from "./component.js";
export { Registry, registerComponent } from "./registry.js";
export { Container, registerProvider, inject } from "./container.js";
export { mount } from "./mount.js";
export { Router } from "./router.js";

import { Container } from "./container.js";
export const container = new Container();
