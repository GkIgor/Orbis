function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("div");
  _el0.setAttribute("class", "container p-6");
  _el0.setAttribute("id", "main");
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
