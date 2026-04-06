-- ------------------------------------------------------------------------------------------------
-- DOWN MIGRATION: ATTRIBUTE DEFINITIONS
--
-- Reverts the attribute_definitions table and associated policies.
-- ------------------------------------------------------------------------------------------------
-- Drop RLS policies
DROP POLICY IF EXISTS attribute_definitions_tenant_isolation ON attribute_definitions;
DROP POLICY IF EXISTS attribute_definitions_admin_access ON attribute_definitions;
DROP POLICY IF EXISTS attribute_definitions_ro_select ON attribute_definitions;

-- Disable RLS
ALTER TABLE
  attribute_definitions DISABLE ROW LEVEL SECURITY;

-- Drop comments from columns
COMMENT ON COLUMN attribute_definitions.encryption_required IS NULL;

COMMENT ON COLUMN attribute_definitions.validation_rules IS NULL;

COMMENT ON COLUMN attribute_definitions.allowed_values IS NULL;

COMMENT ON COLUMN attribute_definitions.is_sensitive IS NULL;

COMMENT ON COLUMN attribute_definitions.category IS NULL;

COMMENT ON COLUMN attribute_definitions.data_type IS NULL;

-- Drop comment from table
COMMENT ON TABLE attribute_definitions IS NULL;

-- Drop the attribute_definitions table
DROP TABLE IF EXISTS attribute_definitions;
