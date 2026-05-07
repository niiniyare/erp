-- Migration 001005: Finance audit outbox for durable CRITICAL audit delivery
--
-- Context:
--   Phase 7 hardens audit reliability. CRITICAL finance events (post, approve,
--   reject, reverse, period-close, reconciliation-complete) must never be silently
--   lost. This outbox table provides the durable guarantee: the audit record is
--   written in the same DB transaction as the financial mutation and delivered to
--   audit.Service by a background worker.
--
-- Delivery states:
--   PENDING    → written, awaiting worker pickup
--   PROCESSING → worker picked up, delivery in flight
--   DELIVERED  → audit.Service confirmed write
--   DEAD       → max retries exceeded, requires manual intervention / alerting

CREATE TABLE IF NOT EXISTS finance_audit_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,

    -- idempotency_key prevents duplicate audit events on Temporal retries.
    -- Format: "{event_type}:{primary_resource_id}[:{qualifier}]"
    idempotency_key TEXT NOT NULL,

    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL, -- serialised audit.CreateAuditEventRequest

    status          TEXT NOT NULL DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING','PROCESSING','DELIVERED','DEAD')),
    retry_count     INT  NOT NULL DEFAULT 0,
    max_retries     INT  NOT NULL DEFAULT 5,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    error_msg       TEXT,

    -- Tenant-scoped idempotency: same event_type+resource cannot appear twice per tenant.
    CONSTRAINT uq_finance_audit_outbox_idempotency UNIQUE (tenant_id, idempotency_key)
);

-- Worker pickup index: PENDING entries ordered by created_at for fair delivery.
CREATE INDEX IF NOT EXISTS idx_finance_audit_outbox_pending
    ON finance_audit_outbox (tenant_id, created_at)
    WHERE status = 'PENDING';

-- Dead-letter index for alerting and gap detection.
CREATE INDEX IF NOT EXISTS idx_finance_audit_outbox_dead
    ON finance_audit_outbox (tenant_id, created_at)
    WHERE status = 'DEAD';
