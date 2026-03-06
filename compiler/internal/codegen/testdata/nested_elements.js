function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("div");
  const _el1 = document.createElement("span");
  _el0.appendChild(_el1);
  const _el2 = document.createElement("p");
  _el0.appendChild(_el2);
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
