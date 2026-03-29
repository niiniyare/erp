# Asset Management

## ️ Overview

The Asset Management module provides  tracking and management of fixed assets throughout their lifecycle, from acquisition to disposal. It includes depreciation calculations, maintenance scheduling, location tracking, and compliance with various accounting standards (GAAP, IFRS).

##  Fixed Asset Management

### Asset Master Data

```sql
-- Fixed assets master table
CREATE TABLE fixed_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Asset identification
    asset_number VARCHAR(50) UNIQUE NOT NULL,
    asset_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Asset classification
    asset_category_id UUID REFERENCES asset_categories(id),
    asset_type VARCHAR(50) NOT NULL, -- building, machinery, vehicle, furniture, it_equipment, etc.
    asset_subtype VARCHAR(100),
    
    -- Manufacturer and model information
    manufacturer VARCHAR(255),
    model VARCHAR(255),
    serial_number VARCHAR(100),
    part_number VARCHAR(100),
    
    -- Physical properties
    specifications JSONB, -- Technical specifications
    dimensions JSONB, -- Length, width, height, unit
    weight DECIMAL(10,4),
    weight_unit VARCHAR(10) DEFAULT 'kg',
    
    -- Asset condition and quality
    condition_rating VARCHAR(20) DEFAULT 'good', -- excellent, good, fair, poor, critical
    quality_grade VARCHAR(10),
    
    -- Location and custody
    current_location_id UUID REFERENCES asset_locations(id),
    assigned_to_employee_id UUID REFERENCES users(id),
    responsible_department_id UUID REFERENCES organizations(id),
    cost_center_id UUID REFERENCES cost_centers(id),
    
    -- Financial information
    acquisition_cost DECIMAL(15,2) NOT NULL,
    acquisition_date DATE NOT NULL,
    acquisition_method VARCHAR(50) DEFAULT 'purchase', -- purchase, lease, donation, construction
    
    -- Depreciation settings
    depreciation_method VARCHAR(50) DEFAULT 'straight_line',
    useful_life_years INTEGER,
    useful_life_months INTEGER,
    salvage_value DECIMAL(15,2) DEFAULT 0,
    depreciation_start_date DATE,
    
    -- Current financial status
    accumulated_depreciation DECIMAL(15,2) DEFAULT 0,
    net_book_value DECIMAL(15,2),
    current_fair_value DECIMAL(15,2),
    last_revaluation_date DATE,
    
    -- Accounting integration
    asset_account_id UUID NOT NULL REFERENCES accounts(id),
    depreciation_account_id UUID NOT NULL REFERENCES accounts(id),
    accumulated_depreciation_account_id UUID NOT NULL REFERENCES accounts(id),
    
    -- Warranty and insurance
    warranty_start_date DATE,
    warranty_end_date DATE,
    warranty_provider VARCHAR(255),
    insurance_policy_number VARCHAR(100),
    insurance_value DECIMAL(15,2),
    insurance_expiry_date DATE,
    
    -- Asset status and lifecycle
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, under_maintenance, disposed, sold
    lifecycle_stage VARCHAR(50) DEFAULT 'operational', -- new, operational, maintenance, disposal
    
    -- Compliance and regulatory
    regulatory_requirements JSONB,
    compliance_certificates JSONB,
    environmental_impact_category VARCHAR(50),
    
    -- Digital assets
    primary_image_url VARCHAR(500),
    images JSONB, -- Array of image URLs
    documents JSONB, -- Array of document attachments
    qr_code TEXT,
    rfid_tag VARCHAR(100),
    
    -- Disposal information
    disposal_date DATE,
    disposal_method VARCHAR(50), -- sale, scrap, donation, trade_in
    disposal_value DECIMAL(15,2),
    disposal_reason TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_asset_type CHECK (asset_type IN (
        'building', 'machinery', 'vehicle', 'furniture', 'it_equipment', 
        'office_equipment', 'tools', 'fixtures', 'land', 'software', 'other'
    )),
    CONSTRAINT valid_condition CHECK (condition_rating IN ('excellent', 'good', 'fair', 'poor', 'critical')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'under_maintenance', 'disposed', 'sold')),
    CONSTRAINT valid_depreciation_method CHECK (depreciation_method IN (
        'straight_line', 'declining_balance', 'double_declining_balance', 
        'sum_of_years_digits', 'units_of_production', 'custom'
    )),
    CONSTRAINT positive_acquisition_cost CHECK (acquisition_cost > 0),
    CONSTRAINT valid_useful_life CHECK (useful_life_years > 0 OR useful_life_months > 0)
);

-- Asset categories for grouping and default settings
CREATE TABLE asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_category_id UUID REFERENCES asset_categories(id),
    
    category_code VARCHAR(50) NOT NULL,
    category_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Default settings for assets in this category
    default_depreciation_method VARCHAR(50) DEFAULT 'straight_line',
    default_useful_life_years INTEGER,
    default_salvage_value_percentage DECIMAL(5,2) DEFAULT 0,
    
    -- Default GL accounts
    default_asset_account_id UUID REFERENCES accounts(id),
    default_depreciation_account_id UUID REFERENCES accounts(id),
    default_accumulated_depreciation_account_id UUID REFERENCES accounts(id),
    
    -- Category properties
    requires_insurance BOOLEAN DEFAULT false,
    requires_maintenance_schedule BOOLEAN DEFAULT false,
    requires_location_tracking BOOLEAN DEFAULT true,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, category_code)
);

-- Asset locations for tracking physical location
CREATE TABLE asset_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_location_id UUID REFERENCES asset_locations(id),
    
    -- Location identification
    location_code VARCHAR(50) NOT NULL,
    location_name VARCHAR(255) NOT NULL,
    location_type VARCHAR(50) DEFAULT 'building', -- building, room, floor, department, warehouse, site
    
    -- Physical address and coordinates
    address JSONB,
    gps_coordinates JSONB, -- {latitude, longitude}
    
    -- Location properties
    floor_number INTEGER,
    room_number VARCHAR(20),
    area_square_meters DECIMAL(10,2),
    capacity_description TEXT,
    
    -- Access and security
    access_restrictions TEXT,
    security_level VARCHAR(20) DEFAULT 'standard', -- public, standard, restricted, confidential
    
    -- Environmental conditions
    climate_controlled BOOLEAN DEFAULT false,
    temperature_range JSONB, -- {min_temp, max_temp, unit}
    humidity_controlled BOOLEAN DEFAULT false,
    
    -- Responsible person
    location_manager_id UUID REFERENCES users(id),
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, location_code),
    CONSTRAINT valid_location_type CHECK (location_type IN (
        'building', 'room', 'floor', 'department', 'warehouse', 'site', 'vehicle', 'external'
    )),
    CONSTRAINT valid_security_level CHECK (security_level IN ('public', 'standard', 'restricted', 'confidential'))
);
```

