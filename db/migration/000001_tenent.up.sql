-- =====================================================
-- ERP/ACCOUNTING SYSTEM DATABASE MIGRATION
-- =====================================================
-- File: 001_create_erp_accounting_system.sql
-- Description: Complete database schema for multi-tenant ERP/Accounting system
-- Author: Generated Migration
-- Date: 2025-07-04
-- Version: 1.0.0
-- =====================================================

-- Start transaction to ensure atomic migration
BEGIN;

-- =====================================================
-- EXTENSIONS
-- =====================================================
-- Enable UUID generation for unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- ROLES AND PERMISSIONS
-- =====================================================
-- Create application role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'application_role') THEN
        CREATE ROLE application_role;
    END IF;
END
$$;

-- =====================================================
-- CORE TENANT MANAGEMENT
-- =====================================================

-- -----------------------------------------------------
-- TENANTS TABLE
-- -----------------------------------------------------
-- Primary table for multi-tenant SaaS architecture
-- Stores tenant information, business details, and configuration
CREATE TABLE tenants (
    -- Primary identifiers
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) UNIQUE NOT NULL,

    -- Contact and access information
    email VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE,

    -- Status and operational settings
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'pending')),
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',

    -- Flexible metadata storage
    metadata JSONB DEFAULT '{}',

    -- Business classification
    industry VARCHAR(50), -- For future industry-specific modules
    company_size VARCHAR(20)
        CHECK (company_size IN ('startup', 'small', 'medium', 'large', 'enterprise')),

    -- Compliance and legal information
    tax_id VARCHAR(50),
    registration_number VARCHAR(50),
    legal_entity_type VARCHAR(50),

    -- Tenant-specific settings
    settings JSONB NOT NULL DEFAULT '{}',

    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ -- Soft delete support
);

-- Create indexes for performance
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain) WHERE subdomain IS NOT NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add comments for documentation
COMMENT ON TABLE tenants IS 'Core tenant management table for multi-tenant SaaS architecture';
COMMENT ON COLUMN tenants.id IS 'Universal unique identifier for external API references';
COMMENT ON COLUMN tenants.slug IS 'URL-friendly tenant identifier';
COMMENT ON COLUMN tenants.metadata IS 'Flexible JSONB storage for additional tenant metadata';
COMMENT ON COLUMN tenants.settings IS 'Tenant-specific configuration settings';
COMMENT ON COLUMN tenants.deleted_at IS 'Soft delete timestamp - NULL means active';

-- -----------------------------------------------------
-- TENANT CONFIGURATIONS TABLE
-- -----------------------------------------------------
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

