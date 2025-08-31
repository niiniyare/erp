-- =====================================================================
--  SIMPLIFIED FEATURE FLAG QUERIES
--  Matching the actual table schema from migrations
-- =====================================================================
-- name: CreateFeatureFlag :one
INSERT INTO
  feature_flags (
    tenant_id,
    name,
    description,
    flag_type,
    default_value,
    rollout_percentage,
    target_audience,
    metadata
  )
VALUES
  (
    current_tenant_id(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
  )
RETURNING
  *;

-- name: GetFeatureFlagByName :one
SELECT
  *
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND name = $1
  AND deleted_at IS NULL;

-- name: GetFeatureFlagByID :one
SELECT
  *
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND id = $1
  AND deleted_at IS NULL;

-- name: UpdateFeatureFlag :one
UPDATE
  feature_flags
SET
  name = $2,
  description = $3,
  flag_type = $4,
  default_value = $5,
  rollout_percentage = $6,
  target_audience = $7,
  metadata = $8,
  updated_at = NOW()
WHERE
  tenant_id = current_tenant_id()
  AND id = $1
  AND deleted_at IS NULL
RETURNING
  *;

-- name: DeleteFeatureFlag :exec
UPDATE
  feature_flags
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE
  tenant_id = current_tenant_id()
  AND id = $1
  AND deleted_at IS NULL;

-- name: ListFeatureFlags :many
SELECT
  *
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    $1::text IS NULL
    OR flag_type = $1
  )
ORDER BY
  name
LIMIT
  $2 OFFSET $3;

-- name: GetActiveFeatureFlags :many
SELECT
  *
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND default_value = TRUE
ORDER BY
  name;

-- name: CleanupExpiredFeatureFlags :one
-- For future use when we add expires_at to metadata
SELECT
  COUNT(*)
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: BulkEvaluateFeatureFlags :many
-- Simplified bulk evaluation based on current schema
SELECT
  name AS flag_key,
  CASE
    WHEN rollout_percentage IS NOT NULL THEN CASE
      WHEN (ABS(HASHTEXT($1::text || name)) % 100) < rollout_percentage THEN default_value
      ELSE false
    END
    ELSE default_value
  END AS evaluated_value,
  default_value AS is_enabled,
  'default' AS source,
  'Basic evaluation' AS reason,
  '' AS variation,
  NOW() AS evaluated_at
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND name = ANY($2::text [])
  AND deleted_at IS NULL;

-- name: GetFeatureFlagStats :one
SELECT
  COUNT(*) AS total_flags,
  COUNT(*) FILTER (
    WHERE
      default_value = TRUE
  ) AS enabled_flags,
  COUNT(*) FILTER (
    WHERE
      rollout_percentage IS NOT NULL
  ) AS rollout_flags,
  AVG(rollout_percentage) AS avg_rollout_percentage
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL;

-- name: SearchFeatureFlags :many
SELECT
  *
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND deleted_at IS NULL
  AND (
    name ILIKE '%' || $1 || '%'
    OR description ILIKE '%' || $1 || '%'
  )
ORDER BY
  name
LIMIT
  $2 OFFSET $3;

-- name: GetFeatureFlagsByType :many
SELECT
  *
FROM
  feature_flags
WHERE
  tenant_id = current_tenant_id()
  AND flag_type = $1
  AND deleted_at IS NULL
ORDER BY
  name;
