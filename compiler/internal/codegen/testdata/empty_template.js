function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  if (ctx.afterRender) ctx.afterRender();
}
