-- ================================================================================================
-- SETTINGS MODULE - Configuration management with 3-level inheritance (System → Tenant → Entity)
-- Also includes IAM-spec setting_definitions, tenant_settings, and user_preferences tables
-- which power session pre-computation (SettingService.ResolveForTenant).
-- ================================================================================================

-- =====================================================================
-- SETTING DEFINITIONS — developer-seeded setting catalogue (no tenant_id)
-- =====================================================================
-- Unlike feature_flag_definitions (auto-seeded by triggers), setting definitions
-- are inserted by developers because settings require human decisions about
-- types, defaults, and validation ranges.
--
-- Naming: setting_key uses the same dot-notation as flags:
--   'finance.transactions.approval_threshold'
--   'finance.budget_control_mode'
--   'iam.mfa.required'
CREATE TABLE IF NOT EXISTS setting_definitions (
  id            UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id     UUID    REFERENCES modules(id)   ON DELETE CASCADE,
  resource_id   UUID    REFERENCES resources(id) ON DELETE CASCADE,
  action_id     UUID    REFERENCES actions(id)   ON DELETE CASCADE,  -- NULL for module/resource settings
  setting_key   TEXT    UNIQUE NOT NULL,    -- dot-notation key
  label         TEXT    NOT NULL,
  description   TEXT,
  value_type    TEXT    NOT NULL CHECK (value_type IN ('bool','int','decimal','text','enum')),
  default_value TEXT,                       -- stored as text, parsed by typed accessors
  enum_options  JSONB,                      -- [{"value":"soft","label":"Warn only"}, ...]
  min_value     TEXT,                       -- for numeric validation
  max_value     TEXT,
  is_system     BOOLEAN NOT NULL DEFAULT false,  -- true = only platform operators can change
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  setting_definitions             IS 'System-wide setting catalogue. No tenant_id — shared across all tenants. Developer-seeded (unlike feature_flag_definitions which are trigger-seeded). Tenant values in tenant_settings.';
COMMENT ON COLUMN setting_definitions.setting_key IS 'Dot-notation key: {module}.{resource?}.{name}. e.g. ''finance.transactions.approval_threshold''. Never free-form — always derived from module/resource slugs.';
COMMENT ON COLUMN setting_definitions.value_type  IS 'Value type: bool | int | decimal | text | enum. Determines which typed session accessor to use.';
COMMENT ON COLUMN setting_definitions.enum_options IS 'For enum type: [{value, label}] array. e.g. [{"value":"soft","label":"Warn only"},{"value":"hard","label":"Block"}]';
COMMENT ON COLUMN setting_definitions.is_system   IS 'When true, only platform operators can change. Shown as read-only/disabled in tenant admin UI.';

CREATE INDEX idx_setting_defs_module   ON setting_definitions(module_id)   WHERE module_id   IS NOT NULL;
CREATE INDEX idx_setting_defs_resource ON setting_definitions(resource_id) WHERE resource_id IS NOT NULL;

-- Globally readable
GRANT SELECT ON setting_definitions TO application_role;
GRANT SELECT ON setting_definitions TO readonly_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON setting_definitions TO admin_role;

-- =====================================================================
-- TENANT SETTINGS — per-tenant setting values
-- =====================================================================
-- One row per tenant per setting they have explicitly configured.
-- Missing rows fall back to setting_definitions.default_value.
CREATE TABLE IF NOT EXISTS tenant_settings (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  setting_id  UUID        NOT NULL REFERENCES setting_definitions(id) ON DELETE CASCADE,
  setting_key TEXT        NOT NULL,    -- denormalized for fast lookup
  value       TEXT        NOT NULL,    -- stored as text regardless of value_type
  set_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
  set_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (tenant_id, setting_id)
);

COMMENT ON TABLE  tenant_settings             IS 'Per-tenant setting values. Rows exist only for explicitly configured settings. Missing rows resolve to setting_definitions.default_value.';
COMMENT ON COLUMN tenant_settings.setting_key IS 'Denormalized from setting_definitions for fast single-table lookups in SettingService.';
COMMENT ON COLUMN tenant_settings.value       IS 'Stored as text. Parsed by typed session helpers: SettingBool/Int/Decimal/String.';

CREATE INDEX idx_ts_tenant  ON tenant_settings(tenant_id);
CREATE INDEX idx_ts_setting ON tenant_settings(setting_id);
CREATE INDEX idx_ts_lookup  ON tenant_settings(tenant_id, setting_key);

ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
CREATE POLICY ts_tenant_isolation ON tenant_settings FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
CREATE POLICY ts_admin_access ON tenant_settings FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_settings TO application_role;
GRANT ALL ON tenant_settings TO admin_role;
GRANT SELECT ON tenant_settings TO readonly_role;

-- =====================================================================
-- USER PREFERENCES — per-user display preferences
-- =====================================================================
-- Not business logic — these are UI/UX preferences that travel in the session.
-- e.g. 'finance.entry_mode' → 'spreadsheet', 'finance.show_account_codes' → 'true'
CREATE TABLE IF NOT EXISTS user_preferences (
  id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  pref_key TEXT        NOT NULL,   -- namespaced: '{module}.{pref_name}'
  value    TEXT        NOT NULL,
  set_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, pref_key)
);

