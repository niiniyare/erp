DROP POLICY IF EXISTS sso_providers_tenant_isolation ON sso_providers;
DROP POLICY IF EXISTS sso_providers_admin_access ON sso_providers;
DROP POLICY IF EXISTS sso_providers_ro_select ON sso_providers;
ALTER TABLE IF EXISTS sso_providers NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS sso_providers;
