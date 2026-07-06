---
title: "Security — Section Overview"
id: sec-000-readme
status: accepted
category: GUIDE
stability: STABLE
audience: [module-authors, operators, framework-authors]
since: "1.0"
normative-level: informative
related:
  - "[Security Model](security-model.md)"
  - "[RBAC](../07-iam/rbac.md)"
  - "[Sessions](../07-iam/sessions.md)"
  - "[RLS](../06-tenancy/rls.md)"
  - "[Architecture Laws](../02-architecture/laws.md)"
  - "[Awo Glossary](../GLOSSARY.md)"
---

# Security

**Section 15 | Security**

Security in Awo is structural — not a layer applied on top of existing logic. Tenant isolation, access control, session management, and data protection are baked into the framework's core primitives.

---

## Contents

| Document | ID | Purpose | Stability |
|---|---|---|---|
| [Security Model](security-model.md) | SEC-001 | Defense-in-depth layers, threat model, trust boundaries, compliance notes | STABLE |

---

## Prerequisites

- [RBAC](../07-iam/rbac.md) — Casbin policy model
- [Sessions](../07-iam/sessions.md) — server-side session management
- [RLS](../06-tenancy/rls.md) — PostgreSQL row-level security
- [Architecture Laws](../02-architecture/laws.md) — LAW-005, LAW-013, LAW-015
- [Invariants](../02-architecture/invariants.md) — INV-001, INV-006, INV-007, INV-010
- [Glossary](../GLOSSARY.md) — Defense in Depth, Row Level Security, Session Token, Sensitive Field
