-- G4: Fix policy_evaluations RLS — replace raw current_setting() with current_tenant_id().
-- Guarded by DO $$ so the migration is a no-op when policy_evaluations does not exist.
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
    END IF;
END
$$;
