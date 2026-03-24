-- ------------------------------------------------------------------------------------------------
-- SET_TENANT_CONTEXT — ACTIVE TENANT GUARD
-- ------------------------------------------------------------------------------------------------
-- Replaces the previous set_tenant_context() with a version that rejects non-ACTIVE tenants.
-- Previously the function accepted any tenant UUID that existed in the tenants table, including
-- PENDING tenants, allowing authentication before onboarding completed.
--
-- Fix: raises restrict_violation (23001) if the tenant's status is not 'ACTIVE'.
-- Must be called inside a transaction (set_config with is_local=true).
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_status TEXT;
BEGIN
    -- Fetch the tenant's status; raises NO DATA FOUND if it doesn't exist.
    SELECT "Status" INTO STRICT v_status
    FROM tenants
    WHERE id = p_tenant_id
      AND deleted_at IS NULL;

    -- Only allow ACTIVE tenants to have a session context.
    IF v_status <> 'ACTIVE' THEN
        RAISE EXCEPTION 'tenant % is not active (status: %)', p_tenant_id, v_status
            USING ERRCODE = 'restrict_violation';  -- 23001
    END IF;

    -- Set the session-local GUC that current_tenant_id() reads.
    PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, true);
END;
$$;

COMMENT ON FUNCTION set_tenant_context(UUID) IS
    'Sets app.current_tenant_id for the current transaction. '
    'Raises restrict_violation if the tenant is not ACTIVE. '
    'Must be called inside a transaction (set_config with is_local=true).';
