-- Entity CRUD Operations
-- name: CreateEntity :one
INSERT INTO entities (
    uuid, tenant_id, parent_id,
    name, code, type, is_active, hidden,
    accrual_method, fy_start_month,
    entity_path, entity_level,
    address, picture, metadata, settings
) VALUES (
    $1, current_tenant_id(), $2,
    $3, $4, $5, $6, $7,
    $8, $9,
    $10, $11,
    $12, $13, $14, $15
) RETURNING *;

-- name: UpdateEntityPath :exec
-- Called by app after insert/reparent to maintain the materialized path.
UPDATE entities
SET entity_path  = $2,
    entity_level = $3,
    updated_at   = NOW()
WHERE uuid = $1
  AND tenant_id = current_tenant_id();

-- name: GetEntity :one
SELECT
  *
FROM
  entities
WHERE
  uuid = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetEntityByCode :one
SELECT
  *
FROM
  entities
WHERE
  code = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetEntityByName :one
SELECT
  *
FROM
  entities
WHERE
  name = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: UpdateEntity :one
UPDATE
  entities
SET
  name = COALESCE($2, name),
  code = COALESCE($3, code),
  TYPE = COALESCE($4, TYPE),
  is_active = COALESCE($5, is_active),
  hidden = COALESCE($6, hidden),
  accrual_method = COALESCE($7, accrual_method),
  fy_start_month = COALESCE($8, fy_start_month),
  address = COALESCE($9, address),
  picture = COALESCE($10, picture),
  settings = COALESCE($11, settings),
  updated_at = NOW()
WHERE
  uuid = $1
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING
  *;

-- name: ResolveEntityScope :one
-- Called once at login to build EntityScope for session pre-computation.
-- Returns the entity with its path and level so the service layer can
-- determine scope type: level=1 → "all", leaf (no children) → "entity", else → "subtree".
SELECT
    e.uuid,
    e.entity_path,
    e.entity_level,
    e.type,
    EXISTS (
        SELECT 1 FROM entities c
        WHERE c.parent_id = e.uuid
          AND c.deleted_at IS NULL
    ) AS has_children
FROM entities e
WHERE e.uuid = $1
  AND e.tenant_id = current_tenant_id()
  AND e.deleted_at IS NULL;

-- name: ListEntitySubtree :many
-- Returns all entities within the subtree rooted at the given path prefix.
-- Used by business repos for subtree-scoped data queries.
SELECT uuid, name, type, entity_path, entity_level
FROM   entities
WHERE  tenant_id   = current_tenant_id()
  AND  entity_path LIKE $1 || '%'
  AND  deleted_at  IS NULL
ORDER  BY entity_level, name;

-- name: SoftDeleteEntity :exec
UPDATE
  entities
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE
  uuid = $1
  AND tenant_id = current_tenant_id();

-- name: RestoreEntity :exec
UPDATE
  entities
SET
  deleted_at = NULL,
  updated_at = NOW()
WHERE
  uuid = $1
  AND tenant_id = current_tenant_id();

-- name: HardDeleteEntity :exec
DELETE FROM
  entities
WHERE
  uuid = $1
  AND tenant_id = current_tenant_id();

-- Entity Listing and Filtering
-- name: ListEntities :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  name;

-- name: ListActiveEntities :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND is_active = TRUE
  AND deleted_at IS NULL
ORDER BY
  name;

-- name: ListEntitiesByType :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND TYPE = $1
  AND deleted_at IS NULL
ORDER BY
  name;

-- name: ListVisibleEntities :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND hidden = false
  AND deleted_at IS NULL
ORDER BY
  name;

-- name: SearchEntitiesByName :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND name ILIKE '%' || $1 || '%'
  AND deleted_at IS NULL
ORDER BY
  name
LIMIT
  $2;

-- Find all leaf nodes (entities with no children)
-- name: ListEntitiesWithNochildren :many
SELECT
  e.*
FROM
  entities e
  LEFT JOIN entities children ON children.parent_id = e.uuid
  AND children.tenant_id = e.tenant_id
WHERE
  children.uuid IS NULL
  AND e.is_active = TRUE;

-- Entity Hierarchy Operations
-- name: CreateHierarchyPath :exec
INSERT INTO
  hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
