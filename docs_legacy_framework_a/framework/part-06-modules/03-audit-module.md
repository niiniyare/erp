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
  - "[Chapter 41b: Platform Module Overview](./implementing-platform-modules.md)"
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-migrations.md)"
---

# Chapter 46: Platform Module — Audit Log

Every business system that handles money, people, or sensitive records must answer the question: **"Who changed that, and when?"** The Audit Log module exists to make that question always answerable — automatically, completely, and tamper-evidently. This chapter explains what the audit log is, why it was designed the way it was, and how every moving part fits together.

---

## 1. What an Audit Log Is — For Stakeholders

Imagine a register at a bank branch. Every time a teller processes a transaction, they stamp the register: the time, their employee number, the customer, and the amount. Nobody can erase an entry from the register. If a dispute arises, the manager pulls out the register and finds the exact entry.

The Awo audit log is that register — but for every data change in the entire ERP system.

**A real example:** On 12 March 2024 at 14:23 EAT, user `john@acme.co.ke` changed invoice `INV-2024-0451`'s total from KES 48,000 to KES 52,000. The audit log records exactly that — the user, the timestamp, the entity, the old value, and the new value — in a format that cannot be altered after the fact.

### Why does this matter?

| Reason | Explanation |
|--------|-------------|
| **Legal compliance** | The Kenya Companies Act 2015 (Section 721) requires financial records to be kept for 7 years from the financial year end. An audit log is direct evidence of compliance. |
| **Fraud detection** | When an employee inflates a supplier invoice, the audit log shows exactly when it changed and which user account made the change. |
| **Debugging** | A support request arrives: "My invoice total is wrong." The audit log answers: "It was changed by the accounts manager at 2pm on Tuesday." |
| **Accountability** | Knowing that all actions are recorded changes behaviour. Users are more careful about accidental deletions or unauthorised changes. |
| **Dispute resolution** | A customer claims they were never informed of a price change. The audit log shows every change to their contract price and when it was made. |

> **Key principle:** The audit log is append-only. Entries are written once and never modified or deleted through normal operations. This is the tamper-evidence guarantee that makes it legally admissible.

---

## 2. What Gets Audited — For Stakeholders

Not everything needs to be audited. Here is how the decision is made:

### Automatically audited

- **Platform entities**: Users, roles, permissions — anything that touches access control. If someone is given admin rights, that's in the audit log.
- **Finance entities**: Invoices, journal entries, payments, credit notes. Any entity that involves money is audited by default.
- **Custom entities that opt in**: Module developers set `Audited: true` on their `EntityDefinition`. The framework handles everything else.

### Explicitly NOT audited

- **Sessions**: Sessions themselves are the mechanism of tracking who is acting. Auditing sessions would be circular.
- **Read operations**: Reading data does not change it. If the framework logged every time a user viewed an invoice, a busy company would generate terabytes of logs per day with zero actionable information. We log writes, not reads.
- **Temporary and derived data**: Caches, computed summaries, background task state — these change constantly and carry no business significance.
- **The audit log itself**: If the audit log audited its own writes, every audit log entry would trigger another audit log entry in an infinite loop. This is the one entity where `Audited` is always `false`.

---

## 3. AuditLog EntityDefinition — For Developers

Here is the complete definition of the audit log entity. Every decision has a reason.

