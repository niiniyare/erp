-- ------------------------------------------------------------------------------------------------
-- ORGUNITS
-- ------------------------------------------------------------------------------------------------
-- Root orgunit/company table with hierarchical structure and per-orgunit accounting preferences.
-- type IN ('COMPANY','SUBSIDIARY','REGION','BRANCH','LOCATION','DEPARTMENT','DIVISION',
--           'COST_CENTER','PROJECT','BUDGET_UNIT').
-- validation_status IN ('PENDING','VALID','WARNING','ERROR').
--
-- NOTE: org_path and org_level are maintained by the application layer on create/reparent.
--       The no_self_parent CHECK and valid_fy_start_month CHECK are added after table creation
--       below. FK to tenants(id) requires tenants migration to run first.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE orgunits (
  uuid                UUID         PRIMARY KEY,
  tenant_id           UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  parent_id           UUID         REFERENCES orgunits(uuid) ON DELETE CASCADE, -- Self-reference for hierarchy
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
  picture             VARCHAR(100),                                              -- File path or URL to orgunit logo
  -- Materialized path for O(1) subtree access checks.
  -- Format: '/root_uuid/parent_uuid/this_uuid/' (leading + trailing slash).
  org_path            TEXT,
  org_level           INTEGER      NOT NULL DEFAULT 1,                           -- 1 = root (COMPANY), increments per level
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

COMMENT ON TABLE   orgunits                    IS 'Master table for business orgunits and organisational units. Supports hierarchical structures for companies, subsidiaries, departments, and other divisions. Each orgunit can maintain its own accounting books, customers, vendors, and fiscal year settings.';
COMMENT ON COLUMN  orgunits.uuid               IS 'Primary key — unique identifier for the orgunit.';
COMMENT ON COLUMN  orgunits.tenant_id          IS 'FK to tenants — associates orgunit with a specific tenant for multi-tenancy.';
COMMENT ON COLUMN  orgunits.parent_id          IS 'Self-referencing FK — creates hierarchical relationship between orgunits (e.g., subsidiary under parent company).';
COMMENT ON COLUMN  orgunits.name               IS 'Business name or title of the orgunit — must be unique within tenant.';
COMMENT ON COLUMN  orgunits.code               IS 'Optional internal reference code — used for abbreviated identification and reporting.';
COMMENT ON COLUMN  orgunits.type               IS 'Classification of orgunit type — defines the organisational level and purpose.';
COMMENT ON COLUMN  orgunits.is_active          IS 'Active status flag — indicates whether the orgunit is currently operational.';
COMMENT ON COLUMN  orgunits.hidden             IS 'Visibility flag — controls whether orgunit appears in UIs and reports.';
COMMENT ON COLUMN  orgunits.accrual_method     IS 'Accounting method: TRUE = accrual, FALSE = cash.';
COMMENT ON COLUMN  orgunits.fy_start_month     IS 'Fiscal year start month — numeric month (1–12) when fiscal year begins for this orgunit.';
COMMENT ON COLUMN  orgunits.address            IS 'Physical address stored as a JSON object with flexible address components.';
COMMENT ON COLUMN  orgunits.picture            IS 'Orgunit logo or image reference — file path or URL.';
COMMENT ON COLUMN  orgunits.org_path           IS 'Materialized path: /uuid1/uuid2/this_uuid/. Enables subtree queries via LIKE ''/root/%''. Populated by app layer on create/reparent. Root orgunits: /uuid/.';
COMMENT ON COLUMN  orgunits.org_level          IS 'Hierarchy depth: 1 = root COMPANY, increments per level (max 8). Used to determine OrgScope: level 1 = all, leaf = orgunit, else = subtree.';
COMMENT ON COLUMN  orgunits.settings           IS 'Orgunit-specific configuration — JSON object storing customisable settings and preferences.';
COMMENT ON COLUMN  orgunits.created_at         IS 'Record creation timestamp — automatically set when orgunit is first created.';
COMMENT ON COLUMN  orgunits.updated_at         IS 'Last modification timestamp — automatically updated on every row change.';
COMMENT ON COLUMN  orgunits.deleted_at         IS 'Soft deletion timestamp — NULL for active records, set when logically deleted.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE UNIQUE INDEX orgunit_code_unique_idx    ON orgunits(tenant_id, code)        WHERE code IS NOT NULL;             -- Orgunit codes unique within tenant (sparse)
CREATE INDEX idx_orgunits_tenant_type          ON orgunits(tenant_id, TYPE);                                           -- Common filter by tenant + type
CREATE INDEX idx_orgunits_parent_id            ON orgunits(parent_id);                                                 -- Hierarchy traversal — direct children
CREATE INDEX idx_orgunits_active               ON orgunits(is_active)              WHERE is_active = TRUE;             -- Partial index for active-only queries
CREATE INDEX idx_orgunits_settings_gin         ON orgunits USING gin(settings);                                        -- JSON containment queries on settings
CREATE INDEX idx_orgunits_path                 ON orgunits(tenant_id, org_path)    WHERE org_path IS NOT NULL AND deleted_at IS NULL; -- Subtree LIKE queries
CREATE INDEX idx_orgunits_level                ON orgunits(tenant_id, org_level)   WHERE deleted_at IS NULL;           -- Org-level filtering
CREATE INDEX idx_orgunits_address_gin          ON orgunits USING gin(address);                                         -- JSON containment queries on address
CREATE INDEX idx_orgunits_deleted_at           ON orgunits(deleted_at)             WHERE deleted_at IS NOT NULL;       -- Soft-deleted record lookups
CREATE INDEX idx_orgunits_tenant               ON orgunits(tenant_id);                                                 -- Tenant-level scans
CREATE INDEX idx_orgunits_parent               ON orgunits(parent_id);                                                 -- Parent-based lookups
CREATE INDEX idx_orgunits_type                 ON orgunits(TYPE);                                                      -- Type-based filtering

COMMENT ON INDEX orgunit_code_unique_idx    IS 'Ensures orgunit codes are unique within each tenant — only applies when code is not NULL.';
COMMENT ON INDEX idx_orgunits_tenant_type   IS 'Optimises queries filtering orgunits by tenant and type — common pattern for orgunit listings.';
COMMENT ON INDEX idx_orgunits_parent_id     IS 'Speeds up hierarchy traversal queries when finding direct children of an orgunit.';
COMMENT ON INDEX idx_orgunits_active        IS 'Optimises queries for active orgunits only — uses partial index to save space.';

-- ------------------------------------------------------------------------------------------------
-- DATA INTEGRITY CONSTRAINTS
-- ------------------------------------------------------------------------------------------------
ALTER TABLE orgunits ADD CONSTRAINT no_self_parent      CHECK (uuid != parent_id);
ALTER TABLE orgunits ADD CONSTRAINT valid_fy_start_month CHECK (fy_start_month BETWEEN 1 AND 12);

COMMENT ON CONSTRAINT no_self_parent       ON orgunits IS 'Prevents circular references where an orgunit is its own parent.';
COMMENT ON CONSTRAINT valid_fy_start_month ON orgunits IS 'Validates fiscal year start month is between 1 (January) and 12 (December).';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE orgunits ENABLE ROW LEVEL SECURITY;
ALTER TABLE orgunits FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON orgunits
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON orgunits
  FOR ALL TO admin_role
  USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY orgunits_ro_select ON orgunits
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON orgunits TO application_role;
-- ------------------------------------------------------------------------------------------------
-- HIERARCHY_PATHS
-- ------------------------------------------------------------------------------------------------
-- Closure table for efficient orgunit hierarchy queries. Stores every ancestor-descendant pair
-- (ancestor_id, descendant_id, depth) — including self-references at depth 0.
-- This is the complete information set; no redundant orgunit_id column is needed.
--
-- NOTE: Depends on orgunits(uuid) and tenants(id) from migration 000201.
--       The no_self_parent CHECK in 000201 prevents immediate self-reference; this table
--       supplements that with full transitive closure.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE hierarchy_paths (
  tenant_id     UUID        NOT NULL REFERENCES tenants(id)   ON DELETE CASCADE,
  ancestor_id   UUID        NOT NULL REFERENCES orgunits(uuid) ON DELETE CASCADE,
  descendant_id UUID        NOT NULL REFERENCES orgunits(uuid) ON DELETE CASCADE,
  depth         INT         NOT NULL CHECK (depth >= 0),       -- 0 = self, 1 = direct parent-child, 2+ = deeper
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

COMMENT ON TABLE   hierarchy_paths              IS 'Closure table for efficient orgunit hierarchy queries. Stores all ancestor-descendant relationships with depth information. Enables fast retrieval of orgunit trees, subtrees, and hierarchy levels without recursive queries.';
COMMENT ON COLUMN  hierarchy_paths.tenant_id    IS 'Tenant identifier — partitions hierarchy data by tenant for multi-tenancy.';
COMMENT ON COLUMN  hierarchy_paths.ancestor_id  IS 'Ancestor orgunit in the relationship — references orgunits.uuid.';
COMMENT ON COLUMN  hierarchy_paths.descendant_id IS 'Descendant orgunit in the relationship — references orgunits.uuid.';
COMMENT ON COLUMN  hierarchy_paths.depth        IS 'Hierarchical distance: 0 = self-reference, 1 = direct parent-child, 2+ = deeper.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_hierarchy_paths_descendant  ON hierarchy_paths(descendant_id);         -- Reverse traversal — find all ancestors of an orgunit
CREATE INDEX idx_hierarchy_paths_depth       ON hierarchy_paths(tenant_id, depth);       -- Queries filtered by hierarchy depth
CREATE INDEX idx_hierarchy_paths_tenant      ON hierarchy_paths(tenant_id);              -- Tenant-level scans
CREATE INDEX idx_hierarchy_paths_ancestor    ON hierarchy_paths(ancestor_id);            -- Forward traversal — find all descendants
CREATE INDEX idx_hierarchy_paths_desc_depth  ON hierarchy_paths(descendant_id, depth);  -- Find typed ancestor (e.g. nearest COMPANY above orgunit X)

COMMENT ON INDEX idx_hierarchy_paths_descendant IS 'Enables efficient reverse hierarchy traversal — finds all ancestors of a given orgunit.';
COMMENT ON INDEX idx_hierarchy_paths_depth      IS 'Optimises queries filtering by hierarchy depth — useful for organisation-level reports.';

-- ------------------------------------------------------------------------------------------------
-- CHECK_ORGUNIT_HIERARCHY_DEPTH
-- ------------------------------------------------------------------------------------------------
-- Trigger function that prevents A→B→A cycles and enforces a maximum hierarchy depth of 8 levels.
-- The no_self_parent CHECK in 000201 handles the immediate self-reference case; this trigger
-- catches multi-hop cycles and depth violations on every INSERT or UPDATE of parent_id.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_orgunit_hierarchy_depth()
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
        'Circular reference in orgunit hierarchy: orgunit % creates a cycle at ancestor %',
        NEW.uuid, v_current
        USING ERRCODE = '23000';
    END IF;

    v_depth := v_depth + 1;
    v_seen  := v_seen || v_current;

    IF v_depth > 8 THEN
      RAISE EXCEPTION
        'Orgunit hierarchy exceeds the maximum depth of 8 levels (orgunit %)',
        NEW.uuid
        USING ERRCODE = '23000';
    END IF;

    SELECT parent_id INTO v_current
      FROM orgunits
     WHERE uuid = v_current;
  END LOOP;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_orgunit_hierarchy_depth IS
  'Prevents circular references and enforces max hierarchy depth of 8 levels on the orgunits table. '
  'Walks the parent chain on every INSERT/UPDATE that sets parent_id.';

CREATE TRIGGER orgunits_check_hierarchy_depth
  BEFORE INSERT OR UPDATE OF parent_id ON orgunits
  FOR EACH ROW
  WHEN (NEW.parent_id IS NOT NULL)
  EXECUTE FUNCTION check_orgunit_hierarchy_depth();

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
-- ORGSCOPE TABLE
-- ------------------------------------------------------------------------------------------------
-- Manages sequential numbering for business documents within orgunits. Tracks the next available
-- sequence number for each document type (invoice, po, estimate, bill, receipt, etc.) by fiscal
-- year and orgunit.
-- config JSONB keys: prefix, suffix, pad_length (INT), reset_frequency (yearly|monthly|never),
--   format_template (STRING). Values override tenant_configurations.settings for this orgunit+doctype.
--   Example: {"prefix":"NORTH-INV-","pad_length":6,"reset_frequency":"yearly"}
--
-- NOTE: orgunit_id and orgunit_sub_id reference orgunits(uuid) with DEFERRABLE INITIALLY DEFERRED
--       to allow insertion within the same transaction that creates the orgunit record.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE orgscope (
  uuid            UUID         PRIMARY KEY,
  tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  fiscal_year     SMALLINT,                                             -- fiscal year for sequence scoping; NULL = no year partitioning
  KEY             VARCHAR(10)  NOT NULL,                                -- document type key (e.g. invoice, po, estimate)
  sequence        BIGINT       NOT NULL,                                -- next available sequence number for this doctype
  orgunit_id      UUID         NOT NULL REFERENCES orgunits(uuid) DEFERRABLE INITIALLY DEFERRED,
  orgunit_sub_id  UUID         REFERENCES orgunits(uuid) DEFERRABLE INITIALLY DEFERRED, -- optional sub-orgunit / department
  config          JSONB        NOT NULL DEFAULT '{}'::jsonb,           -- per-orgunit formatting overrides (see header note)
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ
);

COMMENT ON TABLE  orgscope                IS 'Manages sequential numbering for business documents within orgunits. Tracks next available sequence numbers for different document types (invoices, purchase orders, estimates, etc.) by fiscal year and orgunit.';
COMMENT ON COLUMN orgscope.uuid           IS 'Primary key — unique identifier for the orgscope record.';
COMMENT ON COLUMN orgscope.tenant_id      IS 'Foreign key to tenants table for multi-tenant isolation.';
COMMENT ON COLUMN orgscope.fiscal_year    IS 'Fiscal year for sequence tracking — allows separate numbering sequences per year.';
COMMENT ON COLUMN orgscope.key            IS 'Document type identifier — specifies the type of document being numbered (invoice, po, estimate, bill, receipt, etc.).';
COMMENT ON COLUMN orgscope.sequence       IS 'Next sequence number — the next available sequential number for this document type.';
COMMENT ON COLUMN orgscope.orgunit_id     IS 'Primary orgunit reference — the main orgunit that owns this sequence numbering.';
COMMENT ON COLUMN orgscope.orgunit_sub_id IS 'Sub-orgunit reference — optional reference to a subsidiary or department within the main orgunit for more granular numbering.';
COMMENT ON COLUMN orgscope.config         IS
  'Document sequence formatting config for this orgunit+doctype combination. '
  'Keys: prefix, suffix, pad_length (INT), reset_frequency (yearly|monthly|never), format_template (STRING). '
  'Overrides tenant_configurations.settings for sequences on this orgunit. '
  'Example: {"prefix":"NORTH-INV-","pad_length":6,"reset_frequency":"yearly"}';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_orgscope_orgunit_key  ON orgscope(orgunit_id, KEY);                     -- sequence lookups by orgunit + document type
CREATE INDEX idx_orgscope_fiscal_year  ON orgscope(orgunit_id, fiscal_year, KEY);        -- filtered by fiscal year

COMMENT ON INDEX idx_orgscope_orgunit_key  IS 'Optimizes sequence number lookups by orgunit and document type';
COMMENT ON INDEX idx_orgscope_fiscal_year  IS 'Supports efficient sequence retrieval filtered by fiscal year';

-- ------------------------------------------------------------------------------------------------
-- DATA INTEGRITY CONSTRAINTS
-- ------------------------------------------------------------------------------------------------
ALTER TABLE orgscope
  ADD CONSTRAINT unique_tenant_orgunit_key_fy UNIQUE (tenant_id, orgunit_id, KEY, fiscal_year);

COMMENT ON CONSTRAINT unique_tenant_orgunit_key_fy ON orgscope IS 'Prevents duplicate sequence trackers for same tenant, orgunit, document type, and fiscal year';

ALTER TABLE orgscope
  ADD CONSTRAINT positive_sequence CHECK (sequence > 0);


-- GIN index for sequence config lookups by prefix or format
CREATE INDEX IF NOT EXISTS idx_orgscope_config_gin
  ON orgscope USING gin(config)
  WHERE config IS NOT NULL AND config <> '{}'::jsonb;

COMMENT ON CONSTRAINT positive_sequence ON orgscope IS 'Ensures sequence numbers are always positive values';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE orgscope ENABLE ROW LEVEL SECURITY;
ALTER TABLE orgscope FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON orgscope FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY admin_full_access_policy ON orgscope FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY orgscope_ro_select ON orgscope
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
-- ------------------------------------------------------------------------------------------------
-- V_TENANT_HIERARCHY VIEW
-- ------------------------------------------------------------------------------------------------
-- Recursive path view that builds a full org-chart path (e.g. "Corp > Region > Dept") for every
-- orgunit in the tree. Useful for breadcrumb rendering and path-aware reports.
--
-- NOTE: No ORDER BY — add ORDER BY in the calling query to avoid unnecessary materialisation.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_tenant_hierarchy AS
WITH RECURSIVE org_chart AS (
  SELECT
    e.uuid       AS orgunit_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    e.name::TEXT AS path,
    0            AS depth
  FROM orgunits e
  WHERE e.parent_id IS NULL
    AND e.deleted_at IS NULL

  UNION ALL

  SELECT
    e.uuid AS orgunit_id,
    e.name,
    e.type,
    e.parent_id,
    e.tenant_id,
    (oc.path || ' > ' || e.name)::TEXT,
    oc.depth + 1
  FROM orgunits e
  JOIN org_chart oc ON e.parent_id = oc.orgunit_id
  WHERE e.deleted_at IS NULL
)
SELECT
  t.name    AS tenant_name,
  oc.orgunit_id,
  oc.name   AS orgunit_name,
  oc.type   AS orgunit_type,
  oc.path   AS full_path,
  oc.depth
FROM org_chart oc
JOIN tenants t ON oc.tenant_id = t.id;

ALTER VIEW v_tenant_hierarchy SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ORG_STRUCTURE VIEW
-- ------------------------------------------------------------------------------------------------
-- Canonical 4-level hierarchy flattened into a single row: COST_CENTER → DEPARTMENT → REGION →
-- COMPANY. Intended for reports that assume the standard 4-level org model.
--
-- NOTE: Tenants with non-standard depth should query hierarchy_paths with the depth column instead.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_structure AS
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
FROM orgunits cc
JOIN orgunits d  ON cc.parent_id = d.uuid  AND d.type  = 'DEPARTMENT'
JOIN orgunits r  ON d.parent_id  = r.uuid  AND r.type  IN ('REGION', 'REGIONAL')
JOIN orgunits c  ON r.parent_id  = c.uuid  AND c.type  = 'COMPANY'
JOIN tenants  t  ON cc.tenant_id = t.id
WHERE cc.type = 'COST_CENTER'
  AND cc.deleted_at IS NULL
  AND d.deleted_at  IS NULL
  AND r.deleted_at  IS NULL
  AND c.deleted_at  IS NULL;

ALTER VIEW v_org_structure SET (security_invoker = true);

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
FROM orgunits cc
JOIN orgunits d ON cc.parent_id = d.uuid
JOIN orgunits r ON d.parent_id  = r.uuid
JOIN orgunits c ON r.parent_id  = c.uuid
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
FROM orgunits d
JOIN orgunits r  ON d.parent_id  = r.uuid
JOIN orgunits c  ON r.parent_id  = c.uuid
JOIN tenants  t  ON d.tenant_id  = t.id
LEFT JOIN orgunits cc ON cc.parent_id = d.uuid
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
-- All orgunits under each company resolved via the hierarchy_paths closure table. Returns every
-- descendant with its depth (levels from the company root).
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_company_structure AS
SELECT
  t.name        AS tenant_name,
  c.uuid        AS company_id,
  c.name        AS company_name,
  c.code        AS company_code,
  e.uuid        AS orgunit_id,
  e.name        AS orgunit_name,
  e.type        AS orgunit_type,
  e.code        AS orgunit_code,
  hp.depth      AS levels_from_company
FROM orgunits c
JOIN hierarchy_paths hp ON c.uuid   = hp.ancestor_id
JOIN orgunits        e  ON hp.descendant_id = e.uuid
JOIN tenants         t  ON c.tenant_id      = t.id
WHERE c.type         = 'COMPANY'
  AND c.deleted_at   IS NULL
  AND e.deleted_at   IS NULL;

ALTER VIEW v_company_structure SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_TENANT_ORG_SUMMARY VIEW
-- ------------------------------------------------------------------------------------------------
-- Orgunit counts per tenant broken down by type (COMPANY, REGIONAL, DEPARTMENT, COST_CENTER,
-- PROJECT) plus active and non-deleted totals. Used for tenant dashboard metrics.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_tenant_org_summary AS
SELECT
  t.id          AS tenant_id,
  t.name        AS tenant_name,
  t."Status"    AS tenant_status,
  COUNT(DISTINCT e.uuid)                                                          AS total_orgunits,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'COMPANY')                       AS company_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type IN ('REGION', 'REGIONAL'))         AS regional_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'DEPARTMENT')                    AS department_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'COST_CENTER')                   AS cost_center_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.type = 'PROJECT')                       AS project_count,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.is_active = TRUE)                       AS active_orgunits,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.deleted_at IS NULL)                     AS non_deleted_orgunits
