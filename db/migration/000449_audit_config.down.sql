-- ------------------------------------------------------------------------------------------------
-- DOWN MIGRATION FOR 000449_AUDIT_CONFIGURATION_TABLES
-- ------------------------------------------------------------------------------------------------
-- Reverts all changes made in the up migration, including:
--   - Dropping the registration functions
--   - Dropping the configuration tables
--   - Removing all platform seed data (handled automatically by DROP TABLE)
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- DROP REGISTRATION FUNCTIONS
-- ------------------------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS register_audit_sensitive_field(
  p_field_name       TEXT,
  p_risk_weight      INTEGER,
  p_compliance_flags JSONB,
  p_module_name      TEXT
) CASCADE;

DROP FUNCTION IF EXISTS register_audit_sensitive_table(
  p_table_name           TEXT,
  p_risk_weight          INTEGER,
  p_event_category       VARCHAR(50),
  p_severity_on_delete   VARCHAR(20),
  p_compliance_flags     JSONB,
  p_module_name          TEXT
) CASCADE;

-- ------------------------------------------------------------------------------------------------
-- DROP CONFIGURATION TABLES
-- ------------------------------------------------------------------------------------------------
-- Note: CASCADE will automatically drop any dependent objects (views, foreign keys, etc.)
-- The seed data is removed automatically when the tables are dropped.
-- ------------------------------------------------------------------------------------------------
DROP TABLE IF EXISTS audit_sensitive_fields CASCADE;
DROP TABLE IF EXISTS audit_sensitive_tables CASCADE;

-- ------------------------------------------------------------------------------------------------
-- VERIFICATION (optional - uncomment if you want to verify cleanup)
-- ------------------------------------------------------------------------------------------------
-- SELECT COUNT(*) FROM audit_sensitive_tables;     -- Should return error (table doesn't exist)
-- SELECT COUNT(*) FROM audit_sensitive_fields;     -- Should return error (table doesn't exist)
