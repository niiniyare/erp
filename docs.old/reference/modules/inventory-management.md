# Inventory Management

##  Overview

The Inventory Management module provides  stock control, warehouse management, procurement, and logistics capabilities. It supports multi-warehouse operations, real-time stock tracking, automated reordering, and advanced inventory optimization techniques including ABC analysis, lot tracking, and serial number management.

## ️ Item Master & Catalog Management

### Item Master Data

```sql
-- Inventory items master table
CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Item identification
    item_code VARCHAR(100) UNIQUE NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    
    -- Item classification
    item_type VARCHAR(50) NOT NULL, -- product, service, kit, variant
    item_category_id UUID REFERENCES item_categories(id),
    item_group VARCHAR(100),
    brand VARCHAR(100),
    manufacturer VARCHAR(255),
    model_number VARCHAR(100),
    
    -- Physical properties
    weight DECIMAL(10,4),
    weight_unit VARCHAR(10) DEFAULT 'kg',
    dimensions JSONB, -- {length, width, height, unit}
    volume DECIMAL(10,4),
    volume_unit VARCHAR(10) DEFAULT 'cubic_meter',
    
    -- Stock management
    is_stock_item BOOLEAN DEFAULT true,
    track_quantity BOOLEAN DEFAULT true,
    track_serial_numbers BOOLEAN DEFAULT false,
    track_batches BOOLEAN DEFAULT false,
    
    -- Units of measure
    base_uom VARCHAR(20) NOT NULL DEFAULT 'each',
    purchase_uom VARCHAR(20),
    sales_uom VARCHAR(20),
    uom_conversion_factor DECIMAL(10,4) DEFAULT 1,
    
    -- Pricing
    standard_cost DECIMAL(15,4),
    average_cost DECIMAL(15,4),
    last_purchase_cost DECIMAL(15,4),
    standard_selling_price DECIMAL(15,4),
    
    -- Inventory parameters
    reorder_level DECIMAL(12,4) DEFAULT 0,
    minimum_stock DECIMAL(12,4) DEFAULT 0,
    maximum_stock DECIMAL(12,4),
    safety_stock DECIMAL(12,4) DEFAULT 0,
    lead_time_days INTEGER DEFAULT 7,
    
    -- Accounting integration
    inventory_account_id UUID REFERENCES accounts(id),
    cost_of_goods_account_id UUID REFERENCES accounts(id),
    income_account_id UUID REFERENCES accounts(id),
    expense_account_id UUID REFERENCES accounts(id),
    
    -- Quality and compliance
    quality_inspection_required BOOLEAN DEFAULT false,
    shelf_life_days INTEGER,
    expiry_tracking_required BOOLEAN DEFAULT false,
    hazardous_material BOOLEAN DEFAULT false,
    
    -- Digital assets
    primary_image_url VARCHAR(500),
    image_urls JSONB, -- Array of image URLs
    documents JSONB, -- Array of document attachments
    
    -- Barcodes and identification
    barcode VARCHAR(100),
    qr_code TEXT,
    sku VARCHAR(100),
    upc VARCHAR(20),
    ean VARCHAR(20),
    
    -- Status and lifecycle
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, discontinued, obsolete
    lifecycle_stage VARCHAR(50), -- new, growth, mature, decline, obsolete
    
    -- Tax and regulatory
    tax_category VARCHAR(50),
    hscode VARCHAR(20), -- Harmonized System code for international trade
    country_of_origin CHAR(2),
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_item_type CHECK (item_type IN ('product', 'service', 'kit', 'variant')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'discontinued', 'obsolete')),
    CONSTRAINT positive_costs CHECK (
        (standard_cost IS NULL OR standard_cost >= 0) AND
        (average_cost IS NULL OR average_cost >= 0) AND
        (last_purchase_cost IS NULL OR last_purchase_cost >= 0)
    )
);

-- Item categories hierarchy
CREATE TABLE item_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_category_id UUID REFERENCES item_categories(id),
    
    category_code VARCHAR(50) NOT NULL,
    category_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Category properties
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    
    -- Default settings for items in this category
    default_inventory_account_id UUID REFERENCES accounts(id),
    default_cogs_account_id UUID REFERENCES accounts(id),
    default_income_account_id UUID REFERENCES accounts(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, category_code)
);

-- Item variants for size, color, style variations
CREATE TABLE item_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    variant_code VARCHAR(100) UNIQUE NOT NULL,
    variant_name VARCHAR(255) NOT NULL,
    
    -- Variant attributes
    attributes JSONB, -- {color: "Red", size: "Large", style: "Modern"}
    
    -- Variant-specific pricing
    price_adjustment_type VARCHAR(20) DEFAULT 'none', -- none, fixed, percentage
    price_adjustment_value DECIMAL(15,4) DEFAULT 0,
    
    -- Variant-specific costs
    cost_adjustment_type VARCHAR(20) DEFAULT 'none',
    cost_adjustment_value DECIMAL(15,4) DEFAULT 0,
    
    -- Stock management
    separate_stock_tracking BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_price_adjustment CHECK (price_adjustment_type IN ('none', 'fixed', 'percentage')),
    CONSTRAINT valid_cost_adjustment CHECK (cost_adjustment_type IN ('none', 'fixed', 'percentage'))
);
```

