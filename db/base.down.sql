-- === Drop Table ===
DROP TABLE IF EXISTS tenants CASCADE;

-- === Drop RLS Policies ===
DROP POLICY IF EXISTS tenant_isolation_policy ON tenants;

DROP POLICY IF EXISTS tenant_isolation_policy ON tenant_configurations;

-- === Drop Defaults Using current_tenant_id() ===
DO
$$
BEGIN
IF EXISTS (
  SELECT
    1
  FROM
    information_schema.columns
  WHERE
    table_name = 'tenants'
    AND column_name = 'tenant_id'
) THEN
ALTER TABLE
  tenants
ALTER COLUMN
  tenant_id DROP DEFAULT;

END IF;

END
$$
;

DO
$$
BEGIN
IF EXISTS (
  SELECT
    1
  FROM
    information_schema.columns
  WHERE
    table_name = 'tenant_configurations'
    AND column_name = 'tenant_id'
) THEN
ALTER TABLE
  tenant_configurations
ALTER COLUMN
  tenant_id DROP DEFAULT;

END IF;

END
$$
;

-- === Drop Views that use current_tenant_id() ===
DROP VIEW IF EXISTS active_tenant_configurations;

DROP VIEW IF EXISTS visible_tenants;

-- === Drop Trigger Functions or Others ===
DROP FUNCTION IF EXISTS enforce_tenant_access();

-- === Drop the Function ===
DROP FUNCTION IF EXISTS current_tenant_id();
-- =====================================================
-- TENANT CONFIGURATIONS DOWN MIGRATION
-- =====================================================
-- Drop trigger on tenants table (from another migration)
DROP TRIGGER IF EXISTS create_tenant_configuration_trigger ON tenants;

-- Drop configuration-specific functions
DROP FUNCTION IF EXISTS create_tenant_configuration_on_insert();

DROP FUNCTION IF EXISTS create_default_tenant_configuration(UUID);

DROP FUNCTION IF EXISTS check_tenant_limits(UUID, VARCHAR, INT);

-- Drop trigger on tenant_configurations table
DROP TRIGGER IF EXISTS update_tenant_configurations_updated_at ON tenant_configurations;

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_configurations_isolation_policy ON tenant_configurations;

-- Disable RLS
ALTER TABLE
  IF EXISTS tenant_configurations DISABLE ROW LEVEL SECURITY;

-- Drop table
DROP TABLE IF EXISTS tenant_configurations;

-- Note: Shared functions like update_updated_at_column() and current_tenant_id()
-- are managed by the core tenants migration and should not be dropped here
-- =====================================================
-- TENANT USAGE STATS DOWN MIGRATION
-- =====================================================
-- Revert check_tenant_limits function to previous version
CREATE
OR REPLACE FUNCTION check_tenant_limits(
  p_tenant_id UUID,
  p_check_type VARCHAR(50),
  p_additional_usage INT DEFAULT 1
) RETURNS BOOLEAN AS
$$
DECLARE
v_config tenant_configurations % ROWTYPE;

v_current_usage INT;

BEGIN
-- Get tenant configuration
SELECT
  * INTO v_config
FROM
  tenant_configurations
WHERE
  tenant_id = p_tenant_id;

-- If no configuration found, deny operation
IF v_config.tenant_id IS NULL THEN RAISE NOTICE 'No configuration found for tenant: %',
p_tenant_id;

RETURN FALSE;

END IF;

-- Check different types of limits
CASE
  p_check_type
  WHEN 'users' THEN
  -- Check user limit (placeholder for when users table exists)
  RETURN TRUE;

WHEN 'entities' THEN
-- Check entity limit (placeholder for when entities table exists)
RETURN TRUE;

WHEN 'transactions' THEN
-- Check monthly transaction limit (placeholder)
RETURN TRUE;

WHEN 'storage' THEN
-- Check storage limit using tenant_usage_stats
-- This will be fully functional after tenant_usage_stats migration
RETURN TRUE;

ELSE
-- Unknown check type
RAISE NOTICE 'Unknown check type: %',
p_check_type;

RETURN FALSE;

END CASE
;

END;

$$
LANGUAGE plpgsql;

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_usage_stats_isolation_policy ON tenant_usage_stats;

-- Disable RLS
ALTER TABLE
  IF EXISTS tenant_usage_stats DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_tenant_usage_stats_tenant_period;

DROP INDEX IF EXISTS idx_tenant_usage_stats_period;