```go
// framework/platform/audit/definition.go

package audit

import (
    "awo.so/framework/definition"
    "awo.so/framework/privacy"
)

var AuditLogDef = definition.EntityDefinition{
    Name:        "audit_log",
    Label:       "Audit Log",
    Table:       "audit_log",
    Module:      "platform",
    OrgScope:    definition.ScopeLevelTenant,

    // The audit log never audits itself — that would be infinite recursion.
    // Every other platform entity sets Audited: true. This one must stay false.
    Audited:    false,

    // Audit log rows are NEVER soft-deleted individually.
    // Retention is managed by dropping entire monthly partitions.
    // See Section 6 for the partitioning strategy.
    SoftDelete: false,

    Fields: []definition.Field{
        {Name: "id",             Type: "uuid",      Required: true, SystemManaged: true},
        {Name: "tenant_id",      Type: "uuid",      Required: true, SystemManaged: true},
        {Name: "entity_name",    Type: "text",      Required: true},  // e.g. "invoice"
        {Name: "record_id",      Type: "uuid",      Required: true},  // the PK of the changed row
        {Name: "operation",      Type: "text",      Required: true},  // CREATE, UPDATE, DELETE
        {Name: "previous_value", Type: "jsonb",     Required: false}, // full row before change
        {Name: "new_value",      Type: "jsonb",     Required: false}, // full row after change
        {Name: "changed_by",     Type: "uuid",      Required: false}, // user_id (null for system ops)
        {Name: "request_id",     Type: "text",      Required: false}, // correlation ID from HTTP request
        {Name: "workflow_id",    Type: "text",      Required: false}, // Temporal workflow ID if applicable
        {Name: "ip_address",     Type: "text",      Required: false}, // source IP from request context
        {Name: "timestamp",      Type: "timestamptz", Required: true, SystemManaged: true},
    },

    Policies: []privacy.Policy{
        // System callers only can create audit log entries.
        // The audit log is written by framework hooks and DB triggers —
        // never directly by user-facing API calls.
        // If a user sends POST /api/v1/audit-log, they receive 403.
        {Role: privacy.RoleSystem,   Op: privacy.OpCreate},

        // Admins and auditors can read audit log entries for their tenant.
        // RLS ensures cross-tenant reads are impossible at the DB level.
        {Role: privacy.RoleAdmin,    Op: privacy.OpRead},
        {Role: "auditor",            Op: privacy.OpRead},

        // NOBODY can write or delete audit log entries through the API.
        // Not even platform admins. This is the tamper-evidence guarantee.
        // Emergency deletion requires direct DB access + a second human approver.
        // (See Section 10: Common Questions)
    },
}
```

### Why these policy decisions?

**`OpCreate` for system only:** The audit log is a write-once system record. It would be meaningless if users could insert their own entries ("I swear I didn't change that, here's my own audit entry saying so"). The framework's hooks write entries; nobody else does.

**`OpWrite | OpDelete` denied for everyone:** Auditability is only meaningful if the records cannot be altered. If an admin could delete an audit entry covering their tracks, the audit log would be worthless. The framework makes this structurally impossible through the policy layer, not just through UI constraints.

**Auditor role for reads:** A compliance officer or external auditor needs to search the audit trail without having full admin access. The `auditor` role grants read-only access to the audit log while having no write access to any business entity.

---

## 4. Dual Write Strategy — Why Two Mechanisms Work Together

The audit log uses two complementary approaches. Understanding why both exist requires understanding what each one can and cannot see.

### Mechanism 1: PostgreSQL Database Trigger

A trigger `audit_log_trigger_fn` fires on EVERY INSERT, UPDATE, and DELETE on any audited table — **at the database level**.

**What this gives us:**
- Completeness. Even if a developer runs `psql` directly to fix a typo, the trigger fires. Emergency data migrations, ORM bugs that bypass application logic, background scripts — all captured.
- Atomicity. The trigger runs inside the same database transaction as the original write. If the original write rolls back, the audit row rolls back too.

**What this cannot give us:**
- Who the user was. The DB trigger sees the database role (e.g. `erp_app`), not the application-level user (e.g. `john@acme.co.ke`).
- Why the change happened. No request ID, no workflow context, no IP address.

### Mechanism 2: Application AfterHook

The framework's AfterHook runs after the database transaction commits for any entity with `Audited: true`. This hook has access to the full request context.

**What this gives us:**
- `user_id`: who triggered the operation (from the JWT-authenticated session)
- `request_id`: the HTTP request correlation ID (for distributed tracing)
- `workflow_id`: if the write was triggered by a Temporal workflow, its ID
- `ip_address`: the client IP from the HTTP request

