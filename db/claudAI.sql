
-- Audit logging (enhanced)
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    entity_id INT REFERENCES entities(id), -- Context entity
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50), -- What was changed
    resource_id BIGINT, -- ID of the changed resource
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    module VARCHAR(50), -- Which module generated the log
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Module-specific tables can be added here
-- For example, for accounting module:
CREATE TABLE chart_of_accounts (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    account_code VARCHAR(20) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(20) NOT NULL
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
    parent_account_id INT REFERENCES chart_of_accounts(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, entity_id, account_code)
);

-- Indexes for performance
CREATE INDEX idx_entities_tenant ON entities(tenant_id);
CREATE INDEX idx_entities_parent ON entities(parent_id);
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_entities_path ON entities USING GIST(path);
CREATE INDEX idx_hierarchy_paths_tenant ON hierarchy_paths(tenant_id);
CREATE INDEX idx_hierarchy_paths_ancestor ON hierarchy_paths(ancestor_id);
CREATE INDEX idx_hierarchy_paths_descendant ON hierarchy_paths(descendant_id);

CREATE INDEX idx_persons_tenant ON persons(tenant_id);
CREATE INDEX idx_persons_type ON persons(person_type);
CREATE INDEX idx_persons_email ON persons(email);
CREATE INDEX idx_persons_name ON persons(first_name, last_name);

CREATE INDEX idx_employees_tenant ON employees(tenant_id);
CREATE INDEX idx_employees_entity ON employees(entity_id);
CREATE INDEX idx_employees_number ON employees(employee_number);
CREATE INDEX idx_employees_status ON employees(employment_status);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_entity ON users(entity_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_type ON users(user_type);

CREATE INDEX idx_projects_tenant ON projects(tenant_id);
CREATE INDEX idx_projects_entity ON projects(entity_id);
CREATE INDEX idx_projects_status ON projects(status);

CREATE INDEX idx_budgets_tenant ON budgets(tenant_id);
CREATE INDEX idx_budgets_entity ON budgets(entity_id);
CREATE INDEX idx_budgets_year ON budgets(fiscal_year);

CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- Row Level Security (apply to all tables)
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE hierarchy_paths ENABLE ROW LEVEL SECURITY;
ALTER TABLE persons ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE chart_of_accounts ENABLE ROW LEVEL SECURITY;

-- Tenant context function (your existing function is good)
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS INT AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id')::INT;
EXCEPTION WHEN OTHERS THEN
    RAISE EXCEPTION 'Tenant context not set';
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- RLS Policies (apply tenant isolation to all tables)
CREATE POLICY tenant_isolation_policy ON tenants USING (id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON entities USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON hierarchy_paths USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON persons USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON employees USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON users USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON roles USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON projects USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON budgets USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON audit_logs USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON chart_of_accounts USING (tenant_id = current_tenant_id());

-- User roles policy (users can only see their own roles)
CREATE POLICY user_roles_policy ON user_roles 
    USING (user_id IN (SELECT id FROM users WHERE tenant_id = current_tenant_id()));

-- Triggers for maintaining hierarchy paths and timestamps
CREATE OR REPLACE FUNCTION update_timestamps() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply timestamp triggers to all relevant tables
CREATE TRIGGER update_tenant_timestamps BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_entity_timestamps BEFORE UPDATE ON entities FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_person_timestamps BEFORE UPDATE ON persons FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_employee_timestamps BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_user_timestamps BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_project_timestamps BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_budget_timestamps BEFORE UPDATE ON budgets FOR EACH ROW EXECUTE FUNCTION update_timestamps();

-- Enhanced hierarchy path maintenance
CREATE OR REPLACE FUNCTION maintain_hierarchy_paths() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Remove all paths involving this entity
        DELETE FROM hierarchy_paths 
        WHERE tenant_id = OLD.tenant_id 
        AND (ancestor_id = OLD.id OR descendant_id = OLD.id);
        RETURN OLD;
    END IF;
    
    -- Clear old paths for this descendant
    DELETE FROM hierarchy_paths 
    WHERE tenant_id = NEW.tenant_id AND descendant_id = NEW.id;
    
    -- Self-reference (depth 0)
    INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
    VALUES (NEW.tenant_id, NEW.id, NEW.id, 0);
    
    -- Update ltree path
    IF NEW.parent_id IS NULL THEN
        NEW.path = NEW.id::TEXT::LTREE;
    ELSE
        SELECT path || NEW.id::TEXT INTO NEW.path
        FROM entities 
        WHERE tenant_id = NEW.tenant_id AND id = NEW.parent_id;
    END IF;
    
    -- Parent paths (all ancestors)
    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
        SELECT NEW.tenant_id, p.ancestor_id, NEW.id, p.depth + 1
        FROM hierarchy_paths p
        WHERE p.tenant_id = NEW.tenant_id AND p.descendant_id = NEW.parent_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_maintain_hierarchy_paths
    BEFORE INSERT OR UPDATE OR DELETE ON entities
    FOR EACH ROW EXECUTE FUNCTION maintain_hierarchy_paths();

-- Function to get entity hierarchy (for easy querying)
CREATE OR REPLACE FUNCTION get_entity_descendants(entity_id INT, max_depth INT DEFAULT NULL)
RETURNS TABLE(id INT, name VARCHAR, type VARCHAR, depth INT) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type, hp.depth
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.descendant_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.ancestor_id = entity_id
    AND (max_depth IS NULL OR hp.depth <= max_depth)
    ORDER BY hp.depth, e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;



-- =====================================================
-- ACCOUNTING MODULE
-- =====================================================

-- Chart of Accounts (Enhanced)
CREATE TABLE chart_of_accounts (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    account_code VARCHAR(20) NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(20) NOT NULL
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
    account_subtype VARCHAR(30), -- Current Asset, Fixed Asset, etc.
    parent_account_id INT REFERENCES chart_of_accounts(id),
    normal_balance VARCHAR(6) CHECK (normal_balance IN ('DEBIT', 'CREDIT')),
    is_active BOOLEAN DEFAULT true,
    is_system_account BOOLEAN DEFAULT false, -- Prevent deletion of system accounts
    description TEXT,
    tax_code VARCHAR(20), -- For tax reporting
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, entity_id, account_code)
);

-- General Ledger
CREATE TABLE general_ledger (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    account_id INT NOT NULL REFERENCES chart_of_accounts(id),
    transaction_date DATE NOT NULL,
    reference_number VARCHAR(50),
    description TEXT NOT NULL,
    debit_amount DECIMAL(15,2) DEFAULT 0,
    credit_amount DECIMAL(15,2) DEFAULT 0,
    balance DECIMAL(15,2), -- Running balance
    journal_entry_id BIGINT, -- Reference to journal entry
    source_document_type VARCHAR(20), -- INVOICE, PAYMENT, JOURNAL, etc.
    source_document_id BIGINT,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT check_debit_or_credit CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR 
        (credit_amount > 0 AND debit_amount = 0)
    )
);

