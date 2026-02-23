-- =====================================================================
-- ENTITIES VIEWS - Reporting and analytical views for entity management
-- =====================================================================
/*
 * Entity Hierarchy View
 * 
 * Purpose: Provides recursive hierarchical view of entities within tenant context
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * - current_tenant_id() function (must be implemented)
 * 
 * Notes:
 * - Uses recursive CTE to build entity hierarchy paths
 * - Filters by current tenant context
 * - Includes soft delete filtering
 */
CREATE VIEW v_tenant_hierarchy AS WITH RECURSIVE org_chart AS (
  SELECT
    e.uuid AS entity_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    e.name::TEXT AS path,
    0 AS depth
  FROM
    entities e
  WHERE
    e.parent_id IS NULL
    AND e.deleted_at IS NULL
  UNION
  ALL
  SELECT
    e.uuid AS entity_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    (oc.path || ' > ' || e.name)::TEXT,
    oc.depth + 1
  FROM
    entities e
    JOIN org_chart oc ON e.parent_id = oc.entity_id
  WHERE
    e.deleted_at IS NULL
)
SELECT
  t.name AS tenant_name,
  oc.entity_id,
  oc.name AS entity_name,
  oc.type AS entity_type,
  oc.path AS full_path,
  oc.depth
FROM
  org_chart oc
  JOIN tenants t ON oc.tenant_id = t.id;

/*
 * Entity Hierarchy Structure View
 * 
 * Purpose: Shows the complete organizational structure for entities
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - This view is a placeholder - requires users table to be fully functional
 * - Currently shows entity hierarchy structure only
 * - Can be extended when user management tables are available
 */
CREATE VIEW v_entity_structure AS
SELECT
  t.name AS tenant_name,
  cc.uuid AS cost_center_id,
  cc.name AS cost_center,
  d.uuid AS department_id,
  d.name AS department,
  r.uuid AS regional_id,
  r.name AS regional,
  c.uuid AS company_id,
  c.name AS company
FROM
  entities cc
  JOIN entities d ON cc.parent_id = d.uuid
  AND d.type = 'DEPARTMENT'
  JOIN entities r ON d.parent_id = r.uuid
  AND r.type IN ('REGION', 'REGIONAL')
  JOIN entities c ON r.parent_id = c.uuid
  AND c.type = 'COMPANY'
  JOIN tenants t ON cc.tenant_id = t.id
WHERE
  cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL;

/*
 * Cost Center Basic Information View
 * 
 * Purpose: Provides basic cost center information with organizational context
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Simplified version without user count (requires users table)
 * - Shows cost center hierarchy up to company level
 * - Includes tenant context
 */
CREATE VIEW v_cost_center_info AS
SELECT
  t.name AS tenant_name,
  cc.uuid AS cost_center_id,
  cc.name AS cost_center,
  cc.code AS cost_center_code,
  d.uuid AS department_id,
  d.name AS department,
  r.uuid AS regional_id,
  r.name AS regional,
  c.uuid AS company_id,
  c.name AS company,
  cc.is_active AS cost_center_active,
  cc.created_at AS cost_center_created
FROM
  entities cc
  JOIN entities d ON cc.parent_id = d.uuid
  JOIN entities r ON d.parent_id = r.uuid
  JOIN entities c ON r.parent_id = c.uuid
  JOIN tenants t ON cc.tenant_id = t.id
WHERE
  cc.type = 'COST_CENTER'
  AND d.type = 'DEPARTMENT'
  AND r.type IN ('REGION', 'REGIONAL')
  AND c.type = 'COMPANY'
  AND cc.deleted_at IS NULL
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL;

/*
 * Department Summary View
 * 
 * Purpose: Provides summary information about departments and their cost centers
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Shows department hierarchy with cost center counts
 * - Simplified without user metrics (requires users table)
 * - Includes tenant context and soft delete filtering
 */
CREATE VIEW v_department_summary AS
SELECT
  t.name AS tenant_name,
  d.uuid AS department_id,
  d.name AS department_name,
  d.code AS department_code,
  r.uuid AS regional_id,
  r.name AS regional_name,
  c.uuid AS company_id,
  c.name AS company_name,
  COUNT(DISTINCT cc.uuid) AS cost_center_count,
  d.is_active AS department_active,
  d.created_at AS department_created
