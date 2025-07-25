DROP TABLE IF EXISTS user_permissions;

DROP POLICY IF EXISTS user_permissions_tenant_isolation ON user_permissions;
DROP POLICY IF EXISTS user_permissions_admin_access ON user_permissions;