### Item Kits and Bundles

```sql
-- Item kits (Bill of Materials for bundled items)
CREATE TABLE item_kits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kit_item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    component_item_id UUID NOT NULL REFERENCES inventory_items(id),
    
    quantity DECIMAL(12,4) NOT NULL DEFAULT 1,
    unit_of_measure VARCHAR(20),
    
    -- Component properties
    is_optional BOOLEAN DEFAULT false,
    can_substitute BOOLEAN DEFAULT false,
    substitute_items JSONB, -- Array of alternative item IDs
    
    -- Cost allocation
    cost_allocation_method VARCHAR(20) DEFAULT 'proportional', -- proportional, fixed, manual
    allocated_cost DECIMAL(15,4),
    
    sort_order INTEGER DEFAULT 0,
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(kit_item_id, component_item_id),
    CONSTRAINT valid_cost_allocation CHECK (cost_allocation_method IN ('proportional', 'fixed', 'manual'))
);
```

##  Multi-Warehouse Management

### Warehouse Structure

```sql
-- Warehouses
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Warehouse identification
    warehouse_code VARCHAR(50) UNIQUE NOT NULL,
    warehouse_name VARCHAR(255) NOT NULL,
    warehouse_type VARCHAR(50) DEFAULT 'standard', -- standard, transit, virtual, consignment
    
    -- Location information
    address JSONB,
    gps_coordinates JSONB, -- {latitude, longitude}
    timezone VARCHAR(50),
    
    -- Warehouse properties
    total_area DECIMAL(10,2),
    available_area DECIMAL(10,2),
    temperature_controlled BOOLEAN DEFAULT false,
    temperature_range JSONB, -- {min_temp, max_temp, unit}
    
    -- Operations
    operating_hours JSONB, -- Daily operating schedule
    manager_id UUID REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    
    -- Cost center integration
    cost_center_id UUID REFERENCES cost_centers(id),
    
    -- Default accounts for this warehouse
    inventory_account_id UUID REFERENCES accounts(id),
    inventory_adjustment_account_id UUID REFERENCES accounts(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_warehouse_type CHECK (warehouse_type IN ('standard', 'transit', 'virtual', 'consignment'))
);

-- Storage locations within warehouses
CREATE TABLE storage_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Location identification
    location_code VARCHAR(50) NOT NULL,
    location_name VARCHAR(255),
    location_type VARCHAR(50) DEFAULT 'bin', -- bin, shelf, rack, zone, area
    
    -- Hierarchical structure
    parent_location_id UUID REFERENCES storage_locations(id),
    location_path VARCHAR(500), -- Full hierarchical path like "Zone-A/Rack-01/Shelf-03/Bin-05"
    
    -- Physical properties
    capacity_weight DECIMAL(10,2),
    capacity_volume DECIMAL(10,2),
    dimensions JSONB,
    
    -- Location constraints
    temperature_controlled BOOLEAN DEFAULT false,
    humidity_controlled BOOLEAN DEFAULT false,
    restricted_access BOOLEAN DEFAULT false,
    hazmat_approved BOOLEAN DEFAULT false,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_available BOOLEAN DEFAULT true,
    
    -- Picking optimization
    pick_sequence INTEGER,
    pick_zone VARCHAR(50),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(warehouse_id, location_code),
    CONSTRAINT valid_location_type CHECK (location_type IN ('bin', 'shelf', 'rack', 'zone', 'area'))
);
```

### Stock Levels and Tracking

