CREATE TABLE finance_fiscal_years (
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID        REFERENCES users(id),
    updated_by      UUID        REFERENCES users(id),

    name            TEXT        NOT NULL,  -- e.g. "FY2025"
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    is_closed       BOOLEAN     NOT NULL DEFAULT FALSE,
    is_locked       BOOLEAN     NOT NULL DEFAULT FALSE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, name),
    CONSTRAINT chk_fy_dates CHECK (end_date > start_date)
);

CREATE INDEX idx_finance_fiscal_years_tenant ON finance_fiscal_years (tenant_id);
CREATE INDEX idx_finance_fiscal_years_dates  ON finance_fiscal_years (tenant_id, start_date, end_date);

ALTER TABLE finance_fiscal_years ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_fiscal_years FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_fiscal_years_app_all ON finance_fiscal_years FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_fiscal_years_ro_select ON finance_fiscal_years FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());
