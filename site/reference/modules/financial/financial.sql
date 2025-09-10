-- =====================================================================
-- MIGRATION 001: CREATE FINANCIAL ENUMS AND TYPES
-- =====================================================================
-- File: 001_create_finance_enums.up.sql
-- Description: Create all enums and custom types for the financial module
-- Dependencies: None
-- =====================================================================

-- Account classification types
CREATE TYPE account_type_enum AS ENUM (
    'asset',      -- Assets (what the company owns)
    'liability',  -- Liabilities (what the company owes)  
    'equity',     -- Equity (owner's interest)
    'income',     -- Income/Revenue (money coming in)
    'expense'     -- Expenses (money going out)
);
COMMENT ON TYPE account_type_enum IS 'Primary account classification following standard accounting principles';

-- Detailed account subtypes for more granular classification
CREATE TYPE account_subtype_enum AS ENUM (
    -- Asset subtypes
    'current_asset',        -- Cash, receivables, inventory (< 1 year)
    'fixed_asset',          -- Property, plant, equipment (> 1 year)
    'other_asset',          -- Intangible assets, investments
    'inventory',            -- Goods for sale
    'accounts_receivable',  -- Money owed by customers
    'cash',                 -- Cash and cash equivalents
    'bank',                 -- Bank accounts
    'investment',           -- Short/long term investments
    
    -- Liability subtypes  
    'current_liability',    -- Payable within 1 year
    'long_term_liability',  -- Payable after 1 year
    'accounts_payable',     -- Money owed to suppliers
    'accrued_liability',    -- Accumulated but unpaid expenses
    'tax_liability',        -- Tax obligations
    
    -- Equity subtypes
    'owner_equity',         -- Owner's capital investment
    'retained_earnings',    -- Accumulated profits
    'capital',              -- Share capital/stock
    
    -- Income subtypes
    'operating_income',     -- Primary business revenue
    'other_income',         -- Non-operating income
    'interest_income',      -- Income from investments
    
    -- Expense subtypes
    'operating_expense',    -- Regular business expenses
    'cost_of_goods_sold',   -- Direct costs of products sold
    'administrative_expense', -- Admin and overhead costs
    'interest_expense',     -- Cost of borrowing
    'tax_expense'           -- Income tax expense
);
COMMENT ON TYPE account_subtype_enum IS 'Detailed account subtypes for precise financial classification and reporting';

-- Balance type indicates normal balance side for an account
CREATE TYPE balance_type_enum AS ENUM ('debit', 'credit');
COMMENT ON TYPE balance_type_enum IS 'Normal balance side: Assets/Expenses=debit, Liabilities/Equity/Income=credit';

-- Transaction types supported by the system
CREATE TYPE transaction_type_enum AS ENUM (
    'journal_entry',    -- Manual accounting entries
    'sales_invoice',    -- Bills to customers
    'purchase_invoice', -- Bills from suppliers
    'payment',          -- Outgoing payments
    'receipt',          -- Incoming payments
    'credit_note',      -- Sales returns/adjustments
    'debit_note',       -- Purchase returns/adjustments
    'opening_balance',  -- Beginning balances
    'closing_entry',    -- Year-end closing entries
    'adjustment',       -- Correcting entries
    'depreciation',     -- Asset depreciation
    'tax_entry'         -- Tax-related entries
);
COMMENT ON TYPE transaction_type_enum IS 'Types of financial transactions supported by the system';

-- Transaction workflow states
CREATE TYPE transaction_status_enum AS ENUM (
    'draft',            -- Being prepared, not yet submitted
    'pending_approval', -- Submitted for approval
    'approved',         -- Approved but not posted
    'posted',           -- Posted to general ledger
    'cancelled',        -- Cancelled before posting
    'reversed'          -- Posted but later reversed
);
COMMENT ON TYPE transaction_status_enum IS 'Transaction workflow states from creation to posting';

-- Types of business parties
CREATE TYPE party_type_enum AS ENUM (
    'customer',  -- Entities that buy from us
    'supplier',  -- Entities we buy from
    'employee',  -- Company employees
    'other'      -- Other parties (banks, government, etc.)
);
COMMENT ON TYPE party_type_enum IS 'Types of business parties for financial transactions';

-- Supported currencies (expandable list)
CREATE TYPE currency_code AS ENUM (
    'USD', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY', 'INR', 'BRL',
    'MXN', 'ZAR', 'SGD', 'HKD', 'NOK', 'SEK', 'DKK', 'PLN', 'CZK', 'HUF'
);
COMMENT ON TYPE currency_code IS 'Supported currency codes following ISO 4217 standard';

-- =====================================================================
-- MIGRATION 002: CHART OF ACCOUNTS
-- =====================================================================
-- File: 002_create_chart_of_accounts.up.sql
-- Description: Create chart of accounts structure for organizing financial accounts
-- Dependencies: finance enums, tenants, entities, users tables
-- =====================================================================

CREATE TABLE finance_accounts (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Chart identification and description
    code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Financial configuration
    base_currency currency_code NOT NULL DEFAULT 'USD',
    fiscal_year_start_month INTEGER NOT NULL DEFAULT 1 
        CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    
    -- Status and defaults
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    -- Business constraints
    CONSTRAINT uk_finance_coa_tenant_code UNIQUE(tenant_id, code),
    CONSTRAINT uk_finance_coa_tenant_name UNIQUE(tenant_id, name),
    CONSTRAINT ck_finance_coa_fiscal_month CHECK (fiscal_year_start_month BETWEEN 1 AND 12)
);

-- Indexes for performance
CREATE INDEX idx_finance_accounts_tenant 
    ON finance_accounts(tenant_id);
CREATE INDEX idx_finance_accounts_entity 
    ON finance_accounts(entity_id);
CREATE INDEX idx_finance_accounts_active 
    ON finance_accounts(tenant_id, is_active) 
    WHERE is_active = true;

-- Row Level Security
ALTER TABLE finance_accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_accounts_tenant_isolation 
    ON finance_accounts
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_accounts_admin_access 
    ON finance_accounts
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_accounts IS 
'Chart of accounts templates for organizing financial accounts by entity. Each entity can have its own chart of accounts structure.';

COMMENT ON COLUMN finance_accounts.id IS 
'Unique identifier for the chart of accounts';
COMMENT ON COLUMN finance_accounts.tenant_id IS 
'Reference to tenant - supports multi-tenancy isolation';
COMMENT ON COLUMN finance_accounts.entity_id IS 
'Reference to entity (subsidiary/department/region) - enables entity-specific charts';
COMMENT ON COLUMN finance_accounts.code IS 
'Short code for chart identification (e.g., "US-GAAP", "IFRS")';
COMMENT ON COLUMN finance_accounts.name IS 
'Descriptive name for the chart of accounts';
COMMENT ON COLUMN finance_accounts.base_currency IS 
'Primary currency for this chart of accounts';
COMMENT ON COLUMN finance_accounts.fiscal_year_start_month IS 
'Month when fiscal year starts (1=January, 4=April, etc.)';
COMMENT ON COLUMN finance_accounts.is_default IS 
'Indicates if this is the default chart for new accounts';

-- =====================================================================
-- MIGRATION 003: ACCOUNTS
-- =====================================================================
-- File: 003_create_accounts.up.sql
-- Description: Create accounts table with nested set hierarchy
-- Dependencies: chart_of_accounts, finance enums
-- =====================================================================

CREATE TABLE finance_accounts (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    chart_id UUID NOT NULL REFERENCES finance_accounts(id) ON DELETE CASCADE,
    
    -- Account identification
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Account classification
    account_type account_type_enum NOT NULL,
    account_subtype account_subtype_enum NOT NULL,
    normal_balance balance_type_enum NOT NULL,
    
    -- Nested set model for hierarchy (efficient tree operations)
    parent_account_id UUID REFERENCES finance_accounts(id) ON DELETE RESTRICT,
    lft INTEGER NOT NULL,
    rgt INTEGER NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0,
    is_group BOOLEAN NOT NULL DEFAULT false,
    
    -- Financial configuration
    currency currency_code NOT NULL,
    allow_manual_entries BOOLEAN NOT NULL DEFAULT true,
    require_cost_center BOOLEAN NOT NULL DEFAULT false,
    require_project BOOLEAN NOT NULL DEFAULT false,
    
    -- Balance information
    opening_balance DECIMAL(19,4) DEFAULT 0.00,
    current_balance DECIMAL(19,4) DEFAULT 0.00,
    credit_limit DECIMAL(19,4),
    
    -- Account control
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_frozen BOOLEAN NOT NULL DEFAULT false,
    freeze_date DATE,
    freeze_reason TEXT,
    
    -- Tax configuration
    default_tax_rate DECIMAL(5,4),
    tax_account_id UUID REFERENCES finance_accounts(id),
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    -- Business constraints
    CONSTRAINT uk_finance_accounts_tenant_chart_code UNIQUE(tenant_id, chart_id, code),
    CONSTRAINT uk_finance_accounts_tenant_chart_name UNIQUE(tenant_id, chart_id, name),
    CONSTRAINT ck_finance_accounts_nested_set CHECK (lft < rgt),
    CONSTRAINT ck_finance_accounts_no_self_parent CHECK (parent_account_id IS NULL OR parent_account_id != id),
    CONSTRAINT ck_finance_accounts_group_balance CHECK (opening_balance IS NULL OR is_group = false),
    CONSTRAINT ck_finance_accounts_credit_limit CHECK (credit_limit IS NULL OR credit_limit >= 0),
    CONSTRAINT ck_finance_accounts_tax_rate CHECK (default_tax_rate IS NULL OR (default_tax_rate >= 0 AND default_tax_rate <= 100))
);

-- Indexes for performance
CREATE INDEX idx_finance_accounts_tenant ON finance_accounts(tenant_id);
CREATE INDEX idx_finance_accounts_entity ON finance_accounts(entity_id);
CREATE INDEX idx_finance_accounts_chart ON finance_accounts(tenant_id, chart_id);
CREATE INDEX idx_finance_accounts_code ON finance_accounts(tenant_id, chart_id, code);
CREATE INDEX idx_finance_accounts_type ON finance_accounts(tenant_id, account_type);
CREATE INDEX idx_finance_accounts_parent ON finance_accounts(tenant_id, parent_account_id);
CREATE INDEX idx_finance_accounts_tree ON finance_accounts(tenant_id, lft, rgt);
CREATE INDEX idx_finance_accounts_active ON finance_accounts(tenant_id, is_active) WHERE is_active = true;

-- Row Level Security
ALTER TABLE finance_accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_accounts_tenant_isolation 
    ON finance_accounts
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_accounts_admin_access 
    ON finance_accounts
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_accounts IS 
'Individual accounts within chart of accounts using nested set model for efficient hierarchy operations. Supports multi-entity accounting with complete audit trail.';

COMMENT ON COLUMN finance_accounts.id IS 'Unique identifier for the account';
COMMENT ON COLUMN finance_accounts.tenant_id IS 'Reference to tenant for multi-tenancy isolation';
COMMENT ON COLUMN finance_accounts.entity_id IS 'Reference to entity (subsidiary/department/region)';
COMMENT ON COLUMN finance_accounts.chart_id IS 'Reference to chart of accounts this account belongs to';
COMMENT ON COLUMN finance_accounts.code IS 'Account code (e.g., "1000", "CASH-001") - unique within chart';
COMMENT ON COLUMN finance_accounts.name IS 'Account name (e.g., "Cash", "Accounts Receivable")';
COMMENT ON COLUMN finance_accounts.account_type IS 'Primary classification: asset, liability, equity, income, expense';
COMMENT ON COLUMN finance_accounts.account_subtype IS 'Detailed classification for reporting granularity';
COMMENT ON COLUMN finance_accounts.normal_balance IS 'Normal balance side: debit for assets/expenses, credit for liabilities/equity/income';
COMMENT ON COLUMN finance_accounts.lft IS 'Left boundary for nested set model - enables efficient tree queries';
COMMENT ON COLUMN finance_accounts.rgt IS 'Right boundary for nested set model - enables efficient tree queries';
COMMENT ON COLUMN finance_accounts.depth IS 'Depth level in account hierarchy (0=root level)';
COMMENT ON COLUMN finance_accounts.is_group IS 'True if account is a group (parent) that contains other accounts';
COMMENT ON COLUMN finance_accounts.currency IS 'Account currency - can differ from chart base currency';
COMMENT ON COLUMN finance_accounts.allow_manual_entries IS 'Whether manual journal entries are allowed for this account';
COMMENT ON COLUMN finance_accounts.require_cost_center IS 'Whether cost center is mandatory for entries to this account';
COMMENT ON COLUMN finance_accounts.require_project IS 'Whether project is mandatory for entries to this account';
COMMENT ON COLUMN finance_accounts.opening_balance IS 'Opening balance for the account (null for group accounts)';
COMMENT ON COLUMN finance_accounts.current_balance IS 'Current calculated balance (updated by triggers)';
COMMENT ON COLUMN finance_accounts.is_frozen IS 'Whether account is frozen to prevent further entries';
COMMENT ON COLUMN finance_accounts.freeze_date IS 'Date when account was frozen';
COMMENT ON COLUMN finance_accounts.default_tax_rate IS 'Default tax rate percentage for this account';

-- =====================================================================
-- MIGRATION 004: COST CENTERS
-- =====================================================================
-- File: 004_create_cost_centers.up.sql
-- Description: Create cost centers for departmental accounting
-- Dependencies: entities, users tables
-- =====================================================================

CREATE TABLE finance_cost_centers (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Cost center identification
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Hierarchy (self-referencing for cost center groups)
    parent_cost_center_id UUID REFERENCES finance_cost_centers(id) ON DELETE RESTRICT,
    is_group BOOLEAN NOT NULL DEFAULT false,
    
    -- Management and control
    is_active BOOLEAN NOT NULL DEFAULT true,
    manager_user_id UUID REFERENCES users(id),
    
    -- Budget information
    annual_budget DECIMAL(19,4),
    budget_currency currency_code,
    budget_start_date DATE,
    budget_end_date DATE,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    -- Business constraints
    CONSTRAINT uk_finance_cost_centers_tenant_code UNIQUE(tenant_id, code),
    CONSTRAINT uk_finance_cost_centers_tenant_name UNIQUE(tenant_id, name),
    CONSTRAINT ck_finance_cost_centers_no_self_parent CHECK (parent_cost_center_id IS NULL OR parent_cost_center_id != id),
    CONSTRAINT ck_finance_cost_centers_budget_dates CHECK (budget_start_date IS NULL OR budget_end_date IS NULL OR budget_start_date <= budget_end_date),
    CONSTRAINT ck_finance_cost_centers_annual_budget CHECK (annual_budget IS NULL OR annual_budget >= 0)
);

-- Indexes for performance
CREATE INDEX idx_finance_cost_centers_tenant ON finance_cost_centers(tenant_id);
CREATE INDEX idx_finance_cost_centers_entity ON finance_cost_centers(entity_id);
CREATE INDEX idx_finance_cost_centers_code ON finance_cost_centers(tenant_id, code);
CREATE INDEX idx_finance_cost_centers_parent ON finance_cost_centers(parent_cost_center_id);
CREATE INDEX idx_finance_cost_centers_active ON finance_cost_centers(tenant_id, is_active) WHERE is_active = true;
CREATE INDEX idx_finance_cost_centers_manager ON finance_cost_centers(manager_user_id);

-- Row Level Security
ALTER TABLE finance_cost_centers ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_cost_centers_tenant_isolation 
    ON finance_cost_centers
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_cost_centers_admin_access 
    ON finance_cost_centers
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_cost_centers IS 
'Cost centers for departmental accounting, budget tracking, and expense allocation. Supports hierarchical organization and budget management.';

COMMENT ON COLUMN finance_cost_centers.id IS 'Unique identifier for the cost center';
COMMENT ON COLUMN finance_cost_centers.tenant_id IS 'Reference to tenant for multi-tenancy isolation';
COMMENT ON COLUMN finance_cost_centers.entity_id IS 'Reference to entity (subsidiary/department/region)';
COMMENT ON COLUMN finance_cost_centers.code IS 'Cost center code (e.g., "SALES", "IT", "HR") - unique within tenant';
COMMENT ON COLUMN finance_cost_centers.name IS 'Cost center name (e.g., "Sales Department", "IT Operations")';
COMMENT ON COLUMN finance_cost_centers.parent_cost_center_id IS 'Reference to parent cost center for hierarchical organization';
COMMENT ON COLUMN finance_cost_centers.is_group IS 'True if this is a group cost center containing other cost centers';
COMMENT ON COLUMN finance_cost_centers.manager_user_id IS 'User responsible for managing this cost center';
COMMENT ON COLUMN finance_cost_centers.annual_budget IS 'Annual budget amount for this cost center';
COMMENT ON COLUMN finance_cost_centers.budget_currency IS 'Currency for budget amounts';
COMMENT ON COLUMN finance_cost_centers.budget_start_date IS 'Budget period start date';
COMMENT ON COLUMN finance_cost_centers.budget_end_date IS 'Budget period end date';

-- =====================================================================
-- MIGRATION 005: PROJECTS
-- =====================================================================
-- File: 005_create_projects.up.sql
-- Description: Create projects for project-based accounting
-- Dependencies: entities, users tables
-- =====================================================================

CREATE TABLE finance_projects (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Project identification
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Project lifecycle
    start_date DATE,
    end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' 
        CHECK (status IN ('planning', 'active', 'on_hold', 'completed', 'cancelled')),
    completion_percentage DECIMAL(5,2) DEFAULT 0.00 
        CHECK (completion_percentage >= 0 AND completion_percentage <= 100),
    
    -- Budget and financial information
    total_budget DECIMAL(19,4),
    budget_currency currency_code,
    actual_cost DECIMAL(19,4) DEFAULT 0.00,
    project_manager_id UUID REFERENCES users(id),
    
    -- Configuration
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_billable BOOLEAN NOT NULL DEFAULT false,
    billing_rate DECIMAL(10,2),
    
    -- Client information (if external project)
    client_id UUID, -- Could reference finance_parties when that table is created
    contract_value DECIMAL(19,4),
    contract_currency currency_code,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    -- Business constraints
    CONSTRAINT uk_finance_projects_tenant_code UNIQUE(tenant_id, code),
    CONSTRAINT uk_finance_projects_tenant_name UNIQUE(tenant_id, name),
    CONSTRAINT ck_finance_projects_dates CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date),
    CONSTRAINT ck_finance_projects_budget CHECK (total_budget IS NULL OR total_budget >= 0),
    CONSTRAINT ck_finance_projects_actual_cost CHECK (actual_cost >= 0),
    CONSTRAINT ck_finance_projects_billing_rate CHECK (billing_rate IS NULL OR billing_rate >= 0),
    CONSTRAINT ck_finance_projects_contract_value CHECK (contract_value IS NULL OR contract_value >= 0)
);