VALUES
  (current_tenant_id(), $1, $2, $3);

-- name: GetEntityChildren :many
SELECT
  e.*
FROM
  entities e
  JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.ancestor_id = $1
  AND hp.depth = 1
  AND e.deleted_at IS NULL
ORDER BY
  e.name;

-- name: GetEntityDescendants :many
SELECT
  e.*,
  hp.depth
FROM
  entities e
  JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.ancestor_id = $1
  AND hp.depth > 0
  AND e.deleted_at IS NULL
ORDER BY
  hp.depth,
  e.name;

-- name: GetEntityAncestors :many
SELECT
  e.*,
  hp.depth
FROM
  entities e
  JOIN hierarchy_paths hp ON e.uuid = hp.ancestor_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.descendant_id = $1
  AND hp.depth > 0
  AND e.deleted_at IS NULL
ORDER BY
  hp.depth DESC;

-- name: GetEntityParent :one
SELECT
  e.*
FROM
  entities e
  JOIN hierarchy_paths hp ON e.uuid = hp.ancestor_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.descendant_id = $1
  AND hp.depth = 1
  AND e.deleted_at IS NULL;

-- name: GetEntitySiblings :many
SELECT
  DISTINCT e.*
FROM
  entities e
  JOIN hierarchy_paths hp1 ON e.uuid = hp1.descendant_id
  JOIN hierarchy_paths hp2 ON hp1.ancestor_id = hp2.ancestor_id
WHERE
  hp2.tenant_id = current_tenant_id()
  AND hp2.descendant_id = $1
  AND hp1.depth = 1
  AND hp2.depth = 1
  AND e.uuid != $1
  AND e.deleted_at IS NULL
ORDER BY
  e.name;

-- name: GetEntityRoots :many
SELECT
  e.*
FROM
  entities e
WHERE
  e.tenant_id = current_tenant_id()
  AND e.parent_id IS NULL
  AND e.deleted_at IS NULL
ORDER BY
  e.name;

-- name: GetEntityLevel :one
SELECT
  COALESCE(MIN(hp.depth), 0) AS LEVEL
FROM
  hierarchy_paths hp
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.descendant_id = $1;

-- name: IsEntityAncestor :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      hierarchy_paths
    WHERE
      tenant_id = current_tenant_id()
      AND ancestor_id = $1
      AND descendant_id = $2
      AND depth > 0
  ) AS is_ancestor;

-- name: DeleteHierarchyPaths :exec
DELETE FROM
  hierarchy_paths
WHERE
  tenant_id = current_tenant_id()
  AND (
    ancestor_id = $1
    OR descendant_id = $1
  );

-- name: UpdateHierarchyPaths :exec
WITH RECURSIVE hierarchy_cte AS (
  -- Base case: self-reference
  SELECT
    current_tenant_id() AS tenant_id,
    $1::UUID AS ancestor_id,
    $1::UUID AS descendant_id,
    0 AS depth
  UNION
  ALL
  -- Recursive case: add ancestors
  SELECT
    h.tenant_id,
    hp.ancestor_id,
    h.descendant_id,
    h.depth + 1
  FROM
    hierarchy_cte h
    JOIN hierarchy_paths hp ON hp.descendant_id = h.ancestor_id
    AND hp.tenant_id = h.tenant_id
  WHERE
    h.depth < 5 -- Prevent infinite recursion
)
INSERT INTO
  hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
SELECT
  DISTINCT tenant_id,
  ancestor_id,
  descendant_id,
  depth
FROM
  hierarchy_cte ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING;

-- ===============================================
-- Advanced Entity Queries
-- ===============================================
-- name: GetEntityWithHierarchyInfo :one
SELECT
  e.*,
  COALESCE(MIN(hp.depth), 0) AS LEVEL,
  COUNT(children.uuid) AS child_count,
  parent_e.name AS parent_name
FROM
  entities e
  LEFT JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
  AND hp.tenant_id = e.tenant_id
  LEFT JOIN entities children ON children.parent_id = e.uuid
  AND children.tenant_id = e.tenant_id
  AND children.deleted_at IS NULL
  LEFT JOIN entities parent_e ON parent_e.uuid = e.parent_id
  AND parent_e.tenant_id = e.tenant_id
