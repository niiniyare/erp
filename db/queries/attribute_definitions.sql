-- Attribute Definitions CRUD Operations

-- name: CreateAttributeDefinition :one
INSERT INTO attribute_definitions (
    tenant_id, name, display_name, description, data_type, category,
    is_required, is_sensitive, default_value, allowed_values, validation_rules, encryption_required, is_active
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetAttributeDefinition :one
SELECT * FROM attribute_definitions
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetAttributeDefinitionByName :one
SELECT * FROM attribute_definitions
WHERE name = $1 AND tenant_id = current_tenant_id();

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
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: DeleteAttributeDefinition :exec
DELETE FROM attribute_definitions
WHERE id = $1 AND tenant_id = current_tenant_id();

-- Attribute Definition Listing and Filtering

-- name: ListAttributeDefinitions :many
SELECT * FROM attribute_definitions
WHERE tenant_id = current_tenant_id()
  AND ($1::VARCHAR IS NULL OR name ILIKE '%' || $1 || '%')
  AND ($2::VARCHAR IS NULL OR category = $2)
  AND ($3::BOOLEAN IS NULL OR is_active = $3)
ORDER BY name ASC
LIMIT $4 OFFSET $5;

-- name: ListAttributeDefinitionsByCategory :many
SELECT * FROM attribute_definitions
WHERE tenant_id = current_tenant_id() AND category = $1 AND is_active = true
ORDER BY name;

-- name: CountAttributeDefinitions :one
SELECT COUNT(*) FROM attribute_definitions 
WHERE tenant_id = current_tenant_id()
  AND ($1::VARCHAR IS NULL OR name ILIKE '%' || $1 || '%')
  AND ($2::VARCHAR IS NULL OR category = $2)
  AND ($3::BOOLEAN IS NULL OR is_active = $3);

-- Attribute Values Operations
-- name: CreateAttributeValue :one
INSERT INTO attribute_values (
    tenant_id, definition_id, entity_id, value, encrypted_value, is_encrypted, 
    version, effective_from, effective_to, created_by
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, 
    $6, $7, $8, $9
) RETURNING *;

-- name: GetAttributeValue :one
SELECT av.*, ad.name as attribute_name, ad.data_type, ad.category
FROM attribute_values av
JOIN attribute_definitions ad ON av.definition_id = ad.id
WHERE av.id = $1 
  AND av.tenant_id = current_tenant_id();

-- name: GetAttributeValuesByEntity :many
SELECT av.*, ad.name as attribute_name, ad.data_type, ad.category
FROM attribute_values av
JOIN attribute_definitions ad ON av.definition_id = ad.id
WHERE av.entity_id = $1 
  AND av.tenant_id = current_tenant_id()
  AND ($2::VARCHAR IS NULL OR ad.category = $2)
  AND (av.effective_to IS NULL OR av.effective_to > NOW())
ORDER BY ad.name ASC;

-- name: GetAttributeValueByEntityAndName :one
SELECT av.*, ad.name as attribute_name, ad.data_type, ad.category
FROM attribute_values av
JOIN attribute_definitions ad ON av.definition_id = ad.id
WHERE av.entity_id = $1 
  AND ad.name = $2
  AND av.tenant_id = current_tenant_id()
  AND (av.effective_to IS NULL OR av.effective_to > NOW())
ORDER BY av.effective_from DESC
LIMIT 1;

-- name: UpdateAttributeValue :one
UPDATE attribute_values 
SET 
    value = COALESCE($2, value),
    encrypted_value = COALESCE($3, encrypted_value),
    is_encrypted = COALESCE($4, is_encrypted),
    version = version + 1,
    effective_from = COALESCE($5, effective_from),
    effective_to = COALESCE($6, effective_to),
    updated_at = NOW(),
    updated_by = $7
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;

-- name: ExpireAttributeValue :exec
UPDATE attribute_values 
SET effective_to = NOW(), updated_at = NOW(), updated_by = $2
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: DeleteAttributeValue :exec
DELETE FROM attribute_values 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetAttributeStats :one
SELECT 
    COUNT(DISTINCT ad.id) as total_definitions,
    COUNT(DISTINCT av.id) as total_values,
    COUNT(DISTINCT av.entity_id) as entities_with_attributes,
    COUNT(DISTINCT CASE WHEN ad.category = 'USER' THEN av.id END) as user_attributes,
    COUNT(DISTINCT CASE WHEN ad.category = 'RESOURCE' THEN av.id END) as resource_attributes,
    COUNT(DISTINCT CASE WHEN ad.category = 'ENVIRONMENT' THEN av.id END) as environment_attributes
FROM attribute_definitions ad
LEFT JOIN attribute_values av ON ad.id = av.definition_id 
    AND av.tenant_id = current_tenant_id()
    AND (av.effective_to IS NULL OR av.effective_to > NOW())
WHERE ad.tenant_id = current_tenant_id()
  AND ad.is_active = true;
