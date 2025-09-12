-- =====================================================================
-- FINANCE ACCOUNT GROUPS AND HEADERS
-- Enhances chart of accounts with grouping and classification structure
-- =====================================================================
-- Account groups/headers table for better organization
CREATE TABLE finance_account_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  -- Group identification
  group_code VARCHAR(20) NOT NULL,
  group_name VARCHAR(255) NOT NULL,
  group_description TEXT,
  -- Group hierarchy (groups can have parent groups)
  parent_group_id UUID REFERENCES finance_account_groups(id) ON DELETE RESTRICT,
  group_level INTEGER NOT NULL DEFAULT 1,
  group_path VARCHAR(500),  -- Materialized path for group hierarchy
  -- Classification alignment
  root_type VARCHAR(20) NOT NULL CHECK (
    root_type IN (
      'ASSET',
      'LIABILITY',
      'EQUITY',
      'REVENUE',
      'EXPENSE'
    )
  ),
  group_category VARCHAR(50),  -- CURRENT_ASSETS, FIXED_ASSETS, OPERATING_EXPENSES, etc.
  -- Financial statement presentation
  financial_statement_section VARCHAR(100),  -- Balance Sheet, Income Statement, Cash Flow
  statement_order INTEGER DEFAULT 999,
  show_in_summary BOOLEAN DEFAULT TRUE,
  consolidation_method VARCHAR(20) DEFAULT 'SUM' CHECK (
    consolidation_method IN ('SUM', 'AVERAGE', 'MAX', 'MIN', 'CUSTOM')
  ),
  -- Display and formatting
  display_format VARCHAR(50) DEFAULT 'STANDARD',  -- STANDARD, PERCENTAGE, CURRENCY, etc.
  indent_level INTEGER DEFAULT 0,
  show_totals BOOLEAN DEFAULT TRUE,
  bold_display BOOLEAN DEFAULT false,
  -- Operational settings
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  is_system_group BOOLEAN NOT NULL DEFAULT false,
  allow_direct_posting BOOLEAN DEFAULT false,  -- Usually false for headers
  -- Reporting and analysis
  budget_category VARCHAR(50),
  variance_analysis_group VARCHAR(50),
  cash_flow_category VARCHAR(50) CHECK (
    cash_flow_category IS NULL
    OR cash_flow_category IN ('OPERATING', 'INVESTING', 'FINANCING')
  ),
  -- Metadata
  group_attributes JSONB DEFAULT '{}'::jsonb,
  -- Standard timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  created_by UUID REFERENCES users(id),
  updated_by UUID REFERENCES users(id),
  -- Constraints
  CONSTRAINT chk_group_parent_not_self CHECK (id != parent_group_id),
  CONSTRAINT chk_group_level_depth CHECK (group_level BETWEEN 1 AND 5),
  -- Unique constraints
  UNIQUE (tenant_id, group_code),
  UNIQUE (tenant_id, group_name, parent_group_id)
);

COMMENT ON TABLE finance_account_groups IS 'Account groups and headers for organizing chart of accounts into logical reporting structures';

-- =====================================================================
-- ACCOUNT GROUP INDEXES
-- =====================================================================
-- Primary lookup indexes for groups
CREATE INDEX idx_account_groups_tenant_code ON finance_account_groups(tenant_id, group_code)
WHERE
  deleted_at IS NULL;

CREATE INDEX idx_account_groups_tenant_type ON finance_account_groups(tenant_id, root_type, group_category)
WHERE
  deleted_at IS NULL;

-- Group hierarchy indexes
CREATE INDEX idx_account_groups_parent ON finance_account_groups(parent_group_id)
WHERE
  deleted_at IS NULL;

CREATE INDEX idx_account_groups_hierarchy_level ON finance_account_groups(tenant_id, group_level, parent_group_id)
WHERE
  deleted_at IS NULL;

-- Financial statement indexes
CREATE INDEX idx_account_groups_statement ON finance_account_groups(
  tenant_id,
  financial_statement_section,
  statement_order
)
WHERE
  deleted_at IS NULL
  AND is_active = TRUE;

-- =====================================================================
-- VIEWS WITH GROUPING
-- =====================================================================
-- =====================================================================
-- RLS AND PERMISSIONS FOR NEW TABLES
-- =====================================================================
-- Enable RLS on account groups
ALTER TABLE
  finance_account_groups ENABLE ROW LEVEL SECURITY;

-- RLS policy for tenant isolation
CREATE POLICY tenant_isolation_policy ON finance_account_groups FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

-- Admin bypass policy
CREATE POLICY admin_full_access_policy ON finance_account_groups FOR ALL TO admin_role USING (TRUE);

-- Grant permissions
GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_groups TO application_role;

GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_groups TO admin_role;

-- Grant permissions on views
-- =====================================================================
-- EXAMPLE USAGE AND BENEFITS
-- =====================================================================
/*
BENEFITS OF ACCOUNT GROUPS/HEADERS:

1. FINANCIAL STATEMENT ORGANIZATION:
 - Clean separation of Current vs Fixed Assets
 - Operating vs Administrative Expenses  
 - Proper Income Statement vs Balance Sheet classification

2. REPORTING:
 - Group-level subtotals automatically calculated
 - Hierarchical financial statements
 - Variance analysis by account group
 - Budget vs Actual by category

3. BETTER USER EXPERIENCE:
 - Logical account organization in dropdowns
 - Collapsible account trees in UI
 - Context-aware account suggestions

4. COMPLIANCE AND STANDARDS:
 - Aligns with GAAP/IFRS presentation requirements
 - Industry-standard account groupings
 - Audit-friendly organization

EXAMPLE USAGE:

-- Get all cash and cash equivalent accounts
SELECT * FROM v_chart_of_accounts_complete 
WHERE group_category = 'CURRENT_ASSETS' 
AND account_type IN ('CASH', 'BANK');

-- Build balance sheet with proper grouping
SELECT 
 statement_section,
 header_name,
 group_name,
 SUM(group_balance) as section_total
FROM v_financial_statement_builder
WHERE statement_section = 'Balance Sheet'
GROUP BY statement_section, header_order, header_name, group_name
ORDER BY header_order;

-- Get all operating expense accounts for budget analysis
SELECT a.* FROM v_finance_accounts_with_groups a
WHERE a.group_category = 'OPERATING_EXPENSES'
AND a.is_active = true;
*/