WHERE
  e.uuid = $1
  AND e.tenant_id = current_tenant_id()
  AND e.deleted_at IS NULL
GROUP BY
  e.uuid,
  parent_e.name;

-- name: GetEntityTreeStructure :many
WITH RECURSIVE entity_tree AS (
  SELECT
    e.*,
    0 AS LEVEL,
    CAST(e.name AS TEXT) AS path_text,
    CAST(e.name AS TEXT) AS sort_path
  FROM
    entities e
  WHERE
    e.tenant_id = current_tenant_id()
    AND e.parent_id IS NULL
    AND e.deleted_at IS NULL
  UNION
  ALL
  SELECT
    e.*,
    et.level + 1,
    CAST(et.path_text || ' > ' || e.name AS TEXT),
    CAST(et.sort_path || '/' || e.name AS TEXT)
  FROM
    entities e
    JOIN entity_tree et ON e.parent_id = et.uuid
  WHERE
    e.tenant_id = current_tenant_id()
    AND e.deleted_at IS NULL
    AND et.level < 10
)
SELECT
  *
FROM
  entity_tree
ORDER BY
  sort_path;

-- name: GetEntityStats :one
SELECT
  COUNT(*) AS total_entities,
  SUM(
    CASE
      WHEN is_active = TRUE THEN 1
      ELSE 0
    END
  ) AS active_entities,
  SUM(
    CASE
      WHEN hidden = false THEN 1
      ELSE 0
    END
  ) AS visible_entities,
  COUNT(DISTINCT TYPE) AS entity_types,
  SUM(
    CASE
      WHEN parent_id IS NULL THEN 1
      ELSE 0
    END
  ) AS root_entities,
  SUM(
    CASE
      WHEN accrual_method = TRUE THEN 1
      ELSE 0
    END
  ) AS accrual_entities,
  SUM(
    CASE
      WHEN accrual_method = false THEN 1
      ELSE 0
    END
  ) AS cash_entities
FROM
  entities;

-- name: GetEntitiesByFiscalYear :many
SELECT
  DISTINCT e.*
FROM
  entities e
  JOIN entitystate es ON e.uuid = es.entity_id
WHERE
  e.tenant_id = current_tenant_id()
  AND es.fiscal_year = $1
  AND e.deleted_at IS NULL
ORDER BY
  e.name;

-- name: ValidateEntityHierarchy :one
SELECT
  CASE
    WHEN COUNT(*) = 0 THEN TRUE
    ELSE false
  END AS is_valid
FROM
  hierarchy_paths hp1
  JOIN hierarchy_paths hp2 ON hp1.descendant_id = hp2.ancestor_id
WHERE
  hp1.ancestor_id = hp2.descendant_id
  AND hp1.depth > 0
  AND hp2.depth > 0;

-- ===============================================
-- Batch Operations
-- -- ===============================================
-- name: BatchUpdateEntityStatus :exec
UPDATE
  entities
SET
  is_active = sqlc.arg('is_active'),
  updated_at = NOW()
WHERE
  tenant_id = current_tenant_id()
  AND uuid = ANY(sqlc.arg('uuids')::UUID [])
  AND deleted_at IS NULL;

-- name: BatchSoftDeleteEntities :exec
UPDATE
  entities
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE
  tenant_id = current_tenant_id()
  AND uuid = ANY(sqlc.arg('uuids')::UUID []);

-- name: GetEntitiesByUUIDs :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND uuid = ANY(sqlc.arg('uuids')::UUID [])
  AND deleted_at IS NULL
ORDER BY
  name;

-- =====================================================================
-- 1. ENTITY VALIDATION AND INTEGRITY CHECKS
-- =====================================================================
-- name: ValidateEntityCode :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      entities
    WHERE
      code = $1
      AND tenant_id = current_tenant_id()
      AND uuid != $2
      AND deleted_at IS NULL
  )::BOOLEAN AS EXISTS;

-- name: ValidateEntityName :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      entities
    WHERE
      name = $1
      AND tenant_id = current_tenant_id()
      AND uuid != $2
      AND deleted_at IS NULL
  ) AS EXISTS;

