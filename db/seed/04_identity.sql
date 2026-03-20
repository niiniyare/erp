-- =====================================================================
-- PERSONS, EMPLOYEES, AND USERS
-- Migrations: 000301 (persons), 000302 (employees), 000303 (users)
-- Notes:
--   - persons.entity_id is NOT NULL (ON DELETE RESTRICT)
--   - employees.hire_date is NOT NULL; employees.entity_id is NOT NULL
--   - users.entity_id is NOT NULL
--   - users.lockout_until must be NULL or in the future (CHECK constraint)
--   - Passwords: 'Password123!' hashed with bcrypt via pgcrypto crypt()
--   - ON CONFLICT DO NOTHING handles EXCLUDE constraints on email/username per tenant
-- =====================================================================
DO $$
DECLARE
  v_acme_id   UUID;
  v_globex_id UUID;
  v_stark_id  UUID;

  -- Entity UUIDs (hardcoded, set in 03_entities.sql)
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

  -- Person ID variables (for linking employees & users)
  v_john_id    UUID;
  v_sarah_id   UUID;
  v_michael_id UUID;
  v_emma_id    UUID;
  v_robert_id  UUID;
  v_kenji_id   UUID;
  v_wei_id     UUID;
  v_tony_id    UUID;
  v_pepper_id  UUID;

  -- Employee ID variables (for linking users)
  v_john_emp_id    UUID;
  v_sarah_emp_id   UUID;
  v_michael_emp_id UUID;
  v_emma_emp_id    UUID;
  v_kenji_emp_id   UUID;
  v_wei_emp_id     UUID;
  v_tony_emp_id    UUID;
  v_pepper_emp_id  UUID;

  v_pw_hash TEXT;
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  v_pw_hash := crypt('Password123!', gen_salt('bf'));

  -- -----------------------------------------------------------------------
  -- ACME PERSONS
  -- -----------------------------------------------------------------------
  INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes)
  VALUES
    (v_acme_id, v_us_ops,   'EMPLOYEE', 'John',   'Smith',   'john.smith@acme.com', '+1-555-1234', '{"clearance":"high","department":"sales"}'),
    (v_acme_id, v_sales,    'EMPLOYEE', 'Sarah',  'Johnson', 'sarah.j@acme.com',    '+1-555-5678', '{"clearance":"medium","region":"us"}'),
    (v_acme_id, v_eng,      'EMPLOYEE', 'Michael','Chen',    'm.chen@acme.com',     '+1-555-9012', '{"clearance":"high","specialization":"ai"}'),
    (v_acme_id, v_eu_hq,    'EMPLOYEE', 'Emma',   'Dubois',  'e.dubois@acme.eu',    '+33-1-2345',  '{"clearance":"medium","language":"fr"}'),
    (v_acme_id, v_acme_co,  'CUSTOMER', 'Robert', 'Brown',   'robert@client.com',   '+44-20-1234', '{"tier":"gold","industry":"finance"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_john_id    FROM persons WHERE tenant_id = v_acme_id AND email = 'john.smith@acme.com';
  SELECT id INTO v_sarah_id   FROM persons WHERE tenant_id = v_acme_id AND email = 'sarah.j@acme.com';
  SELECT id INTO v_michael_id FROM persons WHERE tenant_id = v_acme_id AND email = 'm.chen@acme.com';
  SELECT id INTO v_emma_id    FROM persons WHERE tenant_id = v_acme_id AND email = 'e.dubois@acme.eu';
  SELECT id INTO v_robert_id  FROM persons WHERE tenant_id = v_acme_id AND email = 'robert@client.com';

  -- -----------------------------------------------------------------------
  -- GLOBEX PERSONS
  -- -----------------------------------------------------------------------
  INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes)
  VALUES
    (v_globex_id, v_asia_div, 'EMPLOYEE', 'Kenji', 'Tanaka', 'kenji.t@globex.jp',    '+81-3-4567',  '{"clearance":"low","plant":"tokyo"}'),
    (v_globex_id, v_mfg,      'EMPLOYEE', 'Wei',   'Zhang',  'wei.zhang@globex.cn',  '+86-10-8910', '{"clearance":"medium","shift":"night"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_kenji_id FROM persons WHERE tenant_id = v_globex_id AND email = 'kenji.t@globex.jp';
  SELECT id INTO v_wei_id   FROM persons WHERE tenant_id = v_globex_id AND email = 'wei.zhang@globex.cn';

  -- -----------------------------------------------------------------------
  -- STARK PERSONS
  -- -----------------------------------------------------------------------
  INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes)
  VALUES
    (v_stark_id, v_rnd,     'EMPLOYEE', 'Tony',   'Stark', 'tony@stark.com',   '+1-212-1111', '{"clearance":"top","projects":["ironman","mk42"]}'),
    (v_stark_id, v_weapons, 'EMPLOYEE', 'Pepper', 'Potts', 'pepper@stark.com', '+1-212-2222', '{"clearance":"high","role":"executive"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_tony_id   FROM persons WHERE tenant_id = v_stark_id AND email = 'tony@stark.com';
  SELECT id INTO v_pepper_id FROM persons WHERE tenant_id = v_stark_id AND email = 'pepper@stark.com';

  -- -----------------------------------------------------------------------
  -- EMPLOYEES
  -- hire_date and entity_id are NOT NULL
  -- employee_number must be unique per tenant (EXCLUDE constraint)
  -- -----------------------------------------------------------------------

  -- ACME employees
  INSERT INTO employees (tenant_id, person_id, employee_number, entity_id, position_title, hire_date, security_level, access_attributes)
  VALUES
    (v_acme_id, v_john_id,    'EMP-00001', v_us_ops,  'Sales Director',          '2019-03-15', 8, '{"clearance":"high","department":"sales"}'),
    (v_acme_id, v_sarah_id,   'EMP-00002', v_sales,   'Sales Manager',           '2020-06-01', 5, '{"clearance":"medium","region":"us"}'),
    (v_acme_id, v_michael_id, 'EMP-00003', v_eng,     'Senior Software Engineer','2021-01-10', 8, '{"clearance":"high","specialization":"ai"}'),
    (v_acme_id, v_emma_id,    'EMP-00004', v_eu_hq,   'Regional Director',       '2018-09-20', 5, '{"clearance":"medium","language":"fr"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_john_emp_id    FROM employees WHERE tenant_id = v_acme_id AND employee_number = 'EMP-00001';
  SELECT id INTO v_sarah_emp_id   FROM employees WHERE tenant_id = v_acme_id AND employee_number = 'EMP-00002';
  SELECT id INTO v_michael_emp_id FROM employees WHERE tenant_id = v_acme_id AND employee_number = 'EMP-00003';
  SELECT id INTO v_emma_emp_id    FROM employees WHERE tenant_id = v_acme_id AND employee_number = 'EMP-00004';

  -- Globex employees
  INSERT INTO employees (tenant_id, person_id, employee_number, entity_id, position_title, hire_date, security_level, access_attributes)
  VALUES
    (v_globex_id, v_kenji_id, 'EMP-00001', v_asia_div, 'Plant Manager',         '2017-04-01', 3, '{"clearance":"low","plant":"tokyo"}'),
    (v_globex_id, v_wei_id,   'EMP-00002', v_mfg,      'Manufacturing Engineer','2019-11-15', 5, '{"clearance":"medium","shift":"night"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_kenji_emp_id FROM employees WHERE tenant_id = v_globex_id AND employee_number = 'EMP-00001';
  SELECT id INTO v_wei_emp_id   FROM employees WHERE tenant_id = v_globex_id AND employee_number = 'EMP-00002';

  -- Stark employees
  INSERT INTO employees (tenant_id, person_id, employee_number, entity_id, position_title, hire_date, security_level, access_attributes)
  VALUES
    (v_stark_id, v_tony_id,   'EMP-00001', v_rnd,     'Chief Technology Officer','2008-05-02', 10, '{"clearance":"top","projects":["ironman","mk42"]}'),
    (v_stark_id, v_pepper_id, 'EMP-00002', v_weapons, 'Chief Operating Officer', '2010-03-01', 8,  '{"clearance":"high","role":"executive"}')
  ON CONFLICT DO NOTHING;

  SELECT id INTO v_tony_emp_id   FROM employees WHERE tenant_id = v_stark_id AND employee_number = 'EMP-00001';
  SELECT id INTO v_pepper_emp_id FROM employees WHERE tenant_id = v_stark_id AND employee_number = 'EMP-00002';

  -- -----------------------------------------------------------------------
  -- USERS
  -- entity_id is NOT NULL; lockout_until must be NULL or in the future
  -- user_type: INTERNAL, CUSTOMER, VENDOR, PARTNER, SYSADMIN
  -- account_status: ACTIVE, INACTIVE, LOCKED, SUSPENDED, PENDING_VERIFICATION
  -- mfa_enabled: TRUE for high/top clearance users
  -- -----------------------------------------------------------------------

  -- ACME users (linked to persons/employees)
  INSERT INTO users (tenant_id, entity_id, person_id, employee_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength)
  VALUES
    (v_acme_id, v_us_ops, v_john_id,    v_john_emp_id,    'john.smith',  'john.smith@acme.com', v_pw_hash, 'INTERNAL', 'ACTIVE', TRUE,  480, 85),
    (v_acme_id, v_sales,  v_sarah_id,   v_sarah_emp_id,   'sarah.j',     'sarah.j@acme.com',    v_pw_hash, 'INTERNAL', 'ACTIVE', FALSE, 480, 70),
    (v_acme_id, v_eng,    v_michael_id, v_michael_emp_id, 'michael.chen','m.chen@acme.com',     v_pw_hash, 'INTERNAL', 'ACTIVE', TRUE,  480, 90),
    (v_acme_id, v_eu_hq,  v_emma_id,    v_emma_emp_id,    'emma.dubois', 'e.dubois@acme.eu',    v_pw_hash, 'INTERNAL', 'ACTIVE', FALSE, 480, 75),
    (v_acme_id, v_acme_co,v_robert_id,  NULL,             'robert.brown','robert@client.com',   v_pw_hash, 'CUSTOMER', 'ACTIVE', FALSE, 240, 60)
  ON CONFLICT DO NOTHING;

  -- ACME sysadmin account (no person/employee link)
  INSERT INTO users (tenant_id, entity_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength)
  VALUES
    (v_acme_id, v_acme_co, 'sysadmin', 'sysadmin@acme-corp.com', v_pw_hash, 'SYSADMIN', 'ACTIVE', TRUE, 480, 95)
  ON CONFLICT DO NOTHING;

  -- Globex users
  INSERT INTO users (tenant_id, entity_id, person_id, employee_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength)
  VALUES
    (v_globex_id, v_asia_div, v_kenji_id, v_kenji_emp_id, 'kenji.t',    'kenji.t@globex.jp',   v_pw_hash, 'INTERNAL', 'ACTIVE', FALSE, 480, 65),
    (v_globex_id, v_mfg,      v_wei_id,   v_wei_emp_id,   'wei.zhang',  'wei.zhang@globex.cn', v_pw_hash, 'INTERNAL', 'ACTIVE', FALSE, 480, 68)
  ON CONFLICT DO NOTHING;

  INSERT INTO users (tenant_id, entity_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength)
  VALUES
    (v_globex_id, v_globex_co, 'sysadmin', 'sysadmin@globex.com', v_pw_hash, 'SYSADMIN', 'ACTIVE', TRUE, 480, 95)
  ON CONFLICT DO NOTHING;

  -- Stark users
  INSERT INTO users (tenant_id, entity_id, person_id, employee_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength)
  VALUES
    (v_stark_id, v_rnd,     v_tony_id,   v_tony_emp_id,   'tony',   'tony@stark.com',   v_pw_hash, 'INTERNAL', 'ACTIVE', TRUE,  480, 95),
    (v_stark_id, v_weapons, v_pepper_id, v_pepper_emp_id, 'pepper', 'pepper@stark.com', v_pw_hash, 'INTERNAL', 'ACTIVE', TRUE,  480, 90)
  ON CONFLICT DO NOTHING;

  INSERT INTO users (tenant_id, entity_id, username, email, password_hash, user_type, account_status, mfa_enabled, session_timeout_minutes, password_strength)
  VALUES
    (v_stark_id, v_stark_co, 'sysadmin', 'sysadmin@stark.com', v_pw_hash, 'SYSADMIN', 'ACTIVE', TRUE, 480, 95)
  ON CONFLICT DO NOTHING;

END $$;
