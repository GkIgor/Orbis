# Framework Proposal — Deterministic Declarative UI Engine

## 1. Overview

This document defines the technical proposal for a deterministic, declarative frontend framework designed primarily for AI-generated code and secondarily for developers who require strict control over rendering behavior.

The framework eliminates implicit reactivity, hidden lifecycle behavior, automatic change detection, and implicit rendering triggers. All state mutations, render cycles, and component lifecycles are explicitly controlled.

The system is composed of two main parts:

- **Compiler (Go)** — Performs Ahead-of-Time (AOT) compilation, template parsing, structural validation, linting, and code generation.
- **Runtime (JavaScript)** — Executes compiled components inside the browser using native DOM APIs and Shadow DOM.

The architecture prioritizes determinism, predictability, and explicit developer intent.

---

## 2. Motivation

Modern frontend frameworks rely heavily on implicit reactivity systems, automatic change detection, virtual DOM reconciliation, and hidden lifecycle behavior. While powerful, these abstractions introduce:

- Non-deterministic re-render behavior
- Implicit side effects
- Hidden performance costs
- Increased cognitive load
- Difficulty for AI systems to reason about runtime behavior

This framework is motivated by the need for:

- Predictable rendering
- Explicit control over UI updates
- Minimal runtime abstraction
- AI-friendly architectural constraints
- Reduced contextual entropy in generated code

The design enforces structure while removing implicit automation.

---

## 3. Design Principles

### 3.1 Determinism
Rendering only occurs when explicitly invoked.
No automatic re-rendering exists.

### 3.2 Declarative Structure
UI is declared using an HTML-based DSL compiled at build time.

### 3.3 No Implicit Side Effects
State mutation does not trigger rendering.
Observable emissions do not trigger rendering.
No hidden scheduling mechanisms exist.

### 3.4 Manual Control
Developers explicitly call:

- `render()`
- `rerender()`
- `fullrender()`

### 3.5 OOP-First Architecture
All components, states, and services are class-based.

### 3.6 Standardized Asynchronous Model
All asynchronous flows use RxJS.
No alternative reactive paradigms are permitted.

### 3.7 Minimal DSL
The template DSL includes only:

- `{{ interpolation }}`
- `<loop for="item in items" index="i">`
- `<if condition="expression">`
- Event binding via `(event)="handler()"`

No pipes, directives, or custom template syntax layers are supported.

---

## 4. Architecture

### 4.1 Compiler (Go)

Responsibilities:

- Parse HTML DSL
- Validate structural correctness
- Enforce template rules
- Perform linting
- Transform template into optimized DOM instructions
- Generate JavaScript render functions
- Detect invalid nesting or misuse at build time

Compilation Output:

- Static render functions using native DOM APIs
- No runtime template interpretation

The compiler must ensure zero template parsing occurs in the browser.

---

### 4.2 Runtime (JavaScript)

Responsibilities:

- Component instantiation
- ShadowRoot (mode: open) creation per component
- Dependency injection
- Render execution
- Lifecycle orchestration
- Component tree management
- Controlled destruction

The runtime does not:

- Perform diffing
- Perform automatic change detection
- Track dependency graphs
- Schedule batched updates

---

## 5. Component Model

Components are defined using decorators.

Example structure:

```
@Component({
  stateful: [CounterState],
  stateless: [ConfigState]
})
class MyComponent {
  constructor(private counter: CounterState) {}

  onInit() {}
  onDestroy() {}

  render() {}
}
```

Each component:

- Owns a ShadowRoot
- Has isolated DOM scope
- Is manually rendered
- Is destroyed explicitly

---

## 6. Rendering Model

### 6.1 render()

- Resolves interpolations
- Instantiates child components
- Attaches DOM
- Executes before/after render lifecycle hooks
- Does NOT re-trigger component lifecycle

### 6.2 rerender()

- Reconstructs component
- Re-executes full component lifecycle
- Rebuilds DOM subtree

### 6.3 fullrender()

- Recursively rerenders component and all descendants

