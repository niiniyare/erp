-- Phase 12 — API Key Authentication
-- Creates api_keys for machine-to-machine (M2M) authentication.
-- The raw bearer token is returned once at creation and NEVER stored.
-- Only the SHA-256 hex hash is persisted (key_hash column).
--
-- Sessions are NOT stored in user_sessions to avoid table bloat on
-- high-frequency API calls.  ResolvedSessions are built fresh on each
-- request and cached in Redis with a 5-minute TTL.

CREATE TABLE IF NOT EXISTS api_keys (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    -- SHA-256 hex of the raw bearer token. Raw token returned once at creation, never stored.
    key_hash     TEXT        NOT NULL UNIQUE,
    -- Allowed permission keys. Acts as a ceiling — cannot exceed the creating user's own permissions.
    scopes       TEXT[]      NOT NULL DEFAULT '{}',
    -- User who created this key. Key inherits at most the creator's permissions at creation time.
    created_by   UUID        NOT NULL REFERENCES users(id),
    expires_at   TIMESTAMPTZ,           -- NULL = never expires
    -- Set to revoke the key immediately. Checked on every request before permission evaluation.
    revoked_at   TIMESTAMPTZ,           -- NULL = active
    last_used_at TIMESTAMPTZ,           -- updated asynchronously on each validated request
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash  ON api_keys (key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys (tenant_id);

-- ─── Row-Level Security ───────────────────────────────────────────────────────

ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

-- application_role: full tenant isolation for all operations.
-- NOTE: GetAPIKeyByHash (cross-tenant hash lookup during auth) MUST be executed
--       by admin_role because no tenant context exists at token-validation time.
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE tablename = 'api_keys' AND policyname = 'api_keys_tenant_isolation'
    ) THEN
        CREATE POLICY api_keys_tenant_isolation
            ON api_keys FOR ALL TO application_role
            USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
            WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
    END IF;
END $$;

-- admin_role: full bypass — required for cross-tenant hash lookups and platform admin.
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE tablename = 'api_keys' AND policyname = 'api_keys_admin_access'
    ) THEN
        CREATE POLICY api_keys_admin_access
            ON api_keys FOR ALL TO admin_role
            USING  (TRUE) WITH CHECK (TRUE);
    END IF;
END $$;
