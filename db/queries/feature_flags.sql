-- =====================================================================
--  FEATURE FLAG QUERIES
--  Two sections:
--    1. feature_flag_definitions + tenant_feature_flags  (IAM session pre-computation)
--    2. feature_flags + tenant_feature_overrides         (legacy per-tenant flags)
-- =====================================================================

-- =====================================================================
-- SECTION 1: IAM SPEC — feature_flag_definitions + tenant_feature_flags
-- =====================================================================

-- name: ResolveAllFlagsForTenant :many
-- THE core query. Called once at login. Returns effective value for every flag.
-- Result stored in sessions.configuration.flags for O(1) per-request lookups.
SELECT
    ffd.flag_key,
    ffd.is_system,
    COALESCE(tff.enabled, ffd.default_value) AS effective_value
FROM feature_flag_definitions ffd
LEFT JOIN tenant_feature_flags tff
    ON tff.flag_id = ffd.id AND tff.tenant_id = $1
ORDER BY ffd.flag_key;

-- name: GetFlagDefinition :one
SELECT * FROM feature_flag_definitions
WHERE flag_key = $1;

-- name: ListFlagDefinitions :many
SELECT * FROM feature_flag_definitions
WHERE ($1::boolean IS NULL OR is_system = $1)
ORDER BY flag_key;

-- name: ListTenantFlagsWithDefinitions :many
-- Used by the tenant admin flags screen: returns all non-system flags with current tenant values.
SELECT
    ffd.id          AS definition_id,
    ffd.flag_key,
    ffd.label,
    ffd.description,
    ffd.default_value,
    ffd.is_system,
    ffd.module_id,
    ffd.resource_id,
    tff.id          AS override_id,
    tff.enabled     AS tenant_enabled,
    tff.set_by,
    tff.set_at,
    COALESCE(tff.enabled, ffd.default_value) AS effective_value
FROM feature_flag_definitions ffd
LEFT JOIN tenant_feature_flags tff
    ON tff.flag_id = ffd.id AND tff.tenant_id = $1
WHERE ($2::boolean IS NULL OR ffd.is_system = $2)
ORDER BY ffd.flag_key;

-- name: UpsertTenantFlag :exec
-- Enable or disable a flag for a specific tenant.
-- FlagService calls InvalidateSessionsByTenant after this for module/resource flags.
INSERT INTO tenant_feature_flags (tenant_id, flag_id, flag_key, enabled, set_by)
SELECT $1, ffd.id, ffd.flag_key, $2, $3
FROM   feature_flag_definitions ffd
WHERE  ffd.flag_key = $4
ON CONFLICT (tenant_id, flag_id)
DO UPDATE SET enabled = EXCLUDED.enabled,
             set_by   = EXCLUDED.set_by,
             set_at   = NOW();

-- name: DeleteTenantFlagOverride :exec
-- Removes tenant override — flag reverts to default_value.
DELETE FROM tenant_feature_flags
WHERE tenant_id = $1
  AND flag_key  = $2;

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
