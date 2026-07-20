DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_journal_entry_line;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_journal_entry_line;
DROP INDEX   IF EXISTS idx_finance_jel_account;
DROP INDEX   IF EXISTS idx_finance_jel_entry;
DROP TABLE   IF EXISTS finance_journal_entry_line;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_journal_entry;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_journal_entry;
DROP INDEX   IF EXISTS idx_finance_je_status;
DROP INDEX   IF EXISTS idx_finance_je_journal;
DROP INDEX   IF EXISTS idx_finance_je_period;
DROP TABLE   IF EXISTS finance_journal_entry;