-- Drop table
DROP TABLE IF EXISTS tenant_usage_stats;
DROP FUNCTION IF EXISTS provision_tenant_complete(
  VARCHAR,
  VARCHAR,
  VARCHAR,
  VARCHAR,
  VARCHAR,
  CHAR,
  VARCHAR,
  JSONB
);
-- =====================================================
-- TENANT BULK OPERATIONS TRACKING ROLLBACK
-- =====================================================
-- 
-- PURPOSE:
-- This rollback migration safely removes all tenant bulk operations 
-- tracking infrastructure created by the corresponding up migration.
-- 
-- REMOVAL ORDER:
-- 1. Drop triggers (dependent on functions)
-- 2. Drop functions (dependent on tables)
-- 3. Drop indexes (performance optimization cleanup)
-- 4. Drop tables (in dependency order)
-- 
-- SAFETY:
-- - Uses IF EXISTS to prevent errors if objects don't exist
-- - Removes dependencies in correct order
-- - Preserves referential integrity
-- 
-- AUTHOR: ERP System Migration
-- VERSION: 1.0
-- DATE: 2024
-- =====================================================

-- Log rollback start
DO $$
BEGIN
  RAISE NOTICE 'Starting tenant bulk operations tracking rollback migration';
END $$;

-- =====================================================
-- DROP TRIGGERS
-- =====================================================

-- Drop automatic count update trigger
DROP TRIGGER IF EXISTS tenant_bulk_operation_results_update_counts ON tenant_bulk_operation_results;

-- Drop timestamp update triggers
DROP TRIGGER IF EXISTS update_tenant_bulk_operation_results_updated_at ON tenant_bulk_operation_results;
DROP TRIGGER IF EXISTS update_tenant_bulk_operations_updated_at ON tenant_bulk_operations;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================

-- Drop trigger function
DROP FUNCTION IF EXISTS trigger_update_bulk_operation_counts();

-- Drop utility functions
DROP FUNCTION IF EXISTS update_bulk_operation_counts(UUID);
DROP FUNCTION IF EXISTS get_bulk_operation_summary(UUID);

-- =====================================================
-- DROP INDEXES
-- =====================================================

-- Drop indexes for tenant_bulk_operation_results table
DROP INDEX IF EXISTS idx_tenant_bulk_operation_results_status;
DROP INDEX IF EXISTS idx_tenant_bulk_operation_results_tenant;

-- Drop indexes for tenant_bulk_operations table
DROP INDEX IF EXISTS idx_tenant_bulk_operations_created_at;
DROP INDEX IF EXISTS idx_tenant_bulk_operations_type;
DROP INDEX IF EXISTS idx_tenant_bulk_operations_status;
DROP INDEX IF EXISTS idx_tenant_bulk_operations_actor;

-- =====================================================
-- DROP TABLES
-- =====================================================

-- Drop tables in dependency order (child tables first)
DROP TABLE IF EXISTS tenant_bulk_operation_results;
DROP TABLE IF EXISTS tenant_bulk_operations;

-- =====================================================
-- ROLLBACK COMPLETION
-- =====================================================

-- Log successful rollback completion
DO $$
BEGIN
  RAISE NOTICE 'Tenant bulk operations tracking rollback completed successfully';
  RAISE NOTICE 'Removed: triggers, functions, indexes, and tables';
  RAISE NOTICE 'Database restored to state before bulk operations tracking migration';
END $$;-- =====================================================================
-- ENTITIES CORE DOWN MIGRATION
-- =====================================================================
--
-- This down migration properly cleans up the entities core table
--
-- Order of operations:
-- 1. Drop table RLS policies
-- 2. Disable RLS on table
-- 3. Drop triggers and functions
-- 4. Drop indexes
-- 5. Drop table constraints
-- 6. Drop table
-- =====================================================================
-- =====================================================================
-- DROP TABLE RLS POLICIES
-- =====================================================================
-- Drop admin bypass policy
DROP POLICY IF EXISTS admin_full_access_policy ON entities;

-- Drop tenant isolation policy
DROP POLICY IF EXISTS tenant_isolation_policy ON entities;

-- =====================================================================
-- DISABLE RLS ON TABLE
-- =====================================================================
ALTER TABLE
  IF EXISTS entities DISABLE ROW LEVEL SECURITY;

-- =====================================================================
-- DROP TRIGGERS AND FUNCTIONS
-- =====================================================================
-- Drop trigger
DROP TRIGGER IF EXISTS entities_maintain_entity_id ON entities;

-- Drop entity-specific trigger function
DROP FUNCTION IF EXISTS maintain_entity_id();

-- Note: Shared functions like current_tenant_id() are managed by tenants_core migration
-- =====================================================================
-- DROP INDEXES
-- =====================================================================
-- Drop performance indexes
DROP INDEX IF EXISTS idx_entities_type;

DROP INDEX IF EXISTS idx_entities_parent;

DROP INDEX IF EXISTS idx_entities_tenant;

DROP INDEX IF EXISTS idx_entities_deleted_at;

