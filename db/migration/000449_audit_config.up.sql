-- ------------------------------------------------------------------------------------------------
-- AUDIT CONFIGURATION TABLES
-- ------------------------------------------------------------------------------------------------
-- Defines the two configuration tables that drive risk scoring, compliance flag detection,
-- event categorisation, and delete severity across the entire audit system.
--
-- DESIGN INTENT:
--   The audit trigger functions (000451) contain zero hardcoded domain knowledge. All
--   sensitivity rules are data, not code. Any module that wants to participate in audit
--   configuration calls register_audit_sensitive_table() or register_audit_sensitive_field()
--   at the end of its own migration — no changes to core audit files are ever needed.
--
--   Platform-level seed rows (platform core tables and generic PII/financial fields) are
--   inserted here. Module-specific rows (finance, HR, petroleum, healthcare, etc.) belong
--   in their own migration files.
--
-- DEPENDENCIES: None (this migration must run before 000450 and 000451).
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- AUDIT_SENSITIVE_TABLES
-- ------------------------------------------------------------------------------------------------
-- One row per table name that should receive elevated risk treatment.
-- Columns:
--   table_name         — exact table name (no schema prefix; triggers fire per-schema anyway)
--   risk_weight        — flat points added to the base operation score (default 20)
--   event_category     — if non-null, overrides determine_event_category() for this table
--   severity_on_delete — if non-null, overrides determine_severity() for DELETE operations
--   compliance_flags   — JSONB merged into the row's compliance_flags when this table is touched
--   module_name        — which module registered this row (documentation only, not enforced)
--   created_at         — insertion timestamp
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_sensitive_tables (
  table_name           TEXT         NOT NULL,
  risk_weight          INTEGER      NOT NULL DEFAULT 20 CHECK (risk_weight BETWEEN 0 AND 100),
  event_category       VARCHAR(50)  CHECK (
    event_category IS NULL OR event_category IN
      ('ACCESS','ADMIN','DATA','AUTH','SYSTEM','COMPLIANCE')
  ),
  severity_on_delete   VARCHAR(20)  CHECK (
    severity_on_delete IS NULL OR severity_on_delete IN
      ('LOW','INFO','WARN','HIGH','CRITICAL')
  ),
  compliance_flags     JSONB        NOT NULL DEFAULT '{}',
  module_name          TEXT,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  PRIMARY KEY (table_name)
);

COMMENT ON TABLE  audit_sensitive_tables IS
  'Registry of tables that receive elevated risk scores, compliance flags, or severity overrides '
  'in the audit system. Populated by module migrations via register_audit_sensitive_table(). '
  'Platform core rows are seeded in 000449.';
COMMENT ON COLUMN audit_sensitive_tables.event_category     IS 'NULL = use default category logic in determine_event_category()';
COMMENT ON COLUMN audit_sensitive_tables.severity_on_delete IS 'NULL = use default severity logic in determine_severity()';
COMMENT ON COLUMN audit_sensitive_tables.compliance_flags   IS 'JSONB map merged into audit_log.compliance_flags when this table is the audit target';
COMMENT ON COLUMN audit_sensitive_tables.module_name        IS 'Informational: which module or migration owns this row';

-- ------------------------------------------------------------------------------------------------
-- AUDIT_SENSITIVE_FIELDS
-- ------------------------------------------------------------------------------------------------
-- One row per field name (column name) that is considered sensitive regardless of which table
-- it appears in. Risk weight is added once per changed instance during UPDATE diff.
-- Columns:
--   field_name       — exact column name
--   risk_weight      — points added per changed field in an UPDATE diff (default 15)
--   compliance_flags — JSONB merged into the row's compliance_flags when this field is present
--   module_name      — which module registered this row
--   created_at       — insertion timestamp
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_sensitive_fields (
  field_name       TEXT        NOT NULL,
  risk_weight      INTEGER     NOT NULL DEFAULT 15 CHECK (risk_weight BETWEEN 0 AND 100),
  compliance_flags JSONB       NOT NULL DEFAULT '{}',
  module_name      TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (field_name)
);

COMMENT ON TABLE  audit_sensitive_fields IS
  'Registry of column names considered sensitive regardless of table. When a sensitive field '
  'is present in (or changed on) an audited row the field''s risk_weight and compliance_flags '
  'are applied. Populated by module migrations via register_audit_sensitive_field().';
COMMENT ON COLUMN audit_sensitive_fields.compliance_flags IS 'JSONB map merged into audit_log.compliance_flags when this field is present in the row being audited';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
-- Both tables are tiny and queried by PK only inside trigger functions;
-- the PKs cover all necessary lookups. No additional indexes are needed.

