-- +migrate Down

BEGIN;

-- Revoke permissions
REVOKE SELECT ON v_financial_statement_builder FROM application_role, admin_role;
REVOKE SELECT ON v_chart_of_accounts_complete FROM application_role, admin_role;
REVOKE SELECT ON v_finance_accounts_with_groups FROM application_role, admin_role;

-- Drop views
DROP VIEW IF EXISTS v_financial_statement_builder;
DROP VIEW IF EXISTS v_chart_of_accounts_complete;
DROP VIEW IF EXISTS v_financial_statement_structure;
DROP VIEW IF EXISTS v_finance_accounts_with_groups;

-- Drop trigger and function for group hierarchy
DROP TRIGGER IF EXISTS trigger_maintain_group_hierarchy ON finance_account_groups;
DROP FUNCTION IF EXISTS maintain_group_hierarchy_path();

-- Drop other functions
DROP FUNCTION IF EXISTS create_standard_account_groups(UUID);
DROP FUNCTION IF EXISTS update_group_balances(UUID);

COMMIT;
