function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("p");
  const _el1 = document.createTextNode("Hello World");
  _el0.appendChild(_el1);
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
