-- ------------------------------------------------------------------------------------------------
-- ORG_UNITS
-- ------------------------------------------------------------------------------------------------
-- Root organisational unit table with hierarchical structure and per-unit accounting preferences.
--
-- An "org unit" is any node in the company hierarchy — the root company itself, a subsidiary,
-- a region, a department, a cost centre, a project, etc. Using a single table with a type
-- discriminator (rather than separate tables per level) keeps hierarchy traversal simple and
-- allows arbitrary depth re-parenting without schema changes.
--
-- type IN ('COMPANY','SUBSIDIARY','REGION','BRANCH','LOCATION','DEPARTMENT','DIVISION',
--           'COST_CENTER','PROJECT','BUDGET_UNIT').
-- validation_status IN ('PENDING','VALID','WARNING','ERROR').
--
-- MAINTENANCE NOTES:
--   • org_unit_path and org_level are maintained by the application layer on create/reparent.
--     See fn_rebuild_org_unit_paths() in migration 000202 for the closure-table counterpart.
--   • The no_self_parent CHECK and valid_fy_start_month CHECK are added after table creation
--     (below) so they carry explicit constraint names useful for error handling.
--   • FK to tenants(id) requires the tenants migration to run first.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE org_units (
  uuid                UUID         PRIMARY KEY,
  tenant_id           UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  parent_id           UUID         REFERENCES org_units(uuid) ON DELETE CASCADE,   -- Self-reference for hierarchy
  name                VARCHAR(255) NOT NULL,
  code                VARCHAR(50),                                                  -- Optional internal reference code
  type                VARCHAR(20)  NOT NULL DEFAULT 'COMPANY' CHECK (
                                     type IN (
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
  hidden              BOOLEAN      NOT NULL DEFAULT FALSE,
  accrual_method      BOOLEAN      NOT NULL,                                        -- TRUE = Accrual, FALSE = Cash
  fy_start_month      INTEGER      NOT NULL,                                        -- Named constraint added below
  address             JSONB        NOT NULL DEFAULT '{}'::jsonb,
  picture             VARCHAR(100),                                                 -- File path or URL to org unit logo
  -- Materialized path for O(1) subtree access checks.
  -- Format: '/root_uuid/parent_uuid/this_uuid/' (leading + trailing slash).
  -- Root org units: '/uuid/'. Maintained by application layer on create/reparent.
  org_unit_path       TEXT,
  org_level           INTEGER      NOT NULL DEFAULT 1,                              -- 1 = root (COMPANY), increments per level (max 8)
  settings            JSONB        NOT NULL DEFAULT '{}'::jsonb,
  metadata            JSONB        NOT NULL DEFAULT '{}'::jsonb,
  version             INTEGER      NOT NULL DEFAULT 1,
  last_validation_run TIMESTAMPTZ,
  validation_status   VARCHAR(20)  NOT NULL DEFAULT 'PENDING' CHECK (
                                     validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
                                   ),
  validation_errors   JSONB        NOT NULL DEFAULT '[]'::jsonb,
  created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at          TIMESTAMPTZ,
  UNIQUE (tenant_id, name)
);

COMMENT ON TABLE  org_units                    IS 'Master table for all organisational units — companies, subsidiaries, regions, departments, cost centres, projects, and any other division. Each unit may maintain its own accounting books, customers, vendors, and fiscal year settings. Units form an arbitrarily deep hierarchy via parent_id (max depth 8, enforced by trigger).';
COMMENT ON COLUMN org_units.uuid               IS 'Primary key — unique identifier for the org unit.';
COMMENT ON COLUMN org_units.tenant_id          IS 'FK to tenants — scopes the org unit to a specific tenant.';
COMMENT ON COLUMN org_units.parent_id          IS 'Self-referencing FK — creates the parent–child relationship between org units (e.g. a department under a region). NULL for root units (top-level companies).';
COMMENT ON COLUMN org_units.name               IS 'Display name of the org unit — must be unique within a tenant.';
COMMENT ON COLUMN org_units.code               IS 'Optional short reference code — used for abbreviated identification in reports and document numbering prefixes. Unique within a tenant when set.';
COMMENT ON COLUMN org_units.type               IS 'Discriminator for the organisational level and purpose of this unit.';
COMMENT ON COLUMN org_units.is_active          IS 'Operational status flag — FALSE indicates the unit is dormant and should be excluded from active listings.';
COMMENT ON COLUMN org_units.hidden             IS 'UI visibility flag — TRUE hides the unit from standard listings and reports without deactivating it.';
COMMENT ON COLUMN org_units.accrual_method     IS 'Accounting method for this unit: TRUE = accrual, FALSE = cash.';
COMMENT ON COLUMN org_units.fy_start_month     IS 'Fiscal year start month (1 = January … 12 = December). Allows each unit to run a different fiscal calendar.';
COMMENT ON COLUMN org_units.address            IS 'Physical address as a flexible JSON object (street, city, state, postcode, country, etc.).';
COMMENT ON COLUMN org_units.picture            IS 'Logo or image reference for the org unit — file path or URL.';
COMMENT ON COLUMN org_units.org_unit_path      IS 'Materialized ancestor path: /uuid1/uuid2/this_uuid/. Enables subtree queries via LIKE ''/root/%''. Root units: /uuid/. Populated by app layer on create/reparent — treat as read-only outside of those operations.';
COMMENT ON COLUMN org_units.org_level          IS 'Hierarchy depth: 1 = root COMPANY, increments by 1 per level (max 8). Used to determine query scope: level 1 = entire tenant, leaf = this unit only, else = subtree.';
COMMENT ON COLUMN org_units.settings           IS 'Unit-specific configuration overrides — JSON object for any customisable preferences (document numbering, display options, etc.).';
COMMENT ON COLUMN org_units.metadata           IS 'Arbitrary key-value metadata for integrations and extensions. Not interpreted by core application logic.';
COMMENT ON COLUMN org_units.version            IS 'Optimistic-locking counter — incremented on every update. Compare-and-swap before writing to detect concurrent modifications.';
COMMENT ON COLUMN org_units.last_validation_run IS 'Timestamp of the most recent validation job run against this unit.';
COMMENT ON COLUMN org_units.validation_status  IS 'Result of the last validation run: PENDING (not yet run), VALID, WARNING (valid but with advisories), ERROR (failed validation).';
COMMENT ON COLUMN org_units.validation_errors  IS 'Structured list of validation error objects from the last run. Empty array when status is VALID or PENDING.';
COMMENT ON COLUMN org_units.created_at         IS 'Record creation timestamp — set once on INSERT.';
COMMENT ON COLUMN org_units.updated_at         IS 'Last modification timestamp — refreshed by the set_updated_at trigger on every UPDATE.';
COMMENT ON COLUMN org_units.deleted_at         IS 'Soft-deletion timestamp — NULL for live records, non-NULL for logically deleted units. Hard DELETE is never used.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------

-- Unique code per tenant — sparse (only when code IS NOT NULL)
CREATE UNIQUE INDEX org_units_tenant_code_uidx
  ON org_units (tenant_id, code)
  WHERE code IS NOT NULL;

-- Common filter: all units of a given type within a tenant
CREATE INDEX org_units_tenant_type_idx
  ON org_units (tenant_id, type);

-- Hierarchy traversal — direct children of a parent
CREATE INDEX org_units_parent_idx
  ON org_units (parent_id);

-- Active-only queries — partial index excludes the inactive minority
CREATE INDEX org_units_active_idx
  ON org_units (is_active)
  WHERE is_active = TRUE;

-- Inactive-only queries — for auditing and reactivation workflows
CREATE INDEX org_units_inactive_idx
  ON org_units (tenant_id, is_active)
  WHERE is_active = FALSE;

-- Subtree LIKE queries via materialized path — excludes deleted rows
CREATE INDEX org_units_path_idx
  ON org_units (tenant_id, org_unit_path)
  WHERE org_unit_path IS NOT NULL
    AND deleted_at IS NULL;

-- Org-level filtering for dashboard tiles and level-scoped reports
CREATE INDEX org_units_level_idx
  ON org_units (tenant_id, org_level)
  WHERE deleted_at IS NULL;

-- JSON containment queries on settings (e.g. units with a specific feature flag)
CREATE INDEX org_units_settings_gin_idx
  ON org_units USING gin (settings)
  WHERE settings != '{}'::jsonb;

-- JSON containment queries on address
CREATE INDEX org_units_address_gin_idx
  ON org_units USING gin (address);

-- Soft-deleted record lookups (audit / recovery)
CREATE INDEX org_units_deleted_idx
  ON org_units (deleted_at)
  WHERE deleted_at IS NOT NULL;

-- Broad tenant-level scans
CREATE INDEX org_units_tenant_idx
  ON org_units (tenant_id);

COMMENT ON INDEX org_units_tenant_code_uidx  IS 'Ensures org unit codes are unique within each tenant. Sparse — only applies when code is set.';
COMMENT ON INDEX org_units_tenant_type_idx   IS 'Optimises listings and counts filtered by tenant + unit type (the most common query pattern).';
COMMENT ON INDEX org_units_parent_idx        IS 'Speeds up direct-children lookups during hierarchy traversal.';
COMMENT ON INDEX org_units_active_idx        IS 'Partial index for active-unit queries — skips the inactive rows to save space and I/O.';
COMMENT ON INDEX org_units_inactive_idx      IS 'Partial index for inactive-unit auditing and reactivation workflows.';
COMMENT ON INDEX org_units_path_idx          IS 'Supports subtree containment checks via LIKE ''/root_uuid/%'' on the materialized path. Excludes deleted rows.';
COMMENT ON INDEX org_units_level_idx         IS 'Enables efficient org-level scoping (e.g. fetch all level-2 units for a tenant). Excludes deleted rows.';
COMMENT ON INDEX org_units_settings_gin_idx  IS 'GIN index for JSONB containment queries on settings. Partial — skips empty-settings rows.';
COMMENT ON INDEX org_units_address_gin_idx   IS 'GIN index for address component containment queries (e.g. all units in a city).';
COMMENT ON INDEX org_units_deleted_idx       IS 'Partial index for soft-deleted row lookups — used by audit and recovery queries.';
COMMENT ON INDEX org_units_tenant_idx        IS 'Broad tenant-scoped scans when no other filter is available.';

-- ------------------------------------------------------------------------------------------------
-- DATA INTEGRITY CONSTRAINTS
-- ------------------------------------------------------------------------------------------------

-- Prevent an org unit from listing itself as its own parent.
-- Multi-hop cycles are caught by the check_org_unit_hierarchy_depth trigger in migration 000202.
ALTER TABLE org_units
  ADD CONSTRAINT no_self_parent
  CHECK (uuid != parent_id);

-- Fiscal year start month must be a valid calendar month.
ALTER TABLE org_units
  ADD CONSTRAINT valid_fy_start_month
  CHECK (fy_start_month BETWEEN 1 AND 12);

COMMENT ON CONSTRAINT no_self_parent       ON org_units IS 'Prevents an org unit from being its own parent. Multi-hop cycle detection is handled by the check_org_unit_hierarchy_depth trigger.';
COMMENT ON CONSTRAINT valid_fy_start_month ON org_units IS 'Fiscal year start month must be between 1 (January) and 12 (December).';

-- ------------------------------------------------------------------------------------------------
-- UPDATED_AT TRIGGER
-- ------------------------------------------------------------------------------------------------
-- Refreshes updated_at automatically on every UPDATE so callers never forget to set it.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_org_unit_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION set_org_unit_updated_at IS
  'Trigger function — sets updated_at to NOW() before every UPDATE on org_units. '
  'Ensures the column stays accurate without requiring callers to set it explicitly.';

CREATE TRIGGER org_units_set_updated_at
  BEFORE UPDATE ON org_units
  FOR EACH ROW
  EXECUTE FUNCTION set_org_unit_updated_at();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE org_units ENABLE ROW LEVEL SECURITY;
ALTER TABLE org_units FORCE  ROW LEVEL SECURITY;

-- application_role: full DML, scoped to the current tenant
CREATE POLICY org_units_tenant_isolation ON org_units
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

-- admin_role: unrestricted access across all tenants (support / migrations)
CREATE POLICY org_units_admin_access ON org_units
  FOR ALL TO admin_role
  USING (TRUE)
  WITH CHECK (TRUE);

-- readonly_role: SELECT only, still tenant-scoped
CREATE POLICY org_units_readonly_select ON org_units
  FOR SELECT TO readonly_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON org_units TO application_role;
