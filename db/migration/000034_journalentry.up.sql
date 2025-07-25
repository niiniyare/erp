-- =====================================================================
-- JOURNALENTRY TABLE
-- =====================================================================
-- This table stores journal entries, the foundation of double-entry bookkeeping.
-- =====================================================================

CREATE TABLE journalentry (
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  posted_by UUID REFERENCES users(id),
  created_by UUID REFERENCES users(id),
  je_number VARCHAR(25) NOT NULL,
  timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  description VARCHAR(70) NULL,
  activity VARCHAR(20) NULL,
  origin VARCHAR(30) NULL,
  posted BOOLEAN NOT NULL,
  locked BOOLEAN NOT NULL,
  ledger_id UUID NOT NULL REFERENCES ledger(id) DEFERRABLE INITIALLY DEFERRED,
  is_closing_entry BOOLEAN NOT NULL
);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE journalentry ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON journalentry
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON journalentry
    FOR ALL TO admin_role
    USING (true);
