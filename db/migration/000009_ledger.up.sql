-- =====================================================
-- ACCOUNTING AND FINANCIAL TABLES
-- =====================================================

-- Ledgers - collections of journal entries for organizational purposes
CREATE TABLE IF NOT EXISTS ledger (
  id SERIAL PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,

  posted_by INT REFERENCES users(id),
  created_by INT REFERENCES users(id),

  name VARCHAR(150) NULL,
  posted BOOLEAN NOT NULL,
  locked BOOLEAN NOT NULL,
  hidden BOOLEAN NOT NULL,

  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),

  ledger_xid VARCHAR(150) NULL
);
COMMENT ON TABLE ledger IS 'Ledgers group related journal entries (e.g., monthly ledgers, project ledgers)';

