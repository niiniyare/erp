DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_fiscal_year;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_fiscal_year;
DROP INDEX   IF EXISTS idx_finance_fiscal_year_status;
DROP TABLE   IF EXISTS finance_fiscal_year;