-- Journal Entries
CREATE TABLE journal_entries (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    entry_number VARCHAR(50) NOT NULL,
    entry_date DATE NOT NULL,
    description TEXT NOT NULL,
    reference VARCHAR(100),
    total_debit DECIMAL(15,2) NOT NULL,
    total_credit DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'POSTED', 'REVERSED')),
    posted_by INT REFERENCES users(id),
    posted_at TIMESTAMPTZ,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, entity_id, entry_number),
    CONSTRAINT check_balanced_entry CHECK (total_debit = total_credit)
);

-- Journal Entry Lines
CREATE TABLE journal_entry_lines (
    id BIGSERIAL PRIMARY KEY,
    journal_entry_id BIGINT NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    line_number INT NOT NULL,
    account_id INT NOT NULL REFERENCES chart_of_accounts(id),
    description TEXT,
    debit_amount DECIMAL(15,2) DEFAULT 0,
    credit_amount DECIMAL(15,2) DEFAULT 0,
    project_id INT REFERENCES projects(id), -- Project accounting
    cost_center_id INT REFERENCES entities(id), -- Cost center tracking
    PRIMARY KEY (journal_entry_id, line_number),
    CONSTRAINT check_line_debit_or_credit CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR 
        (credit_amount > 0 AND debit_amount = 0)
    )
);

-- =====================================================
-- INVENTORY MODULE
-- =====================================================

-- Item Categories
CREATE TABLE item_categories (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
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

-- =====================================================
-- CUSTOMER RELATIONSHIP MANAGEMENT (CRM)
-- =====================================================

-- Customer/Vendor master (extends persons)
CREATE TABLE business_partners (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    person_id INT REFERENCES persons(id) ON DELETE CASCADE, -- For individual customers
    partner_type VARCHAR(20) NOT NULL
        CHECK (partner_type IN ('CUSTOMER', 'VENDOR', 'BOTH')),
    company_name VARCHAR(255), -- For business customers
    business_registration VARCHAR(50),
    tax_registration VARCHAR(50),
    credit_limit DECIMAL(15,2) DEFAULT 0,
    payment_terms VARCHAR(50) DEFAULT 'NET_30',
    currency_code VARCHAR(3) DEFAULT 'USD',
    default_warehouse_id INT REFERENCES warehouses(id),
    sales_rep_id INT REFERENCES employees(id),
    account_manager_id INT REFERENCES employees(id),
    status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'INACTIVE', 'BLOCKED')),
    billing_address JSONB,
    shipping_address JSONB,
    contact_info JSONB, -- Phone, email, website, etc.
    preferences JSONB DEFAULT '{}'::jsonb,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sales Opportunities
