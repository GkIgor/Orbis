function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("div");
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
