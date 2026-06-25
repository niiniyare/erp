-- ------------------------------------------------------------------------------------------------
-- DOC_SEQUENCES
-- ------------------------------------------------------------------------------------------------
-- Tracks the next available sequential number for each business document type (invoice, purchase
-- order, estimate, bill, receipt, etc.) scoped by org unit and optional fiscal year.
--
-- Design decisions:
--   • One row per (tenant, org_unit, doc_type, fiscal_year) combination.
--   • fiscal_year is nullable: NULL means the sequence never resets by year.
--   • sequence holds the NEXT number to issue, not the last issued number.
--     Callers must use SELECT … FOR UPDATE or an advisory lock before reading and incrementing
--     to avoid gaps under concurrent document creation.
--   • format_config overrides tenant-level defaults for prefix/suffix/padding/reset behaviour
--     for this specific org_unit + doc_type combination.
--
-- format_config JSONB keys:
--   prefix            STRING  — prepended to the formatted number (e.g. "NORTH-INV-")
--   suffix            STRING  — appended after the formatted number
--   pad_length        INT     — zero-pad the number to this width (e.g. 6 → "000042")
--   reset_frequency   STRING  — "yearly" | "monthly" | "never"
--   format_template   STRING  — full override template; other keys are ignored when set
--
-- Example: {"prefix":"NORTH-INV-","pad_length":6,"reset_frequency":"yearly"}
--
-- NOTE: FK references to org_units(uuid) are DEFERRABLE INITIALLY DEFERRED to permit
--       inserting a doc_sequences row in the same transaction that creates the org unit.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE doc_sequences (
  uuid           UUID         PRIMARY KEY,
  tenant_id      UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  org_unit_id    UUID         NOT NULL REFERENCES org_units(uuid) DEFERRABLE INITIALLY DEFERRED,
  org_sub_unit_id UUID                 REFERENCES org_units(uuid) DEFERRABLE INITIALLY DEFERRED, -- Optional sub-unit / department override
  doc_type       VARCHAR(10)  NOT NULL,                          -- Document type key: invoice, po, estimate, bill, receipt, …
  fiscal_year    SMALLINT,                                       -- NULL = sequences never reset by year
  sequence       BIGINT       NOT NULL,                          -- Next available sequence number (always > 0)
  format_config  JSONB        NOT NULL DEFAULT '{}'::jsonb,      -- Per-org-unit formatting overrides (see header)
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at     TIMESTAMPTZ
);

COMMENT ON TABLE  doc_sequences                 IS 'Tracks the next available sequential document number for each org unit, document type, and fiscal year. One row per (tenant, org_unit, doc_type, fiscal_year) combination. Callers must lock the row (SELECT … FOR UPDATE) before reading and incrementing to prevent duplicate numbers under concurrent load.';
COMMENT ON COLUMN doc_sequences.uuid            IS 'Primary key — unique row identifier.';
COMMENT ON COLUMN doc_sequences.tenant_id       IS 'FK to tenants — scopes the sequence to a specific tenant.';
COMMENT ON COLUMN doc_sequences.org_unit_id     IS 'FK to org_units — the primary org unit that owns this sequence. Deferred to allow creation in the same transaction as the org unit.';
COMMENT ON COLUMN doc_sequences.org_sub_unit_id IS 'Optional FK to org_units — a sub-unit or department that overrides the org unit sequence for finer-grained numbering. Deferred FK.';
COMMENT ON COLUMN doc_sequences.doc_type        IS 'Document type discriminator (e.g. "invoice", "po", "estimate", "bill", "receipt"). Used as a lookup key together with org_unit_id and fiscal_year.';
COMMENT ON COLUMN doc_sequences.fiscal_year     IS 'Fiscal year that scopes this sequence row. NULL means the sequence is perpetual (never reset by year). When reset_frequency = "yearly", a new row is inserted each fiscal year.';
COMMENT ON COLUMN doc_sequences.sequence        IS 'The NEXT sequence number to issue — must always be > 0. After issuing, increment this value atomically (SELECT … FOR UPDATE + UPDATE).';
COMMENT ON COLUMN doc_sequences.format_config   IS 'Document number formatting overrides for this org unit + doc type combination. Keys: prefix (STRING), suffix (STRING), pad_length (INT), reset_frequency ("yearly"|"monthly"|"never"), format_template (STRING). Overrides tenant-level defaults in tenant_configurations.settings. Example: {"prefix":"NORTH-INV-","pad_length":6,"reset_frequency":"yearly"}.';
COMMENT ON COLUMN doc_sequences.created_at      IS 'Row creation timestamp.';
COMMENT ON COLUMN doc_sequences.updated_at      IS 'Last modification timestamp — refreshed by the set_doc_sequence_updated_at trigger.';
COMMENT ON COLUMN doc_sequences.deleted_at      IS 'Soft-deletion timestamp — NULL for active sequences.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------

