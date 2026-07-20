-- Finance 003: finance_currency — ISO 4217 currency master.
--
-- One record per active currency per tenant. code is immutable after creation.

CREATE TABLE IF NOT EXISTS finance_currency (
    id             uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at     timestamptz  NOT NULL DEFAULT NOW(),
    updated_at     timestamptz  NOT NULL DEFAULT NOW(),

    code           varchar(3)   NOT NULL,
    name           varchar(100) NOT NULL,
    symbol         varchar(10),
    decimal_places int          NOT NULL DEFAULT 2,
    active         boolean      NOT NULL DEFAULT true,

    CONSTRAINT uq_finance_currency_tenant_code UNIQUE (tenant_id, code)
);

COMMENT ON TABLE finance_currency IS
    'ISO 4217 currency master. code (e.g. "KES") is immutable after creation.';

ALTER TABLE finance_currency ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_currency FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_currency
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_currency
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_currency_active ON finance_currency (tenant_id, active);
