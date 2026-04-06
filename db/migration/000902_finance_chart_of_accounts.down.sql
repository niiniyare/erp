-- Drop chart of accounts RLS policies and table
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_accounts;
DROP POLICY IF EXISTS admin_full_access_policy ON finance_accounts;
DROP POLICY IF EXISTS finance_accounts_ro_select ON finance_accounts;
ALTER TABLE IF EXISTS finance_accounts NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS finance_accounts CASCADE;
