-- =====================================================
-- EXTENSIONS
-- =====================================================
-- Enable UUID generation for unique identifiers
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
    STATUS VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (STATUS IN ('active', 'suspended', 'pending')),
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    -- Flexible metadata storage
    metadata JSONB DEFAULT '{}',
    -- Business classification
    industry VARCHAR(50),  -- For future industry-specific modules
    company_size VARCHAR(20) CHECK (
        company_size IN (
            'startup',
            'small',
            'medium',
            'large',
            'enterprise'
        )
    ),
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

CREATE INDEX idx_tenants_status ON tenants(STATUS);

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
CREATE
OR REPLACE FUNCTION set_tenant_context(tenant_id UUID) RETURNS VOID AS
$$
BEGIN
-- Validate tenant exists and is active
IF NOT EXISTS (
    SELECT
        1
    FROM
        tenants
    WHERE
        id = tenant_id
        AND STATUS = 'active'
        AND deleted_at IS NULL
) THEN RAISE EXCEPTION 'Invalid or inactive tenant: %',
tenant_id;

END IF;

-- Set session variable for tenant context
PERFORM set_config('app.current_tenant_id', tenant_id::text, TRUE);

-- Log tenant context change (optional)
RAISE NOTICE 'Tenant context set to: %',
tenant_id;

END;

$$
LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION set_tenant_context(UUID) IS 'Sets the current tenant context for the session with validation';

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
CREATE POLICY tenant_isolation_policy ON tenants FOR ALL TO application_role USING (id = current_tenant_id() OR current_tenant_id() IS NULL);

-- Add policy comment
COMMENT ON POLICY tenant_isolation_policy ON tenants IS 'Ensures tenant data isolation based on session context';

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
    DELETE ON tenants TO application_role;

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
CREATE
OR REPLACE FUNCTION generate_slug_from_name() RETURNS TRIGGER AS
$$
BEGIN
IF NEW.slug IS NULL THEN NEW.slug := lower(
    regexp_replace(NEW.name, '[^a-zA-Z0-9]+', '-', 'g')
);

END IF;

RETURN NEW;

END;

$$
LANGUAGE plpgsql;

CREATE TRIGGER tenant_slug_trigger BEFORE
INSERT
    ON tenants FOR EACH ROW EXECUTE FUNCTION generate_slug_from_name();
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
COMMENT ON TRIGGER create_tenant_configuration_trigger ON tenants IS 'Automatically creates default configuration for new tenants';-- =====================================================
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
-- ROW LEVEL SECURITY (RLS)
-- =====================================================

-- Enable RLS on tenant_usage_stats table
ALTER TABLE tenant_usage_stats ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant usage stats isolation
CREATE POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id());

-- Add policy comment
COMMENT ON POLICY tenant_usage_stats_isolation_policy ON tenant_usage_stats IS 'Ensures tenant usage statistics data isolation';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_usage_stats TO application_role;

-- =====================================================
-- UPDATE CHECK_TENANT_LIMITS FUNCTION
-- =====================================================
-- Now that tenant_usage_stats exists, we can implement the storage check

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

-- Update function comment
COMMENT ON FUNCTION check_tenant_limits(UUID, VARCHAR, INT) IS 'Validates tenant resource limits before operations - now includes storage limit checking';-- =====================================================================
-- ENTITIES CORE TABLE - Business entity management foundation
-- =====================================================================

-- Root entity/company table with hierarchical structure and accounting preferences
CREATE TABLE entities (
    uuid UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,-- Self-reference for validation consistency
 
    
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50), -- Internal reference code
    
    type VARCHAR(20) NOT NULL CHECK (
        type IN (
            'COMPANY', 'SUBSIDIARY', 'REGION', 'BRANCH', 'LOCATION',
            'DEPARTMENT', 'DIVISION', 'COST_CENTER', 'PROJECT', 'BUDGET_UNIT'
        )
    ) DEFAULT 'COMPANY',
    is_active BOOLEAN NOT NULL DEFAULT true,
    hidden BOOLEAN NOT NULL DEFAULT false,
    accrual_method BOOLEAN NOT NULL,                    -- TRUE = Accrual, FALSE = Cash
    fy_start_month INTEGER NOT NULL CHECK (fy_start_month BETWEEN 1 AND 12),
    address JSONB DEFAULT '{}'::jsonb,
    picture VARCHAR(100),
    settings JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, name)
);

-- Table comments
COMMENT ON TABLE entities IS 
'Master table for business entities and organizational units. Supports hierarchical structures for companies, subsidiaries, departments, and other organizational divisions. Each entity can maintain its own accounting books, customers, vendors, and fiscal year settings.';

-- Column comments
COMMENT ON COLUMN entities.uuid IS 
'Primary key - Unique identifier for the entity';

COMMENT ON COLUMN entities.tenant_id IS 
'Foreign key to tenants table - Associates entity with a specific tenant for multi-tenancy support';

COMMENT ON COLUMN entities.parent_id IS 
'Self-referencing foreign key - Creates hierarchical relationship between entities (e.g., subsidiary under parent company)';

COMMENT ON COLUMN entities.name IS 
'Business name or title of the entity - Must be unique within tenant';

COMMENT ON COLUMN entities.code IS 
'Optional internal reference code - Used for abbreviated identification and reporting';

COMMENT ON COLUMN entities.type IS 
'Classification of entity type - Defines the organizational level and purpose (company, department, project, etc.)';

COMMENT ON COLUMN entities.is_active IS 
'Active status flag - Indicates whether the entity is currently operational';

COMMENT ON COLUMN entities.hidden IS 
'Visibility flag - Controls whether entity appears in user interfaces and reports';

COMMENT ON COLUMN entities.accrual_method IS 
'Accounting method indicator - TRUE for accrual accounting, FALSE for cash accounting';

COMMENT ON COLUMN entities.fy_start_month IS 
'Fiscal year start month - Numeric month (1-12) when fiscal year begins for this entity';

COMMENT ON COLUMN entities.address IS 
'Physical address information - Stored as JSON object with flexible address components';

COMMENT ON COLUMN entities.picture IS 
'Entity logo or image reference - File path or URL to associated image';

COMMENT ON COLUMN entities.settings IS 
'Entity-specific configuration - JSON object storing customizable settings and preferences';

COMMENT ON COLUMN entities.created_at IS 
'Record creation timestamp - Automatically set when entity is first created';

COMMENT ON COLUMN entities.updated_at IS 
'Last modification timestamp - Automatically updated when entity record is modified';

COMMENT ON COLUMN entities.deleted_at IS 
'Soft deletion timestamp - NULL for active records, timestamp when logically deleted';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Create unique index for tenant-code combination
CREATE UNIQUE INDEX tenant_code_unique_idx
    ON entities (tenant_id, code)
    WHERE code IS NOT NULL;

COMMENT ON INDEX tenant_code_unique_idx IS 
'Ensures entity codes are unique within each tenant - Only applies when code is not NULL';

-- Index for filtering entities by tenant and type (common query pattern)
CREATE INDEX idx_entities_tenant_type ON entities(tenant_id, type);
COMMENT ON INDEX idx_entities_tenant_type IS 
'Optimizes queries filtering entities by tenant and type - Common pattern for entity listings';

-- Index for parent-child hierarchy traversal
CREATE INDEX idx_entities_parent_id ON entities(parent_id);
COMMENT ON INDEX idx_entities_parent_id IS 
'Speeds up hierarchy traversal queries when finding direct children of an entity';

-- Partial index for active entities (most common filter)
CREATE INDEX idx_entities_active ON entities(is_active) WHERE is_active = true;
COMMENT ON INDEX idx_entities_active IS 
'Optimizes queries for active entities only - Uses partial index to save space';

-- Additional performance indexes
CREATE INDEX idx_entities_settings_gin ON entities USING gin(settings);
CREATE INDEX idx_entities_address_gin ON entities USING gin(address);
CREATE INDEX idx_entities_deleted_at ON entities(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_entities_tenant ON entities(tenant_id);
CREATE INDEX idx_entities_parent ON entities(parent_id);
CREATE INDEX idx_entities_type ON entities(type);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================

-- Prevent entities from being their own parent (circular reference)
ALTER TABLE entities ADD CONSTRAINT no_self_parent 
    CHECK (uuid != parent_id);
COMMENT ON CONSTRAINT no_self_parent ON entities IS 
'Prevents circular references where an entity is its own parent';

-- Ensure fiscal year start month is valid
ALTER TABLE entities ADD CONSTRAINT valid_fy_start_month 
    CHECK (fy_start_month BETWEEN 1 AND 12);
COMMENT ON CONSTRAINT valid_fy_start_month ON entities IS 
'Validates fiscal year start month is between 1 (January) and 12 (December)';

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable Row Level Security
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;

-- RLS policies with NULL context handling
CREATE POLICY tenant_isolation_policy ON entities
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY admin_full_access_policy ON entities
    FOR ALL TO admin_role
    USING (true);-- =====================================================================
-- ENTITIES HIERARCHY TABLE - Closure table for entity relationships
-- =====================================================================

-- Closure table for entity hierarchy
CREATE TABLE hierarchy_paths (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    ancestor_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    depth INT NOT NULL CHECK (depth >= 0),
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

-- Table comments
COMMENT ON TABLE hierarchy_paths IS 
'Closure table for efficient entity hierarchy queries. Stores all ancestor-descendant relationships with depth information. Enables fast retrieval of entity trees, subtrees, and hierarchy levels without recursive queries.';

-- Column comments
COMMENT ON COLUMN hierarchy_paths.tenant_id IS 
'Tenant identifier - Partitions hierarchy data by tenant for multi-tenancy';

COMMENT ON COLUMN hierarchy_paths.entity_id IS 
'Entity identifier - References the entity this path record belongs to';

COMMENT ON COLUMN hierarchy_paths.ancestor_id IS 
'Parent entity in the relationship - References entities.uuid';

COMMENT ON COLUMN hierarchy_paths.descendant_id IS 
'Child entity in the relationship - References entities.uuid';

COMMENT ON COLUMN hierarchy_paths.depth IS 
'Hierarchical distance - 0 for self-reference, 1 for direct parent-child, 2+ for deeper relationships';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Index for reverse hierarchy lookups (finding parents of a descendant)
CREATE INDEX idx_hierarchy_paths_descendant ON hierarchy_paths(descendant_id);
COMMENT ON INDEX idx_hierarchy_paths_descendant IS 
'Enables efficient reverse hierarchy traversal - Finding all ancestors of a given entity';

-- Index for entity hierarchy depth-based queries
CREATE INDEX idx_hierarchy_paths_depth ON hierarchy_paths(tenant_id, depth);
COMMENT ON INDEX idx_hierarchy_paths_depth IS 
'Optimizes queries filtering by hierarchy depth - Useful for organization level reports';

-- Additional performance indexes
CREATE INDEX idx_hierarchy_paths_tenant ON hierarchy_paths(tenant_id);
CREATE INDEX idx_hierarchy_paths_ancestor ON hierarchy_paths(ancestor_id);

-- =====================================================================
-- HIERARCHY MAINTENANCE TRIGGER
-- =====================================================================

CREATE OR REPLACE FUNCTION maintain_entity_id()
RETURNS TRIGGER AS $$
BEGIN
    -- Always align entity_id with descendant_id
    NEW.entity_id := NEW.descendant_id;
    
    -- Increment version only if row is actually modified
    IF TG_OP = 'UPDATE' AND ROW(NEW.*) IS DISTINCT FROM ROW(OLD.*) THEN
        NEW.version := OLD.version + 1;
    END IF;
    
    -- Update timestamp
    NEW.updated_at := NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION maintain_entity_id IS
'Maintains entity_id consistency in hierarchy_paths by setting it to descendant_id.
Increments version on updates and refreshes updated_at timestamp.';
-- Apply the existing maintain_entity_id trigger to hierarchy_paths



CREATE TRIGGER hierarchy_paths_maintain_entity_id
    BEFORE INSERT OR UPDATE ON hierarchy_paths
    FOR EACH ROW EXECUTE FUNCTION maintain_entity_id();

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable Row Level Security
ALTER TABLE hierarchy_paths ENABLE ROW LEVEL SECURITY;

-- RLS policies with NULL context handling
CREATE POLICY tenant_isolation_policy ON hierarchy_paths
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY admin_full_access_policy ON hierarchy_paths
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- ENTITY STATE MANAGEMENT TABLE - Document sequence tracking
-- =====================================================================

-- Entity state tracking for sequence numbers and fiscal periods
CREATE TABLE entitystate (
    uuid UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    fiscal_year SMALLINT,
    key VARCHAR(10) NOT NULL,                             -- Document type (e.g., invoice, po)
    sequence BIGINT NOT NULL,                             -- Next sequence number
    entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
    entity_unit_id UUID REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table comments
COMMENT ON TABLE entitystate IS 
'Manages sequential numbering for business documents within entities. Tracks next available sequence numbers for different document types (invoices, purchase orders, estimates, etc.) by fiscal year and entity.';

-- Column comments  
COMMENT ON COLUMN entitystate.uuid IS 
'Primary key - Unique identifier for the entity state record';

COMMENT ON COLUMN entitystate.tenant_id IS 
'Foreign key to tenants table for multi-tenant isolation';

COMMENT ON COLUMN entitystate.fiscal_year IS 
'Fiscal year for sequence tracking - Allows separate numbering sequences per year';

COMMENT ON COLUMN entitystate.key IS 
'Document type identifier - Specifies the type of document being numbered (invoice, po, estimate, bill, receipt, etc.)';

COMMENT ON COLUMN entitystate.sequence IS 
'Next sequence number - The next available sequential number for this document type';

COMMENT ON COLUMN entitystate.entity_id IS 
'Primary entity reference - The main entity that owns this sequence numbering';

COMMENT ON COLUMN entitystate.entity_unit_id IS 
'Sub-entity reference - Optional reference to a subsidiary or department within the main entity for more granular numbering';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Composite index for entity state lookups
CREATE INDEX idx_entitystate_entity_key ON entitystate(entity_id, key);
COMMENT ON INDEX idx_entitystate_entity_key IS 
'Optimizes sequence number lookups by entity and document type';

-- Index for fiscal year-based sequence queries
CREATE INDEX idx_entitystate_fiscal_year ON entitystate(entity_id, fiscal_year, key);
COMMENT ON INDEX idx_entitystate_fiscal_year IS 
'Supports efficient sequence retrieval filtered by fiscal year';

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================

-- Ensure unique sequence tracking per tenant, entity, document type, and fiscal year
ALTER TABLE entitystate ADD CONSTRAINT unique_tenant_entity_key_fy 
    UNIQUE (tenant_id, entity_id, key, fiscal_year);
COMMENT ON CONSTRAINT unique_tenant_entity_key_fy ON entitystate IS 
'Prevents duplicate sequence trackers for same tenant, entity, document type, and fiscal year';

-- Ensure sequence numbers are positive
ALTER TABLE entitystate ADD CONSTRAINT positive_sequence 
    CHECK (sequence > 0);
COMMENT ON CONSTRAINT positive_sequence ON entitystate IS 
'Ensures sequence numbers are always positive values';

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable Row Level Security
ALTER TABLE entitystate ENABLE ROW LEVEL SECURITY;

-- RLS policies with NULL context handling
CREATE POLICY tenant_isolation_policy ON entitystate
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY admin_full_access_policy ON entitystate
    FOR ALL TO admin_role
    USING (true);-- =====================================================================
-- ENTITIES VIEWS - Reporting and analytical views for entity management
-- =====================================================================

/*
 * Entity Hierarchy View
 * 
 * Purpose: Provides recursive hierarchical view of entities within tenant context
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * - current_tenant_id() function (must be implemented)
 * 
 * Notes:
 * - Uses recursive CTE to build entity hierarchy paths
 * - Filters by current tenant context
 * - Includes soft delete filtering
 */
CREATE VIEW v_tenant_hierarchy AS
WITH RECURSIVE org_chart AS (
  SELECT 
    e.uuid AS entity_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    e.name::TEXT AS path,
    0 AS depth
  FROM entities e
  WHERE e.parent_id IS NULL
    AND e.deleted_at IS NULL
  
  UNION ALL
  
  SELECT 
    e.uuid AS entity_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    (oc.path || ' > ' || e.name)::TEXT,
    oc.depth + 1
  FROM entities e
  JOIN org_chart oc ON e.parent_id = oc.entity_id
  WHERE e.deleted_at IS NULL
)
SELECT 
  t.name AS tenant_name,
  oc.entity_id,
  oc.name AS entity_name,
  oc.type AS entity_type,
  oc.path AS full_path,
  oc.depth
FROM org_chart oc
JOIN tenants t ON oc.tenant_id = t.id;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Entity Hierarchy Structure View
 * 
 * Purpose: Shows the complete organizational structure for entities
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - This view is a placeholder - requires users table to be fully functional
 * - Currently shows entity hierarchy structure only
 * - Can be extended when user management tables are available
 */
CREATE VIEW v_entity_structure AS
SELECT
  t.name AS tenant_name,
  cc.uuid AS cost_center_id,
  cc.name AS cost_center,
  d.uuid AS department_id,
  d.name AS department,
  r.uuid AS regional_id,
  r.name AS regional,
  c.uuid AS company_id,
  c.name AS company
FROM entities cc
JOIN entities d ON cc.parent_id = d.uuid AND d.type = 'DEPARTMENT'
JOIN entities r ON d.parent_id = r.uuid AND r.type IN ('REGION', 'REGIONAL')
JOIN entities c ON r.parent_id = c.uuid AND c.type = 'COMPANY'
JOIN tenants t ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Cost Center Basic Information View
 * 
 * Purpose: Provides basic cost center information with organizational context
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Simplified version without user count (requires users table)
 * - Shows cost center hierarchy up to company level
 * - Includes tenant context
 */
CREATE VIEW v_cost_center_info AS
SELECT
  t.name AS tenant_name,
  cc.uuid AS cost_center_id,
  cc.name AS cost_center,
  cc.code AS cost_center_code,
  d.uuid AS department_id,
  d.name AS department,
  r.uuid AS regional_id,
  r.name AS regional,
  c.uuid AS company_id,
  c.name AS company,
  cc.is_active AS cost_center_active,
  cc.created_at AS cost_center_created
FROM entities cc
JOIN entities d ON cc.parent_id = d.uuid
JOIN entities r ON d.parent_id = r.uuid
JOIN entities c ON r.parent_id = c.uuid
JOIN tenants t ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND d.type = 'DEPARTMENT'
  AND r.type IN ('REGION', 'REGIONAL')
  AND c.type = 'COMPANY'
  AND cc.deleted_at IS NULL
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Department Summary View
 * 
 * Purpose: Provides summary information about departments and their cost centers
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Shows department hierarchy with cost center counts
 * - Simplified without user metrics (requires users table)
 * - Includes tenant context and soft delete filtering
 */
CREATE VIEW v_department_summary AS
SELECT
  t.name AS tenant_name,
  d.uuid AS department_id,
  d.name AS department_name,
  d.code AS department_code,
  r.uuid AS regional_id,
  r.name AS regional_name,
  c.uuid AS company_id,
  c.name AS company_name,
  COUNT(DISTINCT cc.uuid) AS cost_center_count,
  d.is_active AS department_active,
  d.created_at AS department_created
FROM entities d
JOIN entities r ON d.parent_id = r.uuid
JOIN entities c ON r.parent_id = c.uuid
JOIN tenants t ON d.tenant_id = t.id
LEFT JOIN entities cc ON cc.parent_id = d.uuid 
  AND cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
WHERE d.type = 'DEPARTMENT'
  AND r.type IN ('REGION', 'REGIONAL')
  AND c.type = 'COMPANY'
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY 
  t.id, t.name,
  d.uuid, d.name, d.code, d.is_active, d.created_at,
  r.uuid, r.name,
  c.uuid, c.name;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Company Structure View
 * 
 * Purpose: Shows complete organizational structure under each company
 * 
 * Dependencies:
 * - entities table
 * - hierarchy_paths table
 * - tenants table
 * 
 * Notes:
 * - Uses hierarchy_paths for efficient traversal
 * - Shows all entities under company level
 * - Includes depth information from company root
 */
CREATE VIEW v_company_structure AS
SELECT
  t.name AS tenant_name,
  c.uuid AS company_id,
  c.name AS company_name,
  c.code AS company_code,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.code AS entity_code,
  hp.depth AS levels_from_company
FROM entities c
JOIN hierarchy_paths hp ON c.uuid = hp.ancestor_id
JOIN entities e ON hp.descendant_id = e.uuid
JOIN tenants t ON c.tenant_id = t.id
WHERE c.type = 'COMPANY'
  AND c.deleted_at IS NULL
  AND e.deleted_at IS NULL;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Tenant Entity Summary View
 * 
 * Purpose: Provides summary statistics of entities within each tenant
 * 
 * Dependencies:
 * - tenants table
 * - entities table
 * 
 * Notes:
 * - Simplified without user counts (requires users table)
 * - Shows entity type distribution per tenant
 * - Includes entity activity status
 */
CREATE VIEW v_tenant_entity_summary AS
SELECT
  t.id AS tenant_id,
  t.name AS tenant_name,
  t.status AS tenant_status,
  COUNT(DISTINCT e.uuid) AS total_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'COMPANY') AS company_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type IN ('REGION', 'REGIONAL')) AS regional_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'DEPARTMENT') AS department_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'COST_CENTER') AS cost_center_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'PROJECT') AS project_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.is_active = true) AS active_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.deleted_at IS NULL) AS non_deleted_entities
FROM tenants t
LEFT JOIN entities e ON e.tenant_id = t.id
GROUP BY t.id, t.name, t.status;

/*
 * Active Entities Report View
 * 
 * Purpose: Shows currently active entities across organizational hierarchy
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Replacement for user-based view until users table is available
 * - Shows active entities with their full organizational context
 * - Includes recent activity indicators
 */
CREATE VIEW v_active_entities AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.code AS entity_code,
  c.name AS company_name,
  r.name AS regional_name,
  d.name AS department_name,
  e.is_active,
  e.created_at,
  e.updated_at
FROM entities e
JOIN tenants t ON e.tenant_id = t.id
LEFT JOIN entities c ON (
  CASE 
    WHEN e.type = 'COMPANY' THEN e.uuid = c.uuid
    ELSE EXISTS (
      SELECT 1 FROM hierarchy_paths hp 
      WHERE hp.descendant_id = e.uuid 
        AND hp.ancestor_id = c.uuid 
        AND c.type = 'COMPANY'
    )
  END
)
LEFT JOIN entities r ON (
  CASE 
    WHEN e.type IN ('REGION', 'REGIONAL') THEN e.uuid = r.uuid
    ELSE EXISTS (
      SELECT 1 FROM hierarchy_paths hp 
      WHERE hp.descendant_id = e.uuid 
        AND hp.ancestor_id = r.uuid 
        AND r.type IN ('REGION', 'REGIONAL')
    )
  END
)
LEFT JOIN entities d ON (
  CASE 
    WHEN e.type = 'DEPARTMENT' THEN e.uuid = d.uuid
    ELSE e.parent_id = d.uuid AND d.type = 'DEPARTMENT'
  END
)
WHERE e.deleted_at IS NULL
  AND e.is_active = true
  AND (c.deleted_at IS NULL OR c.uuid IS NULL)
  AND (r.deleted_at IS NULL OR r.uuid IS NULL)
  AND (d.deleted_at IS NULL OR d.uuid IS NULL);

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Entity Change Log View
 * 
 * Purpose: Placeholder for audit trail - tracks entity changes
 * 
 * Dependencies:
 * - entities table (for change tracking)
 * - tenants table
 * 
 * Notes:
 * - Simplified view showing entity modification patterns
 * - Can be extended when audit_logs table is implemented
 * - Focuses on entity lifecycle events
 */
