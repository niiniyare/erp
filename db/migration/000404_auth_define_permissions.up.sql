-- =============================================================================
-- V2.0 RESERVED — ABAC migration. DO NOT DROP. Not active in v1.0.
-- internal/core/access/ is gated with //go:build ignore until v2.0.
-- =============================================================================

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Granular permissions combining resources and actions with ABAC conditions, data filters,
-- and field restrictions for fine-grained access control.
-- effect IN ('ALLOW', 'DENY'): ALLOW grants access; DENY explicitly blocks it.
--
-- NOTE: Depends on tenants(id), resources(id), and actions(id). RLS uses current_tenant_id().
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permissions (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  resource_id        UUID         NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  action_id          UUID         NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
  name               VARCHAR(200) NOT NULL,
  display_name       VARCHAR(250),
  description        TEXT,
  effect             VARCHAR(5)   DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
  conditions         JSONB        DEFAULT '{}'::jsonb,  -- ABAC evaluation conditions
  data_filters       JSONB        DEFAULT '{}'::jsonb,  -- Row-level security filters
  field_restrictions JSONB        DEFAULT '{}'::jsonb,  -- Column-level restrictions
  is_active          BOOLEAN      DEFAULT TRUE,
  created_at         TIMESTAMPTZ  DEFAULT NOW(),
  CONSTRAINT permissions_unique_per_resource_action UNIQUE (tenant_id, resource_id, action_id, name)
);

COMMENT ON TABLE permissions IS 'Granular permissions combining resources and actions with ABAC conditions, data filters, and field restrictions for fine-grained access control.';

COMMENT ON COLUMN permissions.effect            IS 'Permission effect: ALLOW (grant access) or DENY (explicitly deny access)';
COMMENT ON COLUMN permissions.conditions        IS 'JSONB containing ABAC evaluation conditions (time, location, attributes, etc.)';
COMMENT ON COLUMN permissions.data_filters      IS 'JSONB containing row-level security filters to limit data access';
COMMENT ON COLUMN permissions.field_restrictions IS 'JSONB containing column-level restrictions to limit field access';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE permissions FORCE  ROW LEVEL SECURITY;

CREATE POLICY permissions_tenant_isolation ON permissions FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY permissions_admin_access ON permissions FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY permissions_ro_select ON permissions
    FOR SELECT TO readonly_role
    USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
