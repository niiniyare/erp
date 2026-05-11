-- =============================================================================
-- V2.0 RESERVED — ABAC migration. DO NOT DROP. Not active in v1.0.
-- =============================================================================

DROP POLICY IF EXISTS roles_tenant_isolation ON roles;
DROP POLICY IF EXISTS roles_admin_access ON roles;
DROP POLICY IF EXISTS roles_ro_select ON roles;
ALTER TABLE IF EXISTS roles NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS roles;
