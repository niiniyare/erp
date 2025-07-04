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

-- name: SetCurrentTenant :exec
SET app.current_tenant = $1 ; -- Example session variable
--
-- name: GetTenantByID :one
SELECT * FROM tenants
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTenantByUUID :one
-- SELECT * FROM tenants
-- WHERE id = $1 AND deleted_at IS NULL;

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
SET 
    name = COALESCE(sqlc.narg(name), name), 
subdomain = COALESCE(sqlc.narg(subdomain), subdomain),
status = COALESCE(sqlc.narg(status), status),
industry = COALESCE(sqlc.narg(industry), industry) ,
updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
  RETURNING *;

-- name: UpdateTenantIndustry :one
UPDATE tenants
SET industry = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;


-- name: UpdateTenantName :one 
UPDATE tenants
SET name = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
-- name: UpdateTenantSubdomain :one 
UPDATE tenants
SET subdomain = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
--
-- name: UpdateTenantStatus :one
UPDATE tenants
SET status = $2, updated_at = NOW()
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

-- name: DeleteTenant :exec
DELETE FROM tenants;


-- name: SearchTenantsByName :many
SELECT * FROM tenants
WHERE name ILIKE '%' || @name || '%' 
  AND deleted_at IS NULL
ORDER BY name
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

--=====================================================
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
-- SELECT 
--     t.id,
--     t.name,
--     tc.storage_quota,
--     COALESCE(tc.storage_quota, 1073741824) as quota_bytes
-- FROM tenants t
-- LEFT JOIN tenant_configurations tc ON t.id = tc.tenant_id
-- WHERE t.id = current_tenant_id() AND t.deleted_at IS NULL;
--
--
--
--
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
SELECT id, name, subdomain, status, industry, created_at, updated_at, deleted_at FROM tenants
WHERE (sqlc.narg('name_filter')::varchar IS NULL OR name ILIKE '%' || sqlc.narg('name_filter') || '%')
  AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter'))
  AND (sqlc.narg('industry_filter')::varchar IS NULL OR industry = sqlc.narg('industry_filter'))
  AND deleted_at IS NULL
ORDER BY 
  CASE WHEN sqlc.arg('sort_by')::varchar = 'name' THEN name END ASC,
  CASE WHEN sqlc.arg('sort_by')::varchar = 'created_at' THEN created_at END DESC,
  CASE WHEN sqlc.arg('sort_by')::varchar = 'updated_at' THEN updated_at END DESC,
  id DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CountFilteredTenants :one
SELECT COUNT(*) FROM tenants
WHERE (sqlc.narg('name_filter')::varchar IS NULL OR name ILIKE '%' || sqlc.narg('name_filter') || '%')
  AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter'))
  AND (sqlc.narg('industry_filter')::varchar IS NULL OR industry = sqlc.narg('industry_filter'))
  AND deleted_at IS NULL;
