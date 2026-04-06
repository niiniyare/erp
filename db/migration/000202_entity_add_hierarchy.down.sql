-- =====================================================================
-- ENTITIES HIERARCHY DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS admin_full_access_policy ON hierarchy_paths;

DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;

DROP POLICY IF EXISTS hierarchy_paths_ro_select ON hierarchy_paths;

-- Disable RLS
ALTER TABLE
  IF EXISTS hierarchy_paths DISABLE ROW LEVEL SECURITY;

-- Drop triggers
DROP TRIGGER IF EXISTS hierarchy_paths_maintain_entity_id ON hierarchy_paths;

-- Drop indexes
DROP INDEX IF EXISTS idx_hierarchy_paths_ancestor;

DROP INDEX IF EXISTS idx_hierarchy_paths_tenant;

DROP INDEX IF EXISTS idx_hierarchy_paths_depth;

DROP INDEX IF EXISTS idx_hierarchy_paths_descendant;

-- Drop table
DROP TABLE IF EXISTS hierarchy_paths;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITIES HIERARCHY DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (hierarchy_paths)';

RAISE NOTICE '- All associated indexes and triggers';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '===================================================================';

END;

$$
;
