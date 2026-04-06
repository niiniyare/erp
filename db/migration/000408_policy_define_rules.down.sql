-- =====================================================================
-- POLICIES DOWN MIGRATION
-- =====================================================================
-- Drop RLS policies
DROP POLICY IF EXISTS policies_tenant_isolation ON policies;
DROP POLICY IF EXISTS policies_admin_access ON policies;
DROP POLICY IF EXISTS policies_ro_select ON policies;

-- Disable RLS
ALTER TABLE
  IF EXISTS policies DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS policies DROP CONSTRAINT IF EXISTS policies_name_unique_per_tenant;

-- Drop table
DROP TABLE IF EXISTS policies;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'POLICIES DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Table dropped: policies';

RAISE NOTICE 'RLS disabled and policy dropped.';

RAISE NOTICE '===================================================================';

END;

$$
;