Rendering is snapshot-based and state-driven.

---

## 7. State System

State is class-based and injected via decorators.

Two primary types:

### 7.1 @Stateless

- Immutable or externally initialized
- No internal mutation requirement

### 7.2 @Stateful

- Mutable
- Holds application data

Characteristics:

- State contains only data
- No automatic observers
- No implicit subscriptions
- No automatic rendering

State does not know who consumes it.
Components explicitly decide when to render.

---

## 8. Global Scope

States are globally injectable.

Any component may inject any declared state.
There are no scoped state boundaries enforced by the framework.

Architectural discipline is delegated to the developer.

---

## 9. Template Execution Model

Templates are compiled into imperative DOM creation instructions.

Example conceptual output:

```
function render(ctx, root) {
  const div = document.createElement("div")
  for (let i = 0; i < ctx.items.length; i++) {
    const p = document.createElement("p")
    p.textContent = ctx.items[i]
    div.appendChild(p)
  }
  root.appendChild(div)
}
```

No string-based rendering is permitted.
No runtime parsing is allowed.

---

## 10. Lifecycle Model

Component lifecycle:

- onInit()
- afterInit()
- onDestroy()

Render lifecycle:

- beforeRender()
- afterRender()

Component lifecycle is independent from render lifecycle.

---

## 11. Styling System

- Default SCSS support
- Native Tailwind compatibility
- Compile-time asset preprocessing
- Style encapsulation via Shadow DOM

---

## 12. Non-Goals (MVP)

The initial release will NOT include:

- Router
- HTTP client
- Form builder
- Devtools
- SSR
- Plugin system

These will be developed only after the core runtime and compiler are stable.

---

## 13. Target Use Cases

- AI-generated UI systems
- Deterministic enterprise dashboards
- Internal tooling
- Systems requiring strict rendering control
- Applications where predictability outweighs convenience

---

## 14. Conclusion

This framework is intentionally restrictive.

It trades automation for control.
It trades abstraction for clarity.
It trades convenience for determinism.

The primary objective is not to compete with existing mainstream frameworks but to provide an alternative model where rendering behavior is explicit, predictable, and structurally enforced at compile time.


---

# 15. Pilot Reference Implementation (Compiler Test Target)

This section defines a minimal but complete pilot system. It will serve as:

- The first real compilation target
- A structural validation benchmark
- A DSL feature coverage example
- A runtime behavior reference

The system is intentionally simple but exercises:

- Component nesting
- Global state injection
- @Stateful and @Stateless
- Manual render control
- <loop>
- <if>
- Event binding
- Interpolation
- Shadow DOM isolation
- SCSS + Tailwind usage

---

# 16. Repository Tree

The repository must follow this structure:

```
root/
│
├── packages/
│   ├── compiler/              # Go compiler
│   ├── runtime/               # Core JS runtime
│   └── cli/                   # CLI wrapper
│
├── examples/
│   └── pilot-app/
│       ├── src/
│       │   ├── main.ts
│       │   ├── states/
│       │   │   ├── app.state.ts
│       │   │   └── config.state.ts
│       │   │
│       │   ├── components/
│       │   │   ├── app/
│       │   │   │   ├── app.component.ts
│       │   │   │   ├── app.component.html
│       │   │   │   └── app.component.scss
│       │   │   │
│       │   │   └── counter-item/
│       │   │       ├── counter-item.component.ts
│       │   │       ├── counter-item.component.html
│       │   │       └── counter-item.component.scss
│       │   │
│       │   └── styles/
│       │       └── global.scss
│       │
│       └── index.html
│
└── docs/
```

---

# 17. System Overview

The pilot application is a deterministic counter list.

Behavior:

- Displays a list of numbers
- Allows selecting an item
- Allows incrementing selected value
- Allows toggling visibility of list
- All updates require explicit render calls

No automatic UI updates occur.

---

# 18. Global States

## 18.1 app.state.ts

```ts
@Stateful
export class AppState {
  public items: number[] = [1, 2, 3]
  public selectedIndex: number | null = null
  public visible: boolean = true
}
```

