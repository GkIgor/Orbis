function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("h1");
  const _el1 = document.createTextNode(String(ctx.config.title));
  _el0.appendChild(_el1);
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
