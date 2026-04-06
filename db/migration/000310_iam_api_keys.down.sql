DROP POLICY IF EXISTS api_keys_tenant_isolation ON api_keys;
DROP POLICY IF EXISTS api_keys_admin_access ON api_keys;
DROP POLICY IF EXISTS api_keys_ro_select ON api_keys;
ALTER TABLE IF EXISTS api_keys NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS api_keys;
