-- ──────────────────────────────────────────────────────────────────────────────
-- 000018_platform_registry.up.sql
--
-- 1. Creates the `permission_catalog` table — canonical dot-notation keys derived
--    from the MRA (Module / Resource / Action) registry.
--    NOTE: named `permission_catalog` (not `permissions`) to avoid collision with
--    the tenant-scoped ABAC `permissions` table in migration 000404.
-- 2. Adds a trigger that auto-inserts a row whenever an action is inserted,
--    so the catalogue stays in sync with no manual bookkeeping.
-- 3. Seeds SYSTEM-scope modules, resources, and actions for the four core
--    modules: finance, people, settings, iam.
--
-- Permission key format: {module.slug}.{resource.slug}.{action.slug}
-- e.g.  finance.transactions.read
--       iam.users.create
-- ──────────────────────────────────────────────────────────────────────────────

-- ─── Permission catalogue ─────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS permission_catalog (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id     UUID        NOT NULL REFERENCES modules(id)   ON DELETE CASCADE,
    resource_id   UUID        NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    action_id     UUID        NOT NULL REFERENCES actions(id)   ON DELETE CASCADE,
    -- Slugs cached here so full_key can be a GENERATED column without joins
    module_slug   VARCHAR(50) NOT NULL,
    resource_slug VARCHAR(50) NOT NULL,
    action_slug   VARCHAR(50) NOT NULL,
    -- Dot-notation permission key — derived, never edited directly
    full_key      TEXT        NOT NULL GENERATED ALWAYS AS (
                      module_slug || '.' || resource_slug || '.' || action_slug
                  ) STORED,
    description   TEXT,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT permission_catalog_action_unique   UNIQUE (action_id),
    CONSTRAINT permission_catalog_full_key_unique UNIQUE (full_key)
);

CREATE INDEX IF NOT EXISTS idx_permission_catalog_full_key  ON permission_catalog (full_key)    WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_permission_catalog_module    ON permission_catalog (module_id)   WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_permission_catalog_resource  ON permission_catalog (resource_id) WHERE is_active = TRUE;

COMMENT ON TABLE  permission_catalog          IS 'Canonical permission catalogue derived from the MRA action registry. full_key is the dot-notation string used in Casbin policies and ResolvedSession.Permissions maps.';
COMMENT ON COLUMN permission_catalog.full_key IS 'GENERATED: {module_slug}.{resource_slug}.{action_slug}. e.g. finance.transactions.read. Never change slugs after seeding — all policies reference these keys.';

-- ─── Auto-create permission on action insert ──────────────────────────────────

CREATE OR REPLACE FUNCTION fn_auto_create_permission()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    v_module_id     UUID;
    v_module_slug   VARCHAR(50);
    v_resource_slug VARCHAR(50);
BEGIN
    SELECT m.id, m.slug, r.slug
    INTO   v_module_id, v_module_slug, v_resource_slug
    FROM   resources r
    JOIN   modules   m ON m.id = r.module_id
    WHERE  r.id = NEW.resource_id;

    IF v_module_id IS NOT NULL THEN
        INSERT INTO permission_catalog (
            module_id, resource_id, action_id,
            module_slug, resource_slug, action_slug,
            description
        ) VALUES (
            v_module_id, NEW.resource_id, NEW.id,
            v_module_slug, v_resource_slug, NEW.slug,
            'Permission for ' || v_module_slug || '.' || v_resource_slug || '.' || NEW.slug
        )
        ON CONFLICT (action_id) DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'trg_auto_create_permission'
    ) THEN
        CREATE TRIGGER trg_auto_create_permission
            AFTER INSERT ON actions
            FOR EACH ROW
            EXECUTE FUNCTION fn_auto_create_permission();
    END IF;
END $$;

-- ─── Seed: FINANCE module ─────────────────────────────────────────────────────

INSERT INTO modules (slug, name, display_name, description, icon, nav_order, category, module_type, scope)
VALUES ('finance', 'Finance', 'Finance Management',
        'Core financial management: accounts, transactions, receivables, payables, reporting.',
        'fa fa-calculator', 10, 'FINANCE', 'CORE', 'SYSTEM')
ON CONFLICT DO NOTHING;

