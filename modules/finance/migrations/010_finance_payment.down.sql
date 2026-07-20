DROP POLICY IF EXISTS tenant_isolation ON finance_bank_transaction;
DROP INDEX  IF EXISTS idx_finance_bt_status;
DROP INDEX  IF EXISTS idx_finance_bt_account_date;
DROP TABLE  IF EXISTS finance_bank_transaction;

DROP POLICY IF EXISTS tenant_isolation ON finance_allocation;
DROP INDEX  IF EXISTS idx_finance_allocation_invoice;
DROP INDEX  IF EXISTS idx_finance_allocation_payment;
DROP TABLE  IF EXISTS finance_allocation;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_payment;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_payment;
DROP INDEX   IF EXISTS idx_finance_payment_date;
DROP INDEX   IF EXISTS idx_finance_payment_status;
DROP TABLE   IF EXISTS finance_payment;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_payment_method;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_payment_method;
DROP TABLE   IF EXISTS finance_payment_method;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_bank_account;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_bank_account;
DROP TABLE   IF EXISTS finance_bank_account;
