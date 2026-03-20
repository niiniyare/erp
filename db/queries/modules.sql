-- name: CreateModule :one
INSERT INTO modules (
    tenant_id, scope, slug, name, display_name,
    description, icon, nav_order, category, module_type, version
) VALUES (
    current_tenant_id(), $1, $2, $3, $4,
    $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetModuleBySlug :one
SELECT * FROM modules
WHERE slug = $1
  AND scope = 'SYSTEM'
  AND is_active = TRUE;

-- name: ListModules :many
SELECT * FROM modules
WHERE is_active = TRUE
  AND (scope = 'SYSTEM' OR tenant_id = current_tenant_id())
ORDER BY nav_order, name;

-- name: ListActiveSystemModules :many
SELECT * FROM modules
WHERE scope = 'SYSTEM' AND is_active = TRUE
ORDER BY nav_order;

-- ============================================================
-- NAV QUERY — powers BootService.BuildAppShell
-- Returns only modules+resources that are enabled for the tenant
-- (flag = true or no override and default_value = true).
-- Permission filter is applied in Go after this returns.
-- ============================================================

-- name: ListEnabledModulesWithResources :many
SELECT
    m.id            AS module_id,
    m.slug          AS module_slug,
    m.name          AS module_name,
    m.display_name  AS module_display_name,
    m.icon,
    m.nav_order     AS module_nav_order,
    r.id            AS resource_id,
    r.slug          AS resource_slug,
    r.name          AS resource_name,
    r.display_name  AS resource_display_name,
    r.nav_url,
    r.nav_order     AS resource_nav_order
FROM modules m
JOIN resources r ON r.module_id = m.id AND r.is_active = TRUE AND r.deleted_at IS NULL

-- Module flag: must be enabled for this tenant
JOIN feature_flag_definitions mfd
    ON mfd.module_id = m.id AND mfd.resource_id IS NULL
LEFT JOIN tenant_feature_flags mtf
    ON mtf.flag_id = mfd.id AND mtf.tenant_id = $1

-- Resource flag: must be enabled (or have no override and default true)
LEFT JOIN feature_flag_definitions rfd ON rfd.resource_id = r.id
LEFT JOIN tenant_feature_flags rtf
    ON rtf.flag_id = rfd.id AND rtf.tenant_id = $1

WHERE m.is_active = TRUE
  AND m.scope = 'SYSTEM'
  AND COALESCE(mtf.enabled, mfd.default_value) = TRUE
  AND (rfd.id IS NULL OR COALESCE(rtf.enabled, rfd.default_value) = TRUE)
ORDER BY m.nav_order, r.nav_order;
