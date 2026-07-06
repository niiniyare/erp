-- Migration: create IAM tables (iam_user, iam_role, iam_session, iam_api_token, iam_user_role)
-- iam_user is a system entity; all others are custom/JSONB-backed but declared
-- here as typed tables for query performance and FK integrity.

CREATE TABLE IF NOT EXISTS iam_user (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid        NOT NULL REFERENCES platform_tenant (id),
    email            varchar(254) NOT NULL,
    password_hash    text        NOT NULL,
    first_name       varchar(100),
    last_name        varchar(100),
    status           varchar(30)  NOT NULL DEFAULT 'pending_verification'
                                  CHECK (status IN ('pending_verification','active','suspended','locked','deleted')),
    failed_attempts  bigint       NOT NULL DEFAULT 0,
    locked_until     timestamptz,
    email_verified_at timestamptz,
    last_login_at    timestamptz,
    mfa_enabled      boolean      NOT NULL DEFAULT FALSE,
    mfa_secret       text,
    locale           varchar(10)  NOT NULL DEFAULT 'en-KE',
    timezone         varchar(64)  NOT NULL DEFAULT 'Africa/Nairobi',
    custom_fields    jsonb        NOT NULL DEFAULT '{}',
    created_at       timestamptz  NOT NULL DEFAULT NOW(),
    updated_at       timestamptz  NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

ALTER TABLE iam_user ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_user FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_user
    USING (tenant_id = current_tenant_id());

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_iam_user_email_trgm
    ON iam_user USING gin (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_iam_user_tenant ON iam_user (tenant_id);

-- iam_role: tenant-scoped named permission sets.
CREATE TABLE IF NOT EXISTS iam_role (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES platform_tenant (id),
    name        varchar(100) NOT NULL,
    label       varchar(255),
    description varchar(1024),
    is_system   boolean     NOT NULL DEFAULT FALSE,
    data        jsonb       NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

ALTER TABLE iam_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_role FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_role
    USING (tenant_id = current_tenant_id());

-- iam_session: active session tokens.
CREATE TABLE IF NOT EXISTS iam_session (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES platform_tenant (id),
    user_id     uuid        NOT NULL REFERENCES iam_user (id),
    token_hash  varchar(128) NOT NULL UNIQUE,
    ip_address  varchar(45),
    user_agent  varchar(1024),
    expires_at  timestamptz NOT NULL,
    revoked     boolean     NOT NULL DEFAULT FALSE,
    revoked_at  timestamptz,
    data        jsonb       NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE iam_session ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_session FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_session
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS idx_iam_session_token ON iam_session (token_hash);
CREATE INDEX IF NOT EXISTS idx_iam_session_user  ON iam_session (user_id);

-- iam_api_token: long-lived machine-to-machine tokens.
CREATE TABLE IF NOT EXISTS iam_api_token (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES platform_tenant (id),
    user_id      uuid        NOT NULL REFERENCES iam_user (id),
    name         varchar(255) NOT NULL,
    token_hash   varchar(128) NOT NULL UNIQUE,
    scopes       jsonb       NOT NULL DEFAULT '[]',
    expires_at   timestamptz,
    last_used_at timestamptz,
    revoked_at   timestamptz,
    data         jsonb       NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE iam_api_token ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_api_token FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_api_token
    USING (tenant_id = current_tenant_id());

-- iam_user_role: many-to-many join table.
CREATE TABLE IF NOT EXISTS iam_user_role (
    id          uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid    NOT NULL REFERENCES platform_tenant (id),
    user_id     uuid    NOT NULL REFERENCES iam_user (id),
    role_id     uuid    NOT NULL REFERENCES iam_role (id),
    org_unit_id uuid    REFERENCES platform_org_unit (id),
    data        jsonb   NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, role_id)
);

ALTER TABLE iam_user_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_user_role FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_user_role
    USING (tenant_id = current_tenant_id());
