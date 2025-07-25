-- =====================================================================
-- USER SESSIONS DOWN MIGRATION
-- =====================================================================

-- Drop functions
DROP FUNCTION IF EXISTS update_session_last_accessed();
DROP FUNCTION IF EXISTS touch_session(VARCHAR);
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Drop RLS policies
DROP POLICY IF EXISTS user_sessions_admin_access ON user_sessions;
DROP POLICY IF EXISTS user_sessions_tenant_isolation ON user_sessions;

-- Disable RLS
ALTER TABLE IF EXISTS user_sessions DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE IF EXISTS user_sessions DROP CONSTRAINT IF EXISTS valid_expiration_time;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_sessions_location_info_gin;
DROP INDEX IF EXISTS idx_user_sessions_device_info_gin;
DROP INDEX IF EXISTS idx_user_sessions_ip;
DROP INDEX IF EXISTS idx_user_sessions_last_accessed;
DROP INDEX IF EXISTS idx_user_sessions_expired;
DROP INDEX IF EXISTS idx_user_sessions_active;
DROP INDEX IF EXISTS idx_user_sessions_refresh_token;
DROP INDEX IF EXISTS idx_user_sessions_token;
DROP INDEX IF EXISTS idx_user_sessions_user;
DROP INDEX IF EXISTS idx_user_sessions_tenant;

-- Drop table
DROP TABLE IF EXISTS user_sessions;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'USER SESSIONS DOWN MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Successfully removed:';
    RAISE NOTICE '- 1 table (user_sessions)';
    RAISE NOTICE '- Session management functions';
    RAISE NOTICE '- All associated indexes and constraints';
    RAISE NOTICE '- All RLS policies and security settings';
    RAISE NOTICE '===================================================================';
END;
$$;