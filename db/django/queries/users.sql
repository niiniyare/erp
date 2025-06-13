-- ==============================================
-- USERS TABLE OPERATIONS
-- ==============================================

-- name: CreateUser :one
INSERT INTO users (
    tenant_id, entity_id, person_id, employee_id, username, email,
    password_hash, user_type, is_active, settings
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users 
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT * FROM users 
WHERE username = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users 
SET 
    entity_id = COALESCE($2, entity_id),
    person_id = COALESCE($3, person_id),
    employee_id = COALESCE($4, employee_id),
    username = COALESCE($5, username),
    email = COALESCE($6, email),
    password_hash = COALESCE($7, password_hash),
    user_type = COALESCE($8, user_type),
    is_active = COALESCE($9, is_active),
    settings = COALESCE($10, settings),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserLogin :exec
UPDATE users 
SET last_login_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $2, password_changed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: SoftDeleteUser :exec
UPDATE users 
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListUsers :many
SELECT u.*, p.first_name, p.last_name, e.employee_number
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
WHERE u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL
ORDER BY u.email;

-- name: ListActiveUsers :many
SELECT u.*, p.first_name, p.last_name, e.employee_number
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
WHERE u.tenant_id = current_tenant_id() AND u.is_active = true AND u.deleted_at IS NULL
ORDER BY u.email;

-- name: ListUsersByType :many
SELECT u.*, p.first_name, p.last_name, e.employee_number
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
WHERE u.tenant_id = current_tenant_id() AND u.user_type = $1 AND u.deleted_at IS NULL
ORDER BY u.email;

-- name: GetUserWithPersonDetails :one
SELECT 
    u.*,
    p.first_name, p.last_name, p.middle_name, p.phone, p.birth_date,
    e.employee_number, e.position_title, e.department_id
FROM users u
LEFT JOIN persons p ON u.person_id = p.id
LEFT JOIN employees e ON u.employee_id = e.id
WHERE u.id = $1 AND u.tenant_id = current_tenant_id() AND u.deleted_at IS NULL;


-- name: GetUserStats :one
SELECT 
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE is_active = true) as active_users,
    COUNT(*) FILTER (WHERE user_type = 'INTERNAL') as internal_users,
    COUNT(*) FILTER (WHERE user_type = 'CUSTOMER') as customer_users,
    COUNT(*) FILTER (WHERE user_type = 'VENDOR') as vendor_users,
    COUNT(*) FILTER (WHERE last_login_at >= NOW() - INTERVAL '30 days') as recent_logins
FROM users
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL;


