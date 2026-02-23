-- ------------------------------------------------------------------------------------------------
-- PLATFORM IAM: DEFERRED CONSTRAINTS, TRIGGERS, RLS, AND GRANTS
-- ------------------------------------------------------------------------------------------------
-- Tables modules, resources, and actions were created in migrations 000015–000017 (early,
-- before tenants/roles/functions existed). This migration wires them up properly now that:
--   • tenants table exists        (000053)
--   • current_tenant_id() exists  (000055)
--   • update_updated_at_column()  (000057)
--   • roles table exists          (000405)
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- FK CONSTRAINTS
-- ------------------------------------------------------------------------------------------------
ALTER TABLE modules
  ADD CONSTRAINT fk_modules_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

ALTER TABLE actions
  ADD CONSTRAINT fk_actions_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- NOTE: fk_actions_approver_role (roles table) is added in 000413 after roles table exists.

-- ------------------------------------------------------------------------------------------------
-- updated_at TRIGGERS
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_modules_updated_at
  BEFORE UPDATE ON modules
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_resources_updated_at
  BEFORE UPDATE ON resources
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_actions_updated_at
  BEFORE UPDATE ON actions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY — modules
-- ------------------------------------------------------------------------------------------------
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;

-- application_role: see all SYSTEM modules + their own TENANT modules
CREATE POLICY modules_read ON modules
  FOR SELECT TO application_role
  USING (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  );

CREATE POLICY modules_write ON modules
  FOR INSERT TO application_role
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY modules_update ON modules
  FOR UPDATE TO application_role
  USING  (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY modules_admin ON modules
  FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

GRANT SELECT, INSERT, UPDATE, DELETE ON modules TO application_role;

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY — actions
-- ------------------------------------------------------------------------------------------------
ALTER TABLE actions ENABLE ROW LEVEL SECURITY;

-- application_role: see all SYSTEM actions + their own TENANT actions
CREATE POLICY actions_read ON actions
  FOR SELECT TO application_role
  USING (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  );

CREATE POLICY actions_write ON actions
  FOR INSERT TO application_role
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY actions_update ON actions
  FOR UPDATE TO application_role
  USING  (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY actions_admin ON actions
  FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

GRANT SELECT, INSERT, UPDATE, DELETE ON actions TO application_role;
