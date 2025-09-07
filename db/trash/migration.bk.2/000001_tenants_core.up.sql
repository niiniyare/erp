-- =====================================================
-- EXTENSIONS
-- =====================================================
-- Enable UUID generation for unique identifiers
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable Row Level Security globally
SET row_security = on;

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

-- Create admin role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'admin_role') THEN
        CREATE ROLE admin_role;
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
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    slug VARCHAR(50)  NOT NULL,
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
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
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
COMMENT ON FUNCTION current_tenant_id() IS 'Retrieves the current tenant ID from session context';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================

-- -----------------------------------------------------
-- ENABLE RLS ON TENANT TABLE
-- -----------------------------------------------------
-- Enable Row Level Security on tenants table
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant isolation
-- Only allow access to tenant data based on current session context
CREATE POLICY tenant_isolation_policy ON tenants
    FOR ALL TO application_role
    USING (id = current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_isolation_policy ON tenants IS 'Ensures tenant data isolation based on session context';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON tenants TO application_role;

-- Grant execute permissions on functions
GRANT EXECUTE ON FUNCTION set_tenant_context(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION current_tenant_id() TO application_role;

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
-- APPLY TRIGGER TO TENANTS TABLE
-- -----------------------------------------------------
-- Trigger for tenants table
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- -----------------------------------------------------
-- SLUG GENERATION TRIGGER
-- -----------------------------------------------------
CREATE OR REPLACE FUNCTION generate_slug_from_name()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.slug IS NULL THEN
        NEW.slug := lower(regexp_replace(NEW.name, '[^a-zA-Z0-9]+', '-', 'g'));
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenant_slug_trigger
    BEFORE INSERT ON tenants FOR EACH ROW
    EXECUTE FUNCTION generate_slug_from_name();