-- name: CreateSession :exec
-- Inserts a fully pre-computed session at login.
-- configuration = {"flags":{...},"settings":{...},"prefs":{...}} built from 5 concurrent queries.
-- entity_scope  = {"type":"all"|"subtree"|"entity","entity_id":"uuid","path_prefix":"/.../"}.
INSERT INTO user_sessions (
    tenant_id,
    user_id,
    user_type,
    session_token,
    permissions,
    principal_id,
    entity_scope,
    configuration,
    ip_address,
    user_agent,
    expires_at,
    risk_score,
    is_active
) VALUES (
    current_tenant_id(),
    $1,  -- user_id
    $2,  -- user_type     (copied from users for SetDBPool without extra join)
    $3,  -- session_token (sha256hex of raw token — raw token never stored)
    $4,  -- permissions   (JSONB: {"finance.transactions.read": true, ...})
    $5,  -- principal_id  (nullable UUID: portal users' business record)
    $6,  -- entity_scope  (JSONB: pre-computed access scope)
    $7,  -- configuration (JSONB: flags + settings + prefs)
    $8,  -- ip_address
    $9,  -- user_agent
    $10, -- expires_at
    $11, -- risk_score
    TRUE
);

-- name: GetSessionByToken :one
-- Validates and returns the full session. Used on every authenticated request.
SELECT
    id,
    user_id,
    tenant_id,
    user_type,
    session_token,
    permissions,
    principal_id,
    entity_scope,
    configuration,
    ip_address,
    user_agent,
    expires_at,
    risk_score,
    is_active,
    last_accessed_at
FROM user_sessions
WHERE session_token = $1
  AND is_active = TRUE
  AND expires_at > NOW();

-- name: TouchAndGetSession :one
-- Atomically updates last_accessed_at and returns the session in one round-trip.
-- Use on every request instead of separate GetSession + UpdateLastSeen.
UPDATE user_sessions
SET    last_accessed_at = NOW()
WHERE  session_token = $1
  AND  is_active = TRUE
  AND  expires_at > NOW()
RETURNING
    id, user_id, tenant_id, user_type, session_token,
    permissions, principal_id, entity_scope, configuration,
    ip_address, user_agent, expires_at, risk_score, is_active, last_accessed_at;

-- name: InvalidateSession :exec
-- Logout: immediately deactivates a single session.
UPDATE user_sessions
SET    is_active = FALSE
WHERE  session_token = $1;

-- name: UpdateSessionLastSeen :exec
-- Async background touch — use when you don't need the session returned.
UPDATE user_sessions
SET    last_accessed_at = NOW()
WHERE  session_token = $1
  AND  is_active = TRUE;

-- name: InvalidateSessionsByUser :exec
-- Called on SuspendUser, TerminateEmployee, ChangePassword.
-- Forces fresh permission/flag recomputation at next login.
UPDATE user_sessions
SET    is_active = FALSE
WHERE  user_id = $1
  AND  is_active = TRUE;

-- name: InvalidateSessionsByTenant :exec
-- Called when a module flag changes (feature appears/disappears for all users).
UPDATE user_sessions
SET    is_active = FALSE
WHERE  tenant_id = $1
  AND  is_active = TRUE;

-- name: CleanupExpiredSessions :exec
-- Run by background job to purge old rows.
DELETE FROM user_sessions
WHERE expires_at < NOW();

-- name: CountActiveSessionsByUser :one
SELECT COUNT(*) FROM user_sessions
WHERE user_id  = $1
  AND is_active = TRUE
  AND expires_at > NOW();

-- =====================================================================
-- API KEY QUERIES
-- =====================================================================

-- -- name: CreateAPIKey :one
-- INSERT INTO api_keys (tenant_id, name, key_hash, scopes, created_by, expires_at)
-- VALUES (current_tenant_id(), $1, $2, $3, $4, $5)
-- RETURNING *;
--
-- -- name: GetAPIKeyByHash :one
-- -- Called on every API-key-authenticated request. Returns nil if revoked or expired.
-- SELECT id, tenant_id, name, key_hash, scopes, created_by, expires_at, revoked_at, last_used_at
-- FROM   api_keys
-- WHERE  key_hash  = $1
--   AND  revoked_at IS NULL
--   AND  (expires_at IS NULL OR expires_at > NOW());
--
-- -- name: TouchAPIKeyLastUsed :exec
-- UPDATE api_keys SET last_used_at = NOW() WHERE id = $1;

-- name: RevokeAPIKey :exec
-- UPDATE api_keys
-- SET    revoked_at = NOW()
-- WHERE  id        = $1
--   AND  tenant_id = current_tenant_id();

-- name: ListAPIKeys :many
-- SELECT id, name, scopes, created_by, expires_at, revoked_at, last_used_at, created_at
-- FROM   api_keys
-- WHERE  tenant_id = current_tenant_id()
-- ORDER  BY created_at DESC;
