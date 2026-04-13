-- Finance Cost Centres
-- Hierarchical cost centre structure with optional cost distribution rules.

CREATE TABLE IF NOT EXISTS finance_cost_centers (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID        NOT NULL,
    code              VARCHAR(50) NOT NULL,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    parent_id         UUID        REFERENCES finance_cost_centers(id) ON DELETE RESTRICT,
    is_group          BOOLEAN     NOT NULL DEFAULT FALSE,
    is_distributed    BOOLEAN     NOT NULL DEFAULT FALSE,
    allocation_method VARCHAR(30) CHECK (allocation_method IN ('PERCENTAGE','HEADCOUNT','SQUARE_FOOTAGE','ACTIVITY_BASED')),
    is_active         BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by        UUID,
    updated_by        UUID,

    CONSTRAINT uq_finance_cost_centers_tenant_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_cost_center_allocation
        CHECK (NOT is_distributed OR allocation_method IS NOT NULL)
);

CREATE INDEX idx_finance_cost_centers_tenant    ON finance_cost_centers(tenant_id);
CREATE INDEX idx_finance_cost_centers_parent    ON finance_cost_centers(parent_id);
CREATE INDEX idx_finance_cost_centers_active    ON finance_cost_centers(tenant_id, is_active);

-- Enable Row Level Security
ALTER TABLE finance_cost_centers ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_cost_centers_tenant_isolation
    ON finance_cost_centers
    USING (tenant_id = current_tenant_id());

-- Cost centre allocation rules (percentage or driver-based distribution)
CREATE TABLE IF NOT EXISTS finance_cost_center_allocations (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID         NOT NULL,
    source_center_id UUID         NOT NULL REFERENCES finance_cost_centers(id) ON DELETE CASCADE,
    target_center_id UUID         NOT NULL REFERENCES finance_cost_centers(id) ON DELETE CASCADE,
    method           VARCHAR(30)  NOT NULL CHECK (method IN ('PERCENTAGE','HEADCOUNT','SQUARE_FOOTAGE','ACTIVITY_BASED')),
    percentage       NUMERIC(7,4) CHECK (percentage IS NULL OR (percentage > 0 AND percentage <= 100)),
    driver_value     NUMERIC(18,4),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_cost_center_allocation UNIQUE (source_center_id, target_center_id),
    CONSTRAINT chk_allocation_self_ref CHECK (source_center_id <> target_center_id)
);

CREATE INDEX idx_finance_cc_allocations_source ON finance_cost_center_allocations(source_center_id);
CREATE INDEX idx_finance_cc_allocations_tenant ON finance_cost_center_allocations(tenant_id);

ALTER TABLE finance_cost_center_allocations ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_cc_allocations_tenant_isolation
    ON finance_cost_center_allocations
    USING (tenant_id = current_tenant_id());
