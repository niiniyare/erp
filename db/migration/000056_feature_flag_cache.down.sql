-- Rollback the feature flag cache materialized view and functions

-- =====================================================
-- DROP TRIGGERS
-- =====================================================
DROP TRIGGER IF EXISTS tenant_overrides_cache_refresh_trigger ON tenant_feature_overrides;
DROP TRIGGER IF EXISTS feature_flags_cache_refresh_trigger ON feature_flags;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS trigger_cache_refresh();
DROP FUNCTION IF EXISTS evaluate_all_feature_flags_cached();
DROP FUNCTION IF EXISTS evaluate_feature_flag_cached(VARCHAR);
DROP FUNCTION IF EXISTS check_cache_freshness(UUID);
DROP FUNCTION IF EXISTS get_feature_flags_cache_stats();
DROP FUNCTION IF EXISTS refresh_feature_flags_cache();

-- =====================================================
-- DROP INDEXES ON MATERIALIZED VIEW
-- =====================================================
DROP INDEX IF EXISTS idx_tenant_feature_cache_value;
DROP INDEX IF EXISTS idx_tenant_feature_cache_metadata;
DROP INDEX IF EXISTS idx_tenant_feature_cache_target_audience;
DROP INDEX IF EXISTS idx_tenant_feature_cache_rollout;
DROP INDEX IF EXISTS idx_tenant_feature_cache_timestamp;
DROP INDEX IF EXISTS idx_tenant_feature_cache_flag_type;
DROP INDEX IF EXISTS idx_tenant_feature_cache_source;
DROP INDEX IF EXISTS idx_tenant_feature_cache_enabled;
DROP INDEX IF EXISTS idx_tenant_feature_cache_tenant;
DROP INDEX IF EXISTS idx_tenant_feature_cache_name_lookup;
DROP INDEX IF EXISTS idx_tenant_feature_cache_pk;

-- =====================================================
-- DROP MATERIALIZED VIEW
-- =====================================================
DROP MATERIALIZED VIEW IF EXISTS tenant_feature_flags_cache;