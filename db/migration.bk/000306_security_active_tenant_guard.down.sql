-- Revert to the original function that does NOT check tenant status.
-- WARNING: this restores the PENDING-tenant security gap.
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, true);
END;
$$;
