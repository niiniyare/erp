-- Migration 001006: Finance autonomous safety tables
-- Creates:
--   finance_integrity_violations — lifecycle tracking for detected integrity violations
--   finance_audit_chain          — tamper-evident hash chain for CRITICAL audit events

-- ── Integrity Violation Lifecycle ────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_integrity_violations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,

    -- Violation identity (upsert key).
    kind         TEXT NOT NULL,            -- e.g. 'UNBALANCED_TRANSACTION'
    entity_id    UUID NOT NULL,            -- primary affected entity

    severity     TEXT NOT NULL            -- CRITICAL | HIGH | MEDIUM | LOW
                 CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW')),
    detail       TEXT NOT NULL,
    lifecycle    TEXT NOT NULL DEFAULT 'OPEN'
                 CHECK (lifecycle IN ('OPEN', 'ACKNOWLEDGED', 'RESOLVED')),

    detected_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acked_at     TIMESTAMPTZ,
    resolved_at  TIMESTAMPTZ,
    acked_by     UUID,
    resolved_by  UUID,
    repair_note  TEXT,

    -- Upsert key: one active violation record per (tenant, kind, entity).
    CONSTRAINT uq_finance_integrity_violations
        UNIQUE (tenant_id, kind, entity_id)
);

-- Fast queries for blocking checks and governance dashboards.
CREATE INDEX IF NOT EXISTS idx_fiv_open_critical
    ON finance_integrity_violations (tenant_id, detected_at)
    WHERE lifecycle IN ('OPEN', 'ACKNOWLEDGED') AND severity = 'CRITICAL';

CREATE INDEX IF NOT EXISTS idx_fiv_open_all
    ON finance_integrity_violations (tenant_id, severity, lifecycle)
    WHERE lifecycle IN ('OPEN', 'ACKNOWLEDGED');


-- ── Audit Hash Chain ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_audit_chain (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    outbox_id   UUID NOT NULL,             -- FK to finance_audit_outbox (soft reference)
    sequence    BIGINT NOT NULL,           -- Monotonically increasing per tenant
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL,            -- JSON-encoded audit event request
    prev_hash   TEXT NOT NULL DEFAULT '', -- ChainHash of previous entry; empty for first
    chain_hash  TEXT NOT NULL,            -- SHA256(prevHash||eventType||payload||createdAt)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Idempotency: one chain entry per outbox delivery.
    CONSTRAINT uq_finance_audit_chain_outbox UNIQUE (outbox_id),
    -- Monotonic uniqueness per tenant.
    CONSTRAINT uq_finance_audit_chain_sequence UNIQUE (tenant_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_fac_tenant_seq
    ON finance_audit_chain (tenant_id, sequence ASC);
