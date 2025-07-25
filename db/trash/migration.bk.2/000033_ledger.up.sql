-- =====================================================================
-- LEDGER TABLE
-- =====================================================================
-- This table stores ledgers, which are collections of journal entries.
-- =====================================================================

CREATE TABLE ledger (
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  posted_by UUID REFERENCES users(id),
  created_by UUID REFERENCES users(id),
  name VARCHAR(150) NULL,
  posted BOOLEAN NOT NULL,
  locked BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),
  ledger_xid VARCHAR(150) NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE ledger ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON ledger
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON ledger
    FOR ALL TO admin_role
    USING (true);
