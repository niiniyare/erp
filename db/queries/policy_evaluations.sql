-- Policy Evaluations CRUD Operations and Cache Management for ABAC

-- name: CacheEvaluationResult :exec
INSERT INTO policy_evaluations (
    tenant_id, user_id, resource_type, resource_id, action, entity_id, 
    context_hash, decision, applicable_policies, policy_decisions, 
    evaluation_time_ms, cache_key, expires_at
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, 
    $6, $7, $8, $9, $10, $11, $12
) ON CONFLICT (tenant_id, user_id, resource_type, resource_id, action, context_hash)
DO UPDATE SET
    decision = EXCLUDED.decision,
    applicable_policies = EXCLUDED.applicable_policies,
    policy_decisions = EXCLUDED.policy_decisions,
    evaluation_time_ms = EXCLUDED.evaluation_time_ms,
    evaluated_at = NOW(),
    expires_at = EXCLUDED.expires_at;

-- name: GetCachedEvaluationResult :one
SELECT * FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND user_id = $1
  AND resource_type = $2
  AND ($3::UUID IS NULL AND resource_id IS NULL OR resource_id = $3)
  AND action = $4
  AND context_hash = $5
  AND expires_at > NOW()
ORDER BY evaluated_at DESC
LIMIT 1;

-- name: InvalidateUserEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND user_id = $1;

-- name: InvalidateResourceEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND resource_type = $1
  AND ($2::UUID IS NULL OR resource_id = $2);

-- name: InvalidateActionEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND action = $1;

-- name: InvalidatePolicyEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND applicable_policies && $1::UUID[];

-- name: InvalidateAllEvaluations :exec
DELETE FROM policy_evaluations 
WHERE tenant_id = current_tenant_id();

-- name: CleanupExpiredEvaluations :exec
DELETE FROM policy_evaluations 
WHERE expires_at < NOW();

-- name: GetEvaluationCacheStats :one
SELECT 
    COUNT(*) as total_cached_evaluations,
    COUNT(CASE WHEN expires_at > NOW() THEN 1 END) as active_evaluations,
    COUNT(CASE WHEN expires_at <= NOW() THEN 1 END) as expired_evaluations,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT resource_type) as unique_resource_types,
    COUNT(DISTINCT action) as unique_actions,
    AVG(evaluation_time_ms) as avg_evaluation_time_ms,
    MIN(evaluation_time_ms) as min_evaluation_time_ms,
    MAX(evaluation_time_ms) as max_evaluation_time_ms,
    ROUND(
        (COUNT(CASE WHEN expires_at > NOW() THEN 1 END)::NUMERIC / 
         NULLIF(COUNT(*), 0)) * 100, 2
    ) as cache_hit_rate
FROM policy_evaluations 
WHERE tenant_id = current_tenant_id();

-- name: GetUserEvaluationHistory :many
SELECT * FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND user_id = $1
  AND ($2::VARCHAR IS NULL OR resource_type = $2)
  AND ($3::VARCHAR IS NULL OR action = $3)
ORDER BY evaluated_at DESC
LIMIT $4 OFFSET $5;

-- name: GetResourceEvaluationHistory :many
SELECT * FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND resource_type = $1
  AND ($2::UUID IS NULL OR resource_id = $2)
  AND ($3::VARCHAR IS NULL OR action = $3)
ORDER BY evaluated_at DESC
LIMIT $4 OFFSET $5;

-- name: GetEvaluationsByDecision :many
SELECT * FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND decision = $1
  AND evaluated_at >= $2
  AND evaluated_at <= $3
ORDER BY evaluated_at DESC
LIMIT $4 OFFSET $5;

-- name: CountEvaluationsByDecision :one
SELECT 
    COUNT(CASE WHEN decision = 'ALLOW' THEN 1 END) as allow_count,
    COUNT(CASE WHEN decision = 'DENY' THEN 1 END) as deny_count,
    COUNT(CASE WHEN decision = 'NOT_APPLICABLE' THEN 1 END) as not_applicable_count,
    COUNT(*) as total_count
FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND evaluated_at >= $1
  AND evaluated_at <= $2;

-- name: GetEvaluationMetrics :one
SELECT 
    AVG(evaluation_time_ms) as avg_evaluation_time,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY evaluation_time_ms) as median_evaluation_time,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY evaluation_time_ms) as p95_evaluation_time,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY evaluation_time_ms) as p99_evaluation_time,
    COUNT(*) as total_evaluations,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT resource_type || ':' || COALESCE(resource_id::TEXT, '')) as unique_resources
FROM policy_evaluations 
WHERE tenant_id = current_tenant_id()
  AND evaluated_at >= $1
  AND evaluated_at <= $2;