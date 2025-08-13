-- Tenant context validation
CREATE OR REPLACE FUNCTION validate_and_set_tenant_context(p_tenant_id UUID)
RETURNS TABLE(tenant_name TEXT, tenant_status TEXT) AS $$
DECLARE
    v_tenant_record RECORD;
BEGIN
    -- Validate and fetch tenant information
    SELECT id, name, status, deleted_at, last_activity_at
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

    -- Update last activity
    UPDATE tenants 
    SET last_activity_at = NOW() 
    WHERE id = p_tenant_id;

    -- Set tenant context
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);

    -- Return tenant information
    RETURN QUERY SELECT v_tenant_record.name, v_tenant_record.status;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;