-- =====================================================================
-- INVENTORY_MOVEMENTS DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON inventory_movements;
DROP POLICY IF EXISTS tenant_isolation_policy ON inventory_movements;
ALTER TABLE IF EXISTS inventory_movements DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS inventory_movements;
