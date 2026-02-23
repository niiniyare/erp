-- ------------------------------------------------------------------------------------------------
-- RESOURCES TABLE
-- ------------------------------------------------------------------------------------------------
-- Defines system resources that can be protected by permissions (APIs, UI components, data, etc.).
--
-- NOTE: updated_at trigger is added in 000062_platform_iam_constraints.up.sql
--       (after update_updated_at_column() exists in 000057).
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS resources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
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
  is_active  BOOLEAN     DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE resources IS 'System resources that can be protected by permissions including APIs, UI components, data objects, files, reports, and workflows.';
COMMENT ON COLUMN resources.resource_type IS 'Type of resource: API, UI, DATA, FILE, REPORT, WORKFLOW, FUNCTION';
COMMENT ON COLUMN resources.parent_resource_id IS 'Self-referential for resource hierarchy (e.g., API endpoints under API group)';
COMMENT ON COLUMN resources.resource_attributes IS 'JSONB containing ABAC attributes like classification level, sensitivity, department ownership';
COMMENT ON COLUMN resources.path IS
  'Resource path convention: use dot-notation for logical resources (e.g. finance.invoice.create) '
  'and slash-notation for HTTP endpoints (e.g. /api/v1/invoices). Be consistent within a module. '
  'Consuming services must agree on the convention they parse.';
COMMENT ON COLUMN resources.updated_at IS 'Updated by trigger on every row change — use for cache invalidation.';
