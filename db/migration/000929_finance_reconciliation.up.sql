-- Finance Bank Reconciliation
-- Bank statement headers and lines with journal entry matching.

CREATE TABLE IF NOT EXISTS finance_bank_statements (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID          NOT NULL,
    account_id           UUID          NOT NULL,  -- GL bank/cash account

    statement_reference  VARCHAR(100)  NOT NULL,
    statement_date       DATE          NOT NULL,
    start_date           DATE          NOT NULL,
    currency_code        CHAR(3)       NOT NULL DEFAULT 'USD',

    opening_balance      NUMERIC(18,4) NOT NULL DEFAULT 0,
    closing_balance      NUMERIC(18,4) NOT NULL DEFAULT 0,

    status               VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                             CHECK (status IN ('DRAFT','IN_PROGRESS','COMPLETED','VOIDED')),

    -- Running tallies (updated as lines are matched)
    matched_count        INT           NOT NULL DEFAULT 0,
    unmatched_count      INT           NOT NULL DEFAULT 0,
    difference_amount    NUMERIC(18,4) NOT NULL DEFAULT 0,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID,
    updated_by  UUID,

    CONSTRAINT uq_bank_statement UNIQUE (tenant_id, account_id, statement_reference)
);

CREATE INDEX idx_finance_bank_statements_tenant  ON finance_bank_statements(tenant_id);
CREATE INDEX idx_finance_bank_statements_account ON finance_bank_statements(account_id);
CREATE INDEX idx_finance_bank_statements_date    ON finance_bank_statements(statement_date);
CREATE INDEX idx_finance_bank_statements_status  ON finance_bank_statements(tenant_id, status);

ALTER TABLE finance_bank_statements ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_statements FORCE ROW LEVEL SECURITY;

CREATE POLICY finance_bank_statements_tenant_isolation
    ON finance_bank_statements
    USING (tenant_id = current_tenant_id());

-- Bank statement lines
CREATE TABLE IF NOT EXISTS finance_bank_statement_lines (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id     UUID          NOT NULL REFERENCES finance_bank_statements(id) ON DELETE CASCADE,
    tenant_id        UUID          NOT NULL,

    transaction_date DATE          NOT NULL,
    value_date       DATE,

    description      TEXT          NOT NULL,
    reference        VARCHAR(200),

    debit_amount     NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (debit_amount  >= 0),
    credit_amount    NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    balance          NUMERIC(18,4) NOT NULL DEFAULT 0,

    -- Reconciliation
    is_reconciled    BOOLEAN       NOT NULL DEFAULT FALSE,
    reconciled_at    TIMESTAMPTZ,
    reconciled_by    UUID,

    -- Matched journal entry line
    matched_entry_id UUID,  -- FK to finance_transaction_entries.id (loose ref — no CASCADE)

    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_finance_bank_stmt_lines_statement   ON finance_bank_statement_lines(statement_id);
CREATE INDEX idx_finance_bank_stmt_lines_tenant      ON finance_bank_statement_lines(tenant_id);
CREATE INDEX idx_finance_bank_stmt_lines_date        ON finance_bank_statement_lines(transaction_date);
CREATE INDEX idx_finance_bank_stmt_lines_unmatched   ON finance_bank_statement_lines(statement_id, is_reconciled);

ALTER TABLE finance_bank_statement_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_bank_statement_lines FORCE ROW LEVEL SECURITY;

CREATE POLICY finance_bank_statement_lines_tenant_isolation
    ON finance_bank_statement_lines
    USING (tenant_id = current_tenant_id());
