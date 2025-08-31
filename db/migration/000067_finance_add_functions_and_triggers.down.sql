-- +migrate Down
BEGIN;

DROP TRIGGER IF EXISTS trigger_update_account_balances ON finance_transactions;
DROP FUNCTION IF EXISTS update_account_balances_after_posting();

DROP TRIGGER IF EXISTS trigger_maintain_account_hierarchy ON finance_accounts;
DROP FUNCTION IF EXISTS maintain_account_hierarchy_path();

DROP TRIGGER IF EXISTS trigger_update_hierarchy_flags ON finance_accounts;
DROP FUNCTION IF EXISTS update_account_hierarchy_flags();

DROP FUNCTION IF EXISTS validate_account_hierarchy(UUID);
DROP FUNCTION IF EXISTS recalculate_account_balance(UUID);

DROP TRIGGER IF EXISTS trigger_validate_transaction_posting ON finance_transactions;
DROP FUNCTION IF EXISTS validate_transaction_before_posting();

DROP FUNCTION IF EXISTS get_account_full_name(UUID);
DROP FUNCTION IF EXISTS get_account_children(UUID, BOOLEAN);
DROP FUNCTION IF EXISTS analyze_finance_tables_performance();

COMMIT;
