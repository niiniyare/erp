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
