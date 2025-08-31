-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE VALUES
-- ------------------------------------------------------------------------------------------------
-- Stores actual attribute values for ABAC policy evaluation.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_values (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  definition_id UUID NOT NULL REFERENCES attribute_definitions(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL,  -- The entity this attribute belongs to (user, resource, etc.)
  value TEXT NOT NULL,  -- The actual attribute value (may be encrypted)
  encrypted_value BYTEA,  -- Encrypted version if encryption is enabled
  is_encrypted BOOLEAN DEFAULT false,
  version INTEGER DEFAULT 1,  -- For versioning/auditing changes
  effective_from TIMESTAMPTZ DEFAULT NOW(),  -- When this value becomes effective
  effective_to TIMESTAMPTZ,  -- When this value expires (nullable for current values)
  created_at TIMESTAMPTZ DEFAULT NOW(),
  created_by UUID REFERENCES users(id),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  updated_by UUID REFERENCES users(id),
  CONSTRAINT attribute_values_unique_current UNIQUE (
    tenant_id,
    definition_id,
    entity_id,
    effective_from
  )
);

COMMENT ON TABLE attribute_values IS 'Stores actual attribute values for entities with versioning, encryption, and temporal support for ABAC policy evaluation.';

COMMENT ON COLUMN attribute_values.entity_id IS 'The UUID of the entity this attribute belongs to (user, resource, document, etc.)';

COMMENT ON COLUMN attribute_values.value IS 'The actual attribute value in string format';

COMMENT ON COLUMN attribute_values.encrypted_value IS 'Encrypted version of the value when encryption is required';

COMMENT ON COLUMN attribute_values.effective_from IS 'When this attribute value becomes effective (for temporal policies)';

COMMENT ON COLUMN attribute_values.effective_to IS 'When this attribute value expires (null for current values)';

-- Enable RLS and create policies
ALTER TABLE
  attribute_values ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_values_tenant_isolation ON attribute_values FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);

-- Create indexes for performance
CREATE INDEX idx_attribute_values_entity_definition ON attribute_values(tenant_id, entity_id, definition_id)
WHERE
  effective_to IS NULL;

CREATE INDEX idx_attribute_values_definition_value ON attribute_values(tenant_id, definition_id, value)
WHERE
  effective_to IS NULL;

CREATE INDEX idx_attribute_values_effective_period ON attribute_values(effective_from, effective_to);

-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE SOURCES
-- ------------------------------------------------------------------------------------------------
-- Defines external sources for attribute collection.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_sources (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  display_name VARCHAR(150),
  description TEXT,
  source_type VARCHAR(50) NOT NULL CHECK (
    source_type IN (
      'LDAP',
      'DATABASE',
      'REST_API',
      'GRAPHQL',
      'FILE',
      'MANUAL'
    )
  ),
  configuration JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Source-specific configuration
  authentication JSONB DEFAULT '{}'::jsonb,  -- Authentication details (encrypted)
  cache_ttl_minutes INTEGER DEFAULT 60,  -- How long to cache attributes from this source
  is_active BOOLEAN DEFAULT TRUE,
  priority INTEGER DEFAULT 100,  -- Source priority for attribute resolution
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT attribute_sources_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE attribute_sources IS 'Defines external sources for attribute collection with configuration, authentication, and caching controls.';

COMMENT ON COLUMN attribute_sources.source_type IS 'Type of attribute source: LDAP, DATABASE, REST_API, GRAPHQL, FILE, MANUAL';

COMMENT ON COLUMN attribute_sources.configuration IS 'JSONB containing source-specific configuration (URLs, queries, etc.)';

COMMENT ON COLUMN attribute_sources.authentication IS 'JSONB containing authentication details (should be encrypted)';

COMMENT ON COLUMN attribute_sources.priority IS 'Source priority for attribute resolution (higher numbers processed first)';

-- Enable RLS and create policies
ALTER TABLE
  attribute_sources ENABLE ROW LEVEL SECURITY;

CREATE POLICY attribute_sources_tenant_isolation ON attribute_sources FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);
