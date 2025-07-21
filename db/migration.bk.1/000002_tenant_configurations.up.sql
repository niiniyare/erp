-- =====================================================
-- TENANT CONFIGURATIONS TABLE
-- =====================================================
-- Stores tenant-specific configuration, limits, and feature flags

CREATE TABLE tenant_configurations (
    -- Primary key and tenant reference
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE PRIMARY KEY,

    -- Resource limits
    max_users INT NOT NULL DEFAULT 100,
    max_entities INT NOT NULL DEFAULT 1000,
    max_transactions_per_month INT NOT NULL DEFAULT 10000,
    storage_quota BIGINT NOT NULL DEFAULT 1073741824, -- 1GB in bytes

    -- Feature flags for module enablement
    features JSONB NOT NULL DEFAULT '{
        "advanced_reporting": false,
        "multi_currency": false,
        "project_tracking": true,
        "inventory_management": true,
        "payroll": false,
        "api_access": false
    }'::jsonb,

    -- Module configuration
    modules_enabled JSONB NOT NULL DEFAULT '["accounting", "inventory"]'::jsonb,

    -- Accounting preferences
    accounting_method VARCHAR(10) NOT NULL DEFAULT 'accrual'
        CHECK (accounting_method IN ('accrual', 'cash')),
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
COMMENT ON COLUMN tenant_configurations.features IS 'JSONB object containing feature flags for module enablement';
COMMENT ON COLUMN tenant_configurations.modules_enabled IS 'Array of enabled modules for the tenant';
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