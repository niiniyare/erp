# Forecourt Management

## ⛽ Overview

The Forecourt Management module transforms the core ERP into a  fuel retail and convenience store management platform. It provides fuel inventory management, pump operations, environmental compliance, fleet card processing, and integrated retail operations designed for gas stations, truck stops, and fuel distribution centers.

## ️ Fuel Management System

### Tank and Fuel Inventory

```sql
-- Fuel storage tanks
CREATE TABLE fuel_tanks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Tank identification
    tank_number VARCHAR(20) NOT NULL,
    tank_name VARCHAR(255),
    site_id UUID NOT NULL REFERENCES retail_stores(id), -- References fuel station location
    
    -- Tank specifications
    tank_type VARCHAR(30) NOT NULL, -- underground, aboveground, mobile
    tank_material VARCHAR(30) DEFAULT 'steel', -- steel, fiberglass, composite
    installation_date DATE,
    manufacturer VARCHAR(100),
    model_number VARCHAR(100),
    
    -- Capacity and dimensions
    total_capacity_liters DECIMAL(12,2) NOT NULL,
    usable_capacity_liters DECIMAL(12,2) NOT NULL,
    minimum_operating_level_liters DECIMAL(10,2) DEFAULT 500,
    maximum_fill_level_liters DECIMAL(12,2),
    
    -- Fuel product
    fuel_product_id UUID NOT NULL REFERENCES fuel_products(id),
    octane_rating INTEGER, -- For gasoline
    fuel_grade VARCHAR(50), -- regular, mid_grade, premium, diesel, etc.
    
    -- Tank status and monitoring
    tank_status VARCHAR(20) DEFAULT 'active', -- active, maintenance, out_of_service, decommissioned
    current_volume_liters DECIMAL(12,2) DEFAULT 0,
    last_delivery_date DATE,
    last_calibration_date DATE,
    next_calibration_due_date DATE,
    
    -- Environmental compliance
    leak_detection_system VARCHAR(50), -- continuous, monthly, manual
    leak_detection_last_test DATE,
    vapor_recovery_system BOOLEAN DEFAULT false,
    spill_containment_volume_liters DECIMAL(10,2),
    
    -- Monitoring equipment
    atg_probe_installed BOOLEAN DEFAULT true, -- Automatic Tank Gauging
    atg_manufacturer VARCHAR(100),
    atg_model VARCHAR(100),
    
    -- Safety and regulations
    permit_number VARCHAR(100),
    permit_expiry_date DATE,
    insurance_coverage DECIMAL(15,2),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_tank_type CHECK (tank_type IN ('underground', 'aboveground', 'mobile')),
    CONSTRAINT valid_tank_status CHECK (tank_status IN ('active', 'maintenance', 'out_of_service', 'decommissioned')),
    UNIQUE(site_id, tank_number)
);

-- Fuel products and grades
CREATE TABLE fuel_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Product identification
    product_code VARCHAR(20) UNIQUE NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    fuel_type VARCHAR(30) NOT NULL, -- gasoline, diesel, e85, biodiesel, propane, natural_gas
    
    -- Product specifications
    octane_rating INTEGER, -- For gasoline (87, 89, 91, 93, etc.)
    cetane_rating INTEGER, -- For diesel
    ethanol_content_percentage DECIMAL(5,2), -- For ethanol blends
    biodiesel_content_percentage DECIMAL(5,2), -- For biodiesel blends
    
    -- Environmental properties
    reid_vapor_pressure DECIMAL(6,3), -- RVP for gasoline
    sulfur_content_ppm INTEGER, -- Parts per million
    carbon_content_percentage DECIMAL(5,2),
    
    -- Regulatory information
    epa_fuel_code VARCHAR(20),
    carb_fuel_code VARCHAR(20), -- California Air Resources Board
    renewable_fuel_standard BOOLEAN DEFAULT false,
    
    -- Pricing and tax
    federal_excise_tax_rate DECIMAL(8,5), -- Per gallon
    state_excise_tax_rate DECIMAL(8,5),
    environmental_fee_rate DECIMAL(8,5),
    
    -- Seasonal availability
    summer_blend BOOLEAN DEFAULT false,
    winter_blend BOOLEAN DEFAULT false,
    reformulated_gasoline BOOLEAN DEFAULT false,
    
    -- Product status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_fuel_type CHECK (fuel_type IN ('gasoline', 'diesel', 'e85', 'biodiesel', 'propane', 'natural_gas'))
);

-- Tank level readings and monitoring
CREATE TABLE tank_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tank_id UUID NOT NULL REFERENCES fuel_tanks(id) ON DELETE CASCADE,
    
    -- Reading details
    reading_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reading_type VARCHAR(20) DEFAULT 'automatic', -- automatic, manual, delivery, calibration
    
    -- Volume measurements
    fuel_volume_liters DECIMAL(12,4) NOT NULL,
    water_volume_liters DECIMAL(8,4) DEFAULT 0,
    product_height_cm DECIMAL(8,2),
    water_height_cm DECIMAL(8,2) DEFAULT 0,
    temperature_celsius DECIMAL(5,2),
    
    -- Calculated values
    ullage_liters DECIMAL(12,2), -- Empty space in tank
    gross_volume_liters DECIMAL(12,4), -- Volume at current temperature
    net_volume_liters DECIMAL(12,4), -- Volume corrected to standard temperature
    
    -- Quality measurements
    density_kg_per_liter DECIMAL(6,4),
    api_gravity DECIMAL(6,3), -- For petroleum products
    
    -- System information
    reading_source VARCHAR(50), -- atg_system, manual_gauge, delivery_receipt
    reading_device_id VARCHAR(100),
    
    -- Alerts and exceptions
    low_level_alert BOOLEAN DEFAULT false,
    high_level_alert BOOLEAN DEFAULT false,
    water_alert BOOLEAN DEFAULT false,
    leak_alert BOOLEAN DEFAULT false,
    temperature_alert BOOLEAN DEFAULT false,
    
    -- Quality and validation
    reading_validated BOOLEAN DEFAULT true,
    validation_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_reading_type CHECK (reading_type IN ('automatic', 'manual', 'delivery', 'calibration'))
);

-- Fuel deliveries
CREATE TABLE fuel_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Delivery identification
    delivery_number VARCHAR(50) UNIQUE NOT NULL,
    bill_of_lading_number VARCHAR(100),
    
    -- Delivery details
    tank_id UUID NOT NULL REFERENCES fuel_tanks(id),
    supplier_id UUID NOT NULL REFERENCES vendors(id),
    carrier_company VARCHAR(255),
    driver_name VARCHAR(255),
    truck_number VARCHAR(50),
    
    -- Delivery timing
    scheduled_delivery_date DATE,
    actual_delivery_date DATE NOT NULL,
    delivery_start_time TIMESTAMPTZ,
    delivery_end_time TIMESTAMPTZ,
    
    -- Fuel quantities
    gross_gallons DECIMAL(12,4) NOT NULL, -- At delivery temperature
    net_gallons DECIMAL(12,4) NOT NULL, -- Corrected to standard temperature
    temperature_fahrenheit DECIMAL(5,2),
    
    -- Before and after readings
    tank_volume_before_liters DECIMAL(12,2),
    tank_volume_after_liters DECIMAL(12,2),
    volume_variance_liters DECIMAL(10,2),
    
    -- Fuel quality
    octane_rating INTEGER,
    specific_gravity DECIMAL(6,4),
    reid_vapor_pressure DECIMAL(6,3),
    water_content_ppm INTEGER DEFAULT 0,
    
    -- Pricing and costs
    unit_price_per_gallon DECIMAL(8,5) NOT NULL,
    total_cost DECIMAL(12,2) NOT NULL,
    freight_cost DECIMAL(8,2) DEFAULT 0,
    taxes_and_fees DECIMAL(8,2) DEFAULT 0,
    
    -- Quality testing
    quality_test_passed BOOLEAN DEFAULT true,
    quality_test_notes TEXT,
    certificates_received BOOLEAN DEFAULT false,
    
    -- Environmental compliance
    vapor_recovery_performed BOOLEAN DEFAULT false,
    spill_occurred BOOLEAN DEFAULT false,
    spill_volume_liters DECIMAL(8,2) DEFAULT 0,
    
    -- Delivery status
    delivery_status VARCHAR(20) DEFAULT 'completed', -- scheduled, in_progress, completed, rejected
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_delivery_status CHECK (delivery_status IN ('scheduled', 'in_progress', 'completed', 'rejected'))
);
```

