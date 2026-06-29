-- name: GetSetting :one
SELECT value FROM settings
WHERE tenant_id = current_tenant_id() AND key = $1;

-- name: UpsertSetting :one
INSERT INTO settings (id, tenant_id, key, value)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2)
ON CONFLICT (tenant_id, key) DO UPDATE
  SET value = EXCLUDED.value, updated_at = NOW()
RETURNING *;

-- name: DeleteSetting :exec
DELETE FROM settings
WHERE tenant_id = current_tenant_id() AND key = $1;

-- name: ListSettings :many
SELECT * FROM settings
WHERE tenant_id = current_tenant_id()
ORDER BY key;
