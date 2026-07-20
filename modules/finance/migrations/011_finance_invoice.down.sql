-- Drop deferred FK first.
ALTER TABLE finance_allocation
    DROP CONSTRAINT IF EXISTS fk_finance_allocation_invoice;

DROP POLICY IF EXISTS tenant_isolation ON finance_tax_entry;
DROP INDEX  IF EXISTS idx_finance_tax_entry_je;
DROP INDEX  IF EXISTS idx_finance_tax_entry_date;
DROP TABLE  IF EXISTS finance_tax_entry;

DROP POLICY IF EXISTS tenant_isolation ON finance_receipt;
DROP INDEX  IF EXISTS idx_finance_receipt_payment;
DROP TABLE  IF EXISTS finance_receipt;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_debit_note;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_debit_note;
DROP TABLE   IF EXISTS finance_debit_note;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_credit_note;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_credit_note;
DROP TABLE   IF EXISTS finance_credit_note;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_invoice_line;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_invoice_line;
DROP INDEX   IF EXISTS idx_finance_invoice_line_invoice;
DROP TABLE   IF EXISTS finance_invoice_line;

DROP TRIGGER IF EXISTS trg_set_updated_at ON finance_invoice;
DROP POLICY  IF EXISTS tenant_isolation   ON finance_invoice;
DROP INDEX   IF EXISTS idx_finance_invoice_due;
DROP INDEX   IF EXISTS idx_finance_invoice_date;
DROP INDEX   IF EXISTS idx_finance_invoice_customer;
DROP INDEX   IF EXISTS idx_finance_invoice_status;
DROP TABLE   IF EXISTS finance_invoice;