-- Indexes for performance
CREATE INDEX idx_finance_projects_tenant ON finance_projects(tenant_id);
CREATE INDEX idx_finance_projects_entity ON finance_projects(entity_id);
CREATE INDEX idx_finance_projects_code ON finance_projects(tenant_id, code);
CREATE INDEX idx_finance_projects_status ON finance_projects(tenant_id, status);
CREATE INDEX idx_finance_projects_dates ON finance_projects(tenant_id, start_date, end_date);
CREATE INDEX idx_finance_projects_manager ON finance_projects(project_manager_id);
CREATE INDEX idx_finance_projects_active ON finance_projects(tenant_id, is_active) WHERE is_active = true;

-- Row Level Security
ALTER TABLE finance_projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_projects_tenant_isolation 
    ON finance_projects
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_projects_admin_access 
    ON finance_projects
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_projects IS 
'Projects for project-based accounting, cost tracking, and revenue recognition. Supports both internal projects and billable client projects.';

COMMENT ON COLUMN finance_projects.id IS 'Unique identifier for the project';
COMMENT ON COLUMN finance_projects.tenant_id IS 'Reference to tenant for multi-tenancy isolation';
COMMENT ON COLUMN finance_projects.entity_id IS 'Reference to entity (subsidiary/department/region)';
COMMENT ON COLUMN finance_projects.code IS 'Project code (e.g., "PROJ-001", "WEB-2024") - unique within tenant';
COMMENT ON COLUMN finance_projects.name IS 'Project name (e.g., "Website Redesign", "ERP Implementation")';
COMMENT ON COLUMN finance_projects.start_date IS 'Project start date';
COMMENT ON COLUMN finance_projects.end_date IS 'Project end date (planned or actual)';
COMMENT ON COLUMN finance_projects.status IS 'Current project status: planning, active, on_hold, completed, cancelled';
COMMENT ON COLUMN finance_projects.completion_percentage IS 'Project completion percentage (0-100)';
COMMENT ON COLUMN finance_projects.total_budget IS 'Total project budget amount';
COMMENT ON COLUMN finance_projects.actual_cost IS 'Actual costs incurred to date';
COMMENT ON COLUMN finance_projects.project_manager_id IS 'User responsible for managing this project';
COMMENT ON COLUMN finance_projects.is_billable IS 'Whether this project is billable to external client';
COMMENT ON COLUMN finance_projects.billing_rate IS 'Hourly billing rate for billable projects';
COMMENT ON COLUMN finance_projects.contract_value IS 'Total contract value for client projects';

-- =====================================================================
-- MIGRATION 006: PARTIES
-- =====================================================================
-- File: 006_create_parties.up.sql
-- Description: Create unified parties table for customers, suppliers, etc.
-- Dependencies: entities, finance_accounts (for receivable/payable accounts)
-- =====================================================================

