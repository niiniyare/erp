---
title: "Chapter 45: PostgreSQL Operations"
part: "Part VIII — Deployment and Operations"
chapter: 45
section: "45-postgresql-operations"
related:
  - "[Chapter 11: Database Migrations](../part-02-entity-system/11-database-migrations.md)"
  - "[Chapter 14: Multi-Tenancy Middleware](../part-03-api/14-multitenancy-middleware.md)"
---

# Chapter 45: PostgreSQL Operations

PostgreSQL is the only persistent datastore for tenant business data. Awo uses a **shared schema, shared database** model: all tenants share the same tables, with Row-Level Security (RLS) enforcing isolation via the `tenant_id` column. This chapter covers the RLS architecture, connection pooling with PgBouncer, backup and point-in-time recovery, and performance optimisation.

---

## 45.1. Shared Schema with Row-Level Security

### 45.1.1. Design Rationale

Awo uses a single `public` schema shared by all tenants. Every tenant-scoped table has a `tenant_id uuid NOT NULL` column. RLS policies enforce that queries only return rows matching the session's active `current_tenant_id()`.

**Why shared schema over schema-per-tenant?**

| Concern | Shared schema + RLS | Schema-per-tenant |
|---|---|---|
| Provisioning speed | Instant — no DDL, just seed rows | Seconds to minutes — `CREATE SCHEMA`, run migrations |
| Migration complexity | One migration file; runs once | Must run migrations across N schemas sequentially |
| Connection pooling | Single pool; PgBouncer routes all tenants | Schema must be set per connection; complicates pooling |
| Cross-tenant platform queries | Simple — no schema prefix needed | Requires dynamic SQL or `pg_catalog` iteration |
| Tenant count scaling | No PostgreSQL catalog growth | `pg_class` grows ~100 entries per tenant |
| Isolation strength | Database-enforced via RLS | OS-level via schema name; similar strength |
| Per-tenant backup (surgical) | `pg_dump --where tenant_id=?` or export script | `pg_dump --schema=tenant_x` |

The shared schema model is operationally simpler at scale. The security guarantee is equivalent: RLS with `FORCE ROW LEVEL SECURITY` and a least-privilege app role provides the same isolation as schema separation, enforced at the database level.

### 45.1.2. Tenant Session Variable

RLS policies use a session-local variable set by the `set_tenant_context()` stored procedure:

```sql
-- Set at the start of every transaction (called by SetTenantContextFromCtx)
SELECT set_tenant_context('a1b2c3d4-...'::uuid);

-- Helper used in RLS policies
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid LANGUAGE sql STABLE AS $$
    SELECT NULLIF(current_setting('app.current_tenant_id', TRUE), '')::uuid;
$$;
```

The `is_local = TRUE` argument to `set_config` inside `set_tenant_context` makes the setting transaction-scoped. On `COMMIT` or `ROLLBACK`, it resets to empty — preventing any cross-request leakage through connection reuse.

### 45.1.3. RLS Policy Pattern

Every tenant-scoped table has this policy:

```sql
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON invoices
    FOR ALL
    USING (tenant_id = current_tenant_id());
```

`FORCE ROW LEVEL SECURITY` ensures the policy applies even to connections using the table-owner role. The application connects as `awo_app` (not the owner) for defence in depth.

### 45.1.4. Global Tables — No RLS

Reference tables shared across all tenants have no `tenant_id` and no RLS:

```sql
-- These tables have no RLS — readable by all authenticated connections
timezones, currencies, countries, paye_bands, nhif_brackets, nssf_tiers
```

The `tenants` table itself lives in `public` without per-row RLS (the middleware reads it before any tenant context is set), but platform-level access controls limit which operations the `awo_app` role can perform on it.

### 45.1.5. Enforcing tenant_id on INSERT

The application always sets `tenant_id` explicitly from context. As a belt-and-suspenders measure, a trigger ensures it cannot be omitted:

```sql
CREATE OR REPLACE FUNCTION enforce_tenant_id()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_id IS NULL THEN
        NEW.tenant_id := current_tenant_id();
    END IF;
    IF NEW.tenant_id IS NULL THEN
        RAISE EXCEPTION 'tenant_id required but no tenant context is set'
            USING ERRCODE = 'P0001';
    END IF;
    IF NEW.tenant_id != current_tenant_id() THEN
        RAISE EXCEPTION 'tenant_id mismatch: cannot insert row for a different tenant'
            USING ERRCODE = 'P0001';
    END IF;
    RETURN NEW;
END;
$$;
```

