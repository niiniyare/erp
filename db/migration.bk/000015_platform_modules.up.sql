-- ------------------------------------------------------------------------------------------------
-- MODULES TABLE
-- ------------------------------------------------------------------------------------------------
-- Organises system functionality into logical groups for permission management and feature control.
-- scope IN ('SYSTEM', 'TENANT'):
--   SYSTEM modules ship with the platform and are readable by all tenants.
--   TENANT modules are custom modules created by a specific tenant (tenant_id NOT NULL).
--
-- NOTE: FK constraint on tenant_id, updated_at trigger, and RLS policies are added in
--       migration 000062_platform_iam_constraints.up.sql (after tenants + trigger fn exist).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS modules (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID,        -- FK to tenants(id) added in 000062
  scope        VARCHAR(10)  NOT NULL DEFAULT 'SYSTEM'
                              CHECK (scope IN ('SYSTEM', 'TENANT')),
  slug         VARCHAR(50)  NOT NULL,   -- machine-readable key segment: 'finance', 'selling', 'iam'
  name         VARCHAR(50)  NOT NULL,
  display_name VARCHAR(100),
  description  TEXT,
  icon         VARCHAR(100),            -- nav icon class: 'fa fa-calculator'
  nav_order    INTEGER      NOT NULL DEFAULT 999,
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

COMMENT ON TABLE   modules             IS 'System modules for organising permissions and features. Enables modular permission management and feature toggles.';
COMMENT ON COLUMN  modules.tenant_id   IS 'NULL for SYSTEM-scope modules. Set for custom TENANT-scope modules. FK enforced in 000062.';
COMMENT ON COLUMN  modules.scope       IS 'SYSTEM = platform-wide, readable by all. TENANT = private custom module.';
COMMENT ON COLUMN  modules.slug        IS 'Machine-readable key segment used in dot-notation: {slug}.{resource}.{action}. e.g. ''finance'', ''selling''. Do not change after seeding — all permission/flag keys depend on it.';
COMMENT ON COLUMN  modules.icon        IS 'Icon class for sidebar nav. e.g. ''fa fa-calculator''.';
COMMENT ON COLUMN  modules.nav_order   IS 'Sidebar display order. Lower = higher. Default 999.';
COMMENT ON COLUMN  modules.category    IS 'Module category: CORE, HR, FINANCE, SALES, INVENTORY, etc.';
COMMENT ON COLUMN  modules.module_type IS 'Helps categorise industry-specific apps: CORE, INDUSTRY, EXTENSION, INTERNAL.';
COMMENT ON COLUMN  modules.version     IS 'Module version for tracking feature updates and compatibility.';
COMMENT ON COLUMN  modules.updated_at  IS 'Updated by trigger on every row change — use for cache invalidation.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_modules_scope     ON modules(scope, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_modules_tenant    ON modules(tenant_id)        WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_modules_category  ON modules(category)         WHERE is_active = TRUE;
CREATE INDEX idx_modules_nav_order ON modules(nav_order, is_active) WHERE is_active = TRUE;

-- Slug uniqueness: SYSTEM slugs globally unique; TENANT slugs unique per tenant.
CREATE UNIQUE INDEX idx_modules_slug_system ON modules(slug)            WHERE scope = 'SYSTEM';
CREATE UNIQUE INDEX idx_modules_slug_tenant ON modules(tenant_id, slug) WHERE scope = 'TENANT';