```sql
-- Current stock levels by warehouse and location
CREATE TABLE stock_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Item and location
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    location_id UUID REFERENCES storage_locations(id),
    
    -- Stock quantities
    on_hand_quantity DECIMAL(12,4) DEFAULT 0,
    available_quantity DECIMAL(12,4) DEFAULT 0, -- On hand minus allocated
    allocated_quantity DECIMAL(12,4) DEFAULT 0, -- Reserved for orders
    on_order_quantity DECIMAL(12,4) DEFAULT 0, -- Pending purchase orders
    
    -- Cost information
    average_cost DECIMAL(15,4) DEFAULT 0,
    total_value DECIMAL(15,2) DEFAULT 0,
    
    -- Stock status
    last_movement_date TIMESTAMPTZ,
    last_count_date DATE,
    cycle_count_due_date DATE,
    
    -- Reorder parameters (can override item defaults)
    reorder_level DECIMAL(12,4),
    minimum_stock DECIMAL(12,4),
    maximum_stock DECIMAL(12,4),
    
    -- Version control for concurrent updates
    version INTEGER DEFAULT 1,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(item_id, warehouse_id, location_id),
    CONSTRAINT non_negative_quantities CHECK (
        on_hand_quantity >= 0 AND
        available_quantity >= 0 AND
        allocated_quantity >= 0 AND
        on_order_quantity >= 0
    )
);

-- Serial number tracking
CREATE TABLE serial_numbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    serial_number VARCHAR(100) NOT NULL,
    
    -- Current location
    warehouse_id UUID REFERENCES warehouses(id),
    location_id UUID REFERENCES storage_locations(id),
    
    -- Status tracking
    status VARCHAR(20) DEFAULT 'in_stock', -- in_stock, allocated, sold, returned, scrapped
    
    -- Transaction history
    received_date TIMESTAMPTZ,
    received_from VARCHAR(255), -- Supplier, transfer, etc.
    
    shipped_date TIMESTAMPTZ,
    shipped_to VARCHAR(255), -- Customer, transfer, etc.
    
    -- Quality and warranty
    quality_status VARCHAR(20) DEFAULT 'passed', -- passed, failed, pending
    warranty_start_date DATE,
    warranty_end_date DATE,
    
    -- Cost tracking
    unit_cost DECIMAL(15,4),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, item_id, serial_number),
    CONSTRAINT valid_status CHECK (status IN ('in_stock', 'allocated', 'sold', 'returned', 'scrapped')),
    CONSTRAINT valid_quality_status CHECK (quality_status IN ('passed', 'failed', 'pending'))
);

-- Batch/Lot tracking
CREATE TABLE batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    batch_number VARCHAR(100) NOT NULL,
    
    -- Batch properties
    manufacturing_date DATE,
    expiry_date DATE,
    supplier_batch_number VARCHAR(100),
    
    -- Quality information
    quality_status VARCHAR(20) DEFAULT 'approved',
    quality_notes TEXT,
    certificates JSONB, -- Quality certificates and test results
    
    -- Current stock in this batch
    current_quantity DECIMAL(12,4) DEFAULT 0,
    allocated_quantity DECIMAL(12,4) DEFAULT 0,
    
    -- Cost tracking
    unit_cost DECIMAL(15,4),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, item_id, batch_number),
    CONSTRAINT valid_quality_status CHECK (quality_status IN ('approved', 'rejected', 'quarantine', 'pending'))
);
```

##  Stock Transactions

### Stock Movement Framework

