-- Finance 005: finance_chart_of_accounts and finance_account.
--
-- finance_chart_of_accounts is the COA header (one per tenant typically).
-- finance_account is the hierarchical GL account tree within a COA.
-- finance_cost_center is the analytical dimension tree.

-- ── Chart of Accounts ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_chart_of_accounts (
    id                uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at        timestamptz  NOT NULL DEFAULT NOW(),
    updated_at        timestamptz  NOT NULL DEFAULT NOW(),

    name              varchar(255) NOT NULL,
    code              varchar(50)  NOT NULL,
    description       varchar(1024),
    is_default        boolean      NOT NULL DEFAULT false,
    root_account_type varchar(20)  NOT NULL
                          CHECK (root_account_type IN ('Asset','Liability','Equity','Income','Expense')),

    CONSTRAINT uq_finance_coa_tenant_code UNIQUE (tenant_id, code)
);

COMMENT ON TABLE finance_chart_of_accounts IS
    'COA header. One tenant typically has one COA. '
    'root_account_type classifies the top-level account structure.';

ALTER TABLE finance_chart_of_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_chart_of_accounts FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_chart_of_accounts
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_chart_of_accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── GL Account ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_account (
    id                     uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at             timestamptz  NOT NULL DEFAULT NOW(),
    updated_at             timestamptz  NOT NULL DEFAULT NOW(),

    code                   varchar(20)  NOT NULL,
    name                   varchar(255) NOT NULL,
    account_type           varchar(20)  NOT NULL
                               CHECK (account_type IN ('Asset','Liability','Equity','Income','Expense')),
    account_subtype        varchar(50),
    parent_id              uuid         REFERENCES finance_account(id) ON DELETE RESTRICT,
    chart_of_accounts_id   uuid         NOT NULL REFERENCES finance_chart_of_accounts(id),
    currency_id            uuid         REFERENCES finance_currency(id),
    is_group               boolean      NOT NULL DEFAULT false,
    balance_type           varchar(10)  NOT NULL CHECK (balance_type IN ('Debit','Credit')),
    is_reconcilable        boolean      NOT NULL DEFAULT false,
    active                 boolean      NOT NULL DEFAULT true,
    description            varchar(1024),

    CONSTRAINT uq_finance_account_tenant_code UNIQUE (tenant_id, code)
);

COMMENT ON TABLE finance_account IS
    'GL account in a COA hierarchy. Self-referencing via parent_id. '
    'is_group=true accounts are parent nodes that cannot receive direct postings.';

ALTER TABLE finance_account ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_account FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_account
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_account
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_account_coa    ON finance_account (chart_of_accounts_id, active);
CREATE INDEX IF NOT EXISTS idx_finance_account_parent ON finance_account (parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_finance_account_type   ON finance_account (tenant_id, account_type, active);

-- ── Cost Center ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_cost_center (
    id          uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at  timestamptz  NOT NULL DEFAULT NOW(),
    updated_at  timestamptz  NOT NULL DEFAULT NOW(),

    name        varchar(255) NOT NULL,
    code        varchar(20)  NOT NULL,
    parent_id   uuid         REFERENCES finance_cost_center(id) ON DELETE RESTRICT,
    is_group    boolean      NOT NULL DEFAULT false,
    active      boolean      NOT NULL DEFAULT true,
    description varchar(1024),

    CONSTRAINT uq_finance_cost_center_tenant_code UNIQUE (tenant_id, code)
);

COMMENT ON TABLE finance_cost_center IS
    'Analytical dimension for cost tracking. Self-referencing hierarchy.';

ALTER TABLE finance_cost_center ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_cost_center FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_cost_center
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_cost_center
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_cost_center_parent ON finance_cost_center (parent_id) WHERE parent_id IS NOT NULL;
