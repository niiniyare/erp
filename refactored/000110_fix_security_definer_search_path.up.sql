-- =============================================================================
-- FIX: Pin search_path on all SECURITY DEFINER functions
-- =============================================================================
-- WHY: SECURITY DEFINER functions run with the privileges of their owner.
--      Without a pinned search_path, a user can CREATE a table or view in
--      their own schema that shadows a real platform table (e.g. "tenants"),
--      causing the function to operate on fake data.
--
-- HOW: We re-declare each affected function with SET search_path = pg_catalog, public
--      using CREATE OR REPLACE — this is a non-destructive in-place update.
--      Function signatures and bodies are unchanged.
--
-- AFFECTED FUNCTIONS (detected by Script 1):
-- Source: 000106_validate_set_tenant_context.up.sql
-- Source: 000451_audit_funcs.up.sql
--   NEEDS PIN: get_current_user_context
--   NEEDS PIN: calculate_audit_risk_score
--   NEEDS PIN: extract_compliance_flags
--   NEEDS PIN: determine_event_category
--   NEEDS PIN: determine_severity
--   NEEDS PIN: audit_trigger_function
--   NEEDS PIN: enable_audit_on_table
--   NEEDS PIN: disable_audit_on_table
--   NEEDS PIN: enable_audit_on_schema
--   NEEDS PIN: get_audit_statistics
-- Source: 000502_user_functions_triggers.up.sql
-- Source: 000503_user_add_on.up.sql
--   NEEDS PIN: assess_session_risk
--   NEEDS PIN: terminate_risky_sessions
--   NEEDS PIN: gdpr_user_deletion
--   NEEDS PIN: enforce_data_retention
--   NEEDS PIN: trigger_security_notification
-- Source: 000804_feature_flag_functions.up.sql
-- Source: 000805_feature_flag_cache.up.sql
-- Source: 000806_feature_flag_cleanup.up.sql
-- =============================================================================

-- set_tenant_context: reads tenants table — must be pinned
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
DECLARE
  v_status TEXT;
BEGIN
  SELECT "Status"
    INTO v_status
    FROM tenants
   WHERE id = p_tenant_id
     AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'set_tenant_context: tenant not found or has been deleted: %',
      p_tenant_id USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'set_tenant_context: tenant is not ACTIVE: % (status: %)',
      p_tenant_id, v_status USING ERRCODE = 'check_violation';
  END IF;

  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
$$;

COMMENT ON FUNCTION set_tenant_context(UUID) IS
  'Sets transaction-local tenant context. '
  'SECURITY DEFINER with search_path pinned to prevent shadow table injection.';

-- validate_tenant_context: reads tenants table — must be pinned
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
    RAISE EXCEPTION 'validate_tenant_context: no tenant context is set'
      USING ERRCODE = 'no_data_found';
  END IF;

  SELECT "Status"
    INTO v_status
    FROM tenants
   WHERE id = v_tid AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'validate_tenant_context: tenant no longer exists or deleted: %', v_tid
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'validate_tenant_context: stale context — status is %', v_status
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN v_tid;
END;
$$;

-- check_subdomain_not_reserved: reads reserved_subdomains — must be pinned
CREATE OR REPLACE FUNCTION check_subdomain_not_reserved()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
BEGIN
  IF NEW.subdomain IS NULL THEN RETURN NEW; END IF;
  IF TG_OP = 'UPDATE'
     AND OLD.subdomain IS NOT DISTINCT FROM NEW.subdomain THEN
    RETURN NEW;
  END IF;
  IF EXISTS (SELECT 1 FROM reserved_subdomains WHERE subdomain = NEW.subdomain) THEN
    RAISE EXCEPTION
      'Subdomain "%" is reserved for platform infrastructure.',
      NEW.subdomain USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;

-- check_tenant_hierarchy_depth: reads tenants — must be pinned
CREATE OR REPLACE FUNCTION check_tenant_hierarchy_depth()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public
AS $$
DECLARE
  v_max_depth CONSTANT INTEGER := 5;
  v_current   UUID;
  v_depth     INTEGER := 0;
BEGIN
  IF NEW.parent_tenant_id IS NULL THEN RETURN NEW; END IF;
  IF NEW.parent_tenant_id = NEW.id THEN
    RAISE EXCEPTION 'Tenant cannot reference itself as parent.'
      USING ERRCODE = 'check_violation';
  END IF;
  v_current := NEW.parent_tenant_id;
  WHILE v_current IS NOT NULL LOOP
    v_depth := v_depth + 1;
    IF v_depth > v_max_depth THEN
      RAISE EXCEPTION 'Hierarchy depth exceeds maximum of %', v_max_depth
        USING ERRCODE = 'check_violation';
    END IF;
    IF v_current = NEW.id THEN
      RAISE EXCEPTION 'Circular parent_tenant_id reference detected.'
        USING ERRCODE = 'check_violation';
    END IF;
    SELECT parent_tenant_id INTO v_current FROM tenants WHERE id = v_current;
  END LOOP;
  RETURN NEW;
END;
$$;