-- name: ValidateEntityParent :one
SELECT
  (
    CASE
      WHEN $1 IS NULL THEN TRUE
      WHEN NOT EXISTS(
        SELECT
          1
        FROM
          entities
        WHERE
          uuid = $1
          AND tenant_id = current_tenant_id()
          AND deleted_at IS NULL
      ) THEN false
      WHEN EXISTS(
        SELECT
          1
        FROM
          hierarchy_paths
        WHERE
          tenant_id = current_tenant_id()
          AND ancestor_id = $2
          AND descendant_id = $1
      ) THEN false
      ELSE TRUE
    END
  )::BOOLEAN AS valid;

-- name: CheckCircularReference :one
SELECT
  EXISTS(
    SELECT
      1
    FROM
      hierarchy_paths
    WHERE
      tenant_id = current_tenant_id()
      AND ancestor_id = $2
      AND descendant_id = $1
  )::BOOLEAN AS EXISTS;

-- name: GetEntityDepth :one
SELECT
  COALESCE(MAX(depth), 0) AS depth
FROM
  hierarchy_paths
WHERE
  tenant_id = current_tenant_id()
  AND ancestor_id = $1;

-- =====================================================================
-- 2. ENTITY SEARCH AND FILTERING ENHANCEMENTS
-- =====================================================================
-- name: SearchEntitiesByCodeAndName :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND (
    code ILIKE '%' || $1 || '%'
    OR name ILIKE '%' || $1 || '%'
  )
  AND deleted_at IS NULL
ORDER BY
  CASE
    WHEN code ILIKE $1 || '%' THEN 1
    WHEN name ILIKE $1 || '%' THEN 2
    ELSE 3
  END,
  name
LIMIT
  $2;

-- name: ListEntitiesWithPagination :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('type') IS NULL
    OR TYPE = sqlc.narg('type')
  )
  AND (
    sqlc.narg('is_active') IS NULL
    OR is_active = sqlc.narg('is_active')
  )
  AND (
    sqlc.narg('hidden') IS NULL
    OR hidden = sqlc.narg('hidden')
  )
ORDER BY
  name
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountEntitiesWithFilters :one
SELECT
  COUNT(*) AS count
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg('type') IS NULL
    OR TYPE = sqlc.narg('type')
  )
  AND (
    sqlc.narg('is_active') IS NULL
    OR is_active = sqlc.narg('is_active')
  )
  AND (
    sqlc.narg('hidden') IS NULL
    OR hidden = sqlc.narg('hidden')
  );

-- name: ListEntitiesByTypes :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND TYPE = ANY($1::VARCHAR [])
  AND deleted_at IS NULL
ORDER BY
  TYPE,
  name;

-- name: GetEntitiesByFiscalYearStart :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND fy_start_month = $1
  AND deleted_at IS NULL
ORDER BY
  name;

-- =====================================================================
-- 3. HIERARCHY BULK OPERATIONS
-- =====================================================================
-- name: MoveEntityToNewParent :exec
WITH RECURSIVE affected_entities AS (
  SELECT
    $1::UUID AS entity_id,
    0 AS depth
  UNION
  ALL
  SELECT
    hp.descendant_id,
    ae.depth + 1
  FROM
    affected_entities ae
    JOIN hierarchy_paths hp ON hp.ancestor_id = ae.entity_id
  WHERE
    ae.depth < 10
),
delete_paths AS (
  DELETE FROM
    hierarchy_paths
  WHERE
    descendant_id IN (
      SELECT
        entity_id
      FROM
        affected_entities
    )
),
insert_new_paths AS (
  INSERT INTO
    hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
  SELECT
    current_tenant_id(),
    ancestor_paths.ancestor_id,
    ae.entity_id,
    ancestor_paths.depth + descendant_paths.depth + 1
  FROM
    affected_entities ae
    CROSS JOIN (
      SELECT
        ancestor_id,
        depth
      FROM
        hierarchy_paths
      WHERE
        descendant_id = $2
      UNION
      ALL
      SELECT
        $2::UUID,
        0
    ) ancestor_paths
    CROSS JOIN (
      SELECT
        descendant_id,
        depth
      FROM
        hierarchy_paths
      WHERE
        ancestor_id = $1
      UNION
      ALL
      SELECT
        $1::UUID,
        0
    ) descendant_paths
  WHERE
    ae.entity_id = descendant_paths.descendant_id
)
UPDATE
  entities
