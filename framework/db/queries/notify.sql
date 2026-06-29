-- name: CreateNotification :one
INSERT INTO notifications (id, tenant_id, recipient_id, channel, subject, body, status)
VALUES (gen_random_uuid(), current_tenant_id(), $1, $2, $3, $4, 'pending')
RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: ListPendingNotifications :many
SELECT * FROM notifications
WHERE tenant_id = current_tenant_id() AND status = 'pending'
ORDER BY created_at
LIMIT $1;

-- name: MarkNotificationSent :exec
UPDATE notifications
SET status = 'sent', sent_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();

-- name: MarkNotificationFailed :exec
UPDATE notifications
SET status = 'failed', error_msg = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = current_tenant_id();
