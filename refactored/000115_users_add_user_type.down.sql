-- =============================================================================
-- ROLLBACK: Remove user_type from users
-- =============================================================================
DROP INDEX IF EXISTS idx_users_non_human;

ALTER TABLE users
  DROP COLUMN IF EXISTS user_type;
