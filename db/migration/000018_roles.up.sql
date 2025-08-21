-- ------------------------------------------------------------------------------------------------
-- ROLES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines roles with module scope, entity context, and hierarchical structure.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    module_id UUID REFERENCES modules(id),         -- Module this role belongs to
    role_type VARCHAR(20) DEFAULT 'CUSTOM'
        CHECK (role_type IN ('SYSTEM', 'TENANT', 'ENTITY', 'CUSTOM', 'FUNCTIONAL')),
    parent_role_id UUID REFERENCES roles(id),      -- Role hierarchy
    level INTEGER DEFAULT 0,                       -- Hierarchy level (calculated)
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb, -- Direct permissions cache for performance
    entity_scope JSONB DEFAULT '{}'::jsonb,        -- Entity-level access rules
    conditions JSONB DEFAULT '{}'::jsonb,          -- Time, location, device conditions
    is_system_role BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,

    -- -- Standard validation columns
    -- version INTEGER NOT NULL DEFAULT 1,
    -- last_validation_run TIMESTAMPTZ,
    -- validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
    --     validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    -- ),
    -- validation_errors JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT roles_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE roles IS 'Roles with module association, entity scoping, and hierarchical structure. Supports both RBAC and ABAC with conditional access rules.';

COMMENT ON COLUMN roles.role_type IS 'Role classification: SYSTEM (built-in), TENANT (tenant-wide), ENTITY (entity-scoped), CUSTOM (user-defined), FUNCTIONAL (job-based)';
COMMENT ON COLUMN roles.parent_role_id IS 'Parent role for inheritance hierarchy';
COMMENT ON COLUMN roles.level IS 'Calculated hierarchy level (0=root, higher=deeper)';
COMMENT ON COLUMN roles.permissions IS 'Cached permissions JSONB for performance optimization';
COMMENT ON COLUMN roles.entity_scope IS 'JSONB defining which entities this role can access';
COMMENT ON COLUMN roles.conditions IS 'JSONB containing time, location, device, and other conditional access rules';

-- Enable RLS and create policies
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY roles_tenant_isolation ON roles
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY roles_admin_access ON roles
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
