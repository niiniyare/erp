-- name: CreateOrgUnit :one
INSERT INTO org_units (id, tenant_id, parent_id, name, code, path)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2, $3,
        CASE WHEN $1::uuid IS NULL THEN gen_random_uuid()::text
             ELSE (SELECT path FROM org_units WHERE id = $1) || '.' || gen_random_uuid()::text
        END)
RETURNING *;

-- name: GetOrgUnit :one
SELECT * FROM org_units
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListOrgUnits :many
SELECT * FROM org_units
WHERE tenant_id = current_tenant_id()
ORDER BY path;

-- name: ListOrgUnitChildren :many
SELECT * FROM org_units
WHERE tenant_id = current_tenant_id()
  AND parent_id = $1
ORDER BY name;

-- name: ListOrgUnitDescendants :many
SELECT * FROM org_units
WHERE tenant_id = current_tenant_id()
  AND path LIKE (SELECT p.path FROM org_units p WHERE p.id = $1) || '.%'
ORDER BY path;

-- name: UpdateOrgUnit :one
UPDATE org_units
SET name = $2, code = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeleteOrgUnit :exec
DELETE FROM org_units
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: InsertOrgUnitPath :exec
INSERT INTO org_unit_paths (ancestor_id, descendant_id, depth)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetAncestors :many
SELECT ancestor_id FROM org_unit_paths
WHERE descendant_id = $1
ORDER BY depth ASC;

-- name: GetDescendants :many
SELECT descendant_id FROM org_unit_paths
WHERE ancestor_id = $1
ORDER BY depth ASC;

-- name: DeleteOrgUnitPaths :exec
DELETE FROM org_unit_paths
WHERE descendant_id = $1;
