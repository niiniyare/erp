-- -- =====================================================
-- -- SQLC QUERIES FOR ERP/ACCOUNTING SYSTEM
-- -- Multi-tenant with Row Level Security (RLS)
-- -- =====================================================
--
-- -- =====================================================
-- -- TENANT MANAGEMENT QUERIES (Admin/System Level)
-- -- Note: These queries are for system administrators managing tenants
-- -- =====================================================
--
-- -- name: CreateTenant :one
-- INSERT INTO tenants (name, subdomain, status, industry)
-- VALUES ($1, $2, $3, $4)
-- RETURNING *;
--
-- -- name: SetCurrentTenant :exec
-- SET app.current_tenant = $1 ; -- Example session variable
-- --
-- -- name: GetTenantByID :one
-- SELECT * FROM tenants
-- WHERE id = $1 AND deleted_at IS NULL;
--
-- -- name: GetTenantByUUID :one
-- -- SELECT * FROM tenants
-- -- WHERE id = $1 AND deleted_at IS NULL;
--
-- -- name: GetTenantBySubdomain :one
-- SELECT * FROM tenants
-- WHERE subdomain = $1 AND deleted_at IS NULL;
--
-- -- name: ListTenants :many
-- SELECT * FROM tenants
-- WHERE deleted_at IS NULL
-- ORDER BY created_at DESC
-- LIMIT $1 OFFSET $2;
--
-- -- name: UpdateTenant :one
-- UPDATE tenants
-- SET 
--     name = COALESCE(sqlc.narg(name), name), 
-- subdomain = COALESCE(sqlc.narg(subdomain), subdomain),
-- status = COALESCE(sqlc.narg(status), status),
-- industry = COALESCE(sqlc.narg(industry), industry) ,
-- updated_at = NOW()
-- WHERE id = @id AND deleted_at IS NULL
--   RETURNING *;
--
-- -- name: UpdateTenantIndustry :one
-- UPDATE tenants
-- SET industry = $2, updated_at = NOW()
-- WHERE id = $1 AND deleted_at IS NULL
-- RETURNING *;
--
--
-- -- name: UpdateTenantName :one 
-- UPDATE tenants
-- SET name = $2, updated_at = NOW()
-- WHERE id = $1 AND deleted_at IS NULL
-- RETURNING *;
-- -- name: UpdateTenantSubdomain :one 
-- UPDATE tenants
-- SET subdomain = $2, updated_at = NOW()
-- WHERE id = $1 AND deleted_at IS NULL
-- RETURNING *;
-- --
-- -- name: UpdateTenantStatus :one
-- UPDATE tenants
-- SET status = $2, updated_at = NOW()
-- WHERE id = $1 AND deleted_at IS NULL
-- RETURNING *;
--
-- -- name: SoftDeleteTenant :exec
-- UPDATE tenants
-- SET deleted_at = NOW(), updated_at = NOW()
-- WHERE id = $1;
--
-- -- name: GetActiveTenants :many
-- SELECT * FROM tenants
-- WHERE status = 'active' AND deleted_at IS NULL
-- ORDER BY name;
--
-- -- name: CountTenants :one
-- SELECT COUNT(*) FROM tenants
-- WHERE deleted_at IS NULL;
--
-- -- name: DeleteTenant :exec
-- DELETE FROM tenants;
--
--
-- -- name: SearchTenantsByName :many
-- SELECT * FROM tenants
-- WHERE name ILIKE '%' || @name || '%' 
--   AND deleted_at IS NULL
-- ORDER BY name
-- LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
--
-- --=====================================================
-- -- CURRENT TENANT QUERIES (RLS-Aware)
-- -- These queries work within the current tenant context
-- -- =====================================================
--
-- -- name: GetCurrentTenant :one
-- SELECT * FROM tenants
-- WHERE id = current_tenant_id() AND deleted_at IS NULL;
--
-- -- name: UpdateCurrentTenant :one
-- UPDATE tenants
-- SET name = $1, subdomain = $2, status = $3, industry = $4, updated_at = NOW()
-- WHERE id = current_tenant_id() AND deleted_at IS NULL
-- RETURNING *;
--
-- -- =====================================================
-- -- UTILITY QUERIES
-- -- =====================================================
--
-- -- Current tenant utilities
-- -- name: GetCurrentTenantID :one
-- SELECT current_tenant_id();
--
-- -- name: CheckCurrentTenantExists :one
-- SELECT EXISTS(
--     SELECT 1 FROM tenants 
--     WHERE id = current_tenant_id() AND deleted_at IS NULL
-- );
--
-- -- name: GetCurrentTenantStorageUsage :one
-- -- SELECT 
-- --     t.id,
-- --     t.name,
-- --     tc.storage_quota,
-- --     COALESCE(tc.storage_quota, 1073741824) as quota_bytes
-- -- FROM tenants t
-- -- LEFT JOIN tenant_configurations tc ON t.id = tc.tenant_id
-- -- WHERE t.id = current_tenant_id() AND t.deleted_at IS NULL;
-- --
-- --
-- --
-- --
-- -- Admin utilities (system-wide)
-- -- name: GetTenantStats :one
-- SELECT 
--     COUNT(*) as total_tenants,
--     COUNT(*) FILTER (WHERE status = 'active') as active_tenants,
--     COUNT(*) FILTER (WHERE status = 'suspended') as suspended_tenants,
--     COUNT(*) FILTER (WHERE status = 'pending') as pending_tenants
-- FROM tenants
-- WHERE deleted_at IS NULL;
--
-- -- name: GetTenantsByIndustry :many
-- SELECT industry, COUNT(*) as tenant_count
-- FROM tenants
-- WHERE deleted_at IS NULL AND industry IS NOT NULL
-- GROUP BY industry
-- ORDER BY tenant_count DESC;
--
-- -- name: GetTenantsCreatedInDateRange :many
-- SELECT * FROM tenants
-- WHERE created_at >= $1 
--   AND created_at <= $2
--   AND deleted_at IS NULL
-- ORDER BY created_at DESC;
--
-- -- name: CheckTenantExists :one
-- SELECT EXISTS(
--     SELECT 1 FROM tenants 
--     WHERE id = $1 AND deleted_at IS NULL
-- );
--
-- -- name: CheckSubdomainExists :one
-- SELECT EXISTS(
--     SELECT 1 FROM tenants 
--     WHERE subdomain = $1 AND deleted_at IS NULL
-- );
--
-- -- name: CheckTenantNameExists :one
-- SELECT EXISTS(
--     SELECT 1 FROM tenants 
--     WHERE name = $1 AND deleted_at IS NULL
-- );
-- -- =====================================================
-- -- BULK OPERATIONS
-- -- =====================================================
-- -- name: BulkUpdateTenantStatus :exec
-- UPDATE tenants
-- SET status = @status, updated_at = NOW()
-- WHERE id = ANY(sqlc.slice('id')::UUID[]) AND deleted_at IS NULL;
--
--
-- -- name: BulkSoftDeleteTenants :exec
-- UPDATE tenants
-- SET deleted_at = NOW(), updated_at = NOW()
-- WHERE id = ANY(sqlc.slice('tenant_ids')::UUID[]);
--
-- -- =====================================================
-- -- ADVANCED QUERIES WITH FILTERS
-- -- =====================================================
--
-- -- name: FilterTenants :many
-- SELECT id, name, subdomain, status, industry, created_at, updated_at, deleted_at FROM tenants
-- WHERE (sqlc.narg('name_filter')::varchar IS NULL OR name ILIKE '%' || sqlc.narg('name_filter') || '%')
--   AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter'))
--   AND (sqlc.narg('industry_filter')::varchar IS NULL OR industry = sqlc.narg('industry_filter'))
--   AND deleted_at IS NULL
-- ORDER BY 
--   CASE WHEN sqlc.arg('sort_by')::varchar = 'name' THEN name END ASC,
--   CASE WHEN sqlc.arg('sort_by')::varchar = 'created_at' THEN created_at END DESC,
--   CASE WHEN sqlc.arg('sort_by')::varchar = 'updated_at' THEN updated_at END DESC,
--   id DESC
-- LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');
--
-- -- name: CountFilteredTenants :one
-- SELECT COUNT(*) FROM tenants
-- WHERE (sqlc.narg('name_filter')::varchar IS NULL OR name ILIKE '%' || sqlc.narg('name_filter') || '%')
--   AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter'))
--   AND (sqlc.narg('industry_filter')::varchar IS NULL OR industry = sqlc.narg('industry_filter'))
--   AND deleted_at IS NULL;
--
--
--
-- ==================================================
--- New vwrsion
-- ==================================================

