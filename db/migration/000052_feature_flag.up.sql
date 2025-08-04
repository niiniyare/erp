-- Creates the core feature_flags table with proper indexing and RLS

-- =====================================================
-- FEATURE FLAGS TABLE
-- =====================================================
CREATE TABLE feature_flags (
    -- Primary identifier
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    
    -- Descriptive information
    description TEXT,
    
    -- Flag configuration
    flag_type VARCHAR(20) NOT NULL DEFAULT 'boolean'
        CHECK (flag_type IN ('boolean', 'string', 'number', 'json')),
    default_value BOOLEAN NOT NULL DEFAULT false,
    
    -- Rollout settings
    rollout_percentage INTEGER
        CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    target_audience JSONB DEFAULT '{}', -- For advanced targeting rules
    
    -- Flexible metadata storage
    metadata JSONB DEFAULT '{}',
    
    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ, -- Soft delete support
    
    -- Unique constraint per tenant
    UNIQUE (tenant_id, name)
);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================
CREATE INDEX idx_feature_flags_tenant ON feature_flags(tenant_id);
CREATE INDEX idx_feature_flags_tenant_name ON feature_flags(tenant_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_feature_flags_type ON feature_flags(flag_type);
CREATE INDEX idx_feature_flags_rollout ON feature_flags(rollout_percentage) WHERE rollout_percentage IS NOT NULL;
CREATE INDEX idx_feature_flags_deleted_at ON feature_flags(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_feature_flags_created_at ON feature_flags(created_at);
CREATE INDEX idx_feature_flags_updated_at ON feature_flags(updated_at);
CREATE INDEX idx_feature_flags_target_audience ON feature_flags USING GIN (target_audience) WHERE target_audience != '{}';
CREATE INDEX idx_feature_flags_metadata ON feature_flags USING GIN (metadata) WHERE metadata != '{}';

-- =====================================================
-- TABLE COMMENTS
-- =====================================================
COMMENT ON TABLE feature_flags IS 'Master feature flags configuration table with tenant isolation';
COMMENT ON COLUMN feature_flags.rollout_percentage IS 'Percentage of tenants that should have this feature enabled (0-100)';
COMMENT ON COLUMN feature_flags.target_audience IS 'Advanced targeting rules (company_size, industry, etc.)';
COMMENT ON COLUMN feature_flags.metadata IS 'Additional metadata like expiration dates, dependencies, etc.';
COMMENT ON COLUMN feature_flags.deleted_at IS 'Soft delete timestamp - NULL means active';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================
ALTER TABLE feature_flags ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy for application role
CREATE POLICY feature_flags_tenant_isolation ON feature_flags
    FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Admin role can access all tenants
CREATE POLICY feature_flags_admin_access ON feature_flags
    FOR ALL TO admin_role
    USING (true);

-- Read-only role for monitoring/analytics
CREATE POLICY feature_flags_readonly_access ON feature_flags
    FOR SELECT TO readonly_role
    USING (true);

-- Policy comments
COMMENT ON POLICY feature_flags_tenant_isolation ON feature_flags IS 'Ensures tenant data isolation for application users';
COMMENT ON POLICY feature_flags_admin_access ON feature_flags IS 'Allows admin role full access across all tenants';
COMMENT ON POLICY feature_flags_readonly_access ON feature_flags IS 'Allows readonly role to view all feature flags for monitoring';

-- =====================================================
-- PERMISSIONS
-- =====================================================
GRANT SELECT, INSERT, UPDATE, DELETE ON feature_flags TO application_role;
GRANT ALL ON feature_flags TO admin_role;
GRANT SELECT ON feature_flags TO readonly_role;

-- =====================================================
-- TRIGGERS
-- =====================================================
-- Auto-update updated_at timestamp
CREATE TRIGGER update_feature_flags_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
