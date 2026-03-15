-- name: UpsertRoleAssignment :exec
INSERT INTO role_assignments
    (id, tenant_id, subject, role_name, domain, assigned_by, delegated_by, expires_at, is_active)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
ON CONFLICT (subject, role_name, domain)
DO UPDATE SET
    is_active    = TRUE,
    assigned_by  = EXCLUDED.assigned_by,
    delegated_by = EXCLUDED.delegated_by,
    expires_at   = EXCLUDED.expires_at;

-- name: DeactivateRoleAssignment :exec
UPDATE role_assignments
SET    is_active = FALSE
WHERE  subject = $1 AND role_name = $2 AND domain = $3;

-- name: ListRoleAssignments :many
SELECT
    id, tenant_id, subject, role_name, domain,
    assigned_by, delegated_by, expires_at, is_active, created_at
FROM   role_assignments
WHERE  subject = $1 AND domain = $2
ORDER  BY created_at DESC;

-- name: ListExpiredActiveRoleNames :many
SELECT role_name
FROM   role_assignments
WHERE  subject    = $1
  AND  domain     = $2
  AND  is_active  = TRUE
  AND  expires_at IS NOT NULL
  AND  expires_at < NOW();