### Asset Transfer and Location Tracking

```sql
-- Asset transfers for tracking movement history
CREATE TABLE asset_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transfer identification
    transfer_number VARCHAR(50) UNIQUE NOT NULL,
    asset_id UUID NOT NULL REFERENCES fixed_assets(id),
    
    -- Transfer details
    transfer_date DATE NOT NULL DEFAULT CURRENT_DATE,
    transfer_type VARCHAR(50) DEFAULT 'location_change', -- location_change, employee_assignment, department_transfer
    
    -- Source and destination
    from_location_id UUID REFERENCES asset_locations(id),
    to_location_id UUID REFERENCES asset_locations(id),
    from_employee_id UUID REFERENCES users(id),
    to_employee_id UUID REFERENCES users(id),
    from_department_id UUID REFERENCES organizations(id),
    to_department_id UUID REFERENCES organizations(id),
    
    -- Transfer workflow
    status VARCHAR(20) DEFAULT 'pending', -- pending, in_transit, completed, cancelled
    requested_by UUID NOT NULL REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Physical transfer details
    shipped_date DATE,
    received_date DATE,
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(100),
    
    -- Condition verification
    condition_before_transfer VARCHAR(20),
    condition_after_transfer VARCHAR(20),
    transfer_notes TEXT,
    damage_reported BOOLEAN DEFAULT false,
    damage_description TEXT,
    
    -- Financial impact
    transfer_cost DECIMAL(10,2) DEFAULT 0,
    insurance_required BOOLEAN DEFAULT false,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_transfer_type CHECK (transfer_type IN (
        'location_change', 'employee_assignment', 'department_transfer', 'maintenance_transfer'
    )),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'in_transit', 'completed', 'cancelled'))
);

-- Trigger to update asset location after transfer completion
CREATE OR REPLACE FUNCTION update_asset_location_on_transfer()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        UPDATE fixed_assets 
        SET 
            current_location_id = COALESCE(NEW.to_location_id, current_location_id),
            assigned_to_employee_id = COALESCE(NEW.to_employee_id, assigned_to_employee_id),
            responsible_department_id = COALESCE(NEW.to_department_id, responsible_department_id),
            updated_at = NOW()
        WHERE id = NEW.asset_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_asset_location
    AFTER UPDATE ON asset_transfers
    FOR EACH ROW
    EXECUTE FUNCTION update_asset_location_on_transfer();
```

