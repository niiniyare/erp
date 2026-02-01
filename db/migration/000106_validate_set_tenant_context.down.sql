DROP FUNCTION IF EXISTS validate_and_set_tenant_context(UUID);

-- -- =====================================================================
-- -- PROJECTS DOWN MIGRATION
-- -- =====================================================================
--
-- DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
-- DROP POLICY IF EXISTS admin_full_access_policy ON projects;
-- DROP POLICY IF EXISTS tenant_isolation_policy ON projects;
-- ALTER TABLE IF EXISTS projects DISABLE ROW LEVEL SECURITY;
-- DROP TABLE IF EXISTS projects;
