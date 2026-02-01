-- Rollback the feature_flags table creation
-- =====================================================
-- DROP TRIGGERS
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
-- DROP TABLE
-- =====================================================
DROP TABLE IF EXISTS feature_flags;
