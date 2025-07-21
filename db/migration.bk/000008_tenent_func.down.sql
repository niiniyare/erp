-- =====================================================
-- ERP/ACCOUNTING SYSTEM - RLS AND HIERARCHY MANAGEMENT (UUID) - DOWN MIGRATION
-- =====================================================
-- Description: Reverts the changes made by the up migration for RLS and hierarchy management with UUID support
-- Author: Abdirahman Ahmed
-- Date: 2025-07-04
-- Version: 1.0.0
-- Dependencies: None
-- =====================================================

-- Start transaction to ensure atomic migration
BEGIN;

-- =====================================================
-- DROP EXAMPLE QUERIES VIEW
-- =====================================================
DROP VIEW IF EXISTS hierarchy_uuid_examples;

-- =====================================================
-- DROP VALIDATION FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS validate_hierarchy_integrity(UUID);
DROP FUNCTION IF EXISTS ensure_valid_uuid(UUID);
DROP FUNCTION IF EXISTS is_valid_uuid(TEXT);

-- =====================================================
-- DROP UTILITY FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS get_tenant_statistics(UUID);
DROP FUNCTION IF EXISTS manage_rls_across_schemas(TEXT[], TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN);
DROP FUNCTION IF EXISTS remove_rls_from_tenant_tables_by_type(TEXT, TEXT, BOOLEAN);
DROP FUNCTION IF EXISTS apply_rls_to_tenant_tables(TEXT, TEXT, TEXT, BOOLEAN);

-- =====================================================
-- DROP HIERARCHY QUERY FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS rebuild_hierarchy_paths(UUID);
DROP FUNCTION IF EXISTS get_entity_children(UUID);
DROP FUNCTION IF EXISTS get_entity_path(UUID, VARCHAR);
DROP FUNCTION IF EXISTS is_entity_descendant(UUID, UUID);
DROP FUNCTION IF EXISTS get_entity_ancestors(UUID, INT);
DROP FUNCTION IF EXISTS get_entity_descendants(UUID, INT);

-- =====================================================
-- DROP HIERARCHY MANAGEMENT TRIGGER
-- =====================================================
DROP TRIGGER IF EXISTS trg_maintain_hierarchy_paths ON entities;

-- =====================================================
-- DROP HIERARCHY PATH MAINTENANCE FUNCTION
-- =====================================================
DROP FUNCTION IF EXISTS maintain_hierarchy_paths();

-- =====================================================
-- DROP TIMESTAMP TRIGGERS
-- =====================================================
DROP TRIGGER IF EXISTS update_budget_timestamps ON budgets;
DROP TRIGGER IF EXISTS update_project_timestamps ON projects;
DROP TRIGGER IF EXISTS update_user_timestamps ON users;
DROP TRIGGER IF EXISTS update_employee_timestamps ON employees;
DROP TRIGGER IF EXISTS update_person_timestamps ON persons;
DROP TRIGGER IF EXISTS update_entity_timestamps ON entities;
DROP TRIGGER IF EXISTS update_tenant_timestamps ON tenants;

-- =====================================================
-- DROP TIMESTAMP UPDATE FUNCTION
-- =====================================================
DROP FUNCTION IF EXISTS update_timestamps();

-- =====================================================
-- DROP ROW LEVEL SECURITY POLICIES
-- =====================================================
DROP POLICY IF EXISTS user_roles_policy ON user_roles;
DROP POLICY IF EXISTS tenant_isolation_policy ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation_policy ON account;
DROP POLICY IF EXISTS tenant_isolation_policy ON chartofaccount;
DROP POLICY IF EXISTS tenant_isolation_policy ON budgets;
DROP POLICY IF EXISTS tenant_isolation_policy ON projects;
DROP POLICY IF EXISTS tenant_isolation_policy ON roles;
DROP POLICY IF EXISTS tenant_isolation_policy ON users;
DROP POLICY IF EXISTS tenant_isolation_policy ON employees;
DROP POLICY IF EXISTS tenant_isolation_policy ON persons;
DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;
DROP POLICY IF EXISTS tenant_isolation_policy ON entities;
-- DROP POLICY IF EXISTS tenant_isolation_policy ON tenants; -- Assuming this policy might be managed elsewhere

-- =====================================================
-- DROP TENANT CONTEXT FUNCTIONS
-- =====================================================
DROP FUNCTION IF EXISTS clear_tenant_context();
DROP FUNCTION IF EXISTS set_tenant_context_uuid(UUID);
-- Commit the transaction
COMMIT;

-- Log completion
DO $$
BEGIN
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'ERP RLS and Hierarchy Management Migration (UUID) - DOWN MIGRATION Completed!';
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'Features Removed:';
    RAISE NOTICE '  - Enhanced tenant context management functions';
    RAISE NOTICE '  - Row Level Security policies for tenant tables';
    RAISE NOTICE '  - Automatic timestamp triggers';
    RAISE NOTICE '  - Hierarchy management functions and trigger';
    RAISE NOTICE '  - UUID validation and utility functions';
    RAISE NOTICE '  - Example queries view';
    RAISE NOTICE '';
    RAISE NOTICE 'Note: This down migration assumes that the corresponding tables (e.g., entities, hierarchy_paths) are managed by other migrations.';
    RAISE NOTICE 'It only reverts the RLS, functions, triggers, and policies created in the up migration.';
    RAISE NOTICE '====================================================';
END;
$$;