CREATE TABLE finance_parties (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Party classification and identification
    party_type party_type_enum NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    legal_name VARCHAR(255),
    
    -- Contact information
    email VARCHAR(254),
    phone VARCHAR(30),
    mobile VARCHAR(30),
    fax VARCHAR(30),
    website VARCHAR(200),
    
    -- Address information (JSONB for flexibility)
    billing_address JSONB DEFAULT '{}'::jsonb,
    shipping_address JSONB DEFAULT '{}'::jsonb,
    
    -- Financial configuration
    currency currency_code NOT NULL DEFAULT 'USD',
    payment_terms_days INTEGER DEFAULT 30 CHECK (payment_terms_days >= 0),
    credit_limit DECIMAL(19,4) DEFAULT 0.00,
    credit_rating VARCHAR(10),
    
    -- Tax information
    tax_id VARCHAR(50),
    tax_exempt BOOLEAN NOT NULL DEFAULT false,
    default_tax_rate DECIMAL(5,4),
    tax_registration_number VARCHAR(50),
    
    -- Banking information
    bank_name VARCHAR(255),
    bank_account_number VARCHAR(50),
    bank_routing_number VARCHAR(30),
    bank_swift_code VARCHAR(20),
    iban VARCHAR(50),
    
    -- Account mappings (will be populated after accounts are created)
    receivable_account_id UUID REFERENCES finance_accounts(id),
    payable_account_id UUID REFERENCES finance_accounts(id),
    default_income_account_id UUID REFERENCES finance_accounts(id),
    default_expense_account_id UUID REFERENCES finance_accounts(id),
    
    -- Status and configuration
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_internal BOOLEAN NOT NULL DEFAULT false,
    
    -- Additional information
    industry VARCHAR(100),
    notes TEXT,
    additional_info JSONB DEFAULT '{}'::jsonb,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    -- Business constraints
    CONSTRAINT uk_finance_parties_tenant_type_code UNIQUE(tenant_id, party_type, code),
    CONSTRAINT uk_finance_parties_tenant_type_name UNIQUE(tenant_id, party_type, name),
    CONSTRAINT ck_finance_parties_credit_limit CHECK (credit_limit >= 0),
    CONSTRAINT ck_finance_parties_payment_terms CHECK (payment_terms_days >= 0),
    CONSTRAINT ck_finance_parties_tax_rate CHECK (default_tax_rate IS NULL OR (default_tax_rate >= 0 AND default_tax_rate <= 100)),
    CONSTRAINT ck_finance_parties_email CHECK (email IS NULL OR email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Indexes for performance
CREATE INDEX idx_finance_parties_tenant ON finance_parties(tenant_id);
CREATE INDEX idx_finance_parties_entity ON finance_parties(entity_id);
CREATE INDEX idx_finance_parties_type ON finance_parties(tenant_id, party_type);
CREATE INDEX idx_finance_parties_code ON finance_parties(tenant_id, party_type, code);
CREATE INDEX idx_finance_parties_name ON finance_parties(tenant_id, name);
CREATE INDEX idx_finance_parties_email ON finance_parties(email) WHERE email IS NOT NULL;
CREATE INDEX idx_finance_parties_active ON finance_parties(tenant_id, is_active) WHERE is_active = true;

-- Row Level Security
ALTER TABLE finance_parties ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_parties_tenant_isolation 
    ON finance_parties
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_parties_admin_access 
    ON finance_parties
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_parties IS 
'Unified table for all business parties: customers, suppliers, employees, and others. Supports contact and financial information management.';

COMMENT ON COLUMN finance_parties.id IS 'Unique identifier for the party';
COMMENT ON COLUMN finance_parties.tenant_id IS 'Reference to tenant for multi-tenancy isolation';
COMMENT ON COLUMN finance_parties.entity_id IS 'Reference to entity (subsidiary/department/region)';
COMMENT ON COLUMN finance_parties.party_type IS 'Type of party: customer, supplier, employee, other';
COMMENT ON COLUMN finance_parties.code IS 'Party code (e.g., "CUST001", "SUPP001") - unique within party type';
COMMENT ON COLUMN finance_parties.name IS 'Party name for business use';
COMMENT ON COLUMN finance_parties.display_name IS 'Name for display purposes (may differ from legal name)';
COMMENT ON COLUMN finance_parties.legal_name IS 'Legal registered name';
COMMENT ON COLUMN finance_parties.billing_address IS 'Billing address in JSON format';
COMMENT ON COLUMN finance_parties.shipping_address IS 'Shipping address in JSON format';
COMMENT ON COLUMN finance_parties.currency IS 'Default currency for transactions with this party';
COMMENT ON COLUMN finance_parties.payment_terms_days IS 'Default payment terms in days';
COMMENT ON COLUMN finance_parties.credit_limit IS 'Maximum credit limit allowed';
COMMENT ON COLUMN finance_parties.tax_id IS 'Tax identification number';
COMMENT ON COLUMN finance_parties.tax_exempt IS 'Whether party is exempt from taxes';
COMMENT ON COLUMN finance_parties.receivable_account_id IS 'Default receivable account for customer transactions';
COMMENT ON COLUMN finance_parties.payable_account_id IS 'Default payable account for supplier transactions';

-- =====================================================================
-- MIGRATION 007: TRANSACTIONS
-- =====================================================================
-- File: 007_create_transactions.up.sql
-- Description: Create main transactions table for all financial documents
-- Dependencies: entities, finance_parties, users tables
-- =====================================================================

CREATE TABLE finance_transactions (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Transaction identification
    number VARCHAR(50) NOT NULL,
    type transaction_type_enum NOT NULL,
    status transaction_status_enum NOT NULL DEFAULT 'draft',
    
    -- Date information
    transaction_date DATE NOT NULL,
    posting_date DATE NOT NULL,
    due_date DATE,
    
    -- Currency and exchange rate
    base_currency currency_code NOT NULL,
    transaction_currency currency_code NOT NULL,
    exchange_rate DECIMAL(10,6) NOT NULL DEFAULT 1.0,
    
    -- Amount information
    total_amount DECIMAL(19,4) NOT NULL,
    base_total_amount DECIMAL(19,4) NOT NULL,
    tax_amount DECIMAL(19,4) DEFAULT 0.00,
    base_tax_amount DECIMAL(19,4) DEFAULT 0.00,
    
    -- Party information
    party_type party_type_enum,
    party_id UUID REFERENCES finance_parties(id),
    
    -- Reference and linking
    reference_type VARCHAR(50),
    reference_id UUID,
    reference_number VARCHAR(100),
    external_reference VARCHAR(100),
    
    -- Description and notes
    description TEXT,
    notes TEXT,
    memo VARCHAR(255),
    
    -- Workflow and approval
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    approval_status VARCHAR(20) DEFAULT 'not_required' 
        CHECK (approval_status IN ('not_required', 'pending', 'approved', 'rejected')),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approval_comments TEXT,
    
    -- Posting information
    posted_by UUID REFERENCES users(id),
    posted_at TIMESTAMPTZ,
    
    -- Reversal information
    is_reversal BOOLEAN NOT NULL DEFAULT false,
    reversed_transaction_id UUID REFERENCES finance_transactions(id),
    reversal_reason TEXT,
    reversal_date DATE,
    
    -- Additional information
    tags TEXT[], -- Array of tags for categorization
    additional_info JSONB DEFAULT '{}'::jsonb,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    -- Business constraints
    CONSTRAINT uk_finance_transactions_tenant_number UNIQUE(tenant_id, number),
    CONSTRAINT ck_finance_transactions_total_amount CHECK (total_amount > 0),
    CONSTRAINT ck_finance_transactions_base_total_amount CHECK (base_total_amount > 0),
    CONSTRAINT ck_finance_transactions_exchange_rate CHECK (exchange_rate > 0),
    CONSTRAINT ck_finance_transactions_tax_amounts CHECK (tax_amount >= 0 AND base_tax_amount >= 0),
    CONSTRAINT ck_finance_transactions_dates CHECK (posting_date >= transaction_date),
    CONSTRAINT ck_finance_transactions_due_date CHECK (due_date IS NULL OR due_date >= transaction_date),
    CONSTRAINT ck_finance_transactions_reversal CHECK (
        (is_reversal = false AND reversed_transaction_id IS NULL) OR 
        (is_reversal = true AND reversed_transaction_id IS NOT NULL)
    )
);

-- Indexes for performance
CREATE INDEX idx_finance_transactions_tenant ON finance_transactions(tenant_id);
CREATE INDEX idx_finance_transactions_entity ON finance_transactions(entity_id);
CREATE INDEX idx_finance_transactions_number ON finance_transactions(tenant_id, number);
CREATE INDEX idx_finance_transactions_type ON finance_transactions(tenant_id, type);
CREATE INDEX idx_finance_transactions_status ON finance_transactions(tenant_id, status);
CREATE INDEX idx_finance_transactions_date ON finance_transactions(tenant_id, transaction_date);
CREATE INDEX idx_finance_transactions_posting_date ON finance_transactions(tenant_id, posting_date);
CREATE INDEX idx_finance_transactions_party ON finance_transactions(tenant_id, party_type, party_id);
CREATE INDEX idx_finance_transactions_reference ON finance_transactions(tenant_id, reference_type, reference_id);
CREATE INDEX idx_finance_transactions_approval ON finance_transactions(tenant_id, approval_status) WHERE approval_status = 'pending';

-- Row Level Security
ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_transactions_tenant_isolation 
    ON finance_transactions
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_transactions_admin_access 
    ON finance_transactions
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_transactions IS 
'Main table for all financial transactions including journal entries, invoices, payments, and other financial documents. Supports multi-currency, approval workflows, and audit trail.';

COMMENT ON COLUMN finance_transactions.id IS 'Unique identifier for the transaction';
COMMENT ON COLUMN finance_transactions.tenant_id IS 'Reference to tenant for multi-tenancy isolation';
COMMENT ON COLUMN finance_transactions.entity_id IS 'Reference to entity (subsidiary/department/region)';
COMMENT ON COLUMN finance_transactions.number IS 'Transaction number (auto-generated, unique within tenant)';
COMMENT ON COLUMN finance_transactions.type IS 'Type of transaction: journal_entry, sales_invoice, purchase_invoice, etc.';
COMMENT ON COLUMN finance_transactions.status IS 'Current status in workflow: draft, pending_approval, approved, posted, etc.';
COMMENT ON COLUMN finance_transactions.transaction_date IS 'Date when transaction occurred';
COMMENT ON COLUMN finance_transactions.posting_date IS 'Date when transaction should be posted to accounts';
COMMENT ON COLUMN finance_transactions.due_date IS 'Due date for payment (applicable to invoices)';
COMMENT ON COLUMN finance_transactions.base_currency IS 'Base currency of the entity/chart of accounts';
COMMENT ON COLUMN finance_transactions.transaction_currency IS 'Currency used for this specific transaction';
COMMENT ON COLUMN finance_transactions.exchange_rate IS 'Exchange rate from transaction currency to base currency';
COMMENT ON COLUMN finance_transactions.total_amount IS 'Total amount in transaction currency';
COMMENT ON COLUMN finance_transactions.base_total_amount IS 'Total amount in base currency';
COMMENT ON COLUMN finance_transactions.party_id IS 'Reference to party involved in transaction (customer, supplier, etc.)';
COMMENT ON COLUMN finance_transactions.reference_type IS 'Type of referenced document (e.g., "sales_order", "purchase_order")';
COMMENT ON COLUMN finance_transactions.reference_id IS 'ID of referenced document';
COMMENT ON COLUMN finance_transactions.requires_approval IS 'Whether transaction requires approval before posting';
COMMENT ON COLUMN finance_transactions.is_reversal IS 'Whether this transaction reverses another transaction';
COMMENT ON COLUMN finance_transactions.reversed_transaction_id IS 'ID of transaction being reversed';

-- =====================================================================
-- MIGRATION 008: TRANSACTION ENTRIES
-- =====================================================================
-- File: 008_create_transaction_entries.up.sql
-- Description: Create transaction entries for double-entry bookkeeping
-- Dependencies: finance_transactions, finance_accounts, finance_cost_centers, finance_projects, finance_parties
-- =====================================================================

CREATE TABLE finance_transaction_entries (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    transaction_id UUID NOT NULL REFERENCES finance_transactions(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES finance_accounts(id) ON DELETE RESTRICT,
    
    -- Entry ordering and identification
    line_number INTEGER NOT NULL,
    entry_type VARCHAR(20) DEFAULT 'standard' 
        CHECK (entry_type IN ('standard', 'tax', 'rounding', 'exchange_difference')),
    
    -- Double-entry amounts in transaction currency
    debit_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    credit_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    
    -- Amounts in base currency
    base_debit_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    base_credit_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    
    -- Cost allocation
    cost_center_id UUID REFERENCES finance_cost_centers(id),
    project_id UUID REFERENCES finance_projects(id),
    
    -- Party information for receivables/payables tracking
    party_type party_type_enum,
    party_id UUID REFERENCES finance_parties(id),
    
    -- Description and memo
    description TEXT,
    memo VARCHAR(255),
    
    -- Tax information
    tax_amount DECIMAL(19,4) DEFAULT 0.00,
    tax_rate DECIMAL(5,4),
    tax_account_id UUID REFERENCES finance_accounts(id),
    tax_code VARCHAR(20),
    
    -- Quantity tracking (for inventory/unit-based entries)
    quantity DECIMAL(15,4),
    unit_price DECIMAL(19,4),
    unit_of_measure VARCHAR(20),
    
    -- Additional dimensions and information
    department_id UUID REFERENCES entities(uuid), -- Reference to department entity
    region_id UUID REFERENCES entities(uuid),     -- Reference to region entity
    additional_info JSONB DEFAULT '{}'::jsonb,
    
    -- Audit timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Business constraints
    CONSTRAINT uk_finance_transaction_entries_line UNIQUE(tenant_id, transaction_id, line_number),
    CONSTRAINT ck_finance_transaction_entries_amounts CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR 
        (credit_amount > 0 AND debit_amount = 0)
    ),
    CONSTRAINT ck_finance_transaction_entries_base_amounts CHECK (
        (base_debit_amount > 0 AND base_credit_amount = 0) OR 
        (base_credit_amount > 0 AND base_debit_amount = 0)
    ),
    CONSTRAINT ck_finance_transaction_entries_non_negative CHECK (
        debit_amount >= 0 AND credit_amount >= 0 AND 
        base_debit_amount >= 0 AND base_credit_amount >= 0
    ),
    CONSTRAINT ck_finance_transaction_entries_tax_amount CHECK (tax_amount >= 0),
    CONSTRAINT ck_finance_transaction_entries_tax_rate CHECK (tax_rate IS NULL OR (tax_rate >= 0 AND tax_rate <= 100)),
    CONSTRAINT ck_finance_transaction_entries_quantity CHECK (quantity IS NULL OR quantity > 0),
    CONSTRAINT ck_finance_transaction_entries_unit_price CHECK (unit_price IS NULL OR unit_price >= 0)
);

-- Indexes for performance
CREATE INDEX idx_finance_transaction_entries_tenant ON finance_transaction_entries(tenant_id);
CREATE INDEX idx_finance_transaction_entries_entity ON finance_transaction_entries(entity_id);
CREATE INDEX idx_finance_transaction_entries_transaction ON finance_transaction_entries(tenant_id, transaction_id);
CREATE INDEX idx_finance_transaction_entries_account ON finance_transaction_entries(tenant_id, account_id);
CREATE INDEX idx_finance_transaction_entries_cost_center ON finance_transaction_entries(tenant_id, cost_center_id);
CREATE INDEX idx_finance_transaction_entries_project ON finance_transaction_entries(tenant_id, project_id);
CREATE INDEX idx_finance_transaction_entries_party ON finance_transaction_entries(tenant_id, party_type, party_id);
CREATE INDEX idx_finance_transaction_entries_department ON finance_transaction_entries(department_id);
CREATE INDEX idx_finance_transaction_entries_region ON finance_transaction_entries(region_id);

-- Row Level Security
ALTER TABLE finance_transaction_entries ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_transaction_entries_tenant_isolation 
    ON finance_transaction_entries
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_transaction_entries_admin_access 
    ON finance_transaction_entries
    FOR ALL TO admin_role 
    USING (true);

-- Table and column comments
COMMENT ON TABLE finance_transaction_entries IS 
'Double-entry bookkeeping entries for each transaction. Each entry represents either a debit or credit to a specific account, ensuring accounting equation balance.';

COMMENT ON COLUMN finance_transaction_entries.id IS 'Unique identifier for the transaction entry';
COMMENT ON COLUMN finance_transaction_entries.tenant_id IS 'Reference to tenant for multi-tenancy isolation';
COMMENT ON COLUMN finance_transaction_entries.entity_id IS 'Reference to entity (subsidiary/department/region)';
COMMENT ON COLUMN finance_transaction_entries.transaction_id IS 'Reference to parent transaction';
COMMENT ON COLUMN finance_transaction_entries.account_id IS 'Account being debited or credited';
COMMENT ON COLUMN finance_transaction_entries.line_number IS 'Line number within transaction (for ordering)';
COMMENT ON COLUMN finance_transaction_entries.entry_type IS 'Type of entry: standard, tax, rounding, exchange_difference';
COMMENT ON COLUMN finance_transaction_entries.debit_amount IS 'Debit amount in transaction currency (positive for debits, zero for credits)';
COMMENT ON COLUMN finance_transaction_entries.credit_amount IS 'Credit amount in transaction currency (positive for credits, zero for debits)';
COMMENT ON COLUMN finance_transaction_entries.base_debit_amount IS 'Debit amount in base currency';
COMMENT ON COLUMN finance_transaction_entries.base_credit_amount IS 'Credit amount in base currency';
COMMENT ON COLUMN finance_transaction_entries.cost_center_id IS 'Cost center for expense allocation and reporting';
COMMENT ON COLUMN finance_transaction_entries.project_id IS 'Project for project-based cost tracking';
COMMENT ON COLUMN finance_transaction_entries.party_id IS 'Party associated with this entry (for receivables/payables)';
COMMENT ON COLUMN finance_transaction_entries.tax_amount IS 'Tax amount for this entry';
COMMENT ON COLUMN finance_transaction_entries.tax_rate IS 'Tax rate percentage applied';
COMMENT ON COLUMN finance_transaction_entries.quantity IS 'Quantity for unit-based entries (inventory, etc.)';
COMMENT ON COLUMN finance_transaction_entries.unit_price IS 'Unit price for unit-based entries';
COMMENT ON COLUMN finance_transaction_entries.department_id IS 'Reference to department entity for organizational reporting';
COMMENT ON COLUMN finance_transaction_entries.region_id IS 'Reference to region entity for geographical reporting';

-- =====================================================================
-- MIGRATION 009: SUPPORTING TABLES
-- =====================================================================
-- File: 009_create_supporting_tables.up.sql
-- Description: Create payment terms, fiscal periods, exchange rates, and document sequences
-- Dependencies: entities, users tables
-- =====================================================================

-- Payment Terms Table
CREATE TABLE finance_payment_terms (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Term identification
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Payment terms configuration
    days INTEGER NOT NULL DEFAULT 30 CHECK (days >= 0),
    discount_days INTEGER DEFAULT 0 CHECK (discount_days >= 0),
    discount_percentage DECIMAL(5,4) DEFAULT 0.00 
        CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_payment_terms_tenant_code UNIQUE(tenant_id, code),
    CONSTRAINT uk_finance_payment_terms_tenant_name UNIQUE(tenant_id, name),
    CONSTRAINT ck_finance_payment_terms_discount_days CHECK (discount_days <= days)
);

-- Fiscal Periods Table
CREATE TABLE finance_fiscal_periods (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Period information
    name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    fiscal_year INTEGER NOT NULL,
    period_number INTEGER NOT NULL CHECK (period_number BETWEEN 1 AND 13),
    
    -- Status and control
    is_closed BOOLEAN NOT NULL DEFAULT false,
    closed_by UUID REFERENCES users(id),
    closed_at TIMESTAMPTZ,
    close_reason TEXT,
    
    -- Special period types
    is_adjustment_period BOOLEAN NOT NULL DEFAULT false,
    is_year_end_period BOOLEAN NOT NULL DEFAULT false,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_fiscal_periods_tenant_year_period UNIQUE(tenant_id, entity_id, fiscal_year, period_number),
    CONSTRAINT ck_finance_fiscal_periods_dates CHECK (start_date < end_date)
);

-- Exchange Rates Table
CREATE TABLE finance_exchange_rates (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Currency pair
    from_currency currency_code NOT NULL,
    to_currency currency_code NOT NULL,
    
    -- Rate information
    rate DECIMAL(10,6) NOT NULL CHECK (rate > 0),
    effective_date DATE NOT NULL,
    expiry_date DATE,
    
    -- Rate source and type
    rate_type VARCHAR(20) NOT NULL DEFAULT 'manual' 
        CHECK (rate_type IN ('manual', 'automatic', 'bank', 'central_bank', 'market')),
    source VARCHAR(100),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    
    CONSTRAINT uk_finance_exchange_rates_unique UNIQUE(tenant_id, entity_id, from_currency, to_currency, effective_date),
    CONSTRAINT ck_finance_exchange_rates_expiry CHECK (expiry_date IS NULL OR expiry_date > effective_date),
    CONSTRAINT ck_finance_exchange_rates_different_currencies CHECK (from_currency != to_currency)
);

-- Document Sequences Table
CREATE TABLE finance_document_sequences (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Sequence configuration
    document_type VARCHAR(50) NOT NULL,
    prefix VARCHAR(20) NOT NULL DEFAULT '',
    suffix VARCHAR(20) NOT NULL DEFAULT '',
    
    -- Number generation
    current_number BIGINT NOT NULL DEFAULT 1,
    increment_by INTEGER NOT NULL DEFAULT 1 CHECK (increment_by > 0),
    pad_length INTEGER NOT NULL DEFAULT 6 CHECK (pad_length > 0),
    
    -- Reset configuration
    reset_frequency VARCHAR(20) DEFAULT 'never' 
        CHECK (reset_frequency IN ('never', 'yearly', 'monthly', 'daily')),
    last_reset_date DATE,
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_document_sequences_unique UNIQUE(tenant_id, entity_id, document_type)
);

-- Indexes and RLS for all supporting tables
CREATE INDEX idx_finance_payment_terms_tenant ON finance_payment_terms(tenant_id);
CREATE INDEX idx_finance_fiscal_periods_tenant ON finance_fiscal_periods(tenant_id);
CREATE INDEX idx_finance_exchange_rates_tenant ON finance_exchange_rates(tenant_id);
CREATE INDEX idx_finance_document_sequences_tenant ON finance_document_sequences(tenant_id);

-- Enable RLS
ALTER TABLE finance_payment_terms ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_fiscal_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_exchange_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_document_sequences ENABLE ROW LEVEL SECURITY;

-- RLS Policies (tenant isolation)
CREATE POLICY finance_payment_terms_tenant_isolation ON finance_payment_terms
    FOR ALL TO application_role USING (tenant_id = current_tenant_id()) WITH CHECK (tenant_id = current_tenant_id());
CREATE POLICY finance_fiscal_periods_tenant_isolation ON finance_fiscal_periods
    FOR ALL TO application_role USING (tenant_id = current_tenant_id()) WITH CHECK (tenant_id = current_tenant_id());
CREATE POLICY finance_exchange_rates_tenant_isolation ON finance_exchange_rates
    FOR ALL TO application_role USING (tenant_id = current_tenant_id()) WITH CHECK (tenant_id = current_tenant_id());
CREATE POLICY finance_document_sequences_tenant_isolation ON finance_document_sequences
    FOR ALL TO application_role USING (tenant_id = current_tenant_id()) WITH CHECK (tenant_id = current_tenant_id());

-- Admin access policies
CREATE POLICY finance_payment_terms_admin_access ON finance_payment_terms FOR ALL TO admin_role USING (true);
CREATE POLICY finance_fiscal_periods_admin_access ON finance_fiscal_periods FOR ALL TO admin_role USING (true);
CREATE POLICY finance_exchange_rates_admin_access ON finance_exchange_rates FOR ALL TO admin_role USING (true);
CREATE POLICY finance_document_sequences_admin_access ON finance_document_sequences FOR ALL TO admin_role USING (true);

-- Comments for supporting tables
COMMENT ON TABLE finance_payment_terms IS 'Payment terms definitions for invoicing, collections, and supplier payments';
COMMENT ON TABLE finance_fiscal_periods IS 'Fiscal periods for financial reporting, budget tracking, and period-end controls';
COMMENT ON TABLE finance_exchange_rates IS 'Currency exchange rates for multi-currency transaction processing and reporting';
COMMENT ON TABLE finance_document_sequences IS 'Auto-numbering sequences for financial documents with configurable format and reset options';

-- =====================================================================
-- MIGRATION 010: TRIGGERS AND FUNCTIONS
-- =====================================================================
-- File: 010_create_triggers_and_functions.up.sql
-- Description: Create business logic functions and triggers
-- Dependencies: All finance tables
-- =====================================================================

-- Function to update updated_at and version columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS 
'Trigger function to automatically update updated_at timestamp and increment version on record updates';

-- Function to validate transaction balance
CREATE OR REPLACE FUNCTION validate_transaction_balance()
RETURNS TRIGGER AS $$
DECLARE
    total_debits DECIMAL(19,4);
    total_credits DECIMAL(19,4);
    base_total_debits DECIMAL(19,4);
    base_total_credits DECIMAL(19,4);
    transaction_rec RECORD;
BEGIN
    -- Get transaction record for context
    SELECT * INTO transaction_rec 
    FROM finance_transactions 
    WHERE id = COALESCE(NEW.transaction_id, OLD.transaction_id)
    AND tenant_id = COALESCE(NEW.tenant_id, OLD.tenant_id);
    
    -- Only validate for posted transactions
    IF transaction_rec.status != 'posted' THEN
        RETURN COALESCE(NEW, OLD);
    END IF;
    
    -- Calculate totals for the transaction
    SELECT 
        COALESCE(SUM(debit_amount), 0),
        COALESCE(SUM(credit_amount), 0),
        COALESCE(SUM(base_debit_amount), 0),
        COALESCE(SUM(base_credit_amount), 0)
    INTO total_debits, total_credits, base_total_debits, base_total_credits
    FROM finance_transaction_entries
    WHERE transaction_id = transaction_rec.id
    AND tenant_id = transaction_rec.tenant_id;
    
    -- Check if transaction is balanced (allow for small rounding differences)
    IF ABS(total_debits - total_credits) > 0.01 THEN
        RAISE EXCEPTION 'Transaction % entries must be balanced. Debits: %, Credits: %', 
            transaction_rec.number, total_debits, total_credits;
    END IF;
    
    IF ABS(base_total_debits - base_total_credits) > 0.01 THEN
        RAISE EXCEPTION 'Transaction % base currency entries must be balanced. Base Debits: %, Base Credits: %', 
            transaction_rec.number, base_total_debits, base_total_credits;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_transaction_balance() IS 
'Trigger function to ensure transaction entries are balanced (total debits = total credits) for posted transactions';

-- Function to auto-generate document numbers
CREATE OR REPLACE FUNCTION generate_document_number(
    p_tenant_id UUID,
    p_entity_id UUID,
    p_document_type VARCHAR,
    p_date DATE DEFAULT CURRENT_DATE
) RETURNS VARCHAR AS $$
DECLARE
    seq_record RECORD;
    next_number BIGINT;
    formatted_number VARCHAR;
    date_suffix VARCHAR := '';
BEGIN
    -- Get the sequence configuration
    SELECT * INTO seq_record
    FROM finance_document_sequences
    WHERE tenant_id = p_tenant_id 
    AND entity_id = p_entity_id
    AND document_type = p_document_type 
    AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'No active document sequence found for tenant %, entity %, type %', 
            p_tenant_id, p_entity_id, p_document_type;
    END IF;
    
    -- Check if we need to reset the sequence
    IF seq_record.reset_frequency != 'never' THEN
        CASE seq_record.reset_frequency
            WHEN 'yearly' THEN
                IF seq_record.last_reset_date IS NULL OR 
                   EXTRACT(YEAR FROM seq_record.last_reset_date) < EXTRACT(YEAR FROM p_date) THEN
                    UPDATE finance_document_sequences 
                    SET current_number = 1, last_reset_date = p_date 
                    WHERE id = seq_record.id;
                    next_number := 1;
                    date_suffix := EXTRACT(YEAR FROM p_date)::TEXT;
                ELSE
                    next_number := seq_record.current_number;
                    date_suffix := EXTRACT(YEAR FROM seq_record.last_reset_date)::TEXT;
                END IF;
                
            WHEN 'monthly' THEN
                IF seq_record.last_reset_date IS NULL OR 
                   DATE_TRUNC('month', seq_record.last_reset_date) < DATE_TRUNC('month', p_date) THEN
                    UPDATE finance_document_sequences 
                    SET current_number = 1, last_reset_date = p_date 
                    WHERE id = seq_record.id;
                    next_number := 1;
                ELSE
                    next_number := seq_record.current_number;
                END IF;
                date_suffix := TO_CHAR(p_date, 'YYYYMM');
                
            WHEN 'daily' THEN
                IF seq_record.last_reset_date IS NULL OR seq_record.last_reset_date < p_date THEN
                    UPDATE finance_document_sequences 
                    SET current_number = 1, last_reset_date = p_date 
                    WHERE id = seq_record.id;
                    next_number := 1;
                ELSE
                    next_number := seq_record.current_number;
                END IF;
                date_suffix := TO_CHAR(p_date, 'YYYYMMDD');
        END CASE;
    ELSE
        next_number := seq_record.current_number;
    END IF;
    
    -- Increment the sequence
    UPDATE finance_document_sequences 
    SET current_number = current_number + seq_record.increment_by,
        updated_at = NOW()
    WHERE id = seq_record.id;
    
    -- Format the number
    formatted_number := seq_record.prefix || 
                       date_suffix ||
                       LPAD(next_number::TEXT, seq_record.pad_length, '0') || 
                       seq_record.suffix;
    
    RETURN formatted_number;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_document_number(UUID, UUID, VARCHAR, DATE) IS 
'Generate next document number for given tenant, entity, and document type with automatic reset based on configuration';

-- Apply triggers to tables with updated_at column
CREATE TRIGGER update_finance_accounts_updated_at 
    BEFORE UPDATE ON finance_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_accounts_updated_at 
    BEFORE UPDATE ON finance_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_cost_centers_updated_at 
    BEFORE UPDATE ON finance_cost_centers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_projects_updated_at 
    BEFORE UPDATE ON finance_projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_parties_updated_at 
    BEFORE UPDATE ON finance_parties
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_transactions_updated_at 
    BEFORE UPDATE ON finance_transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_payment_terms_updated_at 
    BEFORE UPDATE ON finance_payment_terms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_fiscal_periods_updated_at 
    BEFORE UPDATE ON finance_fiscal_periods
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_finance_document_sequences_updated_at 
    BEFORE UPDATE ON finance_document_sequences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Apply balance validation trigger
CREATE TRIGGER validate_transaction_balance_trigger
    AFTER INSERT OR UPDATE OR DELETE ON finance_transaction_entries
    FOR EACH ROW EXECUTE FUNCTION validate_transaction_balance();

COMMENT ON TRIGGER validate_transaction_balance_trigger ON finance_transaction_entries IS 
'Ensures transaction entries are balanced (debits = credits) for posted transactions';

-- =====================================================================
-- MIGRATION 011: ACCOUNT BALANCE TRACKING
-- =====================================================================
-- File: 011_create_account_balances.up.sql
-- Description: Create account balance tracking for period-wise balance history
-- Dependencies: finance_accounts, finance_fiscal_periods
-- =====================================================================

CREATE TABLE finance_account_balances (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES finance_accounts(id) ON DELETE CASCADE,
    
    -- Period and date information
    fiscal_period_id UUID REFERENCES finance_fiscal_periods(id),
    balance_date DATE NOT NULL,
    
    -- Balance amounts
    opening_balance DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    closing_balance DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    debit_total DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    credit_total DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    
    -- Currency information
    currency currency_code NOT NULL,
    base_currency_opening_balance DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    base_currency_closing_balance DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    
    -- Balance validation
    is_balanced BOOLEAN NOT NULL DEFAULT true,
    last_calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_finance_account_balances_unique UNIQUE(tenant_id, account_id, balance_date),
    CONSTRAINT ck_finance_account_balances_calculation CHECK (
        ABS((opening_balance + debit_total - credit_total) - closing_balance) < 0.01
    )
);

-- Indexes
CREATE INDEX idx_finance_account_balances_tenant ON finance_account_balances(tenant_id);
CREATE INDEX idx_finance_account_balances_account ON finance_account_balances(account_id);
CREATE INDEX idx_finance_account_balances_date ON finance_account_balances(tenant_id, balance_date);
CREATE INDEX idx_finance_account_balances_period ON finance_account_balances(fiscal_period_id);

-- RLS
ALTER TABLE finance_account_balances ENABLE ROW LEVEL SECURITY;
CREATE POLICY finance_account_balances_tenant_isolation ON finance_account_balances
    FOR ALL TO application_role USING (tenant_id = current_tenant_id()) WITH CHECK (tenant_id = current_tenant_id());
CREATE POLICY finance_account_balances_admin_access ON finance_account_balances FOR ALL TO admin_role USING (true);

COMMENT ON TABLE finance_account_balances IS 
'Historical account balance tracking for each period and date, enabling balance sheet preparation and audit trails';

-- =====================================================================
-- MIGRATION 012: BANK ACCOUNTS AND BANK MANAGEMENT
-- =====================================================================
-- File: 012_create_bank_management.up.sql
-- Description: Create bank account management and bank transaction processing
-- Dependencies: finance_accounts, finance_parties
-- =====================================================================

-- Bank Account Master
CREATE TABLE finance_bank_accounts (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Bank account identification
    account_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    bank_name VARCHAR(255) NOT NULL,
    bank_code VARCHAR(20),
    branch_name VARCHAR(255),
    branch_code VARCHAR(20),
    
    -- International banking
    iban VARCHAR(50),
    swift_code VARCHAR(20),
    routing_number VARCHAR(30),
    sort_code VARCHAR(10),
    
    -- Account details
    account_type VARCHAR(30) NOT NULL DEFAULT 'checking' 
        CHECK (account_type IN ('checking', 'savings', 'current', 'deposit', 'loan', 'credit_card')),
    currency currency_code NOT NULL,
    
    -- Linked GL account
    gl_account_id UUID NOT NULL REFERENCES finance_accounts(id),
    
    -- Bank account configuration
    opening_balance DECIMAL(19,4) DEFAULT 0.00,
    current_balance DECIMAL(19,4) DEFAULT 0.00,
    minimum_balance DECIMAL(19,4) DEFAULT 0.00,
    overdraft_limit DECIMAL(19,4) DEFAULT 0.00,
    
    -- Status and control
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    allow_online_banking BOOLEAN NOT NULL DEFAULT false,
    
    -- Contact information
    contact_person VARCHAR(255),
    contact_phone VARCHAR(30),
    contact_email VARCHAR(254),
    
    -- Address
    bank_address JSONB DEFAULT '{}'::jsonb,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_bank_accounts_tenant_number UNIQUE(tenant_id, account_number),
    CONSTRAINT ck_finance_bank_accounts_balances CHECK (minimum_balance >= 0 AND overdraft_limit >= 0)
);

-- Bank Statements
CREATE TABLE finance_bank_statements (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    bank_account_id UUID NOT NULL REFERENCES finance_bank_accounts(id) ON DELETE CASCADE,
    
    -- Statement information
    statement_number VARCHAR(50),
    statement_date DATE NOT NULL,
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    
    -- Balance information
    opening_balance DECIMAL(19,4) NOT NULL,
    closing_balance DECIMAL(19,4) NOT NULL,
    
    -- Reconciliation status
    is_reconciled BOOLEAN NOT NULL DEFAULT false,
    reconciled_by UUID REFERENCES users(id),
    reconciled_at TIMESTAMPTZ,
    
    -- Import information
    imported_from VARCHAR(100),
    import_reference VARCHAR(100),
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_bank_statements_unique UNIQUE(tenant_id, bank_account_id, statement_date),
    CONSTRAINT ck_finance_bank_statements_dates CHECK (from_date <= to_date)
);

-- Bank Transactions
CREATE TABLE finance_bank_transactions (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    bank_account_id UUID NOT NULL REFERENCES finance_bank_accounts(id) ON DELETE CASCADE,
    bank_statement_id UUID REFERENCES finance_bank_statements(id),
    
    -- Transaction details
    transaction_date DATE NOT NULL,
    value_date DATE,
    description TEXT NOT NULL,
    reference_number VARCHAR(100),
    bank_reference VARCHAR(100),
    
    -- Amount information
    debit_amount DECIMAL(19,4) DEFAULT 0.00,
    credit_amount DECIMAL(19,4) DEFAULT 0.00,
    balance_after DECIMAL(19,4),
    
    -- Transaction classification
    transaction_type VARCHAR(30) 
        CHECK (transaction_type IN ('deposit', 'withdrawal', 'transfer', 'fee', 'interest', 'dividend', 'other')),
    
    -- Reconciliation
    is_reconciled BOOLEAN NOT NULL DEFAULT false,
    reconciled_with_id UUID REFERENCES finance_transactions(id),
    reconciliation_difference DECIMAL(19,4) DEFAULT 0.00,
    
    -- Party information
    party_name VARCHAR(255),
    party_account VARCHAR(50),
    
    -- Additional information
    additional_info JSONB DEFAULT '{}'::jsonb,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT ck_finance_bank_transactions_amounts CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR (credit_amount > 0 AND debit_amount = 0)
    )
);

-- Indexes and RLS for bank tables
CREATE INDEX idx_finance_bank_accounts_tenant ON finance_bank_accounts(tenant_id);
CREATE INDEX idx_finance_bank_statements_tenant ON finance_bank_statements(tenant_id);
CREATE INDEX idx_finance_bank_transactions_tenant ON finance_bank_transactions(tenant_id);

ALTER TABLE finance_bank_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_statements ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_transactions ENABLE ROW LEVEL SECURITY;

-- Comments
COMMENT ON TABLE finance_bank_accounts IS 'Bank account master data with banking details and GL account linkage';
COMMENT ON TABLE finance_bank_statements IS 'Bank statements for reconciliation and balance tracking';
COMMENT ON TABLE finance_bank_transactions IS 'Individual bank transactions from statements for reconciliation';

-- =====================================================================
-- MIGRATION 013: FINANCE BOOKS AND TAX MANAGEMENT
-- =====================================================================
-- File: 013_create_finance_books_and_tax.up.sql
-- Description: Create multiple finance books and tax management
-- Dependencies: finance_accounts, finance_transactions
-- =====================================================================

-- Finance Books (Multiple Books Accounting)
CREATE TABLE finance_books (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Book identification
    book_code VARCHAR(20) NOT NULL,
    book_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Book configuration
    book_type VARCHAR(30) NOT NULL DEFAULT 'general' 
        CHECK (book_type IN ('general', 'tax', 'management', 'statutory', 'ifrs', 'gaap')),
    base_currency currency_code NOT NULL,
    
    -- Book behavior
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    allow_negative_stock BOOLEAN NOT NULL DEFAULT false,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_books_tenant_code UNIQUE(tenant_id, book_code),
    CONSTRAINT uk_finance_books_tenant_name UNIQUE(tenant_id, book_name)
);

-- Tax Categories
CREATE TABLE finance_tax_categories (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Tax category identification
    category_code VARCHAR(20) NOT NULL,
    category_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Tax configuration
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_tax_categories_tenant_code UNIQUE(tenant_id, category_code)
);

-- Tax Rules
CREATE TABLE finance_tax_rules (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Tax rule identification
    rule_name VARCHAR(255) NOT NULL,
    tax_type VARCHAR(30) NOT NULL 
        CHECK (tax_type IN ('sales_tax', 'purchase_tax', 'vat', 'gst', 'withholding', 'excise', 'customs')),
    
    -- Rule conditions
    use_for_shopping_cart BOOLEAN NOT NULL DEFAULT false,
    tax_category_id UUID REFERENCES finance_tax_categories(id),
    
    -- Geographic and party conditions
    customer_group VARCHAR(100),
    supplier_group VARCHAR(100),
    item_group VARCHAR(100),
    
    -- Tax calculation
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.00,
    tax_amount DECIMAL(19,4),
    
    -- Tax accounts
    account_head_id UUID NOT NULL REFERENCES finance_accounts(id),
    cost_center_id UUID REFERENCES finance_cost_centers(id),
    
    -- Validity
    valid_from DATE,
    valid_upto DATE,
    priority INTEGER DEFAULT 1,
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT ck_finance_tax_rules_rate_or_amount CHECK (
        (tax_rate > 0 AND tax_amount IS NULL) OR (tax_amount > 0 AND tax_rate = 0)
    ),
    CONSTRAINT ck_finance_tax_rules_dates CHECK (valid_from IS NULL OR valid_upto IS NULL OR valid_from <= valid_upto)
);

-- Tax Withholding Categories
CREATE TABLE finance_tax_withholding_categories (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Withholding category
    category_name VARCHAR(255) NOT NULL,
    category_code VARCHAR(20) NOT NULL,
    description TEXT,
    
    -- Withholding configuration
    withholding_type VARCHAR(20) NOT NULL 
        CHECK (withholding_type IN ('tds', 'tcs', 'income_tax', 'professional_tax')),
    
    -- Tax rates for different slabs
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.00,
    single_threshold DECIMAL(19,4),
    cumulative_threshold DECIMAL(19,4),
    
    -- Tax accounts
    account_head_id UUID NOT NULL REFERENCES finance_accounts(id),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_tax_withholding_tenant_code UNIQUE(tenant_id, category_code)
);

-- Indexes and RLS
CREATE INDEX idx_finance_books_tenant ON finance_books(tenant_id);
CREATE INDEX idx_finance_tax_categories_tenant ON finance_tax_categories(tenant_id);
CREATE INDEX idx_finance_tax_rules_tenant ON finance_tax_rules(tenant_id);
CREATE INDEX idx_finance_tax_withholding_categories_tenant ON finance_tax_withholding_categories(tenant_id);

ALTER TABLE finance_books ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_withholding_categories ENABLE ROW LEVEL SECURITY;

-- Comments
COMMENT ON TABLE finance_books IS 'Multiple finance books for different accounting standards and reporting requirements';
COMMENT ON TABLE finance_tax_categories IS 'Tax category classifications for grouping tax rules and calculations';
COMMENT ON TABLE finance_tax_rules IS 'Tax calculation rules with conditions and rates for automated tax processing';
COMMENT ON TABLE finance_tax_withholding_categories IS 'Tax withholding categories for TDS/TCS and other withholding taxes';

-- =====================================================================
-- MIGRATION 014: PAYMENT PROCESSING AND COLLECTIONS
-- =====================================================================
-- File: 014_create_payment_processing.up.sql
-- Description: Create payment requests, payment orders, and dunning management
-- Dependencies: finance_transactions, finance_parties
-- =====================================================================

-- Payment Requests
CREATE TABLE finance_payment_requests (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Request identification
    request_number VARCHAR(50) NOT NULL,
    request_type VARCHAR(20) NOT NULL DEFAULT 'incoming' 
        CHECK (request_type IN ('incoming', 'outgoing')),
    
    -- Request details
    party_type party_type_enum NOT NULL,
    party_id UUID NOT NULL REFERENCES finance_parties(id),
    
    -- Amount and currency
    requested_amount DECIMAL(19,4) NOT NULL,
    currency currency_code NOT NULL,
    
    -- Request dates
    request_date DATE NOT NULL,
    due_date DATE,
    
    -- Payment method
    payment_method VARCHAR(30) 
        CHECK (payment_method IN ('cash', 'check', 'bank_transfer', 'card', 'online', 'mobile')),
    bank_account_id UUID REFERENCES finance_bank_accounts(id),
    
    -- Status and workflow
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'submitted', 'approved', 'paid', 'cancelled', 'rejected')),
    
    -- Reference information
    reference_type VARCHAR(50),
    reference_id UUID,
    reference_number VARCHAR(100),
    
    -- Additional information
    description TEXT,
    notes TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_payment_requests_tenant_number UNIQUE(tenant_id, request_number),
    CONSTRAINT ck_finance_payment_requests_amount CHECK (requested_amount > 0)
);

