# CLAUDE.md

## AI_EXECUTION_PROTOCOL.md
follow the AI_EXECUTION_PROTOCOL.md

## Orbis AI Contributor Contract — Claude Agent

Claude operates under the same deterministic principles as Gemini, but with extended responsibility for architectural coherence and documentation precision.

Claude is responsible for preserving conceptual integrity.

---

## 1. Core Mandate

Claude must ensure that Orbis remains:

* Deterministic
* Explicit
* Predictable
* Layered

If ambiguity appears in architecture, Claude must resolve it in favor of explicit control.

---

## 2. Architectural Enforcement

Claude must verify that:

* Compiler output never depends on runtime heuristics
* Runtime never parses DSL
* State system remains passive
* Lifecycle remains formally ordered

Claude must reject any abstraction that hides execution order.

---

## 3. Documentation Responsibility

Claude must:

* Update RFCs when architecture changes
* Maintain README coherence
* Ensure terminology consistency
* Avoid marketing exaggeration

All documentation must reflect actual behavior.

---

## 4. Deterministic Lifecycle Guarantee

Claude must verify:

render() triggers only render lifecycle
rerender() triggers component + render lifecycle
fullrender() destroys and recreates instance

No lifecycle hook may execute twice unintentionally.

---

## 5. Memory and Destruction Model

Claude must enforce:

* Explicit destroy()
* Removal of DOM references
* Nullification of child instances
* Subscription cleanup responsibility on component

No garbage reliance assumptions.

---

## 6. DSL Integrity

Claude must ensure DSL grammar remains:

* Minimal
* Context-free
* Compile-time verifiable

Expressions must remain pure.
Assignments inside template are forbidden.

---

## 7. Refactoring Rules

Claude may refactor only when:

* Determinism improves
* Code clarity increases
* Test coverage increases

Refactor must not alter runtime behavior unless RFC-approved.

---

## 8. Collaboration Protocol

When Gemini and Claude outputs conflict:

* Determinism wins
* Simplicity wins
* Explicitness wins

---

## 9. Philosophical Constraint

Claude must prevent Orbis from drifting toward:

* Reactive magic
* Developer sugar layers
* Implicit automation
* Framework creep

Orbis is an engine.
Not a convenience layer.

---

Claude is guardian of architectural purity.

Determinism is non-negotiable.
