-- Finance 004: finance_exchange_rate — point-in-time FX snapshot.
--
-- IMMUTABLE: once created, rate cannot be changed. Create a new record instead.
-- effective_date is the date the rate applies (not a range).

CREATE TABLE IF NOT EXISTS finance_exchange_rate (
    id              uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid           NOT NULL REFERENCES platform_tenant(id),
    created_at      timestamptz    NOT NULL DEFAULT NOW(),
    -- No updated_at: immutable after creation.

    from_currency   uuid           NOT NULL REFERENCES finance_currency(id),
    to_currency     uuid           NOT NULL REFERENCES finance_currency(id),
    rate            numeric(20,10) NOT NULL CHECK (rate > 0),
    effective_date  date           NOT NULL,
    source          varchar(100),  -- e.g. "CBK", "ECB", "manual"

    CONSTRAINT chk_finance_exchange_rate_diff_currencies
        CHECK (from_currency <> to_currency)
);

COMMENT ON TABLE finance_exchange_rate IS
    'Immutable point-in-time FX rate. Create a new record to update a rate. '
    'rate is the multiplier: amount_in_from_currency * rate = amount_in_to_currency.';

ALTER TABLE finance_exchange_rate ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_exchange_rate FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_exchange_rate
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

-- Lookup: find latest rate for a currency pair on a given date.
CREATE INDEX IF NOT EXISTS idx_finance_exchange_rate_pair_date
    ON finance_exchange_rate (tenant_id, from_currency, to_currency, effective_date DESC);