-- =====================================================
-- SQLC QUERIES FOR ERP/ACCOUNTING SYSTEM
-- Multi-tenant with Row Level Security (RLS)
-- =====================================================
--
-- RLS APPROACH:
-- This file uses Row Level Security (RLS) to automatically filter data by tenant.
-- Instead of explicitly adding "WHERE tenant_id = $1" to every query, we rely on:
--
-- 1. set_tenant_context($1) - Sets current tenant for the session
-- 2. get_current_tenant_id() - Returns current tenant from session
-- 3. RLS policies - Automatically add tenant filtering to queries
--
-- QUERY TYPES:
-- - Tenant-level queries: Rely on RLS, no explicit tenant_id needed
-- - Admin-level queries: Include explicit tenant_id for cross-tenant operations
-- - System-level queries: Global operations (stats, analytics across all tenants)
--
-- USAGE:
-- 1. Call SetTenantContext(tenantID) once per request/session
-- 2. All subsequent queries automatically filter by that tenant
-- 3. No need to pass tenant_id to individual queries
-- =====================================================

-- =====================================================
-- TENANT CONFIGURATIONS QUERIES (RLS-AWARE)
-- =====================================================

-- name: CreateTenantConfiguration :one
INSERT INTO tenant_configurations (
    tenant_id, max_users, max_entities, max_transactions_per_month, storage_quota,
    features, modules_enabled, accounting_method, fiscal_year_start_month, 
    default_currency, date_format, number_format, language_code,
    password_policy, webhook_endpoints, api_rate_limits
) VALUES (
    get_current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetTenantConfiguration :one
SELECT * FROM tenant_configurations WHERE tenant_id = current_setting('app.current_tenant_id')::uuid;

-- name: UpdateTenantConfiguration :one
UPDATE tenant_configurations
SET 
    max_users = COALESCE(sqlc.narg('max_users'), max_users),
    max_entities = COALESCE(sqlc.narg('max_entities'), max_entities),
    max_transactions_per_month = COALESCE(sqlc.narg('max_transactions_per_month'), max_transactions_per_month),
    storage_quota = COALESCE(sqlc.narg('storage_quota'), storage_quota),
    features = COALESCE(sqlc.narg('features'), features),
    modules_enabled = COALESCE(sqlc.narg('modules_enabled'), modules_enabled),
    accounting_method = COALESCE(sqlc.narg('accounting_method'), accounting_method),
    fiscal_year_start_month = COALESCE(sqlc.narg('fiscal_year_start_month'), fiscal_year_start_month),
    default_currency = COALESCE(sqlc.narg('default_currency'), default_currency),
    date_format = COALESCE(sqlc.narg('date_format'), date_format),
    number_format = COALESCE(sqlc.narg('number_format'), number_format),
    language_code = COALESCE(sqlc.narg('language_code'), language_code),
    password_policy = COALESCE(sqlc.narg('password_policy'), password_policy),
    webhook_endpoints = COALESCE(sqlc.narg('webhook_endpoints'), webhook_endpoints),
    api_rate_limits = COALESCE(sqlc.narg('api_rate_limits'), api_rate_limits),
    updated_at = NOW()
RETURNING *;

-- name: UpdateTenantFeatures :one
UPDATE tenant_configurations
SET features = $1, updated_at = NOW()
RETURNING *;

-- name: UpdateTenantModules :one
UPDATE tenant_configurations
SET modules_enabled = $1, updated_at = NOW()
RETURNING *;

-- name: UpdateTenantLimits :one
UPDATE tenant_configurations
SET 
    max_users = $1,
    max_entities = $2,
    max_transactions_per_month = $3,
    storage_quota = $4,
    updated_at = NOW()
RETURNING *;

-- name: DeleteTenantConfiguration :exec
DELETE FROM tenant_configurations;

-- =====================================================
-- TENANT USAGE STATISTICS QUERIES (RLS-AWARE)
-- =====================================================

-- name: CreateTenantUsageStats :one
INSERT INTO tenant_usage_stats (
    tenant_id, period_start, period_end, active_users, total_entities,
    total_transactions, storage_used, api_calls, avg_response_time,
    error_rate, monthly_revenue
) VALUES (
    get_current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetTenantUsageStats :one
SELECT * FROM tenant_usage_stats
WHERE period_start = $1;

-- name: GetTenantUsageStatsRange :many
SELECT * FROM tenant_usage_stats
WHERE period_start >= $1 
  AND period_end <= $2
ORDER BY period_start DESC;

-- name: UpdateTenantUsageStats :one
UPDATE tenant_usage_stats
SET 
    active_users = COALESCE(sqlc.narg('active_users'), active_users),
    total_entities = COALESCE(sqlc.narg('total_entities'), total_entities),
    total_transactions = COALESCE(sqlc.narg('total_transactions'), total_transactions),
    storage_used = COALESCE(sqlc.narg('storage_used'), storage_used),
    api_calls = COALESCE(sqlc.narg('api_calls'), api_calls),
    avg_response_time = COALESCE(sqlc.narg('avg_response_time'), avg_response_time),
    error_rate = COALESCE(sqlc.narg('error_rate'), error_rate),
    monthly_revenue = COALESCE(sqlc.narg('monthly_revenue'), monthly_revenue)
WHERE period_start = @period_start
RETURNING *;

-- name: DeleteTenantUsageStats :exec
DELETE FROM tenant_usage_stats
WHERE period_start = $1;

-- name: GetLatestTenantUsageStats :one
SELECT * FROM tenant_usage_stats
ORDER BY period_start DESC
LIMIT 1;

-- =====================================================
-- TENANT CONTEXT AND LIMITS QUERIES
-- =====================================================

-- name: SetTenantContext :exec
SELECT set_tenant_context($1);

-- name: CheckTenantLimits :one
SELECT check_tenant_limits(get_current_tenant_id(), $1, $2);

-- name: CreateDefaultTenantConfiguration :exec
SELECT create_default_tenant_configuration(get_current_tenant_id());

-- =====================================================
-- ENHANCED TENANT QUERIES WITH NEW FIELDS
-- =====================================================

-- name: CreateTenantComplete :one
INSERT INTO tenants (
    name, slug, email, subdomain, status, timezone, currency_code,
    metadata, industry, company_size, tax_id, registration_number,
    legal_entity_type, settings
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetTenantBySlug :one
SELECT * FROM tenants
WHERE slug = $1 AND deleted_at IS NULL;

-- name: GetTenantByEmail :one
SELECT * FROM tenants
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateTenantComplete :one
UPDATE tenants
SET 
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    email = COALESCE(sqlc.narg('email'), email),
    subdomain = COALESCE(sqlc.narg('subdomain'), subdomain),
    status = COALESCE(sqlc.narg('status'), status),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    currency_code = COALESCE(sqlc.narg('currency_code'), currency_code),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    industry = COALESCE(sqlc.narg('industry'), industry),
    company_size = COALESCE(sqlc.narg('company_size'), company_size),
    tax_id = COALESCE(sqlc.narg('tax_id'), tax_id),
    registration_number = COALESCE(sqlc.narg('registration_number'), registration_number),
    legal_entity_type = COALESCE(sqlc.narg('legal_entity_type'), legal_entity_type),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTenantMetadata :one
UPDATE tenants
SET metadata = $1, updated_at = NOW()
WHERE id = get_current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTenantSettings :one
UPDATE tenants
SET settings = $1, updated_at = NOW()
WHERE id = get_current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: GetTenantsByCompanySize :many
SELECT company_size, COUNT(*) as tenant_count
FROM tenants
WHERE deleted_at IS NULL AND company_size IS NOT NULL
GROUP BY company_size
ORDER BY tenant_count DESC;

-- name: GetTenantsByTimezone :many
SELECT timezone, COUNT(*) as tenant_count
FROM tenants
WHERE deleted_at IS NULL
GROUP BY timezone
ORDER BY tenant_count DESC;

-- name: GetTenantsByCurrency :many
SELECT currency_code, COUNT(*) as tenant_count
FROM tenants
WHERE deleted_at IS NULL
GROUP BY currency_code
ORDER BY tenant_count DESC;

-- =====================================================
-- ADVANCED ANALYTICS QUERIES
-- =====================================================

-- Admin-level analytics (requires explicit tenant_id for cross-tenant queries)
-- name: GetTenantGrowthStats :many
SELECT 
    DATE_TRUNC('month', created_at) as month,
    COUNT(*) as new_tenants,
    COUNT(*) FILTER (WHERE status = 'active') as active_new_tenants
FROM tenants
WHERE deleted_at IS NULL
  AND created_at >= $1
  AND created_at <= $2
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY month;

-- name: GetTenantStatusDistribution :one
SELECT 
    COUNT(*) as total_tenants,
    COUNT(*) FILTER (WHERE status = 'active') as active_tenants,
    COUNT(*) FILTER (WHERE status = 'suspended') as suspended_tenants,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_tenants,
    ROUND(COUNT(*) FILTER (WHERE status = 'active') * 100.0 / COUNT(*), 2) as active_percentage,
    ROUND(COUNT(*) FILTER (WHERE status = 'suspended') * 100.0 / COUNT(*), 2) as suspended_percentage,
    ROUND(COUNT(*) FILTER (WHERE status = 'pending') * 100.0 / COUNT(*), 2) as pending_percentage
FROM tenants
WHERE deleted_at IS NULL;

-- Admin-level storage analytics (cross-tenant view)
-- name: GetAllTenantsStorageAnalytics :many
SELECT 
    t.id,
    t.name,
    t.status,
    tc.storage_quota,
    COALESCE(tus.storage_used, 0) as current_storage_used,
    ROUND(COALESCE(tus.storage_used, 0) * 100.0 / tc.storage_quota, 2) as storage_usage_percentage
FROM tenants t
JOIN tenant_configurations tc ON t.id = tc.tenant_id
LEFT JOIN tenant_usage_stats tus ON t.id = tus.tenant_id 
    AND tus.period_start <= CURRENT_DATE 
    AND tus.period_end >= CURRENT_DATE
WHERE t.deleted_at IS NULL
ORDER BY storage_usage_percentage DESC NULLS LAST;

-- Current tenant storage analytics (RLS-aware)
-- name: GetCurrentTenantStorageUsage :one
SELECT 
    t.id,
    t.name,
    t.status,
    tc.storage_quota,
    COALESCE(tus.storage_used, 0) as current_storage_used,
    ROUND(COALESCE(tus.storage_used, 0) * 100.0 / tc.storage_quota, 2) as storage_usage_percentage
FROM tenants t
JOIN tenant_configurations tc ON t.id = tc.tenant_id
LEFT JOIN tenant_usage_stats tus ON t.id = tus.tenant_id 
    AND tus.period_start <= CURRENT_DATE 
    AND tus.period_end >= CURRENT_DATE
WHERE t.deleted_at IS NULL;

-- Admin-level revenue analytics (cross-tenant view)
-- name: GetAllTenantsRevenueAnalytics :many
SELECT 
    t.id,
    t.name,
    t.industry,
    SUM(tus.monthly_revenue) as total_revenue,
    AVG(tus.monthly_revenue) as avg_monthly_revenue,
    COUNT(tus.period_start) as months_tracked
FROM tenants t
LEFT JOIN tenant_usage_stats tus ON t.id = tus.tenant_id
WHERE t.deleted_at IS NULL
  AND (tus.period_start IS NULL OR tus.period_start >= $1)
  AND (tus.period_end IS NULL OR tus.period_end <= $2)
GROUP BY t.id, t.name, t.industry
ORDER BY total_revenue DESC NULLS LAST;

-- Current tenant revenue analytics (RLS-aware)
-- name: GetCurrentTenantRevenueAnalytics :one
SELECT 
    t.id,
    t.name,
    t.industry,
    SUM(tus.monthly_revenue) as total_revenue,
    AVG(tus.monthly_revenue) as avg_monthly_revenue,
    COUNT(tus.period_start) as months_tracked
FROM tenants t
LEFT JOIN tenant_usage_stats tus ON t.id = tus.tenant_id
WHERE t.deleted_at IS NULL
  AND (tus.period_start IS NULL OR tus.period_start >= $1)
  AND (tus.period_end IS NULL OR tus.period_end <= $2)
GROUP BY t.id, t.name, t.industry;

-- =====================================================
-- TENANT MANAGEMENT QUERIES (Admin/System Level)
-- Note: These queries are for system administrators managing tenants
-- =====================================================

-- name: CreateTenant :one
INSERT INTO tenants (name, slug, email, subdomain, status, industry)
VALUES ($1, $2, $3, $4, $5, $6)
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
WHERE id = get_current_tenant_id() AND deleted_at IS NULL;

-- name: UpdateCurrentTenant :one
UPDATE tenants
SET name = $1, subdomain = $2, status = $3, industry = $4, updated_at = NOW()
WHERE id = get_current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- =====================================================
-- UTILITY QUERIES
-- =====================================================

-- Current tenant utilities
-- name: GetCurrentTenantID :one
SELECT get_current_tenant_id();

-- name: CheckCurrentTenantExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE id = get_current_tenant_id() AND deleted_at IS NULL
);

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
WHERE id = ANY(sqlc.slice('id')::UUID[]) AND deleted_at IS NULL;


-- name: BulkSoftDeleteTenants :exec
UPDATE tenants
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = ANY(sqlc.slice('tenant_ids')::UUID[]);

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
