-- Finance 010: payment infrastructure.
--
-- finance_bank_account  — tenant bank account (maps to GL)
-- finance_payment_method — payment method master
-- finance_payment        — payment record (MANDATORY)
-- finance_allocation     — payment-to-invoice allocation
-- finance_bank_transaction — imported bank statement line (immutable)

-- ── Bank Account ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_bank_account (
    id             uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at     timestamptz   NOT NULL DEFAULT NOW(),
    updated_at     timestamptz   NOT NULL DEFAULT NOW(),

    account_name   varchar(255)  NOT NULL,
    bank_name      varchar(255)  NOT NULL,
    -- Sensitive: excluded from logs and API responses.
    account_number varchar(50)   NOT NULL,
    swift_code     varchar(11),
    branch_code    varchar(20),
    currency_id    uuid          NOT NULL REFERENCES finance_currency(id),
    gl_account_id  uuid          NOT NULL REFERENCES finance_account(id),
    active         boolean       NOT NULL DEFAULT true,
    balance        numeric(20,4) NOT NULL DEFAULT 0,

    CONSTRAINT uq_finance_bank_account_tenant_num UNIQUE (tenant_id, account_number)
);

ALTER TABLE finance_bank_account ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_account FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_bank_account
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_bank_account
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── Payment Method ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_payment_method (
    id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at      timestamptz  NOT NULL DEFAULT NOW(),
    updated_at      timestamptz  NOT NULL DEFAULT NOW(),

    name            varchar(100) NOT NULL,
    code            varchar(20)  NOT NULL,
    payment_type    varchar(20)  NOT NULL
                        CHECK (payment_type IN ('cash','bank_transfer','cheque','mobile_money','card','other')),
    bank_account_id uuid         REFERENCES finance_bank_account(id) ON DELETE SET NULL,
    gl_account_id   uuid         REFERENCES finance_account(id) ON DELETE SET NULL,
    active          boolean      NOT NULL DEFAULT true,

    CONSTRAINT uq_finance_payment_method_tenant_code UNIQUE (tenant_id, code)
);

ALTER TABLE finance_payment_method ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_payment_method FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_payment_method
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_payment_method
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── Payment ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_payment (
    id                   uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at           timestamptz   NOT NULL DEFAULT NOW(),
    updated_at           timestamptz   NOT NULL DEFAULT NOW(),

    -- NamingSeries: PAY-{YYYY}-{SEQ:6}, tenant-overridable prefix.
    payment_number       varchar(100),
    payment_type         varchar(20)   NOT NULL CHECK (payment_type IN ('receive','send','internal')),
    payment_date         date          NOT NULL,
    payment_method_id    uuid          NOT NULL REFERENCES finance_payment_method(id),
    party_type           varchar(20)   CHECK (party_type IN ('customer','supplier','employee','other')),
    party_id             varchar(36),
    party_name           varchar(255),
    amount               numeric(20,4) NOT NULL CHECK (amount > 0),
    currency_id          uuid          NOT NULL REFERENCES finance_currency(id),
    exchange_rate        numeric(20,10) NOT NULL DEFAULT 1.0,
    reference            varchar(100),
    status               varchar(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft','submitted','processed','cancelled','reconciled')),
    bank_account_id      uuid          REFERENCES finance_bank_account(id) ON DELETE SET NULL,
    gl_account_id        uuid          NOT NULL REFERENCES finance_account(id),
    notes                varchar(1024),
    accounting_period_id uuid          NOT NULL REFERENCES finance_accounting_period(id),
    -- Set by framework when payment is processed (journal entry created).
    journal_entry_id     uuid          REFERENCES finance_journal_entry(id) ON DELETE SET NULL
);

COMMENT ON TABLE finance_payment IS
    'Payment record (MANDATORY system entity). Covers both incoming (receive) '
    'and outgoing (send) payments. journal_entry_id set by framework on processing.';

ALTER TABLE finance_payment ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_payment FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_payment
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_payment
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_payment_status ON finance_payment (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_finance_payment_date   ON finance_payment (tenant_id, payment_date DESC);

-- ── Allocation ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_allocation (
    id               uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at       timestamptz   NOT NULL DEFAULT NOW(),
    -- No updated_at: immutable after creation.

    payment_id       uuid          NOT NULL REFERENCES finance_payment(id) ON DELETE CASCADE,
    -- invoice_id references finance_invoice which is created in a later migration.
    -- Declared as uuid without FK to avoid circular dependency at migration time.
    -- Application layer enforces referential integrity.
    invoice_id       uuid          NOT NULL,
    amount_allocated numeric(20,4) NOT NULL CHECK (amount_allocated > 0),
    currency_id      uuid          NOT NULL REFERENCES finance_currency(id),
    allocation_date  date          NOT NULL
);

COMMENT ON TABLE finance_allocation IS
    'Payment-to-invoice allocation. Many allocations per payment. '
    'invoice_id FK enforced at application layer (avoids circular migration dependency).';

ALTER TABLE finance_allocation ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_allocation FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_allocation
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS idx_finance_allocation_payment ON finance_allocation (payment_id);
CREATE INDEX IF NOT EXISTS idx_finance_allocation_invoice ON finance_allocation (invoice_id);

-- ── Bank Transaction ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_bank_transaction (
    id                 uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at         timestamptz   NOT NULL DEFAULT NOW(),
    -- No updated_at: amount, date, description are immutable after import.

    -- NamingSeries: BT-{YYYY}-{SEQ:6}
    transaction_number varchar(100),
    bank_account_id    uuid          NOT NULL REFERENCES finance_bank_account(id),
    transaction_date   date          NOT NULL,
    value_date         date,
    transaction_type   varchar(10)   NOT NULL CHECK (transaction_type IN ('credit','debit')),
    amount             numeric(20,4) NOT NULL CHECK (amount > 0),
    currency_id        uuid          NOT NULL REFERENCES finance_currency(id),
    description        varchar(255),
    reference          varchar(100),
    status             varchar(20)   NOT NULL DEFAULT 'unreconciled'
                           CHECK (status IN ('unreconciled','matched','reconciled','ignored')),
    payment_id         uuid          REFERENCES finance_payment(id) ON DELETE SET NULL,
    journal_entry_id   uuid          REFERENCES finance_journal_entry(id) ON DELETE SET NULL
);

COMMENT ON TABLE finance_bank_transaction IS
    'Imported bank statement line. Amount/date/description are immutable after import. '
    'status tracks reconciliation progress against payments and journal entries.';

ALTER TABLE finance_bank_transaction ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_transaction FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_bank_transaction
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS idx_finance_bt_account_date
    ON finance_bank_transaction (bank_account_id, transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_finance_bt_status
    ON finance_bank_transaction (tenant_id, status);
