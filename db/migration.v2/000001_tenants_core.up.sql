-- =====================================================
-- EXTENSIONS
-- =====================================================
-- Enable UUID generation for unique identifiers
-- Enable required extensions


-- CREATE SCHEMA IF NOT EXISTS ledger;

-- SET search_path TO ledger;

-- Enable Row Level Security globally
SET
  row_security = ON;

-- =====================================================
-- ROLES AND PERMISSIONS
-- =====================================================
-- Create application role if it doesn't exist
DO
$$
BEGIN
IF NOT EXISTS (
  SELECT
    1
  FROM
    pg_roles
  WHERE
    rolname = 'application_role'
) THEN CREATE ROLE application_role;

END IF;

END
$$
;

-- Create admin role if it doesn't exist
DO
$$
BEGIN
IF NOT EXISTS (
  SELECT
    1
  FROM
    pg_roles
  WHERE
    rolname = 'admin_role'
) THEN CREATE ROLE admin_role;

END IF;

END
$$
;

DO
$$
BEGIN
IF NOT EXISTS (
  SELECT
    1
  FROM
    pg_roles
  WHERE
    rolname = 'readonly_role'
) THEN CREATE ROLE readonly_role;

END IF;

END
$$
;

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
  slug VARCHAR(50) NOT NULL,
  name VARCHAR(255) UNIQUE NOT NULL,
  -- Contact and access information
  email VARCHAR(255) NOT NULL,
  subdomain VARCHAR(63) UNIQUE,
  -- Status and operational settings
  Status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (Status IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED')),
  timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
  currency_code CHAR(3) NOT NULL DEFAULT 'USD',
  -- Flexible metadata storage
  metadata JSONB DEFAULT '{}',
  -- Business classification
  industry VARCHAR(50),  -- For future industry-specific modules
  company_size VARCHAR(20) CHECK (
    company_size IN (
      'Startup',
      'Small',
      'Medium',
      'Large',
      'Enterprise'
    )
  ),
  -- Compliance and legal information
  tax_id VARCHAR(50),
  registration_number VARCHAR(50),
  legal_entity_type VARCHAR(50),
  -- Tenant-specific settings
  last_activity_at TIMESTAMPTZ DEFAULT NOW(),
  settings JSONB NOT NULL DEFAULT '{}',
  -- Audit timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ -- Soft delete support
);

-- Create indexes for performance
CREATE INDEX idx_tenants_slug ON tenants(slug);

CREATE INDEX idx_tenants_status ON tenants(Status);

CREATE INDEX idx_tenants_subdomain ON tenants(subdomain)
WHERE
  subdomain IS NOT NULL;

CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at)
WHERE
  deleted_at IS NOT NULL;

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
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID, user_role TEXT DEFAULT 'application_role') 
  RETURNS VOID AS $$
DECLARE
  tenant_status TEXT;
BEGIN
  -- Get tenant status in one query
  SELECT status INTO tenant_status 
  FROM tenants 
  WHERE id = tenant_id AND deleted_at IS NULL;
  
  IF NOT FOUND THEN
    RAISE EXCEPTION 'Tenant not found: %', tenant_id;
  END IF;
  
  IF tenant_status != 'ACTIVE' THEN
    RAISE EXCEPTION 'Tenant is not active: % (status: %)', tenant_id, tenant_status;
  END IF;
  
  -- Set multiple context variables
  PERFORM set_config('app.current_tenant_id', tenant_id::text, true);
  PERFORM set_config('app.tenant_status', tenant_status, true);
  PERFORM set_config('app.context_set_at', NOW()::text, true);
  
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION set_tenant_context(UUID,TEXT) IS 'Sets the current tenant context for the session with validation';

-- -----------------------------------------------------
-- GET CURRENT TENANT FUNCTION
-- -----------------------------------------------------
-- Utility function to retrieve current tenant ID from session
CREATE
OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS
$$
BEGIN
-- Return current tenant ID from session variable, default to NULL if not set
RETURN COALESCE(
  nullif(
    current_setting('app.current_tenant_id', FALSE),
    ''
  ),
  NULL
)::UUID;

EXCEPTION
WHEN OTHERS THEN
-- Return NULL if any error occurs (e.g., invalid cast)
RETURN NULL;

END;

$$
LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION current_tenant_id() IS 'Retrieves the current tenant ID from session context';

-- Function to clear tenant context (important for connection pooling)
CREATE OR REPLACE FUNCTION clear_tenant_context() 
  RETURNS VOID AS $$
