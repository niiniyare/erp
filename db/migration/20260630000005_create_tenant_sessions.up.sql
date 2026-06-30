-- tenant_sessions: durable, server-side session store (v2.1 opaque tokens).
-- Redis is the hot path; this table is the system of record for audit,
-- reporting, and rebuilding the Redis cache after a restart.
--
-- Only hashes of tokens are ever written here — the raw credential stays
-- in the client's hands alone.

CREATE TABLE IF NOT EXISTS tenant_sessions (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES tenant_users(id) ON DELETE CASCADE,
    tenant_id           UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    actor_type          TEXT        NOT NULL DEFAULT 'user'
                            CHECK (actor_type IN ('user', 'service_account', 'api_key')),
    plane               TEXT        NOT NULL DEFAULT 'tenant' CHECK (plane = 'tenant'),

    -- SHA-256 hashes of the opaque tokens.
    -- Used as Redis keys and for revocation / audit lookups.
    access_token_hash   TEXT        NOT NULL UNIQUE,
    refresh_token_hash  TEXT        NOT NULL UNIQUE,

    -- Permission snapshot baked at login; refreshed on token rotation.
    permissions         TEXT[]      NOT NULL DEFAULT '{}',
    roles               TEXT[]      NOT NULL DEFAULT '{}',
    org_unit_id         UUID        REFERENCES org_units(id) ON DELETE SET NULL,

    ip_address          INET,
    user_agent          TEXT,
    mfa_verified        BOOLEAN     NOT NULL DEFAULT FALSE,

    issued_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    access_expires_at   TIMESTAMPTZ NOT NULL,
    refresh_expires_at  TIMESTAMPTZ NOT NULL,
    last_active_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- NULL = active session; set on logout, rotation, or forced revocation.
    revoked_at          TIMESTAMPTZ,
    revoke_reason       TEXT
);

-- Fast lookups by token hash (middleware path).
CREATE INDEX IF NOT EXISTS tenant_sessions_access_hash_idx
    ON tenant_sessions (access_token_hash)
    WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS tenant_sessions_refresh_hash_idx
    ON tenant_sessions (refresh_token_hash)
    WHERE revoked_at IS NULL;

-- Bulk revocation by user or tenant (password change, tenant suspend, etc.).
CREATE INDEX IF NOT EXISTS tenant_sessions_user_id_idx
    ON tenant_sessions (user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS tenant_sessions_tenant_id_idx
    ON tenant_sessions (tenant_id)
    WHERE revoked_at IS NULL;

-- RLS: each tenant can only see its own session rows.
ALTER TABLE tenant_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_sessions FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_sessions_isolation ON tenant_sessions
    USING (tenant_id = current_tenant_id());
