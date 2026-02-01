-- =====================================================
-- TENANT USAGE STATISTICS TABLE
-- =====================================================
-- Tracks tenant resource usage and performance metrics
CREATE TABLE tenant_usage_stats (
  -- Composite primary key
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  period_start DATE NOT NULL,
  period_end DATE NOT NULL,
  -- Usage metrics
  active_users INT NOT NULL DEFAULT 0,
  total_entities INT NOT NULL DEFAULT 0,
  total_transactions INT NOT NULL DEFAULT 0,
  storage_used BIGINT NOT NULL DEFAULT 0,
  api_calls INT NOT NULL DEFAULT 0,
  -- Performance metrics
  avg_response_time NUMERIC(10, 2),
  error_rate NUMERIC(5, 4),
  -- Financial metrics
  monthly_revenue NUMERIC(12, 2),
  -- Audit timestamp
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- Primary key constraint
  PRIMARY KEY (tenant_id, period_start)
);

-- Create indexes for performance
CREATE INDEX idx_tenant_usage_stats_period ON tenant_usage_stats(period_start, period_end);

CREATE INDEX idx_tenant_usage_stats_tenant_period ON tenant_usage_stats(tenant_id, period_start);

-- Add comments for documentation
COMMENT ON TABLE tenant_usage_stats IS 'Tracks tenant resource usage and performance metrics over time';

COMMENT ON COLUMN tenant_usage_stats.storage_used IS 'Storage used in bytes';

COMMENT ON COLUMN tenant_usage_stats.avg_response_time IS 'Average response time in milliseconds';

COMMENT ON COLUMN tenant_usage_stats.error_rate IS 'Error rate as decimal (0.0001 = 0.01%)';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================
-- Enable RLS on tenant_usage_stats table
ALTER TABLE
  tenant_usage_stats ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant usage stats isolation
CREATE POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats FOR ALL TO application_role USING (tenant_id = current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats IS 'Ensures tenant usage statistics data isolation';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================
-- Grant necessary permissions to application role
GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON tenant_usage_stats TO application_role;

-- =====================================================
-- UPDATE CHECK_TENANT_LIMITS FUNCTION
-- =====================================================
-- Now that tenant_usage_stats exists, we can implement the storage check
CREATE
OR REPLACE FUNCTION check_tenant_limits(
  p_tenant_id UUID,
  p_check_type VARCHAR(50),
  p_additional_usage INT DEFAULT 1
) RETURNS BOOLEAN AS
$$
DECLARE
v_config tenant_configurations % ROWTYPE;

v_current_usage INT;

BEGIN
-- Get tenant configuration
SELECT
  * INTO v_config
FROM
  tenant_configurations
WHERE
  tenant_id = p_tenant_id;

-- If no configuration found, deny operation
IF v_config.tenant_id IS NULL THEN RAISE NOTICE 'No configuration found for tenant: %',
p_tenant_id;

RETURN FALSE;

END IF;

-- Check different types of limits
CASE
  p_check_type
  WHEN 'users' THEN
  -- Check user limit (placeholder for when users table exists)
  RETURN TRUE;

WHEN 'entities' THEN
-- Check entity limit (placeholder for when entities table exists)
RETURN TRUE;

WHEN 'transactions' THEN
-- Check monthly transaction limit (placeholder)
RETURN TRUE;

WHEN 'storage' THEN
-- Check storage limit using tenant_usage_stats
SELECT
  COALESCE(storage_used, 0) INTO v_current_usage
FROM
  tenant_usage_stats
WHERE
  tenant_id = p_tenant_id
  AND period_start <= CURRENT_DATE
  AND period_end >= CURRENT_DATE
ORDER BY
  period_start DESC
LIMIT
  1;

RETURN (
  COALESCE(v_current_usage, 0) + p_additional_usage
) <= v_config.storage_quota;

ELSE
-- Unknown check type
RAISE NOTICE 'Unknown check type: %',
p_check_type;

RETURN FALSE;

END CASE
;

END;

$$
LANGUAGE plpgsql;

-- Update function comment
COMMENT ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) IS 'Validates tenant resource limits before operations - now includes storage limit checking';