DROP INDEX IF EXISTS idx_entities_address_gin;

DROP INDEX IF EXISTS idx_entities_settings_gin;

DROP INDEX IF EXISTS idx_entities_active;

DROP INDEX IF EXISTS idx_entities_parent_id;

DROP INDEX IF EXISTS idx_entities_tenant_type;

DROP INDEX IF EXISTS tenant_code_unique_idx;

-- =====================================================================
-- DROP TABLE CONSTRAINTS
-- =====================================================================
-- Drop check constraints
ALTER TABLE
  IF EXISTS entities DROP CONSTRAINT IF EXISTS valid_fy_start_month;

ALTER TABLE
  IF EXISTS entities DROP CONSTRAINT IF EXISTS no_self_parent;

-- =====================================================================
-- DROP TABLE
-- =====================================================================
-- Drop entities table
DROP TABLE IF EXISTS entities;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITIES CORE DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (entities)';

RAISE NOTICE '- All associated indexes, triggers, and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '===================================================================';

END;

$$
;
-- =====================================================================
-- ENTITIES HIERARCHY DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS admin_full_access_policy ON hierarchy_paths;

DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;

-- Disable RLS
ALTER TABLE
  IF EXISTS hierarchy_paths DISABLE ROW LEVEL SECURITY;

-- Drop triggers
DROP TRIGGER IF EXISTS hierarchy_paths_maintain_entity_id ON hierarchy_paths;

-- Drop indexes
DROP INDEX IF EXISTS idx_hierarchy_paths_ancestor;

DROP INDEX IF EXISTS idx_hierarchy_paths_tenant;

DROP INDEX IF EXISTS idx_hierarchy_paths_depth;

DROP INDEX IF EXISTS idx_hierarchy_paths_descendant;

-- Drop table
DROP TABLE IF EXISTS hierarchy_paths;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITIES HIERARCHY DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (hierarchy_paths)';

RAISE NOTICE '- All associated indexes and triggers';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '===================================================================';

END;

$$
;
-- =====================================================================
-- ENTITY STATE DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS admin_full_access_policy ON entitystate;

DROP POLICY IF EXISTS tenant_isolation_policy ON entitystate;

-- Disable RLS
ALTER TABLE
  IF EXISTS entitystate DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS entitystate DROP CONSTRAINT IF EXISTS positive_sequence;

ALTER TABLE
  IF EXISTS entitystate DROP CONSTRAINT IF EXISTS unique_tenant_entity_key_fy;

-- Drop indexes
DROP INDEX IF EXISTS idx_entitystate_fiscal_year;

DROP INDEX IF EXISTS idx_entitystate_entity_key;

-- Drop table
DROP TABLE IF EXISTS entitystate;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITY STATE DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (entitystate)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '===================================================================';

END;

$$
;
-- =====================================================================
-- ENTITIES VIEWS DOWN MIGRATION
-- =====================================================================
-- Drop all entity views in reverse order of creation
DROP VIEW IF EXISTS v_tenant_resource_utilization;

DROP VIEW IF EXISTS v_entity_paths;

DROP VIEW IF EXISTS v_entity_changes;

DROP VIEW IF EXISTS v_active_entities;

DROP VIEW IF EXISTS v_tenant_entity_summary;

DROP VIEW IF EXISTS v_company_structure;

DROP VIEW IF EXISTS v_department_summary;

DROP VIEW IF EXISTS v_cost_center_info;

DROP VIEW IF EXISTS v_entity_structure;

DROP VIEW IF EXISTS v_tenant_hierarchy;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITIES VIEWS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 10 entity views for reporting and analytics';

RAISE NOTICE '- v_tenant_hierarchy (recursive hierarchy view)';

RAISE NOTICE '- v_entity_structure (organizational structure)';

RAISE NOTICE '- v_cost_center_info (cost center details)';

RAISE NOTICE '- v_department_summary (department aggregation)';

RAISE NOTICE '- v_company_structure (company organization)';

RAISE NOTICE '- v_tenant_entity_summary (tenant statistics)';

RAISE NOTICE '- v_active_entities (active entity report)';

RAISE NOTICE '- v_entity_changes (change tracking)';

RAISE NOTICE '- v_entity_paths (hierarchy paths)';

RAISE NOTICE '- v_tenant_resource_utilization (resource usage)';

RAISE NOTICE '===================================================================';

END;

$$
;
-- =====================================================================
-- PERSONS DOWN MIGRATION
-- =====================================================================
-- Drop triggers
DROP TRIGGER IF EXISTS update_persons_updated_at ON persons;

-- Drop RLS policies
DROP POLICY IF EXISTS persons_admin_access ON persons;

