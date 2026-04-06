-- ------------------------------------------------------------------------------------------------
-- POLICY_EVALUATIONS — FIX RLS TO USE CURRENT_TENANT_ID()
-- ------------------------------------------------------------------------------------------------
-- Replaces raw current_setting() calls in policy_evaluations RLS policies with the
-- current_tenant_id() helper function for consistency with all other tables.
-- Guarded by a DO $$ block so this migration is a no-op when policy_evaluations does not exist.
-- ------------------------------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = 'policy_evaluations'
          AND n.nspname = current_schema()
    ) THEN
        DROP POLICY IF EXISTS policy_evaluations_tenant_isolation ON policy_evaluations;
        DROP POLICY IF EXISTS policy_evaluations_rls_policy       ON policy_evaluations;
        DROP POLICY IF EXISTS policy_evaluations_tenant_rls       ON policy_evaluations;

        EXECUTE $p$
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
                )
        $p$;

        DROP POLICY IF EXISTS policy_evaluations_admin_access ON policy_evaluations;

        EXECUTE $p$
            CREATE POLICY policy_evaluations_admin_access
                ON policy_evaluations
                FOR ALL
                TO admin_role
                USING (TRUE)
                WITH CHECK (TRUE)
        $p$;

        EXECUTE $p$
            ALTER TABLE policy_evaluations FORCE ROW LEVEL SECURITY
        $p$;

        DROP POLICY IF EXISTS policy_evaluations_ro_select ON policy_evaluations;

        EXECUTE $p$
            CREATE POLICY policy_evaluations_ro_select
                ON policy_evaluations
                FOR SELECT
                TO readonly_role
                USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
        $p$;
    END IF;
END
$$;
