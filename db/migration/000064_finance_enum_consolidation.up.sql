-- +migrate Up
-- =====================================================================
-- FINANCE MODULE - ENUM CONSOLIDATION MIGRATION
-- Consolidates existing VARCHAR CHECK constraints into proper PostgreSQL ENUMs
-- with uppercase values for consistency and better performance
-- =====================================================================

BEGIN;

-- =====================================================================
-- STEP 1: Create new enum types
-- =====================================================================

-- Root type enumeration (replacing existing CHECK constraint)
CREATE TYPE root_type_enum AS ENUM (
    'ASSET',     -- Economic resources controlled by the entity
    'LIABILITY', -- Present obligations arising from past events
    'EQUITY',    -- Residual interest in assets after deducting liabilities
    'REVENUE',   -- Increases in equity from business operations
    'EXPENSE'    -- Decreases in equity from business operations
);

-- Normal balance enumeration (replacing existing CHECK constraint)
CREATE TYPE normal_balance_enum AS ENUM (
    'DEBIT',  -- Left side increases (Assets, Expenses)
    'CREDIT'  -- Right side increases (Liabilities, Equity, Revenue)
);

-- Transaction type enumeration (replacing existing CHECK constraint)
CREATE TYPE transaction_type_enum AS ENUM (
    'MANUAL',     -- User-created journal entries
    'SYSTEM',     -- System-generated transactions
    'IMPORTED',   -- External system imports
    'RECURRING',  -- Auto-generated recurring entries
    'ADJUSTMENT', -- Correcting entries
    'CLOSING',    -- Period-end closings
    'JOURNAL',    -- General journal entries
    'INVOICE',    -- Invoice-related transactions
    'PAYMENT',    -- Payment transactions
    'PURCHASE',   -- Purchase transactions
    'REVERSAL'    -- Transaction reversals
);

-- Transaction status enumeration (replacing existing CHECK constraint)
CREATE TYPE transaction_status_enum AS ENUM (
    'DRAFT',            -- Initial state, editable
    'PENDING_APPROVAL', -- Awaiting approval
    'APPROVED',         -- Approved but not posted
    'POSTED',           -- Posted to ledger, affects balances
    'CANCELLED',        -- Cancelled before posting
    'REVERSED'          -- Posted but subsequently reversed
);

-- Approval status enumeration (replacing existing CHECK constraint)
CREATE TYPE approval_status_enum AS ENUM (
    'NOT_REQUIRED', -- No approval needed
    'PENDING',      -- Waiting for approver
    'APPROVED',     -- Approved by authorized user
    'REJECTED'      -- Rejected by approver
);

-- Validation status enumeration (replacing existing CHECK constraint)
CREATE TYPE validation_status_enum AS ENUM (
    'PENDING', -- Validation not performed
    'VALID',   -- Passed all validations
    'WARNING', -- Has warnings but acceptable
    'ERROR'    -- Has validation errors
);

-- Recurring frequency enumeration (replacing existing CHECK constraint)
CREATE TYPE recurring_frequency_enum AS ENUM (
    'DAILY',
    'WEEKLY',
    'MONTHLY',
    'QUARTERLY',
    'YEARLY'
);

