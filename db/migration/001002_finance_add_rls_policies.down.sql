-- +migrate Down
BEGIN
;

-- Revoke permissions on views
REVOKE
SELECT
  ON v_finance_account_activity
FROM
  application_role,
  admin_role;

REVOKE
SELECT
  ON v_finance_transaction_summary
FROM
  application_role,
  admin_role;

REVOKE
SELECT
  ON v_finance_accounts_hierarchy
FROM
  application_role,
  admin_role;

-- Revoke permissions on new tables
REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_validation_rules
FROM
  application_role,
  admin_role;

REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_balances
FROM
  application_role,
  admin_role;

-- Drop admin policies for new tables
DROP POLICY IF EXISTS admin_full_access_policy ON finance_account_validation_rules;

DROP POLICY IF EXISTS admin_full_access_policy ON finance_account_balances;

-- Drop RLS policies for new tables
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_validation_rules;

DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_balances;

-- Disable RLS on new tables
ALTER TABLE
  finance_account_validation_rules DISABLE ROW LEVEL SECURITY;

ALTER TABLE
  finance_account_balances DISABLE ROW LEVEL SECURITY;

COMMIT;
