-- Finance 006: finance_tax_group and finance_tax.

CREATE TABLE IF NOT EXISTS finance_tax_group (
    id          uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at  timestamptz  NOT NULL DEFAULT NOW(),
    updated_at  timestamptz  NOT NULL DEFAULT NOW(),

    name        varchar(100) NOT NULL,
    description varchar(1024),
    active      boolean      NOT NULL DEFAULT true,

    CONSTRAINT uq_finance_tax_group_tenant_name UNIQUE (tenant_id, name)
);

ALTER TABLE finance_tax_group ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_group FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_tax_group
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_tax_group
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_tax (
    id           uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid           NOT NULL REFERENCES platform_tenant(id),
    created_at   timestamptz    NOT NULL DEFAULT NOW(),
    updated_at   timestamptz    NOT NULL DEFAULT NOW(),

    name         varchar(255)   NOT NULL,
    code         varchar(20)    NOT NULL,
    tax_group_id uuid           REFERENCES finance_tax_group(id) ON DELETE SET NULL,
    tax_type     varchar(20)    NOT NULL DEFAULT 'percentage'
                     CHECK (tax_type IN ('percentage', 'fixed', 'compound')),
    -- For percentage: 16.0000 = 16%. For fixed: amount in base currency.
    rate         numeric(20,4)  NOT NULL CHECK (rate >= 0),
    account_id   uuid           NOT NULL REFERENCES finance_account(id),
    is_inclusive boolean        NOT NULL DEFAULT false,
    applies_to   varchar(20)    NOT NULL DEFAULT 'both'
                     CHECK (applies_to IN ('sales', 'purchase', 'both')),
    active       boolean        NOT NULL DEFAULT true,
    description  varchar(1024),

    CONSTRAINT uq_finance_tax_tenant_code UNIQUE (tenant_id, code)
);

COMMENT ON TABLE finance_tax IS
    'Individual tax rate. rate=16.0000 means 16% for percentage type, '
    'or KES 100.00 for fixed type. account_id is the GL account for tax postings.';

ALTER TABLE finance_tax ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_tax
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_tax
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_tax_applies ON finance_tax (tenant_id, applies_to, active);
