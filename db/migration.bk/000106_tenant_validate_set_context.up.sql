-- ------------------------------------------------------------------------------------------------
-- VALIDATE_AND_SET_TENANT_CONTEXT
-- ------------------------------------------------------------------------------------------------
-- Validates that a tenant exists, is not soft-deleted, and is in an operational status before
-- setting the app.current_tenant_id session variable used by RLS policies.
-- Allowed statuses: 'active', 'pending'. Raises an exception for any other state.
--
-- NOTE: Depends on tenants(id), tenants.deleted_at, and tenants.last_activity_at columns
--       defined in the tenants core migrations. Uses SECURITY DEFINER so it can update
--       last_activity_at regardless of the caller's role.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION validate_and_set_tenant_context(p_tenant_id UUID)
RETURNS TABLE(tenant_name TEXT, tenant_status TEXT) AS $$
DECLARE
  v_tenant_record RECORD;
BEGIN
  -- Validate and fetch tenant information
  SELECT id, name, STATUS, deleted_at, last_activity_at
    INTO v_tenant_record
    FROM tenants
   WHERE id = p_tenant_id;

  -- Check if tenant exists
  IF v_tenant_record.id IS NULL THEN
    RAISE EXCEPTION 'Tenant not found: %', p_tenant_id;
  END IF;

  -- Check if tenant is soft-deleted
  IF v_tenant_record.deleted_at IS NOT NULL THEN
    RAISE EXCEPTION 'Tenant is deleted: %', p_tenant_id;
  END IF;

  -- Check tenant status
  IF v_tenant_record.status NOT IN ('active', 'pending') THEN
    RAISE EXCEPTION 'Tenant is not active: % (status: %)', p_tenant_id, v_tenant_record.status;
  END IF;

  -- Update last activity timestamp
  UPDATE tenants
     SET last_activity_at = NOW()
   WHERE id = p_tenant_id;

  -- Set tenant context for the current transaction
  PERFORM set_config('app.current_tenant_id', p_tenant_id::text, TRUE);

  -- Return tenant information to caller
  RETURN QUERY
  SELECT v_tenant_record.name,
         v_tenant_record.status;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
