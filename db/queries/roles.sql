

-- ==============================================
-- ROLES TABLE OPERATIONS
-- ==============================================

-- name: CreateRole :one
INSERT INTO roles (
    tenant_id, entity_id, name, description, module, permissions,
    entity_scope, is_system_role
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetRole :one
SELECT * FROM roles 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetRoleByName :one
SELECT * FROM roles 
WHERE name = $1 AND tenant_id = current_tenant_id();

-- name: UpdateRole :one
UPDATE roles 
SET 
    entity_id = COALESCE($2, entity_id),
    name = COALESCE($3, name),
    description = COALESCE($4, description),
    module = COALESCE($5, module),
    permissions = COALESCE($6, permissions),
    entity_scope = COALESCE($7, entity_scope)
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles 
WHERE id = $1 AND tenant_id = current_tenant_id() AND is_system_role = false;

-- name: ListRoles :many
SELECT * FROM roles 
WHERE tenant_id = current_tenant_id()
ORDER BY name;

-- name: ListRolesByModule :many
SELECT * FROM roles 
WHERE tenant_id = current_tenant_id() AND module = $1
ORDER BY name;

-- name: ListSystemRoles :many
SELECT * FROM roles 
WHERE tenant_id = current_tenant_id() AND is_system_role = true
ORDER BY name;

-- name: ListCustomRoles :many
SELECT * FROM roles 
WHERE tenant_id = current_tenant_id() AND is_system_role = false
ORDER BY name;

-- ==============================================
-- USER ROLES TABLE OPERATIONS
-- ==============================================

-- name: AssignUserRole :one
INSERT INTO user_roles (user_id, role_id, entity_id, assigned_by, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, role_id, entity_id) DO UPDATE SET
    assigned_at = NOW(),
    assigned_by = EXCLUDED.assigned_by,
    expires_at = EXCLUDED.expires_at
RETURNING *;

-- name: RevokeUserRole :exec
DELETE FROM user_roles 
WHERE user_id = $1 AND role_id = $2 AND entity_id = $3;

-- name: GetUserRoles :many
SELECT ur.*, r.name as role_name, r.description as role_description, r.permissions
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_id = $1 
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
ORDER BY r.name;

-- name: GetRoleUsers :many
SELECT ur.*, u.email, u.username, p.first_name, p.last_name
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
LEFT JOIN persons p ON u.person_id = p.id
WHERE ur.role_id = $1 
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    AND u.deleted_at IS NULL
ORDER BY p.last_name, p.first_name, u.email;

-- name: GetUserPermissions :many
SELECT DISTINCT r.module, r.permissions
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_id = $1 
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

-- name: CheckUserPermission :one
SELECT EXISTS(
    SELECT 1 
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = $1 
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
        AND r.permissions ? $2
) as has_permission;

-- name: ListExpiredUserRoles :many
SELECT ur.*, u.email, r.name as role_name
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE ur.expires_at IS NOT NULL 
    AND ur.expires_at <= NOW()
    AND u.tenant_id = current_tenant_id()
ORDER BY ur.expires_at DESC;

-- name: CleanupExpiredUserRoles :exec
DELETE FROM user_roles 
WHERE expires_at IS NOT NULL AND expires_at <= NOW() - INTERVAL '30 days';




-- name: SearchUsersWithRoles :many
SELECT DISTINCT 
    u.id, u.username, u.email, u.user_type, u.is_active,
    p.first_name, p.last_name,
    string_agg(r.name, ', ') as roles
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN user_roles ur ON u.id = ur.user_id AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.tenant_id = current_tenant_id() 
    AND u.deleted_at IS NULL
    AND (
        u.email ILIKE '%' || $1 || '%' OR
        u.username ILIKE '%' || $1 || '%' OR
        p.first_name ILIKE '%' || $1 || '%' OR
        p.last_name ILIKE '%' || $1 || '%'
    )
GROUP BY u.id, u.username, u.email, u.user_type, u.is_active, p.first_name, p.last_name
ORDER BY p.last_name, p.first_name, u.email
LIMIT $2;
