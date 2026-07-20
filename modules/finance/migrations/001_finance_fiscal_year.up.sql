-- Finance 001: finance_fiscal_year — fiscal year master.
--
-- Lifecycle: open → closed → locked
-- Closed years may be re-opened by admin action.
-- Locked years are permanently sealed (no further journal entries).

CREATE TABLE IF NOT EXISTS finance_fiscal_year (
    id          uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at  timestamptz  NOT NULL DEFAULT NOW(),
    updated_at  timestamptz  NOT NULL DEFAULT NOW(),

    name        varchar(100) NOT NULL,
    start_date  date         NOT NULL,
    end_date    date         NOT NULL,
    status      varchar(20)  NOT NULL DEFAULT 'open'
                    CHECK (status IN ('open', 'closed', 'locked')),
    is_current  boolean      NOT NULL DEFAULT false,

    CONSTRAINT uq_finance_fiscal_year_tenant_name UNIQUE (tenant_id, name),
    CONSTRAINT chk_finance_fiscal_year_dates CHECK (end_date > start_date)
);

COMMENT ON TABLE finance_fiscal_year IS
    'Fiscal year master. Lifecycle: open → closed → locked. '
    'Journal entries can only be posted to open accounting periods within an open fiscal year.';

ALTER TABLE finance_fiscal_year ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_fiscal_year FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_fiscal_year
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_fiscal_year
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_fiscal_year_status ON finance_fiscal_year (tenant_id, status);