**What this cannot give us:**
- Coverage for direct SQL writes. If someone bypasses the application entirely, the AfterHook never fires.

### How they coordinate

```
User action → API handler → Service → Repository (writes to DB)
                                              │
                               ┌──────────────▼──────────────┐
                               │  PostgreSQL transaction       │
                               │  1. INSERT/UPDATE/DELETE      │
                               │     the business row          │
                               │  2. Trigger fires immediately │
                               │     → inserts audit_log row   │
                               │     (with entity, record_id,  │
                               │      operation, before/after) │
                               └──────────────┬──────────────┘
                                              │ COMMIT
                                              ▼
                               AfterHook fires (post-commit)
                               Has request context: user_id,
                               request_id, ip_address
                               → UPDATE audit_log SET
                                   changed_by   = $user_id,
                                   request_id   = $request_id,
                                   ip_address   = $ip_address
                                 WHERE record_id = X
                                   AND entity_name = Y
                                   AND timestamp > now() - interval '5 seconds'
                                   AND changed_by IS NULL
```

The trigger provides **completeness**. The AfterHook provides **context**. Together, every audit entry is complete and attributed.

> **What if the AfterHook fails?** The original write has already committed. The trigger row (without application context) is preserved in the audit log. Audit *coverage* is never sacrificed — you always know *what* changed. You might occasionally not know *who* (the `changed_by` field stays null), but the record still exists.

---

## 5. The Trigger Function — Full SQL

```sql
-- db/migrations/000005_audit_log.up.sql

-- This function is called by triggers on every audited table.
-- SECURITY DEFINER means it runs as the function OWNER (typically 'erp_migrations'),
-- not as the calling role ('erp_app'). This ensures the function can always
-- write to audit_log even if the application role has no direct INSERT permission
-- on the audit_log table.
CREATE OR REPLACE FUNCTION audit_log_trigger_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    INSERT INTO audit_log (
        id,
        tenant_id,
        entity_name,
        record_id,
        operation,
        previous_value,
        new_value,
        timestamp
    )
    VALUES (
        gen_random_uuid(),

        -- COALESCE handles both INSERT (only NEW exists) and DELETE (only OLD exists).
        -- UPDATE has both; we prefer NEW.tenant_id.
        COALESCE(NEW.tenant_id, OLD.tenant_id),

        -- TG_TABLE_NAME is a PostgreSQL magic variable: the name of the table
        -- that fired the trigger. This is how one function serves all audited tables.
        TG_TABLE_NAME,

        -- The primary key of the affected row.
        -- COALESCE handles DELETE (only OLD.id) and INSERT (only NEW.id).
        COALESCE(NEW.id, OLD.id),

        -- TG_OP is another magic variable: 'INSERT', 'UPDATE', or 'DELETE'.
        TG_OP,

        -- to_jsonb() converts the full row to a JSON object.
        -- WHY the full row: we want the complete before-state, not just changed columns.
        -- This makes audit searches simple — no need to reconstruct state by replaying diffs.
        -- WHY NULL for INSERT: there is no "before" state for a new record.
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE to_jsonb(OLD) END,

        -- WHY NULL for DELETE: the row is gone. The previous_value captures the final state.
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE to_jsonb(NEW) END,

        -- Use clock_timestamp(), not now().
        -- now() returns the transaction start time — all rows in one transaction
        -- would have the same timestamp, making sequencing ambiguous.
        -- clock_timestamp() gives the actual wall-clock time of this specific statement.
        clock_timestamp()
    );

    -- Triggers must return NEW for INSERT/UPDATE, OLD for DELETE.
    -- Returning NULL would cancel the original operation — never do that here.
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$;

-- Apply the trigger to the invoices table (example).
-- The framework's migration generator creates this statement for every
-- EntityDefinition where Audited: true.
CREATE TRIGGER audit_log_invoice_trigger
    AFTER INSERT OR UPDATE OR DELETE
    ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION audit_log_trigger_fn();
```