##  Depreciation Management

### Depreciation Calculation Engine

```typescript
interface DepreciationParams {
  asset_id: string;
  depreciation_method: 'straight_line' | 'declining_balance' | 'double_declining_balance' | 'sum_of_years_digits' | 'units_of_production';
  acquisition_cost: number;
  salvage_value: number;
  useful_life_years?: number;
  useful_life_months?: number;
  depreciation_start_date: Date;
  declining_balance_rate?: number; // For declining balance methods
  current_period_usage?: number; // For units of production
  total_estimated_usage?: number; // For units of production
}

interface DepreciationSchedule {
  asset_id: string;
  periods: DepreciationPeriod[];
  total_depreciation: number;
  remaining_book_value: number;
}

interface DepreciationPeriod {
  period_number: number;
  period_start_date: Date;
  period_end_date: Date;
  opening_book_value: number;
  depreciation_amount: number;
  accumulated_depreciation: number;
  closing_book_value: number;
}

class DepreciationService {
  calculateDepreciationSchedule(params: DepreciationParams): DepreciationSchedule {
    const depreciableAmount = params.acquisition_cost - params.salvage_value;
    const totalPeriods = this.calculateTotalPeriods(params);
    
    switch (params.depreciation_method) {
      case 'straight_line':
        return this.calculateStraightLineDepreciation(params, depreciableAmount, totalPeriods);
      
      case 'declining_balance':
        return this.calculateDecliningBalanceDepreciation(params, depreciableAmount, totalPeriods);
      
      case 'double_declining_balance':
        return this.calculateDoubleDecliningBalanceDepreciation(params, depreciableAmount, totalPeriods);
      
      case 'sum_of_years_digits':
        return this.calculateSumOfYearsDigitsDepreciation(params, depreciableAmount, totalPeriods);
      
      case 'units_of_production':
        return this.calculateUnitsOfProductionDepreciation(params, depreciableAmount);
      
      default:
        throw new Error(`Unsupported depreciation method: ${params.depreciation_method}`);
    }
  }
  
  private calculateStraightLineDepreciation(
    params: DepreciationParams, 
    depreciableAmount: number, 
    totalPeriods: number
  ): DepreciationSchedule {
    const periodicDepreciation = depreciableAmount / totalPeriods;
    const periods: DepreciationPeriod[] = [];
    
    let accumulatedDepreciation = 0;
    let bookValue = params.acquisition_cost;
    
    for (let period = 1; period <= totalPeriods; period++) {
      const periodStartDate = this.addMonths(params.depreciation_start_date, period - 1);
      const periodEndDate = this.addMonths(periodStartDate, 1);
      
      const openingBookValue = bookValue;
      const depreciationAmount = Math.min(periodicDepreciation, bookValue - params.salvage_value);
      
      accumulatedDepreciation += depreciationAmount;
      bookValue -= depreciationAmount;
      
      periods.push({
        period_number: period,
        period_start_date: periodStartDate,
        period_end_date: periodEndDate,
        opening_book_value: openingBookValue,
        depreciation_amount: depreciationAmount,
        accumulated_depreciation: accumulatedDepreciation,
        closing_book_value: bookValue
      });
      
      if (bookValue <= params.salvage_value) {
        break;
      }
    }
    
    return {
      asset_id: params.asset_id,
      periods,
      total_depreciation: accumulatedDepreciation,
      remaining_book_value: bookValue
    };
  }
  
  private calculateDecliningBalanceDepreciation(
    params: DepreciationParams, 
    depreciableAmount: number, 
    totalPeriods: number
  ): DepreciationSchedule {
    const rate = params.declining_balance_rate || (1 / (params.useful_life_years || 1));
    const monthlyRate = rate / 12;
    const periods: DepreciationPeriod[] = [];
    
    let accumulatedDepreciation = 0;
    let bookValue = params.acquisition_cost;
    
    for (let period = 1; period <= totalPeriods; period++) {
      const periodStartDate = this.addMonths(params.depreciation_start_date, period - 1);
      const periodEndDate = this.addMonths(periodStartDate, 1);
      
      const openingBookValue = bookValue;
      const depreciationAmount = Math.min(
        bookValue * monthlyRate,
        bookValue - params.salvage_value
      );
      
      accumulatedDepreciation += depreciationAmount;
      bookValue -= depreciationAmount;
      
      periods.push({
        period_number: period,
        period_start_date: periodStartDate,
        period_end_date: periodEndDate,
        opening_book_value: openingBookValue,
        depreciation_amount: depreciationAmount,
        accumulated_depreciation: accumulatedDepreciation,
        closing_book_value: bookValue
      });
      
      if (bookValue <= params.salvage_value) {
        break;
      }
    }
    
    return {
      asset_id: params.asset_id,
      periods,
      total_depreciation: accumulatedDepreciation,
      remaining_book_value: bookValue
    };
  }
  
  private calculateSumOfYearsDigitsDepreciation(
    params: DepreciationParams, 
    depreciableAmount: number, 
    totalPeriods: number
  ): DepreciationSchedule {
    const totalPeriodYears = Math.ceil(totalPeriods / 12);
    const sumOfYears = (totalPeriodYears * (totalPeriodYears + 1)) / 2;
    const periods: DepreciationPeriod[] = [];
    
    let accumulatedDepreciation = 0;
    let bookValue = params.acquisition_cost;
    
    for (let period = 1; period <= totalPeriods; period++) {
      const currentYear = Math.ceil(period / 12);
      const remainingYears = totalPeriodYears - currentYear + 1;
      const yearlyFraction = remainingYears / sumOfYears;
      const monthlyDepreciation = (depreciableAmount * yearlyFraction) / 12;
      
      const periodStartDate = this.addMonths(params.depreciation_start_date, period - 1);
      const periodEndDate = this.addMonths(periodStartDate, 1);
      
      const openingBookValue = bookValue;
      const depreciationAmount = Math.min(monthlyDepreciation, bookValue - params.salvage_value);
      
      accumulatedDepreciation += depreciationAmount;
      bookValue -= depreciationAmount;
      
      periods.push({
        period_number: period,
        period_start_date: periodStartDate,
        period_end_date: periodEndDate,
        opening_book_value: openingBookValue,
        depreciation_amount: depreciationAmount,
        accumulated_depreciation: accumulatedDepreciation,
        closing_book_value: bookValue
      });
      
      if (bookValue <= params.salvage_value) {
        break;
      }
    }
    
    return {
      asset_id: params.asset_id,
      periods,
      total_depreciation: accumulatedDepreciation,
      remaining_book_value: bookValue
    };
  }
}
```

