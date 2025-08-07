-- name: CreateModule :one
INSERT INTO modules (
    tenant_id,
    name,
    display_name,
    category
) VALUES (
    current_tenant_id(), $1, $2, $3
) RETURNING *;
