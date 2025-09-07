-- Creates the tenant_feature_overrides table with proper indexing and RLS
-- =====================================================
-- TENANT FEATURE OVERRIDES TABLE
-- =====================================================
CREATE TABLE tenant_feature_overrides (
  -- Primary identifier
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  -- References
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
  feature_flag_name VARCHAR(100) NOT NULL,
  -- Override settings
  enabled BOOLEAN NOT NULL,
  value JSONB DEFAULT '{}',  -- For complex feature values
  reason TEXT,  -- Why this override was set
  -- Audit timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Unique constraint - one override per tenant per feature
  UNIQUE (tenant_id, feature_flag_id)
);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================
CREATE INDEX idx_tenant_overrides_tenant ON tenant_feature_overrides(tenant_id);

CREATE INDEX idx_tenant_overrides_feature_id ON tenant_feature_overrides(feature_flag_id);

CREATE INDEX idx_tenant_overrides_feature_name ON tenant_feature_overrides(feature_flag_name);

CREATE INDEX idx_tenant_overrides_enabled ON tenant_feature_overrides(enabled);

CREATE INDEX idx_tenant_overrides_tenant_enabled ON tenant_feature_overrides(tenant_id, enabled);

CREATE INDEX idx_tenant_overrides_created_at ON tenant_feature_overrides(created_at);

CREATE INDEX idx_tenant_overrides_updated_at ON tenant_feature_overrides(updated_at);

CREATE INDEX idx_tenant_overrides_value ON tenant_feature_overrides USING GIN (value)
WHERE
  value != '{}';

CREATE INDEX idx_tenant_overrides_lookup ON tenant_feature_overrides(tenant_id, feature_flag_name, enabled);

-- =====================================================
-- TABLE COMMENTS
-- =====================================================
COMMENT ON TABLE tenant_feature_overrides IS 'Tenant-specific feature flag overrides with audit trail';

COMMENT ON COLUMN tenant_feature_overrides.value IS 'Complex feature values for non-boolean flags (JSON format)';

COMMENT ON COLUMN tenant_feature_overrides.reason IS 'Business justification for the override';

COMMENT ON COLUMN tenant_feature_overrides.feature_flag_name IS 'Denormalized feature flag name for faster lookups';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================
ALTER TABLE
  tenant_feature_overrides ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy for application role
CREATE POLICY tenant_overrides_tenant_isolation ON tenant_feature_overrides FOR ALL TO application_role USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);

-- Admin role can access all tenants
CREATE POLICY tenant_overrides_admin_access ON tenant_feature_overrides FOR ALL TO admin_role USING (TRUE);

-- Read-only role for monitoring/analytics
CREATE POLICY tenant_overrides_readonly_access ON tenant_feature_overrides FOR
SELECT
  TO readonly_role USING (TRUE);

-- Policy comments
COMMENT ON POLICY tenant_overrides_tenant_isolation ON tenant_feature_overrides IS 'Ensures tenant data isolation for application users';

COMMENT ON POLICY tenant_overrides_admin_access ON tenant_feature_overrides IS 'Allows admin role full access across all tenants';

COMMENT ON POLICY tenant_overrides_readonly_access ON tenant_feature_overrides IS 'Allows readonly role to view all overrides for monitoring';

-- =====================================================
-- PERMISSIONS
-- =====================================================
GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON tenant_feature_overrides TO application_role;

GRANT ALL ON tenant_feature_overrides TO admin_role;

GRANT
SELECT
  ON tenant_feature_overrides TO readonly_role;

-- =====================================================
-- TRIGGERS
-- =====================================================
-- Auto-update updated_at timestamp
CREATE TRIGGER update_tenant_overrides_updated_at BEFORE
UPDATE
  ON tenant_feature_overrides FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sync feature_flag_name on insert/update
CREATE
OR REPLACE FUNCTION sync_feature_flag_name() RETURNS TRIGGER AS
$$
BEGIN
-- Update the denormalized feature_flag_name from the feature_flags table
SELECT
  name INTO NEW.feature_flag_name
FROM
  feature_flags
WHERE
  id = NEW.feature_flag_id;

IF NEW.feature_flag_name IS NULL THEN RAISE EXCEPTION 'Feature flag not found for ID: %',
NEW.feature_flag_id;

END IF;

RETURN NEW;

END;

$$
LANGUAGE plpgsql;

CREATE TRIGGER sync_tenant_overrides_feature_name BEFORE
INSERT
  OR
UPDATE
  ON tenant_feature_overrides FOR EACH ROW EXECUTE FUNCTION sync_feature_flag_name();

-- Add trigger comments
COMMENT ON TRIGGER sync_tenant_overrides_feature_name ON tenant_feature_overrides IS 'Maintains denormalized feature_flag_name for performance';

COMMENT ON FUNCTION sync_feature_flag_name() IS 'Syncs feature flag name in overrides table';
