-- =====================================================
-- DOWN MIGRATION FOR ERP/ACCOUNTING SYSTEM DATABASE
-- =====================================================
-- File: 001_down_erp_accounting_system.sql
-- Description: Reverts the changes made by 001_create_erp_accounting_system.sql
-- Author: Generated Migration
-- Date: 2025-07-04
-- Version: 1.0.0
-- =====================================================

-- Start transaction for atomic rollback

-- =====================================================
-- STEP 1: DROP DEPENDENT FOREIGN KEY CONSTRAINTS
-- =====================================================
-- Drop foreign key constraints on 'tenants' from other modules.
-- Ensure all tables that reference 'tenants.id' have their FKs dropped first.

-- Core ERP module constraints
ALTER TABLE IF EXISTS persons DROP CONSTRAINT IF EXISTS persons_tenant_id_fkey;
ALTER TABLE IF EXISTS employees DROP CONSTRAINT IF EXISTS employees_tenant_id_fkey;
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_tenant_id_fkey;
ALTER TABLE IF EXISTS roles DROP CONSTRAINT IF EXISTS roles_tenant_id_fkey;
ALTER TABLE IF EXISTS user_roles DROP CONSTRAINT IF EXISTS user_roles_tenant_id_fkey;

-- Inventory module constraints
ALTER TABLE IF EXISTS warehouses DROP CONSTRAINT IF EXISTS warehouses_tenant_id_fkey;
ALTER TABLE IF EXISTS item_categories DROP CONSTRAINT IF EXISTS item_categories_tenant_id_fkey;
ALTER TABLE IF EXISTS items DROP CONSTRAINT IF EXISTS items_tenant_id_fkey;
ALTER TABLE IF EXISTS inventory_balances DROP CONSTRAINT IF EXISTS inventory_balances_tenant_id_fkey;
ALTER TABLE IF EXISTS inventory_movements DROP CONSTRAINT IF EXISTS inventory_movements_tenant_id_fkey;

-- Entity and hierarchy constraints
ALTER TABLE IF EXISTS entities DROP CONSTRAINT IF EXISTS entities_tenant_id_fkey;
ALTER TABLE IF EXISTS hierarchy_paths DROP CONSTRAINT IF EXISTS hierarchy_paths_tenant_id_fkey;

-- UOM and accounting constraints
ALTER TABLE IF EXISTS uom DROP CONSTRAINT IF EXISTS uom_tenant_id_fkey;
ALTER TABLE IF EXISTS uom_conversion DROP CONSTRAINT IF EXISTS uom_conversion_tenant_id_fkey;
ALTER TABLE IF EXISTS chartofaccount DROP CONSTRAINT IF EXISTS chartofaccount_tenant_id_fkey;
ALTER TABLE IF EXISTS account DROP CONSTRAINT IF EXISTS account_tenant_id_fkey;

-- Project and audit constraints
ALTER TABLE IF EXISTS projects DROP CONSTRAINT IF EXISTS projects_tenant_id_fkey;
ALTER TABLE IF EXISTS budgets DROP CONSTRAINT IF EXISTS budgets_tenant_id_fkey;
ALTER TABLE IF EXISTS audit_logs DROP CONSTRAINT IF EXISTS audit_logs_tenant_id_fkey;

-- Journal entry constraints
ALTER TABLE IF EXISTS journal_entries DROP CONSTRAINT IF EXISTS journal_entries_tenant_id_fkey;
ALTER TABLE IF EXISTS journal_entry_lines DROP CONSTRAINT IF EXISTS journal_entry_lines_tenant_id_fkey;

-- Vendor and customer constraints
ALTER TABLE IF EXISTS vendors DROP CONSTRAINT IF EXISTS vendors_tenant_id_fkey;
ALTER TABLE IF EXISTS customers DROP CONSTRAINT IF EXISTS customers_tenant_id_fkey;

-- =====================================================
-- STEP 2: REVERT RLS POLICIES FROM OTHER TABLES
-- =====================================================
-- Drop RLS policies that reference functions we're about to drop

-- Entity management policies
DO $$
BEGIN
    DROP POLICY IF EXISTS tenant_isolation_policy ON entities;
    DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;
EXCEPTION
    WHEN undefined_table THEN
        -- Table doesn't exist, skip
        NULL;
END $$;

-- Person and employee policies
DO $$
BEGIN
    DROP POLICY IF EXISTS tenant_isolation_policy ON persons;
    DROP POLICY IF EXISTS tenant_isolation_policy ON employees;
EXCEPTION
    WHEN undefined_table THEN
        -- Table doesn't exist, skip
        NULL;
END $$;

