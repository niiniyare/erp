-- =============================================================================
-- V2.0 RESERVED — ABAC migration. DO NOT DROP. Not active in v1.0.
-- internal/core/access/ is gated with //go:build ignore until v2.0.
-- =============================================================================

-- ------------------------------------------------------------------------------------------------
-- USER_ROLES TABLE
-- ------------------------------------------------------------------------------------------------
-- Assigns roles to users with entity context, delegation support, and temporal controls
-- for dynamic authorization.
-- assignment_type IN ('DIRECT','INHERITED','DELEGATED','TEMPORARY'):
--   DIRECT = explicitly assigned; INHERITED = from entity hierarchy;
--   DELEGATED = granted by another user; TEMPORARY = time-limited.
--
-- NOTE: Depends on users(id), roles(id), and entities(uuid). RLS checks tenant membership
--       via a subquery on users.tenant_id since user_roles has no direct tenant_id column.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_roles (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id         UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  entity_id       UUID        NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
  assignment_type VARCHAR(20) DEFAULT 'DIRECT' CHECK (
    assignment_type IN ('DIRECT', 'INHERITED', 'DELEGATED', 'TEMPORARY')
  ),
  delegated_by    UUID        REFERENCES users(id),  -- If delegated assignment
  assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  assigned_by     UUID        REFERENCES users(id),
  expires_at      TIMESTAMPTZ,                        -- For temporary assignments
  conditions      JSONB       DEFAULT '{}'::jsonb,    -- Time/location/device conditions
  is_active       BOOLEAN     DEFAULT TRUE,
  CONSTRAINT user_roles_unique_assignment UNIQUE (user_id, role_id, entity_id)
);

COMMENT ON TABLE user_roles IS 'Assigns roles to users with entity context, delegation support, and temporal controls for dynamic authorization.';

COMMENT ON COLUMN user_roles.assignment_type IS 'Type of assignment: DIRECT (explicitly assigned), INHERITED (from hierarchy), DELEGATED (from another user), TEMPORARY (time-limited)';
COMMENT ON COLUMN user_roles.delegated_by    IS 'User who delegated this role assignment (for DELEGATED type)';
COMMENT ON COLUMN user_roles.expires_at      IS 'Expiration timestamp for temporary role assignments';
COMMENT ON COLUMN user_roles.conditions      IS 'JSONB containing conditional access rules (time, location, device, etc.)';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles FORCE  ROW LEVEL SECURITY;

CREATE POLICY user_roles_tenant_isolation ON user_roles FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND EXISTS (
            SELECT 1 FROM users u
            WHERE u.id = user_roles.user_id
              AND u.tenant_id = current_tenant_id()
        )
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND EXISTS (
            SELECT 1 FROM users u
            WHERE u.id = user_roles.user_id
              AND u.tenant_id = current_tenant_id()
        )
    );

CREATE POLICY user_roles_admin_access ON user_roles FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY user_roles_ro_select ON user_roles
    FOR SELECT TO readonly_role
    USING (
        current_tenant_id() IS NOT NULL
        AND EXISTS (
            SELECT 1 FROM users u
            WHERE u.id = user_roles.user_id
              AND u.tenant_id = current_tenant_id()
        )
    );
