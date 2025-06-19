-- =====================================================
-- SQLC QUERIES FOR ERP/ACCOUNTING SYSTEM
-- Multi-tenant with Row Level Security (RLS)
-- =====================================================

-- =====================================================
-- TENANT MANAGEMENT QUERIES (Admin/System Level)
-- Note: These queries are for system administrators managing tenants
-- =====================================================

-- name: CreateTenant :one
INSERT INTO tenants (name, subdomain, status, industry)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTenantByUUID :one
SELECT * FROM tenants
WHERE uuid = $1 AND deleted_at IS NULL;

-- name: GetTenantBySubdomain :one
SELECT * FROM tenants
WHERE subdomain = $1 AND deleted_at IS NULL;

-- name: ListTenants :many
SELECT * FROM tenants
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, subdomain = $3, status = $4, industry = $5, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteTenant :exec
UPDATE tenants
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetActiveTenants :many
SELECT * FROM tenants
WHERE status = 'active' AND deleted_at IS NULL
ORDER BY name;

-- name: CountTenants :one
SELECT COUNT(*) FROM tenants
WHERE deleted_at IS NULL;

-- name: SearchTenantsByName :many
SELECT * FROM tenants
WHERE name ILIKE '%' || $1 || '%' 
  AND deleted_at IS NULL
ORDER BY name
LIMIT $2 OFFSET $3;

-- =====================================================
-- CURRENT TENANT QUERIES (RLS-Aware)
-- These queries work within the current tenant context
-- =====================================================

-- name: GetCurrentTenant :one
SELECT * FROM tenants
WHERE id = current_tenant_id() AND deleted_at IS NULL;

-- name: UpdateCurrentTenant :one
UPDATE tenants
SET name = $1, subdomain = $2, status = $3, industry = $4, updated_at = NOW()
WHERE id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- =====================================================
-- TENANT CONFIGURATION QUERIES
-- =====================================================

-- Admin-level configuration management
-- name: CreateTenantConfiguration :one
INSERT INTO tenant_configurations (tenant_id, max_users, storage_quota, features, modules_enabled)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTenantConfiguration :one
SELECT * FROM tenant_configurations
WHERE tenant_id = $1;

-- name: UpdateTenantConfiguration :one
UPDATE tenant_configurations
SET max_users = $2, storage_quota = $3, features = $4, modules_enabled = $5
WHERE tenant_id = $1
RETURNING *;

-- Current tenant configuration queries (RLS-aware)
-- name: GetCurrentTenantConfiguration :one
SELECT * FROM tenant_configurations
WHERE tenant_id = current_tenant_id();

-- name: UpdateCurrentTenantMaxUsers :one
UPDATE tenant_configurations
SET max_users = $1
WHERE tenant_id = current_tenant_id()
RETURNING *;

-- name: UpdateCurrentTenantStorageQuota :one
UPDATE tenant_configurations
SET storage_quota = $1
WHERE tenant_id = current_tenant_id()
RETURNING *;

-- name: UpdateCurrentTenantFeatures :one
UPDATE tenant_configurations
SET features = $1
WHERE tenant_id = current_tenant_id()
RETURNING *;

-- name: UpdateCurrentTenantModules :one
UPDATE tenant_configurations
SET modules_enabled = $1
WHERE tenant_id = current_tenant_id()
RETURNING *;

-- name: CreateCurrentTenantConfiguration :one
INSERT INTO tenant_configurations (tenant_id, max_users, storage_quota, features, modules_enabled)
VALUES (current_tenant_id(), $1, $2, $3, $4)
RETURNING *;

-- Combined queries with current tenant context
-- name: GetCurrentTenantWithConfiguration :one
SELECT 
    t.*,
    tc.max_users,
    tc.storage_quota,
    tc.features,
    tc.modules_enabled
FROM tenants t
LEFT JOIN tenant_configurations tc ON t.id = tc.tenant_id
WHERE t.id = current_tenant_id() AND t.deleted_at IS NULL;

-- name: CheckCurrentTenantHasFeature :one
SELECT EXISTS(
    SELECT 1 FROM tenant_configurations 
    WHERE tenant_id = current_tenant_id() 
    AND features ? $1
);

-- name: CheckCurrentTenantHasModule :one
SELECT EXISTS(
    SELECT 1 FROM tenant_configurations 
    WHERE tenant_id = current_tenant_id() 
    AND modules_enabled ? $1
);

-- Admin queries (for system administration)
-- name: GetTenantWithConfiguration :one
SELECT 
    t.*,
    tc.max_users,
    tc.storage_quota,
    tc.features,
    tc.modules_enabled
