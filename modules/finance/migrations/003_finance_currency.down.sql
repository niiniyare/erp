DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_currency;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_currency;
DROP INDEX   IF EXISTS idx_finance_currency_active;
DROP TABLE   IF EXISTS finance_currency;
