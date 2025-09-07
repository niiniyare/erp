-- Journal entries - the foundation of double-entry bookkeeping
CREATE TABLE IF NOT EXISTS journalentry (
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated TIMESTAMP WITHOUT TIME ZONE NULL,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) DEFERRABLE INITIALLY DEFERRED,
  posted_by UUID REFERENCES users(id),
  created_by UUID REFERENCES users(id),
  je_number VARCHAR(25) NOT NULL,
  timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  description VARCHAR(70) NULL,
  activity VARCHAR(20) NULL,
  origin VARCHAR(30) NULL,
  posted BOOLEAN NOT NULL,
  locked BOOLEAN NOT NULL,
  ledger_id UUID NOT NULL REFERENCES ledger(id) DEFERRABLE INITIALLY DEFERRED,
  is_closing_entry BOOLEAN NOT NULL
);
COMMENT ON TABLE journalentry IS 'Journal entries for double-entry bookkeeping with audit trail';
COMMENT ON COLUMN journalentry.posted IS 'Whether the journal entry affects account balances';
COMMENT ON COLUMN journalentry.origin IS 'Source system: invoice, bill, manual, etc.';