DROP POLICY IF EXISTS persons_tenant_isolation ON persons;

-- Disable RLS
ALTER TABLE
  IF EXISTS persons DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_persons_metadata_gin;

DROP INDEX IF EXISTS idx_persons_security_attributes_gin;

DROP INDEX IF EXISTS idx_persons_address_gin;

DROP INDEX IF EXISTS idx_persons_name;

DROP INDEX IF EXISTS idx_persons_email_lower;

DROP INDEX IF EXISTS idx_persons_deleted_at;

DROP INDEX IF EXISTS idx_persons_active;

DROP INDEX IF EXISTS idx_persons_entity;

DROP INDEX IF EXISTS idx_persons_type;

DROP INDEX IF EXISTS idx_persons_tenant;

-- Drop constraints
ALTER TABLE
  IF EXISTS persons DROP CONSTRAINT IF EXISTS persons_national_id_unique_active;

ALTER TABLE
  IF EXISTS persons DROP CONSTRAINT IF EXISTS persons_email_unique_active;

-- Drop table
DROP TABLE IF EXISTS persons;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'PERSONS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (persons)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '- All triggers and permissions';

RAISE NOTICE '===================================================================';

END;

$$
;
-- =====================================================================
-- EMPLOYEES DOWN MIGRATION
-- =====================================================================
-- Drop triggers
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;

-- Drop RLS policies
DROP POLICY IF EXISTS employees_admin_access ON employees;

DROP POLICY IF EXISTS employees_tenant_isolation ON employees;

-- Disable RLS
ALTER TABLE
  IF EXISTS employees DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS employees DROP CONSTRAINT IF EXISTS valid_security_level;

ALTER TABLE
  IF EXISTS employees DROP CONSTRAINT IF EXISTS valid_termination_date;

ALTER TABLE
  IF EXISTS employees DROP CONSTRAINT IF EXISTS employees_number_unique_active;

-- Drop indexes
DROP INDEX IF EXISTS idx_employees_access_attributes_gin;

DROP INDEX IF EXISTS idx_employees_work_schedule_gin;

DROP INDEX IF EXISTS idx_employees_salary_info_gin;

DROP INDEX IF EXISTS idx_employees_deleted_at;

DROP INDEX IF EXISTS idx_employees_security_level;

DROP INDEX IF EXISTS idx_employees_number;

DROP INDEX IF EXISTS idx_employees_active;

DROP INDEX IF EXISTS idx_employees_status;

DROP INDEX IF EXISTS idx_employees_manager;

DROP INDEX IF EXISTS idx_employees_department;

DROP INDEX IF EXISTS idx_employees_entity;

DROP INDEX IF EXISTS idx_employees_person;

DROP INDEX IF EXISTS idx_employees_tenant;

-- Drop table
DROP TABLE IF EXISTS employees;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'EMPLOYEES DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (employees)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '- All triggers and permissions';

RAISE NOTICE '===================================================================';

END;

$$
;
-- =====================================================================
-- USERS DOWN MIGRATION
-- =====================================================================
-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop RLS policies
DROP POLICY IF EXISTS users_admin_access ON users;

DROP POLICY IF EXISTS users_tenant_isolation ON users;

-- Disable RLS
ALTER TABLE
  IF EXISTS users DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS valid_lockout_time;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS valid_session_timeout;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS valid_failed_attempts;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS users_username_unique_active;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS users_email_unique_active;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_settings_gin;

DROP INDEX IF EXISTS idx_users_attributes_gin;

DROP INDEX IF EXISTS idx_users_deleted_at;

DROP INDEX IF EXISTS idx_users_mfa;

DROP INDEX IF EXISTS idx_users_lockout;

DROP INDEX IF EXISTS idx_users_failed_attempts;

DROP INDEX IF EXISTS idx_users_active;

DROP INDEX IF EXISTS idx_users_account_status;

DROP INDEX IF EXISTS idx_users_type;

DROP INDEX IF EXISTS idx_users_username_lower;

DROP INDEX IF EXISTS idx_users_email_lower;

DROP INDEX IF EXISTS idx_users_employee;

DROP INDEX IF EXISTS idx_users_person;

DROP INDEX IF EXISTS idx_users_entity;

DROP INDEX IF EXISTS idx_users_tenant;

-- Drop table
DROP TABLE IF EXISTS users;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'USERS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (users)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '- All triggers and permissions';

RAISE NOTICE '===================================================================';

END;

$$
;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS actions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
-- =====================================================================
-- POLICIES DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS policies_tenant_isolation ON policies;

-- Disable RLS
ALTER TABLE
  IF EXISTS policies DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS policies DROP CONSTRAINT IF EXISTS policies_name_unique_per_tenant;

