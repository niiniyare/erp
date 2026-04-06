DROP POLICY IF EXISTS permissions_tenant_isolation ON permissions;
DROP POLICY IF EXISTS permissions_admin_access ON permissions;
DROP POLICY IF EXISTS permissions_ro_select ON permissions;
ALTER TABLE IF EXISTS permissions NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS permissions;
