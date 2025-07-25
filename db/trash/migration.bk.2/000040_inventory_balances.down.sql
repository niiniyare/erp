-- =====================================================================
-- INVENTORY_BALANCES DOWN MIGRATION
-- =====================================================================

DROP TRIGGER IF EXISTS update_inventory_balances_updated_at ON inventory_balances;
DROP POLICY IF EXISTS admin_full_access_policy ON inventory_balances;
DROP POLICY IF EXISTS tenant_isolation_policy ON inventory_balances;
ALTER TABLE IF EXISTS inventory_balances DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS inventory_balances;
