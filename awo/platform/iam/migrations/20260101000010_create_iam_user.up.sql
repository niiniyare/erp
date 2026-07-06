-- iam_user: canonical user identity. Mandatory system entity.
CREATE TABLE IF NOT EXISTS iam_user (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES platform_tenant (id),
    email           varchar(254) NOT NULL,
    name            varchar(255) NOT NULL,
    password_hash   varchar(255) NOT NULL,
    status          varchar(20)  NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active','inactive','locked')),
    failed_attempts integer      NOT NULL DEFAULT 0,
    locked_until    timestamptz,
    last_login_at   timestamptz,
    mfa_enabled     boolean      NOT NULL DEFAULT false,
    mfa_secret      varchar(255),
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT iam_user_email_unique UNIQUE (tenant_id, email)
);

ALTER TABLE iam_user ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_user FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_user
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS iam_user_tenant_idx ON iam_user (tenant_id);
CREATE INDEX IF NOT EXISTS iam_user_email_trgm ON iam_user USING GIN (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS iam_user_name_trgm  ON iam_user USING GIN (name gin_trgm_ops);

-- iam_role: RBAC role definition.
CREATE TABLE IF NOT EXISTS iam_role (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid REFERENCES platform_tenant (id),
    name        varchar(255) NOT NULL,
    label       varchar(255) NOT NULL,
    is_system   boolean      NOT NULL DEFAULT false,
    permissions jsonb,
    description varchar(1024),
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE iam_role ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_role FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_role
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS iam_role_tenant_idx ON iam_role (tenant_id);
CREATE INDEX IF NOT EXISTS iam_role_name_trgm  ON iam_role USING GIN (name gin_trgm_ops);

-- iam_session: active user sessions.
CREATE TABLE IF NOT EXISTS iam_session (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid NOT NULL REFERENCES iam_user (id),
    tenant_id           uuid NOT NULL REFERENCES platform_tenant (id),
    access_token_hash   varchar(64) NOT NULL UNIQUE,
    refresh_token_hash  varchar(64) NOT NULL UNIQUE,
    ip_address          varchar(45),
    user_agent          varchar(1024),
    expires_at          timestamptz NOT NULL,
    revoked_at          timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE iam_session ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_session FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_session
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS iam_session_user_idx          ON iam_session (user_id);
CREATE INDEX IF NOT EXISTS iam_session_tenant_idx        ON iam_session (tenant_id);
CREATE INDEX IF NOT EXISTS iam_session_access_token_idx  ON iam_session (access_token_hash);
CREATE INDEX IF NOT EXISTS iam_session_refresh_token_idx ON iam_session (refresh_token_hash);
CREATE INDEX IF NOT EXISTS iam_session_expires_idx       ON iam_session (expires_at);

-- iam_api_token: long-lived machine tokens.
CREATE TABLE IF NOT EXISTS iam_api_token (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES iam_user (id),
    tenant_id    uuid NOT NULL REFERENCES platform_tenant (id),
    name         varchar(255) NOT NULL,
    token_hash   varchar(64)  NOT NULL UNIQUE,
    scopes       jsonb,
    expires_at   timestamptz,
    last_used_at timestamptz,
    revoked_at   timestamptz,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE iam_api_token ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam_api_token FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam_api_token
    USING (tenant_id = current_tenant_id());

CREATE INDEX IF NOT EXISTS iam_api_token_tenant_idx     ON iam_api_token (tenant_id);
CREATE INDEX IF NOT EXISTS iam_api_token_hash_idx       ON iam_api_token (token_hash);
CREATE INDEX IF NOT EXISTS iam_api_token_name_trgm      ON iam_api_token USING GIN (name gin_trgm_ops);

-- iam_user_role: user→role assignments (with optional org unit scope).
CREATE TABLE IF NOT EXISTS iam_user_role (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES iam_user (id),
    role_id     uuid NOT NULL REFERENCES iam_role (id),
    org_unit_id uuid REFERENCES platform_org_unit (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT iam_user_role_unique UNIQUE (user_id, role_id, org_unit_id)
);

CREATE INDEX IF NOT EXISTS iam_user_role_user_idx ON iam_user_role (user_id);
CREATE INDEX IF NOT EXISTS iam_user_role_role_idx ON iam_user_role (role_id);