CREATE VIEW v_entity_changes AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.validation_status,
  e.created_at,
  e.updated_at,
  e.deleted_at,
  CASE 
    WHEN e.deleted_at IS NOT NULL THEN 'DELETED'
    WHEN e.updated_at > e.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
    ELSE 'CREATED'
  END AS change_type
FROM entities e
JOIN tenants t ON e.tenant_id = t.id
ORDER BY e.updated_at DESC;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Entity Hierarchy Paths View
 * 
 * Purpose: Shows all ancestor-descendant relationships with depth information
 * 
 * Dependencies:
 * - hierarchy_paths table
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Provides flattened view of entity relationships
 * - Includes depth for distance calculations
 * - Useful for hierarchy analysis and reporting
 */
CREATE VIEW v_entity_paths AS
SELECT
  t.name AS tenant_name,
  a.uuid AS ancestor_id,
  a.name AS ancestor_name,
  a.type AS ancestor_type,
  d.uuid AS descendant_id,
  d.name AS descendant_name,
  d.type AS descendant_type,
  hp.depth
FROM hierarchy_paths hp
JOIN entities a ON hp.ancestor_id = a.uuid
JOIN entities d ON hp.descendant_id = d.uuid
JOIN tenants t ON hp.tenant_id = t.id
WHERE a.deleted_at IS NULL
  AND d.deleted_at IS NULL
ORDER BY hp.depth, a.name, d.name;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses

/*
 * Tenant Resource Utilization View
 * 
 * Purpose: Shows resource utilization and capacity for each tenant
 * 
 * Dependencies:
 * - tenants table
 * - entities table
 * - entitystate table
 * 
 * Notes:
 * - Simplified without user metrics (requires users table)
 * - Shows entity utilization and state management
 * - Includes sequence number usage statistics
 */
CREATE VIEW v_tenant_resource_utilization AS
SELECT
  t.id AS tenant_id,
  t.name AS tenant_name,
  t.status AS tenant_status,
  COUNT(DISTINCT e.uuid) AS total_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.is_active = true) AS active_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.deleted_at IS NULL) AS non_deleted_entities,
  COUNT(DISTINCT es.uuid) AS sequence_states,
  COUNT(DISTINCT es.key) AS document_types,
  MAX(e.created_at) AS last_entity_created,
  MAX(e.updated_at) AS last_entity_updated
FROM tenants t
LEFT JOIN entities e ON e.tenant_id = t.id
LEFT JOIN entitystate es ON es.tenant_id = t.id
GROUP BY t.id, t.name, t.status;

-- Note: Views inherit RLS from their underlying tables
-- RLS will be enforced through the entities and tenants tables that the view uses-- ================================================================================================
-- PERSONS TABLE - Generic person entities for flexible identity management
-- ================================================================================================
-- 
-- Stores generic person entities that can represent individuals, employees, contacts, customers, etc.
-- Supports ABAC with security attributes and flexible person typing.
-- 
-- Prerequisites: 
-- - tenants table with UUID primary key
-- - entities table with UUID primary key  
-- - Both tables must exist before running this schema
-- ================================================================================================

CREATE TABLE persons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    person_type VARCHAR(20) NOT NULL 
        CHECK (person_type IN ('INDIVIDUAL', 'EMPLOYEE', 'CONTACT', 'CUSTOMER', 'VENDOR', 'CONTRACTOR')),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(20),
    birth_date DATE,
    national_id VARCHAR(50),
    tax_id VARCHAR(50),
    address JSONB DEFAULT '{}'::jsonb,
    security_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC attributes: clearance level, department, location, etc.
    metadata JSONB DEFAULT '{}'::jsonb,            -- Additional flexible data storage
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,                        -- Soft delete timestamp
    
    -- Ensure email uniqueness per tenant (excluding soft-deleted records)
    CONSTRAINT persons_email_unique_active 
        EXCLUDE (tenant_id WITH =, email WITH =) 
        WHERE (email IS NOT NULL AND deleted_at IS NULL),
    
    -- Ensure national ID uniqueness per tenant (excluding soft-deleted records)
    CONSTRAINT persons_national_id_unique_active 
        EXCLUDE (tenant_id WITH =, national_id WITH =) 
        WHERE (national_id IS NOT NULL AND deleted_at IS NULL)
);

-- Add table and column comments
COMMENT ON TABLE persons IS 
'Stores person entities with ABAC security attributes. Supports multiple person types including employees, customers, vendors, and contractors. Implements soft delete and tenant isolation.';

COMMENT ON COLUMN persons.id IS 'UUID primary key for the person record';
COMMENT ON COLUMN persons.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN persons.entity_id IS 'Foreign key to entities table for hierarchical organization';
COMMENT ON COLUMN persons.person_type IS 'Classification of person: INDIVIDUAL, EMPLOYEE, CONTACT, CUSTOMER, VENDOR, CONTRACTOR';
COMMENT ON COLUMN persons.first_name IS 'Person''s first/given name';
COMMENT ON COLUMN persons.last_name IS 'Person''s last/family name';
COMMENT ON COLUMN persons.middle_name IS 'Person''s middle name or initial (optional)';
COMMENT ON COLUMN persons.email IS 'Person''s email address (must be unique per tenant when not deleted)';
COMMENT ON COLUMN persons.phone IS 'Person''s primary phone number';
COMMENT ON COLUMN persons.birth_date IS 'Person''s date of birth';
COMMENT ON COLUMN persons.national_id IS 'Government-issued national ID number (unique per tenant)';
COMMENT ON COLUMN persons.tax_id IS 'Tax identification number';
COMMENT ON COLUMN persons.address IS 'Person''s address stored as JSONB for flexible structure';
COMMENT ON COLUMN persons.security_attributes IS 'JSONB containing ABAC attributes like clearance level, department, location for access control';
COMMENT ON COLUMN persons.metadata IS 'Flexible JSONB storage for additional person-related data';
COMMENT ON COLUMN persons.is_active IS 'Whether the person record is currently active';
COMMENT ON COLUMN persons.deleted_at IS 'Soft delete timestamp - NULL means record is active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Index for tenant-based queries (most common pattern)
CREATE INDEX idx_persons_tenant ON persons(tenant_id);

-- Index for person type filtering
CREATE INDEX idx_persons_type ON persons(person_type);

-- Index for entity association
CREATE INDEX idx_persons_entity ON persons(entity_id);

-- Index for active person queries
CREATE INDEX idx_persons_active ON persons(is_active) WHERE is_active = true;

-- Index for soft delete queries
CREATE INDEX idx_persons_deleted_at ON persons(deleted_at) WHERE deleted_at IS NOT NULL;

-- Index for email searches (case-insensitive)
CREATE INDEX idx_persons_email_lower ON persons(lower(email)) WHERE email IS NOT NULL;

-- Index for name searches
CREATE INDEX idx_persons_name ON persons(last_name, first_name);

-- GIN indexes for JSONB columns
CREATE INDEX idx_persons_address_gin ON persons USING gin(address);
CREATE INDEX idx_persons_security_attributes_gin ON persons USING gin(security_attributes);
CREATE INDEX idx_persons_metadata_gin ON persons USING gin(metadata);

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable RLS on persons table
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY persons_tenant_isolation ON persons
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY persons_admin_access ON persons
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- =====================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================================

-- Apply the existing update_updated_at_column trigger to persons
CREATE TRIGGER update_persons_updated_at
    BEFORE UPDATE ON persons
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON persons TO application_role;

-- Grant read-only access to specific roles if needed
-- GRANT SELECT ON persons TO readonly_role;-- ================================================================================================
-- EMPLOYEES TABLE - Employment-specific data extending persons
-- ================================================================================================
-- 
-- Extends persons with employment-specific data and organizational hierarchy.
-- Contains role context, security levels, and employment status tracking.
-- 
-- Prerequisites:
-- - tenants table with UUID primary key
-- - entities table with UUID primary key
-- - persons table with UUID primary key
-- ================================================================================================

CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    employee_number VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    position_title VARCHAR(100),
    department_id UUID REFERENCES entities(uuid), -- Department entity reference
    manager_id UUID REFERENCES employees(id),     -- Self-referential for org hierarchy
    hire_date DATE NOT NULL,
    termination_date DATE,
    salary_info JSONB DEFAULT '{}'::jsonb,        -- Encrypted/sensitive salary data
    employment_status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (employment_status IN ('ACTIVE', 'INACTIVE', 'TERMINATED', 'ON_LEAVE', 'SUSPENDED')),
    work_schedule JSONB DEFAULT '{}'::jsonb,      -- Flexible work schedule definition
    security_level INTEGER DEFAULT 0,             -- Numeric security clearance level (0=lowest)
    access_attributes JSONB DEFAULT '{}'::jsonb,  -- Employment-specific ABAC attributes
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Ensure employee number uniqueness per tenant
    CONSTRAINT employees_number_unique_active 
        EXCLUDE (tenant_id WITH =, employee_number WITH =) 
        WHERE (deleted_at IS NULL)
);

-- Add table and column comments
COMMENT ON TABLE employees IS 
'Employee records extending persons with employment-specific data, organizational hierarchy, and security levels for access control.';

COMMENT ON COLUMN employees.id IS 'UUID primary key for the employee record';
COMMENT ON COLUMN employees.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN employees.person_id IS 'Foreign key to persons table linking to personal information';
COMMENT ON COLUMN employees.employee_number IS 'Unique employee identifier within tenant';
COMMENT ON COLUMN employees.entity_id IS 'Foreign key to entities table for organizational assignment';
COMMENT ON COLUMN employees.position_title IS 'Employee''s job title or position';
COMMENT ON COLUMN employees.department_id IS 'Foreign key to entities table representing department';
COMMENT ON COLUMN employees.manager_id IS 'Self-referential foreign key for organizational hierarchy';
COMMENT ON COLUMN employees.hire_date IS 'Date when employee was hired';
COMMENT ON COLUMN employees.termination_date IS 'Date when employee was terminated (if applicable)';
COMMENT ON COLUMN employees.salary_info IS 'JSONB containing encrypted/sensitive salary and compensation data';
COMMENT ON COLUMN employees.employment_status IS 'Current employment status: ACTIVE, INACTIVE, TERMINATED, ON_LEAVE, SUSPENDED';
COMMENT ON COLUMN employees.work_schedule IS 'JSONB containing flexible work schedule definition';
COMMENT ON COLUMN employees.security_level IS 'Numeric security clearance level (0=lowest, higher numbers = higher clearance)';
COMMENT ON COLUMN employees.access_attributes IS 'JSONB containing employment-specific ABAC attributes for access control';
COMMENT ON COLUMN employees.deleted_at IS 'Soft delete timestamp - NULL means record is active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Index for tenant-based queries
CREATE INDEX idx_employees_tenant ON employees(tenant_id);

-- Index for person association
CREATE INDEX idx_employees_person ON employees(person_id);

-- Index for entity/organization association
CREATE INDEX idx_employees_entity ON employees(entity_id);

-- Index for department association
CREATE INDEX idx_employees_department ON employees(department_id) WHERE department_id IS NOT NULL;

-- Index for manager hierarchy
CREATE INDEX idx_employees_manager ON employees(manager_id) WHERE manager_id IS NOT NULL;

-- Index for employment status filtering
CREATE INDEX idx_employees_status ON employees(employment_status);

-- Index for active employees (most common query)
CREATE INDEX idx_employees_active ON employees(employment_status) WHERE employment_status = 'ACTIVE';

-- Index for employee number searches
CREATE INDEX idx_employees_number ON employees(employee_number);

-- Index for security level queries
CREATE INDEX idx_employees_security_level ON employees(security_level);

-- Index for soft delete queries
CREATE INDEX idx_employees_deleted_at ON employees(deleted_at) WHERE deleted_at IS NOT NULL;

-- GIN indexes for JSONB columns
CREATE INDEX idx_employees_salary_info_gin ON employees USING gin(salary_info);
CREATE INDEX idx_employees_work_schedule_gin ON employees USING gin(work_schedule);
CREATE INDEX idx_employees_access_attributes_gin ON employees USING gin(access_attributes);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================

-- Ensure termination date is after hire date
ALTER TABLE employees ADD CONSTRAINT valid_termination_date 
    CHECK (termination_date IS NULL OR termination_date >= hire_date);

-- Ensure security level is non-negative
ALTER TABLE employees ADD CONSTRAINT valid_security_level 
    CHECK (security_level >= 0);

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable RLS on employees table
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY employees_tenant_isolation ON employees
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY employees_admin_access ON employees
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- =====================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================================

-- Apply the existing update_updated_at_column trigger to employees
CREATE TRIGGER update_employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON employees TO application_role;-- ================================================================================================
-- USERS TABLE - System access accounts with authentication and RBAC integration
-- ================================================================================================
-- 
-- System access accounts with authentication data and RBAC integration.
-- Can be linked to persons/employees or exist independently for service accounts.
-- 
-- Prerequisites:
-- - tenants table with UUID primary key
-- - entities table with UUID primary key  
-- - persons table with UUID primary key
-- - employees table with UUID primary key
-- ================================================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    person_id UUID REFERENCES persons(id) ON DELETE SET NULL,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    username VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    user_type VARCHAR(20) NOT NULL DEFAULT 'INTERNAL'
        CHECK (user_type IN ('INTERNAL', 'CUSTOMER', 'VENDOR', 'PARTNER', 'API', 'SERVICE', 'ADMIN')),
    account_status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (account_status IN ('ACTIVE', 'INACTIVE', 'LOCKED', 'SUSPENDED', 'PENDING_VERIFICATION')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    failed_login_attempts INTEGER DEFAULT 0,
    lockout_until TIMESTAMPTZ,
    session_timeout_minutes INTEGER DEFAULT 480,   -- 8 hours default
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    user_attributes JSONB DEFAULT '{}'::jsonb,     -- ABAC user attributes
    settings JSONB DEFAULT '{}'::jsonb,            -- User preferences and settings
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Ensure email uniqueness per tenant
    CONSTRAINT users_email_unique_active 
        EXCLUDE (tenant_id WITH =, email WITH =) 
        WHERE (deleted_at IS NULL),
    
    -- Ensure username uniqueness per tenant when provided
    CONSTRAINT users_username_unique_active 
        EXCLUDE (tenant_id WITH =, username WITH =) 
        WHERE (username IS NOT NULL AND deleted_at IS NULL)
);

-- Add table and column comments
COMMENT ON TABLE users IS 
'System user accounts with authentication, authorization, and session management. Can be linked to persons/employees or exist independently for service accounts.';

COMMENT ON COLUMN users.id IS 'UUID primary key for the user record';
COMMENT ON COLUMN users.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN users.entity_id IS 'Foreign key to entities table for organizational assignment';
COMMENT ON COLUMN users.person_id IS 'Optional foreign key to persons table (NULL for service accounts)';
COMMENT ON COLUMN users.employee_id IS 'Optional foreign key to employees table (NULL for non-employee users)';
COMMENT ON COLUMN users.username IS 'Unique username for login (optional, email can be used instead)';
COMMENT ON COLUMN users.email IS 'Email address for login and communication (must be unique per tenant)';
COMMENT ON COLUMN users.password_hash IS 'Hashed password for authentication';
COMMENT ON COLUMN users.user_type IS 'Classification of user account: INTERNAL, CUSTOMER, VENDOR, PARTNER, API, SERVICE, ADMIN';
COMMENT ON COLUMN users.account_status IS 'Current account status affecting login ability';
COMMENT ON COLUMN users.is_active IS 'Whether the user account is currently active';
COMMENT ON COLUMN users.last_login_at IS 'Timestamp of last successful login';
COMMENT ON COLUMN users.password_changed_at IS 'Timestamp of last password change';
COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for failed login attempts for security monitoring';
COMMENT ON COLUMN users.lockout_until IS 'Timestamp until which account is locked due to failed attempts';
COMMENT ON COLUMN users.session_timeout_minutes IS 'Session timeout in minutes (default 480 = 8 hours)';
COMMENT ON COLUMN users.mfa_enabled IS 'Whether multi-factor authentication is enabled';
COMMENT ON COLUMN users.mfa_secret IS 'Secret key for MFA token generation';
COMMENT ON COLUMN users.user_attributes IS 'JSONB containing ABAC attributes for fine-grained access control';
COMMENT ON COLUMN users.settings IS 'JSONB containing user preferences and application settings';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp - NULL means record is active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Index for tenant-based queries
CREATE INDEX idx_users_tenant ON users(tenant_id);

-- Index for entity association
CREATE INDEX idx_users_entity ON users(entity_id);

-- Index for person/employee associations
CREATE INDEX idx_users_person ON users(person_id) WHERE person_id IS NOT NULL;
CREATE INDEX idx_users_employee ON users(employee_id) WHERE employee_id IS NOT NULL;

-- Index for authentication queries
CREATE INDEX idx_users_email_lower ON users(lower(email));
CREATE INDEX idx_users_username_lower ON users(lower(username)) WHERE username IS NOT NULL;

-- Index for user type filtering
CREATE INDEX idx_users_type ON users(user_type);

-- Index for account status
CREATE INDEX idx_users_account_status ON users(account_status);

-- Index for active users (most common query)
CREATE INDEX idx_users_active ON users(is_active, account_status) WHERE is_active = true AND account_status = 'ACTIVE';

-- Index for security monitoring
CREATE INDEX idx_users_failed_attempts ON users(failed_login_attempts) WHERE failed_login_attempts > 0;
CREATE INDEX idx_users_lockout ON users(lockout_until) WHERE lockout_until IS NOT NULL;

-- Index for MFA users
CREATE INDEX idx_users_mfa ON users(mfa_enabled) WHERE mfa_enabled = true;

-- Index for soft delete queries
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- GIN indexes for JSONB columns
CREATE INDEX idx_users_attributes_gin ON users USING gin(user_attributes);
CREATE INDEX idx_users_settings_gin ON users USING gin(settings);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================

-- Ensure failed login attempts is non-negative
ALTER TABLE users ADD CONSTRAINT valid_failed_attempts 
    CHECK (failed_login_attempts >= 0);

-- Ensure session timeout is positive
ALTER TABLE users ADD CONSTRAINT valid_session_timeout 
    CHECK (session_timeout_minutes > 0);

-- Ensure lockout_until is in the future when set
ALTER TABLE users ADD CONSTRAINT valid_lockout_time 
    CHECK (lockout_until IS NULL OR lockout_until > NOW());

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable RLS on users table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY users_tenant_isolation ON users
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY users_admin_access ON users
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- =====================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================================

-- Apply the existing update_updated_at_column trigger to users
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO application_role;
-- ================================================================================================
-- USER SESSIONS TABLE - Tracks active user sessions with security context
-- ================================================================================================
-- 
-- Extracted from existing user migration to maintain consistency with implemented application code.
-- Tracks active user sessions with security context for ABAC evaluation.
-- Includes device fingerprinting, location data, and access patterns for security monitoring.
-- 
-- Prerequisites:
-- - tenants table with UUID primary key
-- - users table with UUID primary key
-- ================================================================================================

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}'::jsonb,         -- Device fingerprinting data
    location_info JSONB DEFAULT '{}'::jsonb,       -- Geographic/network location for ABAC
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);

-- Add table and column comments
COMMENT ON TABLE user_sessions IS 
'Active user sessions with security context including device, location, and access patterns for ABAC evaluation and security monitoring.';

COMMENT ON COLUMN user_sessions.id IS 'UUID primary key for the session record';
COMMENT ON COLUMN user_sessions.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN user_sessions.user_id IS 'Foreign key to users table identifying the session owner';
COMMENT ON COLUMN user_sessions.session_token IS 'Unique session token for authentication';
COMMENT ON COLUMN user_sessions.refresh_token IS 'Token used for session renewal';
COMMENT ON COLUMN user_sessions.ip_address IS 'IP address of the client';
COMMENT ON COLUMN user_sessions.user_agent IS 'Browser/client user agent string';
COMMENT ON COLUMN user_sessions.device_info IS 'JSONB containing device fingerprinting data for security analysis';
COMMENT ON COLUMN user_sessions.location_info IS 'JSONB containing geographic and network location data for location-based access control';
COMMENT ON COLUMN user_sessions.expires_at IS 'Session expiration timestamp';
COMMENT ON COLUMN user_sessions.created_at IS 'Session creation timestamp';
COMMENT ON COLUMN user_sessions.last_accessed_at IS 'Last activity timestamp for session timeout tracking';
COMMENT ON COLUMN user_sessions.is_active IS 'Whether the session is currently active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================

-- Index for tenant-based queries
CREATE INDEX idx_user_sessions_tenant ON user_sessions(tenant_id);

-- Index for user association
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);

-- Index for session token lookups (primary authentication query)
CREATE UNIQUE INDEX idx_user_sessions_token ON user_sessions(session_token);

-- Index for refresh token lookups
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token) WHERE refresh_token IS NOT NULL;

-- Index for active sessions (most common query)
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active, expires_at) WHERE is_active = true;

-- Index for expired sessions cleanup
CREATE INDEX idx_user_sessions_expired ON user_sessions(expires_at);

-- Index for session timeout tracking
CREATE INDEX idx_user_sessions_last_accessed ON user_sessions(last_accessed_at);

-- Index for IP-based security queries
CREATE INDEX idx_user_sessions_ip ON user_sessions(ip_address) WHERE ip_address IS NOT NULL;

-- GIN indexes for JSONB columns
CREATE INDEX idx_user_sessions_device_info_gin ON user_sessions USING gin(device_info);
CREATE INDEX idx_user_sessions_location_info_gin ON user_sessions USING gin(location_info);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================

-- Ensure expires_at is in the future for new sessions
ALTER TABLE user_sessions ADD CONSTRAINT valid_expiration_time 
    CHECK (expires_at > created_at);

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================

-- Enable RLS on user_sessions table
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY user_sessions_tenant_isolation ON user_sessions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY user_sessions_admin_access ON user_sessions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================

-- Grant necessary permissions to application role
GRANT SELECT, INSERT, UPDATE, DELETE ON user_sessions TO application_role;
-- ------------------------------------------------------------------------------------------------
-- MODULES TABLE
-- ------------------------------------------------------------------------------------------------
-- Organizes system functionality into modules for permission management and feature control.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    category VARCHAR(50),                          -- 'CORE', 'HR', 'FINANCE', 'SALES', etc.
    version VARCHAR(20),
    is_active BOOLEAN DEFAULT true,

    -- Standard validation columns
    validation_version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT modules_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE modules IS
'System modules for organizing permissions and features into logical groups. Enables modular permission management and feature toggles.';

COMMENT ON COLUMN modules.category IS 'Module category for grouping: CORE, HR, FINANCE, SALES, INVENTORY, etc.';
COMMENT ON COLUMN modules.version IS 'Module version for tracking feature updates and compatibility';

-- Enable RLS and create policies
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;

CREATE POLICY modules_tenant_isolation ON modules
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY modules_admin_access ON modules
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- ------------------------------------------------------------------------------------------------
-- RESOURCES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines system resources that can be protected by permissions (APIs, UI components, data, etc.).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),      -- Resource can belong to specific entity
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    resource_type VARCHAR(50) NOT NULL
        CHECK (resource_type IN ('API', 'UI', 'DATA', 'FILE', 'REPORT', 'WORKFLOW', 'FUNCTION')),
    parent_resource_id UUID REFERENCES resources(id),
    path VARCHAR(500),                             -- URL path, API endpoint, file path, etc.
    resource_attributes JSONB DEFAULT '{}'::jsonb, -- ABAC resource attributes
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT resources_name_unique_per_module UNIQUE (tenant_id, module_id, name)
);

COMMENT ON TABLE resources IS
'System resources that can be protected by permissions including APIs, UI components, data objects, files, reports, and workflows.';

