> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Compliance Notes
portal: 4 — Backend Engineering
section: 00-module-development-guide/09-audit-trail
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-audit-overview.md
    title: Audit Trail Overview
---

# Compliance Notes

The audit trail supports financial and operational compliance requirements. This page documents the requirements and how the implementation satisfies them.

## Immutability

Audit records are never updated or deleted by application code. The `audit_log` table has:
- No `DELETE` permission granted to the application DB role.
- No `UPDATE` permission granted to the application DB role.
- A database trigger that rejects mutations.

If a correction is needed, a new audit event is appended with `Action: "audit.correction"` and a reference to the original event.

## Retention

Default audit retention: 7 years (configurable per tenant for regulated industries).

Retention is enforced by a scheduled maintenance job, not by the application code. Business module code never touches retention logic.

## Tamper Evidence

Audit records include a hash chain (each record's hash includes the previous record's hash). This is implemented in the platform audit service — business modules do not need to implement this.

## What the Contracts Module Must Guarantee

1. **Every write is audited.** No state-changing operation returns without firing an audit goroutine.
2. **Before state is captured.** Every update fetches the current entity before modifying it and passes it as `Before`.
3. **Actor is recorded.** Every audit event includes `ActorID` from the session's `UserID`.
4. **TenantID is recorded.** Every audit event includes `TenantID`.
5. **Audit failure is logged but not propagated.** A broken audit service must not prevent contract operations.

## Querying for Compliance Reports

Compliance teams can query the audit trail via the platform admin API:

```
GET /platform/admin/audit?tenant_id=...&resource_type=contract&from=...&to=...
```

Business module developers do not need to build compliance report UI — the platform audit module handles it.
