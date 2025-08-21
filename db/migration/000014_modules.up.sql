-- ------------------------------------------------------------------------------------------------
-- MODULES TABLE
-- ------------------------------------------------------------------------------------------------
-- Organizes system functionality into modules for permission management and feature control.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    category VARCHAR(50),                          -- 'CORE', 'HR', 'FINANCE', 'SALES', etc.
    version VARCHAR(20),
    is_active BOOLEAN DEFAULT true,

    -- -- Standard validation columns
    -- validation_version INTEGER NOT NULL DEFAULT 1,
    -- last_validation_run TIMESTAMPTZ,
    -- validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
    --     validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    -- ),
    -- validation_errors JSONB DEFAULT '[]'::jsonb,
    --
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT modules_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE modules IS
'System modules for organizing permissions and features into logical groups. Enables modular permission management and feature toggles.';

COMMENT ON COLUMN modules.category IS 'Module category for grouping: CORE, HR, FINANCE, SALES, INVENTORY, etc.';
COMMENT ON COLUMN modules.version IS 'Module version for tracking feature updates and compatibility';

-- Enable RLS and create policies
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;

CREATE POLICY modules_tenant_isolation ON modules
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY modules_admin_access ON modules
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