-- User management policies
DO $$
BEGIN
    DROP POLICY IF EXISTS tenant_isolation_policy ON users;
    DROP POLICY IF EXISTS tenant_isolation_policy ON roles;
    DROP POLICY IF EXISTS user_roles_policy ON user_roles;
EXCEPTION
    WHEN undefined_table THEN
        -- Table doesn't exist, skip
        NULL;
END $$;

-- Project management policies
DO $$
BEGIN
    DROP POLICY IF EXISTS tenant_isolation_policy ON projects;
    DROP POLICY IF EXISTS tenant_isolation_policy ON budgets;
EXCEPTION
    WHEN undefined_table THEN
        -- Table doesn't exist, skip
        NULL;
END $$;

-- Audit and accounting policies
DO $$
BEGIN
    DROP POLICY IF EXISTS tenant_isolation_policy ON audit_logs;
    DROP POLICY IF EXISTS tenant_isolation_policy ON account;
    DROP POLICY IF EXISTS tenant_isolation_policy ON chartofaccount;
EXCEPTION
    WHEN undefined_table THEN
        -- Table doesn't exist, skip
        NULL;
END $$;

-- =====================================================
-- STEP 3: REVERT TRIGGERS FROM OTHER TABLES
-- =====================================================
-- Drop triggers that depend on functions we're about to drop

DROP TRIGGER IF EXISTS update_entity_timestamps ON entities;
DROP TRIGGER IF EXISTS update_person_timestamps ON persons;
DROP TRIGGER IF EXISTS update_employee_timestamps ON employees;
DROP TRIGGER IF EXISTS update_user_timestamps ON users;
DROP TRIGGER IF EXISTS update_project_timestamps ON projects;
DROP TRIGGER IF EXISTS update_budget_timestamps ON budgets;
DROP TRIGGER IF EXISTS trg_maintain_hierarchy_paths ON entities;

-- =====================================================
-- STEP 4: REVERT INITIAL DATA SETUP
-- =====================================================

-- Drop the trigger that creates default tenant configuration
DROP TRIGGER IF EXISTS create_tenant_configuration_trigger ON tenants;

-- Drop the function that creates default tenant configuration
DROP FUNCTION IF EXISTS create_default_tenant_configuration(UUID);
DROP FUNCTION IF EXISTS create_tenant_configuration_on_insert();

-- =====================================================
-- STEP 5: REVERT TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================

-- Drop triggers for updated_at column
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_tenant_configurations_updated_at ON tenant_configurations;

-- Drop the generic function to update updated_at timestamp
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS update_timestamps();

-- =====================================================
-- STEP 6: REVERT RLS FUNCTIONS AND HIERARCHY FUNCTIONS
-- =====================================================

-- Drop hierarchy management functions
DROP FUNCTION IF EXISTS get_entity_descendants(UUID, INT);
DROP FUNCTION IF EXISTS get_entity_ancestors(UUID, INT);
DROP FUNCTION IF EXISTS is_entity_descendant(UUID, UUID);
DROP FUNCTION IF EXISTS get_entity_path(UUID, VARCHAR);
DROP FUNCTION IF EXISTS get_entity_children(UUID);
DROP FUNCTION IF EXISTS rebuild_hierarchy_paths(UUID);
DROP FUNCTION IF EXISTS maintain_hierarchy_paths();

-- Drop validation functions
DROP FUNCTION IF EXISTS validate_hierarchy_integrity(UUID);

-- Drop utility functions
DROP FUNCTION IF EXISTS get_tenant_statistics(UUID);
DROP FUNCTION IF EXISTS is_valid_uuid(TEXT);
DROP FUNCTION IF EXISTS ensure_valid_uuid(UUID);

-- Drop RLS management functions
DROP FUNCTION IF EXISTS apply_rls_to_tenant_tables(TEXT, TEXT, TEXT, BOOLEAN);
DROP FUNCTION IF EXISTS remove_rls_from_tenant_tables_by_type(TEXT, TEXT, BOOLEAN);
DROP FUNCTION IF EXISTS manage_rls_across_schemas(TEXT[], TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN);

-- Drop tenant context functions
DROP FUNCTION IF EXISTS set_tenant_context_uuid(UUID);
DROP FUNCTION IF EXISTS clear_tenant_context();
DROP FUNCTION IF EXISTS current_tenant_id();

-- =====================================================
-- STEP 7: REVERT PERMISSIONS AND GRANTS
-- =====================================================

-- Revoke permissions from application_role
REVOKE ALL ON tenants FROM application_role;
REVOKE ALL ON tenant_configurations FROM application_role;
REVOKE ALL ON tenant_usage_stats FROM application_role;

