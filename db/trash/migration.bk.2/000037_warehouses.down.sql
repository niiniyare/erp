-- =====================================================================
-- WAREHOUSES DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON warehouses;
DROP POLICY IF EXISTS tenant_isolation_policy ON warehouses;
ALTER TABLE IF EXISTS warehouses DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS warehouses;
