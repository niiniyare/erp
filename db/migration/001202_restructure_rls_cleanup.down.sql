-- Restore verbose IS NOT NULL guard (matches pre-001202 state).
DROP POLICY IF EXISTS tenant_isolation_policy ON finance_account_balances;
CREATE POLICY tenant_isolation_policy ON finance_account_balances
  FOR ALL TO application_role
  USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

DROP POLICY IF EXISTS readonly_policy ON finance_account_balances;
CREATE POLICY finance_account_balances_ro_select ON finance_account_balances
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
