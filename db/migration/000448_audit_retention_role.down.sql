-- ------------------------------------------------------------------------------------------------
-- DOWN MIGRATION FOR AUDIT RETENTION ROLE AND RLS POLICY
-- ------------------------------------------------------------------------------------------------
-- Reverts all changes made in the up migration.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- REVOKE EXECUTE ON HELPER FUNCTION
-- ------------------------------------------------------------------------------------------------
REVOKE EXECUTE ON FUNCTION delete_audit_logs_older_than(INTEGER, BOOLEAN) FROM audit_retention_role;

-- ------------------------------------------------------------------------------------------------
-- DROP HELPER FUNCTION
-- ------------------------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS delete_audit_logs_older_than(INTEGER, BOOLEAN) CASCADE;

-- ------------------------------------------------------------------------------------------------
-- DROP RLS POLICY
-- ------------------------------------------------------------------------------------------------
DROP POLICY IF EXISTS audit_log_retention_delete ON audit_log;

-- ------------------------------------------------------------------------------------------------
-- REVOKE DELETE PRIVILEGE
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- DROP THE ROLE
-- ------------------------------------------------------------------------------------------------
-- Note: Cannot drop role if it's still logged in or owns objects.
-- The CASCADE option transfers ownership of any objects owned by the role to the current user
-- (which should be a superuser in migration context), then drops the role.
-- ------------------------------------------------------------------------------------------------
DROP OWNED BY audit_retention_role CASCADE;
DROP ROLE IF EXISTS audit_retention_role;

-- ------------------------------------------------------------------------------------------------
-- VERIFICATION (commented out - for reference only)
-- ------------------------------------------------------------------------------------------------
-- -- Verify role no longer exists:
-- SELECT rolname FROM pg_roles WHERE rolname = 'audit_retention_role';  -- Should return 0 rows
-- 
-- -- Verify policy no longer exists:
-- SELECT policyname FROM pg_policies 
-- WHERE tablename = 'audit_log' AND policyname = 'audit_retention_delete';  -- Should return 0 rows
