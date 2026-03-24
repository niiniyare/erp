-- -- name: CreateAction :one
-- INSERT INTO
--   actions (tenant_id, name, action_type)
-- VALUES
--   (current_tenant_id(), $1, $2)
-- RETURNING
--   *;
--
-- ============================================================
-- ACTIONS QUERIES
-- ============================================================

-- name: CreateAction :one
INSERT INTO actions (
  scope,
  name,
  display_name,
  description,
  action_type,
  action_category,
  risk_level,
  requires_approval,
  approver_role_id,
  is_active
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetActionByID :one
SELECT *
FROM actions
WHERE id = $1
LIMIT 1;

-- name: GetActionByName :one
SELECT *
FROM actions
WHERE name  = $1
  AND scope = $2
  AND (
    ($3::uuid IS NULL AND tenant_id IS NULL)
    OR tenant_id = $3
  )
LIMIT 1;

-- name: ListSystemActions :many
SELECT *
FROM actions
WHERE scope     = 'SYSTEM'
  AND is_active = TRUE
ORDER BY action_type, name;

-- name: ListTenantActions :many
SELECT *
FROM actions
WHERE tenant_id = $1
  AND scope     = 'TENANT'
  AND is_active = TRUE
ORDER BY action_type, name;

-- name: ListAllActionsForTenant :many
-- Returns all SYSTEM actions plus the given tenant's custom TENANT actions.
SELECT *
FROM actions
WHERE is_active = TRUE
  AND (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = $1)
  )
ORDER BY scope, action_type, name;

-- name: ListActionsByType :many
SELECT *
FROM actions
WHERE action_type = $1
  AND is_active   = TRUE
  AND (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = $2)
  )
ORDER BY name;

-- name: ListActionsByCategory :many
SELECT *
FROM actions
WHERE action_category = $1
  AND is_active       = TRUE
  AND (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = $2)
  )
ORDER BY action_type, name;

-- name: ListActionsByRiskLevel :many
SELECT *
FROM actions
WHERE risk_level = $1
  AND is_active  = TRUE
  AND (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = $2)
  )
ORDER BY action_type, name;

-- name: ListActionsRequiringApproval :many
SELECT *
FROM actions
WHERE requires_approval = TRUE
  AND is_active         = TRUE
  AND (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = $1)
  )
ORDER BY risk_level DESC, action_type, name;

-- name: ListActionsByApproverRole :many
-- All actions a given role is responsible for approving.
SELECT *
FROM actions
WHERE approver_role_id  = $1
  AND requires_approval = TRUE
  AND is_active         = TRUE
ORDER BY risk_level DESC, name;

-- name: UpdateAction :one
UPDATE actions
SET
  display_name      = $2,
  description       = $3,
  action_type       = $4,
  action_category   = $5,
  risk_level        = $6,
  requires_approval = $7,
  approver_role_id  = $8,
  is_active         = $9,
  updated_at        = NOW()
WHERE id = $1
RETURNING *;

-- name: SetActionApproval :one
-- Enable or disable approval requirement and assign/clear the approver role in one call.
UPDATE actions
SET
  requires_approval = $2,
  approver_role_id  = $3,   -- must be NOT NULL when $2 = TRUE (enforced by DB constraint)
  updated_at        = NOW()
WHERE id = $1
RETURNING *;

-- name: ActivateAction :one
UPDATE actions
SET
  is_active  = TRUE,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivateAction :one
UPDATE actions
SET
  is_active  = FALSE,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAction :exec
DELETE FROM actions
WHERE id = $1;

-- name: CountActionsByRiskLevel :many
-- Summary of active action counts grouped by risk level — useful for dashboards.
SELECT
  risk_level,
  COUNT(*) AS total
FROM actions
WHERE is_active = TRUE
  AND (
    scope = 'SYSTEM'
    OR (scope = 'TENANT' AND tenant_id = $1)
  )
GROUP BY risk_level
ORDER BY risk_level;
