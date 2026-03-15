-- name: CreateSession :exec
INSERT INTO user_sessions (
    tenant_id,
    user_id,
    session_token,
    permissions,
    principal_id,
    ip_address,
    user_agent,
    expires_at,
    risk_score,
    is_active
) VALUES (
    current_tenant_id(),
    $1, -- user_id
    $2, -- session_token  (sha256hex of raw token — never store raw)
    $3, -- permissions    (JSONB map)
    $4, -- principal_id   (nullable UUID)
    $5, -- ip_address
    $6, -- user_agent
    $7, -- expires_at
    $8, -- risk_score
    TRUE
);

-- name: GetSessionByToken :one
SELECT
    id,
    user_id,
    tenant_id,
    session_token,
    permissions,
    principal_id,
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

-- name: InvalidateSession :exec
UPDATE user_sessions
SET    is_active = FALSE
WHERE  session_token = $1;

-- name: UpdateSessionLastSeen :exec
UPDATE user_sessions
SET    last_accessed_at = NOW()
WHERE  session_token = $1
  AND  is_active = TRUE;
