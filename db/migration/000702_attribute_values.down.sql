-- Drop attribute_sources policies and table
DROP POLICY IF EXISTS attribute_sources_tenant_isolation ON attribute_sources;
DROP POLICY IF EXISTS attribute_sources_admin_access ON attribute_sources;
DROP POLICY IF EXISTS attribute_sources_ro_select ON attribute_sources;
ALTER TABLE IF EXISTS attribute_sources NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS attribute_sources CASCADE;

-- Drop attribute_values policies and table
DROP POLICY IF EXISTS attribute_values_tenant_isolation ON attribute_values;
DROP POLICY IF EXISTS attribute_values_admin_access ON attribute_values;
DROP POLICY IF EXISTS attribute_values_ro_select ON attribute_values;
ALTER TABLE IF EXISTS attribute_values NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS attribute_values CASCADE;
