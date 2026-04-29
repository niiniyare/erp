-- ------------------------------------------------------------------------------------------------
-- FINANCE_ACCOUNT_GROUPS TABLE
-- ------------------------------------------------------------------------------------------------
-- Account groups and headers for organizing the chart of accounts into logical reporting
-- structures. Supports multi-level hierarchies (up to 5 levels) and financial statement mapping.
-- root_type IN ('ASSET','LIABILITY','EQUITY','REVENUE','EXPENSE').
-- consolidation_method IN ('SUM','AVERAGE','MAX','MIN','CUSTOM').
-- cash_flow_category IN ('OPERATING','INVESTING','FINANCING') or NULL.
--
-- NOTE: finance_accounts (000902) references this table via account_group_id and
--       account_header_id. The hierarchy path trigger (maintain_group_hierarchy_path) is
--       created in 000906_finance_mod.up.sql.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE finance_account_groups (
  id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id                   UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id                   UUID          REFERENCES entities(uuid) ON DELETE CASCADE,
  -- Group identification
  group_code                  VARCHAR(20)   NOT NULL,
  group_name                  VARCHAR(255)  NOT NULL,
  group_description           TEXT,
  -- Group hierarchy (groups can have parent groups)
  parent_group_id             UUID          REFERENCES finance_account_groups(id) ON DELETE RESTRICT,
  group_level                 INTEGER       NOT NULL DEFAULT 1,
  group_path                  VARCHAR(500),             -- Materialized path for group hierarchy
  -- Classification alignment
  root_type                   VARCHAR(20)   NOT NULL CHECK (
    root_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')
  ),
  group_category              VARCHAR(50),              -- CURRENT_ASSETS, FIXED_ASSETS, OPERATING_EXPENSES, etc.
  -- Financial statement presentation
  financial_statement_section VARCHAR(100),             -- Balance Sheet, Income Statement, Cash Flow
  statement_order             INTEGER       DEFAULT 999,
  show_in_summary             BOOLEAN       DEFAULT TRUE,
  consolidation_method        VARCHAR(20)   DEFAULT 'SUM' CHECK (
    consolidation_method IN ('SUM', 'AVERAGE', 'MAX', 'MIN', 'CUSTOM')
  ),
  -- Display and formatting
  display_format              VARCHAR(50)   DEFAULT 'STANDARD', -- STANDARD, PERCENTAGE, CURRENCY, etc.
  indent_level                INTEGER       DEFAULT 0,
  show_totals                 BOOLEAN       DEFAULT TRUE,
  bold_display                BOOLEAN       DEFAULT false,
  -- Operational settings
  is_active                   BOOLEAN       NOT NULL DEFAULT TRUE,
  is_system_group             BOOLEAN       NOT NULL DEFAULT false,
  allow_direct_posting        BOOLEAN       DEFAULT false,      -- Usually false for headers
  -- Reporting and analysis
  budget_category             VARCHAR(50),
  variance_analysis_group     VARCHAR(50),
  cash_flow_category          VARCHAR(50)   CHECK (
    cash_flow_category IS NULL
    OR cash_flow_category IN ('OPERATING', 'INVESTING', 'FINANCING')
  ),
  -- Metadata
  group_attributes            JSONB         DEFAULT '{}'::jsonb,
  -- Standard timestamps
  created_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  deleted_at                  TIMESTAMPTZ,
  created_by                  UUID          NOT NULL REFERENCES users(id),
  updated_by                  UUID          REFERENCES users(id),
  -- Constraints
  CONSTRAINT chk_group_parent_not_self CHECK (id != parent_group_id),
  CONSTRAINT chk_group_level_depth     CHECK (group_level BETWEEN 1 AND 5),
  -- Unique constraints
  UNIQUE (tenant_id, group_code),
  UNIQUE (tenant_id, group_name, parent_group_id)
);

COMMENT ON TABLE  finance_account_groups                            IS 'Account groups and headers for organizing chart of accounts into logical reporting structures.';
COMMENT ON COLUMN finance_account_groups.group_path                IS 'Materialized path for efficient hierarchy queries — maintained by trigger in 000906.';
COMMENT ON COLUMN finance_account_groups.root_type                 IS 'High-level classification: ASSET, LIABILITY, EQUITY, REVENUE, or EXPENSE.';
COMMENT ON COLUMN finance_account_groups.financial_statement_section IS 'Target financial statement: Balance Sheet, Income Statement, or Cash Flow.';
COMMENT ON COLUMN finance_account_groups.consolidation_method      IS 'How child balances are rolled up: SUM (default), AVERAGE, MAX, MIN, or CUSTOM.';
COMMENT ON COLUMN finance_account_groups.is_system_group           IS 'TRUE for platform-seeded groups that cannot be deleted by tenants.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_account_groups_tenant_code ON finance_account_groups(tenant_id, group_code)
  WHERE deleted_at IS NULL; -- Primary lookup by code within tenant

CREATE INDEX idx_account_groups_tenant_type ON finance_account_groups(tenant_id, root_type, group_category)
  WHERE deleted_at IS NULL; -- Filtering groups by classification

CREATE INDEX idx_account_groups_parent ON finance_account_groups(parent_group_id)
  WHERE deleted_at IS NULL; -- Hierarchy traversal — find children of a group

CREATE INDEX idx_account_groups_hierarchy_level ON finance_account_groups(tenant_id, group_level, parent_group_id)
  WHERE deleted_at IS NULL; -- Level-by-level hierarchy expansion

CREATE INDEX idx_account_groups_statement ON finance_account_groups(
  tenant_id,
  financial_statement_section,
  statement_order
)
  WHERE deleted_at IS NULL
    AND is_active = TRUE; -- Financial statement ordering queries

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE finance_account_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_account_groups FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON finance_account_groups
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON finance_account_groups
  FOR ALL TO admin_role
  USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY finance_account_groups_ro_select ON finance_account_groups
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_account_groups TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_account_groups TO admin_role;