### Depreciation Transactions

```sql
-- Depreciation entries table for tracking all depreciation calculations
CREATE TABLE depreciation_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Entry identification
    entry_number VARCHAR(50) UNIQUE NOT NULL,
    asset_id UUID NOT NULL REFERENCES fixed_assets(id),
    
    -- Period information
    depreciation_period DATE NOT NULL, -- Usually month-end date
    fiscal_year INTEGER NOT NULL,
    fiscal_period INTEGER NOT NULL,
    
    -- Depreciation calculation
    opening_book_value DECIMAL(15,2) NOT NULL,
    depreciation_amount DECIMAL(15,2) NOT NULL,
    accumulated_depreciation DECIMAL(15,2) NOT NULL,
    closing_book_value DECIMAL(15,2) NOT NULL,
    
    -- Calculation details
    depreciation_method VARCHAR(50) NOT NULL,
    calculation_base DECIMAL(15,2), -- For units of production
    calculation_rate DECIMAL(8,6), -- Rate used for calculation
    
    -- Accounting integration
    journal_entry_id UUID REFERENCES journal_entries(id),
    posted_to_gl BOOLEAN DEFAULT false,
    posted_at TIMESTAMPTZ,
    
    -- Entry status
    status VARCHAR(20) DEFAULT 'calculated', -- calculated, approved, posted, reversed
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Reversal handling
    reversed_entry_id UUID REFERENCES depreciation_entries(id),
    reversal_reason TEXT,
    
    -- Metadata
    calculated_at TIMESTAMPTZ DEFAULT NOW(),
    calculated_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_status CHECK (status IN ('calculated', 'approved', 'posted', 'reversed')),
    CONSTRAINT positive_amounts CHECK (
        opening_book_value >= 0 AND 
        depreciation_amount >= 0 AND 
        accumulated_depreciation >= 0 AND 
        closing_book_value >= 0
    )
);

-- Automated depreciation calculation function
CREATE OR REPLACE FUNCTION calculate_monthly_depreciation(
    p_tenant_id UUID,
    p_period_date DATE DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    asset_id UUID,
    asset_number VARCHAR,
    depreciation_amount DECIMAL,
    status VARCHAR
) AS $$
DECLARE
    v_asset RECORD;
    v_last_depreciation depreciation_entries%ROWTYPE;
    v_depreciation_amount DECIMAL(15,2);
    v_accumulated_depreciation DECIMAL(15,2);
    v_new_book_value DECIMAL(15,2);
    v_entry_number VARCHAR(50);
BEGIN
    FOR v_asset IN 
        SELECT fa.*, ac.default_depreciation_method
        FROM fixed_assets fa
        LEFT JOIN asset_categories ac ON fa.asset_category_id = ac.id
        WHERE fa.tenant_id = p_tenant_id
          AND fa.status = 'active'
          AND fa.depreciation_start_date <= p_period_date
          AND (fa.disposal_date IS NULL OR fa.disposal_date > p_period_date)
          AND NOT EXISTS (
              SELECT 1 FROM depreciation_entries de 
              WHERE de.asset_id = fa.id 
                AND de.depreciation_period = p_period_date
                AND de.status != 'reversed'
          )
    LOOP
        -- Get last depreciation entry
        SELECT * INTO v_last_depreciation
        FROM depreciation_entries
        WHERE asset_id = v_asset.id
          AND status != 'reversed'
        ORDER BY depreciation_period DESC
        LIMIT 1;
        
        -- Calculate this period's depreciation
        IF v_asset.depreciation_method = 'straight_line' THEN
            v_depreciation_amount := (v_asset.acquisition_cost - v_asset.salvage_value) / 
                                   (COALESCE(v_asset.useful_life_months, v_asset.useful_life_years * 12));
        ELSIF v_asset.depreciation_method = 'declining_balance' THEN
            v_depreciation_amount := COALESCE(v_last_depreciation.closing_book_value, v_asset.acquisition_cost) * 
                                   (1.0 / COALESCE(v_asset.useful_life_years, 1)) / 12;
        ELSE
            -- Default to straight line if method not implemented
            v_depreciation_amount := (v_asset.acquisition_cost - v_asset.salvage_value) / 
                                   (COALESCE(v_asset.useful_life_months, v_asset.useful_life_years * 12));
        END IF;
        
        -- Ensure we don't depreciate below salvage value
        v_accumulated_depreciation := COALESCE(v_last_depreciation.accumulated_depreciation, 0) + v_depreciation_amount;
        IF v_accumulated_depreciation > (v_asset.acquisition_cost - v_asset.salvage_value) THEN
            v_depreciation_amount := (v_asset.acquisition_cost - v_asset.salvage_value) - COALESCE(v_last_depreciation.accumulated_depreciation, 0);
            v_accumulated_depreciation := v_asset.acquisition_cost - v_asset.salvage_value;
        END IF;
        
        v_new_book_value := v_asset.acquisition_cost - v_accumulated_depreciation;
        
        -- Generate entry number
        v_entry_number := 'DEP-' || TO_CHAR(p_period_date, 'YYYY-MM') || '-' || v_asset.asset_number;
        
        -- Create depreciation entry
        IF v_depreciation_amount > 0 THEN
            INSERT INTO depreciation_entries (
                tenant_id, entry_number, asset_id, depreciation_period, fiscal_year, fiscal_period,
                opening_book_value, depreciation_amount, accumulated_depreciation, closing_book_value,
                depreciation_method, status
            ) VALUES (
                p_tenant_id, v_entry_number, v_asset.id, p_period_date, 
                EXTRACT(YEAR FROM p_period_date), EXTRACT(MONTH FROM p_period_date),
                COALESCE(v_last_depreciation.closing_book_value, v_asset.acquisition_cost),
                v_depreciation_amount, v_accumulated_depreciation, v_new_book_value,
                v_asset.depreciation_method, 'calculated'
            );
            
            -- Update asset with current depreciation info
            UPDATE fixed_assets 
            SET 
                accumulated_depreciation = v_accumulated_depreciation,
                net_book_value = v_new_book_value,
                updated_at = NOW()
            WHERE id = v_asset.id;
            
            RETURN QUERY SELECT v_asset.id, v_asset.asset_number, v_depreciation_amount, 'calculated'::VARCHAR;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

##  Maintenance Management

### Maintenance Scheduling

```sql
-- Maintenance schedules for preventive maintenance
CREATE TABLE maintenance_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Schedule identification
    schedule_name VARCHAR(255) NOT NULL,
    asset_id UUID NOT NULL REFERENCES fixed_assets(id),
    
    -- Schedule parameters
    maintenance_type VARCHAR(50) NOT NULL, -- preventive, corrective, predictive, emergency
    frequency_type VARCHAR(20) NOT NULL, -- days, weeks, months, years, usage_hours, mileage
    frequency_value INTEGER NOT NULL,
    
    -- Schedule timing
    start_date DATE NOT NULL,
    end_date DATE,
    next_due_date DATE,
    
    -- Maintenance details
    estimated_duration_hours DECIMAL(6,2),
    estimated_cost DECIMAL(12,2),
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    
    -- Resource requirements
    required_skills JSONB, -- Array of required skills/certifications
    required_parts JSONB, -- Array of {item_id, quantity} for spare parts
    required_tools JSONB, -- Array of tool requirements
    
    -- Maintenance instructions
    maintenance_procedure TEXT,
    safety_requirements TEXT,
    quality_checkpoints JSONB,
    
    -- Assignment
    assigned_to_team VARCHAR(100),
    default_assignee_id UUID REFERENCES users(id),
    
    -- Schedule status
    is_active BOOLEAN DEFAULT true,
    auto_generate_work_orders BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_maintenance_type CHECK (maintenance_type IN ('preventive', 'corrective', 'predictive', 'emergency')),
    CONSTRAINT valid_frequency_type CHECK (frequency_type IN ('days', 'weeks', 'months', 'years', 'usage_hours', 'mileage')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical'))
);

