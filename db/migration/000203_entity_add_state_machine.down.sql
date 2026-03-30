-- =====================================================================
-- ENTITY STATE DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS admin_full_access_policy ON entitystate;

DROP POLICY IF EXISTS tenant_isolation_policy ON entitystate;

-- Disable RLS
ALTER TABLE
  IF EXISTS entitystate DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS entitystate DROP CONSTRAINT IF EXISTS positive_sequence;

ALTER TABLE
  IF EXISTS entitystate DROP CONSTRAINT IF EXISTS unique_tenant_entity_key_fy;

-- Drop indexes
DROP INDEX IF EXISTS idx_entitystate_fiscal_year;

DROP INDEX IF EXISTS idx_entitystate_entity_key;

DROP INDEX IF EXISTS idx_entitystate_config_gin;
-- Drop table
DROP TABLE IF EXISTS entitystate;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITY STATE DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (entitystate)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '===================================================================';

END;

$$
;
