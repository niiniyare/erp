-- ------------------------------------------------------------------------------------------------
-- AUDIT RETENTION ROLE AND RLS POLICY
-- ------------------------------------------------------------------------------------------------
-- Creates a dedicated database role for audit log retention operations (deletions).
-- This role is intended ONLY for service accounts (archival jobs, GDPR erasure, TTL cleanup)
-- and should NEVER be assigned to interactive users.
--
-- DESIGN INTENT:
--   - Principle of least privilege: Only this role can delete from audit_log
--   - No other roles (including table owners) should have DELETE privileges
--   - RLS policy ensures the role can delete any row (simplified for service accounts)
--   - Separates retention operations from regular audit logging
--
-- DEPENDENCIES: Requires audit_log table to exist (typically from migration 000450 or 000451)
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- CREATE THE RETENTION ROLE
-- ------------------------------------------------------------------------------------------------
-- Note: CREATE ROLE is idempotent with IF NOT EXISTS (PostgreSQL 9.6+)
-- The role has no login capability by default - it's meant to be granted to service accounts
-- or used with SET ROLE from a trusted authenticator.
-- ------------------------------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'audit_retention_role') THEN
    CREATE ROLE audit_retention_role NOLOGIN NOSUPERUSER INHERIT NOCREATEDB NOCREATEROLE NOREPLICATION;
  END IF;
END
$$;

COMMENT ON ROLE audit_retention_role IS
  'Dedicated role for audit log retention operations (deletions only). '
  'Assigned exclusively to service accounts (archival jobs, GDPR erasure, TTL cleanup). '
  'Never assign to interactive user roles.';

-- ------------------------------------------------------------------------------------------------
-- GRANT DELETE PRIVILEGE ON AUDIT_LOG
-- ------------------------------------------------------------------------------------------------
-- Only grant DELETE - this role should never need SELECT, INSERT, or UPDATE on audit_log.
-- The table owner or superusers may need to grant this privilege.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- RLS POLICY FOR RETENTION DELETIONS
-- ------------------------------------------------------------------------------------------------
-- Row Security Policy that allows audit_retention_role to delete any row.
-- Using TRUE as the USING expression because:
--   1. The role is already tightly controlled (never assigned to users)
--   2. Retention jobs need to delete based on age/retention rules, not row-level conditions
--   3. Additional filtering should happen in the application/job logic, not RLS
--
-- Note: Assumes RLS is already enabled on audit_log table.
-- If not, run: ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
-- ------------------------------------------------------------------------------------------------

-- First, drop the policy if it exists (for idempotency)
-- DROP POLICY IF EXISTS audit_log_retention_delete ON audit_log;

-- Create the policy
-- CREATE POLICY audit_log_retention_delete ON audit_log
--   FOR DELETE 
--   TO audit_retention_role
--   USING (TRUE);


-- ------------------------------------------------------------------------------------------------
-- VERIFICATION QUERIES (commented out - for reference only)
-- ------------------------------------------------------------------------------------------------
-- -- Check if role exists:
-- SELECT rolname FROM pg_roles WHERE rolname = 'audit_retention_role';
-- 
-- -- Check if DELETE privilege is granted:
-- SELECT grantee, privilege_type 
-- FROM information_schema.role_table_grants 
-- WHERE table_name = 'audit_log' AND privilege_type = 'DELETE';
-- 
-- -- Check if policy exists:
-- SELECT policyname, tablename, cmd, roles 
-- FROM pg_policies 
-- WHERE tablename = 'audit_log' AND policyname = 'audit_log_retention_delete';

-- ------------------------------------------------------------------------------------------------
-- HELPER FUNCTION: SAFE_AUDIT_DELETE_BY_AGE
-- ------------------------------------------------------------------------------------------------
-- Example helper function for retention jobs to delete old audit logs.
-- This function executes with the privileges of the caller - must be called by 
-- audit_retention_role or a role that can SET ROLE to it.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION delete_audit_logs_older_than(
  p_days INTEGER,
  p_dry_run BOOLEAN DEFAULT FALSE
)
RETURNS TABLE(
  deleted_count BIGINT,
  oldest_retained_date TIMESTAMPTZ,
  message TEXT
) AS $$
DECLARE
  v_cutoff_date TIMESTAMPTZ;
  v_deleted_count BIGINT;
BEGIN
  -- Calculate cutoff date
  v_cutoff_date := NOW() - (p_days || ' days')::INTERVAL;
  
  -- Check if we're in dry-run mode
  IF p_dry_run THEN
    SELECT COUNT(*), MIN(created_at)
    INTO v_deleted_count, oldest_retained_date
    FROM audit_log
    WHERE created_at < v_cutoff_date;
    
    RETURN QUERY SELECT 
      v_deleted_count, 
      oldest_retained_date,
      format('DRY RUN: Would delete %s records older than %s', v_deleted_count, v_cutoff_date)::TEXT;
  ELSE
    -- Perform actual deletion
    WITH deleted AS (
      DELETE FROM audit_log
      WHERE created_at < v_cutoff_date
      RETURNING id, created_at
    )
    SELECT COUNT(*), MIN(created_at)
    INTO v_deleted_count, oldest_retained_date
    FROM deleted;
    
    RETURN QUERY SELECT 
      v_deleted_count, 
      oldest_retained_date,
      format('Deleted %s records older than %s', v_deleted_count, v_cutoff_date)::TEXT;
  END IF;
END;
$$ LANGUAGE plpgsql SECURITY INVOKER;

COMMENT ON FUNCTION delete_audit_logs_older_than IS
  'Helper function for retention jobs to delete audit logs older than specified days. '
  'SECURITY INVOKER means it runs with caller''s privileges - must be called by audit_retention_role. '
  'Use p_dry_run TRUE to preview deletions without actually removing data.';

-- ------------------------------------------------------------------------------------------------
-- GRANT EXECUTE ON HELPER FUNCTION
-- ------------------------------------------------------------------------------------------------
GRANT EXECUTE ON FUNCTION delete_audit_logs_older_than(INTEGER, BOOLEAN) TO audit_retention_role;
