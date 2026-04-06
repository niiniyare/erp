-- Revert to raw current_setting() (unsafe — restores the original vulnerability).
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
        DROP POLICY IF EXISTS policy_evaluations_admin_access     ON policy_evaluations;
        DROP POLICY IF EXISTS policy_evaluations_ro_select        ON policy_evaluations;

        EXECUTE $p$
            CREATE POLICY policy_evaluations_tenant_isolation
                ON policy_evaluations
                FOR ALL
                TO application_role
                USING (
                    current_setting('app.current_tenant_id', true) <> ''
                    AND tenant_id = current_setting('app.current_tenant_id', true)::UUID
                )
                WITH CHECK (
                    current_setting('app.current_tenant_id', true) <> ''
                    AND tenant_id = current_setting('app.current_tenant_id', true)::UUID
                )
        $p$;

        EXECUTE $p$
            CREATE POLICY policy_evaluations_admin_access
                ON policy_evaluations
                FOR ALL
                TO admin_role
                USING (TRUE)
                WITH CHECK (TRUE)
        $p$;

        EXECUTE $p$
            ALTER TABLE policy_evaluations NO FORCE ROW LEVEL SECURITY
        $p$;
    END IF;
END
$$;
