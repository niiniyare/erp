-- Rollback: 001003_finance_db_constraints

DROP INDEX IF EXISTS idx_finance_txn_recurring_due;
DROP INDEX IF EXISTS idx_finance_txn_approval_status;

ALTER TABLE finance_transactions
    DROP CONSTRAINT IF EXISTS chk_posting_date_reasonable,
    DROP CONSTRAINT IF EXISTS chk_transaction_type_valid;

ALTER TABLE finance_reversal_history
    DROP CONSTRAINT IF EXISTS fk_reversal_history_reversal_txn;

ALTER TABLE finance_transaction_entries
    DROP CONSTRAINT IF EXISTS chk_entry_at_least_one_side,
    DROP CONSTRAINT IF EXISTS chk_entry_not_both_sides,
    DROP CONSTRAINT IF EXISTS chk_entry_amounts_non_negative;