CREATE TABLE opportunities (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    opportunity_number VARCHAR(50) NOT NULL,
    partner_id INT NOT NULL REFERENCES business_partners(id),
    contact_person_id INT REFERENCES persons(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    value DECIMAL(15,2),
    probability DECIMAL(5,2) CHECK (probability >= 0 AND probability <= 1),
    stage VARCHAR(20) DEFAULT 'PROSPECTING'
        CHECK (stage IN ('PROSPECTING', 'QUALIFYING', 'PROPOSAL', 'NEGOTIATION', 'CLOSED_WON', 'CLOSED_LOST')),
    source VARCHAR(50), -- Website, referral, cold call, etc.
    expected_close_date DATE,
    actual_close_date DATE,
    next_follow_up DATE,
    assigned_to INT REFERENCES employees(id),
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, opportunity_number)
);

-- =====================================================
-- PAYROLL MODULE
-- =====================================================

-- Pay Periods
CREATE TABLE pay_periods (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    period_name VARCHAR(50) NOT NULL,
    period_type VARCHAR(20) NOT NULL
        CHECK (period_type IN ('WEEKLY', 'BIWEEKLY', 'SEMIMONTHLY', 'MONTHLY')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    pay_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'PROCESSING', 'CLOSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, entity_id, period_name)
);

-- Employee Compensation
CREATE TABLE employee_compensation (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    compensation_type VARCHAR(20) NOT NULL
        CHECK (compensation_type IN ('SALARY', 'HOURLY', 'COMMISSION', 'BONUS')),
    amount DECIMAL(15,2) NOT NULL,
    currency_code VARCHAR(3) DEFAULT 'USD',
    frequency VARCHAR(20) -- ANNUAL, MONTHLY, HOURLY, etc.
        CHECK (frequency IN ('ANNUAL', 'MONTHLY', 'BIWEEKLY', 'WEEKLY', 'HOURLY')),
    effective_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Time Tracking
CREATE TABLE time_entries (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    project_id INT REFERENCES projects(id),
    entry_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    regular_hours DECIMAL(5,2) DEFAULT 0,
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    break_time DECIMAL(5,2) DEFAULT 0,
    entry_type VARCHAR(20) DEFAULT 'WORK'
        CHECK (entry_type IN ('WORK', 'VACATION', 'SICK', 'HOLIDAY', 'PERSONAL')),
    description TEXT,
    status VARCHAR(20) DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'SUBMITTED', 'APPROVED', 'REJECTED')),
    approved_by INT REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payroll Runs
CREATE TABLE payroll_runs (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    pay_period_id INT NOT NULL REFERENCES pay_periods(id),
    run_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'CALCULATED', 'APPROVED', 'PAID')),
    total_gross_pay DECIMAL(15,2) DEFAULT 0,
    total_deductions DECIMAL(15,2) DEFAULT 0,
    total_net_pay DECIMAL(15,2) DEFAULT 0,
    processed_by INT REFERENCES users(id),
    approved_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payroll Details
