-- =====================================================================
-- LEDGER DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON ledger;
DROP POLICY IF EXISTS tenant_isolation_policy ON ledger;
ALTER TABLE IF EXISTS ledger DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS ledger;