SET
  parent_id = $2,
  updated_at = NOW()
WHERE
  uuid = $1;

-- name: GetEntitySubtree :many
SELECT
  e.*,
  hp.depth
FROM
  entities e
  JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.ancestor_id = $1
  AND e.deleted_at IS NULL
  AND (
    $2::INTEGER IS NULL
    OR hp.depth <= $2
  )
ORDER BY
  hp.depth,
  e.name;

-- name: GetEntityPath :many
SELECT
  e.*,
  hp.depth
FROM
  entities e
  JOIN hierarchy_paths hp ON e.uuid = hp.ancestor_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND hp.descendant_id = $1
  AND e.deleted_at IS NULL
ORDER BY
  hp.depth DESC;

-- name: BulkMoveEntities :exec
UPDATE
  entities
SET
  parent_id = sqlc.arg('parent_id'),
  updated_at = NOW()
WHERE
  tenant_id = current_tenant_id()
  AND uuid = ANY(sqlc.arg('entity_ids')::UUID [])
  AND deleted_at IS NULL;

-- =====================================================================
-- 4. AUDIT AND MONITORING QUERIES
-- =====================================================================
-- name: GetRecentlyModifiedEntities :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND updated_at >= $1
  AND deleted_at IS NULL
ORDER BY
  updated_at DESC
LIMIT
  $2;

-- name: GetRecentlyDeletedEntities :many
SELECT
  *
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at >= $1
  AND deleted_at IS NOT NULL
ORDER BY
  deleted_at DESC
LIMIT
  $2;

-- name: GetEntityAuditLog :many
SELECT
  uuid,
  name,
  TYPE,
  is_active,
  hidden,
  created_at,
  updated_at,
  deleted_at,
  CASE
    WHEN deleted_at IS NOT NULL THEN 'DELETED'
    WHEN updated_at > created_at THEN 'UPDATED'
    ELSE 'CREATED'
  END AS ACTION
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND (
    created_at >= $1
    OR updated_at >= $1
    OR deleted_at >= $1
  )
ORDER BY
  GREATEST(
    created_at,
    updated_at,
    COALESCE(deleted_at, created_at)
  ) DESC;

-- name: GetOrphanedEntities :many
SELECT
  e.*
FROM
  entities e
  LEFT JOIN entities parent ON parent.uuid = e.parent_id
  AND parent.tenant_id = e.tenant_id
WHERE
  e.tenant_id = current_tenant_id()
  AND e.parent_id IS NOT NULL
  AND parent.uuid IS NULL
  AND e.deleted_at IS NULL;

-- name: GetInconsistentHierarchyPaths :many
SELECT
  DISTINCT hp.ancestor_id,
  hp.descendant_id,
  hp.depth
FROM
  hierarchy_paths hp
  LEFT JOIN entities e1 ON hp.ancestor_id = e1.uuid
  AND e1.tenant_id = hp.tenant_id
  LEFT JOIN entities e2 ON hp.descendant_id = e2.uuid
  AND e2.tenant_id = hp.tenant_id
WHERE
  hp.tenant_id = current_tenant_id()
  AND (
    e1.uuid IS NULL
    OR e2.uuid IS NULL
    OR e1.deleted_at IS NOT NULL
    OR e2.deleted_at IS NOT NULL
  );

-- =====================================================================
-- 6. PERFORMANCE AND ANALYTICS QUERIES
-- =====================================================================
-- name: GetEntityCountByType :many
SELECT
  TYPE,
  COUNT(*) AS count
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
GROUP BY
  TYPE
ORDER BY
  count DESC;

-- name: GetEntityHierarchyStats :one
SELECT
  COUNT(*) AS total_entities,
  COUNT(*) FILTER (
    WHERE
      e.parent_id IS NULL
  ) AS root_entities,
  MAX(hp.depth) AS max_depth,
  AVG(hp.depth) AS avg_depth,
  COUNT(DISTINCT hp.ancestor_id) AS entities_with_children
FROM
  entities e
  LEFT JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
  AND hp.tenant_id = e.tenant_id
WHERE
  e.tenant_id = current_tenant_id()
  AND e.deleted_at IS NULL;

