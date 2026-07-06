---
title: "IAM — Section Overview"
id: iam-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, framework-authors, operators]
since: "1.0"
normative-level: informative
related:
  - "[RBAC](rbac.md)"
  - "[Sessions](sessions.md)"
  - "[Authentication](authentication.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# IAM

**Section 07 | Identity and Access Management**

The IAM module provides authentication, session management, and role-based access control for all tenant-scoped operations.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [RBAC](rbac.md) | IAM-001 | Casbin model, role hierarchy, permission evaluation, system roles | FROZEN |
| [Sessions](sessions.md) | IAM-002 | Session lifecycle, Redis storage, token format, expiry | FROZEN |
| [Authentication](authentication.md) | IAM-003 | Credential verification, session creation, logout | STABLE |

---

## Prerequisites

- [Tenant Model](../06-tenancy/tenant-model.md) — tenant context required before IAM checks
- [Architecture Laws](../02-architecture/laws.md) — LAW-005 (context required), LAW-013 (sensitive fields)
- [Architecture Invariants](../02-architecture/invariants.md) — INV-007 (Redis required for session validation)
- [Glossary](../GLOSSARY.md) — Actor, Session, RBAC, Casbin, Role, Permission
