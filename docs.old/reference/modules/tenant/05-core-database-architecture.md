[<-- Back to Index](README.md)

## Core Database Architecture

### Table Overview

The tenant module uses five core tables:

```markdown
TENANT DATABASE SCHEMA:

┌─────────────────────────┐
│       tenants            │  Master tenant record
├─────────────────────────┤
│ id (PK)                 │
│ slug                    │
│ name, email             │
│ subdomain               │
│ status                  │
│ industry, company_size  │
│ currency_code, timezone │
│ settings (JSONB)        │
│ metadata (JSONB)        │
│ tax_id                  │
│ deleted_at (soft delete)│
│ last_activity_at        │
└──────────┬──────────────┘
           │ 1:1
           ▼
┌─────────────────────────┐
│ tenant_configurations    │  Resource limits & settings
├─────────────────────────┤
│ id (PK)                 │
│ tenant_id (FK, UNIQUE)  │
│ max_users               │
│ max_entities            │
│ max_transactions/month  │
│ storage_quota_mb        │
│ allowed_modules (JSONB) │
│ accounting_method       │
│ fiscal_year_start_month │
│ password_policy (JSONB) │
│ api_rate_limits (JSONB) │
└──────────┬──────────────┘
           │
           │ (same tenant_id FK)
           ▼
┌─────────────────────────┐
│ tenant_usage_stats       │  Resource consumption tracking
├─────────────────────────┤
│ id (PK)                 │
│ tenant_id (FK, UNIQUE)  │
│ active_users            │
│ total_entities          │
│ total_transactions      │
│ storage_used_mb         │
│ api_calls_today         │
│ avg_response_time_ms    │
│ error_rate              │
│ monthly_revenue         │
│ last_calculated_at      │
└─────────────────────────┘

┌─────────────────────────┐
│ tenant_bulk_operations   │  Bulk admin operations
├─────────────────────────┤
│ id (PK)                 │
│ operation_type          │
│ actor_id, actor_name    │
│ total_tenants           │
│ successful/failed_count │
│ status                  │
│ parameters (JSONB)      │
│ started_at, completed_at│
└──────────┬──────────────┘
           │ 1:N
           ▼
┌─────────────────────────┐
│ tenant_bulk_op_results   │  Per-tenant operation results
├─────────────────────────┤
│ operation_id (PK, FK)   │
│ tenant_id (PK, FK)      │
│ status                  │
│ message, error_details  │
│ started_at, completed_at│
└─────────────────────────┘
```

### Core Functions

```markdown
DATABASE FUNCTIONS:

Tenant Context:
├── set_tenant_context(tenant_id UUID)
│   Sets: app.current_tenant_id session variable
│   Scope: Current transaction only (local = true)
│
├── current_tenant_id() → UUID
│   Reads: app.current_tenant_id from session
│   Used by: All RLS policies
│
├── clear_tenant_context()
│   Clears: app.current_tenant_id session variable
│
└── validate_and_set_tenant_context(tenant_id UUID) → (name, status)
    Validates: Tenant exists, not deleted, status is active/pending
    Updates: last_activity_at timestamp
    Sets: Tenant context if valid
    Raises: Exception if tenant not found, deleted, or inactive

Provisioning:
└── provision_tenant_complete(name, email, subdomain, ...) → UUID
    Creates: Tenant record with PENDING status
    Generates: UUID and unique slug
    Returns: New tenant ID

Limit Checking:
└── check_tenant_limits(tenant_id, resource_type, current_count)
    Compares: Current usage vs configured limits
    Checks: User count, entity count, transaction count, storage
    Raises: Exception if limit exceeded

Bulk Operations:
├── get_bulk_operation_summary(operation_id) → summary row
│   Returns: Counts, status, duration for monitoring
│
├── update_bulk_operation_counts(operation_id)
│   Recalculates: successful/failed counts from results
│   Updates: Overall operation status
│   Sets: completed_at when all items done
│
└── trigger_update_bulk_operation_counts()
    Trigger: AFTER INSERT/UPDATE/DELETE on results table
    Calls: update_bulk_operation_counts() automatically
```

### Indexes

```markdown
PERFORMANCE INDEXES:

tenants:
├── idx_tenants_slug         (slug) WHERE deleted_at IS NULL
├── idx_tenants_subdomain    (subdomain) WHERE deleted_at IS NULL
├── idx_tenants_status       (status)
└── idx_tenants_email        (email)

tenant_configurations:
└── idx_tenant_configurations_tenant (tenant_id) UNIQUE

tenant_usage_stats:
└── idx_tenant_usage_stats_tenant (tenant_id) UNIQUE

tenant_bulk_operations:
├── idx_bulk_ops_actor       (actor_id)
├── idx_bulk_ops_status      (status)
├── idx_bulk_ops_type        (operation_type)
└── idx_bulk_ops_created_at  (created_at)

tenant_bulk_operation_results:
├── idx_bulk_results_tenant  (tenant_id)
└── idx_bulk_results_status  (status)
```

### Triggers

```markdown
AUTOMATIC TRIGGERS:

1. generate_tenant_slug
   Table: tenants (BEFORE INSERT)
   Action: Auto-generates URL-safe slug from name

2. create_default_tenant_config
   Table: tenants (AFTER INSERT)
   Action: Creates tenant_configurations row with defaults

3. update_*_updated_at
   Tables: All tenant tables (BEFORE UPDATE)
   Action: Sets updated_at = NOW()

4. tenant_bulk_operation_results_update_counts
   Table: tenant_bulk_operation_results (AFTER INSERT/UPDATE/DELETE)
   Action: Recalculates parent operation counts and status
```

---

Next: [Tenant Provisioning](./06-tenant-provisioning.md)