Behavior:

- Holds mutable application data
- Does not notify
- Does not trigger rendering

Expected behavior:
State mutation alone does nothing.
UI updates only after manual render.

---

## 18.2 config.state.ts

```ts
@Stateless
export class ConfigState {
  public title: string = "Deterministic Counter"
}
```

Behavior:

- Immutable configuration
- Injected but not mutated

---

# 19. Root Component

## 19.1 app.component.ts

```ts
@Component({
  stateful: [AppState],
  stateless: [ConfigState]
})
export class AppComponent {

  constructor(
    private app: AppState,
    private config: ConfigState
  ) {}

  onInit() {}

  toggle() {
    this.app.visible = !this.app.visible
    this.render()
  }

  select(i: number) {
    this.app.selectedIndex = i
    this.render()
  }

  increment() {
    if (this.app.selectedIndex !== null) {
      this.app.items[this.app.selectedIndex]++
      this.render()
    }
  }
}
```

Behavior created:

- Explicit state mutation
- Explicit render calls
- No hidden observers

---

## 19.2 app.component.html

```html
<div class="p-6 bg-gray-100 rounded-xl container">

  <h1 class="text-2xl font-bold mb-4 title">
    {{ config.title }}
  </h1>

  <button class="btn" (click)="toggle()">
    Toggle List
  </button>

  <if condition="app.visible">
    <div class="mt-4">
      <loop for="item in app.items" index="i">
        <CounterItem
          (click)="select(i)">
        </CounterItem>
      </loop>
    </div>
  </if>

  <if condition="app.selectedIndex !== null">
    <div class="mt-4 flex gap-2">
      <button class="btn-primary" (click)="increment()">
        Increment Selected
      </button>
      <span>
        Selected: {{ app.selectedIndex }}
      </span>
    </div>
  </if>

</div>
```

Features used:

- Interpolation
- <if>
- <loop>
- Event binding
- Nested component instantiation
- Tailwind classes
- SCSS class usage

Expected output (initial):

- Title displayed
- Three counter items rendered
- No selection visible

---

## 19.3 app.component.scss

```scss
.container {
  border: 1px solid #e5e7eb;
}

.title {
  color: #111827;
}

.btn {
  @apply px-4 py-2 bg-gray-300 rounded-lg;
}

.btn-primary {
  @apply px-4 py-2 bg-blue-500 text-white rounded-lg;
}
```

Behavior:

- SCSS processed at compile time
- Tailwind applied natively
- Styles encapsulated in ShadowRoot

---

# 20. Child Component

## 20.1 counter-item.component.ts

```ts
@Component({
  stateful: [AppState]
})
export class CounterItem {

  constructor(private app: AppState) {}

  get value(): number {
    return this.app.items[this.context.index]
  }
}
```

Behavior:

- Injects global state
- Reads value using loop index context
- No local state

---

## 20.2 counter-item.component.html

```html
<div class="item p-2 bg-white rounded-lg shadow mb-2">
  {{ value }}
</div>
```

Expected output for items [1,2,3]:

<div>1</div>
<div>2</div>
<div>3</div>

---

## 20.3 counter-item.component.scss

```scss
.item {
  border: 1px solid #d1d5db;
  cursor: pointer;
}
```

---

# 21. Entry Point

## main.ts

```ts
bootstrap(AppComponent, {
  mount: document.getElementById("app")
})
```

Behavior:

- Instantiates root component
- Engine performs first render automatically
- No implicit future renders

---

# 22. Expected Interaction Flow

1. Initial render shows items 1,2,3
2. Clicking an item calls select(i)
3. State updates
4. render() is called manually
5. UI updates
6. Clicking increment mutates state
7. render() called
8. UI updates
9. Toggle hides list
10. Toggle again recreates component instances

Destroy behavior:

- <if> false removes subtree
- Component instances destroyed
- ShadowRoot disposed

---

# 23. Full File Placement Summary

