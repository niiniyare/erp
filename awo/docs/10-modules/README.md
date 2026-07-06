---
title: "Modules — Section Overview"
id: mod-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, framework-authors]
since: "1.0"
normative-level: informative
related:
  - "[Module System](module-system.md)"
  - "[Platform Modules](platform-modules.md)"
  - "[EntityDefinition](../03-kernel/entity-definition.md)"
  - "[Registry](../03-kernel/registry.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Modules

**Section 10 | Module System**

A module is a cohesive collection of entity definitions, hooks, policies, activities, and migrations that addresses a specific business domain. Every capability in Awo — including the built-in platform capabilities — is delivered through modules.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [Module System](module-system.md) | MOD-001 | Module structure, manifest, registration, dependency resolution | STABLE |
| [Platform Modules](platform-modules.md) | MOD-002 | Built-in platform modules: Tenant, IAM, Flags, Settings, Audit, Metadata, Registry | STABLE |

---

## Prerequisites

- [EntityDefinition](../03-kernel/entity-definition.md) — the central primitive modules build on
- [Registry](../03-kernel/registry.md) — how modules register their definitions
- [Startup Sequence](../03-kernel/startup-sequence.md) — when module init() runs
- [Glossary](../GLOSSARY.md) — Module, ModuleManifest, Platform Module, Business Module