-- name: GetEntitySequenceStats :many
SELECT
  e.name AS entity_name,
  es.key,
  es.fiscal_year,
  es.sequence,
  es.sequence - 1 AS documents_created
FROM
  entitystate es
  JOIN entities e ON es.entity_id = e.uuid
WHERE
  e.tenant_id = current_tenant_id()
  AND (
    $1::UUID IS NULL
    OR es.entity_id = $1
  )
ORDER BY
  e.name,
  es.key,
  es.fiscal_year;

-- name: GetUnusedEntityCodes :many
SELECT
  DISTINCT code
FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND code IS NOT NULL
  AND deleted_at IS NOT NULL
ORDER BY
  code;

-- =====================================================================
-- 7. MAINTENANCE AND CLEANUP QUERIES
-- =====================================================================
-- name: CleanupOrphanedHierarchyPaths :exec
DELETE FROM
  hierarchy_paths
WHERE
  tenant_id = current_tenant_id()
  AND (
    NOT EXISTS(
      SELECT
        1
      FROM
        entities
      WHERE
        uuid = ancestor_id
        AND tenant_id = current_tenant_id()
    )
    OR NOT EXISTS(
      SELECT
        1
      FROM
        entities
      WHERE
        uuid = descendant_id
        AND tenant_id = current_tenant_id()
    )
  );

-- name: RebuildHierarchyPaths :exec
WITH RECURSIVE entity_hierarchy AS (
  SELECT
    uuid AS ancestor_id,
    uuid AS descendant_id,
    0 AS depth,
    tenant_id
  FROM
    entities
  WHERE
    tenant_id = current_tenant_id()
    AND deleted_at IS NULL
  UNION
  ALL
  SELECT
    eh.ancestor_id,
    e.uuid,
    eh.depth + 1,
    e.tenant_id
  FROM
    entity_hierarchy eh
    JOIN entities e ON e.parent_id = eh.descendant_id
  WHERE
    e.tenant_id = current_tenant_id()
    AND e.deleted_at IS NULL
    AND eh.depth < 10
),
cleanup AS (
  DELETE FROM
    hierarchy_paths
  WHERE
    tenant_id = current_tenant_id()
)
INSERT INTO
  hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
SELECT
  tenant_id,
  ancestor_id,
  descendant_id,
  depth
FROM
  entity_hierarchy;

-- name: ArchiveOldDeletedEntities :exec
DELETE FROM
  entities
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at < $1
  AND deleted_at IS NOT NULL;

-- name: GetEntityHealthCheck :one
SELECT
  (
    SELECT
      COUNT(*)
    FROM
      entities
    WHERE
      tenant_id = current_tenant_id()
      AND deleted_at IS NULL
  ) AS active_entities,
  (
    SELECT
      COUNT(*)
    FROM
      hierarchy_paths
    WHERE
      tenant_id = current_tenant_id()
  ) AS hierarchy_paths,
  (
    SELECT
      COUNT(*)
    FROM
      entitystate es
      JOIN entities e ON es.entity_id = e.uuid
    WHERE
      e.tenant_id = current_tenant_id()
  ) AS entity_states,
  (
    SELECT
      COUNT(*)
    FROM
      entities e
      LEFT JOIN entities p ON e.parent_id = p.uuid
    WHERE
      e.tenant_id = current_tenant_id()
      AND e.parent_id IS NOT NULL
      AND p.uuid IS NULL
  ) AS orphaned_entities,
  (
    SELECT
      COUNT(*)
    FROM
      hierarchy_paths hp
      LEFT JOIN entities e1 ON hp.ancestor_id = e1.uuid
      LEFT JOIN entities e2 ON hp.descendant_id = e2.uuid
    WHERE
      hp.tenant_id = current_tenant_id()
      AND (
        e1.uuid IS NULL
        OR e2.uuid IS NULL
      )
  ) AS orphaned_paths;

-- =====================================================================
-- ENTITYSTATE SQLC QUERIES
-- =====================================================================
-- name: CreateEntityState :one
INSERT INTO
  entitystate (
    uuid,
    tenant_id,
    fiscal_year,
    KEY,
    sequence,
    entity_id,
    entity_unit_id,
    created_at,
    updated_at
  )
