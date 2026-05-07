-- +migrate Down
BEGIN
;

DROP TABLE IF EXISTS finance_account_validation_rules;

DROP TABLE IF EXISTS finance_account_balances;

-- is_leaf is GENERATED ALWAYS AS (NOT has_children) — must drop before has_children.
ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS is_leaf;

ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS is_leaf_account;

ALTER TABLE
  finance_accounts DROP COLUMN IF EXISTS has_children;

COMMIT;
