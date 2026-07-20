DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_accounting_period;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_accounting_period;
DROP INDEX   IF EXISTS idx_finance_period_dates;
DROP INDEX   IF EXISTS idx_finance_period_fy;
DROP TABLE   IF EXISTS finance_accounting_period;
