-- =====================================================================
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
    USING (true);