-- Additional financial enums for extended functionality
CREATE TYPE account_type_enum AS ENUM (
    -- Asset account types
    'BANK',
    'CASH',
    'PETTY_CASH',
    'ACCOUNTS_RECEIVABLE',
    'OTHER_RECEIVABLE',
    'INVENTORY',
    'PREPAID_EXPENSES',
    'FIXED_ASSETS',
    'ACCUMULATED_DEPRECIATION',
    'INTANGIBLE_ASSETS',
    'INVESTMENTS',
    'OTHER_CURRENT_ASSETS',
    'OTHER_ASSETS',
    
    -- Liability account types
    'ACCOUNTS_PAYABLE',
    'OTHER_PAYABLE',
    'ACCRUED_EXPENSES',
    'SHORT_TERM_DEBT',
    'LONG_TERM_DEBT',
    'DEFERRED_REVENUE',
    'TAX_PAYABLE',
    'OTHER_CURRENT_LIABILITIES',
    'OTHER_LIABILITIES',
    
    -- Equity account types
    'OWNERS_EQUITY',
    'RETAINED_EARNINGS',
    'COMMON_STOCK',
    'PREFERRED_STOCK',
    'ADDITIONAL_PAID_IN_CAPITAL',
    'TREASURY_STOCK',
    'OTHER_EQUITY',
    
    -- Revenue/Income account types
    'OPERATING_REVENUE',
    'SERVICE_REVENUE',
    'PRODUCT_REVENUE',
    'INTEREST_INCOME',
    'DIVIDEND_INCOME',
    'GAIN_ON_SALE',
    'OTHER_INCOME',
    
    -- Expense account types
    'COST_OF_GOODS_SOLD',
    'SALARIES_WAGES',
    'BENEFITS',
    'RENT_EXPENSE',
    'UTILITIES',
    'OFFICE_EXPENSES',
    'PROFESSIONAL_FEES',
    'MARKETING_ADVERTISING',
    'TRAVEL_EXPENSES',
    'DEPRECIATION_EXPENSE',
    'INTEREST_EXPENSE',
    'TAX_EXPENSE',
    'OTHER_EXPENSES'
);

CREATE TYPE currency_code_enum AS ENUM (
    'USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'CNY',
    'SEK', 'NOK', 'DKK', 'PLN', 'CZK', 'HUF', 'RUB', 'INR',
    'BRL', 'MXN', 'ZAR', 'KRW', 'SGD', 'HKD', 'NZD', 'TRY',
    'AED', 'SAR', 'QAR', 'KWD', 'BHD', 'OMR', 'JOD', 'LBP',
    'EGP', 'MAD', 'TND', 'DZD', 'LYD', 'SDG', 'ETB', 'KES',
    'UGX', 'TZS', 'RWF', 'BIF', 'DJF', 'SOS', 'MGA', 'MUR',
    'SCR', 'MZN', 'ZWL', 'BWP', 'SZL', 'LSL', 'NAD', 'AOA',
    'ZMW', 'MWK', 'GMD', 'GHS', 'NGN', 'XOF', 'XAF'
);

CREATE TYPE payment_method_enum AS ENUM (
    'CASH',
    'CHECK',
    'CREDIT_CARD',
    'DEBIT_CARD',
    'BANK_TRANSFER',
    'WIRE_TRANSFER',
    'ACH',
    'PAYPAL',
    'STRIPE',
    'SQUARE',
    'MOBILE_MONEY',
    'CRYPTOCURRENCY',
    'OTHER'
);

-- =====================================================================
-- STEP 2: Add temporary columns with enum types
-- =====================================================================

-- Add temporary enum columns to finance_chart_of_accounts
ALTER TABLE finance_chart_of_accounts 
ADD COLUMN root_type_new root_type_enum,
ADD COLUMN normal_balance_new normal_balance_enum,
ADD COLUMN validation_status_new validation_status_enum;

-- Add temporary enum columns to finance_transactions
ALTER TABLE finance_transactions
ADD COLUMN transaction_type_new transaction_type_enum,
ADD COLUMN transaction_status_new transaction_status_enum,
ADD COLUMN approval_status_new approval_status_enum,
ADD COLUMN validation_status_new validation_status_enum,
ADD COLUMN recurring_frequency_new recurring_frequency_enum;

-- =====================================================================
-- STEP 3: Migrate data from VARCHAR to enum columns
-- =====================================================================

-- Migrate finance_chart_of_accounts data
UPDATE finance_chart_of_accounts 
SET root_type_new = root_type::root_type_enum;

UPDATE finance_chart_of_accounts 
SET normal_balance_new = normal_balance::normal_balance_enum;

UPDATE finance_chart_of_accounts 
SET validation_status_new = validation_status::validation_status_enum;

