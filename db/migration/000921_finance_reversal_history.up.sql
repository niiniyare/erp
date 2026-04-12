-- =====================================================================
-- FINANCE MODULE — REVERSAL HISTORY TABLE
-- One row per reversal event, linked to both original and reversal
-- transactions. Used by the double-reversal guard and audit UI.
-- =====================================================================

CREATE TABLE IF NOT EXISTS finance_reversal_history (
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    original_transaction_id UUID        NOT NULL,
    reversal_transaction_id UUID        NOT NULL,
    reason                  TEXT        NOT NULL DEFAULT '',
    reversed_by             UUID,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id),
    -- A transaction may only be reversed once per tenant.
    UNIQUE (tenant_id, original_transaction_id)
);

CREATE INDEX idx_finance_reversal_history_tenant
    ON finance_reversal_history (tenant_id);

CREATE INDEX idx_finance_reversal_history_original
    ON finance_reversal_history (tenant_id, original_transaction_id);

CREATE INDEX idx_finance_reversal_history_reversal
    ON finance_reversal_history (tenant_id, reversal_transaction_id);

-- Row-level security: each tenant sees only its own records.
ALTER TABLE finance_reversal_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_reversal_history FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_reversal_history_app_all ON finance_reversal_history
    FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_reversal_history_ro_select ON finance_reversal_history
    FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());
