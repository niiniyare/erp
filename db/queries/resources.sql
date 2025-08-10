-- name: CreateResource :one
INSERT INTO
  resources (
    tenant_id,
    module_id,
    name,
    resource_type
  )
VALUES
  (current_tenant_id(), $1, $2, $3)
RETURNING
  *;
