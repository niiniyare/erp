-- =====================================================================
-- VENDOR DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON vendor;
DROP POLICY IF EXISTS tenant_isolation_policy ON vendor;
ALTER TABLE IF EXISTS vendor DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS vendor;
