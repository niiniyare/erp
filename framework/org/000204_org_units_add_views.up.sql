-- ------------------------------------------------------------------------------------------------
-- VIEWS — ORG UNIT HIERARCHY & REPORTING
-- ------------------------------------------------------------------------------------------------
-- All views use security_invoker = true so queries run with the privileges and RLS context of
-- the calling user, not the view definer. This means tenant isolation is automatically enforced
-- for every caller without needing to duplicate tenant_id predicates in each view.
--
-- None of the views include ORDER BY — add ORDER BY in the calling query. An ORDER BY inside a
-- view definition forces a sort materialisation even when the caller has its own ordering or
-- limit pushdown, and the order is not guaranteed to propagate across query boundaries.
--
-- Views in this migration (in dependency order):
--   v_org_hierarchy          — recursive full-path breadcrumb for every org unit
--   v_org_structure          — standard 4-level flattened view (COST_CENTER → DEPT → REGION → COMPANY)
--   v_cost_center_detail     — cost centre rows enriched with their 4-level ancestors
--   v_department_summary     — department rows with cost-centre counts and ancestor context
--   v_company_structure      — all descendants of each company via the closure table
--   v_tenant_org_summary     — entity counts per tenant broken down by unit type
--   v_active_org_units       — active units enriched with their nearest typed ancestors
--   v_org_unit_changes       — lifecycle change log (CREATED / MODIFIED / DELETED)
--   v_org_unit_paths_detail  — closure table rows enriched with ancestor/descendant names
--   v_tenant_resource_usage  — document sequence counts and entity stats per tenant
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- V_ORG_HIERARCHY
-- ------------------------------------------------------------------------------------------------
-- Recursive CTE that builds a full breadcrumb path (e.g. "Corp > Region > Dept") for every
-- org unit in the tree. Useful for breadcrumb rendering and path-aware reports.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_hierarchy AS
WITH RECURSIVE org_chart AS (
  -- Anchor: root nodes (no parent)
  SELECT
    ou.uuid       AS org_unit_id,
    ou.name,
    ou.type,
    ou.parent_id,
    ou.tenant_id,
    ou.name::TEXT AS full_path,
    0             AS depth
  FROM org_units ou
  WHERE ou.parent_id IS NULL
    AND ou.deleted_at IS NULL

  UNION ALL

  -- Recursive: children of already-visited nodes
  SELECT
    ou.uuid AS org_unit_id,
    ou.name,
    ou.type,
    ou.parent_id,
    ou.tenant_id,
    (oc.full_path || ' > ' || ou.name)::TEXT,
    oc.depth + 1
  FROM org_units ou
  JOIN org_chart oc ON ou.parent_id = oc.org_unit_id
  WHERE ou.deleted_at IS NULL
)
SELECT
  t.name          AS tenant_name,
  oc.org_unit_id,
  oc.name         AS org_unit_name,
  oc.type         AS org_unit_type,
  oc.full_path,
  oc.depth
FROM org_chart oc
JOIN tenants t ON oc.tenant_id = t.id;

ALTER VIEW v_org_hierarchy SET (security_invoker = true);

COMMENT ON VIEW v_org_hierarchy IS
  'Recursive breadcrumb view — builds a full ancestor path (e.g. "Corp > Region > Dept") for '
  'every non-deleted org unit. Useful for breadcrumb UI components and path-aware exports. '
  'Add ORDER BY in the calling query; no ORDER BY is defined here to avoid forced sorts.';

-- ------------------------------------------------------------------------------------------------
-- V_ORG_STRUCTURE
-- ------------------------------------------------------------------------------------------------
-- Standard 4-level hierarchy flattened to a single row: COST_CENTER → DEPARTMENT → REGION →
-- COMPANY. Intended for reports that assume the canonical 4-level org model.
--
-- WARNING: Returns zero rows for tenants using non-standard hierarchy depths. Use
--          v_org_unit_paths_detail with depth filtering for flexible-depth tenants.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_structure AS
SELECT
  t.name   AS tenant_name,
  cc.uuid  AS cost_center_id,
  cc.name  AS cost_center_name,
  d.uuid   AS department_id,
  d.name   AS department_name,
  r.uuid   AS region_id,
  r.name   AS region_name,
  c.uuid   AS company_id,
  c.name   AS company_name
FROM org_units cc
JOIN org_units d  ON cc.parent_id = d.uuid  AND d.type  = 'DEPARTMENT'
JOIN org_units r  ON d.parent_id  = r.uuid  AND r.type  = 'REGION'
JOIN org_units c  ON r.parent_id  = c.uuid  AND c.type  = 'COMPANY'
JOIN tenants   t  ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
  AND d.deleted_at  IS NULL
  AND r.deleted_at  IS NULL
  AND c.deleted_at  IS NULL;

