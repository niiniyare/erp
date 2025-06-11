-- Projects table (can be under any entity)
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    parent_project_id INT REFERENCES projects(id), -- Sub-projects
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    description TEXT,
    project_manager_id INT REFERENCES employees(id),
    start_date DATE,
    end_date DATE,
    budget_amount DECIMAL(15,2),
    actual_cost DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'PLANNING'
        CHECK (status IN ('PLANNING', 'ACTIVE', 'ON_HOLD', 'COMPLETED', 'CANCELLED')),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX tenant_code_entity_unique_idx
ON projects (tenant_id, entity_id, code)
WHERE code IS NOT NULL;

-- Budget management
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    project_id INT REFERENCES projects(id), -- Optional project budget
    name VARCHAR(255) NOT NULL,
    budget_type VARCHAR(20) NOT NULL
        CHECK (budget_type IN ('OPERATIONAL', 'CAPITAL', 'PROJECT', 'DEPARTMENT')),
    fiscal_year INT NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    allocated_amount DECIMAL(15,2) DEFAULT 0,
    spent_amount DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'APPROVED', 'ACTIVE', 'CLOSED')),
    approved_by INT REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

