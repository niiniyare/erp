-- name: CreateTenant :one
INSERT INTO tenants (name, slug, plan, company_size, country, currency, timezone, locale)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenantStatus :one
UPDATE tenants SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, plan = $3, company_size = $4, country = $5,
    currency = $6, timezone = $7, locale = $8, metadata = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2;
