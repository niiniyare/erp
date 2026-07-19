-- Finance module — journal tables: journal master, journal entry, lines, ledger.
-- Depends on: migration 000001 (finance_account, finance_accounting_period, finance_currency).

-- ─── Journal ──────────────────────────────────────────────────────────────────

CREATE TABLE finance_journal (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID         NOT NULL REFERENCES platform_tenant(id),
    name               VARCHAR(255) NOT NULL,
    code               VARCHAR(10)  NOT NULL,
    journal_type       VARCHAR(20)  NOT NULL
                           CHECK (journal_type IN ('General', 'Sales', 'Purchase', 'Bank', 'Cash', 'Opening')),
    description        TEXT,
    default_account_id UUID         REFERENCES finance_account(id),
    active             BOOLEAN      NOT NULL DEFAULT true,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

ALTER TABLE finance_journal ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_journal FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_journal
    USING (tenant_id = current_tenant_id());

-- ─── Journal Entry ────────────────────────────────────────────────────────────

CREATE TABLE finance_journal_entry (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL REFERENCES platform_tenant(id),
    entry_number          VARCHAR(50),
    journal_id            UUID          NOT NULL REFERENCES finance_journal(id),
    posting_date          DATE          NOT NULL,
    accounting_period_id  UUID          NOT NULL REFERENCES finance_accounting_period(id),
    status                VARCHAR(20)   NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft', 'submitted', 'posted', 'cancelled', 'reversed')),
    memo                  TEXT,
    reference             VARCHAR(100),
    currency_id           UUID          NOT NULL REFERENCES finance_currency(id),
    exchange_rate         NUMERIC(20,4) NOT NULL DEFAULT 1.0000,
    total_debit           NUMERIC(20,4) NOT NULL DEFAULT 0,
    total_credit          NUMERIC(20,4) NOT NULL DEFAULT 0,
    is_reversal           BOOLEAN       NOT NULL DEFAULT false,
    reversal_of_id        UUID          REFERENCES finance_journal_entry(id),
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (exchange_rate > 0),
    CHECK (total_debit >= 0),
    CHECK (total_credit >= 0)
);

ALTER TABLE finance_journal_entry ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_journal_entry FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_journal_entry
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_journal_entry_date
    ON finance_journal_entry (tenant_id, posting_date);
CREATE INDEX idx_journal_entry_period
    ON finance_journal_entry (tenant_id, accounting_period_id);
CREATE INDEX idx_journal_entry_status
    ON finance_journal_entry (tenant_id, status);
CREATE INDEX idx_journal_entry_journal
    ON finance_journal_entry (tenant_id, journal_id);

-- ─── Journal Entry Line ───────────────────────────────────────────────────────

CREATE TABLE finance_journal_entry_line (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID          NOT NULL REFERENCES platform_tenant(id),
    journal_entry_id UUID          NOT NULL REFERENCES finance_journal_entry(id) ON DELETE CASCADE,
    account_id       UUID          NOT NULL REFERENCES finance_account(id),
    cost_center_id   UUID          REFERENCES finance_cost_center(id),
    debit_amount     NUMERIC(20,4) NOT NULL DEFAULT 0,
    credit_amount    NUMERIC(20,4) NOT NULL DEFAULT 0,
    memo             TEXT,
    reference_type   VARCHAR(100),
    reference_id     VARCHAR(36),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    -- Exactly one side must be zero: debit XOR credit
    CHECK (NOT (debit_amount > 0 AND credit_amount > 0)),
    CHECK (debit_amount >= 0 AND credit_amount >= 0)
);

ALTER TABLE finance_journal_entry_line ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_journal_entry_line FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_journal_entry_line
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_jel_entry
    ON finance_journal_entry_line (tenant_id, journal_entry_id);
CREATE INDEX idx_jel_account
    ON finance_journal_entry_line (tenant_id, account_id);

-- ─── Ledger Entry ─────────────────────────────────────────────────────────────
-- Immutable: populated only by PostingService when a journal entry is posted.
-- No UPDATE or DELETE permissions granted to the app role.

CREATE TABLE finance_ledger_entry (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID          NOT NULL REFERENCES platform_tenant(id),
    entry_number          VARCHAR(50),
    account_id            UUID          NOT NULL REFERENCES finance_account(id),
    journal_entry_id      UUID          NOT NULL REFERENCES finance_journal_entry(id),
    journal_entry_line_id UUID          NOT NULL REFERENCES finance_journal_entry_line(id),
    posting_date          DATE          NOT NULL,
    debit_amount          NUMERIC(20,4) NOT NULL DEFAULT 0,
    credit_amount         NUMERIC(20,4) NOT NULL DEFAULT 0,
    running_balance       NUMERIC(20,4) NOT NULL DEFAULT 0,
    currency_id           UUID          NOT NULL REFERENCES finance_currency(id),
    exchange_rate         NUMERIC(20,4) NOT NULL DEFAULT 1.0000,
    memo                  TEXT,
    cost_center_id        UUID          REFERENCES finance_cost_center(id),
    is_cancelled          BOOLEAN       NOT NULL DEFAULT false,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CHECK (debit_amount >= 0 AND credit_amount >= 0),
    CHECK (exchange_rate > 0)
);

ALTER TABLE finance_ledger_entry ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_ledger_entry FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON finance_ledger_entry
    USING (tenant_id = current_tenant_id());

CREATE INDEX idx_ledger_entry_account
    ON finance_ledger_entry (tenant_id, account_id);
CREATE INDEX idx_ledger_entry_date
    ON finance_ledger_entry (tenant_id, posting_date);
CREATE INDEX idx_ledger_entry_journal
    ON finance_ledger_entry (tenant_id, journal_entry_id);
CREATE INDEX idx_ledger_entry_account_date
    ON finance_ledger_entry (tenant_id, account_id, posting_date);