-- Maintenance work orders
CREATE TABLE maintenance_work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Work order identification
    work_order_number VARCHAR(50) UNIQUE NOT NULL,
    asset_id UUID NOT NULL REFERENCES fixed_assets(id),
    maintenance_schedule_id UUID REFERENCES maintenance_schedules(id),
    
    -- Work order details
    work_order_type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium',
    description TEXT NOT NULL,
    
    -- Timing
    requested_date DATE NOT NULL DEFAULT CURRENT_DATE,
    scheduled_start_date TIMESTAMPTZ,
    scheduled_end_date TIMESTAMPTZ,
    actual_start_date TIMESTAMPTZ,
    actual_end_date TIMESTAMPTZ,
    
    -- Assignment and approval
    requested_by UUID NOT NULL REFERENCES users(id),
    assigned_to UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Work order status
    status VARCHAR(20) DEFAULT 'open', -- open, assigned, in_progress, completed, closed, cancelled
    
    -- Cost tracking
    estimated_cost DECIMAL(12,2),
    actual_labor_cost DECIMAL(12,2) DEFAULT 0,
    actual_parts_cost DECIMAL(12,2) DEFAULT 0,
    actual_other_cost DECIMAL(12,2) DEFAULT 0,
    total_actual_cost DECIMAL(12,2) DEFAULT 0,
    
    -- Downtime tracking
    planned_downtime_hours DECIMAL(8,2),
    actual_downtime_hours DECIMAL(8,2),
    
    -- Work performed
    work_performed TEXT,
    parts_used JSONB, -- Array of {item_id, quantity_used, cost}
    labor_hours DECIMAL(8,2),
    
    -- Quality and completion
    quality_check_passed BOOLEAN,
    quality_notes TEXT,
    completion_notes TEXT,
    
    -- Follow-up requirements
    requires_follow_up BOOLEAN DEFAULT false,
    follow_up_date DATE,
    follow_up_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_work_order_type CHECK (work_order_type IN ('preventive', 'corrective', 'predictive', 'emergency')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT valid_status CHECK (status IN ('open', 'assigned', 'in_progress', 'completed', 'closed', 'cancelled'))
);