-- Payment Orders (Batch Payments)
CREATE TABLE finance_payment_orders (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Order identification
    order_number VARCHAR(50) NOT NULL,
    order_date DATE NOT NULL,
    
    -- Bank and payment details
    bank_account_id UUID NOT NULL REFERENCES finance_bank_accounts(id),
    payment_method VARCHAR(30) NOT NULL
        CHECK (payment_method IN ('ach', 'wire', 'check', 'eft', 'rtgs', 'neft')),
    
    -- Totals
    total_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    total_transactions INTEGER NOT NULL DEFAULT 0,
    
    -- Status and processing
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'approved', 'sent_to_bank', 'processed', 'failed', 'cancelled')),
    
    -- Processing information
    processed_date DATE,
    processed_by UUID REFERENCES users(id),
    bank_reference VARCHAR(100),
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_payment_orders_tenant_number UNIQUE(tenant_id, order_number)
);

-- Payment Order Items
CREATE TABLE finance_payment_order_items (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    payment_order_id UUID NOT NULL REFERENCES finance_payment_orders(id) ON DELETE CASCADE,
    payment_request_id UUID REFERENCES finance_payment_requests(id),
    
    -- Payment details
    party_id UUID NOT NULL REFERENCES finance_parties(id),
    amount DECIMAL(19,4) NOT NULL,
    
    -- Bank details for the payee
    payee_account_number VARCHAR(50),
    payee_bank_code VARCHAR(20),
    payee_bank_name VARCHAR(255),
    
    -- Processing status
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processed', 'failed', 'cancelled')),
    failure_reason TEXT,
    
    -- Reference
    reference_number VARCHAR(100),
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT ck_finance_payment_order_items_amount CHECK (amount > 0)
);

