---
title: "Persistence — Section Overview"
id: pers-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, framework-authors]
since: "1.0"
normative-level: informative
related:
  - "[EntityRepository](entity-repository.md)"
  - "[Filter DSL](filter-dsl.md)"
  - "[Five-Layer Architecture](../02-architecture/five-layer.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Persistence

**Section 05 | Persistence Layer**

The persistence section specifies the interfaces and DSLs through which all entity data is stored and queried. The Store Layer (which implements these interfaces) is the only layer that communicates directly with PostgreSQL.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [EntityRepository](entity-repository.md) | PERS-001 | Full interface specification: all methods, options, transaction semantics | FROZEN |
| [Filter DSL](filter-dsl.md) | PERS-002 | Declarative predicate DSL: operators, composition, wire format | FROZEN |
| [System Entities](system-entities.md) | PERS-003 | SQL-backed entities: schema conventions, mandatory types, migration requirements | FROZEN |
| [Custom Entities](custom-entities.md) | PERS-004 | JSONB-backed entities: storage model, custom fields, escalation criteria | STABLE |

---

## Prerequisites

- [EntityDefinition](../03-kernel/entity-definition.md)
- [Five-Layer Architecture §6](../02-architecture/five-layer.md#6-store-layer) — Store Layer responsibilities
- [Architecture Laws](../02-architecture/laws.md) — LAW-005 (tenant context), LAW-014 (filter versioning)
- [Tenancy Model](../06-tenancy/tenant-model.md) — RLS, set_tenant_context()
- [Glossary](../GLOSSARY.md) — EntityRepository, Filter, QueryOption, System Entity, Custom Entity