FROM tenants t
LEFT JOIN orgunits e ON e.tenant_id = t.id
GROUP BY t.id, t.name, t."Status";

ALTER VIEW v_tenant_org_summary SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ACTIVE_ORGUNITS VIEW
-- ------------------------------------------------------------------------------------------------
-- Active (non-deleted, is_active=TRUE) orgunits enriched with their nearest typed ancestor names
-- (company, regional, department) resolved via the hierarchy_paths closure table.
--
-- Uses CTEs with DISTINCT ON to find the nearest ancestor of each type — avoids the
-- CASE WHEN EXISTS correlated-subquery anti-pattern that forces a per-row subquery evaluation.
--
-- NOTE: No ORDER BY — add ORDER BY in the calling query.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_active_orgunits AS
WITH company_anc AS (
  -- Nearest COMPANY ancestor for each orgunit (depth ASC picks closest)
  SELECT DISTINCT ON (hp.descendant_id)
    hp.descendant_id,
    a.name
  FROM hierarchy_paths hp
  JOIN orgunits a ON a.uuid = hp.ancestor_id
  WHERE a.type = 'COMPANY'
    AND a.deleted_at IS NULL
  ORDER BY hp.descendant_id, hp.depth ASC
),
regional_anc AS (
  SELECT DISTINCT ON (hp.descendant_id)
    hp.descendant_id,
    a.name
  FROM hierarchy_paths hp
  JOIN orgunits a ON a.uuid = hp.ancestor_id
  WHERE a.type IN ('REGION', 'REGIONAL')
    AND a.deleted_at IS NULL
  ORDER BY hp.descendant_id, hp.depth ASC
),
dept_anc AS (
  SELECT DISTINCT ON (hp.descendant_id)
    hp.descendant_id,
    a.name
  FROM hierarchy_paths hp
  JOIN orgunits a ON a.uuid = hp.ancestor_id
  WHERE a.type = 'DEPARTMENT'
    AND a.deleted_at IS NULL
  ORDER BY hp.descendant_id, hp.depth ASC
)
SELECT
  t.name        AS tenant_name,
  e.uuid        AS orgunit_id,
  e.name        AS orgunit_name,
  e.type        AS orgunit_type,
  e.code        AS orgunit_code,
  ca.name       AS company_name,
  ra.name       AS regional_name,
  da.name       AS department_name,
  e.is_active,
  e.created_at,
  e.updated_at
