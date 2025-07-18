-- Tenant context function
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM tenants
        WHERE id = tenant_id AND status = 'active'
    ) THEN
        RAISE EXCEPTION 'Invalid or inactive tenant: %', tenant_id;
    END IF;

    PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
    RAISE NOTICE 'Tenant context set to: %', tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Current tenant function
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN COALESCE(nullif(current_setting('app.current_tenant_id', true), ''))::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Enable RLS for tenants table
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenants_isolation ON tenants;
CREATE POLICY tenants_isolation ON tenants
    FOR ALL TO PUBLIC
    USING (id = current_tenant_id());

-- Enable RLS for entities table
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS entities_isolation ON entities;
CREATE POLICY entities_isolation ON entities
    FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

-- Enable RLS for hierarchy_paths table
ALTER TABLE hierarchy_paths ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS hierarchy_paths_isolation ON hierarchy_paths;
CREATE POLICY hierarchy_paths_isolation ON hierarchy_paths
    FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

-- Enable RLS for persons table
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS persons_isolation ON persons;
CREATE POLICY persons_isolation ON persons
    FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

-- Enable RLS for employees table
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS employees_isolation ON employees;
CREATE POLICY employees_isolation ON employees
    FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

-- Enable RLS for users table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS users_isolation ON users;
CREATE POLICY users_isolation ON users
    FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());
