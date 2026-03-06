# Orbis

## Orbis — Deterministic UI Engine for AI-First Systems

Orbis is a declarative, deterministic UI engine designed for AI-generated code and engineers who demand explicit architectural control.

It eliminates hidden reactivity, automatic rendering, implicit scheduling, and runtime template interpretation.

If you don’t call `render()`, nothing happens.
This is not a limitation.
This is the design.

---

# Why Orbis Exists

Modern UI frameworks optimize for convenience.
Orbis optimizes for determinism.

Reactive abstractions, virtual DOM diffing, change detection, signals, and automatic subscriptions introduce invisible execution paths. Those behaviors increase cognitive complexity and make AI-generated code unpredictable.

Orbis enforces:

- Explicit rendering
- Compile-time guarantees
- Deterministic lifecycle
- No implicit side effects
- No runtime template parsing

Orbis is built for systems where predictability matters more than convenience.

---

# Core Principles

## 1. Deterministic Rendering

Rendering only occurs when explicitly invoked:

- `render()`
- `rerender()`
- `fullrender()`

No state mutation triggers UI updates.
No observable emission triggers UI updates.

---

## 2. Compile-Time Enforcement

Templates are compiled Ahead-of-Time using the Orbis compiler (written in Go).

There is:

- No runtime template parser
- No string-based rendering
- No expression interpreter in the browser

All template logic becomes static DOM instructions.

---

## 3. Explicit State Model

States are plain classes annotated with:

- `@Stateful`
- `@Stateless`

They:

- Contain only data
- Have no lifecycle
- Trigger no rendering
- Perform no automatic observation

Components decide when rendering occurs.

---

## 4. OOP + Standardized Async

- Components are class-based
- Constructor-only dependency injection
- RxJS is the only asynchronous pattern
- No alternative reactive paradigms

Architectural discipline is enforced, not suggested.

---

# Architecture Overview

Orbis is composed of two primary parts:

## Compiler (Go)

Responsibilities:

- Parse HTML-based DSL
- Validate structure
- Enforce grammar rules
- Perform linting
- Generate optimized render functions
- Detect structural errors at build time

Parallelized using goroutines.

---

## Runtime (JavaScript)

Responsibilities:

- Component instantiation
- Shadow DOM creation (open mode)
- Manual rendering
- Lifecycle orchestration
- Controlled destruction

The runtime does NOT:

- Perform diffing
- Perform automatic change detection
- Batch updates
- Track reactive dependencies

---

# DSL Overview

The Orbis template DSL is intentionally minimal.

Supported constructs:

- `{{ interpolation }}`
- `<loop for="item in items" index="i">`
- `<if condition="expression">`
- `(event)="handler()"`

No pipes.
No directives.
No custom template language layers.

Expressions must be pure.
Assignments and side-effect statements are rejected at compile time.

---

# Lifecycle Model

## Component Lifecycle

- constructor
- dependency injection
- onInit()
- afterInit()
- onDestroy()

Triggered only during instantiation or `rerender()`.

## Render Lifecycle

- beforeRender()
- DOM construction
- afterRender()

Triggered during `render()`.

Component lifecycle and render lifecycle are independent.

---

# Installation

```bash
npm install @orbis/runtime
npm install -D @orbis/cli
```

Add scripts to your `package.json`:

```json
{
  "scripts": {
    "dev": "orbis dev",
    "build": "orbis build",
    "lint": "orbis lint"
  }
}
```

---

# CLI Commands

## Create Project

```bash
orbis create my-app
```

## Development Mode

```bash
orbis dev
```

## Production Build

```bash
orbis build
```

## Lint Templates

```bash
orbis lint
```

---

# Security Model

- Interpolations use `textContent` (no raw HTML injection)
- No runtime template evaluation
- No expression string execution
- Compile-time expression validation

Security is enforced by design, not by convention.

---

# Versioning

Orbis follows Semantic Versioning (SemVer):

MAJOR.MINOR.PATCH

Breaking changes require:

- Published RFC
- Migration guide
- Deprecation cycle

Deterministic behavior is a non-negotiable constraint.

---

# Governance

Orbis follows an RFC-based governance model.

All architectural changes require:

1. RFC proposal
2. Community discussion
3. Maintainer approval

Any feature introducing implicit reactivity or automatic rendering will be rejected.

---

# Who Should Use Orbis?

- Teams building AI-generated UI systems
- Engineers requiring strict render control
- Internal tooling platforms
- Deterministic enterprise dashboards

If you want automatic magic, Orbis is not for you.

If you want explicit control and structural guarantees, welcome.

---

# Philosophy

Orbis assumes competence.

It does not prevent architectural mistakes.
It enforces structural clarity.
It rejects hidden behavior.

Determinism scales.
Magic does not.

---

# Status

Orbis is currently in early development.

The initial milestone focuses on:

- Core compiler
- Minimal runtime
- Deterministic rendering engine
- State injection system

Router, SSR, devtools, and extended tooling will be introduced only after the core is stable.

---

# License

MIT License (proposed)

---

Orbis — Deterministic UI Engine for AI-First Systems

Explicit UI.
Zero surprises.

