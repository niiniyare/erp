-- =====================================================================
-- IAM: ROLES, PERMISSIONS, ROLE ASSIGNMENTS, USER ROLE ASSIGNMENTS
-- Migrations: 000404 (permissions), 000405 (roles), 000406 (role_permissions),
--             000407 (user_roles), 000411 (user_permissions)
-- Notes:
--   - roles.entity_id is NOT NULL — assign to root company entity
--   - permissions bind: tenant_id + resource_id + action_id
--   - role_permissions.entity_scope is nullable (NULL = applies everywhere)
--   - user_roles.entity_id = entity where role applies
-- =====================================================================
DO $$
DECLARE
  v_acme_id   UUID;
  v_globex_id UUID;
  v_stark_id  UUID;

  -- Entity UUIDs (from 03_entities.sql)
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001';
  v_us_ops    UUID := 'a0000000-0000-0000-0000-000000000002';
  v_sales     UUID := 'a0000000-0000-0000-0000-000000000003';
  v_eng       UUID := 'a0000000-0000-0000-0000-000000000004';
  v_eu_hq     UUID := 'a0000000-0000-0000-0000-000000000005';
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001';
  v_asia_div  UUID := 'b0000000-0000-0000-0000-000000000002';
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001';
  v_rnd       UUID := 'c0000000-0000-0000-0000-000000000002';
  v_weapons   UUID := 'c0000000-0000-0000-0000-000000000003';

  -- Role IDs per tenant (populated after role insert)
  v_acme_sysadmin_role UUID;
  v_acme_admin_role    UUID;
  v_acme_manager_role  UUID;
  v_acme_employee_role UUID;
  v_acme_finance_role  UUID;

  v_globex_sysadmin_role UUID;
  v_globex_manager_role  UUID;
  v_globex_employee_role UUID;

  v_stark_sysadmin_role UUID;
  v_stark_admin_role    UUID;
  v_stark_manager_role  UUID;
  v_stark_employee_role UUID;

  -- Action IDs (SYSTEM scope)
  v_act_create UUID;
  v_act_read   UUID;
  v_act_update UUID;
  v_act_delete UUID;
  v_act_approve UUID;
  v_act_export  UUID;

  -- Resource IDs (SYSTEM scope)
  v_res_employees   UUID;
  v_res_payroll     UUID;
  v_res_timesheets  UUID;
  v_res_invoices    UUID;
  v_res_accounts    UUID;
  v_res_fin_reports UUID;
  v_res_orders      UUID;
  v_res_customers   UUID;
  v_res_users       UUID;
  v_res_audit       UUID;

  -- User IDs
  v_john_uid    UUID;
  v_sarah_uid   UUID;
  v_michael_uid UUID;
  v_emma_uid    UUID;
  v_kenji_uid   UUID;
  v_wei_uid     UUID;
  v_tony_uid    UUID;
  v_pepper_uid  UUID;
  v_acme_sys_uid   UUID;
  v_globex_sys_uid UUID;
  v_stark_sys_uid  UUID;

BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  -- Action IDs
  SELECT id INTO v_act_create  FROM actions WHERE name = 'create'  AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_read    FROM actions WHERE name = 'read'    AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_update  FROM actions WHERE name = 'update'  AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_delete  FROM actions WHERE name = 'delete'  AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_approve FROM actions WHERE name = 'approve' AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_export  FROM actions WHERE name = 'export'  AND scope = 'SYSTEM' AND tenant_id IS NULL;

  -- Resource IDs
  SELECT id INTO v_res_employees   FROM resources WHERE name = 'employees';
  SELECT id INTO v_res_payroll     FROM resources WHERE name = 'payroll';
  SELECT id INTO v_res_timesheets  FROM resources WHERE name = 'timesheets';
  SELECT id INTO v_res_invoices    FROM resources WHERE name = 'invoices';
  SELECT id INTO v_res_accounts    FROM resources WHERE name = 'accounts';
  SELECT id INTO v_res_fin_reports FROM resources WHERE name = 'financial-reports';
  SELECT id INTO v_res_orders      FROM resources WHERE name = 'orders';
  SELECT id INTO v_res_customers   FROM resources WHERE name = 'customers';
  SELECT id INTO v_res_users       FROM resources WHERE name = 'users';
  SELECT id INTO v_res_audit       FROM resources WHERE name = 'audit-log';

  -- User IDs
  SELECT id INTO v_john_uid    FROM users WHERE email = 'john.smith@acme.com';
  SELECT id INTO v_sarah_uid   FROM users WHERE email = 'sarah.j@acme.com';
  SELECT id INTO v_michael_uid FROM users WHERE email = 'm.chen@acme.com';
  SELECT id INTO v_emma_uid    FROM users WHERE email = 'e.dubois@acme.eu';
  SELECT id INTO v_kenji_uid   FROM users WHERE email = 'kenji.t@globex.jp';
  SELECT id INTO v_wei_uid     FROM users WHERE email = 'wei.zhang@globex.cn';
  SELECT id INTO v_tony_uid    FROM users WHERE email = 'tony@stark.com';
  SELECT id INTO v_pepper_uid  FROM users WHERE email = 'pepper@stark.com';
  SELECT id INTO v_acme_sys_uid   FROM users WHERE email = 'sysadmin@acme-corp.com';
  SELECT id INTO v_globex_sys_uid FROM users WHERE email = 'sysadmin@globex.com';
  SELECT id INTO v_stark_sys_uid  FROM users WHERE email = 'sysadmin@stark.com';

  -- =====================================================================
  -- ROLES (per tenant, assigned to root company entity)
  -- role_type: SYSTEM, TENANT, ENTITY, CUSTOM, FUNCTIONAL
  -- is_system_role=TRUE for built-in roles
  -- =====================================================================

  -- ACME roles
  INSERT INTO roles (tenant_id, entity_id, name, display_name, role_type, is_system_role, is_active)
  VALUES
    (v_acme_id, v_acme_co, 'sysadmin',          'System Administrator', 'SYSTEM',     TRUE,  TRUE),
    (v_acme_id, v_acme_co, 'tenant-admin',       'Tenant Administrator', 'TENANT',     TRUE,  TRUE),
    (v_acme_id, v_acme_co, 'manager',            'Department Manager',   'ENTITY',     FALSE, TRUE),
    (v_acme_id, v_acme_co, 'employee',           'Regular Employee',     'TENANT',     FALSE, TRUE),
    (v_acme_id, v_acme_co, 'finance-specialist', 'Finance Specialist',   'FUNCTIONAL', FALSE, TRUE)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- Globex roles
  INSERT INTO roles (tenant_id, entity_id, name, display_name, role_type, is_system_role, is_active)
  VALUES
    (v_globex_id, v_globex_co, 'sysadmin',    'System Administrator', 'SYSTEM', TRUE,  TRUE),
    (v_globex_id, v_globex_co, 'manager',     'Department Manager',   'ENTITY', FALSE, TRUE),
    (v_globex_id, v_globex_co, 'employee',    'Regular Employee',     'TENANT', FALSE, TRUE)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- Stark roles
  INSERT INTO roles (tenant_id, entity_id, name, display_name, role_type, is_system_role, is_active)
  VALUES
    (v_stark_id, v_stark_co, 'sysadmin',          'System Administrator', 'SYSTEM',     TRUE,  TRUE),
    (v_stark_id, v_stark_co, 'tenant-admin',       'Tenant Administrator', 'TENANT',     TRUE,  TRUE),
    (v_stark_id, v_stark_co, 'manager',            'Department Manager',   'ENTITY',     FALSE, TRUE),
    (v_stark_id, v_stark_co, 'employee',           'Regular Employee',     'TENANT',     FALSE, TRUE)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  -- Fetch role IDs
  SELECT id INTO v_acme_sysadmin_role FROM roles WHERE tenant_id = v_acme_id AND name = 'sysadmin';
  SELECT id INTO v_acme_admin_role    FROM roles WHERE tenant_id = v_acme_id AND name = 'tenant-admin';
  SELECT id INTO v_acme_manager_role  FROM roles WHERE tenant_id = v_acme_id AND name = 'manager';
  SELECT id INTO v_acme_employee_role FROM roles WHERE tenant_id = v_acme_id AND name = 'employee';
  SELECT id INTO v_acme_finance_role  FROM roles WHERE tenant_id = v_acme_id AND name = 'finance-specialist';

  SELECT id INTO v_globex_sysadmin_role FROM roles WHERE tenant_id = v_globex_id AND name = 'sysadmin';
  SELECT id INTO v_globex_manager_role  FROM roles WHERE tenant_id = v_globex_id AND name = 'manager';
  SELECT id INTO v_globex_employee_role FROM roles WHERE tenant_id = v_globex_id AND name = 'employee';

  SELECT id INTO v_stark_sysadmin_role FROM roles WHERE tenant_id = v_stark_id AND name = 'sysadmin';
  SELECT id INTO v_stark_admin_role    FROM roles WHERE tenant_id = v_stark_id AND name = 'tenant-admin';
  SELECT id INTO v_stark_manager_role  FROM roles WHERE tenant_id = v_stark_id AND name = 'manager';
  SELECT id INTO v_stark_employee_role FROM roles WHERE tenant_id = v_stark_id AND name = 'employee';

  -- =====================================================================
  -- PERMISSIONS (tenant × resource × action)
  -- name format: resource:action
  -- effect: ALLOW or DENY
  -- =====================================================================

  -- Helper: create all standard CRUD permissions for a given tenant and resource
  -- We insert ALLOW permissions for all CRUD + approve + export actions

  -- ACME permissions
  INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
  SELECT v_acme_id, r.res_id, a.act_id, r.res_name || ':' || a.act_name, 'ALLOW', '{}'::jsonb
  FROM (VALUES
    (v_res_employees,   'employees'),
    (v_res_payroll,     'payroll'),
    (v_res_timesheets,  'timesheets'),
    (v_res_invoices,    'invoices'),
    (v_res_accounts,    'accounts'),
    (v_res_fin_reports, 'financial-reports'),
    (v_res_orders,      'orders'),
    (v_res_customers,   'customers'),
    (v_res_users,       'security-users'),
    (v_res_audit,       'audit-log')
  ) AS r(res_id, res_name)
  CROSS JOIN (VALUES
    (v_act_create, 'create'),
    (v_act_read,   'read'),
    (v_act_update, 'update'),
    (v_act_delete, 'delete'),
    (v_act_approve,'approve'),
    (v_act_export, 'export')
  ) AS a(act_id, act_name)
  ON CONFLICT (tenant_id, resource_id, action_id, name) DO NOTHING;

  -- Payroll requires MFA — override payroll:create/update/delete conditions
  UPDATE permissions
  SET conditions = '{"require_mfa": true}'::jsonb
  WHERE tenant_id = v_acme_id
    AND resource_id = v_res_payroll
    AND name LIKE 'payroll:%'
    AND conditions = '{}'::jsonb;

  -- Globex permissions
  INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
  SELECT v_globex_id, r.res_id, a.act_id, r.res_name || ':' || a.act_name, 'ALLOW', '{}'::jsonb
  FROM (VALUES
    (v_res_employees,   'employees'),
    (v_res_payroll,     'payroll'),
    (v_res_invoices,    'invoices'),
    (v_res_orders,      'orders'),
    (v_res_customers,   'customers')
  ) AS r(res_id, res_name)
  CROSS JOIN (VALUES
    (v_act_create, 'create'),
    (v_act_read,   'read'),
    (v_act_update, 'update'),
    (v_act_delete, 'delete'),
    (v_act_approve,'approve')
  ) AS a(act_id, act_name)
  ON CONFLICT (tenant_id, resource_id, action_id, name) DO NOTHING;

  -- Stark permissions
  INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
  SELECT v_stark_id, r.res_id, a.act_id, r.res_name || ':' || a.act_name, 'ALLOW', '{}'::jsonb
  FROM (VALUES
    (v_res_employees,   'employees'),
    (v_res_payroll,     'payroll'),
    (v_res_invoices,    'invoices'),
    (v_res_accounts,    'accounts'),
    (v_res_orders,      'orders'),
    (v_res_users,       'security-users'),
    (v_res_audit,       'audit-log')
  ) AS r(res_id, res_name)
  CROSS JOIN (VALUES
    (v_act_create, 'create'),
    (v_act_read,   'read'),
    (v_act_update, 'update'),
    (v_act_delete, 'delete'),
    (v_act_approve,'approve'),
    (v_act_export, 'export')
  ) AS a(act_id, act_name)
  ON CONFLICT (tenant_id, resource_id, action_id, name) DO NOTHING;

  -- =====================================================================
  -- ROLE PERMISSIONS
  -- Assign permission sets to roles (entity_scope NULL = applies to all entities)
  -- =====================================================================

  -- ACME: sysadmin + tenant-admin → all ALLOW permissions
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_sysadmin_role, p.id, TRUE
  FROM permissions p
  WHERE p.tenant_id = v_acme_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_admin_role, p.id, TRUE
  FROM permissions p
  WHERE p.tenant_id = v_acme_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ACME: manager → READ + APPROVE + EXPORT
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_manager_role, p.id, TRUE
  FROM permissions p
  JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id AND a.name IN ('read','approve','export') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ACME: employee → READ only
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_employee_role, p.id, TRUE
  FROM permissions p
  JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id AND a.name = 'read' AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ACME: finance-specialist → full access to finance resources
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_finance_role, p.id, TRUE
  FROM permissions p
  JOIN resources r ON p.resource_id = r.id
  JOIN actions a   ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id
    AND r.name IN ('invoices','accounts','financial-reports')
    AND a.name IN ('create','read','update','export')
    AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Globex: sysadmin → all
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_globex_id, v_globex_sysadmin_role, p.id, TRUE
  FROM permissions p WHERE p.tenant_id = v_globex_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Globex: manager → READ + APPROVE
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_globex_id, v_globex_manager_role, p.id, TRUE
  FROM permissions p
  JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_globex_id AND a.name IN ('read','approve') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Globex: employee → READ
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_globex_id, v_globex_employee_role, p.id, TRUE
  FROM permissions p
  JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_globex_id AND a.name = 'read' AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Stark: sysadmin + tenant-admin → all
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, v_stark_sysadmin_role, p.id, TRUE
  FROM permissions p WHERE p.tenant_id = v_stark_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, v_stark_admin_role, p.id, TRUE
  FROM permissions p WHERE p.tenant_id = v_stark_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Stark: manager → READ + APPROVE + EXPORT
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, v_stark_manager_role, p.id, TRUE
  FROM permissions p
  JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_stark_id AND a.name IN ('read','approve','export') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Stark: employee → READ
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, v_stark_employee_role, p.id, TRUE
  FROM permissions p
  JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_stark_id AND a.name = 'read' AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- =====================================================================
  -- USER ROLE ASSIGNMENTS
  -- UNIQUE constraint: (user_id, role_id, entity_id)
  -- =====================================================================

  -- ACME
  INSERT INTO user_roles (user_id, role_id, entity_id, assignment_type, is_active)
  VALUES
    -- sysadmin user → sysadmin role at company level
    (v_acme_sys_uid,   v_acme_sysadmin_role, v_acme_co, 'DIRECT', TRUE),
    -- John Smith → manager at US Ops
    (v_john_uid,       v_acme_manager_role,  v_us_ops,  'DIRECT', TRUE),
    -- John → employee at company level too
    (v_john_uid,       v_acme_employee_role, v_acme_co, 'DIRECT', TRUE),
    -- Sarah Johnson → manager at Sales
    (v_sarah_uid,      v_acme_manager_role,  v_sales,   'DIRECT', TRUE),
    (v_sarah_uid,      v_acme_employee_role, v_acme_co, 'DIRECT', TRUE),
    -- Michael Chen → employee + finance-specialist
    (v_michael_uid,    v_acme_employee_role, v_acme_co, 'DIRECT', TRUE),
    (v_michael_uid,    v_acme_finance_role,  v_acme_co, 'DIRECT', TRUE),
    -- Emma Dubois → manager at EU HQ
    (v_emma_uid,       v_acme_manager_role,  v_eu_hq,   'DIRECT', TRUE),
    (v_emma_uid,       v_acme_employee_role, v_acme_co, 'DIRECT', TRUE)
  ON CONFLICT (user_id, role_id, entity_id) DO NOTHING;

  -- Globex
  INSERT INTO user_roles (user_id, role_id, entity_id, assignment_type, is_active)
  VALUES
    (v_globex_sys_uid, v_globex_sysadmin_role, v_globex_co, 'DIRECT', TRUE),
    (v_kenji_uid,      v_globex_manager_role,  v_asia_div,  'DIRECT', TRUE),
    (v_kenji_uid,      v_globex_employee_role, v_globex_co, 'DIRECT', TRUE),
    (v_wei_uid,        v_globex_employee_role, v_globex_co, 'DIRECT', TRUE)
  ON CONFLICT (user_id, role_id, entity_id) DO NOTHING;

  -- Stark
  INSERT INTO user_roles (user_id, role_id, entity_id, assignment_type, is_active)
  VALUES
    (v_stark_sys_uid, v_stark_sysadmin_role, v_stark_co, 'DIRECT', TRUE),
    -- Tony Stark → tenant-admin (CEO override)
    (v_tony_uid,      v_stark_admin_role,    v_stark_co, 'DIRECT', TRUE),
    (v_tony_uid,      v_stark_employee_role, v_stark_co, 'DIRECT', TRUE),
    -- Pepper Potts → manager at company level
    (v_pepper_uid,    v_stark_manager_role,  v_stark_co, 'DIRECT', TRUE),
    (v_pepper_uid,    v_stark_employee_role, v_stark_co, 'DIRECT', TRUE)
  ON CONFLICT (user_id, role_id, entity_id) DO NOTHING;

  -- =====================================================================
  -- DIRECT USER PERMISSIONS (exceptional grants outside roles)
  -- =====================================================================

  -- Tony Stark: direct ALLOW on users:delete (CEO-level override)
  INSERT INTO user_permissions (tenant_id, user_id, permission_id, entity_id, effect, reason, is_active)
  SELECT v_stark_id, v_tony_uid, p.id, v_stark_co, 'ALLOW',
         'CEO override — full platform access', TRUE
  FROM permissions p
  JOIN resources r ON p.resource_id = r.id
  JOIN actions a   ON p.action_id = a.id
  WHERE p.tenant_id = v_stark_id
    AND r.name = 'security-users'
    AND a.name = 'delete'
  ON CONFLICT (tenant_id, user_id, permission_id, entity_id) DO NOTHING;

  -- Michael Chen: direct DENY on payroll:delete (safety restriction)
  INSERT INTO user_permissions (tenant_id, user_id, permission_id, entity_id, effect, reason, is_active)
  SELECT v_acme_id, v_michael_uid, p.id, v_acme_co, 'DENY',
         'Finance safety restriction — engineers cannot delete payroll records', TRUE
  FROM permissions p
  JOIN resources r ON p.resource_id = r.id
  JOIN actions a   ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id
    AND r.name = 'payroll'
    AND a.name = 'delete'
  ON CONFLICT (tenant_id, user_id, permission_id, entity_id) DO NOTHING;

END $$;