```sql
-- Stock transactions (all inventory movements)
CREATE TABLE stock_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transaction identification
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    transaction_type VARCHAR(50) NOT NULL, -- receipt, issue, transfer, adjustment, opening
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posting_date DATE NOT NULL DEFAULT CURRENT_DATE,
    
    -- Item and location details
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    from_warehouse_id UUID REFERENCES warehouses(id),
    to_warehouse_id UUID REFERENCES warehouses(id),
    from_location_id UUID REFERENCES storage_locations(id),
    to_location_id UUID REFERENCES storage_locations(id),
    
    -- Quantity and valuation
    quantity DECIMAL(12,4) NOT NULL,
    unit_of_measure VARCHAR(20) NOT NULL,
    unit_cost DECIMAL(15,4),
    total_value DECIMAL(15,2),
    
    -- Batch and serial tracking
    batch_id UUID REFERENCES batches(id),
    serial_number_id UUID REFERENCES serial_numbers(id),
    
    -- Transaction source
    source_document_type VARCHAR(50), -- purchase_order, sales_order, transfer_order, etc.
    source_document_id UUID,
    source_document_line_id UUID,
    
    -- Approval and workflow
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, posted, cancelled
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Financial impact
    affects_valuation BOOLEAN DEFAULT true,
    gl_posted BOOLEAN DEFAULT false,
    gl_entry_id UUID REFERENCES journal_entries(id),
    
    -- Quality control
    quality_inspection_required BOOLEAN DEFAULT false,
    quality_status VARCHAR(20) DEFAULT 'pending',
    quality_approved_by UUID REFERENCES users(id),
    quality_approved_at TIMESTAMPTZ,
    
    -- Reference and notes
    reference VARCHAR(255),
    notes TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_transaction_type CHECK (transaction_type IN (
        'receipt', 'issue', 'transfer', 'adjustment', 'opening', 'closing'
    )),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'submitted', 'approved', 'posted', 'cancelled')),
    CONSTRAINT valid_quality_status CHECK (quality_status IN ('pending', 'passed', 'failed', 'not_required')),
    CONSTRAINT quantity_not_zero CHECK (quantity != 0)
);

-- Stock transaction triggers for automatic stock level updates
CREATE OR REPLACE FUNCTION update_stock_levels()
RETURNS TRIGGER AS $$
BEGIN
    -- Handle different transaction types
    IF NEW.transaction_type IN ('receipt', 'opening') THEN
        -- Increase stock in destination warehouse
        INSERT INTO stock_levels (tenant_id, item_id, warehouse_id, location_id, on_hand_quantity, average_cost, total_value)
        VALUES (NEW.tenant_id, NEW.item_id, NEW.to_warehouse_id, NEW.to_location_id, NEW.quantity, NEW.unit_cost, NEW.total_value)
        ON CONFLICT (item_id, warehouse_id, COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID))
        DO UPDATE SET
            on_hand_quantity = stock_levels.on_hand_quantity + NEW.quantity,
            average_cost = CASE 
                WHEN stock_levels.on_hand_quantity + NEW.quantity > 0 
                THEN ((stock_levels.on_hand_quantity * stock_levels.average_cost) + NEW.total_value) / (stock_levels.on_hand_quantity + NEW.quantity)
                ELSE NEW.unit_cost
            END,
            total_value = stock_levels.total_value + NEW.total_value,
            last_movement_date = NEW.transaction_date,
            updated_at = NOW();
            
    ELSIF NEW.transaction_type IN ('issue', 'closing') THEN
        -- Decrease stock in source warehouse
        UPDATE stock_levels 
        SET 
            on_hand_quantity = on_hand_quantity - NEW.quantity,
            total_value = total_value - NEW.total_value,
            last_movement_date = NEW.transaction_date,
            updated_at = NOW()
        WHERE item_id = NEW.item_id 
          AND warehouse_id = NEW.from_warehouse_id 
          AND COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID) = COALESCE(NEW.from_location_id, '00000000-0000-0000-0000-000000000000'::UUID);
          
    ELSIF NEW.transaction_type = 'transfer' THEN
        -- Decrease from source
        UPDATE stock_levels 
        SET 
            on_hand_quantity = on_hand_quantity - NEW.quantity,
            total_value = total_value - NEW.total_value,
            last_movement_date = NEW.transaction_date,
            updated_at = NOW()
        WHERE item_id = NEW.item_id 
          AND warehouse_id = NEW.from_warehouse_id 
          AND COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID) = COALESCE(NEW.from_location_id, '00000000-0000-0000-0000-000000000000'::UUID);
          
        -- Increase in destination
        INSERT INTO stock_levels (tenant_id, item_id, warehouse_id, location_id, on_hand_quantity, average_cost, total_value)
        VALUES (NEW.tenant_id, NEW.item_id, NEW.to_warehouse_id, NEW.to_location_id, NEW.quantity, NEW.unit_cost, NEW.total_value)
        ON CONFLICT (item_id, warehouse_id, COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID))
        DO UPDATE SET
            on_hand_quantity = stock_levels.on_hand_quantity + NEW.quantity,
            total_value = stock_levels.total_value + NEW.total_value,
            last_movement_date = NEW.transaction_date,
            updated_at = NOW();
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_stock_levels
    AFTER INSERT ON stock_transactions
    FOR EACH ROW
    WHEN (NEW.status = 'posted')
    EXECUTE FUNCTION update_stock_levels();
```

##  Procurement Management

### Purchase Order Management

