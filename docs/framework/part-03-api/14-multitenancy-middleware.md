---
title: "Chapter 14: Multi-Tenancy Middleware"
part: "Part III — The API Layer"
chapter: 14
section: "14-multitenancy-middleware"
related:
  - "[Chapter 13: Middleware Pipeline](13-middleware-pipeline.md)"
  - "[Chapter 38: Tenant Lifecycle](../part-07-multitenancy/38-tenant-lifecycle.md)"
  - "[Chapter 9: Privacy Policies](../part-02-entity-system/09-privacy-policies.md)"
---

# Chapter 14: Multi-Tenancy Middleware

The tenant middleware is the foundation of Awo's multi-tenancy. Every API request must resolve to exactly one tenant before authentication, permission evaluation, or any database query can occur. Awo uses **PostgreSQL Row-Level Security (RLS)** on a shared schema — all tenants share the same tables, and isolation is enforced at the database level via a transaction-local session variable.

---

## 14.1. Tenant Identification

Awo resolves the tenant from three sources in priority order:

### 14.1.1. Priority 1: `X-Tenant-ID` Header

For API clients, mobile apps, and service-to-service calls:

```
GET /api/v1/invoices
X-Tenant-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890
Authorization: Bearer {token}
```

The `X-Tenant-ID` header accepts a UUID (canonical tenant ID) and takes highest priority.

### 14.1.2. Priority 2: Subdomain Parsing

For browser clients using the standard multi-tenant URL:

```
https://acme.awo.app/api/v1/invoices
         ^^^^
         tenant slug → looked up in tenants table → tenant UUID
```

The subdomain is extracted from the `Host` header. Known prefixes (`bo`, `portal`, `app`, `api`) are stripped first — `bo.acme.awo.app` resolves to tenant slug `acme`.

### 14.1.3. Priority 3: `tenant_id` Query Parameter

Supported for webhook callbacks and development. In production (`AWO_ENV=production`), this is disabled — query parameters appear in server logs and browser history.

### 14.1.4. Tenant Not Found — 404, Not 400

When the resolved identifier does not match any tenant, Awo returns **404 Not Found**, not 400. Returning 404 prevents tenant enumeration: the caller cannot distinguish "does not exist" from "exists but you cannot access it."

---

## 14.2. Tenant Context Propagation

### 14.2.1. TenantContext Structure

```go
type TenantContext struct {
    ID       uuid.UUID
    Slug     string
    Status   string   // "active", "suspended", "trial"
    Config   *TenantConfig
    Features FeatureFlagSet
}
```

No `SchemaName` field — there is no per-tenant schema. All tenants share the `public` schema. Isolation comes from RLS, not schema routing.

### 14.2.2. Storing Tenant Context in Request Scope

```go
// Fiber locals (handlers)
c.Locals(TenantIDKey, tenantID)   // uuid.UUID
c.Locals(TenantKey, tenantIDStr)  // string
c.Locals(EntityKey, tenantEntity) // *tenant.Tenant

// Go context (services, repositories)
ctx = shared.WithTenantID(c.Context(), tenantID)
c.SetUserContext(ctx)
```

### 14.2.3. PostgreSQL RLS Session Variable

After resolving and validating the tenant, the middleware activates RLS for the request by calling the `set_tenant_context()` stored procedure:

```go
// In tenant middleware — step 12
if err := config.Store.SetTenantContextFromCtx(ctx); err != nil {
    return sendErrorResponse(c, "Database context setup failed",
        fiber.StatusInternalServerError, requestID)
}
```

`SetTenantContextFromCtx` calls:

```sql
SELECT set_tenant_context($1::uuid)
```

The stored procedure:

```sql
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id uuid)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER AS $$
BEGIN
    -- Verify tenant exists and is ACTIVE before setting context
    IF NOT EXISTS (
        SELECT 1 FROM tenants
        WHERE id = p_tenant_id AND status = 'ACTIVE' AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'tenant_not_found_or_inactive'
            USING ERRCODE = 'P0002';
    END IF;

    -- is_local = TRUE → transaction-scoped, resets on COMMIT/ROLLBACK
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, TRUE);
END;
$$;
```

**Why a stored procedure instead of raw `SET LOCAL`?**

`SET LOCAL app.current_tenant_id = 'uuid'` would set the variable without verifying the tenant is active. The stored procedure enforces that:
1. The tenant exists in the `tenants` table
2. The tenant `status = 'ACTIVE'`
3. The tenant is not soft-deleted

A suspended tenant whose UUID is somehow known cannot bypass the status check by calling the API directly — the stored procedure will raise an exception.

### 14.2.4. Transaction-Local Scope

The `is_local = TRUE` argument to `set_config` limits the setting to the current transaction. On `COMMIT` or `ROLLBACK`, the variable resets to its previous value (empty string, treated as NULL by `current_tenant_id()`).

```sql
-- Helper function used in RLS policies
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid LANGUAGE sql STABLE AS $$
    SELECT NULLIF(current_setting('app.current_tenant_id', TRUE), '')::uuid;
$$;
```

This means:
- **No cross-request leakage**: even if PgBouncer reuses a connection, the tenant context does not carry over — it was reset on the previous transaction's COMMIT
- **PgBouncer transaction mode is safe**: the transaction-local scope makes connection reuse safe by design

