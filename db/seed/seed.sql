-- =====================================================================
-- DEMO DATA GENERATION FOR MULTI-TENANT RBAC/ABAC SCHEMA
-- =====================================================================

-- Disable triggers for data generation
SET session_replication_role = replica;

-- =====================================================================
-- TENANT DATA
-- =====================================================================
INSERT INTO tenants (slug, name, email, subdomain, industry, company_size, tax_id) VALUES
('acme-corp', 'ACME Corporation', 'admin@acme-corp.com', 'acme', 'Technology', 'large', 'TAX-ACME-12345'),
('globex', 'Globex Corporation', 'info@globex.com', 'globex', 'Manufacturing', 'enterprise', 'TAX-GLOBEX-67890'),
('stark-ind', 'Stark Industries', 'contact@stark.com', 'stark', 'Defense', 'enterprise', 'TAX-STARK-45678');

-- =====================================================================
-- ENTITY HIERARCHY
-- =====================================================================
-- ACME Entities
WITH acme AS (SELECT id FROM tenants WHERE slug = 'acme-corp')
INSERT INTO entities (tenant_id, name, code, type, parent_id, accrual_method, fy_start_month) VALUES
((SELECT id FROM acme), 'ACME Corporation', 'ACME', 'COMPANY', NULL, true, 1),
((SELECT id FROM acme), 'US Operations', 'US-OPS', 'REGION', (SELECT uuid FROM entities WHERE name = 'ACME Corporation'), true, 1),
((SELECT id FROM acme), 'Sales Department', 'SALES', 'DEPARTMENT', (SELECT uuid FROM entities WHERE name = 'US Operations'), true, 1),
((SELECT id FROM acme), 'Engineering', 'ENG', 'DEPARTMENT', (SELECT uuid FROM entities WHERE name = 'US Operations'), true, 1),
((SELECT id FROM acme), 'Europe HQ', 'EU-HQ', 'REGION', (SELECT uuid FROM entities WHERE name = 'ACME Corporation'), false, 1);

-- Globex Entities
WITH globex AS (SELECT id FROM tenants WHERE slug = 'globex')
INSERT INTO entities (tenant_id, name, code, type, parent_id, accrual_method, fy_start_month) VALUES
((SELECT id FROM globex), 'Globex Corp', 'GLOBEX', 'COMPANY', NULL, false, 4),
((SELECT id FROM globex), 'Asia Division', 'ASIA', 'REGION', (SELECT uuid FROM entities WHERE name = 'Globex Corp'), false, 4),
((SELECT id FROM globex), 'Manufacturing', 'MFG', 'DEPARTMENT', (SELECT uuid FROM entities WHERE name = 'Asia Division'), false, 4);

-- Stark Entities
WITH stark AS (SELECT id FROM tenants WHERE slug = 'stark-ind')
INSERT INTO entities (tenant_id, name, code, type, parent_id, accrual_method, fy_start_month) VALUES
((SELECT id FROM stark), 'Stark Industries', 'STARK', 'COMPANY', NULL, true, 10),
((SELECT id FROM stark), 'R&D Division', 'RND', 'DEPARTMENT', (SELECT uuid FROM entities WHERE name = 'Stark Industries'), true, 10),
((SELECT id FROM stark), 'Advanced Weapons', 'WEAPONS', 'COST_CENTER', (SELECT uuid FROM entities WHERE name = 'R&D Division'), true, 10);

-- =====================================================================
-- PERSON AND EMPLOYEE DATA
-- =====================================================================
-- ACME Employees
WITH acme AS (SELECT id FROM tenants WHERE slug = 'acme-corp'),
     us_ops AS (SELECT uuid FROM entities WHERE name = 'US Operations'),
     sales_dept AS (SELECT uuid FROM entities WHERE name = 'Sales Department'),
     eng_dept AS (SELECT uuid FROM entities WHERE name = 'Engineering'),
     eu_hq AS (SELECT uuid FROM entities WHERE name = 'Europe HQ')
INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes) VALUES
((SELECT id FROM acme), (SELECT uuid FROM us_ops), 'EMPLOYEE', 'John', 'Smith', 'john.smith@acme.com', '+1-555-1234', '{"clearance": "high", "department": "sales"}'),
((SELECT id FROM acme), (SELECT uuid FROM sales_dept), 'EMPLOYEE', 'Sarah', 'Johnson', 'sarah.j@acme.com', '+1-555-5678', '{"clearance": "medium", "region": "us"}'),
((SELECT id FROM acme), (SELECT uuid FROM eng_dept), 'EMPLOYEE', 'Michael', 'Chen', 'm.chen@acme.com', '+1-555-9012', '{"clearance": "high", "specialization": "ai"}'),
((SELECT id FROM acme), (SELECT uuid FROM eu_hq), 'EMPLOYEE', 'Emma', 'Dubois', 'e.dubois@acme.eu', '+33-1-2345', '{"clearance": "medium", "language": "fr"}'),
((SELECT id FROM acme), NULL, 'CUSTOMER', 'Robert', 'Brown', 'robert@client.com', '+44-20-1234', '{"tier": "gold", "industry": "finance"}');

-- Globex Employees
WITH globex AS (SELECT id FROM tenants WHERE slug = 'globex'),
     asia_div AS (SELECT uuid FROM entities WHERE name = 'Asia Division'),
     mfg_dept AS (SELECT uuid FROM entities WHERE name = 'Manufacturing')
INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes) VALUES
((SELECT id FROM globex), (SELECT uuid FROM asia_div), 'EMPLOYEE', 'Kenji', 'Tanaka', 'kenji.t@globex.jp', '+81-3-4567', '{"clearance": "low", "plant": "tokyo"}'),
((SELECT id FROM globex), (SELECT uuid FROM mfg_dept), 'EMPLOYEE', 'Wei', 'Zhang', 'wei.zhang@globex.cn', '+86-10-8910', '{"clearance": "medium", "shift": "night"}');

-- Stark Employees
WITH stark AS (SELECT id FROM tenants WHERE slug = 'stark-ind'),
     rnd_div AS (SELECT uuid FROM entities WHERE name = 'R&D Division'),
     weapons_cc AS (SELECT uuid FROM entities WHERE name = 'Advanced Weapons')
INSERT INTO persons (tenant_id, entity_id, person_type, first_name, last_name, email, phone, security_attributes) VALUES
((SELECT id FROM stark), (SELECT uuid FROM rnd_div), 'EMPLOYEE', 'Tony', 'Stark', 'tony@stark.com', '+1-212-1111', '{"clearance": "top", "projects": ["ironman"]}'),
((SELECT id FROM stark), (SELECT uuid FROM weapons_cc), 'EMPLOYEE', 'Pepper', 'Potts', 'pepper@stark.com', '+1-212-2222', '{"clearance": "high", "role": "executive"}');

-- Create employee records
INSERT INTO employees (tenant_id, person_id, employee_number, entity_id, position_title, hire_date, security_level, access_attributes) 
SELECT 
    p.tenant_id, 
    p.id,
    'EMP-' || LPAD(ROW_NUMBER() OVER (PARTITION BY p.tenant_id)::text, 5, '0'),
    p.entity_id,
    CASE 
        WHEN p.first_name = 'Tony' THEN 'Chief Technology Officer'
        WHEN p.first_name = 'Pepper' THEN 'Chief Operating Officer'
        WHEN p.first_name = 'John' THEN 'Sales Director'
        ELSE INITCAP(p.person_type) || ' Specialist'
    END,
    NOW() - INTERVAL '1 year' * RANDOM() * 5,
    CASE 
        WHEN p.security_attributes->>'clearance' = 'top' THEN 10
        WHEN p.security_attributes->>'clearance' = 'high' THEN 8
        WHEN p.security_attributes->>'clearance' = 'medium' THEN 5
        ELSE 3
    END,
    p.security_attributes
FROM persons p
WHERE p.person_type = 'EMPLOYEE';

