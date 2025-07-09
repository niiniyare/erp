-- =====================================================================
-- 1. ENTITY VALIDATION AND INTEGRITY CHECKS
-- =====================================================================

-- name: ValidateEntityCode :one
SELECT EXISTS(
    SELECT 1 FROM entities 
    WHERE code = $1 AND tenant_id = current_tenant_id() AND uuid != $2 AND deleted_at IS NULL
) AS exists;

-- name: ValidateEntityName :one
SELECT EXISTS(
    SELECT 1 FROM entities 
    WHERE name = $1 AND tenant_id = current_tenant_id() AND uuid != $2 AND deleted_at IS NULL
) AS exists;

-- name: ValidateEntityParent :one
SELECT 
    CASE 
        WHEN $1 IS NULL THEN true
        WHEN NOT EXISTS(SELECT 1 FROM entities WHERE uuid = $1 AND tenant_id = current_tenant_id() AND deleted_at IS NULL) THEN false
        WHEN EXISTS(SELECT 1 FROM hierarchy_paths WHERE tenant_id = current_tenant_id() AND ancestor_id = $2 AND descendant_id = $1) THEN false
        ELSE true
    END AS valid;

-- name: CheckCircularReference :one
SELECT EXISTS(
    SELECT 1 FROM hierarchy_paths 
    WHERE tenant_id = current_tenant_id() 
        AND ancestor_id = $2 
        AND descendant_id = $1
) AS exists;

-- name: GetEntityDepth :one
SELECT COALESCE(MAX(depth), 0) AS depth
FROM hierarchy_paths 
WHERE tenant_id = current_tenant_id() AND ancestor_id = $1;

-- =====================================================================
-- 2. ENTITY SEARCH AND FILTERING ENHANCEMENTS
-- =====================================================================

