-- Revert to raw current_setting() (unsafe — restores the original vulnerability).
DROP POLICY IF EXISTS policy_evaluations_tenant_isolation ON policy_evaluations;
DROP POLICY IF EXISTS policy_evaluations_admin_access     ON policy_evaluations;

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
    );

CREATE POLICY policy_evaluations_admin_access
    ON policy_evaluations
    FOR ALL
    TO admin_role
    USING (TRUE)
    WITH CHECK (TRUE);