-- =====================================================================
-- USER ACCOUNTS
-- =====================================================================
INSERT INTO users (tenant_id, entity_id, person_id, employee_id, username, email, password_hash, user_type, account_status, session_timeout_minutes, mfa_enabled) 
SELECT 
    p.tenant_id,
    p.entity_id,
    p.id,
    e.id,
    LOWER(SPLIT_PART(p.email, '@', 1)),
    p.email,
    crypt('Password123!', gen_salt('bf')),
    CASE 
        WHEN p.person_type = 'EMPLOYEE' THEN 'INTERNAL'
        ELSE 'CUSTOMER'
    END,
    'ACTIVE',
    480,
    (p.security_attributes->>'clearance') IN ('high', 'top')
FROM persons p
LEFT JOIN employees e ON p.id = e.person_id;

-- Create service accounts
WITH tenants AS (SELECT id FROM tenants)
INSERT INTO users (tenant_id, entity_id, username, email, password_hash, user_type, account_status) 
SELECT 
    id,
    NULL,
    'api-service',
    'api@company.com',
    crypt('ServiceAccountSecure!', gen_salt('bf')),
    'API',
    'ACTIVE'
FROM tenants;

-- =====================================================================
-- MODULES AND RESOURCES
-- =====================================================================
-- Create modules for each tenant
INSERT INTO modules (tenant_id, name, display_name, category, is_active)
SELECT 
    id,
    mods.name,
    mods.display_name,
    mods.category,
    true
FROM tenants
CROSS JOIN (
    VALUES 
    ('HR', 'Human Resources', 'CORE'),
    ('FINANCE', 'Financial Systems', 'FINANCE'),
    ('INVENTORY', 'Inventory Management', 'OPERATIONS'),
    ('SALES', 'Sales CRM', 'SALES')
) AS mods(name, display_name, category);

-- Create resources
INSERT INTO resources (tenant_id, module_id, entity_id, name, display_name, resource_type, path) 
SELECT 
    m.tenant_id,
    m.id,
    e.uuid,
    res.name,
    res.display_name,
    res.resource_type,
    res.path
FROM modules m
JOIN entities e ON e.tenant_id = m.tenant_id
CROSS JOIN (
    VALUES 
    ('employees', 'Employee Records', 'DATA', '/api/employees'),
    ('payroll', 'Payroll System', 'API', '/api/payroll'),
    ('inventory', 'Inventory Database', 'DATA', '/db/inventory'),
    ('sales-orders', 'Sales Orders', 'API', '/api/orders'),
    ('weapons-db', 'Weapons Database', 'DATA', '/db/weapons'),
    ('financial-reports', 'Financial Reports', 'REPORT', '/reports/financial')
) AS res(name, display_name, resource_type, path);

-- =====================================================================
-- ACTIONS AND PERMISSIONS
-- =====================================================================
-- Standard actions
INSERT INTO actions (tenant_id, name, display_name, action_type, risk_level, requires_approval)
SELECT 
    id,
    acts.name,
    acts.display_name,
    acts.action_type,
    acts.risk_level,
    acts.requires_approval
FROM tenants
CROSS JOIN (
    VALUES 
    ('CREATE', 'Create Records', 'CREATE', 'MEDIUM', false),
    ('READ', 'View Records', 'READ', 'LOW', false),
    ('UPDATE', 'Modify Records', 'UPDATE', 'HIGH', true),
    ('DELETE', 'Delete Records', 'DELETE', 'CRITICAL', true),
    ('APPROVE', 'Approve Actions', 'APPROVE', 'HIGH', false)
) AS acts(name, display_name, action_type, risk_level, requires_approval);

-- Create permissions
INSERT INTO permissions (tenant_id, resource_id, action_id, name, effect, conditions)
SELECT 
    r.tenant_id,
    r.id,
    a.id,
    CONCAT(r.name, ':', a.name),
    CASE 
        WHEN r.name = 'weapons-db' AND a.name != 'READ' THEN 'DENY'
        ELSE 'ALLOW'
    END,
    CASE 
        WHEN r.name = 'payroll' THEN '{"require_mfa": true}'::jsonb
        ELSE '{}'::jsonb
    END
FROM resources r
JOIN actions a ON a.tenant_id = r.tenant_id
WHERE a.name IN ('CREATE', 'READ', 'UPDATE', 'DELETE');

