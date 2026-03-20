DROP TABLE IF EXISTS api_keys;

ALTER TABLE user_sessions
    DROP COLUMN IF EXISTS permissions,
    DROP COLUMN IF EXISTS principal_id;
