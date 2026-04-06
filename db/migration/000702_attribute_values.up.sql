-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE VALUES
-- ------------------------------------------------------------------------------------------------
-- Stores actual attribute values for ABAC policy evaluation with versioning,
-- encryption, and temporal (effective_from / effective_to) support.
--
-- NOTE: Depends on attribute_definitions (000701).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_values (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  definition_id   UUID        NOT NULL REFERENCES attribute_definitions(id) ON DELETE CASCADE,
  entity_id       UUID        NOT NULL,  -- the entity this attribute belongs to (user, resource, etc.)
  value           TEXT        NOT NULL,  -- the actual attribute value (may be encrypted)
  encrypted_value BYTEA,                 -- encrypted version if encryption is enabled
  is_encrypted    BOOLEAN     DEFAULT false,
  version         INTEGER     DEFAULT 1, -- for versioning/auditing changes
  effective_from  TIMESTAMPTZ DEFAULT NOW(),  -- when this value becomes effective
  effective_to    TIMESTAMPTZ,               -- when this value expires (nullable for current values)
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  created_by      UUID        REFERENCES users(id),
  updated_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_by      UUID        REFERENCES users(id),
  CONSTRAINT attribute_values_unique_current UNIQUE (
    tenant_id,
    definition_id,
    entity_id,
    effective_from
  )
);

COMMENT ON TABLE attribute_values IS 'Stores actual attribute values for entities with versioning, encryption, and temporal support for ABAC policy evaluation.';

COMMENT ON COLUMN attribute_values.entity_id       IS 'The UUID of the entity this attribute belongs to (user, resource, document, etc.)';
COMMENT ON COLUMN attribute_values.value           IS 'The actual attribute value in string format';
COMMENT ON COLUMN attribute_values.encrypted_value IS 'Encrypted version of the value when encryption is required';
COMMENT ON COLUMN attribute_values.effective_from  IS 'When this attribute value becomes effective (for temporal policies)';
COMMENT ON COLUMN attribute_values.effective_to    IS 'When this attribute value expires (null for current values)';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE attribute_values ENABLE ROW LEVEL SECURITY;
ALTER TABLE attribute_values FORCE  ROW LEVEL SECURITY;

CREATE POLICY attribute_values_tenant_isolation ON attribute_values FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY attribute_values_admin_access ON attribute_values FOR ALL TO admin_role
    USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY attribute_values_ro_select ON attribute_values
    FOR SELECT TO readonly_role
    USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_attribute_values_entity_definition ON attribute_values(tenant_id, entity_id, definition_id)
  WHERE effective_to IS NULL;  -- active values only

CREATE INDEX idx_attribute_values_definition_value ON attribute_values(tenant_id, definition_id, value)
  WHERE effective_to IS NULL;  -- active values only

CREATE INDEX idx_attribute_values_effective_period ON attribute_values(effective_from, effective_to);

-- ------------------------------------------------------------------------------------------------
-- ATTRIBUTE SOURCES
-- ------------------------------------------------------------------------------------------------
-- Defines external sources for attribute collection with configuration,
-- authentication, and caching controls.
-- source_type IN ('LDAP','DATABASE','REST_API','GRAPHQL','FILE','MANUAL').
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attribute_sources (
  id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name           VARCHAR(100) NOT NULL,
  display_name   VARCHAR(150),
  description    TEXT,
  source_type    VARCHAR(50)  NOT NULL CHECK (
    source_type IN (
      'LDAP',
      'DATABASE',
      'REST_API',
      'GRAPHQL',
      'FILE',
      'MANUAL'
    )
  ),
  configuration  JSONB        NOT NULL DEFAULT '{}'::jsonb,  -- source-specific configuration
  authentication JSONB        DEFAULT '{}'::jsonb,           -- authentication details (encrypted)
  cache_ttl_minutes INTEGER   DEFAULT 60,   -- how long to cache attributes from this source
  is_active      BOOLEAN      DEFAULT TRUE,
  priority       INTEGER      DEFAULT 100,  -- source priority for attribute resolution
  created_at     TIMESTAMPTZ  DEFAULT NOW(),
  updated_at     TIMESTAMPTZ  DEFAULT NOW(),
  CONSTRAINT attribute_sources_name_unique_per_tenant UNIQUE (tenant_id, name)
);

COMMENT ON TABLE attribute_sources IS 'Defines external sources for attribute collection with configuration, authentication, and caching controls.';

COMMENT ON COLUMN attribute_sources.source_type    IS 'Type of attribute source: LDAP, DATABASE, REST_API, GRAPHQL, FILE, MANUAL';
COMMENT ON COLUMN attribute_sources.configuration  IS 'JSONB containing source-specific configuration (URLs, queries, etc.)';
COMMENT ON COLUMN attribute_sources.authentication IS 'JSONB containing authentication details (should be encrypted)';
COMMENT ON COLUMN attribute_sources.priority       IS 'Source priority for attribute resolution (higher numbers processed first)';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE attribute_sources ENABLE ROW LEVEL SECURITY;
ALTER TABLE attribute_sources FORCE  ROW LEVEL SECURITY;

CREATE POLICY attribute_sources_tenant_isolation ON attribute_sources FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY attribute_sources_admin_access ON attribute_sources FOR ALL TO admin_role
    USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY attribute_sources_ro_select ON attribute_sources
    FOR SELECT TO readonly_role
    USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
