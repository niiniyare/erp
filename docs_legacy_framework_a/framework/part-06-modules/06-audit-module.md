> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module: Audit Log"
part: "Part VI — Platform Entities"
chapter: 46
section: "platform-audit-module"
related:
  - "[Chapter 43: IAM Module](./platform-iam-module.md)"
  - "[Chapter 42: Tenant & Organisation](./platform-tenant-module.md)"
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
---

# Chapter 46 — Audit Log Module

> **Primary source for:** the `audit_log` partitioned table, dual-write strategy (DB trigger + application hook), monthly partitioning for 7-year retention, the audit search API, and how module developers opt into automatic auditing.
>
> **Audience:** developers building auditable entities, compliance officers who need to understand what is captured and how, and anyone asking "who changed this record?"

---

## 46.1 What Is an Audit Log?

An audit log is a permanent, tamper-evident record of who did what to which data and when.

### Why It Exists

**Legal compliance.** The Kenya Companies Act 2015 requires businesses to maintain financial records for seven years from the end of the financial year. Without an audit log, there is no way to prove that a record has not been silently edited since it was created.

**Fraud detection.** "Who changed this invoice total from KES 48,000 to KES 52,000 two hours before payment?" The audit log answers this in milliseconds.

**Debugging.** "This customer's credit limit was 500,000 last week and is now 100,000. Who changed it?" Without an audit log, you are guessing.

**Accountability.** Users make better decisions when they know their actions are recorded. The audit log is a deterrent as much as a detective control.

### What the Audit Log Is Not

- **Not a system log.** Server logs (nginx, application logs) are operational. The audit log is a business record of data changes.
- **Not a read log.** We do not record every SELECT query — that would be terabytes per day with no compliance value. Only CREATE, UPDATE, and DELETE operations are recorded.
- **Not auditing the audit log itself.** The audit log does not audit itself (infinite regress). Audit log rows are immutable by design.

---

## 46.2 What Gets Audited

Any EntityDefinition that declares `Audited: true` automatically generates audit log entries. The framework wires the trigger and AfterHook without additional code from the module developer.

**Platform entities — all audited:**

| Entity | Why |
|---|---|
| Tenant | Status changes, plan changes — business-critical |
| User | Access changes, status changes — security-critical |
| Role / UserRole | Permission changes — security-critical |
| FeatureFlagOverride | Feature enablement decisions |
| ConfigValue | Settings changes affect all transactions |
| OrgNode | Structural changes affect reporting |

**Business entities — opt in per entity:**

| Entity | `Audited` | Why |
|---|---|---|
| Invoice | `true` | Financial record — must be auditable |
| JournalEntry | `true` | GL entries — immutable by accounting standard |
| Employee | `true` | Payroll data changes — HR compliance |
| StockMovement | `true` | Inventory accuracy — regulatory |
| SalesOrder | `true` | Revenue recognition |
| SystemConfig | `true` | Configuration changes have system-wide impact |
| UIPreference | `false` | User display settings — no compliance value |
| Notification | `false` | Ephemeral; not a business record |

---

## 46.3 Directory Layout

```
framework/platform/audit/
├── audit.go      ← init() — registers AuditLogDef
├── definition.go ← AuditLogDef
├── policy.go     ← requireRole("admin"), requireRole("auditor")
├── service.go    ← AuditService: Search(), GetHistory()
└── migrations/
    ├── 20240101000040_create_audit_log.up.sql
    └── 20240101000040_create_audit_log.down.sql
```

---

## 46.4 AuditLog EntityDefinition

