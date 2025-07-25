-- =====================================================================
-- ITEMS DOWN MIGRATION
-- =====================================================================

DROP TRIGGER IF EXISTS update_items_updated_at ON items;
DROP POLICY IF EXISTS admin_full_access_policy ON items;
DROP POLICY IF EXISTS tenant_isolation_policy ON items;
ALTER TABLE IF EXISTS items DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS items;
