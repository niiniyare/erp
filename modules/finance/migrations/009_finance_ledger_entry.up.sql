-- Finance 009: finance_ledger_entry — immutable GL ledger (MANDATORY system entity).
--
-- Written exclusively by the framework when a journal entry is posted.
-- API cannot Create, Write, or Delete. Provides the authoritative GL balance.
-- running_balance is updated by the framework after posting.

CREATE TABLE IF NOT EXISTS finance_ledger_entry (
    id                      uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               uuid           NOT NULL REFERENCES platform_tenant(id),
    created_at              timestamptz    NOT NULL DEFAULT NOW(),
    -- No updated_at: immutable after creation.

    -- NamingSeries: LE-{YYYY}-{SEQ:7}
    entry_number            varchar(100),
    account_id              uuid           NOT NULL REFERENCES finance_account(id),
    journal_entry_id        uuid           NOT NULL REFERENCES finance_journal_entry(id),
    journal_entry_line_id   uuid           NOT NULL REFERENCES finance_journal_entry_line(id),
    posting_date            date           NOT NULL,
    debit_amount            numeric(20,4)  NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
    credit_amount           numeric(20,4)  NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    running_balance         numeric(20,4)  NOT NULL DEFAULT 0,
    currency_id             uuid           REFERENCES finance_currency(id),
    exchange_rate           numeric(20,10) NOT NULL DEFAULT 1.0,
    memo                    varchar(1024),
    cost_center_id          uuid           REFERENCES finance_cost_center(id),
    is_cancelled            boolean        NOT NULL DEFAULT false
);

COMMENT ON TABLE finance_ledger_entry IS
    'Immutable GL ledger (MANDATORY system entity). '
    'Written by framework only when journal entry is posted. '
    'running_balance is the account balance after this entry.';

ALTER TABLE finance_ledger_entry ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_ledger_entry FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_ledger_entry
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

-- No updated_at trigger — immutable.

CREATE INDEX IF NOT EXISTS idx_finance_le_account_date
    ON finance_ledger_entry (account_id, posting_date DESC);
CREATE INDEX IF NOT EXISTS idx_finance_le_je
    ON finance_ledger_entry (journal_entry_id);
