-- ------------------------------------------------------------------------------------------------
-- ACTIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines actions that can be performed on resources, with risk and approval requirements.
-- scope IN ('SYSTEM', 'TENANT'):
--   SYSTEM actions are standard platform actions (CREATE, READ, APPROVE, etc.).
--   TENANT actions are custom actions defined by a specific tenant.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS actions (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID        REFERENCES tenants(id) ON DELETE CASCADE,
  scope        VARCHAR(10) NOT NULL DEFAULT 'SYSTEM'
                             CHECK (scope IN ('SYSTEM', 'TENANT')),
  name         VARCHAR(100) NOT NULL,
  display_name VARCHAR(150),
  description  TEXT,
  action_type  VARCHAR(50)  NOT NULL CHECK (
    action_type IN (
      'CREATE', 'READ', 'UPDATE', 'DELETE',
      'EXECUTE', 'APPROVE', 'REJECT', 'EXPORT', 'IMPORT'
    )
  ),
  action_category VARCHAR(50) DEFAULT 'STANDARD' CHECK (
    action_category IN ('STANDARD', 'ADMINISTRATIVE', 'SENSITIVE', 'BULK', 'SYSTEM')
  ),
  risk_level VARCHAR(20) DEFAULT 'LOW' CHECK (
    risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')
  ),
  -- When requires_approval=TRUE, approver_role_id MUST be set.
  -- A flag without an approver cannot drive a workflow.
  requires_approval BOOLEAN DEFAULT false,
  approver_role_id  UUID REFERENCES roles(id) ON DELETE SET NULL,
  is_active    BOOLEAN     DEFAULT TRUE,
  created_at   TIMESTAMPTZ DEFAULT NOW(),
  updated_at   TIMESTAMPTZ DEFAULT NOW(),
  -- SYSTEM actions must have tenant_id=NULL; TENANT actions must have a tenant_id
  CONSTRAINT actions_scope_tenant_check CHECK (
    (scope = 'SYSTEM' AND tenant_id IS NULL)
    OR (scope = 'TENANT' AND tenant_id IS NOT NULL)
  ),
  -- When approval is required, an approver role must be designated
  CONSTRAINT actions_approval_requires_role CHECK (
    NOT requires_approval OR approver_role_id IS NOT NULL
  )
);

COMMENT ON TABLE  actions                   IS 'Defines actions that can be performed on resources with risk assessment and approval workflow requirements.';
COMMENT ON COLUMN actions.tenant_id         IS 'NULL for SYSTEM-scope actions. Set for custom TENANT-scope actions.';
COMMENT ON COLUMN actions.scope             IS 'SYSTEM = platform-wide standard action. TENANT = custom action for one tenant.';
COMMENT ON COLUMN actions.action_type       IS 'Standard action type: CREATE, READ, UPDATE, DELETE, EXECUTE, APPROVE, REJECT, EXPORT, IMPORT';
COMMENT ON COLUMN actions.action_category   IS 'Risk category: STANDARD, ADMINISTRATIVE, SENSITIVE, BULK, SYSTEM';
COMMENT ON COLUMN actions.risk_level        IS 'Risk level for audit and approval routing: LOW, MEDIUM, HIGH, CRITICAL';
COMMENT ON COLUMN actions.requires_approval IS 'Whether this action requires explicit approval. If TRUE, approver_role_id MUST be set.';
COMMENT ON COLUMN actions.approver_role_id  IS 'Role whose members can approve this action. Required when requires_approval=TRUE.';
COMMENT ON COLUMN actions.updated_at        IS 'Updated by trigger on every row change — use for cache invalidation.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_actions_scope    ON actions(scope, is_active)    WHERE is_active = TRUE;
CREATE INDEX idx_actions_tenant   ON actions(tenant_id)           WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_actions_type     ON actions(action_type, scope);
CREATE INDEX idx_actions_approval ON actions(approver_role_id)    WHERE requires_approval = TRUE;

-- ------------------------------------------------------------------------------------------------
-- updated_at TRIGGER
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_actions_updated_at
  BEFORE UPDATE ON actions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE actions ENABLE ROW LEVEL SECURITY;

-- application_role: see all SYSTEM actions + their own TENANT actions
CREATE POLICY actions_read ON actions
  FOR SELECT TO application_role
  USING (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  );

CREATE POLICY actions_write ON actions
  FOR INSERT TO application_role
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY actions_update ON actions
  FOR UPDATE TO application_role
  USING  (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY actions_admin ON actions
  FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

GRANT SELECT, INSERT, UPDATE, DELETE ON actions TO application_role;