```
examples/pilot-app/src/

main.ts

states/
  app.state.ts
  config.state.ts

components/app/
  app.component.ts
  app.component.html
  app.component.scss

components/counter-item/
  counter-item.component.ts
  counter-item.component.html
  counter-item.component.scss
```

This pilot fully exercises:

- Deterministic rendering
- Manual state mutation
- Explicit lifecycle
- DSL parsing
- Nested components
- Shadow DOM isolation
- SCSS + Tailwind integration
- AOT template compilation

This implementation becomes the canonical reference for validating the compiler and runtime behavior.

---

# 24. Technical Specification — Compiler (Go) and Core Architecture

This section formally defines the compiler architecture, DSL grammar, AST structure, runtime contract, lifecycle model, memory model, state system, repository structure, and testing strategy.

---

# 25. Formal DSL Specification (Grammar)

The DSL is HTML-based with restricted structural extensions.

It is parsed as a deterministic grammar.

## 25.1 Lexical Rules

- Identifiers: `[a-zA-Z_][a-zA-Z0-9_]*`
- Interpolation: `{{ expression }}`
- Event binding: `(event)="expression"`
- Loop: `<loop for="item in expression" index="i">`
- If: `<if condition="expression">`

Expressions are valid TypeScript expressions without statements.

---

## 25.2 Structural Grammar (Simplified EBNF)

```
Template      ::= Node*

Node          ::= Element
               | Component
               | Loop
               | If
               | Text
               | Interpolation

Element       ::= "<" Identifier Attribute* ">" Node* "</" Identifier ">"

Component     ::= "<" CapitalizedIdentifier Attribute* ">" Node* "</" CapitalizedIdentifier ">"

Loop          ::= "<loop" "for=" Expression "" "index=" Identifier "" ">" Node* "</loop>"

If            ::= "<if" "condition=" Expression "" ">" Node* "</if>"

Interpolation ::= "{{" Expression "}}"
```

Constraints:

- `<loop>` requires `for` and `index`
- `<if>` requires `condition`
- Component tags must start with uppercase
- Nested structures are unlimited

---

# 26. AST Specification

The compiler produces a strongly typed AST.

## 26.1 Core Node Interface

```
type Node interface {
    Type() NodeType
    Location() SourceLocation
}
```

## 26.2 Node Types

```
type NodeType int

const (
    ElementNode NodeType
    ComponentNode
    LoopNode
    IfNode
    TextNode
    InterpolationNode
)
```

---

## 26.3 Concrete AST Structures

### ElementNode

```
type Element struct {
    Tag        string
    Attributes []Attribute
    Children   []Node
}
```

### ComponentNode

```
type Component struct {
    Name       string
    Attributes []Attribute
    Children   []Node
}
```

### LoopNode

```
type Loop struct {
    Iterator   string
    Expression string
    Index      string
    Children   []Node
}
```

### IfNode

```
type If struct {
    Condition string
    Children  []Node
}
```

### InterpolationNode

```
type Interpolation struct {
    Expression string
}
```

---

# 27. Compiler Architecture (Go)

The compiler is structured in deterministic stages.

## 27.1 Pipeline

1. File Loader
2. Lexer
3. Parser → AST
4. Semantic Analyzer
5. Template Validator
6. Linter
7. Code Generator
8. Bundler Output

Each stage runs in isolated goroutines where safe.

Parallelization:

- Independent component compilation
- Independent style compilation
- Concurrent linting

No shared mutable state between goroutines.

---

# 28. Runtime Internal Architecture

The runtime is minimal and deterministic.

## 28.1 Core Modules

- ComponentRegistry
- Injector
- Renderer
- LifecycleManager
- DestroyManager

---

## 28.2 Renderer Model

Render functions are generated at build time.

Signature:

```
function render(ctx, root, context)
```

Where:

- ctx → component instance
- root → ShadowRoot
- context → loop context (index, iterator)

No diffing.
No reconciliation.
Render clears and rebuilds subtree deterministically.

---

# 29. Lifecycle Contract

## 29.1 Component Lifecycle

Executed only during instantiation or rerender():