```go
// framework/platform/audit/definition.go
package audit

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var AuditLogDef = definition.EntityDefinition{
    Name:     "audit_log",
    Label:    "Audit Log",
    Module:   "Platform",
    Table:    "audit_log",
    OrgScope: org.ScopeLevelTenant,

    // CRITICAL: Audited must be false on the audit log itself.
    // If Audited were true, every write to audit_log would trigger
    // another write to audit_log, which would trigger another, infinitely.
    Audited: false,

    // No soft-delete. Audit log rows are never individually deleted.
    // Entire monthly partitions are dropped when retention period expires.
    SoftDelete: false,

    Fields: []*definition.FieldDef{
        {Name: "entity_name",    Type: definition.FieldTypeData,     Label: "Entity",
            Description: "EntityDefinition.Name of the affected entity: 'invoice', 'user'"},
        {Name: "record_id",      Type: definition.FieldTypeUUID,     Label: "Record ID",
            Description: "Primary key of the affected row."},
        {Name: "operation",      Type: definition.FieldTypeSelect,   Label: "Operation",
            Options: []string{"CREATE", "UPDATE", "DELETE"}},

        // Nullable: system/workflow operations have no human actor.
        {Name: "user_id",     Type: definition.FieldTypeLink,     Label: "Actor",
            TargetEntity: "user",
            Description: "The user who triggered this change. " +
                "NULL for system-initiated changes (Temporal workflows, scheduled jobs)."},
        {Name: "workflow_id", Type: definition.FieldTypeData,     Label: "Workflow ID",
            Description: "Temporal workflow ID if change came from a workflow."},
        {Name: "request_id",  Type: definition.FieldTypeData,     Label: "Request ID",
            Description: "HTTP request ID for correlating multiple changes in one request."},
        {Name: "ip",          Type: definition.FieldTypeData,     Label: "IP Address"},
        {Name: "timestamp",   Type: definition.FieldTypeDateTime, Label: "When"},

        // Full JSON snapshots of the row before and after the change.
        // NULL for INSERT (no previous value). NULL for DELETE (no new value).
        // Sensitive fields (Sensitive:true) are stripped before storage.
        {Name: "previous_value", Type: definition.FieldTypeJSON, Label: "Before"},
        {Name: "new_value",      Type: definition.FieldTypeJSON, Label: "After"},
    },

    Policies: []definition.PolicyDef{
        // System callers (the framework trigger AfterHook, Temporal activities)
        // are the only writers. No user can create an audit log entry.
        definition.Policy(definition.OpCreate, definition.AllowSystem),

        // Nobody — not even platform admins — can update or delete audit entries.
        // This tamper-evidence is the core value of the audit log.
        // If an admin could edit the audit log, it would be worthless as evidence.
        definition.Policy(definition.OpUpdate|definition.OpDelete, definition.DenyAll),

        // Tenant admins can search their own audit log.
        definition.Policy(definition.OpRead, requireRole("admin")),
        // Auditors get read-only access (for external/internal audits).
        definition.Policy(definition.OpRead, requireRole("auditor")),

        // All others denied.
        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

---

## 46.5 Dual Write Strategy

The audit log uses two mechanisms simultaneously to capture every change. Understanding why both are needed is important.

### Mechanism 1 — PostgreSQL Trigger (Completeness)

A trigger fires on every `INSERT`, `UPDATE`, and `DELETE` on any auditable table. It runs *inside the same transaction* as the data change.

```
BEGIN TRANSACTION
    UPDATE invoices SET total = 52000 WHERE id = $1
    → trigger fires: INSERT INTO audit_log (...) VALUES (...)
COMMIT
```

If the transaction commits, the audit entry commits. If the transaction rolls back, the audit entry rolls back. The audit log is always consistent with the data.

**Why a trigger?**
- Captures ALL writes — including emergency fixes via `psql`, data migrations run directly on the DB, and any other tool that bypasses the application.
- Zero code needed in the application for the entry to be created.
- Cannot be accidentally omitted by a developer who forgets to call an audit service.

### Mechanism 2 — Application AfterHook (Context Enrichment)

The trigger creates a minimal audit entry with `entity_name`, `record_id`, `operation`, `timestamp`, `previous_value`, `new_value`. It cannot know the HTTP request context.

The application's `AfterHook` (registered by the framework on all `Audited: true` entities) runs *after* the transaction commits and enriches the audit entry:

- `user_id`: who made the change (from `ViewerContext.ActorID()`)
- `request_id`: the HTTP request ID (from the request context)
- `workflow_id`: the Temporal workflow ID (if in a workflow context)
- `ip`: the client's IP address

**Why not just the hook?**

If only the hook ran, direct SQL access (emergency maintenance, data migrations) would leave no audit trail. Combining trigger + hook gives:
- **Trigger**: completeness (nothing escapes)
- **Hook**: context (the why and who)

---

## 46.6 The Audit Trigger Function

```sql
-- Creates audit entries for ANY table that calls this trigger function.
-- Registered on each auditable table's migration.
CREATE OR REPLACE FUNCTION audit_log_trigger_fn()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (
        id,
        tenant_id,
        entity_name,
        record_id,
        operation,
        timestamp,
        previous_value,
        new_value
    ) VALUES (
        gen_random_uuid(),

        -- COALESCE because DELETE only has OLD, INSERT only has NEW.
        COALESCE(NEW.tenant_id, OLD.tenant_id),

        -- TG_TABLE_NAME: magic variable — the name of the table that fired the trigger.
        -- This is how one function serves all auditable tables.
        TG_TABLE_NAME,

        COALESCE(NEW.id, OLD.id),

        -- TG_OP: 'INSERT', 'UPDATE', or 'DELETE'.
        TG_OP,

        now(),

        -- For INSERT, there is no "before" value — previous_value is NULL.
        CASE WHEN TG_OP != 'INSERT' THEN to_jsonb(OLD) ELSE NULL END,

        -- For DELETE, there is no "after" value — new_value is NULL.
        CASE WHEN TG_OP != 'DELETE' THEN to_jsonb(NEW) ELSE NULL END
    );

    -- AFTER trigger: must return COALESCE(NEW, OLD) for the original operation to proceed.
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- SECURITY DEFINER explanation:
-- The trigger runs as its OWNER (typically the migration user / postgres),
-- not as the CALLER (the application role). This guarantees the trigger can
-- always INSERT into audit_log even if the application role does not have
-- direct INSERT permission on audit_log. Platform security-by-design.
```

### Attaching the Trigger to an Auditable Table

Module migrations attach this trigger to their entity tables:

```sql
-- In Finance module migration for invoices:
CREATE TRIGGER audit_invoices
    AFTER INSERT OR UPDATE OR DELETE ON invoices
    FOR EACH ROW EXECUTE FUNCTION audit_log_trigger_fn();