```sql
-- Purchase orders
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- PO identification
    po_number VARCHAR(50) UNIQUE NOT NULL,
    vendor_id UUID NOT NULL REFERENCES vendors(id),
    
    -- PO dates
    po_date DATE NOT NULL DEFAULT CURRENT_DATE,
    required_date DATE,
    promised_date DATE,
    
    -- Delivery information
    delivery_address JSONB,
    delivery_instructions TEXT,
    
    -- Financial details
    subtotal DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    freight_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    
    -- Currency and payment
    currency_code CHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1,
    payment_terms_days INTEGER DEFAULT 30,
    
    -- PO status and workflow
    status VARCHAR(20) DEFAULT 'draft', -- draft, sent, confirmed, received, closed, cancelled
    approval_status VARCHAR(20) DEFAULT 'pending',
    
    -- Quantities tracking
    total_ordered_qty DECIMAL(12,4) DEFAULT 0,
    total_received_qty DECIMAL(12,4) DEFAULT 0,
    total_invoiced_qty DECIMAL(12,4) DEFAULT 0,
    
    -- References
    material_request_id UUID REFERENCES material_requests(id),
    contract_id UUID REFERENCES contracts(id),
    
    -- Workflow tracking
    requested_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Terms and conditions
    terms_and_conditions TEXT,
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_status CHECK (status IN ('draft', 'sent', 'confirmed', 'received', 'closed', 'cancelled')),
    CONSTRAINT valid_approval_status CHECK (approval_status IN ('pending', 'approved', 'rejected'))
);

-- Purchase order line items
CREATE TABLE purchase_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    
    -- Item details
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    item_code VARCHAR(100),
    item_description TEXT,
    
    -- Quantity and pricing
    ordered_quantity DECIMAL(12,4) NOT NULL,
    received_quantity DECIMAL(12,4) DEFAULT 0,
    invoiced_quantity DECIMAL(12,4) DEFAULT 0,
    
    unit_of_measure VARCHAR(20),
    unit_price DECIMAL(15,4) NOT NULL,
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    line_total DECIMAL(15,2),
    
    -- Delivery
    required_date DATE,
    promised_date DATE,
    warehouse_id UUID REFERENCES warehouses(id),
    location_id UUID REFERENCES storage_locations(id),
    
    -- Quality requirements
    quality_inspection_required BOOLEAN DEFAULT false,
    quality_specifications TEXT,
    
    -- Tax
    tax_code VARCHAR(20),
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    
    -- Line status
    line_status VARCHAR(20) DEFAULT 'open', -- open, received, closed, cancelled
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(purchase_order_id, line_number),
    CONSTRAINT valid_line_status CHECK (line_status IN ('open', 'received', 'closed', 'cancelled'))
);
```

### Material Request Workflow

```sql
-- Material requests (internal requisitions)
CREATE TABLE material_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Request identification
    request_number VARCHAR(50) UNIQUE NOT NULL,
    request_type VARCHAR(20) DEFAULT 'purchase', -- purchase, transfer, manufacture
    
    -- Request details
    request_date DATE NOT NULL DEFAULT CURRENT_DATE,
    required_date DATE NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, urgent
    
    -- Requesting details
    requested_by UUID NOT NULL REFERENCES users(id),
    department_id UUID REFERENCES organizations(id),
    cost_center_id UUID REFERENCES cost_centers(id),
    project_id UUID REFERENCES projects(id),
    
    -- Delivery information
    deliver_to_warehouse_id UUID REFERENCES warehouses(id),
    deliver_to_location_id UUID REFERENCES storage_locations(id),
    
    -- Approval workflow
    status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, approved, rejected, completed
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- Purpose and justification
    purpose TEXT,
    justification TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_request_type CHECK (request_type IN ('purchase', 'transfer', 'manufacture')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'submitted', 'approved', 'rejected', 'completed'))
);

-- Material request line items
CREATE TABLE material_request_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_request_id UUID NOT NULL REFERENCES material_requests(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    
    -- Item details
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    item_code VARCHAR(100),
    item_description TEXT,
    
    -- Quantity requirements
    requested_quantity DECIMAL(12,4) NOT NULL,
    approved_quantity DECIMAL(12,4),
    ordered_quantity DECIMAL(12,4) DEFAULT 0,
    received_quantity DECIMAL(12,4) DEFAULT 0,
    
    unit_of_measure VARCHAR(20),
    estimated_cost DECIMAL(15,4),
    
    -- Requirements
    required_date DATE,
    quality_specifications TEXT,
    
    -- Source preference
    preferred_vendor_id UUID REFERENCES vendors(id),
    
    -- Line status
    line_status VARCHAR(20) DEFAULT 'pending', -- pending, approved, ordered, received, cancelled
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(material_request_id, line_number),
    CONSTRAINT valid_line_status CHECK (line_status IN ('pending', 'approved', 'ordered', 'received', 'cancelled'))
);
```

##  Inventory Analytics & Optimization

### ABC Analysis and Classification