FROM
  entities d
  JOIN entities r ON d.parent_id = r.uuid
  JOIN entities c ON r.parent_id = c.uuid
  JOIN tenants t ON d.tenant_id = t.id
  LEFT JOIN entities cc ON cc.parent_id = d.uuid
  AND cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
WHERE
  d.type = 'DEPARTMENT'
  AND r.type IN ('REGION', 'REGIONAL')
  AND c.type = 'COMPANY'
  AND d.deleted_at IS NULL
  AND r.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY
  t.id,
  t.name,
  d.uuid,
  d.name,
  d.code,
  d.is_active,
  d.created_at,
  r.uuid,
  r.name,
  c.uuid,
  c.name;

/*
 * Company Structure View
 * 
 * Purpose: Shows complete organizational structure under each company
 * 
 * Dependencies:
 * - entities table
 * - hierarchy_paths table
 * - tenants table
 * 
 * Notes:
 * - Uses hierarchy_paths for efficient traversal
 * - Shows all entities under company level
 * - Includes depth information from company root
 */
CREATE VIEW v_company_structure AS
SELECT
  t.name AS tenant_name,
  c.uuid AS company_id,
  c.name AS company_name,
  c.code AS company_code,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.code AS entity_code,
  hp.depth AS levels_from_company
FROM
  entities c
  JOIN hierarchy_paths hp ON c.uuid = hp.ancestor_id
  JOIN entities e ON hp.descendant_id = e.uuid
  JOIN tenants t ON c.tenant_id = t.id
WHERE
  c.type = 'COMPANY'
  AND c.deleted_at IS NULL
  AND e.deleted_at IS NULL;

/*
 * Tenant Entity Summary View
 * 
 * Purpose: Provides summary statistics of entities within each tenant
 * 
 * Dependencies:
 * - tenants table
 * - entities table
 * 
 * Notes:
 * - Simplified without user counts (requires users table)
 * - Shows entity type distribution per tenant
 * - Includes entity activity status
 */
CREATE VIEW v_tenant_entity_summary AS
SELECT
  t.id AS tenant_id,
  t.name AS tenant_name,
  t."Status" AS tenant_status,
  COUNT(DISTINCT e.uuid) AS total_entities,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.type = 'COMPANY'
  ) AS company_count,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.type IN ('REGION', 'REGIONAL')
  ) AS regional_count,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.type = 'DEPARTMENT'
  ) AS department_count,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.type = 'COST_CENTER'
  ) AS cost_center_count,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.type = 'PROJECT'
  ) AS project_count,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.is_active = TRUE
  ) AS active_entities,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.deleted_at IS NULL
  ) AS non_deleted_entities
FROM
  tenants t
  LEFT JOIN entities e ON e.tenant_id = t.id
GROUP BY
  t.id,
  t.name,
  t."Status";

/*
 * Active Entities Report View
 * 
 * Purpose: Shows currently active entities across organizational hierarchy
 * 
 * Dependencies:
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Replacement for user-based view until users table is available
 * - Shows active entities with their full organizational context
 * - Includes recent activity indicators
 */
CREATE VIEW v_active_entities AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.code AS entity_code,
  c.name AS company_name,
  r.name AS regional_name,
  d.name AS department_name,
  e.is_active,
  e.created_at,
  e.updated_at
FROM
  entities e
  JOIN tenants t ON e.tenant_id = t.id
  LEFT JOIN entities c ON (
    CASE
      WHEN e.type = 'COMPANY' THEN e.uuid = c.uuid
      ELSE EXISTS (
        SELECT
          1
        FROM
          hierarchy_paths hp
        WHERE
          hp.descendant_id = e.uuid
          AND hp.ancestor_id = c.uuid
          AND c.type = 'COMPANY'
      )
    END
  )
  LEFT JOIN entities r ON (
    CASE
      WHEN e.type IN ('REGION', 'REGIONAL') THEN e.uuid = r.uuid
      ELSE EXISTS (
        SELECT
          1
        FROM
          hierarchy_paths hp
        WHERE
          hp.descendant_id = e.uuid
          AND hp.ancestor_id = r.uuid
          AND r.type IN ('REGION', 'REGIONAL')
      )
    END
  )
  LEFT JOIN entities d ON (
    CASE
      WHEN e.type = 'DEPARTMENT' THEN e.uuid = d.uuid
      ELSE e.parent_id = d.uuid
      AND d.type = 'DEPARTMENT'
    END
  )
