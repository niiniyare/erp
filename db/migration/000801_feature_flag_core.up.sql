-- Creates the core feature_flags table with proper indexing and RLS.
-- Also adds the IAM-spec feature_flag_definitions catalogue (system-wide, no tenant_id)
-- which is the authoritative source for module/resource flags resolved into sessions.

-- =====================================================
-- FEATURE FLAG DEFINITIONS — system catalogue (no tenant_id)
-- =====================================================
-- Auto-seeded by triggers on modules and resources (see bottom of this file).
-- Every tenant's flag resolution starts from this catalogue:
--   COALESCE(tenant_feature_flags.enabled, feature_flag_definitions.default_value)
CREATE TABLE IF NOT EXISTS feature_flag_definitions (
  id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id   UUID    REFERENCES modules(id) ON DELETE CASCADE,    -- NULL = global/platform flag
  resource_id UUID    REFERENCES resources(id) ON DELETE CASCADE,  -- NULL = module-level flag
  flag_key    TEXT    UNIQUE NOT NULL,  -- 'finance' | 'finance.transactions'
  label       TEXT    NOT NULL,
  description TEXT,
  default_value BOOLEAN NOT NULL DEFAULT false,  -- effective value when no tenant override exists
  is_system   BOOLEAN NOT NULL DEFAULT false,    -- true = only platform operators can toggle
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  feature_flag_definitions             IS 'System-wide feature flag catalogue. No tenant_id — shared across all tenants. Auto-seeded by triggers on modules and resources. Tenant overrides live in tenant_feature_flags.';
COMMENT ON COLUMN feature_flag_definitions.flag_key    IS 'Dot-notation key derived from MRA slugs: module.slug or module.slug.resource.slug. e.g. ''finance'', ''finance.transactions''. Never write free-form strings — always derive from slugs.';
COMMENT ON COLUMN feature_flag_definitions.default_value IS 'Value tenants get without any configuration. Modules default false (off); resources default true (on once module is on).';
COMMENT ON COLUMN feature_flag_definitions.is_system   IS 'When true, only platform operators (admin_role) can toggle. Tenant admins cannot see or change these.';

CREATE INDEX idx_ffd_module   ON feature_flag_definitions(module_id)   WHERE module_id   IS NOT NULL;
CREATE INDEX idx_ffd_resource ON feature_flag_definitions(resource_id) WHERE resource_id IS NOT NULL;

-- Globally readable — no RLS needed (no tenant data)
GRANT SELECT ON feature_flag_definitions TO application_role;
GRANT SELECT ON feature_flag_definitions TO readonly_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON feature_flag_definitions TO admin_role;

-- =====================================================
-- AUTO-SEED TRIGGERS
-- =====================================================
-- When a module row is inserted, seed a module-level flag definition automatically.
CREATE OR REPLACE FUNCTION seed_module_flag_definition()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO feature_flag_definitions (module_id, flag_key, label, default_value, is_system)
    VALUES (NEW.id, NEW.slug, NEW.display_name || ' module', false, false)
    ON CONFLICT (flag_key) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION seed_module_flag_definition IS 'Auto-seeds a feature_flag_definitions row when a module is inserted. Idempotent (ON CONFLICT DO NOTHING).';

CREATE TRIGGER trg_seed_module_flag
    AFTER INSERT ON modules
    FOR EACH ROW EXECUTE FUNCTION seed_module_flag_definition();

-- When a resource row is inserted, seed a resource-level flag definition automatically.
CREATE OR REPLACE FUNCTION seed_resource_flag_definition()
RETURNS TRIGGER AS $$
DECLARE v_module_slug TEXT;
BEGIN
    SELECT slug INTO v_module_slug FROM modules WHERE id = NEW.module_id;
    INSERT INTO feature_flag_definitions (module_id, resource_id, flag_key, label, default_value)
    VALUES (NEW.module_id, NEW.id, v_module_slug || '.' || NEW.slug, NEW.display_name, true)
    -- resources default true: enabled once their parent module is enabled
    ON CONFLICT (flag_key) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION seed_resource_flag_definition IS 'Auto-seeds a feature_flag_definitions row when a resource is inserted. Resources default to true (enabled when module is on).';

CREATE TRIGGER trg_seed_resource_flag
    AFTER INSERT ON resources
    FOR EACH ROW EXECUTE FUNCTION seed_resource_flag_definition();

-- =====================================================
-- FEATURE FLAGS TABLE (tenant-scoped, pre-existing design)
-- =====================================================
CREATE TABLE feature_flags (
  -- Primary identifier
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  -- Descriptive information
  description TEXT,
  -- Flag configuration
  flag_type VARCHAR(20) NOT NULL DEFAULT 'boolean' CHECK (
    flag_type IN ('boolean', 'string', 'number', 'json')
  ),
  default_value BOOLEAN NOT NULL DEFAULT false,
  -- Rollout settings
  rollout_percentage INTEGER CHECK (
    rollout_percentage >= 0
    AND rollout_percentage <= 100
  ),
  target_audience JSONB DEFAULT '{}',  -- For advanced targeting rules
  -- Flexible metadata storage
  metadata JSONB DEFAULT '{}',
  -- Audit timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,  -- Soft delete support
  -- Unique constraint per tenant
  UNIQUE (tenant_id, name)
);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================
CREATE INDEX idx_feature_flags_tenant ON feature_flags(tenant_id);

CREATE INDEX idx_feature_flags_tenant_name ON feature_flags(tenant_id, name)
WHERE
  deleted_at IS NULL;

CREATE INDEX idx_feature_flags_type ON feature_flags(flag_type);

CREATE INDEX idx_feature_flags_rollout ON feature_flags(rollout_percentage)
WHERE
  rollout_percentage IS NOT NULL;

CREATE INDEX idx_feature_flags_deleted_at ON feature_flags(deleted_at)
WHERE
  deleted_at IS NOT NULL;

CREATE INDEX idx_feature_flags_created_at ON feature_flags(created_at);

CREATE INDEX idx_feature_flags_updated_at ON feature_flags(updated_at);

CREATE INDEX idx_feature_flags_target_audience ON feature_flags USING GIN (target_audience)
WHERE
  target_audience != '{}';

CREATE INDEX idx_feature_flags_metadata ON feature_flags USING GIN (metadata)
WHERE
  metadata != '{}';

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
ALTER TABLE feature_flags FORCE  ROW LEVEL SECURITY;

-- Tenant isolation policy for application role
CREATE POLICY feature_flags_tenant_isolation ON feature_flags FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- Admin role can access all tenants
CREATE POLICY feature_flags_admin_access ON feature_flags FOR ALL TO admin_role
    USING (TRUE) WITH CHECK (TRUE);

-- Read-only role for tenant-scoped analytics
CREATE POLICY feature_flags_readonly_access ON feature_flags
    FOR SELECT TO readonly_role
    USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- Policy comments
COMMENT ON POLICY feature_flags_tenant_isolation ON feature_flags IS 'Ensures tenant data isolation for application users';

COMMENT ON POLICY feature_flags_admin_access ON feature_flags IS 'Allows admin role full access across all tenants';

COMMENT ON POLICY feature_flags_readonly_access ON feature_flags IS 'Allows readonly role to view all feature flags for monitoring';

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
  DELETE ON feature_flags TO application_role;

GRANT ALL ON feature_flags TO admin_role;

GRANT
SELECT
  ON feature_flags TO readonly_role;

-- =====================================================
-- TRIGGERS
-- =====================================================
-- Auto-update updated_at timestamp
CREATE TRIGGER update_feature_flags_updated_at BEFORE
UPDATE
  ON feature_flags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
