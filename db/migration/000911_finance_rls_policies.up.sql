-- +migrate Up
BEGIN
;

-- ADDITIONAL RLS POLICIES FOR NEW TABLES
-- Enable RLS on new tables
ALTER TABLE
  finance_account_balances ENABLE ROW LEVEL SECURITY;

ALTER TABLE
  finance_account_validation_rules ENABLE ROW LEVEL SECURITY;

-- RLS policies for new tables
CREATE POLICY tenant_isolation_policy ON finance_account_balances FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

CREATE POLICY tenant_isolation_policy ON finance_account_validation_rules FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

-- Admin policies for new tables
CREATE POLICY admin_full_access_policy ON finance_account_balances FOR ALL TO admin_role USING (TRUE);

CREATE POLICY admin_full_access_policy ON finance_account_validation_rules FOR ALL TO admin_role USING (TRUE);

-- Grant permissions on new tables
GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_balances TO application_role,
  admin_role;

GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON finance_account_validation_rules TO application_role,
  admin_role;

-- Grant permissions on views (already created in previous migration)
GRANT
SELECT
  ON v_finance_accounts_hierarchy TO application_role,
  admin_role;

GRANT
SELECT
  ON v_finance_transaction_summary TO application_role,
  admin_role;

GRANT
SELECT
  ON v_finance_account_activity TO application_role,
  admin_role;

COMMIT;
