-- Migration 001003: Finance DB-level financial constraints
-- Enforces double-entry bookkeeping and data integrity at the database layer.

-- ── Transaction entries: balanced debits/credits ────────────────────────────
-- Each entry row may not have both debit and credit simultaneously, and amounts
-- must be non-negative. The balance invariant (∑debit = ∑credit per transaction)
-- is enforced at service layer; here we prevent individual row pathologies.

ALTER TABLE finance_transaction_entries
    ADD CONSTRAINT chk_entry_amounts_non_negative
        CHECK (debit_amount >= 0 AND credit_amount >= 0),
    ADD CONSTRAINT chk_entry_not_both_sides
        CHECK (NOT (debit_amount > 0 AND credit_amount > 0)),
    ADD CONSTRAINT chk_entry_at_least_one_side
        CHECK (debit_amount > 0 OR credit_amount > 0);

-- ── Transactions: valid type enum ──────────────────────────────────────────
-- Guard against values not covered by domain.TransactionType.
-- Extend this list when new types are added to the domain.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_transaction_type_valid'
    ) THEN
        ALTER TABLE finance_transactions
            ADD CONSTRAINT chk_transaction_type_valid
                CHECK (transaction_type IN (
                    'JOURNAL', 'INVOICE', 'PAYMENT', 'RECEIPT',
                    'CREDIT_NOTE', 'DEBIT_NOTE', 'ADJUSTMENT', 'REVERSAL',
                    'OPENING_BALANCE', 'CLOSING_BALANCE', 'ACCRUAL',
                    'DEPRECIATION', 'PAYROLL', 'BANK_TRANSFER', 'TAX'
                ));
    END IF;
END $$;

-- ── Reversal history: FK integrity ─────────────────────────────────────────
-- Ensure reversal_transaction_id always points to an existing transaction.
-- If the table/column already has a FK, this is a no-op.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_reversal_history_reversal_txn'
    ) THEN
        ALTER TABLE finance_reversal_history
            ADD CONSTRAINT fk_reversal_history_reversal_txn
                FOREIGN KEY (reversal_transaction_id)
                REFERENCES finance_transactions (id)
                ON DELETE RESTRICT;
    END IF;
END $$;

-- ── Posting date sanity ─────────────────────────────────────────────────────
-- Posting date must not precede transaction date by more than 1 year, and must
-- not be more than 1 year in the future. Prevents fat-finger posting errors.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_posting_date_reasonable'
    ) THEN
        ALTER TABLE finance_transactions
            ADD CONSTRAINT chk_posting_date_reasonable
                CHECK (
                    posting_date IS NULL OR (
                        posting_date >= transaction_date - INTERVAL '1 year'
                        AND posting_date <= NOW() + INTERVAL '1 year'
                    )
                );
    END IF;
END $$;

-- ── Index: efficient approval-queue lookups ─────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_finance_txn_approval_status
    ON finance_transactions (tenant_id, approval_status)
    WHERE approval_status = 'PENDING_APPROVAL';

-- ── Index: recurring transactions due-date scan ─────────────────────────────
CREATE INDEX IF NOT EXISTS idx_finance_txn_recurring_due
    ON finance_transactions (tenant_id, next_recurring_date)
    WHERE is_recurring = TRUE AND next_recurring_date IS NOT NULL;