This prevents a bug where a service passes the wrong `tenant_id` to an INSERT — the trigger rejects it before the row reaches the table.

### 45.1.6. Platform Admin Queries (Cross-Tenant)

Platform operations (billing, health monitoring, usage analytics) must query all tenants. These run under a separate `awo_platform` role that bypasses RLS (`BYPASSRLS`):

```sql
CREATE ROLE awo_platform WITH LOGIN BYPASSRLS;
```

This role is used exclusively by platform admin processes (billing service, usage aggregation worker). It is **never** used for request-scoped queries from the API server. The API server connects as `awo_app` which has RLS enforced.

---

## 45.2. Connection Pooling with PgBouncer

### 45.2.1. Why PgBouncer?

Each PostgreSQL connection is a separate OS process (~5MB memory). At 1,000 concurrent tenant requests, raw connections would consume 5GB just for backend processes. PgBouncer multiplexes many lightweight application connections onto a small pool of real PostgreSQL connections.

### 45.2.2. Transaction Mode Is Required

PgBouncer must run in **transaction pool mode**:

```ini
[pgbouncer]
pool_mode = transaction
server_reset_query = DISCARD ALL
```

**Why transaction mode?** The RLS session variable is set transaction-locally (`is_local = TRUE` in `set_config`). On `COMMIT`, it resets automatically. This means:
- A connection returned to the pool after a transaction has no residual tenant context
- The next transaction that acquires the connection starts clean
- There is no need for `SET LOCAL` cleanup on connection return

In session mode, the variable would persist across transactions on the same connection, causing tenant context leakage.

**`DISCARD ALL`** in `server_reset_query` is a belt-and-suspenders measure: it resets any residual session state (prepared statements, temp tables, settings) when a connection is returned to the pool after a transaction.

### 45.2.3. PgBouncer Configuration

```ini
[databases]
awo = host=postgres-primary port=5432 dbname=awo

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 5432
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 50
server_reset_query = DISCARD ALL
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt
```

### 45.2.4. Read Replica Routing

Reporting queries (trial balance, P&L, stock valuation) route to the read replica via a separate pool:

```go
// Read replica pool — stale by ~100ms, acceptable for reports
func (r *GLRepo) TrialBalance(ctx context.Context, input TrialBalanceInput) (Report, error) {
    db := r.dbPool.ReadOnly()
    // Must also set tenant context on the read replica connection
    db.SetTenantContextFromCtx(ctx)
    // ...
}
```

The read replica also has RLS enforced — the same policies apply. The `set_tenant_context()` function must be called on any connection, primary or replica, before queries.

---

## 45.3. Backup and Recovery

### 45.3.1. Continuous WAL Archiving

Write-Ahead Log (WAL) archiving provides point-in-time recovery:

```bash
# postgresql.conf
archive_mode = on
archive_command = 'wal-g wal-push %p'
archive_timeout = 60  # ship WAL segment every 60s even if not full
```

### 45.3.2. Base Backup Schedule

Daily full base backup via WAL-G:

```bash
# 2 AM EAT (23:00 UTC previous day)
0 23 * * * wal-g backup-push $PGDATA
```

### 45.3.3. Recovery Time Objectives

| Scenario | Recovery procedure | RTO |
|---|---|---|
| Single tenant data corruption | Export from backup, re-import with tenant_id filter | ~30 min |
| Full database failure | Restore from WAL-G base backup + replay WAL | ~2 hours |
| Point-in-time recovery | Restore base backup + replay WAL to specific LSN | ~2 hours |

### 45.3.4. Single-Tenant Data Export for Recovery

Since tenant data is identified by `tenant_id` (not schema), a per-tenant extract uses `COPY` with a `WHERE` clause:

```bash
# Export all invoices for a specific tenant
psql $DB_URL -c "\COPY (
    SELECT * FROM invoices WHERE tenant_id = '$TENANT_ID'
) TO '/tmp/invoices_${TENANT_ID}.csv' CSV HEADER"
```

