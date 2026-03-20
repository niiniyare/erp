-- =====================================================================
-- USER SESSIONS AND AUDIT LOG
-- Migrations: 000304 (user_sessions), 000305 (sessions + permissions col),
--             000450 (audit_log)
-- Notes:
--   - user_sessions.expires_at must be > created_at (CHECK constraint)
--   - user_sessions.permissions JSONB added in 000305 (NOT NULL DEFAULT '{}')
--   - audit_log.severity: LOW/INFO/WARN/HIGH/CRITICAL
--   - audit_log.event_category: ACCESS/ADMIN/DATA/AUTH/SYSTEM/COMPLIANCE
-- =====================================================================
DO $$
DECLARE
  v_acme_id   UUID;
  v_globex_id UUID;
  v_stark_id  UUID;

  v_john_uid    UUID;
  v_sarah_uid   UUID;
  v_michael_uid UUID;
  v_emma_uid    UUID;
  v_kenji_uid   UUID;
  v_tony_uid    UUID;
  v_pepper_uid  UUID;
  v_acme_sys_uid   UUID;
  v_stark_sys_uid  UUID;

  v_res_employees UUID;
  v_res_payroll   UUID;
  v_res_invoices  UUID;

  v_act_read   UUID;
  v_act_update UUID;
  v_act_delete UUID;