### 14.2.5. Propagating Tenant Context into Temporal Workflows

When starting a Temporal workflow from a request context, the tenant UUID is serialised into the workflow params. Activities re-establish the RLS context before any DB operation:

```go
type WorkflowParams struct {
    TenantID uuid.UUID       `json:"tenant_id"`
    ActorID  uuid.UUID       `json:"actor_id"`
    Payload  json.RawMessage `json:"payload"`
}

// Inside activity — re-establish RLS context
func (a *Activities) PostJournalEntry(ctx context.Context, params WorkflowParams) error {
    ctx = shared.WithTenantID(ctx, params.TenantID)
    if err := a.store.SetTenantContextFromCtx(ctx); err != nil {
        return err
    }
    // All subsequent DB queries are RLS-filtered to params.TenantID
}
```

---

## 14.3. RLS Policy Structure

### 14.3.1. Standard Tenant Isolation Policy

Every tenant-scoped table has an RLS policy using `current_tenant_id()`:

```sql
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON invoices
    FOR ALL
    USING (tenant_id = current_tenant_id());
```

`FORCE ROW LEVEL SECURITY` ensures the policy applies even to the table owner. Without it, a connection with the `awo` role (table owner) would bypass RLS.

### 14.3.2. Global Tables — No RLS

Tables shared across all tenants are not tenant-scoped and have no RLS policy:

| Table | Reason |
|---|---|
| `tenants` | Platform-level — read by middleware before tenant context is set |
| `timezones` | Reference data shared by all tenants |
| `currencies` | Reference data |
| `countries` | Reference data |
| `paye_bands` | Kenya statutory rates — shared (tenants can override via config) |
| `platform_admins` | Platform-level users |
| `audit_log` | Append-only; platform admins query across tenants |

### 14.3.3. Re-Validating Tenant Context in Long-Running Operations

Background jobs (Temporal activities, nightly workflows) may run for minutes. A tenant can be suspended while the workflow is running. Use `ValidateTenantContext()` to re-check:

```go
// In a long-running activity, validate tenant is still active mid-execution
tenantID, err := store.ValidateTenantContext(ctx)
if err != nil {
    return temporal.NewNonRetryableApplicationError(
        "tenant suspended or deleted mid-execution", "TENANT_INACTIVE", err)
}
```

The `validate_tenant_context()` stored procedure re-reads the `tenants` table within the current transaction and returns the UUID if active, raises an exception otherwise.

---

## 14.4. Tenant State Handling

### 14.4.1. Active Tenant — Normal Flow

Status `ACTIVE`: tenant context established, RLS set, request proceeds.

### 14.4.2. Suspended Tenant — 402

Status `SUSPENDED`: middleware short-circuits before any RLS setup:

```go
if tenantEntity.Status == tenant.StatusSuspended {
    return c.Status(402).JSON(fiber.Map{
        "code":    "TENANT_SUSPENDED",
        "message": "Account suspended. Contact support@awo.so.",
        "reason":  tenantEntity.SuspensionReason,
    })
}
```

The `set_tenant_context()` stored procedure would also reject a suspended tenant, but the middleware check is earlier and avoids the DB round-trip.

### 14.4.3. Trial Tenant — Feature Flag Restrictions

Status `TRIAL`: request is allowed. Trial limitations are enforced via feature flags (`max_users`, `max_invoices_per_month`) evaluated from the `FeatureFlagSet` in the tenant context.

### 14.4.4. Archived Tenant — 410 Gone

Status `ARCHIVED`: middleware returns 410 Gone. The data still exists in the database (retention period), but the tenant is operationally closed.

---

## 14.5. Security Properties of RLS Isolation

### 14.5.1. What RLS Guarantees

When `tenant_id = current_tenant_id()` is enforced by RLS:

- A query `SELECT * FROM invoices` with tenant context set to `acme` returns **only** acme's invoices — even if the app code mistakenly omits a `WHERE tenant_id = ?` clause
- An `INSERT INTO invoices (...)` without a `tenant_id` column will violate the NOT NULL constraint on `tenant_id`, preventing phantom rows
- A `UPDATE invoices SET amount = 0` without `WHERE` only updates the current tenant's rows

RLS is a second line of defence behind the application's own tenant scoping. An application bug that forgets `WHERE tenant_id = ?` is caught at the database level, not at the application level.

### 14.5.2. What RLS Does Not Prevent

- **Timing side-channels**: aggregate queries (COUNT, SUM) are correctly scoped, but a carefully-crafted sequence of queries could infer information about another tenant's data volume from timing
- **Schema-level information**: table structure, function definitions, and index names are visible to all application connections (there is only one schema)
- **Superuser bypass**: a superuser connection bypasses RLS — production database access must be limited to the `awo_app` role

### 14.5.3. FORCE ROW LEVEL SECURITY

Without `FORCE ROW LEVEL SECURITY`, the table owner (`awo` role) bypasses policies. Awo sets FORCE on every tenant-scoped table:

```sql
-- Applied in every tenant table migration
ALTER TABLE {table} FORCE ROW LEVEL SECURITY;
```

The application connects as `awo_app`, not `awo` (the owner), to prevent accidental bypasses even without FORCE. FORCE is a belt-and-suspenders measure.