-- Dunning Management
CREATE TABLE finance_dunning (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Dunning identification
    dunning_number VARCHAR(50) NOT NULL,
    dunning_date DATE NOT NULL,
    
    -- Customer and invoice details
    customer_id UUID NOT NULL REFERENCES finance_parties(id),
    sales_invoice_id UUID REFERENCES finance_transactions(id),
    
    -- Outstanding details
    outstanding_amount DECIMAL(19,4) NOT NULL,
    overdue_days INTEGER NOT NULL,
    
    -- Dunning level
    dunning_level INTEGER NOT NULL DEFAULT 1 CHECK (dunning_level BETWEEN 1 AND 5),
    dunning_fee DECIMAL(19,4) DEFAULT 0.00,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'sent', 'acknowledged', 'resolved', 'escalated', 'legal')),
    
    -- Communication details
    sent_date DATE,
    sent_via VARCHAR(20) CHECK (sent_via IN ('email', 'post', 'sms', 'phone', 'courier')),
    contact_person VARCHAR(255),
    
    -- Response tracking
    customer_response TEXT,
    response_date DATE,
    next_action_date DATE,
    
    -- Legal escalation
    escalated_to_legal BOOLEAN NOT NULL DEFAULT false,
    legal_notice_date DATE,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_dunning_tenant_number UNIQUE(tenant_id, dunning_number),
    CONSTRAINT ck_finance_dunning_amounts CHECK (outstanding_amount > 0 AND dunning_fee >= 0)
);

