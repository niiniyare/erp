CREATE TABLE finance_exchange_rates (
    id             UUID          NOT NULL DEFAULT gen_random_uuid(),
    tenant_id      UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by     UUID          REFERENCES users(id),

    from_currency  CHAR(3)       NOT NULL,
    to_currency    CHAR(3)       NOT NULL,
    rate           NUMERIC(18,8) NOT NULL CHECK (rate > 0),
    rate_type      TEXT          NOT NULL DEFAULT 'SPOT'
                       CHECK (rate_type IN ('SPOT', 'CLOSING', 'AVERAGE', 'HISTORICAL', 'FIXED', 'OFFICIAL')),
    effective_date DATE          NOT NULL,
    expiry_date    DATE,
    source         TEXT          NOT NULL DEFAULT 'manual',


    PRIMARY KEY (id),
    UNIQUE (tenant_id, from_currency, to_currency, rate_type, effective_date)
);

CREATE INDEX idx_finance_exchange_rates_tenant ON finance_exchange_rates (tenant_id);
CREATE INDEX idx_finance_exchange_rates_pair   ON finance_exchange_rates (tenant_id, from_currency, to_currency);
CREATE INDEX idx_finance_exchange_rates_date   ON finance_exchange_rates (tenant_id, from_currency, to_currency, effective_date DESC);

ALTER TABLE finance_exchange_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_exchange_rates FORCE  ROW LEVEL SECURITY;

CREATE POLICY finance_exchange_rates_app_all ON finance_exchange_rates FOR ALL TO application_role
    USING     (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY finance_exchange_rates_ro_select ON finance_exchange_rates FOR SELECT TO readonly_role
    USING (tenant_id = current_tenant_id());
