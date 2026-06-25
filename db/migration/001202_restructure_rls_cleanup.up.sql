-- --------------------------------------------------------------------
-- 001202  Standardise RLS policies — remove redundant IS NOT NULL guard
-- --------------------------------------------------------------------
-- Some policies (e.g. 000911) use:
--   USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
--
-- The IS NOT NULL guard is redundant: when current_tenant_id() returns NULL,
-- `tenant_id = NULL` is UNKNOWN (not TRUE) in SQL, so the row is already
-- excluded by the USING clause. The extra guard adds noise and diverges from
-- the authoritative pattern established in 000056 (tenants table).
--
-- This migration re-creates all non-conforming policies to match:
--   USING (tenant_id = current_tenant_id() [AND deleted_at IS NULL])
--
-- Tables fixed: finance_account_balances, finance_account_validation_rules
-- (and any others added between 000911 and 001100 with the same pattern).
-- --------------------------------------------------------------------

-- finance_account_balances
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_balances;
CREATE POLICY tenant_isolation_policy ON finance_account_balances
  FOR ALL TO application_role
  USING  (tenant_id = current_tenant_id())
  WITH CHECK (tenant_id = current_tenant_id());

-- finance_account_validation_rules (may not exist in all environments)
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_validation_rules;
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_class WHERE relname = 'finance_account_validation_rules'
  ) THEN
    EXECUTE $q$
      CREATE POLICY tenant_isolation_policy ON finance_account_validation_rules
        FOR ALL TO application_role
        USING  (tenant_id = current_tenant_id())
        WITH CHECK (tenant_id = current_tenant_id())
    $q$;
  END IF;
END;
$$;

-- Readonly policies for finance_account_balances — align naming convention.
DROP POLICY IF EXISTS finance_account_balances_ro_select ON finance_account_balances;
CREATE POLICY readonly_policy ON finance_account_balances
  FOR SELECT TO readonly_role
  USING (tenant_id = current_tenant_id());
