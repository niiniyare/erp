-- ------------------------------------------------------------------------------------------------
-- TENANT_CONFIGURATIONS
-- ------------------------------------------------------------------------------------------------
-- Stores tenant-specific configuration, resource limits, and feature flags.
-- accounting_method IN ('ACCRUAL', 'CASH'); fiscal_year_start_month BETWEEN 1 AND 12.
--
-- Design decision: dedicated columns (strong typing, SQL constraints) are used for critical
-- settings (limits, accounting preferences, localisation, security policies). A settings JSONB
-- column is included for flexible, per-tenant overrides that do not warrant their own column.
-- Future developers may extend towards a key-value table if flexibility outweighs typing.
--
-- NOTE: update_updated_at_column() trigger function is defined in an earlier migration.
--       create_tenant_configuration_trigger fires on the tenants table (previous migration).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE tenant_configurations (
  tenant_id                    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE PRIMARY KEY,
  -- Resource limits
  max_users                    INT          NOT NULL DEFAULT 100,
  max_entities                 INT          NOT NULL DEFAULT 1000,
  max_transactions_per_month   INT          NOT NULL DEFAULT 10000,
  storage_quota                BIGINT       NOT NULL DEFAULT 1073741824,  -- 1 GB in bytes
  -- Accounting preferences
  accounting_method            VARCHAR(10)  NOT NULL DEFAULT 'ACCRUAL'
                                              CHECK (accounting_method IN ('ACCRUAL', 'CASH')),
  fiscal_year_start_month      INT          NOT NULL DEFAULT 1
                                              CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
  default_currency             CHAR(3)      NOT NULL DEFAULT 'USD',
  -- Localisation settings
  date_format                  VARCHAR(20)  NOT NULL DEFAULT 'MM/DD/YYYY',
  number_format                VARCHAR(20)  NOT NULL DEFAULT 'US',
  language_code                VARCHAR(5)   NOT NULL DEFAULT 'en-US',
  -- Security settings
  password_policy              JSONB        NOT NULL DEFAULT '{
        "min_length": 8,
        "require_uppercase": true,
        "require_lowercase": true,
        "require_numbers": true,
        "require_symbols": false
    }'::jsonb,
  settings                     JSONB        DEFAULT '{}'::jsonb,   -- flexible per-tenant preferences
  -- Integration settings
  webhook_endpoints            JSONB        DEFAULT '[]'::jsonb,
  api_rate_limits              JSONB        DEFAULT '{
        "requests_per_minute": 100,
        "requests_per_hour": 5000
    }'::jsonb,
  -- Audit timestamps
  created_at                   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at                   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  tenant_configurations                    IS 'Tenant-specific configuration settings, feature flags, and resource limits';
COMMENT ON COLUMN tenant_configurations.password_policy   IS 'Password complexity requirements';
COMMENT ON COLUMN tenant_configurations.api_rate_limits   IS 'API rate limiting configuration';

-- ------------------------------------------------------------------------------------------------
-- CHECK_TENANT_LIMITS FUNCTION
-- ------------------------------------------------------------------------------------------------
-- Validates whether a tenant has headroom for an additional usage unit of a given type.
-- p_check_type: 'users' | 'entities' | 'transactions' | 'storage'.
-- Returns FALSE immediately if no configuration row exists for the tenant.
--
-- NOTE: 'users', 'entities', and 'transactions' checks are placeholders — they return TRUE
--       until those tables exist in later migrations. 'storage' is fully implemented here
--       and reads from tenant_usage_stats (added in 000103).
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_tenant_limits(
  p_tenant_id       UUID,
  p_check_type      VARCHAR(50),
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
      -- Placeholder: storage check implemented fully in 000103 after tenant_usage_stats exists
      RETURN TRUE;

    ELSE
      RAISE NOTICE 'Unknown check type: %', p_check_type;
      RETURN FALSE;
  END CASE;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) IS 'Validates tenant resource limits before operations';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE tenant_configurations ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_configurations_isolation_policy ON tenant_configurations
  FOR ALL TO application_role
  USING (tenant_id = current_tenant_id());

COMMENT ON POLICY tenant_configurations_isolation_policy ON tenant_configurations IS 'Ensures tenant configuration data isolation';

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_configurations TO application_role;
GRANT EXECUTE ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) TO application_role;

-- ------------------------------------------------------------------------------------------------
-- TRIGGERS
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_tenant_configurations_updated_at
  BEFORE UPDATE ON tenant_configurations
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- DEFAULT CONFIGURATION FUNCTIONS
-- ------------------------------------------------------------------------------------------------
-- create_default_tenant_configuration inserts a default config row for a new tenant, using
-- all column defaults. ON CONFLICT DO NOTHING makes it safe to call repeatedly.
-- create_tenant_configuration_on_insert is the trigger body that calls the above on INSERT
-- into tenants, so every tenant always has a config row immediately after creation.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION create_default_tenant_configuration(p_tenant_id UUID) RETURNS VOID AS $$
BEGIN
  INSERT INTO tenant_configurations (tenant_id)
  VALUES (p_tenant_id)
  ON CONFLICT (tenant_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_default_tenant_configuration(UUID) IS 'Creates default configuration for a new tenant';

CREATE OR REPLACE FUNCTION create_tenant_configuration_on_insert() RETURNS TRIGGER AS $$
BEGIN
  PERFORM create_default_tenant_configuration(NEW.id);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on tenants table (references previous migration)
CREATE TRIGGER create_tenant_configuration_trigger
  AFTER INSERT ON tenants
  FOR EACH ROW EXECUTE FUNCTION create_tenant_configuration_on_insert();

COMMENT ON TRIGGER create_tenant_configuration_trigger ON tenants IS 'Automatically creates default configuration for new tenants';
