-- =====================================================
-- TENANTS CORE DOWN MIGRATION
-- =====================================================================
-- 
-- IMPORTANT: This migration should only be run when ALL dependent 
-- migrations have been rolled back first, as it drops shared functions
-- and roles used by other migrations.
-- 
-- Order of operations:
-- 1. Drop table-specific triggers and policies
-- 2. Drop tenant table and indexes  
-- 3. Leave shared functions/roles for last migration rollback
-- =====================================================================

-- Drop triggers specific to tenants table
DROP TRIGGER IF EXISTS tenant_slug_trigger ON tenants;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

-- Drop RLS policies specific to tenants table
DROP POLICY IF EXISTS tenant_isolation_policy ON tenants;

-- Disable RLS on tenants table
ALTER TABLE IF EXISTS tenants DISABLE ROW LEVEL SECURITY;

-- Drop indexes specific to tenants table
DROP INDEX IF EXISTS idx_tenants_deleted_at;
DROP INDEX IF EXISTS idx_tenants_subdomain;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_slug;

-- Drop tenants table
DROP TABLE IF EXISTS tenants;

-- Drop tenant-specific functions that no other migrations use
DROP FUNCTION IF EXISTS generate_slug_from_name();

-- =====================================================================
-- SHARED FUNCTIONS AND ROLES CLEANUP
-- =====================================================================
-- Only drop these if this is the LAST migration being rolled back
-- These functions/roles are used by multiple migrations:
-- - current_tenant_id() - used by all RLS policies
-- - set_tenant_context() - used by application
-- - update_updated_at_column() - used by multiple tables
-- - application_role/admin_role - used by all RLS policies

-- Check if any other tables still exist before dropping shared resources
DO $$
BEGIN
    -- Only drop shared functions if no other tenant-isolated tables exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name IN ('tenant_configurations', 'tenant_usage_stats', 'entities')
    ) THEN
        -- Safe to drop shared functions and roles
        DROP FUNCTION IF EXISTS update_updated_at_column();
        DROP FUNCTION IF EXISTS current_tenant_id();
        DROP FUNCTION IF EXISTS set_tenant_context(UUID);
        
        -- Drop roles only if no policies reference them
        DROP ROLE IF EXISTS admin_role;
        DROP ROLE IF EXISTS application_role;
        
        RAISE NOTICE 'Cleaned up shared functions and roles - no dependent objects found';
    ELSE
        RAISE NOTICE 'Keeping shared functions and roles - other dependent tables still exist';
    END IF;
END;
$$;