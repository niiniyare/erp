-- ------------------------------------------------------------------------------------------------
-- IAM — PASSWORD RESET TOKENS + PASSWORD HISTORY
-- ------------------------------------------------------------------------------------------------
-- password_reset_tokens: single-use, time-limited tokens for the forgot-password flow.
-- token_hash: SHA-256 of the raw token (raw token is emailed; hash is stored).
-- used_at: set on first use; NULL means token is still available.
--
-- password_history: JSONB array of the last-5 bcrypt hashes stored on the users row.
-- Allows the service to reject re-use of recent passwords.
-- ------------------------------------------------------------------------------------------------

-- 1. Password reset tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_hash
    ON password_reset_tokens (token_hash);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user
    ON password_reset_tokens (tenant_id, user_id);

-- 2. Add password_history column to users
-- Stores the last 5 bcrypt hashes as a JSON array.
-- Checked during ChangePassword and ResetPassword to prevent re-use.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_history JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON TABLE  password_reset_tokens IS
    'Single-use tokens for the forgot-password email flow. '
    'token_hash is SHA-256 of the raw token. '
    'used_at IS NOT NULL means the token has been consumed.';

COMMENT ON COLUMN users.password_history IS
    'JSON array of the last 5 bcrypt hashes. '
    'Used to prevent password re-use. Capped at 5 entries.';

-- 3. Row-Level Security for password_reset_tokens
ALTER TABLE password_reset_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE password_reset_tokens FORCE  ROW LEVEL SECURITY;

CREATE POLICY password_reset_tokens_tenant_isolation
    ON password_reset_tokens FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY password_reset_tokens_admin_access
    ON password_reset_tokens FOR ALL TO admin_role
    USING  (TRUE)
    WITH CHECK (TRUE);
