-- =====================================================================
-- ENTITIES MODULE DOWN MIGRATION
-- =====================================================================
-- 
-- This down migration properly cleans up all objects created in the
-- entities up migration in reverse order of creation
-- 
-- Order of operations:
-- 1. Drop view RLS policies
-- 2. Drop views
-- 3. Drop table triggers and functions
-- 4. Drop table RLS policies
-- 5. Disable RLS on tables
-- 6. Drop indexes
-- 7. Drop tables in dependency order
-- =====================================================================

-- =====================================================================
-- VIEWS CLEANUP
-- =====================================================================
-- Note: Views don't have direct RLS policies, they inherit RLS from underlying tables
-- DROP VIEWS
-- =====================================================================

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
-- DROP TRIGGERS AND FUNCTIONS
-- =====================================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS hierarchy_paths_maintain_entity_id ON hierarchy_paths;
DROP TRIGGER IF EXISTS entities_maintain_entity_id ON entities;

-- Drop trigger functions
DROP FUNCTION IF EXISTS maintain_entity_id();

-- =====================================================================
-- DROP TABLE RLS POLICIES
-- =====================================================================

-- Drop admin bypass policies
DROP POLICY IF EXISTS admin_full_access_policy ON entitystate;
DROP POLICY IF EXISTS admin_full_access_policy ON hierarchy_paths;
DROP POLICY IF EXISTS admin_full_access_policy ON entities;

-- Drop tenant isolation policies
DROP POLICY IF EXISTS tenant_isolation_policy ON entitystate;
DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;
DROP POLICY IF EXISTS tenant_isolation_policy ON entities;

-- =====================================================================
-- DISABLE RLS ON TABLES
-- =====================================================================

ALTER TABLE IF EXISTS entitystate DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS hierarchy_paths DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS entities DISABLE ROW LEVEL SECURITY;

-- =====================================================================
-- DROP INDEXES
-- =====================================================================

-- Drop performance indexes
DROP INDEX IF EXISTS idx_entities_settings_gin;
DROP INDEX IF EXISTS idx_entities_address_gin;
DROP INDEX IF EXISTS idx_entities_deleted_at;
DROP INDEX IF EXISTS idx_hierarchy_paths_ancestor;
DROP INDEX IF EXISTS idx_hierarchy_paths_tenant;
DROP INDEX IF EXISTS idx_entities_type;
DROP INDEX IF EXISTS idx_entities_parent;
DROP INDEX IF EXISTS idx_entities_tenant;
DROP INDEX IF EXISTS idx_entitystate_fiscal_year;
DROP INDEX IF EXISTS idx_entitystate_entity_key;
DROP INDEX IF EXISTS idx_hierarchy_paths_depth;
DROP INDEX IF EXISTS idx_hierarchy_paths_descendant;
DROP INDEX IF EXISTS idx_entities_active;
DROP INDEX IF EXISTS idx_entities_parent_id;
DROP INDEX IF EXISTS idx_entities_tenant_type;
DROP INDEX IF EXISTS tenant_code_unique_idx;

-- =====================================================================
-- DROP FOREIGN KEY CONSTRAINTS FROM DEPENDENT TABLES
-- =====================================================================

-- Drop foreign key constraints from other modules that reference entities
-- These constraints might exist if other migrations have been run
ALTER TABLE IF EXISTS chartofaccount DROP CONSTRAINT IF EXISTS chartofaccount_entity_id_fkey;
ALTER TABLE IF EXISTS account DROP CONSTRAINT IF EXISTS account_entity_id_fkey;

-- Drop constraints that might exist from user module
DO $$
BEGIN
    -- Check if users table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        EXECUTE 'ALTER TABLE users DROP CONSTRAINT IF EXISTS users_entity_id_fkey';
    END IF;
    
    -- Check if employees table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'employees') THEN
        EXECUTE 'ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_entity_id_fkey';
        EXECUTE 'ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_department_id_fkey';
    END IF;
    
    -- Check if persons table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'persons') THEN
        EXECUTE 'ALTER TABLE persons DROP CONSTRAINT IF EXISTS persons_entity_id_fkey';
    END IF;
    
    -- Check if roles table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        EXECUTE 'ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_entity_id_fkey';
    END IF;
    
    -- Check if user_roles table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_roles') THEN
        EXECUTE 'ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_entity_id_fkey';
    END IF;
    
    -- Check if resources table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'resources') THEN
        EXECUTE 'ALTER TABLE resources DROP CONSTRAINT IF EXISTS resources_entity_id_fkey';
    END IF;
    
    -- Check if policies table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'policies') THEN
        EXECUTE 'ALTER TABLE policies DROP CONSTRAINT IF EXISTS policies_entity_id_fkey';
    END IF;
    
    -- Check if access_requests table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'access_requests') THEN
        EXECUTE 'ALTER TABLE access_requests DROP CONSTRAINT IF EXISTS access_requests_entity_id_fkey';
    END IF;
    
    -- Check if audit_log table exists and drop entity_id constraint
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'audit_log') THEN
        EXECUTE 'ALTER TABLE audit_log DROP CONSTRAINT IF EXISTS audit_log_entity_id_fkey';
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Some constraints may not exist, continuing with migration';
END;
$$;

-- =====================================================================
-- DROP TABLES IN DEPENDENCY ORDER
-- =====================================================================

-- Drop entitystate table (depends on entities)
DROP TABLE IF EXISTS entitystate;

-- Drop hierarchy_paths table (depends on entities)
DROP TABLE IF EXISTS hierarchy_paths;

-- Drop entities table (no dependencies)
DROP TABLE IF EXISTS entities;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '=====================================================================';
    RAISE NOTICE 'ENTITIES MODULE DOWN MIGRATION COMPLETED';
    RAISE NOTICE '=====================================================================';
    RAISE NOTICE 'Successfully removed:';
    RAISE NOTICE '- 10 views (RLS inherited from underlying tables)';
    RAISE NOTICE '- 3 tables (entities, hierarchy_paths, entitystate)';
    RAISE NOTICE '- All associated indexes, triggers, and constraints';
    RAISE NOTICE '- All RLS policies and security settings';
    RAISE NOTICE '=====================================================================';
END;
$$;