-- name: CreateAuditEvent :one
INSERT INTO audit_log (
  tenant_id,
  user_id,
  event_type,
  event_category,
  severity,
  entity_id,
  decision,
  reason,
  context
) VALUES (
  current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;