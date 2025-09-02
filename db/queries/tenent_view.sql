-- =====================================================================
-- DATABASE VIEWS QUERIES FOR SQLC GENERATION
-- Queries for database views and materialized views with proper tenant isolation
-- Admin queries use tenant_id parameter, user queries use current_tenant_id()
-- =====================================================================

-- =====================================================================
-- MATERIALIZED VIEWS QUERIES
-- =====================================================================

-- name: RefreshTenantFeatureFlagsCache :exec
REFRESH MATERIALIZED VIEW mv_tenant_feature_flags_cache;

-- name: GetTenantFeatureFlagsCacheAdmin :many
SELECT 
  tenant_id,
  feature_flag_name,
  enabled,
  rollout_percentage,
  cache_timestamp
FROM mv_tenant_feature_flags_cache 
WHERE tenant_id = $1;

-- name: GetTenantFeatureFlagsCacheUser :many  
SELECT 
  tenant_id,
  feature_flag_name,
  enabled,
  rollout_percentage,
  cache_timestamp
FROM mv_tenant_feature_flags_cache 
WHERE tenant_id = current_tenant_id();

-- name: CountTenantFeatureFlagsAdmin :one
SELECT COUNT(*) as total_flags
FROM mv_tenant_feature_flags_cache 
WHERE tenant_id = $1;

-- name: CountTenantFeatureFlagsUser :one
SELECT COUNT(*) as total_flags
FROM mv_tenant_feature_flags_cache 
WHERE tenant_id = current_tenant_id();

-- =====================================================================
-- STANDARD VIEWS QUERIES  
-- =====================================================================

-- name: GetTenantResourceUtilizationAdmin :many
SELECT 
  tenant_id,
  tenant_name,
  tenant_status,
  total_entities,
  active_entities,
  non_deleted_entities,
  sequence_states,
  document_types,
  last_entity_created,
  last_entity_updated
FROM v_tenant_resource_utilization 
WHERE tenant_id = $1;

-- name: GetTenantResourceUtilizationUser :many
SELECT 
  tenant_id,
  tenant_name,
  tenant_status,
  total_entities,
  active_entities,
  non_deleted_entities,
  sequence_states,
  document_types,
  last_entity_created,
  last_entity_updated
FROM v_tenant_resource_utilization 
WHERE tenant_id = current_tenant_id();

-- name: GetTenantEntitySummaryAdmin :many
SELECT 
  tenant_id,
  tenant_name,
  tenant_status,
  total_entities,
  company_count,
  regional_count,
  department_count,
  cost_center_count,
  project_count,
  active_entities,
  non_deleted_entities
FROM v_tenant_entity_summary 
WHERE tenant_id = $1;

-- name: GetTenantEntitySummaryUser :many
SELECT 
  tenant_id,
  tenant_name,
  tenant_status,
  total_entities,
  company_count,
  regional_count,
  department_count,
  cost_center_count,
  project_count,
  active_entities,
  non_deleted_entities
FROM v_tenant_entity_summary 
WHERE tenant_id = current_tenant_id();

-- name: GetFinancialStatementBuilderAdmin :many
SELECT 
  tenant_id,
  statement_section,
  header_code,
  header_name,
  header_order,
  group_code,
  group_name,
  group_category,
  account_count,
  group_balance,
  active_balance,
  running_total,
  percentage_of_section
FROM v_financial_statement_builder 
WHERE tenant_id = $1;

-- name: GetFinancialStatementBuilderUser :many
SELECT 
  tenant_id,
  statement_section,
  header_code,
  header_name,
  header_order,
  group_code,
  group_name,
  group_category,
  account_count,
  group_balance,
  active_balance,
  running_total,
  percentage_of_section
FROM v_financial_statement_builder 
WHERE tenant_id = current_tenant_id();

-- name: GetSecurityThreatDashboardAdmin :many
SELECT 
  user_id,
  username,
  email,
  high_risk_sessions,
  max_risk_score,
  critical_events,
  last_suspicious_activity
FROM v_security_threat_dashboard 
WHERE user_id = $1;

-- name: GetSecurityThreatDashboardUser :many
SELECT 
  user_id,
  username,
  email,
  high_risk_sessions,
  max_risk_score,
  critical_events,
  last_suspicious_activity
FROM v_security_threat_dashboard;

-- =====================================================================
-- TENANT ANALYTICS QUERIES
-- =====================================================================

