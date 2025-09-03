-- ================================================================================================
-- SETTINGS MODULE ROLLBACK - Remove all settings-related tables and modifications
-- ================================================================================================

-- =====================================================================
-- REMOVE PERMISSIONS AND GRANTS
-- =====================================================================
REVOKE ALL ON template_applications FROM application_role;
REVOKE ALL ON configuration_audit FROM application_role;
REVOKE ALL ON configuration_templates FROM application_role;
REVOKE ALL ON config_definitions FROM application_role;

-- =====================================================================
-- DROP TRIGGERS
-- =====================================================================
DROP TRIGGER IF EXISTS update_configuration_templates_updated_at ON configuration_templates;
DROP TRIGGER IF EXISTS update_config_definitions_updated_at ON config_definitions;

-- =====================================================================
-- DROP RLS POLICIES
-- =====================================================================
DROP POLICY IF EXISTS template_applications_admin_access ON template_applications;
DROP POLICY IF EXISTS template_applications_tenant_isolation ON template_applications;
DROP POLICY IF EXISTS configuration_audit_admin_access ON configuration_audit;
DROP POLICY IF EXISTS configuration_audit_tenant_isolation ON configuration_audit;
DROP POLICY IF EXISTS configuration_templates_modify ON configuration_templates;
DROP POLICY IF EXISTS configuration_templates_read ON configuration_templates;
DROP POLICY IF EXISTS config_definitions_modify ON config_definitions;
DROP POLICY IF EXISTS config_definitions_read ON config_definitions;

-- =====================================================================
-- DROP INDEXES
-- =====================================================================

-- JSONB GIN indexes
DROP INDEX IF EXISTS idx_configuration_templates_configs_gin;
-- DROP INDEX IF EXISTS idx_entities_settings_gin;
DROP INDEX IF EXISTS idx_tenant_configurations_settings_gin;

-- Enhanced tenant_configurations indexes
DROP INDEX IF EXISTS idx_tenant_configurations_settings_version;
DROP INDEX IF EXISTS idx_tenant_configurations_template;

-- Entity settings index
DROP INDEX IF EXISTS idx_entities_settings_tenant;

-- Template applications indexes
DROP INDEX IF EXISTS idx_template_applications_correlation;
DROP INDEX IF EXISTS idx_template_applications_tenant;
DROP INDEX IF EXISTS idx_template_applications_template;

-- Configuration audit indexes
DROP INDEX IF EXISTS idx_configuration_audit_user_time;
DROP INDEX IF EXISTS idx_configuration_audit_correlation;
DROP INDEX IF EXISTS idx_configuration_audit_config_key;
DROP INDEX IF EXISTS idx_configuration_audit_entity;
DROP INDEX IF EXISTS idx_configuration_audit_tenant;

-- Configuration templates indexes
DROP INDEX IF EXISTS idx_configuration_templates_created_by;
DROP INDEX IF EXISTS idx_configuration_templates_active;
DROP INDEX IF EXISTS idx_configuration_templates_category;

-- Config definitions indexes
DROP INDEX IF EXISTS idx_config_definitions_overridable;
DROP INDEX IF EXISTS idx_config_definitions_module_key;
DROP INDEX IF EXISTS idx_config_definitions_module;

-- =====================================================================
-- DROP CONSTRAINTS
-- =====================================================================

-- Remove constraints from tenant_configurations
ALTER TABLE tenant_configurations DROP CONSTRAINT IF EXISTS valid_settings_version;

-- Remove constraints from template_applications
ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS valid_conflict_count;
ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS valid_skipped_configs;
ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS valid_applied_configs;

-- =====================================================================
-- REMOVE ENHANCED COLUMNS FROM EXISTING TABLES
-- =====================================================================
ALTER TABLE tenant_configurations DROP COLUMN IF EXISTS template_applied_at;
ALTER TABLE tenant_configurations DROP COLUMN IF EXISTS last_template_applied;
ALTER TABLE tenant_configurations DROP COLUMN IF EXISTS settings_version;

-- =====================================================================
-- DROP TABLES IN REVERSE DEPENDENCY ORDER
-- =====================================================================
DROP TABLE IF EXISTS template_applications;
DROP TABLE IF EXISTS configuration_audit;
DROP TABLE IF EXISTS configuration_templates;
DROP TABLE IF EXISTS config_definitions;
