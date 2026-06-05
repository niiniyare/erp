-- =====================================================
-- CONTRACTS MODULE SCHEMA
-- =====================================================

CREATE TABLE IF NOT EXISTS contracts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES tenants(id),
    entity_id       UUID        NOT NULL REFERENCES entities(uuid),
    number          TEXT        NOT NULL,
    title           TEXT        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','PENDING_APPROVAL','ACTIVE','EXPIRED','TERMINATED')),
    contract_type   TEXT        NOT NULL DEFAULT 'VENDOR'
                        CHECK (contract_type IN ('VENDOR','CUSTOMER','EMPLOYEE','SERVICE','LEASE','OTHER')),
    counterparty_name  TEXT     NOT NULL,
    counterparty_email TEXT,
    start_date      DATE        NOT NULL,
    end_date        DATE,
    value           NUMERIC(20,6) NOT NULL DEFAULT 0,
    currency_code   CHAR(3)     NOT NULL DEFAULT 'USD',
    description     TEXT,
    terms           TEXT,
    signed_by       UUID,
    signed_at       TIMESTAMPTZ,
    metadata        JSONB       NOT NULL DEFAULT '{}',
    version         INTEGER     NOT NULL DEFAULT 1,
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (tenant_id, number)
);

CREATE INDEX IF NOT EXISTS idx_contracts_tenant_id    ON contracts (tenant_id);
CREATE INDEX IF NOT EXISTS idx_contracts_entity_id    ON contracts (entity_id);
CREATE INDEX IF NOT EXISTS idx_contracts_status       ON contracts (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contracts_number       ON contracts (tenant_id, number);
CREATE INDEX IF NOT EXISTS idx_contracts_end_date     ON contracts (tenant_id, end_date) WHERE deleted_at IS NULL;

-- RLS
ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;

CREATE POLICY contracts_tenant_isolation ON contracts
    USING (tenant_id = current_tenant_id());
