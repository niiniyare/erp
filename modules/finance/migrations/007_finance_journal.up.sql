-- Finance 007: finance_journal — journal master.
--
-- Journals classify double-entry postings: General, Sales, Purchase, Bank, Cash, Opening.

CREATE TABLE IF NOT EXISTS finance_journal (
    id                 uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          uuid         NOT NULL REFERENCES platform_tenant(id),
    created_at         timestamptz  NOT NULL DEFAULT NOW(),
    updated_at         timestamptz  NOT NULL DEFAULT NOW(),

    name               varchar(255) NOT NULL,
    code               varchar(10)  NOT NULL,
    journal_type       varchar(20)  NOT NULL
                           CHECK (journal_type IN ('General','Sales','Purchase','Bank','Cash','Opening')),
    description        varchar(1024),
    default_account_id uuid         REFERENCES finance_account(id) ON DELETE SET NULL,
    active             boolean      NOT NULL DEFAULT true,

    CONSTRAINT uq_finance_journal_tenant_code UNIQUE (tenant_id, code)
);

COMMENT ON TABLE finance_journal IS
    'Journal master. Classifies entries by type. '
    'default_account_id is the default GL account for new entries in this journal.';

ALTER TABLE finance_journal ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_journal FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON finance_journal
    AS PERMISSIVE FOR ALL TO PUBLIC
    USING (tenant_id = current_tenant_id());

CREATE TRIGGER trg_set_updated_at
    BEFORE UPDATE ON finance_journal
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX IF NOT EXISTS idx_finance_journal_type ON finance_journal (tenant_id, journal_type, active);