-- Indexes and RLS
CREATE INDEX idx_finance_payment_requests_tenant ON finance_payment_requests(tenant_id);
CREATE INDEX idx_finance_payment_orders_tenant ON finance_payment_orders(tenant_id);
CREATE INDEX idx_finance_payment_order_items_tenant ON finance_payment_order_items(tenant_id);
CREATE INDEX idx_finance_dunning_tenant ON finance_dunning(tenant_id);

ALTER TABLE finance_payment_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_payment_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_payment_order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_dunning ENABLE ROW LEVEL SECURITY;

-- Comments
COMMENT ON TABLE finance_payment_requests IS 'Payment collection requests for incoming and outgoing payments';
COMMENT ON TABLE finance_payment_orders IS 'Batch payment orders for processing multiple payments together';
COMMENT ON TABLE finance_payment_order_items IS 'Individual payment items within a payment order';
COMMENT ON TABLE finance_dunning IS 'Collections management and dunning process for overdue receivables';

-- =====================================================================
-- MIGRATION 015: PERIOD MANAGEMENT AND BUDGETING
-- =====================================================================
-- File: 015_create_period_management_budgeting.up.sql
-- Description: Create period closing, accounting periods, and budget management
-- Dependencies: finance_accounts, finance_fiscal_periods, finance_cost_centers
-- =====================================================================

-- Period Closing Vouchers
CREATE TABLE finance_period_closing_vouchers (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Closing voucher details
    voucher_number VARCHAR(50) NOT NULL,
    fiscal_period_id UUID NOT NULL REFERENCES finance_fiscal_periods(id),
    closing_date DATE NOT NULL,
    
    -- Closing type
    closing_type VARCHAR(20) NOT NULL DEFAULT 'monthly'
        CHECK (closing_type IN ('monthly', 'quarterly', 'yearly', 'adjustment')),
    
    -- Financial totals
    total_income DECIMAL(19,4) DEFAULT 0.00,
    total_expense DECIMAL(19,4) DEFAULT 0.00,
    net_profit_loss DECIMAL(19,4) DEFAULT 0.00,
    
    -- Retained earnings transfer
    retained_earnings_account_id UUID REFERENCES finance_accounts(id),
    closing_transaction_id UUID REFERENCES finance_transactions(id),
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'posted', 'cancelled')),
    
    -- Processing information
    closing_remarks TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_pcv_tenant_number UNIQUE(tenant_id, voucher_number),
    CONSTRAINT uk_finance_pcv_period UNIQUE(tenant_id, entity_id, fiscal_period_id, closing_type)
);

-- Accounting Periods (Enhanced)
CREATE TABLE finance_accounting_periods (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Period identification
    period_name VARCHAR(100) NOT NULL,
    period_code VARCHAR(20) NOT NULL,
    
    -- Period dates
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Period classification
    period_type VARCHAR(20) NOT NULL DEFAULT 'regular'
        CHECK (period_type IN ('regular', 'adjustment', 'closing', 'opening')),
    fiscal_year INTEGER NOT NULL,
    quarter INTEGER CHECK (quarter BETWEEN 1 AND 4),
    
    -- Status and control
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_closed BOOLEAN NOT NULL DEFAULT false,
    
    -- Closing information
    closed_by UUID REFERENCES users(id),
    closed_at TIMESTAMPTZ,
    closing_balance DECIMAL(19,4),
    
    -- Lock controls
    lock_transactions BOOLEAN NOT NULL DEFAULT false,
    lock_reason TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_accounting_periods_tenant_code UNIQUE(tenant_id, entity_id, period_code),
    CONSTRAINT ck_finance_accounting_periods_dates CHECK (start_date < end_date)
);

-- Budget Management
CREATE TABLE finance_budgets (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Budget identification
    budget_name VARCHAR(255) NOT NULL,
    budget_code VARCHAR(20) NOT NULL,
    description TEXT,
    
    -- Budget period
    fiscal_year INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Budget scope
    budget_type VARCHAR(20) NOT NULL DEFAULT 'annual'
        CHECK (budget_type IN ('annual', 'quarterly', 'monthly', 'project', 'department')),
    
    -- Budget currency
    currency currency_code NOT NULL,
    
    -- Budget totals
    total_budget_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    total_allocated_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    total_actual_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    
    -- Status and workflow
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'submitted', 'approved', 'active', 'closed', 'cancelled')),
    
    -- Approval workflow
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    
    -- Control settings
    allow_overspend BOOLEAN NOT NULL DEFAULT false,
    overspend_limit_percentage DECIMAL(5,2) DEFAULT 0.00,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_budgets_tenant_code UNIQUE(tenant_id, entity_id, budget_code),
    CONSTRAINT ck_finance_budgets_dates CHECK (start_date < end_date),
    CONSTRAINT ck_finance_budgets_amounts CHECK (total_budget_amount >= 0)
);

