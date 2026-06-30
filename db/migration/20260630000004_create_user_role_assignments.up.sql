-- Phase 9: IAM — user_role_assignments table.
--
-- Binds a tenant_user to a role, optionally scoped to an org unit.
-- org_unit_id=NULL means tenant-wide (no OU restriction on permissions;
-- RLS still limits data visibility to the user's primary OU subtree).
-- expires_at=NULL means no expiry.

CREATE TABLE user_role_assignments (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id    UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES tenant_users(id) ON DELETE CASCADE,
    role_id      UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    org_unit_id  UUID        REFERENCES orgunits(uuid) ON DELETE CASCADE,  -- NULL = tenant-wide
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    expires_at   TIMESTAMPTZ,
    granted_by   UUID        REFERENCES tenant_users(id) ON DELETE SET NULL,
    reason       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, role_id, org_unit_id)
);

CREATE INDEX user_role_assignments_user_id    ON user_role_assignments (user_id);
CREATE INDEX user_role_assignments_tenant_id  ON user_role_assignments (tenant_id);
CREATE INDEX user_role_assignments_role_id    ON user_role_assignments (role_id);

ALTER TABLE user_role_assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_role_assignments FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON user_role_assignments
    USING (tenant_id = current_tenant_id());