-- Primary lookup path: find the sequence row for a given org unit + document type
CREATE INDEX doc_sequences_org_unit_doc_type_idx
  ON doc_sequences (org_unit_id, doc_type);

-- Fiscal-year-scoped lookup (the common case when reset_frequency = "yearly")
CREATE INDEX doc_sequences_org_unit_fiscal_year_idx
  ON doc_sequences (org_unit_id, fiscal_year, doc_type);

-- GIN index for format_config containment queries (e.g. find all sequences using a given prefix)
CREATE INDEX doc_sequences_format_config_gin_idx
  ON doc_sequences USING gin (format_config)
  WHERE format_config != '{}'::jsonb;

COMMENT ON INDEX doc_sequences_org_unit_doc_type_idx    IS 'Primary sequence lookup by org unit and document type.';
COMMENT ON INDEX doc_sequences_org_unit_fiscal_year_idx IS 'Fiscal-year-scoped sequence lookup — used when sequences reset annually or monthly.';
COMMENT ON INDEX doc_sequences_format_config_gin_idx    IS 'GIN index for format_config JSONB containment queries. Partial — skips rows with default empty config.';

-- ------------------------------------------------------------------------------------------------
-- DATA INTEGRITY CONSTRAINTS
-- ------------------------------------------------------------------------------------------------

-- Exactly one sequence row per org unit, document type, and fiscal year within a tenant.
-- For the NULL fiscal_year case (perpetual sequences), a separate partial unique index is
-- required because NULL != NULL in SQL and the regular UNIQUE constraint permits duplicate NULLs.
ALTER TABLE doc_sequences
  ADD CONSTRAINT doc_sequences_unique_scoped
  UNIQUE (tenant_id, org_unit_id, doc_type, fiscal_year);

-- Partial unique index for perpetual sequences (fiscal_year IS NULL).
-- Ensures at most one "no year" sequence per tenant + org unit + doc type.
CREATE UNIQUE INDEX doc_sequences_no_fiscal_year_uidx
  ON doc_sequences (tenant_id, org_unit_id, doc_type)
  WHERE fiscal_year IS NULL;

-- Sequence counter must always be a positive integer.
ALTER TABLE doc_sequences
  ADD CONSTRAINT doc_sequences_positive_sequence
  CHECK (sequence > 0);

COMMENT ON CONSTRAINT doc_sequences_unique_scoped        ON doc_sequences IS 'Prevents duplicate sequence trackers for the same tenant, org unit, document type, and fiscal year. Works for non-NULL fiscal years; the doc_sequences_no_fiscal_year_uidx partial index covers the NULL case.';
COMMENT ON CONSTRAINT doc_sequences_positive_sequence    ON doc_sequences IS 'Sequence counter must always be a positive integer — 0 is never a valid next-number.';
COMMENT ON INDEX      doc_sequences_no_fiscal_year_uidx                   IS 'Partial unique index ensuring at most one perpetual (no fiscal year) sequence exists per tenant + org unit + doc type. Required because the UNIQUE constraint on (…, fiscal_year) permits multiple NULLs.';

-- ------------------------------------------------------------------------------------------------
-- UPDATED_AT TRIGGER
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_doc_sequence_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION set_doc_sequence_updated_at IS
  'Trigger function — sets updated_at to NOW() before every UPDATE on doc_sequences.';

CREATE TRIGGER doc_sequences_set_updated_at
  BEFORE UPDATE ON doc_sequences
  FOR EACH ROW
  EXECUTE FUNCTION set_doc_sequence_updated_at();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE doc_sequences ENABLE ROW LEVEL SECURITY;
ALTER TABLE doc_sequences FORCE  ROW LEVEL SECURITY;

CREATE POLICY doc_sequences_tenant_isolation ON doc_sequences
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY doc_sequences_admin_access ON doc_sequences
  FOR ALL TO admin_role
  USING (TRUE)
  WITH CHECK (TRUE);

CREATE POLICY doc_sequences_readonly_select ON doc_sequences
  FOR SELECT TO readonly_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON doc_sequences TO application_role;