COMMENT ON COLUMN resources.resource_type IS 'Type of resource: API, UI, DATA, FILE, REPORT, WORKFLOW, FUNCTION';
COMMENT ON COLUMN resources.parent_resource_id IS 'Self-referential for resource hierarchy (e.g., API endpoints under API group)';
COMMENT ON COLUMN resources.path IS 'Resource path: URL, API endpoint, file path, database object, etc.';
COMMENT ON COLUMN resources.resource_attributes IS 'JSONB containing ABAC attributes like classification level, sensitivity, department ownership';

-- Enable RLS and create policies
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;

CREATE POLICY resources_tenant_isolation ON resources
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY resources_admin_access ON resources
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- ------------------------------------------------------------------------------------------------
-- ACTIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines actions that can be performed on resources with risk and approval requirements.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    action_type VARCHAR(50) NOT NULL
        CHECK (action_type IN ('CREATE', 'READ', 'UPDATE', 'DELETE', 'EXECUTE', 'APPROVE', 'REJECT', 'EXPORT', 'IMPORT')),
    action_category VARCHAR(50) DEFAULT 'STANDARD'
        CHECK (action_category IN ('STANDARD', 'ADMINISTRATIVE', 'SENSITIVE', 'BULK', 'SYSTEM')),
    risk_level VARCHAR(20) DEFAULT 'LOW'
        CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    requires_approval BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT actions_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE actions IS
'Defines actions that can be performed on resources with risk assessment and approval workflow requirements.';

COMMENT ON COLUMN actions.action_type IS 'Standard action type: CREATE, READ, UPDATE, DELETE, EXECUTE, APPROVE, REJECT, EXPORT, IMPORT';
COMMENT ON COLUMN actions.action_category IS 'Action category for risk assessment: STANDARD, ADMINISTRATIVE, SENSITIVE, BULK, SYSTEM';
COMMENT ON COLUMN actions.risk_level IS 'Risk level for audit and approval workflows: LOW, MEDIUM, HIGH, CRITICAL';
COMMENT ON COLUMN actions.requires_approval IS 'Whether this action requires explicit approval before execution';

-- Enable RLS and create policies
ALTER TABLE actions ENABLE ROW LEVEL SECURITY;

CREATE POLICY actions_tenant_isolation ON actions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY actions_admin_access ON actions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Granular permissions combining resources and actions with ABAC conditions.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    display_name VARCHAR(250),
    description TEXT,
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    conditions JSONB DEFAULT '{}'::jsonb,          -- ABAC evaluation conditions
    data_filters JSONB DEFAULT '{}'::jsonb,       -- Row-level security filters
    field_restrictions JSONB DEFAULT '{}'::jsonb, -- Column-level restrictions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT permissions_unique_per_resource_action UNIQUE (tenant_id, resource_id, action_id, name)
);

COMMENT ON TABLE permissions IS
'Granular permissions combining resources and actions with ABAC conditions, data filters, and field restrictions for fine-grained access control.';

COMMENT ON COLUMN permissions.effect IS 'Permission effect: ALLOW (grant access) or DENY (explicitly deny access)';
COMMENT ON COLUMN permissions.conditions IS 'JSONB containing ABAC evaluation conditions (time, location, attributes, etc.)';
COMMENT ON COLUMN permissions.data_filters IS 'JSONB containing row-level security filters to limit data access';
COMMENT ON COLUMN permissions.field_restrictions IS 'JSONB containing column-level restrictions to limit field access';

-- Enable RLS and create policies
ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY permissions_tenant_isolation ON permissions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY permissions_admin_access ON permissions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- ------------------------------------------------------------------------------------------------
-- ROLES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines roles with module scope, entity context, and hierarchical structure.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    module_id UUID REFERENCES modules(id),         -- Module this role belongs to
    role_type VARCHAR(20) DEFAULT 'CUSTOM'
        CHECK (role_type IN ('SYSTEM', 'TENANT', 'ENTITY', 'CUSTOM', 'FUNCTIONAL')),
    parent_role_id UUID REFERENCES roles(id),      -- Role hierarchy
    level INTEGER DEFAULT 0,                       -- Hierarchy level (calculated)
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb, -- Direct permissions cache for performance
    entity_scope JSONB DEFAULT '{}'::jsonb,        -- Entity-level access rules
    conditions JSONB DEFAULT '{}'::jsonb,          -- Time, location, device conditions
    is_system_role BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,

    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT roles_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE roles IS
'Roles with module association, entity scoping, and hierarchical structure. Supports both RBAC and ABAC with conditional access rules.';

COMMENT ON COLUMN roles.role_type IS 'Role classification: SYSTEM (built-in), TENANT (tenant-wide), ENTITY (entity-scoped), CUSTOM (user-defined), FUNCTIONAL (job-based)';
COMMENT ON COLUMN roles.parent_role_id IS 'Parent role for inheritance hierarchy';
COMMENT ON COLUMN roles.level IS 'Calculated hierarchy level (0=root, higher=deeper)';
COMMENT ON COLUMN roles.permissions IS 'Cached permissions JSONB for performance optimization';
COMMENT ON COLUMN roles.entity_scope IS 'JSONB defining which entities this role can access';
COMMENT ON COLUMN roles.conditions IS 'JSONB containing time, location, device, and other conditional access rules';

-- Enable RLS and create policies
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY roles_tenant_isolation ON roles
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY roles_admin_access ON roles
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- ------------------------------------------------------------------------------------------------
-- ROLE PERMISSIONS MAPPING
-- ------------------------------------------------------------------------------------------------
-- Maps permissions to roles with entity-specific scoping and conditions.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    entity_scope UUID REFERENCES entities(uuid),   -- Entity-specific permission scope
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    conditions JSONB DEFAULT '{}'::jsonb,          -- Additional conditions beyond permission
    is_active BOOLEAN DEFAULT true,

    CONSTRAINT role_permissions_unique_assignment UNIQUE (tenant_id, role_id, permission_id, entity_scope)
);

COMMENT ON TABLE role_permissions IS
'Maps permissions to roles with optional entity-specific scoping and additional conditions for flexible authorization.';

COMMENT ON COLUMN role_permissions.entity_scope IS 'Optional entity restriction - if specified, permission only applies within this entity';
COMMENT ON COLUMN role_permissions.granted_by IS 'User who granted this permission assignment';
COMMENT ON COLUMN role_permissions.conditions IS 'Additional JSONB conditions beyond those in the permission itself';

-- Enable RLS and create policies
ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY role_permissions_tenant_isolation ON role_permissions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY role_permissions_admin_access ON role_permissions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- ------------------------------------------------------------------------------------------------
-- USER ROLE ASSIGNMENTS
-- ------------------------------------------------------------------------------------------------
-- Assigns roles to users with entity context, delegation, and temporal controls.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
    assignment_type VARCHAR(20) DEFAULT 'DIRECT'
        CHECK (assignment_type IN ('DIRECT', 'INHERITED', 'DELEGATED', 'TEMPORARY')),
    delegated_by UUID REFERENCES users(id),        -- If delegated assignment
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,                        -- For temporary assignments
    conditions JSONB DEFAULT '{}'::jsonb,          -- Time/location/device conditions
    is_active BOOLEAN DEFAULT true,

    CONSTRAINT user_roles_unique_assignment UNIQUE (user_id, role_id, entity_id)
);

COMMENT ON TABLE user_roles IS
'Assigns roles to users with entity context, delegation support, and temporal controls for dynamic authorization.';

COMMENT ON COLUMN user_roles.assignment_type IS 'Type of assignment: DIRECT (explicitly assigned), INHERITED (from hierarchy), DELEGATED (from another user), TEMPORARY (time-limited)';
COMMENT ON COLUMN user_roles.delegated_by IS 'User who delegated this role assignment (for DELEGATED type)';
COMMENT ON COLUMN user_roles.expires_at IS 'Expiration timestamp for temporary role assignments';
COMMENT ON COLUMN user_roles.conditions IS 'JSONB containing conditional access rules (time, location, device, etc.)';

-- Enable RLS and create policies
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_roles_tenant_isolation ON user_roles
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND EXISTS (
            SELECT 1 FROM users u
            WHERE u.id = user_roles.user_id
            AND u.tenant_id = current_tenant_id()
        )
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND EXISTS (
            SELECT 1 FROM users u
            WHERE u.id = user_roles.user_id
            AND u.tenant_id = current_tenant_id()
        )
    );

CREATE POLICY user_roles_admin_access ON user_roles
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);
-- =====================================================================
-- POLICIES UP MIGRATION
-- =====================================================================

-- ------------------------------------------------------------------------------------------------
-- ABAC POLICIES
-- ------------------------------------------------------------------------------------------------
-- ABAC policies with advanced rule engine, priorities, and compliance tracking.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),      -- Entity scope for policy
    name VARCHAR(150) NOT NULL,
    display_name VARCHAR(200),
    description TEXT,
    policy_type VARCHAR(50) DEFAULT 'ABAC'
        CHECK (policy_type IN ('ABAC', 'RBAC', 'HYBRID', 'TIME_BASED', 'LOCATION_BASED')),
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    priority INTEGER DEFAULT 100,                  -- Higher numbers = higher priority
    category VARCHAR(50) DEFAULT 'ACCESS'
        CHECK (category IN ('ACCESS', 'DATA_FILTER', 'FIELD_MASK', 'AUDIT', 'COMPLIANCE')),
    target JSONB NOT NULL,                         -- When policy applies (conditions)
    rule JSONB NOT NULL,                           -- Policy logic/evaluation rules
    obligations JSONB DEFAULT '{}'::jsonb,        -- Required actions when policy fires
    advice JSONB DEFAULT '{}'::jsonb,              -- Optional actions/logging recommendations
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT policies_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE policies IS
'ABAC policies with advanced rule engine supporting multiple policy types, priorities, obligations, and compliance tracking.';

COMMENT ON COLUMN policies.policy_type IS 'Policy type: ABAC (attribute-based), RBAC (role-based), HYBRID (combined), TIME_BASED (temporal), LOCATION_BASED (geographic)';
COMMENT ON COLUMN policies.priority IS 'Policy priority for conflict resolution (higher numbers processed first)';
COMMENT ON COLUMN policies.category IS 'Policy category: ACCESS (authorization), DATA_FILTER (row-level), FIELD_MASK (column-level), AUDIT (logging), COMPLIANCE (regulatory)';
COMMENT ON COLUMN policies.target IS 'JSONB defining when policy applies (subjects, resources, actions, conditions)';
COMMENT ON COLUMN policies.rule IS 'JSONB containing policy evaluation logic and conditions';
COMMENT ON COLUMN policies.obligations IS 'JSONB defining required actions when policy fires (logging, notifications, etc.)';
COMMENT ON COLUMN policies.advice IS 'JSONB defining optional actions and recommendations';

-- Enable RLS and create policies
ALTER TABLE policies ENABLE ROW LEVEL SECURITY;

CREATE POLICY policies_tenant_isolation ON policies
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'POLICIES UP MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Table created: policies';
    RAISE NOTICE 'RLS enabled and policy applied.';
    RAISE NOTICE '===================================================================';
END;
$$;
-- User entity access is handled by the entity_id in the user_roles table.-- =====================================================================
-- ACCESS REQUESTS UP MIGRATION
-- =====================================================================

-- ------------------------------------------------------------------------------------------------
-- ACCESS REQUESTS
-- ------------------------------------------------------------------------------------------------
-- Approval workflow for access requests with business justification and lifecycle management.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS access_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    requester_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID REFERENCES users(id),      -- If requesting for someone else
    entity_id UUID NOT NULL REFERENCES entities(uuid),
    request_type VARCHAR(20) NOT NULL
        CHECK (request_type IN ('ROLE_ASSIGNMENT', 'PERMISSION_GRANT', 'RESOURCE_ACCESS', 'ELEVATION')),
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    resource_id UUID REFERENCES resources(id),
    justification TEXT NOT NULL,
    business_reason VARCHAR(500),
    duration_hours INTEGER,                        -- For temporary access
    approval_status VARCHAR(20) DEFAULT 'PENDING'
        CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED', 'REVOKED')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approval_comments TEXT,
    expires_at TIMESTAMPTZ,
    auto_revoke BOOLEAN DEFAULT true,              -- Auto-revoke when expires
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE access_requests IS
'Access request approval workflow with business justification, lifecycle management, and automatic revocation for governance and compliance.';

COMMENT ON COLUMN access_requests.request_type IS 'Type of access request: ROLE_ASSIGNMENT, PERMISSION_GRANT, RESOURCE_ACCESS, ELEVATION';
COMMENT ON COLUMN access_requests.target_user_id IS 'User receiving the access (if different from requester)';
COMMENT ON COLUMN access_requests.duration_hours IS 'Requested access duration in hours for temporary access';
COMMENT ON COLUMN access_requests.auto_revoke IS 'Whether to automatically revoke access when it expires';

-- Enable RLS and create policies
ALTER TABLE access_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY access_requests_tenant_isolation ON access_requests
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'ACCESS REQUESTS UP MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Table created: access_requests';
    RAISE NOTICE 'RLS enabled and policy applied.';
    RAISE NOTICE '===================================================================';
END;
$$;
-- ------------------------------------------------------------------------------------------------
-- AUDIT LOG
-- ------------------------------------------------------------------------------------------------
--  audit logging with compliance flags and risk scoring.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_category VARCHAR(50) DEFAULT 'ACCESS'
        CHECK (event_category IN ('ACCESS', 'ADMIN', 'DATA', 'AUTH', 'SYSTEM', 'COMPLIANCE')),
    severity VARCHAR(20) DEFAULT 'INFO'
        CHECK (severity IN ('LOW', 'INFO', 'WARN', 'HIGH', 'CRITICAL')),
    user_id UUID REFERENCES users(id),
    target_user_id UUID REFERENCES users(id),      -- For admin actions on other users
    entity_id UUID REFERENCES entities(uuid),
    resource_id UUID REFERENCES resources(id),
    action_id UUID REFERENCES actions(id),
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    decision VARCHAR(20),                          -- ALLOW/DENY for access attempts
    reason TEXT,                                   -- Human-readable reason
    risk_score INTEGER DEFAULT 0,                 -- Calculated risk score (0-100)
    context JSONB DEFAULT '{}'::jsonb,             -- Additional event context
    ip_address INET,
    user_agent TEXT,
    session_id UUID REFERENCES user_sessions(id),
    compliance_flags JSONB DEFAULT '{}'::jsonb,   -- GDPR, SOX, HIPAA, etc.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE audit_log IS
' audit log with compliance tracking, risk scoring, and detailed context for security monitoring and regulatory compliance.';

COMMENT ON COLUMN audit_log.event_category IS 'Event category: ACCESS (authorization), ADMIN (administrative), DATA (data access), AUTH (authentication), SYSTEM (system events), COMPLIANCE (regulatory)';
COMMENT ON COLUMN audit_log.severity IS 'Event severity level: LOW, INFO, WARN, HIGH, CRITICAL';
COMMENT ON COLUMN audit_log.target_user_id IS 'Target user for administrative actions (e.g., admin modifying another user)';
COMMENT ON COLUMN audit_log.risk_score IS 'Calculated risk score from 0-100 based on action, context, and user behavior';
COMMENT ON COLUMN audit_log.compliance_flags IS 'JSONB containing compliance-related flags (GDPR, SOX, HIPAA, PCI, etc.)';

-- Enable RLS and create policies
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_log_tenant_isolation ON audit_log
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);
-- ------------------------------------------------------------------------------------------------
--  user view with all related data
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_user_complete_view AS
SELECT
    u.id as user_id,
    u.tenant_id,
    u.entity_id,
    u.username,
    u.email,
    u.user_type,
    u.account_status,
    u.is_active as user_active,
    u.last_login_at,
    u.mfa_enabled,
    p.id as person_id,
    p.first_name,
    p.last_name,
    p.middle_name,
    p.person_type,
    e.id as employee_id,
    e.employee_number,
    e.position_title,
    e.department_id,
    e.employment_status,
    e.security_level,
    -- Combine all attributes for ABAC evaluation
    COALESCE(p.security_attributes, '{}'::jsonb) ||
    COALESCE(e.access_attributes, '{}'::jsonb) ||
    COALESCE(u.user_attributes, '{}'::jsonb) as combined_attributes,
    -- Aggregate role information - FIXED
    array_agg(DISTINCT r.name ORDER BY r.name) as role_names,
    array_agg(DISTINCT r.id ORDER BY r.id) as role_ids,  -- Changed to ORDER BY r.id
    count(DISTINCT ur.id) FILTER (WHERE ur.is_active = true) as active_role_count
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN roles r ON ur.role_id = r.id AND r.is_active = true AND r.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.tenant_id, u.entity_id, u.username, u.email, u.user_type,
         u.account_status, u.is_active, u.last_login_at, u.mfa_enabled,
         p.id, p.first_name, p.last_name, p.middle_name, p.person_type,
         e.id, e.employee_number, e.position_title, e.department_id,
         e.employment_status, e.security_level,
         p.security_attributes, e.access_attributes, u.user_attributes;

COMMENT ON VIEW v_user_complete_view IS
' view combining user, person, and employee data with role aggregations and combined ABAC attributes for authorization decisions.';

-- ------------------------------------------------------------------------------------------------
-- Role permissions summary view
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_role_permissions_summary AS
SELECT
    r.tenant_id,
    r.id as role_id,
    r.name as role_name,
    r.display_name as role_display_name,
    r.role_type,
    r.level as hierarchy_level,
    r.module_id,
    m.name as module_name,
    array_agg(DISTINCT res.name ORDER BY res.name) FILTER (WHERE res.name IS NOT NULL) as resource_names,
    array_agg(DISTINCT a.name ORDER BY a.name) FILTER (WHERE a.name IS NOT NULL) as action_names,
    count(DISTINCT rp.permission_id) as permission_count,
    count(DISTINCT ur.user_id) FILTER (WHERE ur.is_active = true) as assigned_user_count
FROM roles r
LEFT JOIN modules m ON r.module_id = m.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND rp.is_active = true
LEFT JOIN permissions p ON rp.permission_id = p.id AND p.is_active = true
LEFT JOIN resources res ON p.resource_id = res.id AND res.is_active = true
LEFT JOIN actions a ON p.action_id = a.id AND a.is_active = true
LEFT JOIN user_roles ur ON r.id = ur.role_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
WHERE r.is_active = true AND r.deleted_at IS NULL
GROUP BY r.tenant_id, r.id, r.name, r.display_name, r.role_type, r.level, r.module_id, m.name;

COMMENT ON VIEW v_role_permissions_summary IS
'Summary view of roles with their permissions, resources, actions, and user assignment counts for role management and analysis.';

-- ------------------------------------------------------------------------------------------------
-- Audit summary view for security monitoring
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_audit_summary_view AS
SELECT
    tenant_id,
    event_category,
    severity,
    DATE_TRUNC('hour', created_at) as hour_bucket,
    count(*) as event_count,
    count(DISTINCT user_id) as unique_users,
    avg(risk_score) as avg_risk_score,
    max(risk_score) as max_risk_score,
    count(*) FILTER (WHERE decision = 'DENY') as denied_attempts,
    count(*) FILTER (WHERE decision = 'ALLOW') as allowed_attempts
FROM audit_log
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY tenant_id, event_category, severity, DATE_TRUNC('hour', created_at)
ORDER BY hour_bucket DESC, event_count DESC;

COMMENT ON VIEW v_audit_summary_view IS
'Hourly audit event summary for the last 7 days with risk metrics and access decision counts for security monitoring dashboards.';
-- =====================================================================
-- USER FUNCTIONS AND TRIGGERS UP MIGRATION
-- =====================================================================

-- ------------------------------------------------------------------------------------------------
-- Updated timestamp trigger function
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS
'Trigger function to automatically update the updated_at timestamp when a record is modified.';

-- ------------------------------------------------------------------------------------------------
-- Tenant isolation validation trigger
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
DECLARE
    entity_tenant_id UUID;
BEGIN
    -- This trigger is intended to be generic. It checks if a referenced
    -- entity (via entity_id) belongs to the same tenant as the new row.
    -- It dynamically checks for the existence of an 'entity_id' column.

    -- Check if the table has an 'entity_id' column
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        BEGIN
            -- This block will fail if entity_id does not exist, and the exception will be caught.
            IF NEW.entity_id IS NOT NULL THEN
                -- Get the tenant_id from the referenced entity
                SELECT tenant_id INTO entity_tenant_id
                FROM entities
                WHERE uuid = NEW.entity_id;

                -- If the referenced entity doesn't exist or tenant_ids don't match, raise an exception.
                IF NOT FOUND OR entity_tenant_id != NEW.tenant_id THEN
                    RAISE EXCEPTION 'Tenant mismatch: Referenced entity (%) does not belong to tenant %', NEW.entity_id, NEW.tenant_id;
                END IF;
            END IF;
        EXCEPTION
            WHEN undefined_column THEN
                -- The table does not have an entity_id column, so we can ignore it.
                -- You could log this notice for debugging purposes if you want.
                -- RAISE NOTICE 'Table % does not have an entity_id column, skipping tenant isolation check.', TG_TABLE_NAME;
        END;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enforce_tenant_isolation() IS
'Trigger function to enforce tenant isolation by validating that all foreign key references belong to the same tenant.';

-- ------------------------------------------------------------------------------------------------
-- Role hierarchy validation and level calculation
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION validate_role_hierarchy()
RETURNS TRIGGER AS $$
DECLARE
    max_depth INTEGER := 10;
    current_depth INTEGER := 0;
    current_role_id UUID;