```typescript
interface ABCAnalysisParams {
  tenant_id: string;
  analysis_period: DateRange;
  classification_criteria: 'value' | 'quantity' | 'profit_margin' | 'turnover_rate';
  category_thresholds: {
    a_category_percentage: number; // e.g., 80% of value
    b_category_percentage: number; // e.g., 15% of value
    c_category_percentage: number; // e.g., 5% of value
  };
}

interface ABCClassificationResult {
  item_id: string;
  item_code: string;
  item_name: string;
  category: 'A' | 'B' | 'C';
  
  analysis_metrics: {
    total_value: number;
    total_quantity: number;
    turnover_rate: number;
    profit_margin: number;
    movement_frequency: number;
  };
  
  percentage_of_total_value: number;
  cumulative_percentage: number;
  
  recommendations: {
    reorder_frequency: string;
    safety_stock_level: number;
    inventory_review_frequency: string;
    storage_priority: 'high' | 'medium' | 'low';
  };
}

class InventoryAnalyticsService {
  async performABCAnalysis(params: ABCAnalysisParams): Promise<ABCClassificationResult[]> {
    // Get inventory movement data for the analysis period
    const movementData = await this.getInventoryMovements(params.tenant_id, params.analysis_period);
    
    // Calculate metrics for each item
    const itemMetrics = await this.calculateItemMetrics(movementData, params.classification_criteria);
    
    // Sort items by the selected criteria (descending)
    itemMetrics.sort((a, b) => b.criteriaValue - a.criteriaValue);
    
    // Calculate cumulative percentages
    const totalValue = itemMetrics.reduce((sum, item) => sum + item.criteriaValue, 0);
    let cumulativeValue = 0;
    
    // Classify items into A, B, C categories
    const results: ABCClassificationResult[] = itemMetrics.map(item => {
      cumulativeValue += item.criteriaValue;
      const cumulativePercentage = (cumulativeValue / totalValue) * 100;
      
      let category: 'A' | 'B' | 'C';
      if (cumulativePercentage <= params.category_thresholds.a_category_percentage) {
        category = 'A';
      } else if (cumulativePercentage <= params.category_thresholds.a_category_percentage + params.category_thresholds.b_category_percentage) {
        category = 'B';
      } else {
        category = 'C';
      }
      
      return {
        item_id: item.item_id,
        item_code: item.item_code,
        item_name: item.item_name,
        category,
        analysis_metrics: item.metrics,
        percentage_of_total_value: (item.criteriaValue / totalValue) * 100,
        cumulative_percentage: cumulativePercentage,
        recommendations: this.generateRecommendations(category, item.metrics)
      };
    });
    
    return results;
  }
  
  private generateRecommendations(category: 'A' | 'B' | 'C', metrics: any) {
    const recommendations = {
      A: {
        reorder_frequency: 'Weekly',
        safety_stock_multiplier: 1.5,
        inventory_review_frequency: 'Daily',
        storage_priority: 'high' as const
      },
      B: {
        reorder_frequency: 'Bi-weekly',
        safety_stock_multiplier: 1.2,
        inventory_review_frequency: 'Weekly',
        storage_priority: 'medium' as const
      },
      C: {
        reorder_frequency: 'Monthly',
        safety_stock_multiplier: 1.0,
        inventory_review_frequency: 'Monthly',
        storage_priority: 'low' as const
      }
    };
    
    const baseRecommendation = recommendations[category];
    
    return {
      reorder_frequency: baseRecommendation.reorder_frequency,
      safety_stock_level: metrics.average_monthly_usage * baseRecommendation.safety_stock_multiplier,
      inventory_review_frequency: baseRecommendation.inventory_review_frequency,
      storage_priority: baseRecommendation.storage_priority
    };
  }
}
```

### Inventory Turnover and Dead Stock Analysis