### Pump Operations & Dispensing

```sql
-- Fuel dispensers (pumps)
CREATE TABLE fuel_dispensers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Dispenser identification
    dispenser_number VARCHAR(20) NOT NULL,
    site_id UUID NOT NULL REFERENCES retail_stores(id),
    
    -- Dispenser specifications
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    serial_number VARCHAR(100),
    installation_date DATE,
    
    -- Dispenser configuration
    number_of_hoses INTEGER DEFAULT 2,
    number_of_grades INTEGER DEFAULT 3,
    multi_product_dispenser BOOLEAN DEFAULT true,
    blended_products_supported BOOLEAN DEFAULT false,
    
    -- Connected tanks
    connected_tanks JSONB, -- Array of tank IDs and grade mappings
    
    -- Display and interface
    display_type VARCHAR(30) DEFAULT 'lcd', -- lcd, led, mechanical
    price_sign_integrated BOOLEAN DEFAULT true,
    payment_terminal_integrated BOOLEAN DEFAULT true,
    
    -- Dispensing capabilities
    maximum_flow_rate_lpm DECIMAL(6,2), -- Liters per minute
    minimum_dispensing_amount DECIMAL(6,2) DEFAULT 0.01,
    maximum_dispensing_amount DECIMAL(8,2) DEFAULT 999.99,
    
    -- Safety features
    emergency_stop_button BOOLEAN DEFAULT true,
    vapor_recovery_system BOOLEAN DEFAULT false,
    leak_detection_system BOOLEAN DEFAULT true,
    automatic_shutoff BOOLEAN DEFAULT true,
    
    -- Maintenance
    last_calibration_date DATE,
    next_calibration_due_date DATE,
    last_maintenance_date DATE,
    maintenance_interval_days INTEGER DEFAULT 30,
    
    -- Status
    dispenser_status VARCHAR(20) DEFAULT 'active', -- active, maintenance, out_of_order, decommissioned
    
    -- Regulatory compliance
    weights_measures_seal_number VARCHAR(50),
    seal_date DATE,
    permit_numbers JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_dispenser_status CHECK (dispenser_status IN ('active', 'maintenance', 'out_of_order', 'decommissioned')),
    UNIQUE(site_id, dispenser_number)
);

-- Fuel transactions (individual fuel sales)
CREATE TABLE fuel_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transaction identification
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    pos_transaction_id VARCHAR(100), -- Link to POS system
    
    -- Dispenser and location
    dispenser_id UUID NOT NULL REFERENCES fuel_dispensers(id),
    hose_number INTEGER NOT NULL,
    
    -- Customer and payment
    customer_id UUID REFERENCES customers(id),
    fleet_card_number VARCHAR(50),
    payment_method VARCHAR(30) NOT NULL, -- cash, credit_card, debit_card, fleet_card, prepaid
    
    -- Fuel details
    fuel_product_id UUID NOT NULL REFERENCES fuel_products(id),
    fuel_grade VARCHAR(50),
    
    -- Transaction amounts
    gallons_dispensed DECIMAL(10,4) NOT NULL,
    price_per_gallon DECIMAL(6,4) NOT NULL,
    fuel_amount DECIMAL(10,2) NOT NULL,
    taxes_amount DECIMAL(8,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    
    -- Transaction timing
    transaction_start_time TIMESTAMPTZ NOT NULL,
    transaction_end_time TIMESTAMPTZ NOT NULL,
    
    -- Vehicle information (for fleet cards)
    vehicle_id VARCHAR(50),
    odometer_reading INTEGER,
    driver_id VARCHAR(50),
    vehicle_license_plate VARCHAR(20),
    
    -- Loyalty and discounts
    loyalty_card_number VARCHAR(50),
    discount_amount DECIMAL(6,2) DEFAULT 0,
    loyalty_points_earned INTEGER DEFAULT 0,
    
    -- Transaction status
    transaction_status VARCHAR(20) DEFAULT 'completed', -- completed, cancelled, voided, disputed
    
    -- Regulatory reporting
    reported_to_state BOOLEAN DEFAULT false,
    state_reporting_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_payment_method CHECK (payment_method IN ('cash', 'credit_card', 'debit_card', 'fleet_card', 'prepaid')),
    CONSTRAINT valid_transaction_status CHECK (transaction_status IN ('completed', 'cancelled', 'voided', 'disputed'))
);

-- Dispenser price management
CREATE TABLE dispenser_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Price identification
    dispenser_id UUID NOT NULL REFERENCES fuel_dispensers(id),
    fuel_product_id UUID NOT NULL REFERENCES fuel_products(id),
    
    -- Pricing details
    current_price DECIMAL(6,4) NOT NULL,
    previous_price DECIMAL(6,4),
    cost_basis DECIMAL(6,4), -- Wholesale cost basis
    margin_amount DECIMAL(6,4),
    margin_percentage DECIMAL(5,2),
    
    -- Price timing
    price_effective_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    price_change_reason VARCHAR(100),
    
    -- Competitive pricing
    competitor_average_price DECIMAL(6,4),
    market_position VARCHAR(20) DEFAULT 'competitive', -- low, competitive, high
    
    -- Promotional pricing
    promotional_price DECIMAL(6,4),
    promotion_start_date TIMESTAMPTZ,
    promotion_end_date TIMESTAMPTZ,
    promotion_description TEXT,
    
    -- Price management
    auto_pricing_enabled BOOLEAN DEFAULT false,
    price_update_frequency_hours INTEGER DEFAULT 24,
    maximum_price_change_percentage DECIMAL(5,2) DEFAULT 10,
    
    -- Approval requirements
    requires_manager_approval BOOLEAN DEFAULT false,
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_market_position CHECK (market_position IN ('low', 'competitive', 'high')),
    UNIQUE(dispenser_id, fuel_product_id)
);
```