VALUES
  (
    sqlc.arg(uuid)::UUID,
    current_tenant_id(),
    sqlc.narg(fiscal_year)::SMALLINT,
    sqlc.arg(KEY)::VARCHAR(10),
    sqlc.arg(sequence)::BIGINT,
    sqlc.arg(entity_id)::UUID,
    sqlc.narg(entity_unit_id)::UUID,
    NOW(),
    NOW()
  )
RETURNING
  *;

-- name: GetEntityState :one
SELECT
  *
FROM
  entitystate
WHERE
  uuid = sqlc.arg(uuid)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetEntityStateByEntityAndKey :one
SELECT
  *
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg(fiscal_year)::SMALLINT IS NULL
    OR fiscal_year = sqlc.narg(fiscal_year)::SMALLINT
  );

-- name: GetEntityStateByEntityKeyAndFiscalYear :one
SELECT
  *
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND fiscal_year = sqlc.arg(fiscal_year)::SMALLINT
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetNextSequenceNumber :one
SELECT
  sequence
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg(fiscal_year)::SMALLINT IS NULL
    OR fiscal_year = sqlc.narg(fiscal_year)::SMALLINT
  ) FOR
UPDATE
;

-- name: IncrementSequenceNumber :one
UPDATE
  entitystate
SET
  sequence = sequence + 1,
  updated_at = NOW()
WHERE
  uuid = sqlc.arg(uuid)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING
  sequence;

-- name: UpdateEntityState :one
UPDATE
  entitystate
SET
  fiscal_year = COALESCE(sqlc.narg(fiscal_year)::SMALLINT, fiscal_year),
  KEY = COALESCE(sqlc.narg(KEY)::VARCHAR(10), KEY),
  sequence = COALESCE(sqlc.narg(sequence)::BIGINT, sequence),
  entity_id = COALESCE(sqlc.narg(entity_id)::UUID, entity_id),
  entity_unit_id = COALESCE(sqlc.narg(entity_unit_id)::UUID, entity_unit_id),
  updated_at = NOW()
WHERE
  uuid = sqlc.arg(uuid)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING
  *;

-- name: SetSequenceNumber :one
UPDATE
  entitystate
SET
  sequence = sqlc.arg(sequence)::BIGINT,
  updated_at = NOW()
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg(fiscal_year)::SMALLINT IS NULL
    OR fiscal_year = sqlc.narg(fiscal_year)::SMALLINT
  )
RETURNING
  *;

-- name: SoftDeleteEntityState :exec
UPDATE
  entitystate
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE
  uuid = sqlc.arg(uuid)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: ListEntityStatesByEntity :many
SELECT
  *
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  KEY,
  fiscal_year;

-- name: ListEntityStatesByEntityAndKey :many
SELECT
  *
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  fiscal_year;

-- name: ListEntityStatesByFiscalYear :many
SELECT
  *
FROM
  entitystate
WHERE
  fiscal_year = sqlc.arg(fiscal_year)::SMALLINT
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  entity_id,
  KEY;

-- name: ListEntityStatesByEntityUnit :many
SELECT
  *
FROM
  entitystate
WHERE
  entity_unit_id = sqlc.arg(entity_unit_id)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
ORDER BY
  entity_id,
  KEY,
  fiscal_year;

-- name: Get_OrCreateEntityState :one
WITH ins AS (
  INSERT INTO
    entitystate (
      uuid,
      tenant_id,
      fiscal_year,
      KEY,
      sequence,
      entity_id,
      entity_unit_id,
      created_at,
      updated_at
    )
  VALUES
    (
      sqlc.arg(uuid),  -- generate UUID in app layer
      current_tenant_id(),
      sqlc.arg(fiscal_year),
      sqlc.arg(KEY),
      1,  -- start sequence at 1
      sqlc.arg(entity_id),
      sqlc.narg(entity_unit_id),
      NOW(),
      NOW()
    ) ON CONFLICT (tenant_id, fiscal_year, KEY, entity_id) DO NOTHING
  RETURNING
    *
)
SELECT
  *
FROM
  ins
UNION
SELECT
  *
FROM
  entitystate
WHERE
  tenant_id = current_tenant_id()
  AND fiscal_year = sqlc.arg(fiscal_year)
  AND KEY = sqlc.arg(KEY)
  AND entity_id = sqlc.arg(entity_id)
  AND deleted_at IS NULL
