DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_tax;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_tax;
DROP INDEX   IF EXISTS idx_finance_tax_applies;
DROP TABLE   IF EXISTS finance_tax;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_tax_group;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_tax_group;
DROP TABLE   IF EXISTS finance_tax_group;
