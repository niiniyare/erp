-- =====================================================================
-- ACCOUNT TABLE
-- =====================================================================
-- This table stores individual accounts within a chart of accounts.
-- =====================================================================

CREATE TABLE account (
  id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  path VARCHAR(255) NOT NULL UNIQUE,
  depth INTEGER NOT NULL CHECK (depth >= 0),
  numchild INTEGER NOT NULL CHECK (numchild >= 0),
  account_code VARCHAR(10) NOT NULL,
  account_name VARCHAR(100) NOT NULL,
  account_type VARCHAR(20) NOT NULL
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
  account_role VARCHAR(30) ,
  balance_type VARCHAR(6) NOT NULL,
  locked BOOLEAN NOT NULL,
  active BOOLEAN NOT NULL,
  coa__id UUID NOT NULL REFERENCES chartofaccount (id) DEFERRABLE INITIALLY DEFERRED,
  role_default BOOLEAN NULL,
  CONSTRAINT unique_code_for_coa_ UNIQUE (coa__id, account_code),
  CONSTRAINT only_one_account_assigned_as_default_for_role UNIQUE (
    coa__id, account_role, role_default
  )
);

-- =====================================================================
-- INDEXES
-- =====================================================================
CREATE INDEX idx_account_tenant ON account(tenant_id);
CREATE INDEX idx_account_entity ON account(entity_id);
CREATE INDEX idx_account_code ON account(account_code);
CREATE INDEX idx_account_type ON account(account_type);

-- =====================================================================
-- RLS
-- =====================================================================
ALTER TABLE account ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON account
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON account
    FOR ALL TO admin_role
    USING (true);
