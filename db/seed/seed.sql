-- =====================================================================
-- MASTER SEED FILE — MULTI-TENANT ERP
-- Generated: 2026-03-20
-- =====================================================================
-- Execution order follows migration dependency graph.
-- Safe to re-run: all inserts use ON CONFLICT DO NOTHING.
--
-- Seeded data overview:
--   Platform   : 5 SYSTEM modules, 7 SYSTEM actions, 15 global resources
--   Tenants    : ACME Corporation, Globex Corporation, Stark Industries
--   Entities   : 11 total (5 ACME, 3 Globex, 3 Stark) with closure table
--   Persons    : 9 (7 employees, 1 customer, 1 per tenant vendor)
--   Employees  : 8 linked employee records
--   Users      : 11 (8 internal, 1 customer, 3 sysadmin accounts)
--   Roles      : 12 (4-5 per tenant)
--   Permissions: ~250 (tenant × resource × action combinations)
--   Sessions   : 9 (7 active, 2 expired historical)
--   Audit logs : 15 events (auth, access, admin, compliance)
--   Feature flags: 16 (6 ACME, 4 Globex, 6 Stark)
--
-- Passwords: all seeded users have password 'Password123!'
-- =====================================================================

-- =====================================================================
-- STEP 1: EXTENSIONS (idempotent; migrations already ran these)
-- =====================================================================
-- CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =====================================================================
-- STEP 2: SYSTEM MODULES AND ACTIONS (migrations 000015–000017)
-- SYSTEM scope → tenant_id must be NULL
-- =====================================================================

INSERT INTO modules (scope, slug, name, display_name, description, category, module_type, version, is_active)
SELECT 'SYSTEM', t.slug, t.slug, t.display_name, t.description, t.category, t.module_type, '1.0', TRUE
FROM (VALUES
  ('hr',        'Human Resources',       'Employee management, payroll, and workforce planning',   'HR',        'CORE'),
  ('finance',   'Financial Management',  'Accounting, invoicing, budgeting, and financial reports','FINANCE',   'CORE'),
  ('inventory', 'Inventory Control',     'Stock management, warehousing, and supply chain ops',   'OPERATIONS','CORE'),
  ('sales',     'Sales & CRM',           'Customer management, sales orders, and pipeline',       'SALES',     'CORE'),
  ('security',  'Security & Compliance', 'IAM, audit logging, and compliance management',         'CORE',      'INTERNAL')
) AS t(slug, display_name, description, category, module_type)
WHERE NOT EXISTS (
  SELECT 1 FROM modules WHERE slug = t.slug AND scope = 'SYSTEM' AND tenant_id IS NULL
);

INSERT INTO actions (scope, slug, name, display_name, action_type, action_category, risk_level, requires_approval)
SELECT 'SYSTEM', t.slug, t.slug, t.display_name, t.action_type, t.action_category, t.risk_level, FALSE
FROM (VALUES
  ('create', 'Create Record',  'CREATE', 'STANDARD',       'MEDIUM'),
  ('read',   'Read Record',    'READ',   'STANDARD',       'LOW'),
  ('update', 'Update Record',  'UPDATE', 'STANDARD',       'HIGH'),
  ('delete', 'Delete Record',  'DELETE', 'SENSITIVE',      'CRITICAL'),
  ('approve','Approve Action', 'APPROVE','ADMINISTRATIVE', 'HIGH'),
  ('export', 'Export Data',    'EXPORT', 'SENSITIVE',      'MEDIUM'),
  ('import', 'Import Data',    'IMPORT', 'BULK',           'HIGH')
) AS t(slug, display_name, action_type, action_category, risk_level)
WHERE NOT EXISTS (
  SELECT 1 FROM actions WHERE slug = t.slug AND scope = 'SYSTEM' AND tenant_id IS NULL
);

DO $$
DECLARE
  v_mod_hr        UUID;
  v_mod_finance   UUID;
  v_mod_inventory UUID;
  v_mod_sales     UUID;
  v_mod_security  UUID;
BEGIN
  SELECT id INTO v_mod_hr        FROM modules WHERE slug = 'hr'        AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_finance   FROM modules WHERE slug = 'finance'   AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_inventory FROM modules WHERE slug = 'inventory' AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_sales     FROM modules WHERE slug = 'sales'     AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_mod_security  FROM modules WHERE slug = 'security'  AND scope = 'SYSTEM' AND tenant_id IS NULL;

  INSERT INTO resources (module_id, slug, name, display_name, resource_type, path)
  VALUES
    (v_mod_hr, 'employees',       'employees',       'Employee Records',       'DATA',    'hr.employees'),
    (v_mod_hr, 'payroll',         'payroll',         'Payroll System',         'API',     'hr.payroll'),
    (v_mod_hr, 'timesheets',      'timesheets',      'Timesheet Records',      'DATA',    'hr.timesheets'),
    (v_mod_hr, 'leave',           'leave',           'Leave Management',       'WORKFLOW','hr.leave'),
    (v_mod_finance, 'invoices',          'invoices',          'Invoices',              'API',    'finance.invoices'),
    (v_mod_finance, 'accounts',          'accounts',          'Chart of Accounts',     'DATA',   'finance.accounts'),
    (v_mod_finance, 'transactions',      'transactions',      'Financial Transactions','DATA',   'finance.transactions'),
    (v_mod_finance, 'financial-reports', 'financial-reports', 'Financial Reports',     'REPORT', 'finance.reports'),
    (v_mod_inventory, 'products',  'products',  'Product Catalogue', 'DATA', 'inventory.products'),
    (v_mod_inventory, 'stock',     'stock',     'Stock Levels',      'API',  'inventory.stock'),
    (v_mod_inventory, 'warehouse', 'warehouse', 'Warehouse Data',    'DATA', 'inventory.warehouse'),
    (v_mod_sales, 'orders',    'orders',    'Sales Orders',     'API',  'sales.orders'),
    (v_mod_sales, 'customers', 'customers', 'Customer Records', 'DATA', 'sales.customers'),
    (v_mod_sales, 'pipeline',  'pipeline',  'Sales Pipeline',   'UI',   'sales.pipeline'),
    (v_mod_security, 'users',     'users',     'User Management', 'DATA',   'security.users'),
    (v_mod_security, 'roles',     'roles',     'Role Management', 'DATA',   'security.roles'),
    (v_mod_security, 'audit-log', 'audit-log', 'Audit Log',       'REPORT', 'security.audit')
  ON CONFLICT DO NOTHING;
END $$;

-- =====================================================================
-- STEP 3: TENANTS (migration 000053)
-- Notes:
--   "Status" column is quoted (capital S) — required by schema
--   company_size: STARTUP/SMALL/MEDIUM/LARGE/ENTERPRISE (uppercase)
--   Inserting a tenant auto-creates tenant_configurations (trigger 000102)
-- =====================================================================

