-- ------------------------------------------------------------------------------------------------
-- POLICY EVALUATIONS CACHE
-- ------------------------------------------------------------------------------------------------
-- Caches policy evaluation results for performance optimization.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policy_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    resource_id UUID NOT NULL REFERENCES resources(id),
    action_id UUID NOT NULL REFERENCES actions(id),
    context_hash VARCHAR(64) NOT NULL,             -- Hash of evaluation context
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
    applicable_policies UUID[] DEFAULT '{}',       -- Array of policy IDs that fired
    evaluation_time_ms INTEGER,                    -- Performance metric
    evaluated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour')
);

COMMENT ON TABLE policy_evaluations IS 
'Caches ABAC policy evaluation results for performance optimization with configurable TTL and context tracking.';

COMMENT ON COLUMN policy_evaluations.context_hash IS 'SHA-256 hash of evaluation context for cache key uniqueness';
COMMENT ON COLUMN policy_evaluations.applicable_policies IS 'Array of policy UUIDs that were evaluated and fired';
COMMENT ON COLUMN policy_evaluations.evaluation_time_ms IS 'Policy evaluation time in milliseconds for performance monitoring';

-- Enable RLS and create policies
ALTER TABLE policy_evaluations ENABLE ROW LEVEL SECURITY;

CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);