FROM tenants t
LEFT JOIN tenant_configurations tc ON t.id = tc.tenant_id
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: ListTenantsWithConfigurations :many
SELECT 
    t.*,
    tc.max_users,
    tc.storage_quota,
    tc.features,
    tc.modules_enabled
FROM tenants t
LEFT JOIN tenant_configurations tc ON t.id = tc.tenant_id
WHERE t.deleted_at IS NULL
ORDER BY t.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetTenantsWithFeature :many
SELECT 
    t.*,
    tc.max_users,
    tc.storage_quota,
    tc.features,
    tc.modules_enabled
FROM tenants t
JOIN tenant_configurations tc ON t.id = tc.tenant_id
WHERE tc.features ? $1 
  AND t.deleted_at IS NULL
ORDER BY t.name;

-- name: GetTenantsWithModule :many
SELECT 
    t.*,
    tc.max_users,
    tc.storage_quota,
    tc.features,
    tc.modules_enabled
FROM tenants t
JOIN tenant_configurations tc ON t.id = tc.tenant_id
WHERE tc.modules_enabled ? $1 
  AND t.deleted_at IS NULL
ORDER BY t.name;

-- name: DeleteCurrentTenantConfiguration :exec
DELETE FROM tenant_configurations
WHERE tenant_id = current_tenant_id();

-- =====================================================
-- UTILITY QUERIES
-- =====================================================

-- Current tenant utilities
-- name: GetCurrentTenantID :one
SELECT current_tenant_id();

-- name: CheckCurrentTenantExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE id = current_tenant_id() AND deleted_at IS NULL
);

-- name: GetCurrentTenantStorageUsage :one
SELECT 
    t.id,
    t.name,
    tc.storage_quota,
    COALESCE(tc.storage_quota, 1073741824) as quota_bytes
FROM tenants t
LEFT JOIN tenant_configurations tc ON t.id = tc.tenant_id
WHERE t.id = current_tenant_id() AND t.deleted_at IS NULL;

-- Admin utilities (system-wide)
-- name: GetTenantStats :one
SELECT 
    COUNT(*) as total_tenants,
    COUNT(*) FILTER (WHERE status = 'active') as active_tenants,
    COUNT(*) FILTER (WHERE status = 'suspended') as suspended_tenants,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_tenants
FROM tenants
WHERE deleted_at IS NULL;

-- name: GetTenantsByIndustry :many
SELECT industry, COUNT(*) as tenant_count
FROM tenants
WHERE deleted_at IS NULL AND industry IS NOT NULL
GROUP BY industry
ORDER BY tenant_count DESC;

-- name: GetTenantsCreatedInDateRange :many
SELECT * FROM tenants
WHERE created_at >= $1 
  AND created_at <= $2
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CheckTenantExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE id = $1 AND deleted_at IS NULL
);

-- name: CheckSubdomainExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE subdomain = $1 AND deleted_at IS NULL
);

-- name: CheckTenantNameExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE name = $1 AND deleted_at IS NULL
);
-- =====================================================
-- BULK OPERATIONS
-- =====================================================
-- name: BulkUpdateTenantStatus :exec
UPDATE tenants
SET status = @status, updated_at = NOW()
WHERE id = ANY(sqlc.slice('id')::int[]) AND deleted_at IS NULL;


-- name: BulkSoftDeleteTenants :exec
UPDATE tenants
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = ANY(sqlc.slice('tenant_ids')::int[]);

-- =====================================================
-- ADVANCED QUERIES WITH FILTERS
-- =====================================================

-- name: FilterTenants :many
SELECT * FROM tenants
WHERE (sqlc.narg('name_filter')::varchar IS NULL OR name ILIKE '%' || sqlc.narg('name_filter') || '%')
  AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter'))
  AND (sqlc.narg('industry_filter')::varchar IS NULL OR industry = sqlc.narg('industry_filter'))
  AND deleted_at IS NULL
ORDER BY 
  CASE WHEN sqlc.arg('sort_by') = 'name' THEN name END ASC,
  CASE WHEN sqlc.arg('sort_by') = 'created_at' THEN created_at END DESC,
  CASE WHEN sqlc.arg('sort_by') = 'updated_at' THEN updated_at END DESC,
  id DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CountFilteredTenants :one
SELECT COUNT(*) FROM tenants
WHERE (sqlc.narg('name_filter')::varchar IS NULL OR name ILIKE '%' || sqlc.narg('name_filter') || '%')
  AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter'))
  AND (sqlc.narg('industry_filter')::varchar IS NULL OR industry = sqlc.narg('industry_filter'))
  AND deleted_at IS NULL;
