-- =====================================================================
-- ITEM_CATEGORIES DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON item_categories;
DROP POLICY IF EXISTS tenant_isolation_policy ON item_categories;
ALTER TABLE IF EXISTS item_categories DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS item_categories;
