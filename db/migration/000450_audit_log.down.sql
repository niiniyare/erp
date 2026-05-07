REVOKE DELETE ON audit_log FROM audit_retention_role;

DROP POLICY IF EXISTS audit_log_tenant_isolation ON audit_log;
DROP POLICY IF EXISTS audit_log_admin_access ON audit_log;
DROP POLICY IF EXISTS audit_log_ro_select ON audit_log;
ALTER TABLE IF EXISTS audit_log NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS audit_log;