-- Drop table
DROP TABLE IF EXISTS policies;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'POLICIES DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Table dropped: policies';

RAISE NOTICE 'RLS disabled and policy dropped.';

RAISE NOTICE '===================================================================';

END;

$$
;
-- User entity access is handled by the entity_id in the user_roles table.
-- =====================================================================
-- ACCESS REQUESTS DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS access_requests_tenant_isolation ON access_requests;

-- Disable RLS
ALTER TABLE
  IF EXISTS access_requests DISABLE ROW LEVEL SECURITY;

-- Drop table
DROP TABLE IF EXISTS access_requests;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ACCESS REQUESTS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Table dropped: access_requests';

RAISE NOTICE 'RLS disabled and policy dropped.';

RAISE NOTICE '===================================================================';

END;

$$
;
DROP TABLE IF EXISTS audit_log;
DROP VIEW IF EXISTS v_user_complete_view;

DROP VIEW IF EXISTS v_role_permissions_summary;

DROP VIEW IF EXISTS v_audit_summary_view;
-- =====================================================================
-- USER FUNCTIONS AND TRIGGERS DOWN MIGRATION
-- =====================================================================
-- Drop triggers in reverse order of creation
DROP TRIGGER IF EXISTS validate_role_hierarchy_trigger ON roles;

DROP TRIGGER IF EXISTS enforce_persons_tenant_isolation ON persons;

DROP TRIGGER IF EXISTS update_access_requests_updated_at ON access_requests;

DROP TRIGGER IF EXISTS update_policies_updated_at ON policies;

DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;

DROP TRIGGER IF EXISTS update_persons_updated_at ON persons;

-- Drop functions
DROP FUNCTION IF EXISTS validate_role_hierarchy() CASCADE;

DROP FUNCTION IF EXISTS enforce_tenant_isolation() CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

DROP FUNCTION IF EXISTS cleanup_expired_data(UUID) CASCADE;

DROP FUNCTION IF EXISTS user_has_permission(UUID, VARCHAR, VARCHAR, UUID, UUID, JSONB) CASCADE;

DROP FUNCTION IF EXISTS create_default_system_data(UUID) CASCADE;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'USER FUNCTIONS AND TRIGGERS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Functions and triggers dropped.';

RAISE NOTICE '===================================================================';

END;

$$
;
DROP FUNCTION IF EXISTS validate_and_set_tenant_context(UUID);

-- -- =====================================================================
-- -- PROJECTS DOWN MIGRATION
-- -- =====================================================================
--
-- DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
-- DROP POLICY IF EXISTS admin_full_access_policy ON projects;
-- DROP POLICY IF EXISTS tenant_isolation_policy ON projects;
-- ALTER TABLE IF EXISTS projects DISABLE ROW LEVEL SECURITY;
-- DROP TABLE IF EXISTS projects;
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
-- =====================================================================
-- ACCOUNT DOWN MIGRATION
-- =====================================================================
DROP POLICY IF EXISTS admin_full_access_policy ON account;

DROP POLICY IF EXISTS tenant_isolation_policy ON account;

ALTER TABLE
  IF EXISTS account DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS account;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS user_permissions;

DROP POLICY IF EXISTS user_permissions_tenant_isolation ON user_permissions;

DROP POLICY IF EXISTS user_permissions_admin_access ON user_permissions;
-- 7. Revert Security Notification System
DROP TRIGGER IF EXISTS notify_role_changes ON user_roles;

DROP TRIGGER IF EXISTS notify_suspicious_login ON user_sessions;

DROP FUNCTION IF EXISTS trigger_security_notification;

DROP TABLE IF EXISTS security_notifications;

-- 6. Revert Index Optimizations
DROP INDEX IF EXISTS idx_active_roles;

DROP INDEX IF EXISTS idx_active_users;

DROP INDEX IF EXISTS idx_policies_rule_gin;

-- DROP INDEX IF EXISTS idx_users_attributes_gin;
DROP INDEX IF EXISTS idx_user_sessions_time_brin;

DROP INDEX IF EXISTS idx_audit_log_time_brin;

-- 5. Revert Advanced Threat Detection View
DROP VIEW IF EXISTS v_security_threat_dashboard;

-- 4. Revert Compliance Enhancements
DROP FUNCTION IF EXISTS enforce_data_retention;

DROP FUNCTION IF EXISTS gdpr_user_deletion;

-- 3. Revert Security Automation Functions
DROP FUNCTION IF EXISTS terminate_risky_sessions;

DROP FUNCTION IF EXISTS assess_session_risk;

-- 2. Revert Performance Optimizations
DROP MATERIALIZED VIEW IF EXISTS mv_user_effective_permissions;

