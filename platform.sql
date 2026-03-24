--db/migration/000015_platform_modules.up.sql-
-- 
-- MODULES TABLE
-- 
-- Organises system functionality into logical groups for permission management and feature control.
-- scope IN ('SYSTEM', 'TENANT'):
--   SYSTEM modules ship with the platform and are readable by all tenants.
--   TENANT modules are custom modules created by a specific tenant (tenant_id NOT NULL).
--
-- 
CREATE TABLE IF NOT EXISTS modules (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  -- tenant_id    UUID,        -- FK to tenants(id) added in 000062
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
COMMENT ON COLUMN  modules.tenant_id  IS 'NULL for SYSTEM-scope modules. Set for custom TENANT-scope modules. FK enforced in 000062.';
COMMENT ON COLUMN  modules.scope      IS 'SYSTEM = platform-wide, readable by all. TENANT = private custom module.';
COMMENT ON COLUMN  modules.category   IS 'Module category: CORE, HR, FINANCE, SALES, INVENTORY, etc.';
COMMENT ON COLUMN  modules.module_type IS 'Helps categorise industry-specific apps: CORE, INDUSTRY, EXTENSION, INTERNAL.';
COMMENT ON COLUMN  modules.version    IS 'Module version for tracking feature updates and compatibility.';
COMMENT ON COLUMN  modules.updated_at IS 'Updated by trigger on every row change — use for cache invalidation.';

-- 
-- INDEXES
-- 
CREATE INDEX idx_modules_scope    ON modules(scope, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_modules_tenant   ON modules(tenant_id)        WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_modules_category ON modules(category)         WHERE is_active = TRUE;
--
--
--db/migration/0000016_platform_resources.up.
--
-- 
-- RESOURCES TABLE
-- 
-- Defines system resources that can be protected by permissions (APIs, UI components, data, etc.).
--
-- 
CREATE TABLE IF NOT EXISTS resources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  display_name VARCHAR(150),
  description TEXT,
  resource_type VARCHAR(50) NOT NULL CHECK (
    resource_type IN (
      'API',
      'UI',
      'DATA',
      'FILE',
      'REPORT',
      'WORKFLOW',
      'FUNCTION'
    )
  ),
  parent_resource_id UUID REFERENCES resources(id),
  path VARCHAR(500),  -- URL path, API endpoint, file path, etc.
  resource_attributes JSONB DEFAULT '{}'::jsonb,  -- ABAC resource attributes
  is_active  BOOLEAN     DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE resources IS 'System resources that can be protected by permissions including APIs, UI components, data objects, files, reports, and workflows.';
COMMENT ON COLUMN resources.resource_type IS 'Type of resource: API, UI, DATA, FILE, REPORT, WORKFLOW, FUNCTION';
COMMENT ON COLUMN resources.parent_resource_id IS 'Self-referential for resource hierarchy (e.g., API endpoints under API group)';
COMMENT ON COLUMN resources.resource_attributes IS 'JSONB containing ABAC attributes like classification level, sensitivity, department ownership';
COMMENT ON COLUMN resources.path IS
  'Resource path convention: use dot-notation for logical resources (e.g. finance.invoice.create) '
  'and slash-notation for HTTP endpoints (e.g. /api/v1/invoices). Be consistent within a module. '
  'Consuming services must agree on the convention they parse.';
COMMENT ON COLUMN resources.updated_at IS 'Updated by trigger on every row change — use for cache invalidation.';
--
--
--db/migration/0000017_platform_actions.up.sql
--
-- 
-- ACTIONS TABLE
-- 
-- Defines actions that can be performed on resources, with risk and approval requirements.
-- scope IN ('SYSTEM', 'TENANT'):
--   SYSTEM actions are standard platform actions (CREATE, READ, APPROVE, etc.).
--   TENANT actions are custom actions defined by a specific tenant.
--
-- 
CREATE TABLE IF NOT EXISTS actions (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  -- tenant_id    UUID,        -- FK to tenants(id) added in 000062
  scope        VARCHAR(10) NOT NULL DEFAULT 'SYSTEM'
                             CHECK (scope IN ('SYSTEM', 'TENANT')),
  name         VARCHAR(100) NOT NULL,
  display_name VARCHAR(150),
  description  TEXT,
  action_type  VARCHAR(50)  NOT NULL CHECK (
    action_type IN (
      'CREATE', 'READ', 'UPDATE', 'DELETE',
      'EXECUTE', 'APPROVE', 'REJECT', 'EXPORT', 'IMPORT'
    )
  ),
  action_category VARCHAR(50) DEFAULT 'STANDARD' CHECK (
    action_category IN ('STANDARD', 'ADMINISTRATIVE', 'SENSITIVE', 'BULK', 'SYSTEM')
  ),
  risk_level VARCHAR(20) DEFAULT 'LOW' CHECK (
    risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')
  ),
  -- When requires_approval=TRUE, approver_role_id MUST be set.
  -- A flag without an approver cannot drive a workflow.
  requires_approval BOOLEAN DEFAULT false,
  approver_role_id  UUID,   -- FK to roles(id) added in 000062
  is_active    BOOLEAN     DEFAULT TRUE,
  created_at   TIMESTAMPTZ DEFAULT NOW(),
  updated_at   TIMESTAMPTZ DEFAULT NOW(),
  -- SYSTEM actions must have tenant_id=NULL; TENANT actions must have a tenant_id
  CONSTRAINT actions_scope_tenant_check CHECK (
    (scope = 'SYSTEM' AND tenant_id IS NULL)
    OR (scope = 'TENANT' AND tenant_id IS NOT NULL)
  ),
  -- When approval is required, an approver role must be designated
  CONSTRAINT actions_approval_requires_role CHECK (
    NOT requires_approval OR approver_role_id IS NOT NULL
  )
);

COMMENT ON TABLE  actions                   IS 'Defines actions that can be performed on resources with risk assessment and approval workflow requirements.';
COMMENT ON COLUMN actions.tenant_id         IS 'NULL for SYSTEM-scope actions. Set for custom TENANT-scope actions. FK enforced in 000062.';
COMMENT ON COLUMN actions.scope             IS 'SYSTEM = platform-wide standard action. TENANT = custom action for one tenant.';
COMMENT ON COLUMN actions.action_type       IS 'Standard action type: CREATE, READ, UPDATE, DELETE, EXECUTE, APPROVE, REJECT, EXPORT, IMPORT';
COMMENT ON COLUMN actions.action_category   IS 'Risk category: STANDARD, ADMINISTRATIVE, SENSITIVE, BULK, SYSTEM';
COMMENT ON COLUMN actions.risk_level        IS 'Risk level for audit and approval routing: LOW, MEDIUM, HIGH, CRITICAL';
COMMENT ON COLUMN actions.requires_approval IS 'Whether this action requires explicit approval. If TRUE, approver_role_id MUST be set.';
COMMENT ON COLUMN actions.approver_role_id  IS 'Role whose members can approve this action. Required when requires_approval=TRUE. FK enforced in 000062.';
COMMENT ON COLUMN actions.updated_at        IS 'Updated by trigger on every row change — use for cache invalidation.';

-- 
-- INDEXES
-- 
CREATE INDEX idx_actions_scope    ON actions(scope, is_active)    WHERE is_active = TRUE;
CREATE INDEX idx_actions_tenant   ON actions(tenant_id)           WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_actions_type     ON actions(action_type, scope);
CREATE INDEX idx_actions_approval ON actions(approver_role_id)    WHERE requires_approval = TRUE;