INSERT INTO tenants (
  slug, name, email, billing_email, billing_contact_name, subdomain,
  "Status", plan_tier, timezone, currency_code, industry, company_size,
  tax_id, registration_number, legal_entity_type, metadata, settings
) VALUES
  ('acme-corp', 'ACME Corporation', 'admin@acme-corp.com', 'billing@acme-corp.com',
   'Finance Department', 'acme', 'ACTIVE', 'ENTERPRISE', 'America/New_York', 'USD',
   'Technology', 'LARGE', 'TAX-ACME-12345', 'REG-ACME-001', 'LLC',
   '{"crm_id":"CRM-ACME-001","stripe_customer_id":"cus_acme_001"}',
   '{"theme":"light","notifications":true}'),

  ('globex', 'Globex Corporation', 'info@globex.com', 'ap@globex.com',
   'Accounts Payable', 'globex', 'ACTIVE', 'ENTERPRISE', 'Asia/Tokyo', 'JPY',
   'Manufacturing', 'ENTERPRISE', 'TAX-GLOBEX-67890', 'REG-GLOBEX-002', 'K.K.',
   '{"crm_id":"CRM-GLOBEX-002"}',
   '{"theme":"light","notifications":true}'),

  ('stark-ind', 'Stark Industries', 'contact@stark.com', 'billing@stark.com',
   'Pepper Potts', 'stark', 'ACTIVE', 'ENTERPRISE', 'America/Los_Angeles', 'USD',
   'Defense', 'ENTERPRISE', 'TAX-STARK-45678', 'REG-STARK-003', 'Inc.',
   '{"crm_id":"CRM-STARK-003","cleared_vendor":true}',
   '{"theme":"dark","notifications":true,"security_level":"high"}')
ON CONFLICT (slug) DO NOTHING;

-- =====================================================================
-- STEP 4: ENTITIES, HIERARCHY PATHS, ENTITY STATE (migrations 000201–000203)
-- entities.uuid has NO DEFAULT — must be supplied
-- entitystate.uuid has NO DEFAULT — must be supplied
-- =====================================================================

DO $$
DECLARE
  v_acme_id   UUID;
  v_globex_id UUID;
  v_stark_id  UUID;
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001';
  v_us_ops    UUID := 'a0000000-0000-0000-0000-000000000002';
  v_sales     UUID := 'a0000000-0000-0000-0000-000000000003';
  v_eng       UUID := 'a0000000-0000-0000-0000-000000000004';
  v_eu_hq     UUID := 'a0000000-0000-0000-0000-000000000005';
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001';
  v_asia_div  UUID := 'b0000000-0000-0000-0000-000000000002';
  v_mfg       UUID := 'b0000000-0000-0000-0000-000000000003';
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001';
  v_rnd       UUID := 'c0000000-0000-0000-0000-000000000002';
  v_weapons   UUID := 'c0000000-0000-0000-0000-000000000003';
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  -- ACME entities
  INSERT INTO entities (uuid, tenant_id, name, code, type, parent_id, accrual_method, fy_start_month) VALUES
    (v_acme_co, v_acme_id, 'ACME Corporation', 'ACME',   'COMPANY',    NULL,      TRUE, 1),
    (v_us_ops,  v_acme_id, 'US Operations',    'US-OPS', 'REGION',     v_acme_co, TRUE, 1),
    (v_sales,   v_acme_id, 'Sales Department', 'SALES',  'DEPARTMENT', v_us_ops,  TRUE, 1),
    (v_eng,     v_acme_id, 'Engineering',      'ENG',    'DEPARTMENT', v_us_ops,  TRUE, 1),
    (v_eu_hq,   v_acme_id, 'Europe HQ',        'EU-HQ',  'REGION',     v_acme_co, FALSE,1)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth) VALUES
    (v_acme_id, v_acme_co, v_acme_co, 0), (v_acme_id, v_us_ops, v_us_ops, 0),
    (v_acme_id, v_sales,   v_sales,   0), (v_acme_id, v_eng,    v_eng,    0),
    (v_acme_id, v_eu_hq,   v_eu_hq,   0),
    (v_acme_id, v_acme_co, v_us_ops,  1), (v_acme_id, v_acme_co, v_eu_hq,  1),
    (v_acme_id, v_us_ops,  v_sales,   1), (v_acme_id, v_us_ops,  v_eng,    1),
    (v_acme_id, v_acme_co, v_sales,   2), (v_acme_id, v_acme_co, v_eng,    2)
  ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

  INSERT INTO entitystate (uuid, tenant_id, fiscal_year, key, sequence, entity_id, config) VALUES
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_acme_co,
     '{"prefix":"ACME-INV-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'PO',  1000, v_acme_co,
     '{"prefix":"ACME-PO-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_acme_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'SO',  1000, v_acme_co,
     '{"prefix":"ACME-SO-","pad_length":6,"reset_frequency":"yearly"}')
  ON CONFLICT DO NOTHING;

  -- Globex entities
  INSERT INTO entities (uuid, tenant_id, name, code, type, parent_id, accrual_method, fy_start_month) VALUES
    (v_globex_co, v_globex_id, 'Globex Corp',   'GLOBEX', 'COMPANY',    NULL,       FALSE, 4),
    (v_asia_div,  v_globex_id, 'Asia Division', 'ASIA',   'REGION',     v_globex_co,FALSE, 4),
    (v_mfg,       v_globex_id, 'Manufacturing', 'MFG',    'DEPARTMENT', v_asia_div,  FALSE, 4)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth) VALUES
    (v_globex_id, v_globex_co, v_globex_co, 0), (v_globex_id, v_asia_div, v_asia_div, 0),
    (v_globex_id, v_mfg,       v_mfg,       0),
    (v_globex_id, v_globex_co, v_asia_div,  1), (v_globex_id, v_asia_div, v_mfg, 1),
    (v_globex_id, v_globex_co, v_mfg,       2)
  ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

  INSERT INTO entitystate (uuid, tenant_id, fiscal_year, key, sequence, entity_id, config) VALUES
    (gen_random_uuid(), v_globex_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_globex_co,
     '{"prefix":"GX-INV-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_globex_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'PO',  1000, v_globex_co,
     '{"prefix":"GX-PO-","pad_length":6,"reset_frequency":"yearly"}')
  ON CONFLICT DO NOTHING;

  -- Stark entities
  INSERT INTO entities (uuid, tenant_id, name, code, type, parent_id, accrual_method, fy_start_month) VALUES
    (v_stark_co, v_stark_id, 'Stark Industries', 'STARK',   'COMPANY',     NULL,      TRUE, 10),
    (v_rnd,      v_stark_id, 'R&D Division',     'RND',     'DEPARTMENT',  v_stark_co,TRUE, 10),
    (v_weapons,  v_stark_id, 'Advanced Weapons', 'WEAPONS', 'COST_CENTER', v_rnd,     TRUE, 10)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth) VALUES
    (v_stark_id, v_stark_co, v_stark_co, 0), (v_stark_id, v_rnd,     v_rnd,     0),
    (v_stark_id, v_weapons,  v_weapons,  0),
    (v_stark_id, v_stark_co, v_rnd,      1), (v_stark_id, v_rnd,     v_weapons, 1),
    (v_stark_id, v_stark_co, v_weapons,  2)
  ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

  INSERT INTO entitystate (uuid, tenant_id, fiscal_year, key, sequence, entity_id, config) VALUES
    (gen_random_uuid(), v_stark_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'INV', 1000, v_stark_co,
     '{"prefix":"SI-INV-","pad_length":6,"reset_frequency":"yearly"}'),
    (gen_random_uuid(), v_stark_id, EXTRACT(YEAR FROM NOW())::SMALLINT, 'PO',  1000, v_stark_co,
     '{"prefix":"SI-PO-","pad_length":6,"reset_frequency":"yearly"}')
  ON CONFLICT DO NOTHING;