COMMENT ON TABLE  user_preferences          IS 'Per-user UI/display preferences. Stored in session configuration.prefs at login. Not business logic.';
COMMENT ON COLUMN user_preferences.pref_key IS 'Preference key namespaced by module. e.g. ''finance.entry_mode'', ''finance.show_account_codes''.';

CREATE INDEX idx_uprefs_user ON user_preferences(user_id);

ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
-- Users can only see/edit their own preferences; admins see all
CREATE POLICY uprefs_self ON user_preferences FOR ALL TO application_role
    USING  (user_id = current_setting('app.user_id', true)::uuid)
    WITH CHECK (user_id = current_setting('app.user_id', true)::uuid);
CREATE POLICY uprefs_admin ON user_preferences FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

GRANT SELECT, INSERT, UPDATE, DELETE ON user_preferences TO application_role;
GRANT ALL ON user_preferences TO admin_role;

-- ================================================================================================
--
-- Core tables for ERP Settings Module implementing configuration inheritance, templates,
-- and audit trails for enterprise configuration management.
--
-- Prerequisites:
-- - tenants table with UUID primary key
-- - entities table with UUID primary key
-- ================================================================================================

-- =====================================================================
-- CONFIG DEFINITIONS - System-wide metadata for all configurations
-- =====================================================================
-- CREATE TABLE config_definitions (
--   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--   entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
--   module_name VARCHAR(50) NOT NULL,
--   config_key VARCHAR(100) NOT NULL,
--   data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('STRING', 'INTEGER', 'BOOLEAN', 'DECIMAL', 'JSON')),
--   default_value JSONB,
--   validation_rules JSONB DEFAULT '{}'::jsonb,
--   description TEXT,
--   required_permission VARCHAR(100),
--   required_feature_flag VARCHAR(100),
--   is_overridable BOOLEAN NOT NULL DEFAULT true,
--   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--   -- Ensure unique configuration keys per module
--   CONSTRAINT config_definitions_module_key_unique UNIQUE (module_name, config_key)
-- );
--
-- -- Add table and column comments
-- COMMENT ON TABLE config_definitions IS 'System-wide configuration metadata defining all possible configuration keys with validation rules and inheritance policies';
--
-- COMMENT ON COLUMN config_definitions.id IS 'UUID primary key for the configuration definition';
-- COMMENT ON COLUMN config_definitions.module_name IS 'ERP module that owns this configuration (finance, hr, inventory, etc.)';
-- COMMENT ON COLUMN config_definitions.config_key IS 'Unique configuration key within the module namespace';
-- COMMENT ON COLUMN config_definitions.data_type IS 'Data type constraint for configuration values (string, integer, boolean, decimal, json)';
-- COMMENT ON COLUMN config_definitions.default_value IS 'Default value for this configuration in JSONB format';
-- COMMENT ON COLUMN config_definitions.validation_rules IS 'JSON schema or validation rules for the configuration value';
-- COMMENT ON COLUMN config_definitions.description IS 'Human-readable description of the configuration purpose';
-- COMMENT ON COLUMN config_definitions.required_permission IS 'Permission required to modify this configuration';
-- COMMENT ON COLUMN config_definitions.required_feature_flag IS 'Feature flag that must be enabled for this configuration';
-- COMMENT ON COLUMN config_definitions.is_overridable IS 'Whether this configuration can be overridden at tenant/entity levels';
--
-- =====================================================================
-- CONFIGURATION TEMPLATES - Bulk configuration deployment
-- =====================================================================
CREATE TABLE configuration_templates (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- NULL for SYSTEM-scoped templates; NOT NULL for TENANT-scoped templates.
  -- Enforced by the scope_tenant_check constraint below.
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  -- SYSTEM = platform-wide (readable by all, writable only by admin_role).
  -- TENANT = private to the owning tenant.
  scope     VARCHAR(10) NOT NULL DEFAULT 'TENANT'
              CHECK (scope IN ('SYSTEM', 'TENANT')),
  name      VARCHAR(255) NOT NULL,
  category  VARCHAR(50)  NOT NULL CHECK (category IN ('INDUSTRY', 'FUNCTIONAL', 'REGIONAL')),
  description TEXT,
  version   VARCHAR(20) NOT NULL,
  configurations        JSONB    NOT NULL,
  applicable_tenant_types TEXT[],
  required_feature_flags  TEXT[],
  conflict_resolution VARCHAR(20) DEFAULT 'MERGE'
                        CHECK (conflict_resolution IN ('MERGE', 'REPLACE', 'PRESERVE')),
  is_active  BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by UUID NOT NULL,
  -- SYSTEM templates have tenant_id=NULL; TENANT templates must have a tenant_id.
  CONSTRAINT configuration_templates_scope_tenant_check CHECK (
    (scope = 'SYSTEM' AND tenant_id IS NULL)
    OR (scope = 'TENANT' AND tenant_id IS NOT NULL)
  )
);