### Attaching the trigger to new tables

Whenever a module developer adds a new entity with `Audited: true`, the framework migration generator produces a `CREATE TRIGGER` statement automatically. Developers do not write trigger SQL — they just set the flag.

---

## 6. Monthly Partitioning Strategy

A large Kenyan ERP deployment might have 500 active businesses, each processing thousands of transactions per day. After 7 years, the audit log could contain 2–3 billion rows. Without partitioning, a single table at that scale becomes operationally painful — even with indexes, full-table operations (VACUUM, backups, index rebuilds) become slow.

### How partitioning works

The `audit_log` table is declared as a **range partition** on the `timestamp` column. PostgreSQL automatically routes each row to the correct sub-table based on its timestamp.

```sql
-- The parent table — this is what the application queries.
-- PostgreSQL routes reads and writes to the correct partition automatically.
CREATE TABLE audit_log (
    id             UUID            NOT NULL,
    tenant_id      UUID            NOT NULL,
    entity_name    TEXT            NOT NULL,
    record_id      UUID            NOT NULL,
    operation      TEXT            NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    previous_value JSONB,
    new_value      JSONB,
    changed_by     UUID,           -- filled by AfterHook
    request_id     TEXT,           -- filled by AfterHook
    workflow_id    TEXT,           -- filled by AfterHook
    ip_address     TEXT,           -- filled by AfterHook
    timestamp      TIMESTAMPTZ     NOT NULL DEFAULT clock_timestamp()
) PARTITION BY RANGE (timestamp);

-- Each month is a separate physical table.
-- Queries with a date range filter only touch relevant partitions (partition pruning).
CREATE TABLE audit_log_2024_01
    PARTITION OF audit_log
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE audit_log_2024_02
    PARTITION OF audit_log
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- ... and so on.
```

### Creating next month's partition

A Temporal scheduled workflow runs on the first day of each month and creates the next month's partition:

```sql
-- Run by the partition-management Temporal workflow
CREATE TABLE IF NOT EXISTS audit_log_2025_01
    PARTITION OF audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Indexes are created on the partition, not the parent table.
-- This keeps index sizes small (one month of data, not all-time).
CREATE INDEX IF NOT EXISTS audit_log_2025_01_entity_record_idx
    ON audit_log_2025_01 (tenant_id, entity_name, record_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS audit_log_2025_01_user_time_idx
    ON audit_log_2025_01 (tenant_id, changed_by, timestamp DESC);
```

### Dropping expired partitions

When a partition ages beyond the 7-year retention period, dropping it is **instantaneous** — a metadata-only operation. Compare this to the alternative:

```sql
-- WRONG approach — takes hours on a large table, locks the table:
DELETE FROM audit_log WHERE timestamp < '2017-01-01';

-- RIGHT approach — instant, no lock contention:
DROP TABLE audit_log_2017_01;
```

> **Why 7 years?** The Kenya Companies Act 2015, Section 721, requires that companies keep financial records for at least 7 years from the end of the financial year to which they relate. The Awo audit log matches this retention period exactly.

---

## 7. Indexes — What and Why

Indexes on the audit log are created per-partition to keep them small and fast. Each partition has two indexes:

### Entity-record index

```sql
CREATE INDEX audit_log_{month}_entity_record_idx
    ON audit_log_{month} (tenant_id, entity_name, record_id, timestamp DESC);
```

**Purpose:** Answers the most common audit query: "Show me all changes to this specific invoice."

```sql
-- "Show me everything that happened to invoice abc-123 in Q1 2024"
SELECT * FROM audit_log
WHERE tenant_id    = 'tenant-uuid'
  AND entity_name  = 'invoice'
  AND record_id    = 'abc-123-uuid'
  AND timestamp   >= '2024-01-01'
  AND timestamp    < '2024-04-01'
ORDER BY timestamp DESC;
```

The `(tenant_id, entity_name, record_id)` prefix is a perfect equality match. The `timestamp DESC` ordering matches the common "most recent first" use case.

