-- Migration 001004: Finance lifecycle performance indexes and period-close constraints
--
-- Context:
--   Phase 5 introduces period-close integrity gating and operational tooling that
--   scans large transaction sets. Without dedicated indexes these scans require
--   full table scans, causing lock contention spikes on large tenants.
--
--   All new indexes use CONCURRENTLY to avoid table locks during deployment.
--   The period-close consistency constraint is added idempotently.

-- ── Posting-date index for period-close and integrity scan queries ────────────
-- ScanPostedTransactions, integrity cron, and period-boundary queries all filter
-- on (tenant_id, posting_date) with status=POSTED. Without this index every
-- scan is a sequential read of the full transactions table.
--
-- CONCURRENTLY: safe to deploy during normal traffic. No table lock taken.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_finance_txn_posting_date
    ON finance_transactions (tenant_id, posting_date)
    WHERE transaction_status = 'POSTED';

-- ── Pending-approval index for orphaned-workflow detection ───────────────────
-- FindOrphanedWorkflows pages through PENDING_APPROVAL transactions.
-- Without this index the scan hits every tenant's transactions.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_finance_txn_pending_approval
    ON finance_transactions (tenant_id, created_at)
    WHERE transaction_status = 'PENDING_APPROVAL';

-- ── Period close consistency constraint ──────────────────────────────────────
-- If a period has been soft or hard closed, closed_at MUST be set.
-- Prevents a partial-close state where status is SOFT_CLOSED but the
-- timestamp was lost (e.g. a failed migration or a direct DB update).
--
-- Wrapped in a DO block to be idempotent across repeated runs.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_period_closed_at_consistency'
          AND conrelid = 'finance_accounting_periods'::regclass
    ) THEN
        ALTER TABLE finance_accounting_periods
            ADD CONSTRAINT chk_period_closed_at_consistency
                CHECK (
                    status NOT IN ('SOFT_CLOSED', 'HARD_CLOSED') OR closed_at IS NOT NULL
                );
    END IF;
END $$;

-- ── Reversal history — tenant index ─────────────────────────────────────────
-- ScanReversalChains calls GetByOriginal per reversed transaction.
-- An index on original_transaction_id speeds this scan significantly.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_finance_reversal_history_original
    ON finance_reversal_history (original_transaction_id);
