-- =============================================================================
-- ROLLBACK: Remove search_path pins from SECURITY DEFINER functions
-- =============================================================================
-- This rollback restores functions to their pre-pin state.
-- WARNING: Rolling back this migration re-introduces the search path
-- injection vulnerability. Only do this for testing purposes.
-- =============================================================================

-- Restore set_tenant_context WITHOUT search_path pin
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
DECLARE v_status TEXT;
BEGIN
  SELECT "Status" INTO v_status FROM tenants
   WHERE id = p_tenant_id AND deleted_at IS NULL;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not found: %', p_tenant_id;
  END IF;
  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not active: %', p_tenant_id;
  END IF;
  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
$$;
