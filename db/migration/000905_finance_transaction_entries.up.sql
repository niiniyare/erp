-- ------------------------------------------------------------------------------------------------
-- FINANCE_TRANSACTION_ENTRIES TABLE
-- ------------------------------------------------------------------------------------------------
-- Individual journal entry lines that make up financial transactions. Implements double-entry
-- bookkeeping — each transaction must have at least one debit and one credit line, and the
-- totals must balance (enforced by trigger in 000909).
-- Each row carries either a debit_amount > 0 or a credit_amount > 0 (never both).
--
-- NOTE: Depends on finance_transactions (000904) and finance_accounts (000902).
--       RLS is driven by tenant_id denormalized from the parent transaction for query efficiency.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE finance_transaction_entries (
  id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id                   UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id                   UUID          REFERENCES entities(uuid) ON DELETE CASCADE,
  -- Transaction relationship
  transaction_id              UUID          NOT NULL REFERENCES finance_transactions(id) ON DELETE CASCADE,
  entry_number                INTEGER       NOT NULL,      -- Sequential line number within the transaction
  -- Account relationship
  account_id                  UUID          NOT NULL REFERENCES finance_accounts(id) ON DELETE RESTRICT,
  -- Entry amounts — DECIMAL(19,4) supports KWD/IQD/OMR (3 dp) and high-value transactions
  debit_amount                DECIMAL(19,4) NOT NULL DEFAULT 0.0000, -- Must be 0 if credit_amount > 0
  credit_amount               DECIMAL(19,4) NOT NULL DEFAULT 0.0000, -- Must be 0 if debit_amount > 0
  -- Entry details
  description                 TEXT          NOT NULL,
  reference                   VARCHAR(100),
  -- Dimensional analysis
  cost_center                 VARCHAR(20),
  cost_center_id              UUID,         -- Preferred FK reference; cost_center kept for compat
  department                  VARCHAR(50),
  project_id                  UUID,         -- TODO: add FK REFERENCES finance_projects(id) when projects table exists
  -- Multi-currency support
  original_currency           CHAR(3),
  original_amount             DECIMAL(19,4),               -- Amount in original currency before conversion
  exchange_rate               DECIMAL(18,8),
  -- Tax information
  tax_code                    VARCHAR(20),
  tax_rate                    DECIMAL(5,4),  -- 4dp supports rates like 7.5000%
  tax_amount                  DECIMAL(19,4),
  -- Reconciliation
  reconciled                  BOOLEAN       DEFAULT false,
  reconciled_date             DATE,
  reconciliation_reference    VARCHAR(100),
  -- Standard timestamps
  created_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  deleted_at                  TIMESTAMPTZ,
  -- Constraints
  CHECK (debit_amount >= 0 AND credit_amount >= 0),
  CHECK (NOT (debit_amount > 0 AND credit_amount > 0)),
  CHECK (debit_amount > 0 OR credit_amount > 0),
  -- Unique entry number per transaction
  UNIQUE (transaction_id, entry_number)
);

COMMENT ON TABLE  finance_transaction_entries IS 'Individual journal entries that make up financial transactions. Implements double-entry bookkeeping with debit and credit amounts.';
COMMENT ON COLUMN finance_transaction_entries.entry_number    IS 'Sequential entry number within transaction — used for ordering and reference.';
COMMENT ON COLUMN finance_transaction_entries.debit_amount    IS 'Debit amount in functional currency — must be 0 if credit_amount > 0.';
COMMENT ON COLUMN finance_transaction_entries.credit_amount   IS 'Credit amount in functional currency — must be 0 if debit_amount > 0.';
COMMENT ON COLUMN finance_transaction_entries.original_amount IS 'Original transaction amount in original currency before conversion.';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE finance_transaction_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE finance_transaction_entries FORCE  ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON finance_transaction_entries
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON finance_transaction_entries
  FOR ALL TO admin_role
  USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY finance_transaction_entries_ro_select ON finance_transaction_entries
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transaction_entries TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transaction_entries TO admin_role;
