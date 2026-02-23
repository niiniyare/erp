-- ------------------------------------------------------------------------------------------------
-- MODULES TABLE
-- ------------------------------------------------------------------------------------------------
-- Organises system functionality into logical groups for permission management and feature control.
-- scope IN ('SYSTEM', 'TENANT'):
--   SYSTEM modules ship with the platform and are readable by all tenants.
--   TENANT modules are custom modules created by a specific tenant (tenant_id NOT NULL).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS modules (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID         REFERENCES tenants(id) ON DELETE CASCADE,
  scope        VARCHAR(10)  NOT NULL DEFAULT 'SYSTEM'
                              CHECK (scope IN ('SYSTEM', 'TENANT')),
  name         VARCHAR(50)  NOT NULL,
  display_name VARCHAR(100),
  description  TEXT,
  category     VARCHAR(50),   -- 'CORE', 'HR', 'FINANCE', 'SALES', etc.
  module_type  VARCHAR(30),   -- 'CORE', 'INDUSTRY', 'EXTENSION', 'INTERNAL'
  version      VARCHAR(20),
  is_active    BOOLEAN      DEFAULT TRUE,
  created_at   TIMESTAMPTZ  DEFAULT NOW(),
  updated_at   TIMESTAMPTZ  DEFAULT NOW(),
  -- SYSTEM modules must have tenant_id=NULL; TENANT modules must have a tenant_id
  CONSTRAINT modules_scope_tenant_check CHECK (
    (scope = 'SYSTEM' AND tenant_id IS NULL)
    OR (scope = 'TENANT' AND tenant_id IS NOT NULL)
  ),
  -- Module names are unique within scope+tenant
  CONSTRAINT modules_name_scope_unique UNIQUE (name, scope, tenant_id)
);

COMMENT ON TABLE   modules            IS 'System modules for organising permissions and features. Enables modular permission management and feature toggles.';
COMMENT ON COLUMN  modules.tenant_id  IS 'NULL for SYSTEM-scope modules. Set for custom TENANT-scope modules.';
COMMENT ON COLUMN  modules.scope      IS 'SYSTEM = platform-wide, readable by all. TENANT = private custom module.';
COMMENT ON COLUMN  modules.category   IS 'Module category: CORE, HR, FINANCE, SALES, INVENTORY, etc.';
COMMENT ON COLUMN  modules.module_type IS 'Helps categorise industry-specific apps: CORE, INDUSTRY, EXTENSION, INTERNAL.';
COMMENT ON COLUMN  modules.version    IS 'Module version for tracking feature updates and compatibility.';
COMMENT ON COLUMN  modules.updated_at IS 'Updated by trigger on every row change — use for cache invalidation.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_modules_scope    ON modules(scope, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_modules_tenant   ON modules(tenant_id)        WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_modules_category ON modules(category)         WHERE is_active = TRUE;

-- ------------------------------------------------------------------------------------------------
-- updated_at TRIGGER
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_modules_updated_at
  BEFORE UPDATE ON modules
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;

-- application_role: see all SYSTEM modules + their own TENANT modules
CREATE POLICY modules_read ON modules
  FOR SELECT TO application_role
  USING (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  );

-- application_role: insert/update only their own TENANT modules
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
