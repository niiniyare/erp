-- =====================================================================
-- REVERT FINANCE ACCOUNT GROUPS AND HEADERS
-- Reverses the changes made by the 'up' migration script
-- =====================================================================
-- Revoke permissions
REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_groups
FROM
  application_role;

REVOKE
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_groups
FROM
  admin_role;

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_groups;

DROP POLICY IF EXISTS admin_full_access_policy ON finance_account_groups;

DROP POLICY IF EXISTS finance_account_groups_ro_select ON finance_account_groups;

-- Disable RLS
ALTER TABLE
  finance_account_groups DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_account_groups_tenant_code;

DROP INDEX IF EXISTS idx_account_groups_tenant_type;

DROP INDEX IF EXISTS idx_account_groups_parent;

DROP INDEX IF EXISTS idx_account_groups_hierarchy_level;

DROP INDEX IF EXISTS idx_account_groups_statement;

-- Drop table
DROP TABLE IF EXISTS finance_account_groups;
