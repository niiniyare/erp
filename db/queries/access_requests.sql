-- name: CreateAccessRequest :one
INSERT INTO access_requests (
    tenant_id, requester_id, target_user_id, entity_id, request_type,
    role_id, permission_id, resource_id, justification, business_reason,
    duration_hours, expires_at, auto_revoke
) VALUES (
    current_tenant_id(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetAccessRequestByID :one
SELECT * FROM access_requests 
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: UpdateAccessRequestStatus :one
UPDATE access_requests 
SET approval_status = $2, approved_by = $3, approved_at = $4,
    approval_comments = $5, duration_hours = $6, expires_at = $7, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id()
RETURNING *;