```sql
-- Dead stock analysis function
CREATE OR REPLACE FUNCTION analyze_dead_stock(
    p_tenant_id UUID,
    p_no_movement_days INTEGER DEFAULT 90,
    p_low_turnover_threshold DECIMAL DEFAULT 2.0
)
RETURNS TABLE (
    item_id UUID,
    item_code VARCHAR,
    item_name VARCHAR,
    warehouse_id UUID,
    current_quantity DECIMAL,
    current_value DECIMAL,
    last_movement_date TIMESTAMPTZ,
    days_without_movement INTEGER,
    annual_turnover_rate DECIMAL,
    category VARCHAR,
    recommended_action VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    WITH item_movements AS (
        SELECT 
            st.item_id,
            st.to_warehouse_id as warehouse_id,
            MAX(st.transaction_date) as last_movement_date,
            COUNT(*) as movement_count,
            SUM(st.quantity) as total_quantity_moved
        FROM stock_transactions st
        WHERE st.tenant_id = p_tenant_id
          AND st.transaction_date >= CURRENT_DATE - INTERVAL '1 year'
          AND st.transaction_type IN ('issue', 'receipt')
        GROUP BY st.item_id, st.to_warehouse_id
    ),
    current_stock AS (
        SELECT 
            sl.item_id,
            sl.warehouse_id,
            sl.on_hand_quantity,
            sl.total_value,
            COALESCE(im.last_movement_date, sl.last_movement_date) as last_movement_date,
            COALESCE(im.total_quantity_moved, 0) as annual_quantity_moved
        FROM stock_levels sl
        LEFT JOIN item_movements im ON sl.item_id = im.item_id AND sl.warehouse_id = im.warehouse_id
        WHERE sl.tenant_id = p_tenant_id
          AND sl.on_hand_quantity > 0
    )
    SELECT 
        cs.item_id,
        ii.item_code,
        ii.item_name,
        cs.warehouse_id,
        cs.on_hand_quantity,
        cs.total_value,
        cs.last_movement_date,
        EXTRACT(DAY FROM (CURRENT_TIMESTAMP - cs.last_movement_date))::INTEGER as days_without_movement,
        CASE 
            WHEN cs.on_hand_quantity > 0 AND cs.annual_quantity_moved > 0 
            THEN cs.annual_quantity_moved / cs.on_hand_quantity
            ELSE 0
        END as annual_turnover_rate,
        CASE 
            WHEN cs.last_movement_date IS NULL OR 
                 EXTRACT(DAY FROM (CURRENT_TIMESTAMP - cs.last_movement_date)) > p_no_movement_days * 2
            THEN 'Dead Stock'
            WHEN EXTRACT(DAY FROM (CURRENT_TIMESTAMP - cs.last_movement_date)) > p_no_movement_days
            THEN 'Slow Moving'
            WHEN cs.annual_quantity_moved / GREATEST(cs.on_hand_quantity, 1) < p_low_turnover_threshold
            THEN 'Low Turnover'
            ELSE 'Normal'
        END as category,
        CASE 
            WHEN cs.last_movement_date IS NULL OR 
                 EXTRACT(DAY FROM (CURRENT_TIMESTAMP - cs.last_movement_date)) > p_no_movement_days * 2
            THEN 'Consider liquidation or write-off'
            WHEN EXTRACT(DAY FROM (CURRENT_TIMESTAMP - cs.last_movement_date)) > p_no_movement_days
            THEN 'Investigate demand patterns, consider promotion'
            WHEN cs.annual_quantity_moved / GREATEST(cs.on_hand_quantity, 1) < p_low_turnover_threshold
            THEN 'Reduce reorder quantities, optimize stock levels'
            ELSE 'Monitor regularly'
        END as recommended_action
    FROM current_stock cs
    JOIN inventory_items ii ON cs.item_id = ii.id
    WHERE cs.last_movement_date IS NULL 
       OR EXTRACT(DAY FROM (CURRENT_TIMESTAMP - cs.last_movement_date)) > p_no_movement_days
       OR (cs.annual_quantity_moved / GREATEST(cs.on_hand_quantity, 1)) < p_low_turnover_threshold
    ORDER BY days_without_movement DESC, cs.total_value DESC;
END;
$$ LANGUAGE plpgsql;
```

##  Cycle Counting & Physical Inventory

### Cycle Count Management

```sql
-- Cycle count schedules
CREATE TABLE cycle_count_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    schedule_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Schedule parameters
    frequency VARCHAR(20) NOT NULL, -- daily, weekly, monthly, quarterly, annual
    schedule_type VARCHAR(20) DEFAULT 'abc_based', -- abc_based, location_based, random, value_based
    
    -- ABC-based scheduling
    count_a_items_frequency INTEGER DEFAULT 30, -- days
    count_b_items_frequency INTEGER DEFAULT 90, -- days
    count_c_items_frequency INTEGER DEFAULT 365, -- days
    
    -- Location-based scheduling
    target_locations JSONB, -- Array of location IDs or patterns
    
    -- Value-based scheduling
    high_value_threshold DECIMAL(15,2),
    high_value_frequency INTEGER DEFAULT 30, -- days
    
    -- Schedule status
    is_active BOOLEAN DEFAULT true,
    next_generation_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_frequency CHECK (frequency IN ('daily', 'weekly', 'monthly', 'quarterly', 'annual')),
    CONSTRAINT valid_schedule_type CHECK (schedule_type IN ('abc_based', 'location_based', 'random', 'value_based'))
);

-- Individual cycle counts
CREATE TABLE cycle_counts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Count identification
    count_number VARCHAR(50) UNIQUE NOT NULL,
    count_date DATE NOT NULL DEFAULT CURRENT_DATE,
    count_type VARCHAR(20) DEFAULT 'cycle', -- cycle, physical, spot
    
    -- Scope
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    location_id UUID REFERENCES storage_locations(id),
    schedule_id UUID REFERENCES cycle_count_schedules(id),
    
    -- Status and workflow
    status VARCHAR(20) DEFAULT 'planned', -- planned, in_progress, completed, approved, posted
    assigned_to UUID REFERENCES users(id),
    counted_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    
    -- Count details
    planned_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    
    -- Results summary
    total_items_counted INTEGER DEFAULT 0,
    items_with_variances INTEGER DEFAULT 0,
    total_variance_value DECIMAL(15,2) DEFAULT 0,
    
    notes TEXT,
    
    CONSTRAINT valid_count_type CHECK (count_type IN ('cycle', 'physical', 'spot')),
    CONSTRAINT valid_status CHECK (status IN ('planned', 'in_progress', 'completed', 'approved', 'posted'))
);

-- Cycle count line items
CREATE TABLE cycle_count_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_count_id UUID NOT NULL REFERENCES cycle_counts(id) ON DELETE CASCADE,
    
    -- Item and location
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    location_id UUID REFERENCES storage_locations(id),
    batch_id UUID REFERENCES batches(id),
    
    -- Expected vs actual quantities
    expected_quantity DECIMAL(12,4) NOT NULL DEFAULT 0,
    counted_quantity DECIMAL(12,4),
    variance_quantity DECIMAL(12,4),
    
    -- Valuation
    unit_cost DECIMAL(15,4),
    variance_value DECIMAL(15,2),
    
    -- Count details
    count_sequence INTEGER,
    counted_at TIMESTAMPTZ,
    counted_by UUID REFERENCES users(id),
    
    -- Variance investigation
    variance_reason VARCHAR(100),
    variance_explanation TEXT,
    requires_investigation BOOLEAN DEFAULT false,
    investigation_completed BOOLEAN DEFAULT false,
    
    -- Adjustment tracking
    adjustment_posted BOOLEAN DEFAULT false,
    stock_transaction_id UUID REFERENCES stock_transactions(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Automated Count Generation

```typescript
interface CycleCountGenerator {
  generateCycleCounts(scheduleId: string, generationDate: Date): Promise<CycleCount[]>;
  generateABCBasedCounts(params: ABCCountParams): Promise<CycleCountLine[]>;
  generateLocationBasedCounts(params: LocationCountParams): Promise<CycleCountLine[]>;
  generateValueBasedCounts(params: ValueCountParams): Promise<CycleCountLine[]>;
}

