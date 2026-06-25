-- --------------------------------------------------------------------
-- 001201  Fix tenants."Status" → tenants.status
-- --------------------------------------------------------------------
-- The quoted "Status" column was an early mistake — quoted identifiers
-- are case-sensitive and require quoting in every query. Renamed to the
-- standard unquoted snake_case `status` to match every other table.
--
-- Impact: any SQLC queries, views, functions, or triggers that reference
-- tenants."Status" must be updated. Known affected objects:
--   • set_tenant_context()    — references "Status"
--   • validate_tenant_context() — references "Status"
--   • company_size CHECK constraint (same migration, different column)
-- --------------------------------------------------------------------

-- Step 1: rename the column.
ALTER TABLE tenants RENAME COLUMN "Status" TO status;

-- Step 2: re-create set_tenant_context() without the quoted column.
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
DECLARE
  v_status TEXT;
BEGIN
  SELECT status
    INTO v_status
    FROM tenants
   WHERE id         = p_tenant_id
     AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'set_tenant_context: tenant not found or has been deleted — id: %',
      p_tenant_id
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'set_tenant_context: tenant is not ACTIVE — id: %, current status: %',
      p_tenant_id, v_status
      USING ERRCODE = 'check_violation';
  END IF;

  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
$$;

-- Step 3: re-create validate_tenant_context() without the quoted column.
CREATE OR REPLACE FUNCTION validate_tenant_context()
  RETURNS UUID
  LANGUAGE plpgsql
  STABLE
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
DECLARE
  v_tid    UUID;
  v_status TEXT;
BEGIN
  v_tid := current_tenant_id();

  IF v_tid IS NULL THEN
    RAISE EXCEPTION
      'validate_tenant_context: no tenant context is set'
      USING ERRCODE = 'no_data_found';
  END IF;

  SELECT status
    INTO v_status
    FROM tenants
   WHERE id         = v_tid
     AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'validate_tenant_context: tenant no longer exists — id: %', v_tid
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'validate_tenant_context: tenant not ACTIVE — id: %, status: %',
      v_tid, v_status
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN v_tid;
END;
$$;

-- Step 4: fix company_size CHECK — existing check allows 'STARTUP' but
-- the domain and EntityDefinition use 'MICRO'. Align them.
ALTER TABLE tenants
  DROP CONSTRAINT IF EXISTS tenants_company_size_check;

ALTER TABLE tenants
  ADD CONSTRAINT tenants_company_size_check
  CHECK (company_size IN ('MICRO','SMALL','MEDIUM','LARGE','ENTERPRISE'));

-- Step 5: fix plan_tier CHECK — add FREE tier that EntityDefinition declares.
ALTER TABLE tenants
  DROP CONSTRAINT IF EXISTS tenants_plan_tier_check;

ALTER TABLE tenants
  ADD CONSTRAINT tenants_plan_tier_check
  CHECK (plan_tier IN ('FREE','STARTER','GROWTH','PROFESSIONAL','ENTERPRISE'));

COMMENT ON COLUMN tenants.status IS
  'Lifecycle state (was "Status" — quoted identifier removed in 001201). '
  'PENDING→ACTIVE→SUSPENDED→ARCHIVED. Set by set_tenant_context validation.';
