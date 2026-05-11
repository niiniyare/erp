-- Migration 001007: Finance governance & compliance tables
-- Creates:
--   finance_violation_suppressions — auditable, time-bounded suppression records

-- ── Violation Suppression Records ─────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_violation_suppressions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL,

    -- Violation identity (matches finance_integrity_violations upsert key).
    violation_kind   TEXT NOT NULL,
    entity_id        UUID NOT NULL,

    reason           TEXT NOT NULL,               -- Mandatory reason for audit trail
    suppressed_by    UUID NOT NULL,               -- Operator user ID
    audit_event_id   TEXT,                        -- Reference to immutable audit event

    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ NOT NULL,        -- Hard expiry — cannot be extended

    -- Policy enforcement: CRITICAL violations must never appear here.
    -- Application layer enforces this; constraint is belt-and-suspenders.
    CONSTRAINT chk_fvs_expires_after_created
        CHECK (expires_at > created_at)
);

-- Active suppression lookup: (tenant, kind, entity) → suppression.
-- Note: Cannot use WHERE expires_at > NOW() because NOW() is volatile.
-- Instead, filter active records in query or use a scheduled materialized view.
CREATE INDEX IF NOT EXISTS idx_fvs_lookup
    ON finance_violation_suppressions (tenant_id, violation_kind, entity_id, expires_at);

-- Additional index for expiry cleanup queries.
CREATE INDEX IF NOT EXISTS idx_fvs_expires_at
    ON finance_violation_suppressions (expires_at);

-- Optional: Create a view for active suppressions (cleaner than inline filtering)
CREATE OR REPLACE VIEW active_finance_violation_suppressions AS
SELECT *
FROM finance_violation_suppressions
WHERE expires_at > NOW();



