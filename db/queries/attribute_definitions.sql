-- Attribute Definitions CRUD Operations

-- name: CreateAttributeDefinition :one
INSERT INTO attribute_definitions (
    id, tenant_id, name, display_name, description, data_type, category,
    is_required, is_sensitive, default_value, allowed_values, validation_rules, encryption_required, is_active
) VALUES (
    $1, current_tenant_id(), $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetAttributeDefinition :one
SELECT * FROM attribute_definitions
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetAttributeDefinitionByName :one
SELECT * FROM attribute_definitions
WHERE name = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: UpdateAttributeDefinition :one
UPDATE attribute_definitions
SET
    name = COALESCE($2, name),
    display_name = COALESCE($3, display_name),
    description = COALESCE($4, description),
    data_type = COALESCE($5, data_type),
    category = COALESCE($6, category),
    is_required = COALESCE($7, is_required),
    is_sensitive = COALESCE($8, is_sensitive),
    default_value = COALESCE($9, default_value),
    allowed_values = COALESCE($10, allowed_values),
    validation_rules = COALESCE($11, validation_rules),
    encryption_required = COALESCE($12, encryption_required),
    is_active = COALESCE($13, is_active),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAttributeDefinition :exec
DELETE FROM attribute_definitions
WHERE id = $1 AND tenant_id = current_tenant_id();

-- Attribute Definition Listing and Filtering

-- name: ListAttributeDefinitions :many
SELECT * FROM attribute_definitions
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
ORDER BY name;

-- name: ListAttributeDefinitionsByCategory :many
SELECT * FROM attribute_definitions
WHERE tenant_id = current_tenant_id() AND category = $1 AND deleted_at IS NULL
ORDER BY name;
