-- =====================================================================
-- PERSONS DOWN MIGRATION
-- =====================================================================
-- Drop triggers
DROP TRIGGER IF EXISTS update_persons_updated_at ON persons;

-- Drop RLS policies
DROP POLICY IF EXISTS persons_admin_access ON persons;

DROP POLICY IF EXISTS persons_tenant_isolation ON persons;

-- Disable RLS
ALTER TABLE
  IF EXISTS persons DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_persons_metadata_gin;

DROP INDEX IF EXISTS idx_persons_security_attributes_gin;

DROP INDEX IF EXISTS idx_persons_address_gin;

DROP INDEX IF EXISTS idx_persons_name;

DROP INDEX IF EXISTS idx_persons_email_lower;

DROP INDEX IF EXISTS idx_persons_deleted_at;

DROP INDEX IF EXISTS idx_persons_active;

DROP INDEX IF EXISTS idx_persons_entity;

DROP INDEX IF EXISTS idx_persons_type;

DROP INDEX IF EXISTS idx_persons_tenant;

-- Drop constraints
ALTER TABLE
  IF EXISTS persons DROP CONSTRAINT IF EXISTS persons_national_id_unique_active;

ALTER TABLE
  IF EXISTS persons DROP CONSTRAINT IF EXISTS persons_email_unique_active;

-- Drop table
DROP TABLE IF EXISTS persons;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'PERSONS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (persons)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '- All triggers and permissions';

RAISE NOTICE '===================================================================';

END;

$$
;
