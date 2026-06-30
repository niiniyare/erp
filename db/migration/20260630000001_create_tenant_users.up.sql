-- Phase 9: IAM — tenant_users table.
--
-- Stores authenticated principals (employees, service accounts) for one tenant.
-- Scoped by RLS: every query must run with set_tenant_context($tenantID) active.
-- password_hash is Argon2id — never returned via API (IsSensitive on field def).

CREATE TABLE tenant_users (
    id                   UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id            UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email                VARCHAR(320) NOT NULL,
    full_name            VARCHAR(255) NOT NULL,
    status               VARCHAR(32)  NOT NULL DEFAULT 'PENDING_VERIFICATION'
                         CHECK (status IN (
                             'PENDING_VERIFICATION', 'ACTIVE', 'SUSPENDED', 'DEACTIVATED'
                         )),
    actor_type           VARCHAR(32)  NOT NULL DEFAULT 'user'
                         CHECK (actor_type IN ('user', 'service_account')),
    password_hash        TEXT,                              -- NULL when SSO-only
    org_unit_id          UUID         REFERENCES orgunits(uuid) ON DELETE SET NULL,
    must_change_password BOOLEAN      NOT NULL DEFAULT FALSE,
    failed_login_count   INTEGER      NOT NULL DEFAULT 0,
    locked_until         TIMESTAMPTZ,
    last_login_at        TIMESTAMPTZ,
    last_login_ip        VARCHAR(45),
    metadata             JSONB        NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Enforce unique email per tenant (not globally).
CREATE UNIQUE INDEX tenant_users_tenant_email  ON tenant_users (tenant_id, email);
CREATE INDEX        tenant_users_tenant_id     ON tenant_users (tenant_id);
CREATE INDEX        tenant_users_org_unit_id   ON tenant_users (org_unit_id)
    WHERE org_unit_id IS NOT NULL;

ALTER TABLE tenant_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_users FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON tenant_users
    USING (tenant_id = current_tenant_id());
