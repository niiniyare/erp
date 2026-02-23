-- ------------------------------------------------------------------------------------------------
-- DIRECT USER PERMISSIONS
-- ------------------------------------------------------------------------------------------------
-- Direct permission grants to users bypassing roles for exceptional access.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid),
  effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
  reason TEXT,  -- Justification for direct permission
  granted_by UUID REFERENCES users(id),
  granted_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ,  -- For temporary permissions
  conditions JSONB DEFAULT '{}'::jsonb,  -- Additional conditions
  is_active BOOLEAN DEFAULT TRUE,
  CONSTRAINT user_permissions_unique_assignment UNIQUE (tenant_id, user_id, permission_id, entity_id)
);

COMMENT ON TABLE user_permissions IS 'Direct permission grants to users bypassing roles. Used for exceptional access, denials, and temporary permissions.';

COMMENT ON COLUMN user_permissions.effect IS 'Permission effect: ALLOW (grant access) or DENY (explicitly deny - overrides role permissions)';

COMMENT ON COLUMN user_permissions.reason IS 'Business justification for this direct permission assignment';

COMMENT ON COLUMN user_permissions.granted_by IS 'User who granted this direct permission';

-- Enable RLS and create policies
ALTER TABLE
  user_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_permissions_tenant_isolation ON user_permissions FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY user_permissions_admin_access ON user_permissions FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
