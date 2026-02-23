-- =====================================================================
-- ENTITIES VIEWS - Reporting and analytical views for entity management
-- =====================================================================

-- =====================================================================
-- v_tenant_hierarchy: Recursive path view for the org chart
-- =====================================================================
CREATE VIEW v_tenant_hierarchy AS
WITH RECURSIVE org_chart AS (
  SELECT
    e.uuid       AS entity_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    e.name::TEXT AS path,
    0            AS depth
  FROM entities e
  WHERE e.parent_id IS NULL
    AND e.deleted_at IS NULL

  UNION ALL

  SELECT
    e.uuid AS entity_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    (oc.path || ' > ' || e.name)::TEXT,
    oc.depth + 1
  FROM entities e
  JOIN org_chart oc ON e.parent_id = oc.entity_id
  WHERE e.deleted_at IS NULL
)
SELECT
  t.name    AS tenant_name,
  oc.entity_id,
  oc.name   AS entity_name,
  oc.type   AS entity_type,
  oc.path   AS full_path,
  oc.depth
FROM org_chart oc
JOIN tenants t ON oc.tenant_id = t.id;

-- =====================================================================
-- v_entity_structure: 4-level canonical view (COST_CENTER→DEPT→REGION→COMPANY)
-- NOTE: Documents the canonical 4-level hierarchy. Tenants with different
-- structures should query hierarchy_paths directly with the depth column.
-- =====================================================================
CREATE VIEW v_entity_structure AS
SELECT
  t.name   AS tenant_name,
  cc.uuid  AS cost_center_id,
  cc.name  AS cost_center,
  d.uuid   AS department_id,
  d.name   AS department,
  r.uuid   AS regional_id,
  r.name   AS regional,
  c.uuid   AS company_id,
  c.name   AS company
FROM entities cc
JOIN entities d  ON cc.parent_id = d.uuid  AND d.type  = 'DEPARTMENT'
JOIN entities r  ON d.parent_id  = r.uuid  AND r.type  IN ('REGION', 'REGIONAL')
JOIN entities c  ON r.parent_id  = c.uuid  AND c.type  = 'COMPANY'
JOIN tenants  t  ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
  AND d.deleted_at  IS NULL
  AND r.deleted_at  IS NULL
  AND c.deleted_at  IS NULL;

-- =====================================================================
-- v_cost_center_info: Cost centre with full 4-level context
-- =====================================================================
CREATE VIEW v_cost_center_info AS
SELECT
  t.name            AS tenant_name,
  cc.uuid           AS cost_center_id,
  cc.name           AS cost_center,
  cc.code           AS cost_center_code,
  d.uuid            AS department_id,
  d.name            AS department,
  r.uuid            AS regional_id,
  r.name            AS regional,
  c.uuid            AS company_id,
  c.name            AS company,
  cc.is_active      AS cost_center_active,
  cc.created_at     AS cost_center_created
FROM entities cc
JOIN entities d ON cc.parent_id = d.uuid
JOIN entities r ON d.parent_id  = r.uuid
JOIN entities c ON r.parent_id  = c.uuid
JOIN tenants  t ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND d.type  = 'DEPARTMENT'
  AND r.type  IN ('REGION', 'REGIONAL')
  AND c.type  = 'COMPANY'
  AND cc.deleted_at IS NULL
  AND d.deleted_at  IS NULL
  AND r.deleted_at  IS NULL
  AND c.deleted_at  IS NULL;

-- =====================================================================
-- v_department_summary: Departments with cost-centre counts
-- =====================================================================
CREATE VIEW v_department_summary AS
SELECT
  t.name                      AS tenant_name,
  d.uuid                      AS department_id,
  d.name                      AS department_name,
  d.code                      AS department_code,
  r.uuid                      AS regional_id,
  r.name                      AS regional_name,
  c.uuid                      AS company_id,
  c.name                      AS company_name,
  COUNT(DISTINCT cc.uuid)     AS cost_center_count,
  d.is_active                 AS department_active,
  d.created_at                AS department_created
FROM entities d
JOIN entities r  ON d.parent_id  = r.uuid
JOIN entities c  ON r.parent_id  = c.uuid
JOIN tenants  t  ON d.tenant_id  = t.id
LEFT JOIN entities cc ON cc.parent_id = d.uuid
                      AND cc.type      = 'COST_CENTER'
                      AND cc.deleted_at IS NULL
WHERE d.type = 'DEPARTMENT'
  AND r.type IN ('REGION', 'REGIONAL')
  AND c.type = 'COMPANY'
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY
  t.id, t.name,
  d.uuid, d.name, d.code, d.is_active, d.created_at,
  r.uuid, r.name,
  c.uuid, c.name;

-- =====================================================================
-- v_company_structure: All entities under each company via closure table
-- =====================================================================
CREATE VIEW v_company_structure AS
SELECT
  t.name        AS tenant_name,
  c.uuid        AS company_id,
  c.name        AS company_name,
  c.code        AS company_code,
  e.uuid        AS entity_id,
  e.name        AS entity_name,
  e.type        AS entity_type,
  e.code        AS entity_code,
  hp.depth      AS levels_from_company
FROM entities c
JOIN hierarchy_paths hp ON c.uuid   = hp.ancestor_id
JOIN entities        e  ON hp.descendant_id = e.uuid
JOIN tenants         t  ON c.tenant_id      = t.id
WHERE c.type         = 'COMPANY'
  AND c.deleted_at   IS NULL
  AND e.deleted_at   IS NULL;

