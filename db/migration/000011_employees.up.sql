-- ================================================================================================
-- EMPLOYEES TABLE - Employment-specific data extending persons
-- ================================================================================================
--
-- Extends persons with employment-specific data and organizational hierarchy.
-- Contains role context, security levels, and employment status tracking.
--
-- Prerequisites:
-- - tenants table with UUID primary key
-- - entities table with UUID primary key
-- - persons table with UUID primary key
-- ================================================================================================
CREATE TABLE employees (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
  employee_number VARCHAR(50) NOT NULL,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
  position_title VARCHAR(100),
  department_id UUID REFERENCES entities(uuid),  -- Department entity reference
  manager_id UUID REFERENCES employees(id),  -- Self-referential for org hierarchy
  hire_date DATE NOT NULL,
  termination_date DATE,
  salary_info JSONB DEFAULT '{}'::jsonb,  -- Encrypted/sensitive salary data
  employment_status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (
    employment_status IN (
      'ACTIVE',
      'INACTIVE',
      'TERMINATED',
      'ON_LEAVE',
      'SUSPENDED'
    )
  ),
  work_schedule JSONB DEFAULT '{}'::jsonb,  -- Flexible work schedule definition
  security_level INTEGER DEFAULT 0,  -- Numeric security clearance level (0=lowest)
  access_attributes JSONB DEFAULT '{}'::jsonb,  -- Employment-specific ABAC attributes
  -- -- Standard validation columns
  -- version INTEGER NOT NULL DEFAULT 1,
  -- last_validation_run TIMESTAMPTZ,
  -- validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
  --     validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
  -- ),
  -- validation_errors JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  -- Ensure employee number uniqueness per tenant
  CONSTRAINT employees_number_unique_active EXCLUDE (tenant_id WITH =, employee_number WITH =)
  WHERE
    (deleted_at IS NULL)
);

-- Add table and column comments
COMMENT ON TABLE employees IS 'Employee records extending persons with employment-specific data, organizational hierarchy, and security levels for access control.';

COMMENT ON COLUMN employees.id IS 'UUID primary key for the employee record';

COMMENT ON COLUMN employees.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';

COMMENT ON COLUMN employees.person_id IS 'Foreign key to persons table linking to personal information';

COMMENT ON COLUMN employees.employee_number IS 'Unique employee identifier within tenant';

COMMENT ON COLUMN employees.entity_id IS 'Foreign key to entities table for organizational assignment';

COMMENT ON COLUMN employees.position_title IS 'Employee''s job title or position';

COMMENT ON COLUMN employees.department_id IS 'Foreign key to entities table representing department';

COMMENT ON COLUMN employees.manager_id IS 'Self-referential foreign key for organizational hierarchy';

COMMENT ON COLUMN employees.hire_date IS 'Date when employee was hired';

COMMENT ON COLUMN employees.termination_date IS 'Date when employee was terminated (if applicable)';

COMMENT ON COLUMN employees.salary_info IS 'JSONB containing encrypted/sensitive salary and compensation data';

COMMENT ON COLUMN employees.employment_status IS 'Current employment status: ACTIVE, INACTIVE, TERMINATED, ON_LEAVE, SUSPENDED';

COMMENT ON COLUMN employees.work_schedule IS 'JSONB containing flexible work schedule definition';

COMMENT ON COLUMN employees.security_level IS 'Numeric security clearance level (0=lowest, higher numbers = higher clearance)';

COMMENT ON COLUMN employees.access_attributes IS 'JSONB containing employment-specific ABAC attributes for access control';

COMMENT ON COLUMN employees.deleted_at IS 'Soft delete timestamp - NULL means record is active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================
-- Index for tenant-based queries
CREATE INDEX idx_employees_tenant ON employees(tenant_id);

-- Index for person association
CREATE INDEX idx_employees_person ON employees(person_id);

-- Index for entity/organization association
CREATE INDEX idx_employees_entity ON employees(entity_id);

-- Index for department association
CREATE INDEX idx_employees_department ON employees(department_id)
WHERE
  department_id IS NOT NULL;

-- Index for manager hierarchy
CREATE INDEX idx_employees_manager ON employees(manager_id)
WHERE
  manager_id IS NOT NULL;

-- Index for employment status filtering
CREATE INDEX idx_employees_status ON employees(employment_status);

-- Index for active employees (most common query)
CREATE INDEX idx_employees_active ON employees(employment_status)
WHERE
  employment_status = 'ACTIVE';

-- Index for employee number searches
CREATE INDEX idx_employees_number ON employees(employee_number);

-- Index for security level queries
CREATE INDEX idx_employees_security_level ON employees(security_level);

-- Index for soft delete queries
CREATE INDEX idx_employees_deleted_at ON employees(deleted_at)
WHERE
  deleted_at IS NOT NULL;

-- GIN indexes for JSONB columns
CREATE INDEX idx_employees_salary_info_gin ON employees USING gin(salary_info);

CREATE INDEX idx_employees_work_schedule_gin ON employees USING gin(work_schedule);

CREATE INDEX idx_employees_access_attributes_gin ON employees USING gin(access_attributes);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================
-- Ensure termination date is after hire date
ALTER TABLE
  employees
ADD
  CONSTRAINT valid_termination_date CHECK (
    termination_date IS NULL
    OR termination_date >= hire_date
  );

-- Ensure security level is non-negative
ALTER TABLE
  employees
ADD
  CONSTRAINT valid_security_level CHECK (security_level >= 0);

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================
-- Enable RLS on employees table
ALTER TABLE
  employees ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY employees_tenant_isolation ON employees FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND deleted_at IS NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND deleted_at IS NULL
  AND tenant_id = current_tenant_id()
);

-- Admin bypass policy
CREATE POLICY employees_admin_access ON employees FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

-- =====================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================================
-- Apply the existing update_updated_at_column trigger to employees
CREATE TRIGGER update_employees_updated_at BEFORE
UPDATE
  ON employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================
-- Grant necessary permissions to application role
GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON employees TO application_role;
