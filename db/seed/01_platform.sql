-- =====================================================================
-- PLATFORM: SYSTEM MODULES, ACTIONS, AND RESOURCES
-- Migrations: 000015 (modules), 000016 (resources), 000017 (actions)
-- These are SYSTEM-scoped and have no tenant_id.
-- =====================================================================

-- -------------------------------------------------------------------------
-- SYSTEM MODULES
-- scope='SYSTEM' requires tenant_id IS NULL (constraint: modules_scope_tenant_check)
-- -------------------------------------------------------------------------
INSERT INTO modules (scope, name, display_name, description, category, module_type, version, is_active)
SELECT 'SYSTEM', t.name, t.display_name, t.description, t.category, t.module_type, '1.0', TRUE
FROM (VALUES
  ('hr',        'Human Resources',       'Employee management, payroll, and workforce planning',   'HR',        'CORE'),
  ('finance',   'Financial Management',  'Accounting, invoicing, budgeting, and financial reports','FINANCE',   'CORE'),
  ('inventory', 'Inventory Control',     'Stock management, warehousing, and supply chain ops',   'OPERATIONS','CORE'),
  ('sales',     'Sales & CRM',           'Customer management, sales orders, and pipeline',       'SALES',     'CORE'),
  ('security',  'Security & Compliance', 'IAM, audit logging, and compliance management',         'CORE',      'INTERNAL')
) AS t(name, display_name, description, category, module_type)
WHERE NOT EXISTS (
  SELECT 1 FROM modules WHERE name = t.name AND scope = 'SYSTEM' AND tenant_id IS NULL
);

-- -------------------------------------------------------------------------
-- SYSTEM ACTIONS
-- scope='SYSTEM' requires tenant_id IS NULL (constraint: actions_scope_tenant_check)
-- requires_approval=FALSE so approver_role_id can be NULL (constraint: actions_approval_requires_role)
-- -------------------------------------------------------------------------
INSERT INTO actions (scope, name, display_name, action_type, action_category, risk_level, requires_approval)
SELECT 'SYSTEM', t.name, t.display_name, t.action_type, t.action_category, t.risk_level, FALSE
FROM (VALUES
  ('create', 'Create Record',  'CREATE', 'STANDARD',       'MEDIUM'),
  ('read',   'Read Record',    'READ',   'STANDARD',       'LOW'),
  ('update', 'Update Record',  'UPDATE', 'STANDARD',       'HIGH'),
  ('delete', 'Delete Record',  'DELETE', 'SENSITIVE',      'CRITICAL'),
  ('approve','Approve Action', 'APPROVE','ADMINISTRATIVE', 'HIGH'),
  ('export', 'Export Data',    'EXPORT', 'SENSITIVE',      'MEDIUM'),
  ('import', 'Import Data',    'IMPORT', 'BULK',           'HIGH')
) AS t(name, display_name, action_type, action_category, risk_level)
WHERE NOT EXISTS (
  SELECT 1 FROM actions WHERE name = t.name AND scope = 'SYSTEM' AND tenant_id IS NULL
);

-- -------------------------------------------------------------------------
-- RESOURCES (linked to SYSTEM modules; no tenant_id on resources table)
-- -------------------------------------------------------------------------
DO $$
DECLARE
  v_mod_hr        UUID;
  v_mod_finance   UUID;
  v_mod_inventory UUID;
  v_mod_sales     UUID;
  v_mod_security  UUID;
BEGIN
  SELECT id INTO v_mod_hr        FROM modules WHERE name = 'hr'        AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_finance   FROM modules WHERE name = 'finance'   AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_inventory FROM modules WHERE name = 'inventory' AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_sales     FROM modules WHERE name = 'sales'     AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_security  FROM modules WHERE name = 'security'  AND scope = 'SYSTEM' AND tenant_id IS NULL;

  -- HR resources
  INSERT INTO resources (module_id, name, display_name, resource_type, path)
  VALUES
    (v_mod_hr, 'employees',  'Employee Records',   'DATA',   'hr.employees'),
    (v_mod_hr, 'payroll',    'Payroll System',     'API',    'hr.payroll'),
    (v_mod_hr, 'timesheets', 'Timesheet Records',  'DATA',   'hr.timesheets'),
    (v_mod_hr, 'leave',      'Leave Management',   'WORKFLOW','hr.leave')
  ON CONFLICT DO NOTHING;

  -- Finance resources
  INSERT INTO resources (module_id, name, display_name, resource_type, path)
  VALUES
    (v_mod_finance, 'invoices',          'Invoices',              'API',    'finance.invoices'),
    (v_mod_finance, 'accounts',          'Chart of Accounts',     'DATA',   'finance.accounts'),
    (v_mod_finance, 'transactions',      'Financial Transactions','DATA',   'finance.transactions'),
    (v_mod_finance, 'financial-reports', 'Financial Reports',     'REPORT', 'finance.reports')
  ON CONFLICT DO NOTHING;

  -- Inventory resources
  INSERT INTO resources (module_id, name, display_name, resource_type, path)
  VALUES
    (v_mod_inventory, 'products',  'Product Catalogue', 'DATA', 'inventory.products'),
    (v_mod_inventory, 'stock',     'Stock Levels',      'API',  'inventory.stock'),
    (v_mod_inventory, 'warehouse', 'Warehouse Data',    'DATA', 'inventory.warehouse')
  ON CONFLICT DO NOTHING;

  -- Sales resources
  INSERT INTO resources (module_id, name, display_name, resource_type, path)
  VALUES
    (v_mod_sales, 'orders',    'Sales Orders',     'API',  'sales.orders'),
    (v_mod_sales, 'customers', 'Customer Records', 'DATA', 'sales.customers'),
    (v_mod_sales, 'pipeline',  'Sales Pipeline',   'UI',   'sales.pipeline')
  ON CONFLICT DO NOTHING;

  -- Security resources
  INSERT INTO resources (module_id, name, display_name, resource_type, path)
  VALUES
    (v_mod_security, 'users',     'User Management', 'DATA',   'security.users'),
    (v_mod_security, 'roles',     'Role Management', 'DATA',   'security.roles'),
    (v_mod_security, 'audit-log', 'Audit Log',       'REPORT', 'security.audit')
  ON CONFLICT DO NOTHING;

END $$;
