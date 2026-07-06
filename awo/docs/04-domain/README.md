---
title: "Domain Layer — Section Overview"
id: domain-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, framework-authors, contributors]
since: "1.0"
normative-level: informative
related:
  - "[EntityDefinition](../03-kernel/entity-definition.md)"
  - "[Five-Layer Architecture](../02-architecture/five-layer.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Domain Layer

**Section 04 | Domain Layer**

The Domain Layer is the stateless core of every module. It contains declarations and logic that define what entities are, how they behave, who may access them, and what happens as they move through their lifecycle.

Domain Layer code has zero external dependencies — it imports only `awo/def/`, `awo/filter/`, and the Go standard library. This property enables the entire domain layer to be tested without running infrastructure.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [Fields](fields.md) | DOM-001 | FieldDef, all FieldType handlers, constraints, serialization | FROZEN |
| [Edges](edges.md) | DOM-002 | EdgeDef, edge types, explicit loading, cascade rules | FROZEN |
| [Hooks](hooks.md) | DOM-003 | Lifecycle hook system, HookRegistration, execution order | FROZEN |
| [Policy Functions](policies.md) | DOM-004 | Row-level filter injection, PolicyFunc composition | FROZEN |
| [Actions](actions.md) | DOM-005 | Custom operations, ActionDef, ActionContext, ActionResult | STABLE |
| [Naming Series](naming-series.md) | DOM-006 | Format string syntax, atomic counter, tenant override, ResetOnYear | STABLE |

---

## Prerequisites

- [EntityDefinition](../03-kernel/entity-definition.md) — the struct that contains all domain declarations
- [Five-Layer Architecture](../02-architecture/five-layer.md) §4 — Domain Layer responsibilities and import rules
- [Architecture Laws](../02-architecture/laws.md) — LAW-007 (hook order), LAW-008 (imports), LAW-013 (sensitive fields), LAW-017 (cross-module hooks)
- [Glossary](../GLOSSARY.md)
