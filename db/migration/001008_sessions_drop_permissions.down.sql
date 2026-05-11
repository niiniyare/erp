-- ------------------------------------------------------------------------------------------------
-- USER_SESSIONS — RESTORE PERMISSIONS COLUMN (rollback of 001008)
-- ------------------------------------------------------------------------------------------------
-- Restores the column as nullable JSONB so existing rows are not affected.
-- Note: data that was in the column before the up migration was applied is lost.
-- ------------------------------------------------------------------------------------------------
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS permissions JSONB NULL;
