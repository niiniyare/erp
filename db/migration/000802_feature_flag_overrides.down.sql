-- Rollback the tenant_feature_overrides table creation
-- =====================================================
-- DROP TRIGGERS
-- =====================================================
DROP TRIGGER IF EXISTS sync_tenant_overrides_feature_name ON tenant_feature_overrides;

DROP TRIGGER IF EXISTS update_tenant_overrides_updated_at ON tenant_feature_overrides;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS sync_feature_flag_name();

-- =====================================================
-- DROP POLICIES
-- =====================================================
DROP POLICY IF EXISTS tenant_overrides_readonly_access ON tenant_feature_overrides;

DROP POLICY IF EXISTS tenant_overrides_admin_access ON tenant_feature_overrides;

DROP POLICY IF EXISTS tenant_overrides_tenant_isolation ON tenant_feature_overrides;

-- =====================================================
-- DROP INDEXES
-- =====================================================
DROP INDEX IF EXISTS idx_tenant_overrides_lookup;

DROP INDEX IF EXISTS idx_tenant_overrides_value;

DROP INDEX IF EXISTS idx_tenant_overrides_updated_at;

DROP INDEX IF EXISTS idx_tenant_overrides_created_at;

DROP INDEX IF EXISTS idx_tenant_overrides_tenant_enabled;

DROP INDEX IF EXISTS idx_tenant_overrides_enabled;

DROP INDEX IF EXISTS idx_tenant_overrides_feature_name;

DROP INDEX IF EXISTS idx_tenant_overrides_feature_id;

DROP INDEX IF EXISTS idx_tenant_overrides_tenant;

-- =====================================================
-- DROP TABLES
-- =====================================================
DROP TABLE IF EXISTS tenant_feature_overrides;
DROP TABLE IF EXISTS tenant_feature_flags;
