-- =====================================================
-- INVENTORY MODULE
-- =====================================================

-- Warehouses/Locations
CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    address JSONB,
    warehouse_type VARCHAR(20) DEFAULT 'GENERAL'
        CHECK (warehouse_type IN ('GENERAL', 'RETAIL', 'TRANSIT', 'QUARANTINE')),
    manager_id INT REFERENCES employees(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);



-- Item Categories
CREATE TABLE item_categories (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    parent_id INT REFERENCES item_categories(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, parent_id, name)
);

-- Items/Products
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    item_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id INT REFERENCES item_categories(id),
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


-- Inventory Balances
CREATE TABLE inventory_balances (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    warehouse_id INT NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
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

-- Inventory Movements/Transactions
CREATE TABLE inventory_movements (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES items(id),
    warehouse_id INT NOT NULL REFERENCES warehouses(id),
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
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

