-- Policies CRUD Operations

-- name: CreatePolicy :one
INSERT INTO policies (
    tenant_id, entity_id, name, display_name, description, policy_type,
    effect, priority, category, target, rule, obligations, advice, is_active, created_by
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetPolicy :one
SELECT * FROM policies
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetPolicyByName :one
SELECT * FROM policies
WHERE name = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: UpdatePolicy :one
UPDATE policies
SET
    entity_id = COALESCE($2, entity_id),
    name = COALESCE($3, name),
    display_name = COALESCE($4, display_name),
    description = COALESCE($5, description),
    policy_type = COALESCE($6, policy_type),
    effect = COALESCE($7, effect),
    priority = COALESCE($8, priority),
    category = COALESCE($9, category),
    target = COALESCE($10, target),
    rule = COALESCE($11, rule),
    obligations = COALESCE($12, obligations),
    advice = COALESCE($13, advice),
    is_active = COALESCE($14, is_active),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePolicy :exec
UPDATE policies
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: HardDeletePolicy :exec
DELETE FROM policies
WHERE id = $1 AND tenant_id = current_tenant_id();

-- Policy Listing and Filtering

-- name: ListPolicies :many
SELECT * FROM policies
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
ORDER BY name;

-- name: ListActivePolicies :many
SELECT * FROM policies
WHERE tenant_id = current_tenant_id() AND is_active = true AND deleted_at IS NULL
ORDER BY name;

-- name: ListPoliciesByCategory :many
SELECT * FROM policies
WHERE tenant_id = current_tenant_id() AND category = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: ListPoliciesByEffect :many
SELECT * FROM policies
WHERE tenant_id = current_tenant_id() AND effect = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: SearchPolicies :many
SELECT * FROM policies
WHERE tenant_id = current_tenant_id()
    AND (name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
    AND deleted_at IS NULL
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: GetApplicablePolicies :many
SELECT p.*
FROM policies p
WHERE p.tenant_id = current_tenant_id()
  AND p.is_active = TRUE
  AND p.deleted_at IS NULL
  AND (
    p.target->'resources' ? $1::text OR p.target->'actions' ? $2::text
  )
ORDER BY p.priority DESC, p.created_at ASC;

-- ABAC-specific queries for policy evaluation

-- name: GetPoliciesForEvaluation :many
SELECT * FROM policies 
WHERE tenant_id = current_tenant_id()
  AND is_active = true 
  AND deleted_at IS NULL
  AND (
    $1::UUID IS NULL OR 
    entity_id IS NULL OR 
    entity_id = $1
  )
  AND (
    target->>'resource_type' = $2 OR 
    target->>'resource_type' = '*' OR 
    target->>'resource_type' IS NULL
  )
  AND (
    target->>'action' = $3 OR 
    target->>'action' = '*' OR 
    target->>'action' IS NULL
  )
ORDER BY priority DESC, created_at ASC;

-- name: GetPoliciesByIDs :many
SELECT * FROM policies 
WHERE id = ANY($1::UUID[]) 
  AND tenant_id = current_tenant_id();

-- name: CountPolicies :one
SELECT COUNT(*) FROM policies 
WHERE tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND ($1::VARCHAR IS NULL OR name ILIKE '%' || $1 || '%')
  AND ($2::VARCHAR IS NULL OR category = $2)
  AND ($3::BOOLEAN IS NULL OR is_active = $3);

-- name: GetPoliciesByEntityID :many
SELECT * FROM policies 
WHERE tenant_id = current_tenant_id()
  AND (entity_id = $1 OR entity_id IS NULL)
  AND is_active = true
  AND deleted_at IS NULL
ORDER BY priority DESC, created_at ASC;

-- name: UpdatePolicyStatus :exec
UPDATE policies 
SET is_active = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();