-- =====================================================================
-- v_tenant_entity_summary: Entity counts per tenant
-- =====================================================================
CREATE VIEW v_tenant_entity_summary AS
SELECT
  t.id          AS tenant_id,
  t.name        AS tenant_name,
  t."Status"    AS tenant_status,
  COUNT(DISTINCT e.uuid)                                                          AS total_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'COMPANY')                       AS company_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type IN ('REGION', 'REGIONAL'))         AS regional_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'DEPARTMENT')                    AS department_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'COST_CENTER')                   AS cost_center_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'PROJECT')                       AS project_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.is_active = TRUE)                       AS active_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.deleted_at IS NULL)                     AS non_deleted_entities
FROM tenants t
LEFT JOIN entities e ON e.tenant_id = t.id
GROUP BY t.id, t.name, t."Status";

-- =====================================================================
-- v_active_entities: Active entities with their typed ancestors
--
-- Uses CTEs with DISTINCT ON to find the NEAREST ancestor of each type
-- via the closure table — avoids the CASE WHEN EXISTS correlated-subquery
-- anti-pattern that forces a per-row subquery evaluation.
-- =====================================================================
CREATE VIEW v_active_entities AS
WITH company_anc AS (
  -- Nearest COMPANY ancestor for each entity (depth ASC picks closest)
  SELECT DISTINCT ON (hp.descendant_id)
    hp.descendant_id,
    a.name
  FROM hierarchy_paths hp
  JOIN entities a ON a.uuid = hp.ancestor_id
  WHERE a.type = 'COMPANY'
    AND a.deleted_at IS NULL
  ORDER BY hp.descendant_id, hp.depth ASC
),
regional_anc AS (
  SELECT DISTINCT ON (hp.descendant_id)
    hp.descendant_id,
    a.name
  FROM hierarchy_paths hp
  JOIN entities a ON a.uuid = hp.ancestor_id
  WHERE a.type IN ('REGION', 'REGIONAL')
    AND a.deleted_at IS NULL
  ORDER BY hp.descendant_id, hp.depth ASC
),
dept_anc AS (
  SELECT DISTINCT ON (hp.descendant_id)
    hp.descendant_id,
    a.name
  FROM hierarchy_paths hp
  JOIN entities a ON a.uuid = hp.ancestor_id
  WHERE a.type = 'DEPARTMENT'
    AND a.deleted_at IS NULL
  ORDER BY hp.descendant_id, hp.depth ASC
)
SELECT
  t.name        AS tenant_name,
  e.uuid        AS entity_id,
  e.name        AS entity_name,
  e.type        AS entity_type,
  e.code        AS entity_code,
  ca.name       AS company_name,
  ra.name       AS regional_name,
  da.name       AS department_name,
  e.is_active,
  e.created_at,
  e.updated_at
FROM entities e
JOIN tenants      t  ON e.tenant_id      = t.id
LEFT JOIN company_anc  ca ON ca.descendant_id = e.uuid
LEFT JOIN regional_anc ra ON ra.descendant_id = e.uuid
LEFT JOIN dept_anc     da ON da.descendant_id = e.uuid
WHERE e.deleted_at IS NULL
  AND e.is_active  = TRUE;

-- =====================================================================
-- v_entity_changes: Lifecycle change log
-- NOTE: No ORDER BY — callers add ORDER BY as needed.
--       ORDER BY inside a view definition is not guaranteed to propagate
--       and forces a sort materialisation even when unused by the caller.
-- =====================================================================
CREATE VIEW v_entity_changes AS
SELECT
  t.name            AS tenant_name,
  e.uuid            AS entity_id,
  e.name            AS entity_name,
  e.type            AS entity_type,
  e.validation_status,
  e.created_at,
  e.updated_at,
  e.deleted_at,
  CASE
    WHEN e.deleted_at IS NOT NULL                              THEN 'DELETED'
    WHEN e.updated_at > e.created_at + INTERVAL '1 minute'    THEN 'MODIFIED'
    ELSE                                                            'CREATED'
  END AS change_type
FROM entities e
JOIN tenants t ON e.tenant_id = t.id;

-- =====================================================================
-- v_entity_paths: Flattened ancestor-descendant relationships
-- NOTE: No ORDER BY — sort in the calling query.
-- =====================================================================
CREATE VIEW v_entity_paths AS
SELECT
  t.name        AS tenant_name,
  a.uuid        AS ancestor_id,
  a.name        AS ancestor_name,
  a.type        AS ancestor_type,
  d.uuid        AS descendant_id,
  d.name        AS descendant_name,
  d.type        AS descendant_type,
  hp.depth
FROM hierarchy_paths hp
JOIN entities a ON hp.ancestor_id   = a.uuid
JOIN entities d ON hp.descendant_id = d.uuid
JOIN tenants  t ON hp.tenant_id     = t.id
WHERE a.deleted_at IS NULL
  AND d.deleted_at IS NULL;

-- =====================================================================
-- v_tenant_resource_utilization: Sequence state and entity counts per tenant
-- =====================================================================
CREATE VIEW v_tenant_resource_utilization AS
SELECT
  t.id                                                              AS tenant_id,
  t.name                                                            AS tenant_name,
  t."Status"                                                        AS tenant_status,
  COUNT(DISTINCT e.uuid)                                            AS total_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.is_active = TRUE)          AS active_entities,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.deleted_at IS NULL)        AS non_deleted_entities,
  COUNT(DISTINCT es.uuid)                                           AS sequence_states,
  COUNT(DISTINCT es.key)                                            AS document_types,
  MAX(e.created_at)                                                 AS last_entity_created,
  MAX(e.updated_at)                                                 AS last_entity_updated
FROM tenants t
LEFT JOIN entities    e  ON e.tenant_id  = t.id
LEFT JOIN entitystate es ON es.tenant_id = t.id
GROUP BY t.id, t.name, t."Status";
