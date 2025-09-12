-- ================================================================================================
-- SETTINGS MODULE - Configuration management with 3-level inheritance (System → Tenant → Entity)
-- ================================================================================================
--
-- Core tables for ERP Settings Module implementing configuration inheritance, templates,
-- and comprehensive audit trails for enterprise configuration management.
--
-- Prerequisites:
-- - tenants table with UUID primary key
-- - entities table with UUID primary key
-- ================================================================================================

-- =====================================================================
-- CONFIG DEFINITIONS - System-wide metadata for all configurations
-- =====================================================================
CREATE TABLE config_definitions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  module_name VARCHAR(50) NOT NULL,
  config_key VARCHAR(100) NOT NULL,
  data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('STRING', 'INTEGER', 'BOOLEAN', 'DECIMAL', 'JSON')),
  default_value JSONB,
  validation_rules JSONB DEFAULT '{}'::jsonb,
  description TEXT,
  required_permission VARCHAR(100),
  required_feature_flag VARCHAR(100),
  is_overridable BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Ensure unique configuration keys per module
  CONSTRAINT config_definitions_module_key_unique UNIQUE (module_name, config_key)
);

-- Add table and column comments
COMMENT ON TABLE config_definitions IS 'System-wide configuration metadata defining all possible configuration keys with validation rules and inheritance policies';

COMMENT ON COLUMN config_definitions.id IS 'UUID primary key for the configuration definition';
COMMENT ON COLUMN config_definitions.module_name IS 'ERP module that owns this configuration (finance, hr, inventory, etc.)';
COMMENT ON COLUMN config_definitions.config_key IS 'Unique configuration key within the module namespace';
COMMENT ON COLUMN config_definitions.data_type IS 'Data type constraint for configuration values (string, integer, boolean, decimal, json)';
COMMENT ON COLUMN config_definitions.default_value IS 'Default value for this configuration in JSONB format';
COMMENT ON COLUMN config_definitions.validation_rules IS 'JSON schema or validation rules for the configuration value';
COMMENT ON COLUMN config_definitions.description IS 'Human-readable description of the configuration purpose';
COMMENT ON COLUMN config_definitions.required_permission IS 'Permission required to modify this configuration';
COMMENT ON COLUMN config_definitions.required_feature_flag IS 'Feature flag that must be enabled for this configuration';
COMMENT ON COLUMN config_definitions.is_overridable IS 'Whether this configuration can be overridden at tenant/entity levels';

-- =====================================================================
-- CONFIGURATION TEMPLATES - Bulk configuration deployment
-- =====================================================================
CREATE TABLE configuration_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  category VARCHAR(50) NOT NULL CHECK (category IN ('INDUSTRY', 'FUNCTIONAL', 'REGIONAL')),
  description TEXT,
  version VARCHAR(20) NOT NULL,
  configurations JSONB NOT NULL,
  applicable_tenant_types TEXT[],
  required_feature_flags TEXT[],
  conflict_resolution VARCHAR(20) DEFAULT 'MERGE' CHECK (conflict_resolution IN ('MERGE', 'REPLACE', 'PRESERVE')),
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by UUID NOT NULL,
  -- Ensure unique template name+version combinations
  CONSTRAINT configuration_templates_name_version_unique UNIQUE (name, version)
);

-- Add table and column comments
COMMENT ON TABLE configuration_templates IS 'Reusable configuration templates for bulk deployment across tenants and entities';

