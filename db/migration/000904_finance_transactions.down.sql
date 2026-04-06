-- Drop finance transactions RLS policies and table
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_transactions;
DROP POLICY IF EXISTS admin_full_access_policy ON finance_transactions;
DROP POLICY IF EXISTS finance_transactions_ro_select ON finance_transactions;
ALTER TABLE IF EXISTS finance_transactions NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS finance_transactions CASCADE;