-- Add table and column comments
COMMENT ON TABLE configuration_templates IS 'Reusable configuration templates for bulk deployment across tenants and entities';

COMMENT ON COLUMN configuration_templates.id        IS 'UUID primary key for the configuration template';
COMMENT ON COLUMN configuration_templates.tenant_id IS 'Owning tenant for TENANT-scoped templates; NULL for SYSTEM-scoped templates';
COMMENT ON COLUMN configuration_templates.scope     IS 'SYSTEM = platform-wide, readable by all. TENANT = private to owner tenant.';
COMMENT ON COLUMN configuration_templates.name      IS 'Template display name';
COMMENT ON COLUMN configuration_templates.category  IS 'Template category: INDUSTRY, FUNCTIONAL, or REGIONAL';
COMMENT ON COLUMN configuration_templates.version   IS 'Semantic version string for template versioning';
COMMENT ON COLUMN configuration_templates.configurations        IS 'JSONB object of module.key → value pairs applied by this template';
COMMENT ON COLUMN configuration_templates.applicable_tenant_types IS 'Tenant types this template is designed for';
COMMENT ON COLUMN configuration_templates.required_feature_flags  IS 'Feature flags that must be enabled for this template to be applicable';
COMMENT ON COLUMN configuration_templates.conflict_resolution IS 'How to handle key conflicts on apply: MERGE, REPLACE, or PRESERVE';
COMMENT ON COLUMN configuration_templates.is_active IS 'Active templates appear in the UI. Deactivate instead of deleting applied templates.';
COMMENT ON COLUMN configuration_templates.created_by IS 'UUID of user who created this template';

