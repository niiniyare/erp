-- =====================================================
-- TENANT CONFIGURATIONS TABLE
-- =====================================================
-- =====================================================
-- NOTE: FOR FUTURE posible changes
-- =====================================================
-- This table stores tenant-specific configurations, limits, and feature flags.
-- ⚠️ IMPORTANT:
-- During design we considered several approaches for tenant configurations:
-- 1. **Dedicated Columns (Current Approach)**
--    - Each configuration (limits, currency, fiscal year, etc.) is represented
--      as a dedicated column with constraints and defaults.
--    - ✅ Pros: Strong typing, easy querying, integrity enforced by Postgres.
--    - ❌ Cons: Schema migrations are required when adding/removing config options.
-- 2. **Key-Value Table**
--    - A normalized table: (tenant_id, key, value).
--    - Value could be TEXT or JSONB.
--    - ✅ Pros: Flexible, supports dynamic additions without schema changes.
--    - ❌ Cons: Weaker typing, harder to enforce constraints, requires parsing logic.
-- 3. **PostgreSQL Composite Type**
--    - Define a custom type, e.g. (key VARCHAR, value JSONB).
--    - Store an array of these in a single column (per tenant).
--    - ✅ Pros: Flexible structure, still uses Postgres typing.
--    - ❌ Cons: Less standard, more complex queries/updates, potential overuse of JSON.
-- 4. **Pure JSONB Column**
--    - Store all tenant configuration in a single JSONB column.
--    - ✅ Pros: Extremely flexible, supports arbitrary nesting.
--    - ❌ Cons: No relational constraints, application logic must enforce validity.
-- 🎯 DECISION:
-- We chose **Approach #1 (Dedicated Columns)** for core/critical settings
-- (limits, accounting preferences, localization, security policies, etc.)
-- because it enforces data integrity and allows direct SQL constraints.
-- However:
-- - A `settings JSONB` column is included for flexible, tenant-specific preferences.
-- - Future developers may extend or migrate towards #2 or #3 if flexibility
--   becomes more important than strong typing.
-- ✅ When modifying this table:
-- - Keep critical, high-value configs as dedicated columns.
-- - Use `settings` JSONB for experimental, low-risk, or per-tenant overrides.
-- - Always add `COMMENT ON COLUMN ...` for clarity.

-- =====================================================
-- Stores tenant-specific configuration, limits, and feature flags
-- =====================================================
CREATE TABLE tenant_configurations (
    -- Primary key and tenant reference
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE PRIMARY KEY,

    -- Resource limits
    max_users INT NOT NULL DEFAULT 100,
    max_entities INT NOT NULL DEFAULT 1000,
    max_transactions_per_month INT NOT NULL DEFAULT 10000,
    storage_quota BIGINT NOT NULL DEFAULT 1073741824, -- 1GB in bytes

    -- Accounting preferences
    accounting_method VARCHAR(10) NOT NULL DEFAULT 'ACCRUAL'
        CHECK (accounting_method IN ('ACCRUAL', 'CASH')),
    fiscal_year_start_month INT NOT NULL DEFAULT 1
        CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    default_currency CHAR(3) NOT NULL DEFAULT 'USD',

    -- Localization settings
    date_format VARCHAR(20) NOT NULL DEFAULT 'MM/DD/YYYY',
    number_format VARCHAR(20) NOT NULL DEFAULT 'US',
    language_code VARCHAR(5) NOT NULL DEFAULT 'en-US',

    -- Security settings
    password_policy JSONB NOT NULL DEFAULT '{
        "min_length": 8,
        "require_uppercase": true,
        "require_lowercase": true,
        "require_numbers": true,
        "require_symbols": false
    }'::jsonb,
    settings JSONB DEFAULT '{}'::jsonb,            -- tenant preferences and settings

    -- Integration settings
    webhook_endpoints JSONB DEFAULT '[]'::jsonb,
    api_rate_limits JSONB DEFAULT '{
        "requests_per_minute": 100,
        "requests_per_hour": 5000
    }'::jsonb,

    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add comments for documentation
COMMENT ON TABLE tenant_configurations IS 'Tenant-specific configuration settings, feature flags, and resource limits';
COMMENT ON COLUMN tenant_configurations.password_policy IS 'Password complexity requirements';
COMMENT ON COLUMN tenant_configurations.api_rate_limits IS 'API rate limiting configuration';

-- =====================================================
-- TENANT LIMITS CHECKING FUNCTION
-- =====================================================
-- Function to validate tenant resource limits before operations
CREATE OR REPLACE FUNCTION check_tenant_limits(
    p_tenant_id UUID,
    p_check_type VARCHAR(50),
    p_additional_usage INT DEFAULT 1
)
RETURNS BOOLEAN AS $$
DECLARE
    v_config tenant_configurations%ROWTYPE;
    v_current_usage INT;
BEGIN
    -- Get tenant configuration
    SELECT * INTO v_config
    FROM tenant_configurations
    WHERE tenant_id = p_tenant_id;

    -- If no configuration found, deny operation
    IF v_config.tenant_id IS NULL THEN
        RAISE NOTICE 'No configuration found for tenant: %', p_tenant_id;
        RETURN FALSE;
    END IF;

    -- Check different types of limits
    CASE p_check_type
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
            RAISE NOTICE 'Unknown check type: %', p_check_type;
            RETURN FALSE;
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) IS 'Validates tenant resource limits before operations';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================

-- Enable RLS on tenant_configurations table
ALTER TABLE tenant_configurations ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant configurations isolation
CREATE POLICY tenant_configurations_isolation_policy ON tenant_configurations
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_configurations_isolation_policy ON tenant_configurations IS 'Ensures tenant configuration data isolation';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_configurations TO application_role;

-- Grant execute permissions on function
GRANT EXECUTE ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) TO application_role;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Trigger for tenant_configurations table
CREATE TRIGGER update_tenant_configurations_updated_at
    BEFORE UPDATE ON tenant_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- DEFAULT TENANT CONFIGURATION SETUP
-- =====================================================

-- Function to create default configuration for new tenants
CREATE OR REPLACE FUNCTION create_default_tenant_configuration(p_tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    INSERT INTO tenant_configurations (tenant_id)
    VALUES (p_tenant_id)
    ON CONFLICT (tenant_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION create_default_tenant_configuration(UUID) IS 'Creates default configuration for a new tenant';

-- Automatically create configuration when new tenant is created
CREATE OR REPLACE FUNCTION create_tenant_configuration_on_insert()
RETURNS TRIGGER AS $$
BEGIN
    -- Create default configuration for new tenant
    PERFORM create_default_tenant_configuration(NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on tenants table (references previous migration)
CREATE TRIGGER create_tenant_configuration_trigger
    AFTER INSERT ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION create_tenant_configuration_on_insert();

-- Add trigger comment
COMMENT ON TRIGGER create_tenant_configuration_trigger ON tenants IS 'Automatically creates default configuration for new tenants';
