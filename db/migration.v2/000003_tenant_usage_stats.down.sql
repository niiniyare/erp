-- =====================================================
-- TENANT USAGE STATS DOWN MIGRATION
-- =====================================================
-- Revert check_tenant_limits function to previous version
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
-- This will be fully functional after tenant_usage_stats migration
RETURN TRUE;

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

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_usage_stats_isolation_policy ON tenant_usage_stats;

-- Disable RLS
ALTER TABLE
  IF EXISTS tenant_usage_stats DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_tenant_usage_stats_tenant_period;

DROP INDEX IF EXISTS idx_tenant_usage_stats_period;

-- Drop table
DROP TABLE IF EXISTS tenant_usage_stats;
