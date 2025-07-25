-- ------------------------------------------------------------------------------------------------
-- USER ROLE ASSIGNMENTS
-- ------------------------------------------------------------------------------------------------
-- Assigns roles to users with entity context, delegation, and temporal controls.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    assignment_type VARCHAR(20) DEFAULT 'DIRECT'
        CHECK (assignment_type IN ('DIRECT', 'INHERITED', 'DELEGATED', 'TEMPORARY')),
    delegated_by UUID REFERENCES users(id),        -- If delegated assignment
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,                        -- For temporary assignments
    conditions JSONB DEFAULT '{}'::jsonb,          -- Time/location/device conditions
    is_active BOOLEAN DEFAULT true,

    CONSTRAINT user_roles_unique_assignment UNIQUE (user_id, role_id, entity_id)
);

COMMENT ON TABLE user_roles IS
'Assigns roles to users with entity context, delegation support, and temporal controls for dynamic authorization.';

COMMENT ON COLUMN user_roles.assignment_type IS 'Type of assignment: DIRECT (explicitly assigned), INHERITED (from hierarchy), DELEGATED (from another user), TEMPORARY (time-limited)';
COMMENT ON COLUMN user_roles.delegated_by IS 'User who delegated this role assignment (for DELEGATED type)';
COMMENT ON COLUMN user_roles.expires_at IS 'Expiration timestamp for temporary role assignments';
COMMENT ON COLUMN user_roles.conditions IS 'JSONB containing conditional access rules (time, location, device, etc.)';

-- Enable RLS and create policies
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_roles_tenant_isolation ON user_roles
    FOR ALL TO application_role
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

CREATE POLICY user_roles_admin_access ON user_roles
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
