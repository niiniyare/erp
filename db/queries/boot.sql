-- Boot schema queries — used by GET /schema/boot to build the AMIS app shell.
-- Run `make sqlc` after modifying this file.
--
-- Modules and resources are SYSTEM-scoped (tenant_id IS NULL) — no tenant
-- context needed.  Permission and feature-flag filtering is done in Go using
-- the pre-computed ResolvedSession (no extra DB round-trips).

-- name: ListActiveSystemModules :many
-- SELECT *
-- FROM   modules
-- WHERE  scope     = 'SYSTEM'
--   AND  is_active = TRUE
-- ORDER  BY nav_order ASC;
--
-- name: ListActiveResourcesByModule :many
SELECT *
FROM   resources
WHERE  module_id  = $1
  AND  is_active  = TRUE
  AND  deleted_at IS NULL
ORDER  BY nav_order ASC;
