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
