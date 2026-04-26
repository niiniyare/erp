-- ------------------------------------------------------------------------------------------------
-- FINANCE_ACCOUNTS TABLE
-- ------------------------------------------------------------------------------------------------
-- Master chart of accounts with hierarchical structure and multi-currency support. Every
-- financial transaction entry (000905) references a row in this table via account_id.
-- root_type IN ('ASSET','LIABILITY','EQUITY','REVENUE','EXPENSE').
-- normal_balance IN ('DEBIT','CREDIT') — DEBIT for assets/expenses, CREDIT for liabilities/equity/revenue.
-- validation_status IN ('PENDING','VALID','WARNING','ERROR').
-- cash_flow_type IN ('OPERATING','INVESTING','FINANCING') or NULL.
--
-- NOTE: Depends on finance_account_groups (000901) for account_group_id and account_header_id.
--       Hierarchy path is maintained by a trigger created in 000909_finance_triggers.up.sql.
--       Indexes are created in 000903_finance_coa_indexes.up.sql.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE finance_accounts (
  id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id                   UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id                   UUID          REFERENCES entities(uuid) ON DELETE CASCADE,
  -- Account identification
  account_code                VARCHAR(20)   NOT NULL,
  account_name                VARCHAR(255)  NOT NULL,
  account_description         TEXT,
  -- Account grouping
  account_group_id            UUID          REFERENCES finance_account_groups(id) ON DELETE SET NULL, -- Link to account group
  account_header_id           UUID          REFERENCES finance_account_groups(id) ON DELETE SET NULL, -- Link to account header
  -- Account hierarchy
  parent_account_id           UUID          REFERENCES finance_accounts(id) ON DELETE RESTRICT,
  account_level               INTEGER       NOT NULL DEFAULT 1,
  account_path                VARCHAR(500),             -- Materialized path for hierarchy queries
  account_category            VARCHAR(50),
  sub_category                VARCHAR(50),
  -- Display and reporting
  display_order               INTEGER       DEFAULT 999,
  show_in_reports             BOOLEAN       DEFAULT TRUE,
  consolidation_account       VARCHAR(50),
  -- Cash flow classification
  cash_flow_type              VARCHAR(20)   CHECK (
    cash_flow_type IS NULL
    OR cash_flow_type IN ('OPERATING', 'INVESTING', 'FINANCING')
  ),
  -- Account classification
  root_type                   VARCHAR(20)   NOT NULL CHECK (
    root_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')
  ),
  account_type                VARCHAR(50)   NOT NULL,
  account_subtype             VARCHAR(50),
  -- Financial attributes
  normal_balance              VARCHAR(10)   NOT NULL CHECK (normal_balance IN ('DEBIT', 'CREDIT')),
  is_control_account          BOOLEAN       NOT NULL DEFAULT false,
  has_children                BOOLEAN       NOT NULL DEFAULT false,
  control_account_id          UUID          REFERENCES finance_accounts(id),
  
  -- Currency and localization
  currency_code               CHAR(3)       DEFAULT 'USD',
  is_multi_currency           BOOLEAN       DEFAULT false,
  currency_revaluation_required BOOLEAN     DEFAULT false,
  -- Operational settings
  is_active                   BOOLEAN       NOT NULL DEFAULT TRUE,
  is_system_account           BOOLEAN       NOT NULL DEFAULT false,
  allow_manual_entries        BOOLEAN       NOT NULL DEFAULT TRUE,
  is_leaf                     BOOLEAN       GENERATED ALWAYS AS (NOT has_children) STORED,
  require_reference           BOOLEAN       NOT NULL DEFAULT false,
  -- Balance tracking
  current_balance             DECIMAL(15,2) DEFAULT 0.00,
  ytd_balance                 DECIMAL(15,2) DEFAULT 0.00,
  last_transaction_date       DATE,
  -- Reporting and analysis
  financial_statement_line    VARCHAR(100),
  report_order                INTEGER       DEFAULT 999,
  -- Budgeting
  is_budgetable               BOOLEAN       DEFAULT TRUE,
  budget_variance_threshold   DECIMAL(5,2)  DEFAULT 10.0,
  -- Audit and validation
  version                     INTEGER       NOT NULL DEFAULT 1,
  last_validation_run         TIMESTAMPTZ,
  validation_status           VARCHAR(20)   DEFAULT 'PENDING' CHECK (
    validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
  ),
  validation_errors           JSONB         DEFAULT '[]'::jsonb,
  -- Metadata and settings
  account_attributes          JSONB         DEFAULT '{}'::jsonb,
  -- Standard timestamps
  created_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  deleted_at                  TIMESTAMPTZ,
  created_by                  UUID          REFERENCES users(id),
  updated_by                  UUID          REFERENCES users(id),
  -- Unique constraints
  UNIQUE (tenant_id, account_code),
  UNIQUE (tenant_id, account_name, parent_account_id)
);

COMMENT ON TABLE  finance_accounts IS 'Master chart of accounts for all financial transactions. Supports hierarchical account structures, multi-currency operations, and financial reporting requirements.';
COMMENT ON COLUMN finance_accounts.account_code    IS 'Unique account code within tenant — used for transaction posting and reporting.';
COMMENT ON COLUMN finance_accounts.account_path    IS 'Materialized path for efficient hierarchy queries — format: /root/parent/child/.';
COMMENT ON COLUMN finance_accounts.root_type       IS 'High-level account classification for balance sheet and income statement categorization.';
COMMENT ON COLUMN finance_accounts.normal_balance  IS 'Normal balance type — DEBIT for assets/expenses, CREDIT for liabilities/equity/revenue.';
COMMENT ON COLUMN finance_accounts.account_group_id  IS 'Link to account group for organizational structure and reporting.';
COMMENT ON COLUMN finance_accounts.account_header_id IS 'Link to account header for financial statement presentation.';
COMMENT ON COLUMN finance_accounts.account_category  IS 'Detailed category within account type (e.g., CURRENT_ASSETS, FIXED_ASSETS).';
COMMENT ON COLUMN finance_accounts.cash_flow_type    IS 'Cash flow statement classification for proper statement presentation.';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE finance_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_accounts FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON finance_accounts
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON finance_accounts
  FOR ALL TO admin_role
  USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY finance_accounts_ro_select ON finance_accounts
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_accounts TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_accounts TO admin_role;