-- name: GetTenantAnalyticsOverviewAdmin :one
SELECT 
  COUNT(DISTINCT CASE WHEN t.status = 'active' THEN t.id END) as active_tenants,
  COUNT(DISTINCT CASE WHEN t.status = 'suspended' THEN t.id END) as suspended_tenants,
  COUNT(DISTINCT CASE WHEN t.created_at > NOW() - INTERVAL '30 days' THEN t.id END) as new_tenants_30d,
  AVG(CASE WHEN t.last_activity_at IS NOT NULL THEN EXTRACT(epoch FROM (NOW() - t.last_activity_at))/86400 END) as avg_days_since_activity
FROM tenants t
WHERE t.deleted_at IS NULL
  AND ($1::UUID IS NULL OR t.id = $1);

-- name: GetTenantAnalyticsOverviewUser :one  
SELECT 
  COUNT(DISTINCT CASE WHEN t.status = 'active' THEN t.id END) as active_tenants,
  COUNT(DISTINCT CASE WHEN t.status = 'suspended' THEN t.id END) as suspended_tenants,
  COUNT(DISTINCT CASE WHEN t.created_at > NOW() - INTERVAL '30 days' THEN t.id END) as new_tenants_30d,
  AVG(CASE WHEN t.last_activity_at IS NOT NULL THEN EXTRACT(epoch FROM (NOW() - t.last_activity_at))/86400 END) as avg_days_since_activity
FROM tenants t
WHERE t.deleted_at IS NULL
  AND t.id = current_tenant_id();

-- name: GetTenantGrowthMetricsAdmin :many
SELECT 
  DATE_TRUNC('month', created_at) as month,
  COUNT(*) as new_tenants,
  COUNT(*) FILTER (WHERE status = 'active') as active_new_tenants
FROM tenants 
WHERE deleted_at IS NULL
  AND created_at >= NOW() - INTERVAL '12 months'
  AND ($1::UUID IS NULL OR id = $1)
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY month;

-- name: GetTenantGrowthMetricsUser :many
SELECT 
  DATE_TRUNC('month', created_at) as month,
  COUNT(*) as new_tenants,
  COUNT(*) FILTER (WHERE status = 'active') as active_new_tenants
FROM tenants 
WHERE deleted_at IS NULL
  AND created_at >= NOW() - INTERVAL '12 months'
  AND id = current_tenant_id()
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY month;

-- =====================================================================
-- VIEW PERFORMANCE AND METADATA QUERIES
-- =====================================================================

-- name: TestViewPerformanceAdmin :exec
-- Test query for view performance with admin access
SELECT COUNT(*) FROM (
  SELECT * FROM v_tenant_resource_utilization WHERE tenant_id = $1 LIMIT 1000
) t;

-- name: TestViewPerformanceUser :exec  
-- Test query for view performance with user access
SELECT COUNT(*) FROM (
  SELECT * FROM v_tenant_resource_utilization WHERE tenant_id = current_tenant_id() LIMIT 1000
) t;

-- name: GetViewMetadata :many
-- Get metadata about available views
SELECT 
  table_schema::text AS schemaname,
  table_name::text AS viewname
FROM information_schema.views 
WHERE table_schema = 'public' 
  AND table_name LIKE 'v_%'
ORDER BY table_name;

-- name: GetMaterializedViewMetadata :many  
-- Get metadata about materialized views
SELECT 
  schemaname::text,
  matviewname::text
FROM pg_catalog.pg_matviews
WHERE schemaname = 'public'
  AND matviewname LIKE 'mv_%'
ORDER BY matviewname;

-- =====================================================================
-- FEATURE FLAG QUERIES WITH VIEW SUPPORT
-- =====================================================================

-- name: GetFeatureFlagStatusAdmin :one
SELECT 
  feature_flag_name,
  enabled,
  rollout_percentage
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = $1 
  AND feature_flag_name = $2;

-- name: GetFeatureFlagStatusUser :one
SELECT 
  feature_flag_name,
  enabled, 
  rollout_percentage
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = current_tenant_id()
  AND feature_flag_name = $1;

-- name: ListActiveFeatureFlagsAdmin :many
SELECT 
  feature_flag_name,
  enabled,
  rollout_percentage,
  cache_timestamp
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = $1
  AND enabled = true
ORDER BY feature_flag_name;

-- name: ListActiveFeatureFlagsUser :many
SELECT 
  feature_flag_name,
  enabled,
  rollout_percentage,
  cache_timestamp  
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = current_tenant_id()
  AND enabled = true
ORDER BY feature_flag_name;