BEGIN
  SELECT id INTO v_acme_id   FROM tenants WHERE slug = 'acme-corp';
  SELECT id INTO v_globex_id FROM tenants WHERE slug = 'globex';
  SELECT id INTO v_stark_id  FROM tenants WHERE slug = 'stark-ind';

  SELECT id INTO v_john_uid    FROM users WHERE email = 'john.smith@acme.com';
  SELECT id INTO v_sarah_uid   FROM users WHERE email = 'sarah.j@acme.com';
  SELECT id INTO v_michael_uid FROM users WHERE email = 'm.chen@acme.com';
  SELECT id INTO v_emma_uid    FROM users WHERE email = 'e.dubois@acme.eu';
  SELECT id INTO v_kenji_uid   FROM users WHERE email = 'kenji.t@globex.jp';
  SELECT id INTO v_tony_uid    FROM users WHERE email = 'tony@stark.com';
  SELECT id INTO v_pepper_uid  FROM users WHERE email = 'pepper@stark.com';
  SELECT id INTO v_acme_sys_uid   FROM users WHERE email = 'sysadmin@acme-corp.com';
  SELECT id INTO v_stark_sys_uid  FROM users WHERE email = 'sysadmin@stark.com';

  SELECT id INTO v_res_employees FROM resources WHERE name = 'employees';
  SELECT id INTO v_res_payroll   FROM resources WHERE name = 'payroll';
  SELECT id INTO v_res_invoices  FROM resources WHERE name = 'invoices';

  SELECT id INTO v_act_read   FROM actions WHERE name = 'read'   AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_update FROM actions WHERE name = 'update' AND scope = 'SYSTEM' AND tenant_id IS NULL;
  SELECT id INTO v_act_delete FROM actions WHERE name = 'delete' AND scope = 'SYSTEM' AND tenant_id IS NULL;

  -- -----------------------------------------------------------------------
  -- USER SESSIONS
  -- session_token must be UNIQUE; permissions JSONB is NOT NULL (DEFAULT '{}')
  -- CHECK (expires_at > created_at): for expired sessions, override created_at
  -- -----------------------------------------------------------------------

  -- Active sessions (created_at = NOW() by default; expires_at is in the future)
  INSERT INTO user_sessions (
    tenant_id, user_id, session_token, ip_address, user_agent,
    expires_at, risk_score, permissions, is_active
  ) VALUES
    (v_acme_id, v_john_uid,    'sess-acme-john-001',    '10.0.1.101', 'Mozilla/5.0 (Windows NT 10.0) Chrome/120',
     NOW() + INTERVAL '8 hours', 10,
     '{"hr.employees.read":true,"hr.payroll.read":true,"sales.orders.read":true}'::jsonb, TRUE),

    (v_acme_id, v_sarah_uid,   'sess-acme-sarah-001',   '10.0.1.102', 'Mozilla/5.0 (Macintosh) Safari/17',
     NOW() + INTERVAL '6 hours', 15,
     '{"hr.employees.read":true,"sales.orders.read":true,"sales.orders.approve":true}'::jsonb, TRUE),

    (v_acme_id, v_michael_uid, 'sess-acme-michael-001', '10.0.1.103', 'Mozilla/5.0 (Linux) Firefox/121',
     NOW() + INTERVAL '8 hours', 5,
     '{"hr.employees.read":true,"finance.invoices.read":true,"finance.accounts.read":true}'::jsonb, TRUE),

    (v_acme_id, v_emma_uid,    'sess-acme-emma-001',    '172.16.0.50','Mozilla/5.0 (Windows NT 10.0) Edge/120',
     NOW() + INTERVAL '4 hours', 20,
     '{"hr.employees.read":true,"sales.orders.read":true}'::jsonb, TRUE),

    (v_globex_id, v_kenji_uid, 'sess-globex-kenji-001', '192.168.10.5','Mozilla/5.0 (Windows NT 10.0) Chrome/119',
     NOW() + INTERVAL '8 hours', 8,
     '{"hr.employees.read":true,"sales.orders.read":true}'::jsonb, TRUE),

    (v_stark_id, v_tony_uid,   'sess-stark-tony-001',   '10.10.0.1',  'Mozilla/5.0 (Macintosh) Chrome/120',
     NOW() + INTERVAL '8 hours', 5,
     '{"hr.employees.read":true,"finance.invoices.create":true,"security.users.read":true}'::jsonb, TRUE),

    (v_stark_id, v_pepper_uid, 'sess-stark-pepper-001', '10.10.0.2',  'Mozilla/5.0 (iPad) Safari/17',
     NOW() + INTERVAL '4 hours', 12,
     '{"hr.employees.read":true,"finance.invoices.read":true}'::jsonb, TRUE)
  ON CONFLICT (session_token) DO NOTHING;

  -- Historical (expired) sessions: set created_at to the past so expires_at > created_at
  INSERT INTO user_sessions (
    tenant_id, user_id, session_token, ip_address, user_agent,
    created_at, expires_at, risk_score, permissions, is_active
  ) VALUES
    (v_acme_id,  v_acme_sys_uid,  'sess-acme-sys-expired', '10.0.0.1',  'curl/8.0',
     NOW() - INTERVAL '9 hours',  NOW() - INTERVAL '1 hour',  0, '{}'::jsonb, FALSE),

    (v_stark_id, v_stark_sys_uid, 'sess-stark-sys-prev',   '10.10.0.1', 'curl/8.0',
     NOW() - INTERVAL '10 hours', NOW() - INTERVAL '2 hours', 0, '{}'::jsonb, FALSE)
  ON CONFLICT (session_token) DO NOTHING;

  -- -----------------------------------------------------------------------
  -- AUDIT LOG ENTRIES
  -- Realistic mix of AUTH, ACCESS, DATA, ADMIN, COMPLIANCE events
  -- -----------------------------------------------------------------------
  INSERT INTO audit_log (tenant_id, event_type, event_category, severity, user_id, resource_id, action_id, decision, risk_score, context)
  VALUES
    -- Authentication events
    (v_acme_id,   'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_john_uid,    NULL,              NULL,          'ALLOW', 5,
     '{"method":"password","ip":"10.0.1.101","mfa_used":true}'),
    (v_acme_id,   'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_sarah_uid,   NULL,              NULL,          'ALLOW', 5,
     '{"method":"password","ip":"10.0.1.102","mfa_used":false}'),
    (v_globex_id, 'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_kenji_uid,   NULL,              NULL,          'ALLOW', 5,
     '{"method":"password","ip":"192.168.10.5","mfa_used":false}'),
    (v_stark_id,  'LOGIN_SUCCESS',     'AUTH',       'INFO',  v_tony_uid,    NULL,              NULL,          'ALLOW', 5,
     '{"method":"password","ip":"10.10.0.1","mfa_used":true}'),

    -- Failed login attempt
    (v_acme_id,   'LOGIN_FAILED',      'AUTH',       'WARN',  NULL,          NULL,              NULL,          'DENY',  30,
     '{"attempted_email":"unknown@acme.com","ip":"45.33.12.99","reason":"invalid_credentials"}'),

    -- Data access events
    (v_acme_id,   'READ_ACCESS',       'ACCESS',     'INFO',  v_john_uid,    v_res_employees,   v_act_read,    'ALLOW', 5,
     '{"entity":"US Operations","records_accessed":12}'),
    (v_acme_id,   'READ_ACCESS',       'ACCESS',     'INFO',  v_sarah_uid,   v_res_invoices,    v_act_read,    'ALLOW', 5,
     '{"entity":"Sales Department","invoice_count":45}'),
    (v_acme_id,   'SENSITIVE_READ',    'ACCESS',     'WARN',  v_john_uid,    v_res_payroll,     v_act_read,    'ALLOW', 40,
     '{"entity":"US Operations","mfa_verified":true,"payroll_period":"2026-02"}'),

    -- Write / mutation events
    (v_acme_id,   'UPDATE_RECORD',     'DATA',       'INFO',  v_sarah_uid,   v_res_invoices,    v_act_update,  'ALLOW', 10,
     '{"invoice_id":"ACME-INV-000042","change":"status:approved"}'),
    (v_stark_id,  'UPDATE_RECORD',     'DATA',       'INFO',  v_tony_uid,    v_res_employees,   v_act_update,  'ALLOW', 10,
     '{"employee":"pepper","field":"security_level","from":8,"to":9}'),

    -- Denied access attempts
    (v_acme_id,   'DELETE_DENIED',     'ACCESS',     'HIGH',  v_michael_uid, v_res_payroll,     v_act_delete,  'DENY',  75,
     '{"reason":"user_permission_deny","resource":"payroll"}'),

    -- Admin actions
    (v_acme_id,   'ROLE_ASSIGNED',     'ADMIN',      'INFO',  v_acme_sys_uid,NULL,              NULL,          'ALLOW', 20,
     '{"target_user":"sarah.j","role":"manager","entity":"Sales Department"}'),
    (v_stark_id,  'PERMISSION_GRANTED','ADMIN',      'WARN',  v_stark_sys_uid,NULL,             NULL,          'ALLOW', 35,
     '{"target_user":"tony","permission":"security-users:delete","reason":"CEO override"}'),

    -- Compliance events
    (v_acme_id,   'PAYROLL_EXPORT',    'COMPLIANCE', 'WARN',  v_acme_sys_uid,v_res_payroll,     v_act_update,  'ALLOW', 60,
     '{"period":"2026-Q1","records":4,"mfa_verified":true,"compliance_tag":"SOX"}'),
    (v_stark_id,  'SENSITIVE_ACCESS',  'COMPLIANCE', 'HIGH',  v_tony_uid,    v_res_employees,   v_act_update,  'ALLOW', 55,
     '{"clearance_required":"top","user_clearance":"top","resource":"security-users","compliance_tag":"ITAR"}')
  ON CONFLICT DO NOTHING;

END $$;