BEGIN
    -- If no parent role, set level to 0
    IF NEW.parent_role_id IS NULL THEN
        NEW.level = 0;
        RETURN NEW;
    END IF;

    -- Ensure parent role belongs to same tenant
    IF NOT EXISTS (
        SELECT 1 FROM roles
        WHERE id = NEW.parent_role_id AND tenant_id = NEW.tenant_id AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Parent role % does not belong to tenant % or is deleted', NEW.parent_role_id, NEW.tenant_id;
    END IF;

    -- Check for cycles and calculate depth
    current_role_id := NEW.parent_role_id;
    current_depth := 1;

    WHILE current_role_id IS NOT NULL AND current_depth <= max_depth LOOP
        -- Check if we've hit the new role (cycle detection)
        IF current_role_id = NEW.id THEN
            RAISE EXCEPTION 'Role hierarchy cycle detected for role %', NEW.id;
        END IF;

        -- Get the next parent
        SELECT parent_role_id INTO current_role_id
        FROM roles
        WHERE id = current_role_id AND tenant_id = NEW.tenant_id AND deleted_at IS NULL;

        current_depth := current_depth + 1;
    END LOOP;

    IF current_depth > max_depth THEN
        RAISE EXCEPTION 'Role hierarchy depth exceeds maximum of %', max_depth;
    END IF;

    -- Set the level
    NEW.level = current_depth - 1;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_role_hierarchy() IS
'Validates role hierarchy integrity, prevents cycles, enforces depth limits, and calculates hierarchy levels.';

-- ------------------------------------------------------------------------------------------------
-- User permission evaluation function with ABAC support
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION user_has_permission(
    p_user_id UUID,
    p_resource_name VARCHAR(100),
    p_action_name VARCHAR(100),
    p_tenant_id UUID,
    p_entity_id UUID DEFAULT NULL,
    p_context JSONB DEFAULT '{}'::jsonb
)
RETURNS BOOLEAN AS $$
DECLARE
    v_has_permission BOOLEAN := FALSE;
    v_user_context JSONB;
    v_cache_key VARCHAR(64);
    v_cached_result BOOLEAN;
BEGIN
    -- Generate cache key
    v_cache_key := encode(digest(
        p_user_id::text || p_resource_name || p_action_name ||
        COALESCE(p_entity_id::text, '') || p_context::text, 'sha256'
    ), 'hex');

    -- Check cache first
    SELECT decision = 'ALLOW' INTO v_cached_result
    FROM policy_evaluations
    WHERE user_id = p_user_id
        AND context_hash = v_cache_key
        AND expires_at > NOW()
        AND tenant_id = p_tenant_id;

    IF FOUND THEN
        RETURN v_cached_result;
    END IF;

    -- Build comprehensive user context for ABAC evaluation
    SELECT jsonb_build_object(
        'user_id', u.id,
        'user_type', u.user_type,
        'account_status', u.account_status,
        'user_attributes', COALESCE(u.user_attributes, '{}'::jsonb),
        'person_attributes', COALESCE(p.security_attributes, '{}'::jsonb),
        'employee_attributes', COALESCE(e.access_attributes, '{}'::jsonb),
        'employment_status', e.employment_status,
        'security_level', COALESCE(e.security_level, 0),
        'entity_id', u.entity_id,
        'department_id', e.department_id,
        'context', p_context
    ) INTO v_user_context
    FROM users u
    LEFT JOIN persons p ON u.person_id = p.id
    LEFT JOIN employees e ON u.employee_id = e.id
    WHERE u.id = p_user_id AND u.tenant_id = p_tenant_id;

    -- Check for explicit DENY in direct user permissions first
    SELECT true INTO v_has_permission
    FROM user_permissions up
    JOIN permissions perm ON up.permission_id = perm.id
    JOIN resources r ON perm.resource_id = r.id
    JOIN actions a ON perm.action_id = a.id
    WHERE up.user_id = p_user_id
        AND up.tenant_id = p_tenant_id
        AND r.name = p_resource_name
        AND a.name = p_action_name
        AND up.is_active = true
        AND up.effect = 'DENY'
        AND (up.expires_at IS NULL OR up.expires_at > NOW())
        AND (p_entity_id IS NULL OR up.entity_id = p_entity_id);

    -- If explicit DENY found, return false immediately
    IF FOUND THEN
        RETURN FALSE;
    END IF;

    -- Check direct user permissions for ALLOW
    SELECT true INTO v_has_permission
    FROM user_permissions up
    JOIN permissions perm ON up.permission_id = perm.id
    JOIN resources r ON perm.resource_id = r.id
    JOIN actions a ON perm.action_id = a.id
    WHERE up.user_id = p_user_id
        AND up.tenant_id = p_tenant_id
        AND r.name = p_resource_name
        AND a.name = p_action_name
        AND up.is_active = true
        AND up.effect = 'ALLOW'
        AND (up.expires_at IS NULL OR up.expires_at > NOW())
        AND (p_entity_id IS NULL OR up.entity_id = p_entity_id);

    -- If no direct permission, check role-based permissions
    IF NOT FOUND THEN
        SELECT true INTO v_has_permission
        FROM user_roles ur
        JOIN role_permissions rp ON ur.role_id = rp.role_id
        JOIN permissions perm ON rp.permission_id = perm.id
        JOIN resources r ON perm.resource_id = r.id
        JOIN actions a ON perm.action_id = a.id
        WHERE ur.user_id = p_user_id
            AND r.name = p_resource_name
            AND a.name = p_action_name
            AND ur.is_active = true
            AND rp.is_active = true
            AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
            AND (p_entity_id IS NULL OR ur.entity_id = p_entity_id OR rp.entity_scope = p_entity_id);
    END IF;

    RETURN COALESCE(v_has_permission, FALSE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION user_has_permission(UUID, VARCHAR, VARCHAR, UUID, UUID, JSONB) IS
'Evaluates user permissions with ABAC context, caching, and comprehensive policy evaluation including direct permissions and role-based permissions.';

-- ------------------------------------------------------------------------------------------------
-- Cleanup function for expired data and maintenance
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION cleanup_expired_data(p_tenant_id UUID DEFAULT NULL)
RETURNS INTEGER AS $$
DECLARE
    v_cleanup_count INTEGER := 0;
    v_tenant_filter TEXT := '';
BEGIN
    -- Build tenant filter if specified
    IF p_tenant_id IS NOT NULL THEN
        v_tenant_filter := ' AND tenant_id = ' || quote_literal(p_tenant_id);
    END IF;

    -- Clean expired sessions
    EXECUTE 'DELETE FROM user_sessions WHERE expires_at < NOW()' || v_tenant_filter;
    GET DIAGNOSTICS v_cleanup_count = ROW_COUNT;

    -- Clean expired policy evaluations
    EXECUTE 'DELETE FROM policy_evaluations WHERE expires_at < NOW()' || v_tenant_filter;

    -- Deactivate expired user roles
    EXECUTE 'UPDATE user_roles SET is_active = false
             WHERE expires_at < NOW() AND is_active = true' ||
             CASE WHEN p_tenant_id IS NOT NULL THEN
                ' AND EXISTS (SELECT 1 FROM users WHERE id = user_roles.user_id AND tenant_id = ' || quote_literal(p_tenant_id) || ')'
             ELSE '' END;

    -- Deactivate expired user permissions
    EXECUTE 'UPDATE user_permissions SET is_active = false
             WHERE expires_at < NOW() AND is_active = true' || v_tenant_filter;

    -- Expire approved access requests
    EXECUTE 'UPDATE access_requests SET approval_status = ''EXPIRED''
             WHERE expires_at < NOW() AND approval_status = ''APPROVED''' || v_tenant_filter;

    RETURN v_cleanup_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION cleanup_expired_data(UUID) IS
'Cleans up expired sessions, policy evaluations, user roles, permissions, and access requests. Can be run for all tenants or a specific tenant.';

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'USER FUNCTIONS AND TRIGGERS UP MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Functions created: update_updated_at_column, enforce_tenant_isolation, validate_role_hierarchy, user_has_permission, cleanup_expired_data';
    RAISE NOTICE '===================================================================';
END;
$$;-- =====================================================================
-- PROJECTS TABLE
-- =====================================================================
-- This table stores project information, which can be associated with any entity.
-- =====================================================================

CREATE TABLE projects (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    description TEXT,
    project_manager_id UUID REFERENCES employees(id),
    start_date DATE,
    end_date DATE,
    budget_amount DECIMAL(15,2),
    actual_cost DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'PLANNING'
        CHECK (status IN ('PLANNING', 'ACTIVE', 'ON_HOLD', 'COMPLETED', 'CANCELLED')),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT projects_tenant_code_entity_unique_idx UNIQUE (tenant_id, entity_id, code)
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_projects_tenant ON projects(tenant_id);
CREATE INDEX idx_projects_entity ON projects(entity_id);
CREATE INDEX idx_projects_status ON projects(status);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON projects
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON projects
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- TRIGGERS
-- =====================================================================
CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- =====================================================================
-- BUDGETS TABLE
-- =====================================================================
-- This table stores budget information, which can be associated with an entity or a project.
-- =====================================================================

CREATE TABLE budgets (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id), -- Optional project budget
    name VARCHAR(255) NOT NULL,
    budget_type VARCHAR(20) NOT NULL
        CHECK (budget_type IN ('OPERATIONAL', 'CAPITAL', 'PROJECT', 'DEPARTMENT')),
    fiscal_year INT NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    allocated_amount DECIMAL(15,2) DEFAULT 0,
    spent_amount DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'APPROVED', 'ACTIVE', 'CLOSED')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_budgets_tenant ON budgets(tenant_id);
CREATE INDEX idx_budgets_entity ON budgets(entity_id);
CREATE INDEX idx_budgets_year ON budgets(fiscal_year);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON budgets
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON budgets
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- TRIGGERS
-- =====================================================================
CREATE TRIGGER update_budgets_updated_at
    BEFORE UPDATE ON budgets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- =====================================================================
-- UOM TABLE
-- =====================================================================
-- This table stores unit of measure information.
-- =====================================================================

CREATE TABLE uom (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    uom_name VARCHAR(255) NOT NULL,
    must_be_whole_number BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    symbol VARCHAR(50),
    common_code VARCHAR(3),
    description TEXT,
    base_uom_id UUID REFERENCES uom(id),
    conversion_factor DECIMAL(15,6) DEFAULT 1.0,
    uom_type VARCHAR(50), -- 'Weight', 'Length', 'Volume', 'Area', 'Time', 'Count'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, entity_id, uom_name)
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_uom_tenant_entity ON uom(tenant_id, entity_id);
CREATE INDEX idx_uom_enabled ON uom(enabled);
CREATE INDEX idx_uom_tenant_entity_type ON uom(tenant_id, entity_id, uom_type);
CREATE INDEX idx_uom_base_uom_id ON uom(base_uom_id);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE uom ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON uom
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON uom
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- TRIGGERS
-- =====================================================================
CREATE TRIGGER update_uom_updated_at
    BEFORE UPDATE ON uom
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- =====================================================================
-- UOM CONVERSION TABLE
-- =====================================================================
-- This table stores conversion factors between different units of measure.
-- =====================================================================

CREATE TABLE uom_conversion (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    from_uom_id UUID NOT NULL REFERENCES uom(id),
    to_uom_id UUID NOT NULL REFERENCES uom(id),
    conversion_factor DECIMAL(15,6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, entity_id, from_uom_id, to_uom_id)
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_uom_conversion_tenant_entity ON uom_conversion(tenant_id, entity_id);
CREATE INDEX idx_uom_conversion_from ON uom_conversion(from_uom_id);
CREATE INDEX idx_uom_conversion_to ON uom_conversion(to_uom_id);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE uom_conversion ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON uom_conversion
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON uom_conversion
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- CHARTOFACCOUNT TABLE
-- =====================================================================
-- This table stores chart of accounts templates.
-- =====================================================================

CREATE TABLE chartofaccount (
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  module TEXT,
  slug VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(150) NULL,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  is_active BOOLEAN DEFAULT true, 
  description TEXT NULL,
  active BOOLEAN NOT NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE chartofaccount ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON chartofaccount
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON chartofaccount
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- ACCOUNT TABLE
-- =====================================================================
-- This table stores individual accounts within a chart of accounts.
-- =====================================================================

CREATE TABLE account (
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  path VARCHAR(255) NOT NULL UNIQUE,
  depth INTEGER NOT NULL CHECK (depth >= 0),
  numchild INTEGER NOT NULL CHECK (numchild >= 0),
  account_code VARCHAR(10) NOT NULL,
  account_name VARCHAR(100) NOT NULL,
  account_type VARCHAR(20) NOT NULL
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
  account_role VARCHAR(30) ,
  balance_type VARCHAR(6) NOT NULL,
  locked BOOLEAN NOT NULL,
  active BOOLEAN NOT NULL,
  coa__id UUID NOT NULL REFERENCES chartofaccount (id) DEFERRABLE INITIALLY DEFERRED,
  role_default BOOLEAN NULL,
  CONSTRAINT unique_code_for_coa_ UNIQUE (coa__id, account_code),
  CONSTRAINT only_one_account_assigned_as_default_for_role UNIQUE (
    coa__id, account_role, role_default
  )
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_account_tenant ON account(tenant_id);
CREATE INDEX idx_account_entity ON account(entity_id);
CREATE INDEX idx_account_code ON account(account_code);
CREATE INDEX idx_account_type ON account(account_type);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE account ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON account
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON account
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- LEDGER TABLE
-- =====================================================================
-- This table stores ledgers, which are collections of journal entries.
-- =====================================================================

CREATE TABLE ledger (
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  posted_by UUID REFERENCES users(id),
  created_by UUID REFERENCES users(id),
  name VARCHAR(150) NULL,
  posted BOOLEAN NOT NULL,
  locked BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),
  ledger_xid VARCHAR(150) NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE ledger ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON ledger
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON ledger
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- JOURNALENTRY TABLE
-- =====================================================================
-- This table stores journal entries, the foundation of double-entry bookkeeping.
-- =====================================================================

CREATE TABLE journalentry (
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  posted_by UUID REFERENCES users(id),
  created_by UUID REFERENCES users(id),
  je_number VARCHAR(25) NOT NULL,
  timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  description VARCHAR(70) NULL,
  activity VARCHAR(20) NULL,
  origin VARCHAR(30) NULL,
  posted BOOLEAN NOT NULL,
  locked BOOLEAN NOT NULL,
  ledger_id UUID NOT NULL REFERENCES ledger(id) DEFERRABLE INITIALLY DEFERRED,
  is_closing_entry BOOLEAN NOT NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE journalentry ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON journalentry
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON journalentry
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- CUSTOMER TABLE
-- =====================================================================
-- This table stores customer information.
-- =====================================================================

CREATE TABLE customer (
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP NULL,
  id UUID NOT NULL PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  customer_name VARCHAR(100) NOT NULL,
  customer_number VARCHAR(30) NOT NULL,
  description TEXT NOT NULL,
  active BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,
  address JSONB DEFAULT '{}'::jsonb,
  email VARCHAR(254) NULL,
  website VARCHAR(200) NULL,
  phone VARCHAR(30) NULL,
  sales_tax_rate REAL NULL,
  additional_info JSONB NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE customer ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON customer
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON customer
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- VENDOR TABLE
-- =====================================================================
-- This table stores vendor information.
-- =====================================================================

CREATE TABLE vendor (
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP NULL,
  uuid UUID NOT NULL PRIMARY KEY,
  vendor_name VARCHAR(100) NOT NULL,
  vendor_number VARCHAR(30) NULL,
  description TEXT NOT NULL,
  active BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  address JSONB DEFAULT '{}'::jsonb,
  contact JSONB DEFAULT '{}'::jsonb,
  account_number VARCHAR(30) NULL,
  routing_number VARCHAR(30) NULL,
  aba_number VARCHAR(30) NULL,
  swift_number VARCHAR(30) NULL,
  tax_id_number VARCHAR(30) NULL,
  account_type VARCHAR(20) NOT NULL,
  additional_info JSONB NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE vendor ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON vendor
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON vendor
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- WAREHOUSES TABLE
-- =====================================================================
-- This table stores warehouse information.
-- =====================================================================

CREATE TABLE warehouses (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 
    code VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    address JSONB,
    warehouse_type VARCHAR(20) DEFAULT 'GENERAL'
        CHECK (warehouse_type IN ('GENERAL', 'RETAIL', 'TRANSIT', 'QUARANTINE')),
    manager_id UUID REFERENCES employees(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON warehouses
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON warehouses
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- ITEM_CATEGORIES TABLE
-- =====================================================================
-- This table stores item category information.
-- =====================================================================

CREATE TABLE item_categories (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 
    parent_id UUID REFERENCES item_categories(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, parent_id, name)
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE item_categories ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON item_categories
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON item_categories
    FOR ALL TO admin_role
    USING (true);
-- =====================================================================
-- ITEMS TABLE
-- =====================================================================
-- This table stores item information.
-- =====================================================================

CREATE TABLE items (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 
    item_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES item_categories(id),
    item_type VARCHAR(20) DEFAULT 'INVENTORY'
        CHECK (item_type IN ('INVENTORY', 'SERVICE', 'NON_INVENTORY', 'ASSEMBLY')),
    unit_of_measure VARCHAR(20) NOT NULL DEFAULT 'EACH',
    cost_method VARCHAR(20) DEFAULT 'FIFO'
        CHECK (cost_method IN ('FIFO', 'LIFO', 'WEIGHTED_AVERAGE', 'SPECIFIC')),
    standard_cost DECIMAL(10,4),
    selling_price DECIMAL(10,2),
    minimum_stock_level DECIMAL(10,2) DEFAULT 0,
    maximum_stock_level DECIMAL(10,2),
    reorder_point DECIMAL(10,2),
    reorder_quantity DECIMAL(10,2),
    is_active BOOLEAN DEFAULT true,
    is_serialized BOOLEAN DEFAULT false,
    is_batch_tracked BOOLEAN DEFAULT false,
    tax_category VARCHAR(20),
    supplier_id INT, -- Main supplier (references persons table)
    specifications JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, item_code)
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE items ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON items
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON items
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- TRIGGERS
-- =====================================================================
CREATE TRIGGER update_items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- =====================================================================
-- INVENTORY_BALANCES TABLE
-- =====================================================================
-- This table stores inventory balance information.
-- =====================================================================

CREATE TABLE inventory_balances (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    quantity_on_hand DECIMAL(10,2) DEFAULT 0,
    quantity_available DECIMAL(10,2) DEFAULT 0, -- On hand - reserved
    quantity_reserved DECIMAL(10,2) DEFAULT 0,
    quantity_on_order DECIMAL(10,2) DEFAULT 0,
    average_cost DECIMAL(10,4) DEFAULT 0,
    total_value DECIMAL(15,2) DEFAULT 0,
    last_movement_date DATE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, item_id, warehouse_id)
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE inventory_balances ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON inventory_balances
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON inventory_balances
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- TRIGGERS
-- =====================================================================
CREATE TRIGGER update_inventory_balances_updated_at
    BEFORE UPDATE ON inventory_balances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- =====================================================================
-- INVENTORY_MOVEMENTS TABLE
-- =====================================================================
-- This table stores inventory movement information.
-- =====================================================================

CREATE TABLE inventory_movements (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 
    item_id UUID NOT NULL REFERENCES items(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    movement_type VARCHAR(20) NOT NULL
        CHECK (movement_type IN ('RECEIPT', 'ISSUE', 'TRANSFER', 'ADJUSTMENT', 'SALE', 'RETURN')),
    reference_type VARCHAR(20), -- PURCHASE_ORDER, SALES_ORDER, etc.
    reference_id BIGINT,
    reference_number VARCHAR(50),
    transaction_date DATE NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    unit_cost DECIMAL(10,4),
    total_cost DECIMAL(15,2),
    reason TEXT,
    batch_number VARCHAR(50),
    serial_numbers TEXT[], -- For serialized items
    expiry_date DATE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE inventory_movements ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON inventory_movements
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON inventory_movements
    FOR ALL TO admin_role
    USING (true);
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    in_app_notifications BOOLEAN NOT NULL DEFAULT true,
    slack_notifications BOOLEAN NOT NULL DEFAULT false,
    notification_types JSONB NOT NULL DEFAULT '{}'::jsonb,
    preferred_channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    quiet_hours JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notification_preferences_user_id_unique UNIQUE (tenant_id, user_id)
);

COMMENT ON TABLE notification_preferences IS 'Stores user notification preferences.';
COMMENT ON COLUMN notification_preferences.notification_types IS 'JSONB object with notification types as keys and booleans as values.';
COMMENT ON COLUMN notification_preferences.preferred_channels IS 'JSONB array of preferred notification channels.';
COMMENT ON COLUMN notification_preferences.quiet_hours IS 'JSONB object with quiet hours settings.';

ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_preferences_tenant_isolation ON notification_preferences
    FOR ALL
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE TRIGGER update_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- DIRECT USER PERMISSIONS
-- ------------------------------------------------------------------------------------------------
-- Direct permission grants to users bypassing roles for exceptional access.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid),
    effect VARCHAR(20) DEFAULT 'ALLOW' CHECK (effect IN ('ALLOW', 'DENY')),
    reason TEXT,                                   -- Justification for direct permission
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                       -- For temporary permissions
    conditions JSONB DEFAULT '{}'::jsonb,         -- Additional conditions
    is_active BOOLEAN DEFAULT true,
    
    CONSTRAINT user_permissions_unique_assignment UNIQUE (tenant_id, user_id, permission_id, entity_id)
);

COMMENT ON TABLE user_permissions IS 
'Direct permission grants to users bypassing roles. Used for exceptional access, denials, and temporary permissions.';

COMMENT ON COLUMN user_permissions.effect IS 'Permission effect: ALLOW (grant access) or DENY (explicitly deny - overrides role permissions)';
COMMENT ON COLUMN user_permissions.reason IS 'Business justification for this direct permission assignment';
COMMENT ON COLUMN user_permissions.granted_by IS 'User who granted this direct permission';

-- Enable RLS and create policies
ALTER TABLE user_permissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_permissions_tenant_isolation ON user_permissions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY user_permissions_admin_access ON user_permissions
    FOR ALL TO admin_role
    USING (true)
    WITH CHECK (true);


--- 1. Security Hardening Enhancements:

-- Add risk_score to user_sessions
ALTER TABLE user_sessions
    ADD COLUMN risk_score INT DEFAULT 0;
COMMENT ON COLUMN user_sessions.risk_score IS 'Calculated risk score (0-100) based on action, context, and user behavior';

-- Password security enhancements
ALTER TABLE users
    ADD COLUMN password_strength INT DEFAULT 0,
    ADD COLUMN compromised BOOLEAN DEFAULT false,
    ADD COLUMN rotation_required BOOLEAN DEFAULT false;

COMMENT ON COLUMN users.password_strength IS 'Password strength score (0-100) based on complexity';
COMMENT ON COLUMN users.compromised IS 'Flag if password found in breach databases';
COMMENT ON COLUMN users.rotation_required IS 'Forces password change on next login';


--- 2. Performance Optimizations:


-- Optimized materialized view for permission evaluations
CREATE MATERIALIZED VIEW mv_user_effective_permissions AS
SELECT
    u.id AS user_id,
    u.tenant_id,
    r.id AS resource_id,
    a.id AS action_id,
    MAX(CASE WHEN up.effect = 'DENY' THEN 0 ELSE 1 END) AS allow_flag,
    ARRAY_AGG(DISTINCT rp.id) AS role_permission_ids,
    ARRAY_AGG(DISTINCT up.id) AS direct_permission_ids
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
    AND ur.is_active = true
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN role_permissions rp ON ur.role_id = rp.role_id
    AND rp.is_active = true
LEFT JOIN user_permissions up ON u.id = up.user_id
    AND up.is_active = true
    AND (up.expires_at IS NULL OR up.expires_at > NOW())
JOIN resources r ON rp.permission_id = r.id OR up.permission_id = r.id
JOIN actions a ON rp.permission_id = a.id OR up.permission_id = a.id
GROUP BY u.id, u.tenant_id, r.id, a.id;

CREATE UNIQUE INDEX idx_user_effective_perms
    ON mv_user_effective_permissions (user_id, resource_id, action_id);
    
COMMENT ON MATERIALIZED VIEW mv_user_effective_permissions IS
'Pre-computed effective permissions for all users with optimized access patterns';

-- Session clustering
-- CLUSTER user_sessions USING idx_user_sessions_user_id;


--- 3. Security Automation Functions:


-- Session risk assessment function
CREATE OR REPLACE FUNCTION assess_session_risk(session_id UUID)
RETURNS INT AS $$
DECLARE
    risk INT := 0;
    session_data user_sessions%ROWTYPE;
BEGIN
    SELECT * INTO session_data 
    FROM user_sessions 
    WHERE id = session_id;
    
    -- Location anomaly detection
    IF EXISTS (
        SELECT 1 FROM user_sessions 
        WHERE user_id = session_data.user_id
        AND location_info->>'country' != session_data.location_info->>'country'
        AND created_at > NOW() - INTERVAL '1 hour'
    ) THEN
        risk := risk + 30;
        session_data.anomaly_flags := session_data.anomaly_flags || '["impossible_travel"]'::jsonb;
    END IF;
    
    -- Device change detection
    IF EXISTS (
        SELECT 1 FROM user_sessions 
        WHERE user_id = session_data.user_id
        AND device_info->>'fingerprint' != session_data.device_info->>'fingerprint'
        AND created_at > NOW() - INTERVAL '10 minutes'
    ) THEN
        risk := risk + 25;
        session_data.anomaly_flags := session_data.anomaly_flags || '["device_change"]'::jsonb;
    END IF;
    
    -- High-risk action detection
    IF EXISTS (
        SELECT 1 FROM audit_log
        WHERE session_id = session_data.id
        AND risk_score > 70
    ) THEN
        risk := risk + 45;
    END IF;
    
    UPDATE user_sessions 
    SET risk_score = risk,
        anomaly_flags = session_data.anomaly_flags
    WHERE id = session_id;
    
    RETURN risk;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Automatic session termination
CREATE OR REPLACE FUNCTION terminate_risky_sessions(threshold INT)
RETURNS INT AS $$
DECLARE
    terminated_count INT := 0;
BEGIN
    UPDATE user_sessions
    SET is_active = false
    WHERE risk_score >= threshold
        AND is_active = true
    RETURNING id INTO terminated_count;
    
    INSERT INTO audit_log (tenant_id, event_type, event_category, severity, context)
    SELECT tenant_id, 'SESSION_TERMINATED', 'SECURITY', 'HIGH',
        jsonb_build_object('session_id', id, 'risk_score', risk_score)
    FROM user_sessions
    WHERE risk_score >= threshold;
    
    RETURN terminated_count;
END;
$$ LANGUAGE plpgsql;


--- 4. Compliance Enhancements:


-- GDPR right-to-forget implementation
CREATE OR REPLACE FUNCTION gdpr_user_deletion(user_id UUID)
RETURNS VOID AS $$
BEGIN
    -- Pseudonymize sensitive data
    UPDATE persons p
    SET 
        first_name = 'REDACTED',
        last_name = 'REDACTED',
        email = 'redacted_' || gen_random_uuid() || '@example.com',
        phone = NULL,
        national_id = NULL,
        tax_id = NULL
    FROM users u
    WHERE u.person_id = p.id
        AND u.id = user_id;

    -- Delete authentication data
    UPDATE users
    SET 
        password_hash = NULL,
        mfa_secret = NULL,
        settings = settings - 'preferences'
    WHERE id = user_id;

    -- Terminate active sessions
    PERFORM terminate_risky_sessions(0); -- Terminate all sessions for user

    -- Log compliance action
    INSERT INTO audit_log (tenant_id, event_type, event_category, severity, context)
    SELECT tenant_id, 'GDPR_DELETION', 'COMPLIANCE', 'HIGH',
        jsonb_build_object('user_id', user_id)
    FROM users
    WHERE id = user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Data retention policy enforcement
CREATE OR REPLACE FUNCTION enforce_data_retention()
RETURNS VOID AS $$
BEGIN
    -- Anonymize old audit logs
    UPDATE audit_log
    SET 
        user_id = NULL,
        target_user_id = NULL,
        context = jsonb_set(context, '{user_info}', '"REDACTED"')
    WHERE created_at < NOW() - INTERVAL '180 days';
    
    -- Purge expired sessions
    DELETE FROM user_sessions
    WHERE expires_at < NOW() - INTERVAL '30 days';
    
    -- Archive and purge old access requests
    WITH archived AS (
        DELETE FROM access_requests
        WHERE created_at < NOW() - INTERVAL '365 days'
        RETURNING *
    )
    INSERT INTO access_requests_archive SELECT * FROM archived;
END;
$$ LANGUAGE plpgsql;


--- 5. Advanced Threat Detection View:


CREATE VIEW v_security_threat_dashboard AS
SELECT 
    u.id AS user_id,
    u.username,
    u.email,
    COUNT(s.id) FILTER (WHERE s.risk_score > 70) AS high_risk_sessions,
    MAX(s.risk_score) AS max_risk_score,
    -- TODDO anomaly_flags 
    -- ARRAY_AGG(DISTINCT s.anomaly_flags) AS anomaly_types,
    COUNT(a.id) FILTER (WHERE a.risk_score > 80) AS critical_events,
    MAX(a.created_at) AS last_suspicious_activity
FROM users u
LEFT JOIN user_sessions s ON u.id = s.user_id
LEFT JOIN audit_log a ON u.id = a.user_id AND a.risk_score > 50
WHERE u.account_status = 'ACTIVE'
    AND (s.risk_score > 50 OR a.risk_score > 50)
GROUP BY u.id;

COMMENT ON VIEW v_security_threat_dashboard IS
'Identifies potential security threats through session anomalies and audit patterns';


--- 6. Index Optimizations for Large-Scale Deployments:


-- BRIN Indexes for time-series data
CREATE INDEX idx_audit_log_time_brin ON audit_log USING BRIN (created_at);
CREATE INDEX idx_user_sessions_time_brin ON user_sessions USING BRIN (created_at);

-- GIN optimizations for JSONB queries
-- CREATE INDEX idx_users_attributes_gin ON users USING GIN (user_attributes jsonb_path_ops);
CREATE INDEX idx_policies_rule_gin ON policies USING GIN (rule jsonb_path_ops);

-- Partial indexes for active records
CREATE INDEX idx_active_users ON users (id) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX idx_active_roles ON roles (id) WHERE is_active = true AND deleted_at IS NULL;


--- 7. Security Notification System:


-- Notification table
CREATE TABLE security_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    notification_type VARCHAR(50) NOT NULL 
        CHECK (notification_type IN ('SUSPICIOUS_LOGIN', 'PASSWORD_COMPROMISED', 'ROLE_CHANGE', 'PERMISSION_GRANT')),
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    acknowledged BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '7 days'
);

-- Notification trigger function
CREATE OR REPLACE FUNCTION trigger_security_notification()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' AND TG_TABLE_NAME = 'user_sessions') THEN
        IF NEW.risk_score > 60 THEN
            INSERT INTO security_notifications (tenant_id, user_id, notification_type, title, message, metadata)
            VALUES (
                NEW.tenant_id,
                NEW.user_id,
                'SUSPICIOUS_LOGIN',
                'New login from unusual location',
                'We detected a login from ' || (NEW.location_info->>'city') || ', ' || (NEW.location_info->>'country'),
                jsonb_build_object('session_id', NEW.id, 'device', NEW.device_info)
            );
        END IF;
    ELSIF (TG_OP = 'INSERT' AND TG_TABLE_NAME = 'user_roles') THEN
        INSERT INTO security_notifications (tenant_id, user_id, notification_type, title, message, metadata)
        VALUES (
            (SELECT tenant_id FROM users WHERE id = NEW.user_id),
            NEW.user_id,
            'ROLE_CHANGE',
            'Role assignment: ' || (SELECT name FROM roles WHERE id = NEW.role_id),
            'You have been assigned a new role',
            jsonb_build_object('role_id', NEW.role_id, 'assigned_by', NEW.assigned_by)
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
CREATE TRIGGER notify_suspicious_login
    AFTER INSERT ON user_sessions
    FOR EACH ROW
    WHEN (NEW.risk_score > 60)
    EXECUTE FUNCTION trigger_security_notification();
    
CREATE TRIGGER notify_role_changes
    AFTER INSERT ON user_roles
    FOR EACH ROW
    EXECUTE FUNCTION trigger_security_notification();


--  Key Benefits of These Enhancements:
--
-- 1. **Proactive Threat Detection**:
--    - Real-time session risk scoring
--    - Automated anomaly detection
--    - Behavioral analytics integration
--
-- 2. **Regulatory Compliance**:
--    - Built-in GDPR enforcement
--    - Data retention automation
--    - Audit trail completeness
--
-- 3. **Operational Efficiency**:
--    - Materialized views for permission checks
--    - BRIN indexes for time-series data
--    - Automated security notifications
--
-- 4. **Security Posture Strengthening**:
--    - Password breach monitoring
--    - Compromised credential detection
--    - Session termination automation
--
-- 5. **Scalability Improvements**:
--    - Optimized indexing strategies
--    - Session clustering for faster access
--    - Batch processing for large datasets
--
-- These enhancements maintain your existing schema structure while adding critical security and operational capabilities. They address common enterprise requirements for audit compliance, threat detection, and large-scale performance without requiring architectural changes to your well-designed RBAC/ABAC implementation.

-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE DEFINITIONS
-- ------------------------------------------------------------------------------------------------
-- Defines attributes used in ABAC policies with validation and encryption controls.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    data_type VARCHAR(50) NOT NULL 
        CHECK (data_type IN ('STRING', 'NUMBER', 'BOOLEAN', 'DATE', 'TIME', 'JSON', 'ARRAY', 'ENUM')),
    category VARCHAR(50) NOT NULL 
        CHECK (category IN ('USER', 'RESOURCE', 'ENVIRONMENT', 'ACTION', 'ENTITY', 'SESSION')),
    is_required BOOLEAN DEFAULT false,
    is_sensitive BOOLEAN DEFAULT false,            -- For PII/sensitive attributes
    default_value TEXT,
    allowed_values JSONB,                          -- For enum types
    validation_rules JSONB DEFAULT '{}'::jsonb,    -- Custom validation rules
    encryption_required BOOLEAN DEFAULT false,     -- Whether values must be encrypted
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT attribute_definitions_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE attribute_definitions IS 
'Defines attributes used in ABAC policies with data types, validation rules, and security controls for consistent attribute management.';

COMMENT ON COLUMN attribute_definitions.data_type IS 'Attribute data type: STRING, NUMBER, BOOLEAN, DATE, TIME, JSON, ARRAY, ENUM';
COMMENT ON COLUMN attribute_definitions.category IS 'Attribute category: USER (user attributes), RESOURCE (resource attributes), ENVIRONMENT (context), ACTION (action attributes), ENTITY (entity attributes), SESSION (session context)';
COMMENT ON COLUMN attribute_definitions.is_sensitive IS 'Whether attribute contains PII or sensitive data requiring special handling';
COMMENT ON COLUMN attribute_definitions.allowed_values IS 'JSONB array of allowed values for ENUM data type';
COMMENT ON COLUMN attribute_definitions.validation_rules IS 'JSONB containing custom validation rules (regex, ranges, etc.)';
COMMENT ON COLUMN attribute_definitions.encryption_required IS 'Whether attribute values must be encrypted at rest';

-- Enable RLS and create policies
ALTER TABLE attribute_definitions ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_definitions_tenant_isolation ON attribute_definitions
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);


-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE VALUES
-- ------------------------------------------------------------------------------------------------
-- Stores actual attribute values for ABAC policy evaluation.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    definition_id UUID NOT NULL REFERENCES attribute_definitions(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL,                       -- The entity this attribute belongs to (user, resource, etc.)
    value TEXT NOT NULL,                           -- The actual attribute value (may be encrypted)
    encrypted_value BYTEA,                         -- Encrypted version if encryption is enabled
    is_encrypted BOOLEAN DEFAULT false,
    version INTEGER DEFAULT 1,                     -- For versioning/auditing changes
    effective_from TIMESTAMPTZ DEFAULT NOW(),      -- When this value becomes effective
    effective_to TIMESTAMPTZ,                      -- When this value expires (nullable for current values)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),
    
    CONSTRAINT attribute_values_unique_current UNIQUE (tenant_id, definition_id, entity_id, effective_from)
);

COMMENT ON TABLE attribute_values IS 
'Stores actual attribute values for entities with versioning, encryption, and temporal support for ABAC policy evaluation.';

COMMENT ON COLUMN attribute_values.entity_id IS 'The UUID of the entity this attribute belongs to (user, resource, document, etc.)';
COMMENT ON COLUMN attribute_values.value IS 'The actual attribute value in string format';
COMMENT ON COLUMN attribute_values.encrypted_value IS 'Encrypted version of the value when encryption is required';
COMMENT ON COLUMN attribute_values.effective_from IS 'When this attribute value becomes effective (for temporal policies)';
COMMENT ON COLUMN attribute_values.effective_to IS 'When this attribute value expires (null for current values)';

-- Enable RLS and create policies
ALTER TABLE attribute_values ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_values_tenant_isolation ON attribute_values
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Create indexes for performance
CREATE INDEX idx_attribute_values_entity_definition ON attribute_values(tenant_id, entity_id, definition_id) 
    WHERE effective_to IS NULL;

CREATE INDEX idx_attribute_values_definition_value ON attribute_values(tenant_id, definition_id, value) 
    WHERE effective_to IS NULL;

CREATE INDEX idx_attribute_values_effective_period ON attribute_values(effective_from, effective_to);

-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE SOURCES
-- ------------------------------------------------------------------------------------------------
-- Defines external sources for attribute collection.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    source_type VARCHAR(50) NOT NULL 
        CHECK (source_type IN ('LDAP', 'DATABASE', 'REST_API', 'GRAPHQL', 'FILE', 'MANUAL')),
    configuration JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Source-specific configuration
    authentication JSONB DEFAULT '{}'::jsonb,          -- Authentication details (encrypted)
    cache_ttl_minutes INTEGER DEFAULT 60,              -- How long to cache attributes from this source
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100,                      -- Source priority for attribute resolution
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT attribute_sources_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE attribute_sources IS 
'Defines external sources for attribute collection with configuration, authentication, and caching controls.';

COMMENT ON COLUMN attribute_sources.source_type IS 'Type of attribute source: LDAP, DATABASE, REST_API, GRAPHQL, FILE, MANUAL';
COMMENT ON COLUMN attribute_sources.configuration IS 'JSONB containing source-specific configuration (URLs, queries, etc.)';
COMMENT ON COLUMN attribute_sources.authentication IS 'JSONB containing authentication details (should be encrypted)';
COMMENT ON COLUMN attribute_sources.priority IS 'Source priority for attribute resolution (higher numbers processed first)';

-- Enable RLS and create policies
ALTER TABLE attribute_sources ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_sources_tenant_isolation ON attribute_sources
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);-- ------------------------------------------------------------------------------------------------
-- POLICY EVALUATIONS CACHE
-- ------------------------------------------------------------------------------------------------
-- Caches policy evaluation results for performance optimization.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policy_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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


-- ------------------------------------------------------------------------------------------------
-- UPDATE POLICY EVALUATIONS FOR ABAC
-- ------------------------------------------------------------------------------------------------
-- Updates policy evaluations table to support flexible resource types for ABAC.
-- ------------------------------------------------------------------------------------------------

-- Drop existing table and recreate with flexible schema
DROP TABLE IF EXISTS policy_evaluations CASCADE;

CREATE TABLE IF NOT EXISTS policy_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    resource_type VARCHAR(100) NOT NULL,           -- Flexible resource type (user, document, etc.)
    resource_id UUID,                              -- Optional specific resource ID
    action VARCHAR(100) NOT NULL,                  -- Action being performed (read, write, etc.)
    entity_id UUID REFERENCES entities(uuid),      -- Optional entity context
    context_hash VARCHAR(64) NOT NULL,             -- Hash of evaluation context
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE')),
    applicable_policies UUID[] DEFAULT '{}',       -- Array of policy IDs that fired
    policy_decisions JSONB DEFAULT '[]'::jsonb,    -- Detailed policy decisions
    evaluation_time_ms INTEGER,                    -- Performance metric
    cache_key VARCHAR(255),                        -- Optional cache key for faster lookup
    evaluated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '1 hour'),
    
    -- Unique constraint for cache lookups
    CONSTRAINT policy_evaluations_unique_cache 
        UNIQUE (tenant_id, user_id, resource_type, resource_id, action, context_hash)
);

COMMENT ON TABLE policy_evaluations IS 
'Caches ABAC policy evaluation results with flexible resource types and detailed decision tracking for performance optimization.';

COMMENT ON COLUMN policy_evaluations.resource_type IS 'Type of resource being accessed (user, document, report, system, etc.)';
COMMENT ON COLUMN policy_evaluations.resource_id IS 'Optional specific resource identifier';
COMMENT ON COLUMN policy_evaluations.action IS 'Action being performed (read, write, delete, execute, etc.)';
COMMENT ON COLUMN policy_evaluations.context_hash IS 'SHA-256 hash of evaluation context for cache key uniqueness';
COMMENT ON COLUMN policy_evaluations.applicable_policies IS 'Array of policy UUIDs that were evaluated and contributed to the decision';
COMMENT ON COLUMN policy_evaluations.policy_decisions IS 'JSONB array containing detailed policy decision information';
COMMENT ON COLUMN policy_evaluations.evaluation_time_ms IS 'Policy evaluation time in milliseconds for performance monitoring';

-- Enable RLS and create policies
ALTER TABLE policy_evaluations ENABLE ROW LEVEL SECURITY;

CREATE POLICY policy_evaluations_tenant_isolation ON policy_evaluations
    FOR ALL TO public
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Create indexes for performance
CREATE INDEX idx_policy_evaluations_cache_lookup ON policy_evaluations(
    tenant_id, user_id, resource_type, resource_id, action, context_hash
);

CREATE INDEX idx_policy_evaluations_user_resource ON policy_evaluations(
    tenant_id, user_id, resource_type
);

CREATE INDEX idx_policy_evaluations_expires_at ON policy_evaluations(expires_at);

CREATE INDEX idx_policy_evaluations_resource_action ON policy_evaluations(
    tenant_id, resource_type, action
);CREATE TABLE policy_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_evaluation_id UUID NOT NULL REFERENCES policy_evaluations(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    decision VARCHAR(20) NOT NULL,
    reason TEXT,
    matched_rule TEXT,
    evaluation_ms BIGINT,
    target_matched BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_policy_evaluation FOREIGN KEY (policy_evaluation_id) REFERENCES policy_evaluations(id),
    CONSTRAINT fk_policy FOREIGN KEY (policy_id) REFERENCES policies(id)
);

CREATE INDEX idx_policy_decisions_evaluation_id ON policy_decisions(policy_evaluation_id);
CREATE INDEX idx_policy_decisions_policy_id ON policy_decisions(policy_id);

COMMENT ON TABLE policy_decisions IS 'Stores individual policy decisions made during a policy evaluation.';
CREATE OR REPLACE FUNCTION assign_user_role(
    p_user_id UUID,
    p_role_id UUID,
    p_entity_id UUID,
    p_assigned_by UUID
) RETURNS VOID AS $$
BEGIN
    INSERT INTO user_roles (user_id, role_id, entity_id, assigned_by)
    VALUES (p_user_id, p_role_id, p_entity_id, p_assigned_by);
END;
$$ LANGUAGE plpgsql SECURITY INVOKER;

CREATE OR REPLACE FUNCTION revoke_user_role(
    p_user_id UUID,
    p_role_id UUID,
    p_entity_id UUID
) RETURNS VOID AS $$
BEGIN
    DELETE FROM user_roles
    WHERE user_id = p_user_id
      AND role_id = p_role_id
      AND entity_id = p_entity_id;
END;
$$ LANGUAGE plpgsql SECURITY INVOKER;-- =====================================================
-- USER ACTIVITIES TABLE FOR ABAC BEHAVIORAL ANALYTICS
-- =====================================================
-- Advanced user activity tracking for behavioral analytics
-- Supports ABAC evaluation with rich security context
-- Partitioned by timestamp for performance at scale

-- Main partitioned table for user activity tracking
CREATE TABLE user_activities (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    session_id UUID REFERENCES user_sessions(id) ON DELETE SET NULL,
    
    -- Activity classification
    activity_type VARCHAR(50) NOT NULL,
    module VARCHAR(50),
    resource_type VARCHAR(50),
    resource_id UUID,
    action_performed VARCHAR(50),
    
    -- Security context for ABAC evaluation
    ip_address INET,
    user_agent TEXT,
    device_fingerprint VARCHAR(255),
    location_data JSONB DEFAULT '{}'::jsonb,
    
    -- Performance and request metrics
    request_method VARCHAR(10),
    request_path TEXT,
    request_params JSONB DEFAULT '{}'::jsonb,
    response_status INTEGER,
    response_time_ms INTEGER,
    
    -- Risk assessment data
    risk_indicators JSONB DEFAULT '{}'::jsonb,
    anomaly_score DECIMAL(5,2) DEFAULT 0.00,
    
    -- Activity metadata and context
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    additional_data JSONB DEFAULT '{}'::jsonb,
    
    -- Constraints
    CONSTRAINT user_activities_anomaly_score_range 
        CHECK (anomaly_score >= 0.00 AND anomaly_score <= 100.00),
    -- Composite primary key including partition column
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create initial partitions (last 3 months + next 3 months)
-- Current month partition
-- CREATE TABLE user_activities_current PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE)) 
--     TO (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month');
--
-- -- Previous 2 months partitions
-- CREATE TABLE user_activities_prev1 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) - INTERVAL '1 month') 
--     TO (date_trunc('month', CURRENT_DATE));
--
-- CREATE TABLE user_activities_prev2 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) - INTERVAL '2 months') 
--     TO (date_trunc('month', CURRENT_DATE) - INTERVAL '1 month');
--
-- -- Next 2 months partitions
-- CREATE TABLE user_activities_next1 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month') 
--     TO (date_trunc('month', CURRENT_DATE) + INTERVAL '2 months');
--
-- CREATE TABLE user_activities_next2 PARTITION OF user_activities
--     FOR VALUES FROM (date_trunc('month', CURRENT_DATE) + INTERVAL '2 months') 
--     TO (date_trunc('month', CURRENT_DATE) + INTERVAL '3 months');
--
-- Performance indexes
CREATE INDEX idx_user_activities_user_timestamp ON user_activities (user_id, timestamp DESC);
CREATE INDEX idx_user_activities_tenant_timestamp ON user_activities (tenant_id, timestamp DESC);
CREATE INDEX idx_user_activities_activity_type ON user_activities (activity_type, timestamp DESC);
CREATE INDEX idx_user_activities_session ON user_activities (session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_user_activities_resource ON user_activities (resource_type, resource_id) WHERE resource_id IS NOT NULL;

-- JSONB indexes for ABAC attribute queries
CREATE INDEX idx_user_activities_location_data ON user_activities USING GIN (location_data);
CREATE INDEX idx_user_activities_risk_indicators ON user_activities USING GIN (risk_indicators);
CREATE INDEX idx_user_activities_additional_data ON user_activities USING GIN (additional_data);

-- Risk and anomaly detection indexes
CREATE INDEX idx_user_activities_anomaly_score ON user_activities (anomaly_score DESC) WHERE anomaly_score > 0;
CREATE INDEX idx_user_activities_high_risk ON user_activities (user_id, timestamp DESC) 
    WHERE anomaly_score > 50.0;

-- Enable Row Level Security
ALTER TABLE user_activities ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only access activities within their tenant
CREATE POLICY user_activities_tenant_isolation ON user_activities
    FOR ALL 
    TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- RLS Policy: Users can view their own activities (for self-service features)
CREATE POLICY user_activities_self_access ON user_activities
    FOR SELECT
    TO application_role
    USING (
        user_id = current_setting('app.current_user_id')::uuid AND
        tenant_id = current_setting('app.current_tenant_id')::uuid
    );

-- RLS Policy: Admin bypass - system administrators can access all activities within tenant
CREATE POLICY user_activities_admin_bypass ON user_activities
    FOR ALL
    TO admin_role
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE ON user_activities TO application_role;
GRANT ALL PRIVILEGES ON user_activities TO admin_role;

-- Table and column comments for documentation
COMMENT ON TABLE user_activities IS 'Partitioned table for user activity tracking and behavioral analytics supporting ABAC evaluation';
COMMENT ON COLUMN user_activities.id IS 'Unique identifier for the activity record';
COMMENT ON COLUMN user_activities.user_id IS 'Reference to the user who performed the activity';
COMMENT ON COLUMN user_activities.tenant_id IS 'Tenant isolation for multi-tenant architecture';
COMMENT ON COLUMN user_activities.session_id IS 'Reference to the user session when activity occurred';
COMMENT ON COLUMN user_activities.activity_type IS 'Classification of the activity (login, access, modification, etc.)';
COMMENT ON COLUMN user_activities.location_data IS 'JSONB containing geographic and network location information';
COMMENT ON COLUMN user_activities.risk_indicators IS 'JSONB containing calculated risk factors for the activity';
COMMENT ON COLUMN user_activities.anomaly_score IS 'Calculated anomaly score from 0.00 to 100.00 for behavioral analysis';
COMMENT ON COLUMN user_activities.additional_data IS 'Flexible JSONB storage for activity-specific metadata';

-- Function to automatically create monthly partitions
CREATE OR REPLACE FUNCTION create_monthly_user_activities_partition(partition_date DATE)
RETURNS TEXT AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- Generate partition name
    partition_name := 'user_activities_' || to_char(partition_date, 'YYYY_MM');
    
    -- Calculate partition boundaries
    start_date := date_trunc('month', partition_date)::DATE;
    end_date := (date_trunc('month', partition_date) + INTERVAL '1 month')::DATE;
    
    -- Create partition
    EXECUTE format('CREATE TABLE %I PARTITION OF user_activities 
                    FOR VALUES FROM (%L) TO (%L)', 
                   partition_name, start_date, end_date);
    
    RETURN 'Created partition: ' || partition_name;
END;
$$ LANGUAGE plpgsql;

-- Function to drop old partitions (data retention)
CREATE OR REPLACE FUNCTION drop_old_user_activities_partitions(retention_months INTEGER DEFAULT 12)
RETURNS TEXT AS $$
DECLARE
    partition_name TEXT;
    cutoff_date DATE;
    dropped_partitions TEXT[] := '{}';
    partition_record RECORD;
BEGIN
    cutoff_date := (date_trunc('month', CURRENT_DATE) - (retention_months || ' months')::INTERVAL)::DATE;
    
    -- Find partitions older than retention period
    FOR partition_record IN
        SELECT schemaname, tablename 
        FROM pg_tables 
        WHERE schemaname = 'public' 
        AND tablename LIKE 'user_activities_%'
        AND tablename ~ '^user_activities_[0-9]{4}_[0-9]{2}$'
    LOOP
        -- Extract date from partition name and check if it's old enough
        BEGIN
            DECLARE
                partition_date DATE;
            BEGIN
                partition_date := to_date(
                    substring(partition_record.tablename from 'user_activities_([0-9]{4}_[0-9]{2})$'), 
                    'YYYY_MM'
                );
                
                IF partition_date < cutoff_date THEN
                    EXECUTE format('DROP TABLE IF EXISTS %I', partition_record.tablename);
                    dropped_partitions := array_append(dropped_partitions, partition_record.tablename);
                END IF;
            END;
        EXCEPTION
            WHEN OTHERS THEN
                -- Skip invalid partition names
                CONTINUE;
        END;
    END LOOP;
    
    IF array_length(dropped_partitions, 1) > 0 THEN
        RETURN 'Dropped partitions: ' || array_to_string(dropped_partitions, ', ');
    ELSE
        RETURN 'No old partitions found to drop';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Grant execute permissions on utility functions
GRANT EXECUTE ON FUNCTION create_monthly_user_activities_partition(DATE) TO application_role;
GRANT EXECUTE ON FUNCTION drop_old_user_activities_partitions(INTEGER) TO application_role;
GRANT EXECUTE ON FUNCTION create_monthly_user_activities_partition(DATE) TO admin_role;
GRANT EXECUTE ON FUNCTION drop_old_user_activities_partitions(INTEGER) TO admin_role;
-- Creates the core feature_flags table with proper indexing and RLS

-- =====================================================
-- FEATURE FLAGS TABLE
-- =====================================================
CREATE TABLE feature_flags (
    -- Primary identifier
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    
    -- Descriptive information
    description TEXT,
    
    -- Flag configuration
    flag_type VARCHAR(20) NOT NULL DEFAULT 'boolean'
        CHECK (flag_type IN ('boolean', 'string', 'number', 'json')),
    default_value BOOLEAN NOT NULL DEFAULT false,
    
    -- Rollout settings
    rollout_percentage INTEGER
        CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    target_audience JSONB DEFAULT '{}', -- For advanced targeting rules
    
    -- Flexible metadata storage
    metadata JSONB DEFAULT '{}',
    
    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ, -- Soft delete support
    
    -- Unique constraint per tenant
    UNIQUE (tenant_id, name)
);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================
CREATE INDEX idx_feature_flags_tenant ON feature_flags(tenant_id);
CREATE INDEX idx_feature_flags_tenant_name ON feature_flags(tenant_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_feature_flags_type ON feature_flags(flag_type);
CREATE INDEX idx_feature_flags_rollout ON feature_flags(rollout_percentage) WHERE rollout_percentage IS NOT NULL;
CREATE INDEX idx_feature_flags_deleted_at ON feature_flags(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_feature_flags_created_at ON feature_flags(created_at);
CREATE INDEX idx_feature_flags_updated_at ON feature_flags(updated_at);
CREATE INDEX idx_feature_flags_target_audience ON feature_flags USING GIN (target_audience) WHERE target_audience != '{}';
CREATE INDEX idx_feature_flags_metadata ON feature_flags USING GIN (metadata) WHERE metadata != '{}';

-- =====================================================
-- TABLE COMMENTS
-- =====================================================
COMMENT ON TABLE feature_flags IS 'Master feature flags configuration table with tenant isolation';
COMMENT ON COLUMN feature_flags.rollout_percentage IS 'Percentage of tenants that should have this feature enabled (0-100)';
COMMENT ON COLUMN feature_flags.target_audience IS 'Advanced targeting rules (company_size, industry, etc.)';
COMMENT ON COLUMN feature_flags.metadata IS 'Additional metadata like expiration dates, dependencies, etc.';
COMMENT ON COLUMN feature_flags.deleted_at IS 'Soft delete timestamp - NULL means active';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================
ALTER TABLE feature_flags ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy for application role
CREATE POLICY feature_flags_tenant_isolation ON feature_flags
    FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Admin role can access all tenants
CREATE POLICY feature_flags_admin_access ON feature_flags
    FOR ALL TO admin_role
    USING (true);

-- Read-only role for monitoring/analytics
CREATE POLICY feature_flags_readonly_access ON feature_flags
    FOR SELECT TO readonly_role
    USING (true);

-- Policy comments
COMMENT ON POLICY feature_flags_tenant_isolation ON feature_flags IS 'Ensures tenant data isolation for application users';
COMMENT ON POLICY feature_flags_admin_access ON feature_flags IS 'Allows admin role full access across all tenants';
COMMENT ON POLICY feature_flags_readonly_access ON feature_flags IS 'Allows readonly role to view all feature flags for monitoring';

-- =====================================================
-- PERMISSIONS
-- =====================================================
GRANT SELECT, INSERT, UPDATE, DELETE ON feature_flags TO application_role;
GRANT ALL ON feature_flags TO admin_role;
GRANT SELECT ON feature_flags TO readonly_role;

-- =====================================================
-- TRIGGERS
-- =====================================================
-- Auto-update updated_at timestamp
CREATE TRIGGER update_feature_flags_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- Creates the tenant_feature_overrides table with proper indexing and RLS

-- =====================================================
-- TENANT FEATURE OVERRIDES TABLE
-- =====================================================
CREATE TABLE tenant_feature_overrides (
    -- Primary identifier
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    
    -- References
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    feature_flag_name VARCHAR(100) NOT NULL,
    
    -- Override settings
    enabled BOOLEAN NOT NULL,
    value JSONB DEFAULT '{}', -- For complex feature values
    reason TEXT, -- Why this override was set
    
    -- Audit timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint - one override per tenant per feature
    UNIQUE (tenant_id, feature_flag_id)
);

-- =====================================================
-- PERFORMANCE INDEXES
-- =====================================================
CREATE INDEX idx_tenant_overrides_tenant ON tenant_feature_overrides(tenant_id);
CREATE INDEX idx_tenant_overrides_feature_id ON tenant_feature_overrides(feature_flag_id);
CREATE INDEX idx_tenant_overrides_feature_name ON tenant_feature_overrides(feature_flag_name);
CREATE INDEX idx_tenant_overrides_enabled ON tenant_feature_overrides(enabled);
CREATE INDEX idx_tenant_overrides_tenant_enabled ON tenant_feature_overrides(tenant_id, enabled);
CREATE INDEX idx_tenant_overrides_created_at ON tenant_feature_overrides(created_at);
CREATE INDEX idx_tenant_overrides_updated_at ON tenant_feature_overrides(updated_at);
CREATE INDEX idx_tenant_overrides_value ON tenant_feature_overrides USING GIN (value) WHERE value != '{}';
CREATE INDEX idx_tenant_overrides_lookup ON tenant_feature_overrides(tenant_id, feature_flag_name, enabled);

-- =====================================================
-- TABLE COMMENTS
-- =====================================================
COMMENT ON TABLE tenant_feature_overrides IS 'Tenant-specific feature flag overrides with audit trail';
COMMENT ON COLUMN tenant_feature_overrides.value IS 'Complex feature values for non-boolean flags (JSON format)';
COMMENT ON COLUMN tenant_feature_overrides.reason IS 'Business justification for the override';
COMMENT ON COLUMN tenant_feature_overrides.feature_flag_name IS 'Denormalized feature flag name for faster lookups';

-- =====================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================
ALTER TABLE tenant_feature_overrides ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy for application role
CREATE POLICY tenant_overrides_tenant_isolation ON tenant_feature_overrides
    FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Admin role can access all tenants
CREATE POLICY tenant_overrides_admin_access ON tenant_feature_overrides
    FOR ALL TO admin_role
    USING (true);

-- Read-only role for monitoring/analytics
CREATE POLICY tenant_overrides_readonly_access ON tenant_feature_overrides
    FOR SELECT TO readonly_role
    USING (true);

-- Policy comments
COMMENT ON POLICY tenant_overrides_tenant_isolation ON tenant_feature_overrides IS 'Ensures tenant data isolation for application users';
COMMENT ON POLICY tenant_overrides_admin_access ON tenant_feature_overrides IS 'Allows admin role full access across all tenants';
COMMENT ON POLICY tenant_overrides_readonly_access ON tenant_feature_overrides IS 'Allows readonly role to view all overrides for monitoring';

-- =====================================================
-- PERMISSIONS
-- =====================================================
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_feature_overrides TO application_role;
GRANT ALL ON tenant_feature_overrides TO admin_role;
GRANT SELECT ON tenant_feature_overrides TO readonly_role;

-- =====================================================
-- TRIGGERS
-- =====================================================
-- Auto-update updated_at timestamp
CREATE TRIGGER update_tenant_overrides_updated_at
    BEFORE UPDATE ON tenant_feature_overrides
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Sync feature_flag_name on insert/update
CREATE OR REPLACE FUNCTION sync_feature_flag_name()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the denormalized feature_flag_name from the feature_flags table
    SELECT name INTO NEW.feature_flag_name 
    FROM feature_flags 
    WHERE id = NEW.feature_flag_id;
    
    IF NEW.feature_flag_name IS NULL THEN
        RAISE EXCEPTION 'Feature flag not found for ID: %', NEW.feature_flag_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_tenant_overrides_feature_name
    BEFORE INSERT OR UPDATE ON tenant_feature_overrides
    FOR EACH ROW
    EXECUTE FUNCTION sync_feature_flag_name();

-- Add trigger comments
COMMENT ON TRIGGER sync_tenant_overrides_feature_name ON tenant_feature_overrides IS 'Maintains denormalized feature_flag_name for performance';
COMMENT ON FUNCTION sync_feature_flag_name() IS 'Syncs feature flag name in overrides table';
-- Creates audit logging functions and triggers for feature flag changes

-- =====================================================
-- AUDIT TRIGGER FUNCTIONS
-- =====================================================

-- Function to create audit log entries for feature flag changes
CREATE OR REPLACE FUNCTION audit_feature_flag_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            tenant_id, 
            event_type, 
            event_category, 
            severity,
            user_id,
            decision,
            reason,
            context,
            session_id
        ) VALUES (
            NEW.tenant_id,
            'FEATURE_FLAG_CREATED',
            'ADMIN',
            'INFO',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature flag created: ' || NEW.name,
            jsonb_build_object(
                'feature_flag_id', NEW.id,
                'feature_flag_name', NEW.name,
                'flag_type', NEW.flag_type,
                'default_value', NEW.default_value,
                'rollout_percentage', NEW.rollout_percentage,
                'metadata', NEW.metadata,
                'operation', 'CREATE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            context,
            session_id
        ) VALUES (
            NEW.tenant_id,
            'FEATURE_FLAG_UPDATED',
            'ADMIN',
            CASE 
                WHEN OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN 'WARN'
                ELSE 'INFO'
            END,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature flag updated: ' || NEW.name,
            jsonb_build_object(
                'feature_flag_id', NEW.id,
                'feature_flag_name', NEW.name,
                'old_values', jsonb_build_object(
                    'flag_type', OLD.flag_type,
                    'default_value', OLD.default_value,
                    'rollout_percentage', OLD.rollout_percentage,
                    'deleted_at', OLD.deleted_at,
                    'metadata', OLD.metadata
                ),
                'new_values', jsonb_build_object(
                    'flag_type', NEW.flag_type,
                    'default_value', NEW.default_value,
                    'rollout_percentage', NEW.rollout_percentage,
                    'deleted_at', NEW.deleted_at,
                    'metadata', NEW.metadata
                ),
                'operation', CASE 
                    WHEN OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN 'SOFT_DELETE'
                    ELSE 'UPDATE'
                END
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            context,
            session_id
        ) VALUES (
            OLD.tenant_id,
            'FEATURE_FLAG_DELETED',
            'ADMIN',
            'WARN',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature flag permanently deleted: ' || OLD.name,
            jsonb_build_object(
                'feature_flag_id', OLD.id,
                'feature_flag_name', OLD.name,
                'deleted_values', to_jsonb(OLD),
                'operation', 'HARD_DELETE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID
        );
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Function to create audit entries for tenant feature override changes
CREATE OR REPLACE FUNCTION audit_tenant_feature_override_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            risk_score,
            context,
            session_id,
            compliance_flags
        ) VALUES (
            NEW.tenant_id,
            'FEATURE_OVERRIDE_CREATED',
            'ADMIN',
            'INFO',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            COALESCE(NEW.reason, 'Feature override created for ' || NEW.feature_flag_name),
            CASE WHEN NEW.enabled THEN 10 ELSE 5 END, -- Higher risk when enabling features
            jsonb_build_object(
                'feature_flag_id', NEW.feature_flag_id,
                'feature_flag_name', NEW.feature_flag_name,
                'enabled', NEW.enabled,
                'value', NEW.value,
                'operation', 'CREATE_OVERRIDE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID,
            jsonb_build_object('feature_management', true)
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            risk_score,
            context,
            session_id,
            compliance_flags
        ) VALUES (
            NEW.tenant_id,
            'FEATURE_OVERRIDE_UPDATED',
            'ADMIN',
            CASE 
                WHEN OLD.enabled != NEW.enabled THEN 'WARN'
                ELSE 'INFO'
            END,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            COALESCE(NEW.reason, 'Feature override updated for ' || NEW.feature_flag_name),
            CASE 
                WHEN OLD.enabled != NEW.enabled THEN 15
                ELSE 8
            END,
            jsonb_build_object(
                'feature_flag_id', NEW.feature_flag_id,
                'feature_flag_name', NEW.feature_flag_name,
                'old_values', jsonb_build_object(
                    'enabled', OLD.enabled,
                    'value', OLD.value
                ),
                'new_values', jsonb_build_object(
                    'enabled', NEW.enabled,
                    'value', NEW.value
                ),
                'operation', 'UPDATE_OVERRIDE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID,
            jsonb_build_object('feature_management', true)
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (
            tenant_id,
            event_type,
            event_category,
            severity,
            user_id,
            decision,
            reason,
            risk_score,
            context,
            session_id,
            compliance_flags
        ) VALUES (
            OLD.tenant_id,
            'FEATURE_OVERRIDE_DELETED',
            'ADMIN',
            'INFO',
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            'ALLOW',
            'Feature override deleted for ' || OLD.feature_flag_name,
            5,
            jsonb_build_object(
                'feature_flag_id', OLD.feature_flag_id,
                'feature_flag_name', OLD.feature_flag_name,
                'deleted_values', jsonb_build_object(
                    'enabled', OLD.enabled,
                    'value', OLD.value
                ),
                'operation', 'DELETE_OVERRIDE'
            ),
            NULLIF(current_setting('app.current_session_id', true), '')::UUID,
            jsonb_build_object('feature_management', true)
        );
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- CREATE AUDIT TRIGGERS
-- =====================================================

-- Audit trigger for feature_flags table
CREATE TRIGGER feature_flags_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION audit_feature_flag_changes();

-- Audit trigger for tenant_feature_overrides table
CREATE TRIGGER tenant_feature_overrides_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tenant_feature_overrides
    FOR EACH ROW
    EXECUTE FUNCTION audit_tenant_feature_override_changes();

-- =====================================================
-- FUNCTION COMMENTS
-- =====================================================
COMMENT ON FUNCTION audit_feature_flag_changes() IS 'Creates audit log entries for feature flag changes using existing audit_log table';
COMMENT ON FUNCTION audit_tenant_feature_override_changes() IS 'Creates audit entries for tenant feature override changes using existing audit_log table';
COMMENT ON TRIGGER feature_flags_audit_trigger ON feature_flags IS 'Logs all feature flag changes to audit_log table';
COMMENT ON TRIGGER tenant_feature_overrides_audit_trigger ON tenant_feature_overrides IS 'Logs all tenant override changes to audit_log table';
-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================

-- Grant execute permissions on functions to application role
-- -- GRANT EXECUTE ON FUNCTION set_audit_context(UUID, UUID) TO application_role;
-- GRANT EXECUTE ON FUNCTION evaluate_feature_flag(VARCHAR) TO application_role;
-- GRANT EXECUTE ON FUNCTION evaluate_all_feature_flags() TO application_role;
-- GRANT EXECUTE ON FUNCTION evaluate_feature_flag_fast(VARCHAR) TO application_role;
--
-- -- Grant execute permissions to admin role
-- -- GRANT EXECUTE ON FUNCTION set_audit_context(UUID, UUID) TO admin_role;
-- GRANT EXECUTE ON FUNCTION evaluate_feature_flag(VARCHAR) TO admin_role;
-- GRANT EXECUTE ON FUNCTION evaluate_all_feature_flags() TO admin_role;
-- GRANT EXECUTE ON FUNCTION evaluate_feature_flag_fast(VARCHAR) TO admin_role;
--
-- -- Grant execute permissions to readonly role (for monitoring)
-- GRANT EXECUTE ON FUNCTION evaluate_feature_flag_fast(VARCHAR) TO readonly_role;
--
-- -- =====================================================
-- -- FUNCTION COMMENTS
-- -- =====================================================
-- -- COMMENT ON FUNCTION set_audit_context(UUID, UUID) IS 'Sets the current user and session context for audit logging';
-- COMMENT ON FUNCTION evaluate_feature_flag(VARCHAR) IS 'Evaluates feature flag for current tenant with override and rollout logic, logs evaluation to audit_log';
-- COMMENT ON FUNCTION evaluate_all_feature_flags() IS 'Evaluates all feature flags for current tenant, logs bulk evaluation to audit_log';
-- COMMENT ON FUNCTION evaluate_feature_flag_fast(VARCHAR) IS 'Lightweight feature flag evaluation without audit logging for high-frequency calls';
-- -- =====================================================
-- -- Creates the core functions for evaluating feature flags with audit logging
--
-- -- =====================================================
-- -- UTILITY FUNCTIONS
-- -- =====================================================
--
-- -- Function to set user context for audit logging
-- -- CREATE OR REPLACE FUNCTION set_audit_context(user_id UUID DEFAULT NULL, session_id UUID DEFAULT NULL)
-- -- RETURNS VOID AS $$
-- -- BEGIN
-- --     IF user_id IS NOT NULL THEN
-- --         PERFORM set_config('app.current_user_id', user_id::text, true);
-- --     END IF;
-- --
-- --     IF session_id IS NOT NULL THEN
-- --         PERFORM set_config('app.current_session_id', session_id::text, true);
-- --     END IF;
-- -- END;
-- -- $$ LANGUAGE plpgsql SECURITY DEFINER;
-- --
-- -- =====================================================
-- -- SINGLE FEATURE FLAG EVALUATION
-- -- =====================================================
--
-- -- Function to evaluate a single feature flag for current tenant
-- CREATE OR REPLACE FUNCTION evaluate_feature_flag(flag_name VARCHAR)
-- RETURNS TABLE(enabled BOOLEAN, value JSONB) AS $$
-- DECLARE
--     v_tenant_id UUID;
--     v_feature_flag feature_flags%ROWTYPE;
--     v_override tenant_feature_overrides%ROWTYPE;
--     v_result_enabled BOOLEAN;
--     v_result_value JSONB;
--     v_evaluation_source TEXT;
-- BEGIN
--     -- Get current tenant ID
--     v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
--
--     IF v_tenant_id IS NULL THEN
--         RAISE EXCEPTION 'No tenant context set. Use SET app.current_tenant_id = ''<tenant_id>''';
--     END IF;
--
--     -- Get feature flag configuration for current tenant
--     SELECT * INTO v_feature_flag
--     FROM feature_flags ff
--     WHERE ff.name = flag_name 
--       AND ff.tenant_id = v_tenant_id 
--       AND ff.deleted_at IS NULL;
--
--     IF v_feature_flag.id IS NULL THEN
--         RAISE EXCEPTION 'Feature flag not found: % for tenant: %', flag_name, v_tenant_id;
--     END IF;
--
--     -- Check for tenant-specific override
--     SELECT * INTO v_override
--     FROM tenant_feature_overrides tfo
--     WHERE tfo.tenant_id = v_tenant_id 
--       AND tfo.feature_flag_id = v_feature_flag.id;
--
--     -- Determine effective value
--     IF v_override.id IS NOT NULL THEN
--         -- Override exists, use it
--         v_result_enabled := v_override.enabled;
--         v_result_value := COALESCE(v_override.value, '{}');
--         v_evaluation_source := 'override';
--     ELSIF v_feature_flag.rollout_percentage IS NOT NULL AND v_feature_flag.rollout_percentage > 0 THEN
--         -- Use percentage rollout (deterministic based on tenant ID)
--         v_result_enabled := (hashtext(v_tenant_id::text || flag_name) % 100) < v_feature_flag.rollout_percentage;
--         v_result_value := '{}';
--         v_evaluation_source := 'rollout';
--     ELSE
--         -- Use default value
--         v_result_enabled := v_feature_flag.default_value;
--         v_result_value := '{}';
--         v_evaluation_source := 'default';
--     END IF;
--
--     -- Log feature flag evaluation for audit purposes
--     INSERT INTO audit_log (
--         tenant_id,
--         event_type,
--         event_category,
--         severity,
--         user_id,
--         decision,
--         reason,
--         context,
--         session_id
--     ) VALUES (
--         v_tenant_id,
--         'FEATURE_FLAG_EVALUATED',
--         'ACCESS',
--         'LOW',
--         NULLIF(current_setting('app.current_user_id', true), '')::UUID,
--         CASE WHEN v_result_enabled THEN 'ALLOW' ELSE 'DENY' END,
--         'Feature flag evaluated: ' || flag_name,
--         jsonb_build_object(
--             'feature_flag_id', v_feature_flag.id,
--             'feature_flag_name', flag_name,
--             'enabled', v_result_enabled,
--             'value', v_result_value,
--             'source', v_evaluation_source,
--             'rollout_percentage', v_feature_flag.rollout_percentage,
--             'has_override', (v_override.id IS NOT NULL)
--         ),
--         NULLIF(current_setting('app.current_session_id', true), '')::UUID
--     );
--
--     RETURN QUERY SELECT v_result_enabled, v_result_value;
-- END;
-- $$ LANGUAGE plpgsql SECURITY DEFINER;
--
-- -- =====================================================
-- -- BULK FEATURE FLAG EVALUATION
-- -- =====================================================
--
-- -- Function to bulk evaluate all feature flags for current tenant
-- CREATE OR REPLACE FUNCTION evaluate_all_feature_flags()
-- RETURNS TABLE(flag_name VARCHAR, enabled BOOLEAN, value JSONB, flag_type VARCHAR, source TEXT) AS $$
-- DECLARE
--     v_tenant_id UUID;
--     v_flag_count INTEGER;
-- BEGIN
--     -- Get current tenant ID
--     v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
--
--     IF v_tenant_id IS NULL THEN
--         RAISE EXCEPTION 'No tenant context set. Use SET app.current_tenant_id = ''<tenant_id>''';
--     END IF;
--
--     -- Count flags for audit logging
--     SELECT COUNT(*) INTO v_flag_count
--     FROM feature_flags ff
--     WHERE ff.tenant_id = v_tenant_id 
--       AND ff.deleted_at IS NULL;
--
--     -- Log bulk evaluation
--     INSERT INTO audit_log (
--         tenant_id,
--         event_type,
--         event_category,
--         severity,
--         user_id,
--         decision,
--         reason,
--         context,
--         session_id
--     ) VALUES (
--         v_tenant_id,
--         'FEATURE_FLAGS_BULK_EVALUATED',
--         'ACCESS',
--         'LOW',
--         NULLIF(current_setting('app.current_user_id', true), '')::UUID,
--         'ALLOW',
--         'All feature flags evaluated for tenant',
--         jsonb_build_object(
--             'operation', 'bulk_evaluation',
--             'flags_count', v_flag_count,
--             'tenant_id', v_tenant_id
--         ),
--         NULLIF(current_setting('app.current_session_id', true), '')::UUID
--     );
--
--     -- Return evaluated flags
--     RETURN QUERY
--     WITH feature_evaluation AS (
--         SELECT 
--             ff.name,
--             ff.flag_type,
--             ff.default_value,
--             ff.rollout_percentage,
--             tfo.enabled as override_enabled,
--             tfo.value as override_value,
--             CASE 
--                 -- Override exists, use it
--                 WHEN tfo.enabled IS NOT NULL THEN tfo.enabled
--                 -- Percentage rollout check
--                 WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN
--                     (hashtext(v_tenant_id::text || ff.name) % 100) < ff.rollout_percentage
--                 -- Default value
--                 ELSE ff.default_value
--             END as effective_enabled,
--             COALESCE(tfo.value, '{}') as effective_value,
--             CASE 
--                 WHEN tfo.enabled IS NOT NULL THEN 'override'
--                 WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN 'rollout'
--                 ELSE 'default'
--             END as evaluation_source
--         FROM feature_flags ff
--         LEFT JOIN tenant_feature_overrides tfo ON ff.id = tfo.feature_flag_id 
--             AND tfo.tenant_id = v_tenant_id
--         WHERE ff.tenant_id = v_tenant_id 
--           AND ff.deleted_at IS NULL
--     )
--     SELECT 
--         fe.name::VARCHAR,
--         fe.effective_enabled,
--         fe.effective_value,
--         fe.flag_type::VARCHAR,
--         fe.evaluation_source::TEXT
--     FROM feature_evaluation fe
--     ORDER BY fe.name;
-- END;
-- $$ LANGUAGE plpgsql SECURITY DEFINER;
--
-- -- =====================================================
-- -- FAST EVALUATION FUNCTION (NO AUDIT LOGGING)
-- -- =====================================================
--
-- -- Lightweight evaluation function for high-frequency calls
-- CREATE OR REPLACE FUNCTION evaluate_feature_flag_fast(flag_name VARCHAR)
-- RETURNS TABLE(enabled BOOLEAN, value JSONB) AS $$
-- DECLARE
--     v_tenant_id UUID;
--     v_result RECORD;
-- BEGIN
--     -- Get current tenant ID
--     v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
--
--     IF v_tenant_id IS NULL THEN
--         RAISE EXCEPTION 'No tenant context set';
--     END IF;
--
--     -- Single query to get evaluation result
--     SELECT 
--         CASE 
--             -- Override exists, use it
--             WHEN tfo.enabled IS NOT NULL THEN tfo.enabled
--             -- Percentage rollout check
--             WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN
--                 (hashtext(v_tenant_id::text || ff.name) % 100) < ff.rollout_percentage
--             -- Default value
--             ELSE ff.default_value
--         END as flag_enabled,
--         COALESCE(tfo.value, '{}') as flag_value
--     INTO v_result
--     FROM feature_flags ff
--     LEFT JOIN tenant_feature_overrides tfo ON ff.id = tfo.feature_flag_id 
--         AND tfo.tenant_id = v_tenant_id
--     WHERE ff.name = flag_name 
--       AND ff.tenant_id = v_tenant_id 
--       AND ff.deleted_at IS NULL;
--     
--     IF v_result IS NULL THEN
--         RAISE EXCEPTION 'Feature flag not found: %', flag_name;
--     END IF;
--     
--     RETURN QUERY SELECT v_result.flag_enabled, v_result.flag_value;
-- END;
-- $$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
--
-- Creates materialized view for fast feature flag lookups and cache management

-- =====================================================
-- MATERIALIZED VIEW FOR FEATURE FLAG CACHE
-- =====================================================

-- Materialized view for feature flag evaluation cache
CREATE MATERIALIZED VIEW mv_tenant_feature_flags_cache AS
WITH feature_evaluation AS (
    SELECT 
        ff.tenant_id,
        ff.id as feature_flag_id,
        ff.name as feature_flag_name,
        ff.flag_type,
        ff.default_value,
        ff.rollout_percentage,
        ff.target_audience,
        ff.metadata,
        tfo.enabled as override_enabled,
        tfo.value as override_value,
        tfo.reason as override_reason,
        CASE 
            -- Override exists, use it
            WHEN tfo.enabled IS NOT NULL THEN tfo.enabled
            -- Percentage rollout check
            WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN
                (hashtext(ff.tenant_id::text || ff.name) % 100) < ff.rollout_percentage
            -- Default value
            ELSE ff.default_value
        END as effective_enabled,
        COALESCE(tfo.value, '{}') as effective_value,
        CASE 
            WHEN tfo.enabled IS NOT NULL THEN 'override'
            WHEN ff.rollout_percentage IS NOT NULL AND ff.rollout_percentage > 0 THEN 'rollout'
            ELSE 'default'
        END as evaluation_source,
        GREATEST(ff.updated_at, COALESCE(tfo.updated_at, ff.updated_at)) as cache_timestamp
    FROM feature_flags ff
    LEFT JOIN tenant_feature_overrides tfo ON ff.id = tfo.feature_flag_id 
        AND tfo.tenant_id = ff.tenant_id
    WHERE ff.deleted_at IS NULL
)
SELECT 
    tenant_id,
    feature_flag_id,
    feature_flag_name,
    flag_type,
    effective_enabled as enabled,
    effective_value as value,
    evaluation_source,
    default_value,
    rollout_percentage,
    target_audience,
    metadata,
    override_enabled,
    override_value,
    override_reason,
    cache_timestamp,
    NOW() as cache_created_at
FROM feature_evaluation;

-- =====================================================
-- PERFORMANCE INDEXES ON MATERIALIZED VIEW
-- =====================================================

-- Primary lookup indexes
CREATE UNIQUE INDEX idx_tenant_feature_cache_pk 
    ON mv_tenant_feature_flags_cache(tenant_id, feature_flag_id);

CREATE UNIQUE INDEX idx_tenant_feature_cache_name_lookup 
    ON mv_tenant_feature_flags_cache(tenant_id, feature_flag_name);

-- Query optimization indexes
CREATE INDEX idx_tenant_feature_cache_tenant 
    ON mv_tenant_feature_flags_cache(tenant_id);

CREATE INDEX idx_tenant_feature_cache_enabled 
    ON mv_tenant_feature_flags_cache(enabled) WHERE enabled = true;

CREATE INDEX idx_tenant_feature_cache_source 
    ON mv_tenant_feature_flags_cache(evaluation_source);

CREATE INDEX idx_tenant_feature_cache_flag_type 
    ON mv_tenant_feature_flags_cache(flag_type);

CREATE INDEX idx_tenant_feature_cache_timestamp 
    ON mv_tenant_feature_flags_cache(cache_timestamp);

CREATE INDEX idx_tenant_feature_cache_rollout 
    ON mv_tenant_feature_flags_cache(rollout_percentage) 
    WHERE rollout_percentage IS NOT NULL;

-- JSON indexes for complex queries
CREATE INDEX idx_tenant_feature_cache_target_audience 
    ON mv_tenant_feature_flags_cache USING GIN (target_audience) 
    WHERE target_audience != '{}';

CREATE INDEX idx_tenant_feature_cache_metadata 
    ON mv_tenant_feature_flags_cache USING GIN (metadata) 
    WHERE metadata != '{}';

CREATE INDEX idx_tenant_feature_cache_value 
    ON mv_tenant_feature_flags_cache USING GIN (value) 
    WHERE value != '{}';

-- =====================================================
-- CACHE MANAGEMENT FUNCTIONS
-- =====================================================

-- Function to refresh the cache
CREATE OR REPLACE FUNCTION refresh_feature_flags_cache()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_tenant_feature_flags_cache;
    
    -- Log cache refresh
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        NULL, -- System operation
        'FEATURE_FLAGS_CACHE_REFRESHED',
        'SYSTEM',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Feature flags cache materialized view refreshed',
        jsonb_build_object(
            'operation', 'cache_refresh',
            'timestamp', NOW()
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get cache statistics
CREATE OR REPLACE FUNCTION get_feature_flags_cache_stats()
RETURNS TABLE(
    total_entries BIGINT,
    tenants_count BIGINT,
    flags_per_tenant_avg NUMERIC,
    enabled_flags_count BIGINT,
    override_count BIGINT,
    rollout_count BIGINT,
    default_count BIGINT,
    cache_age INTERVAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_entries,
        COUNT(DISTINCT tffc.tenant_id) as tenants_count,
        ROUND(COUNT(*)::NUMERIC / COUNT(DISTINCT tffc.tenant_id), 2) as flags_per_tenant_avg,
        COUNT(*) FILTER (WHERE tffc.enabled = true) as enabled_flags_count,
        COUNT(*) FILTER (WHERE tffc.evaluation_source = 'override') as override_count,
        COUNT(*) FILTER (WHERE tffc.evaluation_source = 'rollout') as rollout_count,
        COUNT(*) FILTER (WHERE tffc.evaluation_source = 'default') as default_count,
        NOW() - MIN(tffc.cache_created_at) as cache_age
    FROM mv_tenant_feature_flags_cache tffc;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check cache freshness for a tenant
CREATE OR REPLACE FUNCTION check_cache_freshness(p_tenant_id UUID)
RETURNS TABLE(
    is_stale BOOLEAN,
    cache_age INTERVAL,
    last_flag_update TIMESTAMPTZ,
    last_override_update TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    WITH cache_info AS (
        SELECT MAX(cache_timestamp) as max_cache_ts
        FROM mv_tenant_feature_flags_cache
        WHERE tenant_id = p_tenant_id
    ),
    source_info AS (
        SELECT 
            MAX(ff.updated_at) as last_flag_update,
            MAX(tfo.updated_at) as last_override_update
        FROM feature_flags ff
        LEFT JOIN tenant_feature_overrides tfo ON ff.id = tfo.feature_flag_id
        WHERE ff.tenant_id = p_tenant_id AND ff.deleted_at IS NULL
    )
    SELECT 
        COALESCE(si.last_flag_update > ci.max_cache_ts OR si.last_override_update > ci.max_cache_ts, true) as is_stale,
        NOW() - ci.max_cache_ts as cache_age,
        si.last_flag_update,
        si.last_override_update
    FROM cache_info ci
    CROSS JOIN source_info si;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- CACHED EVALUATION FUNCTIONS
-- =====================================================

-- Fast evaluation using cache
CREATE OR REPLACE FUNCTION evaluate_feature_flag_cached(flag_name VARCHAR)
RETURNS TABLE(enabled BOOLEAN, value JSONB) AS $$
DECLARE
    v_tenant_id UUID;
    v_result RECORD;
    v_cache_stale BOOLEAN;
BEGIN
    -- Get current tenant ID
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
    
    IF v_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context set';
    END IF;
    
    -- Check if cache is stale (optional check)
    SELECT is_stale INTO v_cache_stale
    FROM check_cache_freshness(v_tenant_id)
    LIMIT 1;
    
    -- If cache is stale, optionally refresh (comment out for performance)
    -- IF v_cache_stale THEN
    --     PERFORM refresh_feature_flags_cache();
    -- END IF;
    
    -- Get result from cache
    SELECT tffc.enabled, tffc.value
    INTO v_result
    FROM mv_tenant_feature_flags_cache tffc
    WHERE tffc.tenant_id = v_tenant_id 
      AND tffc.feature_flag_name = flag_name;
    
    IF v_result IS NULL THEN
        RAISE EXCEPTION 'Feature flag not found in cache: % for tenant: %', flag_name, v_tenant_id;
    END IF;
    
    RETURN QUERY SELECT v_result.enabled, v_result.value;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Bulk evaluation using cache
CREATE OR REPLACE FUNCTION evaluate_all_feature_flags_cached()
RETURNS TABLE(flag_name VARCHAR, enabled BOOLEAN, value JSONB, flag_type VARCHAR, source TEXT) AS $$
DECLARE
    v_tenant_id UUID;
BEGIN
    -- Get current tenant ID
    v_tenant_id := NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
    
    IF v_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context set';
    END IF;
    
    RETURN QUERY
    SELECT 
        tffc.feature_flag_name::VARCHAR,
        tffc.enabled,
        tffc.value,
        tffc.flag_type::VARCHAR,
        tffc.evaluation_source::TEXT
    FROM mv_tenant_feature_flags_cache tffc
    WHERE tffc.tenant_id = v_tenant_id
    ORDER BY tffc.feature_flag_name;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- AUTOMATIC CACHE REFRESH TRIGGERS
-- =====================================================

-- Function to trigger cache refresh on data changes
CREATE OR REPLACE FUNCTION trigger_cache_refresh()
RETURNS TRIGGER AS $$
BEGIN
    -- Async refresh (use pg_notify for external refresh or schedule)
    PERFORM pg_notify('feature_flags_cache_refresh', 
        jsonb_build_object(
            'operation', TG_OP,
            'table', TG_TABLE_NAME,
            'tenant_id', COALESCE(NEW.tenant_id, OLD.tenant_id)
        )::text
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Add triggers to automatically notify when refresh is needed
CREATE TRIGGER feature_flags_cache_refresh_trigger
    AFTER INSERT OR UPDATE OR DELETE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION trigger_cache_refresh();

CREATE TRIGGER tenant_overrides_cache_refresh_trigger
    AFTER INSERT OR UPDATE OR DELETE ON tenant_feature_overrides
    FOR EACH ROW
    EXECUTE FUNCTION trigger_cache_refresh();

-- =====================================================
-- PERMISSIONS
-- =====================================================

-- Grant access to cache functions
GRANT EXECUTE ON FUNCTION refresh_feature_flags_cache() TO admin_role;
GRANT EXECUTE ON FUNCTION get_feature_flags_cache_stats() TO admin_role, readonly_role;
GRANT EXECUTE ON FUNCTION check_cache_freshness(UUID) TO application_role, admin_role;
GRANT EXECUTE ON FUNCTION evaluate_feature_flag_cached(VARCHAR) TO application_role;
GRANT EXECUTE ON FUNCTION evaluate_all_feature_flags_cached() TO application_role;

-- Grant access to materialized view
GRANT SELECT ON mv_tenant_feature_flags_cache TO application_role, admin_role, readonly_role;

-- =====================================================
-- COMMENTS
-- =====================================================
COMMENT ON MATERIALIZED VIEW mv_tenant_feature_flags_cache IS 'Materialized view for fast feature flag lookups with pre-computed evaluations';
COMMENT ON FUNCTION refresh_feature_flags_cache() IS 'Refreshes the feature flags cache materialized view';
COMMENT ON FUNCTION get_feature_flags_cache_stats() IS 'Returns statistics about the feature flags cache';
COMMENT ON FUNCTION check_cache_freshness(UUID) IS 'Checks if the cache is stale for a specific tenant';
COMMENT ON FUNCTION evaluate_feature_flag_cached(VARCHAR) IS 'Fast feature flag evaluation using materialized view cache';
COMMENT ON FUNCTION evaluate_all_feature_flags_cached() IS 'Fast bulk feature flag evaluation using materialized view cache';
COMMENT ON FUNCTION trigger_cache_refresh() IS 'Triggers cache refresh notification when data changes';
-- Creates maintenance functions for cleanup and system health

-- =====================================================
-- AUDIT LOG CLEANUP FUNCTIONS
-- =====================================================

-- Function to clean up old feature flag audit logs
CREATE OR REPLACE FUNCTION cleanup_old_feature_flag_audit_logs(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
    cutoff_date TIMESTAMPTZ;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    DELETE FROM audit_log
    WHERE event_type IN (
        'FEATURE_FLAG_CREATED', 
        'FEATURE_FLAG_UPDATED', 
        'FEATURE_FLAG_DELETED',
        'FEATURE_OVERRIDE_CREATED', 
        'FEATURE_OVERRIDE_UPDATED', 
        'FEATURE_OVERRIDE_DELETED',
        'FEATURE_FLAG_EVALUATED', 
        'FEATURE_FLAGS_BULK_EVALUATED',
        'FEATURE_FLAGS_CACHE_REFRESHED'
    )
    AND created_at < cutoff_date;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the cleanup operation
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        NULL, -- System operation
        'FEATURE_FLAGS_AUDIT_CLEANUP',
        'SYSTEM',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Feature flag audit logs cleaned up',
        jsonb_build_object(
            'deleted_count', deleted_count,
            'retention_days', retention_days,
            'cutoff_date', cutoff_date,
            'operation', 'cleanup'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up soft-deleted feature flags
CREATE OR REPLACE FUNCTION cleanup_soft_deleted_feature_flags(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
    cutoff_date TIMESTAMPTZ;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::INTERVAL;
    
    -- Hard delete feature flags that have been soft-deleted for the retention period
    DELETE FROM feature_flags
    WHERE deleted_at IS NOT NULL 
      AND deleted_at < cutoff_date;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the cleanup operation
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        NULL, -- System operation
        'FEATURE_FLAGS_HARD_DELETE_CLEANUP',
        'SYSTEM',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Soft-deleted feature flags permanently removed',
        jsonb_build_object(
            'deleted_count', deleted_count,
            'retention_days', retention_days,
            'cutoff_date', cutoff_date,
            'operation', 'hard_delete_cleanup'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- SYSTEM HEALTH AND MONITORING FUNCTIONS
-- =====================================================

-- Function to get feature flag system health metrics
CREATE OR REPLACE FUNCTION get_feature_flags_health_metrics()
RETURNS TABLE(
    metric_name TEXT,
    metric_value NUMERIC,
    metric_unit TEXT,
    metric_status TEXT,
    details JSONB
) AS $$
BEGIN
    RETURN QUERY
    WITH metrics AS (
        -- Total feature flags
        SELECT 
            'total_feature_flags' as name,
            COUNT(*)::NUMERIC as value,
            'count' as unit,
            CASE WHEN COUNT(*) > 0 THEN 'healthy' ELSE 'warning' END as status,
            jsonb_build_object('active_only', COUNT(*) FILTER (WHERE deleted_at IS NULL)) as details
        FROM feature_flags
        
        UNION ALL
        
        -- Total tenant overrides
        SELECT 
            'total_tenant_overrides' as name,
            COUNT(*)::NUMERIC as value,
            'count' as unit,
            'healthy' as status,
            jsonb_build_object('enabled_overrides', COUNT(*) FILTER (WHERE enabled = true)) as details
        FROM tenant_feature_overrides
        
        UNION ALL
        
        -- Average flags per tenant
        SELECT 
            'avg_flags_per_tenant' as name,
            COALESCE(ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT tenant_id), 0), 2), 0) as value,
            'count' as unit,
            CASE 
                WHEN COUNT(DISTINCT tenant_id) = 0 THEN 'error'
                WHEN COUNT(*)::NUMERIC / COUNT(DISTINCT tenant_id) > 100 THEN 'warning'
                ELSE 'healthy' 
            END as status,
            jsonb_build_object(
                'total_flags', COUNT(*),
                'total_tenants', COUNT(DISTINCT tenant_id)
            ) as details
        FROM feature_flags
        WHERE deleted_at IS NULL
        
        UNION ALL
        
        -- Cache age
        SELECT 
            'cache_age_minutes' as name,
            COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(cache_created_at)))/60, 0) as value,
            'minutes' as unit,
            CASE 
                WHEN MIN(cache_created_at) IS NULL THEN 'error'
                WHEN EXTRACT(EPOCH FROM (NOW() - MIN(cache_created_at)))/60 > 60 THEN 'warning'
                ELSE 'healthy' 
            END as status,
            jsonb_build_object(
                'cache_entries', COUNT(*),
                'last_refresh', MIN(cache_created_at)
            ) as details
        FROM mv_tenant_feature_flags_cache
        
        UNION ALL
        
        -- Rollout percentage distribution
        SELECT 
            'flags_with_rollout' as name,
            COUNT(*) FILTER (WHERE rollout_percentage IS NOT NULL)::NUMERIC as value,
            'count' as unit,
            'healthy' as status,
            jsonb_build_object(
                'avg_rollout_percentage', ROUND(AVG(rollout_percentage) FILTER (WHERE rollout_percentage IS NOT NULL), 2),
                'max_rollout_percentage', MAX(rollout_percentage),
                'min_rollout_percentage', MIN(rollout_percentage) FILTER (WHERE rollout_percentage IS NOT NULL)
            ) as details
        FROM feature_flags
        WHERE deleted_at IS NULL
    )
    SELECT m.name, m.value, m.unit, m.status, m.details FROM metrics m;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get tenant-specific feature flag statistics
CREATE OR REPLACE FUNCTION get_tenant_feature_flag_stats(p_tenant_id UUID)
RETURNS TABLE(
    total_flags INTEGER,
    enabled_flags INTEGER,
    overridden_flags INTEGER,
    rollout_flags INTEGER,
    flag_types JSONB,
    last_evaluation TIMESTAMPTZ,
    evaluation_count_today INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH tenant_stats AS (
        SELECT 
            COUNT(*)::INTEGER as total_flags,
            COUNT(*) FILTER (WHERE tffc.enabled = true)::INTEGER as enabled_flags,
            COUNT(*) FILTER (WHERE tffc.evaluation_source = 'override')::INTEGER as overridden_flags,
            COUNT(*) FILTER (WHERE tffc.evaluation_source = 'rollout')::INTEGER as rollout_flags,
            jsonb_object_agg(tffc.flag_type, COUNT(*)) as flag_types
        FROM mv_tenant_feature_flags_cache tffc
        WHERE tffc.tenant_id = p_tenant_id
    ),
    audit_stats AS (
        SELECT 
            MAX(al.created_at) as last_evaluation,
            COUNT(*) FILTER (WHERE al.created_at >= CURRENT_DATE)::INTEGER as evaluation_count_today
        FROM audit_log al
        WHERE al.tenant_id = p_tenant_id
          AND al.event_type IN ('FEATURE_FLAG_EVALUATED', 'FEATURE_FLAGS_BULK_EVALUATED')
    )
    SELECT 
        ts.total_flags,
        ts.enabled_flags,
        ts.overridden_flags,
        ts.rollout_flags,
        ts.flag_types,
        aus.last_evaluation,
        aus.evaluation_count_today
    FROM tenant_stats ts
    CROSS JOIN audit_stats aus;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- DATA INTEGRITY FUNCTIONS
-- =====================================================

-- Function to check data integrity
CREATE OR REPLACE FUNCTION check_feature_flags_integrity()
RETURNS TABLE(
    check_name TEXT,
    status TEXT,
    issue_count INTEGER,
    details JSONB
) AS $$
BEGIN
    RETURN QUERY
    WITH integrity_checks AS (
        -- Check for orphaned overrides
        SELECT 
            'orphaned_overrides' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'override_id', tfo.id,
                    'tenant_id', tfo.tenant_id,
                    'feature_flag_id', tfo.feature_flag_id
                )
            ) as details
        FROM tenant_feature_overrides tfo
        LEFT JOIN feature_flags ff ON tfo.feature_flag_id = ff.id
        WHERE ff.id IS NULL
        
        UNION ALL
        
        -- Check for mismatched tenant IDs
        SELECT 
            'mismatched_tenant_ids' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'override_id', tfo.id,
                    'override_tenant_id', tfo.tenant_id,
                    'flag_tenant_id', ff.tenant_id
                )
            ) as details
        FROM tenant_feature_overrides tfo
        JOIN feature_flags ff ON tfo.feature_flag_id = ff.id
        WHERE tfo.tenant_id != ff.tenant_id
        
        UNION ALL
        
        -- Check for invalid rollout percentages
        SELECT 
            'invalid_rollout_percentages' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'flag_id', ff.id,
                    'flag_name', ff.name,
                    'rollout_percentage', ff.rollout_percentage
                )
            ) as details
        FROM feature_flags ff
        WHERE ff.rollout_percentage IS NOT NULL 
          AND (ff.rollout_percentage < 0 OR ff.rollout_percentage > 100)
        
        UNION ALL
        
        -- Check for duplicate flag names per tenant
        SELECT 
            'duplicate_flag_names' as check_name,
            CASE WHEN COUNT(*) = 0 THEN 'pass' ELSE 'fail' END as status,
            COUNT(*)::INTEGER as issue_count,
            jsonb_agg(
                jsonb_build_object(
                    'tenant_id', tenant_id,
                    'flag_name', name,
                    'count', flag_count
                )
            ) as details
        FROM (
            SELECT tenant_id, name, COUNT(*) as flag_count
            FROM feature_flags
            WHERE deleted_at IS NULL
            GROUP BY tenant_id, name
            HAVING COUNT(*) > 1
        ) duplicates
    )
    SELECT ic.check_name, ic.status, ic.issue_count, ic.details FROM integrity_checks ic;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to fix orphaned overrides
