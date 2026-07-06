---
title: "Awo Framework Documentation"
id: docs-root
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, framework-authors, operators]
since: "1.0"
normative-level: informative
---

# Awo Framework Documentation

Awo is a Go-native multi-tenant ERP framework. This documentation set is the authoritative reference for framework contributors, module authors, and operators.

---

## Documentation Sections

| Section | Contents | Audience |
|---|---|---|
| [00 Documentation](00-documentation/README.md) | Governance: writing standards, review process, templates | All |
| [01 Introduction](01-introduction/README.md) | Philosophy, design goals, architecture overview | All |
| [02 Architecture](02-architecture/README.md) | Laws, invariants, five-layer specification (normative) | All |
| [03 Kernel](03-kernel/README.md) | EntityDefinition, Registry, compilation pipeline, startup | Framework authors, Module authors |
| [04 Domain](04-domain/README.md) | Fields, edges, hooks, policies, actions | Module authors |
| [05 Persistence](05-persistence/README.md) | EntityRepository, Filter DSL, system entities, custom entities | Module authors |
| [06 Tenancy](06-tenancy/README.md) | Tenant model, RLS, tenant lifecycle | Framework authors |
| [07 IAM](07-iam/README.md) | RBAC, sessions, authentication | Framework authors, Operators |
| [08 SDUI](08-sdui/README.md) | Page schemas, page builders, amis integration | Module authors |
| [09 Workflow](09-workflow/README.md) | Temporal integration, activities, sagas, outbox | Module authors |
| [10 Modules](10-modules/README.md) | Module system, platform modules | Module authors |
| [11 API](11-api/README.md) | API conventions, error handling | Module authors |
| [12 Configuration](12-configuration/README.md) | Environment variables, secrets | Operators |
| [13 Observability](13-observability/README.md) | Logging, metrics, health, tracing | Operators, Module authors |
| [14 Operations](14-operations/README.md) | Migrations, deployment | Operators |
| [15 Security](15-security/README.md) | Security model, defense-in-depth | All |
| [16 Module Dev Guide](16-module-dev-guide/README.md) | Step-by-step tutorial for building a module | Module authors |
| [17 ADR](17-adr/README.md) | Architecture Decision Records | All |

---

## Entry Points by Role

### New Module Author
Start here:
1. [Philosophy](01-introduction/philosophy.md) — understand the five axioms
2. [Architecture Overview](01-introduction/architecture-overview.md) — mental model
3. [Architecture Laws](02-architecture/laws.md) — the 20 constraints
4. [Module Developer Guide](16-module-dev-guide/README.md) — build your first module

### Framework Contributor
Start here:
1. [Architecture Laws](02-architecture/laws.md) + [Invariants](02-architecture/invariants.md)
2. [Compilation Pipeline](03-kernel/compilation-pipeline.md)
3. [Registry](03-kernel/registry.md)
4. [ADRs](17-adr/README.md) — decisions already made

### Operator
Start here:
1. [Configuration](12-configuration/configuration.md)
2. [Deployment](14-operations/deployment.md)
3. [Migrations](14-operations/migrations.md)
4. [Observability](13-observability/observability.md)

---

## Governing Documents (read first)

These documents have the highest authority in the documentation hierarchy:

1. [Architecture Laws](02-architecture/laws.md) — 20 permanently binding rules
2. [Architecture Invariants](02-architecture/invariants.md) — 12 runtime properties that must hold
3. [Documentation Architecture](00-documentation/documentation-architecture.md) — how docs are organized
4. [Documentation Standards](00-documentation/documentation-standards.md) — how docs are written
5. [GLOSSARY](GLOSSARY.md) — canonical term definitions

---

## Glossary

[GLOSSARY.md](GLOSSARY.md) — 105 canonical term definitions, A–Z with category index.

All documentation uses only terms defined in the Glossary. New terms must be added to the Glossary before use.
