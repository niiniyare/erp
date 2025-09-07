-- Revert to the original policy_evaluations table structure
DROP TABLE IF EXISTS policy_evaluations CASCADE;

-- Recreate original table (this should match the original 000047 migration)
CREATE TABLE IF NOT EXISTS policy_evaluations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  resource_id UUID NOT NULL REFERENCES resources(id),
  action_id UUID NOT NULL REFERENCES actions(id),
  context_hash VARCHAR(64) NOT NULL,
  decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
  applicable_policies UUID [] DEFAULT '{}',
  evaluation_time_ms INTEGER,
  evaluated_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour')
);

-- Enable RLS and create policies
ALTER TABLE
  policy_evaluations ENABLE ROW LEVEL SECURITY;

CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);
