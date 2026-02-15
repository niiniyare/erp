-- ------------------------------------------------------------------------------------------------
-- ACTIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines actions that can be performed on resources with risk and approval requirements.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS actions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  display_name VARCHAR(150),
  description TEXT,
  action_type VARCHAR(50) NOT NULL CHECK (
    action_type IN (
      'CREATE',
      'READ',
      'UPDATE',
      'DELETE',
      'EXECUTE',
      'APPROVE',
      'REJECT',
      'EXPORT',
      'IMPORT'
    )
  ),
  action_category VARCHAR(50) DEFAULT 'STANDARD' CHECK (
    action_category IN (
      'STANDARD',
      'ADMINISTRATIVE',
      'SENSITIVE',
      'BULK',
      'SYSTEM'
    )
  ),
  risk_level VARCHAR(20) DEFAULT 'LOW' CHECK (
    risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')
  ),
  requires_approval BOOLEAN DEFAULT false,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  -- CONSTRAINT actions_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE actions IS 'Defines actions that can be performed on resources with risk assessment and approval workflow requirements.';

COMMENT ON COLUMN actions.action_type IS 'Standard action type: CREATE, READ, UPDATE, DELETE, EXECUTE, APPROVE, REJECT, EXPORT, IMPORT';

COMMENT ON COLUMN actions.action_category IS 'Action category for risk assessment: STANDARD, ADMINISTRATIVE, SENSITIVE, BULK, SYSTEM';

COMMENT ON COLUMN actions.risk_level IS 'Risk level for audit and approval workflows: LOW, MEDIUM, HIGH, CRITICAL';

COMMENT ON COLUMN actions.requires_approval IS 'Whether this action requires explicit approval before execution';

-- Enable RLS and create policies
-- ALTER TABLE
--   actions ENABLE ROW LEVEL SECURITY;
--
-- CREATE POLICY actions_tenant_isolation ON actions FOR ALL TO application_role USING (
--   current_tenant_id() IS NOT NULL
--   AND tenant_id = current_tenant_id()
-- ) WITH CHECK (
--   current_tenant_id() IS NOT NULL
--   AND tenant_id = current_tenant_id()
-- );
--
-- CREATE POLICY actions_admin_access ON actions FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
