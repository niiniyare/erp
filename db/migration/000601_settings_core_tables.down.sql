-- ================================================================================================
-- DOWN MIGRATION: Revert Settings Module core tables
-- ================================================================================================

-- Triggers
DROP TRIGGER IF EXISTS update_configuration_templates_updated_at ON configuration_templates;
DROP TRIGGER IF EXISTS update_config_definitions_updated_at ON config_definitions;

-- RLS policies
DROP POLICY IF EXISTS template_applications_admin_access       ON template_applications;
DROP POLICY IF EXISTS template_applications_tenant_isolation   ON template_applications;
DROP POLICY IF EXISTS template_applications_ro_select          ON template_applications;
DROP POLICY IF EXISTS configuration_audit_admin_access         ON configuration_audit;
DROP POLICY IF EXISTS configuration_audit_tenant_isolation     ON configuration_audit;
DROP POLICY IF EXISTS configuration_audit_ro_select            ON configuration_audit;
DROP POLICY IF EXISTS configuration_templates_modify           ON configuration_templates;
DROP POLICY IF EXISTS configuration_templates_update           ON configuration_templates;
DROP POLICY IF EXISTS configuration_templates_insert           ON configuration_templates;
DROP POLICY IF EXISTS configuration_templates_read             ON configuration_templates;
DROP POLICY IF EXISTS configuration_templates_ro_select        ON configuration_templates;
DROP POLICY IF EXISTS config_definitions_modify                ON config_definitions;
DROP POLICY IF EXISTS config_definitions_read                  ON config_definitions;
DROP POLICY IF EXISTS ts_tenant_isolation                      ON tenant_settings;
DROP POLICY IF EXISTS ts_admin_access                          ON tenant_settings;
DROP POLICY IF EXISTS ts_ro_select                             ON tenant_settings;
DROP POLICY IF EXISTS uprefs_self                              ON user_preferences;
DROP POLICY IF EXISTS uprefs_admin                             ON user_preferences;

-- Disable RLS
ALTER TABLE IF EXISTS template_applications    DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS configuration_audit      DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS configuration_templates  DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS config_definitions       DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS tenant_settings          DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS user_preferences         DISABLE ROW LEVEL SECURITY;

-- Constraints on enhanced columns
ALTER TABLE IF EXISTS tenant_configurations DROP CONSTRAINT IF EXISTS valid_settings_version;
ALTER TABLE IF EXISTS template_applications DROP CONSTRAINT IF EXISTS valid_conflict_count;
ALTER TABLE IF EXISTS template_applications DROP CONSTRAINT IF EXISTS valid_skipped_configs;
ALTER TABLE IF EXISTS template_applications DROP CONSTRAINT IF EXISTS valid_applied_configs;

-- Indexes
DROP INDEX IF EXISTS idx_configuration_templates_configs_gin;
DROP INDEX IF EXISTS idx_tenant_configurations_settings_gin;
DROP INDEX IF EXISTS idx_entities_settings_tenant;
DROP INDEX IF EXISTS idx_tenant_configurations_settings_version;
DROP INDEX IF EXISTS idx_tenant_configurations_template;
DROP INDEX IF EXISTS idx_template_applications_correlation;
DROP INDEX IF EXISTS idx_template_applications_tenant;
DROP INDEX IF EXISTS idx_template_applications_template;
DROP INDEX IF EXISTS idx_configuration_audit_user_time;
DROP INDEX IF EXISTS idx_configuration_audit_correlation;
DROP INDEX IF EXISTS idx_configuration_audit_key;
DROP INDEX IF EXISTS idx_configuration_audit_module;
DROP INDEX IF EXISTS idx_configuration_audit_entity;
DROP INDEX IF EXISTS idx_configuration_audit_tenant;
DROP INDEX IF EXISTS configuration_templates_name_version_unique;
DROP INDEX IF EXISTS idx_configuration_templates_tenant;
DROP INDEX IF EXISTS idx_configuration_templates_scope;
DROP INDEX IF EXISTS idx_configuration_templates_created_by;
DROP INDEX IF EXISTS idx_configuration_templates_active;
DROP INDEX IF EXISTS idx_configuration_templates_category;
DROP INDEX IF EXISTS idx_config_definitions_module_key;
DROP INDEX IF EXISTS idx_config_definitions_module;

-- Enhanced columns on existing tables
ALTER TABLE IF EXISTS tenant_configurations DROP COLUMN IF EXISTS template_applied_at;
ALTER TABLE IF EXISTS tenant_configurations DROP COLUMN IF EXISTS last_template_applied;
ALTER TABLE IF EXISTS tenant_configurations DROP COLUMN IF EXISTS settings_version;

-- Drop tables in FK-safe order (referencing tables first)
DROP TABLE IF EXISTS template_applications;
DROP TABLE IF EXISTS configuration_audit;
DROP TABLE IF EXISTS configuration_templates;

-- IAM settings tables
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS tenant_settings;
DROP TABLE IF EXISTS setting_definitions;
