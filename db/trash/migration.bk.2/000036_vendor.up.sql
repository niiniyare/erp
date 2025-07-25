-- =====================================================================
-- VENDOR TABLE
-- =====================================================================
-- This table stores vendor information.
-- =====================================================================

CREATE TABLE vendor (
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP NULL,
  uuid UUID NOT NULL PRIMARY KEY,
  vendor_name VARCHAR(100) NOT NULL,
  vendor_number VARCHAR(30) NULL,
  description TEXT NOT NULL,
  active BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  address JSONB DEFAULT '{}'::jsonb,
  contact JSONB DEFAULT '{}'::jsonb,
  account_number VARCHAR(30) NULL,
  routing_number VARCHAR(30) NULL,
  aba_number VARCHAR(30) NULL,
  swift_number VARCHAR(30) NULL,
  tax_id_number VARCHAR(30) NULL,
  account_type VARCHAR(20) NOT NULL,
  additional_info JSONB NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE vendor ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON vendor
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON vendor
    FOR ALL TO admin_role
    USING (true);
