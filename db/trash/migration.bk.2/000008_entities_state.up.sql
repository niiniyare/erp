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
    USING (true);