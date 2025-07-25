-- =====================================================================
-- CUSTOMER TABLE
-- =====================================================================
-- This table stores customer information.
-- =====================================================================

CREATE TABLE customer (
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP NULL,
  id UUID NOT NULL PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  customer_name VARCHAR(100) NOT NULL,
  customer_number VARCHAR(30) NOT NULL,
  description TEXT NOT NULL,
  active BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,
  address JSONB DEFAULT '{}'::jsonb,
  email VARCHAR(254) NULL,
  website VARCHAR(200) NULL,
  phone VARCHAR(30) NULL,
  sales_tax_rate REAL NULL,
  additional_info JSONB NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE customer ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON customer
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON customer
    FOR ALL TO admin_role
    USING (true);