WITH mod AS (SELECT id FROM modules WHERE slug = 'finance' AND scope = 'SYSTEM')
INSERT INTO resources (module_id, slug, name, display_name, resource_type, nav_url, nav_order)
SELECT mod.id, r.slug, r.name, r.display_name, r.resource_type::VARCHAR, r.nav_url, r.nav_order
FROM   mod,
       (VALUES
           ('accounts',     'Accounts',     'Chart of Accounts',  'UI',     '/finance/accounts',     10),
           ('transactions', 'Transactions', 'Transactions',        'UI',     '/finance/transactions', 20),
           ('receivables',  'Receivables',  'Accounts Receivable', 'UI',     '/finance/receivables',  30),
           ('payables',     'Payables',     'Accounts Payable',    'UI',     '/finance/payables',     40),
           ('reports',      'Reports',      'Financial Reports',   'REPORT', '/finance/reports',      50)
       ) AS r(slug, name, display_name, resource_type, nav_url, nav_order)
ON CONFLICT (module_id, slug) DO NOTHING;

-- Standard CRUD on all finance resources
WITH res AS (
    SELECT r.id
    FROM   resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'finance' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level)
SELECT res.id, a.slug, a.name, a.action_type, a.http_method, 'SYSTEM', a.risk_level
FROM   res,
       (VALUES
           ('read',   'Read',   'READ',   'GET',    'LOW'),
           ('create', 'Create', 'CREATE', 'POST',   'LOW'),
           ('update', 'Update', 'UPDATE', 'PUT',    'MEDIUM'),
           ('delete', 'Delete', 'DELETE', 'DELETE', 'HIGH')
       ) AS a(slug, name, action_type, http_method, risk_level)
ON CONFLICT (resource_id, slug) DO NOTHING;

-- Finance-specific actions on transactions
WITH tx AS (
    SELECT r.id FROM resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'finance' AND r.slug = 'transactions' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level, action_category)
SELECT tx.id, a.slug, a.name, a.action_type, a.http_method, 'SYSTEM', a.risk_level, a.category
FROM   tx,
       (VALUES
           ('approve', 'Approve', 'APPROVE', 'POST', 'HIGH',     'SENSITIVE'),
           ('post',    'Post',    'EXECUTE', 'POST', 'HIGH',     'SENSITIVE'),
           ('void',    'Void',    'EXECUTE', 'POST', 'CRITICAL', 'SENSITIVE'),
           ('export',  'Export',  'EXPORT',  'GET',  'LOW',      'STANDARD')
       ) AS a(slug, name, action_type, http_method, risk_level, category)
ON CONFLICT (resource_id, slug) DO NOTHING;

-- Export on reports
WITH rpt AS (
    SELECT r.id FROM resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'finance' AND r.slug = 'reports' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level)
SELECT rpt.id, 'export', 'Export', 'EXPORT', 'GET', 'SYSTEM', 'LOW'
FROM   rpt
ON CONFLICT (resource_id, slug) DO NOTHING;

-- ─── Seed: PEOPLE module ──────────────────────────────────────────────────────

INSERT INTO modules (slug, name, display_name, description, icon, nav_order, category, module_type, scope)
VALUES ('people', 'People', 'People & HR', 'Employee and people management.',
        'fa fa-users', 20, 'HR', 'CORE', 'SYSTEM')
ON CONFLICT DO NOTHING;

WITH mod AS (SELECT id FROM modules WHERE slug = 'people' AND scope = 'SYSTEM')
INSERT INTO resources (module_id, slug, name, display_name, resource_type, nav_url, nav_order)
SELECT mod.id, r.slug, r.name, r.display_name, 'UI', r.nav_url, r.nav_order
FROM   mod,
       (VALUES
           ('employees', 'Employees', 'Employees', '/people/employees', 10),
           ('persons',   'Persons',   'Persons',   '/people/persons',   20)
       ) AS r(slug, name, display_name, nav_url, nav_order)
ON CONFLICT (module_id, slug) DO NOTHING;

WITH res AS (
    SELECT r.id FROM resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'people' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level)
SELECT res.id, a.slug, a.name, a.action_type, a.http_method, 'SYSTEM', a.risk_level
FROM   res,
       (VALUES
           ('read',   'Read',   'READ',   'GET',    'LOW'),
           ('create', 'Create', 'CREATE', 'POST',   'LOW'),
           ('update', 'Update', 'UPDATE', 'PUT',    'MEDIUM'),
           ('delete', 'Delete', 'DELETE', 'DELETE', 'HIGH')
       ) AS a(slug, name, action_type, http_method, risk_level)
