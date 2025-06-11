-- Root entity/company table with hierarchical structure and accounting preferences
CREATE TABLE entities (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_id INT REFERENCES entities(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50), -- For internal reference codes
    type VARCHAR(20) NOT NULL 
        CHECK (type IN ('COMPANY', 'SUBSIDIARY', 'REGION', 'BRANCH', 'LOCATION', 
                       'DEPARTMENT', 'DIVISION', 'COST_CENTER', 'PROJECT', 'BUDGET_UNIT')),
                  -- TODO ltree is not working for termuxfix later
    -- parent_id INT REFERENCES entities(id), -- For efficient hierarchy queries
    is_active BOOLEAN NOT NULL DEFAULT true,
    hidden BOOLEAN NOT NULL,                             -- Whether entity is hidden
    accrual_method BOOLEAN NOT NULL,                     -- True=Accrual, False=Cash accounting
    fy_start_month INTEGER NOT NULL,                     -- Fiscal year start month (1-12)
  
    address JSONB DEFAULT '{}'::jsonb, -- Entity-specific address
    picture VARCHAR(100) NULL,                           -- Logo/picture file path
    settings JSONB DEFAULT '{}'::jsonb, -- Entity-specific settings
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, parent_id, name)
);

CREATE UNIQUE INDEX tenant_code_unique_idx
ON entities (tenant_id, code)
WHERE code IS NOT NULL;

-- Enhanced closure table (keep your existing structure)
CREATE TABLE hierarchy_paths (
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ancestor_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    descendant_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    depth INT NOT NULL,
    PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

