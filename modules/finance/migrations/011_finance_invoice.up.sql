-- Finance 011: finance_invoice, finance_invoice_line, finance_credit_note,
--              finance_debit_note, finance_receipt, finance_tax_entry.

-- ── Invoice ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_invoice (
    id                   uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at           timestamptz   NOT NULL DEFAULT NOW(),
    updated_at           timestamptz   NOT NULL DEFAULT NOW(),

    -- NamingSeries: INV-{YYYY}-{SEQ:6}, tenant-overridable prefix.
    invoice_number       varchar(100),
    -- customer_id is a plain text reference to an external CRM/party record.
    customer_id          varchar(36)   NOT NULL,
    customer_name        varchar(255)  NOT NULL,
    invoice_date         date          NOT NULL,
    due_date             date,
    status               varchar(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft','submitted','approved','sent','partial','paid','overdue','cancelled')),
    currency_id          uuid          NOT NULL REFERENCES finance_currency(id),
    exchange_rate        numeric(20,10) NOT NULL DEFAULT 1.0,
    subtotal             numeric(20,4) NOT NULL DEFAULT 0,
    tax_amount           numeric(20,4) NOT NULL DEFAULT 0,
    total                numeric(20,4) NOT NULL DEFAULT 0,
    amount_paid          numeric(20,4) NOT NULL DEFAULT 0,
    amount_outstanding   numeric(20,4) NOT NULL DEFAULT 0,
    payment_terms        varchar(100),
    notes                varchar(1024),
    accounting_period_id uuid          NOT NULL REFERENCES finance_accounting_period(id),
    -- Set by framework on approval.
    journal_entry_id     uuid          REFERENCES finance_journal_entry(id) ON DELETE SET NULL,
    -- org_id for multi-org tenants (optional).
    organization_id      varchar(36)
);

COMMENT ON TABLE finance_invoice IS
    'Customer invoice. Lifecycle: draft→submitted→approved→sent→partial|paid|overdue|cancelled. '
    'journal_entry_id set by framework when invoice is approved and posted.';

ALTER TABLE finance_invoice ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_invoice FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_invoice
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_invoice
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_invoice_status     ON finance_invoice (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_finance_invoice_customer   ON finance_invoice (tenant_id, customer_id);
CREATE INDEX IF NOT EXISTS idx_finance_invoice_date       ON finance_invoice (tenant_id, invoice_date DESC);
CREATE INDEX IF NOT EXISTS idx_finance_invoice_due        ON finance_invoice (tenant_id, due_date)
    WHERE status NOT IN ('paid','cancelled');

-- ── Invoice Lines ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_invoice_line (
    id               uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at       timestamptz   NOT NULL DEFAULT NOW(),
    updated_at       timestamptz   NOT NULL DEFAULT NOW(),

    invoice_id       uuid          NOT NULL REFERENCES finance_invoice(id) ON DELETE CASCADE,
    description      varchar(255)  NOT NULL,
    quantity         numeric(20,4) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price       numeric(20,4) NOT NULL CHECK (unit_price >= 0),
    tax_id           uuid          REFERENCES finance_tax(id) ON DELETE SET NULL,
    tax_amount       numeric(20,4) NOT NULL DEFAULT 0,
    line_total       numeric(20,4) NOT NULL DEFAULT 0,
    account_id       uuid          REFERENCES finance_account(id) ON DELETE SET NULL,
    cost_center_id   uuid          REFERENCES finance_cost_center(id) ON DELETE SET NULL,
    sort_order       int           NOT NULL DEFAULT 0
);

ALTER TABLE finance_invoice_line ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_invoice_line FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_invoice_line
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_invoice_line
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_invoice_line_invoice ON finance_invoice_line (invoice_id, sort_order);

-- ── Credit Note ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_credit_note (
    id                   uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at           timestamptz   NOT NULL DEFAULT NOW(),
    updated_at           timestamptz   NOT NULL DEFAULT NOW(),

    -- NamingSeries: CN-{YYYY}-{SEQ:5}
    credit_number        varchar(100),
    original_invoice_id  uuid          NOT NULL REFERENCES finance_invoice(id),
    reason               varchar(1024) NOT NULL,
    issue_date           date          NOT NULL,
    status               varchar(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft','approved','applied','cancelled')),
    currency_id          uuid          NOT NULL REFERENCES finance_currency(id),
    total_amount         numeric(20,4) NOT NULL CHECK (total_amount > 0),
    accounting_period_id uuid          NOT NULL REFERENCES finance_accounting_period(id)
);

ALTER TABLE finance_credit_note ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_credit_note FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_credit_note
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_credit_note
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── Debit Note ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_debit_note (
    id                   uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at           timestamptz   NOT NULL DEFAULT NOW(),
    updated_at           timestamptz   NOT NULL DEFAULT NOW(),

    -- NamingSeries: DN-{YYYY}-{SEQ:5}
    debit_number         varchar(100),
    original_invoice_id  uuid          REFERENCES finance_invoice(id) ON DELETE SET NULL,
    reason               varchar(1024) NOT NULL,
    issue_date           date          NOT NULL,
    status               varchar(20)   NOT NULL DEFAULT 'draft'
                             CHECK (status IN ('draft','approved','applied','cancelled')),
    currency_id          uuid          NOT NULL REFERENCES finance_currency(id),
    total_amount         numeric(20,4) NOT NULL CHECK (total_amount > 0),
    accounting_period_id uuid          NOT NULL REFERENCES finance_accounting_period(id)
);

ALTER TABLE finance_debit_note ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_debit_note FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_debit_note
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_debit_note
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── Receipt ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_receipt (
    id           uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at   timestamptz   NOT NULL DEFAULT NOW(),
    -- No updated_at: immutable after issuance.

    -- NamingSeries: RCT-{YYYY}-{SEQ:5}
    receipt_number  varchar(100),
    payment_id      uuid          NOT NULL REFERENCES finance_payment(id),
    invoice_id      uuid          REFERENCES finance_invoice(id) ON DELETE SET NULL,
    receipt_date    date          NOT NULL,
    amount          numeric(20,4) NOT NULL CHECK (amount > 0),
    currency_id     uuid          NOT NULL REFERENCES finance_currency(id),
    notes           varchar(1024)
);

COMMENT ON TABLE finance_receipt IS
    'Payment receipt (immutable after creation). Issued when payment is processed.';

ALTER TABLE finance_receipt ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_receipt FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_receipt
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS idx_finance_receipt_payment ON finance_receipt (payment_id);

-- ── Tax Entry ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_tax_entry (
    id                uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         uuid          NOT NULL REFERENCES platform_tenant(id),
    created_at        timestamptz   NOT NULL DEFAULT NOW(),
    -- No updated_at: immutable (framework-written only).

    -- NamingSeries: TAX-{YYYY}-{SEQ:6}
    entry_number      varchar(100),
    tax_id            uuid          NOT NULL REFERENCES finance_tax(id),
    journal_entry_id  uuid          NOT NULL REFERENCES finance_journal_entry(id),
    reference_type    varchar(100),
    reference_id      varchar(36),
    tax_base_amount   numeric(20,4) NOT NULL CHECK (tax_base_amount >= 0),
    tax_amount        numeric(20,4) NOT NULL CHECK (tax_amount >= 0),
    posting_date      date          NOT NULL,
    is_reverse_charge boolean       NOT NULL DEFAULT false
);

COMMENT ON TABLE finance_tax_entry IS
    'Immutable tax record for KRA eTIMS compliance (MANDATORY system entity). '
    'Written exclusively by the framework. API cannot Create/Write/Delete.';

ALTER TABLE finance_tax_entry ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_tax_entry FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_tax_entry
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS idx_finance_tax_entry_date ON finance_tax_entry (tenant_id, posting_date DESC);
CREATE INDEX IF NOT EXISTS idx_finance_tax_entry_je   ON finance_tax_entry (journal_entry_id);

-- ── FK: finance_allocation → finance_invoice ──────────────────────────────────
-- Now that finance_invoice exists, add the FK constraint deferred from step 010.
ALTER TABLE finance_allocation
    ADD CONSTRAINT fk_finance_allocation_invoice
    FOREIGN KEY (invoice_id) REFERENCES finance_invoice(id) ON DELETE RESTRICT;