COMMENT ON COLUMN configuration_templates.id IS 'UUID primary key for the configuration template';
COMMENT ON COLUMN configuration_templates.name IS 'Template display name';
COMMENT ON COLUMN configuration_templates.category IS 'Template category: industry, functional, or regional';
COMMENT ON COLUMN configuration_templates.version IS 'Semantic version string for template versioning';
COMMENT ON COLUMN configuration_templates.configurations IS 'JSON object containing all configuration key-value pairs';
COMMENT ON COLUMN configuration_templates.applicable_tenant_types IS 'Array of tenant types this template applies to';
COMMENT ON COLUMN configuration_templates.required_feature_flags IS 'Array of feature flags required for this template';
COMMENT ON COLUMN configuration_templates.conflict_resolution IS 'Strategy for handling configuration conflicts: merge, replace, or preserve';
COMMENT ON COLUMN configuration_templates.is_active IS 'Whether this template is active and available for use';
COMMENT ON COLUMN configuration_templates.created_by IS 'UUID of user who created this template';

-- =====================================================================
-- CONFIGURATION AUDIT - Complete audit trail for all changes
-- =====================================================================
CREATE TABLE configuration_audit (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE SET NULL,
  config_key VARCHAR(150) NOT NULL,
  old_value JSONB,
  new_value JSONB,
source VARCHAR(20) NOT NULL CHECK (source IN ('SYSTEM', 'TENANT', 'ENTITY', 'TEMPLATE')),
operation VARCHAR(20) NOT NULL CHECK (operation IN ('CREATE', 'UPDATE', 'DELETE', 'RESET', 'TEMPLATE_APPLY')),
  user_id UUID NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  session_id VARCHAR(100),
  correlation_id VARCHAR(100)
);

-- Add table and column comments
COMMENT ON TABLE configuration_audit IS 'Complete audit trail of all configuration changes for compliance and troubleshooting';

COMMENT ON COLUMN configuration_audit.id IS 'UUID primary key for the audit record';
COMMENT ON COLUMN configuration_audit.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN configuration_audit.entity_id IS 'Optional foreign key to entities table for entity-level changes';
COMMENT ON COLUMN configuration_audit.config_key IS 'Full configuration key (module.key) that was modified';
COMMENT ON COLUMN configuration_audit.old_value IS 'Previous configuration value in JSONB format';
COMMENT ON COLUMN configuration_audit.new_value IS 'New configuration value in JSONB format';
COMMENT ON COLUMN configuration_audit.source IS 'Source level where change occurred: system, tenant, entity, or template';
COMMENT ON COLUMN configuration_audit.operation IS 'Type of operation: create, update, delete, reset, or template_apply';
COMMENT ON COLUMN configuration_audit.user_id IS 'UUID of user who made the change';
COMMENT ON COLUMN configuration_audit.session_id IS 'Session identifier for tracking related changes';
COMMENT ON COLUMN configuration_audit.correlation_id IS 'Correlation ID for tracking bulk operations';

-- =====================================================================
-- TEMPLATE APPLICATIONS - History of template deployments
-- =====================================================================
CREATE TABLE template_applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  template_id UUID NOT NULL REFERENCES configuration_templates(id) ON DELETE CASCADE,
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
CREATE INDEX idx_config_definitions_overridable ON config_definitions(module_name) WHERE is_overridable = true;

-- Configuration templates indexes
CREATE INDEX idx_configuration_templates_category ON configuration_templates(category);
CREATE INDEX idx_configuration_templates_active ON configuration_templates(is_active) WHERE is_active = true;
CREATE INDEX idx_configuration_templates_created_by ON configuration_templates(created_by);

-- Configuration audit indexes
CREATE INDEX idx_configuration_audit_tenant ON configuration_audit(tenant_id);
CREATE INDEX idx_configuration_audit_entity ON configuration_audit(tenant_id, entity_id) WHERE entity_id IS NOT NULL;
CREATE INDEX idx_configuration_audit_config_key ON configuration_audit(tenant_id, config_key, applied_at);
CREATE INDEX idx_configuration_audit_correlation ON configuration_audit(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_configuration_audit_user_time ON configuration_audit(user_id, applied_at);

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

-- Configuration templates are globally readable, only system admins can modify
CREATE POLICY configuration_templates_read ON configuration_templates 
  FOR SELECT TO application_role USING (true);

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
