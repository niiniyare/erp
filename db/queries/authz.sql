-- name: UpsertRoleAssignment :exec
INSERT INTO role_assignments (
    id, tenant_id, subject, role_name, role_slug, domain,
    assigned_by, granted_by, delegated_by, expires_at, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, TRUE
) ON CONFLICT (subject, role_name, domain)
DO UPDATE SET
    is_active    = TRUE,
    role_slug    = EXCLUDED.role_slug,
    assigned_by  = EXCLUDED.assigned_by,
    granted_by   = EXCLUDED.granted_by,
    delegated_by = EXCLUDED.delegated_by,
    expires_at   = EXCLUDED.expires_at,
    revoked_at   = NULL,
    revoke_reason = NULL;

-- name: RevokeRoleAssignment :exec
-- Explicit revocation — sets revoked_at + reason, distinct from expiry.
UPDATE role_assignments
SET    is_active     = FALSE,
       revoked_at    = NOW(),
       revoke_reason = $4
WHERE  subject   = $1
  AND  role_name = $2
  AND  domain    = $3;

-- name: DeactivateRoleAssignment :exec
UPDATE role_assignments
SET    is_active = FALSE
WHERE  subject = $1 AND role_name = $2 AND domain = $3;

-- name: ListRoleAssignments :many
SELECT
    id, tenant_id, subject, role_name, role_slug, domain,
    assigned_by, granted_by, delegated_by,
    expires_at, revoked_at, revoke_reason,
    is_active, created_at
FROM   role_assignments
WHERE  subject = $1 AND domain = $2
ORDER  BY created_at DESC;

-- name: ListActiveRoleAssignments :many
-- Returns roles that are currently active and not expired — used by permission computation.
SELECT
    id, subject, role_name, role_slug, domain,
    expires_at, granted_by
FROM   role_assignments
WHERE  subject    = $1
  AND  domain     = $2
  AND  is_active  = TRUE
  AND  revoked_at IS NULL
  AND  (expires_at IS NULL OR expires_at > NOW())
ORDER  BY created_at DESC;

-- name: ListExpiredActiveRoleNames :many
SELECT role_name
FROM   role_assignments
WHERE  subject    = $1
  AND  domain     = $2
  AND  is_active  = TRUE
  AND  expires_at IS NOT NULL
  AND  expires_at < NOW();

-- name: UserHasRole :one
-- Guard 3 delegation check: does the granter hold this role?
SELECT EXISTS (
    SELECT 1 FROM role_assignments
    WHERE subject   = $1
      AND role_name = $2
      AND domain    = $3
      AND is_active = TRUE
      AND revoked_at IS NULL
      AND (expires_at IS NULL OR expires_at > NOW())
) AS has_role;

-- name: ListAssignmentHistory :many
-- Audit trail: all assignments including revoked/expired.
SELECT
    id, subject, role_name, role_slug, domain,
    assigned_by, granted_by, delegated_by,
    expires_at, revoked_at, revoke_reason,
    is_active, created_at
FROM   role_assignments
WHERE  tenant_id = $1
  AND  ($2::text IS NULL OR subject = $2)
ORDER  BY created_at DESC
LIMIT  $3 OFFSET $4;
