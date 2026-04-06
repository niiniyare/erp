DROP POLICY IF EXISTS role_permissions_tenant_isolation ON role_permissions;
DROP POLICY IF EXISTS role_permissions_admin_access ON role_permissions;
DROP POLICY IF EXISTS role_permissions_ro_select ON role_permissions;
ALTER TABLE IF EXISTS role_permissions NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS role_permissions;
