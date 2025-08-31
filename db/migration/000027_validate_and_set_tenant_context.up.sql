-- Tenant context validation
CREATE
OR REPLACE FUNCTION validate_and_set_tenant_context(p_tenant_id UUID) RETURNS TABLE(tenant_name TEXT, tenant_status TEXT) AS
$$
DECLARE
v_tenant_record RECORD;

BEGIN
-- Validate and fetch tenant information
SELECT
  id,
  name,
  STATUS,
  deleted_at,
  last_activity_at INTO v_tenant_record
FROM
  tenants
WHERE
  id = p_tenant_id;

-- Check if tenant exists
IF v_tenant_record.id IS NULL THEN RAISE EXCEPTION 'Tenant not found: %',
p_tenant_id;

END IF;

-- Check if tenant is soft-deleted
IF v_tenant_record.deleted_at IS NOT NULL THEN RAISE EXCEPTION 'Tenant is deleted: %',
p_tenant_id;

END IF;

-- Check tenant status
IF v_tenant_record.status NOT IN ('active', 'pending') THEN RAISE EXCEPTION 'Tenant is not active: % (status: %)',
p_tenant_id,
v_tenant_record.status;

END IF;

-- Update last activity
UPDATE
  tenants
SET
  last_activity_at = NOW()
WHERE
  id = p_tenant_id;

-- Set tenant context
PERFORM set_config('app.current_tenant_id', p_tenant_id::text, TRUE);

-- Return tenant information
RETURN QUERY
SELECT
  v_tenant_record.name,
  v_tenant_record.status;

END;

$$
LANGUAGE plpgsql SECURITY DEFINER;

-- -- =====================================================================
-- -- PROJECTS TABLE
-- -- =====================================================================
-- -- This table stores project information, which can be associated with any entity.
-- -- =====================================================================
--
-- CREATE TABLE projects (
--     id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
--     name VARCHAR(255) NOT NULL,
--     code VARCHAR(50),
--     description TEXT,
--     project_manager_id UUID REFERENCES employees(id),
--     start_date DATE,
--     end_date DATE,
--     budget_amount DECIMAL(15,2),
--     actual_cost DECIMAL(15,2) DEFAULT 0,
--     status VARCHAR(20) DEFAULT 'PLANNING'
--         CHECK (status IN ('PLANNING', 'ACTIVE', 'ON_HOLD', 'COMPLETED', 'CANCELLED')),
--     metadata JSONB DEFAULT '{}'::jsonb,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     CONSTRAINT projects_tenant_code_entity_unique_idx UNIQUE (tenant_id, entity_id, code)
-- );
--
-- -- =====================================================================
-- -- INDEXES
-- -- =====================================================================
-- CREATE INDEX idx_projects_tenant ON projects(tenant_id);
-- CREATE INDEX idx_projects_entity ON projects(entity_id);
-- CREATE INDEX idx_projects_status ON projects(status);
--
-- -- =====================================================================
-- -- RLS
-- -- =====================================================================
-- ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
--
-- CREATE POLICY tenant_isolation_policy ON projects
--     FOR ALL TO application_role
--     USING (tenant_id = current_tenant_id())
--     WITH CHECK (tenant_id = current_tenant_id());
--
-- CREATE POLICY admin_full_access_policy ON projects
--     FOR ALL TO admin_role
--     USING (true);
--
-- -- =====================================================================
-- -- TRIGGERS
-- -- =====================================================================
-- CREATE TRIGGER update_projects_updated_at
--     BEFORE UPDATE ON projects
--     FOR EACH ROW
--     EXECUTE FUNCTION update_updated_at_column();
