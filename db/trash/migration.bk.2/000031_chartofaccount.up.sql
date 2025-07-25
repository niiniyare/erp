-- =====================================================================
-- CHARTOFACCOUNT TABLE
-- =====================================================================
-- This table stores chart of accounts templates.
-- =====================================================================

CREATE TABLE chartofaccount (
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  module TEXT,
  slug VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(150) NULL,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  is_active BOOLEAN DEFAULT true, 
  description TEXT NULL,
  active BOOLEAN NOT NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE chartofaccount ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON chartofaccount
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON chartofaccount
    FOR ALL TO admin_role
    USING (true);