BEGIN
  PERFORM set_config('app.current_tenant_id', NULL, true);
  PERFORM set_config('app.tenant_status', NULL, true);
  PERFORM set_config('app.context_set_at', NULL, true);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION clear_tenant_context() IS 'clear tenant context (important for connection pooling)';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================
-- -----------------------------------------------------
-- ENABLE RLS ON TENANT TABLE
-- -----------------------------------------------------
-- Enable Row Level Security on tenants table
ALTER TABLE
  tenants ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant isolation
-- Only allow access to tenant data based on current session context
-- FIXME: I am not sure if the tenants table can take this policy
CREATE POLICY tenant_isolation_policy ON tenants FOR ALL TO application_role USING (
  id = current_tenant_id()
  OR current_tenant_id() IS NULL
);

CREATE POLICY admin_full_access_policy ON tenants FOR ALL TO admin_role USING (true);

CREATE POLICY readonly_access_policy ON tenants FOR SELECT TO readonly_role USING (true);

--
-- Add policy comment
COMMENT ON POLICY tenant_isolation_policy ON tenants IS 'Ensures tenant data isolation based on session context';
COMMENT ON POLICY admin_full_access_policy ON tenants IS 'Allows admin_role full access to all tenant data';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================
-- Grant necessary permissions to application role
GRANT SELECT , INSERT ,UPDATE , DELETE ON tenants TO application_role;

-- Grant necessary permissions to admin_role
GRANT ALL PRIVILEGES ON tenants TO admin_role;


-- Grant necessary permissions t readonly_role
GRANT 
  SELECT
   ON tenants TO readonly_role;


-- Grant execute permissions on functions
GRANT EXECUTE ON FUNCTION set_tenant_context(UUID,TEXT) TO application_role;

GRANT EXECUTE ON FUNCTION current_tenant_id() TO application_role;
GRANT EXECUTE ON FUNCTION current_tenant_id() TO readonly_role;

-- =====================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================
-- -----------------------------------------------------
-- UPDATED_AT TRIGGER FUNCTION
-- -----------------------------------------------------
-- Generic function to update the updated_at timestamp
CREATE
OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS
$$
BEGIN
NEW.updated_at = NOW();

RETURN NEW;

END;

$$
LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION update_updated_at_column() IS 'Generic trigger function to update updated_at timestamp';

-- -----------------------------------------------------
-- APPLY TRIGGER TO TENANTS TABLE
-- -----------------------------------------------------
-- Trigger for tenants table
CREATE TRIGGER update_tenants_updated_at BEFORE
UPDATE
  ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- -----------------------------------------------------
-- SLUG GENERATION TRIGGER
-- -----------------------------------------------------
CREATE OR REPLACE FUNCTION generate_unique_slug_from_name() 
  RETURNS TRIGGER AS $$
DECLARE
  base_slug TEXT;
  final_slug TEXT;
  counter INTEGER := 1;
BEGIN
  IF NEW.slug IS NULL THEN
    -- Create base slug with better sanitization
    base_slug := lower(trim(both '-' from 
      regexp_replace(
        regexp_replace(NEW.name, '[^\w\s-]', '', 'g'),
        '\s+', '-', 'g'
      )
    ));
    
    -- Ensure slug is not empty
    IF base_slug = '' THEN
      base_slug := 'tenant';
    END IF;
    
    final_slug := base_slug;
    
    -- Handle slug collisions
    WHILE EXISTS(SELECT 1 FROM tenants WHERE slug = final_slug AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::UUID)) LOOP
      final_slug := base_slug || '-' || counter;
      counter := counter + 1;
    END LOOP;
    
    NEW.slug := final_slug;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenant_slug_trigger BEFORE
INSERT
  ON tenants FOR EACH ROW EXECUTE FUNCTION generate_unique_slug_from_name();


-- Better email validation
ALTER TABLE tenants ADD CONSTRAINT valid_email
  CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- Subdomain validation
ALTER TABLE tenants ADD CONSTRAINT valid_subdomain 
  CHECK (subdomain IS NULL OR subdomain ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$');

-- Slug validation
ALTER TABLE tenants ADD CONSTRAINT valid_slug 
  CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$');

-- Currency code validation (ISO 4217)
ALTER TABLE tenants ADD CONSTRAINT valid_currency 
  CHECK (currency_code ~* '^[A-Z]{3}$');
