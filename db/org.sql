-- ------------------------------------------------------------------------------------------------
-- ENTITIES
-- ------------------------------------------------------------------------------------------------
-- Root entity/company table with hierarchical structure and per-entity accounting preferences.
-- type IN ('COMPANY','SUBSIDIARY','REGION','BRANCH','LOCATION','DEPARTMENT','DIVISION',
--           'COST_CENTER','PROJECT','BUDGET_UNIT').
-- validation_status IN ('PENDING','VALID','WARNING','ERROR').
--
-- NOTE: entity_path and entity_level are maintained by the application layer on create/reparent.
--       The no_self_parent CHECK and valid_fy_start_month CHECK are added after table creation
--       below. FK to tenants(id) requires tenants migration to run first.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE entities (
  uuid                UUID         PRIMARY KEY,
  tenant_id           UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  parent_id           UUID         REFERENCES entities(uuid) ON DELETE CASCADE, -- Self-reference for hierarchy
  name                VARCHAR(255) NOT NULL,
  code                VARCHAR(50),                                               -- Optional internal reference code
  TYPE                VARCHAR(20)  NOT NULL DEFAULT 'COMPANY' CHECK (
                                     TYPE IN (
                                       'COMPANY',
                                       'SUBSIDIARY',
                                       'REGION',
                                       'BRANCH',
                                       'LOCATION',
                                       'DEPARTMENT',
                                       'DIVISION',
                                       'COST_CENTER',
                                       'PROJECT',
                                       'BUDGET_UNIT'
                                     )
                                   ),
  is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
  hidden              BOOLEAN      NOT NULL DEFAULT false,
  accrual_method      BOOLEAN      NOT NULL,                                     -- TRUE = Accrual, FALSE = Cash
  fy_start_month      INTEGER      NOT NULL CHECK (fy_start_month BETWEEN 1 AND 12),
  address             JSONB        DEFAULT '{}'::jsonb,
  picture             VARCHAR(100),                                              -- File path or URL to entity logo
  -- Materialized path for O(1) subtree access checks.
  -- Format: '/root_uuid/parent_uuid/this_uuid/' (leading + trailing slash).
  entity_path         TEXT,
  entity_level        INTEGER      NOT NULL DEFAULT 1,                           -- 1 = root (COMPANY), increments per level
  settings            JSONB        DEFAULT '{}'::jsonb,
  metadata            JSONB        DEFAULT '{}'::jsonb,
  version             INTEGER      NOT NULL DEFAULT 1,
  last_validation_run TIMESTAMPTZ,
  validation_status   VARCHAR(20)  DEFAULT 'PENDING' CHECK (
                                     validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
                                   ),
  validation_errors   JSONB        DEFAULT '[]'::jsonb,
  created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ,
  UNIQUE (tenant_id, name)
);