ON CONFLICT (resource_id, slug) DO NOTHING;

-- ─── Seed: SETTINGS module ────────────────────────────────────────────────────

INSERT INTO modules (slug, name, display_name, description, icon, nav_order, category, module_type, scope)
VALUES ('settings', 'Settings', 'Settings', 'Tenant and system configuration.',
        'fa fa-cog', 90, 'CORE', 'CORE', 'SYSTEM')
ON CONFLICT DO NOTHING;

WITH mod AS (SELECT id FROM modules WHERE slug = 'settings' AND scope = 'SYSTEM')
INSERT INTO resources (module_id, slug, name, display_name, resource_type, nav_url, nav_order)
SELECT mod.id, r.slug, r.name, r.display_name, 'UI', r.nav_url, r.nav_order
FROM   mod,
       (VALUES
           ('iam',     'IAM',     'Access Control',  '/settings/iam',     10),
           ('general', 'General', 'General',          '/settings/general', 20),
           ('modules', 'Modules', 'Module Registry',  '/settings/modules', 30)
       ) AS r(slug, name, display_name, nav_url, nav_order)
ON CONFLICT (module_id, slug) DO NOTHING;

WITH res AS (
    SELECT r.id FROM resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'settings' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level)
SELECT res.id, a.slug, a.name, a.action_type, a.http_method, 'SYSTEM', a.risk_level
FROM   res,
       (VALUES
           ('read',   'Read',   'READ',   'GET', 'LOW'),
           ('update', 'Update', 'UPDATE', 'PUT', 'HIGH')
       ) AS a(slug, name, action_type, http_method, risk_level)
ON CONFLICT (resource_id, slug) DO NOTHING;

-- ─── Seed: IAM module ─────────────────────────────────────────────────────────

INSERT INTO modules (slug, name, display_name, description, icon, nav_order, category, module_type, scope)
VALUES ('iam', 'IAM', 'Identity & Access', 'Users, roles, and permission management.',
        'fa fa-shield', 80, 'CORE', 'CORE', 'SYSTEM')
ON CONFLICT DO NOTHING;

WITH mod AS (SELECT id FROM modules WHERE slug = 'iam' AND scope = 'SYSTEM')
INSERT INTO resources (module_id, slug, name, display_name, resource_type, nav_url, nav_order)
SELECT mod.id, r.slug, r.name, r.display_name, 'UI', r.nav_url, r.nav_order
FROM   mod,
       (VALUES
           ('users',    'Users',    'Users',     '/iam/users',    10),
           ('roles',    'Roles',    'Roles',     '/iam/roles',    20),
           ('policies', 'Policies', 'Policies',  '/iam/policies', 30),
           ('api-keys', 'API Keys', 'API Keys',  '/iam/api-keys', 40),
           ('sessions', 'Sessions', 'Sessions',  '/iam/sessions', 50)
       ) AS r(slug, name, display_name, nav_url, nav_order)
ON CONFLICT (module_id, slug) DO NOTHING;

WITH res AS (
    SELECT r.id FROM resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'iam' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level)
SELECT res.id, a.slug, a.name, a.action_type, a.http_method, 'SYSTEM', a.risk_level
FROM   res,
       (VALUES
           ('read',   'Read',   'READ',   'GET',    'LOW'),
           ('create', 'Create', 'CREATE', 'POST',   'MEDIUM'),
           ('update', 'Update', 'UPDATE', 'PUT',    'HIGH'),
           ('delete', 'Delete', 'DELETE', 'DELETE', 'HIGH')
       ) AS a(slug, name, action_type, http_method, risk_level)
ON CONFLICT (resource_id, slug) DO NOTHING;

-- Session revoke is IAM-specific
WITH sess_res AS (
    SELECT r.id FROM resources r JOIN modules m ON m.id = r.module_id
    WHERE  m.slug = 'iam' AND r.slug = 'sessions' AND m.scope = 'SYSTEM'
)
INSERT INTO actions (resource_id, slug, name, action_type, http_method, scope, risk_level, action_category)
SELECT sess_res.id, 'revoke', 'Revoke', 'EXECUTE', 'POST', 'SYSTEM', 'HIGH', 'SENSITIVE'
FROM   sess_res
ON CONFLICT (resource_id, slug) DO NOTHING;