1. constructor
2. dependency injection
3. onInit()
4. afterInit()

On destroy:

5. onDestroy()

---

## 29.2 Render Lifecycle

Executed during render():

1. beforeRender()
2. DOM construction
3. afterRender()

render() does NOT trigger component lifecycle.
rerender() triggers full component lifecycle.

---

# 30. Memory and Destruction Model

The framework enforces explicit destruction.

## 30.1 Destruction Triggers

- <if> condition becomes false
- Loop item removed
- Parent destroyed
- fullrender()

---

## 30.2 Destruction Process

1. Call onDestroy()
2. Remove event listeners
3. Clear ShadowRoot
4. Remove DOM references
5. Nullify internal references

No retained closures allowed.

Garbage collection is delegated to browser after references are cleared.

---

# 31. State System Specification

States are plain classes annotated with:

- @Stateful
- @Stateless

## 31.1 Registration

At bootstrap, states are registered in a global container.

## 31.2 Injection

Constructor-based only.

## 31.3 Rules

- No automatic observers
- No mutation interception
- No lifecycle
- No internal subscriptions

States are passive memory containers.

---

# 32. Official Repository Structure

```
root/
├── packages/
│   ├── compiler/
│   │   ├── lexer/
│   │   ├── parser/
│   │   ├── ast/
│   │   ├── analyzer/
│   │   ├── generator/
│   │   └── main.go
│   │
│   ├── runtime/
│   │   ├── core/
│   │   ├── renderer/
│   │   ├── lifecycle/
│   │   ├── injector/
│   │   └── destroy/
│   │
│   └── cli/
│       └── main.go
│
├── examples/
├── docs/
└── tests/
```

---

# 33. Testing Strategy

Testing is mandatory at multiple levels.

## 33.1 Compiler Tests

- DSL parsing tests
- AST shape validation
- Invalid structure detection
- Snapshot tests for generated render code
- Concurrency safety tests

## 33.2 Runtime Tests

- Lifecycle order verification
- Destroy correctness
- Event binding behavior
- Nested component creation/destruction
- Memory leak detection (integration tests)

## 33.3 Integration Tests

- Compile pilot app
- Execute in headless browser
- Validate DOM output
- Validate deterministic behavior

## 33.4 Stress Tests

- Deep nested loops
- 300 nested ifs
- 1,000 component instances

The framework must behave deterministically under extreme nesting.

---

# 34. Final Architectural Statement

The system is built on five non-negotiable pillars:

1. Determinism
2. Explicit rendering
3. Zero hidden reactivity
4. Compile-time enforcement
5. Manual architectural responsibility

Any future feature must preserve these constraints.

---

# RFC 0001 — Render Engine Deterministic Model

## 1. Purpose

This RFC formally defines the deterministic rendering engine. The objective is to guarantee that rendering behavior is:

- Explicit
- Predictable
- Side-effect free (unless explicitly coded)
- Free from implicit scheduling or reactivity

The render engine is snapshot-based and manually triggered.

---

## 2. Core Principles

1. Rendering only occurs when explicitly invoked.
2. State mutation does not trigger rendering.
3. Observable emission does not trigger rendering.
4. No diffing algorithm is used.
5. No virtual DOM is used.
6. No automatic batching exists.

---

## 3. Render Types

### 3.1 render()

- Clears ShadowRoot
- Executes compiled render function
- Instantiates children
- Executes render lifecycle hooks

Does NOT:
- Recreate component instance
- Re-run constructor
- Re-run onInit()

---

### 3.2 rerender()

- Calls onDestroy()
- Destroys subtree
- Reinstantiates component
- Re-injects dependencies
- Re-runs full lifecycle
- Rebuilds DOM

---

### 3.3 fullrender()

- Recursively applies rerender() to component and all descendants

---

## 4. Determinism Guarantee

Given identical state input and identical render invocation sequence, the DOM output must be identical.

There is no time-based behavior.
There is no scheduler.
There is no internal state caching.

---

## 5. Render Execution Flow

