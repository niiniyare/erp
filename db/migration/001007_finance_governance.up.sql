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
CREATE INDEX IF NOT EXISTS idx_fvs_active
    ON finance_violation_suppressions (tenant_id, violation_kind, entity_id, expires_at)
    WHERE expires_at > NOW();

-- Expiry cleanup: find all expired suppressions efficiently.
CREATE INDEX IF NOT EXISTS idx_fvs_expires_at
    ON finance_violation_suppressions (expires_at);
