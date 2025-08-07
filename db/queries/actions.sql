-- name: CreateAction :one
INSERT INTO actions (
    tenant_id,
    name,
    action_type
) VALUES (
    current_tenant_id(), $1, $2
) RETURNING *;
