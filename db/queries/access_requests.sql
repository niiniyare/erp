-- name: CreateAccessRequest :one
INSERT INTO
  access_requests (
    tenant_id,
    requester_id,
    target_user_id,
    entity_id,
    request_type,
    role_id,
    permission_id,
    resource_id,
    justification,
    business_reason,
    duration_hours,
    expires_at,
    auto_revoke
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
  )
RETURNING
  *;

-- name: GetAccessRequestByID :one
SELECT
  *
FROM
  access_requests
WHERE
  id = $1
  AND tenant_id = current_tenant_id();

-- name: UpdateAccessRequestStatus :one
UPDATE
  access_requests
SET
  approval_status = $2,
  approved_by = $3,
  approved_at = $4,
  approval_comments = $5,
  duration_hours = $6,
  expires_at = $7,
  updated_at = NOW()
WHERE
  id = $1
  AND tenant_id = current_tenant_id()
RETURNING
  *;

-- name: ListAccessRequestsByStatus :many
SELECT
  *
FROM
  access_requests
WHERE
  tenant_id = current_tenant_id()
  AND (
    $1::text IS NULL
    OR approval_status = $1
  )
  AND (
    $2::uuid IS NULL
    OR requester_id = $2
  )
  AND (
    $3::uuid IS NULL
    OR target_user_id = $3
  )
ORDER BY
  created_at DESC
LIMIT
  $4 OFFSET $5;

-- name: GetPendingAccessRequests :many
SELECT
  *
FROM
  access_requests
WHERE
  tenant_id = current_tenant_id()
  AND approval_status = 'PENDING'
  AND (
    expires_at IS NULL
    OR expires_at > NOW()
  )
ORDER BY
  created_at ASC;

-- name: GetExpiredAccessRequests :many
SELECT
  *
FROM
  access_requests
WHERE
  tenant_id = current_tenant_id()
  AND approval_status = 'APPROVED'
  AND expires_at IS NOT NULL
  AND expires_at < NOW()
  AND auto_revoke = TRUE;

-- name: GetUserAccessRequestHistory :many
SELECT
  *
FROM
  access_requests
WHERE
  tenant_id = current_tenant_id()
  AND requester_id = $1
ORDER BY
  created_at DESC
LIMIT
  $2 OFFSET $3;

-- name: CountAccessRequestsByStatus :one
SELECT
  COUNT(*) AS total_requests,
  SUM(
    CASE
      WHEN approval_status = 'PENDING' THEN 1
      ELSE 0
    END
  ) AS pending_requests,
  SUM(
    CASE
      WHEN approval_status = 'APPROVED' THEN 1
      ELSE 0
    END
  ) AS approved_requests,
  SUM(
    CASE
      WHEN approval_status = 'REJECTED' THEN 1
      ELSE 0
    END
  ) AS rejected_requests,
  SUM(
    CASE
      WHEN approval_status = 'EXPIRED' THEN 1
      ELSE 0
    END
  ) AS expired_requests
FROM
  access_requests
WHERE
  tenant_id = current_tenant_id()
  AND created_at >= $1
  AND created_at <= $2;
