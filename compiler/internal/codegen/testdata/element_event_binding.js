function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("button");
  _el0.setAttribute("class", "btn");
  _el0.addEventListener("click", function() { ctx.toggle(); });
  const _el1 = document.createTextNode("Toggle");
  _el0.appendChild(_el1);
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
