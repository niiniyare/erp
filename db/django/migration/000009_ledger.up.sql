-- =====================================================
-- ACCOUNTING AND FINANCIAL TABLES
-- =====================================================

-- Ledgers - collections of journal entries for organizational purposes
CREATE TABLE IF NOT EXISTS ledger (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  -- entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE, 
  posted_by INT REFERENCES users(id),   
  created_by INT REFERENCES users(id),
  id SERIAL PRIMARY KEY,
  name VARCHAR(150) NULL,                              -- Ledger name/description
  posted BOOLEAN NOT NULL,                             -- Whether ledger is posted
  locked BOOLEAN NOT NULL,                             -- Whether ledger is locked
  hidden BOOLEAN NOT NULL,                             -- Whether ledger is hidden from UI
  entity_id INT NOT NULL REFERENCES entities(id) DEFERRABLE INITIALLY DEFERRED,
  additional_info TEXT NULL CHECK (
    (additional_info IS NULL OR additional_info::TEXT ~ '^[\s]*(\{.*\}|null)[\s]*$')
  ),
  ledger_xid VARCHAR(150) NULL                         -- External ledger identifier
);
COMMENT ON TABLE ledger IS 'Ledgers group related journal entries (e.g., monthly ledgers, project ledgers)';

