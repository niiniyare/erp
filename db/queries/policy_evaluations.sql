-- Policy Evaluations CRUD Operations and Cache Management

-- name: CreatePolicyEvaluation :one
INSERT INTO policy_evaluations (
    id, tenant_id, user_id, resource_id, action_id, context_hash,
    decision, applicable_policies, evaluation_time_ms, expires_at
) VALUES (
    $1, current_tenant_id(), $2, $3, $4, $5,
    $6, $7, $8, $9
) RETURNING *;

-- name: GetPolicyEvaluation :one
SELECT * FROM policy_evaluations
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: GetPolicyEvaluationByContextHash :one
SELECT * FROM policy_evaluations
WHERE tenant_id = current_tenant_id()
    AND user_id = $1
    AND resource_id = $2
    AND action_id = $3
    AND context_hash = $4
    AND expires_at > NOW();

-- name: DeletePolicyEvaluation :exec
DELETE FROM policy_evaluations
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: DeleteExpiredPolicyEvaluations :exec
DELETE FROM policy_evaluations
WHERE expires_at <= NOW();

-- name: ListPolicyEvaluationsForUser :many
SELECT * FROM policy_evaluations
WHERE tenant_id = current_tenant_id() AND user_id = $1
ORDER BY evaluated_at DESC
LIMIT $2 OFFSET $3;