-- Maintenance history for tracking all maintenance activities
CREATE TABLE maintenance_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    asset_id UUID NOT NULL REFERENCES fixed_assets(id),
    work_order_id UUID REFERENCES maintenance_work_orders(id),
    
    -- Maintenance event details
    maintenance_date DATE NOT NULL,
    maintenance_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    
    -- Personnel involved
    performed_by UUID REFERENCES users(id),
    supervised_by UUID REFERENCES users(id),
    
    -- Time and cost
    duration_hours DECIMAL(8,2),
    labor_cost DECIMAL(12,2),
    parts_cost DECIMAL(12,2),
    total_cost DECIMAL(12,2),
    
    -- Condition assessment
    condition_before VARCHAR(20),
    condition_after VARCHAR(20),
    
    -- Parts and materials used
    parts_used JSONB,
    
    -- Documentation
    documents JSONB, -- Array of document attachments
    photos JSONB, -- Array of photo URLs
    
    -- Next maintenance recommendation
    next_maintenance_date DATE,
    next_maintenance_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_maintenance_type CHECK (maintenance_type IN ('preventive', 'corrective', 'predictive', 'emergency'))
);
```

### Automated Work Order Generation

```typescript
interface MaintenanceScheduler {
  generateWorkOrders(tenantId: string, generateDate: Date): Promise<WorkOrder[]>;
  updateMaintenanceSchedules(tenantId: string): Promise<void>;
  calculateNextDueDate(schedule: MaintenanceSchedule, lastMaintenanceDate: Date): Date;
}

