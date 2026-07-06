---
title: "Documentation System — Section Overview"
id: doc-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [all]
since: "1.0"
normative-level: informative
related:
  - "[Documentation Architecture](documentation-architecture.md)"
  - "[Documentation Standards](documentation-standards.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Documentation System

**Section 00 | Governance**

This section governs the Awo Framework documentation system. Every document it contains applies to the act of writing documentation, not to the framework itself.

These documents do not describe Awo. They define the rules by which Awo is described.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [Documentation Architecture](documentation-architecture.md) | DAS-001 | Complete blueprint of the documentation system: taxonomy, dependency hierarchy, metadata standard, lifecycle, stability model, governance | FROZEN |
| [Documentation Standards](documentation-standards.md) | DSS-001 | Normative writing constitution: templates, naming conventions, cross-reference rules, diagram standards, code example standards, AI authoring rules | FROZEN |

---

## Who Must Read This

**All contributors** who author or review documents in `awo/docs/` must read both documents in this section before doing so. No exceptions.

**Core team members** enforce these governance documents in every documentation review. Failure to conform to either document is grounds for blocking a documentation pull request.

**External contributors** must read Documentation Standards before submitting documentation PRs. The documentation review checklist in that document must be self-applied before submission.

---

## Governing Hierarchy

When conflicts arise between documents at different levels, resolution order is:

1. Approved Architecture Decision Records (ADRs)
2. Kernel specifications (`03-kernel/`)
3. Documentation Standards (`documentation-standards.md`)
4. Documentation Architecture (`documentation-architecture.md`)
5. All remaining documentation

No document in `awo/docs/` may contradict a document ranked above it in this hierarchy.

---

## Documentation and Source Code

Documentation is the authoritative specification of the Awo Framework. Source code is one implementation of that specification.

When documentation and source code conflict:

- The source code is wrong.
- The source code must be corrected to match the documentation.
- The documentation may not be silently altered to match a divergent implementation.

A behavior present in source code but absent from documentation is not a public feature. It is an implementation detail that may change at any time without notice.

This principle is stated in Documentation Architecture §1 and repeated here because its implications are non-obvious: it means that a change to documented behavior requires a documentation change first, then a source code change, not the reverse.

---

## Navigation

New to the documentation system? Read in this order:

1. [Documentation Architecture](documentation-architecture.md) — understand the structure and rules
2. [Documentation Standards](documentation-standards.md) — understand how to write

Then proceed to [01-introduction](../01-introduction/README.md) for framework content.