### User-time index

```sql
CREATE INDEX audit_log_{month}_user_time_idx
    ON audit_log_{month} (tenant_id, changed_by, timestamp DESC);
```

**Purpose:** Answers compliance and investigation queries: "Show me everything this user did between Monday and Friday."

```sql
-- "What did john@acme.co.ke do last week?"
SELECT * FROM audit_log
WHERE tenant_id  = 'tenant-uuid'
  AND changed_by = 'john-user-uuid'
  AND timestamp >= now() - interval '7 days'
ORDER BY timestamp DESC;
```

### Why no full-text search index?

The `previous_value` and `new_value` columns are JSONB. Full-text search over JSONB requires a GIN index, which is large and slow to maintain on a high-write table. The two indexes above cover the vast majority of audit queries. GIN can be added per-deployment by any client whose compliance requirements demand keyword search across audit data.

---

## 8. The Audit Search API

```
GET /api/v1/audit-log
    ?entity_name=invoice
    &record_id=abc-123-uuid
    &from=2024-01-01
    &to=2024-03-31
    &changed_by=john-user-uuid   (optional)
    &limit=100
    &cursor=                     (pagination cursor)
```

All parameters are optional except that at least one of `entity_name+record_id` or `changed_by` must be provided — open-ended audit log scans are not permitted (they would be too expensive).

**Row Level Security applies:** The RLS policy on `audit_log` ensures that a query can only return rows where `tenant_id = current_setting('app.tenant_id')`. Cross-tenant audit reads are impossible at the database level, not just at the application level.

