-- Rollback: drop indexes and constraint added in 001004

DROP INDEX CONCURRENTLY IF EXISTS idx_finance_reversal_history_original;

ALTER TABLE finance_accounting_periods
    DROP CONSTRAINT IF EXISTS chk_period_closed_at_consistency;

DROP INDEX CONCURRENTLY IF EXISTS idx_finance_txn_pending_approval;
DROP INDEX CONCURRENTLY IF EXISTS idx_finance_txn_posting_date;
