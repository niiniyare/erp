---
title: "API Layer — Section Overview"
id: api-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, framework-authors]
since: "1.0"
normative-level: informative
related:
  - "[API Conventions](conventions.md)"
  - "[Error Handling](error-handling.md)"
  - "[RBAC](../07-iam/rbac.md)"
  - "[Sessions](../07-iam/sessions.md)"
  - "[Architecture Laws](../02-architecture/laws.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# API Layer

**Section 11 | API Layer**

The API Layer is Awo's external boundary. It translates HTTP requests into domain operations and HTTP responses into structured envelopes. The framework auto-generates standard CRUD routes from `EntityDefinition`s — custom handlers are for exceptional cases only.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [API Conventions](conventions.md) | API-001 | URL structure, response envelope, pagination, datetime serialization | STABLE |
| [Error Handling](error-handling.md) | API-002 | Error types, HTTP status mapping, envelope format, field-level errors | STABLE |

---

## Prerequisites

- [Five-Layer Architecture](../02-architecture/five-layer.md) — API Layer constraints
- [EntityDefinition](../03-kernel/entity-definition.md) — what drives route generation
- [RBAC](../07-iam/rbac.md) — how permissions are checked in route handlers
- [Sessions](../07-iam/sessions.md) — session validation middleware
- [Glossary](../GLOSSARY.md) — API Layer, Route Handler, Action, Response Envelope