class MaintenanceService implements MaintenanceScheduler {
  async generateWorkOrders(tenantId: string, generateDate: Date = new Date()): Promise<WorkOrder[]> {
    const dueSchedules = await this.getDueMaintenanceSchedules(tenantId, generateDate);
    const workOrders: WorkOrder[] = [];
    
    for (const schedule of dueSchedules) {
      try {
        const workOrder = await this.createWorkOrderFromSchedule(schedule, generateDate);
        workOrders.push(workOrder);
        
        // Update the schedule's next due date
        await this.updateScheduleNextDueDate(schedule);
        
      } catch (error) {
        console.error(`Failed to create work order for schedule ${schedule.id}:`, error);
      }
    }
    
    return workOrders;
  }
  
  private async createWorkOrderFromSchedule(
    schedule: MaintenanceSchedule, 
    requestDate: Date
  ): Promise<WorkOrder> {
    const workOrderNumber = await this.generateWorkOrderNumber(schedule.tenant_id);
    
    const workOrder: WorkOrder = {
      id: uuid(),
      tenant_id: schedule.tenant_id,
      work_order_number: workOrderNumber,
      asset_id: schedule.asset_id,
      maintenance_schedule_id: schedule.id,
      work_order_type: schedule.maintenance_type,
      priority: schedule.priority,
      description: `Scheduled ${schedule.maintenance_type} maintenance: ${schedule.schedule_name}`,
      requested_date: requestDate,
      scheduled_start_date: this.calculateScheduledStartDate(schedule, requestDate),
      scheduled_end_date: this.calculateScheduledEndDate(schedule, requestDate),
      requested_by: schedule.created_by,
      assigned_to: schedule.default_assignee_id,
      status: 'open',
      estimated_cost: schedule.estimated_cost,
      planned_downtime_hours: schedule.estimated_duration_hours,
      created_at: new Date()
    };
    
    return await this.saveWorkOrder(workOrder);
  }
  
  calculateNextDueDate(schedule: MaintenanceSchedule, lastMaintenanceDate: Date): Date {
    const nextDue = new Date(lastMaintenanceDate);
    
    switch (schedule.frequency_type) {
      case 'days':
        nextDue.setDate(nextDue.getDate() + schedule.frequency_value);
        break;
      
      case 'weeks':
        nextDue.setDate(nextDue.getDate() + (schedule.frequency_value * 7));
        break;
      
      case 'months':
        nextDue.setMonth(nextDue.getMonth() + schedule.frequency_value);
        break;
      
      case 'years':
        nextDue.setFullYear(nextDue.getFullYear() + schedule.frequency_value);
        break;
      
      case 'usage_hours':
      case 'mileage':
        // These require usage tracking data
        return this.calculateUsageBasedDueDate(schedule, lastMaintenanceDate);
      
      default:
        throw new Error(`Unsupported frequency type: ${schedule.frequency_type}`);
    }
    
    return nextDue;
  }
  
  private async calculateUsageBasedDueDate(
    schedule: MaintenanceSchedule, 
    lastMaintenanceDate: Date
  ): Promise<Date> {
    const asset = await this.getAsset(schedule.asset_id);
    const currentUsage = await this.getCurrentAssetUsage(schedule.asset_id, schedule.frequency_type);
    const lastMaintenanceUsage = await this.getUsageAtDate(schedule.asset_id, lastMaintenanceDate, schedule.frequency_type);
    
    const usageSinceLastMaintenance = currentUsage - lastMaintenanceUsage;
    const usageRemaining = schedule.frequency_value - usageSinceLastMaintenance;
    
    if (usageRemaining <= 0) {
      // Already due
      return new Date();
    }
    
    // Estimate when the usage threshold will be reached
    const averageDailyUsage = await this.getAverageDailyUsage(schedule.asset_id, schedule.frequency_type, 30); // Last 30 days
    
    if (averageDailyUsage > 0) {
      const estimatedDaysToReachThreshold = usageRemaining / averageDailyUsage;
      const estimatedDueDate = new Date();
      estimatedDueDate.setDate(estimatedDueDate.getDate() + Math.ceil(estimatedDaysToReachThreshold));
      return estimatedDueDate;
    }
    
    // Fallback: assume usage will reach threshold in 30 days
    const fallbackDate = new Date();
    fallbackDate.setDate(fallbackDate.getDate() + 30);
    return fallbackDate;
  }
}
```

##  Asset Performance Analytics

### Key Performance Indicators

```typescript
interface AssetKPIs {
  asset_id: string;
  reporting_period: DateRange;
  
