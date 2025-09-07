-- =====================================================================
-- ITEM_CATEGORIES TABLE
-- =====================================================================
-- This table stores item category information.
-- =====================================================================

CREATE TABLE item_categories (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE, 
    parent_id UUID REFERENCES item_categories(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, parent_id, name)
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE item_categories ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON item_categories
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON item_categories
    FOR ALL TO admin_role
    USING (true);