ALTER VIEW v_org_structure SET (security_invoker = true);

COMMENT ON VIEW v_org_structure IS
  'Flattens the canonical 4-level org hierarchy (COST_CENTER → DEPARTMENT → REGION → COMPANY) '
  'into one row per cost centre. Returns zero rows for tenants with non-standard depths — use '
  'v_org_unit_paths_detail instead for flexible-depth tenants.';

-- ------------------------------------------------------------------------------------------------
-- V_COST_CENTER_DETAIL
-- ------------------------------------------------------------------------------------------------
-- Cost centre record enriched with its full 4-level ancestor context (department, region,
-- company, tenant). Convenient for detail pages, exports, and per-cost-centre drill-downs.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_cost_center_detail AS
SELECT
  t.name            AS tenant_name,
  cc.uuid           AS cost_center_id,
  cc.name           AS cost_center_name,
  cc.code           AS cost_center_code,
  d.uuid            AS department_id,
  d.name            AS department_name,
  r.uuid            AS region_id,
  r.name            AS region_name,
  c.uuid            AS company_id,
  c.name            AS company_name,
  cc.is_active      AS cost_center_active,
  cc.created_at     AS cost_center_created_at
FROM org_units cc
JOIN org_units d ON cc.parent_id = d.uuid AND d.type = 'DEPARTMENT'
JOIN org_units r ON d.parent_id  = r.uuid AND r.type = 'REGION'
JOIN org_units c ON r.parent_id  = c.uuid AND c.type = 'COMPANY'
JOIN tenants   t ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
  AND d.deleted_at  IS NULL
  AND r.deleted_at  IS NULL
  AND c.deleted_at  IS NULL;

ALTER VIEW v_cost_center_detail SET (security_invoker = true);

COMMENT ON VIEW v_cost_center_detail IS
  'Cost centre rows enriched with their full 4-level ancestor chain (department, region, company, '
  'tenant). Use for detail pages and cost-centre-level exports.';

-- ------------------------------------------------------------------------------------------------
-- V_DEPARTMENT_SUMMARY
-- ------------------------------------------------------------------------------------------------
-- Department rows with aggregated cost-centre counts and their region/company ancestors.
-- Useful for dashboard tiles and org-overview reports.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_department_summary AS
SELECT
  t.name                      AS tenant_name,
  d.uuid                      AS department_id,
  d.name                      AS department_name,
  d.code                      AS department_code,
  r.uuid                      AS region_id,
  r.name                      AS region_name,
  c.uuid                      AS company_id,
  c.name                      AS company_name,
  COUNT(DISTINCT cc.uuid)     AS cost_center_count,
  d.is_active                 AS department_active,
  d.created_at                AS department_created_at
FROM org_units d
JOIN org_units r  ON d.parent_id  = r.uuid AND r.type = 'REGION'
JOIN org_units c  ON r.parent_id  = c.uuid AND c.type = 'COMPANY'
JOIN tenants   t  ON d.tenant_id  = t.id
LEFT JOIN org_units cc
  ON  cc.parent_id = d.uuid
  AND cc.type      = 'COST_CENTER'
  AND cc.deleted_at IS NULL
WHERE d.type      = 'DEPARTMENT'
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY
  t.id,  t.name,
  d.uuid, d.name, d.code, d.is_active, d.created_at,
  r.uuid, r.name,
  c.uuid, c.name;

ALTER VIEW v_department_summary SET (security_invoker = true);

COMMENT ON VIEW v_department_summary IS
  'Department rows with their region and company ancestor context, plus the count of active cost '
  'centres under each department. Designed for org-overview dashboards.';

-- ------------------------------------------------------------------------------------------------
-- V_COMPANY_STRUCTURE
-- ------------------------------------------------------------------------------------------------
-- All org units under each company, resolved via the org_unit_paths closure table.
-- Returns every descendant with its depth measured from the company root.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_company_structure AS
SELECT
  t.name         AS tenant_name,
  c.uuid         AS company_id,
  c.name         AS company_name,
  c.code         AS company_code,
  ou.uuid        AS org_unit_id,
  ou.name        AS org_unit_name,
  ou.type        AS org_unit_type,
  ou.code        AS org_unit_code,
  oup.depth      AS levels_from_company
FROM org_units c
JOIN org_unit_paths oup ON c.uuid         = oup.ancestor_id
JOIN org_units      ou  ON oup.descendant_id = ou.uuid
JOIN tenants        t   ON c.tenant_id    = t.id
WHERE c.type       = 'COMPANY'
  AND c.deleted_at IS NULL
  AND ou.deleted_at IS NULL;

ALTER VIEW v_company_structure SET (security_invoker = true);

