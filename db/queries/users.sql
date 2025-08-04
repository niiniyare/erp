-- name: CreateUser :one
INSERT INTO users (
    entity_id, person_id, employee_id, username, email, password_hash, user_type, account_status, session_timeout_minutes, mfa_enabled, user_attributes, settings
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: GetUserPasswordByID :one
SELECT password_hash FROM users WHERE id = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: UpdateUser :one
UPDATE users
SET
    username = COALESCE(sqlc.narg(username), username),
    email = COALESCE(sqlc.narg(email), email),
    user_type = COALESCE(sqlc.narg(user_type), user_type),
    account_status = COALESCE(sqlc.narg(account_status), account_status),
    session_timeout_minutes = COALESCE(sqlc.narg(session_timeout_minutes), session_timeout_minutes),
    mfa_enabled = COALESCE(sqlc.narg(mfa_enabled), mfa_enabled),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND tenant_id = current_tenant_id()
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: RestoreSoftDeletedUser :exec
UPDATE users SET deleted_at = NULL WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListUsers :many
SELECT * FROM users
WHERE
    (sqlc.narg(user_type)::text IS NULL OR user_type = sqlc.narg(user_type)::text)
AND (sqlc.narg(account_status)::text IS NULL OR account_status = sqlc.narg(account_status)::text)
AND deleted_at IS NULL
AND tenant_id = current_tenant_id()
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: GetCompleteUserProfile :one
SELECT
    u.*,
    p.*,
    e.*
FROM
    users u
LEFT JOIN
    persons p ON u.person_id = p.id AND p.tenant_id = u.tenant_id
LEFT JOIN
    employees e ON u.employee_id = e.id AND e.tenant_id = u.tenant_id
WHERE
    u.id = $1 AND u.deleted_at IS NULL;

-- name: CheckEmailAvailability :one
SELECT COUNT(*) = 0 FROM users WHERE email = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: CheckUsernameAvailability :one
SELECT COUNT(*) = 0 FROM users WHERE username = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: CheckEmployeeNumberAvailability :one
SELECT COUNT(*) = 0 FROM employees WHERE employee_number = $1 AND deleted_at IS NULL AND tenant_id = current_tenant_id();

-- name: SearchUsersAdvanced :many
SELECT * FROM users
WHERE
    (username ILIKE '%' || sqlc.arg(query) || '%' OR email ILIKE '%' || sqlc.arg(query) || '%')
AND deleted_at IS NULL
LIMIT $1
OFFSET $2;

-- name: IncrementFailedLogins :exec
UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: UnlockUser :exec
UPDATE users SET failed_login_attempts = 0, lockout_until = NULL WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = NOW() WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: AssignUserRole :one
SELECT assign_user_role($1, $2, $3, $4);

-- name: RevokeUserRole :exec
SELECT revoke_user_role($1, $2, $3);

