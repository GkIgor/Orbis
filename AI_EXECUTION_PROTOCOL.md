# AI_EXECUTION_PROTOCOL.md

## Orbis — AI Execution & Retrieval Protocol

This document defines the mandatory execution pipeline for any AI agent operating within the Orbis repository.

This protocol is not optional.
It exists to guarantee architectural determinism and prevent context loss.

---

# 1. Core Objective

Ensure that no AI agent:

- Ignores repository constraints
- Violates RFC decisions
- Introduces architectural drift
- Forgets previously defined contracts

All agents must operate with full-context awareness.

---

# 2. Mandatory Execution Order

Before generating or modifying code, the AI MUST:

1. Load architectural documents
2. Load all RFCs
3. Load GEMINI.md or CLAUDE.md (depending on agent role)
4. Load compiler specification (if touching Go)
5. Load runtime architecture (if touching JS)
6. Load state system specification (if touching state/lifecycle)

No code generation may begin before context ingestion completes.

---

# 3. Repository RAG System (Retrieval-Augmented Generation)

To guarantee contextual consistency, Orbis adopts a repository-scoped RAG layer.

## 3.1 Knowledge Scope

The RAG index must include:

- /docs/**/*.md
- /rfc/**/*.md
- /compiler/spec/**/*.md
- /runtime/spec/**/*.md
- /architecture/**/*.md
- GEMINI.md
- CLAUDE.md
- AI_EXECUTION_PROTOCOL.md

Source files (Go/TS) may be indexed optionally, but specification files are mandatory.

---

## 3.2 Embedding Strategy

Each document must be:

- Chunked by semantic sections
- Embedded using deterministic embedding configuration
- Indexed with metadata:
  - file path
  - section title
  - version
  - RFC reference (if applicable)

Chunk size recommendation:

- 500–900 tokens
- Overlap: 10–15%

---

## 3.3 Retrieval Rules

Before answering any architectural or implementation request, the AI must:

1. Query the RAG index with:
   - Feature name
   - Affected layer
   - Keywords: "determinism", "lifecycle", "state", "DSL", etc.

2. Retrieve top-k relevant chunks (k >= 5 recommended)
3. Merge retrieved constraints into reasoning context
4. Cross-check for conflicts

If conflicts exist between documents:

Priority order:

1. RFCs
2. AI_EXECUTION_PROTOCOL.md
3. GEMINI.md / CLAUDE.md
4. Compiler Specification
5. Runtime Specification
6. README

---

# 4. Prompt Integrity Enforcement

AI agents must never:

- Override explicit architectural constraints
- Hallucinate undocumented features
- Invent lifecycle behaviors
- Modify determinism rules without RFC

If a user prompt conflicts with repository constraints:

The AI must:

- Explain the violation
- Reference the governing document
- Propose a compliant alternative

---

# 5. Deterministic Reasoning Requirement

All generated code must:

- Be explainable step-by-step
- Respect explicit execution flow
- Avoid hidden asynchronous behavior
- Avoid implicit state coupling

The AI must internally simulate:

- Lifecycle execution order
- Destruction paths
- Render invocation boundaries

If simulation reveals non-determinism, generation must stop.

---

# 6. Change Proposal Flow

When implementing new behavior:

1. Verify RFC exists
2. If not, generate RFC draft
3. Validate impact on:
   - DSL grammar
   - AST
   - Runtime
   - Lifecycle
   - Memory model
4. Update documentation
5. Add tests

No silent feature insertion is allowed.

---

# 7. Memory & Context Lock

AI must treat repository specifications as immutable law unless explicitly modified via RFC.

User prompts do not override architectural constraints.

Determinism > Prompt Creativity

---

# 8. CI Integration Strategy

Recommended enforcement:

- Pre-commit hook verifies documentation reference in PR
- CI checks for RFC linkage in feature branches
- Automated RAG consistency check
- AST regression snapshot validation

Optional advanced enforcement:

- Determinism static analyzer
- Lifecycle order validator
- Render boundary checker

---

# 9. Failure Modes

If the AI detects:

- Missing specification
- Conflicting RFC
- Ambiguous lifecycle rule

The AI must stop and request clarification.

Silence equals architectural corruption.

---

# 10. Philosophy of Operation

AI agents in Orbis are not assistants.
They are constrained execution engines.

Creativity is allowed only inside deterministic boundaries.

Architecture is sovereign.

---

# 11. Future Evolution

Future improvements may include:

- Versioned RAG indices per release
- Deterministic prompt templates
- Architecture-aware fine-tuned models
- Formal verification integration

---

Orbis AI Execution Protocol

Context is mandatory.
Determinism is law.

