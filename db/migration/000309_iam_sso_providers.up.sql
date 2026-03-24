-- ------------------------------------------------------------------------------------------------
-- IAM — SSO PROVIDER CONFIGURATIONS
-- ------------------------------------------------------------------------------------------------
-- sso_providers: per-tenant OAuth/OIDC provider configurations.
-- client_secret_enc: AES-256-GCM encrypted client secret (raw secret is never stored).
-- extra_params: provider-specific JSON config, e.g. {"tenant": "common"} for Microsoft Azure AD.
-- auto_provision: when true, unknown SSO users are created automatically (JIT provisioning).
-- ------------------------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sso_providers (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider          TEXT        NOT NULL,
    client_id         TEXT        NOT NULL,
    client_secret_enc TEXT        NOT NULL,
    scopes            TEXT[]      NOT NULL DEFAULT '{openid,email,profile}',
    redirect_uri      TEXT        NOT NULL,
    extra_params      JSONB       NOT NULL DEFAULT '{}',
    auto_provision      BOOLEAN     NOT NULL DEFAULT FALSE,
    default_entity_id   UUID        NULL,               -- entity assigned to JIT-provisioned users
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sso_provider_tenant_provider UNIQUE (tenant_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_sso_providers_tenant_active
    ON sso_providers (tenant_id) WHERE is_active = TRUE;

-- Row-Level Security
ALTER TABLE sso_providers ENABLE ROW LEVEL SECURITY;

CREATE POLICY sso_providers_tenant_isolation
    ON sso_providers FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY sso_providers_admin_access
    ON sso_providers FOR ALL TO admin_role
    USING  (TRUE)
    WITH CHECK (TRUE);

COMMENT ON TABLE sso_providers IS
    'Per-tenant OAuth/OIDC SSO provider configurations. '
    'client_secret_enc is AES-256-GCM encrypted. '
    'extra_params holds provider-specific config (e.g. Azure AD tenant for Microsoft). '
    'auto_provision=true enables just-in-time user creation on first SSO login.';