class CycleCountService implements CycleCountGenerator {
  async generateCycleCounts(scheduleId: string, generationDate: Date): Promise<CycleCount[]> {
    const schedule = await this.getSchedule(scheduleId);
    const counts: CycleCount[] = [];
    
    switch (schedule.schedule_type) {
      case 'abc_based':
        const abcCounts = await this.generateABCBasedCounts({
          tenant_id: schedule.tenant_id,
          generation_date: generationDate,
          a_frequency_days: schedule.count_a_items_frequency,
          b_frequency_days: schedule.count_b_items_frequency,
          c_frequency_days: schedule.count_c_items_frequency
        });
        counts.push(...abcCounts);
        break;
        
      case 'location_based':
        const locationCounts = await this.generateLocationBasedCounts({
          tenant_id: schedule.tenant_id,
          target_locations: schedule.target_locations,
          generation_date: generationDate
        });
        counts.push(...locationCounts);
        break;
        
      case 'value_based':
        const valueCounts = await this.generateValueBasedCounts({
          tenant_id: schedule.tenant_id,
          high_value_threshold: schedule.high_value_threshold,
          high_value_frequency: schedule.high_value_frequency,
          generation_date: generationDate
        });
        counts.push(...valueCounts);
        break;
    }
    
    return counts;
  }
  
  async generateABCBasedCounts(params: ABCCountParams): Promise<CycleCount[]> {
    // Get ABC classification results
    const abcAnalysis = await this.getLatestABCAnalysis(params.tenant_id);
    
    // Group items by category and determine which need counting
    const itemsToCount = abcAnalysis.filter(item => {
      const daysSinceLastCount = this.getDaysSinceLastCount(item.item_id);
      
      switch (item.category) {
        case 'A':
          return daysSinceLastCount >= params.a_frequency_days;
        case 'B':
          return daysSinceLastCount >= params.b_frequency_days;
        case 'C':
          return daysSinceLastCount >= params.c_frequency_days;
        default:
          return false;
      }
    });
    
    // Group by warehouse and create cycle counts
    const countsByWarehouse = this.groupItemsByWarehouse(itemsToCount);
    const cycleCounts: CycleCount[] = [];
    
    for (const [warehouseId, items] of countsByWarehouse.entries()) {
      const cycleCount = await this.createCycleCount({
        tenant_id: params.tenant_id,
        warehouse_id: warehouseId,
        count_type: 'cycle',
        count_date: params.generation_date,
        items: items
      });
      
      cycleCounts.push(cycleCount);
    }
    
    return cycleCounts;
  }
}
```

This  inventory management system provides the foundation for efficient stock control, procurement optimization, and warehouse operations while maintaining accurate financial valuation and supporting various industry-specific requirements.
