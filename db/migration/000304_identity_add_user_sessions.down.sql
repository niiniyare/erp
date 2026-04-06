DROP POLICY IF EXISTS user_sessions_tenant_isolation ON user_sessions;
DROP POLICY IF EXISTS user_sessions_admin_access ON user_sessions;
DROP POLICY IF EXISTS user_sessions_ro_select ON user_sessions;
ALTER TABLE IF EXISTS user_sessions NO FORCE ROW LEVEL SECURITY;
DROP INDEX IF EXISTS idx_user_sessions_user_type;
DROP TABLE IF EXISTS user_sessions;