  availability_metrics: {
    total_hours: number;
    operational_hours: number;
    downtime_hours: number;
    availability_percentage: number;
  };
  
  reliability_metrics: {
    mtbf: number; // Mean Time Between Failures
    mttr: number; // Mean Time To Repair
    failure_count: number;
    repair_count: number;
  };
  
  cost_metrics: {
    acquisition_cost: number;
    accumulated_depreciation: number;
    current_book_value: number;
    maintenance_cost_ytd: number;
    maintenance_cost_per_hour: number;
    total_cost_of_ownership: number;
  };
  
  utilization_metrics: {
    planned_usage_hours: number;
    actual_usage_hours: number;
    utilization_percentage: number;
    efficiency_rating: number;
  };
  
  financial_metrics: {
    roi_percentage: number;
    payback_period_months: number;
    net_present_value: number;
    annual_savings: number;
  };
}

class AssetAnalyticsService {
  async calculateAssetKPIs(assetId: string, period: DateRange): Promise<AssetKPIs> {
    const asset = await this.getAsset(assetId);
    const maintenanceHistory = await this.getMaintenanceHistory(assetId, period);
    const usageData = await this.getUsageData(assetId, period);
    const costData = await this.getCostData(assetId, period);
    
    return {
      asset_id: assetId,
      reporting_period: period,
      availability_metrics: await this.calculateAvailabilityMetrics(assetId, period),
      reliability_metrics: await this.calculateReliabilityMetrics(maintenanceHistory),
      cost_metrics: await this.calculateCostMetrics(asset, costData),
      utilization_metrics: await this.calculateUtilizationMetrics(usageData),
      financial_metrics: await this.calculateFinancialMetrics(asset, costData)
    };
  }
  
  private async calculateAvailabilityMetrics(assetId: string, period: DateRange) {
    const totalHours = this.calculateHoursBetweenDates(period.start_date, period.end_date);
    const downtimeHours = await this.calculateDowntimeHours(assetId, period);
    const operationalHours = totalHours - downtimeHours;
    
    return {
      total_hours: totalHours,
      operational_hours: operationalHours,
      downtime_hours: downtimeHours,
      availability_percentage: (operationalHours / totalHours) * 100
    };
  }
  
  private async calculateReliabilityMetrics(maintenanceHistory: MaintenanceRecord[]) {
    const failures = maintenanceHistory.filter(record => 
      record.maintenance_type === 'corrective' || record.maintenance_type === 'emergency'
    );
    
    const repairs = maintenanceHistory.filter(record => 
      record.work_performed && record.status === 'completed'
    );
    
    // Calculate MTBF (Mean Time Between Failures)
    let totalOperatingHours = 0;
    let mtbf = 0;
    
    if (failures.length > 1) {
      const timeBetweenFailures = [];
      for (let i = 1; i < failures.length; i++) {
        const timeDiff = failures[i].maintenance_date.getTime() - failures[i - 1].maintenance_date.getTime();
        timeBetweenFailures.push(timeDiff / (1000 * 60 * 60)); // Convert to hours
      }
      mtbf = timeBetweenFailures.reduce((sum, time) => sum + time, 0) / timeBetweenFailures.length;
    }
    
    // Calculate MTTR (Mean Time To Repair)
    const repairTimes = repairs
      .filter(repair => repair.duration_hours)
      .map(repair => repair.duration_hours);
    
    const mttr = repairTimes.length > 0 
      ? repairTimes.reduce((sum, time) => sum + time, 0) / repairTimes.length 
      : 0;
    
    return {
      mtbf,
      mttr,
      failure_count: failures.length,
      repair_count: repairs.length
    };
  }
}
```

This  asset management system provides complete lifecycle tracking, automated depreciation, proactive maintenance scheduling, and detailed performance analytics to optimize asset utilization and minimize total cost of ownership.
