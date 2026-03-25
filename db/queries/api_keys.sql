-- API key management queries.
-- Run `make sqlc` after modifying this file.
--
-- GetAPIKeyByHash intentionally omits key_hash from the SELECT list (no need
-- to return the hash to Go code) and runs WITHOUT tenant context — the
-- key_hash column is globally unique and the lookup must work cross-tenant
-- during the authentication middleware.  This query MUST be executed by
-- admin_role (or a BYPASSRLS role) at the DB layer.

-- name: CreateAPIKey :one
INSERT INTO api_keys (
    tenant_id,
    name,
    key_hash,
    scopes,
    expires_at,
    created_by
) VALUES (
    current_tenant_id(),
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetAPIKeyByHash :one
-- Called on every API-key-authenticated request. Returns nil if revoked or expired.
-- NOTE: runs without tenant context (admin_role required); key_hash is globally unique.
SELECT
    id,
    tenant_id,
    name,
    scopes,
    created_by,
    expires_at,
    revoked_at,
    last_used_at,
    created_at
FROM api_keys
WHERE  key_hash   = $1
  AND  revoked_at IS NULL
  AND  (expires_at IS NULL OR expires_at > NOW());

-- name: RevokeAPIKey :exec
UPDATE api_keys
SET    revoked_at = NOW()
WHERE  tenant_id  = current_tenant_id()
  AND  id         = $1
  AND  revoked_at IS NULL;

-- name: ListAPIKeys :many
SELECT
    id,
    tenant_id,
    name,
    scopes,
    created_by,
    expires_at,
    revoked_at,
    last_used_at,
    created_at
FROM api_keys
WHERE  tenant_id = current_tenant_id()
ORDER BY created_at DESC;
