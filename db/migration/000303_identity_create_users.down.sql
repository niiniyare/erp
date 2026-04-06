-- =====================================================================
-- USERS DOWN MIGRATION
-- =====================================================================
-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop RLS policies
DROP POLICY IF EXISTS users_admin_access ON users;

DROP POLICY IF EXISTS users_tenant_isolation ON users;

DROP POLICY IF EXISTS users_ro_select ON users;

-- Disable RLS
ALTER TABLE
  IF EXISTS users DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS valid_lockout_time;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS valid_session_timeout;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS valid_failed_attempts;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS users_username_unique_active;

ALTER TABLE
  IF EXISTS users DROP CONSTRAINT IF EXISTS users_email_unique_active;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_settings_gin;

DROP INDEX IF EXISTS idx_users_attributes_gin;

DROP INDEX IF EXISTS idx_users_deleted_at;

DROP INDEX IF EXISTS idx_users_principal;
DROP INDEX IF EXISTS idx_users_mfa;

DROP INDEX IF EXISTS idx_users_lockout;

DROP INDEX IF EXISTS idx_users_failed_attempts;

DROP INDEX IF EXISTS idx_users_active;

DROP INDEX IF EXISTS idx_users_account_status;

DROP INDEX IF EXISTS idx_users_type;

DROP INDEX IF EXISTS idx_users_username_lower;

DROP INDEX IF EXISTS idx_users_email_lower;

DROP INDEX IF EXISTS idx_users_employee;

DROP INDEX IF EXISTS idx_users_person;

DROP INDEX IF EXISTS idx_users_entity;

DROP INDEX IF EXISTS idx_users_tenant;

-- Drop table
DROP TABLE IF EXISTS users;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'USERS DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (users)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '- All triggers and permissions';

RAISE NOTICE '===================================================================';

END;

$$
;
