-- Reverse: drop document tables in reverse dependency order.

DROP TABLE IF EXISTS finance_bank_transaction;
DROP TABLE IF EXISTS finance_receipt;
DROP TABLE IF EXISTS finance_debit_note;
DROP TABLE IF EXISTS finance_credit_note;
DROP TABLE IF EXISTS finance_allocation;
DROP TABLE IF EXISTS finance_invoice_line;
DROP TABLE IF EXISTS finance_invoice;
DROP TABLE IF EXISTS finance_payment;
DROP TABLE IF EXISTS finance_payment_method;
DROP TABLE IF EXISTS finance_bank_account;
DROP TABLE IF EXISTS finance_tax_entry;
DROP TABLE IF EXISTS finance_tax;
DROP TABLE IF EXISTS finance_tax_group;
