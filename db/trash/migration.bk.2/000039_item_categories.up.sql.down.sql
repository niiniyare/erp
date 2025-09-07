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
