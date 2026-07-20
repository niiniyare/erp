-- Finance 002: finance_accounting_period — monthly/quarterly period within a fiscal year.
--
-- Journal entries require an open period. Closing a period prevents new postings.
-- Locked periods are permanently sealed.

CREATE TABLE IF NOT EXISTS finance_accounting_period (
    id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at      timestamptz  NOT NULL DEFAULT NOW(),
    updated_at      timestamptz  NOT NULL DEFAULT NOW(),

    fiscal_year_id  uuid         NOT NULL REFERENCES finance_fiscal_year(id),
    name            varchar(100) NOT NULL,
    period_number   int          NOT NULL,
    start_date      date         NOT NULL,
    end_date        date         NOT NULL,
    status          varchar(20)  NOT NULL DEFAULT 'open'
                        CHECK (status IN ('open', 'closed', 'locked')),
    is_adjustment   boolean      NOT NULL DEFAULT false,

    CONSTRAINT uq_finance_period_fy_num UNIQUE (fiscal_year_id, period_number),
    CONSTRAINT chk_finance_period_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_finance_period_num   CHECK (period_number BETWEEN 1 AND 16)
);

COMMENT ON TABLE finance_accounting_period IS
    'Monthly or quarterly accounting period. Journal entries require an open period. '
    'is_adjustment marks the 13th period (year-end adjustments).';

ALTER TABLE finance_accounting_period ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_accounting_period FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_accounting_period
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_accounting_period
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_period_fy     ON finance_accounting_period (fiscal_year_id, status);
CREATE INDEX IF NOT EXISTS idx_finance_period_dates  ON finance_accounting_period (tenant_id, start_date, end_date);
