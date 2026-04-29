CREATE TABLE finance_accounting_periods (
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    fiscal_year_id  UUID        NOT NULL REFERENCES finance_fiscal_years(id) ON DELETE CASCADE,
    period_number   INT         NOT NULL CHECK (period_number >= 1),
    name            TEXT        NOT NULL,  -- e.g. "Jan 2025"
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'OPEN'
                        CHECK (status IN ('OPEN', 'SOFT_CLOSED', 'HARD_CLOSED', 'LOCKED')),
    closed_at       TIMESTAMPTZ,
    closed_by       UUID        REFERENCES users(id),
    locked_at       TIMESTAMPTZ,
    locked_by       UUID        REFERENCES users(id),

    PRIMARY KEY (id),
    -- NOTE: UNIQUE(tenant_id, id) removed — id is already globally unique (PK);
    -- the composite unique added index bloat with zero correctness benefit.
    UNIQUE (tenant_id, fiscal_year_id, period_number),
    UNIQUE (tenant_id, fiscal_year_id, start_date),
    CONSTRAINT chk_period_dates CHECK (end_date > start_date)
);

CREATE INDEX idx_finance_periods_tenant      ON finance_accounting_periods (tenant_id);
CREATE INDEX idx_finance_periods_fiscal_year ON finance_accounting_periods (tenant_id, fiscal_year_id);
CREATE INDEX idx_finance_periods_date        ON finance_accounting_periods (tenant_id, start_date, end_date);
CREATE INDEX idx_finance_periods_open        ON finance_accounting_periods (tenant_id, status)
    WHERE status = 'OPEN';

ALTER TABLE finance_accounting_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_accounting_periods FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_accounting_periods_app_all ON finance_accounting_periods FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_accounting_periods_ro_select ON finance_accounting_periods FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());

-- admin_role: full access without tenant restriction (needed for period close/reopen operations)
CREATE POLICY finance_accounting_periods_admin_all ON finance_accounting_periods FOR ALL TO admin_role
    USING (TRUE) WITH CHECK (TRUE);

-- Currencies table (ISO 4217 currencies supported by the tenant)
CREATE TABLE finance_currencies (
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      UUID        REFERENCES users(id),
    updated_by      UUID        REFERENCES users(id),

    code            CHAR(3)     NOT NULL,   -- ISO 4217 e.g. "KES"
    name            TEXT        NOT NULL,   -- e.g. "Kenyan Shilling"
    symbol          TEXT        NOT NULL,   -- e.g. "KSh"
    decimal_places  INT         NOT NULL DEFAULT 2 CHECK (decimal_places BETWEEN 0 AND 4),
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    is_base         BOOLEAN     NOT NULL DEFAULT FALSE,  -- functional/base currency

    PRIMARY KEY (id),
    UNIQUE (tenant_id, code)
);

CREATE INDEX idx_finance_currencies_tenant ON finance_currencies (tenant_id);

ALTER TABLE finance_currencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_currencies FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_currencies_app_all ON finance_currencies FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_currencies_ro_select ON finance_currencies FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());

-- admin_role: full access needed for currency management operations
CREATE POLICY finance_currencies_admin_all ON finance_currencies FOR ALL TO admin_role
    USING (TRUE) WITH CHECK (TRUE);
