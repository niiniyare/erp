-- =====================================================================
-- BUDGETS TABLE
-- =====================================================================
-- This table stores budget information, which can be associated with an entity or a project.
-- =====================================================================

CREATE TABLE budgets (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id), -- Optional project budget
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
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_budgets_tenant ON budgets(tenant_id);
CREATE INDEX idx_budgets_entity ON budgets(entity_id);
CREATE INDEX idx_budgets_year ON budgets(fiscal_year);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON budgets
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON budgets
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- TRIGGERS
-- =====================================================================
CREATE TRIGGER update_budgets_updated_at
    BEFORE UPDATE ON budgets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
