-- Revert status → "Status" and restore original constraints.
ALTER TABLE tenants RENAME COLUMN status TO "Status";

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_company_size_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_company_size_check
  CHECK (company_size IN ('STARTUP','SMALL','MEDIUM','LARGE','ENTERPRISE'));

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_plan_tier_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_plan_tier_check
  CHECK (plan_tier IN ('STARTER','GROWTH','PROFESSIONAL','ENTERPRISE'));

-- Restore original functions referencing "Status".
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID LANGUAGE plpgsql SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
DECLARE v_status TEXT;
BEGIN
  SELECT "Status" INTO v_status FROM tenants
   WHERE id = p_tenant_id AND deleted_at IS NULL;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not found — id: %', p_tenant_id
      USING ERRCODE = 'no_data_found';
  END IF;
  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not ACTIVE — id: %, status: %',
      p_tenant_id, v_status USING ERRCODE = 'check_violation';
  END IF;
  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
$$;
