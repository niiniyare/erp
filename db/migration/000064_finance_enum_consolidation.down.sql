-- +migrate Down
-- =====================================================================
-- FINANCE MODULE - ENUM CONSOLIDATION ROLLBACK
-- Reverts enum types back to VARCHAR with CHECK constraints
-- =====================================================================

BEGIN;

-- =====================================================================
-- STEP 1: Add temporary VARCHAR columns
-- =====================================================================

-- Add temporary VARCHAR columns to finance_accounts
ALTER TABLE finance_accounts 
ADD COLUMN root_type_old VARCHAR(20),
ADD COLUMN normal_balance_old VARCHAR(10),
ADD COLUMN validation_status_old VARCHAR(20);

-- Add temporary VARCHAR columns to finance_transactions
ALTER TABLE finance_transactions
ADD COLUMN transaction_type_old VARCHAR(30),
ADD COLUMN transaction_status_old VARCHAR(20),
ADD COLUMN approval_status_old VARCHAR(20),
ADD COLUMN validation_status_old VARCHAR(20),
ADD COLUMN recurring_frequency_old VARCHAR(20);

-- =====================================================================
-- STEP 2: Migrate data from enum to VARCHAR columns
-- =====================================================================

-- Migrate finance_accounts data
UPDATE finance_accounts 
SET root_type_old = root_type::text;

UPDATE finance_accounts 
SET normal_balance_old = normal_balance::text;

UPDATE finance_accounts 
SET validation_status_old = validation_status::text;

-- Migrate finance_transactions data
UPDATE finance_transactions 
SET transaction_type_old = transaction_type::text;

UPDATE finance_transactions 
SET transaction_status_old = transaction_status::text;

UPDATE finance_transactions 
SET approval_status_old = approval_status::text;

UPDATE finance_transactions 
SET validation_status_old = validation_status::text;

UPDATE finance_transactions 
SET recurring_frequency_old = recurring_frequency::text
WHERE recurring_frequency IS NOT NULL;

-- =====================================================================
-- STEP 3: Drop enum columns and rename old columns
-- =====================================================================

-- Update finance_accounts
ALTER TABLE finance_accounts 
DROP COLUMN root_type,
DROP COLUMN normal_balance,
DROP COLUMN validation_status;

ALTER TABLE finance_accounts 
RENAME COLUMN root_type_old TO root_type;

ALTER TABLE finance_accounts 
RENAME COLUMN normal_balance_old TO normal_balance;

ALTER TABLE finance_accounts 
RENAME COLUMN validation_status_old TO validation_status;

-- Update finance_transactions
ALTER TABLE finance_transactions 
DROP COLUMN transaction_type,
DROP COLUMN transaction_status,
DROP COLUMN approval_status,
DROP COLUMN validation_status,
DROP COLUMN recurring_frequency;

ALTER TABLE finance_transactions 
RENAME COLUMN transaction_type_old TO transaction_type;

ALTER TABLE finance_transactions 
RENAME COLUMN transaction_status_old TO transaction_status;

ALTER TABLE finance_transactions 
RENAME COLUMN approval_status_old TO approval_status;

ALTER TABLE finance_transactions 
RENAME COLUMN validation_status_old TO validation_status;

ALTER TABLE finance_transactions 
RENAME COLUMN recurring_frequency_old TO recurring_frequency;

-- =====================================================================
-- STEP 4: Add back CHECK constraints and set NOT NULL/defaults
-- =====================================================================

-- Add CHECK constraints to finance_accounts
ALTER TABLE finance_accounts 
ALTER COLUMN root_type SET NOT NULL,
ADD CONSTRAINT finance_accounts_root_type_check 
    CHECK (root_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE'));

ALTER TABLE finance_accounts 
ALTER COLUMN normal_balance SET NOT NULL,
ADD CONSTRAINT finance_accounts_normal_balance_check 
    CHECK (normal_balance IN ('DEBIT', 'CREDIT'));

ALTER TABLE finance_accounts 
ALTER COLUMN validation_status SET DEFAULT 'PENDING',
ADD CONSTRAINT finance_accounts_validation_status_check 
    CHECK (validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR'));

-- Add CHECK constraints to finance_transactions
ALTER TABLE finance_transactions 
ALTER COLUMN transaction_type SET NOT NULL,
ADD CONSTRAINT finance_transactions_transaction_type_check 
    CHECK (transaction_type IN ('MANUAL', 'SYSTEM', 'IMPORTED', 'RECURRING', 'ADJUSTMENT', 'CLOSING'));

ALTER TABLE finance_transactions 
ALTER COLUMN transaction_status SET NOT NULL,
ALTER COLUMN transaction_status SET DEFAULT 'DRAFT',
ADD CONSTRAINT finance_transactions_transaction_status_check 
    CHECK (transaction_status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'POSTED', 'CANCELLED', 'REVERSED'));

ALTER TABLE finance_transactions 
ALTER COLUMN approval_status SET DEFAULT 'NOT_REQUIRED',
ADD CONSTRAINT finance_transactions_approval_status_check 
    CHECK (approval_status IN ('NOT_REQUIRED', 'PENDING', 'APPROVED', 'REJECTED'));

ALTER TABLE finance_transactions 
ALTER COLUMN validation_status SET DEFAULT 'PENDING',
ADD CONSTRAINT finance_transactions_validation_status_check 
    CHECK (validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR'));

ALTER TABLE finance_transactions 
ADD CONSTRAINT finance_transactions_recurring_frequency_check 
    CHECK (recurring_frequency IS NULL OR 
           recurring_frequency IN ('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'YEARLY'));

-- =====================================================================
-- STEP 5: Drop enum types
-- =====================================================================

DROP TYPE IF EXISTS root_type_enum CASCADE;
DROP TYPE IF EXISTS normal_balance_enum CASCADE;
DROP TYPE IF EXISTS transaction_type_enum CASCADE;
DROP TYPE IF EXISTS transaction_status_enum CASCADE;
DROP TYPE IF EXISTS approval_status_enum CASCADE;
DROP TYPE IF EXISTS validation_status_enum CASCADE;
DROP TYPE IF EXISTS recurring_frequency_enum CASCADE;
DROP TYPE IF EXISTS account_type_enum CASCADE;
DROP TYPE IF EXISTS currency_code_enum CASCADE;
DROP TYPE IF EXISTS payment_method_enum CASCADE;

COMMIT;