DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_cost_center;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_cost_center;
DROP INDEX   IF EXISTS idx_finance_cost_center_parent;
DROP TABLE   IF EXISTS finance_cost_center;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_account;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_account;
DROP INDEX   IF EXISTS idx_finance_account_type;
DROP INDEX   IF EXISTS idx_finance_account_parent;
DROP INDEX   IF EXISTS idx_finance_account_coa;
DROP TABLE   IF EXISTS finance_account;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_chart_of_accounts;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_chart_of_accounts;
DROP TABLE   IF EXISTS finance_chart_of_accounts;