CREATE TABLE payroll_details (
    id BIGSERIAL PRIMARY KEY,
    payroll_run_id INT NOT NULL REFERENCES payroll_runs(id) ON DELETE CASCADE,
    employee_id INT NOT NULL REFERENCES employees(id),
    regular_hours DECIMAL(5,2) DEFAULT 0,
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    regular_pay DECIMAL(15,2) DEFAULT 0,
    overtime_pay DECIMAL(15,2) DEFAULT 0,
    bonus_pay DECIMAL(15,2) DEFAULT 0,
    commission_pay DECIMAL(15,2) DEFAULT 0,
    gross_pay DECIMAL(15,2) DEFAULT 0,
    tax_deductions DECIMAL(15,2) DEFAULT 0,
    other_deductions DECIMAL(15,2) DEFAULT 0,
    total_deductions DECIMAL(15,2) DEFAULT 0,
    net_pay DECIMAL(15,2) DEFAULT 0,
    pay_breakdown JSONB DEFAULT '{}'::jsonb, -- Detailed breakdown
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (payroll_run_id, employee_id)
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Accounting indexes
CREATE INDEX idx_gl_tenant_account ON general_ledger(tenant_id, account_id);
CREATE INDEX idx_gl_transaction_date ON general_ledger(transaction_date);
CREATE INDEX idx_gl_source_document ON general_ledger(source_document_type, source_document_id);
CREATE INDEX idx_journal_entries_date ON journal_entries(entry_date);
CREATE INDEX idx_journal_entries_status ON journal_entries(status);

-- Inventory indexes
CREATE INDEX idx_items_tenant_code ON items(tenant_id, item_code);
CREATE INDEX idx_items_category ON items(category_id);
CREATE INDEX idx_inventory_balances_item_warehouse ON inventory_balances(item_id, warehouse_id);
CREATE INDEX idx_inventory_movements_item ON inventory_movements(item_id);
CREATE INDEX idx_inventory_movements_date ON inventory_movements(transaction_date);
CREATE INDEX idx_inventory_movements_reference ON inventory_movements(reference_type, reference_id);

-- CRM indexes
CREATE INDEX idx_business_partners_type ON business_partners(partner_type);
CREATE INDEX idx_business_partners_sales_rep ON business_partners(sales_rep_id);
CREATE INDEX idx_opportunities_stage ON opportunities(stage);
CREATE INDEX idx_opportunities_assigned ON opportunities(assigned_to);
CREATE INDEX idx_opportunities_close_date ON opportunities(expected_close_date);

-- Payroll indexes
CREATE INDEX idx_time_entries_employee_date ON time_entries(employee_id, entry_date);
CREATE INDEX idx_time_entries_project ON time_entries(project_id);
CREATE INDEX idx_payroll_details_employee ON payroll_details(employee_id);
CREATE INDEX idx_employee_compensation_employee ON employee_compensation(employee_id);

-- =====================================================
-- ROW LEVEL SECURITY FOR MODULES
-- =====================================================

-- Apply RLS to all module tables
ALTER TABLE chart_of_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE general_ledger ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_entry_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE items ENABLE ROW LEVEL SECURITY;
ALTER TABLE item_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_balances ENABLE ROW LEVEL SECURITY;
ALTER TABLE inventory_movements ENABLE ROW LEVEL SECURITY;
ALTER TABLE business_partners ENABLE ROW LEVEL SECURITY;
ALTER TABLE opportunities ENABLE ROW LEVEL SECURITY;
ALTER TABLE pay_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_compensation ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE payroll_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE payroll_details ENABLE ROW LEVEL SECURITY;

-- Create tenant isolation policies for all module tables
CREATE POLICY tenant_isolation_policy ON chart_of_accounts USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON general_ledger USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON journal_entries USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON items USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON item_categories USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON warehouses USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON inventory_balances USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON inventory_movements USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON business_partners USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON opportunities USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON pay_periods USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON employee_compensation USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON time_entries USING (tenant_id = current_tenant_id());
CREATE POLICY tenant_isolation_policy ON payroll_runs USING (tenant_id = current_tenant_id());

-- Journal entry lines inherit from parent journal entry
CREATE POLICY tenant_isolation_policy ON journal_entry_lines 
USING (journal_entry_id IN (
    SELECT id FROM journal_entries WHERE tenant_id = current_tenant_id()
));

-- Payroll details inherit from payroll run
CREATE POLICY tenant_isolation_policy ON payroll_details 
USING (payroll_run_id IN (
    SELECT id FROM payroll_runs WHERE tenant_id = current_tenant_id()
));

-- =====================================================
-- TRIGGERS FOR MODULE TABLES
-- =====================================================

-- Add timestamp triggers
CREATE TRIGGER update_chart_of_accounts_timestamps BEFORE UPDATE ON chart_of_accounts FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_items_timestamps BEFORE UPDATE ON items FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_inventory_balances_timestamps BEFORE UPDATE ON inventory_balances FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_business_partners_timestamps BEFORE UPDATE ON business_partners FOR EACH ROW EXECUTE FUNCTION update_timestamps();
CREATE TRIGGER update_opportunities_timestamps BEFORE UPDATE ON opportunities FOR EACH ROW EXECUTE FUNCTION update_timestamps();

-- =====================================================
-- RETAIL INDUSTRY MODULE
-- =====================================================

-- Point of Sale Terminals
CREATE TABLE pos_terminals (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    warehouse_id INT NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    terminal_code VARCHAR(20) NOT NULL,
    terminal_name VARCHAR(100) NOT NULL,
    ip_address INET,
    status VARCHAR(20) DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'INACTIVE', 'MAINTENANCE')),
    cashier_id INT REFERENCES employees(id),
    opened_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    opening_cash DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, terminal_code)
);

-- Sales Transactions
CREATE TABLE sales_transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    transaction_number VARCHAR(50) NOT NULL,
    pos_terminal_id INT REFERENCES pos_terminals(id),
    customer_id INT REFERENCES business_partners(id),
    cashier_id INT REFERENCES employees(id),
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    subtotal DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(20) DEFAULT 'CASH'
        CHECK (payment_method IN ('CASH', 'CARD', 'CHECK', 'STORE_CREDIT', 'MOBILE')),
    status VARCHAR(20) DEFAULT 'COMPLETED'
        CHECK (status IN ('PENDING', 'COMPLETED',
