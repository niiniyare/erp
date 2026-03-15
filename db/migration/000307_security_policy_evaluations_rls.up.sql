-- ================================================================================================
-- G4: Fix policy_evaluations RLS — replace raw current_setting() with current_tenant_id().
--
-- The original RLS policy used:
--   current_setting('app.current_tenant_id')::UUID
-- directly.  This is unsafe because current_setting() raises an ERROR when the
-- GUC is unset (e.g. superuser connections, background workers, migrations) and
-- the cast to UUID raises an error on empty string.
--
-- current_tenant_id() is the approved wrapper that returns NULL safely when the
-- setting is absent.  NULL propagation means the policy USING clause evaluates
-- to NULL (treated as false) which safely denies access rather than erroring.
-- ================================================================================================

-- Drop the existing unsafe policies.
DROP POLICY IF EXISTS policy_evaluations_tenant_isolation  ON policy_evaluations;
DROP POLICY IF EXISTS policy_evaluations_rls_policy        ON policy_evaluations;
DROP POLICY IF EXISTS policy_evaluations_tenant_rls        ON policy_evaluations;

-- Re-create with the safe wrapper.
CREATE POLICY policy_evaluations_tenant_isolation
    ON policy_evaluations
    FOR ALL
    TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass remains unchanged.
DROP POLICY IF EXISTS policy_evaluations_admin_access ON policy_evaluations;

CREATE POLICY policy_evaluations_admin_access
    ON policy_evaluations
    FOR ALL
    TO admin_role
    USING (TRUE)
    WITH CHECK (TRUE);
