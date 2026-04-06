-- ------------------------------------------------------------------------------------------------
-- FINANCE_ACCOUNTS — PERFORMANCE INDEXES
-- ------------------------------------------------------------------------------------------------
-- Covers the primary access patterns for the chart of accounts: tenant-scoped lookups,
-- type filtering, hierarchy traversal, financial reporting, and multi-currency queries.
--
-- NOTE: Depends on finance_accounts created in 000902_finance_chart_of_accounts.up.sql.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_chart_of_accounts_tenant_code ON finance_accounts(tenant_id, account_code)
  WHERE deleted_at IS NULL; -- Most common query pattern: account lookup by code within tenant

COMMENT ON INDEX idx_chart_of_accounts_tenant_code IS 'Optimizes account lookup by code within tenant — most common query pattern.';

CREATE INDEX idx_chart_of_accounts_tenant_type ON finance_accounts(tenant_id, account_type, root_type)
  WHERE deleted_at IS NULL; -- Filtering accounts by type classification

CREATE INDEX idx_chart_of_accounts_parent ON finance_accounts(parent_account_id)
  WHERE deleted_at IS NULL; -- Hierarchy traversal — find children of an account

CREATE INDEX idx_chart_of_accounts_statement_line ON finance_accounts(
  tenant_id,
  financial_statement_line,
  report_order
)
  WHERE deleted_at IS NULL
    AND is_active = TRUE; -- Financial reporting ordered by statement line

CREATE INDEX idx_chart_of_accounts_balance_tracking ON finance_accounts(
  tenant_id,
  last_transaction_date DESC,
  current_balance
)
  WHERE deleted_at IS NULL
    AND current_balance != 0; -- Accounts with non-zero balance sorted by last activity

CREATE INDEX idx_chart_of_accounts_control ON finance_accounts(control_account_id)
  WHERE is_control_account = false
    AND deleted_at IS NULL; -- Find sub-accounts linked to a control account

CREATE INDEX idx_chart_of_accounts_currency ON finance_accounts(tenant_id, currency_code, is_multi_currency)
  WHERE deleted_at IS NULL; -- Multi-currency account queries

CREATE INDEX idx_chart_of_accounts_entity ON finance_accounts(entity_id)
  WHERE deleted_at IS NULL; -- Lookup accounts belonging to a specific entity

CREATE INDEX idx_chart_of_accounts_active ON finance_accounts(is_active)
  WHERE is_active = TRUE
    AND deleted_at IS NULL; -- Active-account-only queries

CREATE INDEX idx_chart_of_accounts_attributes_gin ON finance_accounts USING gin(account_attributes); -- JSON attribute search

CREATE INDEX idx_accounts_header_relationship ON finance_accounts(tenant_id, account_header_id, account_category)
  WHERE deleted_at IS NULL; -- Accounts grouped under a header for statement rendering

CREATE INDEX idx_accounts_cash_flow_type ON finance_accounts(tenant_id, cash_flow_type, account_type)
  WHERE cash_flow_type IS NOT NULL
    AND deleted_at IS NULL; -- Cash flow statement classification queries