-- 1. Revert Security Hardening Enhancements
ALTER TABLE
  user_sessions DROP COLUMN IF EXISTS anomaly_flags,
  DROP COLUMN IF EXISTS mfa_verified_at;
-- ------------------------------------------------------------------------------------------------
-- DOWN MIGRATION: ATTRIBUTE DEFINITIONS
--
-- Reverts the attribute_definitions table and associated policies.
-- ------------------------------------------------------------------------------------------------
-- Drop RLS policy
DROP POLICY IF EXISTS attribute_definitions_tenant_isolation ON attribute_definitions;

-- Disable RLS
ALTER TABLE
  attribute_definitions DISABLE ROW LEVEL SECURITY;

-- Drop comments from columns
COMMENT ON COLUMN attribute_definitions.encryption_required IS NULL;

COMMENT ON COLUMN attribute_definitions.validation_rules IS NULL;

COMMENT ON COLUMN attribute_definitions.allowed_values IS NULL;

COMMENT ON COLUMN attribute_definitions.is_sensitive IS NULL;

COMMENT ON COLUMN attribute_definitions.category IS NULL;

COMMENT ON COLUMN attribute_definitions.data_type IS NULL;

-- Drop comment from table
COMMENT ON TABLE attribute_definitions IS NULL;

-- Drop the attribute_definitions table
DROP TABLE IF EXISTS attribute_definitions;
-- Drop attribute_sources table
DROP TABLE IF EXISTS attribute_sources CASCADE;

-- Drop attribute_values table
DROP TABLE IF EXISTS attribute_values CASCADE;
-- ------------------------------------------------------------------------------------------------
-- DOWN MIGRATION: POLICY EVALUATIONS CACHE
--
-- Reverts the policy_evaluations table and associated policies.
-- ------------------------------------------------------------------------------------------------
-- Drop RLS policy
DROP POLICY IF EXISTS policy_evaluations_tenant_isolation ON policy_evaluations;

-- Disable RLS
ALTER TABLE
  policy_evaluations DISABLE ROW LEVEL SECURITY;

-- Drop comments from columns
COMMENT ON COLUMN policy_evaluations.evaluation_time_ms IS NULL;

COMMENT ON COLUMN policy_evaluations.applicable_policies IS NULL;

COMMENT ON COLUMN policy_evaluations.context_hash IS NULL;

-- Drop comment from table
COMMENT ON TABLE policy_evaluations IS NULL;

-- Drop the policy_evaluations table
DROP TABLE IF EXISTS policy_evaluations;
-- Revert to the original policy_evaluations table structure
DROP TABLE IF EXISTS policy_evaluations CASCADE;

-- Recreate original table (this should match the original 000047 migration)
CREATE TABLE IF NOT EXISTS policy_evaluations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  resource_id UUID NOT NULL REFERENCES resources(id),
  action_id UUID NOT NULL REFERENCES actions(id),
  context_hash VARCHAR(64) NOT NULL,
  decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
  applicable_policies UUID [] DEFAULT '{}',
  evaluation_time_ms INTEGER,
  evaluated_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour')
);

-- Enable RLS and create policies
ALTER TABLE
  policy_evaluations ENABLE ROW LEVEL SECURITY;

CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);
DROP TABLE IF EXISTS policy_decisions;
DROP FUNCTION IF EXISTS assign_user_role(UUID, UUID, UUID, UUID);

DROP FUNCTION IF EXISTS revoke_user_role(UUID, UUID, UUID);
-- =====================================================
-- ROLLBACK USER ACTIVITIES TABLE MIGRATION
-- =====================================================
-- Drops user activities table and related objects
-- Drop utility functions
DROP FUNCTION IF EXISTS create_monthly_user_activities_partition(DATE);

DROP FUNCTION IF EXISTS drop_old_user_activities_partitions(INTEGER);

-- Drop all user activities partitions
-- DROP TABLE IF EXISTS user_activities_prev2;
-- DROP TABLE IF EXISTS user_activities_prev1;
-- DROP TABLE IF EXISTS user_activities_current;
-- DROP TABLE IF EXISTS user_activities_next1;
-- DROP TABLE IF EXISTS user_activities_next2;
--
-- Drop the main partitioned table (this will cascade to any remaining partitions)
DROP TABLE IF EXISTS user_activities;
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
-- DROP TABLE
-- =====================================================
DROP TABLE IF EXISTS tenant_feature_overrides;
-- Rollback audit functions and triggers for feature flag changes
-- =====================================================
-- DROP TRIGGERS
-- =====================================================
DROP TRIGGER IF EXISTS tenant_feature_overrides_audit_trigger ON tenant_feature_overrides;

DROP TRIGGER IF EXISTS feature_flags_audit_trigger ON feature_flags;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS audit_tenant_feature_override_changes();