-- =====================================================================
-- CONFIGURATION AUDIT - Complete audit trail for all changes
-- =====================================================================
CREATE TABLE configuration_audit (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id       UUID REFERENCES entities(uuid) ON DELETE SET NULL,
  -- Separate columns enable B-tree indexing on each part.
  -- Replaces the legacy combined "module.key" string pattern.
  module_name     VARCHAR(50)  NOT NULL,
  config_key_name VARCHAR(100) NOT NULL,
  old_value       JSONB,
  new_value       JSONB,
  source    VARCHAR(20) NOT NULL CHECK (source    IN ('SYSTEM', 'TENANT', 'ENTITY', 'TEMPLATE')),
  operation VARCHAR(20) NOT NULL CHECK (operation IN ('CREATE', 'UPDATE', 'DELETE', 'RESET', 'TEMPLATE_APPLY')),
  user_id         UUID      NOT NULL,
  applied_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  session_id      VARCHAR(100),
  correlation_id  VARCHAR(100)
);

-- Add table and column comments
COMMENT ON TABLE configuration_audit IS 'Complete audit trail of all configuration changes for compliance and troubleshooting';

COMMENT ON COLUMN configuration_audit.id              IS 'UUID primary key for the audit record';
COMMENT ON COLUMN configuration_audit.tenant_id       IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN configuration_audit.entity_id       IS 'Optional FK to entities — set when the change is entity-level';
COMMENT ON COLUMN configuration_audit.module_name     IS 'Module owning this config key (e.g. finance, hr, inventory)';
COMMENT ON COLUMN configuration_audit.config_key_name IS 'Config key within the module (e.g. invoice_prefix, valuation_method)';
COMMENT ON COLUMN configuration_audit.old_value       IS 'Previous configuration value (JSONB)';
COMMENT ON COLUMN configuration_audit.new_value       IS 'New configuration value (JSONB)';
COMMENT ON COLUMN configuration_audit.source          IS 'Level where change occurred: SYSTEM, TENANT, ENTITY, TEMPLATE';
COMMENT ON COLUMN configuration_audit.operation       IS 'Operation type: CREATE, UPDATE, DELETE, RESET, TEMPLATE_APPLY';
COMMENT ON COLUMN configuration_audit.user_id         IS 'UUID of user who made the change';
COMMENT ON COLUMN configuration_audit.session_id      IS 'Session identifier for grouping related changes';
COMMENT ON COLUMN configuration_audit.correlation_id  IS 'Correlation ID for tracking bulk/template operations';

-- =====================================================================
-- TEMPLATE APPLICATIONS - History of template deployments
-- =====================================================================
CREATE TABLE template_applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- ON DELETE RESTRICT: a template that has been applied cannot be deleted.
  -- Deactivate templates (is_active=false) instead — preserves deployment history for compliance.
  template_id UUID NOT NULL REFERENCES configuration_templates(id) ON DELETE RESTRICT,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE SET NULL,
  target_type VARCHAR(10) NOT NULL CHECK (target_type IN ('TENANT', 'ENTITY')),
  applied_configs INTEGER NOT NULL DEFAULT 0,
  skipped_configs INTEGER NOT NULL DEFAULT 0,
  conflict_count INTEGER NOT NULL DEFAULT 0,
  application_summary JSONB,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  applied_by UUID NOT NULL,
  correlation_id VARCHAR(100)
);

-- Add table and column comments
COMMENT ON TABLE template_applications IS 'History of template applications with detailed results and statistics';

COMMENT ON COLUMN template_applications.id IS 'UUID primary key for the template application record';
COMMENT ON COLUMN template_applications.template_id IS 'Foreign key to configuration_templates table';
COMMENT ON COLUMN template_applications.tenant_id IS 'Foreign key to tenants table';
COMMENT ON COLUMN template_applications.entity_id IS 'Optional foreign key to entities table for entity-level applications';
COMMENT ON COLUMN template_applications.target_type IS 'Target type: tenant or entity';
COMMENT ON COLUMN template_applications.applied_configs IS 'Number of configurations successfully applied';
COMMENT ON COLUMN template_applications.skipped_configs IS 'Number of configurations skipped due to conflicts or policies';
COMMENT ON COLUMN template_applications.conflict_count IS 'Number of configuration conflicts encountered';
COMMENT ON COLUMN template_applications.application_summary IS 'Detailed JSON summary of the application results';
COMMENT ON COLUMN template_applications.applied_by IS 'UUID of user who applied the template';
COMMENT ON COLUMN template_applications.correlation_id IS 'Correlation ID for tracking related operations';