### Sample response

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "7f3a1b2c-...",
        "entity_name": "invoice",
        "record_id": "abc-123-uuid",
        "operation": "UPDATE",
        "previous_value": {
          "id": "abc-123-uuid",
          "total": "48000.00",
          "status": "draft",
          "updated_at": "2024-03-11T09:15:00Z"
        },
        "new_value": {
          "id": "abc-123-uuid",
          "total": "52000.00",
          "status": "draft",
          "updated_at": "2024-03-12T14:23:07Z"
        },
        "changed_by": "john-user-uuid",
        "changed_by_email": "john@acme.co.ke",
        "request_id": "req_9fXk2m",
        "ip_address": "41.89.12.44",
        "timestamp": "2024-03-12T14:23:07+03:00"
      }
    ],
    "cursor": "eyJ0aW1lc3RhbXAiOi4uLn0=",
    "has_more": false
  }
}
```

**Field guide:**
- `previous_value`: The complete row *before* the change. Null for CREATE operations.
- `new_value`: The complete row *after* the change. Null for DELETE operations.
- `changed_by_email`: Joined at query time from the users table. Not stored in the audit row itself — if the user is later deleted, the UUID is preserved even if the email is gone.
- `cursor`: An opaque pagination token. Pass it as the `cursor` parameter to get the next page.

---

## 9. How Module Developers Opt In

Enabling auditing for a new entity requires exactly one line:

```go
var InvoiceDef = definition.EntityDefinition{
    Name:       "invoice",
    Label:      "Invoice",
    Table:      "invoices",
    Module:     "finance",
    OrgScope:   definition.ScopeLevelTenant,
    Audited:    true,           // <-- this is all that is required
    SoftDelete: true,
    Fields: []definition.Field{
        {Name: "id",              Type: "uuid",         Required: true, SystemManaged: true},
        {Name: "invoice_number",  Type: "text",         Required: true},
        {Name: "total",           Type: "numeric(20,4)", Required: true},
        {Name: "status",          Type: "text",         Required: true},
        // ...
    },
}
```

The framework migration generator detects `Audited: true` and automatically:
1. Creates the `CREATE TRIGGER` statement for this table
2. Creates the first month's partition indexes when the table is first populated
3. Registers the entity in the partition management workflow's watched list

### What an audit entry looks like for a JournalEntry update

When an accountant updates a journal entry amount, here is the audit log entry the system produces:

```json
{
  "entity_name":    "journal_entry",
  "record_id":      "je-uuid-5678",
  "operation":      "UPDATE",
  "previous_value": {
    "id":          "je-uuid-5678",
    "amount":      "15000.00",
    "description": "Office supplies - March",
    "status":      "draft",
    "created_by":  "sarah-uuid",
    "posted_at":   null
  },
  "new_value": {
    "id":          "je-uuid-5678",
    "amount":      "17500.00",
    "description": "Office supplies - March (corrected)",
    "status":      "draft",
    "created_by":  "sarah-uuid",
    "posted_at":   null
  },
  "changed_by":  "sarah-uuid",
  "timestamp":   "2024-03-15T10:44:12+03:00"
}
```

### Sensitive fields

If an entity has fields marked `Sensitive: true`, those field values are **never** written to `previous_value` or `new_value`. The field key appears in the JSON but its value is replaced with `"[REDACTED]"`:

```json
{
  "previous_value": {
    "id": "emp-uuid",
    "name": "Alice Kamau",
    "salary": "[REDACTED]",
    "national_id": "[REDACTED]"
  }
}
```

The audit log records *that* the field changed, and *when*, but not *what it changed to*. This protects payroll and identity data from appearing in audit logs that might be seen by a broader set of auditors.

---

## 10. Common Questions

### "Can an admin delete audit log entries?"

**No.** The framework's policy layer denies `OpDelete` on `audit_log` for all roles, including `RoleAdmin` and `RolePlatformAdmin`. The API will return `403 Forbidden`.

Emergency deletion (e.g. a court order requiring data removal, or accidental ingestion of PII) requires:
1. Direct database access (bypassing the application entirely)
2. A second human approver on the operation
3. A separate record of what was deleted and why (stored outside the audit log)

This is by design. If admins could delete audit entries, the audit log would not be audit-worthy.

### "What if an audit log write fails?"

The dual-write design ensures graceful degradation:

- The DB trigger runs inside the original transaction. If the trigger write fails, the original operation rolls back too. You'll never have a business change without an audit entry.
- The AfterHook (application context enrichment) runs after commit, in a separate operation. If it fails, the trigger row already exists with `changed_by = NULL`. You lose the attribution context but not the fact of the change.

Audit *coverage* is never sacrificed for application context enrichment.

### "Does auditing slow down writes?"

The trigger adds approximately 1–2 milliseconds to each write operation (one extra `INSERT` into the audit table). At ERP scale — thousands of transactions per hour, not millions per second — this overhead is negligible.

For extremely high-throughput entities (e.g. IoT meter readings at a large forecourt network), consider whether full auditing is required or whether a summary log (daily reconciliation snapshots) would suffice. The `Audited` flag can be set to `false` for such entities with an explicit architectural note explaining the decision.

### "How do I search for changes by value — e.g. 'find all invoices changed from draft to approved'?"

```
GET /api/v1/audit-log
    ?entity_name=invoice
    &operation=UPDATE
    &value_path=status
    &value_before=draft
    &value_after=approved
    &from=2024-01-01
```

The API translates this to a JSONB containment query:
```sql
WHERE previous_value->>'status' = 'draft'
  AND new_value->>'status'      = 'approved'
```

This uses the JSONB index for good performance within a partition. Cross-partition value searches (no date range) are not recommended and will be rejected with a clear error message asking you to provide a date range.

---

## Summary

The Audit Log module achieves tamper-evident, legally-compliant record-keeping through:

1. **Dual write**: DB triggers for completeness, AfterHooks for context
2. **Append-only policy**: no Update or Delete via any API, for any role
3. **Monthly partitioning**: keeps the table manageable across 7-year retention
4. **Opt-in simplicity**: one flag (`Audited: true`) activates everything
5. **Sensitive field redaction**: audit coverage without leaking PII into broader-access logs

The result is a system where every stakeholder — legal, finance, operations, and security — can always answer the question: "Who changed that, and when?"
