-- Rollback: feature_flags + feature_flag_definitions + auto-seed triggers
-- =====================================================
-- DROP TRIGGERS (auto-seed)
-- =====================================================
DROP TRIGGER IF EXISTS trg_seed_resource_flag ON resources;
DROP TRIGGER IF EXISTS trg_seed_module_flag   ON modules;
DROP FUNCTION IF EXISTS seed_resource_flag_definition();
DROP FUNCTION IF EXISTS seed_module_flag_definition();

-- =====================================================
-- DROP TRIGGERS (updated_at)
-- =====================================================
DROP TRIGGER IF EXISTS update_feature_flags_updated_at ON feature_flags;

-- =====================================================
-- DROP POLICIES
-- =====================================================
DROP POLICY IF EXISTS feature_flags_readonly_access ON feature_flags;

DROP POLICY IF EXISTS feature_flags_admin_access ON feature_flags;

DROP POLICY IF EXISTS feature_flags_tenant_isolation ON feature_flags;

-- =====================================================
-- DROP INDEXES
-- =====================================================
DROP INDEX IF EXISTS idx_feature_flags_metadata;

DROP INDEX IF EXISTS idx_feature_flags_target_audience;

DROP INDEX IF EXISTS idx_feature_flags_updated_at;

DROP INDEX IF EXISTS idx_feature_flags_created_at;

DROP INDEX IF EXISTS idx_feature_flags_deleted_at;

DROP INDEX IF EXISTS idx_feature_flags_rollout;

DROP INDEX IF EXISTS idx_feature_flags_type;

DROP INDEX IF EXISTS idx_feature_flags_tenant_name;

DROP INDEX IF EXISTS idx_feature_flags_tenant;

-- =====================================================
-- DROP TABLES
-- =====================================================
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS feature_flag_definitions;
