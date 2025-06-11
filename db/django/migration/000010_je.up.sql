-- Journal entries - the foundation of double-entry bookkeeping
CREATE TABLE IF NOT EXISTS journalentry (
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  -- entity_id INT NOT NULL REFERENCES entities(id) ON DELETE CASCADE 
  posted_by INT REFERENCES users(id),   
  created_by INT REFERENCES users(id),
  uuid CHAR(32) NOT NULL PRIMARY KEY,
  je_number VARCHAR(25) NOT NULL,                      -- Journal entry number
  timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,     -- Transaction timestamp
  description VARCHAR(70) NULL,                        -- JE description
  activity VARCHAR(20) NULL,                           -- Business activity category
  origin VARCHAR(30) NULL,                             -- Source of the journal entry
  posted BOOLEAN NOT NULL,                             -- Whether JE is posted to the books
  locked BOOLEAN NOT NULL,                             -- Whether JE is locked from changes
  entity_unit_id INT NULL REFERENCES entities(id) DEFERRABLE INITIALLY DEFERRED,
  ledger_id INT NOT NULL REFERENCES ledger (id) DEFERRABLE INITIALLY DEFERRED,
  is_closing_entry BOOLEAN NOT NULL                    -- Whether this is a period-end closing entry
);
COMMENT ON TABLE journalentry IS 'Journal entries for double-entry bookkeeping with audit trail';
COMMENT ON COLUMN journalentry.posted IS 'Whether the journal entry affects account balances';
COMMENT ON COLUMN journalentry.origin IS 'Source system: invoice, bill, manual, etc.';


