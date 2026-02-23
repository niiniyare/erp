-- =============================================================================
-- MIGRATION 006 DOWN: Row Level Security Policies
-- =============================================================================
-- Disabling RLS will expose ALL rows to all roles immediately.
-- Only run this in a maintenance window or in a development teardown.
-- =============================================================================

DROP POLICY IF EXISTS readonly_access_policy    ON tenants;
DROP POLICY IF EXISTS admin_full_access_policy  ON tenants;
DROP POLICY IF EXISTS tenant_isolation_policy   ON tenants;

ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;

