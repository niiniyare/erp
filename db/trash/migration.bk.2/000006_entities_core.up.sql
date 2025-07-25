-- =====================================================================
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
    USING (true);