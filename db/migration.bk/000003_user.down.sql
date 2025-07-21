-- Drop views
DROP VIEW IF EXISTS audit_summary_view CASCADE;
DROP VIEW IF EXISTS role_permissions_summary CASCADE;
DROP VIEW IF EXISTS user_complete_view CASCADE;

-- Drop admin access policies first
DROP POLICY IF EXISTS access_requests_admin_access ON access_requests;
DROP POLICY IF EXISTS audit_log_admin_access ON audit_log;
DROP POLICY IF EXISTS user_permissions_admin_access ON user_permissions;
DROP POLICY IF EXISTS user_roles_admin_access ON user_roles;
DROP POLICY IF EXISTS role_permissions_admin_access ON role_permissions;
DROP POLICY IF EXISTS permissions_admin_access ON permissions;
DROP POLICY IF EXISTS roles_admin_access ON roles;
DROP POLICY IF EXISTS actions_admin_access ON actions;
DROP POLICY IF EXISTS resources_admin_access ON resources;
DROP POLICY IF EXISTS modules_admin_access ON modules;
DROP POLICY IF EXISTS user_sessions_admin_access ON user_sessions;
DROP POLICY IF EXISTS users_admin_access ON users;
DROP POLICY IF EXISTS employees_admin_access ON employees;
DROP POLICY IF EXISTS persons_admin_access ON persons;

-- Drop tenant isolation policies
DROP POLICY IF EXISTS access_requests_tenant_isolation ON access_requests;
DROP POLICY IF EXISTS audit_log_tenant_isolation ON audit_log;
DROP POLICY IF EXISTS user_permissions_tenant_isolation ON user_permissions;
DROP POLICY IF EXISTS user_roles_tenant_isolation ON user_roles;
DROP POLICY IF EXISTS role_permissions_tenant_isolation ON role_permissions;
DROP POLICY IF EXISTS permissions_tenant_isolation ON permissions;
DROP POLICY IF EXISTS roles_tenant_isolation ON roles;
DROP POLICY IF EXISTS actions_tenant_isolation ON actions;
DROP POLICY IF EXISTS resources_tenant_isolation ON resources;
DROP POLICY IF EXISTS modules_tenant_isolation ON modules;
DROP POLICY IF EXISTS user_sessions_tenant_isolation ON user_sessions;
DROP POLICY IF EXISTS users_tenant_isolation ON users;
DROP POLICY IF EXISTS employees_tenant_isolation ON employees;
DROP POLICY IF EXISTS persons_tenant_isolation ON persons;

-- Drop RLS policies from other tables
DROP POLICY IF EXISTS policy_evaluations_tenant_isolation ON policy_evaluations;
DROP POLICY IF EXISTS policies_tenant_isolation ON policies;
DROP POLICY IF EXISTS attribute_definitions_tenant_isolation ON attribute_definitions;

-- Disable RLS
ALTER TABLE access_requests DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log DISABLE ROW LEVEL SECURITY;
ALTER TABLE policy_evaluations DISABLE ROW LEVEL SECURITY;
ALTER TABLE policies DISABLE ROW LEVEL SECURITY;
ALTER TABLE attribute_definitions DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_permissions DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles DISABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions DISABLE ROW LEVEL SECURITY;
ALTER TABLE permissions DISABLE ROW LEVEL SECURITY;
ALTER TABLE roles DISABLE ROW LEVEL SECURITY;
ALTER TABLE actions DISABLE ROW LEVEL SECURITY;
ALTER TABLE resources DISABLE ROW LEVEL SECURITY;
ALTER TABLE modules DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_sessions DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE employees DISABLE ROW LEVEL SECURITY;
ALTER TABLE persons DISABLE ROW LEVEL SECURITY;

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
DROP FUNCTION IF EXISTS user_has_permission(UUID, VARCHAR, VARCHAR, INTEGER, UUID, JSONB) CASCADE;
DROP FUNCTION IF EXISTS create_default_system_data(UUID) CASCADE;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS access_requests CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS policy_evaluations CASCADE;
DROP TABLE IF EXISTS policies CASCADE;
DROP TABLE IF EXISTS attribute_definitions CASCADE;
DROP TABLE IF EXISTS user_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS actions CASCADE;
DROP TABLE IF EXISTS resources CASCADE;
DROP TABLE IF EXISTS modules CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS employees CASCADE;
DROP TABLE IF EXISTS persons CASCADE;