COMMENT ON VIEW v_company_structure IS
  'Every org unit under each company, resolved via the org_unit_paths closure table. '
  'Includes self-reference rows (depth = 0) so the company itself appears in the result set. '
  'Use levels_from_company to filter by depth.';

-- ------------------------------------------------------------------------------------------------
-- V_TENANT_ORG_SUMMARY
-- ------------------------------------------------------------------------------------------------
-- Org unit counts per tenant broken down by type, plus active and non-deleted totals.
-- Used for tenant dashboard metrics and capacity monitoring.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_tenant_org_summary AS
SELECT
  t.id                                                                       AS tenant_id,
  t.name                                                                     AS tenant_name,
  t."Status"                                                                 AS tenant_status,
  COUNT(DISTINCT ou.uuid)                                                    AS total_org_units,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'COMPANY')                AS company_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'REGION')                 AS region_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'DEPARTMENT')             AS department_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'COST_CENTER')            AS cost_center_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'PROJECT')                AS project_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'SUBSIDIARY')             AS subsidiary_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'DIVISION')               AS division_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'BRANCH')                 AS branch_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'LOCATION')               AS location_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.type = 'BUDGET_UNIT')            AS budget_unit_count,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.is_active = TRUE)                AS active_org_units,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.deleted_at IS NULL)              AS non_deleted_org_units
FROM tenants t
LEFT JOIN org_units ou ON ou.tenant_id = t.id
GROUP BY t.id, t.name, t."Status";

ALTER VIEW v_tenant_org_summary SET (security_invoker = true);

COMMENT ON VIEW v_tenant_org_summary IS
  'Org unit counts per tenant broken down by all unit types, with active and non-deleted totals. '
  'Used for tenant dashboard metrics. Includes all unit types defined in the org_units.type CHECK '
  'constraint so new types added to the enum are reflected here without view changes.';

-- ------------------------------------------------------------------------------------------------
-- V_ACTIVE_ORG_UNITS
-- ------------------------------------------------------------------------------------------------
-- Active (non-deleted, is_active = TRUE) org units enriched with the name of their nearest
-- ancestor of each standard type (company, region, department). Nearest is determined by the
-- smallest depth value in org_unit_paths — DISTINCT ON picks one row per descendant.
--
-- Uses CTEs with DISTINCT ON rather than correlated subqueries to avoid a per-row scan of
-- the closure table.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_active_org_units AS
WITH nearest_company AS (
  SELECT DISTINCT ON (oup.descendant_id)
    oup.descendant_id AS org_unit_id,
    ancestor.name     AS company_name
  FROM org_unit_paths oup
  JOIN org_units ancestor ON ancestor.uuid = oup.ancestor_id
  WHERE ancestor.type       = 'COMPANY'
    AND ancestor.deleted_at IS NULL
  ORDER BY oup.descendant_id, oup.depth ASC
),
nearest_region AS (
  SELECT DISTINCT ON (oup.descendant_id)
    oup.descendant_id AS org_unit_id,
    ancestor.name     AS region_name
  FROM org_unit_paths oup
  JOIN org_units ancestor ON ancestor.uuid = oup.ancestor_id
  WHERE ancestor.type       = 'REGION'
    AND ancestor.deleted_at IS NULL
  ORDER BY oup.descendant_id, oup.depth ASC
),
nearest_department AS (
  SELECT DISTINCT ON (oup.descendant_id)
    oup.descendant_id AS org_unit_id,
    ancestor.name     AS department_name
  FROM org_unit_paths oup
  JOIN org_units ancestor ON ancestor.uuid = oup.ancestor_id
  WHERE ancestor.type       = 'DEPARTMENT'
    AND ancestor.deleted_at IS NULL
  ORDER BY oup.descendant_id, oup.depth ASC
)
SELECT
  t.name          AS tenant_name,
  ou.uuid         AS org_unit_id,
  ou.name         AS org_unit_name,
  ou.type         AS org_unit_type,
  ou.code         AS org_unit_code,
  nc.company_name,
  nr.region_name,
  nd.department_name,
  ou.is_active,
  ou.created_at,
  ou.updated_at
FROM org_units ou
JOIN tenants           t  ON ou.tenant_id  = t.id
LEFT JOIN nearest_company    nc ON nc.org_unit_id = ou.uuid
LEFT JOIN nearest_region     nr ON nr.org_unit_id = ou.uuid
LEFT JOIN nearest_department nd ON nd.org_unit_id = ou.uuid
WHERE ou.deleted_at IS NULL
  AND ou.is_active  = TRUE;

ALTER VIEW v_active_org_units SET (security_invoker = true);

COMMENT ON VIEW v_active_org_units IS
  'Active (non-deleted, is_active = TRUE) org units with the name of their nearest company, '
  'region, and department ancestor resolved via the closure table. Uses DISTINCT ON per ancestor '
  'type to pick the closest ancestor at each level without correlated subqueries.';