FROM orgunits e
JOIN tenants      t  ON e.tenant_id      = t.id
LEFT JOIN company_anc  ca ON ca.descendant_id = e.uuid
LEFT JOIN regional_anc ra ON ra.descendant_id = e.uuid
LEFT JOIN dept_anc     da ON da.descendant_id = e.uuid
WHERE e.deleted_at IS NULL
  AND e.is_active  = TRUE;

ALTER VIEW v_active_orgunits SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ORG_CHANGES VIEW
-- ------------------------------------------------------------------------------------------------
-- Lifecycle change log classifying each orgunit row as CREATED, MODIFIED, or DELETED based on
-- its timestamp fields.
--
-- NOTE: No ORDER BY — callers add ORDER BY as needed. ORDER BY inside a view definition is not
--       guaranteed to propagate and forces a sort materialisation even when unused by the caller.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_changes AS
SELECT
  t.name            AS tenant_name,
  e.uuid            AS orgunit_id,
  e.name            AS orgunit_name,
  e.type            AS orgunit_type,
  e.validation_status,
  e.created_at,
  e.updated_at,
  e.deleted_at,
  CASE
    WHEN e.deleted_at IS NOT NULL                              THEN 'DELETED'
    WHEN e.updated_at > e.created_at + INTERVAL '1 minute'    THEN 'MODIFIED'
    ELSE                                                            'CREATED'
  END AS change_type