-- Budget Line Items
CREATE TABLE finance_budget_items (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    budget_id UUID NOT NULL REFERENCES finance_budgets(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES finance_accounts(id),
    
    -- Optional dimensions
    cost_center_id UUID REFERENCES finance_cost_centers(id),
    project_id UUID REFERENCES finance_projects(id),
    
    -- Budget amounts
    budget_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    allocated_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    actual_amount DECIMAL(19,4) NOT NULL DEFAULT 0.00,
    
    -- Period distribution
    q1_amount DECIMAL(19,4) DEFAULT 0.00,
    q2_amount DECIMAL(19,4) DEFAULT 0.00,
    q3_amount DECIMAL(19,4) DEFAULT 0.00,
    q4_amount DECIMAL(19,4) DEFAULT 0.00,
    
    -- Variance tracking
    variance_amount DECIMAL(19,4) DEFAULT 0.00,
    variance_percentage DECIMAL(5,2) DEFAULT 0.00,
    
    -- Notes and justification
    notes TEXT,
    justification TEXT,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_finance_budget_items_unique UNIQUE(tenant_id, budget_id, account_id, cost_center_id, project_id),
    CONSTRAINT ck_finance_budget_items_amounts CHECK (budget_amount >= 0)
);

-- Indexes and RLS
CREATE INDEX idx_finance_period_closing_vouchers_tenant ON finance_period_closing_vouchers(tenant_id);
CREATE INDEX idx_finance_accounting_periods_tenant ON finance_accounting_periods(tenant_id);
CREATE INDEX idx_finance_budgets_tenant ON finance_budgets(tenant_id);
CREATE INDEX idx_finance_budget_items_tenant ON finance_budget_items(tenant_id);

ALTER TABLE finance_period_closing_vouchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_accounting_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_budget_items ENABLE ROW LEVEL SECURITY;

-- Comments
COMMENT ON TABLE finance_period_closing_vouchers IS 'Period-end closing vouchers for transferring income/expense balances';
COMMENT ON TABLE finance_accounting_periods IS 'Accounting period definitions with status and control mechanisms';
COMMENT ON TABLE finance_budgets IS 'Budget master records for financial planning and control';
COMMENT ON TABLE finance_budget_items IS 'Individual budget line items with dimensional analysis';

-- =====================================================================
-- MIGRATION 016: PRICING AND LOYALTY MANAGEMENT
-- =====================================================================
-- File: 016_create_pricing_and_loyalty.up.sql
-- Description: Create pricing rules, promotional schemes, and loyalty programs
-- Dependencies: finance_parties
-- =====================================================================

-- Pricing Rules
CREATE TABLE finance_pricing_rules (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Rule identification
    rule_name VARCHAR(255) NOT NULL,
    rule_code VARCHAR(50) NOT NULL,
    description TEXT,
    
    -- Rule application
    applies_to VARCHAR(20) NOT NULL DEFAULT 'item'
        CHECK (applies_to IN ('item', 'item_group', 'brand', 'customer', 'customer_group', 'all')),
    
    -- Conditions
    min_qty DECIMAL(15,4) DEFAULT 0,
    max_qty DECIMAL(15,4),
    min_amount DECIMAL(19,4) DEFAULT 0,
    max_amount DECIMAL(19,4),
    
    -- Customer conditions
    customer_id UUID REFERENCES finance_parties(id),
    customer_group VARCHAR(100),
    
    -- Geographic conditions
    territory VARCHAR(100),
    
    -- Pricing configuration
    price_or_product_discount VARCHAR(10) NOT NULL DEFAULT 'price'
        CHECK (price_or_product_discount IN ('price', 'discount')),
    
    -- Price/discount values
    rate DECIMAL(19,4),
    discount_percentage DECIMAL(5,2),
    discount_amount DECIMAL(19,4),
    
    -- Currency
    currency currency_code,
    
    -- Validity
    valid_from DATE,
    valid_upto DATE,
    
    -- Priority and mixing
    priority INTEGER DEFAULT 1,
    disable BOOLEAN NOT NULL DEFAULT false,
    
    -- Conditions for application
    for_price_list VARCHAR(100),
    warehouse VARCHAR(100),
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_pricing_rules_tenant_code UNIQUE(tenant_id, rule_code),
    CONSTRAINT ck_finance_pricing_rules_qty CHECK (max_qty IS NULL OR min_qty <= max_qty),
    CONSTRAINT ck_finance_pricing_rules_amount CHECK (max_amount IS NULL OR min_amount <= max_amount),
    CONSTRAINT ck_finance_pricing_rules_dates CHECK (valid_from IS NULL OR valid_upto IS NULL OR valid_from <= valid_upto)
);

-- Promotional Schemes
CREATE TABLE finance_promotional_schemes (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Scheme identification
    scheme_name VARCHAR(255) NOT NULL,
    scheme_code VARCHAR(50) NOT NULL,
    description TEXT,
    
    -- Scheme type
    scheme_type VARCHAR(20) NOT NULL DEFAULT 'discount'
        CHECK (scheme_type IN ('discount', 'bogo', 'bundle', 'cashback', 'points')),
    
    -- Discount configuration
    discount_type VARCHAR(15) NOT NULL DEFAULT 'percentage'
        CHECK (discount_type IN ('percentage', 'amount', 'fixed_price')),
    discount_value DECIMAL(19,4) NOT NULL,
    max_discount_amount DECIMAL(19,4),
    
    -- Conditions
    min_purchase_amount DECIMAL(19,4) DEFAULT 0,
    min_qty DECIMAL(15,4) DEFAULT 0,
    
    -- Customer eligibility
    applicable_for VARCHAR(20) DEFAULT 'all'
        CHECK (applicable_for IN ('all', 'new_customers', 'existing_customers', 'specific_group')),
    customer_group VARCHAR(100),
    
    -- Usage limits
    usage_limit_per_customer INTEGER,
    total_usage_limit INTEGER,
    current_usage_count INTEGER DEFAULT 0,
    
    -- Validity
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    auto_apply BOOLEAN NOT NULL DEFAULT false,
    
    -- Terms and conditions
    terms_and_conditions TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_promotional_schemes_tenant_code UNIQUE(tenant_id, scheme_code),
    CONSTRAINT ck_finance_promotional_schemes_dates CHECK (start_date <= end_date),
    CONSTRAINT ck_finance_promotional_schemes_limits CHECK (
        usage_limit_per_customer IS NULL OR usage_limit_per_customer > 0
    )
);

-- Loyalty Programs
CREATE TABLE finance_loyalty_programs (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Program identification
    program_name VARCHAR(255) NOT NULL,
    program_code VARCHAR(50) NOT NULL,
    description TEXT,
    
    -- Program configuration
    loyalty_point_name VARCHAR(50) NOT NULL DEFAULT 'Points',
    conversion_factor DECIMAL(10,4) NOT NULL DEFAULT 1.0, -- Points per currency unit
    
    -- Earning rules
    earn_points_on_purchase BOOLEAN NOT NULL DEFAULT true,
    minimum_purchase_amount DECIMAL(19,4) DEFAULT 0,
    earning_rate DECIMAL(5,4) NOT NULL DEFAULT 1.0, -- Points per unit spent
    
    -- Redemption rules
    allow_redemption BOOLEAN NOT NULL DEFAULT true,
    minimum_redemption_points INTEGER DEFAULT 100,
    redemption_value DECIMAL(10,4) NOT NULL DEFAULT 0.01, -- Currency value per point
    max_redemption_percentage DECIMAL(5,2) DEFAULT 100.00,
    
    -- Membership tiers
    has_tiers BOOLEAN NOT NULL DEFAULT false,
    
    -- Expiry configuration
    points_expire BOOLEAN NOT NULL DEFAULT false,
    expiry_days INTEGER,
    
    -- Validity
    start_date DATE NOT NULL,
    end_date DATE,
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    auto_enroll BOOLEAN NOT NULL DEFAULT false,
    
    -- Terms and conditions
    terms_and_conditions TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_loyalty_programs_tenant_code UNIQUE(tenant_id, program_code),
    CONSTRAINT ck_finance_loyalty_programs_dates CHECK (end_date IS NULL OR start_date <= end_date),
    CONSTRAINT ck_finance_loyalty_programs_conversion CHECK (conversion_factor > 0),
    CONSTRAINT ck_finance_loyalty_programs_redemption CHECK (redemption_value > 0)
);

-- Customer Loyalty Points
CREATE TABLE finance_customer_loyalty_points (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationships
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES finance_parties(id) ON DELETE CASCADE,
    loyalty_program_id UUID NOT NULL REFERENCES finance_loyalty_programs(id) ON DELETE CASCADE,
    
    -- Points tracking
    total_points_earned INTEGER NOT NULL DEFAULT 0,
    total_points_redeemed INTEGER NOT NULL DEFAULT 0,
    current_balance INTEGER NOT NULL DEFAULT 0,
    expired_points INTEGER NOT NULL DEFAULT 0,
    
    -- Membership details
    enrollment_date DATE NOT NULL,
    membership_number VARCHAR(50),
    tier_level VARCHAR(50),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_activity_date DATE,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_finance_customer_loyalty_unique UNIQUE(tenant_id, customer_id, loyalty_program_id),
    CONSTRAINT ck_finance_customer_loyalty_balance CHECK (current_balance >= 0)
);

-- Indexes and RLS
CREATE INDEX idx_finance_pricing_rules_tenant ON finance_pricing_rules(tenant_id);
CREATE INDEX idx_finance_promotional_schemes_tenant ON finance_promotional_schemes(tenant_id);
CREATE INDEX idx_finance_loyalty_programs_tenant ON finance_loyalty_programs(tenant_id);
CREATE INDEX idx_finance_customer_loyalty_points_tenant ON finance_customer_loyalty_points(tenant_id);

ALTER TABLE finance_pricing_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_promotional_schemes ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_loyalty_programs ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_customer_loyalty_points ENABLE ROW LEVEL SECURITY;

-- Comments
COMMENT ON TABLE finance_pricing_rules IS 'Dynamic pricing rules based on quantity, amount, customer, and other conditions';
COMMENT ON TABLE finance_promotional_schemes IS 'Sales promotion schemes with discounts, offers, and usage tracking';
COMMENT ON TABLE finance_loyalty_programs IS 'Customer loyalty program configurations with points and redemption rules';
COMMENT ON TABLE finance_customer_loyalty_points IS 'Customer loyalty points balance and transaction tracking';

-- =====================================================================
-- MIGRATION 017: ADVANCED FEATURES
-- =====================================================================
-- File: 017_create_advanced_features.up.sql
-- Description: Create subscriptions, shareholding, guarantees, and cheque templates
-- Dependencies: finance_parties, finance_accounts
-- =====================================================================

-- Subscriptions (Recurring Billing)
CREATE TABLE finance_subscriptions (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Subscription identification
    subscription_number VARCHAR(50) NOT NULL,
    subscription_name VARCHAR(255) NOT NULL,
    
    -- Customer information
    customer_id UUID NOT NULL REFERENCES finance_parties(id),
    
    -- Subscription configuration
    billing_frequency VARCHAR(20) NOT NULL DEFAULT 'monthly'
        CHECK (billing_frequency IN ('daily', 'weekly', 'monthly', 'quarterly', 'yearly')),
    billing_amount DECIMAL(19,4) NOT NULL,
    currency currency_code NOT NULL,
    
    -- Billing dates
    start_date DATE NOT NULL,
    end_date DATE,
    next_billing_date DATE NOT NULL,
    last_billing_date DATE,
    
    -- Trial period
    trial_period_days INTEGER DEFAULT 0,
    trial_end_date DATE,
    
    -- Status and control
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('draft', 'active', 'paused', 'cancelled', 'expired')),
    auto_renew BOOLEAN NOT NULL DEFAULT true,
    
    -- Payment method
    payment_method VARCHAR(30)
        CHECK (payment_method IN ('auto_charge', 'invoice', 'bank_transfer')),
    
    -- Accounting
    income_account_id UUID REFERENCES finance_accounts(id),
    cost_center_id UUID REFERENCES finance_cost_centers(id),
    
    -- Cancellation
    cancellation_date DATE,
    cancellation_reason TEXT,
    
    -- Additional information
    description TEXT,
    terms_and_conditions TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_subscriptions_tenant_number UNIQUE(tenant_id, subscription_number),
    CONSTRAINT ck_finance_subscriptions_dates CHECK (end_date IS NULL OR start_date <= end_date),
    CONSTRAINT ck_finance_subscriptions_amount CHECK (billing_amount > 0)
);

