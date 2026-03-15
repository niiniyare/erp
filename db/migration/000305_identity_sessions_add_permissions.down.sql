ALTER TABLE user_sessions
    DROP COLUMN IF EXISTS permissions,
    DROP COLUMN IF EXISTS principal_id;