END $$;

-- =====================================================================
-- STEP 5: PERSONS, EMPLOYEES, USERS (migrations 000301–000303)
-- persons.entity_id NOT NULL; employees.hire_date NOT NULL
-- users.entity_id NOT NULL; lockout_until must be NULL or future
-- =====================================================================

DO $$
DECLARE
  v_acme_id   UUID;  v_globex_id UUID;  v_stark_id  UUID;
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001';
  v_us_ops    UUID := 'a0000000-0000-0000-0000-000000000002';
  v_sales     UUID := 'a0000000-0000-0000-0000-000000000003';
  v_eng       UUID := 'a0000000-0000-0000-0000-000000000004';
  v_eu_hq     UUID := 'a0000000-0000-0000-0000-000000000005';
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001';
  v_asia_div  UUID := 'b0000000-0000-0000-0000-000000000002';
  v_mfg       UUID := 'b0000000-0000-0000-0000-000000000003';
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001';
  v_rnd       UUID := 'c0000000-0000-0000-0000-000000000002';
  v_weapons   UUID := 'c0000000-0000-0000-0000-000000000003';
  -- Person IDs
  v_john_id UUID;  v_sarah_id UUID;  v_michael_id UUID;  v_emma_id UUID;  v_robert_id UUID;
  v_kenji_id UUID; v_wei_id UUID;
  v_tony_id UUID;  v_pepper_id UUID;
  -- Employee IDs
  v_john_emp UUID;  v_sarah_emp UUID;  v_michael_emp UUID;  v_emma_emp UUID;
  v_kenji_emp UUID; v_wei_emp UUID;
  v_tony_emp UUID;  v_pepper_emp UUID;
  v_pw TEXT;
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';
  v_pw := crypt('Password123!', gen_salt('bf'));

  -- Persons
  INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes) VALUES
    (v_acme_id, v_us_ops,  'EMPLOYEE', 'John',   'Smith',   'john.smith@acme.com', '+1-555-1234', '{"clearance":"high","department":"sales"}'),
    (v_acme_id, v_sales,   'EMPLOYEE', 'Sarah',  'Johnson', 'sarah.j@acme.com',    '+1-555-5678', '{"clearance":"medium","region":"us"}'),
    (v_acme_id, v_eng,     'EMPLOYEE', 'Michael','Chen',    'm.chen@acme.com',     '+1-555-9012', '{"clearance":"high","specialization":"ai"}'),
    (v_acme_id, v_eu_hq,   'EMPLOYEE', 'Emma',   'Dubois',  'e.dubois@acme.eu',    '+33-1-2345',  '{"clearance":"medium","language":"fr"}'),
    (v_acme_id, v_acme_co, 'CUSTOMER', 'Robert', 'Brown',   'robert@client.com',   '+44-20-1234', '{"tier":"gold","industry":"finance"}'),
    (v_globex_id, v_asia_div,'EMPLOYEE','Kenji',  'Tanaka', 'kenji.t@globex.jp',   '+81-3-4567',  '{"clearance":"low","plant":"tokyo"}'),
    (v_globex_id, v_mfg,   'EMPLOYEE', 'Wei',    'Zhang',   'wei.zhang@globex.cn', '+86-10-8910', '{"clearance":"medium","shift":"night"}'),
    (v_stark_id, v_rnd,    'EMPLOYEE', 'Tony',   'Stark',   'tony@stark.com',      '+1-212-1111', '{"clearance":"top","projects":["ironman","mk42"]}'),
    (v_stark_id, v_weapons,'EMPLOYEE', 'Pepper', 'Potts',   'pepper@stark.com',    '+1-212-2222', '{"clearance":"high","role":"executive"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_john_id    FROM persons WHERE tenant_id = v_acme_id   AND email = 'john.smith@acme.com';
  SELECT id INTO v_sarah_id   FROM persons WHERE tenant_id = v_acme_id   AND email = 'sarah.j@acme.com';
  SELECT id INTO v_michael_id FROM persons WHERE tenant_id = v_acme_id   AND email = 'm.chen@acme.com';
  SELECT id INTO v_emma_id    FROM persons WHERE tenant_id = v_acme_id   AND email = 'e.dubois@acme.eu';
  SELECT id INTO v_robert_id  FROM persons WHERE tenant_id = v_acme_id   AND email = 'robert@client.com';
  SELECT id INTO v_kenji_id   FROM persons WHERE tenant_id = v_globex_id AND email = 'kenji.t@globex.jp';
  SELECT id INTO v_wei_id     FROM persons WHERE tenant_id = v_globex_id AND email = 'wei.zhang@globex.cn';
  SELECT id INTO v_tony_id    FROM persons WHERE tenant_id = v_stark_id  AND email = 'tony@stark.com';
  SELECT id INTO v_pepper_id  FROM persons WHERE tenant_id = v_stark_id  AND email = 'pepper@stark.com';

  -- Employees
  INSERT INTO employees (tenant_id, person_id, employee_number, entity_id, position_title, hire_date, security_level, access_attributes) VALUES
    (v_acme_id,   v_john_id,   'EMP-00001', v_us_ops,   'Sales Director',           '2019-03-15', 8,  '{"clearance":"high"}'),
    (v_acme_id,   v_sarah_id,  'EMP-00002', v_sales,    'Sales Manager',            '2020-06-01', 5,  '{"clearance":"medium"}'),
    (v_acme_id,   v_michael_id,'EMP-00003', v_eng,      'Senior Software Engineer', '2021-01-10', 8,  '{"clearance":"high","specialization":"ai"}'),
    (v_acme_id,   v_emma_id,   'EMP-00004', v_eu_hq,    'Regional Director',        '2018-09-20', 5,  '{"clearance":"medium"}'),
    (v_globex_id, v_kenji_id,  'EMP-00001', v_asia_div, 'Plant Manager',            '2017-04-01', 3,  '{"clearance":"low","plant":"tokyo"}'),
    (v_globex_id, v_wei_id,    'EMP-00002', v_mfg,      'Manufacturing Engineer',   '2019-11-15', 5,  '{"clearance":"medium"}'),
    (v_stark_id,  v_tony_id,   'EMP-00001', v_rnd,      'Chief Technology Officer', '2008-05-02', 10, '{"clearance":"top"}'),
    (v_stark_id,  v_pepper_id, 'EMP-00002', v_weapons,  'Chief Operating Officer',  '2010-03-01', 8,  '{"clearance":"high"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_john_emp    FROM employees WHERE tenant_id = v_acme_id   AND employee_number = 'EMP-00001';
  SELECT id INTO v_sarah_emp   FROM employees WHERE tenant_id = v_acme_id   AND employee_number = 'EMP-00002';
  SELECT id INTO v_michael_emp FROM employees WHERE tenant_id = v_acme_id   AND employee_number = 'EMP-00003';
  SELECT id INTO v_emma_emp    FROM employees WHERE tenant_id = v_acme_id   AND employee_number = 'EMP-00004';
  SELECT id INTO v_kenji_emp   FROM employees WHERE tenant_id = v_globex_id AND employee_number = 'EMP-00001';
  SELECT id INTO v_wei_emp     FROM employees WHERE tenant_id = v_globex_id AND employee_number = 'EMP-00002';
  SELECT id INTO v_tony_emp    FROM employees WHERE tenant_id = v_stark_id  AND employee_number = 'EMP-00001';
  SELECT id INTO v_pepper_emp  FROM employees WHERE tenant_id = v_stark_id  AND employee_number = 'EMP-00002';

  -- Users
  INSERT INTO users (tenant_id, entity_id, person_id, employee_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength) VALUES
    (v_acme_id,   v_us_ops,   v_john_id,    v_john_emp,    'john.smith',   'john.smith@acme.com', v_pw, 'INTERNAL', 'ACTIVE', TRUE,  480, 85),
    (v_acme_id,   v_sales,    v_sarah_id,   v_sarah_emp,   'sarah.j',      'sarah.j@acme.com',    v_pw, 'INTERNAL', 'ACTIVE', FALSE, 480, 70),
    (v_acme_id,   v_eng,      v_michael_id, v_michael_emp, 'michael.chen', 'm.chen@acme.com',     v_pw, 'INTERNAL', 'ACTIVE', TRUE,  480, 90),
    (v_acme_id,   v_eu_hq,    v_emma_id,    v_emma_emp,    'emma.dubois',  'e.dubois@acme.eu',    v_pw, 'INTERNAL', 'ACTIVE', FALSE, 480, 75),
    (v_acme_id,   v_acme_co,  v_robert_id,  NULL,          'robert.brown', 'robert@client.com',   v_pw, 'CUSTOMER', 'ACTIVE', FALSE, 240, 60),
    (v_globex_id, v_asia_div, v_kenji_id,   v_kenji_emp,   'kenji.t',      'kenji.t@globex.jp',   v_pw, 'INTERNAL', 'ACTIVE', FALSE, 480, 65),
    (v_globex_id, v_mfg,      v_wei_id,     v_wei_emp,     'wei.zhang',    'wei.zhang@globex.cn', v_pw, 'INTERNAL', 'ACTIVE', FALSE, 480, 68),
    (v_stark_id,  v_rnd,      v_tony_id,    v_tony_emp,    'tony',         'tony@stark.com',      v_pw, 'INTERNAL', 'ACTIVE', TRUE,  480, 95),
    (v_stark_id,  v_weapons,  v_pepper_id,  v_pepper_emp,  'pepper',       'pepper@stark.com',    v_pw, 'INTERNAL', 'ACTIVE', TRUE,  480, 90)
  ON CONFLICT DO NOTHING;

  -- Sysadmin accounts (no person/employee link)
  INSERT INTO users (tenant_id, entity_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength) VALUES
    (v_acme_id,   v_acme_co,   'sysadmin', 'sysadmin@acme-corp.com', v_pw, 'SYSADMIN', 'ACTIVE', TRUE, 480, 95),
    (v_globex_id, v_globex_co, 'sysadmin', 'sysadmin@globex.com',    v_pw, 'SYSADMIN', 'ACTIVE', TRUE, 480, 95),
    (v_stark_id,  v_stark_co,  'sysadmin', 'sysadmin@stark.com',     v_pw, 'SYSADMIN', 'ACTIVE', TRUE, 480, 95)
  ON CONFLICT DO NOTHING;
END $$;

-- =====================================================================
-- STEP 6: ROLES, PERMISSIONS, ROLE ASSIGNMENTS (migrations 000404–000413)
-- =====================================================================

DO $$
DECLARE
  v_acme_id   UUID;  v_globex_id UUID;  v_stark_id  UUID;
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001';
  v_us_ops    UUID := 'a0000000-0000-0000-0000-000000000002';
  v_sales     UUID := 'a0000000-0000-0000-0000-000000000003';
  v_eu_hq     UUID := 'a0000000-0000-0000-0000-000000000005';
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001';
  v_asia_div  UUID := 'b0000000-0000-0000-0000-000000000002';
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001';
  v_weapons   UUID := 'c0000000-0000-0000-0000-000000000003';
  -- Roles
  v_acme_sysadmin UUID;  v_acme_admin UUID;  v_acme_mgr UUID;
  v_acme_emp UUID;  v_acme_fin UUID;
  v_globex_sysadmin UUID;  v_globex_mgr UUID;  v_globex_emp UUID;
  v_stark_sysadmin UUID;  v_stark_admin UUID;  v_stark_mgr UUID;  v_stark_emp UUID;
  -- Actions
  v_a_create UUID; v_a_read UUID; v_a_update UUID;
  v_a_delete UUID; v_a_approve UUID; v_a_export UUID;
  -- Resources
  v_r_employees UUID; v_r_payroll UUID; v_r_timesheets UUID;
  v_r_invoices UUID;  v_r_accounts UUID; v_r_finreports UUID;
  v_r_orders UUID;    v_r_customers UUID;
  v_r_users UUID;     v_r_audit UUID;
  -- Users
  v_john UUID; v_sarah UUID; v_michael UUID; v_emma UUID;
  v_kenji UUID; v_tony UUID; v_pepper UUID;
  v_acme_sys UUID; v_globex_sys UUID; v_stark_sys UUID;
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  SELECT id INTO v_a_create  FROM actions WHERE name = 'create'  AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_read    FROM actions WHERE name = 'read'    AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_update  FROM actions WHERE name = 'update'  AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_delete  FROM actions WHERE name = 'delete'  AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_approve FROM actions WHERE name = 'approve' AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_export  FROM actions WHERE name = 'export'  AND scope = 'SYSTEM' AND tenant_id IS NULL;

  SELECT id INTO v_r_employees  FROM resources WHERE name = 'employees';
  SELECT id INTO v_r_payroll    FROM resources WHERE name = 'payroll';
  SELECT id INTO v_r_timesheets FROM resources WHERE name = 'timesheets';
  SELECT id INTO v_r_invoices   FROM resources WHERE name = 'invoices';
  SELECT id INTO v_r_accounts   FROM resources WHERE name = 'accounts';
  SELECT id INTO v_r_finreports FROM resources WHERE name = 'financial-reports';
  SELECT id INTO v_r_orders     FROM resources WHERE name = 'orders';
  SELECT id INTO v_r_customers  FROM resources WHERE name = 'customers';
  SELECT id INTO v_r_users      FROM resources WHERE name = 'users';
  SELECT id INTO v_r_audit      FROM resources WHERE name = 'audit-log';

  SELECT id INTO v_john    FROM users WHERE email = 'john.smith@acme.com';
  SELECT id INTO v_sarah   FROM users WHERE email = 'sarah.j@acme.com';
  SELECT id INTO v_michael FROM users WHERE email = 'm.chen@acme.com';
  SELECT id INTO v_emma    FROM users WHERE email = 'e.dubois@acme.eu';
  SELECT id INTO v_kenji   FROM users WHERE email = 'kenji.t@globex.jp';
  SELECT id INTO v_tony    FROM users WHERE email = 'tony@stark.com';
  SELECT id INTO v_pepper  FROM users WHERE email = 'pepper@stark.com';
  SELECT id INTO v_acme_sys   FROM users WHERE email = 'sysadmin@acme-corp.com';
  SELECT id INTO v_globex_sys FROM users WHERE email = 'sysadmin@globex.com';
  SELECT id INTO v_stark_sys  FROM users WHERE email = 'sysadmin@stark.com';

  -- ---- Roles ----
  INSERT INTO roles (tenant_id, entity_id, name, display_name, role_type, is_system_role, is_active) VALUES
    (v_acme_id,   v_acme_co,   'sysadmin',          'System Administrator', 'SYSTEM',     TRUE,  TRUE),
    (v_acme_id,   v_acme_co,   'tenant-admin',       'Tenant Administrator', 'TENANT',     TRUE,  TRUE),
    (v_acme_id,   v_acme_co,   'manager',            'Department Manager',   'ENTITY',     FALSE, TRUE),
    (v_acme_id,   v_acme_co,   'employee',           'Regular Employee',     'TENANT',     FALSE, TRUE),
    (v_acme_id,   v_acme_co,   'finance-specialist', 'Finance Specialist',   'FUNCTIONAL', FALSE, TRUE),
    (v_globex_id, v_globex_co, 'sysadmin',           'System Administrator', 'SYSTEM',     TRUE,  TRUE),
    (v_globex_id, v_globex_co, 'manager',            'Department Manager',   'ENTITY',     FALSE, TRUE),
    (v_globex_id, v_globex_co, 'employee',           'Regular Employee',     'TENANT',     FALSE, TRUE),
    (v_stark_id,  v_stark_co,  'sysadmin',           'System Administrator', 'SYSTEM',     TRUE,  TRUE),
    (v_stark_id,  v_stark_co,  'tenant-admin',       'Tenant Administrator', 'TENANT',     TRUE,  TRUE),
    (v_stark_id,  v_stark_co,  'manager',            'Department Manager',   'ENTITY',     FALSE, TRUE),
    (v_stark_id,  v_stark_co,  'employee',           'Regular Employee',     'TENANT',     FALSE, TRUE)
  ON CONFLICT (tenant_id, name) DO NOTHING;

  SELECT id INTO v_acme_sysadmin FROM roles WHERE tenant_id = v_acme_id   AND name = 'sysadmin';
  SELECT id INTO v_acme_admin    FROM roles WHERE tenant_id = v_acme_id   AND name = 'tenant-admin';
  SELECT id INTO v_acme_mgr      FROM roles WHERE tenant_id = v_acme_id   AND name = 'manager';
  SELECT id INTO v_acme_emp      FROM roles WHERE tenant_id = v_acme_id   AND name = 'employee';
  SELECT id INTO v_acme_fin      FROM roles WHERE tenant_id = v_acme_id   AND name = 'finance-specialist';
  SELECT id INTO v_globex_sysadmin FROM roles WHERE tenant_id = v_globex_id AND name = 'sysadmin';
  SELECT id INTO v_globex_mgr      FROM roles WHERE tenant_id = v_globex_id AND name = 'manager';
  SELECT id INTO v_globex_emp      FROM roles WHERE tenant_id = v_globex_id AND name = 'employee';
  SELECT id INTO v_stark_sysadmin FROM roles WHERE tenant_id = v_stark_id  AND name = 'sysadmin';
  SELECT id INTO v_stark_admin    FROM roles WHERE tenant_id = v_stark_id  AND name = 'tenant-admin';
  SELECT id INTO v_stark_mgr      FROM roles WHERE tenant_id = v_stark_id  AND name = 'manager';
  SELECT id INTO v_stark_emp      FROM roles WHERE tenant_id = v_stark_id  AND name = 'employee';

  -- ---- Permissions (tenant × resource × action) ----
  INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
  SELECT v_acme_id, r.rid, a.aid, r.rname || ':' || a.aname, 'ALLOW', '{}'::jsonb
  FROM (VALUES
    (v_r_employees,'employees'), (v_r_payroll,'payroll'), (v_r_timesheets,'timesheets'),
    (v_r_invoices,'invoices'),   (v_r_accounts,'accounts'), (v_r_finreports,'financial-reports'),
    (v_r_orders,'orders'),       (v_r_customers,'customers'),
    (v_r_users,'security-users'),(v_r_audit,'audit-log')
  ) r(rid,rname)
  CROSS JOIN (VALUES (v_a_create,'create'),(v_a_read,'read'),(v_a_update,'update'),
                     (v_a_delete,'delete'),(v_a_approve,'approve'),(v_a_export,'export')) a(aid,aname)
  ON CONFLICT (tenant_id, resource_id, action_id, name) DO NOTHING;

  -- Payroll MFA condition
  UPDATE permissions SET conditions = '{"require_mfa":true}'::jsonb
  WHERE tenant_id = v_acme_id AND resource_id = v_r_payroll AND conditions = '{}'::jsonb;

  INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
  SELECT v_globex_id, r.rid, a.aid, r.rname || ':' || a.aname, 'ALLOW', '{}'::jsonb
  FROM (VALUES (v_r_employees,'employees'),(v_r_payroll,'payroll'),
               (v_r_invoices,'invoices'),(v_r_orders,'orders'),(v_r_customers,'customers')) r(rid,rname)
  CROSS JOIN (VALUES (v_a_create,'create'),(v_a_read,'read'),(v_a_update,'update'),
                     (v_a_delete,'delete'),(v_a_approve,'approve')) a(aid,aname)
  ON CONFLICT (tenant_id, resource_id, action_id, name) DO NOTHING;

  INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
  SELECT v_stark_id, r.rid, a.aid, r.rname || ':' || a.aname, 'ALLOW', '{}'::jsonb
  FROM (VALUES (v_r_employees,'employees'),(v_r_payroll,'payroll'),
               (v_r_invoices,'invoices'),(v_r_accounts,'accounts'),(v_r_orders,'orders'),
               (v_r_users,'security-users'),(v_r_audit,'audit-log')) r(rid,rname)
  CROSS JOIN (VALUES (v_a_create,'create'),(v_a_read,'read'),(v_a_update,'update'),
                     (v_a_delete,'delete'),(v_a_approve,'approve'),(v_a_export,'export')) a(aid,aname)
  ON CONFLICT (tenant_id, resource_id, action_id, name) DO NOTHING;

  -- ---- Role Permissions ----
  -- ACME sysadmin + admin → all ALLOW
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, r.rid, p.id, TRUE FROM (VALUES (v_acme_sysadmin),(v_acme_admin)) r(rid)
  CROSS JOIN permissions p WHERE p.tenant_id = v_acme_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ACME manager → READ, APPROVE, EXPORT
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_mgr, p.id, TRUE FROM permissions p JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id AND a.name IN ('read','approve','export') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ACME employee → READ
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_emp, p.id, TRUE FROM permissions p JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id AND a.name = 'read' AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ACME finance-specialist → finance resources CRUD
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_acme_id, v_acme_fin, p.id, TRUE FROM permissions p
  JOIN resources r ON p.resource_id = r.id JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id AND r.name IN ('invoices','accounts','financial-reports')
    AND a.name IN ('create','read','update','export') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Globex sysadmin → all; manager → READ+APPROVE; employee → READ
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_globex_id, v_globex_sysadmin, p.id, TRUE
  FROM permissions p WHERE p.tenant_id = v_globex_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_globex_id, v_globex_mgr, p.id, TRUE FROM permissions p JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_globex_id AND a.name IN ('read','approve') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_globex_id, v_globex_emp, p.id, TRUE FROM permissions p JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_globex_id AND a.name = 'read' AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- Stark sysadmin + admin → all; manager → READ+APPROVE+EXPORT; employee → READ
  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, r.rid, p.id, TRUE FROM (VALUES (v_stark_sysadmin),(v_stark_admin)) r(rid)
  CROSS JOIN permissions p WHERE p.tenant_id = v_stark_id AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, v_stark_mgr, p.id, TRUE FROM permissions p JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_stark_id AND a.name IN ('read','approve','export') AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  INSERT INTO role_permissions (tenant_id, role_id, permission_id, is_active)
  SELECT v_stark_id, v_stark_emp, p.id, TRUE FROM permissions p JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_stark_id AND a.name = 'read' AND p.effect = 'ALLOW'
  ON CONFLICT (tenant_id, role_id, permission_id, entity_scope) DO NOTHING;

  -- ---- User Role Assignments ----
  INSERT INTO user_roles (user_id, role_id, entity_id, assignment_type, is_active) VALUES
    (v_acme_sys,  v_acme_sysadmin, v_acme_co,  'DIRECT', TRUE),
    (v_john,      v_acme_mgr,      v_us_ops,   'DIRECT', TRUE),
    (v_john,      v_acme_emp,      v_acme_co,  'DIRECT', TRUE),
    (v_sarah,     v_acme_mgr,      v_sales,    'DIRECT', TRUE),
    (v_sarah,     v_acme_emp,      v_acme_co,  'DIRECT', TRUE),
    (v_michael,   v_acme_emp,      v_acme_co,  'DIRECT', TRUE),
    (v_michael,   v_acme_fin,      v_acme_co,  'DIRECT', TRUE),
    (v_emma,      v_acme_mgr,      v_eu_hq,    'DIRECT', TRUE),
    (v_emma,      v_acme_emp,      v_acme_co,  'DIRECT', TRUE),
    (v_globex_sys,v_globex_sysadmin,v_globex_co,'DIRECT',TRUE),
    (v_kenji,     v_globex_mgr,    v_asia_div, 'DIRECT', TRUE),
    (v_kenji,     v_globex_emp,    v_globex_co,'DIRECT', TRUE),
    (v_stark_sys, v_stark_sysadmin,v_stark_co, 'DIRECT', TRUE),
    (v_tony,      v_stark_admin,   v_stark_co, 'DIRECT', TRUE),
    (v_tony,      v_stark_emp,     v_stark_co, 'DIRECT', TRUE),
    (v_pepper,    v_stark_mgr,     v_stark_co, 'DIRECT', TRUE),
    (v_pepper,    v_stark_emp,     v_stark_co, 'DIRECT', TRUE)
  ON CONFLICT (user_id, role_id, entity_id) DO NOTHING;

  -- ---- Direct User Permissions ----
  -- Tony Stark: CEO override on security-users:delete
  INSERT INTO user_permissions (tenant_id, user_id, permission_id, entity_id, effect, reason, is_active)
  SELECT v_stark_id, v_tony, p.id, v_stark_co, 'ALLOW', 'CEO override — full platform access', TRUE
  FROM permissions p JOIN resources r ON p.resource_id = r.id JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_stark_id AND r.name = 'security-users' AND a.name = 'delete'
  ON CONFLICT (tenant_id, user_id, permission_id, entity_id) DO NOTHING;

  -- Michael Chen: DENY payroll:delete (safety restriction)
  INSERT INTO user_permissions (tenant_id, user_id, permission_id, entity_id, effect, reason, is_active)
  SELECT v_acme_id, v_michael, p.id, v_acme_co, 'DENY',
         'Engineers cannot delete payroll records', TRUE
  FROM permissions p JOIN resources r ON p.resource_id = r.id JOIN actions a ON p.action_id = a.id
  WHERE p.tenant_id = v_acme_id AND r.name = 'payroll' AND a.name = 'delete'
  ON CONFLICT (tenant_id, user_id, permission_id, entity_id) DO NOTHING;
END $$;

-- =====================================================================
-- STEP 7: USER SESSIONS AND AUDIT LOG (migrations 000304, 000305, 000450)
-- =====================================================================

DO $$
DECLARE
  v_acme_id   UUID;  v_globex_id UUID;  v_stark_id UUID;
  v_john UUID;  v_sarah UUID;  v_michael UUID;  v_emma UUID;
  v_kenji UUID; v_tony UUID;  v_pepper UUID;
  v_acme_sys UUID; v_stark_sys UUID;
  v_r_employees UUID; v_r_payroll UUID; v_r_invoices UUID;
  v_a_read UUID;  v_a_update UUID;  v_a_delete UUID;
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';
  SELECT id INTO v_john    FROM users WHERE email = 'john.smith@acme.com';
  SELECT id INTO v_sarah   FROM users WHERE email = 'sarah.j@acme.com';
  SELECT id INTO v_michael FROM users WHERE email = 'm.chen@acme.com';
  SELECT id INTO v_emma    FROM users WHERE email = 'e.dubois@acme.eu';
  SELECT id INTO v_kenji   FROM users WHERE email = 'kenji.t@globex.jp';
  SELECT id INTO v_tony    FROM users WHERE email = 'tony@stark.com';
  SELECT id INTO v_pepper  FROM users WHERE email = 'pepper@stark.com';
  SELECT id INTO v_acme_sys   FROM users WHERE email = 'sysadmin@acme-corp.com';
  SELECT id INTO v_stark_sys  FROM users WHERE email = 'sysadmin@stark.com';
  SELECT id INTO v_r_employees FROM resources WHERE name = 'employees';
  SELECT id INTO v_r_payroll   FROM resources WHERE name = 'payroll';
  SELECT id INTO v_r_invoices  FROM resources WHERE name = 'invoices';
  SELECT id INTO v_a_read   FROM actions WHERE name = 'read'   AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_update FROM actions WHERE name = 'update' AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_a_delete FROM actions WHERE name = 'delete' AND scope = 'SYSTEM' AND tenant_id IS NULL;

  -- User sessions (active)
  -- created_at defaults to NOW(); expires_at must be > created_at (CHECK constraint)
  INSERT INTO user_sessions (tenant_id, user_id, session_token, ip_address, user_agent, expires_at, risk_score, permissions, is_active) VALUES
    (v_acme_id,   v_john,   'sess-acme-john-001',    '10.0.1.101',  'Chrome/120 Windows',
     NOW()+INTERVAL '8h', 10, '{"hr.employees.read":true,"hr.payroll.read":true}'::jsonb, TRUE),
    (v_acme_id,   v_sarah,  'sess-acme-sarah-001',   '10.0.1.102',  'Safari/17 Mac',
     NOW()+INTERVAL '6h', 15, '{"hr.employees.read":true,"sales.orders.approve":true}'::jsonb, TRUE),
    (v_acme_id,   v_michael,'sess-acme-michael-001', '10.0.1.103',  'Firefox/121 Linux',
     NOW()+INTERVAL '8h', 5,  '{"hr.employees.read":true,"finance.invoices.read":true}'::jsonb, TRUE),
    (v_acme_id,   v_emma,   'sess-acme-emma-001',    '172.16.0.50', 'Edge/120 Windows',
     NOW()+INTERVAL '4h', 20, '{"hr.employees.read":true,"sales.orders.read":true}'::jsonb, TRUE),
    (v_globex_id, v_kenji,  'sess-globex-kenji-001', '192.168.10.5','Chrome/119 Windows',
     NOW()+INTERVAL '8h', 8,  '{"hr.employees.read":true}'::jsonb, TRUE),
    (v_stark_id,  v_tony,   'sess-stark-tony-001',   '10.10.0.1',   'Chrome/120 Mac',
     NOW()+INTERVAL '8h', 5,  '{"finance.invoices.create":true,"security.users.read":true}'::jsonb, TRUE),
    (v_stark_id,  v_pepper, 'sess-stark-pepper-001', '10.10.0.2',   'Safari/17 iPad',
     NOW()+INTERVAL '4h', 12, '{"hr.employees.read":true,"finance.invoices.read":true}'::jsonb, TRUE)
  ON CONFLICT (session_token) DO NOTHING;

  -- Historical (expired) sessions: created_at set to past so expires_at > created_at
  INSERT INTO user_sessions (tenant_id, user_id, session_token, ip_address, user_agent, created_at, expires_at, risk_score, permissions, is_active) VALUES
    (v_acme_id,  v_acme_sys,  'sess-acme-sys-prev',  '10.0.0.1',  'curl/8.0',
     NOW()-INTERVAL '9h', NOW()-INTERVAL '1h', 0, '{}'::jsonb, FALSE),
    (v_stark_id, v_stark_sys, 'sess-stark-sys-prev', '10.10.0.1', 'curl/8.0',
     NOW()-INTERVAL '10h', NOW()-INTERVAL '2h', 0, '{}'::jsonb, FALSE)
  ON CONFLICT (session_token) DO NOTHING;

  -- Audit log
  INSERT INTO audit_log (tenant_id, event_type, event_category, severity, user_id, resource_id, action_id, decision, risk_score, context) VALUES
    (v_acme_id,   'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_john,     NULL,         NULL,        'ALLOW', 5,
     '{"method":"password","ip":"10.0.1.101","mfa_used":true}'),
    (v_acme_id,   'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_sarah,    NULL,         NULL,        'ALLOW', 5,
     '{"method":"password","ip":"10.0.1.102","mfa_used":false}'),
    (v_globex_id, 'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_kenji,    NULL,         NULL,        'ALLOW', 5,
     '{"method":"password","ip":"192.168.10.5"}'),
    (v_stark_id,  'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_tony,     NULL,         NULL,        'ALLOW', 5,
     '{"method":"password","ip":"10.10.0.1","mfa_used":true}'),
    (v_acme_id,   'LOGIN_FAILED',      'AUTH',       'WARN',  NULL,       NULL,         NULL,        'DENY',  30,
     '{"attempted_email":"unknown@acme.com","ip":"45.33.12.99","reason":"invalid_credentials"}'),
    (v_acme_id,   'READ_ACCESS',       'ACCESS',     'INFO',  v_john,     v_r_employees,v_a_read,    'ALLOW', 5,
     '{"entity":"US Operations","records_accessed":12}'),
    (v_acme_id,   'READ_ACCESS',       'ACCESS',     'INFO',  v_sarah,    v_r_invoices, v_a_read,    'ALLOW', 5,
     '{"entity":"Sales Department","invoice_count":45}'),
    (v_acme_id,   'SENSITIVE_READ',    'ACCESS',     'WARN',  v_john,     v_r_payroll,  v_a_read,    'ALLOW', 40,
     '{"mfa_verified":true,"payroll_period":"2026-02"}'),
    (v_acme_id,   'UPDATE_RECORD',     'DATA',       'INFO',  v_sarah,    v_r_invoices, v_a_update,  'ALLOW', 10,
     '{"invoice_id":"ACME-INV-000042","change":"status:approved"}'),
    (v_stark_id,  'UPDATE_RECORD',     'DATA',       'INFO',  v_tony,     v_r_employees,v_a_update,  'ALLOW', 10,
     '{"employee":"pepper","field":"security_level","from":8,"to":9}'),
    (v_acme_id,   'DELETE_DENIED',     'ACCESS',     'HIGH',  v_michael,  v_r_payroll,  v_a_delete,  'DENY',  75,
     '{"reason":"user_permission_deny","resource":"payroll"}'),
    (v_acme_id,   'ROLE_ASSIGNED',     'ADMIN',      'INFO',  v_acme_sys, NULL,         NULL,        'ALLOW', 20,
     '{"target_user":"sarah.j","role":"manager","entity":"Sales Department"}'),
    (v_stark_id,  'PERMISSION_GRANTED','ADMIN',      'WARN',  v_stark_sys,NULL,         NULL,        'ALLOW', 35,
     '{"target_user":"tony","permission":"security-users:delete","reason":"CEO override"}'),
    (v_acme_id,   'PAYROLL_EXPORT',    'COMPLIANCE', 'WARN',  v_acme_sys, v_r_payroll,  v_a_update,  'ALLOW', 60,
     '{"period":"2026-Q1","records":4,"mfa_verified":true,"compliance_tag":"SOX"}'),
    (v_stark_id,  'SENSITIVE_ACCESS',  'COMPLIANCE', 'HIGH',  v_tony,     v_r_employees,v_a_update,  'ALLOW', 55,
     '{"clearance_required":"top","user_clearance":"top","compliance_tag":"ITAR"}')
  ON CONFLICT DO NOTHING;
END $$;

-- =====================================================================
-- STEP 8: FEATURE FLAGS (migrations 000801+)
-- =====================================================================

DO $$
DECLARE
  v_acme_id   UUID;  v_globex_id UUID;  v_stark_id UUID;
  v_acme_co   UUID := 'a0000000-0000-0000-0000-000000000001';
  v_globex_co UUID := 'b0000000-0000-0000-0000-000000000001';
  v_stark_co  UUID := 'c0000000-0000-0000-0000-000000000001';
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  INSERT INTO feature_flags (tenant_id, entity_id, name, description, flag_type, default_value, rollout_percentage, metadata) VALUES
    -- ACME
    (v_acme_id, v_acme_co, 'advanced_reporting',    'Advanced analytics and custom report builder',    'boolean', TRUE,  100, '{"module":"finance","tier":"ENTERPRISE"}'),
    (v_acme_id, v_acme_co, 'ai_invoice_matching',   'AI-powered automatic invoice to PO matching',     'boolean', FALSE,  50, '{"module":"finance","beta":true}'),
    (v_acme_id, v_acme_co, 'multi_currency_support', 'Transactions and reporting in multiple currencies','boolean', TRUE, 100, '{"module":"finance"}'),
    (v_acme_id, v_acme_co, 'hr_self_service',        'Employee self-service portal',                   'boolean', TRUE,  100, '{"module":"hr"}'),
    (v_acme_id, v_acme_co, 'mobile_app_access',      'Mobile application access for field employees',  'boolean', FALSE,  30, '{"module":"core","beta":true}'),
    (v_acme_id, v_acme_co, 'bulk_import_limit',      'Maximum records per bulk import',                'number',  FALSE, NULL,'{"module":"core","default_value":5000}'),
    -- Globex
    (v_globex_id, v_globex_co, 'advanced_reporting',        'Advanced analytics and custom report builder', 'boolean', FALSE,   0, '{"module":"finance"}'),
    (v_globex_id, v_globex_co, 'inventory_lot_tracking',    'Mandatory lot and batch tracking',             'boolean', TRUE,  100, '{"module":"inventory","compliance":"ISO9001"}'),
    (v_globex_id, v_globex_co, 'quality_inspection_workflow','Inline quality inspection for orders',        'boolean', TRUE,  100, '{"module":"inventory","industry":"manufacturing"}'),
    (v_globex_id, v_globex_co, 'multi_plant_consolidation', 'Consolidated reporting across all plants',     'boolean', FALSE,  20, '{"module":"finance","beta":true}'),
    -- Stark
    (v_stark_id, v_stark_co, 'advanced_reporting',      'Advanced analytics and custom report builder',       'boolean', TRUE,  100, '{"module":"finance"}'),
    (v_stark_id, v_stark_co, 'classified_data_masking', 'Mask classified fields by security clearance level', 'boolean', TRUE,  100, '{"module":"security","compliance":"ITAR,DFARS","required":true}'),
    (v_stark_id, v_stark_co, 'enhanced_audit_trail',    'Granular field-level change history',               'boolean', TRUE,  100, '{"module":"security","compliance":"ITAR","retention_days":2555}'),
    (v_stark_id, v_stark_co, 'ai_invoice_matching',     'AI-powered automatic invoice to PO matching',        'boolean', TRUE,  100, '{"module":"finance"}'),
    (v_stark_id, v_stark_co, 'mfa_required_all',        'Require MFA for all user logins',                   'boolean', TRUE,  100, '{"module":"security","policy":"zero_trust"}'),
    (v_stark_id, v_stark_co, 'project_cost_tracking',   'Link expenses to R&D project codes',                'boolean', TRUE,  100, '{"module":"finance","industry":"defense"}')
  ON CONFLICT (tenant_id, name) DO NOTHING;
END $$;

-- =====================================================================
-- VERIFICATION SUMMARY
-- =====================================================================
SELECT 'Modules'         AS "Table", COUNT(*) AS "Rows" FROM modules        WHERE scope = 'SYSTEM'
UNION ALL
SELECT 'Actions (SYSTEM)',            COUNT(*) FROM actions        WHERE scope = 'SYSTEM'
UNION ALL
SELECT 'Resources',                   COUNT(*) FROM resources
UNION ALL
SELECT 'Tenants',                     COUNT(*) FROM tenants        WHERE deleted_at IS NULL
UNION ALL
SELECT 'Entities',                    COUNT(*) FROM entities       WHERE deleted_at IS NULL
UNION ALL
SELECT 'Hierarchy Paths',             COUNT(*) FROM hierarchy_paths
UNION ALL
SELECT 'Persons',                     COUNT(*) FROM persons        WHERE deleted_at IS NULL
UNION ALL
SELECT 'Employees',                   COUNT(*) FROM employees      WHERE deleted_at IS NULL
UNION ALL
SELECT 'Users',                       COUNT(*) FROM users          WHERE deleted_at IS NULL
UNION ALL
SELECT 'Roles',                       COUNT(*) FROM roles          WHERE deleted_at IS NULL
UNION ALL
SELECT 'Permissions',                 COUNT(*) FROM permissions    WHERE is_active = TRUE
UNION ALL
SELECT 'Role Permissions',            COUNT(*) FROM role_permissions WHERE is_active = TRUE
UNION ALL
SELECT 'User Roles',                  COUNT(*) FROM user_roles     WHERE is_active = TRUE
UNION ALL
SELECT 'User Permissions',            COUNT(*) FROM user_permissions WHERE is_active = TRUE
UNION ALL
SELECT 'User Sessions',               COUNT(*) FROM user_sessions
UNION ALL
SELECT 'Audit Log Entries',           COUNT(*) FROM audit_log
UNION ALL
SELECT 'Feature Flags',               COUNT(*) FROM feature_flags  WHERE deleted_at IS NULL
ORDER BY 1;
