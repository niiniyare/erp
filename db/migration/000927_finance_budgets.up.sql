-- Finance Budgets
-- Budget header + line item tables with approval workflow support.

CREATE TABLE IF NOT EXISTS finance_budgets (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID         NOT NULL,
    fiscal_year_id      UUID         NOT NULL REFERENCES finance_fiscal_years(id) ON DELETE RESTRICT,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    budget_type         VARCHAR(30)  NOT NULL CHECK (budget_type IN ('ANNUAL','QUARTERLY','PROJECT','DEPARTMENT','CAPEX')),
    status              VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED','REVISED','CLOSED')),
    currency_code       CHAR(3)      NOT NULL DEFAULT 'USD',
    cost_center_id      UUID         REFERENCES finance_cost_centers(id) ON DELETE SET NULL,

    -- Approval metadata
    submitted_at        TIMESTAMPTZ,
    submitted_by        UUID,
    approved_at         TIMESTAMPTZ,
    approved_by         UUID,
    rejected_at         TIMESTAMPTZ,
    rejected_by         UUID,
    reject_note         TEXT,

    -- Revision tracking
    version             INT          NOT NULL DEFAULT 1,
    original_budget_id  UUID         REFERENCES finance_budgets(id) ON DELETE SET NULL,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by          UUID,
    updated_by          UUID
);

CREATE INDEX idx_finance_budgets_tenant        ON finance_budgets(tenant_id);
CREATE INDEX idx_finance_budgets_fiscal_year   ON finance_budgets(fiscal_year_id);
CREATE INDEX idx_finance_budgets_status        ON finance_budgets(tenant_id, status);
CREATE INDEX idx_finance_budgets_cost_center   ON finance_budgets(cost_center_id);

ALTER TABLE finance_budgets ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_budgets_tenant_isolation
    ON finance_budgets
    USING (tenant_id = current_tenant_id());

-- Budget line items (one row per account / cost-centre / period combination)
CREATE TABLE IF NOT EXISTS finance_budget_line_items (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id       UUID         NOT NULL REFERENCES finance_budgets(id) ON DELETE CASCADE,
    tenant_id       UUID         NOT NULL,
    account_id      UUID         NOT NULL,
    cost_center_id  UUID         REFERENCES finance_cost_centers(id) ON DELETE SET NULL,
    period_id       UUID         REFERENCES finance_accounting_periods(id) ON DELETE SET NULL,

    budgeted_amount NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (budgeted_amount >= 0),
    actual_amount   NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes           TEXT,

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by      UUID,
    updated_by      UUID,

    CONSTRAINT uq_budget_line UNIQUE (budget_id, account_id, cost_center_id, period_id)
);

CREATE INDEX idx_finance_budget_lines_budget      ON finance_budget_line_items(budget_id);
CREATE INDEX idx_finance_budget_lines_tenant      ON finance_budget_line_items(tenant_id);
CREATE INDEX idx_finance_budget_lines_account     ON finance_budget_line_items(account_id);
CREATE INDEX idx_finance_budget_lines_cost_center ON finance_budget_line_items(cost_center_id);

ALTER TABLE finance_budget_line_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_budget_lines_tenant_isolation
    ON finance_budget_line_items
    USING (tenant_id = current_tenant_id());