-- ------------------------------------------------------------------------------------------------
-- REGISTER_AUDIT_SENSITIVE_TABLE
-- ------------------------------------------------------------------------------------------------
-- Upserts a table registration into audit_sensitive_tables.
-- Idempotent: re-running with the same table_name updates in place.
-- Intended to be called from module migration files, not from application code.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION register_audit_sensitive_table(
  p_table_name           TEXT,
  p_risk_weight          INTEGER     DEFAULT 20,
  p_event_category       VARCHAR(50) DEFAULT NULL,
  p_severity_on_delete   VARCHAR(20) DEFAULT NULL,
  p_compliance_flags     JSONB       DEFAULT '{}',
  p_module_name          TEXT        DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
  INSERT INTO audit_sensitive_tables (
    table_name, risk_weight, event_category,
    severity_on_delete, compliance_flags, module_name
  )
  VALUES (
    p_table_name, p_risk_weight, p_event_category,
    p_severity_on_delete, p_compliance_flags, p_module_name
  )
  ON CONFLICT (table_name) DO UPDATE SET
    risk_weight        = EXCLUDED.risk_weight,
    event_category     = EXCLUDED.event_category,
    severity_on_delete = EXCLUDED.severity_on_delete,
    compliance_flags   = EXCLUDED.compliance_flags,
    module_name        = EXCLUDED.module_name;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION register_audit_sensitive_table IS
  'Upserts a table entry into audit_sensitive_tables. Call from module migration files to '
  'register domain-specific sensitivity rules without touching core audit migrations.';

-- ------------------------------------------------------------------------------------------------
-- REGISTER_AUDIT_SENSITIVE_FIELD
-- ------------------------------------------------------------------------------------------------
-- Upserts a field registration into audit_sensitive_fields.
-- Idempotent: re-running with the same field_name updates in place.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION register_audit_sensitive_field(
  p_field_name       TEXT,
  p_risk_weight      INTEGER DEFAULT 15,
  p_compliance_flags JSONB   DEFAULT '{}',
  p_module_name      TEXT    DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
  INSERT INTO audit_sensitive_fields (
    field_name, risk_weight, compliance_flags, module_name
  )
  VALUES (
    p_field_name, p_risk_weight, p_compliance_flags, p_module_name
  )
  ON CONFLICT (field_name) DO UPDATE SET
    risk_weight      = EXCLUDED.risk_weight,
    compliance_flags = EXCLUDED.compliance_flags,
    module_name      = EXCLUDED.module_name;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION register_audit_sensitive_field IS
  'Upserts a field entry into audit_sensitive_fields. Call from module migration files to '
  'register domain-specific field sensitivity without touching core audit migrations.';

-- ------------------------------------------------------------------------------------------------
-- PLATFORM SEED DATA
-- ------------------------------------------------------------------------------------------------
-- These rows represent the platform layer only — auth, IAM, tenancy, and system tables that
-- every deployment will have. No ERP module, industry, or jurisdiction is referenced here.
--
-- MODULE EXAMPLE (to be placed in the relevant module migration, NOT here):
--
--   -- finance module (000600_finance_accounts.up.sql)
--   SELECT register_audit_sensitive_table(
--     'finance_accounts', 20, 'DATA', 'HIGH', '{"SOX": true}', 'finance'
--   );
--   SELECT register_audit_sensitive_field(
--     'balance', 15, '{"SOX": true}', 'finance'
--   );
-- ------------------------------------------------------------------------------------------------

-- Platform core: IAM tables
SELECT register_audit_sensitive_table('users',         20, 'AUTH',  'HIGH',     '{}',              'platform');
SELECT register_audit_sensitive_table('user_sessions', 15, 'AUTH',  'HIGH',     '{}',              'platform');
SELECT register_audit_sensitive_table('auth_tokens',   15, 'AUTH',  'HIGH',     '{}',              'platform');
SELECT register_audit_sensitive_table('roles',         20, 'ADMIN', 'HIGH',     '{}',              'platform');
SELECT register_audit_sensitive_table('permissions',   20, 'ADMIN', 'HIGH',     '{}',              'platform');
SELECT register_audit_sensitive_table('policies',      20, 'ADMIN', 'HIGH',     '{}',              'platform');
SELECT register_audit_sensitive_table('tenants',       30, 'ADMIN', 'CRITICAL', '{}',              'platform');

-- Platform core: generic PII fields (GDPR-relevant regardless of which table they appear in)
SELECT register_audit_sensitive_field('email',           15, '{"GDPR": true}', 'platform');
SELECT register_audit_sensitive_field('phone',           15, '{"GDPR": true}', 'platform');
SELECT register_audit_sensitive_field('address',         15, '{"GDPR": true}', 'platform');
SELECT register_audit_sensitive_field('date_of_birth',   15, '{"GDPR": true}', 'platform');
SELECT register_audit_sensitive_field('national_id',     15, '{"GDPR": true}', 'platform');
SELECT register_audit_sensitive_field('passport_number', 15, '{"GDPR": true}', 'platform');
SELECT register_audit_sensitive_field('ssn',             15, '{"GDPR": true}', 'platform');

-- Platform core: credentials (always sensitive, no specific compliance flag)
SELECT register_audit_sensitive_field('password',        15, '{}',             'platform');
SELECT register_audit_sensitive_field('password_hash',   15, '{}',             'platform');
SELECT register_audit_sensitive_field('secret',          15, '{}',             'platform');
SELECT register_audit_sensitive_field('token',           10, '{}',             'platform');
SELECT register_audit_sensitive_field('refresh_token',   10, '{}',             'platform');

-- Platform core: payment card fields (PCI-DSS regardless of table)
SELECT register_audit_sensitive_field('credit_card',     15, '{"PCI_DSS": true}', 'platform');
SELECT register_audit_sensitive_field('card_number',     15, '{"PCI_DSS": true}', 'platform');
SELECT register_audit_sensitive_field('cvv',             15, '{"PCI_DSS": true}', 'platform');
SELECT register_audit_sensitive_field('card_expiry',     10, '{"PCI_DSS": true}', 'platform');