DROP FUNCTION IF EXISTS audit_feature_flag_changes();
-- Rollback the feature flag evaluation functions
-- This migration file is mostly commented out functions,
-- so there's nothing to rollback in the current implementation.
-- If functions were actually created, they would be dropped here.
-- Example of what would be here if functions were created:
-- DROP FUNCTION IF EXISTS evaluate_feature_flag_fast(VARCHAR);
-- DROP FUNCTION IF EXISTS evaluate_all_feature_flags();
-- DROP FUNCTION IF EXISTS evaluate_feature_flag(VARCHAR);
-- DROP FUNCTION IF EXISTS set_audit_context(UUID, UUID);
-- Rollback the feature flag cache materialized view and functions
-- =====================================================
-- DROP TRIGGERS
-- =====================================================
DROP TRIGGER IF EXISTS tenant_overrides_cache_refresh_trigger ON tenant_feature_overrides;

DROP TRIGGER IF EXISTS feature_flags_cache_refresh_trigger ON feature_flags;

-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS trigger_cache_refresh();

DROP FUNCTION IF EXISTS evaluate_all_feature_flags_cached();

DROP FUNCTION IF EXISTS evaluate_feature_flag_cached(VARCHAR);

DROP FUNCTION IF EXISTS check_cache_freshness(UUID);

DROP FUNCTION IF EXISTS get_feature_flags_cache_stats();

DROP FUNCTION IF EXISTS refresh_feature_flags_cache();

-- =====================================================
-- DROP INDEXES ON MATERIALIZED VIEW
-- =====================================================
DROP INDEX IF EXISTS idx_tenant_feature_cache_value;

DROP INDEX IF EXISTS idx_tenant_feature_cache_metadata;

DROP INDEX IF EXISTS idx_tenant_feature_cache_target_audience;

DROP INDEX IF EXISTS idx_tenant_feature_cache_rollout;

DROP INDEX IF EXISTS idx_tenant_feature_cache_timestamp;

DROP INDEX IF EXISTS idx_tenant_feature_cache_flag_type;

DROP INDEX IF EXISTS idx_tenant_feature_cache_source;

DROP INDEX IF EXISTS idx_tenant_feature_cache_enabled;

DROP INDEX IF EXISTS idx_tenant_feature_cache_tenant;

DROP INDEX IF EXISTS idx_tenant_feature_cache_name_lookup;

DROP INDEX IF EXISTS idx_tenant_feature_cache_pk;

-- =====================================================
-- DROP MATERIALIZED VIEW
-- =====================================================
DROP MATERIALIZED VIEW IF EXISTS mv_tenant_feature_flags_cache;
-- Rollback maintenance and cleanup functions
-- =====================================================
-- DROP FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS export_tenant_feature_flags(UUID);

DROP FUNCTION IF EXISTS bulk_create_feature_flags(JSONB, UUID);

DROP FUNCTION IF EXISTS bulk_update_rollout_percentage(TEXT [], INTEGER, UUID);

DROP FUNCTION IF EXISTS fix_orphaned_overrides();

DROP FUNCTION IF EXISTS check_feature_flags_integrity();

DROP FUNCTION IF EXISTS get_tenant_feature_flag_stats(UUID);

DROP FUNCTION IF EXISTS get_feature_flags_health_metrics();

DROP FUNCTION IF EXISTS cleanup_soft_deleted_feature_flags(INTEGER);

DROP FUNCTION IF EXISTS cleanup_old_feature_flag_audit_logs(INTEGER);
-- Rollback usage examples and documentation
-- This migration contains only documentation and examples,
-- so there's nothing to rollback.
-- =====================================================================
-- REVERT FINANCE ACCOUNT GROUPS AND HEADERS
-- Reverses the changes made by the 'up' migration script
-- =====================================================================
-- Revoke permissions
REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_groups
FROM
  application_role;

REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_groups
FROM
  admin_role;

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_groups;

DROP POLICY IF EXISTS admin_full_access_policy ON finance_account_groups;

-- Disable RLS
ALTER TABLE
  finance_account_groups DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_account_groups_tenant_code;

DROP INDEX IF EXISTS idx_account_groups_tenant_type;

DROP INDEX IF EXISTS idx_account_groups_parent;

DROP INDEX IF EXISTS idx_account_groups_hierarchy_level;

DROP INDEX IF EXISTS idx_account_groups_statement;

-- Drop table
DROP TABLE IF EXISTS finance_account_groups;
-- Drop chart of accounts table
DROP TABLE IF EXISTS finance_accounts CASCADE;
-- Drop chart of accounts indexes
DROP INDEX IF EXISTS idx_chart_of_accounts_attributes_gin;

DROP INDEX IF EXISTS idx_chart_of_accounts_active;