COMMENT ON TABLE   entities                    IS 'Master table for business entities and organisational units. Supports hierarchical structures for companies, subsidiaries, departments, and other divisions. Each entity can maintain its own accounting books, customers, vendors, and fiscal year settings.';
COMMENT ON COLUMN  entities.uuid               IS 'Primary key — unique identifier for the entity.';
COMMENT ON COLUMN  entities.tenant_id          IS 'FK to tenants — associates entity with a specific tenant for multi-tenancy.';
COMMENT ON COLUMN  entities.parent_id          IS 'Self-referencing FK — creates hierarchical relationship between entities (e.g., subsidiary under parent company).';
COMMENT ON COLUMN  entities.name               IS 'Business name or title of the entity — must be unique within tenant.';
COMMENT ON COLUMN  entities.code               IS 'Optional internal reference code — used for abbreviated identification and reporting.';
COMMENT ON COLUMN  entities.type               IS 'Classification of entity type — defines the organisational level and purpose.';
COMMENT ON COLUMN  entities.is_active          IS 'Active status flag — indicates whether the entity is currently operational.';
COMMENT ON COLUMN  entities.hidden             IS 'Visibility flag — controls whether entity appears in UIs and reports.';
COMMENT ON COLUMN  entities.accrual_method     IS 'Accounting method: TRUE = accrual, FALSE = cash.';
COMMENT ON COLUMN  entities.fy_start_month     IS 'Fiscal year start month — numeric month (1–12) when fiscal year begins for this entity.';
COMMENT ON COLUMN  entities.address            IS 'Physical address stored as a JSON object with flexible address components.';
COMMENT ON COLUMN  entities.picture            IS 'Entity logo or image reference — file path or URL.';
COMMENT ON COLUMN  entities.entity_path        IS 'Materialized path: /uuid1/uuid2/this_uuid/. Enables subtree queries via LIKE ''/root/%''. Populated by app layer on create/reparent. Root entities: /uuid/.';
COMMENT ON COLUMN  entities.entity_level       IS 'Hierarchy depth: 1 = root COMPANY, increments per level (max 8). Used to determine EntityScope: level 1 = all, leaf = entity, else = subtree.';
COMMENT ON COLUMN  entities.settings           IS 'Entity-specific configuration — JSON object storing customisable settings and preferences.';
COMMENT ON COLUMN  entities.created_at         IS 'Record creation timestamp — automatically set when entity is first created.';
COMMENT ON COLUMN  entities.updated_at         IS 'Last modification timestamp — automatically updated on every row change.';
COMMENT ON COLUMN  entities.deleted_at         IS 'Soft deletion timestamp — NULL for active records, set when logically deleted.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE UNIQUE INDEX tenant_code_unique_idx    ON entities(tenant_id, code)        WHERE code IS NOT NULL;             -- Entity codes unique within tenant (sparse)
CREATE INDEX idx_entities_tenant_type         ON entities(tenant_id, TYPE);                                           -- Common filter by tenant + type
CREATE INDEX idx_entities_parent_id           ON entities(parent_id);                                                 -- Hierarchy traversal — direct children
CREATE INDEX idx_entities_active              ON entities(is_active)              WHERE is_active = TRUE;             -- Partial index for active-only queries
CREATE INDEX idx_entities_settings_gin        ON entities USING gin(settings);                                        -- JSON containment queries on settings
CREATE INDEX idx_entities_path                ON entities(tenant_id, entity_path) WHERE entity_path IS NOT NULL AND deleted_at IS NULL; -- Subtree LIKE queries
CREATE INDEX idx_entities_level               ON entities(tenant_id, entity_level) WHERE deleted_at IS NULL;          -- Org-level filtering
CREATE INDEX idx_entities_address_gin         ON entities USING gin(address);                                         -- JSON containment queries on address
CREATE INDEX idx_entities_deleted_at          ON entities(deleted_at)             WHERE deleted_at IS NOT NULL;       -- Soft-deleted record lookups
CREATE INDEX idx_entities_tenant              ON entities(tenant_id);                                                 -- Tenant-level scans
CREATE INDEX idx_entities_parent              ON entities(parent_id);                                                 -- Parent-based lookups
CREATE INDEX idx_entities_type                ON entities(TYPE);                                                      -- Type-based filtering

COMMENT ON INDEX tenant_code_unique_idx    IS 'Ensures entity codes are unique within each tenant — only applies when code is not NULL.';
COMMENT ON INDEX idx_entities_tenant_type  IS 'Optimises queries filtering entities by tenant and type — common pattern for entity listings.';
COMMENT ON INDEX idx_entities_parent_id    IS 'Speeds up hierarchy traversal queries when finding direct children of an entity.';
COMMENT ON INDEX idx_entities_active       IS 'Optimises queries for active entities only — uses partial index to save space.';

-- ------------------------------------------------------------------------------------------------
-- DATA INTEGRITY CONSTRAINTS
-- ------------------------------------------------------------------------------------------------
ALTER TABLE entities ADD CONSTRAINT no_self_parent      CHECK (uuid != parent_id);
ALTER TABLE entities ADD CONSTRAINT valid_fy_start_month CHECK (fy_start_month BETWEEN 1 AND 12);

