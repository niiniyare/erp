ALTER TABLE users DROP COLUMN IF EXISTS password_history;
DROP POLICY IF EXISTS password_reset_tenant_isolation ON password_reset_tokens;
DROP POLICY IF EXISTS password_reset_admin_access ON password_reset_tokens;
DROP POLICY IF EXISTS password_reset_ro_select ON password_reset_tokens;
ALTER TABLE IF EXISTS password_reset_tokens NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS password_reset_tokens;
