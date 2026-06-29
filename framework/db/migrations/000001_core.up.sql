-- ============================================================
-- 000001_core — framework SQL functions and RLS helpers
-- ============================================================

-- ── Tenant-context functions ─────────────────────────────────

CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $func$
  SELECT current_setting('app.current_tenant_id', true)::uuid
$func$;

COMMENT ON FUNCTION current_tenant_id() IS
  'Returns the tenant UUID for the current transaction, set by set_tenant_context().
   Returns NULL (not an error) when called outside a tenant context.';

CREATE OR REPLACE FUNCTION set_tenant_context(tid uuid)
RETURNS void
LANGUAGE sql
AS $func$
  SELECT set_config('app.current_tenant_id', tid::text, true)
$func$;

COMMENT ON FUNCTION set_tenant_context(uuid) IS
  'Binds tid as the current tenant for the duration of the transaction.
   Must be called once inside every write transaction before any DML.
   The setting is transaction-local (true flag) and cleared on commit/rollback.';

-- ── Database roles (idempotent) ──────────────────────────────

DO $body$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'application_role') THEN
    CREATE ROLE application_role NOLOGIN;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'admin_role') THEN
    CREATE ROLE admin_role NOLOGIN;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'readonly_role') THEN
    CREATE ROLE readonly_role NOLOGIN;
  END IF;
END;
$body$;

-- ── RLS helper: apply standard tenant isolation to a table ────

CREATE OR REPLACE FUNCTION apply_tenant_rls(tbl regclass)
RETURNS void
LANGUAGE plpgsql
AS $func$
BEGIN
  EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', tbl);
  EXECUTE format('ALTER TABLE %s FORCE  ROW LEVEL SECURITY', tbl);

  EXECUTE format($policy$
    CREATE POLICY tenant_isolation_policy ON %s
      FOR ALL TO application_role
      USING  (tenant_id = current_tenant_id())
      WITH CHECK (tenant_id = current_tenant_id())
  $policy$, tbl);

  EXECUTE format($policy$
    CREATE POLICY admin_full_access_policy ON %s
      FOR ALL TO admin_role
      USING (TRUE) WITH CHECK (TRUE)
  $policy$, tbl);

  EXECUTE format($policy$
    CREATE POLICY readonly_policy ON %s
      FOR SELECT TO readonly_role
      USING (tenant_id = current_tenant_id())
  $policy$, tbl);
END;
$func$;

COMMENT ON FUNCTION apply_tenant_rls(regclass) IS
  'Enables RLS on tbl and installs the three standard tenant-isolation policies.
   Call once per tenant-scoped table in its CREATE TABLE migration.';

-- ── RLS helper: global (non-tenant) tables ────────────────────

CREATE OR REPLACE FUNCTION apply_global_rls(tbl regclass)
RETURNS void
LANGUAGE plpgsql
AS $func$
BEGIN
  EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', tbl);
  EXECUTE format('ALTER TABLE %s FORCE  ROW LEVEL SECURITY', tbl);

  EXECUTE format($policy$
    CREATE POLICY admin_full_access_policy ON %s
      FOR ALL TO admin_role
      USING (TRUE) WITH CHECK (TRUE)
  $policy$, tbl);

  EXECUTE format($policy$
    CREATE POLICY app_read_policy ON %s
      FOR SELECT TO application_role USING (TRUE)
  $policy$, tbl);

  EXECUTE format($policy$
    CREATE POLICY readonly_policy ON %s
      FOR SELECT TO readonly_role USING (TRUE)
  $policy$, tbl);
END;
$func$;

COMMENT ON FUNCTION apply_global_rls(regclass) IS
  'Enables RLS on tbl with policies appropriate for global (non-tenant) tables.
   application_role gets SELECT only; admin_role gets full access.';
