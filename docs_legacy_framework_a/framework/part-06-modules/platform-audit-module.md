> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module — Audit Log"
part: "Part VI — Platform Entities"
chapter: 46
section: "platform-audit-module"
related:
  - "[Chapter 41: Platform Modules Overview](./platform-module-overview.md)"
  - "[Chapter 42: Tenant Module](./platform-tenant-module.md)"
  - "[Chapter 45: Settings & Configuration Module](./platform-settings-module.md)"
---

# Chapter 46 — Audit Log Module

> **Who should read this?** Developers adding auditable entities, compliance engineers verifying tamper-evidence, and platform engineers tuning log retention.

---

## 46.1 Why a Centralised Audit Log?

Every regulated ERP deployment must answer: *"Who changed what, when, and from what to what?"* The Audit Log module answers that question for all entities that declare `Audited: true` in their `EntityDefinition`.

The log is:

- **Immutable** — rows are never updated or deleted by application code. Deletion requires a separate `archive_audit_logs` background job (controlled by settings).
- **Tenant-scoped** — each tenant's rows are isolated via Row-Level Security; no cross-tenant query is possible.
- **Partitioned** — one PostgreSQL partition per month per tenant avoids hot-table lock contention and allows fast archival.
- **Sensitive-field redacting** — fields marked `IsSensitive: true` in the entity definition are stored as `"[REDACTED]"` so the audit trail never contains passwords or tokens.

---

## 46.2 What Gets Logged

The framework's `AfterSaveHook` (`HookAfterSave`) fires for every entity with `Audited: true`. It writes one row per mutation:

| Op | Logged? | Notes |
|---|---|---|
| `OpCreate` | Yes | `before` is null; `after` is the new record |
| `OpUpdate` | Yes | Both `before` and `after` are stored |
| `OpDelete` / soft-delete | Yes | `after` is null (hard delete) or `after.deleted_at` set (soft delete) |
| List / FindByID | No | Reads are not audited — use query logs for read access tracking |

---

## 46.3 EntityDefinition: AuditLog

```go
// internal/platform/audit/definition.go
package audit

import "awo.so/framework/definition"

var AuditLogDef = definition.EntityDefinition{
    Name:        "audit_log",
    Label:       "Audit Log",
    Description: "Immutable record of every create, update, and delete on audited entities.",
    Module:      "Platform",
    Table:       "audit_logs",
    OrgScope:    org.ScopeLevelTenant,
    SoftDelete:  false, // never soft-delete an audit row

    Fields: []*definition.FieldDef{
        definition.Field("entity_name").
            OfType(definition.FieldTypeData).
            WithLabel("Entity").
            RequiredField().
            ImmutableField().
            SearchableField(),

        definition.Field("entity_id").
            OfType(definition.FieldTypeUUID).
            WithLabel("Record ID").
            RequiredField().
            ImmutableField(),

        definition.Field("op").
            OfType(definition.FieldTypeSelect).
            WithLabel("Operation").
            WithOptions("create", "update", "delete").
            RequiredField().
            ImmutableField(),

        definition.Field("actor_id").
            OfType(definition.FieldTypeUUID).
            WithLabel("Actor").
            RequiredField().
            ImmutableField(),

        definition.Field("before").
            OfType(definition.FieldTypeJSON).
            WithLabel("Before"),

        definition.Field("after").
            OfType(definition.FieldTypeJSON).
            WithLabel("After"),

        definition.Field("ip_address").
            OfType(definition.FieldTypeData).
            WithLabel("IP Address").
            ImmutableField(),

        definition.Field("user_agent").
            OfType(definition.FieldTypeData).
            WithLabel("User Agent").
            ImmutableField(),
    },

    Policies: []definition.PolicyDef{
        {Ops: definition.OpCreate, Fn: definition.DenyAll},   // only framework writes
        {Ops: definition.OpRead,   Fn: requireRole("auditor", "tenant_admin")},
        {Ops: definition.OpUpdate, Fn: definition.DenyAll},   // immutable
        {Ops: definition.OpDelete, Fn: definition.DenyAll},   // immutable
    },

    Audited: false, // do not audit the audit log itself
}
```

### Why `DenyAll` on Create?

Application code never calls the Audit Log API directly. The framework's internal audit hook writes rows using a privileged connection that bypasses the policy layer. Exposing a Create endpoint would allow callers to forge audit records.

---

## 46.4 The Audit Hook

The hook is registered by the framework automatically for every entity with `Audited: true`. It fires after the primary mutation succeeds (`HookAfterSave`), so a failed save produces no audit row.

```go
// framework/audit/hook.go
package audit

import (
    "context"
    "encoding/json"

    "awo.so/framework/definition"
)

// WriteAuditLog writes one row to audit_logs for the given mutation.
// Sensitive fields in Before/After are replaced with "[REDACTED]".
func WriteAuditLog(svc *Service) definition.HookFunc {
    return func(ctx context.Context, m *definition.Mutation) error {
        if !m.Def.Audited {
            return nil
        }

        before := redact(m.Def, recordToMap(m.Before))
        after  := redact(m.Def, recordToMap(m.After))

        return svc.write(ctx, AuditEntry{
            TenantID:   m.TenantID,
            EntityName: m.Def.Name,
            EntityID:   entityID(m),
            Op:         m.Op.String(),
            ActorID:    m.ActorID,
            Before:     mustJSON(before),
            After:      mustJSON(after),
            IPAddress:  ipFromCtx(ctx),
            UserAgent:  uaFromCtx(ctx),
        })
    }
}

// redact replaces sensitive field values with "[REDACTED]".
func redact(def *definition.EntityDefinition, m map[string]any) map[string]any {
    if m == nil {
        return nil
    }
    for _, f := range def.Fields {
        if f.IsSensitive {
            if _, ok := m[f.Name]; ok {
                m[f.Name] = "[REDACTED]"
            }
        }
    }
    return m
}

func mustJSON(v any) json.RawMessage {
    if v == nil {
        return nil
    }
    b, _ := json.Marshal(v)
    return b
}
```

