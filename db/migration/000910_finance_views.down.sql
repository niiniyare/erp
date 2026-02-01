-- +migrate Down
BEGIN
;

DROP VIEW IF EXISTS v_finance_accounts_hierarchy;

DROP VIEW IF EXISTS v_finance_transaction_summary;

DROP VIEW IF EXISTS v_finance_account_activity;

COMMIT;
