-- Finance module — core tables: currency, exchange rate, fiscal periods, COA, accounts, cost centers.
-- All tables are tenant-isolated with RLS enforced via current_tenant_id().

-- ─── Currency ─────────────────────────────────────────────────────────────────

CREATE TABLE finance_currency (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID        NOT NULL REFERENCES platform_tenant(id),
    code           VARCHAR(3)  NOT NULL,
    name           VARCHAR(100) NOT NULL,
    symbol         VARCHAR(10),
    decimal_places INT         NOT NULL DEFAULT 2,
    active         BOOLEAN     NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_currency ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_currency FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_currency
    USING (tenant_id = current_tenant_id());

-- ─── Exchange Rate ────────────────────────────────────────────────────────────

CREATE TABLE finance_exchange_rate (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID         NOT NULL REFERENCES platform_tenant(id),
    from_currency  UUID         NOT NULL REFERENCES finance_currency(id),
    to_currency    UUID         NOT NULL REFERENCES finance_currency(id),
    rate           NUMERIC(20,4) NOT NULL,
    effective_date DATE         NOT NULL,
    source         VARCHAR(100),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (rate > 0),
    CHECK (from_currency <> to_currency)
);

ALTER TABLE finance_exchange_rate ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_exchange_rate FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_exchange_rate
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_exchange_rate_pair_date
    ON finance_exchange_rate (tenant_id, from_currency, to_currency, effective_date DESC);

-- ─── Fiscal Year ──────────────────────────────────────────────────────────────

CREATE TABLE finance_fiscal_year (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID         NOT NULL REFERENCES platform_tenant(id),
    name       VARCHAR(100) NOT NULL,
    start_date DATE         NOT NULL,
    end_date   DATE         NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'open'
                   CHECK (status IN ('open', 'closed', 'locked')),
    is_current BOOLEAN      NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (start_date < end_date),
    UNIQUE (tenant_id, name)
);

ALTER TABLE finance_fiscal_year ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_fiscal_year FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_fiscal_year
    USING (tenant_id = current_tenant_id());

-- At most one current fiscal year per tenant.
CREATE UNIQUE INDEX idx_fiscal_year_current
    ON finance_fiscal_year (tenant_id)
    WHERE is_current = true;

-- ─── Accounting Period ────────────────────────────────────────────────────────

CREATE TABLE finance_accounting_period (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID         NOT NULL REFERENCES platform_tenant(id),
    fiscal_year_id  UUID         NOT NULL REFERENCES finance_fiscal_year(id),
    name            VARCHAR(100) NOT NULL,
    period_number   INT          NOT NULL,
    start_date      DATE         NOT NULL,
    end_date        DATE         NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'open'
                        CHECK (status IN ('open', 'closed', 'locked')),
    is_adjustment   BOOLEAN      NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (start_date < end_date),
    CHECK (period_number BETWEEN 1 AND 13),
    UNIQUE (tenant_id, fiscal_year_id, period_number)
);

ALTER TABLE finance_accounting_period ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_accounting_period FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_accounting_period
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_accounting_period_fy
    ON finance_accounting_period (tenant_id, fiscal_year_id);
CREATE INDEX idx_accounting_period_dates
    ON finance_accounting_period (tenant_id, start_date, end_date);

-- ─── Chart of Accounts ────────────────────────────────────────────────────────

CREATE TABLE finance_chart_of_accounts (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID         NOT NULL REFERENCES platform_tenant(id),
    name              VARCHAR(255) NOT NULL,
    code              VARCHAR(50)  NOT NULL,
    description       TEXT,
    is_default        BOOLEAN      NOT NULL DEFAULT false,
    root_account_type VARCHAR(20)  NOT NULL
                          CHECK (root_account_type IN ('Asset', 'Liability', 'Equity', 'Income', 'Expense')),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_chart_of_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_chart_of_accounts FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_chart_of_accounts
    USING (tenant_id = current_tenant_id());

CREATE UNIQUE INDEX idx_coa_default
    ON finance_chart_of_accounts (tenant_id)
    WHERE is_default = true;

-- ─── Account ──────────────────────────────────────────────────────────────────

CREATE TABLE finance_account (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID         NOT NULL REFERENCES platform_tenant(id),
    code                 VARCHAR(20)  NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    account_type         VARCHAR(20)  NOT NULL
                             CHECK (account_type IN ('Asset', 'Liability', 'Equity', 'Income', 'Expense')),
    account_subtype      VARCHAR(50),
    parent_id            UUID         REFERENCES finance_account(id),
    chart_of_accounts_id UUID         NOT NULL REFERENCES finance_chart_of_accounts(id),
    currency_id          UUID         REFERENCES finance_currency(id),
    is_group             BOOLEAN      NOT NULL DEFAULT false,
    balance_type         VARCHAR(10)  NOT NULL CHECK (balance_type IN ('Debit', 'Credit')),
    is_reconcilable      BOOLEAN      NOT NULL DEFAULT false,
    active               BOOLEAN      NOT NULL DEFAULT true,
    description          TEXT,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (parent_id IS NULL OR parent_id <> id),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_account ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_account FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_account
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_account_parent ON finance_account (tenant_id, parent_id);
CREATE INDEX idx_account_coa ON finance_account (tenant_id, chart_of_accounts_id);
CREATE INDEX idx_account_type ON finance_account (tenant_id, account_type);

-- ─── Cost Center ─────────────────────────────────────────────────────────────

CREATE TABLE finance_cost_center (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID         NOT NULL REFERENCES platform_tenant(id),
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(20)  NOT NULL,
    parent_id   UUID         REFERENCES finance_cost_center(id),
    is_group    BOOLEAN      NOT NULL DEFAULT false,
    active      BOOLEAN      NOT NULL DEFAULT true,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CHECK (parent_id IS NULL OR parent_id <> id),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_cost_center ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_cost_center FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_cost_center
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_cost_center_parent ON finance_cost_center (tenant_id, parent_id);
