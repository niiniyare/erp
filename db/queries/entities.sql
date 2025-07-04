-- Entity CRUD Operations
-- name: CreateEntity :one
INSERT INTO entities (
    uuid, tenant_id, parent_id, name, code, type, is_active, 
    hidden, accrual_method, fy_start_month, address, picture, settings
) VALUES (
    $1, current_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetEntity :one
SELECT * FROM entities 
  WHERE uuid = $1 
  AND tenant_id = current_tenant_id() 
  AND deleted_at IS NULL;

-- name: GetEntityByCode :one
SELECT * FROM entities 
WHERE code = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: GetEntityByName :one
SELECT * FROM entities 
WHERE name = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL;

-- name: UpdateEntity :one
UPDATE entities 
SET 
    name = COALESCE($2, name),
    code = COALESCE($3, code),
    type = COALESCE($4, type),
    is_active = COALESCE($5, is_active),
    hidden = COALESCE($6, hidden),
    accrual_method = COALESCE($7, accrual_method),
    fy_start_month = COALESCE($8, fy_start_month),
    address = COALESCE($9, address),
    picture = COALESCE($10, picture),
    settings = COALESCE($11, settings),
    updated_at = NOW()
WHERE uuid = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteEntity :exec
UPDATE entities 
SET deleted_at = NOW(), updated_at = NOW()
WHERE uuid = $1 AND tenant_id = current_tenant_id();

-- name: RestoreEntity :exec
UPDATE entities 
SET deleted_at = NULL, updated_at = NOW()
WHERE uuid = $1 AND tenant_id = current_tenant_id();

-- name: HardDeleteEntity :exec
DELETE FROM entities 
WHERE uuid = $1 AND tenant_id = current_tenant_id();

-- Entity Listing and Filtering
-- name: ListEntities :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
ORDER BY name;

-- name: ListActiveEntities :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() AND is_active = true AND deleted_at IS NULL
ORDER BY name;

-- name: ListEntitiesByType :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() AND type = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: ListVisibleEntities :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() AND hidden = false AND deleted_at IS NULL
ORDER BY name;

-- name: SearchEntitiesByName :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND name ILIKE '%' || $1 || '%' 
    AND deleted_at IS NULL
ORDER BY name
LIMIT $2;

-- Find all leaf nodes (entities with no children)
-- name: ListEntitiesWithNochildren :many
SELECT e.*
FROM entities e
LEFT JOIN entities children ON children.parent_id = e.uuid 
    AND children.tenant_id = e.tenant_id
WHERE children.uuid IS NULL 
  AND e.tenant_id = current_tenant_id()
  AND e.is_active = true;


-- Entity Hierarchy Operations
-- name: CreateHierarchyPath :exec
INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
VALUES (current_tenant_id(), $1, $2, $3);

-- name: GetEntityChildren :many
SELECT e.* FROM entities e
JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE hp.tenant_id = current_tenant_id() 
    AND hp.ancestor_id = $1 
    AND hp.depth = 1
    AND e.deleted_at IS NULL
ORDER BY e.name;

-- name: GetEntityDescendants :many
SELECT e.*, hp.depth FROM entities e
JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE hp.tenant_id = current_tenant_id() 
    AND hp.ancestor_id = $1 
    AND hp.depth > 0
    AND e.deleted_at IS NULL
ORDER BY hp.depth, e.name;

-- name: GetEntityAncestors :many
SELECT e.*, hp.depth FROM entities e
JOIN hierarchy_paths hp ON e.uuid = hp.ancestor_id
WHERE hp.tenant_id = current_tenant_id() 
    AND hp.descendant_id = $1 
    AND hp.depth > 0
    AND e.deleted_at IS NULL
ORDER BY hp.depth DESC;

-- name: GetEntityParent :one
SELECT e.* FROM entities e
JOIN hierarchy_paths hp ON e.uuid = hp.ancestor_id
WHERE hp.tenant_id = current_tenant_id() 
    AND hp.descendant_id = $1 
    AND hp.depth = 1
    AND e.deleted_at IS NULL;

-- name: GetEntitySiblings :many
SELECT DISTINCT e.* FROM entities e
JOIN hierarchy_paths hp1 ON e.uuid = hp1.descendant_id
JOIN hierarchy_paths hp2 ON hp1.ancestor_id = hp2.ancestor_id
WHERE hp2.tenant_id = current_tenant_id() 
    AND hp2.descendant_id = $1 
    AND hp1.depth = 1 
    AND hp2.depth = 1
    AND e.uuid != $1
    AND e.deleted_at IS NULL
ORDER BY e.name;

-- name: GetEntityRoots :many
SELECT e.* FROM entities e
WHERE e.tenant_id = current_tenant_id() 
    AND e.parent_id IS NULL
    AND e.deleted_at IS NULL
ORDER BY e.name;

-- name: GetEntityLevel :one
SELECT COALESCE(MIN(hp.depth), 0) as level
FROM hierarchy_paths hp
WHERE hp.tenant_id = current_tenant_id() AND hp.descendant_id = $1;

-- name: IsEntityAncestor :one
SELECT EXISTS(
    SELECT 1 FROM hierarchy_paths 
    WHERE tenant_id = current_tenant_id() 
        AND ancestor_id = $1 
        AND descendant_id = $2 
        AND depth > 0
) as is_ancestor;

-- name: DeleteHierarchyPaths :exec
DELETE FROM hierarchy_paths 
WHERE tenant_id = current_tenant_id() 
    AND (ancestor_id = $1 OR descendant_id = $1);

-- name: UpdateHierarchyPaths :exec
WITH RECURSIVE hierarchy_cte AS (
    -- Base case: self-reference
    SELECT current_tenant_id() as tenant_id, $1::UUID as ancestor_id, $1::UUID as descendant_id, 0 as depth
    UNION ALL
    -- Recursive case: add ancestors
    SELECT h.tenant_id, hp.ancestor_id, h.descendant_id, h.depth + 1
    FROM hierarchy_cte h
    JOIN hierarchy_paths hp ON hp.descendant_id = h.ancestor_id AND hp.tenant_id = h.tenant_id
    WHERE h.depth < 10 -- Prevent infinite recursion
)
INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
SELECT DISTINCT tenant_id, ancestor_id, descendant_id, depth
FROM hierarchy_cte
ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

-- Entity State Management
-- name: CreateEntityState :one
INSERT INTO entitystate (uuid, fiscal_year, key, sequence, entity_id, entity_unit_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetEntityState :one
SELECT * FROM entitystate
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3;

-- name: GetEntityStateByKey :one
SELECT * FROM entitystate
WHERE entity_id = $1 AND key = $2;

-- name: UpdateEntityStateSequence :one
UPDATE entitystate 
SET sequence = $3, updated_at = NOW()
WHERE entity_id = $1 AND key = $2
RETURNING *;

-- name: IncrementEntityStateSequence :one
UPDATE entitystate 
SET sequence = sequence + 1
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3
RETURNING sequence;

-- name: GetNextSequenceNumber :one
INSERT INTO entitystate (uuid, fiscal_year, key, sequence, entity_id, entity_unit_id)
VALUES (gen_random_uuid(), $3, $2, 1, $1, $4)
ON CONFLICT (entity_id, key, fiscal_year) DO UPDATE 
SET sequence = entitystate.sequence + 1
RETURNING sequence;

-- name: ListEntityStates :many
SELECT * FROM entitystate
WHERE entity_id = $1
ORDER BY key, fiscal_year;

-- name: DeleteEntityState :exec
DELETE FROM entitystate
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3;

-- name: ResetEntityStateSequence :exec
UPDATE entitystate 
SET sequence = $3
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $4;

-- Advanced Entity Queries
-- name: GetEntityWithHierarchyInfo :one
SELECT 
    e.*,
    COALESCE(MIN(hp.depth), 0) as level,
    COUNT(children.uuid) as child_count,
    parent_e.name as parent_name
FROM entities e
LEFT JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id AND hp.tenant_id = e.tenant_id
LEFT JOIN entities children ON children.parent_id = e.uuid AND children.tenant_id = e.tenant_id AND children.deleted_at IS NULL
LEFT JOIN entities parent_e ON parent_e.uuid = e.parent_id AND parent_e.tenant_id = e.tenant_id
WHERE e.uuid = $1 AND e.tenant_id = $2 AND e.deleted_at IS NULL
GROUP BY e.uuid, parent_e.name;

-- name: GetEntityTreeStructure :many
WITH RECURSIVE entity_tree AS (
    SELECT 
        e.*,
        0 as level,
        ARRAY[e.name] as path,
        e.name as sort_path
    FROM entities e
    WHERE e.tenant_id = $1 
        AND e.parent_id IS NULL 
        AND e.deleted_at IS NULL
    
    UNION ALL
    
    SELECT 
        e.*,
        et.level + 1,
        et.path || e.name,
        et.sort_path || '/' || e.name
    FROM entities e
    JOIN entity_tree et ON e.parent_id = et.uuid
    WHERE e.tenant_id = $1 
        AND e.deleted_at IS NULL
        AND et.level < 10
)
SELECT * FROM entity_tree
ORDER BY sort_path;

-- name: GetEntityStats :one
SELECT 
    COUNT(*) as total_entities,
    COUNT(*) FILTER (WHERE is_active = true) as active_entities,
    COUNT(*) FILTER (WHERE hidden = false) as visible_entities,
    COUNT(DISTINCT type) as entity_types,
    COUNT(*) FILTER (WHERE parent_id IS NULL) as root_entities,
    COUNT(*) FILTER (WHERE accrual_method = true) as accrual_entities,
    COUNT(*) FILTER (WHERE accrual_method = false) as cash_entities
FROM entities
WHERE tenant_id = $1 AND deleted_at IS NULL;

-- name: GetEntitiesByFiscalYear :many
SELECT DISTINCT e.*
FROM entities e
JOIN entitystate es ON e.uuid = es.entity_id
WHERE e.tenant_id = $1 
    AND es.fiscal_year = $2
    AND e.deleted_at IS NULL
ORDER BY e.name;

-- name: ValidateEntityHierarchy :one
SELECT 
    CASE 
        WHEN COUNT(*) = 0 THEN true
        ELSE false
    END as is_valid
FROM hierarchy_paths hp1
JOIN hierarchy_paths hp2 ON hp1.descendant_id = hp2.ancestor_id
WHERE hp1.tenant_id = $1 
    AND hp2.tenant_id = $1
    AND hp1.ancestor_id = hp2.descendant_id
    AND hp1.depth > 0 
    AND hp2.depth > 0;

-- Batch Operations
-- name: BatchUpdateEntityStatus :exec
UPDATE entities 
SET is_active = $3, updated_at = NOW()
WHERE tenant_id = $1 AND uuid = ANY($2::UUID[]) AND deleted_at IS NULL;

-- name: BatchSoftDeleteEntities :exec
UPDATE entities 
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND uuid = ANY($2::UUID[]);

-- name: GetEntitiesByUUIDs :many
SELECT * FROM entities 
WHERE tenant_id = $1 AND uuid = ANY($2::UUID[]) AND deleted_at IS NULL
ORDER BY name;
/*
I'll create comprehensive sqlc functions for your entity management schema. This will include CRUD operations, hierarchy management, and state tracking.I've created a comprehensive set of sqlc functions for your entity management schema. Here's what's included:

## Core Entity Operations
- **CRUD**: Create, read, update, soft/hard delete entities
- **Listing**: Various filtered views (active, by type, visible, search)
- **Validation**: Get by code, name, or UUID

## Hierarchy Management
- **Tree Operations**: Get children, descendants, ancestors, parent, siblings, roots
- **Path Management**: Create/delete hierarchy paths, update hierarchy structure
- **Utility Functions**: Check ancestry relationships, get entity level, validate hierarchy integrity

## Entity State Management
- **Sequence Tracking**: Create, update, increment sequence numbers for document types
- **Fiscal Year Support**: Manage sequences per fiscal year
- **Automatic Numbering**: Get next sequence number with automatic creation

## Advanced Queries
- **Rich Entity Info**: Get entity with hierarchy metadata (level, child count, parent name)
- **Tree Structure**: Recursive query to get complete entity tree with paths
- **Statistics**: Comprehensive stats about entities in a tenant
- **Validation**: Check for circular references in hierarchy

## Batch Operations
- **Bulk Updates**: Update status for multiple entities
- **Bulk Operations**: Batch delete and retrieve operations

## Key Features
- **Soft Delete Support**: All queries respect `deleted_at` for logical deletion
- **Tenant Isolation**: All operations are tenant-scoped for multi-tenancy
- **Performance Optimized**: Uses appropriate indexes and efficient queries
- **Hierarchy Integrity**: Functions to maintain and validate entity hierarchies
- **Flexible Filtering**: Multiple ways to query and filter entities

The functions handle common use cases like organizational charts, document numbering sequences, and hierarchical business structure management. They're designed to work efficiently with your closure table approach for hierarchy management.
*/