DROP INDEX IF EXISTS idx_chart_of_accounts_entity;

DROP INDEX IF EXISTS idx_chart_of_accounts_currency;

DROP INDEX IF EXISTS idx_chart_of_accounts_control;

DROP INDEX IF EXISTS idx_chart_of_accounts_balance_tracking;

DROP INDEX IF EXISTS idx_chart_of_accounts_statement_line;

DROP INDEX IF EXISTS idx_chart_of_accounts_parent;

DROP INDEX IF EXISTS idx_chart_of_accounts_tenant_type;

DROP INDEX IF EXISTS idx_chart_of_accounts_tenant_code;
-- Drop finance transactions table
DROP TABLE IF EXISTS finance_transactions CASCADE;
-- Drop finance transaction entries table
DROP TABLE IF EXISTS finance_transaction_entries CASCADE;
-- +migrate Down
BEGIN
;

-- Revoke permissions
REVOKE
SELECT
  ON v_financial_statement_builder
FROM
  application_role,
  admin_role;

REVOKE
SELECT
  ON v_chart_of_accounts_complete
FROM
  application_role,
  admin_role;

REVOKE
SELECT
  ON v_finance_accounts_with_groups
FROM
  application_role,
  admin_role;

-- Drop views
DROP VIEW IF EXISTS v_financial_statement_builder;

DROP VIEW IF EXISTS v_chart_of_accounts_complete;

DROP VIEW IF EXISTS v_financial_statement_structure;

DROP VIEW IF EXISTS v_finance_accounts_with_groups;

-- Drop trigger and function for group hierarchy
DROP TRIGGER IF EXISTS trigger_maintain_group_hierarchy ON finance_account_groups;

DROP FUNCTION IF EXISTS maintain_group_hierarchy_path();

-- Drop other functions
DROP FUNCTION IF EXISTS create_standard_account_groups(UUID);

DROP FUNCTION IF EXISTS update_group_balances(UUID);

COMMIT;
-- +migrate Down
BEGIN
;

DROP TABLE IF EXISTS finance_account_validation_rules;

DROP TABLE IF EXISTS finance_account_balances;

ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS has_children;

ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS is_leaf_account;

COMMIT;

-- +migrate Down
BEGIN
;

DROP TRIGGER IF EXISTS trigger_update_account_balances ON finance_transactions;

DROP FUNCTION IF EXISTS update_account_balances_after_posting();

DROP TRIGGER IF EXISTS trigger_maintain_account_hierarchy ON finance_accounts;

DROP FUNCTION IF EXISTS maintain_account_hierarchy_path();

DROP TRIGGER IF EXISTS trigger_update_hierarchy_flags ON finance_accounts;

DROP FUNCTION IF EXISTS update_account_hierarchy_flags();

DROP FUNCTION IF EXISTS validate_account_hierarchy(UUID);

DROP FUNCTION IF EXISTS recalculate_account_balance(UUID);

DROP TRIGGER IF EXISTS trigger_validate_transaction_posting ON finance_transactions;

DROP FUNCTION IF EXISTS validate_transaction_before_posting();

DROP FUNCTION IF EXISTS get_account_full_name(UUID);

DROP FUNCTION IF EXISTS get_account_children(UUID, BOOLEAN);

DROP FUNCTION IF EXISTS analyze_finance_tables_performance();

COMMIT;
-- +migrate Down
BEGIN
;

DROP VIEW IF EXISTS v_finance_accounts_hierarchy;

DROP VIEW IF EXISTS v_finance_transaction_summary;

DROP VIEW IF EXISTS v_finance_account_activity;

COMMIT;
-- +migrate Down
BEGIN
;

-- Revoke permissions on views
REVOKE
SELECT
  ON v_finance_account_activity
FROM
  application_role,
  admin_role;

REVOKE
SELECT
  ON v_finance_transaction_summary
FROM
  application_role,
  admin_role;

REVOKE
SELECT
  ON v_finance_accounts_hierarchy
FROM
  application_role,
  admin_role;

-- Revoke permissions on new tables
REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_validation_rules
FROM
  application_role,
  admin_role;

REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_balances
FROM
  application_role,
  admin_role;

-- Drop admin policies for new tables
DROP POLICY IF EXISTS admin_full_access_policy ON finance_account_validation_rules;

DROP POLICY IF EXISTS admin_full_access_policy ON finance_account_balances;

-- Drop RLS policies for new tables
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_validation_rules;

DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_balances;

-- Disable RLS on new tables
ALTER TABLE
  finance_account_validation_rules DISABLE ROW LEVEL SECURITY;

ALTER TABLE
  finance_account_balances DISABLE ROW LEVEL SECURITY;

COMMIT;
