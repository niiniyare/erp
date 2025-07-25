-- =====================================================================
-- CUSTOMER DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON customer;
DROP POLICY IF EXISTS tenant_isolation_policy ON customer;
ALTER TABLE IF EXISTS customer DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS customer;
