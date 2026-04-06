-- ------------------------------------------------------------------------------------------------
-- TENANTS TABLE — ROW LEVEL SECURITY POLICIES
-- ------------------------------------------------------------------------------------------------
-- Defines the three RLS policies for the tenants table.
--
-- Policy design principles:
--   • Fail-closed: application_role sees ZERO rows when no tenant context is set.
--     There is NO fallback "OR current_tenant_id() IS NULL" clause. Fail-open RLS
--     (the v1 behaviour) is a data exfiltration vulnerability in pooled connection
--     environments where a missed set_tenant_context() call would expose all tenants.
--   • Soft-delete exclusion in the policy itself (not just in set_tenant_context).
--     Defense-in-depth: if an app bug bypasses the function and sets a deleted
--     tenant's ID directly via SET LOCAL, the policy still prevents access.
--   • admin_role uses an explicit USING(TRUE) policy rather than BYPASSRLS.
--     The explicit policy is visible in pg_policies and therefore auditable.
--     BYPASSRLS is invisible in standard role inspection and easier to accidentally
--     grant broadly.
--
-- NOTE: current_tenant_id() function must exist (migration 000055) before this runs.
--       GRANTs are in migration 000059.
-- ------------------------------------------------------------------------------------------------

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Force RLS even for the table owner (typically the migration user).
-- Without this, the table owner bypasses all policies.
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;

-- Drop all existing policies before redefining — idempotent re-runs.
DROP POLICY IF EXISTS tenant_isolation_policy   ON tenants;
DROP POLICY IF EXISTS admin_full_access_policy  ON tenants;
DROP POLICY IF EXISTS readonly_access_policy    ON tenants;

-- ------------------------------------------------------------------------------------------------
-- application_role — strict tenant isolation (fail-closed)
-- ------------------------------------------------------------------------------------------------
-- USING clause (read filter): rows visible in SELECT, UPDATE, DELETE.
-- WITH CHECK clause (write filter): rows allowed in INSERT, UPDATE.
-- Both must pass for any operation to succeed.
--
-- An application bug that fails to call set_tenant_context() will result
-- in an empty result set (SELECT) or a policy violation error (INSERT/UPDATE)
-- — never in cross-tenant data exposure.
-- ------------------------------------------------------------------------------------------------
CREATE POLICY tenant_isolation_policy
  ON tenants
  FOR ALL
  TO application_role
  USING (
    id         = current_tenant_id()
    AND deleted_at IS NULL
  )
  WITH CHECK (
    id         = current_tenant_id()
    AND deleted_at IS NULL
  );

-- ------------------------------------------------------------------------------------------------
-- admin_role — unrestricted access
-- ------------------------------------------------------------------------------------------------
-- Admin tooling needs to see soft-deleted rows for audit, restore, and
-- hard-delete retention workflows. No deleted_at filter here.
-- Explicit policy (not BYPASSRLS) for auditability via pg_policies.
-- ------------------------------------------------------------------------------------------------
CREATE POLICY admin_full_access_policy
  ON tenants
  FOR ALL
  TO admin_role
  USING (TRUE)
  WITH CHECK (TRUE);

-- ------------------------------------------------------------------------------------------------
-- readonly_role — SELECT on active rows only
-- ------------------------------------------------------------------------------------------------
-- Reporting and export jobs should not see deleted tenant data unless they
-- specifically switch to admin_role (which requires explicit GRANT).
-- ------------------------------------------------------------------------------------------------
CREATE POLICY readonly_access_policy
  ON tenants
  FOR SELECT
  TO readonly_role
  USING (deleted_at IS NULL);

-- ------------------------------------------------------------------------------------------------
-- POLICY COMMENTS
-- ------------------------------------------------------------------------------------------------
COMMENT ON POLICY admin_full_access_policy ON tenants IS
  'Full unrestricted access for admin_role including soft-deleted rows. '
  'Explicit policy preferred over BYPASSRLS for auditability in pg_policies.';

COMMENT ON POLICY readonly_access_policy   ON tenants IS
  'SELECT-only access for readonly_role. Active (non-deleted) rows only.';