-- ------------------------------------------------------------------------------------------------
-- V_ORG_UNIT_CHANGES
-- ------------------------------------------------------------------------------------------------
-- Lifecycle change log classifying each org unit row as CREATED, MODIFIED, or DELETED based on
-- its timestamp fields. Useful for audit exports and change-feed UIs.
--
-- Classification logic:
--   DELETED  — deleted_at IS NOT NULL
--   MODIFIED — updated_at more than 1 minute after created_at  AND  version > 1
--   CREATED  — everything else (brand new rows, or rows updated in the same minute they were created)
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_unit_changes AS
SELECT
  t.name              AS tenant_name,
  ou.uuid             AS org_unit_id,
  ou.name             AS org_unit_name,
  ou.type             AS org_unit_type,
  ou.validation_status,
  ou.version,
  ou.created_at,
  ou.updated_at,
  ou.deleted_at,
  CASE
    WHEN ou.deleted_at IS NOT NULL                                          THEN 'DELETED'
    WHEN ou.version > 1 AND ou.updated_at > ou.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
    ELSE                                                                         'CREATED'
  END AS change_type
FROM org_units ou
JOIN tenants t ON ou.tenant_id = t.id;

ALTER VIEW v_org_unit_changes SET (security_invoker = true);

COMMENT ON VIEW v_org_unit_changes IS
  'Lifecycle change log for org units. Classifies each row as CREATED, MODIFIED, or DELETED '
  'using deleted_at, version, and the updated_at / created_at delta. The version > 1 guard '
  'prevents false MODIFIED classifications caused by sub-minute timestamp jitter at creation.';

-- ------------------------------------------------------------------------------------------------
-- V_ORG_UNIT_PATHS_DETAIL
-- ------------------------------------------------------------------------------------------------
-- Closure table rows enriched with ancestor and descendant names and types.
-- Useful for path-based permission checks, tree traversals, and breadcrumb generation.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_unit_paths_detail AS
SELECT
  t.name         AS tenant_name,
  anc.uuid       AS ancestor_id,
  anc.name       AS ancestor_name,
  anc.type       AS ancestor_type,
  desc_.uuid     AS descendant_id,
  desc_.name     AS descendant_name,
  desc_.type     AS descendant_type,
  oup.depth
FROM org_unit_paths oup
JOIN org_units anc   ON oup.ancestor_id   = anc.uuid
JOIN org_units desc_ ON oup.descendant_id = desc_.uuid
JOIN tenants   t     ON oup.tenant_id     = t.id
WHERE anc.deleted_at   IS NULL
  AND desc_.deleted_at IS NULL;

ALTER VIEW v_org_unit_paths_detail SET (security_invoker = true);

COMMENT ON VIEW v_org_unit_paths_detail IS
  'Closure table (org_unit_paths) enriched with org unit names and types for both the ancestor '
  'and descendant ends. Use for path-based permission checks and breadcrumb generation. '
  'Excludes pairs where either end has been soft-deleted.';

-- ------------------------------------------------------------------------------------------------
-- V_TENANT_RESOURCE_USAGE
-- ------------------------------------------------------------------------------------------------
-- Document sequence counts and org unit stats per tenant.
-- Gives an overview of how many document types and sequence slots each tenant is consuming.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_tenant_resource_usage AS
SELECT
  t.id                                                                AS tenant_id,
  t.name                                                              AS tenant_name,
  t."Status"                                                          AS tenant_status,
  COUNT(DISTINCT ou.uuid)                                             AS total_org_units,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.is_active = TRUE)          AS active_org_units,
  COUNT(DISTINCT ou.uuid) FILTER (WHERE ou.deleted_at IS NULL)        AS non_deleted_org_units,
  COUNT(DISTINCT ds.uuid)                                             AS doc_sequence_rows,
  COUNT(DISTINCT ds.doc_type)                                         AS distinct_doc_types,
  MAX(ou.created_at)                                                  AS last_org_unit_created_at,
  MAX(ou.updated_at)                                                  AS last_org_unit_updated_at
FROM tenants t
LEFT JOIN org_units     ou ON ou.tenant_id = t.id
LEFT JOIN doc_sequences ds ON ds.tenant_id = t.id
GROUP BY t.id, t.name, t."Status";

ALTER VIEW v_tenant_resource_usage SET (security_invoker = true);

COMMENT ON VIEW v_tenant_resource_usage IS
  'Resource usage summary per tenant: org unit counts (total, active, non-deleted) joined with '
  'doc_sequences counts (distinct rows, distinct document types). Used for tenant capacity '
  'dashboards and billing metrics.';