-- Revoke execute permissions on functions (if they still exist)
DO $$
BEGIN
    REVOKE EXECUTE ON FUNCTION set_tenant_context(UUID) FROM application_role;
EXCEPTION
    WHEN undefined_function THEN
        -- Function doesn't exist, skip
        NULL;
END $$;

DO $$
BEGIN
    REVOKE EXECUTE ON FUNCTION get_current_tenant_id() FROM application_role;
EXCEPTION
    WHEN undefined_function THEN
        -- Function doesn't exist, skip
        NULL;
END $$;

-- =====================================================
-- STEP 8: REVERT ROW LEVEL SECURITY (RLS)
-- =====================================================

-- Drop RLS policies on tenant tables
DROP POLICY IF EXISTS tenant_isolation_policy ON tenants;
DROP POLICY IF EXISTS tenant_configurations_isolation_policy ON tenant_configurations;
DROP POLICY IF EXISTS tenant_usage_stats_isolation_policy ON tenant_usage_stats;

-- Disable RLS on tables
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_configurations DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_usage_stats DISABLE ROW LEVEL SECURITY;

-- =====================================================
-- STEP 9: REVERT UTILITY FUNCTIONS
-- =====================================================

-- Drop tenant limits checking function
DROP FUNCTION IF EXISTS check_tenant_limits(UUID, VARCHAR, INT);

-- Drop get current tenant ID function
DROP FUNCTION IF EXISTS get_current_tenant_id();

-- Drop set tenant context function
DROP FUNCTION IF EXISTS set_tenant_context(UUID);

-- =====================================================
-- STEP 10: DROP VIEWS
-- =====================================================

DROP VIEW IF EXISTS hierarchy_uuid_examples;

-- =====================================================
-- STEP 11: REVERT CORE TENANT MANAGEMENT TABLES
-- =====================================================

-- Drop tenant_usage_stats table and its indexes
DROP INDEX IF EXISTS idx_tenant_usage_stats_tenant_period;
DROP INDEX IF EXISTS idx_tenant_usage_stats_period;
DROP TABLE IF EXISTS tenant_usage_stats;

-- Drop tenant_configurations table
DROP TABLE IF EXISTS tenant_configurations;

-- Drop tenants table and its indexes
DROP INDEX IF EXISTS idx_tenants_deleted_at;
DROP INDEX IF EXISTS idx_tenants_subdomain;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_slug;
DROP TABLE IF EXISTS tenants;

-- =====================================================
-- STEP 12: REVERT ROLES AND PERMISSIONS
-- =====================================================

-- Drop application role
-- Note: This will fail if the role is still owned by objects or has privileges granted outside this script.
DROP ROLE IF EXISTS application_role;

-- =====================================================
-- STEP 13: EXTENSIONS (CAUTIOUS APPROACH)
-- =====================================================

-- DO NOT DROP uuid-ossp extension in down migration
-- This is because:
-- 1. Other tables/migrations may depend on it
-- 2. Extensions are typically shared across the entire database
-- 3. Dropping extensions can cause cascading failures

-- If you absolutely need to drop the extension, you would need to:
-- 1. First identify all dependent objects: SELECT * FROM pg_depend WHERE refobjid = 'uuid-ossp'::regtype;
-- 2. Drop or modify all dependent objects first
-- 3. Then drop the extension

-- For safety, we leave the extension in place
-- DROP EXTENSION IF EXISTS "uuid-ossp"; -- COMMENTED OUT FOR SAFETY

-- RAISE NOTICE 'UUID-OSSP extension left in place to avoid dependency conflicts';
-- RAISE NOTICE 'Other migrations may depend on this extension';
--
-- -- =====================================================
-- -- MIGRATION ROLLBACK COMPLETION
-- -- =====================================================
--
-- -- Commit the transaction
-- COMMIT;
--
-- -- Log completion
-- DO $$
-- BEGIN
--     RAISE NOTICE '====================================================';
--     RAISE NOTICE 'ERP/Accounting System Database Down Migration Completed Successfully!';
--     RAISE NOTICE '====================================================';
--     RAISE NOTICE 'Reverted tables: tenants, tenant_configurations, tenant_usage_stats';
--     RAISE NOTICE 'Reverted functions, RLS policies, and triggers';
--     RAISE NOTICE 'Reverted hierarchy management and utility functions';
--     RAISE NOTICE 'Cleaned up dependencies from other modules';
--     RAISE NOTICE '';
--     RAISE NOTICE 'NOTE: uuid-ossp extension was NOT dropped to avoid conflicts';
--     RAISE NOTICE 'Other parts of your system may depend on this extension';
--     RAISE NOTICE '====================================================';
-- END;
-- $$;
