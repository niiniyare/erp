-- ================================================================================================
-- DOWN MIGRATION: Reverting the Settings Module changes
-- ================================================================================================
-- This migration safely reverts all changes from 000070_settings_core_tables.up.sql
-- All operations use defensive programming to handle cases where objects may not exist

DO $$
DECLARE
    table_exists boolean;
BEGIN
    -- =====================================================================
    -- REVERT TRIGGERS
    -- =====================================================================
    -- Drop triggers only if tables exist
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'configuration_templates') INTO table_exists;
    IF table_exists THEN
        DROP TRIGGER IF EXISTS update_configuration_templates_updated_at ON configuration_templates;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'config_definitions') INTO table_exists;
    IF table_exists THEN
        DROP TRIGGER IF EXISTS update_config_definitions_updated_at ON config_definitions;
    END IF;

    -- =====================================================================
    -- REVERT ROW LEVEL SECURITY (RLS) POLICIES
    -- =====================================================================
    -- Drop policies only if tables exist
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'template_applications') INTO table_exists;
    IF table_exists THEN
        DROP POLICY IF EXISTS template_applications_admin_access ON template_applications;
        DROP POLICY IF EXISTS template_applications_tenant_isolation ON template_applications;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'configuration_audit') INTO table_exists;
    IF table_exists THEN
        DROP POLICY IF EXISTS configuration_audit_admin_access ON configuration_audit;
        DROP POLICY IF EXISTS configuration_audit_tenant_isolation ON configuration_audit;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'configuration_templates') INTO table_exists;
    IF table_exists THEN
        DROP POLICY IF EXISTS configuration_templates_modify ON configuration_templates;
        DROP POLICY IF EXISTS configuration_templates_read ON configuration_templates;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'config_definitions') INTO table_exists;
    IF table_exists THEN
        DROP POLICY IF EXISTS config_definitions_modify ON config_definitions;
        DROP POLICY IF EXISTS config_definitions_read ON config_definitions;
    END IF;

    -- =====================================================================
    -- DISABLE RLS ON TABLES
    -- =====================================================================
    -- Disable RLS on all tables (only if they exist)
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'template_applications') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE template_applications DISABLE ROW LEVEL SECURITY;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'configuration_audit') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE configuration_audit DISABLE ROW LEVEL SECURITY;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'configuration_templates') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE configuration_templates DISABLE ROW LEVEL SECURITY;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'config_definitions') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE config_definitions DISABLE ROW LEVEL SECURITY;
    END IF;

    -- =====================================================================
    -- REVERT DATA INTEGRITY CONSTRAINTS
    -- =====================================================================
    -- Drop constraints only if tables exist
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'tenant_configurations') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE tenant_configurations DROP CONSTRAINT IF EXISTS valid_settings_version;
    END IF;
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'template_applications') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS valid_conflict_count;
        ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS valid_skipped_configs;
        ALTER TABLE template_applications DROP CONSTRAINT IF EXISTS valid_applied_configs;
    END IF;

    -- =====================================================================
    -- REVERT PERFORMANCE OPTIMIZATION INDEXES
    -- =====================================================================
    -- Drop all indexes (IF EXISTS handles non-existent indexes)
    DROP INDEX IF EXISTS idx_configuration_templates_configs_gin;
    DROP INDEX IF EXISTS idx_entities_settings_gin;
    DROP INDEX IF EXISTS idx_tenant_configurations_settings_gin;
    DROP INDEX IF EXISTS idx_entities_settings_tenant;
    DROP INDEX IF EXISTS idx_tenant_configurations_settings_version;
    DROP INDEX IF EXISTS idx_tenant_configurations_template;
    DROP INDEX IF EXISTS idx_template_applications_correlation;
    DROP INDEX IF EXISTS idx_template_applications_tenant;
    DROP INDEX IF EXISTS idx_template_applications_template;
    DROP INDEX IF EXISTS idx_configuration_audit_user_time;
    DROP INDEX IF EXISTS idx_configuration_audit_correlation;
    DROP INDEX IF EXISTS idx_configuration_audit_config_key;
    DROP INDEX IF EXISTS idx_configuration_audit_entity;
    DROP INDEX IF EXISTS idx_configuration_audit_tenant;
    DROP INDEX IF EXISTS idx_configuration_templates_created_by;
    DROP INDEX IF EXISTS idx_configuration_templates_active;
    DROP INDEX IF EXISTS idx_configuration_templates_category;
    DROP INDEX IF EXISTS idx_config_definitions_overridable;
    DROP INDEX IF EXISTS idx_config_definitions_module_key;
    DROP INDEX IF EXISTS idx_config_definitions_module;

    -- =====================================================================
    -- REVERT ENHANCEMENTS TO EXISTING TABLES
    -- =====================================================================
    -- Drop columns only if table exists
    
    SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'tenant_configurations') INTO table_exists;
    IF table_exists THEN
        ALTER TABLE tenant_configurations DROP COLUMN IF EXISTS template_applied_at;
        ALTER TABLE tenant_configurations DROP COLUMN IF EXISTS last_template_applied;
        ALTER TABLE tenant_configurations DROP COLUMN IF EXISTS settings_version;
    END IF;

    -- =====================================================================
    -- DROP ALL NEW TABLES (CORRECT ORDER)
    -- =====================================================================
    -- Drop tables that reference others first (all use IF EXISTS for safety)
    DROP TABLE IF EXISTS template_applications;
    DROP TABLE IF EXISTS configuration_audit;

    -- Then drop the tables that were referenced
    DROP TABLE IF EXISTS configuration_templates;
    DROP TABLE IF EXISTS config_definitions;

END $$;