-- Shareholder Management
CREATE TABLE finance_shareholders (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Shareholder identification
    shareholder_id VARCHAR(50) NOT NULL,
    shareholder_name VARCHAR(255) NOT NULL,
    
    -- Contact information
    email VARCHAR(254),
    phone VARCHAR(30),
    address JSONB DEFAULT '{}'::jsonb,
    
    -- Shareholding details
    share_type VARCHAR(20) NOT NULL DEFAULT 'equity'
        CHECK (share_type IN ('equity', 'preference', 'convertible', 'warrant')),
    shares_held INTEGER NOT NULL DEFAULT 0,
    nominal_value_per_share DECIMAL(19,4) NOT NULL,
    total_nominal_value DECIMAL(19,4) NOT NULL DEFAULT 0,
    
    -- Ownership percentage
    ownership_percentage DECIMAL(5,4) DEFAULT 0.00,
    
    -- Investment details
    initial_investment_date DATE,
    initial_investment_amount DECIMAL(19,4),
    current_market_value DECIMAL(19,4),
    
    -- Rights and restrictions
    voting_rights BOOLEAN NOT NULL DEFAULT true,
    dividend_rights BOOLEAN NOT NULL DEFAULT true,
    transfer_restrictions TEXT,
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_founder BOOLEAN NOT NULL DEFAULT false,
    
    -- Tax information
    tax_id VARCHAR(50),
    tax_resident_country VARCHAR(3),
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_shareholders_tenant_id UNIQUE(tenant_id, shareholder_id),
    CONSTRAINT ck_finance_shareholders_shares CHECK (shares_held >= 0),
    CONSTRAINT ck_finance_shareholders_values CHECK (
        nominal_value_per_share > 0 AND 
        total_nominal_value = shares_held * nominal_value_per_share
    )
);

-- Cheque Print Templates
CREATE TABLE finance_cheque_print_templates (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Template identification
    template_name VARCHAR(255) NOT NULL,
    template_code VARCHAR(50) NOT NULL,
    description TEXT,
    
    -- Bank account association
    bank_account_id UUID REFERENCES finance_bank_accounts(id),
    
    -- Cheque dimensions (in mm)
    cheque_width DECIMAL(5,2) NOT NULL DEFAULT 152.4, -- 6 inches
    cheque_height DECIMAL(5,2) NOT NULL DEFAULT 76.2,  -- 3 inches
    
    -- Field positions (x, y coordinates in mm from top-left)
    date_x DECIMAL(5,2) NOT NULL DEFAULT 120,
    date_y DECIMAL(5,2) NOT NULL DEFAULT 15,
    payee_x DECIMAL(5,2) NOT NULL DEFAULT 25,
    payee_y DECIMAL(5,2) NOT NULL DEFAULT 25,
    amount_in_words_x DECIMAL(5,2) NOT NULL DEFAULT 25,
    amount_in_words_y DECIMAL(5,2) NOT NULL DEFAULT 35,
    amount_in_figures_x DECIMAL(5,2) NOT NULL DEFAULT 120,
    amount_in_figures_y DECIMAL(5,2) NOT NULL DEFAULT 35,
    memo_x DECIMAL(5,2) NOT NULL DEFAULT 25,
    memo_y DECIMAL(5,2) NOT NULL DEFAULT 50,
    
    -- Font settings
    font_name VARCHAR(50) DEFAULT 'Arial',
    font_size INTEGER DEFAULT 12,
    
    -- Print settings
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    
    -- Security features
    micr_line VARCHAR(100), -- Magnetic Ink Character Recognition
    security_features TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_cheque_templates_tenant_code UNIQUE(tenant_id, template_code)
);

-- Bank Guarantees
CREATE TABLE finance_bank_guarantees (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Multi-tenancy and entity hierarchy
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Guarantee identification
    guarantee_number VARCHAR(50) NOT NULL,
    guarantee_type VARCHAR(30) NOT NULL DEFAULT 'performance'
        CHECK (guarantee_type IN ('performance', 'advance_payment', 'bid_bond', 'warranty', 'financial')),
    
    -- Parties involved
    beneficiary_id UUID NOT NULL REFERENCES finance_parties(id), -- Who receives the guarantee
    applicant_id UUID REFERENCES finance_parties(id),           -- Who requests the guarantee
    bank_id UUID NOT NULL REFERENCES finance_parties(id),       -- Issuing bank
    
    -- Guarantee details
    guarantee_amount DECIMAL(19,4) NOT NULL,
    currency currency_code NOT NULL,
    
    -- Important dates
    issue_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    claim_period_days INTEGER DEFAULT 30,
    
    -- Underlying contract/project
    contract_number VARCHAR(100),
    contract_value DECIMAL(19,4),
    project_id UUID REFERENCES finance_projects(id),
    
    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('draft', 'issued', 'active', 'expired', 'cancelled', 'claimed', 'returned')),
    
    -- Financial impact
    margin_amount DECIMAL(19,4), -- Security deposit with bank
    commission_rate DECIMAL(5,4),
    commission_amount DECIMAL(19,4),
    
    -- Bank account for charges
    bank_account_id UUID REFERENCES finance_bank_accounts(id),
    
    -- Terms and conditions
    terms_and_conditions TEXT,
    special_conditions TEXT,
    
    -- Claim information
    claim_date DATE,
    claimed_amount DECIMAL(19,4),
    claim_reason TEXT,
    
    -- Audit and versioning
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    
    CONSTRAINT uk_finance_bank_guarantees_tenant_number UNIQUE(tenant_id, guarantee_number),
    CONSTRAINT ck_finance_bank_guarantees_dates CHECK (issue_date <= expiry_date),
    CONSTRAINT ck_finance_bank_guarantees_amounts CHECK (guarantee_amount > 0)
);

-- Indexes and RLS
CREATE INDEX idx_finance_subscriptions_tenant ON finance_subscriptions(tenant_id);
CREATE INDEX idx_finance_shareholders_tenant ON finance_shareholders(tenant_id);
CREATE INDEX idx_finance_cheque_print_templates_tenant ON finance_cheque_print_templates(tenant_id);
CREATE INDEX idx_finance_bank_guarantees_tenant ON finance_bank_guarantees(tenant_id);

ALTER TABLE finance_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_shareholders ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_cheque_print_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_guarantees ENABLE ROW LEVEL SECURITY;

-- Comments
COMMENT ON TABLE finance_subscriptions IS 'Recurring billing subscriptions with flexible frequency and payment options';
COMMENT ON TABLE finance_shareholders IS 'Shareholder and equity management with ownership tracking and rights management';
COMMENT ON TABLE finance_cheque_print_templates IS 'Cheque printing templates with precise positioning and formatting';
COMMENT ON TABLE finance_bank_guarantees IS 'Bank guarantee management for performance bonds, bid bonds, and other guarantees';

-- =====================================================================
-- MIGRATION 018: ADDITIONAL BUSINESS LOGIC AND VIEWS
-- =====================================================================
-- File: 018_create_additional_functions_views.up.sql
-- Description: Create additional business functions, views, and procedures
-- Dependencies: All finance tables
-- =====================================================================

-- Function to calculate account hierarchy balance
CREATE OR REPLACE FUNCTION get_account_hierarchy_balance(
    p_tenant_id UUID,
    p_account_id UUID,
    p_as_of_date DATE DEFAULT CURRENT_DATE
) RETURNS DECIMAL(19,4) AS $$
DECLARE
    account_rec RECORD;
    child_balance DECIMAL(19,4) := 0;
    total_balance DECIMAL(19,4) := 0;
BEGIN
    -- Get account details
    SELECT * INTO account_rec 
    FROM finance_accounts 
    WHERE id = p_account_id AND tenant_id = p_tenant_id;
    
    IF NOT FOUND THEN
        RETURN 0;
    END IF;
    
    -- If it's a group account, sum all children
    IF account_rec.is_group THEN
        SELECT COALESCE(SUM(get_account_hierarchy_balance(p_tenant_id, id, p_as_of_date)), 0)
        INTO total_balance
        FROM finance_accounts
        WHERE tenant_id = p_tenant_id 
        AND parent_account_id = p_account_id
        AND is_active = true;
    ELSE
        -- Calculate balance from transaction entries
        SELECT 
            COALESCE(SUM(base_debit_amount), 0) - COALESCE(SUM(base_credit_amount), 0)
        INTO total_balance
        FROM finance_transaction_entries fte
        JOIN finance_transactions ft ON fte.transaction_id = ft.id
        WHERE fte.tenant_id = p_tenant_id
        AND fte.account_id = p_account_id
        AND ft.status = 'posted'
        AND ft.posting_date <= p_as_of_date;
    END IF;
    
    RETURN total_balance;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION get_account_hierarchy_balance(UUID, UUID, DATE) IS 
'Calculate balance for an account including all child accounts if it is a group account';

-- Function to get exchange rate
CREATE OR REPLACE FUNCTION get_exchange_rate(
    p_tenant_id UUID,
    p_entity_id UUID,
    p_from_currency currency_code,
    p_to_currency currency_code,
    p_rate_date DATE DEFAULT CURRENT_DATE
) RETURNS DECIMAL(10,6) AS $$
DECLARE
    exchange_rate DECIMAL(10,6);
BEGIN
    -- If same currency, return 1
    IF p_from_currency = p_to_currency THEN
        RETURN 1.0;
    END IF;
    
    -- Get the most recent rate
    SELECT rate INTO exchange_rate
    FROM finance_exchange_rates
    WHERE tenant_id = p_tenant_id
    AND entity_id = p_entity_id
    AND from_currency = p_from_currency
    AND to_currency = p_to_currency
    AND effective_date <= p_rate_date
    AND is_active = true
    ORDER BY effective_date DESC
    LIMIT 1;
    
    -- If not found, try reverse rate
    IF exchange_rate IS NULL THEN
        SELECT (1.0 / rate) INTO exchange_rate
        FROM finance_exchange_rates
        WHERE tenant_id = p_tenant_id
        AND entity_id = p_entity_id
        AND from_currency = p_to_currency
        AND to_currency = p_from_currency
        AND effective_date <= p_rate_date
        AND is_active = true
        ORDER BY effective_date DESC
        LIMIT 1;
    END IF;
    
    -- Default to 1 if no rate found
    RETURN COALESCE(exchange_rate, 1.0);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION get_exchange_rate(UUID, UUID, currency_code, currency_code, DATE) IS 
'Get exchange rate between two currencies for a specific date';

-- View for Trial Balance
CREATE VIEW finance_trial_balance AS
SELECT 
    fa.tenant_id,
    fa.entity_id,
    fa.id as account_id,
    fa.code as account_code,
    fa.name as account_name,
    fa.account_type,
    fa.normal_balance,
    COALESCE(SUM(fte.base_debit_amount), 0) as total_debits,
    COALESCE(SUM(fte.base_credit_amount), 0) as total_credits,
    CASE fa.normal_balance
        WHEN 'debit' THEN COALESCE(SUM(fte.base_debit_amount), 0) - COALESCE(SUM(fte.base_credit_amount), 0)
        WHEN 'credit' THEN COALESCE(SUM(fte.base_credit_amount), 0) - COALESCE(SUM(fte.base_debit_amount), 0)
    END as balance
FROM finance_accounts fa
LEFT JOIN finance_transaction_entries fte ON fa.id = fte.account_id AND fa.tenant_id = fte.tenant_id
LEFT JOIN finance_transactions ft ON fte.transaction_id = ft.id
WHERE fa.is_active = true
AND fa.is_group = false
AND (ft.status = 'posted' OR ft.status IS NULL)
GROUP BY fa.tenant_id, fa.entity_id, fa.id, fa.code, fa.name, fa.account_type, fa.normal_balance;

COMMENT ON VIEW finance_trial_balance IS 
'Trial balance view showing account balances for all active leaf accounts';

-- View for Customer Outstanding
CREATE VIEW finance_customer_outstanding AS
SELECT 
    fp.tenant_id,
    fp.entity_id,
    fp.id as customer_id,
    fp.code as customer_code,
    fp.name as customer_name,
    COUNT(ft.id) as total_invoices,
    COALESCE(SUM(ft.total_amount), 0) as total_invoiced,
    COALESCE(SUM(CASE WHEN ft.due_date < CURRENT_DATE THEN ft.total_amount ELSE 0 END), 0) as overdue_amount,
    COALESCE(SUM(CASE WHEN ft.due_date >= CURRENT_DATE THEN ft.total_amount ELSE 0 END), 0) as current_amount
FROM finance_parties fp
LEFT JOIN finance_transactions ft ON fp.id = ft.party_id 
    AND ft.party_type = 'customer' 
    AND ft.type = 'sales_invoice'
    AND ft.status = 'posted'
WHERE fp.party_type = 'customer'
AND fp.is_active = true
GROUP BY fp.tenant_id, fp.entity_id, fp.id, fp.code, fp.name;

COMMENT ON VIEW finance_customer_outstanding IS 
'Customer outstanding balances with aging analysis';

-- Sample initial data script placeholder
-- This would contain INSERT statements for default data
