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
