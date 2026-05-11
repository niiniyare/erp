-- =============================================================================
-- V2.0 RESERVED — ABAC migration. DO NOT DROP. Not active in v1.0.
-- =============================================================================

DROP POLICY IF EXISTS user_roles_tenant_isolation ON user_roles;
DROP POLICY IF EXISTS user_roles_admin_access ON user_roles;
DROP POLICY IF EXISTS user_roles_ro_select ON user_roles;
ALTER TABLE IF EXISTS user_roles NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS user_roles;
