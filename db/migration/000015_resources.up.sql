-- ------------------------------------------------------------------------------------------------
-- RESOURCES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines system resources that can be protected by permissions (APIs, UI components, data, etc.).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resources (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid),  -- Resource can belong to specific entity
  name VARCHAR(100) NOT NULL,
  display_name VARCHAR(150),
  description TEXT,
  resource_type VARCHAR(50) NOT NULL CHECK (
    resource_type IN (
      'API',
      'UI',
      'DATA',
      'FILE',
      'REPORT',
      'WORKFLOW',
      'FUNCTION'
    )
  ),
  parent_resource_id UUID REFERENCES resources(id),
  path VARCHAR(500),  -- URL path, API endpoint, file path, etc.
  resource_attributes JSONB DEFAULT '{}'::jsonb,  -- ABAC resource attributes
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT resources_name_unique_per_module UNIQUE (tenant_id, module_id, name)
);

COMMENT ON TABLE resources IS 'System resources that can be protected by permissions including APIs, UI components, data objects, files, reports, and workflows.';

COMMENT ON COLUMN resources.resource_type IS 'Type of resource: API, UI, DATA, FILE, REPORT, WORKFLOW, FUNCTION';

COMMENT ON COLUMN resources.parent_resource_id IS 'Self-referential for resource hierarchy (e.g., API endpoints under API group)';

COMMENT ON COLUMN resources.path IS 'Resource path: URL, API endpoint, file path, database object, etc.';

COMMENT ON COLUMN resources.resource_attributes IS 'JSONB containing ABAC attributes like classification level, sensitivity, department ownership';

-- Enable RLS and create policies
ALTER TABLE
  resources ENABLE ROW LEVEL SECURITY;

CREATE POLICY resources_tenant_isolation ON resources FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY resources_admin_access ON resources FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);
