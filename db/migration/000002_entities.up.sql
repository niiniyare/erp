-- =====================================================================
-- ENTITIES TABLE - Core business entity management
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
-- ENTITY HIERARCHY MANAGEMENT
-- =====================================================================

-- Create unique index for tenant-code combination
CREATE UNIQUE INDEX tenant_code_unique_idx
    ON entities (tenant_id, code)
    WHERE code IS NOT NULL;

COMMENT ON INDEX tenant_code_unique_idx IS 
'Ensures entity codes are unique within each tenant - Only applies when code is not NULL';

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

COMMENT ON COLUMN hierarchy_paths.ancestor_id IS 
'Parent entity in the relationship - References entities.uuid';

COMMENT ON COLUMN hierarchy_paths.descendant_id IS 
'Child entity in the relationship - References entities.uuid';

COMMENT ON COLUMN hierarchy_paths.depth IS 
'Hierarchical distance - 0 for self-reference, 1 for direct parent-child, 2+ for deeper relationships';

-- =====================================================================
-- ENTITY STATE MANAGEMENT
-- =====================================================================

-- Entity state tracking for sequence numbers and fiscal periods
CREATE TABLE IF NOT EXISTS entitystate (
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

-- Index for reverse hierarchy lookups (finding parents of a descendant)
CREATE INDEX idx_hierarchy_paths_descendant ON hierarchy_paths(descendant_id);
COMMENT ON INDEX idx_hierarchy_paths_descendant IS 
'Enables efficient reverse hierarchy traversal - Finding all ancestors of a given entity';

-- Index for entity hierarchy depth-based queries
CREATE INDEX idx_hierarchy_paths_depth ON hierarchy_paths(tenant_id, depth);
COMMENT ON INDEX idx_hierarchy_paths_depth IS 
'Optimizes queries filtering by hierarchy depth - Useful for organization level reports';

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

-- Validation trigger to maintain entity_id consistency
CREATE OR REPLACE FUNCTION maintain_entity_id()
RETURNS TRIGGER AS $$
BEGIN
    -- Set entity_id to uuid if not provided (for entities table)
    IF TG_TABLE_NAME = 'entities' AND NEW.entity_id IS NULL THEN
        NEW.entity_id := NEW.uuid;
    END IF;
    
    -- For hierarchy_paths, entity_id should reference ancestor
    IF TG_TABLE_NAME = 'hierarchy_paths' AND NEW.entity_id IS NULL THEN
        NEW.entity_id := NEW.ancestor_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER entities_maintain_entity_id
    BEFORE INSERT OR UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION maintain_entity_id();

CREATE TRIGGER hierarchy_paths_maintain_entity_id
    BEFORE INSERT OR UPDATE ON hierarchy_paths
    FOR EACH ROW EXECUTE FUNCTION maintain_entity_id();

-- Ensure sequence numbers are positive
ALTER TABLE entitystate ADD CONSTRAINT positive_sequence 
    CHECK (sequence > 0);
COMMENT ON CONSTRAINT positive_sequence ON entitystate IS 
'Ensures sequence numbers are always positive values';

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
-- ADDITIONAL PERFORMANCE CONSIDERATIONS
-- =====================================================================

-- Consider these additional optimizations based on usage patterns:

-- 1. For frequent entity name searches (if supporting partial matching):
-- CREATE INDEX idx_entities_name_gin ON entities USING gin(name gin_trgm_ops);

-- 2. For JSON-based queries on settings or address:
CREATE INDEX idx_entities_settings_gin ON entities USING gin(settings);
CREATE INDEX idx_entities_address_gin ON entities USING gin(address);

-- 3. For temporal queries on creation/modification:
-- CREATE INDEX idx_entities_created_at ON entities(created_at);
-- CREATE INDEX idx_entities_updated_at ON entities(updated_at);

-- 4. For soft deletion queries:
CREATE INDEX idx_entities_deleted_at ON entities(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX idx_entities_tenant ON entities(tenant_id);
CREATE INDEX idx_entities_parent ON entities(parent_id);
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_hierarchy_paths_tenant ON hierarchy_paths(tenant_id);
CREATE INDEX idx_hierarchy_paths_ancestor ON hierarchy_paths(ancestor_id);



-- Enable Row Level Security
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE hierarchy_paths ENABLE ROW LEVEL SECURITY;
ALTER TABLE entitystate ENABLE ROW LEVEL SECURITY;

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

-- Admin bypass policies
CREATE POLICY admin_full_access_policy ON entities
    FOR ALL TO admin_role
    USING (true);

CREATE POLICY admin_full_access_policy ON hierarchy_paths
    FOR ALL TO admin_role
    USING (true);

CREATE POLICY admin_full_access_policy ON entitystate
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- MAINTENANCE CONSIDERATIONS
-- =====================================================================

-- Regular maintenance tasks to consider:

-- 1. Hierarchy integrity check (ensure no orphaned records in hierarchy_paths)
-- 2. Sequence number gap analysis (check for missing sequences)
-- 3. Soft deletion cleanup (archive old deleted records)
-- 4. Index maintenance (REINDEX for heavily updated tables)
-- 5. Statistics updates (ANALYZE for query planner optimization)




-- -- Root entity/company table with hierarchical structure and accounting preferences
-- CREATE TABLE entities (
--     uuid UUID PRIMARY KEY,
--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     parent_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
--
--     name VARCHAR(255) NOT NULL,
--     code VARCHAR(50), -- Internal reference code
--
--     type VARCHAR(20) NOT NULL CHECK (
--         type IN (
--             'COMPANY', 'SUBSIDIARY', 'REGION', 'BRANCH', 'LOCATION',
--             'DEPARTMENT', 'DIVISION', 'COST_CENTER', 'PROJECT', 'BUDGET_UNIT'
--         )
--     ) DEFAULT 'COMPANY',
--
--     is_active BOOLEAN NOT NULL DEFAULT true,
--     hidden BOOLEAN NOT NULL DEFAULT false,
--     accrual_method BOOLEAN NOT NULL,                    -- TRUE = Accrual, FALSE = Cash
--     fy_start_month INTEGER NOT NULL CHECK (fy_start_month BETWEEN 1 AND 12),
--
--     address JSONB DEFAULT '{}'::jsonb,
--     picture VARCHAR(100),
--     settings JSONB DEFAULT '{}'::jsonb,
--
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     deleted_at TIMESTAMPTZ,
--
--     UNIQUE (tenant_id, name)
-- );
--
-- COMMENT ON TABLE entities IS
-- 'Purpose: Stores business entities/organizations/companies.
--  Description: Core table representing different business entities that can
--               have their own accounting books, customers, vendors, etc.
--               Uses tree structure for hierarchical organization relationships.';
--
-- CREATE UNIQUE INDEX tenant_code_unique_idx
--     ON entities (tenant_id, code)
--     WHERE code IS NOT NULL;
--
-- -- Enhanced closure table for entity hierarchy
-- CREATE TABLE hierarchy_paths (
--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     ancestor_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
--     descendant_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
--     depth INT NOT NULL CHECK (depth >= 0),
--
--     PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
-- );
--
-- -- Entity state tracking for sequence numbers and fiscal periods
-- CREATE TABLE IF NOT EXISTS entitystate (
--     uuid UUID PRIMARY KEY,
--     fiscal_year SMALLINT,
--     key VARCHAR(10) NOT NULL,                             -- Document type (e.g., invoice, po)
--     sequence BIGINT NOT NULL,                             -- Next sequence number
--     entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
--     entity_unit_id UUID REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED
-- );
--
-- COMMENT ON TABLE entitystate IS
-- 'Manages sequence numbers for document numbering (invoices, POs, etc.)';
-- COMMENT ON COLUMN entitystate.key IS
-- 'Document type: invoice, po, estimate, bill, etc.';
--
--
