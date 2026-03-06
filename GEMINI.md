# GEMINI.md

## AI_EXECUTION_PROTOCOL.md
follow the AI_EXECUTION_PROTOCOL.md

## Orbis AI Contributor Contract — Gemini Agent

This document defines the operational behavior, architectural constraints, and execution order that Gemini must follow when contributing to the Orbis repository.

Gemini is not a creative participant.
Gemini is a deterministic engineering agent.

---

## 1. Primary Directive

Gemini must preserve the deterministic nature of Orbis.

Under no circumstance may Gemini:

* Introduce implicit reactivity
* Introduce automatic rendering
* Introduce runtime template interpretation
* Add hidden dependency tracking
* Add silent side-effects

If a requested feature violates determinism, Gemini must refuse and propose a deterministic alternative.

---

## 2. Architectural Awareness

Before modifying any code, Gemini must identify which layer is being affected:

* Compiler (Go)
* Runtime (JavaScript)
* DSL grammar
* State system
* Lifecycle engine

Gemini must never mix responsibilities between layers.

Compiler changes must not leak runtime logic.
Runtime changes must not introduce compiler assumptions.

---

## 3. Order of Implementation

When implementing a new feature:

1. Validate RFC existence
2. Update DSL grammar (if required)
3. Update AST definitions
4. Update compiler transformation logic
5. Update runtime support (if needed)
6. Update tests
7. Update documentation

No feature is complete without tests.

---

## 4. Coding Standards

### Go (Compiler)

* Pure functions whenever possible
* No global mutable state
* Deterministic parsing
* No hidden goroutine side-effects
* Errors must be explicit
* No panic for user errors

All compiler diagnostics must:

* Include file
* Include line
* Include column
* Provide actionable message

---

### JavaScript (Runtime)

* Class-based components only
* Constructor-only dependency injection
* No dynamic property injection
* No proxy-based tracking
* No mutation observers

Shadow DOM must always be `open`.

---

## 5. State Rules

Gemini must treat state as:

* Pure data containers
* No lifecycle
* No rendering knowledge
* No side effects

Gemini must never:

* Auto-subscribe to state
* Auto-render on state mutation
* Inject hidden observers

---

## 6. Testing Strategy

Gemini must write:

* Unit tests for compiler parsing
* Snapshot tests for AST output
* Runtime instantiation tests
* Deterministic lifecycle tests

Tests must prove:

* No implicit rendering occurs
* Lifecycle order is preserved
* Destroy fully releases references

---

## 7. Prohibited Patterns

Gemini must reject:

* Virtual DOM implementation
* Signal-based state tracking
* Automatic batching
* Template evaluation via eval/new Function
* Expression strings evaluated at runtime

---

## 8. Commit Discipline

Every commit must:

* Be atomic
* Reference RFC if applicable
* Not mix refactor + feature
* Pass full test suite

---

## 9. Performance Philosophy

Gemini must optimize for:

* Predictable execution
* Compile-time guarantees
* Minimal runtime overhead

Not for:

* Developer convenience
* API sugar
* Hidden automation

---

Gemini operates under strict determinism.
Convenience is secondary.
