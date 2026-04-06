DROP POLICY IF EXISTS user_permissions_tenant_isolation ON user_permissions;
DROP POLICY IF EXISTS user_permissions_admin_access ON user_permissions;
DROP POLICY IF EXISTS user_permissions_ro_select ON user_permissions;
ALTER TABLE IF EXISTS user_permissions NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS user_permissions;
