-- ------------------------------------------------------------------------------------------------
-- Comprehensive user view with all related data
-- ------------------------------------------------------------------------------------------------
CREATE VIEW user_complete_view AS
SELECT
    u.id as user_id,
    u.tenant_id,
    u.entity_id,
    u.username,
    u.email,
    u.user_type,
    u.account_status,
    u.is_active as user_active,
    u.last_login_at,
    u.mfa_enabled,
    p.id as person_id,
    p.first_name,
    p.last_name,
    p.middle_name,
    p.person_type,
    e.id as employee_id,
    e.employee_number,
    e.position_title,
    e.department_id,
    e.employment_status,
    e.security_level,
    -- Combine all attributes for ABAC evaluation
    COALESCE(p.security_attributes, '{}'::jsonb) ||
    COALESCE(e.access_attributes, '{}'::jsonb) ||
    COALESCE(u.user_attributes, '{}'::jsonb) as combined_attributes,
    -- Aggregate role information - FIXED
    array_agg(DISTINCT r.name ORDER BY r.name) as role_names,
    array_agg(DISTINCT r.id ORDER BY r.id) as role_ids,  -- Changed to ORDER BY r.id
    count(DISTINCT ur.id) FILTER (WHERE ur.is_active = true) as active_role_count
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN roles r ON ur.role_id = r.id AND r.is_active = true AND r.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.tenant_id, u.entity_id, u.username, u.email, u.user_type,
         u.account_status, u.is_active, u.last_login_at, u.mfa_enabled,
         p.id, p.first_name, p.last_name, p.middle_name, p.person_type,
         e.id, e.employee_number, e.position_title, e.department_id,
         e.employment_status, e.security_level,
         p.security_attributes, e.access_attributes, u.user_attributes;

COMMENT ON VIEW user_complete_view IS
'Comprehensive view combining user, person, and employee data with role aggregations and combined ABAC attributes for authorization decisions.';

-- ------------------------------------------------------------------------------------------------
-- Role permissions summary view
-- ------------------------------------------------------------------------------------------------
CREATE VIEW role_permissions_summary AS
SELECT
    r.tenant_id,
    r.id as role_id,
    r.name as role_name,
    r.display_name as role_display_name,
    r.role_type,
    r.level as hierarchy_level,
    r.module_id,
    m.name as module_name,
    array_agg(DISTINCT res.name ORDER BY res.name) FILTER (WHERE res.name IS NOT NULL) as resource_names,
    array_agg(DISTINCT a.name ORDER BY a.name) FILTER (WHERE a.name IS NOT NULL) as action_names,
    count(DISTINCT rp.permission_id) as permission_count,
    count(DISTINCT ur.user_id) FILTER (WHERE ur.is_active = true) as assigned_user_count
FROM roles r
LEFT JOIN modules m ON r.module_id = m.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND rp.is_active = true
LEFT JOIN permissions p ON rp.permission_id = p.id AND p.is_active = true
LEFT JOIN resources res ON p.resource_id = res.id AND res.is_active = true
LEFT JOIN actions a ON p.action_id = a.id AND a.is_active = true
LEFT JOIN user_roles ur ON r.id = ur.role_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
WHERE r.is_active = true AND r.deleted_at IS NULL
GROUP BY r.tenant_id, r.id, r.name, r.display_name, r.role_type, r.level, r.module_id, m.name;

COMMENT ON VIEW role_permissions_summary IS
'Summary view of roles with their permissions, resources, actions, and user assignment counts for role management and analysis.';

-- ------------------------------------------------------------------------------------------------
-- Audit summary view for security monitoring
-- ------------------------------------------------------------------------------------------------
CREATE VIEW audit_summary_view AS
SELECT
    tenant_id,
    event_category,
    severity,
    DATE_TRUNC('hour', created_at) as hour_bucket,
    count(*) as event_count,
    count(DISTINCT user_id) as unique_users,
    avg(risk_score) as avg_risk_score,
    max(risk_score) as max_risk_score,
    count(*) FILTER (WHERE decision = 'DENY') as denied_attempts,
    count(*) FILTER (WHERE decision = 'ALLOW') as allowed_attempts
FROM audit_log
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY tenant_id, event_category, severity, DATE_TRUNC('hour', created_at)
ORDER BY hour_bucket DESC, event_count DESC;

COMMENT ON VIEW audit_summary_view IS
'Hourly audit event summary for the last 7 days with risk metrics and access decision counts for security monitoring dashboards.';