-- =====================================================================
-- ENHANCE EXISTING TABLES - Add settings integration columns
-- =====================================================================
-- Enhance tenant_configurations table for better settings integration
ALTER TABLE tenant_configurations ADD COLUMN IF NOT EXISTS settings_version INTEGER DEFAULT 1;
ALTER TABLE tenant_configurations ADD COLUMN IF NOT EXISTS last_template_applied UUID REFERENCES configuration_templates(id) ON DELETE SET NULL;
ALTER TABLE tenant_configurations ADD COLUMN IF NOT EXISTS template_applied_at TIMESTAMPTZ;

-- Add comments for new columns
COMMENT ON COLUMN tenant_configurations.settings_version IS 'Version counter for optimistic locking of tenant settings';
COMMENT ON COLUMN tenant_configurations.last_template_applied IS 'Reference to last template applied to this tenant';
COMMENT ON COLUMN tenant_configurations.template_applied_at IS 'Timestamp when template was last applied';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Config definitions indexes
CREATE INDEX idx_config_definitions_module ON config_definitions(module_name);
CREATE INDEX idx_config_definitions_module_key ON config_definitions(module_name, config_key);
-- Configuration templates indexes
CREATE INDEX idx_configuration_templates_category   ON configuration_templates(category);
CREATE INDEX idx_configuration_templates_active     ON configuration_templates(is_active) WHERE is_active = true;
CREATE INDEX idx_configuration_templates_created_by ON configuration_templates(created_by);
CREATE INDEX idx_configuration_templates_scope      ON configuration_templates(scope, is_active) WHERE is_active = true;
CREATE INDEX idx_configuration_templates_tenant     ON configuration_templates(tenant_id, scope, is_active)
  WHERE tenant_id IS NOT NULL AND is_active = true;

-- Unique name+version per scope (SYSTEM templates unique globally; TENANT templates unique per tenant)
CREATE UNIQUE INDEX configuration_templates_name_version_unique
  ON configuration_templates (name, version, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid));

