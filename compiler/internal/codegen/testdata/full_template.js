function render(ctx, root) {
  if (ctx.beforeRender) ctx.beforeRender();
  root.innerHTML = "";
  const _el0 = document.createElement("div");
  _el0.setAttribute("class", "p-6 bg-gray-100 container");
  const _el1 = document.createElement("h1");
  _el1.setAttribute("class", "title");
  const _el2 = document.createTextNode(String(ctx.config.title));
  _el1.appendChild(_el2);
  _el0.appendChild(_el1);
  const _el3 = document.createElement("button");
  _el3.setAttribute("class", "btn");
  _el3.addEventListener("click", function() { ctx.toggle(); });
  const _el4 = document.createTextNode("Toggle List");
  _el3.appendChild(_el4);
  _el0.appendChild(_el3);
  if (ctx.app.visible) {
    const _el5 = document.createElement("div");
    _el5.setAttribute("class", "mt-4");
    for (let i = 0; i < ctx.app.items.length; i++) {
      const item = ctx.app.items[i];
      const _el7 = document.createElement("counteritem");
      _el7.addEventListener("click", function() { ctx.select(i); });
      const _el6 = new CounterItem();
      const _el6Shadow = _el7.attachShadow({ mode: "open" });
      _el6.render(_el6, _el6Shadow);
      _el5.appendChild(_el7);
    }
    _el0.appendChild(_el5);
  }
  root.appendChild(_el0);
  if (ctx.afterRender) ctx.afterRender();
}
