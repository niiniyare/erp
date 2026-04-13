-- =====================================================================
-- FINANCE MODULE — APPROVAL WORKFLOW TABLES
-- finance_workflow_records  : one row per transaction approval lifecycle
-- finance_approval_history  : one row per approval decision / action
-- =====================================================================

-- ── finance_workflow_records ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS finance_workflow_records (
    id             UUID        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id      UUID        NOT NULL REFERENCES tenants(id),
    transaction_id UUID        NOT NULL,
    status         TEXT        NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending','in_progress','completed','rejected','cancelled')),
    current_tier   SMALLINT    NOT NULL DEFAULT 1,
    initiated_by   UUID,
    due_at         TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id),
    -- Only one active workflow per transaction per tenant.
    UNIQUE (tenant_id, transaction_id)
);

CREATE INDEX idx_finance_workflow_records_tenant
    ON finance_workflow_records (tenant_id);

CREATE INDEX idx_finance_workflow_records_pending
    ON finance_workflow_records (tenant_id, status)
    WHERE status IN ('pending', 'in_progress');

ALTER TABLE finance_workflow_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_workflow_records FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_workflow_records_app_all ON finance_workflow_records
    FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_workflow_records_ro_select ON finance_workflow_records
    FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());


-- ── finance_approval_history ──────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS finance_approval_history (
    id             UUID        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id      UUID        NOT NULL REFERENCES tenants(id),
    workflow_id    UUID        NOT NULL REFERENCES finance_workflow_records(id),
    transaction_id UUID        NOT NULL,
    tier           SMALLINT    NOT NULL DEFAULT 1,
    action         TEXT        NOT NULL
                       CHECK (action IN ('submitted','approved','rejected','escalated','delegated')),
    performed_by   UUID,
    notes          TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id)
);

CREATE INDEX idx_finance_approval_history_tenant
    ON finance_approval_history (tenant_id);

CREATE INDEX idx_finance_approval_history_transaction
    ON finance_approval_history (tenant_id, transaction_id, created_at);

CREATE INDEX idx_finance_approval_history_workflow
    ON finance_approval_history (workflow_id);

ALTER TABLE finance_approval_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_approval_history FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_approval_history_app_all ON finance_approval_history
    FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_approval_history_ro_select ON finance_approval_history
    FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());