##  Convenience Store Integration

### Point of Sale Integration

```sql
-- POS transactions integrated with fuel sales
CREATE TABLE forecourt_pos_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transaction identification
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    register_number INTEGER NOT NULL,
    shift_id UUID REFERENCES employee_shifts(id),
    
    -- Customer and timing
    customer_id UUID REFERENCES customers(id),
    transaction_date DATE NOT NULL,
    transaction_time TIMESTAMPTZ NOT NULL,
    
    -- Transaction type
    transaction_type VARCHAR(30) DEFAULT 'sale', -- sale, return, void, no_sale, payout
    
    -- Associated fuel transaction
    fuel_transaction_id UUID REFERENCES fuel_transactions(id),
    combo_transaction BOOLEAN DEFAULT false, -- Fuel + convenience items
    
    -- Financial totals
    subtotal DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(8,2) DEFAULT 0,
    discount_amount DECIMAL(8,2) DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0,
    
    -- Payment details
    payment_method VARCHAR(30) NOT NULL,
    cash_tendered DECIMAL(10,2),
    change_given DECIMAL(8,2),
    card_last_four CHAR(4),
    authorization_code VARCHAR(20),
    
    -- Cashier information
    cashier_id UUID NOT NULL REFERENCES employees(id),
    manager_override BOOLEAN DEFAULT false,
    override_reason VARCHAR(100),
    
    -- Transaction status
    transaction_status VARCHAR(20) DEFAULT 'completed',
    void_reason VARCHAR(100),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_transaction_type CHECK (transaction_type IN ('sale', 'return', 'void', 'no_sale', 'payout')),
    CONSTRAINT valid_payment_method CHECK (payment_method IN ('cash', 'credit_card', 'debit_card', 'fleet_card', 'mobile_payment', 'gift_card')),
    CONSTRAINT valid_transaction_status CHECK (transaction_status IN ('completed', 'voided', 'returned', 'disputed'))
);

-- POS transaction line items
CREATE TABLE forecourt_pos_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pos_transaction_id UUID NOT NULL REFERENCES forecourt_pos_transactions(id) ON DELETE CASCADE,
    
    -- Item identification
    item_id UUID REFERENCES retail_products(id),
    upc VARCHAR(20),
    sku VARCHAR(100),
    item_description VARCHAR(255) NOT NULL,
    
    -- Category and classification
    category VARCHAR(100),
    item_type VARCHAR(50) DEFAULT 'merchandise', -- merchandise, food_service, lottery, tobacco, alcohol
    
    -- Quantity and pricing
    quantity DECIMAL(8,4) NOT NULL DEFAULT 1,
    unit_price DECIMAL(8,4) NOT NULL,
    regular_price DECIMAL(8,4),
    discount_amount DECIMAL(6,2) DEFAULT 0,
    line_total DECIMAL(10,2) NOT NULL,
    
    -- Tax information
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(6,2) DEFAULT 0,
    tax_exempt BOOLEAN DEFAULT false,
    
    -- Special handling
    age_verification_required BOOLEAN DEFAULT false,
    age_verified BOOLEAN DEFAULT false,
    quantity_restricted BOOLEAN DEFAULT false,
    
    -- Promotions and loyalty
    promotion_applied VARCHAR(100),
    loyalty_points_earned INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Convenience store inventory specific to forecourt
CREATE TABLE convenience_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Product and location
    product_id UUID NOT NULL REFERENCES retail_products(id),
    site_id UUID NOT NULL REFERENCES retail_stores(id),
    
    -- Inventory levels
    current_stock INTEGER DEFAULT 0,
    minimum_stock INTEGER DEFAULT 0,
    maximum_stock INTEGER DEFAULT 0,
    reorder_point INTEGER DEFAULT 0,
    
    -- Product placement
    planogram_position VARCHAR(100), -- Aisle and shelf position
    facing_count INTEGER DEFAULT 1,
    
    -- Sales performance
    daily_sales_velocity DECIMAL(8,2) DEFAULT 0,
    weekly_sales_velocity DECIMAL(8,2) DEFAULT 0,
    monthly_sales_velocity DECIMAL(8,2) DEFAULT 0,
    
    -- Special characteristics
    temperature_controlled BOOLEAN DEFAULT false,
    age_restricted BOOLEAN DEFAULT false,
    high_theft_item BOOLEAN DEFAULT false,
    
    -- Vendor and ordering
    primary_vendor_id UUID REFERENCES vendors(id),
    vendor_pack_size INTEGER DEFAULT 1,
    minimum_order_quantity INTEGER DEFAULT 1,
    
    -- Last activity
    last_sale_date DATE,
    last_received_date DATE,
    last_count_date DATE,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(product_id, site_id)
);
```