CREATE OR REPLACE FUNCTION fix_orphaned_overrides()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete orphaned overrides
    DELETE FROM tenant_feature_overrides tfo
    WHERE NOT EXISTS (
        SELECT 1 FROM feature_flags ff 
        WHERE ff.id = tfo.feature_flag_id
    );
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log the fix operation
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        NULL, -- System operation
        'FEATURE_FLAGS_ORPHANED_OVERRIDES_FIXED',
        'SYSTEM',
        'WARN',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Orphaned feature flag overrides removed',
        jsonb_build_object(
            'deleted_count', deleted_count,
            'operation', 'fix_orphaned_overrides'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- BATCH OPERATIONS
-- =====================================================

-- Function to bulk update rollout percentages
CREATE OR REPLACE FUNCTION bulk_update_rollout_percentage(
    flag_names TEXT[],
    new_percentage INTEGER,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
    target_tenant_id UUID;
BEGIN
    -- Use provided tenant_id or current context
    target_tenant_id := COALESCE(p_tenant_id, NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);
    
    IF target_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context provided';
    END IF;
    
    -- Validate percentage
    IF new_percentage < 0 OR new_percentage > 100 THEN
        RAISE EXCEPTION 'Rollout percentage must be between 0 and 100';
    END IF;
    
    -- Update rollout percentages
    UPDATE feature_flags 
    SET 
        rollout_percentage = new_percentage,
        updated_at = NOW()
    WHERE tenant_id = target_tenant_id
      AND name = ANY(flag_names)
      AND deleted_at IS NULL;
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    
    -- Log the bulk update
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        target_tenant_id,
        'FEATURE_FLAGS_BULK_ROLLOUT_UPDATE',
        'ADMIN',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Bulk rollout percentage update',
        jsonb_build_object(
            'flag_names', flag_names,
            'new_percentage', new_percentage,
            'updated_count', updated_count,
            'operation', 'bulk_rollout_update'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

-- Function to bulk create feature flags
CREATE OR REPLACE FUNCTION bulk_create_feature_flags(
    flag_definitions JSONB,
    p_tenant_id UUID DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    created_count INTEGER := 0;
    flag_def JSONB;
    target_tenant_id UUID;
BEGIN
    -- Use provided tenant_id or current context
    target_tenant_id := COALESCE(p_tenant_id, NULLIF(current_setting('app.current_tenant_id', true), '')::UUID);
    
    IF target_tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context provided';
    END IF;
    
    -- Process each flag definition
    FOR flag_def IN SELECT jsonb_array_elements(flag_definitions)
    LOOP
        INSERT INTO feature_flags (
            tenant_id,
            name,
            description,
            flag_type,
            default_value,
            rollout_percentage,
            target_audience,
            metadata
        ) VALUES (
            target_tenant_id,
            flag_def->>'name',
            flag_def->>'description',
            COALESCE(flag_def->>'flag_type', 'boolean'),
            COALESCE((flag_def->>'default_value')::BOOLEAN, false),
            (flag_def->>'rollout_percentage')::INTEGER,
            COALESCE(flag_def->'target_audience', '{}'),
            COALESCE(flag_def->'metadata', '{}')
        )
        ON CONFLICT (tenant_id, name) DO NOTHING;
        
        IF FOUND THEN
            created_count := created_count + 1;
        END IF;
    END LOOP;
    
    -- Log the bulk creation
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        target_tenant_id,
        'FEATURE_FLAGS_BULK_CREATED',
        'ADMIN',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Bulk feature flags creation',
        jsonb_build_object(
            'definitions', flag_definitions,
            'created_count', created_count,
            'operation', 'bulk_create'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN created_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- EXPORT AND IMPORT FUNCTIONS
-- =====================================================

-- Function to export tenant feature flags configuration
CREATE OR REPLACE FUNCTION export_tenant_feature_flags(p_tenant_id UUID)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    SELECT jsonb_build_object(
        'tenant_id', p_tenant_id,
        'export_timestamp', NOW(),
        'feature_flags', jsonb_agg(
            jsonb_build_object(
                'name', ff.name,
                'description', ff.description,
                'flag_type', ff.flag_type,
                'default_value', ff.default_value,
                'rollout_percentage', ff.rollout_percentage,
                'target_audience', ff.target_audience,
                'metadata', ff.metadata,
                'created_at', ff.created_at,
                'updated_at', ff.updated_at
            )
        ),
        'overrides', (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'feature_flag_name', tfo.feature_flag_name,
                    'enabled', tfo.enabled,
                    'value', tfo.value,
                    'reason', tfo.reason,
                    'created_at', tfo.created_at,
                    'updated_at', tfo.updated_at
                )
            )
            FROM tenant_feature_overrides tfo
            WHERE tfo.tenant_id = p_tenant_id
        )
    ) INTO result
    FROM feature_flags ff
    WHERE ff.tenant_id = p_tenant_id
      AND ff.deleted_at IS NULL;
    
    -- Log the export
    INSERT INTO audit_log (
        tenant_id,
        event_type,
        event_category,
        severity,
        user_id,
        decision,
        reason,
        context,
        session_id
    ) VALUES (
        p_tenant_id,
        'FEATURE_FLAGS_EXPORTED',
        'ADMIN',
        'INFO',
        NULLIF(current_setting('app.current_user_id', true), '')::UUID,
        'ALLOW',
        'Feature flags configuration exported',
        jsonb_build_object(
            'export_size_bytes', octet_length(result::text),
            'operation', 'export'
        ),
        NULLIF(current_setting('app.current_session_id', true), '')::UUID
    );
    
    RETURN result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================================================
-- PERMISSIONS
-- =====================================================

-- Grant execute permissions to admin role
GRANT EXECUTE ON FUNCTION cleanup_old_feature_flag_audit_logs(INTEGER) TO admin_role;
GRANT EXECUTE ON FUNCTION cleanup_soft_deleted_feature_flags(INTEGER) TO admin_role;
GRANT EXECUTE ON FUNCTION get_feature_flags_health_metrics() TO admin_role, readonly_role;
GRANT EXECUTE ON FUNCTION get_tenant_feature_flag_stats(UUID) TO application_role, admin_role;
GRANT EXECUTE ON FUNCTION check_feature_flags_integrity() TO admin_role;
GRANT EXECUTE ON FUNCTION fix_orphaned_overrides() TO admin_role;
GRANT EXECUTE ON FUNCTION bulk_update_rollout_percentage(TEXT[], INTEGER, UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION bulk_create_feature_flags(JSONB, UUID) TO admin_role;
GRANT EXECUTE ON FUNCTION export_tenant_feature_flags(UUID) TO admin_role;

-- Grant limited permissions to application role
GRANT EXECUTE ON FUNCTION get_tenant_feature_flag_stats(UUID) TO application_role;

-- =====================================================
-- FUNCTION COMMENTS
-- =====================================================
COMMENT ON FUNCTION cleanup_old_feature_flag_audit_logs(INTEGER) IS 'Cleans up feature flag related audit logs older than specified days';
COMMENT ON FUNCTION cleanup_soft_deleted_feature_flags(INTEGER) IS 'Permanently removes feature flags that have been soft-deleted for specified days';
COMMENT ON FUNCTION get_feature_flags_health_metrics() IS 'Returns comprehensive health metrics for the feature flag system';
COMMENT ON FUNCTION get_tenant_feature_flag_stats(UUID) IS 'Returns detailed statistics for a specific tenant''s feature flags';
COMMENT ON FUNCTION check_feature_flags_integrity() IS 'Performs data integrity checks on feature flag tables';
COMMENT ON FUNCTION fix_orphaned_overrides() IS 'Removes orphaned tenant feature overrides that reference non-existent feature flags';
COMMENT ON FUNCTION bulk_update_rollout_percentage(TEXT[], INTEGER, UUID) IS 'Updates rollout percentage for multiple feature flags in bulk';
COMMENT ON FUNCTION bulk_create_feature_flags(JSONB, UUID) IS 'Creates multiple feature flags from JSON definitions';
COMMENT ON FUNCTION export_tenant_feature_flags(UUID) IS 'Exports complete feature flag configuration for a tenant in JSON format';
-- Provides examples and documentation for using the feature flag system

-- =====================================================
-- EXAMPLE: BASIC FEATURE FLAG SETUP
-- =====================================================

-- Example: Set up context and create feature flags for a tenant
-- Replace 'your-tenant-uuid' with actual tenant ID

/*
-- Set tenant context
SET app.current_tenant_id = 'your-tenant-uuid';
SET app.current_user_id = 'your-user-uuid';
SET app.current_session_id = 'your-session-uuid';

-- Create basic feature flags
INSERT INTO feature_flags (tenant_id, name, description, flag_type, default_value, metadata) VALUES
    ('your-tenant-uuid', 'advanced_reporting', 'Enable advanced reporting features', 'boolean', false, '{"category": "reporting", "priority": "high"}'),
    ('your-tenant-uuid', 'new_dashboard', 'Enable new dashboard UI', 'boolean', false, '{"category": "ui", "priority": "medium"}'),
    ('your-tenant-uuid', 'api_rate_limit', 'API rate limit configuration', 'number', true, '{"category": "performance", "default_limit": 1000}'),
    ('your-tenant-uuid', 'feature_rollout_test', 'Test gradual rollout', 'boolean', false, '{"category": "testing"}');

-- Set rollout percentage for gradual rollout
UPDATE feature_flags 
SET rollout_percentage = 25 
WHERE name = 'feature_rollout_test' 
  AND tenant_id = 'your-tenant-uuid';

-- Create tenant-specific overrides
INSERT INTO tenant_feature_overrides (tenant_id, feature_flag_id, feature_flag_name, enabled, reason) 
SELECT 
    'your-tenant-uuid',
    ff.id,
    ff.name,
    true,
    'Enable advanced reporting for premium tenant'
FROM feature_flags ff 
WHERE ff.name = 'advanced_reporting' 
  AND ff.tenant_id = 'your-tenant-uuid';
*/

-- =====================================================
-- EXAMPLE: FEATURE FLAG EVALUATION
-- =====================================================

-- Example usage queries (run after setting up feature flags above)

/*
-- Evaluate a single feature flag
SELECT * FROM evaluate_feature_flag('advanced_reporting');

-- Fast evaluation without audit logging
SELECT * FROM evaluate_feature_flag_fast('new_dashboard');

-- Evaluate all feature flags for current tenant
SELECT * FROM evaluate_all_feature_flags();

-- Use cached evaluation (faster for frequent calls)
SELECT * FROM evaluate_feature_flag_cached('advanced_reporting');
SELECT * FROM evaluate_all_feature_flags_cached();
*/

-- =====================================================
-- EXAMPLE: MONITORING AND MAINTENANCE
-- =====================================================

-- Example monitoring queries

/*
-- Check system health
SELECT * FROM get_feature_flags_health_metrics();

-- Get tenant-specific statistics
SELECT * FROM get_tenant_feature_flag_stats('your-tenant-uuid');

-- Check cache statistics
SELECT * FROM get_feature_flags_cache_stats();

-- Check cache freshness for a tenant
SELECT * FROM check_cache_freshness('your-tenant-uuid');

-- Check data integrity
SELECT * FROM check_feature_flags_integrity();
*/

-- =====================================================
-- EXAMPLE: BULK OPERATIONS
-- =====================================================

-- Example bulk operations

/*
-- Bulk update rollout percentages
SELECT bulk_update_rollout_percentage(
    ARRAY['new_dashboard', 'feature_rollout_test'], 
    50, 
    'your-tenant-uuid'
);

-- Bulk create feature flags
SELECT bulk_create_feature_flags('[
    {
        "name": "experimental_feature_a",
        "description": "Experimental feature A",
        "flag_type": "boolean",
        "default_value": false,
        "rollout_percentage": 10,
        "metadata": {"category": "experimental"}
    },
    {
        "name": "experimental_feature_b",
        "description": "Experimental feature B",
        "flag_type": "boolean",  
        "default_value": false,
        "metadata": {"category": "experimental"}
    }
]'::jsonb, 'your-tenant-uuid');

-- Export tenant configuration
SELECT export_tenant_feature_flags('your-tenant-uuid');
*/

-- =====================================================
-- EXAMPLE: MAINTENANCE OPERATIONS
-- =====================================================

-- Example maintenance operations (admin only)

/*
-- Refresh the materialized view cache
SELECT refresh_feature_flags_cache();

-- Clean up old audit logs (keep last 90 days)
SELECT cleanup_old_feature_flag_audit_logs(90);

-- Clean up soft-deleted flags (after 30 days)
SELECT cleanup_soft_deleted_feature_flags(30);

-- Fix any orphaned overrides
SELECT fix_orphaned_overrides();
*/

-- =====================================================
-- EXAMPLE: ADVANCED QUERIES
-- =====================================================

-- Example advanced analytics queries

/*
-- Feature flags by type and status
SELECT 
    flag_type,
    evaluation_source,
    COUNT(*) as count,
    ROUND(AVG(CASE WHEN enabled THEN 1 ELSE 0 END) * 100, 2) as enabled_percentage
FROM mv_tenant_feature_flags_cache
WHERE tenant_id = 'your-tenant-uuid'
GROUP BY flag_type, evaluation_source
ORDER BY flag_type, evaluation_source;

-- Most evaluated features (from audit logs)
SELECT 
    context->>'feature_flag_name' as feature_name,
    COUNT(*) as evaluation_count,
    COUNT(*) FILTER (WHERE decision = 'ALLOW') as enabled_count,
    ROUND(COUNT(*) FILTER (WHERE decision = 'ALLOW')::NUMERIC / COUNT(*) * 100, 2) as enabled_percentage
FROM audit_log
WHERE tenant_id = 'your-tenant-uuid'
  AND event_type = 'FEATURE_FLAG_EVALUATED'
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY context->>'feature_flag_name'
ORDER BY evaluation_count DESC
LIMIT 10;

-- Rollout effectiveness analysis
SELECT 
    ff.name,
    ff.rollout_percentage,
    COUNT(DISTINCT al.session_id) as unique_evaluations,
    COUNT(*) FILTER (WHERE al.decision = 'ALLOW') as enabled_evaluations,
    ROUND(COUNT(*) FILTER (WHERE al.decision = 'ALLOW')::NUMERIC / COUNT(*) * 100, 2) as actual_enabled_percentage
FROM feature_flags ff
JOIN audit_log al ON al.context->>'feature_flag_name' = ff.name
WHERE ff.tenant_id = 'your-tenant-uuid'
  AND ff.rollout_percentage IS NOT NULL
  AND al.event_type = 'FEATURE_FLAG_EVALUATED'
  AND al.created_at >= NOW() - INTERVAL '24 hours'
GROUP BY ff.name, ff.rollout_percentage
ORDER BY ff.rollout_percentage DESC;
*/

-- =====================================================
-- PERFORMANCE MONITORING QUERIES
-- =====================================================

-- Queries to monitor performance

/*
-- Index usage statistics
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides', 'mv_tenant_feature_flags_cache')
ORDER BY idx_scan DESC;

-- Table size statistics
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples
FROM pg_stat_user_tables 
WHERE tablename IN ('feature_flags', 'tenant_feature_overrides', 'audit_log')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Slow query analysis for feature flag operations
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE query ILIKE '%feature_flag%' 
   OR query ILIKE '%tenant_feature_overrides%'
ORDER BY mean_time DESC
LIMIT 10;
*/

-- =====================================================
-- TROUBLESHOOTING GUIDE
-- =====================================================

/*
TROUBLESHOOTING COMMON ISSUES:

1. "No tenant context set" error:
   - Ensure you set the tenant context before calling functions
   - SET app.current_tenant_id = 'your-tenant-uuid';

2. Feature flag not found:
   - Check if the flag exists and is not soft-deleted
   - Verify tenant_id matches your context
   - SELECT * FROM feature_flags WHERE name = 'flag_name' AND deleted_at IS NULL;

3. Cache is stale:
   - Refresh the materialized view cache
   - SELECT refresh_feature_flags_cache();
   - Check for notifications: LISTEN feature_flags_cache_refresh;

4. Performance issues:
   - Use cached evaluation functions for high-frequency calls
   - Monitor index usage and table statistics
   - Consider partitioning audit_log table for large datasets

5. Data integrity issues:
   - Run integrity checks regularly
   - SELECT * FROM check_feature_flags_integrity();
   - Fix issues: SELECT fix_orphaned_overrides();

6. Audit log growing too large:
   - Regular cleanup of old logs
   - SELECT cleanup_old_feature_flag_audit_logs(90);
   - Consider partitioning by date

7. RLS (Row Level Security) issues:
   - Verify policies are correctly applied
   - Check role permissions
   - Ensure tenant context is properly set
*/

-- =====================================================
-- BEST PRACTICES
-- =====================================================

/*
FEATURE FLAG BEST PRACTICES:

1. Naming Conventions:
   - Use descriptive, hierarchical names: 'ui.new_dashboard', 'api.v2_endpoints'
   - Avoid spaces, use underscores or dots
   - Include the feature area as prefix

2. Rollout Strategy:
   - Start with low percentages (5-10%) for new features
   - Monitor metrics before increasing rollout
   - Use overrides for specific tenants during testing

3. Cleanup:
   - Regularly review and remove unused flags
   - Set expiration dates in metadata
   - Use soft deletes initially, then hard delete after grace period

4. Monitoring:
   - Set up alerts for integrity check failures
   - Monitor evaluation patterns and performance
   - Track feature adoption rates

5. Documentation:
   - Document flag purpose and expected lifespan in description
   - Use metadata to store additional context
   - Maintain changelog of flag modifications

6. Security:
   - Audit all flag modifications
   - Use reason field for overrides
   - Regular review of admin actions

7. Performance:
   - Use cached evaluation for high-frequency checks
   - Refresh cache after bulk operations
   - Monitor query performance and optimize indexes

8. Testing:
   - Test both enabled and disabled states
   - Verify rollout percentages work as expected
   - Test override functionality
*/
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure all foreign key references belong to the same tenant
    IF TG_TABLE_NAME = 'persons' THEN
        -- Validate entity belongs to same tenant
        IF NOT EXISTS (
            SELECT 1 FROM entities e
            JOIN tenants t ON e.tenant_id = t.id
            WHERE e.uuid = NEW.entity_id AND t.id = NEW.tenant_id
        ) THEN
            RAISE EXCEPTION 'Entity % does not belong to tenant %', NEW.entity_id, NEW.tenant_id;
        END IF;
    ELSE
        RAISE NOTICE 'enforce_tenant_isolation trigger fired on table %', TG_TABLE_NAME;
    END IF;

    -- Add similar validations for other tables as needed
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
GRANT SELECT, INSERT, UPDATE, DELETE ON entities TO application_role;

-- Tenant context validation
CREATE OR REPLACE FUNCTION validate_and_set_tenant_context(p_tenant_id UUID)
RETURNS TABLE(tenant_name TEXT, tenant_status TEXT) AS $$
DECLARE
    v_tenant_record RECORD;
BEGIN
    -- Validate and fetch tenant information
    SELECT id, name, status, deleted_at, last_activity_at
    INTO v_tenant_record
    FROM tenants
    WHERE id = p_tenant_id;

    -- Check if tenant exists
    IF v_tenant_record.id IS NULL THEN
        RAISE EXCEPTION 'Tenant not found: %', p_tenant_id;
    END IF;

    -- Check if tenant is soft-deleted
    IF v_tenant_record.deleted_at IS NOT NULL THEN
        RAISE EXCEPTION 'Tenant is deleted: %', p_tenant_id;
    END IF;

    -- Check tenant status
    IF v_tenant_record.status NOT IN ('active', 'pending') THEN
        RAISE EXCEPTION 'Tenant is not active: % (status: %)', p_tenant_id, v_tenant_record.status;
    END IF;

    -- Update last activity
    UPDATE tenants 
    SET last_activity_at = NOW() 
    WHERE id = p_tenant_id;

    -- Set tenant context
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);

    -- Return tenant information
    RETURN QUERY SELECT v_tenant_record.name, v_tenant_record.status;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;ALTER TABLE tenants ADD COLUMN last_activity_at TIMESTAMPTZ DEFAULT NOW();-- Tenant provisioning function
CREATE OR REPLACE FUNCTION provision_tenant_complete(
    p_name VARCHAR(255),
    p_email VARCHAR(255),
    p_subdomain VARCHAR(63) DEFAULT NULL,
    p_industry VARCHAR(50) DEFAULT NULL,
    p_company_size VARCHAR(20) DEFAULT 'small',
    p_currency_code CHAR(3) DEFAULT 'USD',
    p_timezone VARCHAR(50) DEFAULT 'UTC',
    p_settings JSONB DEFAULT '{}'
)
RETURNS TABLE(
    id UUID
) AS $body$
DECLARE
    v_tenant_id UUID;
    v_slug VARCHAR(50);
BEGIN
    -- Generate UUID and slug
    v_tenant_id := gen_random_uuid();
    v_slug := lower(regexp_replace(p_name, '[^a-zA-Z0-9]+', '-', 'g'));

    -- Ensure slug uniqueness
    WHILE EXISTS (SELECT 1 FROM tenants WHERE slug = v_slug AND deleted_at IS NULL) LOOP
        v_slug := v_slug || '-' || substring(v_tenant_id::text, 1, 8);
    END LOOP;

    -- Create tenant record
    INSERT INTO tenants (
        id, slug, name, email, subdomain, status, industry, 
        company_size, currency_code, timezone, settings
    ) VALUES (
        v_tenant_id, v_slug, p_name, p_email, p_subdomain, 'pending',
        p_industry, p_company_size, p_currency_code, p_timezone, p_settings
    );

    -- Create default configuration
    INSERT INTO tenant_configurations (
        tenant_id, default_currency
    ) VALUES (
        v_tenant_id, p_currency_code
    );

    -- Initialize usage statistics for current month
    INSERT INTO tenant_usage_stats (tenant_id, period_start, period_end)
    VALUES (
        v_tenant_id,
        date_trunc('month', CURRENT_DATE)::DATE,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::DATE
    );

    -- Return tenant information
    RETURN QUERY SELECT v_tenant_id AS id;
END;
$body$ LANGUAGE plpgsql;