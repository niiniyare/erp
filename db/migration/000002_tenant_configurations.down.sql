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
ALTER TABLE IF EXISTS tenant_configurations DISABLE ROW LEVEL SECURITY;

-- Drop table
DROP TABLE IF EXISTS tenant_configurations;

-- Note: Shared functions like update_updated_at_column() and current_tenant_id() 
-- are managed by the core tenants migration and should not be dropped here