## ⚖️ Environmental Compliance & Safety

### Regulatory Compliance Management

```sql
-- Environmental compliance tracking
CREATE TABLE environmental_compliance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Compliance identification
    site_id UUID NOT NULL REFERENCES retail_stores(id),
    compliance_type VARCHAR(50) NOT NULL, -- epa, state, local, osha, dot
    regulation_name VARCHAR(255) NOT NULL,
    regulation_code VARCHAR(100),
    
    -- Compliance requirements
    requirement_description TEXT NOT NULL,
    compliance_frequency VARCHAR(30) NOT NULL, -- daily, weekly, monthly, quarterly, annual, as_needed
    responsible_party VARCHAR(100),
    
    -- Status tracking
    compliance_status VARCHAR(20) DEFAULT 'compliant', -- compliant, non_compliant, pending, exempt
    last_inspection_date DATE,
    next_inspection_due_date DATE,
    
    -- Documentation
    required_documents JSONB, -- Array of required document types
    document_urls JSONB, -- Array of stored document URLs
    certificate_number VARCHAR(100),
    certificate_expiry_date DATE,
    
    -- Violations and issues
    violation_count INTEGER DEFAULT 0,
    last_violation_date DATE,
    total_fines_assessed DECIMAL(12,2) DEFAULT 0,
    
    -- Reporting requirements
    requires_periodic_reporting BOOLEAN DEFAULT false,
    last_report_submitted_date DATE,
    next_report_due_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_compliance_type CHECK (compliance_type IN ('epa', 'state', 'local', 'osha', 'dot')),
    CONSTRAINT valid_compliance_status CHECK (compliance_status IN ('compliant', 'non_compliant', 'pending', 'exempt')),
    CONSTRAINT valid_compliance_frequency CHECK (compliance_frequency IN ('daily', 'weekly', 'monthly', 'quarterly', 'annual', 'as_needed'))
);

-- Leak detection and monitoring
CREATE TABLE leak_detection_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Test identification
    tank_id UUID REFERENCES fuel_tanks(id),
    dispenser_id UUID REFERENCES fuel_dispensers(id),
    piping_segment_id UUID REFERENCES piping_segments(id),
    
    -- Test details
    test_date DATE NOT NULL,
    test_type VARCHAR(50) NOT NULL, -- continuous, monthly, line_tightness, vapor_monitor
    test_method VARCHAR(100),
    conducted_by VARCHAR(255),
    
    -- Test parameters
    test_duration_hours DECIMAL(6,2),
    pressure_test_psi DECIMAL(8,3),
    volume_test_gallons DECIMAL(10,4),
    
    -- Test results
    test_result VARCHAR(20) NOT NULL, -- pass, fail, inconclusive
    leak_detected BOOLEAN DEFAULT false,
    estimated_leak_rate DECIMAL(8,4), -- Gallons per hour
    
    -- Environmental impact
    product_released_gallons DECIMAL(10,4) DEFAULT 0,
    soil_contamination_detected BOOLEAN DEFAULT false,
    groundwater_contamination_detected BOOLEAN DEFAULT false,
    
    -- Response actions
    immediate_actions_taken TEXT,
    repair_required BOOLEAN DEFAULT false,
    repair_completed_date DATE,
    follow_up_testing_required BOOLEAN DEFAULT false,
    
    -- Regulatory reporting
    reported_to_authorities BOOLEAN DEFAULT false,
    reporting_date DATE,
    incident_number VARCHAR(100),
    
    -- Documentation
    test_report_url VARCHAR(500),
    photos_urls JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_test_type CHECK (test_type IN ('continuous', 'monthly', 'line_tightness', 'vapor_monitor')),
    CONSTRAINT valid_test_result CHECK (test_result IN ('pass', 'fail', 'inconclusive'))
);

-- Spill and incident reporting
CREATE TABLE environmental_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Incident identification
    incident_number VARCHAR(50) UNIQUE NOT NULL,
    site_id UUID NOT NULL REFERENCES retail_stores(id),
    
    -- Incident details
    incident_date TIMESTAMPTZ NOT NULL,
    incident_type VARCHAR(50) NOT NULL, -- spill, leak, vapor_release, fire, explosion, other
    incident_location VARCHAR(255),
    
    -- Product involved
    fuel_product_id UUID REFERENCES fuel_products(id),
    estimated_volume_released_gallons DECIMAL(10,4),
    
    -- Cause and description
    root_cause VARCHAR(100),
    incident_description TEXT NOT NULL,
    weather_conditions VARCHAR(100),
    
    -- Environmental impact
    soil_contamination BOOLEAN DEFAULT false,
    surface_water_contamination BOOLEAN DEFAULT false,
    groundwater_contamination BOOLEAN DEFAULT false,
    air_quality_impact BOOLEAN DEFAULT false,
    
    -- Response and cleanup
    immediate_response_actions TEXT,
    cleanup_contractor VARCHAR(255),
    cleanup_start_date DATE,
    cleanup_completion_date DATE,
    cleanup_cost DECIMAL(12,2),
    
    -- Regulatory reporting
    reported_to_epa BOOLEAN DEFAULT false,
    reported_to_state BOOLEAN DEFAULT false,
    reported_to_local BOOLEAN DEFAULT false,
    nrc_report_number VARCHAR(100), -- National Response Center
    
    -- Investigation and follow-up
    investigation_completed BOOLEAN DEFAULT false,
    corrective_actions_required TEXT,
    corrective_actions_completed BOOLEAN DEFAULT false,
    
    -- Documentation
    incident_photos JSONB,
    incident_reports JSONB,
    regulatory_correspondence JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_incident_type CHECK (incident_type IN ('spill', 'leak', 'vapor_release', 'fire', 'explosion', 'other'))
);
```

