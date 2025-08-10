-- Policy Evaluations CRUD Operations and Cache Management for ABAC
-- name: CacheEvaluationResult :exec
INSERT INTO
  policy_evaluations (
    tenant_id,
    user_id,
    resource_type,
    resource_id,
    ACTION,
    entity_id,
    context_hash,
    decision,
    applicable_policies,
    policy_decisions,
    evaluation_time_ms,
    cache_key,
    expires_at
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
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
  ) ON CONFLICT (
    tenant_id,
    user_id,
    resource_type,
    resource_id,
    ACTION,
    context_hash
  ) DO
UPDATE
SET
  decision = EXCLUDED.decision,
  applicable_policies = EXCLUDED.applicable_policies,
  policy_decisions = EXCLUDED.policy_decisions,
  evaluation_time_ms = EXCLUDED.evaluation_time_ms,
  evaluated_at = NOW(),
  expires_at = EXCLUDED.expires_at;

-- name: GetCachedEvaluationResult :one
SELECT
  *
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND user_id = $1
  AND resource_type = $2
  AND (
    $3::UUID IS NULL
    AND resource_id IS NULL
    OR resource_id = $3
  )
  AND ACTION = $4
  AND context_hash = $5
  AND expires_at > NOW()
ORDER BY
  evaluated_at DESC
LIMIT
  1;

-- name: InvalidateUserEvaluations :exec
DELETE FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND user_id = $1;

-- name: InvalidateResourceEvaluations :exec
DELETE FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND resource_type = $1
  AND (
    $2::UUID IS NULL
    OR resource_id = $2
  );

-- name: InvalidateActionEvaluations :exec
DELETE FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND ACTION = $1;

-- name: InvalidatePolicyEvaluations :exec
DELETE FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND applicable_policies && $1::UUID [];

-- name: InvalidateAllEvaluations :exec
DELETE FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id();

-- name: CleanupExpiredEvaluations :exec
DELETE FROM
  policy_evaluations
WHERE
  expires_at < NOW()
  AND tenant_id = current_tenant_id();

-- name: GetEvaluationCacheStats :one
SELECT
  COUNT(*) AS total_cached_evaluations,
  COUNT(
    CASE
      WHEN expires_at > NOW() THEN 1
    END
  ) AS active_evaluations,
  COUNT(
    CASE
      WHEN expires_at <= NOW() THEN 1
    END
  ) AS expired_evaluations,
  COUNT(DISTINCT user_id) AS unique_users,
  COUNT(DISTINCT resource_type) AS unique_resource_types,
  COUNT(DISTINCT ACTION) AS unique_actions,
  AVG(evaluation_time_ms) AS avg_evaluation_time_ms,
  MIN(evaluation_time_ms) AS min_evaluation_time_ms,
  MAX(evaluation_time_ms) AS max_evaluation_time_ms,
  ROUND(
    (
      COUNT(
        CASE
          WHEN expires_at > NOW() THEN 1
        END
      )::NUMERIC / NULLIF(COUNT(*), 0)
    ) * 100,
    2
  ) AS cache_hit_rate
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id();

-- name: GetUserEvaluationHistory :many
SELECT
  *
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND user_id = $1
  AND (
    $2::VARCHAR IS NULL
    OR resource_type = $2
  )
  AND (
    $3::VARCHAR IS NULL
    OR ACTION = $3
  )
ORDER BY
  evaluated_at DESC
LIMIT
  $4 OFFSET $5;

-- name: GetResourceEvaluationHistory :many
SELECT
  *
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND resource_type = $1
  AND (
    $2::UUID IS NULL
    OR resource_id = $2
  )
  AND (
    $3::VARCHAR IS NULL
    OR ACTION = $3
  )
ORDER BY
  evaluated_at DESC
LIMIT
  $4 OFFSET $5;

-- name: GetEvaluationsByDecision :many
SELECT
  *
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND decision = $1
  AND evaluated_at >= $2
  AND evaluated_at <= $3
ORDER BY
  evaluated_at DESC
LIMIT
  $4 OFFSET $5;

-- name: CountEvaluationsByDecision :one
SELECT
  COUNT(
    CASE
      WHEN decision = 'ALLOW' THEN 1
    END
  ) AS allow_count,
  COUNT(
    CASE
      WHEN decision = 'DENY' THEN 1
    END
  ) AS deny_count,
  COUNT(
    CASE
      WHEN decision = 'NOT_APPLICABLE' THEN 1
    END
  ) AS not_applicable_count,
  COUNT(*) AS total_count
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND evaluated_at >= $1
  AND evaluated_at <= $2;

-- name: GetEvaluationMetrics :one
SELECT
  AVG(evaluation_time_ms) AS avg_evaluation_time,
  PERCENTILE_CONT(0.5) WITHIN GROUP (
    ORDER BY
      evaluation_time_ms
  ) AS median_evaluation_time,
  PERCENTILE_CONT(0.95) WITHIN GROUP (
    ORDER BY
      evaluation_time_ms
  ) AS p95_evaluation_time,
  PERCENTILE_CONT(0.99) WITHIN GROUP (
    ORDER BY
      evaluation_time_ms
  ) AS p99_evaluation_time,
  COUNT(*) AS total_evaluations,
  COUNT(DISTINCT user_id) AS unique_users,
  COUNT(
    DISTINCT resource_type || ':' || COALESCE(resource_id::TEXT, '')
  ) AS unique_resources
FROM
  policy_evaluations
WHERE
  tenant_id = current_tenant_id()
  AND evaluated_at >= $1
  AND evaluated_at <= $2;

-- name: CreatePolicyEvaluation :one
INSERT INTO
  policy_evaluations (
    user_id,
    resource_id,
    ACTION,
    context_hash,
    decision
  )
VALUES
  ($1, $2, $3, $4, $5)
RETURNING
  *;
