-- ------------------------------------------------------------------------------------------------
-- TENANT_USAGE_STATS
-- ------------------------------------------------------------------------------------------------
-- Tracks tenant resource usage and performance metrics over time, keyed by tenant + period.
-- Each row represents one billing/reporting period (period_start inclusive, period_end inclusive).
-- storage_used is in bytes; avg_response_time is in milliseconds; error_rate is a decimal
-- fraction (e.g. 0.0001 = 0.01%).
--
-- NOTE: This table is referenced by check_tenant_limits() defined in 000102, which is
--       re-implemented here (CREATE OR REPLACE) to add the live storage check now that
--       the table exists.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE tenant_usage_stats (
  tenant_id          UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  period_start       DATE          NOT NULL,
  period_end         DATE          NOT NULL,
  -- Usage metrics
  active_users       INT           NOT NULL DEFAULT 0,
  total_entities     INT           NOT NULL DEFAULT 0,
  total_transactions INT           NOT NULL DEFAULT 0,
  storage_used       BIGINT        NOT NULL DEFAULT 0,    -- bytes
  api_calls          INT           NOT NULL DEFAULT 0,
  -- Performance metrics
  avg_response_time  NUMERIC(10,2),                       -- milliseconds
  error_rate         NUMERIC(5,4),                        -- decimal fraction (0.0001 = 0.01%)
  -- Financial metrics
  monthly_revenue    NUMERIC(12,2),
  -- Audit timestamp
  created_at         TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, period_start)
);

COMMENT ON TABLE  tenant_usage_stats                   IS 'Tracks tenant resource usage and performance metrics over time';
COMMENT ON COLUMN tenant_usage_stats.storage_used      IS 'Storage used in bytes';
COMMENT ON COLUMN tenant_usage_stats.avg_response_time IS 'Average response time in milliseconds';
COMMENT ON COLUMN tenant_usage_stats.error_rate        IS 'Error rate as decimal (0.0001 = 0.01%)';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_tenant_usage_stats_period        ON tenant_usage_stats(period_start, period_end); -- period range queries
CREATE INDEX idx_tenant_usage_stats_tenant_period ON tenant_usage_stats(tenant_id, period_start);  -- per-tenant history

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE tenant_usage_stats ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id());

COMMENT ON POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats IS 'Ensures tenant usage statistics data isolation';

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_usage_stats TO application_role;

-- ------------------------------------------------------------------------------------------------
-- CHECK_TENANT_LIMITS — STORAGE CHECK IMPLEMENTATION
-- ------------------------------------------------------------------------------------------------
-- Re-implements check_tenant_limits() (originally stubbed in 000102) now that tenant_usage_stats
-- exists. The 'storage' branch performs a real check against the current period's storage_used
-- value. All other branches remain TRUE placeholders until their respective tables are created.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_tenant_limits(
  p_tenant_id        UUID,
  p_check_type       VARCHAR(50),
  p_additional_usage INT DEFAULT 1
) RETURNS BOOLEAN AS $$
DECLARE
  v_config        tenant_configurations%ROWTYPE;
  v_current_usage INT;
BEGIN
  -- Get tenant configuration
  SELECT * INTO v_config
  FROM   tenant_configurations
  WHERE  tenant_id = p_tenant_id;

  -- If no configuration found, deny operation
  IF v_config.tenant_id IS NULL THEN
    RAISE NOTICE 'No configuration found for tenant: %', p_tenant_id;
    RETURN FALSE;
  END IF;

  -- Check different types of limits
  CASE p_check_type
    WHEN 'users' THEN
      -- Placeholder: check user limit (users table does not exist yet)
      RETURN TRUE;

    WHEN 'entities' THEN
      -- Placeholder: check entity limit (entities table does not exist yet)
      RETURN TRUE;

    WHEN 'transactions' THEN
      -- Placeholder: check monthly transaction limit
      RETURN TRUE;

    WHEN 'storage' THEN
      -- Live storage check against tenant_usage_stats current period
      SELECT COALESCE(storage_used, 0) INTO v_current_usage
      FROM   tenant_usage_stats
      WHERE  tenant_id    = p_tenant_id
        AND  period_start <= CURRENT_DATE
        AND  period_end   >= CURRENT_DATE
      ORDER  BY period_start DESC
      LIMIT  1;

      RETURN (COALESCE(v_current_usage, 0) + p_additional_usage) <= v_config.storage_quota;

    ELSE
      RAISE NOTICE 'Unknown check type: %', p_check_type;
      RETURN FALSE;
  END CASE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) IS 'Validates tenant resource limits before operations - now includes storage limit checking';