##  Fleet Card & Commercial Fuel Management

### Fleet Customer Management

```typescript
interface FleetManagementSystem {
  processFleetTransaction(transactionData: FleetTransactionData): Promise<FleetTransactionResult>;
  validateFleetCard(cardNumber: string, pin?: string): Promise<CardValidationResult>;
  manageFuelLimits(fleetAccountId: string, limits: FuelLimits): Promise<void>;
  generateFleetReports(fleetAccountId: string, period: DateRange): Promise<FleetReport>;
}

interface FleetTransactionData {
  card_number: string;
  pin?: string;
  vehicle_id?: string;
  driver_id?: string;
  odometer_reading?: number;
  fuel_type: string;
  gallons: number;
  price_per_gallon: number;
  location_id: string;
  transaction_timestamp: Date;
}

interface FleetTransactionResult {
  approved: boolean;
  transaction_id?: string;
  authorization_code?: string;
  decline_reason?: string;
  remaining_limits: {
    daily_gallons: number;
    daily_amount: number;
    monthly_gallons: number;
    monthly_amount: number;
  };
}

interface FuelLimits {
  daily_gallon_limit: number;
  daily_dollar_limit: number;
  monthly_gallon_limit: number;
  monthly_dollar_limit: number;
  allowed_fuel_types: string[];
  allowed_locations: string[];
  time_restrictions: TimeRestriction[];
  product_restrictions: ProductRestriction[];
}

// Fleet accounts and customers
const fleetAccountsSchema = `
CREATE TABLE fleet_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Account identification
    account_number VARCHAR(50) UNIQUE NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    
    -- Billing and contact
    billing_address JSONB NOT NULL,
    primary_contact_name VARCHAR(255),
    primary_contact_email VARCHAR(255),
    primary_contact_phone VARCHAR(20),
    
    -- Account settings
    account_type VARCHAR(30) DEFAULT 'commercial', -- commercial, government, non_profit
    credit_limit DECIMAL(15,2) DEFAULT 0,
    payment_terms_days INTEGER DEFAULT 30,
    
    -- Fuel program settings
    fuel_discount_percentage DECIMAL(5,2) DEFAULT 0,
    volume_discount_tiers JSONB, -- Volume-based discount structure
    
    -- Controls and restrictions
    pin_required BOOLEAN DEFAULT true,
    odometer_required BOOLEAN DEFAULT false,
    driver_id_required BOOLEAN DEFAULT false,
    
    -- Default limits (can be overridden per card)
    default_daily_gallon_limit DECIMAL(8,2) DEFAULT 999.99,
    default_daily_dollar_limit DECIMAL(10,2) DEFAULT 9999.99,
    default_monthly_gallon_limit DECIMAL(10,2) DEFAULT 99999.99,
    default_monthly_dollar_limit DECIMAL(12,2) DEFAULT 999999.99,
    
    -- Allowed products and locations
    allowed_fuel_products JSONB, -- Array of fuel product IDs
    allowed_locations JSONB, -- Array of site IDs or 'all'
    restricted_locations JSONB, -- Array of restricted site IDs
    
    -- Account status
    account_status VARCHAR(20) DEFAULT 'active', -- active, suspended, closed
    credit_status VARCHAR(20) DEFAULT 'approved', -- approved, hold, suspended
    
    -- Billing and reporting
    billing_cycle VARCHAR(20) DEFAULT 'monthly', -- weekly, monthly, custom
    statement_delivery_method VARCHAR(20) DEFAULT 'email', -- email, postal, electronic
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_account_type CHECK (account_type IN ('commercial', 'government', 'non_profit')),
    CONSTRAINT valid_account_status CHECK (account_status IN ('active', 'suspended', 'closed')),
    CONSTRAINT valid_credit_status CHECK (credit_status IN ('approved', 'hold', 'suspended')),
    CONSTRAINT valid_billing_cycle CHECK (billing_cycle IN ('weekly', 'monthly', 'custom'))
);

