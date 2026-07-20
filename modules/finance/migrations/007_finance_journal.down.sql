DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_journal;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_journal;
DROP INDEX   IF EXISTS idx_finance_journal_type;
DROP TABLE   IF EXISTS finance_journal;
