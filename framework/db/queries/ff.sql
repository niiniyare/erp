-- name: GetFlag :one
SELECT * FROM feature_flags
WHERE key = $1
  AND (tenant_id IS NULL OR tenant_id = current_tenant_id())
ORDER BY tenant_id NULLS LAST
LIMIT 1;

-- name: UpsertFlag :one
INSERT INTO feature_flags (id, tenant_id, key, enabled, rollout_pct)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2, $3)
ON CONFLICT (tenant_id, key) DO UPDATE
  SET enabled = EXCLUDED.enabled,
      rollout_pct = EXCLUDED.rollout_pct,
      updated_at = NOW()
RETURNING *;

-- name: UpsertGlobalFlag :one
INSERT INTO feature_flags (id, tenant_id, key, enabled, rollout_pct)
VALUES (gen_random_uuid(), NULL, $1, $2, $3)
ON CONFLICT (tenant_id, key) DO UPDATE
  SET enabled = EXCLUDED.enabled,
      rollout_pct = EXCLUDED.rollout_pct,
      updated_at = NOW()
RETURNING *;

-- name: ListFlags :many
SELECT * FROM feature_flags
WHERE tenant_id IS NULL OR tenant_id = current_tenant_id()
ORDER BY key, tenant_id NULLS LAST;
