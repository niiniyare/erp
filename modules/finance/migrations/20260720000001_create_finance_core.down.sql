-- Reverse: drop core finance tables in reverse dependency order.

DROP TABLE IF EXISTS finance_cost_center;
DROP TABLE IF EXISTS finance_account;
DROP TABLE IF EXISTS finance_chart_of_accounts;
DROP TABLE IF EXISTS finance_accounting_period;
DROP TABLE IF EXISTS finance_fiscal_year;
DROP TABLE IF EXISTS finance_exchange_rate;
DROP TABLE IF EXISTS finance_currency;