FROM orgunits e
JOIN tenants t ON e.tenant_id = t.id;

ALTER VIEW v_org_changes SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_ORG_PATHS VIEW
-- ------------------------------------------------------------------------------------------------
-- Flattened ancestor-descendant relationships from the hierarchy_paths closure table enriched
-- with orgunit names and types. Handy for path-based permission checks and tree traversals.
--
-- NOTE: No ORDER BY — sort in the calling query.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_org_paths AS
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
JOIN orgunits a ON hp.ancestor_id   = a.uuid
JOIN orgunits d ON hp.descendant_id = d.uuid
JOIN tenants  t ON hp.tenant_id     = t.id
WHERE a.deleted_at IS NULL
  AND d.deleted_at IS NULL;

ALTER VIEW v_org_paths SET (security_invoker = true);

-- ------------------------------------------------------------------------------------------------
-- V_TENANT_RESOURCE_UTILIZATION VIEW
-- ------------------------------------------------------------------------------------------------
-- Sequence state and orgunit counts per tenant. Combines orgunit stats with orgscope stats to
-- give an overview of how many document types and sequence slots each tenant is using.
-- ------------------------------------------------------------------------------------------------
CREATE VIEW v_tenant_resource_utilization AS
SELECT
  t.id                                                              AS tenant_id,
  t.name                                                            AS tenant_name,
  t."Status"                                                        AS tenant_status,
  COUNT(DISTINCT e.uuid)                                            AS total_orgunits,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.is_active = TRUE)          AS active_orgunits,
  COUNT(DISTINCT e.uuid) FILTER (WHERE e.deleted_at IS NULL)        AS non_deleted_orgunits,
  COUNT(DISTINCT es.uuid)                                           AS sequence_states,
  COUNT(DISTINCT es.key)                                            AS document_types,
  MAX(e.created_at)                                                 AS last_orgunit_created,
  MAX(e.updated_at)                                                 AS last_orgunit_updated
FROM tenants t
LEFT JOIN orgunits  e  ON e.tenant_id  = t.id
LEFT JOIN orgscope  es ON es.tenant_id = t.id
GROUP BY t.id, t.name, t."Status";

ALTER VIEW v_tenant_resource_utilization SET (security_invoker = true);
