DROP POLICY IF EXISTS tenant_isolation ON finance_ledger_entry;
DROP INDEX  IF EXISTS idx_finance_le_je;
DROP INDEX  IF EXISTS idx_finance_le_account_date;
DROP TABLE  IF EXISTS finance_ledger_entry;
