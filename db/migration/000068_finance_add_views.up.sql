
-- VIEWS FOR COMMON QUERIES
-- Account hierarchy view with computed fields (Fixed type casting)
CREATE VIEW v_finance_accounts_hierarchy AS WITH RECURSIVE account_tree AS (
  -- Root accounts
  SELECT
    id,
    tenant_id,
    account_code,
    account_name,
    parent_account_id,
    root_type,
    account_type,
    normal_balance,
    current_balance,
    1 AS LEVEL,
    account_code::VARCHAR(500) AS full_path,
    account_name::VARCHAR(500) AS full_name
  FROM
    finance_accounts
  WHERE
    parent_account_id IS NULL
    AND deleted_at IS NULL
  UNION
  ALL
  -- Child accounts
  SELECT
    a.id,
    a.tenant_id,
    a.account_code,
    a.account_name,
    a.parent_account_id,
    a.root_type,
    a.account_type,
    a.normal_balance,
    a.current_balance,
    t.level + 1,
    (t.full_path || '.' || a.account_code)::VARCHAR(500),
    (t.full_name || ' > ' || a.account_name)::VARCHAR(500)
  FROM
    finance_accounts a
    JOIN account_tree t ON a.parent_account_id = t.id
  WHERE
    a.deleted_at IS NULL
),
account_tree_with_children AS (
  SELECT
    id,
    tenant_id,
    account_code,
    account_name,
    parent_account_id,
    root_type,
    account_type,
    normal_balance,
    current_balance,
    LEVEL,
    full_path,
    full_name
  FROM
    account_tree
)
SELECT
  at.id,
  at.tenant_id,
  at.account_code,
  at.account_name,
  at.parent_account_id,
  at.root_type,
  at.account_type,
  at.normal_balance,
  at.current_balance,
  at.level,
  at.full_path,
  at.full_name,
  COALESCE(child_counts.child_count, 0) AS child_count
FROM
  account_tree_with_children at
  LEFT JOIN (
    SELECT
      parent_account_id,
      COUNT(*) AS child_count
    FROM
      finance_accounts
    WHERE
      deleted_at IS NULL
    GROUP BY
      parent_account_id
  ) child_counts ON at.id = child_counts.parent_account_id;

COMMENT ON VIEW v_finance_accounts_hierarchy IS 'Hierarchical view of chart of accounts with computed paths and levels';

-- Transaction summary view
CREATE VIEW v_finance_transaction_summary AS
SELECT
  t.id,
  t.tenant_id,
  t.transaction_number,
  t.transaction_date,
  t.description,
  t.transaction_status,
  t.currency_code,
  t.total_debit_amount,
  t.total_credit_amount,
  COUNT(e.id) AS entry_count,
  COUNT(DISTINCT e.account_id) AS unique_accounts,
  BOOL_AND(e.reconciled) AS all_entries_reconciled,
  STRING_AGG(
    DISTINCT a.account_code,
    ', '
    ORDER BY
      a.account_code
  ) AS account_codes
FROM
  finance_transactions t
  LEFT JOIN finance_transaction_entries e ON t.id = e.transaction_id
  LEFT JOIN finance_accounts a ON e.account_id = a.id
WHERE
  t.deleted_at IS NULL
GROUP BY
  t.id,
  t.tenant_id,
  t.transaction_number,
  t.transaction_date,
  t.description,
  t.transaction_status,
  t.currency_code,
  t.total_debit_amount,
  t.total_credit_amount;

COMMENT ON VIEW v_finance_transaction_summary IS 'Summary view of transactions with entry counts and reconciliation status';

-- ===============================
-- PERFORMANCE MONITORING VIEWS
-- ===============================

-- Account activity monitoring
CREATE VIEW v_finance_account_activity AS
SELECT
  a.tenant_id,
  a.id AS account_id,
  a.account_code,
  a.account_name,
  a.current_balance,
  a.last_transaction_date,
  COUNT(e.id) AS total_entries,
  COUNT(
    CASE
      WHEN t.transaction_date >= CURRENT_DATE - INTERVAL '30 days' THEN 1
    END
  ) AS entries_last_30_days,
  SUM(
    CASE
      WHEN t.transaction_date >= CURRENT_DATE - INTERVAL '30 days' THEN e.debit_amount
      ELSE 0
    END
  ) AS debits_last_30_days,
  SUM(
    CASE
      WHEN t.transaction_date >= CURRENT_DATE - INTERVAL '30 days' THEN e.credit_amount
      ELSE 0
    END
  ) AS credits_last_30_days
FROM
  finance_accounts a
  LEFT JOIN finance_transaction_entries e ON a.id = e.account_id
  LEFT JOIN finance_transactions t ON e.transaction_id = t.id
  AND t.transaction_status = 'POSTED'
WHERE
  a.deleted_at IS NULL
GROUP BY
  a.tenant_id,
  a.id,
  a.account_code,
  a.account_name,
  a.current_balance,
  a.last_transaction_date;

COMMENT ON VIEW v_finance_account_activity IS 'Account activity summary for monitoring and analysis';

