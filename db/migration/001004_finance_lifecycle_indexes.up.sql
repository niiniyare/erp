-- Migration 001004: Finance lifecycle performance indexes and period-close constraints
--
-- Context:
--   Phase 5 introduces period-close integrity gating and operational tooling that
--   scans large transaction sets. Without dedicated indexes these scans require
--   full table scans, causing lock contention spikes on large tenants.
--
-- Note: indexes use plain CREATE INDEX (not CONCURRENTLY) because migration
-- runners execute inside a transaction block. CONCURRENTLY is incompatible
-- with transaction blocks. For zero-downtime production deployments on large
-- tables, run these index statements manually outside a transaction first,
-- then apply the migration.

-- ── Posting-date index for period-close and integrity scan queries ────────────
-- ScanPostedTransactions, integrity cron, and period-boundary queries all filter
-- on (tenant_id, posting_date) with status=POSTED. Without this index every
-- scan is a sequential read of the full transactions table.
CREATE INDEX IF NOT EXISTS idx_finance_txn_posting_date
    ON finance_transactions (tenant_id, posting_date)
    WHERE transaction_status = 'POSTED';

-- ── Pending-approval index for orphaned-workflow detection ───────────────────
-- FindOrphanedWorkflows pages through PENDING_APPROVAL transactions.
-- Without this index the scan hits every tenant's transactions.
CREATE INDEX IF NOT EXISTS idx_finance_txn_pending_approval
    ON finance_transactions (tenant_id, created_at)
    WHERE transaction_status = 'PENDING_APPROVAL';

-- ── Reversal history — original transaction index ────────────────────────────
-- ScanReversalChains calls GetByOriginal per reversed transaction.
-- An index on original_transaction_id speeds this scan significantly.
CREATE INDEX IF NOT EXISTS idx_finance_reversal_history_original
    ON finance_reversal_history (original_transaction_id);

-- ── Period close consistency constraint ──────────────────────────────────────
-- If a period is SOFT_CLOSED or HARD_CLOSED, closed_at MUST be set.
-- Prevents a partial-close state where status is SOFT_CLOSED but the
-- timestamp was lost (e.g. failed migration or direct DB update).
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
