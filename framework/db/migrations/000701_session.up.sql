
CREATE TABLE session (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,
    tenant_id  uuid        NOT NULL DEFAULT current_tenant_id(),
    org_unit_id UUID NOT NULL,

    plane TEXT NOT NULL,

    -- PostgreSQL arrays map naturally to []string
    roles TEXT[] NOT NULL DEFAULT '{}',
    permissions TEXT[] NOT NULL DEFAULT '{}',

    issued_at TIMESTAMPTZ NOT NULL,
    access_expires_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    access_token_hash TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Fast lookups
CREATE INDEX idx_session_user_id ON session(user_id);
CREATE INDEX idx_session_tenant_id ON session(tenant_id);
CREATE INDEX idx_session_org_unit_id ON session(org_unit_id);

-- Token lookups during authentication/revocation
CREATE UNIQUE INDEX idx_session_access_token_hash
    ON session(access_token_hash);

CREATE UNIQUE INDEX idx_session_refresh_token_hash
    ON session(refresh_token_hash);

-- Expired session cleanup
CREATE INDEX idx_session_access_expires_at
    ON session(access_expires_at);

CREATE INDEX idx_session_expires_at
    ON session(expires_at);