-- -----------------------------------------------------
-- TENANT USAGE STATISTICS TABLE
-- -----------------------------------------------------
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
    avg_response_time NUMERIC(10,2),
    error_rate NUMERIC(5,4),

    -- Financial metrics
    monthly_revenue NUMERIC(12,2),

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
-- UTILITY FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- TENANT CONTEXT MANAGEMENT
-- -----------------------------------------------------
-- Function to set tenant context for the current session
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    -- Validate tenant exists and is active
    IF NOT EXISTS (
        SELECT 1 FROM tenants
        WHERE id = tenant_id AND status = 'active' AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Invalid or inactive tenant: %', tenant_id;
    END IF;

    -- Set session variable for tenant context
    PERFORM set_config('app.current_tenant_id', tenant_id::text, true);

    -- Log tenant context change (optional)
    RAISE NOTICE 'Tenant context set to: %', tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION set_tenant_context(UUID) IS 'Sets the current tenant context for the session with validation';

-- -----------------------------------------------------
-- GET CURRENT TENANT FUNCTION
-- -----------------------------------------------------
-- Utility function to retrieve current tenant ID from session
CREATE OR REPLACE FUNCTION get_current_tenant_id()
RETURNS UUID AS $$
BEGIN
    -- Return current tenant ID from session variable, default to NULL if not set
    RETURN COALESCE(nullif(current_setting('app.current_tenant_id', true), ''), NULL)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        -- Return NULL if any error occurs (e.g., invalid cast)
        RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION get_current_tenant_id() IS 'Retrieves the current tenant ID from session context';

-- -----------------------------------------------------
-- TENANT LIMITS CHECKING FUNCTION
-- -----------------------------------------------------
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
            -- Check user limit (assuming users table exists)
            -- This is a placeholder for when users table is created
            RETURN TRUE;

        WHEN 'entities' THEN
            -- Check entity limit (assuming entities table exists)
            -- This is a placeholder for when entities table is created
            RETURN TRUE;

        WHEN 'transactions' THEN
            -- Check monthly transaction limit (assuming journal_entries table exists)
            -- This is a placeholder for when journal_entries table is created
            RETURN TRUE;

        WHEN 'storage' THEN
            -- Check storage limit
            SELECT COALESCE(storage_used, 0) INTO v_current_usage
            FROM tenant_usage_stats
            WHERE tenant_id = p_tenant_id
              AND period_start <= CURRENT_DATE
              AND period_end >= CURRENT_DATE
            ORDER BY period_start DESC
            LIMIT 1;

            RETURN (COALESCE(v_current_usage, 0) + p_additional_usage) <= v_config.storage_quota;

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

-- -----------------------------------------------------
-- ENABLE RLS ON TENANT TABLES
-- -----------------------------------------------------
-- Enable Row Level Security on tenants table
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant isolation
-- Only allow access to tenant data based on current session context
CREATE POLICY tenant_isolation_policy ON tenants
    FOR ALL TO application_role
    USING (id = get_current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_isolation_policy ON tenants IS 'Ensures tenant data isolation based on session context';

-- Enable RLS on tenant_configurations table
ALTER TABLE tenant_configurations ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant configurations isolation
CREATE POLICY tenant_configurations_isolation_policy ON tenant_configurations
    FOR ALL TO application_role
    USING (tenant_id = get_current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_configurations_isolation_policy ON tenant_configurations IS 'Ensures tenant configuration data isolation';

-- Enable RLS on tenant_usage_stats table
ALTER TABLE tenant_usage_stats ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant usage stats isolation
CREATE POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats
    FOR ALL TO application_role
    USING (tenant_id = get_current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats IS 'Ensures tenant usage statistics data isolation';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON tenants TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_configurations TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_usage_stats TO application_role;

-- Grant execute permissions on functions
GRANT EXECUTE ON FUNCTION set_tenant_context(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION get_current_tenant_id() TO application_role;
GRANT EXECUTE ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) TO application_role;

-- =====================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================

-- -----------------------------------------------------
-- UPDATED_AT TRIGGER FUNCTION
-- -----------------------------------------------------
-- Generic function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION update_updated_at_column() IS 'Generic trigger function to update updated_at timestamp';

-- -----------------------------------------------------
-- APPLY TRIGGERS TO TABLES
-- -----------------------------------------------------
-- Trigger for tenants table
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for tenant_configurations table
CREATE TRIGGER update_tenant_configurations_updated_at
    BEFORE UPDATE ON tenant_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- INITIAL DATA SETUP
-- =====================================================

-- -----------------------------------------------------
-- DEFAULT TENANT CONFIGURATION
-- -----------------------------------------------------
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

-- -----------------------------------------------------
-- TENANT CREATION TRIGGER
-- -----------------------------------------------------
-- Automatically create configuration when new tenant is created
CREATE OR REPLACE FUNCTION create_tenant_configuration_on_insert()
RETURNS TRIGGER AS $$
BEGIN
    -- Create default configuration for new tenant
    PERFORM create_default_tenant_configuration(NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER create_tenant_configuration_trigger
    AFTER INSERT ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION create_tenant_configuration_on_insert();

-- Add trigger comment
COMMENT ON TRIGGER create_tenant_configuration_trigger ON tenants IS 'Automatically creates default configuration for new tenants';

-- =====================================================
-- SAMPLE DATA (OPTIONAL)
-- =====================================================

-- Uncomment the following to insert sample data for testing
/*
-- Insert sample tenant
INSERT INTO tenants (
    slug, name, email, subdomain, industry, company_size, tax_id
) VALUES (
    'acme-corp',
    'ACME Corporation',
    'admin@acme-corp.com',
    'acme',
    'technology',
    'medium',
    'TAX123456789'
);
*/

-- =====================================================
-- MIGRATION COMPLETION
-- =====================================================

-- Commit the transaction
COMMIT;

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'ERP/Accounting System Database Migration Completed Successfully!';
    RAISE NOTICE 'Created tables: tenants, tenant_configurations, tenant_usage_stats';
    RAISE NOTICE 'Created functions: set_tenant_context, get_current_tenant_id, check_tenant_limits';
    RAISE NOTICE 'Enabled Row Level Security with tenant isolation policies';
    RAISE NOTICE 'Next steps: Create entities, accounts, users, and journal_entries tables';
END;
$$;