```

The framework's migration generator does this automatically when it detects `Audited: true` on the EntityDefinition. Module developers don't write this trigger manually.

---

## 46.7 Monthly Partitioning

The `audit_log` table is partitioned by month. This is the most important operational decision in the module's design.

### Why Partition?

A busy ERP deployment might write 100,000 audit entries per day. After 7 years, that is 255 million rows. Queries that touch the whole table (compliance reports, "all changes by user X in 2021") become table scans on a 255M-row table without partitioning.

With partitioning:
- Each month is a separate table (partition).
- PostgreSQL prunes irrelevant partitions at query time — a query for March 2024 only scans `audit_log_2024_03`.
- Expiring old data is instant: `DROP TABLE audit_log_2017_01` is a metadata operation (milliseconds), not a `DELETE WHERE timestamp < ...` (hours scanning billions of rows).

### Partition Schema

```sql
-- The parent table. All queries go here; PostgreSQL routes to the right partition.
CREATE TABLE audit_log (
    id             uuid        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id      uuid        NOT NULL REFERENCES tenants(id),
    entity_name    text        NOT NULL,
    record_id      uuid        NOT NULL,
    operation      text        NOT NULL CHECK (operation IN ('INSERT','UPDATE','DELETE')),
    user_id        uuid,       -- NULL for system operations
    workflow_id    text,
    request_id     text,
    ip             text,
    timestamp      timestamptz NOT NULL DEFAULT now(),
    previous_value jsonb,
    new_value      jsonb,

    -- timestamp is part of the PK because partitioned tables require the
    -- partition key to be in the primary key.
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_log_tenant_isolation ON audit_log
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

### Partition Creation

A Temporal schedule creates next month's partition on the 1st of each month:

```sql
-- Example: created on 2024-01-01 for February 2024
CREATE TABLE audit_log_2024_02 PARTITION OF audit_log
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
```

### Partition Expiry

Seven years of monthly partitions = 84 partitions maximum. The same schedule drops the partition from 7 years ago:

```sql
-- Drop January 2017 on January 1, 2024 (7 years later)
DROP TABLE IF EXISTS audit_log_2017_01;
```

The framework's `cmd/awo partition-rotate` command manages this automatically.

---

## 46.8 Indexes

Two indexes cover the two most common audit searches:

```sql
-- "Show me all changes to invoice INV-2024-0451"
-- (Most common: compliance review of a specific record's history)
CREATE INDEX audit_log_entity_record_idx
    ON audit_log (tenant_id, entity_name, record_id, timestamp DESC);

-- "Show me everything user john@acme.co.ke did this week"
-- (Security investigation: suspicious activity by a user)
CREATE INDEX audit_log_user_time_idx
    ON audit_log (tenant_id, user_id, timestamp DESC);
```

Both indexes include `tenant_id` as the first column. This is required — without it, a query for Tenant A's records would scan Tenant B's rows too before RLS filters them.

`timestamp DESC` on both indexes means "newest first" queries (the most common direction for human-facing audit UIs) use the index without sorting.

---

## 46.9 The Audit Search API

The framework generates a standard list endpoint for `AuditLogDef`:

```
GET /api/v1/audit-log
    ?entity_name=invoice
    &record_id=550e8400-e29b-41d4-a716-446655440000
    &user_id=...
    &from=2024-01-01T00:00:00Z
    &to=2024-03-31T23:59:59Z
    &limit=50
    &cursor=...
```

**RLS applies.** Searching the audit log only returns entries for the current tenant. A tenant admin cannot query another tenant's audit history even with a crafted API call.

Sample response:

```json
{
  "data": [
    {
      "id": "uuid",
      "entity_name": "invoice",
      "record_id": "550e8400-...",
      "operation": "UPDATE",
      "user_id": "usr-john-uuid",
      "timestamp": "2024-03-12T14:23:07Z",
      "ip": "41.90.64.1",
      "request_id": "req-abc123",
      "previous_value": {
        "id": "550e8400-...",
        "total": "48000.0000",
        "status": "DRAFT"
      },
      "new_value": {
        "id": "550e8400-...",
        "total": "52000.0000",
        "status": "DRAFT"
      }
    }
  ],
  "meta": {
    "total": 1,
    "cursor": null,
    "has_next": false
  }
}
```

---

## 46.10 How Module Developers Opt In

Opting in is a single field on the EntityDefinition:

```go
// In Finance module — InvoiceDef:
var InvoiceDef = definition.EntityDefinition{
    Name:     "invoice",
    Label:    "Invoice",
    Module:   "Finance",
    Table:    "invoices",
    OrgScope: org.ScopeLevelUnit,
    Audited:  true,   // ← THIS IS ALL THAT IS NEEDED
    // ...
}
```

The framework then:
1. Generates `CREATE TRIGGER audit_invoices AFTER INSERT OR UPDATE OR DELETE ON invoices FOR EACH ROW EXECUTE FUNCTION audit_log_trigger_fn();` in the migration.
2. Registers an AfterHook that enriches the trigger-created audit entry with user_id, request_id, workflow_id, ip.
3. Strips any fields marked `Sensitive: true` from `previous_value` and `new_value` before writing (passwords, API keys never appear in the audit log).

**That's it.** No audit service import. No manual `auditService.Log(ctx, ...)` calls. Zero lines of module code.

---

## 46.11 Sensitive Fields in Audit Entries

Fields marked `Sensitive: true` on the EntityDefinition are automatically redacted from audit log JSON:

```go
// UserDef has:
{Name: "password_hash", Type: definition.FieldTypeData, Sensitive: true}
{Name: "mfa_secret",    Type: definition.FieldTypeData, Sensitive: true}
```

In the audit log entry for a user update, these fields appear as:

```json
{
  "previous_value": {
    "email": "old@example.com",
    "name": "John Old",
    "password_hash": "[REDACTED]",
    "mfa_secret": "[REDACTED]"
  },
  "new_value": {
    "email": "new@example.com",
    "name": "John New",
    "password_hash": "[REDACTED]",
    "mfa_secret": "[REDACTED]"
  }
}
```

The redaction happens in the AfterHook before the enrichment update. The trigger-created entry from PostgreSQL contains the raw values (PostgreSQL doesn't know about Go's `Sensitive` flag). The AfterHook replaces them before the entry is visible to any application query.

---

## 46.12 Frequently Asked Questions

**"Can a platform admin delete audit log entries?"**

No. The `OpUpdate|OpDelete` policy is `DenyAll` — not `requireRole("platform_admin")`. No API path allows deletion. Emergency deletion requires direct database access (which itself is audited at the infrastructure layer) and is a two-person operation per security policy.

**"What if the AfterHook fails to enrich the entry?"**

The original data write has already committed. The trigger-created audit row exists. If the AfterHook fails, the audit entry is preserved with NULL user_id/request_id — incomplete but present. We never sacrifice audit coverage for enrichment. The hook logs its failure for investigation.

**"Does auditing add latency to every write?"**

The trigger adds approximately 1ms to every write operation on an auditable table (one extra INSERT into a local partition). At ERP scale (hundreds of transactions per minute, not millions per second), this is imperceptible. Partition pruning ensures the INSERT always targets a small, recently-created partition — not a 255M-row table scan.

**"Can we audit read operations for GDPR or similar?"**

Read auditing is not implemented and intentionally not planned for the core framework. Read logging at scale is expensive (every SELECT on a customer record generates an audit entry) and noisy (hard to distinguish meaningful access from background queries). For GDPR access requests, the framework's `export-tenant-data` command generates the required data subject report without read-level logging.

**"What about the previous_value/new_value size?"**

Large JSONB rows (an entity with 100 fields and large text fields) produce large audit entries. The `audit_log` table does not have a size limit per entry — PostgreSQL handles arbitrary JSONB. For truly large entities (document attachments stored in entity rows), consider storing the attachment in S3 and only the metadata in the DB row. The audit log then captures metadata changes, not binary content changes.
