-- =====================================================================
-- ENTITIES CORE DOWN MIGRATION
-- =====================================================================
--
-- This down migration properly cleans up the entities core table
--
-- Order of operations:
-- 1. Drop table RLS policies
-- 2. Disable RLS on table
-- 3. Drop triggers and functions
-- 4. Drop indexes
-- 5. Drop table constraints
-- 6. Drop table
-- =====================================================================
-- =====================================================================
-- DROP TABLE RLS POLICIES
-- =====================================================================
-- Drop admin bypass policy
DROP POLICY IF EXISTS admin_full_access_policy ON entities;

-- Drop tenant isolation policy
DROP POLICY IF EXISTS tenant_isolation_policy ON entities;

-- =====================================================================
-- DISABLE RLS ON TABLE
-- =====================================================================
ALTER TABLE
  IF EXISTS entities DISABLE ROW LEVEL SECURITY;

-- =====================================================================
-- DROP TRIGGERS AND FUNCTIONS
-- =====================================================================
-- Drop trigger
DROP TRIGGER IF EXISTS entities_maintain_entity_id ON entities;

-- Drop entity-specific trigger function
DROP FUNCTION IF EXISTS maintain_entity_id();

-- Note: Shared functions like current_tenant_id() are managed by tenants_core migration
-- =====================================================================
-- DROP INDEXES
-- =====================================================================
-- Drop performance indexes
DROP INDEX IF EXISTS idx_entities_type;

DROP INDEX IF EXISTS idx_entities_parent;

DROP INDEX IF EXISTS idx_entities_tenant;

DROP INDEX IF EXISTS idx_entities_deleted_at;

DROP INDEX IF EXISTS idx_entities_address_gin;

DROP INDEX IF EXISTS idx_entities_level;
DROP INDEX IF EXISTS idx_entities_path;
DROP INDEX IF EXISTS idx_entities_settings_gin;

DROP INDEX IF EXISTS idx_entities_active;

DROP INDEX IF EXISTS idx_entities_parent_id;

DROP INDEX IF EXISTS idx_entities_tenant_type;

DROP INDEX IF EXISTS tenant_code_unique_idx;

-- =====================================================================
-- DROP TABLE CONSTRAINTS
-- =====================================================================
-- Drop check constraints
ALTER TABLE
  IF EXISTS entities DROP CONSTRAINT IF EXISTS valid_fy_start_month;

ALTER TABLE
  IF EXISTS entities DROP CONSTRAINT IF EXISTS no_self_parent;

-- =====================================================================
-- DROP TABLE
-- =====================================================================
-- Drop entities table
DROP TABLE IF EXISTS entities;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'ENTITIES CORE DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (entities)';

RAISE NOTICE '- All associated indexes, triggers, and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '===================================================================';

END;

$$
;
