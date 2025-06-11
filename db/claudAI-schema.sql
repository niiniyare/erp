

-- =====================================================
-- CORE ENTITY DEFINITION
-- =====================================================

-- Main entity/company table - the root of the organizational hierarchy
-- CREATE TABLE IF NOT EXISTS entity (
--   slug VARCHAR(50) NOT NULL UNIQUE,                    -- URL-friendly identifier
--   created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
--   updated TIMESTAMP WITHOUT TIME ZONE NULL,
--   -- Address information
--   address_1 VARCHAR(70) NOT NULL,
--   address_2 VARCHAR(70) NULL,
--   city VARCHAR(70) NULL,
--   state VARCHAR(70) NULL,
--   zip_code VARCHAR(20) NULL,
--   country VARCHAR(70) NULL,
--   -- Contact information
--   email VARCHAR(254) NULL,
--   website VARCHAR(200) NULL,
--   phone VARCHAR(30) NULL,
--   -- Tree structure fields for hierarchical organizations
--   path VARCHAR(255) NOT NULL UNIQUE,                   -- Materialized path for tree queries
--   depth INTEGER NOT NULL CHECK (depth >= 0),        -- Depth in organization tree
--   numchild INTEGER NOT NULL CHECK (numchild >= 0),  -- Number of direct children
--   uuid CHAR(32) NOT NULL PRIMARY KEY,
--   name VARCHAR(150) NOT NULL,                          -- Entity/company name
--   hidden BOOLEAN NOT NULL,                             -- Whether entity is hidden
--   accrual_method BOOLEAN NOT NULL,                     -- True=Accrual, False=Cash accounting
--   fy_start_month INTEGER NOT NULL,                     -- Fiscal year start month (1-12)
--   picture VARCHAR(100) NULL,                           -- Logo/picture file path
--   admin_id INTEGER NOT NULL REFERENCES auth_user (id) DEFERRABLE INITIALLY DEFERRED,
--   default_coa_id CHAR(32) NULL UNIQUE REFERENCES chartofaccount (uuid) DEFERRABLE INITIALLY DEFERRED,
--   last_closing_date DATE NULL,                         -- Last period closing date
--   meta TEXT NULL                                       -- Additional metadata
-- );
-- COMMENT ON TABLE entity IS 'Root entity/company table with hierarchical structure and accounting preferences';
-- COMMENT ON COLUMN entity.accrual_method IS 'Accounting method: True=Accrual, False=Cash basis';
-- COMMENT ON COLUMN entity.fy_start_month IS 'Fiscal year start month (1=Jan, 2=Feb, etc.)';

-- =====================================================
-- ACCOUNTING AND FINANCIAL TABLES
-- =====================================================

-- Journal entries - the foundation of double-entry bookkeeping
-- CREATE TABLE IF NOT EXISTS journalentry (
--   created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
--   updated TIMESTAMP WITHOUT TIME ZONE NULL,
--   tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--   -- entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE 
--   posted_by INT REFERENCES users(id),   
--   created_by INT REFERENCES users(id),
--   uuid CHAR(32) NOT NULL PRIMARY KEY,
--   je_number VARCHAR(25) NOT NULL,                      -- Journal entry number
--   timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,     -- Transaction timestamp
--   description VARCHAR(70) NULL,                        -- JE description
--   activity VARCHAR(20) NULL,                           -- Business activity category
--   origin VARCHAR(30) NULL,                             -- Source of the journal entry
--   posted BOOLEAN NOT NULL,                             -- Whether JE is posted to the books
--   locked BOOLEAN NOT NULL,                             -- Whether JE is locked from changes
--   entity_unit_id CHAR(32) NULL REFERENCES entities(id) (uuid) DEFERRABLE INITIALLY DEFERRED,
--   ledger_id CHAR(32) NOT NULL REFERENCES ledger (uuid) DEFERRABLE INITIALLY DEFERRED,
--   is_closing_entry BOOLEAN NOT NULL                    -- Whether this is a period-end closing entry
-- );
-- COMMENT ON TABLE journalentry IS 'Journal entries for double-entry bookkeeping with audit trail';
-- COMMENT ON COLUMN journalentry.posted IS 'Whether the journal entry affects account balances';
-- COMMENT ON COLUMN journalentry.origin IS 'Source system: invoice, bill, manual, etc.';

