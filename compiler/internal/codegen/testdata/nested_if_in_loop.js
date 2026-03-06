function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  for (let i = 0; i < ctx.items.length; i++) {
    const item = ctx.items[i];
    if (item.active) {
      const _el0 = document.createElement("span");
      const _el1 = document.createTextNode(String(item.name));
      _el0.appendChild(_el1);
      root.appendChild(_el0);
    }
  }
  if (ctx.afterRender) ctx.afterRender();
}
