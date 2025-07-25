-- =====================================================================
-- CHARTOFACCOUNT DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON chartofaccount;
DROP POLICY IF EXISTS tenant_isolation_policy ON chartofaccount;
ALTER TABLE IF EXISTS chartofaccount DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS chartofaccount;