-- =====================================================================
-- ROLES AND ASSIGNMENTS
-- =====================================================================
-- Create standard roles
INSERT INTO roles (tenant_id, entity_id, name, display_name, role_type, permissions) 
SELECT 
    id,
    (SELECT uuid FROM entities WHERE name = 'ACME Corporation' AND tenant_id = t.id),
    roles.name,
    roles.display_name,
    roles.role_type,
    '{}'::jsonb
FROM tenants t
CROSS JOIN (
    VALUES 
    ('admin', 'Administrator', 'SYSTEM'),
    ('manager', 'Department Manager', 'ENTITY'),
    ('employee', 'Regular Employee', 'TENANT'),
    ('finance', 'Finance Specialist', 'FUNCTIONAL')
) AS roles(name, display_name, role_type);

-- Assign permissions to roles
INSERT INTO role_permissions (tenant_id, role_id, permission_id)
SELECT 
    r.tenant_id,
    r.id,
    p.id
FROM roles r
JOIN permissions p ON p.tenant_id = r.tenant_id
WHERE 
    (r.name = 'admin' AND p.effect = 'ALLOW') OR
    (r.name = 'manager' AND p.action_id IN (SELECT id FROM actions WHERE name IN ('READ', 'APPROVE'))) OR
    (r.name = 'employee' AND p.action_id IN (SELECT id FROM actions WHERE name = 'READ'));

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id, entity_id, assignment_type)
SELECT 
    u.id,
    r.id,
    u.entity_id,
    'DIRECT'
FROM users u
JOIN roles r ON r.tenant_id = u.tenant_id
WHERE 
    (u.username = 'tony' AND r.name = 'admin') OR
    (u.username LIKE 'sarah%' AND r.name = 'manager') OR
    (r.name = 'employee');

-- Special permission for Tony Stark
INSERT INTO user_permissions (tenant_id, user_id, permission_id, effect, reason)
SELECT
    u.tenant_id,
    u.id,
    p.id,
    'ALLOW',
    'CEO override'
FROM users u
JOIN permissions p ON p.tenant_id = u.tenant_id
WHERE u.username = 'tony' AND p.name = 'weapons-db:UPDATE';

-- =====================================================================
-- ABAC COMPONENTS
-- =====================================================================
-- Attribute definitions
INSERT INTO attribute_definitions (tenant_id, name, display_name, data_type, category, is_sensitive)
SELECT 
    id,
    attrs.name,
    attrs.display_name,
    attrs.data_type,
    attrs.category,
    attrs.is_sensitive
FROM tenants
CROSS JOIN (
    VALUES 
    ('security_clearance', 'Security Clearance', 'NUMBER', 'USER', true),
    ('department', 'Department', 'STRING', 'USER', false),
    ('resource_sensitivity', 'Resource Sensitivity', 'STRING', 'RESOURCE', true),
    ('time_of_day', 'Time of Access', 'TIME', 'ENVIRONMENT', false)
) AS attrs(name, display_name, data_type, category, is_sensitive);

-- ABAC Policies
INSERT INTO policies (tenant_id, name, policy_type, effect, priority, target, rule)
SELECT
    id,
    'HighSecurityAccess',
    'ABAC',
    'ALLOW',
    100,
    '{"resource.attributes.sensitivity": "high"}',
    '{"user.attributes.clearance": {"$gte": 8}, "environment.time_of_day": {"$between": ["09:00", "17:00"]}}'
FROM tenants;

-- =====================================================================
-- AUDIT AND SESSION DATA
-- =====================================================================
-- Generate user sessions
INSERT INTO user_sessions (tenant_id, user_id, session_token, ip_address, user_agent, expires_at)
SELECT
    u.tenant_id,
    u.id,
    gen_random_uuid()::text,
    ('192.168.' || FLOOR(RANDOM()*255) || '.' || FLOOR(RANDOM()*255))::inet,
    CASE FLOOR(RANDOM()*3)
        WHEN 0 THEN 'Chrome/Windows'
        WHEN 1 THEN 'Safari/Mac'
        ELSE 'Firefox/Linux'
    END,
    NOW() + INTERVAL '1 day'
FROM users u;