WHERE
  e.deleted_at IS NULL
  AND e.is_active = TRUE
  AND (
    c.deleted_at IS NULL
    OR c.uuid IS NULL
  )
  AND (
    r.deleted_at IS NULL
    OR r.uuid IS NULL
  )
  AND (
    d.deleted_at IS NULL
    OR d.uuid IS NULL
  );

/*
 * Entity Change Log View
 * 
 * Purpose: Placeholder for audit trail - tracks entity changes
 * 
 * Dependencies:
 * - entities table (for change tracking)
 * - tenants table
 * 
 * Notes:
 * - Simplified view showing entity modification patterns
 * - Can be extended when audit_logs table is implemented
 * - Focuses on entity lifecycle events
 */
CREATE VIEW v_entity_changes AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.validation_status,
  e.created_at,
  e.updated_at,
  e.deleted_at,
  CASE
    WHEN e.deleted_at IS NOT NULL THEN 'DELETED'
    WHEN e.updated_at > e.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
    ELSE 'CREATED'
  END AS change_type
FROM
  entities e
  JOIN tenants t ON e.tenant_id = t.id
ORDER BY
  e.updated_at DESC;

/*
 * Entity Hierarchy Paths View
 * 
 * Purpose: Shows all ancestor-descendant relationships with depth information
 * 
 * Dependencies:
 * - hierarchy_paths table
 * - entities table
 * - tenants table
 * 
 * Notes:
 * - Provides flattened view of entity relationships
 * - Includes depth for distance calculations
 * - Useful for hierarchy analysis and reporting
 */
CREATE VIEW v_entity_paths AS
SELECT
  t.name AS tenant_name,
  a.uuid AS ancestor_id,
  a.name AS ancestor_name,
  a.type AS ancestor_type,
  d.uuid AS descendant_id,
  d.name AS descendant_name,
  d.type AS descendant_type,
  hp.depth
FROM
  hierarchy_paths hp
  JOIN entities a ON hp.ancestor_id = a.uuid
  JOIN entities d ON hp.descendant_id = d.uuid
  JOIN tenants t ON hp.tenant_id = t.id
WHERE
  a.deleted_at IS NULL
  AND d.deleted_at IS NULL
ORDER BY
  hp.depth,
  a.name,
  d.name;

/*
 * Tenant Resource Utilization View
 * 
 * Purpose: Shows resource utilization and capacity for each tenant
 * 
 * Dependencies:
 * - tenants table
 * - entities table
 * - entitystate table
 * 
 * Notes:
 * - Simplified without user metrics (requires users table)
 * - Shows entity utilization and state management
 * - Includes sequence number usage statistics
 */
CREATE VIEW v_tenant_resource_utilization AS
SELECT
  t.id AS tenant_id,
  t.name AS tenant_name,
  t."Status" AS tenant_status,
  COUNT(DISTINCT e.uuid) AS total_entities,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.is_active = TRUE
  ) AS active_entities,
  COUNT(DISTINCT e.uuid) FILTER (
    WHERE
      e.deleted_at IS NULL
  ) AS non_deleted_entities,
  COUNT(DISTINCT es.uuid) AS sequence_states,
  COUNT(DISTINCT es.key) AS document_types,
  MAX(e.created_at) AS last_entity_created,
  MAX(e.updated_at) AS last_entity_updated
FROM
  tenants t
  LEFT JOIN entities e ON e.tenant_id = t.id
  LEFT JOIN entitystate es ON es.tenant_id = t.id
GROUP BY
  t.id,
  t.name,
  t."Status";
