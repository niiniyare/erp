-- SSO provider configuration queries.
-- Run `make sqlc` after modifying this file.

-- name: GetSSOProvider :one
SELECT *
FROM sso_providers
WHERE tenant_id = current_tenant_id()
  AND provider  = $1
  AND is_active = TRUE
LIMIT 1;

-- name: UpsertSSOProvider :one
INSERT INTO sso_providers (
    tenant_id,
    provider,
    client_id,
    client_secret_enc,
    scopes,
    redirect_uri,
    extra_params,
    auto_provision,
    default_entity_id,
    is_active
) VALUES (current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, TRUE)
ON CONFLICT (tenant_id, provider) DO UPDATE SET
    client_id          = EXCLUDED.client_id,
    client_secret_enc  = EXCLUDED.client_secret_enc,
    scopes             = EXCLUDED.scopes,
    redirect_uri       = EXCLUDED.redirect_uri,
    extra_params       = EXCLUDED.extra_params,
    auto_provision     = EXCLUDED.auto_provision,
    default_entity_id  = EXCLUDED.default_entity_id,
    is_active          = TRUE,
    updated_at         = NOW()
RETURNING *;

-- name: DeactivateSSOProvider :exec
UPDATE sso_providers
SET    is_active  = FALSE,
       updated_at = NOW()
WHERE  tenant_id = current_tenant_id()
  AND  provider  = $1;

-- name: ListSSOProviders :many
SELECT *
FROM sso_providers
WHERE tenant_id = current_tenant_id()
ORDER BY provider;
