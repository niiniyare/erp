-- Finance 008: finance_journal_entry and finance_journal_entry_line.
--
-- finance_journal_entry is the double-entry header (MANDATORY system entity).
-- Lifecycle: draft → submitted → posted → reversed | cancelled
-- Lines must balance (Σdebit = Σcredit) before posting.

CREATE TABLE IF NOT EXISTS finance_journal_entry (
    id                    uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             uuid           NOT NULL REFERENCES platform_tenant(id),
    created_at            timestamptz    NOT NULL DEFAULT NOW(),
    updated_at            timestamptz    NOT NULL DEFAULT NOW(),

    -- NamingSeries populated by framework AfterCreate hook.
    entry_number          varchar(100),
    journal_id            uuid           NOT NULL REFERENCES finance_journal(id),
    posting_date          date           NOT NULL,
    accounting_period_id  uuid           NOT NULL REFERENCES finance_accounting_period(id),
    status                varchar(20)    NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft','submitted','posted','cancelled','reversed')),
    memo                  varchar(1024),
    reference             varchar(100),
    currency_id           uuid           NOT NULL REFERENCES finance_currency(id),
    exchange_rate         numeric(20,10) NOT NULL DEFAULT 1.0,
    total_debit           numeric(20,4)  NOT NULL DEFAULT 0,
    total_credit          numeric(20,4)  NOT NULL DEFAULT 0,
    is_reversal           boolean        NOT NULL DEFAULT false,
    -- If this entry reverses another, link to the original.
    reversal_of_id        uuid           REFERENCES finance_journal_entry(id) ON DELETE RESTRICT
);

COMMENT ON TABLE finance_journal_entry IS
    'Double-entry journal header (MANDATORY system entity). '
    'Lines must balance (Σdebit = Σcredit) before status can be set to posted. '
    'Once posted, the entry is immutable — only reversal is permitted.';

ALTER TABLE finance_journal_entry ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_journal_entry FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_journal_entry
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_journal_entry
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_je_period  ON finance_journal_entry (accounting_period_id, status);
CREATE INDEX IF NOT EXISTS idx_finance_je_journal ON finance_journal_entry (journal_id, posting_date DESC);
CREATE INDEX IF NOT EXISTS idx_finance_je_status  ON finance_journal_entry (tenant_id, status);

-- ── Journal Entry Lines ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS finance_journal_entry_line (
    id               uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid           NOT NULL REFERENCES platform_tenant(id),
    created_at       timestamptz    NOT NULL DEFAULT NOW(),
    updated_at       timestamptz    NOT NULL DEFAULT NOW(),

    journal_entry_id uuid           NOT NULL REFERENCES finance_journal_entry(id) ON DELETE CASCADE,
    account_id       uuid           NOT NULL REFERENCES finance_account(id),
    cost_center_id   uuid           REFERENCES finance_cost_center(id) ON DELETE SET NULL,
    debit_amount     numeric(20,4)  NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
    credit_amount    numeric(20,4)  NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    memo             varchar(1024),
    -- Polymorphic reference to source document (invoice, payment, etc.).
    reference_type   varchar(100),
    reference_id     varchar(36),

    -- Exactly one of debit_amount or credit_amount must be non-zero.
    CONSTRAINT chk_finance_jel_single_side CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (debit_amount = 0 AND credit_amount > 0)
    )
);

COMMENT ON TABLE finance_journal_entry_line IS
    'Debit/credit line for a journal entry. '
    'Exactly one of debit_amount, credit_amount must be non-zero (enforced by CHECK constraint). '
    'Cascade deleted with parent journal_entry.';

ALTER TABLE finance_journal_entry_line ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_journal_entry_line FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_journal_entry_line
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_journal_entry_line
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_jel_entry   ON finance_journal_entry_line (journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_finance_jel_account ON finance_journal_entry_line (account_id);