-- Configuration audit indexes
CREATE INDEX idx_configuration_audit_tenant      ON configuration_audit(tenant_id);
CREATE INDEX idx_configuration_audit_entity      ON configuration_audit(tenant_id, entity_id) WHERE entity_id IS NOT NULL;
-- Per-column indexes replace the old combined config_key index; each supports B-tree seeks
CREATE INDEX idx_configuration_audit_module      ON configuration_audit(tenant_id, module_name, applied_at);
CREATE INDEX idx_configuration_audit_key         ON configuration_audit(tenant_id, module_name, config_key_name, applied_at);
CREATE INDEX idx_configuration_audit_correlation ON configuration_audit(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_configuration_audit_user_time   ON configuration_audit(user_id, applied_at);

-- Template applications indexes
CREATE INDEX idx_template_applications_template ON template_applications(template_id, applied_at);
CREATE INDEX idx_template_applications_tenant ON template_applications(tenant_id, applied_at);
CREATE INDEX idx_template_applications_correlation ON template_applications(correlation_id) WHERE correlation_id IS NOT NULL;

-- Enhanced tenant_configurations indexes
CREATE INDEX idx_tenant_configurations_template ON tenant_configurations(tenant_id, last_template_applied) WHERE last_template_applied IS NOT NULL;
CREATE INDEX idx_tenant_configurations_settings_version ON tenant_configurations(tenant_id, settings_version);

-- Entity settings index (leverages existing entities.settings)
CREATE INDEX idx_entities_settings_tenant ON entities(tenant_id) INCLUDE (settings) WHERE deleted_at IS NULL AND settings IS NOT NULL;

-- JSONB GIN indexes for efficient configuration lookup
CREATE INDEX idx_tenant_configurations_settings_gin ON tenant_configurations USING gin(settings);
-- CREATE INDEX idx_entities_settings_gin ON entities USING gin(settings) WHERE settings IS NOT NULL;
CREATE INDEX idx_configuration_templates_configs_gin ON configuration_templates USING gin(configurations);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================

-- Ensure applied configs counts are non-negative
ALTER TABLE template_applications ADD CONSTRAINT valid_applied_configs CHECK (applied_configs >= 0);
ALTER TABLE template_applications ADD CONSTRAINT valid_skipped_configs CHECK (skipped_configs >= 0);
ALTER TABLE template_applications ADD CONSTRAINT valid_conflict_count CHECK (conflict_count >= 0);

-- Ensure settings version is positive
ALTER TABLE tenant_configurations ADD CONSTRAINT valid_settings_version CHECK (settings_version > 0);

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable RLS on all new tables
ALTER TABLE config_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE configuration_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE configuration_audit ENABLE ROW LEVEL SECURITY;
ALTER TABLE template_applications ENABLE ROW LEVEL SECURITY;

-- Config definitions are globally readable, only system admins can modify
CREATE POLICY config_definitions_read ON config_definitions 
  FOR SELECT TO application_role USING (true);

CREATE POLICY config_definitions_modify ON config_definitions 
  FOR ALL TO admin_role USING (true) WITH CHECK (true);

-- SYSTEM-scoped templates are readable by all tenants.
-- TENANT-scoped templates are readable/writable only by the owning tenant.
CREATE POLICY configuration_templates_read ON configuration_templates
  FOR SELECT TO application_role
  USING (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  );

CREATE POLICY configuration_templates_insert ON configuration_templates
  FOR INSERT TO application_role
  WITH CHECK (
    scope = 'TENANT'
    AND current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY configuration_templates_update ON configuration_templates
  FOR UPDATE TO application_role
  USING  (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- admin_role: full access including SYSTEM-scoped templates
CREATE POLICY configuration_templates_modify ON configuration_templates
  FOR ALL TO admin_role USING (true) WITH CHECK (true);

-- Configuration audit is tenant-isolated
CREATE POLICY configuration_audit_tenant_isolation ON configuration_audit 
  FOR ALL TO application_role 
  USING (
    current_tenant_id() IS NOT NULL 
    AND tenant_id = current_tenant_id()
  ) 
  WITH CHECK (
    current_tenant_id() IS NOT NULL 
    AND tenant_id = current_tenant_id()
  );

-- Admin bypass for configuration audit
CREATE POLICY configuration_audit_admin_access ON configuration_audit 
  FOR ALL TO admin_role USING (true) WITH CHECK (true);

-- Template applications are tenant-isolated
CREATE POLICY template_applications_tenant_isolation ON template_applications 
  FOR ALL TO application_role 
  USING (
    current_tenant_id() IS NOT NULL 
    AND tenant_id = current_tenant_id()
  ) 
  WITH CHECK (
    current_tenant_id() IS NOT NULL 
    AND tenant_id = current_tenant_id()
  );

-- Admin bypass for template applications
CREATE POLICY template_applications_admin_access ON template_applications 
  FOR ALL TO admin_role USING (true) WITH CHECK (true);

-- =====================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================================

-- Apply the existing update_updated_at_column trigger to new tables
-- config_definitions trigger may already exist from an earlier migration — drop first to be safe
DROP TRIGGER IF EXISTS update_config_definitions_updated_at ON config_definitions;
CREATE TRIGGER update_config_definitions_updated_at
  BEFORE UPDATE ON config_definitions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_configuration_templates_updated_at 
  BEFORE UPDATE ON configuration_templates 
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON config_definitions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON configuration_templates TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON configuration_audit TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON template_applications TO application_role;
