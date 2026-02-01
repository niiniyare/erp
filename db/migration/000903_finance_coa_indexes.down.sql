-- Drop chart of accounts indexes
DROP INDEX IF EXISTS idx_chart_of_accounts_attributes_gin;

DROP INDEX IF EXISTS idx_chart_of_accounts_active;

DROP INDEX IF EXISTS idx_chart_of_accounts_entity;

DROP INDEX IF EXISTS idx_chart_of_accounts_currency;

DROP INDEX IF EXISTS idx_chart_of_accounts_control;

DROP INDEX IF EXISTS idx_chart_of_accounts_balance_tracking;

DROP INDEX IF EXISTS idx_chart_of_accounts_statement_line;

DROP INDEX IF EXISTS idx_chart_of_accounts_parent;

DROP INDEX IF EXISTS idx_chart_of_accounts_tenant_type;

DROP INDEX IF EXISTS idx_chart_of_accounts_tenant_code;
