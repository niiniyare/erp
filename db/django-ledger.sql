-- =============================================================================
-- ACCOUNTING SYSTEM DATABASE SCHEMA
-- =============================================================================
-- This schema represents a accounting system with support for:
-- - Chart of Accounts management
-- - Bank transaction imports and processing
-- - Customer/Vendor management
-- - Invoice/Bill processing
-- - Inventory and purchase order management
-- - Journal entries and ledger management
-- =============================================================================
-- =============================================================================
-- COMPREHENSIVE ACCOUNTING AND BUSINESS MANAGEMENT DATABASE SCHEMA
-- =============================================================================
-- This schema supports a complete business accounting system including:
-- - Customer and vendor management
-- - Estimates, invoices, and purchase orders
-- - Inventory and item tracking
-- - Chart of accounts and general ledger
-- - Bank account management and transaction import
-- - Journal entries and closing procedures
-- =============================================================================

-- =============================================================================
-- ENTITY MANAGEMENT TABLES
-- =============================================================================



CREATE TABLE IF NOT EXISTS entities (
-- Hierarchical organization fields (tree structure)
  path VARCHAR(255) NOT NULL UNIQUE, 
  depth INTEGER NOT NULL CHECK (depth >= 0), 
  numchild INTEGER NOT NULL CHECK (numchild >= 0), --

-- Audit and identification fields
  slug VARCHAR(50) NOT NULL UNIQUE, -- URL-friendly identifier
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier (32-char hex)

-- Core business information
  name VARCHAR(150) NOT NULL, -- Legal/business name of entity
  hidden BOOLEAN NOT NULL, -- Whether entity is hidden from lists
  accrual_method BOOLEAN NOT NULL, -- TRUE=accrual accounting, FALSE=cash basis
  fy_start_month INTEGER NOT NULL, -- Fiscal year start month (1-12)
  picture VARCHAR(100) NULL, -- Path to entity logo/picture
  admin_id INTEGER NOT NULL, -- Primary administrator user ID
  default_coa_id UUID NULL UNIQUE, -- Default chart of accounts
  last_closing_date DATE NULL, -- Last period closing date
  meta TEXT NULL, -- Additional metadata (JSON/text)

-- Contact information
  address_1 VARCHAR(70) NOT NULL, -- Primary address line
  address_2 VARCHAR(70) NULL, -- Secondary address line
  city VARCHAR(70) NULL, -- City name
  state VARCHAR(70) NULL, -- State/province
  zip_code VARCHAR(20) NULL, -- Postal code
  country VARCHAR(70) NULL, -- Country name
  email VARCHAR(254) NULL, -- Primary email address
  website VARCHAR(200) NULL, -- Website URL
  phone VARCHAR(30) NULL -- Primary phone number
);

COMMENT ON TABLE entities IS ' * Purpose: Stores business entities/organizations/companies
 * Description: Core table representing different business entities that can
 *              have their own accounting books, customers, vendors, etc.
 *              Uses tree structure for hierarchical organization relationships.
';

COMMENT ON COLUMN entities.path IS 'Tree path for hierarchical queries';
COMMENT ON COLUMN entities.depth IS 'Depth in organization tree';
COMMENT ON COLUMN entities.numchild IS 'Number of child entities';



CREATE TABLE users (
    uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
    email character varying(255) NOT NULL,
    full_name character varying(255) NOT NULL,
    entity_id UUID NOT NULL References entities (uuid),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_login_at timestamp with time zone,
    deleted_at timestamp with time zone
);

/*
 * Table: entitystate
 * Purpose: Maintains sequence counters and state information for each entity
 * Description: Tracks fiscal year settings and auto-increment sequences for
 *              document numbers (invoices, estimates, etc.) per entity
 */
CREATE TABLE IF NOT EXISTS entitystate (
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  fiscal_year SMALLINT NULL, -- Current fiscal year
  key VARCHAR(10) NOT NULL, -- Type of sequence (INV, EST, PO, etc.)
  sequence BIGINT NOT NULL, -- Current sequence number
  entity__id UUID NOT NULL, -- Reference to parent entity
  entity_unit_id uuid NULL -- Reference to entity unit (if applicable)
);

/*
 * Table: entitymanagement
 * Purpose: Manages user permissions for business entities
 * Description: Controls which users can access and manage specific entities
 *              with different permission levels
 */
CREATE TABLE IF NOT EXISTS entitymanagement (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  permission_level VARCHAR(10) NOT NULL, -- Permission level (ADMIN, READ, WRITE, etc.)
  entity_id UUID NOT NULL, -- Reference to entity
  user_id INTEGER NOT NULL -- Reference to user account
);

-- =============================================================================
-- CUSTOMER AND VENDOR MANAGEMENT TABLES
-- =============================================================================

/*
 * Table: customer
 * Purpose: Stores customer information and contact details
 * Description: Maintains customer database with contact information, billing
 *              details, and sales tax rates for invoicing purposes
 */