COMMENT ON CONSTRAINT no_self_parent       ON entities IS 'Prevents circular references where an entity is its own parent.';
COMMENT ON CONSTRAINT valid_fy_start_month ON entities IS 'Validates fiscal year start month is between 1 (January) and 12 (December).';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON entities
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON entities
  FOR ALL TO admin_role
  USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY entities_ro_select ON entities
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON entities TO application_role;
-- ------------------------------------------------------------------------------------------------
-- HIERARCHY_PATHS
-- ------------------------------------------------------------------------------------------------
-- Closure table for efficient entity hierarchy queries. Stores every ancestor-descendant pair
-- (ancestor_id, descendant_id, depth) — including self-references at depth 0.
-- This is the complete information set; no redundant entity_id column is needed.
--
-- NOTE: Depends on entities(uuid) and tenants(id) from migration 000201.
--       The no_self_parent CHECK in 000201 prevents immediate self-reference; this table
--       supplements that with full transitive closure.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE hierarchy_paths (
  tenant_id     UUID        NOT NULL REFERENCES tenants(id)  ON DELETE CASCADE,
  ancestor_id   UUID        NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  descendant_id UUID        NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  depth         INT         NOT NULL CHECK (depth >= 0),       -- 0 = self, 1 = direct parent-child, 2+ = deeper
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

COMMENT ON TABLE   hierarchy_paths              IS 'Closure table for efficient entity hierarchy queries. Stores all ancestor-descendant relationships with depth information. Enables fast retrieval of entity trees, subtrees, and hierarchy levels without recursive queries.';
COMMENT ON COLUMN  hierarchy_paths.tenant_id    IS 'Tenant identifier — partitions hierarchy data by tenant for multi-tenancy.';
COMMENT ON COLUMN  hierarchy_paths.ancestor_id  IS 'Ancestor entity in the relationship — references entities.uuid.';
COMMENT ON COLUMN  hierarchy_paths.descendant_id IS 'Descendant entity in the relationship — references entities.uuid.';
COMMENT ON COLUMN  hierarchy_paths.depth        IS 'Hierarchical distance: 0 = self-reference, 1 = direct parent-child, 2+ = deeper.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_hierarchy_paths_descendant  ON hierarchy_paths(descendant_id);         -- Reverse traversal — find all ancestors of an entity
CREATE INDEX idx_hierarchy_paths_depth       ON hierarchy_paths(tenant_id, depth);       -- Queries filtered by hierarchy depth
CREATE INDEX idx_hierarchy_paths_tenant      ON hierarchy_paths(tenant_id);              -- Tenant-level scans
CREATE INDEX idx_hierarchy_paths_ancestor    ON hierarchy_paths(ancestor_id);            -- Forward traversal — find all descendants
CREATE INDEX idx_hierarchy_paths_desc_depth  ON hierarchy_paths(descendant_id, depth);  -- Find typed ancestor (e.g. nearest COMPANY above entity X)

COMMENT ON INDEX idx_hierarchy_paths_descendant IS 'Enables efficient reverse hierarchy traversal — finds all ancestors of a given entity.';
COMMENT ON INDEX idx_hierarchy_paths_depth      IS 'Optimises queries filtering by hierarchy depth — useful for organisation-level reports.';

-- ------------------------------------------------------------------------------------------------
-- CHECK_ENTITY_HIERARCHY_DEPTH
-- ------------------------------------------------------------------------------------------------
-- Trigger function that prevents A→B→A cycles and enforces a maximum hierarchy depth of 8 levels.
-- The no_self_parent CHECK in 000201 handles the immediate self-reference case; this trigger
-- catches multi-hop cycles and depth violations on every INSERT or UPDATE of parent_id.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_entity_hierarchy_depth()
RETURNS TRIGGER AS $$
DECLARE
  v_depth   INTEGER := 0;
  v_current UUID    := NEW.parent_id;
  v_seen    UUID[]  := ARRAY[NEW.uuid];
BEGIN
  WHILE v_current IS NOT NULL LOOP
    -- Cycle detection
    IF v_current = ANY(v_seen) THEN
      RAISE EXCEPTION
        'Circular reference in entity hierarchy: entity % creates a cycle at ancestor %',
        NEW.uuid, v_current
        USING ERRCODE = '23000';
    END IF;

    v_depth := v_depth + 1;
    v_seen  := v_seen || v_current;

    IF v_depth > 8 THEN
      RAISE EXCEPTION
        'Entity hierarchy exceeds the maximum depth of 8 levels (entity %)',
        NEW.uuid
        USING ERRCODE = '23000';
    END IF;

    SELECT parent_id INTO v_current
      FROM entities
     WHERE uuid = v_current;
  END LOOP;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_entity_hierarchy_depth IS
  'Prevents circular references and enforces max hierarchy depth of 8 levels on the entities table. '
  'Walks the parent chain on every INSERT/UPDATE that sets parent_id.';

CREATE TRIGGER entities_check_hierarchy_depth
  BEFORE INSERT OR UPDATE OF parent_id ON entities
  FOR EACH ROW
  WHEN (NEW.parent_id IS NOT NULL)
  EXECUTE FUNCTION check_entity_hierarchy_depth();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE hierarchy_paths ENABLE ROW LEVEL SECURITY;
ALTER TABLE hierarchy_paths FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON hierarchy_paths
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON hierarchy_paths
  FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY hierarchy_paths_ro_select ON hierarchy_paths
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON hierarchy_paths TO application_role;
-- ------------------------------------------------------------------------------------------------
-- ENTITYSTATE TABLE
-- ------------------------------------------------------------------------------------------------
-- Manages sequential numbering for business documents within entities. Tracks the next available
-- sequence number for each document type (invoice, po, estimate, bill, receipt, etc.) by fiscal
-- year and entity.
-- config JSONB keys: prefix, suffix, pad_length (INT), reset_frequency (yearly|monthly|never),
--   format_template (STRING). Values override tenant_configurations.settings for this entity+doctype.
--   Example: {"prefix":"NORTH-INV-","pad_length":6,"reset_frequency":"yearly"}
--
-- NOTE: entity_id and entity_unit_id reference entities(uuid) with DEFERRABLE INITIALLY DEFERRED
--       to allow insertion within the same transaction that creates the entity record.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE entitystate (
  uuid            UUID         PRIMARY KEY,
  tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  fiscal_year     SMALLINT,                                             -- fiscal year for sequence scoping; NULL = no year partitioning
  KEY             VARCHAR(10)  NOT NULL,                                -- document type key (e.g. invoice, po, estimate)
  sequence        BIGINT       NOT NULL,                                -- next available sequence number for this doctype
  entity_id       UUID         NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  entity_unit_id  UUID         REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED, -- optional sub-entity / department
  config          JSONB        NOT NULL DEFAULT '{}'::jsonb,           -- per-entity formatting overrides (see header note)
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ
);

COMMENT ON TABLE  entitystate                IS 'Manages sequential numbering for business documents within entities. Tracks next available sequence numbers for different document types (invoices, purchase orders, estimates, etc.) by fiscal year and entity.';
COMMENT ON COLUMN entitystate.uuid           IS 'Primary key — unique identifier for the entity state record.';
COMMENT ON COLUMN entitystate.tenant_id      IS 'Foreign key to tenants table for multi-tenant isolation.';
COMMENT ON COLUMN entitystate.fiscal_year    IS 'Fiscal year for sequence tracking — allows separate numbering sequences per year.';
COMMENT ON COLUMN entitystate.key            IS 'Document type identifier — specifies the type of document being numbered (invoice, po, estimate, bill, receipt, etc.).';
COMMENT ON COLUMN entitystate.sequence       IS 'Next sequence number — the next available sequential number for this document type.';
COMMENT ON COLUMN entitystate.entity_id      IS 'Primary entity reference — the main entity that owns this sequence numbering.';
COMMENT ON COLUMN entitystate.entity_unit_id IS 'Sub-entity reference — optional reference to a subsidiary or department within the main entity for more granular numbering.';
COMMENT ON COLUMN entitystate.config         IS
  'Document sequence formatting config for this entity+doctype combination. '
  'Keys: prefix, suffix, pad_length (INT), reset_frequency (yearly|monthly|never), format_template (STRING). '
  'Overrides tenant_configurations.settings for sequences on this entity. '
  'Example: {"prefix":"NORTH-INV-","pad_length":6,"reset_frequency":"yearly"}';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_entitystate_entity_key  ON entitystate(entity_id, KEY);                     -- sequence lookups by entity + document type
CREATE INDEX idx_entitystate_fiscal_year ON entitystate(entity_id, fiscal_year, KEY);        -- filtered by fiscal year

COMMENT ON INDEX idx_entitystate_entity_key  IS 'Optimizes sequence number lookups by entity and document type';
COMMENT ON INDEX idx_entitystate_fiscal_year IS 'Supports efficient sequence retrieval filtered by fiscal year';

-- ------------------------------------------------------------------------------------------------
-- DATA INTEGRITY CONSTRAINTS
-- ------------------------------------------------------------------------------------------------
ALTER TABLE entitystate
  ADD CONSTRAINT unique_tenant_entity_key_fy UNIQUE (tenant_id, entity_id, KEY, fiscal_year);

COMMENT ON CONSTRAINT unique_tenant_entity_key_fy ON entitystate IS 'Prevents duplicate sequence trackers for same tenant, entity, document type, and fiscal year';

ALTER TABLE entitystate
  ADD CONSTRAINT positive_sequence CHECK (sequence > 0);


-- GIN index for sequence config lookups by prefix or format
CREATE INDEX IF NOT EXISTS idx_entitystate_config_gin
  ON entitystate USING gin(config)
  WHERE config IS NOT NULL AND config <> '{}'::jsonb;

COMMENT ON CONSTRAINT positive_sequence ON entitystate IS 'Ensures sequence numbers are always positive values';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE entitystate ENABLE ROW LEVEL SECURITY;
ALTER TABLE entitystate FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON entitystate FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY admin_full_access_policy ON entitystate FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY entitystate_ro_select ON entitystate
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
-- ------------------------------------------------------------------------------------------------
-- V_TENANT_HIERARCHY VIEW
-- ------------------------------------------------------------------------------------------------
-- Recursive path view that builds a full org-chart path (e.g. "Corp > Region > Dept") for every
-- entity in the tree. Useful for breadcrumb rendering and path-aware reports.
--
-- NOTE: No ORDER BY — add ORDER BY in the calling query to avoid unnecessary materialisation.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_tenant_hierarchy SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ENTITY_STRUCTURE VIEW
-- ------------------------------------------------------------------------------------------------
-- Canonical 4-level hierarchy flattened into a single row: COST_CENTER → DEPARTMENT → REGION →
-- COMPANY. Intended for reports that assume the standard 4-level org model.
--
-- NOTE: Tenants with non-standard depth should query hierarchy_paths with the depth column instead.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_entity_structure SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_COST_CENTER_INFO VIEW
-- ------------------------------------------------------------------------------------------------
-- Cost centre record enriched with its full 4-level ancestor context (department, regional,
-- company, tenant). Convenient for detail pages and exports.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_cost_center_info SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_DEPARTMENT_SUMMARY VIEW
-- ------------------------------------------------------------------------------------------------
-- Department rows with aggregated cost-centre counts and their regional/company ancestors.
-- Useful for dashboard tiles and org-overview reports.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_department_summary SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_COMPANY_STRUCTURE VIEW
-- ------------------------------------------------------------------------------------------------
-- All entities under each company resolved via the hierarchy_paths closure table. Returns every
-- descendant with its depth (levels from the company root).
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_company_structure SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_TENANT_ENTITY_SUMMARY VIEW
-- ------------------------------------------------------------------------------------------------
-- Entity counts per tenant broken down by type (COMPANY, REGIONAL, DEPARTMENT, COST_CENTER,
-- PROJECT) plus active and non-deleted totals. Used for tenant dashboard metrics.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_tenant_entity_summary SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ACTIVE_ENTITIES VIEW
-- ------------------------------------------------------------------------------------------------
-- Active (non-deleted, is_active=TRUE) entities enriched with their nearest typed ancestor names
-- (company, regional, department) resolved via the hierarchy_paths closure table.
--
-- Uses CTEs with DISTINCT ON to find the nearest ancestor of each type — avoids the
-- CASE WHEN EXISTS correlated-subquery anti-pattern that forces a per-row subquery evaluation.
--
-- NOTE: No ORDER BY — add ORDER BY in the calling query.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_active_entities SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ENTITY_CHANGES VIEW
-- ------------------------------------------------------------------------------------------------
-- Lifecycle change log classifying each entity row as CREATED, MODIFIED, or DELETED based on
-- its timestamp fields.
--
-- NOTE: No ORDER BY — callers add ORDER BY as needed. ORDER BY inside a view definition is not
--       guaranteed to propagate and forces a sort materialisation even when unused by the caller.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_entity_changes SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ENTITY_PATHS VIEW
-- ------------------------------------------------------------------------------------------------
-- Flattened ancestor-descendant relationships from the hierarchy_paths closure table enriched
-- with entity names and types. Handy for path-based permission checks and tree traversals.
--
-- NOTE: No ORDER BY — sort in the calling query.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_entity_paths SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_TENANT_RESOURCE_UTILIZATION VIEW
-- ------------------------------------------------------------------------------------------------
-- Sequence state and entity counts per tenant. Combines entity stats with entitystate stats to
-- give an overview of how many document types and sequence slots each tenant is using.
-- ------------------------------------------------------------------------------------------------
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

ALTER VIEW v_tenant_resource_utilization SET (security_invoker = true);
