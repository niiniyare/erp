-- =============================================================================
-- MIGRATION 006 UP: Row Level Security Policies
-- =============================================================================
-- Architecture Decision (ADR-016): Fail-closed by design
--   The application_role policy contains NO fallback clause like
--   "OR current_tenant_id() IS NULL". This is intentional. When no tenant
--   context is set, application_role sees ZERO rows — it does not get a
--   list of all tenants. Fail-open RLS (the v1 behaviour) is a data
--   exfiltration vulnerability in pooled connection environments where a
--   missed set_tenant_context() call exposes the full tenant table.
--
-- Architecture Decision (ADR-017): Soft-delete exclusion in the policy
--   Soft-deleted rows are excluded in the RLS policy itself, not just in
--   set_tenant_context(). This is defense-in-depth: if an application bug
--   bypasses the function and sets a deleted tenant's ID directly via
--   SET LOCAL, the policy still prevents access to that row.
--
-- Architecture Decision (ADR-018): admin_role BYPASSRLS vs explicit policy
--   We use an explicit USING (TRUE) policy for admin_role rather than
--   ALTER ROLE admin_role BYPASSRLS. The explicit policy is visible in
--   pg_policies and therefore auditable. BYPASSRLS is invisible in standard
--   role inspection and easier to accidentally grant broadly.
--   The tradeoff: explicit policies require the table to have RLS enabled,
--   which adds marginal overhead per row — acceptable for an admin path
--   that is never in the hot query path.
-- =============================================================================

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- Force RLS even for the table owner (typically the migration user).
-- Without this, the table owner bypasses all policies.
-- Only enable this if your migration user is NOT a superuser.
-- ALTER TABLE tenants FORCE ROW LEVEL SECURITY;

-- Drop all existing policies before redefining — idempotent re-runs.
DROP POLICY IF EXISTS tenant_isolation_policy   ON tenants;
DROP POLICY IF EXISTS admin_full_access_policy  ON tenants;
DROP POLICY IF EXISTS readonly_access_policy    ON tenants;

-- -------------------------------------------------------------------------
-- application_role — strict tenant isolation (fail-closed)
-- -------------------------------------------------------------------------
-- USING clause (read filter): rows visible in SELECT, UPDATE, DELETE.
-- WITH CHECK clause (write filter): rows allowed in INSERT, UPDATE.
-- Both must pass for any operation to succeed.
--
-- An application bug that fails to call set_tenant_context() will result
-- in an empty result set (SELECT) or a policy violation error (INSERT/UPDATE)
-- — never in cross-tenant data exposure.
-- -------------------------------------------------------------------------
CREATE POLICY tenant_isolation_policy
  ON tenants
  FOR ALL
  TO application_role
  USING (
    -- Row must belong to the session's tenant AND must not be soft-deleted.
    -- current_tenant_id() returns NULL when no context is set,
    -- so NULL = NULL evaluates to NULL (not TRUE) — the row is excluded.
    id         = current_tenant_id()
    AND deleted_at IS NULL
  )
  WITH CHECK (
    id         = current_tenant_id()
    AND deleted_at IS NULL
  );

-- -------------------------------------------------------------------------
-- admin_role — unrestricted access (see ADR-018)
-- -------------------------------------------------------------------------
-- Admin tooling needs to see soft-deleted rows for audit, restore,
-- and hard-delete retention workflows. No deleted_at filter here.
-- -------------------------------------------------------------------------
CREATE POLICY admin_full_access_policy
  ON tenants
  FOR ALL
  TO admin_role
  USING (TRUE)
  WITH CHECK (TRUE);

-- -------------------------------------------------------------------------
-- readonly_role — SELECT on active rows only
-- -------------------------------------------------------------------------
-- Reporting and export jobs should not see deleted tenant data unless they
-- specifically switch to admin_role (which requires explicit GRANT).
-- -------------------------------------------------------------------------
CREATE POLICY readonly_access_policy
  ON tenants
  FOR SELECT
  TO readonly_role
  USING (deleted_at IS NULL);

-- -------------------------------------------------------------------------
-- POLICY COMMENTS
-- -------------------------------------------------------------------------
COMMENT ON POLICY tenant_isolation_policy  ON tenants IS
  'Strict tenant isolation for application_role. '
  'Fail-closed: zero rows visible when no context is set. '
  'Soft-deleted rows excluded at policy level (defense-in-depth).';

COMMENT ON POLICY admin_full_access_policy ON tenants IS
  'Full unrestricted access for admin_role including soft-deleted rows. '
  'See ADR-018: explicit policy preferred over BYPASSRLS for auditability.';

COMMENT ON POLICY readonly_access_policy   ON tenants IS
  'SELECT-only access for readonly_role. Active (non-deleted) rows only.';