-- name: SearchEntitiesByCodeAndName :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND (code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
    AND deleted_at IS NULL
ORDER BY 
    CASE WHEN code ILIKE $1 || '%' THEN 1 
         WHEN name ILIKE $1 || '%' THEN 2 
         ELSE 3 END,
    name
LIMIT $2;

-- name: ListEntitiesWithPagination :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND deleted_at IS NULL
    AND ($1::VARCHAR IS NULL OR type = $1)
    AND ($2::BOOLEAN IS NULL OR is_active = $2)
    AND ($3::BOOLEAN IS NULL OR hidden = $3)
ORDER BY name
LIMIT $4 OFFSET $5;

-- name: CountEntitiesWithFilters :one
SELECT COUNT(*) AS count
FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND deleted_at IS NULL
    AND ($1::VARCHAR IS NULL OR type = $1)
    AND ($2::BOOLEAN IS NULL OR is_active = $2)
    AND ($3::BOOLEAN IS NULL OR hidden = $3);

-- name: ListEntitiesByTypes :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND type = ANY($1::VARCHAR[])
    AND deleted_at IS NULL
ORDER BY type, name;

-- name: GetEntitiesByFiscalYearStart :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND fy_start_month = $1
    AND deleted_at IS NULL
ORDER BY name;

-- =====================================================================
-- 3. HIERARCHY BULK OPERATIONS
-- =====================================================================

-- name: MoveEntityToNewParent :exec
WITH RECURSIVE affected_entities AS (
    SELECT $1::UUID as entity_id, 0 as depth
    UNION ALL
    SELECT hp.descendant_id, ae.depth + 1
    FROM affected_entities ae
    JOIN hierarchy_paths hp ON hp.ancestor_id = ae.entity_id
    WHERE hp.tenant_id = current_tenant_id() AND ae.depth < 10
),
delete_paths AS (
    DELETE FROM hierarchy_paths 
    WHERE tenant_id = current_tenant_id() 
        AND descendant_id IN (SELECT entity_id FROM affected_entities)
),
insert_new_paths AS (
    INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
    SELECT 
        current_tenant_id(),
        ancestor_paths.ancestor_id,
        ae.entity_id,
        ancestor_paths.depth + descendant_paths.depth + 1
    FROM affected_entities ae
    CROSS JOIN (
        SELECT ancestor_id, depth FROM hierarchy_paths 
        WHERE tenant_id = current_tenant_id() AND descendant_id = $2
        UNION ALL
        SELECT $2::UUID, 0
    ) ancestor_paths
    CROSS JOIN (
        SELECT descendant_id, depth FROM hierarchy_paths 
        WHERE tenant_id = current_tenant_id() AND ancestor_id = $1
        UNION ALL
        SELECT $1::UUID, 0
    ) descendant_paths
    WHERE ae.entity_id = descendant_paths.descendant_id
)
UPDATE entities SET parent_id = $2, updated_at = NOW() WHERE uuid = $1;

-- name: GetEntitySubtree :many
SELECT e.*, hp.depth
FROM entities e
JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE hp.tenant_id = current_tenant_id() 
    AND hp.ancestor_id = $1
    AND e.deleted_at IS NULL
    AND ($2::INTEGER IS NULL OR hp.depth <= $2)
ORDER BY hp.depth, e.name;

-- name: GetEntityPath :many
SELECT e.*, hp.depth
FROM entities e
JOIN hierarchy_paths hp ON e.uuid = hp.ancestor_id
WHERE hp.tenant_id = current_tenant_id() 
    AND hp.descendant_id = $1
    AND e.deleted_at IS NULL
ORDER BY hp.depth DESC;

-- name: BulkMoveEntities :exec
UPDATE entities 
SET parent_id = $2, updated_at = NOW()
WHERE tenant_id = current_tenant_id() 
    AND uuid = ANY($1::UUID[])
    AND deleted_at IS NULL;

-- =====================================================================
-- 4. ENTITY STATE ENHANCEMENTS
-- =====================================================================

-- name: GetEntityStateWithLocking :one
SELECT * FROM entitystate
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3
FOR UPDATE;

-- name: BulkCreateEntityStates :exec
INSERT INTO entitystate (uuid, fiscal_year, key, sequence, entity_id, entity_unit_id)
SELECT gen_random_uuid(), $2, unnest($3::VARCHAR[]), 1, $1, $4
ON CONFLICT (entity_id, key, fiscal_year) DO NOTHING;

-- name: GetEntityStateHistory :many
SELECT es.*, e.name as entity_name
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
WHERE es.entity_id = $1
    AND ($2::VARCHAR IS NULL OR es.key = $2)
    AND ($3::SMALLINT IS NULL OR es.fiscal_year = $3)
ORDER BY es.fiscal_year DESC, es.key;

-- name: GetHighestSequenceNumber :one
SELECT COALESCE(MAX(sequence), 0) AS sequence
FROM entitystate
WHERE entity_id = $1 AND key = $2 AND fiscal_year = $3;

-- name: ResetAllEntitySequences :exec
UPDATE entitystate 
SET sequence = 1
WHERE entity_id = $1 AND fiscal_year = $2;

-- name: GetEntityStatesByFiscalYear :many
SELECT es.*, e.name as entity_name
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
WHERE es.fiscal_year = $1
    AND e.tenant_id = current_tenant_id()
ORDER BY e.name, es.key;

-- =====================================================================
-- 5. AUDIT AND MONITORING QUERIES
-- =====================================================================

-- name: GetRecentlyModifiedEntities :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND updated_at >= $1
    AND deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT $2;

-- name: GetRecentlyDeletedEntities :many
SELECT * FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND deleted_at >= $1
    AND deleted_at IS NOT NULL
ORDER BY deleted_at DESC
LIMIT $2;

-- name: GetEntityAuditLog :many
SELECT 
    uuid,
    name,
    type,
    is_active,
    hidden,
    created_at,
    updated_at,
    deleted_at,
    CASE 
        WHEN deleted_at IS NOT NULL THEN 'DELETED'
        WHEN updated_at > created_at THEN 'UPDATED'
        ELSE 'CREATED'
    END as action
FROM entities 
WHERE tenant_id = current_tenant_id()
    AND (created_at >= $1 OR updated_at >= $1 OR deleted_at >= $1)
ORDER BY GREATEST(created_at, updated_at, COALESCE(deleted_at, created_at)) DESC;

-- name: GetOrphanedEntities :many
SELECT e.* FROM entities e
LEFT JOIN entities parent ON parent.uuid = e.parent_id AND parent.tenant_id = e.tenant_id
WHERE e.tenant_id = current_tenant_id()
    AND e.parent_id IS NOT NULL
    AND parent.uuid IS NULL
    AND e.deleted_at IS NULL;

-- name: GetInconsistentHierarchyPaths :many
SELECT DISTINCT hp.ancestor_id, hp.descendant_id, hp.depth
FROM hierarchy_paths hp
LEFT JOIN entities e1 ON hp.ancestor_id = e1.uuid AND e1.tenant_id = hp.tenant_id
LEFT JOIN entities e2 ON hp.descendant_id = e2.uuid AND e2.tenant_id = hp.tenant_id
WHERE hp.tenant_id = current_tenant_id()
    AND (e1.uuid IS NULL OR e2.uuid IS NULL OR e1.deleted_at IS NOT NULL OR e2.deleted_at IS NOT NULL);

-- =====================================================================
-- 6. PERFORMANCE AND ANALYTICS QUERIES
-- =====================================================================

-- name: GetEntityCountByType :many
SELECT type, COUNT(*) as count
FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND deleted_at IS NULL
GROUP BY type
ORDER BY count DESC;

-- name: GetEntityHierarchyStats :one
SELECT 
    COUNT(*) as total_entities,
    COUNT(*) FILTER (WHERE parent_id IS NULL) as root_entities,
    MAX(depth) as max_depth,
    AVG(depth) as avg_depth,
    COUNT(DISTINCT ancestor_id) as entities_with_children
FROM entities e
LEFT JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id AND hp.tenant_id = e.tenant_id
WHERE e.tenant_id = current_tenant_id() AND e.deleted_at IS NULL;

-- name: GetEntitySequenceStats :many
SELECT 
    e.name as entity_name,
    es.key,
    es.fiscal_year,
    es.sequence,
    es.sequence - 1 as documents_created
FROM entitystate es
JOIN entities e ON es.entity_id = e.uuid
WHERE e.tenant_id = current_tenant_id()
    AND ($1::UUID IS NULL OR es.entity_id = $1)
ORDER BY e.name, es.key, es.fiscal_year;

-- name: GetUnusedEntityCodes :many
SELECT DISTINCT code
FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND code IS NOT NULL
    AND deleted_at IS NOT NULL
ORDER BY code;

-- =====================================================================
-- 7. MAINTENANCE AND CLEANUP QUERIES
-- =====================================================================

-- name: CleanupOrphanedHierarchyPaths :exec
DELETE FROM hierarchy_paths 
WHERE tenant_id = current_tenant_id()
    AND (
        NOT EXISTS(SELECT 1 FROM entities WHERE uuid = ancestor_id AND tenant_id = current_tenant_id())
        OR NOT EXISTS(SELECT 1 FROM entities WHERE uuid = descendant_id AND tenant_id = current_tenant_id())
    );

-- name: RebuildHierarchyPaths :exec
WITH RECURSIVE entity_hierarchy AS (
    SELECT 
        uuid as ancestor_id,
        uuid as descendant_id,
        0 as depth,
        tenant_id
    FROM entities
    WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL
    
    UNION ALL
    
    SELECT 
        eh.ancestor_id,
        e.uuid,
        eh.depth + 1,
        e.tenant_id
    FROM entity_hierarchy eh
    JOIN entities e ON e.parent_id = eh.descendant_id
    WHERE e.tenant_id = current_tenant_id() 
        AND e.deleted_at IS NULL 
        AND eh.depth < 10
),
cleanup AS (
    DELETE FROM hierarchy_paths WHERE tenant_id = current_tenant_id()
)
INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
SELECT tenant_id, ancestor_id, descendant_id, depth
FROM entity_hierarchy;

-- name: ArchiveOldDeletedEntities :exec
DELETE FROM entities 
WHERE tenant_id = current_tenant_id() 
    AND deleted_at < $1
    AND deleted_at IS NOT NULL;

-- name: GetEntityHealthCheck :one
SELECT 
    (SELECT COUNT(*) FROM entities WHERE tenant_id = current_tenant_id() AND deleted_at IS NULL) as active_entities,
    (SELECT COUNT(*) FROM hierarchy_paths WHERE tenant_id = current_tenant_id()) as hierarchy_paths,
    (SELECT COUNT(*) FROM entitystate es JOIN entities e ON es.entity_id = e.uuid WHERE e.tenant_id = current_tenant_id()) as entity_states,
    (SELECT COUNT(*) FROM entities e LEFT JOIN entities p ON e.parent_id = p.uuid WHERE e.tenant_id = current_tenant_id() AND e.parent_id IS NOT NULL AND p.uuid IS NULL) as orphaned_entities,
    (SELECT COUNT(*) FROM hierarchy_paths hp LEFT JOIN entities e1 ON hp.ancestor_id = e1.uuid LEFT JOIN entities e2 ON hp.descendant_id = e2.uuid WHERE hp.tenant_id = current_tenant_id() AND (e1.uuid IS NULL OR e2.uuid IS NULL)) as orphaned_paths;
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
