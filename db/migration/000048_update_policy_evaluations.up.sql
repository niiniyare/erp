-- ------------------------------------------------------------------------------------------------
-- UPDATE POLICY EVALUATIONS FOR ABAC
-- ------------------------------------------------------------------------------------------------
-- Updates policy evaluations table to support flexible resource types for ABAC.
-- ------------------------------------------------------------------------------------------------
-- Drop existing table and recreate with flexible schema
DROP TABLE IF EXISTS policy_evaluations CASCADE;

CREATE TABLE IF NOT EXISTS policy_evaluations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  resource_type VARCHAR(100) NOT NULL,  -- Flexible resource type (user, document, etc.)
  resource_id UUID,  -- Optional specific resource ID
  ACTION VARCHAR(100) NOT NULL,  -- Action being performed (read, write, etc.)
  entity_id UUID REFERENCES entities(uuid),  -- Optional entity context
  context_hash VARCHAR(64) NOT NULL,  -- Hash of evaluation context
  decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
  applicable_policies UUID [] DEFAULT '{}',  -- Array of policy IDs that fired
  policy_decisions JSONB DEFAULT '[]'::jsonb,  -- Detailed policy decisions
  evaluation_time_ms INTEGER,  -- Performance metric
  cache_key VARCHAR(255),  -- Optional cache key for faster lookup
  evaluated_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour'),
  -- Unique constraint for cache lookups
  CONSTRAINT policy_evaluations_unique_cache UNIQUE (
    tenant_id,
    user_id,
    resource_type,
    resource_id,
    ACTION,
    context_hash
  )
);

COMMENT ON TABLE policy_evaluations IS 'Caches ABAC policy evaluation results with flexible resource types and detailed decision tracking for performance optimization.';

COMMENT ON COLUMN policy_evaluations.resource_type IS 'Type of resource being accessed (user, document, report, system, etc.)';

COMMENT ON COLUMN policy_evaluations.resource_id IS 'Optional specific resource identifier';

COMMENT ON COLUMN policy_evaluations.action IS 'Action being performed (read, write, delete, execute, etc.)';

COMMENT ON COLUMN policy_evaluations.context_hash IS 'SHA-256 hash of evaluation context for cache key uniqueness';

COMMENT ON COLUMN policy_evaluations.applicable_policies IS 'Array of policy UUIDs that were evaluated and contributed to the decision';

COMMENT ON COLUMN policy_evaluations.policy_decisions IS 'JSONB array containing detailed policy decision information';

COMMENT ON COLUMN policy_evaluations.evaluation_time_ms IS 'Policy evaluation time in milliseconds for performance monitoring';

-- Enable RLS and create policies
ALTER TABLE
  policy_evaluations ENABLE ROW LEVEL SECURITY;

CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);

-- Create indexes for performance
CREATE INDEX idx_policy_evaluations_cache_lookup ON policy_evaluations(
  tenant_id,
  user_id,
  resource_type,
  resource_id,
  ACTION,
  context_hash
);

CREATE INDEX idx_policy_evaluations_user_resource ON policy_evaluations(tenant_id, user_id, resource_type);

CREATE INDEX idx_policy_evaluations_expires_at ON policy_evaluations(expires_at);

CREATE INDEX idx_policy_evaluations_resource_action ON policy_evaluations(tenant_id, resource_type, ACTION);
