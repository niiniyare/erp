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