-- Fleet cards issued to account
CREATE TABLE fleet_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fleet_account_id UUID NOT NULL REFERENCES fleet_accounts(id) ON DELETE CASCADE,
    
    -- Card identification
    card_number VARCHAR(20) UNIQUE NOT NULL,
    card_type VARCHAR(30) DEFAULT 'fuel_only', -- fuel_only, fuel_and_maintenance, universal
    
    -- Vehicle and driver assignment
    vehicle_id VARCHAR(50),
    vehicle_description VARCHAR(255),
    license_plate VARCHAR(20),
    vin VARCHAR(17),
    assigned_driver_id VARCHAR(50),
    assigned_driver_name VARCHAR(255),
    
    -- Card security
    pin_hash VARCHAR(255), -- Hashed PIN
    security_code VARCHAR(10),
    
    -- Card limits (overrides account defaults if set)
    daily_gallon_limit DECIMAL(8,2),
    daily_dollar_limit DECIMAL(10,2),
    monthly_gallon_limit DECIMAL(10,2),
    monthly_dollar_limit DECIMAL(12,2),
    
    -- Usage restrictions
    allowed_fuel_products JSONB, -- Override account settings if specified
    allowed_day_of_week JSONB, -- Array of allowed days (1-7)
    allowed_time_start TIME,
    allowed_time_end TIME,
    
    -- Current usage tracking
    current_month_gallons DECIMAL(10,2) DEFAULT 0,
    current_month_amount DECIMAL(12,2) DEFAULT 0,
    current_day_gallons DECIMAL(8,2) DEFAULT 0,
    current_day_amount DECIMAL(10,2) DEFAULT 0,
    last_reset_date DATE DEFAULT CURRENT_DATE,
    
    -- Card status
    card_status VARCHAR(20) DEFAULT 'active', -- active, suspended, lost, stolen, expired
    issue_date DATE NOT NULL,
    expiry_date DATE,
    last_used_date DATE,
    
    -- Security flags
    compromised BOOLEAN DEFAULT false,
    fraud_suspected BOOLEAN DEFAULT false,
    velocity_check_enabled BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_card_type CHECK (card_type IN ('fuel_only', 'fuel_and_maintenance', 'universal')),
    CONSTRAINT valid_card_status CHECK (card_status IN ('active', 'suspended', 'lost', 'stolen', 'expired'))
);`;

class FleetCardService implements FleetManagementSystem {
  async processFleetTransaction(transactionData: FleetTransactionData): Promise<FleetTransactionResult> {
    // Step 1: Validate card
    const cardValidation = await this.validateFleetCard(transactionData.card_number, transactionData.pin);
    if (!cardValidation.valid) {
      return {
        approved: false,
        decline_reason: cardValidation.decline_reason
      };
    }
    
    const fleetCard = cardValidation.card;
    const fleetAccount = await this.getFleetAccount(fleetCard.fleet_account_id);
    
    // Step 2: Check limits
    const limitsCheck = await this.checkTransactionLimits(fleetCard, transactionData);
    if (!limitsCheck.approved) {
      return {
        approved: false,
        decline_reason: limitsCheck.reason
      };
    }
    
    // Step 3: Check restrictions
    const restrictionsCheck = this.checkTransactionRestrictions(fleetCard, fleetAccount, transactionData);
    if (!restrictionsCheck.approved) {
      return {
        approved: false,
        decline_reason: restrictionsCheck.reason
      };
    }
    
    // Step 4: Process transaction
    const transaction = await this.createFleetTransaction({
      fleet_card_id: fleetCard.id,
      fleet_account_id: fleetAccount.id,
      ...transactionData,
      authorized: true,
      authorization_code: this.generateAuthorizationCode()
    });
    
    // Step 5: Update usage tracking
    await this.updateCardUsage(fleetCard.id, transactionData.gallons, 
      transactionData.gallons * transactionData.price_per_gallon);
    
    // Step 6: Apply discounts
    const discountedAmount = this.calculateFleetDiscount(
      transactionData.gallons * transactionData.price_per_gallon,
      fleetAccount.fuel_discount_percentage,
      fleetAccount.volume_discount_tiers,
      await this.getMonthlyVolume(fleetAccount.id)
    );
    
    return {
      approved: true,
      transaction_id: transaction.id,
      authorization_code: transaction.authorization_code,
      remaining_limits: await this.getRemainingLimits(fleetCard.id)
    };
  }
  
  async validateFleetCard(cardNumber: string, pin?: string): Promise<CardValidationResult> {
    const card = await this.getFleetCardByNumber(cardNumber);
    
    if (!card) {
      return { valid: false, decline_reason: 'Invalid card number' };
    }
    
    if (card.card_status !== 'active') {
      return { valid: false, decline_reason: `Card is ${card.card_status}` };
    }
    
    if (card.expiry_date && new Date(card.expiry_date) < new Date()) {
      return { valid: false, decline_reason: 'Card expired' };
    }
    
    // Validate PIN if required
    const account = await this.getFleetAccount(card.fleet_account_id);
    if (account.pin_required && pin) {
      const pinValid = await this.validatePin(card.pin_hash, pin);
      if (!pinValid) {
        return { valid: false, decline_reason: 'Invalid PIN' };
      }
    }
    
    // Check account status
    if (account.account_status !== 'active') {
      return { valid: false, decline_reason: `Account is ${account.account_status}` };
    }
    
    if (account.credit_status !== 'approved') {
      return { valid: false, decline_reason: 'Account credit hold' };
    }
    
    return { valid: true, card, account };
  }
  
  private async checkTransactionLimits(card: FleetCard, transaction: FleetTransactionData): Promise<{approved: boolean, reason?: string}> {
    const transactionAmount = transaction.gallons * transaction.price_per_gallon;
    
    // Check daily gallon limit
    const dailyGallonLimit = card.daily_gallon_limit || card.fleet_account.default_daily_gallon_limit;
    if (card.current_day_gallons + transaction.gallons > dailyGallonLimit) {
      return { approved: false, reason: 'Daily gallon limit exceeded' };
    }
    
    // Check daily dollar limit
    const dailyDollarLimit = card.daily_dollar_limit || card.fleet_account.default_daily_dollar_limit;
    if (card.current_day_amount + transactionAmount > dailyDollarLimit) {
      return { approved: false, reason: 'Daily dollar limit exceeded' };
    }
    
    // Check monthly limits
    const monthlyGallonLimit = card.monthly_gallon_limit || card.fleet_account.default_monthly_gallon_limit;
    if (card.current_month_gallons + transaction.gallons > monthlyGallonLimit) {
      return { approved: false, reason: 'Monthly gallon limit exceeded' };
    }
    
    const monthlyDollarLimit = card.monthly_dollar_limit || card.fleet_account.default_monthly_dollar_limit;
    if (card.current_month_amount + transactionAmount > monthlyDollarLimit) {
      return { approved: false, reason: 'Monthly dollar limit exceeded' };
    }
    
    return { approved: true };
  }
  
  private calculateFleetDiscount(
    baseAmount: number, 
    discountPercentage: number, 
    volumeTiers: any[], 
    monthlyVolume: number
  ): number {
    
    let discount = baseAmount * (discountPercentage / 100);
    
    // Apply volume-based discounts
    if (volumeTiers && volumeTiers.length > 0) {
      for (const tier of volumeTiers.sort((a, b) => b.minimum_gallons - a.minimum_gallons)) {
        if (monthlyVolume >= tier.minimum_gallons) {
          const volumeDiscount = baseAmount * (tier.discount_percentage / 100);
          discount = Math.max(discount, volumeDiscount);
          break;
        }
      }
    }
    
    return baseAmount - discount;
  }
}
```

This  forecourt management system provides sophisticated fuel inventory control, pump operations, environmental compliance, and fleet card processing capabilities specifically designed for gas stations and fuel retail operations while integrating seamlessly with the core ERP financial, inventory, and customer management modules.
