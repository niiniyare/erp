-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE DEFINITIONS
-- ------------------------------------------------------------------------------------------------
-- Defines attributes used in ABAC policies with validation and encryption controls.
-- data_type IN ('STRING','NUMBER','BOOLEAN','DATE','TIME','JSON','ARRAY','ENUM').
-- category IN ('USER','RESOURCE','ENVIRONMENT','ACTION','ENTITY','SESSION').
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_definitions (
  id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name                VARCHAR(100) NOT NULL,
  display_name        VARCHAR(150),
  description         TEXT,
  data_type           VARCHAR(50)  NOT NULL CHECK (
    data_type IN (
      'STRING',
      'NUMBER',
      'BOOLEAN',
      'DATE',
      'TIME',
      'JSON',
      'ARRAY',
      'ENUM'
    )
  ),
  category            VARCHAR(50)  NOT NULL CHECK (
    category IN (
      'USER',
      'RESOURCE',
      'ENVIRONMENT',
      'ACTION',
      'ENTITY',
      'SESSION'
    )
  ),
  is_required         BOOLEAN      DEFAULT false,
  is_sensitive        BOOLEAN      DEFAULT false,   -- for PII/sensitive attributes
  default_value       TEXT,
  allowed_values      JSONB,                        -- for enum types
  validation_rules    JSONB        DEFAULT '{}'::jsonb,  -- custom validation rules
  encryption_required BOOLEAN      DEFAULT false,   -- whether values must be encrypted
  is_active           BOOLEAN      DEFAULT TRUE,
  created_at          TIMESTAMPTZ  DEFAULT NOW(),
  CONSTRAINT attribute_definitions_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE attribute_definitions IS 'Defines attributes used in ABAC policies with data types, validation rules, and security controls for consistent attribute management.';

COMMENT ON COLUMN attribute_definitions.data_type          IS 'Attribute data type: STRING, NUMBER, BOOLEAN, DATE, TIME, JSON, ARRAY, ENUM';
COMMENT ON COLUMN attribute_definitions.category           IS 'Attribute category: USER (user attributes), RESOURCE (resource attributes), ENVIRONMENT (context), ACTION (action attributes), ENTITY (entity attributes), SESSION (session context)';
COMMENT ON COLUMN attribute_definitions.is_sensitive       IS 'Whether attribute contains PII or sensitive data requiring special handling';
COMMENT ON COLUMN attribute_definitions.allowed_values     IS 'JSONB array of allowed values for ENUM data type';
COMMENT ON COLUMN attribute_definitions.validation_rules   IS 'JSONB containing custom validation rules (regex, ranges, etc.)';
COMMENT ON COLUMN attribute_definitions.encryption_required IS 'Whether attribute values must be encrypted at rest';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE attribute_definitions ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_definitions_tenant_isolation ON attribute_definitions FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);
