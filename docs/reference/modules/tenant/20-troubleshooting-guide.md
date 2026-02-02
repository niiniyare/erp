[<-- Back to Index](README.md)

## Troubleshooting Guide

### Common Issues

#### 1. "Tenant not found" Error

```markdown
SYMPTOM:
  API returns 404: "Tenant not found: <uuid>"

CAUSES:
  a) Incorrect tenant UUID
  b) Tenant was soft-deleted (deleted_at IS NOT NULL)
  c) Cache stale after recent deletion

RESOLUTION:
  1. Verify UUID is correct
  2. Check database: SELECT id, deleted_at FROM tenants WHERE id = '<uuid>'
  3. If deleted_at is set → Tenant was soft-deleted
  4. Clear cache: InvalidateTenantCache(tenantID)
  5. If tenant should exist → Check if wrong UUID was used
```

#### 2. "Tenant is not active" Error

```markdown
SYMPTOM:
  validate_and_set_tenant_context() raises:
  "Tenant is not active: <uuid> (status: SUSPENDED)"

CAUSES:
  a) Tenant was suspended (payment, policy, security)
  b) Tenant is still in PENDING status
  c) Tenant was archived

RESOLUTION:
  1. Check status: SELECT status FROM tenants WHERE id = '<uuid>'
  2. SUSPENDED → Contact billing team or platform admin
  3. PENDING → Complete activation process
  4. ARCHIVED → Tenant cannot be reactivated (by policy)
```

#### 3. User Limit Exceeded

```markdown
SYMPTOM:
  "Tenant has reached maximum user limit (10)"

CAUSES:
  a) Tenant on Basic plan (max 10 users)
  b) Deactivated users still counted

RESOLUTION:
  1. Check current usage:
     SELECT active_users FROM tenant_usage_stats WHERE tenant_id = '<uuid>'
  2. Check limit:
     SELECT max_users FROM tenant_configurations WHERE tenant_id = '<uuid>'
  3. Options:
     a) Deactivate unused user accounts
     b) Upgrade to higher plan
     c) Platform admin can override limit (Enterprise)
```

#### 4. Module Access Denied

```markdown
SYMPTOM:
  403: "Module not available on current plan"

CAUSES:
  a) Module not in allowed_modules for tenant's plan
  b) Module was recently disabled

RESOLUTION:
  1. Check allowed modules:
     SELECT allowed_modules FROM tenant_configurations WHERE tenant_id = '<uuid>'
  2. If module not listed → Plan upgrade required
  3. If module listed but still failing → Clear config cache
```

#### 5. Cross-Tenant Data Leak (Should Never Happen)

```markdown
SYMPTOM:
  User sees data from another tenant

INVESTIGATION:
  1. Check RLS is enabled:
     SELECT relname, relrowsecurity FROM pg_class WHERE relname = '<table>';
     → relrowsecurity must be TRUE

  2. Check tenant context is set:
     SELECT current_setting('app.current_tenant_id', true);
     → Must return a UUID, not empty

  3. Check RLS policy exists:
     SELECT * FROM pg_policies WHERE tablename = '<table>';
     → Must have tenant_isolation_policy

  4. Check database role:
     SELECT current_user;
     → Must be application_role (not superuser which bypasses RLS)

CRITICAL: If confirmed → Immediately suspend affected tenants,
          investigate, patch, and audit all access logs.
```

#### 6. Slow Tenant Resolution

```markdown
SYMPTOM:
  Tenant lookup taking > 100ms

CAUSES:
  a) Cache miss (Redis down or cold cache)
  b) Missing database index
  c) High database load

RESOLUTION:
  1. Check Redis connectivity
  2. Check cache hit rate in metrics
  3. Verify indexes exist:
     \d idx_tenants_slug
     \d idx_tenants_subdomain
  4. Check database connection pool utilization
  5. Monitor: avg_response_time_ms in tenant_usage_stats
```

#### 7. Bulk Operation Stuck in IN_PROGRESS

```markdown
SYMPTOM:
  Bulk operation status remains IN_PROGRESS indefinitely

CAUSES:
  a) Processing node crashed mid-operation
  b) Individual result stuck in PROCESSING status
  c) Trigger failed to update counts

RESOLUTION:
  1. Check individual results:
     SELECT status, COUNT(*) FROM tenant_bulk_operation_results
     WHERE operation_id = '<uuid>' GROUP BY status;

  2. If PROCESSING items exist → Mark as FAILED:
     UPDATE tenant_bulk_operation_results
     SET status = 'FAILED', error_details = 'Processing interrupted'
     WHERE operation_id = '<uuid>' AND status = 'PROCESSING';

  3. Manually trigger count update:
     SELECT update_bulk_operation_counts('<uuid>');

  4. Verify operation status updated to COMPLETED/PARTIAL_SUCCESS
```

#### 8. Provisioning Fails

```markdown
SYMPTOM:
  New tenant provisioning returns error

COMMON ERRORS:

  "duplicate key value violates unique constraint"
  → Subdomain already taken
  → Resolution: Choose different subdomain

  "Tenant email already exists"
  → Email registered to another tenant
  → Resolution: Use different email or check existing account

  "Connection refused"
  → Database unavailable
  → Resolution: Check database cluster health

  "Permission denied"
  → Database role lacks INSERT permission
  → Resolution: GRANT INSERT ON tenants TO application_role
```

---

Next: [Business Rules & Validation](./21-business-rules-and-validation.md)
