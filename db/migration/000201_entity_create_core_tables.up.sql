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
