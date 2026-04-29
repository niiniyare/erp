-- =====================================================================
-- FINANCE MODULE — APPROVAL WORKFLOW QUERIES
-- =====================================================================

-- name: InsertWorkflowRecord :exec
INSERT INTO finance_workflow_records
  (tenant_id, transaction_id, status, current_tier, initiated_by, due_at)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5);

-- name: GetWorkflowByTransaction :one
SELECT id, transaction_id, status, current_tier, initiated_by, due_at, created_at, updated_at
FROM   finance_workflow_records
WHERE  tenant_id      = current_tenant_id()
  AND  transaction_id = $1
LIMIT  1;

-- name: UpdateWorkflowStatus :exec
UPDATE finance_workflow_records
SET    status       = $2,
       current_tier = $3,
       updated_at   = NOW()
WHERE  tenant_id = current_tenant_id()
  AND  id        = $1;

-- name: InsertApprovalHistory :exec
INSERT INTO finance_approval_history
  (tenant_id, workflow_id, transaction_id, tier, action, performed_by, notes)
VALUES
  (current_tenant_id(), $1, $2, $3, $4, $5, $6);

-- name: GetApprovalHistoryByTransaction :many
SELECT id, workflow_id, transaction_id, tier, action, performed_by, notes, created_at
FROM   finance_approval_history
WHERE  tenant_id      = current_tenant_id()
  AND  transaction_id = $1
ORDER  BY created_at ASC;

-- name: GetPendingWorkflowsByUser :many
-- Returns pending/in-progress workflow records initiated by a specific user.
-- Filter pushed to SQL to avoid full-tenant fetch on every dashboard load.
-- Paginated: callers must supply limit/offset.
SELECT id, transaction_id, status, current_tier, initiated_by, due_at, created_at, updated_at
FROM   finance_workflow_records
WHERE  tenant_id    = current_tenant_id()
  AND  status       IN ('pending', 'in_progress')
  AND  initiated_by = $1
ORDER  BY created_at ASC
LIMIT  $2
OFFSET $3;
