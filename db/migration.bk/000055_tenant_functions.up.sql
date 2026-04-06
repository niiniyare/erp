-- ------------------------------------------------------------------------------------------------
-- TENANT CONTEXT FUNCTIONS
-- ------------------------------------------------------------------------------------------------
-- Transaction-local session functions for setting, reading, and clearing the current tenant.
-- These functions are the sole mechanism for establishing tenant context before any
-- RLS-protected query. They are designed for pooled connections — context is cleared
-- automatically on COMMIT or ROLLBACK (is_local = TRUE in set_config).
--
-- Function inventory:
--   A. current_tenant_id()       — STABLE, INVOKER: reads app.current_tenant_id from session
--   B. set_tenant_context(UUID)  — DEFINER, validates tenant, sets context
--   C. clear_tenant_context()    — INVOKER: clears context (defensive pool hygiene)
--   D. validate_tenant_context() — DEFINER, STABLE: re-validates mid-transaction
--
-- Security design:
--   • set_config(..., TRUE) uses transaction-local scope — cleared on COMMIT/ROLLBACK.
--   • All SECURITY DEFINER functions pin SET search_path = pg_catalog, public to prevent
--     search path injection (a malicious user cannot shadow the tenants table).
--   • current_tenant_id() uses SECURITY INVOKER — it reads only a session variable,
--     so no privilege escalation is needed or desired.
--   • Only app.current_tenant_id is stored in session state. Old app.tenant_status and
--     app.context_set_at variables are intentionally omitted — they were never read
--     by any RLS policy and created misleading noise in pg_settings output.
--
-- NOTE: GRANTs for these functions are in migration 000059.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- A. current_tenant_id()
-- ------------------------------------------------------------------------------------------------
-- Returns the UUID of the tenant set for the current transaction.
-- Returns NULL if no context is set or the stored value is empty/invalid.
-- STABLE because it reads a session variable — safe to call multiple times
-- in a single query without re-execution. Inlined by the planner into RLS.
--
-- Error handling:
--   WHEN invalid_text_representation — catches malformed UUID strings.
--   Returning NULL on a bad UUID is correct because this function runs
--   inside RLS policies — a hard error would crash every query for the
--   session rather than simply denying access.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION current_tenant_id()
  RETURNS UUID
  LANGUAGE plpgsql
  STABLE
  SECURITY INVOKER
  -- No SECURITY DEFINER here — this function reads only a session variable,
  -- not any table, so no privilege escalation is needed or desired.
AS $$
BEGIN
  RETURN COALESCE(
    NULLIF(current_setting('app.current_tenant_id', TRUE), ''),
    NULL
  )::UUID;

EXCEPTION
  WHEN invalid_text_representation THEN
    -- Emit a WARNING so the malformed value appears in server logs.
    -- The calling session will continue with NULL tenant context (= no access).
    RAISE WARNING
      'current_tenant_id(): session variable "app.current_tenant_id" '
      'contains an invalid UUID — returning NULL. '
      'Investigate how this value was set.';
    RETURN NULL;
END;
$$;

COMMENT ON FUNCTION current_tenant_id() IS
  'Returns the UUID stored in app.current_tenant_id for the current '
  'transaction-local context. Returns NULL when no context is set, when '
  'the variable is empty, or when the value is not a valid UUID (emits WARNING). '
  'Called by RLS policies — never raises an exception.';

-- ------------------------------------------------------------------------------------------------
-- B. set_tenant_context(UUID)
-- ------------------------------------------------------------------------------------------------
-- Validates the tenant and sets the transaction-local context.
-- Raises an exception on any validation failure — callers must handle it.
--
-- Validation checks (in order):
--   1. Tenant exists and is not soft-deleted.
--   2. Tenant Status is ACTIVE (not SUSPENDED, PENDING, or ARCHIVED).
--
-- Only app.current_tenant_id is written. Context is automatically cleared
-- at transaction end — safe for connection pools.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_tenant_context(p_tenant_id UUID)
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public   -- search path pin: prevents shadow table injection
AS $$
DECLARE
  v_status TEXT;
BEGIN
  -- Single query: existence + soft-delete check + status fetch.
  -- NOT FOUND means either the ID doesn't exist OR the row is soft-deleted.
  SELECT "Status"
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

  -- TRUE = transaction-local: automatically cleared on COMMIT / ROLLBACK.
  -- This is the only safe behaviour for pooled connections.
  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
$$;

COMMENT ON FUNCTION set_tenant_context(UUID) IS
  'Sets the transaction-local tenant context after verifying the tenant '
  'exists (not soft-deleted) and has ACTIVE status. '
  'Context is automatically cleared at transaction end — safe for connection pools. '
  'Raises an exception with a descriptive message on any validation failure.';

-- ------------------------------------------------------------------------------------------------
-- C. clear_tenant_context()
-- ------------------------------------------------------------------------------------------------
-- Explicitly clears the tenant context. Call before returning a connection
-- to a pool as a defensive measure, even though transaction-local context
-- is cleared automatically at transaction end.
--
-- Uses empty string '' (not NULL) because set_config() requires TEXT.
-- NULLIF in current_tenant_id() translates '' back to NULL.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION clear_tenant_context()
  RETURNS VOID
  LANGUAGE plpgsql
  SECURITY INVOKER
  -- No SECURITY DEFINER — only writes a session variable, no table access.
AS $$
BEGIN
  PERFORM set_config('app.current_tenant_id', '', TRUE);
END;
$$;

COMMENT ON FUNCTION clear_tenant_context() IS
  'Clears the transaction-local tenant context. '
  'Defensive call before returning connections to a pool. '
  'Uses empty string (not NULL) as sentinel — current_tenant_id() '
  'translates it back to NULL via NULLIF.';

-- ------------------------------------------------------------------------------------------------
-- D. validate_tenant_context()
-- ------------------------------------------------------------------------------------------------
-- Re-validates the active tenant mid-transaction or mid-job.
-- Designed for long-running background jobs and pooled connections where
-- tenant status may change after the context was first set.
-- Returns the tenant UUID on success; raises on any failure.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION validate_tenant_context()
  RETURNS UUID
  LANGUAGE plpgsql
  STABLE
  SECURITY DEFINER
  SET search_path = pg_catalog, public   -- search path pin
AS $$
DECLARE
  v_tid    UUID;
  v_status TEXT;
BEGIN
  v_tid := current_tenant_id();

  IF v_tid IS NULL THEN
    RAISE EXCEPTION
      'validate_tenant_context: no tenant context is set — '
      'call set_tenant_context() before proceeding'
      USING ERRCODE = 'no_data_found';
  END IF;

  SELECT "Status"
    INTO v_status
    FROM tenants
   WHERE id         = v_tid
     AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION
      'validate_tenant_context: tenant no longer exists or has been '
      'soft-deleted — id: %', v_tid
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION
      'validate_tenant_context: tenant context is stale — '
      'id: %, current status: %', v_tid, v_status
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN v_tid;
END;
$$;

COMMENT ON FUNCTION validate_tenant_context() IS
  'Re-validates the current transaction-local tenant context against the '
  'live tenants table. Use in long-running jobs or pooled connections '
  'where tenant status might change after the context was first set. '
  'Returns the current tenant UUID on success; raises on any failure.';
