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

COMMENT ON CONSTRAINT positive_sequence ON entitystate IS 'Ensures sequence numbers are always positive values';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE entitystate ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON entitystate FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY admin_full_access_policy ON entitystate FOR ALL TO admin_role USING (TRUE);
