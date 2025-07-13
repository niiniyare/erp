
-- ================================================================================================
-- PERSONS QUERIES
-- ================================================================================================

-- name: CreatePerson :one
INSERT INTO persons (
    tenant_id, entity_id, person_type, first_name, last_name, middle_name,
    email, phone, birth_date, national_id, tax_id, address, 
    security_attributes, metadata
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetPersonByID :one
SELECT * FROM persons 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetPersonByEmail :one
SELECT * FROM persons 
WHERE email = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: ListPersons :many
SELECT * FROM persons 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
  AND ($1::varchar IS NULL OR person_type = $1)
ORDER BY last_name, first_name
LIMIT $2 OFFSET $3;

-- name: UpdatePerson :one
UPDATE persons 
SET first_name = $2, last_name = $3, middle_name = $4, email = $5, 
    phone = $6, address = $7, security_attributes = $8, metadata = $9,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePerson :exec
UPDATE persons 
SET deleted_at = NOW() 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: SearchPersons :many
SELECT * FROM persons 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
  AND (
    first_name ILIKE '%' || $1 || '%' OR
    last_name ILIKE '%' || $1 || '%' OR
    email ILIKE '%' || $1 || '%'
  )
ORDER BY last_name, first_name
LIMIT $2 OFFSET $3;

-- ================================================================================================
-- EMPLOYEES QUERIES
-- ================================================================================================

-- name: CreateEmployee :one
INSERT INTO employees (
    tenant_id, person_id, employee_number, entity_id, position_title,
    department_id, manager_id, hire_date, salary_info, employment_status,
    work_schedule, security_level, access_attributes
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetEmployeeByID :one
SELECT * FROM employees 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetEmployeeByNumber :one
SELECT * FROM employees 
WHERE employee_number = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetEmployeeByPersonID :one
SELECT * FROM employees 
WHERE person_id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: ListEmployeesByDepartment :many
SELECT * FROM employees 
WHERE department_id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
ORDER BY position_title, hire_date;

-- name: GetEmployeeHierarchy :many
WITH RECURSIVE emp_hierarchy AS (
    SELECT e.id, e.person_id, e.employee_number, e.position_title, e.manager_id, 0 as level
    FROM employees e
    WHERE e.id = $1 AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
    
    UNION ALL
    
    SELECT e.id, e.person_id, e.employee_number, e.position_title, e.manager_id, eh.level + 1
    FROM employees e
    INNER JOIN emp_hierarchy eh ON e.manager_id = eh.id
    WHERE e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL AND eh.level < 10
)
SELECT * FROM emp_hierarchy ORDER BY level, position_title;

-- name: UpdateEmployee :one
UPDATE employees 
SET position_title = $2, department_id = $3, manager_id = $4, 
    employment_status = $5, security_level = $6, access_attributes = $7,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: TerminateEmployee :exec
UPDATE employees 
SET employment_status = 'TERMINATED', termination_date = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- USERS QUERIES
-- ================================================================================================

-- name: CreateUser :one
INSERT INTO users (
    tenant_id, entity_id, person_id, employee_id, username, email, 
    password_hash, user_type, account_status, session_timeout_minutes,
    mfa_enabled, user_attributes, settings
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT * FROM users 
WHERE username = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetUserComplete :one
SELECT * FROM user_complete_view 
WHERE user_id = $1 AND tenant_id = current_tenant_id();

-- name: ListUsers :many
SELECT * FROM users 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
  AND ($1::varchar IS NULL OR user_type = $1)
  AND ($2::varchar IS NULL OR account_status = $2)
ORDER BY email
LIMIT $3 OFFSET $4;

-- name: UpdateUser :one
UPDATE users 
SET username = $2, email = $3, user_type = $4, account_status = $5,
    session_timeout_minutes = $6, mfa_enabled = $7, user_attributes = $8,
    settings = $9, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $2, password_changed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: UpdateUserLastLogin :exec
UPDATE users 
SET last_login_at = NOW(), failed_login_attempts = 0, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: IncrementFailedLogins :exec
UPDATE users 
SET failed_login_attempts = failed_login_attempts + 1,
    lockout_until = CASE 
        WHEN failed_login_attempts >= 4 THEN NOW() + INTERVAL '30 minutes'
        ELSE lockout_until 
    END,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: UnlockUser :exec
UPDATE users 
SET failed_login_attempts = 0, lockout_until = NULL, account_status = 'ACTIVE',
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: SoftDeleteUser :exec
UPDATE users 
SET deleted_at = NOW(), account_status = 'INACTIVE'
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: SearchUsers :many
SELECT * FROM user_complete_view
WHERE tenant_id = current_tenant_id()
  AND (
    user_complete_view.first_name ILIKE '%' || $1 || '%' OR
    user_complete_view.last_name ILIKE '%' || $1 || '%' OR
    user_complete_view.email ILIKE '%' || $1 || '%' OR
    user_complete_view.username ILIKE '%' || $1 || '%'
  )
ORDER BY user_complete_view.last_name, user_complete_view.first_name
LIMIT $2 OFFSET $3;

-- ================================================================================================
-- USER SESSIONS QUERIES
-- ================================================================================================

-- name: CreateSession :one
INSERT INTO user_sessions (
    tenant_id, user_id, session_token, refresh_token, ip_address,
    user_agent, device_info, location_info, expires_at
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM user_sessions 
WHERE session_token = $1 AND tenant_id = current_tenant_id() 
  AND is_active = true AND expires_at > NOW();

-- name: GetSessionByRefreshToken :one
SELECT * FROM user_sessions 
WHERE refresh_token = $1 AND tenant_id = current_tenant_id()
  AND is_active = true AND expires_at > NOW();

-- name: GetUserSessions :many
SELECT * FROM user_sessions 
WHERE user_id = $1 AND tenant_id = current_tenant_id() AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateSessionAccess :exec
UPDATE user_sessions 
SET last_accessed_at = NOW() 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: RefreshSession :one
UPDATE user_sessions 
SET refresh_token = $2, expires_at = $3, last_accessed_at = NOW()
WHERE session_token = $1 AND tenant_id = current_tenant_id() AND is_active = true
RETURNING *;

-- name: InvalidateSession :exec
UPDATE user_sessions 
SET is_active = false 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: InvalidateUserSessions :exec
UPDATE user_sessions 
SET is_active = false 
WHERE user_id = $1 AND tenant_id = current_tenant_id();

-- name: CleanupExpiredUserSessions :exec
DELETE FROM user_sessions 
WHERE tenant_id = current_tenant_id() AND expires_at < NOW();

-- ================================================================================================
-- MODULES QUERIES
-- ================================================================================================

-- name: CreateModule :one
INSERT INTO modules (
    tenant_id, name, display_name, description, category, version
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetModuleByID :one
SELECT * FROM modules 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetModuleByName :one
SELECT * FROM modules 
WHERE name = $1 AND tenant_id = current_tenant_id();

-- name: ListModules :many
SELECT * FROM modules 
WHERE tenant_id = current_tenant_id() AND is_active = true
  AND ($1::varchar IS NULL OR category = $1)
ORDER BY category, name;

-- name: UpdateModule :one
UPDATE modules 
SET display_name = $2, description = $3, category = $4, version = $5
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeactivateModule :exec
UPDATE modules 
SET is_active = false 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- RESOURCES QUERIES
-- ================================================================================================

-- name: CreateResource :one
INSERT INTO resources (
    tenant_id, module_id, entity_id, name, display_name, description,
    resource_type, parent_resource_id, path, resource_attributes
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetResourceByID :one
SELECT * FROM resources 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetResourceByName :one
SELECT * FROM resources 
WHERE name = $1 AND module_id = $2 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: ListResourcesByModule :many
SELECT * FROM resources 
WHERE module_id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
  AND ($2::varchar IS NULL OR resource_type = $2)
ORDER BY name;

-- name: GetResourceHierarchy :many
WITH RECURSIVE resource_tree AS (
    SELECT r.id, r.name, r.display_name, r.parent_resource_id, 0 as level
    FROM resources r
    WHERE r.id = $1 AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
    
    UNION ALL
    
    SELECT r.id, r.name, r.display_name, r.parent_resource_id, rt.level + 1
    FROM resources r
    INNER JOIN resource_tree rt ON r.parent_resource_id = rt.id
    WHERE r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL AND rt.level < 10
)
SELECT * FROM resource_tree ORDER BY level, name;

-- name: UpdateResource :one
UPDATE resources 
SET display_name = $2, description = $3, path = $4, resource_attributes = $5
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteResource :exec
UPDATE resources 
SET deleted_at = NOW() 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- ACTIONS QUERIES
-- ================================================================================================

-- name: CreateAction :one
INSERT INTO actions (
    tenant_id, name, display_name, description, action_type,
    action_category, risk_level, requires_approval
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetActionByID :one
SELECT * FROM actions 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetActionByName :one
SELECT * FROM actions 
WHERE name = $1 AND tenant_id = current_tenant_id();

-- name: ListActions :many
SELECT * FROM actions 
WHERE tenant_id = current_tenant_id() AND is_active = true
  AND ($1::varchar IS NULL OR action_type = $1)
  AND ($2::varchar IS NULL OR action_category = $2)
ORDER BY action_category, name;

-- name: UpdateAction :one
UPDATE actions 
SET display_name = $2, description = $3, action_category = $4, 
    risk_level = $5, requires_approval = $6
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeactivateAction :exec
UPDATE actions 
SET is_active = false 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- ROLES QUERIES
-- ================================================================================================

-- name: CreateRole :one
INSERT INTO roles (
    tenant_id, entity_id, name, display_name, description, module_id,
    role_type, parent_role_id, entity_scope, conditions
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetRoleByName :one
SELECT * FROM roles 
WHERE name = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: ListRoles :many
SELECT * FROM roles 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
  AND ($1::varchar IS NULL OR role_type = $1)
  AND ($2::uuid IS NULL OR entity_id = $2)
  AND is_active = true
ORDER BY level, name;

-- name: GetRoleHierarchy :many
WITH RECURSIVE role_hierarchy AS (
    SELECT r.id, r.tenant_id, r.name, r.display_name, r.parent_role_id, r.level, 0 as depth
    FROM roles r
    WHERE r.id = $1 AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
    
    UNION ALL
    
    SELECT r.id, r.tenant_id, r.name, r.display_name, r.parent_role_id, r.level, rh.depth + 1
    FROM roles r
    INNER JOIN role_hierarchy rh ON r.parent_role_id = rh.id
    WHERE r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL AND rh.depth < 10
)
SELECT * FROM role_hierarchy ORDER BY depth, name;

-- name: GetRolesSummary :many
SELECT * FROM role_permissions_summary 
WHERE tenant_id = current_tenant_id()
  AND ($1::uuid IS NULL OR role_id = $1)
ORDER BY role_name;

-- name: UpdateRole :one
UPDATE roles 
SET display_name = $2, description = $3, entity_scope = $4, 
    conditions = $5, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteRole :exec
UPDATE roles 
SET deleted_at = NOW() 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- PERMISSIONS QUERIES
-- ================================================================================================

-- name: CreatePermission :one
INSERT INTO permissions (
    tenant_id, resource_id, action_id, name, display_name, description,
    effect, conditions, data_filters, field_restrictions
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPermissionByID :one
SELECT * FROM permissions 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetPermissionByResourceAction :one
SELECT * FROM permissions 
WHERE resource_id = $1 AND action_id = $2 AND name = $3 
  AND tenant_id = current_tenant_id();

-- name: ListPermissions :many
SELECT p.*, r.name as resource_name, a.name as action_name
FROM permissions p
JOIN resources r ON p.resource_id = r.id
JOIN actions a ON p.action_id = a.id
WHERE p.tenant_id = current_tenant_id() AND p.is_active = true
  AND ($1::uuid IS NULL OR p.resource_id = $1)
  AND ($2::uuid IS NULL OR p.action_id = $2)
ORDER BY r.name, a.name;

-- name: UpdatePermission :one
UPDATE permissions 
SET display_name = $2, description = $3, conditions = $4,
    data_filters = $5, field_restrictions = $6
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeactivatePermission :exec
UPDATE permissions 
SET is_active = false 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: CheckUserPermission :one
SELECT user_has_permission($1, $2, $3, current_tenant_id(), $4, $5) as has_permission;

-- ================================================================================================
-- ROLE PERMISSIONS QUERIES
-- ================================================================================================

-- name: GrantPermissionToRole :one
INSERT INTO role_permissions (
    tenant_id, role_id, permission_id, entity_scope, granted_by, conditions
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5
) ON CONFLICT (tenant_id, role_id, permission_id, entity_scope)
DO UPDATE SET 
    granted_by = EXCLUDED.granted_by,
    granted_at = NOW(),
    conditions = EXCLUDED.conditions,
    is_active = true
RETURNING *;

-- name: GetRolePermissions :many
SELECT p.*, r.name as resource_name, a.name as action_name, rp.entity_scope, rp.conditions
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
JOIN resources r ON p.resource_id = r.id
JOIN actions a ON p.action_id = a.id
WHERE rp.role_id = $1 AND rp.tenant_id = current_tenant_id() 
  AND rp.is_active = true AND p.is_active = true;

-- name: RevokePermissionFromRole :exec
UPDATE role_permissions 
SET is_active = false 
WHERE role_id = $1 AND permission_id = $2 AND tenant_id = current_tenant_id()
  AND ($3::uuid IS NULL OR entity_scope = $3);

-- name: ListRolePermissionsByResource :many
SELECT rp.*, p.name as permission_name, r.name as resource_name, a.name as action_name
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
JOIN resources r ON p.resource_id = r.id
JOIN actions a ON p.action_id = a.id
WHERE r.id = $1 AND rp.tenant_id = current_tenant_id() AND rp.is_active = true;

-- ================================================================================================
-- USER ROLES QUERIES
-- ================================================================================================

-- name: AssignRoleToUser :one
INSERT INTO user_roles (
    user_id, role_id, entity_id, assignment_type, delegated_by, 
    assigned_by, expires_at, conditions
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (user_id, role_id, entity_id) 
DO UPDATE SET 
    assignment_type = EXCLUDED.assignment_type,
    delegated_by = EXCLUDED.delegated_by,
    assigned_by = EXCLUDED.assigned_by,
    expires_at = EXCLUDED.expires_at,
    conditions = EXCLUDED.conditions,
    is_active = true
RETURNING *;

-- name: GetUserRoles :many
SELECT r.*, ur.assignment_type, ur.expires_at, ur.conditions, ur.assigned_at
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_id = $1 AND r.tenant_id = current_tenant_id()
  AND ur.is_active = true 
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
  AND r.is_active = true 
  AND r.deleted_at IS NULL;

-- name: GetRoleUsers :many
SELECT u.*, ur.assignment_type, ur.expires_at, ur.assigned_at
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
WHERE ur.role_id = $1 AND u.tenant_id = current_tenant_id()
  AND ur.is_active = true 
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
  AND u.deleted_at IS NULL;

-- name: RevokeUserRole :exec
UPDATE user_roles 
SET is_active = false 
WHERE user_id = $1 AND role_id = $2 AND entity_id = $3;

-- name: GetUserEffectivePermissions :many
SELECT DISTINCT p.*, r.name as resource_name, a.name as action_name, 
       'ROLE' as permission_source, ro.name as source_role
FROM user_roles ur
JOIN role_permissions rp ON ur.role_id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
JOIN resources r ON p.resource_id = r.id
JOIN actions a ON p.action_id = a.id
JOIN roles ro ON ur.role_id = ro.id
WHERE ur.user_id = $1 AND rp.tenant_id = current_tenant_id()
  AND ur.is_active = true 
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
  AND rp.is_active = true 
  AND p.is_active = true

UNION

SELECT DISTINCT p.*, r.name as resource_name, a.name as action_name,
       'DIRECT' as permission_source, NULL as source_role
FROM user_permissions up
JOIN permissions p ON up.permission_id = p.id
JOIN resources r ON p.resource_id = r.id
JOIN actions a ON p.action_id = a.id
WHERE up.user_id = $1 AND up.tenant_id = current_tenant_id()
  AND up.is_active = true 
  AND (up.expires_at IS NULL OR up.expires_at > NOW())
  AND up.effect = 'ALLOW'
  AND p.is_active = true

ORDER BY resource_name, action_name;

-- ================================================================================================
-- USER PERMISSIONS QUERIES
-- ================================================================================================

-- name: GrantDirectPermission :one
INSERT INTO user_permissions (
    tenant_id, user_id, permission_id, entity_id, effect, reason,
    granted_by, expires_at, conditions
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (tenant_id, user_id, permission_id, entity_id)
DO UPDATE SET 
    effect = EXCLUDED.effect,
    reason = EXCLUDED.reason,
    granted_by = EXCLUDED.granted_by,
    granted_at = NOW(),
    expires_at = EXCLUDED.expires_at,
    conditions = EXCLUDED.conditions,
    is_active = true
RETURNING *;

-- name: GetUserDirectPermissions :many
SELECT p.*, up.effect, up.reason, up.expires_at, up.granted_at,
       r.name as resource_name, a.name as action_name
FROM user_permissions up
JOIN permissions p ON up.permission_id = p.id
JOIN resources r ON p.resource_id = r.id
JOIN actions a ON p.action_id = a.id
WHERE up.user_id = $1 AND up.tenant_id = current_tenant_id()
  AND up.is_active = true 
  AND (up.expires_at IS NULL OR up.expires_at > NOW());

-- name: RevokeDirectPermission :exec
UPDATE user_permissions 
SET is_active = false 
WHERE user_id = $1 AND permission_id = $2 AND tenant_id = current_tenant_id()
  AND ($3::uuid IS NULL OR entity_id = $3);

-- name: GetPermissionUsers :many
SELECT u.*, up.effect, up.reason, up.expires_at, up.granted_at
FROM user_permissions up
JOIN users u ON up.user_id = u.id
WHERE up.permission_id = $1 AND up.tenant_id = current_tenant_id()
  AND up.is_active = true 
  AND (up.expires_at IS NULL OR up.expires_at > NOW())
  AND u.deleted_at IS NULL;

-- ================================================================================================
-- ATTRIBUTE DEFINITIONS QUERIES
-- ================================================================================================

-- name: CreateAttributeDefinition :one
INSERT INTO attribute_definitions (
    tenant_id, name, display_name, description, data_type, category,
    is_required, is_sensitive, default_value, allowed_values, 
    validation_rules, encryption_required
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetAttributeDefinitionByID :one
SELECT * FROM attribute_definitions 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetAttributeDefinitionByName :one
SELECT * FROM attribute_definitions 
WHERE name = $1 AND tenant_id = current_tenant_id();

-- name: ListAttributeDefinitions :many
SELECT * FROM attribute_definitions 
WHERE tenant_id = current_tenant_id() AND is_active = true
  AND ($1::varchar IS NULL OR category = $1)
ORDER BY category, name;

-- name: UpdateAttributeDefinition :one
UPDATE attribute_definitions 
SET display_name = $2, description = $3, default_value = $4,
    allowed_values = $5, validation_rules = $6, encryption_required = $7
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeactivateAttributeDefinition :exec
UPDATE attribute_definitions 
SET is_active = false 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- POLICIES QUERIES
-- ================================================================================================

-- name: CreatePolicy :one
INSERT INTO policies (
    tenant_id, entity_id, name, display_name, description, policy_type,
    effect, priority, category, target, rule, obligations, advice, created_by
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetPolicyByID :one
SELECT * FROM policies 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetPolicyByName :one
SELECT * FROM policies 
WHERE name = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: ListPolicies :many
SELECT * FROM policies 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
  AND ($1::varchar IS NULL OR policy_type = $1)
  AND ($2::varchar IS NULL OR category = $2)
  AND is_active = true
ORDER BY priority DESC, name;

-- name: GetActivePoliciesByPriority :many
SELECT * FROM policies 
WHERE tenant_id = current_tenant_id() AND is_active = true AND deleted_at IS NULL
ORDER BY priority DESC, created_at;

-- name: UpdatePolicy :one
UPDATE policies 
SET display_name = $2, description = $3, priority = $4, target = $5,
    rule = $6, obligations = $7, advice = $8, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePolicy :exec
UPDATE policies 
SET deleted_at = NOW() 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ActivatePolicy :exec
UPDATE policies 
SET is_active = true, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: DeactivatePolicy :exec
UPDATE policies 
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- ================================================================================================
-- POLICY EVALUATIONS QUERIES
-- ================================================================================================

-- name: CachePolicyEvaluation :one
INSERT INTO policy_evaluations (
    tenant_id, user_id, resource_id, action_id, context_hash,
    decision, applicable_policies, evaluation_time_ms, expires_at
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetCachedPolicyEvaluation :one
SELECT * FROM policy_evaluations
WHERE user_id = $1 AND resource_id = $2 AND action_id = $3 
  AND context_hash = $4 AND tenant_id = current_tenant_id()
  AND expires_at > NOW();

-- name: CleanupExpiredEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id() AND expires_at < NOW();

-- name: GetPolicyEvaluationStats :many
SELECT decision, COUNT(*) as count, AVG(evaluation_time_ms) as avg_time_ms
FROM policy_evaluations
WHERE tenant_id = current_tenant_id() 
  AND evaluated_at >= $1 AND evaluated_at <= $2
GROUP BY decision;

-- ================================================================================================
-- AUDIT LOG QUERIES
-- ================================================================================================

-- name: LogAuditEvent :one
INSERT INTO audit_log (
    tenant_id, event_type, event_category, severity, user_id, target_user_id,
    entity_id, resource_id, action_id, role_id, permission_id, decision,
    reason, risk_score, context, ip_address, user_agent, session_id, compliance_flags
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
) RETURNING *;

-- name: GetAuditEvents :many
SELECT * FROM audit_log
WHERE tenant_id = current_tenant_id()
  AND created_at >= $1 AND created_at <= $2
  AND ($3::varchar IS NULL OR event_category = $3)
  AND ($4::varchar IS NULL OR severity = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: GetUserAuditEvents :many
SELECT * FROM audit_log
WHERE user_id = $1 AND tenant_id = current_tenant_id()
  AND created_at >= $2 AND created_at <= $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetHighRiskEvents :many
SELECT * FROM audit_log
WHERE tenant_id = current_tenant_id() 
  AND risk_score >= $1
  AND created_at >= $2
ORDER BY risk_score DESC, created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetAuditSummary :many
SELECT * FROM audit_summary_view
WHERE tenant_id = current_tenant_id()
  AND ($1::timestamptz IS NULL OR hour_bucket >= $1);

-- name: GetFailedLoginAttempts :many
SELECT * FROM audit_log
WHERE event_type = 'LOGIN_FAILED' AND tenant_id = current_tenant_id()
  AND created_at >= $1
  AND ($2::uuid IS NULL OR user_id = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetResourceAccessEvents :many
SELECT * FROM audit_log
WHERE resource_id = $1 AND tenant_id = current_tenant_id()
  AND created_at >= $2
  AND event_category = 'ACCESS'
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- ================================================================================================
-- ACCESS REQUESTS QUERIES
-- ================================================================================================

-- name: CreateAccessRequest :one
INSERT INTO access_requests (
    tenant_id, requester_id, target_user_id, entity_id, request_type,
    role_id, permission_id, resource_id, justification, business_reason,
    duration_hours, expires_at, auto_revoke
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetAccessRequestByID :one
SELECT * FROM access_requests 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListAccessRequests :many
SELECT * FROM access_requests 
WHERE tenant_id = current_tenant_id()
  AND ($1::varchar IS NULL OR approval_status = $1)
  AND ($2::uuid IS NULL OR requester_id = $2)
  AND ($3::uuid IS NULL OR target_user_id = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetPendingAccessRequests :many
SELECT * FROM access_requests 
WHERE approval_status = 'PENDING' AND tenant_id = current_tenant_id()
ORDER BY created_at;

-- name: ApproveAccessRequest :one
UPDATE access_requests 
SET approval_status = 'APPROVED', approved_by = $2, approved_at = NOW(),
    approval_comments = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND approval_status = 'PENDING'
RETURNING *;

-- name: RejectAccessRequest :one
UPDATE access_requests 
SET approval_status = 'REJECTED', approved_by = $2, approved_at = NOW(),
    approval_comments = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND approval_status = 'PENDING'
RETURNING *;

-- name: ExpireAccessRequest :exec
UPDATE access_requests 
SET approval_status = 'EXPIRED', updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetExpiredAccessRequests :many
SELECT * FROM access_requests 
WHERE expires_at < NOW() AND approval_status = 'APPROVED' 
  AND auto_revoke = true AND tenant_id = current_tenant_id();

-- ================================================================================================
-- MAINTENANCE AND UTILITY QUERIES
-- ================================================================================================

-- name: CleanupExpiredData :exec
SELECT cleanup_expired_data(current_tenant_id());

-- name: GetUserSecuritySummary :one
SELECT 
    u.id as user_id, u.email, u.account_status, u.last_login_at,
    u.failed_login_attempts, u.mfa_enabled,
    COUNT(DISTINCT ur.role_id) FILTER (WHERE ur.is_active = true) as active_roles,
    COUNT(DISTINCT up.permission_id) FILTER (WHERE up.is_active = true AND up.effect = 'ALLOW') as direct_permissions,
    COUNT(DISTINCT up.permission_id) FILTER (WHERE up.is_active = true AND up.effect = 'DENY') as denied_permissions,
    COUNT(DISTINCT s.id) FILTER (WHERE s.is_active = true) as active_sessions,
    MAX(e.security_level) as max_security_level
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN user_permissions up ON u.id = up.user_id AND (up.expires_at IS NULL OR up.expires_at > NOW())
LEFT JOIN user_sessions s ON u.id = s.user_id
LEFT JOIN employees e ON u.employee_id = e.id
WHERE u.id = $1 AND u.tenant_id = current_tenant_id()
GROUP BY u.id, u.email, u.account_status, u.last_login_at, u.failed_login_attempts, u.mfa_enabled;

-- name: GetTenantStatistics :one
SELECT 
    COUNT(DISTINCT u.id) FILTER (WHERE u.deleted_at IS NULL) as total_users,
    COUNT(DISTINCT u.id) FILTER (WHERE u.account_status = 'ACTIVE' AND u.deleted_at IS NULL) as active_users,
    COUNT(DISTINCT r.id) FILTER (WHERE r.deleted_at IS NULL) as total_roles,
    COUNT(DISTINCT p.id) as total_permissions,
    COUNT(DISTINCT res.id) FILTER (WHERE res.deleted_at IS NULL) as total_resources,
    COUNT(DISTINCT s.id) FILTER (WHERE s.is_active = true AND s.expires_at > NOW()) as active_sessions,
    COUNT(DISTINCT pol.id) FILTER (WHERE pol.is_active = true AND pol.deleted_at IS NULL) as active_policies
FROM users u
CROSS JOIN roles r
CROSS JOIN permissions p  
CROSS JOIN resources res
CROSS JOIN user_sessions s
CROSS JOIN policies pol
WHERE u.tenant_id = current_tenant_id()
  AND r.tenant_id = current_tenant_id()
  AND p.tenant_id = current_tenant_id()
  AND res.tenant_id = current_tenant_id()
  AND s.tenant_id = current_tenant_id()
  AND pol.tenant_id = current_tenant_id();

-- name: CreateDefaultSystemData :exec
SELECT create_default_system_data(current_tenant_id());

-- name: GetOrphanedRecords :many
SELECT 'users' as table_name, id, 'person_id not found' as issue
FROM users 
WHERE person_id IS NOT NULL 
  AND person_id NOT IN (SELECT id FROM persons WHERE tenant_id = current_tenant_id())
  AND tenant_id = current_tenant_id()

UNION ALL

SELECT 'employees' as table_name, id, 'person_id not found' as issue  
FROM employees
WHERE person_id NOT IN (SELECT id FROM persons WHERE tenant_id = current_tenant_id())
  AND tenant_id = current_tenant_id()

UNION ALL

SELECT 'user_roles' as table_name, user_id, 'user not found' as issue
FROM user_roles ur
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ur.user_id AND u.tenant_id = current_tenant_id());

-- name: ValidateRoleHierarchy :many
WITH RECURSIVE role_cycles AS (
    SELECT id, parent_role_id, ARRAY[id] as path, 0 as depth
    FROM roles 
    WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
    
    UNION ALL
    
    SELECT r.id, r.parent_role_id, rc.path || r.id, rc.depth + 1
    FROM roles r
    INNER JOIN role_cycles rc ON r.id = rc.parent_role_id
    WHERE r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
      AND r.id = ANY(rc.path) AND rc.depth < 20
)
SELECT id, path, 'Circular reference detected' as issue
FROM role_cycles 
WHERE depth > 0;



-- ================================================================================================
--  RBAC QUERIES - Integrating Useful Patterns from Existing Queries
-- ================================================================================================

-- ================================================================================================
-- IMPROVED UPDATE PATTERNS (Using COALESCE for partial updates)
-- ================================================================================================

-- name: UpdateUserPartial :one
UPDATE users 
SET 
    entity_id = COALESCE($2, entity_id),
    person_id = COALESCE($3, person_id),
    employee_id = COALESCE($4, employee_id),
    username = COALESCE($5, username),
    email = COALESCE($6, email),
    password_hash = COALESCE($7, password_hash),
    user_type = COALESCE($8, user_type),
    account_status = COALESCE($9, account_status),
    session_timeout_minutes = COALESCE($10, session_timeout_minutes),
    mfa_enabled = COALESCE($11, mfa_enabled),
    user_attributes = COALESCE($12, user_attributes),
    settings = COALESCE($13, settings),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: UpdatePersonPartial :one
UPDATE persons 
SET 
    entity_id = COALESCE($2, entity_id),
    person_type = COALESCE($3, person_type),
    first_name = COALESCE($4, first_name),
    last_name = COALESCE($5, last_name),
    middle_name = COALESCE($6, middle_name),
    email = COALESCE($7, email),
    phone = COALESCE($8, phone),
    birth_date = COALESCE($9, birth_date),
    national_id = COALESCE($10, national_id),
    tax_id = COALESCE($11, tax_id),
    address = COALESCE($12, address),
    security_attributes = COALESCE($13, security_attributes),
    metadata = COALESCE($14, metadata),
    is_active = COALESCE($15, is_active),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEmployeePartial :one
UPDATE employees 
SET 
    employee_number = COALESCE($2, employee_number),
    entity_id = COALESCE($3, entity_id),
    position_title = COALESCE($4, position_title),
    department_id = COALESCE($5, department_id),
    manager_id = COALESCE($6, manager_id),
    hire_date = COALESCE($7, hire_date),
    termination_date = COALESCE($8, termination_date),
    salary_info = COALESCE($9, salary_info),
    employment_status = COALESCE($10, employment_status),
    work_schedule = COALESCE($11, work_schedule),
    security_level = COALESCE($12, security_level),
    access_attributes = COALESCE($13, access_attributes),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: UpdateRolePartial :one
UPDATE roles 
SET 
    entity_id = COALESCE($2, entity_id),
    name = COALESCE($3, name),
    display_name = COALESCE($4, display_name),
    description = COALESCE($5, description),
    module_id = COALESCE($6, module_id),
    role_type = COALESCE($7, role_type),
    parent_role_id = COALESCE($8, parent_role_id),
    permissions = COALESCE($9, permissions),
    entity_scope = COALESCE($10, entity_scope),
    conditions = COALESCE($11, conditions),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- ================================================================================================
--  COMPLEX JOINS AND AGGREGATIONS
-- ================================================================================================

-- name: GetCompleteUserProfile :one
SELECT 
    u.*,
    p.first_name, p.last_name, p.middle_name, p.email as person_email, 
    p.phone, p.birth_date, p.national_id, p.address, p.security_attributes as person_security_attributes,
    e.id as employee_id, e.employee_number, e.position_title, e.department_id,
    e.manager_id, e.hire_date, e.employment_status, e.security_level, e.access_attributes,
    (CASE 
        WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
        THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
        ELSE p.first_name || ' ' || p.last_name
    END)::text as full_name,
    -- Active roles count
    (SELECT COUNT(*) FROM user_roles ur 
     WHERE ur.user_id = u.id AND ur.is_active = true 
     AND (ur.expires_at IS NULL OR ur.expires_at > NOW())) as active_roles_count,
    -- Active sessions count  
    (SELECT COUNT(*) FROM user_sessions us 
     WHERE us.user_id = u.id AND us.is_active = true 
     AND us.expires_at > NOW()) as active_sessions_count
FROM users u
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
LEFT JOIN employees e ON u.employee_id = e.id AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
WHERE u.id = $1 AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL;

-- name: ListUsersWithDetails :many
SELECT 
    u.id, u.username, u.email, u.user_type, u.account_status, u.is_active, u.last_login_at,
    p.first_name, p.last_name, 
    (CASE 
        WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
        THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
        ELSE p.first_name || ' ' || p.last_name
    END)::text as full_name,
    e.employee_number, e.position_title, e.employment_status,
    string_agg(DISTINCT r.name, ', ' ORDER BY r.name) as roles,
    COUNT(DISTINCT ur.role_id) FILTER (WHERE ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())) as roles_count
FROM users u
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
LEFT JOIN employees e ON u.employee_id = e.id AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN roles r ON ur.role_id = r.id AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
WHERE u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
  AND ($1::varchar IS NULL OR u.user_type = $1)
  AND ($2::varchar IS NULL OR u.account_status = $2)
GROUP BY u.id, u.username, u.email, u.user_type, u.account_status, u.is_active, u.last_login_at,
         p.first_name, p.last_name, p.middle_name, e.employee_number, e.position_title, e.employment_status
ORDER BY p.last_name, p.first_name, u.email
LIMIT $3 OFFSET $4;

-- name: SearchUsersAdvanced :many
SELECT DISTINCT 
    u.id, u.username, u.email, u.user_type, u.account_status, u.is_active,
    p.first_name, p.last_name,
    (CASE 
        WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
        THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
        ELSE p.first_name || ' ' || p.last_name
    END)::text as full_name,
    e.employee_number, e.position_title,
    string_agg(DISTINCT r.name, ', ' ORDER BY r.name) as roles
FROM users u
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
LEFT JOIN employees e ON u.employee_id = e.id AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
LEFT JOIN user_roles ur ON u.id = ur.user_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN roles r ON ur.role_id = r.id AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
WHERE u.tenant_id = current_tenant_id() 
    AND u.deleted_at IS NULL
    AND (
        u.email ILIKE '%' || $1 || '%' OR
        u.username ILIKE '%' || $1 || '%' OR
        p.first_name ILIKE '%' || $1 || '%' OR
        p.last_name ILIKE '%' || $1 || '%' OR
        (p.first_name || ' ' || p.last_name) ILIKE '%' || $1 || '%' OR
        e.employee_number ILIKE '%' || $1 || '%' OR
        e.position_title ILIKE '%' || $1 || '%'
    )
GROUP BY u.id, u.username, u.email, u.user_type, u.account_status, u.is_active, 
         p.first_name, p.last_name, p.middle_name, e.employee_number, e.position_title
ORDER BY p.last_name, p.first_name, u.email
LIMIT $2 OFFSET $3;

-- ================================================================================================
--  ROLE AND PERMISSION QUERIES
-- ================================================================================================

-- name: GetRoleWithPermissionDetails :one
SELECT 
    r.*,
    m.name as module_name, m.display_name as module_display_name,
    COUNT(DISTINCT rp.permission_id) as permissions_count,
    COUNT(DISTINCT ur.user_id) FILTER (WHERE ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())) as users_count,
    array_agg(DISTINCT p.name ORDER BY p.name) FILTER (WHERE p.name IS NOT NULL) as permission_names
FROM roles r
LEFT JOIN modules m ON r.module_id = m.id AND m.tenant_id = current_tenant_id()
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND rp.tenant_id = current_tenant_id() AND rp.is_active = true
LEFT JOIN permissions p ON rp.permission_id = p.id AND p.tenant_id = current_tenant_id() AND p.is_active = true
LEFT JOIN user_roles ur ON r.id = ur.role_id AND ur.is_active = true AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
WHERE r.id = $1 AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
GROUP BY r.id, r.tenant_id, r.entity_id, r.name, r.display_name, r.description, r.module_id, 
         r.role_type, r.parent_role_id, r.level, r.permissions, r.entity_scope, r.conditions,
         r.is_system_role, r.is_active, r.created_at, r.updated_at, r.deleted_at,
         m.name, m.display_name;

-- name: GetUserRolesDetailed :many
SELECT 
    ur.*,
    r.name as role_name, r.display_name as role_display_name, r.description as role_description,
    r.role_type, r.level, r.is_system_role,
    m.name as module_name, m.display_name as module_display_name,
    CASE 
        WHEN ur.expires_at IS NOT NULL AND ur.expires_at <= NOW() THEN 'EXPIRED'
        WHEN ur.expires_at IS NOT NULL AND ur.expires_at > NOW() THEN 'TEMPORARY'
        ELSE 'PERMANENT'
    END as assignment_status
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
LEFT JOIN modules m ON r.module_id = m.id AND m.tenant_id = current_tenant_id()
WHERE ur.user_id = $1 AND ur.is_active = true
ORDER BY r.level DESC, r.name;

-- name: GetRoleUsersDetailed :many
SELECT 
    ur.*,
    u.username, u.email, u.user_type, u.account_status, u.is_active as user_active,
    p.first_name, p.last_name,
    (CASE 
        WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
        THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
        ELSE p.first_name || ' ' || p.last_name
    END)::text as full_name,
    e.employee_number, e.position_title, e.employment_status,
    CASE 
        WHEN ur.expires_at IS NOT NULL AND ur.expires_at <= NOW() THEN 'EXPIRED'
        WHEN ur.expires_at IS NOT NULL AND ur.expires_at > NOW() THEN 'TEMPORARY'
        ELSE 'PERMANENT'
    END as assignment_status
FROM user_roles ur
JOIN users u ON ur.user_id = u.id AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
LEFT JOIN employees e ON u.employee_id = e.id AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
WHERE ur.role_id = $1 AND ur.is_active = true
ORDER BY p.last_name, p.first_name, u.email;

-- ================================================================================================
--  STATISTICS AND REPORTING QUERIES
-- ================================================================================================

-- name: GetComprehensiveUserStats :one
SELECT 
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE is_active = true) as active_users,
    COUNT(*) FILTER (WHERE account_status = 'ACTIVE') as account_active_users,
    COUNT(*) FILTER (WHERE account_status = 'LOCKED') as locked_users,
    COUNT(*) FILTER (WHERE account_status = 'SUSPENDED') as suspended_users,
    COUNT(*) FILTER (WHERE user_type = 'INTERNAL') as internal_users,
    COUNT(*) FILTER (WHERE user_type = 'CUSTOMER') as customer_users,
    COUNT(*) FILTER (WHERE user_type = 'VENDOR') as vendor_users,
    COUNT(*) FILTER (WHERE user_type = 'ADMIN') as admin_users,
    COUNT(*) FILTER (WHERE mfa_enabled = true) as mfa_enabled_users,
    COUNT(*) FILTER (WHERE last_login_at >= NOW() - INTERVAL '24 hours') as last_24h_logins,
    COUNT(*) FILTER (WHERE last_login_at >= NOW() - INTERVAL '7 days') as last_7d_logins,
    COUNT(*) FILTER (WHERE last_login_at >= NOW() - INTERVAL '30 days') as last_30d_logins,
    COUNT(*) FILTER (WHERE failed_login_attempts > 0) as users_with_failed_logins,
    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '30 days') as new_users_30d
FROM users
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetRoleStatistics :one
SELECT 
    COUNT(*) as total_roles,
    COUNT(*) FILTER (WHERE is_system_role = true) as system_roles,
    COUNT(*) FILTER (WHERE is_system_role = false) as custom_roles,
    COUNT(*) FILTER (WHERE role_type = 'SYSTEM') as system_type_roles,
    COUNT(*) FILTER (WHERE role_type = 'FUNCTIONAL') as functional_roles,
    COUNT(*) FILTER (WHERE role_type = 'CUSTOM') as custom_type_roles,
    COUNT(*) FILTER (WHERE parent_role_id IS NOT NULL) as child_roles,
    COUNT(*) FILTER (WHERE parent_role_id IS NULL) as root_roles,
    COUNT(DISTINCT module_id) as modules_with_roles
FROM roles
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetEmployeeStatistics :one
SELECT 
    COUNT(*) as total_employees,
    COUNT(*) FILTER (WHERE employment_status = 'ACTIVE') as active_employees,
    COUNT(*) FILTER (WHERE employment_status = 'TERMINATED') as terminated_employees,
    COUNT(*) FILTER (WHERE employment_status = 'ON_LEAVE') as on_leave_employees,
    COUNT(*) FILTER (WHERE employment_status = 'SUSPENDED') as suspended_employees,
    COUNT(DISTINCT department_id) as departments_count,
    COUNT(*) FILTER (WHERE manager_id IS NULL) as top_level_employees,
    COUNT(*) FILTER (WHERE security_level >= 5) as high_security_employees,
    COUNT(*) FILTER (WHERE hire_date >= NOW() - INTERVAL '90 days') as new_hires_90d,
    AVG(security_level)::DECIMAL(3,2) as avg_security_level
FROM employees
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetSessionStatistics :one
SELECT 
    COUNT(*) as total_sessions,
    COUNT(*) FILTER (WHERE is_active = true AND expires_at > NOW()) as active_sessions,
    COUNT(*) FILTER (WHERE expires_at <= NOW()) as expired_sessions,
    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours') as sessions_24h,
    COUNT(DISTINCT user_id) as unique_users_with_sessions,
    COUNT(DISTINCT ip_address) as unique_ip_addresses,
    AVG(EXTRACT(EPOCH FROM (expires_at - created_at))/3600)::DECIMAL(5,2) as avg_session_duration_hours
FROM user_sessions
WHERE tenant_id = current_tenant_id();

-- ================================================================================================
--  HIERARCHY AND ORGANIZATIONAL QUERIES
-- ================================================================================================

-- name: GetCompleteEmployeeHierarchy :many
WITH RECURSIVE employee_hierarchy AS (
    -- Base case: Start with the specified employee
    SELECT 
        e.id, e.person_id, e.employee_number, e.position_title, e.manager_id,
        e.department_id, e.security_level, e.employment_status,
        p.first_name, p.last_name,
        (CASE 
            WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
            THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
            ELSE p.first_name || ' ' || p.last_name
        END)::text as full_name,
        0 as level,
        ARRAY[e.id] as path
    FROM employees e
    JOIN persons p ON e.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
    WHERE e.id = $1 AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
    
    UNION ALL
    
    -- Recursive case: Get subordinates
    SELECT 
        e.id, e.person_id, e.employee_number, e.position_title, e.manager_id,
        e.department_id, e.security_level, e.employment_status,
        p.first_name, p.last_name,
        (CASE 
            WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
            THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
            ELSE p.first_name || ' ' || p.last_name
        END)::text as full_name,
        eh.level + 1,
        eh.path || e.id
    FROM employees e
    JOIN persons p ON e.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
    JOIN employee_hierarchy eh ON e.manager_id = eh.id
    WHERE e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL 
      AND eh.level < 10 AND NOT (e.id = ANY(eh.path)) -- Prevent cycles
)
SELECT *, array_length(path, 1) as depth
FROM employee_hierarchy
ORDER BY level, last_name, first_name;

-- name: GetOrganizationalChart :many
SELECT 
    e.id, e.employee_number, e.position_title, e.manager_id, e.department_id,
    e.security_level, e.employment_status,
    p.first_name, p.last_name,
    (CASE 
        WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
        THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
        ELSE p.first_name || ' ' || p.last_name
    END)::text as full_name,
    u.email, u.user_type, u.account_status,
    -- Manager info
    mp.first_name as manager_first_name,
    mp.last_name as manager_last_name,
    (CASE 
        WHEN mp.middle_name IS NOT NULL AND mp.middle_name != '' 
        THEN mp.first_name || ' ' || mp.middle_name || ' ' || mp.last_name
        ELSE mp.first_name || ' ' || mp.last_name
    END)::text as manager_full_name,
    -- Direct reports count
    (SELECT COUNT(*) FROM employees sub 
     WHERE sub.manager_id = e.id AND sub.tenant_id = current_tenant_id() 
     AND sub.deleted_at IS NULL AND sub.employment_status = 'ACTIVE') as direct_reports_count
FROM employees e
JOIN persons p ON e.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
LEFT JOIN users u ON e.id = u.employee_id AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
LEFT JOIN employees me ON e.manager_id = me.id AND me.tenant_id = current_tenant_id() AND me.deleted_at IS NULL
LEFT JOIN persons mp ON me.person_id = mp.id AND mp.tenant_id = current_tenant_id() AND mp.deleted_at IS NULL
WHERE e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
  AND ($1::varchar IS NULL OR e.employment_status = $1)
  AND ($2::uuid IS NULL OR e.department_id = $2)
ORDER BY e.department_id, p.last_name, p.first_name;

-- ================================================================================================
-- MAINTENANCE AND CLEANUP QUERIES
-- ================================================================================================

-- name: RestoreSoftDeletedPerson :exec
UPDATE persons 
SET deleted_at = NULL, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NOT NULL;

-- name: RestoreSoftDeletedUser :exec
UPDATE users 
SET deleted_at = NULL, account_status = 'ACTIVE', updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NOT NULL;

-- name: RestoreSoftDeletedRole :exec
UPDATE roles 
SET deleted_at = NULL, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NOT NULL;

-- name: ListExpiredUserRolesDetailed :many
SELECT 
    ur.*,
    u.email, u.username,
    p.first_name, p.last_name,
    r.name as role_name, r.display_name as role_display_name,
    r.role_type,
    EXTRACT(DAYS FROM (NOW() - ur.expires_at))::INT as days_expired
FROM user_roles ur
JOIN users u ON ur.user_id = u.id AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
JOIN roles r ON ur.role_id = r.id AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
WHERE ur.expires_at IS NOT NULL 
    AND ur.expires_at <= NOW()
    AND ur.is_active = true
ORDER BY ur.expires_at DESC
LIMIT $1 OFFSET $2;

-- name: GetInactiveUsers :many
SELECT 
    u.id, u.username, u.email, u.user_type, u.account_status, u.last_login_at,
    p.first_name, p.last_name,
    EXTRACT(DAYS FROM (NOW() - COALESCE(u.last_login_at, u.created_at)))::INT as days_inactive
FROM users u
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
WHERE u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
    AND (u.last_login_at IS NULL OR u.last_login_at < NOW() - INTERVAL '90 days')
    AND u.created_at < NOW() - INTERVAL '30 days'  -- Exclude very new accounts
ORDER BY COALESCE(u.last_login_at, u.created_at)
LIMIT $1 OFFSET $2;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions 
WHERE tenant_id = current_tenant_id() 
    AND (expires_at < NOW() - INTERVAL '7 days' OR created_at < NOW() - INTERVAL '90 days');

-- name: CleanupExpiredPolicyEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id() 
    AND (expires_at < NOW() OR evaluated_at < NOW() - INTERVAL '24 hours');

-- name: CleanupOldAuditLogs :exec
DELETE FROM audit_log 
WHERE tenant_id = current_tenant_id() 
    AND created_at < NOW() - INTERVAL '2 years'
    AND severity IN ('LOW', 'INFO');

-- ================================================================================================
-- UTILITY AND HELPER QUERIES
-- ================================================================================================

-- name: GetPersonFullNameById :one
SELECT 
    CASE 
        WHEN middle_name IS NOT NULL AND middle_name != '' 
        THEN first_name || ' ' || middle_name || ' ' || last_name
        ELSE first_name || ' ' || last_name
    END as full_name
FROM persons 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetUserDisplayInfo :one
SELECT 
    u.id, u.username, u.email, u.user_type,
    COALESCE(
        CASE 
            WHEN p.middle_name IS NOT NULL AND p.middle_name != '' 
            THEN p.first_name || ' ' || p.middle_name || ' ' || p.last_name
            ELSE p.first_name || ' ' || p.last_name
        END,
        u.email,
        u.username
    ) as display_name,
    e.employee_number, e.position_title
FROM users u
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id() AND p.deleted_at IS NULL
LEFT JOIN employees e ON u.employee_id = e.id AND e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL
WHERE u.id = $1 AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL;

-- name: CheckUsernameAvailability :one
SELECT NOT EXISTS(
    SELECT 1 FROM users 
    WHERE username = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
) as available;

-- name: CheckEmailAvailability :one
SELECT NOT EXISTS(
    SELECT 1 FROM users 
    WHERE users.email = $1 AND users.tenant_id = current_tenant_id() AND users.deleted_at IS NULL
    UNION
    SELECT 1 FROM persons
    WHERE persons.email = $1 AND persons.tenant_id = current_tenant_id() AND persons.deleted_at IS NULL
) as available;

-- name: CheckEmployeeNumberAvailability :one
SELECT NOT EXISTS(
    SELECT 1 FROM employees 
    WHERE employee_number = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
) as available;

-- name: GetRolePermissionSummary :one
SELECT 
    r.id, r.name, r.display_name,
    COUNT(DISTINCT rp.permission_id) as total_permissions,
    COUNT(DISTINCT rp.permission_id) FILTER (WHERE p.effect = 'ALLOW') as allow_permissions,
    COUNT(DISTINCT rp.permission_id) FILTER (WHERE p.effect = 'DENY') as deny_permissions,
    array_agg(DISTINCT res.name ORDER BY res.name) FILTER (WHERE res.name IS NOT NULL) as resources,
    array_agg(DISTINCT a.name ORDER BY a.name) FILTER (WHERE a.name IS NOT NULL) as actions
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id AND rp.tenant_id = current_tenant_id() AND rp.is_active = true
LEFT JOIN permissions p ON rp.permission_id = p.id AND p.tenant_id = current_tenant_id() AND p.is_active = true
LEFT JOIN resources res ON p.resource_id = res.id AND res.tenant_id = current_tenant_id() AND res.deleted_at IS NULL
LEFT JOIN actions a ON p.action_id = a.id AND a.tenant_id = current_tenant_id() AND a.is_active = true
WHERE r.id = $1 AND r.tenant_id = current_tenant_id() AND r.deleted_at IS NULL
GROUP BY r.id, r.name, r.display_name;

-- ================================================================================================
-- DASHBOARD AND MONITORING QUERIES
-- ================================================================================================

-- name: GetTenantDashboardStats :one
SELECT 
    -- User stats
    (SELECT COUNT(*) FROM users WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL) as total_users,
    (SELECT COUNT(*) FROM users WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL AND is_active = true) as active_users,
    (SELECT COUNT(*) FROM users WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL AND last_login_at >= NOW() - INTERVAL '24 hours') as users_logged_in_24h,
    
    -- Employee stats
    (SELECT COUNT(*) FROM employees WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL AND employment_status = 'ACTIVE') as active_employees,
    
    -- Role stats
    (SELECT COUNT(*) FROM roles WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL) as total_roles,
    
    -- Session stats
    (SELECT COUNT(*) FROM user_sessions WHERE tenant_id = current_tenant_id() AND is_active = true AND expires_at > NOW()) as active_sessions,
    
    -- Security stats
    (SELECT COUNT(*) FROM audit_log WHERE tenant_id = current_tenant_id() AND created_at >= NOW() - INTERVAL '24 hours' AND severity IN ('HIGH', 'CRITICAL')) as high_risk_events_24h,
    (SELECT COUNT(*) FROM users WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL AND failed_login_attempts > 0) as users_with_failed_logins,
    
    -- Access request stats
    (SELECT COUNT(*) FROM access_requests WHERE tenant_id = current_tenant_id() AND approval_status = 'PENDING') as pending_access_requests;

-- name: GetRecentSecurityEvents :many
SELECT 
    al.event_type, al.event_category, al.severity, al.created_at, al.risk_score,
    u.username, u.email,
    p.first_name, p.last_name,
    al.ip_address, al.reason
FROM audit_log al
LEFT JOIN users u ON al.user_id = u.id AND u.tenant_id = current_tenant_id()
LEFT JOIN persons p ON u.person_id = p.id AND p.tenant_id = current_tenant_id()
WHERE al.tenant_id = current_tenant_id()
    AND al.created_at >= NOW() - INTERVAL '24 hours'
    AND al.severity IN ('HIGH', 'CRITICAL')
ORDER BY al.created_at DESC, al.risk_score DESC
LIMIT $1;