-- -- Ledgers - collections of journal entries for organizational purposes
-- CREATE TABLE IF NOT EXISTS ledger (
--   created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
--   updated TIMESTAMP WITHOUT TIME ZONE NULL,
--   tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--   entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE 
--   posted_by INT REFERENCES users(id),   
--   created_by INT REFERENCES users(id),
--   id SERIAL PRIMARY KEY,
--   name VARCHAR(150) NULL,                              -- Ledger name/description
--   posted BOOLEAN NOT NULL,                             -- Whether ledger is posted
--   locked BOOLEAN NOT NULL,                             -- Whether ledger is locked
--   hidden BOOLEAN NOT NULL,                             -- Whether ledger is hidden from UI
--   entity_id INT NOT NULL REFERENCES entities(id) DEFERRABLE INITIALLY DEFERRED,
--   additional_info TEXT NULL CHECK (
--     (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
--   ),
--   ledger_xid VARCHAR(150) NULL                         -- External ledger identifier
-- );
-- COMMENT ON TABLE ledger IS 'Ledgers group related journal entries (e.g., monthly ledgers, project ledgers)';

-- Period closing entries and transactions
CREATE TABLE IF NOT EXISTS closingentrytransaction (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE 
  posted_by INT REFERENCES users(id),   
  created_by INT REFERENCES users(id),
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  activity VARCHAR(20) NULL,                           -- Business activity
  tx_type VARCHAR(10) NOT NULL,                        -- Transaction type (debit/credit)
  balance DECIMAL NOT NULL,                            -- Account balance being closed
  account__id CHAR(32) NOT NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  closing_entry__id CHAR(32) NOT NULL REFERENCES closingentry (uuid) DEFERRABLE INITIALLY DEFERRED,
  unit__id CHAR(32) NULL REFERENCES entityunit (uuid) DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT unique_closing_entry UNIQUE (
    closing_entry__id, account__id,
    unit__id, activity
  )
);
COMMENT ON TABLE closingentrytransaction IS 'Individual account balances captured during period closing';

-- Period closing entries
CREATE TABLE IF NOT EXISTS closingentry (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE 
  posted_by INT REFERENCES users(id),   
  created_by INT REFERENCES users(id),
  markdown_notes TEXT NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  closing_date DATE NOT NULL,                          -- Period end date
  posted BOOLEAN NOT NULL,                             -- Whether closing is posted
  entity__id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  ledger__id CHAR(32) NOT NULL UNIQUE REFERENCES ledger (uuid) DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT unique_entity_closing_date UNIQUE (
    entity__id, closing_date
  )
);
COMMENT ON TABLE closingentry IS 'Period-end closing entries for financial reporting periods';



-- =====================================================
-- CORE ENTITY MANAGEMENT TABLES
-- =====================================================

-- Main entity table - represents companies/organizations in the system
-- This is the root table that most other tables reference
CREATE TABLE IF NOT EXISTS estimate (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,        -- Record creation timestamp
  updated TIMESTAMP WITHOUT TIME ZONE NULL,            -- Last update timestamp
  markdown_notes TEXT NULL,                            -- Rich text notes in markdown format
  uuid CHAR(32) NOT NULL PRIMARY KEY,                  -- Unique identifier (32-char hex UUID)
  estimate_number VARCHAR(20) NOT NULL,                -- Human-readable estimate number
  terms VARCHAR(10) NOT NULL,                          -- Payment terms (e.g., NET30, COD)
  title VARCHAR(250) NOT NULL,                         -- Estimate title/description
  status VARCHAR(10) NOT NULL,                         -- Current status (draft, review, approved, etc.)
  date_draft DATE NULL,                                -- Date estimate was drafted
  date_in_review DATE NULL,                            -- Date estimate entered review
  date_approved DATE NULL,                             -- Date estimate was approved
  date_completed DATE NULL,                            -- Date work was completed
  date_canceled DATE NULL,                             -- Date estimate was canceled
  date_void DATE NULL,                                 -- Date estimate was voided
  revenue_estimate DECIMAL NOT NULL,                   -- Total estimated revenue
  labor_estimate DECIMAL NOT NULL,                     -- Estimated labor costs
  material_estimate DECIMAL NOT NULL,                  -- Estimated material costs
  equipment_estimate DECIMAL NOT NULL,                 -- Estimated equipment costs
  other_estimate DECIMAL NOT NULL,                     -- Other estimated costs
  customer_id CHAR(32) NOT NULL REFERENCES customer (uuid) DEFERRABLE INITIALLY DEFERRED,
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE estimate IS 'Project estimates/quotes with cost breakdowns and workflow status tracking';
COMMENT ON COLUMN estimate.status IS 'Workflow status: draft, in_review, approved, completed, canceled, void';
COMMENT ON COLUMN estimate.terms IS 'Payment terms like NET30, NET15, COD, etc.';

-- Item transactions - line items for estimates, invoices, bills, and purchase orders
-- This is the central table that connects items to various business documents
CREATE TABLE IF NOT EXISTS itemtransaction (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  -- Invoice/Bill quantities and pricing
  quantity REAL NULL,                                  -- Actual quantity for invoice/bill
  unit_cost REAL NULL,                                 -- Actual unit cost for invoice/bill
  total_amount DECIMAL NULL,                           -- Total amount for invoice/bill line
  -- Purchase Order quantities and pricing
  po_quantity REAL NULL,                               -- PO quantity
  po_unit_cost REAL NULL,                              -- PO unit cost
  po_total_amount DECIMAL NULL,                        -- PO total amount
  po_item_status VARCHAR(15) NULL,                     -- PO item status (pending, received, etc.)
  -- Cost Estimate quantities and pricing
  ce_quantity REAL NULL,                               -- Estimate quantity
  ce_unit_cost_estimate REAL NULL,                     -- Estimated unit cost
  ce_cost_estimate DECIMAL NULL,                       -- Total estimated cost
  ce_unit_revenue_estimate REAL NULL,                  -- Estimated unit revenue
  ce_revenue_estimate DECIMAL NULL,                    -- Total estimated revenue
  item_notes VARCHAR(400) NULL,                        -- Notes specific to this line item
  -- Foreign key relationships to various business documents
  bill__id CHAR(32) NULL REFERENCES bill (uuid) DEFERRABLE INITIALLY DEFERRED,
  ce__id CHAR(32) NULL REFERENCES estimate (uuid) DEFERRABLE INITIALLY DEFERRED,
  entity_unit_id CHAR(32) NULL REFERENCES entityunit (uuid) DEFERRABLE INITIALLY DEFERRED,
  invoice__id CHAR(32) NULL REFERENCES invoice (uuid) DEFERRABLE INITIALLY DEFERRED,
  item__id CHAR(32) NOT NULL REFERENCES item (uuid) DEFERRABLE INITIALLY DEFERRED,
  po__id CHAR(32) NULL REFERENCES purchaseorder (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE itemtransaction IS 'Line items for all business documents - estimates, invoices, bills, and purchase orders';
COMMENT ON COLUMN itemtransaction.po_item_status IS 'Purchase order item status: pending, partial, received, canceled';

-- Unit of measure definitions (e.g., each, hour, pound, gallon)
CREATE TABLE IF NOT EXISTS unitofmeasure (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,                           -- Full name (e.g., Each, Hour, Pound)
  unit_abbr VARCHAR(10) NOT NULL,                      -- Abbreviation (e.g., ea, hr, lb)
  is_active BOOLEAN NOT NULL,                          -- Whether this UOM is active for use
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE unitofmeasure IS 'Units of measure for items (each, hour, pound, gallon, etc.)';

-- =====================================================
-- PURCHASE ORDER MANAGEMENT
-- =====================================================

-- Purchase orders for procurement
CREATE TABLE IF NOT EXISTS purchaseorder (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  markdown_notes TEXT NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  po_number VARCHAR(20) NOT NULL,                      -- Purchase order number
  po_title VARCHAR(250) NOT NULL,                      -- PO title/description
  po_status VARCHAR(10) NOT NULL,                      -- PO status (draft, approved, fulfilled, etc.)
  po_amount DECIMAL NOT NULL,                          -- Total PO amount
  po_amount_received DECIMAL NOT NULL,                 -- Amount of goods/services received
  date_draft DATE NULL,
  date_in_review DATE NULL,
  date_approved DATE NULL,
  date_void DATE NULL,
  date_fulfilled DATE NULL,                            -- Date PO was completely fulfilled
  date_canceled DATE NULL,
  ce__id CHAR(32) NULL REFERENCES estimate (uuid) DEFERRABLE INITIALLY DEFERRED,
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE purchaseorder IS 'Purchase orders for procuring goods and services from vendors';
COMMENT ON COLUMN purchaseorder.po_amount_received IS 'Tracks partial fulfillment of purchase orders';

-- =====================================================
-- INVOICE MANAGEMENT
-- =====================================================

-- Customer invoices with revenue recognition
CREATE TABLE IF NOT EXISTS invoice (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  -- Financial amounts for revenue recognition
  amount_due DECIMAL NOT NULL,                         -- Total amount due from customer
  amount_paid DECIMAL NOT NULL,                        -- Amount already paid
  amount_receivable DECIMAL NOT NULL,                  -- Amount still receivable
  amount_unearned DECIMAL NOT NULL,                    -- Unearned revenue (prepayments)
  amount_earned DECIMAL NOT NULL,                      -- Earned revenue
  accrue BOOLEAN NOT NULL,                             -- Whether to accrue revenue
  progress DECIMAL NOT NULL,                           -- Project completion percentage (0-100)
  terms VARCHAR(10) NOT NULL,                          -- Payment terms
  date_due DATE NULL,                                  -- Payment due date
  markdown_notes TEXT NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  invoice_number VARCHAR(20) NOT NULL,                 -- Invoice number
  invoice_status VARCHAR(10) NOT NULL,                 -- Invoice status
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),                                                     -- JSON data for additional information
  -- Status dates
  date_draft DATE NULL,
  date_in_review DATE NULL,
  date_approved DATE NULL,
  date_paid DATE NULL,
  date_void DATE NULL,
  date_canceled DATE NULL,
  -- Account relationships for proper accounting
  cash_account_id CHAR(32) NOT NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  ce__id CHAR(32) NULL REFERENCES estimate (uuid) DEFERRABLE INITIALLY DEFERRED,
  customer_id CHAR(32) NOT NULL REFERENCES customer (uuid) DEFERRABLE INITIALLY DEFERRED,
  ledger_id CHAR(32) NOT NULL UNIQUE REFERENCES ledger (uuid) DEFERRABLE INITIALLY DEFERRED,
  prepaid_account_id CHAR(32) NOT NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  unearned_account_id CHAR(32) NOT NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE invoice IS 'Customer invoices with revenue recognition and payment tracking';
COMMENT ON COLUMN invoice.progress IS 'Project completion percentage for revenue recognition (0-100)';
COMMENT ON COLUMN invoice.accrue IS 'Whether to use accrual accounting for this invoice';

-- =====================================================
-- ENTITY STATE AND MANAGEMENT
-- =====================================================

-- Entity state tracking for sequence numbers and fiscal periods
CREATE TABLE IF NOT EXISTS entitystate (
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  fiscal_year SMALLINT NULL,                           -- Fiscal year for this state
  key VARCHAR(10) NOT NULL,                            -- State key (e.g., invoice, po)
  sequence BIGINT NOT NULL,                            -- Next sequence number for this key
  entity__id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  entity_unit_id CHAR(32) NULL REFERENCES entityunit (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE entitystate IS 'Manages sequence numbers for document numbering (invoices, POs, etc.)';
COMMENT ON COLUMN entitystate.key IS 'Document type: invoice, po, estimate, bill, etc.';

-- Entity management - user permissions for entities
CREATE TABLE IF NOT EXISTS entitymanagement (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  permission_level VARCHAR(10) NOT NULL,               -- Permission level (read, write, admin)
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  user_id INTEGER NOT NULL REFERENCES auth_user (id) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE entitymanagement IS 'User access permissions for entities/companies';
COMMENT ON COLUMN entitymanagement.permission_level IS 'Permission level: read, write, admin';

-- =====================================================
-- CUSTOMER AND VENDOR MANAGEMENT
-- =====================================================

-- Customer master data
CREATE TABLE IF NOT EXISTS customer (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  -- Address information
  address_1 VARCHAR(70) NOT NULL,
  address_2 VARCHAR(70) NULL,
  city VARCHAR(70) NULL,
  state VARCHAR(70) NULL,
  zip_code VARCHAR(20) NULL,
  country VARCHAR(70) NULL,
  -- Contact information
  email VARCHAR(254) NULL,
  website VARCHAR(200) NULL,
  phone VARCHAR(30) NULL,
  sales_tax_rate REAL NULL,                            -- Default sales tax rate for this customer
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  customer_name VARCHAR(100) NOT NULL,                 -- Customer company/person name
  customer_number VARCHAR(30) NOT NULL,                -- Customer reference number
  description TEXT NOT NULL,                           -- Customer description/notes
  active BOOLEAN NOT NULL,                             -- Whether customer is active
  hidden BOOLEAN NOT NULL,                             -- Whether to hide from lists
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),                                                     -- JSON for additional customer data
  entity__id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE customer IS 'Customer master data with contact info and billing preferences';
COMMENT ON COLUMN customer.sales_tax_rate IS 'Default sales tax rate as decimal (e.g., 0.0825 for 8.25%)';

-- =====================================================
-- BILL MANAGEMENT (VENDOR BILLS)
-- =====================================================

-- Vendor bills (accounts payable)
CREATE TABLE IF NOT EXISTS bill (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  -- Financial amounts
  amount_due DECIMAL NOT NULL,
  amount_paid DECIMAL NOT NULL,
  amount_receivable DECIMAL NOT NULL,
  amount_unearned DECIMAL NOT NULL,
  amount_earned DECIMAL NOT NULL,
  accrue BOOLEAN NOT NULL,
  progress DECIMAL NOT NULL,
  terms VARCHAR(10) NOT NULL,
  date_due DATE NULL,
  markdown_notes TEXT NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  bill_number VARCHAR(20) NOT NULL,
  bill_status VARCHAR(10) NOT NULL,
  xref VARCHAR(50) NULL,                               -- External reference (vendor's invoice number)
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),
  -- Status dates
  date_draft DATE NULL,
  date_in_review DATE NULL,
  date_approved DATE NULL,
  date_paid DATE NULL,
  date_void DATE NULL,
  date_canceled DATE NULL,
  -- Account relationships
  cash_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  ce__id CHAR(32) NULL REFERENCES estimate (uuid) DEFERRABLE INITIALLY DEFERRED,
  ledger_id CHAR(32) NOT NULL UNIQUE REFERENCES ledger (uuid) DEFERRABLE INITIALLY DEFERRED,
  prepaid_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  unearned_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  vendor_id CHAR(32) NOT NULL REFERENCES vendor (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE bill IS 'Vendor bills (accounts payable) with expense tracking';
COMMENT ON COLUMN bill.xref IS 'Vendor invoice/reference number for cross-referencing';

-- =====================================================
-- ITEM/PRODUCT MANAGEMENT
-- =====================================================

-- Item/product/service master data
CREATE TABLE IF NOT EXISTS item (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,                          -- Item name/description
  item_type VARCHAR(1) NULL,                           -- Item type code
  sku VARCHAR(50) NULL,                                -- Stock Keeping Unit
  upc VARCHAR(50) NULL,                                -- Universal Product Code
  item_id VARCHAR(50) NULL,                            -- Internal item ID
  item_number VARCHAR(30) NOT NULL,                    -- Item number for referencing
  is_active BOOLEAN NOT NULL,                          -- Whether item is active
  default_amount DECIMAL NOT NULL,                     -- Default price/cost
  for_inventory BOOLEAN NOT NULL,                      -- Whether this item is tracked in inventory
  is_product_or_service BOOLEAN NOT NULL,              -- True=Product, False=Service
  sold_as_unit BOOLEAN NOT NULL,                       -- Whether sold as whole units only
  inventory_received DECIMAL NULL,                     -- Quantity received in inventory
  inventory_received_value DECIMAL NULL,               -- Value of inventory received
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),
  -- Account relationships for proper accounting
  cogs_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,     -- Cost of Goods Sold
  earnings_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED, -- Revenue account
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  expense_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,  -- Expense account
  inventory_account_id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED, -- Inventory asset account
  uom_id CHAR(32) NOT NULL REFERENCES unitofmeasure (uuid) DEFERRABLE INITIALLY DEFERRED,
  item_role VARCHAR(10) NULL                           -- Item role/category
);
COMMENT ON TABLE item IS 'Master catalog of items, products, and services with accounting integration';
COMMENT ON COLUMN item.is_product_or_service IS 'True for physical products, False for services';
COMMENT ON COLUMN item.for_inventory IS 'Whether to track inventory quantities for this item';

-- =====================================================
-- BANK INTEGRATION AND IMPORT MANAGEMENT
-- =====================================================

-- Import jobs for bank statement processing
CREATE TABLE IF NOT EXISTS importjob (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  description VARCHAR(200) NOT NULL,                   -- Import job description
  completed BOOLEAN NOT NULL,                          -- Whether import is completed
  bank_account__id CHAR(32) NOT NULL REFERENCES bankaccount (uuid) DEFERRABLE INITIALLY DEFERRED,
  ledger__id CHAR(32) NULL UNIQUE REFERENCES ledger (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE importjob IS 'Bank statement import jobs for automated transaction processing';

-- Staged transactions from bank imports before they become actual transactions
CREATE TABLE IF NOT EXISTS stagedtransaction (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  fit_id VARCHAR(100) NOT NULL,                        -- Financial Institution Transaction ID
  date_posted DATE NOT NULL,                           -- Bank posting date
  amount DECIMAL NULL,                                 -- Transaction amount
  amount_split DECIMAL NULL,                           -- Split portion amount
  name VARCHAR(200) NULL,                              -- Payee/merchant name
  memo VARCHAR(200) NULL,                              -- Transaction memo/description
  account__id CHAR(32) NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  import_job_id CHAR(32) NOT NULL REFERENCES importjob (uuid) DEFERRABLE INITIALLY DEFERRED,
  transaction__id CHAR(32) NULL UNIQUE REFERENCES transaction (uuid) DEFERRABLE INITIALLY DEFERRED,
  unit__id CHAR(32) NULL REFERENCES entityunit (uuid) DEFERRABLE INITIALLY DEFERRED,
  activity VARCHAR(20) NULL,
  bundle_split BOOLEAN NOT NULL,                       -- Whether this is part of a split transaction
  parent_id CHAR(32) NULL REFERENCES stagedtransaction (uuid) DEFERRABLE INITIALLY DEFERRED
);
COMMENT ON TABLE stagedtransaction IS 'Imported bank transactions staged for review before posting';
COMMENT ON COLUMN stagedtransaction.fit_id IS 'Unique bank transaction ID from financial institution';

-- =====================================================
-- CHART OF ACCOUNTS STRUCTURE
-- =====================================================

-- Chart of accounts - templates for account structures
CREATE TABLE IF NOT EXISTS chartofaccount (
  slug VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(150) NULL,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  description TEXT NULL,
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  active BOOLEAN NOT NULL
);
COMMENT ON TABLE chartofaccount IS 'Chart of accounts templates (e.g., Standard, Manufacturing, Retail)';

-- Individual accounts within a chart of accounts
CREATE TABLE IF NOT EXISTS account (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  -- Tree structure for account hierarchy
  path VARCHAR(255) NOT NULL UNIQUE,
  depth INTEGER NOT NULL CHECK (depth >= 0),
  numchild INTEGER NOT NULL CHECK (numchild >= 0),
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  code VARCHAR(10) NOT NULL,                           -- Account code (e.g., 1000, 4000)
  name VARCHAR(100) NOT NULL,                          -- Account name (e.g., Cash, Sales)
  role VARCHAR(30) NOT NULL,                           -- Account role (cash, ar, ap, revenue, etc.)
  balance_type VARCHAR(6) NOT NULL,                    -- DEBIT or CREDIT normal balance
  locked BOOLEAN NOT NULL,                             -- Whether account is locked from changes
  active BOOLEAN NOT NULL,                             -- Whether account is active
  coa__id CHAR(32) NOT NULL REFERENCES chartofaccount (uuid) DEFERRABLE INITIALLY DEFERRED,
  role_default BOOLEAN NULL,                          -- Whether this is the default account for this role
  CONSTRAINT unique_code_for_coa_ UNIQUE (coa__id, code),
  CONSTRAINT only_one_account_assigned_as_default_for_role UNIQUE (
    coa__id, role, role_default
  )
);
COMMENT ON TABLE account IS 'Individual accounts within chart of accounts with hierarchical structure';
COMMENT ON COLUMN account.role IS 'Account role: cash, ar, ap, inventory, revenue, expense, equity, etc.';
COMMENT ON COLUMN account.balance_type IS 'Normal balance type: DEBIT (assets, expenses) or CREDIT (liabilities, equity, revenue)';

-- Individual transactions that make up journal entries
CREATE TABLE IF NOT EXISTS transaction (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  tx_type VARCHAR(10) NOT NULL,                        -- Transaction type: DEBIT or CREDIT
  amount DECIMAL NOT NULL,                             -- Transaction amount (always positive)
  description VARCHAR(100) NULL,                       -- Transaction description
  account_id CHAR(32) NOT NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  journal_entry_id CHAR(32) NOT NULL REFERENCES journalentry (uuid) DEFERRABLE INITIALLY DEFERRED,
  cleared BOOLEAN NOT NULL,                            -- Whether transaction has cleared the bank
  reconciled BOOLEAN NOT NULL                          -- Whether transaction has been reconciled
);
COMMENT ON TABLE transaction IS 'Individual debit/credit transactions that comprise journal entries';
COMMENT ON COLUMN transaction.tx_type IS 'DEBIT or CREDIT - follows double-entry bookkeeping rules';
COMMENT ON COLUMN transaction.cleared IS 'Whether this transaction has cleared the bank (for cash accounts)';

-- =====================================================
-- BANK ACCOUNT MANAGEMENT
-- =====================================================

-- Bank accounts linked to chart of accounts
CREATE TABLE IF NOT EXISTS bankaccount (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  -- Bank account details
  account_number VARCHAR(30) NULL,                     -- Bank account number
  routing_number VARCHAR(30) NULL,                     -- Bank routing number
  aba_number VARCHAR(30) NULL,                         -- ABA routing number
  swift_number VARCHAR(30) NULL,                       -- SWIFT code for international
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  name VARCHAR(150) NULL,                              -- Account nickname/description
  active BOOLEAN NOT NULL,                             -- Whether account is active
  hidden BOOLEAN NOT NULL,                             -- Whether to hide from UI
  entity__id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  account__id CHAR(32) NOT NULL REFERENCES account (uuid) DEFERRABLE INITIALLY DEFERRED,
  account_type VARCHAR(20) NOT NULL                    -- Account type: checking, savings, etc.
);
COMMENT ON TABLE bankaccount IS 'Bank accounts linked to general ledger accounts for reconciliation';
COMMENT ON COLUMN bankaccount.account_type IS 'Bank account type: checking, savings, money_market, etc.';

-- =====================================================
-- VENDOR MANAGEMENT
-- =====================================================

-- Vendor master data (suppliers/service providers)
CREATE TABLE IF NOT EXISTS vendor (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  -- Address information
  address_1 VARCHAR(70) NOT NULL,
  address_2 VARCHAR(70) NULL,
  city VARCHAR(70) NULL,
  state VARCHAR(70) NULL,
  zip_code VARCHAR(20) NULL,
  country VARCHAR(70) NULL,
  -- Contact information
  email VARCHAR(254) NULL,
  website VARCHAR(200) NULL,
  phone VARCHAR(30) NULL,
  -- Banking information for payments
  account_number VARCHAR(30) NULL,                     -- Vendor's bank account number
  routing_number VARCHAR(30) NULL,                     -- Vendor's bank routing number
  aba_number VARCHAR(30) NULL,                         -- ABA number
  swift_number VARCHAR(30) NULL,                       -- SWIFT code
  tax_id_number VARCHAR(30) NULL,                      -- Vendor's tax ID (EIN/SSN)
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  vendor_number VARCHAR(30) NULL,                      -- Vendor reference number
  vendor_name VARCHAR(100) NOT NULL,                   -- Vendor company/person name
  description TEXT NOT NULL,                           -- Vendor description/notes
  active BOOLEAN NOT NULL,                             -- Whether vendor is active
  hidden BOOLEAN NOT NULL,                             -- Whether to hide from lists
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*)
  ),                                                     -- JSON for additional vendor data
  entity__id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  account_type VARCHAR(20) NOT NULL                    -- Vendor account type classification
);
COMMENT ON TABLE vendor IS 'Vendor/supplier master data with payment and tax information';
COMMENT ON COLUMN vendor.tax_id_number IS 'Vendor tax ID for 1099 reporting and tax compliance';
COMMENT ON COLUMN vendor.account_type IS 'Vendor classification for reporting and payment processing';

-- =====================================================
-- DJANGO SESSION MANAGEMENT
-- =====================================================

-- Django session storage (framework-specific)
CREATE TABLE IF NOT EXISTS django_session (
  session_key VARCHAR(40) NOT NULL PRIMARY KEY,       -- Session identifier
  session_data TEXT NOT NULL,                         -- Serialized session data
  expire_date TIMESTAMP WITHOUT TIME ZONE NOT NULL    -- Session expiration timestamp
);
COMMENT ON TABLE django_session IS 'Django framework session storage for user authentication';

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Add indexes for commonly queried fields
CREATE INDEX IF NOT EXISTS idx_estimate_customer ON estimate(customer_id);
CREATE INDEX IF NOT EXISTS idx_estimate_entity ON estimate(entity_id);
CREATE INDEX IF NOT EXISTS idx_estimate_status ON estimate(status);
CREATE INDEX IF NOT EXISTS idx_estimate_number ON estimate(estimate_number);

CREATE INDEX IF NOT EXISTS idx_itemtransaction_item ON itemtransaction(item__id);
CREATE INDEX IF NOT EXISTS idx_itemtransaction_invoice ON itemtransaction(invoice__id);
CREATE INDEX IF NOT EXISTS idx_itemtransaction_bill ON itemtransaction(bill__id);
CREATE INDEX IF NOT EXISTS idx_itemtransaction_po ON itemtransaction(po__id);

CREATE INDEX IF NOT EXISTS idx_invoice_customer ON invoice(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoice_status ON invoice(invoice_status);
CREATE INDEX IF NOT EXISTS idx_invoice_number ON invoice(invoice_number);

CREATE INDEX IF NOT EXISTS idx_bill_vendor ON bill(vendor_id);
CREATE INDEX IF NOT EXISTS idx_bill_status ON bill(bill_status);

CREATE INDEX IF NOT EXISTS idx_journalentry_ledger ON journalentry(ledger_id);
CREATE INDEX IF NOT EXISTS idx_journalentry_posted ON journalentry(posted);
CREATE INDEX IF NOT EXISTS idx_journalentry_timestamp ON journalentry(timestamp);

CREATE INDEX IF NOT EXISTS idx_transaction_account ON transaction(account_id);
CREATE INDEX IF NOT EXISTS idx_transaction_je ON transaction(journal_entry_id);

CREATE INDEX IF NOT EXISTS idx_account_coa ON account(coa__id);
CREATE INDEX IF NOT EXISTS idx_account_role ON account(role);
CREATE INDEX IF NOT EXISTS idx_account_code ON account(code);

-- =====================================================
-- SUMMARY OF KEY RELATIONSHIPS
-- =====================================================

/*
ENTITY HIERARCHY:
- entity: Root company/organization
- entitystate: Manages document numbering sequences
- entitymanagement: User permissions for entities

CUSTOMER RELATIONSHIP MANAGEMENT:
- customer: Customer master data
- estimate: Project estimates/quotes
- invoice: Customer invoices with revenue recognition

VENDOR RELATIONSHIP MANAGEMENT:
- vendor: Vendor/supplier master data
- purchaseorder: Purchase orders
- bill: Vendor bills (accounts payable)

ITEM/INVENTORY MANAGEMENT:
- item: Product/service catalog
- unitofmeasure: Units of measure
- itemtransaction: Line items across all documents

ACCOUNTING SYSTEM:
- chartofaccount: Account structure templates
- account: Individual GL accounts
- ledger: Collections of journal entries
- journalentry: Double-entry journal entries
- transaction: Individual debits/credits

BANK INTEGRATION:
- bankaccount: Bank account master data
- importjob: Bank statement import jobs
- stagedtransaction: Imported transactions awaiting review

PERIOD CLOSING:
- closingentry: Period-end closing entries
- closingentrytransaction: Account balances at period end
*/