1. beforeRender()
2. Clear ShadowRoot
3. Execute compiled DOM instructions
4. Attach children
5. afterRender()

No hidden branching.

---

# RFC 0002 — Dependency Injection Contract

## 1. Purpose

Define a deterministic and minimal dependency injection system.

The DI container exists to:

- Instantiate states
- Provide shared instances
- Resolve constructor dependencies

It does NOT:

- Handle lifecycle
- Track scopes dynamically
- Perform reflection at runtime

---

## 2. Registration Phase

At bootstrap:

- All @Stateful and @Stateless classes are registered
- Single global container is built

---

## 3. Injection Rules

1. Constructor-only injection
2. No property injection
3. No runtime token resolution
4. No circular dependency support

Circular dependencies must fail at compile time.

---

## 4. Lifetime Rules

- Stateful: single instance
- Stateless: single instance

Future scoped injection must not violate determinism.

---

# RFC 0003 — DSL Expression Engine

## 1. Purpose

Define how expressions inside:

- {{ }}
- <if>
- <loop>
- (event)="..."

are parsed and validated.

---

## 2. Expression Constraints

Allowed:
- Property access
- Method call
- Arithmetic
- Logical operators
- Array access
- Ternary operator

Forbidden:
- Assignment
- Variable declaration
- Function declaration
- Await
- New expressions
- Side-effect statements

Expressions must be pure.

---

## 3. Compile-Time Validation

The compiler must:

- Validate symbol existence
- Validate method existence
- Validate loop variable scope
- Detect illegal mutation inside template

---

## 4. Output Strategy

Expressions are embedded directly into generated render functions.
No runtime expression interpreter exists.

---

# CLI Specification

## 1. Purpose

Provide a deterministic interface for:

- Project creation
- Compilation
- Development mode
- Production build

---

## 2. Commands

### create

```
framework create my-app
```

Scaffolds:
- Directory structure
- tsconfig
- Tailwind config
- Example app

---

### build

```
framework build
```

Actions:
- Compile DSL
- Generate JS
- Compile SCSS
- Bundle output
- Minify (production mode)

---

### dev

```
framework dev
```

Features:
- File watcher
- Incremental compile
- Error overlay

No hot re-rendering.
Developer must trigger UI updates via render().

---

### lint

```
framework lint
```

Validates:
- Template correctness
- DSL constraints
- State misuse
- Circular injection

---

# npm Integration Documentation

## 1. Package Structure

Published packages:

- @framework/runtime
- @framework/cli

---

## 2. Installation

```
npm install @framework/runtime
npm install -D @framework/cli
```

---

## 3. package.json Integration

```
{
  "scripts": {
    "dev": "framework dev",
    "build": "framework build",
    "lint": "framework lint"
  }
}
```

---

## 4. Browser Output

The CLI outputs:

- Compiled JS bundle
- Compiled CSS
- Asset folder

No runtime template parsing shipped.

---

# Philosophical Manifesto

This framework is not built for convenience.
It is built for control.

It rejects:

- Hidden reactivity
- Implicit rendering
- Scheduler-based magic
- Mutation tracking
- Virtual DOM illusions

It embraces:

- Explicit mutation
- Manual rendering
- Deterministic behavior
- Structural discipline
- Compile-time guarantees

If you forget to call render(), nothing happens.
This is not a flaw.
This is the design.

The framework assumes competence.
It does not protect developers from architectural decisions.

It enforces structure for clarity, not for restriction.
It limits abstraction to reduce entropy.

It is designed for:

- AI-generated code
- Predictable enterprise systems
- Developers who prefer explicit control over convenience

The guiding belief:

Determinism scales.
Magic does not.

---

# Threat Model and Security Architecture

Security is treated as a first-class architectural constraint. The framework must not introduce new classes of vulnerabilities beyond those already inherent to the browser platform.

## 1. XSS (Cross-Site Scripting)

### 1.1 Interpolation Safety

All `{{ expression }}` interpolations are compiled into:

```
node.textContent = value
```

Never:

