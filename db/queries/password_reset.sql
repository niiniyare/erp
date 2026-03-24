-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (tenant_id, user_id, token_hash, expires_at)
VALUES (current_tenant_id(), $1, $2, $3);

-- name: GetPasswordResetToken :one
-- Returns the token row regardless of used_at so the caller can detect already-used tokens.
SELECT id, user_id, tenant_id, expires_at, used_at
FROM password_reset_tokens
WHERE token_hash = $1
  AND tenant_id  = current_tenant_id()
LIMIT 1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE token_hash = $1
  AND tenant_id  = current_tenant_id();

-- name: GetUserPasswordHistory :one
SELECT password_history
FROM users
WHERE id = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: UpdatePasswordAndHistory :exec
-- Updates the password hash and prepends the new hash to password_history,
-- keeping only the last 5 entries.
UPDATE users
SET password_hash    = $2,
    password_history = (
        SELECT jsonb_agg(h)
        FROM (
            SELECT jsonb_array_elements_text($3::jsonb) AS h
            LIMIT 5
        ) sub
    ),
    updated_at = NOW()
WHERE id = $1
  AND tenant_id = current_tenant_id();
