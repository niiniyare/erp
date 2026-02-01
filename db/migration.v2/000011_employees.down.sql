-- =====================================================================
-- EMPLOYEES DOWN MIGRATION
-- =====================================================================
-- Drop triggers
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;

-- Drop RLS policies
DROP POLICY IF EXISTS employees_admin_access ON employees;

DROP POLICY IF EXISTS employees_tenant_isolation ON employees;

-- Disable RLS
ALTER TABLE
  IF EXISTS employees DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE
  IF EXISTS employees DROP CONSTRAINT IF EXISTS valid_security_level;

ALTER TABLE
  IF EXISTS employees DROP CONSTRAINT IF EXISTS valid_termination_date;

ALTER TABLE
  IF EXISTS employees DROP CONSTRAINT IF EXISTS employees_number_unique_active;

-- Drop indexes
DROP INDEX IF EXISTS idx_employees_access_attributes_gin;

DROP INDEX IF EXISTS idx_employees_work_schedule_gin;

DROP INDEX IF EXISTS idx_employees_salary_info_gin;

DROP INDEX IF EXISTS idx_employees_deleted_at;

DROP INDEX IF EXISTS idx_employees_security_level;

DROP INDEX IF EXISTS idx_employees_number;

DROP INDEX IF EXISTS idx_employees_active;

DROP INDEX IF EXISTS idx_employees_status;

DROP INDEX IF EXISTS idx_employees_manager;

DROP INDEX IF EXISTS idx_employees_department;

DROP INDEX IF EXISTS idx_employees_entity;

DROP INDEX IF EXISTS idx_employees_person;

DROP INDEX IF EXISTS idx_employees_tenant;

-- Drop table
DROP TABLE IF EXISTS employees;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================
DO
$$
BEGIN
RAISE NOTICE '===================================================================';

RAISE NOTICE 'EMPLOYEES DOWN MIGRATION COMPLETED';

RAISE NOTICE '===================================================================';

RAISE NOTICE 'Successfully removed:';

RAISE NOTICE '- 1 table (employees)';

RAISE NOTICE '- All associated indexes and constraints';

RAISE NOTICE '- All RLS policies and security settings';

RAISE NOTICE '- All triggers and permissions';

RAISE NOTICE '===================================================================';

END;

$$
;
