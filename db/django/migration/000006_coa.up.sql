-- Module-specific tables can be added here
-- For example, for accounting module:
CREATE TABLE IF NOT EXISTS chartofaccount (
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  module TEXT,
  slug VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(150) NULL,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  is_active BOOLEAN DEFAULT true, 
  description TEXT NULL,
  entity_id CHAR(32) NOT NULL REFERENCES entity (uuid) DEFERRABLE INITIALLY DEFERRED,
  active BOOLEAN NOT NULL
);
COMMENT ON TABLE chartofaccount IS 'Chart of accounts templates (e.g., Standard, Manufacturing, Retail)';




-- Individual accounts within a chart of accounts
CREATE TABLE IF NOT EXISTS account (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  -- Tree structure for account hierarchy
  path VARCHAR(255) NOT NULL UNIQUE,
  depth INTEGER NOT NULL CHECK (depth >= 0),
  numchild INTEGER NOT NULL CHECK (numchild >= 0),
  uuid UUID NOT NULL PRIMARY KEY,
  account_code VARCHAR(10) NOT NULL,                           -- Account code (e.g., 1000, 4000)
  account_name VARCHAR(100) NOT NULL,                          -- Account name (e.g., Cash, Sales)

  account_type VARCHAR(20) NOT NULL
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),

  account_role VARCHAR(30) ,                           -- Account role (cash, ar, ap, revenue, etc.)
  balance_type VARCHAR(6) NOT NULL,                    -- DEBIT or CREDIT normal balance
  locked BOOLEAN NOT NULL,                             -- Whether account is locked from changes
  active BOOLEAN NOT NULL,                             -- Whether account is active
  coa__id CHAR(32) NOT NULL REFERENCES chartofaccount (uuid) DEFERRABLE INITIALLY DEFERRED,
  role_default BOOLEAN NULL,                          -- Whether this is the default account for this role
  CONSTRAINT unique_code_for_coa_ UNIQUE (coa__id, code),
  CONSTRAINT only_one_account_assigned_as_default_for_role UNIQUE (
    coa__id, role, role_default
  )
);
COMMENT ON TABLE account IS 'Individual accounts within chart of accounts with hierarchical structure';
COMMENT ON COLUMN account.role IS 'Account role: cash, ar, ap, inventory, revenue, expense, equity, etc.';
COMMENT ON COLUMN account.balance_type IS 'Normal balance type: DEBIT (assets, expenses) or CREDIT (liabilities, equity, revenue)';