### Why AfterSave, Not BeforeSave?

The mutation might be rejected by a database constraint or a later hook. Writing the audit log before the row exists would produce a false record. `HookAfterSave` runs inside the same transaction, so if the transaction rolls back the audit row also rolls back.

---

## 46.5 Registering the Audit Hook

The framework adds the audit hook automatically when building a `Handler` for an audited entity:

```go
// framework/api/handler.go (abbreviated)
func NewHandler(def *definition.EntityDefinition, ...) *Handler {
    h := &Handler{...}
    if def.Audited {
        h.hooks.Register(audit.AfterSaveHook(auditService))
    }
    return h
}
```

Module developers need only set `Audited: true` — no manual hook registration required.

---

## 46.6 Dual-Write Safety

The audit write shares the primary transaction:

```
BEGIN
  UPDATE invoices SET status = 'paid' WHERE id = $1   ← primary mutation
  INSERT INTO audit_logs (...) VALUES (...)            ← audit write (same tx)
COMMIT
```

If the commit fails both writes roll back together. There is no window where the invoice is updated but the audit row is missing.

---

## 46.7 Partitioning Strategy

`audit_logs` is partitioned by `(tenant_id, created_at)` using PostgreSQL range partitions on `created_at`, with one partition per calendar month:

```sql
-- audit_logs is the parent table
CREATE TABLE audit_logs (
    id          uuid          NOT NULL,
    tenant_id   uuid          NOT NULL,
    created_at  timestamptz   NOT NULL DEFAULT NOW(),
    ...
) PARTITION BY RANGE (created_at);

-- Monthly child partition (created by background job)
CREATE TABLE audit_logs_2025_01
    PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

### Why Not Partition by Tenant?

A large deployment may have thousands of tenants. Partitioning by tenant produces thousands of partitions — PostgreSQL planner overhead increases quadratically. Partitioning by month keeps the partition count bounded and predictable. RLS enforces tenant isolation within each partition.

### Partition Management

A nightly Temporal workflow (`CreateNextMonthPartition`) creates next month's partition before it is needed. A separate `ArchiveOldPartitions` workflow detaches and dumps old partitions to cold storage according to `platform.audit_retention_months` (default: 84 months / 7 years, matching Kenyan tax law).

---

## 46.8 Querying the Audit Log

### REST API

```
GET /api/audit_logs?entity_name=invoice&entity_id=<uuid>
GET /api/audit_logs?actor_id=<uuid>&order_by=created_at&dir=desc
```

Responses are paginated. Tenant admins and users with the `auditor` role can read their own tenant's records.

### Direct SQL (for reports)

```sql
-- All changes to a specific invoice in the last 30 days
SELECT
    created_at,
    op,
    actor_id,
    before -> 'status' AS status_before,
    after  -> 'status' AS status_after
FROM audit_logs
WHERE entity_name = 'invoice'
  AND entity_id   = $1
  AND created_at  > NOW() - INTERVAL '30 days'
ORDER BY created_at DESC;
```

RLS ensures `current_setting('app.current_tenant_id')` scopes the query automatically.

---

## 46.9 Migration

```sql
-- internal/platform/audit/migrations/20240101000001_create_audit_logs.up.sql

CREATE TABLE audit_logs (
    id          uuid        NOT NULL DEFAULT uuid_generate_v4(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    entity_name text        NOT NULL,
    entity_id   uuid        NOT NULL,
    op          text        NOT NULL CHECK (op IN ('create','update','delete')),
    actor_id    uuid,
    before      jsonb,
    after       jsonb,
    ip_address  text,
    user_agent  text,
    PRIMARY KEY (id, created_at)        -- partition key must be in PK
) PARTITION BY RANGE (created_at);

-- Index: look up all changes to one record
CREATE INDEX audit_entity_idx ON audit_logs (tenant_id, entity_name, entity_id, created_at DESC);

-- Index: look up all actions by one actor
CREATE INDEX audit_actor_idx ON audit_logs (tenant_id, actor_id, created_at DESC);

ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON audit_logs
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- First partition: bootstrap for current month
CREATE TABLE audit_logs_2025_01
    PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

---

## 46.10 FAQ

**Q: Can a tenant admin delete their own audit logs?**
A: No. The `OpDelete` policy is `DenyAll`. Deletion is only possible via the `ArchiveOldPartitions` workflow, which is controlled by the platform team and bounded by `platform.audit_retention_months`.

**Q: What happens if the audit write fails?**
A: The entire transaction rolls back — including the primary mutation. The user sees a 500 error. This is intentional: for audited entities, a mutation without an audit record is considered unsafe.

**Q: Are list/read operations logged?**
A: No. The framework only logs writes. If you need read access logging (e.g. for HIPAA), add a separate `BeforeValidateHook` on `OpRead` that writes to a `read_access_logs` table — but be aware of the performance cost.

**Q: How do I view diffs in the UI?**
A: The SDUI audit log page renders `before` and `after` as side-by-side JSON diff using AMIS's `diff-editor` control. Each changed key is highlighted.