```
innerHTML = value
```

This guarantees automatic escaping by the browser.

---

### 1.2 Explicit Unsafe HTML

If raw HTML injection is ever supported in the future, it must require:

- Explicit opt-in directive
- Compile-time warning
- Runtime sanitization hook

Unsafe HTML must never be the default behavior.

---

## 2. Template Injection

Templates are compiled AOT.

There is:
- No runtime template parsing
- No runtime expression evaluation
- No string-to-code execution

This eliminates template injection vectors common in string-based systems.

---

## 3. Expression Sandboxing

Expressions are statically analyzed.

Forbidden operations:
- Assignment
- `new`
- `eval`
- `Function` constructor
- Global object access
- `window`, `document` direct reference in DSL

The compiler must reject any expression attempting to escape component scope.

---

## 4. Event Binding Safety

Event handlers are direct function references.

No string-based handler evaluation is allowed.

Example generated code:

```
button.addEventListener("click", ctx.increment.bind(ctx))
```

---

## 5. State Isolation

States are plain classes.

There is:
- No implicit sharing of DOM references
- No automatic cross-component mutation

Security boundaries remain explicit and visible.

---

## 6. Shadow DOM Boundary

Each component uses `ShadowRoot (open)`.

Benefits:
- Style encapsulation
- Reduced accidental DOM traversal
- Scoped event handling

Shadow DOM is not a security boundary but reduces accidental surface exposure.

---

# Versioning Model and Breaking Policy

The framework follows Semantic Versioning (SemVer):

MAJOR.MINOR.PATCH

## 1. PATCH

- Bug fixes
- Performance improvements
- Internal refactors

No public API changes.

---

## 2. MINOR

- Backward-compatible feature additions
- New CLI commands
- Extended DSL capabilities (non-breaking)

Existing applications must continue compiling without modification.

---

## 3. MAJOR

Breaking changes include:

- DSL grammar changes
- Lifecycle contract modifications
- Render engine behavior modification
- DI contract change
- State model alteration

---

## 4. Breaking Change Policy

Before any MAJOR release:

1. RFC must be published
2. Migration guide must be written
3. Deprecation warnings must exist for at least one MINOR cycle

No silent breaking changes are allowed.

---

# Branding and Public Identity

## 1. Technical Positioning

The framework is positioned as:

"A deterministic UI engine for developers and AI systems who demand explicit control."

It does not compete on convenience.
It competes on predictability.

---

## 2. Identity Pillars

- Deterministic
- Explicit
- Compile-time enforced
- Minimal runtime
- Architecturally strict

---

## 3. Tone of Communication

Clear. Technical. Direct.
No hype-driven messaging.
No buzzword alignment.

---

## 4. Target Audience

- Engineers building internal systems
- Teams requiring deterministic rendering
- AI-generated frontend pipelines
- Developers fatigued by reactive magic

---

# Open-Source Governance Model

The project follows a structured governance model to ensure architectural consistency.

---

## 1. Maintainers

Core maintainers are responsible for:

- Approving RFCs
- Reviewing PRs
- Enforcing architectural constraints
- Managing releases

Maintainers must preserve determinism principles.

---

## 2. RFC Process

All significant changes require an RFC.

RFC flow:

1. Proposal submitted via /rfcs directory
2. Community discussion (minimum review window defined)
3. Maintainer review
4. Approval or rejection
5. Implementation phase

No major feature enters the core without RFC approval.

---

## 3. Contribution Model

Contribution types:

- Bug fix
- Documentation improvement
- Performance optimization
- Feature proposal (RFC required)

All PRs must:

- Include tests
- Pass linting
- Preserve deterministic guarantees

---

## 4. Architectural Protection Rule

Any contribution introducing:

- Implicit reactivity
- Automatic rendering
- Hidden scheduling
- Runtime template parsing

Will be rejected.

The core philosophy is non-negotiable.

---

## 5. Decision Authority

Final architectural decisions belong to core maintainers.

The goal is not maximal flexibility.
The goal is long-term structural integrity.



