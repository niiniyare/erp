-- =====================================================================
-- UOM DOWN MIGRATION
-- =====================================================================

DROP TRIGGER IF EXISTS update_uom_updated_at ON uom;
DROP POLICY IF EXISTS admin_full_access_policy ON uom;
DROP POLICY IF EXISTS tenant_isolation_policy ON uom;
ALTER TABLE IF EXISTS uom DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS uom;
