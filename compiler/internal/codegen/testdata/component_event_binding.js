function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el1 = document.createElement("counteritem");
  _el1.addEventListener("click", function() { ctx.select(i); });
  const _el0 = new CounterItem();
  const _el0Shadow = _el1.attachShadow({ mode: "open" });
  _el0.render(_el0, _el0Shadow);
  root.appendChild(_el1);
  if (ctx.afterRender) ctx.afterRender();
}
