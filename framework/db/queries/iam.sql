-- name: CreateUser :one
INSERT INTO iam_users (id, tenant_id, email, password_hash, status)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2, 'active')
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM iam_users
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetUserByEmail :one
SELECT * FROM iam_users
WHERE email = $1 AND tenant_id = current_tenant_id();

-- name: ListUsers :many
SELECT * FROM iam_users
WHERE tenant_id = current_tenant_id()
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUserStatus :one
UPDATE iam_users
SET status = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: CreateRole :one
INSERT INTO iam_roles (id, tenant_id, name, description)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2)
RETURNING *;

-- name: ListRoles :many
SELECT * FROM iam_roles
WHERE tenant_id = current_tenant_id()
ORDER BY name;

-- name: AssignRole :exec
INSERT INTO iam_user_roles (user_id, role_id, tenant_id)
VALUES ($1, $2, current_tenant_id())
ON CONFLICT DO NOTHING;

-- name: RevokeRole :exec
DELETE FROM iam_user_roles
WHERE user_id = $1 AND role_id = $2 AND tenant_id = current_tenant_id();

-- name: ListUserRoles :many
SELECT r.* FROM iam_roles r
JOIN iam_user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1 AND ur.tenant_id = current_tenant_id();
