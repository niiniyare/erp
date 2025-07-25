-- =====================================================================
-- UOM CONVERSION DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON uom_conversion;
DROP POLICY IF EXISTS tenant_isolation_policy ON uom_conversion;
ALTER TABLE IF EXISTS uom_conversion DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS uom_conversion;
