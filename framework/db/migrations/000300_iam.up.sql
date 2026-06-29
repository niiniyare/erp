-- ============================================================
-- 000300_iam — identity and access management
-- ============================================================

CREATE TABLE IF NOT EXISTS iam_users (
    id            uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id     uuid        NOT NULL DEFAULT current_tenant_id(),
    email         text        NOT NULL,
    password_hash text        NOT NULL DEFAULT '',
    status        text        NOT NULL DEFAULT 'active'
                              CHECK (status IN ('active','inactive','locked')),
    metadata      jsonb       NOT NULL DEFAULT '{}',
    created_at    timestamptz NOT NULL DEFAULT NOW(),
    updated_at    timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS iam_roles (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id   uuid        NOT NULL DEFAULT current_tenant_id(),
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    permissions jsonb       NOT NULL DEFAULT '[]',
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS iam_user_roles (
    user_id    uuid NOT NULL REFERENCES iam_users (id) ON DELETE CASCADE,
    role_id    uuid NOT NULL REFERENCES iam_roles (id) ON DELETE CASCADE,
    tenant_id  uuid NOT NULL DEFAULT current_tenant_id(),
    PRIMARY KEY (user_id, role_id)
);

SELECT apply_tenant_rls('iam_users');
SELECT apply_tenant_rls('iam_roles');
SELECT apply_tenant_rls('iam_user_roles');
