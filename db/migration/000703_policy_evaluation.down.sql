-- ------------------------------------------------------------------------------------------------
-- DOWN MIGRATION: POLICY EVALUATIONS CACHE
--
-- Reverts the policy_evaluations table and associated policies.
-- ------------------------------------------------------------------------------------------------
-- Drop RLS policies
DROP POLICY IF EXISTS policy_evaluations_tenant_isolation ON policy_evaluations;
DROP POLICY IF EXISTS policy_evaluations_admin_access ON policy_evaluations;

-- Disable RLS
ALTER TABLE
  policy_evaluations DISABLE ROW LEVEL SECURITY;

-- Drop comments from columns
COMMENT ON COLUMN policy_evaluations.evaluation_time_ms IS NULL;

COMMENT ON COLUMN policy_evaluations.applicable_policies IS NULL;

COMMENT ON COLUMN policy_evaluations.context_hash IS NULL;

-- Drop comment from table
COMMENT ON TABLE policy_evaluations IS NULL;

-- Drop the policy_evaluations table
DROP TABLE IF EXISTS policy_evaluations;
