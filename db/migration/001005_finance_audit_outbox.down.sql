-- Rollback: drop finance audit outbox added in 001005

DROP INDEX IF EXISTS idx_finance_audit_outbox_dead;
DROP INDEX IF EXISTS idx_finance_audit_outbox_pending;
DROP TABLE IF EXISTS finance_audit_outbox;
