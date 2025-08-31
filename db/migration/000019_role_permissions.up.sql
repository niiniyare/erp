-- ------------------------------------------------------------------------------------------------
-- ROLE PERMISSIONS MAPPING
-- ------------------------------------------------------------------------------------------------
-- Maps permissions to roles with entity-specific scoping and conditions.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS role_permissions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  entity_scope UUID REFERENCES entities(uuid),  -- Entity-specific permission scope
  granted_by UUID REFERENCES users(id),
  granted_at TIMESTAMPTZ DEFAULT NOW(),
  conditions JSONB DEFAULT '{}'::jsonb,  -- Additional conditions beyond permission
  is_active BOOLEAN DEFAULT TRUE,
  CONSTRAINT role_permissions_unique_assignment UNIQUE (tenant_id, role_id, permission_id, entity_scope)
);

COMMENT ON TABLE role_permissions IS 'Maps permissions to roles with optional entity-specific scoping and additional conditions for flexible authorization.';

COMMENT ON COLUMN role_permissions.entity_scope IS 'Optional entity restriction - if specified, permission only applies within this entity';

COMMENT ON COLUMN role_permissions.granted_by IS 'User who granted this permission assignment';

COMMENT ON COLUMN role_permissions.conditions IS 'Additional JSONB conditions beyond those in the permission itself';

-- Enable RLS and create policies
ALTER TABLE
  role_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY role_permissions_tenant_isolation ON role_permissions FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY role_permissions_admin_access ON role_permissions FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
