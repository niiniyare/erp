-- ------------------------------------------------------------------------------------------------
-- FINANCE_TRANSACTIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Header table for all financial transactions. Stores transaction metadata, approval workflow,
-- and debit/credit summary amounts. Individual journal lines are stored in
-- finance_transaction_entries (000905).
-- transaction_type IN ('MANUAL','SYSTEM','IMPORTED','RECURRING','ADJUSTMENT','CLOSING').
-- transaction_status IN ('DRAFT','PENDING_APPROVAL','APPROVED','POSTED','CANCELLED','REVERSED').
-- approval_status IN ('NOT_REQUIRED','PENDING','APPROVED','REJECTED').
-- recurring_frequency IN ('DAILY','WEEKLY','MONTHLY','QUARTERLY','YEARLY') or NULL.
-- validation_status IN ('PENDING','VALID','WARNING','ERROR').
--
-- NOTE: Depends on tenants, entities, users, and finance_accounts (000902).
--       Balance-update and validation triggers are created in 000909_finance_triggers.up.sql.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE finance_transactions (
  id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id                   UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id                   UUID          REFERENCES entities(uuid) ON DELETE CASCADE,
  -- Transaction identification
  transaction_number          VARCHAR(50)   NOT NULL,
  transaction_type            VARCHAR(30)   NOT NULL CHECK (
    transaction_type IN (
      'MANUAL', 'SYSTEM', 'IMPORTED', 'RECURRING', 'ADJUSTMENT', 'CLOSING'
    )
  ),
  transaction_status          VARCHAR(20)   NOT NULL DEFAULT 'DRAFT' CHECK (
    transaction_status IN (
      'DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'POSTED', 'CANCELLED', 'REVERSED'
    )
  ),
  -- Transaction dates
  transaction_date            DATE          NOT NULL,                -- Business date for accounting purposes
  posting_date                DATE,                                  -- Required when status = POSTED
  due_date                    DATE,                                  -- For AP/AR and cash management
  -- Transaction details
  description                 TEXT          NOT NULL,
  reference_number            VARCHAR(100),                          -- Internal ref: invoice number, check number, etc.
  external_reference          VARCHAR(100),                          -- External ref: bank reference, vendor invoice number
  memo                        TEXT,
  -- Financial information
  currency_code               CHAR(3)       NOT NULL DEFAULT 'USD',
  exchange_rate               DECIMAL(18,8) DEFAULT 1.0 CHECK (exchange_rate > 0),
  total_debit_amount          DECIMAL(15,4) NOT NULL DEFAULT 0.00 CHECK (total_debit_amount >= 0),
  total_credit_amount         DECIMAL(15,4) NOT NULL DEFAULT 0.00 CHECK (total_credit_amount >= 0),
  -- Source and traceability
  source_module               VARCHAR(50),                           -- AP, AR, GL, PAYROLL, etc.
  source_document_type        VARCHAR(50),                           -- INVOICE, PAYMENT, JOURNAL_ENTRY, etc.
  source_document_id          UUID,
  batch_id                    UUID,                                  -- Groups related transactions (imports, bulk ops)
  -- Approval workflow
  approval_required           BOOLEAN       DEFAULT false,
  approval_status             VARCHAR(20)   DEFAULT 'NOT_REQUIRED' CHECK (
    approval_status IN ('NOT_REQUIRED', 'PENDING', 'APPROVED', 'REJECTED')
  ),
  approved_by                 UUID          REFERENCES users(id),
  approved_at                 TIMESTAMPTZ,
  approval_notes              TEXT,
  -- Recurring transaction
  is_recurring                BOOLEAN       DEFAULT false,
  recurring_frequency         VARCHAR(20)   CHECK (
    recurring_frequency IS NULL
    OR recurring_frequency IN ('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'YEARLY')
  ),
  next_recurring_date         DATE,
  -- Reversal tracking
  is_reversed                 BOOLEAN       DEFAULT false,
  reversed_by_transaction_id  UUID          REFERENCES finance_transactions(id),
  reversal_reason             TEXT,
  -- Audit and validation
  version                     INTEGER       NOT NULL DEFAULT 1 CHECK (version > 0), -- Optimistic locking
  validation_status           VARCHAR(20)   DEFAULT 'PENDING' CHECK (
    validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
  ),
  validation_errors           JSONB         DEFAULT '[]'::jsonb,
  -- Metadata and attachments
  transaction_attributes      JSONB         DEFAULT '{}'::jsonb,
  attachment_ids              TEXT[],
  tags                        VARCHAR(25)[],
  -- Standard timestamps and audit
  created_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  deleted_at                  TIMESTAMPTZ,
  created_by                  UUID          NOT NULL REFERENCES users(id),
  updated_by                  UUID          REFERENCES users(id),
  posted_by                   UUID          REFERENCES users(id),
  posted_at                   TIMESTAMPTZ,
  -- Business rule constraints
  CONSTRAINT balanced_transaction CHECK (
    CASE
      WHEN transaction_status IN ('POSTED', 'APPROVED')
        THEN total_debit_amount = total_credit_amount
      ELSE TRUE
    END
  ),
  CONSTRAINT posting_date_logic CHECK (
    CASE
      WHEN transaction_status = 'POSTED'
        THEN posting_date IS NOT NULL AND posted_by IS NOT NULL AND posted_at IS NOT NULL
      ELSE TRUE
    END
  ),
  CONSTRAINT approval_logic CHECK (
    CASE
      WHEN approval_required = TRUE AND transaction_status IN ('APPROVED', 'POSTED')
        THEN approved_by IS NOT NULL AND approved_at IS NOT NULL
      ELSE TRUE
    END
  ),
  CONSTRAINT recurring_logic CHECK (
    CASE
      WHEN is_recurring = TRUE THEN recurring_frequency IS NOT NULL
      ELSE recurring_frequency IS NULL
    END
  ),
  CONSTRAINT reversal_logic CHECK (
    CASE
      WHEN is_reversed = TRUE THEN reversed_by_transaction_id IS NOT NULL
      ELSE reversed_by_transaction_id IS NULL
    END
  ),
  -- Unique constraints
  UNIQUE (tenant_id, transaction_number)
);

COMMENT ON TABLE  finance_transactions IS 'Header table for all financial transactions. Contains transaction metadata, approval workflow, and summary amounts.';
COMMENT ON COLUMN finance_transactions.transaction_number IS 'Unique transaction number within tenant — auto-generated or user-provided.';
COMMENT ON COLUMN finance_transactions.transaction_type  IS 'Type of transaction — determines behavior and validation rules.';
COMMENT ON COLUMN finance_transactions.transaction_status IS 'Current status in transaction lifecycle — controls what operations are allowed.';
COMMENT ON COLUMN finance_transactions.transaction_date  IS 'Date when the transaction occurred — business date for accounting purposes.';
COMMENT ON COLUMN finance_transactions.posting_date      IS 'Date when transaction was posted to the general ledger — required when status is POSTED.';
COMMENT ON COLUMN finance_transactions.due_date          IS 'Due date for payment transactions — used for AP/AR and cash management.';
COMMENT ON COLUMN finance_transactions.currency_code     IS 'ISO 4217 currency code — defaults to USD but supports multi-currency.';
COMMENT ON COLUMN finance_transactions.exchange_rate     IS 'Exchange rate from transaction currency to functional currency — defaults to 1.0 for same currency.';
COMMENT ON COLUMN finance_transactions.total_debit_amount  IS 'Sum of all debit entries — must equal total_credit_amount for balanced transactions.';
COMMENT ON COLUMN finance_transactions.total_credit_amount IS 'Sum of all credit entries — must equal total_debit_amount for balanced transactions.';
COMMENT ON COLUMN finance_transactions.source_module       IS 'Source module that created this transaction: AP, AR, GL, PAYROLL, etc.';
COMMENT ON COLUMN finance_transactions.source_document_type IS 'Type of source document: INVOICE, PAYMENT, JOURNAL_ENTRY, etc.';
COMMENT ON COLUMN finance_transactions.source_document_id  IS 'ID of the source document that generated this transaction.';
COMMENT ON COLUMN finance_transactions.batch_id            IS 'Batch ID for grouping related transactions — useful for imports and bulk operations.';
COMMENT ON COLUMN finance_transactions.version             IS 'Version number for optimistic locking — prevents concurrent modifications.';
COMMENT ON COLUMN finance_transactions.validation_errors   IS 'JSON array of validation errors and warnings — helps with troubleshooting.';
COMMENT ON COLUMN finance_transactions.attachment_ids      IS 'Array of attachment/document IDs — links to supporting documents.';
COMMENT ON COLUMN finance_transactions.tags                IS 'Array of tags for categorization and filtering — max 25 chars each.';
COMMENT ON COLUMN finance_transactions.posted_at           IS 'Timestamp when transaction was posted — required when status is POSTED.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_finance_transactions_tenant_date   ON finance_transactions(tenant_id, transaction_date); -- Primary date range queries
CREATE INDEX idx_finance_transactions_tenant_number ON finance_transactions(tenant_id, transaction_number); -- Lookup by transaction number
CREATE INDEX idx_finance_transactions_tenant_status ON finance_transactions(tenant_id, transaction_status); -- Status-based filtering
CREATE INDEX idx_finance_transactions_entity_date   ON finance_transactions(entity_id, transaction_date)
  WHERE entity_id IS NOT NULL; -- Entity-scoped date range queries

CREATE INDEX idx_finance_transactions_approval ON finance_transactions(tenant_id, approval_status)
  WHERE approval_required = TRUE; -- Pending approval queue

CREATE INDEX idx_finance_transactions_validation ON finance_transactions(tenant_id, validation_status); -- Validation error review

CREATE INDEX idx_finance_transactions_source ON finance_transactions(source_document_type, source_document_id)
  WHERE source_document_id IS NOT NULL; -- Reverse-lookup from source documents

CREATE INDEX idx_finance_transactions_batch ON finance_transactions(batch_id)
  WHERE batch_id IS NOT NULL; -- Batch grouping queries

CREATE INDEX idx_finance_transactions_recurring ON finance_transactions(tenant_id, next_recurring_date)
  WHERE is_recurring = TRUE; -- Scheduled recurring transaction generation

CREATE INDEX idx_finance_transactions_created ON finance_transactions(tenant_id, created_at); -- Audit and reporting queries

CREATE INDEX idx_finance_transactions_posted ON finance_transactions(tenant_id, posted_at)
  WHERE posted_at IS NOT NULL; -- Posted-transaction reporting

CREATE INDEX idx_finance_transactions_tags ON finance_transactions USING GIN(tags)
  WHERE tags IS NOT NULL; -- Tag-based search (GIN index for array operations)

-- Additional index for group relationship (referenced in 000906)
CREATE INDEX idx_accounts_group_relationship ON finance_accounts(tenant_id, account_group_id, display_order)
  WHERE deleted_at IS NULL; -- Ordered account lists within a group

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON finance_transactions
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
    AND deleted_at IS NULL
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON finance_transactions
  FOR ALL TO admin_role
  USING (TRUE);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO admin_role;

-- ------------------------------------------------------------------------------------------------
-- TRIGGERS
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_finance_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_finance_transactions_updated_at
  BEFORE UPDATE ON finance_transactions
  FOR EACH ROW
  EXECUTE FUNCTION update_finance_transactions_updated_at();

-- ------------------------------------------------------------------------------------------------
-- UPDATE_FINANCE_TRANSACTION_VERSION (TRIGGER FUNCTION)
-- ------------------------------------------------------------------------------------------------
-- Increments the version counter on every UPDATE for optimistic locking support.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION increment_finance_transaction_version()
RETURNS TRIGGER AS $$
BEGIN
  NEW.version = OLD.version + 1;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_finance_transaction_version
  BEFORE UPDATE ON finance_transactions
  FOR EACH ROW
  EXECUTE FUNCTION increment_finance_transaction_version();
