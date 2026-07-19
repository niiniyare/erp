-- Finance module — document tables: tax, payments, invoices, banking.
-- Depends on: migrations 000001 + 000002.

-- ─── Tax Group ────────────────────────────────────────────────────────────────

CREATE TABLE finance_tax_group (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID         NOT NULL REFERENCES platform_tenant(id),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    active      BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

ALTER TABLE finance_tax_group ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_group FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_tax_group
    USING (tenant_id = current_tenant_id());

-- ─── Tax ─────────────────────────────────────────────────────────────────────

CREATE TABLE finance_tax (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID          NOT NULL REFERENCES platform_tenant(id),
    name          VARCHAR(255)  NOT NULL,
    code          VARCHAR(20)   NOT NULL,
    tax_group_id  UUID          REFERENCES finance_tax_group(id),
    tax_type      VARCHAR(20)   NOT NULL DEFAULT 'percentage'
                      CHECK (tax_type IN ('percentage', 'fixed', 'compound')),
    rate          NUMERIC(20,4) NOT NULL,
    account_id    UUID          NOT NULL REFERENCES finance_account(id),
    is_inclusive  BOOLEAN       NOT NULL DEFAULT false,
    applies_to    VARCHAR(20)   NOT NULL DEFAULT 'both'
                      CHECK (applies_to IN ('sales', 'purchase', 'both')),
    active        BOOLEAN       NOT NULL DEFAULT true,
    description   TEXT,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (rate >= 0),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_tax ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_tax
    USING (tenant_id = current_tenant_id());

-- ─── Tax Entry (mandatory) ────────────────────────────────────────────────────

CREATE TABLE finance_tax_entry (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID          NOT NULL REFERENCES platform_tenant(id),
    entry_number      VARCHAR(50),
    tax_id            UUID          NOT NULL REFERENCES finance_tax(id),
    journal_entry_id  UUID          NOT NULL REFERENCES finance_journal_entry(id),
    reference_type    VARCHAR(100),
    reference_id      VARCHAR(36),
    tax_base_amount   NUMERIC(20,4) NOT NULL,
    tax_amount        NUMERIC(20,4) NOT NULL,
    posting_date      DATE          NOT NULL,
    is_reverse_charge BOOLEAN       NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (tax_base_amount >= 0),
    CHECK (tax_amount >= 0)
);

ALTER TABLE finance_tax_entry ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_entry FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_tax_entry
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_tax_entry_tax ON finance_tax_entry (tenant_id, tax_id);
CREATE INDEX idx_tax_entry_date ON finance_tax_entry (tenant_id, posting_date);

-- ─── Bank Account ─────────────────────────────────────────────────────────────

CREATE TABLE finance_bank_account (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID          NOT NULL REFERENCES platform_tenant(id),
    account_name   VARCHAR(255)  NOT NULL,
    bank_name      VARCHAR(255)  NOT NULL,
    account_number VARCHAR(50)   NOT NULL,
    swift_code     VARCHAR(11),
    branch_code    VARCHAR(20),
    currency_id    UUID          NOT NULL REFERENCES finance_currency(id),
    gl_account_id  UUID          NOT NULL REFERENCES finance_account(id),
    active         BOOLEAN       NOT NULL DEFAULT true,
    balance        NUMERIC(20,4) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, account_number)
);

ALTER TABLE finance_bank_account ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_account FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_bank_account
    USING (tenant_id = current_tenant_id());

-- ─── Payment Method ───────────────────────────────────────────────────────────

CREATE TABLE finance_payment_method (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID         NOT NULL REFERENCES platform_tenant(id),
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(20)  NOT NULL,
    payment_type    VARCHAR(30)  NOT NULL
                        CHECK (payment_type IN ('cash', 'bank_transfer', 'cheque', 'mobile_money', 'card', 'other')),
    bank_account_id UUID         REFERENCES finance_bank_account(id),
    gl_account_id   UUID         REFERENCES finance_account(id),
    active          BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_payment_method ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_payment_method FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_payment_method
    USING (tenant_id = current_tenant_id());

-- ─── Payment (mandatory) ──────────────────────────────────────────────────────

CREATE TABLE finance_payment (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID          NOT NULL REFERENCES platform_tenant(id),
    payment_number       VARCHAR(50),
    payment_type         VARCHAR(20)   NOT NULL
                             CHECK (payment_type IN ('receive', 'send', 'internal')),
    payment_date         DATE          NOT NULL,
    payment_method_id    UUID          NOT NULL REFERENCES finance_payment_method(id),
    party_type           VARCHAR(30)   CHECK (party_type IN ('customer', 'supplier', 'employee', 'other')),
    party_id             VARCHAR(36),
    party_name           VARCHAR(255),
    amount               NUMERIC(20,4) NOT NULL,
    currency_id          UUID          NOT NULL REFERENCES finance_currency(id),
    exchange_rate        NUMERIC(20,4) NOT NULL DEFAULT 1.0000,
    reference            VARCHAR(100),
    status               VARCHAR(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft', 'submitted', 'processed', 'cancelled', 'reconciled')),
    bank_account_id      UUID          REFERENCES finance_bank_account(id),
    gl_account_id        UUID          NOT NULL REFERENCES finance_account(id),
    notes                TEXT,
    accounting_period_id UUID          NOT NULL REFERENCES finance_accounting_period(id),
    journal_entry_id     UUID          REFERENCES finance_journal_entry(id),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (amount > 0),
    CHECK (exchange_rate > 0)
);

ALTER TABLE finance_payment ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_payment FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_payment
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_payment_status ON finance_payment (tenant_id, status);
CREATE INDEX idx_payment_date ON finance_payment (tenant_id, payment_date);
CREATE INDEX idx_payment_period ON finance_payment (tenant_id, accounting_period_id);

-- ─── Invoice ──────────────────────────────────────────────────────────────────

CREATE TABLE finance_invoice (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID          NOT NULL REFERENCES platform_tenant(id),
    invoice_number       VARCHAR(50),
    customer_id          VARCHAR(36)   NOT NULL,
    customer_name        VARCHAR(255)  NOT NULL,
    invoice_date         DATE          NOT NULL,
    due_date             DATE,
    status               VARCHAR(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft', 'submitted', 'approved', 'sent', 'partial', 'paid', 'overdue', 'cancelled')),
    currency_id          UUID          NOT NULL REFERENCES finance_currency(id),
    exchange_rate        NUMERIC(20,4) NOT NULL DEFAULT 1.0000,
    subtotal             NUMERIC(20,4) NOT NULL DEFAULT 0,
    tax_amount           NUMERIC(20,4) NOT NULL DEFAULT 0,
    total                NUMERIC(20,4) NOT NULL DEFAULT 0,
    amount_paid          NUMERIC(20,4) NOT NULL DEFAULT 0,
    amount_outstanding   NUMERIC(20,4) NOT NULL DEFAULT 0,
    payment_terms        VARCHAR(100),
    notes                TEXT,
    accounting_period_id UUID          NOT NULL REFERENCES finance_accounting_period(id),
    journal_entry_id     UUID          REFERENCES finance_journal_entry(id),
    organization_id      VARCHAR(36),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (exchange_rate > 0),
    CHECK (subtotal >= 0 AND tax_amount >= 0 AND total >= 0),
    CHECK (amount_paid >= 0 AND amount_outstanding >= 0)
);

ALTER TABLE finance_invoice ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_invoice FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_invoice
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_invoice_customer ON finance_invoice (tenant_id, customer_id);
CREATE INDEX idx_invoice_status ON finance_invoice (tenant_id, status);
CREATE INDEX idx_invoice_period ON finance_invoice (tenant_id, accounting_period_id);
CREATE INDEX idx_invoice_date ON finance_invoice (tenant_id, invoice_date);
CREATE INDEX idx_invoice_org ON finance_invoice (tenant_id, organization_id)
    WHERE organization_id IS NOT NULL;

-- ─── Invoice Line ─────────────────────────────────────────────────────────────

CREATE TABLE finance_invoice_line (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID          NOT NULL REFERENCES platform_tenant(id),
    invoice_id     UUID          NOT NULL REFERENCES finance_invoice(id) ON DELETE CASCADE,
    description    VARCHAR(255)  NOT NULL,
    quantity       NUMERIC(20,4) NOT NULL DEFAULT 1.0000,
    unit_price     NUMERIC(20,4) NOT NULL,
    tax_id         UUID          REFERENCES finance_tax(id),
    tax_amount     NUMERIC(20,4) NOT NULL DEFAULT 0,
    line_total     NUMERIC(20,4) NOT NULL DEFAULT 0,
    account_id     UUID          REFERENCES finance_account(id),
    cost_center_id UUID          REFERENCES finance_cost_center(id),
    sort_order     INT           NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (quantity > 0),
    CHECK (unit_price >= 0),
    CHECK (tax_amount >= 0 AND line_total >= 0)
);

ALTER TABLE finance_invoice_line ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_invoice_line FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_invoice_line
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_invoice_line_invoice ON finance_invoice_line (tenant_id, invoice_id);

-- ─── Allocation ───────────────────────────────────────────────────────────────

CREATE TABLE finance_allocation (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID          NOT NULL REFERENCES platform_tenant(id),
    payment_id       UUID          NOT NULL REFERENCES finance_payment(id),
    invoice_id       UUID          NOT NULL REFERENCES finance_invoice(id),
    amount_allocated NUMERIC(20,4) NOT NULL,
    currency_id      UUID          NOT NULL REFERENCES finance_currency(id),
    allocation_date  DATE          NOT NULL,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (amount_allocated > 0),
    UNIQUE (tenant_id, payment_id, invoice_id)
);

ALTER TABLE finance_allocation ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_allocation FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_allocation
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_allocation_payment ON finance_allocation (tenant_id, payment_id);
CREATE INDEX idx_allocation_invoice ON finance_allocation (tenant_id, invoice_id);

-- ─── Credit Note ──────────────────────────────────────────────────────────────

CREATE TABLE finance_credit_note (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID          NOT NULL REFERENCES platform_tenant(id),
    credit_number        VARCHAR(50),
    original_invoice_id  UUID          REFERENCES finance_invoice(id),
    reason               TEXT          NOT NULL,
    issue_date           DATE          NOT NULL,
    status               VARCHAR(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft', 'approved', 'applied', 'cancelled')),
    currency_id          UUID          NOT NULL REFERENCES finance_currency(id),
    total_amount         NUMERIC(20,4) NOT NULL,
    accounting_period_id UUID          NOT NULL REFERENCES finance_accounting_period(id),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (total_amount >= 0)
);

ALTER TABLE finance_credit_note ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_credit_note FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_credit_note
    USING (tenant_id = current_tenant_id());

-- ─── Debit Note ───────────────────────────────────────────────────────────────

CREATE TABLE finance_debit_note (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID          NOT NULL REFERENCES platform_tenant(id),
    debit_number         VARCHAR(50),
    original_invoice_id  UUID          REFERENCES finance_invoice(id),
    reason               TEXT          NOT NULL,
    issue_date           DATE          NOT NULL,
    status               VARCHAR(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft', 'approved', 'applied', 'cancelled')),
    currency_id          UUID          NOT NULL REFERENCES finance_currency(id),
    total_amount         NUMERIC(20,4) NOT NULL,
    accounting_period_id UUID          NOT NULL REFERENCES finance_accounting_period(id),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (total_amount >= 0)
);

ALTER TABLE finance_debit_note ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_debit_note FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_debit_note
    USING (tenant_id = current_tenant_id());

-- ─── Receipt ──────────────────────────────────────────────────────────────────

CREATE TABLE finance_receipt (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID          NOT NULL REFERENCES platform_tenant(id),
    receipt_number VARCHAR(50),
    payment_id     UUID          NOT NULL REFERENCES finance_payment(id),
    invoice_id     UUID          REFERENCES finance_invoice(id),
    receipt_date   DATE          NOT NULL,
    amount         NUMERIC(20,4) NOT NULL,
    currency_id    UUID          NOT NULL REFERENCES finance_currency(id),
    notes          TEXT,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (amount > 0)
);

ALTER TABLE finance_receipt ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_receipt FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_receipt
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_receipt_payment ON finance_receipt (tenant_id, payment_id);

-- ─── Bank Transaction ─────────────────────────────────────────────────────────

CREATE TABLE finance_bank_transaction (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID          NOT NULL REFERENCES platform_tenant(id),
    transaction_number VARCHAR(50),
    bank_account_id  UUID          NOT NULL REFERENCES finance_bank_account(id),
    transaction_date DATE          NOT NULL,
    value_date       DATE,
    transaction_type VARCHAR(10)   NOT NULL CHECK (transaction_type IN ('credit', 'debit')),
    amount           NUMERIC(20,4) NOT NULL,
    currency_id      UUID          NOT NULL REFERENCES finance_currency(id),
    description      VARCHAR(255),
    reference        VARCHAR(100),
    status           VARCHAR(20)   NOT NULL DEFAULT 'unreconciled'
                         CHECK (status IN ('unreconciled', 'matched', 'reconciled', 'ignored')),
    payment_id       UUID          REFERENCES finance_payment(id),
    journal_entry_id UUID          REFERENCES finance_journal_entry(id),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (amount > 0)
);

ALTER TABLE finance_bank_transaction ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_transaction FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_bank_transaction
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_bank_txn_account ON finance_bank_transaction (tenant_id, bank_account_id);
CREATE INDEX idx_bank_txn_date ON finance_bank_transaction (tenant_id, transaction_date);
CREATE INDEX idx_bank_txn_status ON finance_bank_transaction (tenant_id, status);