For a full tenant extract (all tables), a script iterates over all tenant-scoped tables:

```bash
#!/bin/bash
TABLES=(invoices journal_entries gl_lines employees payslips ...)
for table in "${TABLES[@]}"; do
    psql $DB_URL -c "\COPY (
        SELECT * FROM $table WHERE tenant_id = '$TENANT_ID'
    ) TO '/tmp/${table}_${TENANT_ID}.csv' CSV HEADER"
done
gzip /tmp/*_${TENANT_ID}.csv
aws s3 cp /tmp/*_${TENANT_ID}.csv.gz s3://awo-backups/tenants/${TENANT_ID}/$(date +%Y-%m-%d)/
```

This per-tenant export runs nightly in addition to the full WAL-G backup, enabling surgical single-tenant restore without touching other tenants' data.

---

## 45.4. Performance Optimisation

### 45.4.1. tenant_id on Every Index

Because all queries are filtered by `tenant_id` (via RLS), indexes on frequently-queried columns should include `tenant_id` as the leading column:

```sql
-- Without tenant_id first: PostgreSQL scans all tenants' invoices, then filters
CREATE INDEX idx_invoices_status ON invoices(status);

-- With tenant_id first: PostgreSQL jumps directly to this tenant's rows
CREATE INDEX CONCURRENTLY idx_invoices_tenant_status
    ON invoices(tenant_id, status);

-- Similarly for date-range queries
CREATE INDEX CONCURRENTLY idx_journal_entries_tenant_date
    ON journal_entries(tenant_id, posting_date);

-- GL lines by account (account ledger)
CREATE INDEX CONCURRENTLY idx_gl_lines_tenant_account_date
    ON gl_lines(tenant_id, account_id, posting_date);
```

The RLS policy `WHERE tenant_id = current_tenant_id()` combined with a leading `tenant_id` index means PostgreSQL can use an index range scan for just this tenant's rows, rather than a full table scan.

### 45.4.2. Partial Indexes for Active Records

Many queries filter on status as well:

```sql
-- Only active employees — most common HR query
CREATE INDEX CONCURRENTLY idx_employees_active
    ON employees(tenant_id, department_id)
    WHERE status = 'active';

-- Open opportunities pipeline
CREATE INDEX CONCURRENTLY idx_opportunities_open
    ON opportunities(tenant_id, owner_id, stage)
    WHERE stage NOT IN ('won', 'lost');
```

Partial indexes are smaller and faster than full indexes — they only index the rows that actually match the filter condition.

### 45.4.3. JSONB GIN Index for Custom Fields

```sql
CREATE INDEX CONCURRENTLY idx_invoices_custom_fields
    ON invoices USING gin(custom_fields jsonb_path_ops);
```

The `tenant_id` filter is not easily incorporated into a GIN index. In practice, the GIN index narrows the result set quickly, and the `tenant_id` RLS filter removes non-tenant rows from the result — acceptable because custom field queries are rare.

### 45.4.4. Table Partitioning for High-Volume Tables

For high-transaction-volume tenants, `gl_lines` and `stock_entries` can be partitioned by month. Since all tenants share the same table, partitioning is by date (not tenant):

```sql
CREATE TABLE gl_lines (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    posting_date date NOT NULL,
    ...
) PARTITION BY RANGE (posting_date);

CREATE TABLE gl_lines_2025_07 PARTITION OF gl_lines
  FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
```

Old partitions can be detached and archived to cold storage without affecting queries on recent partitions. Indexes on each partition include `tenant_id` as the leading column.

### 45.4.5. Autovacuum Tuning for High-Write Tables

```sql
ALTER TABLE stock_ledger SET (
    autovacuum_vacuum_scale_factor = 0.02,
    autovacuum_analyze_scale_factor = 0.01
);
```

### 45.4.6. Slow Query Monitoring

```bash
# postgresql.conf
log_min_duration_statement = 1000  # log queries > 1 second
log_line_prefix = '%t [%p]: tenant=%a '
```

`application_name` is set to the tenant slug on each connection:

```go
conn.Exec("SET application_name = 'awo_" + tenantSlug + "'")
```

Slow query logs show which tenant's queries are slow without parsing query parameters.
