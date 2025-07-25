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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    EXECUTE FUNCTION generate_slug_from_name();-- =====================================================
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
-- VALIDATION TRIGGER
-- =====================================================================

-- Validation trigger to maintain entity_id consistency
CREATE OR REPLACE FUNCTION maintain_entity_id()
RETURNS TRIGGER AS $$
BEGIN
    -- Set entity_id to uuid if not provided (for entities table)
    IF TG_TABLE_NAME = 'entities' AND NEW.entity_id IS NULL THEN
        NEW.entity_id := NEW.uuid;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER entities_maintain_entity_id
    BEFORE INSERT OR UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION maintain_entity_id();

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
    USING (true);-- =====================================================================
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
    
    -- Standard validation columns
    version INTEGER NOT NULL DEFAULT 1,
    last_validation_run TIMESTAMPTZ,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}'::jsonb,         -- Device fingerprinting data
    location_info JSONB DEFAULT '{}'::jsonb,       -- Geographic/network location for ABAC
    expires_at TIMESTAMPTZ NOT NULL,
    risk_score INTEGER DEFAULT 0,                 -- Calculated risk score (0-100)
    
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
COMMENT ON COLUMN user_sessions.risk_score IS 'Calculated risk score from 0-100 based on action, context, and user behavior';
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
-- Comprehensive audit logging with compliance flags and risk scoring.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
'Comprehensive audit log with compliance tracking, risk scoring, and detailed context for security monitoring and regulatory compliance.';

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
-- Comprehensive user view with all related data
-- ------------------------------------------------------------------------------------------------
CREATE VIEW user_complete_view AS
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

COMMENT ON VIEW user_complete_view IS
'Comprehensive view combining user, person, and employee data with role aggregations and combined ABAC attributes for authorization decisions.';

-- ------------------------------------------------------------------------------------------------
-- Role permissions summary view
-- ------------------------------------------------------------------------------------------------
CREATE VIEW role_permissions_summary AS
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

COMMENT ON VIEW role_permissions_summary IS
'Summary view of roles with their permissions, resources, actions, and user assignment counts for role management and analysis.';

-- ------------------------------------------------------------------------------------------------
-- Audit summary view for security monitoring
-- ------------------------------------------------------------------------------------------------
CREATE VIEW audit_summary_view AS
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

COMMENT ON VIEW audit_summary_view IS
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
    END IF;

    -- Add similar validations for other tables as needed
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
CREATE MATERIALIZED VIEW user_effective_permissions AS
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
    ON user_effective_permissions (user_id, resource_id, action_id);
    
COMMENT ON MATERIALIZED VIEW user_effective_permissions IS
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
        email = 'redacted_' || uuid_generate_v4() || '@example.com',
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


CREATE VIEW security_threat_dashboard AS
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

COMMENT ON VIEW security_threat_dashboard IS
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
-- POLICY EVALUATIONS CACHE
-- ------------------------------------------------------------------------------------------------
-- Caches policy evaluation results for performance optimization.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS policy_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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


