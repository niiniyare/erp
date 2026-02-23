-- =============================================================================
-- FIX: Add user_type to users table for non-human identity support
-- =============================================================================
-- WHY: The identity model is Person → Employee → User, which works for
--      human users but has no representation for:
--        - Background workers (scheduled jobs, queue processors)
--        - API integration users (webhooks, third-party connectors)
--        - Service accounts (internal microservice calls)
--      These identities need to participate in the IAM model (role assignment,
--      permission checks, audit logging) but cannot be "persons" or "employees".
--
-- HOW: Add user_type column with HUMAN as default (preserves all existing rows).
--      Non-human users will have person_id = NULL (allow NULL on FK in users).
--      Application code should enforce: if user_type != 'HUMAN' then person_id IS NULL.
-- =============================================================================

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS user_type VARCHAR(20) NOT NULL DEFAULT 'HUMAN'
    CHECK (user_type IN ('HUMAN', 'SERVICE_ACCOUNT', 'API_KEY', 'INTEGRATION', 'SYSTEM'));

COMMENT ON COLUMN users.user_type IS
  'Identity classification. '
  'HUMAN = real person with a person_id FK. '
  'SERVICE_ACCOUNT = internal background worker or scheduled job. '
  'API_KEY = external system integration (person_id IS NULL). '
  'INTEGRATION = third-party connector (Stripe, Salesforce, etc.). '
  'SYSTEM = platform-level automated process.';

-- Partial index: finding service accounts is a common ops/security query
CREATE INDEX IF NOT EXISTS idx_users_non_human
  ON users(user_type, tenant_id)
  WHERE user_type <> 'HUMAN';

-- Verify all existing rows correctly got HUMAN default
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM users WHERE user_type IS NULL) THEN
    RAISE EXCEPTION 'Unexpected NULL user_type after migration — investigate.';
  END IF;
END;
$$;