-- Migrate finance_transactions data
UPDATE finance_transactions 
SET transaction_type_new = transaction_type::transaction_type_enum;

UPDATE finance_transactions 
SET transaction_status_new = transaction_status::transaction_status_enum;

UPDATE finance_transactions 
SET approval_status_new = approval_status::approval_status_enum;

UPDATE finance_transactions 
SET validation_status_new = validation_status::validation_status_enum;

UPDATE finance_transactions 
SET recurring_frequency_new = recurring_frequency::recurring_frequency_enum
WHERE recurring_frequency IS NOT NULL;

-- =====================================================================
-- STEP 4: Drop old columns and constraints, rename new columns
-- =====================================================================

-- Update finance_chart_of_accounts
ALTER TABLE finance_chart_of_accounts 
DROP COLUMN root_type,
DROP COLUMN normal_balance,
DROP COLUMN validation_status;

ALTER TABLE finance_chart_of_accounts 
RENAME COLUMN root_type_new TO root_type;

ALTER TABLE finance_chart_of_accounts 
RENAME COLUMN normal_balance_new TO normal_balance;

ALTER TABLE finance_chart_of_accounts 
RENAME COLUMN validation_status_new TO validation_status;

-- Set NOT NULL constraints and defaults
ALTER TABLE finance_chart_of_accounts 
ALTER COLUMN root_type SET NOT NULL,
ALTER COLUMN normal_balance SET NOT NULL,
ALTER COLUMN validation_status SET DEFAULT 'PENDING';

-- Update finance_transactions
ALTER TABLE finance_transactions 
DROP COLUMN transaction_type,
DROP COLUMN transaction_status,
DROP COLUMN approval_status,
DROP COLUMN validation_status,
DROP COLUMN recurring_frequency;

ALTER TABLE finance_transactions 
RENAME COLUMN transaction_type_new TO transaction_type;

ALTER TABLE finance_transactions 
RENAME COLUMN transaction_status_new TO transaction_status;

ALTER TABLE finance_transactions 
RENAME COLUMN approval_status_new TO approval_status;

ALTER TABLE finance_transactions 
RENAME COLUMN validation_status_new TO validation_status;

ALTER TABLE finance_transactions 
RENAME COLUMN recurring_frequency_new TO recurring_frequency;

-- Set NOT NULL constraints and defaults
ALTER TABLE finance_transactions 
ALTER COLUMN transaction_type SET NOT NULL,
ALTER COLUMN transaction_status SET NOT NULL,
ALTER COLUMN transaction_status SET DEFAULT 'DRAFT',
ALTER COLUMN approval_status SET DEFAULT 'NOT_REQUIRED',
ALTER COLUMN validation_status SET DEFAULT 'PENDING';

-- =====================================================================
-- STEP 5: Add enum documentation
-- =====================================================================

COMMENT ON TYPE root_type_enum IS 'Primary account classifications following the fundamental accounting equation: Assets = Liabilities + Equity';
COMMENT ON TYPE normal_balance_enum IS 'Determines which side of the accounting equation increases account balance';
COMMENT ON TYPE transaction_type_enum IS 'Transaction categorization by source and business purpose for audit trail and reporting';
COMMENT ON TYPE transaction_status_enum IS 'Transaction lifecycle states from draft creation through final posting';
COMMENT ON TYPE approval_status_enum IS 'Workflow approval states for transactions requiring authorization';
COMMENT ON TYPE validation_status_enum IS 'Data validation states ensuring financial data integrity';
COMMENT ON TYPE recurring_frequency_enum IS 'Frequency patterns for recurring financial transactions';
COMMENT ON TYPE account_type_enum IS 'Detailed account classifications within each root type following standard chart of accounts structure';
COMMENT ON TYPE currency_code_enum IS 'ISO 4217 standard currency codes for multi-currency financial operations';
COMMENT ON TYPE payment_method_enum IS 'Supported payment mechanisms for receivables and payables processing';

COMMIT;