CREATE TABLE IF NOT EXISTS customer (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Core customer information
  customer_name VARCHAR(100) NOT NULL, -- Customer business/personal name
  customer_number VARCHAR(30) NOT NULL, -- Internal customer reference number
  description TEXT NOT NULL, -- Customer notes/description
  active BOOLEAN NOT NULL, -- Whether customer is currently active
  hidden BOOLEAN NOT NULL, -- Whether to hide from customer lists
  entity__id UUID NOT NULL, -- Reference to owning entity

-- Contact information
  address_1 VARCHAR(70) NOT NULL, -- Primary billing address
  address_2 VARCHAR(70) NULL, -- Secondary address line
  city VARCHAR(70) NULL, -- City name
  state VARCHAR(70) NULL, -- State/province
  zip_code VARCHAR(20) NULL, -- Postal/ZIP code
  country VARCHAR(70) NULL, -- Country name
  email VARCHAR(254) NULL, -- Primary email for invoices
  website VARCHAR(200) NULL, -- Customer website
  phone VARCHAR(30) NULL, -- Primary contact phone

-- Business settings
  sales_tax_rate REAL NULL, -- Default sales tax rate (decimal)
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

/*
 * Table: vendor
 * Purpose: Stores vendor/supplier information and payment details
 * Description: Maintains vendor database with contact information and
 *              banking details for bill payments and purchase orders
 */
CREATE TABLE IF NOT EXISTS vendor (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Core vendor information
  vendor_name VARCHAR(100) NOT NULL, -- Vendor business name
  vendor_number VARCHAR(30) NULL, -- Internal vendor reference number
  description TEXT NOT NULL, -- Vendor notes/description
  active BOOLEAN NOT NULL, -- Whether vendor is currently active
  hidden BOOLEAN NOT NULL, -- Whether to hide from vendor lists
  entity__id UUID NOT NULL, -- Reference to owning entity

-- Contact information
  address_1 VARCHAR(70) NOT NULL, -- Primary address
  address_2 VARCHAR(70) NULL, -- Secondary address line
  city VARCHAR(70) NULL, -- City name
  state VARCHAR(70) NULL, -- State/province
  zip_code VARCHAR(20) NULL, -- Postal/ZIP code
  country VARCHAR(70) NULL, -- Country name
  email VARCHAR(254) NULL, -- Primary contact email
  website VARCHAR(200) NULL, -- Vendor website
  phone VARCHAR(30) NULL, -- Primary contact phone

-- Banking and payment information
  account_number VARCHAR(30) NULL, -- Bank account number for payments
  routing_number VARCHAR(30) NULL, -- Bank routing number
  aba_number VARCHAR(30) NULL, -- ABA routing number
  swift_number VARCHAR(30) NULL, -- SWIFT code for international transfers
  tax_id_number VARCHAR(30) NULL, -- Tax ID/EIN for 1099 reporting
  account_type VARCHAR(20) NOT NULL, -- Account type (CHECKING, SAVINGS, etc.)
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

-- =============================================================================
-- INVENTORY AND ITEM MANAGEMENT TABLES
-- =============================================================================

/*
 * Table: unitofmeasure
 * Purpose: Defines units of measurement for inventory items
 * Description: Standardizes units like pieces, pounds, gallons, hours, etc.
 *              for consistent item quantification across the system
 */
CREATE TABLE IF NOT EXISTS unitofmeasure (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  name VARCHAR(50) NOT NULL, -- Full unit name (e.g., Pounds)
  unit_abbr VARCHAR(10) NOT NULL, -- Unit abbreviation (e.g., lbs)
  is_active BOOLEAN NOT NULL, -- Whether unit is available for use
  entity_id UUID NOT NULL -- Reference to owning entity
);

/*
 * Table: item
 * Purpose: Stores inventory items, products, and services
 * Description: Central catalog of all sellable items, materials, and services
 *              with pricing, inventory tracking, and account mapping
 */
CREATE TABLE IF NOT EXISTS item (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Item identification
  name VARCHAR(100) NOT NULL, -- Item display name
  item_number VARCHAR(30) NOT NULL, -- Internal item number/SKU
  item_type VARCHAR(1) NULL, -- Item type code (P=Product, S=Service, etc.)
  sku VARCHAR(50) NULL, -- Stock Keeping Unit code
  upc VARCHAR(50) NULL, -- Universal Product Code
  item_id VARCHAR(50) NULL, -- External/vendor item ID
  item_role VARCHAR(10) NULL, -- Item role (MATERIAL, LABOR, etc.)

-- Item configuration
  is_active BOOLEAN NOT NULL, -- Whether item is available for use
  default_amount DECIMAL NOT NULL, -- Default selling price
  for_inventory BOOLEAN NOT NULL, -- Whether item is tracked in inventory
  is_product_or_service BOOLEAN NOT NULL, -- TRUE=product, FALSE=service
  sold_as_unit BOOLEAN NOT NULL, -- Whether sold in whole units only

-- Inventory tracking
  inventory_received DECIMAL NULL, -- Total quantity received
  inventory_received_value DECIMAL NULL, -- Total value of inventory received

-- Account mappings for different transaction types
  cogs_account_id uuid NULL, -- Cost of Goods Sold account
  earnings_account_id uuid NULL, -- Revenue/Sales account
  expense_account_id uuid NULL, -- Expense account for purchases
  inventory_account_id uuid NULL, -- Inventory asset account

-- Relationships
  entity_id UUID NOT NULL, -- Reference to owning entity
  uom_id UUID NOT NULL, -- Reference to unit of measure
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

/*
 * Table: itemtransaction
 * Purpose: Stores line items for all transaction types (invoices, bills, POs, estimates)
 * Description: Individual line items that appear on business documents,
 *              tracking quantities, costs, and linking to various document types
 */
CREATE TABLE IF NOT EXISTS itemtransaction (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Basic transaction details
  quantity REAL NULL, -- Quantity ordered/invoiced/received
  unit_cost REAL NULL, -- Cost per unit
  total_amount DECIMAL NULL, -- Total line amount (quantity × unit_cost)
  item_notes VARCHAR(400) NULL, -- Line item specific notes

-- Purchase Order specific fields
  po_quantity REAL NULL, -- PO quantity
  po_unit_cost REAL NULL, -- PO unit cost
  po_total_amount DECIMAL NULL, -- PO total amount
  po_item_status VARCHAR(15) NULL, -- PO line status (PENDING, RECEIVED, etc.)

-- Cost Estimate specific fields (ce = cost estimate)
  ce_quantity REAL NULL, -- Estimated quantity
  ce_unit_cost_estimate REAL NULL, -- Estimated unit cost
  ce_cost_estimate DECIMAL NULL, -- Total estimated cost
  ce_unit_revenue_estimate REAL NULL, -- Estimated unit selling price
  ce_revenue_estimate DECIMAL NULL, -- Total estimated revenue

-- Document relationships (each line item belongs to one document type)
  bill__id uuid NULL, -- Reference to vendor bill
  ce__id uuid NULL, -- Reference to cost estimate
  invoice__id uuid NULL, -- Reference to customer invoice
  po__id uuid NULL, -- Reference to purchase order
  item__id UUID NOT NULL, -- Reference to item being transacted
  entity_unit_id uuid NULL -- Reference to entity unit
);

-- =============================================================================
-- SALES AND PURCHASING DOCUMENT TABLES
-- =============================================================================

/*
 * Table: estimate
 * Purpose: Stores project estimates and quotes for customers
 * Description: Tracks estimates/quotes with approval workflow, cost breakdowns,
 *              and conversion to invoices and purchase orders
 */
CREATE TABLE IF NOT EXISTS estimate (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Estimate identification
  estimate_number VARCHAR(20) NOT NULL, -- Sequential estimate number
  title VARCHAR(250) NOT NULL, -- Estimate/project title
  terms VARCHAR(10) NOT NULL, -- Payment terms (NET30, COD, etc.)
  markdown_notes TEXT NULL, -- Detailed project notes (Markdown format)

-- Workflow status and dates
  status VARCHAR(10) NOT NULL, -- Current status (DRAFT, REVIEW, APPROVED, etc.)
  date_draft DATE NULL, -- Date estimate was created
  date_in_review DATE NULL, -- Date sent for review
  date_approved DATE NULL, -- Date customer approved
  date_completed DATE NULL, -- Date project completed
  date_canceled DATE NULL, -- Date estimate was canceled
  date_void DATE NULL, -- Date estimate was voided

-- Financial estimates by category
  revenue_estimate DECIMAL NOT NULL, -- Total estimated revenue
  labor_estimate DECIMAL NOT NULL, -- Estimated labor costs
  material_estimate DECIMAL NOT NULL, -- Estimated material costs
  equipment_estimate DECIMAL NOT NULL, -- Estimated equipment costs
  other_estimate DECIMAL NOT NULL, -- Other estimated costs

-- Relationships
  customer_id UUID NOT NULL, -- Reference to customer
  entity_id UUID NOT NULL -- Reference to owning entity
);

/*
 * Table: invoice
 * Purpose: Stores customer invoices and billing information
 * Description: Tracks customer invoices with payment status, accounting entries,
 *              and revenue recognition for accrual accounting
 */
CREATE TABLE IF NOT EXISTS invoice (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Invoice identification
  invoice_number VARCHAR(20) NOT NULL, -- Sequential invoice number
  terms VARCHAR(10) NOT NULL, -- Payment terms (NET30, COD, etc.)
  date_due DATE NULL, -- Payment due date
  markdown_notes TEXT NULL, -- Invoice notes (Markdown format)

-- Financial amounts
  amount_due DECIMAL NOT NULL, -- Total amount due from customer
  amount_paid DECIMAL NOT NULL, -- Amount already paid
  amount_receivable DECIMAL NOT NULL, -- Amount still receivable
  amount_unearned DECIMAL NOT NULL, -- Unearned revenue (prepayments)
  amount_earned DECIMAL NOT NULL, -- Earned revenue recognized

-- Accounting settings
  accrue BOOLEAN NOT NULL, -- Whether to accrue revenue
  progress DECIMAL NOT NULL, -- Project completion percentage

-- Workflow status and dates
  invoice_status VARCHAR(10) NOT NULL, -- Current status (DRAFT, SENT, PAID, etc.)
  date_draft DATE NULL, -- Date invoice was created
  date_in_review DATE NULL, -- Date sent for review
  date_approved DATE NULL, -- Date approved for sending
  date_paid DATE NULL, -- Date fully paid
  date_void DATE NULL, -- Date invoice was voided
  date_canceled DATE NULL, -- Date invoice was canceled

-- Account mappings for accounting entries
  cash_account_id UUID NOT NULL, -- Cash account for payments
  prepaid_account_id UUID NOT NULL, -- Prepaid/unearned revenue account
  unearned_account_id UUID NOT NULL, -- Unearned revenue liability account

-- Relationships
  customer_id UUID NOT NULL, -- Reference to customer
  ce__id uuid NULL, -- Reference to originating estimate
  ledger_id UUID NOT NULL UNIQUE, -- Reference to accounting ledger
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

/*
 * Table: purchaseorder
 * Purpose: Stores purchase orders to vendors
 * Description: Tracks purchase orders with approval workflow and receiving status
 */
CREATE TABLE IF NOT EXISTS purchaseorder (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Purchase order identification
  po_number VARCHAR(20) NOT NULL, -- Sequential PO number
  po_title VARCHAR(250) NOT NULL, -- PO title/description
  markdown_notes TEXT NULL, -- PO notes (Markdown format)

-- Financial amounts
  po_amount DECIMAL NOT NULL, -- Total PO amount
  po_amount_received DECIMAL NOT NULL, -- Amount of goods/services received

-- Workflow status and dates
  po_status VARCHAR(10) NOT NULL, -- Current status (DRAFT, SENT, FULFILLED, etc.)
  date_draft DATE NULL, -- Date PO was created
  date_in_review DATE NULL, -- Date sent for review
  date_approved DATE NULL, -- Date approved for sending
  date_fulfilled DATE NULL, -- Date completely fulfilled
  date_canceled DATE NULL, -- Date PO was canceled
  date_void DATE NULL, -- Date PO was voided

-- Relationships
  ce__id uuid NULL, -- Reference to originating estimate
  entity_id UUID NOT NULL -- Reference to owning entity
);

/*
 * Table: bill
 * Purpose: Stores vendor bills and payables
 * Description: Tracks vendor bills with payment status and accounting entries
 */
CREATE TABLE IF NOT EXISTS bill (
-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Bill identification
  bill_number VARCHAR(20) NOT NULL, -- Sequential bill number
  xref VARCHAR(50) NULL, -- Vendor's invoice/reference number
  terms VARCHAR(10) NOT NULL, -- Payment terms (NET30, COD, etc.)
  date_due DATE NULL, -- Payment due date
  markdown_notes TEXT NULL, -- Bill notes (Markdown format)

-- Financial amounts
  amount_due DECIMAL NOT NULL, -- Total amount due to vendor
  amount_paid DECIMAL NOT NULL, -- Amount already paid
  amount_receivable DECIMAL NOT NULL, -- Amount still payable
  amount_unearned DECIMAL NOT NULL, -- Prepaid expenses
  amount_earned DECIMAL NOT NULL, -- Expenses recognized

-- Accounting settings
  accrue BOOLEAN NOT NULL, -- Whether to accrue expenses
  progress DECIMAL NOT NULL, -- Service completion percentage

-- Workflow status and dates
  bill_status VARCHAR(10) NOT NULL, -- Current status (DRAFT, APPROVED, PAID, etc.)
  date_draft DATE NULL, -- Date bill was entered
  date_in_review DATE NULL, -- Date sent for review
  date_approved DATE NULL, -- Date approved for payment
  date_paid DATE NULL, -- Date fully paid
  date_void DATE NULL, -- Date bill was voided
  date_canceled DATE NULL, -- Date bill was canceled

-- Account mappings for accounting entries
  cash_account_id uuid NULL, -- Cash account for payments
  prepaid_account_id uuid NULL, -- Prepaid expense account
  unearned_account_id uuid NULL, -- Unearned/prepaid liability account

-- Relationships
  vendor_id UUID NOT NULL, -- Reference to vendor
  ce__id uuid NULL, -- Reference to originating estimate
  ledger_id UUID NOT NULL UNIQUE, -- Reference to accounting ledger
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

-- =============================================================================
-- ACCOUNTING AND FINANCIAL TABLES
-- =============================================================================

/*
 * Table: chartofaccount
 * Purpose: Defines chart of accounts templates
 * Description: Templates for account structures that can be applied to entities
 */
CREATE TABLE IF NOT EXISTS chartofaccount (
  slug VARCHAR(50) NOT NULL UNIQUE, -- URL-friendly identifier
  name VARCHAR(150) NULL, -- Chart of accounts name
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  description TEXT NULL, -- Chart description
  entity_id UUID NOT NULL, -- Reference to owning entity
  active BOOLEAN NOT NULL -- Whether chart is active
);

/*
 * Table: account
 * Purpose: Stores individual accounts within chart of accounts
 * Description: Individual GL accounts with hierarchical structure supporting
 *              standard accounting roles (Assets, Liabilities, Equity, etc.)
 */
CREATE TABLE IF NOT EXISTS account (
-- Hierarchical structure fields
  path VARCHAR(255) NOT NULL UNIQUE, -- Tree path for hierarchical queries
  depth INTEGER NOT NULL CHECK (depth >= 0), -- Depth in account tree
  numchild INTEGER NOT NULL CHECK (numchild >= 0), -- Number of child accounts

-- Audit fields
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier

-- Account identification
  code VARCHAR(10) NOT NULL, -- Account code/number
  name VARCHAR(100) NOT NULL, -- Account name
  role VARCHAR(30) NOT NULL, -- Account role (ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE)
  balance_type VARCHAR(6) NOT NULL, -- Normal balance type (DEBIT or CREDIT)

-- Account settings
  locked BOOLEAN NOT NULL, -- Whether account is locked from editing
  active BOOLEAN NOT NULL, -- Whether account is available for use
  role_default BOOLEAN NULL, -- Whether this is the default account for its role

-- Relationships
  coa__id UUID NOT NULL, -- Reference to chart of accounts

-- Constraints ensure unique codes within each chart and only one default per role
  CONSTRAINT unique_code_for_coa_ UNIQUE (coa__id, code),
  CONSTRAINT only_one_account_assigned_as_default_for_role UNIQUE (
    coa__id, role, role_default
  )
);

/*
 * Table: ledger
 * Purpose: Groups related journal entries into ledgers
 * Description: Organizes journal entries by transaction type, period, or business unit
 */
CREATE TABLE IF NOT EXISTS ledger (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  name VARCHAR(150) NULL, -- Ledger name/description
  posted BOOLEAN NOT NULL, -- Whether ledger entries are posted
  locked BOOLEAN NOT NULL, -- Whether ledger is locked from changes
  hidden BOOLEAN NOT NULL, -- Whether to hide from ledger lists
  entity_id UUID NOT NULL, -- Reference to owning entity
  ledger_xid VARCHAR(150) NULL, -- External ledger identifier
  additional_info JSONB NULL -- Additional custom fields (JSON)
);

/*
 * Table: journalentry
 * Purpose: Stores accounting journal entry headers
 * Description: Groups related debits and credits into complete journal entries
 *              with audit trail and posting controls
 */
CREATE TABLE IF NOT EXISTS journalentry (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  je_number VARCHAR(25) NOT NULL, -- Journal entry number
  timestamp TIMESTAMP NOT NULL, -- Transaction date/time
  description VARCHAR(70) NULL, -- Entry description
  activity VARCHAR(20) NULL, -- Activity type (SALES, PURCHASES, etc.)
  origin VARCHAR(30) NULL, -- Source of entry (MANUAL, IMPORT, AUTO)
  posted BOOLEAN NOT NULL, -- Whether entry is posted to GL
  locked BOOLEAN NOT NULL, -- Whether entry is locked from changes
  is_closing_entry BOOLEAN NOT NULL, -- Whether this is a period closing entry
  entity_unit_id uuid NULL, -- Reference to entity unit
  ledger_id UUID NOT NULL -- Reference to containing ledger
);

/*
 * Table: transaction
 * Purpose: Stores individual debit and credit transactions
 * Description: Individual debits and credits that make up journal entries,
 *              linked to specific GL accounts with bank reconciliation support
 */
CREATE TABLE IF NOT EXISTS transaction (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  tx_type VARCHAR(10) NOT NULL, -- Transaction type (DEBIT or CREDIT)
  amount DECIMAL NOT NULL, -- Transaction amount (always positive)
  description VARCHAR(100) NULL, -- Transaction description
  cleared BOOLEAN NOT NULL, -- Whether transaction has cleared bank
  reconciled BOOLEAN NOT NULL, -- Whether transaction is bank reconciled
  account_id UUID NOT NULL, -- Reference to GL account
  journal_entry_id UUID NOT NULL -- Reference to parent journal entry
);

-- =============================================================================
-- PERIOD CLOSING AND YEAR-END TABLES
-- =============================================================================

/*
 * Table: closingentry
 * Purpose: Stores period-end closing entries
 * Description: Tracks period closing procedures with date controls to prevent
 *              posting transactions to closed periods
 */
CREATE TABLE IF NOT EXISTS closingentry (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  closing_date DATE NOT NULL, -- Period end date being closed
  posted BOOLEAN NOT NULL, -- Whether closing is posted
  markdown_notes TEXT NULL, -- Closing notes (Markdown format)
  entity__id UUID NOT NULL, -- Reference to entity being closed
  ledger__id UUID NOT NULL UNIQUE, -- Reference to closing ledger

-- Ensure only one closing per entity per date
  CONSTRAINT unique_entity_closing_date UNIQUE (
    entity__id, closing_date
  )
);

/*
 * Table: closingentrytransaction
 * Purpose: Stores individual account balances for closing entries
 * Description: Captures account balances at period close with activity breakdown
 */
CREATE TABLE IF NOT EXISTS closingentrytransaction (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  activity VARCHAR(20) NULL, -- Activity type (OPERATING, FINANCING, etc.)
  tx_type VARCHAR(10) NOT NULL, -- Transaction type (DEBIT or CREDIT)
  balance DECIMAL NOT NULL, -- Account balance at closing
  account__id UUID NOT NULL, -- Reference to GL account
  closing_entry__id UUID NOT NULL, -- Reference to closing entry
  unit__id uuid NULL, -- Reference to entity unit

-- Ensure unique closing entry per account/unit/activity combination
  CONSTRAINT unique_closing_entry UNIQUE (
    closing_entry__id, account__id,
    unit__id, activity
  )
);

-- =============================================================================
-- BANK MANAGEMENT AND IMPORT TABLES
-- =============================================================================

/*
 * Table: bankaccount
 * Purpose: Stores bank account information
 * Description: Manages bank accounts linked to GL accounts for cash management
 *              and automated transaction import
 */
CREATE TABLE IF NOT EXISTS bankaccount (
  created TIMESTAMP NOT NULL, -- Record creation timestamp
  updated TIMESTAMP NULL, -- Last modification timestamp
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier
  name VARCHAR(150) NULL, -- Bank account nickname
  active BOOLEAN NOT NULL, -- Whether account is active
  hidden BOOLEAN NOT NULL, -- Whether to hide from lists

-- Banking details
  account_number VARCHAR(30) NULL, -- Bank account number
  routing_number VARCHAR(30) NULL, -- Bank routing number
  aba_number VARCHAR(30) NULL, -- ABA routing number
  swift_number VARCHAR(30) NULL, -- SWIFT code for international
  account_type VARCHAR(20) NOT NULL, -- Account type (CHECKING, SAVINGS, etc.)

-- Relationships
  entity__id UUID NOT NULL, -- Reference to owning entity
  account__id UUID NOT NULL -- Reference to linked GL account
);


-- -----------------------------------------------------------------------------
-- IMPORT JOB MANAGEMENT
-- -----------------------------------------------------------------------------
-- Purpose: Manages batch import jobs for bank transactions from external sources
-- Use Case: When importing bank statements or transaction files, create an import
--           job to track the process and group related staged transactions
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS importjob (
-- Audit trail fields
  created TIMESTAMP NOT NULL, -- When the import job was created
  updated TIMESTAMP NULL, -- Last modification timestamp

-- Primary identifier
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier for the import job

-- Job details
  description VARCHAR(200) NOT NULL, -- Human-readable description of what's being imported
  completed BOOLEAN NOT NULL, -- Whether the import job has finished processing

-- Relationships
  bank_account__id UUID NOT NULL, -- Which bank account these transactions belong to
  ledger__id uuid NULL UNIQUE -- Optional: Associated ledger for accounting entries
);

-- -----------------------------------------------------------------------------
-- STAGED TRANSACTION PROCESSING
-- -----------------------------------------------------------------------------
-- Purpose: Temporary storage for imported transactions before they're processed
--          into the main accounting system
-- Use Case: Bank transactions are first staged here, reviewed/categorized, then
--           converted to proper accounting transactions
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stagedtransaction (
-- Audit trail fields
  created TIMESTAMP NOT NULL, -- When the transaction was staged
  updated TIMESTAMP NULL, -- Last modification timestamp

-- Primary identifier
  uuid UUID NOT NULL PRIMARY KEY, -- Unique identifier for staged transaction

-- Transaction identification (from bank/external system)
  fit_id VARCHAR(100) NOT NULL, -- Financial Institution Transaction ID

-- Core transaction data
  date_posted DATE NOT NULL, -- Date the transaction was posted by the bank
  amount DECIMAL NULL, -- Transaction amount (positive or negative)
  amount_split DECIMAL NULL, -- If transaction is split, this portion's amount
  name VARCHAR(200) NULL, -- Payee/Payer name from bank
  memo VARCHAR(200) NULL, -- Transaction description/memo from bank

-- Processing and categorization
  account__id uuid NULL, -- Which accounting account to assign this to
  activity VARCHAR(20) NULL, -- Type of activity (DEBIT, CREDIT, etc.)
  bundle_split BOOLEAN NOT NULL, -- Whether this transaction can be split into multiple entries

-- Hierarchical structure for split transactions
  parent_id uuid NULL, -- Reference to parent transaction if this is a split

-- Relationships
  import_job_id UUID NOT NULL, -- Which import job this transaction belongs to
  transaction__id uuid NULL UNIQUE, -- Reference to final accounting transaction (after processing)
  unit__id uuid NULL -- Business unit/department assignment
);


-- -- =============================================================================
-- -- FOREIGN KEY CONSTRAINTS AND RELATIONSHIPS
-- -- =============================================================================
-- These constraints ensure data integrity and define the relationships between
-- tables. All constraints use DEFERRABLE INITIALLY DEFERRED to allow for
-- flexible transaction processing while maintaining referential integrity.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Import Job Relationships
-- -----------------------------------------------------------------------------
-- Links import jobs to bank accounts and optionally to ledgers
ALTER TABLE importjob
ADD CONSTRAINT importjob_bank_account__id_fkey
FOREIGN KEY (bank_account__id)
REFERENCES bankaccount (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE importjob
ADD CONSTRAINT importjob_ledger__id_fkey
FOREIGN KEY (ledger__id)
REFERENCES ledger (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- -----------------------------------------------------------------------------
-- Staged Transaction Relationships
-- -----------------------------------------------------------------------------
-- Links staged transactions to accounts, import jobs, final transactions, and units
ALTER TABLE stagedtransaction
ADD CONSTRAINT stagedtransaction_account__id_fkey
FOREIGN KEY (account__id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE stagedtransaction
ADD CONSTRAINT stagedtransaction_import_job_id_fkey
FOREIGN KEY (import_job_id)
REFERENCES importjob (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE stagedtransaction
ADD CONSTRAINT stagedtransaction_transaction__id_fkey
FOREIGN KEY (transaction__id)
REFERENCES transaction (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- ALTER TABLE stagedtransaction
-- ADD CONSTRAINT stagedtransaction_unit__id_fkey
-- FOREIGN KEY (unit__id)
-- REFERENCES entityunit (uuid)
-- DEFERRABLE INITIALLY DEFERRED;

-- Self-referencing constraint for split transaction hierarchy
ALTER TABLE stagedtransaction
ADD CONSTRAINT stagedtransaction_parent_id_fkey
FOREIGN KEY (parent_id)
REFERENCES stagedtransaction (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- -----------------------------------------------------------------------------
-- Chart of Accounts Relationships
-- -----------------------------------------------------------------------------
-- Links charts of accounts to their owning entities
ALTER TABLE chartofaccount
ADD CONSTRAINT chartofaccount_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- -----------------------------------------------------------------------------
-- Account  Relationships
-- -----------------------------------------------------------------------------
-- Links individual accounts to their chart of accounts
ALTER TABLE account
ADD CONSTRAINT account_coa__id_fkey
FOREIGN KEY (coa__id)
REFERENCES chartofaccount (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- -----------------------------------------------------------------------------
-- Transaction  Relationships
-- -----------------------------------------------------------------------------
-- Links transactions to their accounts and journal entries
ALTER TABLE transaction
ADD CONSTRAINT transaction_account_id_fkey
FOREIGN KEY (account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE transaction
ADD CONSTRAINT transaction_journal_entry_id_fkey
FOREIGN KEY (journal_entry_id)
REFERENCES journalentry (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- -----------------------------------------------------------------------------
-- Bank Account Relationships
-- -----------------------------------------------------------------------------
-- Links bank accounts to entities and their corresponding accounting accounts
ALTER TABLE bankaccount
ADD CONSTRAINT bankaccount_entity__id_fkey
FOREIGN KEY (entity__id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE bankaccount
ADD CONSTRAINT bankaccount_account__id_fkey
FOREIGN KEY (account__id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- -----------------------------------------------------------------------------
-- Vendor  Relationships
-- -----------------------------------------------------------------------------
-- Links vendors to their managing entities
ALTER TABLE vendor
ADD CONSTRAINT vendor_entity__id_fkey
FOREIGN KEY (entity__id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- =============================================================================
-- ADDITIONAL FOREIGN KEY CONSTRAINTS FROM RELATED TABLES
-- =============================================================================
-- Note: The following constraints reference tables not fully defined above
-- but are part of the complete accounting system schema
-- =============================================================================

-- Estimate  constraints (Customer Relationship Management)
ALTER TABLE estimate
ADD CONSTRAINT estimate_customer_id_fkey
FOREIGN KEY (customer_id)
REFERENCES customer (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE estimate
ADD CONSTRAINT estimate_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Item transaction  constraints (Inventory Management)
ALTER TABLE itemtransaction
ADD CONSTRAINT itemtransaction_bill__id_fkey
FOREIGN KEY (bill__id)
REFERENCES bill (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE itemtransaction
ADD CONSTRAINT itemtransaction_ce__id_fkey
FOREIGN KEY (ce__id)
REFERENCES estimate (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- ALTER TABLE itemtransaction
-- ADD CONSTRAINT itemtransaction_entity_unit_id_fkey
-- FOREIGN KEY (entity_unit_id)
-- REFERENCES entityunit (uuid)
-- DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE itemtransaction
ADD CONSTRAINT itemtransaction_invoice__id_fkey
FOREIGN KEY (invoice__id)
REFERENCES invoice (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE itemtransaction
ADD CONSTRAINT itemtransaction_item__id_fkey
FOREIGN KEY (item__id)
REFERENCES item (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE itemtransaction
ADD CONSTRAINT itemtransaction_po__id_fkey
FOREIGN KEY (po__id)
REFERENCES purchaseorder (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Unit of measure constraints
ALTER TABLE unitofmeasure
ADD CONSTRAINT unitofmeasure_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Purchase order constraints
ALTER TABLE purchaseorder
ADD CONSTRAINT purchaseorder_ce__id_fkey
FOREIGN KEY (ce__id)
REFERENCES estimate (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE purchaseorder
ADD CONSTRAINT purchaseorder_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Invoice  constraints (Accounts Receivable)
ALTER TABLE invoice
ADD CONSTRAINT invoice_cash_account_id_fkey
FOREIGN KEY (cash_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE invoice
ADD CONSTRAINT invoice_ce__id_fkey
FOREIGN KEY (ce__id)
REFERENCES estimate (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE invoice
ADD CONSTRAINT invoice_customer_id_fkey
FOREIGN KEY (customer_id)
REFERENCES customer (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE invoice
ADD CONSTRAINT invoice_ledger_id_fkey
FOREIGN KEY (ledger_id)
REFERENCES ledger (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE invoice
ADD CONSTRAINT invoice_prepaid_account_id_fkey
FOREIGN KEY (prepaid_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE invoice
ADD CONSTRAINT invoice_unearned_account_id_fkey
FOREIGN KEY (unearned_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Entity state and management constraints
ALTER TABLE entitystate
ADD CONSTRAINT entitystate_entity__id_fkey
FOREIGN KEY (entity__id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;
-- TODO
-- ALTER TABLE entitystate
-- ADD CONSTRAINT entitystate_entity_unit_id_fkey
-- FOREIGN KEY (entity_unit_id)
-- REFERENCES entityunit (uuid)
-- DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE entitymanagement
ADD CONSTRAINT entitymanagement_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;
--- TODO
-- ALTER TABLE entitymanagement
-- ADD CONSTRAINT entitymanagement_user_id_fkey
-- FOREIGN KEY (user_id)
-- REFERENCES auth_user (id)
-- DEFERRABLE INITIALLY DEFERRED;

-- Customer  constraints
ALTER TABLE customer
ADD CONSTRAINT customer_entity__id_fkey
FOREIGN KEY (entity__id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Bill  constraints (Accounts Payable)
ALTER TABLE bill
ADD CONSTRAINT bill_cash_account_id_fkey
FOREIGN KEY (cash_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE bill
ADD CONSTRAINT bill_ce__id_fkey
FOREIGN KEY (ce__id)
REFERENCES estimate (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE bill
ADD CONSTRAINT bill_ledger_id_fkey
FOREIGN KEY (ledger_id)
REFERENCES ledger (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE bill
ADD CONSTRAINT bill_prepaid_account_id_fkey
FOREIGN KEY (prepaid_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE bill
ADD CONSTRAINT bill_unearned_account_id_fkey
FOREIGN KEY (unearned_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE bill
ADD CONSTRAINT bill_vendor_id_fkey
FOREIGN KEY (vendor_id)
REFERENCES vendor (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Item  constraints (Inventory)
ALTER TABLE item
ADD CONSTRAINT item_cogs_account_id_fkey
FOREIGN KEY (cogs_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE item
ADD CONSTRAINT item_earnings_account_id_fkey
FOREIGN KEY (earnings_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE item
ADD CONSTRAINT item_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE item
ADD CONSTRAINT item_expense_account_id_fkey
FOREIGN KEY (expense_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE item
ADD CONSTRAINT item_inventory_account_id_fkey
FOREIGN KEY (inventory_account_id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE item
ADD CONSTRAINT item_uom_id_fkey
FOREIGN KEY (uom_id)
REFERENCES unitofmeasure (uuid)
DEFERRABLE INITIALLY DEFERRED;
-- -- TODO
-- -- Entity  constraints
-- ALTER TABLE entities
-- ADD CONSTRAINT entities_admin_id_fkey
-- FOREIGN KEY (admin_id)
-- REFERENCES auth_user (id)
-- DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE entities
ADD CONSTRAINT entities_default_coa_id_fkey
FOREIGN KEY (default_coa_id)
REFERENCES chartofaccount (uuid)
DEFERRABLE INITIALLY DEFERRED;
-- TODO
-- -- Journal entry constraints
-- ALTER TABLE journalentry
-- ADD CONSTRAINT journalentry_entity_unit_id_fkey
-- FOREIGN KEY (entity_unit_id)
-- REFERENCES entityunit (uuid)
-- DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE journalentry
ADD CONSTRAINT journalentry_ledger_id_fkey
FOREIGN KEY (ledger_id)
REFERENCES ledger (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Ledger  constraints
ALTER TABLE ledger
ADD CONSTRAINT ledger_entity_id_fkey
FOREIGN KEY (entity_id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- Closing entry constraints (Year-end closing)
ALTER TABLE closingentrytransaction
ADD CONSTRAINT closingentrytransaction_account__id_fkey
FOREIGN KEY (account__id)
REFERENCES account (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE closingentrytransaction
ADD CONSTRAINT closingentrytransaction_closing_entry__id_fkey
FOREIGN KEY (closing_entry__id)
REFERENCES closingentry (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- ALTER TABLE closingentrytransaction
-- ADD CONSTRAINT closingentrytransaction_unit__id_fkey
-- FOREIGN KEY (unit__id)
-- REFERENCES entityunit (uuid)
-- DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE closingentry
ADD CONSTRAINT closingentry_entity__id_fkey
FOREIGN KEY (entity__id)
REFERENCES entities (uuid)
DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE closingentry
ADD CONSTRAINT closingentry_ledger__id_fkey
FOREIGN KEY (ledger__id)
REFERENCES ledger (uuid)
DEFERRABLE INITIALLY DEFERRED;

-- =============================================================================
--Tax Management
-- =============================================================================




CREATE TABLE IF NOT EXISTS taxrate (
  uuid UUID NOT NULL PRIMARY KEY,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    name VARCHAR(100) NOT NULL,
    rate DECIMAL(5,4) NOT NULL, -- e.g., 0.0825 for 8.25%
    tax_type VARCHAR(20) NOT NULL, -- sales, use, vat, etc.
    jurisdiction VARCHAR(100), -- State, County, City
    effective_date DATE NOT NULL,
    expiry_date DATE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    entity_id UUID NOT NULL REFERENCES entities (uuid)
  );



CREATE TABLE IF NOT EXISTS taxline (
    uuid UUID NOT NULL PRIMARY KEY,
    taxable_amount DECIMAL NOT NULL,
    tax_amount DECIMAL NOT NULL,
    tax_rate_id UUID NOT NULL REFERENCES taxrate (uuid),
    invoice_id uuid REFERENCES invoice (uuid),
    bill_id uuid REFERENCES bill (uuid)
  );



-- =============================================================================
-- Payment Tracking 
-- =============================================================================

CREATE TABLE IF NOT EXISTS payment (
  uuid UUID NOT NULL PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  payment_number VARCHAR(20) NOT NULL,
  payment_date DATE NOT NULL,
  amount DECIMAL NOT NULL,
  payment_method VARCHAR(20) NOT NULL, -- check, card, ach, wire
  reference_number VARCHAR(50),
  notes TEXT,
  bank_account_id uuid REFERENCES bankaccount (uuid),
  ledger_id UUID NOT NULL REFERENCES ledger (uuid),
  entity_id UUID NOT NULL REFERENCES entities (uuid)
);

CREATE TABLE IF NOT EXISTS paymentallocation (
  uuid UUID NOT NULL PRIMARY KEY,
  amount_allocated DECIMAL NOT NULL,
  payment_id UUID NOT NULL REFERENCES payment (uuid),
  invoice_id uuid REFERENCES invoice (uuid),
  bill_id uuid REFERENCES bill (uuid)
);




-- =============================================================================
--  Document Workflow State Machine
-- =============================================================================



CREATE TABLE IF NOT EXISTS workflowstate (
  uuid UUID NOT NULL PRIMARY KEY,
  document_type VARCHAR(20) NOT NULL, -- invoice, estimate, po, bill
  from_status VARCHAR(20) NOT NULL,
  to_status VARCHAR(20) NOT NULL,
  required_permission VARCHAR(20),
  auto_transition BOOLEAN DEFAULT FALSE,
  entity_id UUID NOT NULL REFERENCES entities (uuid)
);

CREATE TABLE IF NOT EXISTS documenthistory (
  uuid UUID NOT NULL PRIMARY KEY,
  timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  document_type VARCHAR(20) NOT NULL,
  document_id UUID NOT NULL,
  from_status VARCHAR(20),
  to_status VARCHAR(20) NOT NULL,
  -- TODO REFERENCES user id 
  user_id INTEGER ,
  -- user_id INTEGER REFERENCES auth_user (id),
  notes TEXT,
  entity_id UUID NOT NULL REFERENCES entities (uuid)
);

-- =============================================================================
--  Materialized Views for Reporting
-- =============================================================================


-- Account balances view (refresh nightly)
CREATE MATERIALIZED VIEW account_balances AS
SELECT 
    a.uuid as account_id,
    a.code,
    a.name,
    a.balance_type,
    a.coa__id,
    COALESCE(SUM(CASE WHEN t.tx_type = 'DEBIT' THEN t.amount ELSE -t.amount END), 0) as balance,
    MAX(je.timestamp) as last_transaction_date
FROM account a
LEFT JOIN transaction t ON a.uuid = t.account_id
LEFT JOIN journalentry je ON t.journal_entry_id = je.uuid AND je.posted = TRUE
GROUP BY a.uuid, a.code, a.name, a.balance_type, a.coa__id;

-- Aging report view
CREATE MATERIALIZED VIEW ar_aging AS
SELECT 
    i.customer_id,
    c.customer_name,
    SUM(CASE WHEN CURRENT_DATE - i.date_due <= 30 THEN i.amount_receivable ELSE 0 END) as current_amount,
    SUM(CASE WHEN CURRENT_DATE - i.date_due BETWEEN 31 AND 60 THEN i.amount_receivable ELSE 0 END) as days_31_60,
    SUM(CASE WHEN CURRENT_DATE - i.date_due BETWEEN 61 AND 90 THEN i.amount_receivable ELSE 0 END) as days_61_90,
    SUM(CASE WHEN CURRENT_DATE - i.date_due > 90 THEN i.amount_receivable ELSE 0 END) as over_90_days
FROM invoice i
JOIN customer c ON i.customer_id = c.uuid
WHERE i.invoice_status NOT IN ('paid', 'void', 'canceled') AND i.amount_receivable > 0
GROUP BY i.customer_id, c.customer_name;





-- =============================================================================
--  Data Validation Functions
-- =============================================================================

-- -- Function to validate journal entry balance
-- CREATE OR REPLACE FUNCTION validate_journal_entry_balance(je_id uuid)
-- RETURNS BOOLEAN AS $$
-- DECLARE
--     debit_total DECIMAL;
--     credit_total DECIMAL;
-- BEGIN
--     SELECT 
--         COALESCE(SUM(CASE WHEN tx_type = 'DEBIT' THEN amount ELSE 0 END), 0),
--         COALESCE(SUM(CASE WHEN tx_type = 'CREDIT' THEN amount ELSE 0 END), 0)
--     INTO debit_total, credit_total
--     FROM transaction 
--     WHERE journal_entry_id = je_id;
--     
--     RETURN debit_total = credit_total;
-- END;
-- $$ LANGUAGE plpgsql;

-- Add constraint to ensure balanced journal entries
ALTER TABLE journalentry 
  ADD CONSTRAINT chk_balanced_je 
  CHECK (NOT posted OR validate_journal_entry_balance(uuid));



-- =============================================================================
--  Multi-Currency Support
-- =============================================================================


CREATE TABLE IF NOT EXISTS currency (
  uuid UUID NOT NULL PRIMARY KEY,
    code CHAR(3) NOT NULL UNIQUE, -- USD, EUR, etc.
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10),
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS exchangerate (
  uuid UUID NOT NULL PRIMARY KEY,
    from_currency_id UUID NOT NULL REFERENCES currency (uuid),
    to_currency_id UUID NOT NULL REFERENCES currency (uuid),
    rate DECIMAL(12,6) NOT NULL,
    effective_date DATE NOT NULL,
    source VARCHAR(50) -- API source, manual, etc.
);

-- Add currency columns to financial tables
ALTER TABLE invoice ADD COLUMN currency_id uuid REFERENCES currency (uuid);
ALTER TABLE bill ADD COLUMN currency_id uuid REFERENCES currency (uuid);


-- =============================================================================
--  Document Attachments
-- =============================================================================

CREATE TABLE IF NOT EXISTS attachment (
  uuid UUID NOT NULL PRIMARY KEY,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    document_type VARCHAR(20) NOT NULL,
    document_id UUID NOT NULL,
    description TEXT,
    entity_id UUID NOT NULL REFERENCES entities (uuid)
  );


-- =============================================================================
-- Budget and Forecast 
-- =============================================================================

-- CREATE TABLE IF NOT EXISTS budget (
--   uuid UUID NOT NULL PRIMARY KEY,
--   created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
--   name VARCHAR(150) NOT NULL,
--   fiscal_year INTEGER NOT NULL,
--   version INTEGER NOT NULL DEFAULT 1,
--   status VARCHAR(20) NOT NULL DEFAULT 'draft',
--   entity_id UUID NOT NULL REFERENCES entities (uuid)
-- );

-- CREATE TABLE IF NOT EXISTS budgetline (
--   uuid UUID NOT NULL PRIMARY KEY,
--   account_id UUID NOT NULL REFERENCES account (uuid),
--   budget_id UUID NOT NULL REFERENCES budget (uuid),
--   period INTEGER NOT NULL, -- 1-12 for months
--   budgeted_amount DECIMAL(15,2) NOT NULL,
--   unit_id uuid REFERENCES entityunit (uuid)
-- );


-- Ensure the uuid-ossp extension is available for 
-- You only need to run this once per database.
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Define an ENUM type for budget status for better data integrity
CREATE TYPE budget_status AS ENUM ('draft', 'pending', 'approved', 'rejected', 'final');

CREATE TABLE IF NOT EXISTS budget (
  uuid UUID NOT NULL PRIMARY KEY,
  created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,   name VARCHAR(150) NOT NULL,
  -- Store with time zone and default to creation time

  fiscal_year INTEGER NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  status budget_status NOT NULL DEFAULT 'draft',   entity_id UUID NOT NULL REFERENCES entities (uuid)
);

CREATE TABLE IF NOT EXISTS budgetline (
  uuid UUID NOT NULL PRIMARY KEY ,
  account_id UUID NOT NULL REFERENCES account (uuid),
  budget_id UUID NOT NULL REFERENCES budget (uuid),
  period INTEGER NOT NULL CHECK (period >= 1 AND period <= 12), 
  budgeted_amount DECIMAL(15,2) NOT NULL,
  unit_id UUID REFERENCES entities (uuid)
);


CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    entity_id INT REFERENCES entities(uuid), -- Context entity
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

 the root of the organizational hierarchy

COMMENT ON TABLE audit_logs IS '';