-- Generate audit logs
INSERT INTO audit_log (tenant_id, event_type, event_category, user_id, resource_id, action_id, decision, risk_score)
SELECT
    u.tenant_id,
    CASE 
        WHEN a.name = 'DELETE' THEN 'DELETE_ATTEMPT'
        ELSE a.name || '_ACCESS'
    END,
    CASE 
        WHEN r.name LIKE '%payroll%' THEN 'SENSITIVE'
        WHEN r.name LIKE '%weapons%' THEN 'CRITICAL'
        ELSE 'STANDARD'
    END,
    u.id,
    r.id,
    a.id,
    CASE 
        WHEN u.username = 'john.smith' AND a.name = 'DELETE' THEN 'DENY'
        ELSE 'ALLOW'
    END,
    CASE 
        WHEN a.name = 'DELETE' THEN 80
        WHEN r.name LIKE '%weapons%' THEN 70
        ELSE 30
    END
FROM users u
CROSS JOIN resources r
CROSS JOIN actions a
WHERE RANDOM() < 0.3;  -- 30% of possible combinations

-- =====================================================================
-- ACCESS REQUESTS
-- =====================================================================
INSERT INTO access_requests (tenant_id, requester_id, target_user_id, entity_id, request_type, role_id, justification, business_reason)
SELECT
    u.tenant_id,
    u.id,
    u.id,
    u.entity_id,
    'ROLE_ASSIGNMENT',
    r.id,
    'Need access for new project responsibilities',
    'Project Alpha requires additional privileges'
FROM users u
JOIN roles r ON r.tenant_id = u.tenant_id
WHERE r.name = 'manager' AND RANDOM() < 0.4;

-- Approve some requests
UPDATE access_requests 
SET 
    approval_status = 'APPROVED',
    approved_by = (SELECT id FROM users WHERE username = 'tony' AND tenant_id = access_requests.tenant_id),
    approved_at = NOW() - INTERVAL '1 day'
WHERE RANDOM() < 0.6;

-- =====================================================================
-- ENTITY STATE AND HIERARCHY
-- =====================================================================
-- Entity state tracking
INSERT INTO entitystate (uuid, fiscal_year, key, sequence, entity_id)
SELECT
    gen_random_uuid(),
    EXTRACT(YEAR FROM NOW()),
    docs.key,
    1000,
    e.uuid
FROM entities e
CROSS JOIN (VALUES ('INV'), ('PO'), ('SO')) AS docs(key);

-- Hierarchy paths (closure table)
INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
SELECT
    e.tenant_id,
    anc.uuid,
    desc.uuid,
    CASE 
        WHEN anc.name = 'ACME Corporation' AND desc.name = 'US Operations' THEN 1
        WHEN anc.name = 'US Operations' AND desc.name = 'Sales Department' THEN 1
        WHEN anc.name = 'ACME Corporation' AND desc.name = 'Sales Department' THEN 2
        ELSE 0
    END
FROM entities desc
JOIN entities anc ON desc.parent_id = anc.uuid;

-- Re-enable triggers
SET session_replication_role = DEFAULT;

-- =====================================================================
-- VERIFICATION QUERIES
-- =====================================================================
-- Check data counts
SELECT 'Tenants' AS table, COUNT(*) FROM tenants
UNION ALL SELECT 'Entities', COUNT(*) FROM entities
UNION ALL SELECT 'Persons', COUNT(*) FROM persons
UNION ALL SELECT 'Employees', COUNT(*) FROM employees
UNION ALL SELECT 'Users', COUNT(*) FROM users
UNION ALL SELECT 'Modules', COUNT(*) FROM modules
UNION ALL SELECT 'Resources', COUNT(*) FROM resources
UNION ALL SELECT 'Permissions', COUNT(*) FROM permissions
UNION ALL SELECT 'User Roles', COUNT(*) FROM user_roles
UNION ALL SELECT 'Audit Logs', COUNT(*) FROM audit_log;

-- Show sample user with permissions
SELECT u.username, r.name AS role, p.name AS permission, res.name AS resource
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
JOIN resources res ON p.resource_id = res.id
WHERE u.username = 'tony';