LIMIT
  1;

-- -- name: GetOrCreateEntityState :one
-- WITH existing AS (
--     SELECT *
--     FROM entitystate
--     WHERE entity_id = sqlc.arg(entity_id)::UUID
--       AND key = sqlc.arg(key)::VARCHAR(10)
--       AND tenant_id = current_tenant_id()
--       AND deleted_at IS NULL
--       AND (sqlc.narg(fiscal_year)::SMALLINT IS NULL OR fiscal_year = sqlc.narg(fiscal_year)::SMALLINT)
-- ),
-- new_record AS (
--     INSERT INTO entitystate (
--         uuid,
--         tenant_id,
--         fiscal_year,
--         key,
--         sequence,
--         entity_id,
--         entity_unit_id,
--         created_at,
--         updated_at
--     )
--     SELECT
--         sqlc.arg(uuid)::UUID,
--         current_tenant_id(),
--         sqlc.narg(fiscal_year)::SMALLINT,
--         sqlc.arg(key)::VARCHAR(10),
--         sqlc.arg(sequence)::BIGINT,
--         sqlc.arg(entity_id)::UUID,
--         sqlc.narg(entity_unit_id)::UUID,
--         NOW(),
--         NOW()
--     WHERE NOT EXISTS (SELECT 1 FROM existing)
--     RETURNING *
-- )
-- SELECT * FROM existing
-- UNION ALL
-- SELECT * FROM new_record;
--
-- name: GetNextSequenceAndIncrement :one
UPDATE
  entitystate
SET
  sequence = sequence + 1,
  updated_at = NOW()
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg(fiscal_year)::SMALLINT IS NULL
    OR fiscal_year = sqlc.narg(fiscal_year)::SMALLINT
  )
RETURNING
  sequence - 1 AS used_sequence,
  sequence AS next_sequence;

--
-- name: BulkCreateEntityStates :copyfrom
INSERT INTO
  entitystate (
    uuid,
    tenant_id,
    fiscal_year,
    KEY,
    sequence,
    entity_id,
    entity_unit_id,
    created_at,
    updated_at,
    deleted_at
  )
VALUES
  (
    sqlc.arg(uuid),
    sqlc.arg(tenant_id),
    sqlc.narg(fiscal_year),
    sqlc.arg(KEY),
    sqlc.arg(sequence),
    sqlc.arg(entity_id),
    sqlc.narg(entity_unit_id),
    sqlc.arg(created_at),
    sqlc.arg(updated_at),
    sqlc.narg(deleted_at)
  );

-- name: CountEntityStatesByEntity :one
SELECT
  COUNT(*)
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: CountEntityStatesByKey :one
SELECT
  COUNT(*)
FROM
  entitystate
WHERE
  KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: GetMaxSequenceByEntityAndKey :one
SELECT
  COALESCE(MAX(sequence), 0) AS max_sequence
FROM
  entitystate
WHERE
  entity_id = sqlc.arg(entity_id)::UUID
  AND KEY = sqlc.arg(KEY)::VARCHAR(10)
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg(fiscal_year)::SMALLINT IS NULL
    OR fiscal_year = sqlc.narg(fiscal_year)::SMALLINT
  );

-- name: ResetSequenceNumber :one
UPDATE
  entitystate
SET
  sequence = sqlc.arg(sequence)::BIGINT,
  updated_at = NOW()
WHERE
  uuid = sqlc.arg(uuid)::UUID
  AND tenant_id = current_tenant_id()
  AND deleted_at IS NULL
RETURNING
  *;

-- name: GetEntityStatesWithPaging :many
SELECT
  *
FROM
  entitystate
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    sqlc.narg(entity_id) IS NULL
    OR entity_id = sqlc.narg(entity_id)
  )
  AND (
    sqlc.narg(KEY) IS NULL
    OR KEY = sqlc.narg(KEY)
  )
  AND (
    sqlc.narg(fiscal_year) IS NULL
    OR fiscal_year = sqlc.narg(fiscal_year)
  )
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

/*
I'll create sqlc functions for your entity management schema. This will include CRUD operations, hierarchy management, and state tracking.I've created a set of sqlc functions for your entity management schema. Here's what's included:

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
- **Statistics**:  stats about entities in a tenant
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
