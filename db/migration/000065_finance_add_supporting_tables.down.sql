-- +migrate Down
BEGIN
;

DROP TABLE IF EXISTS finance_account_validation_rules;

DROP TABLE IF EXISTS finance_account_balances;

ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS has_children;

ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS is_leaf_account;

COMMIT;
