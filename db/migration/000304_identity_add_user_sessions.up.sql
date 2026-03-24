-- ------------------------------------------------------------------------------------------------
-- USER_SESSIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- Tracks active user sessions with security context for ABAC evaluation and security monitoring.
-- Includes device fingerprinting, location data, and access patterns.
--
-- NOTE: Depends on tenants(id) and users(id). RLS relies on current_tenant_id() defined in an
--       earlier migration. set_tenant_context() guard is tightened in 000306.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE user_sessions (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_token    VARCHAR(255) UNIQUE NOT NULL,
  refresh_token    VARCHAR(255),
  ip_address       INET,
  user_agent       TEXT,
  device_info      JSONB       DEFAULT '{}'::jsonb,   -- Device fingerprinting data
  location_info    JSONB       DEFAULT '{}'::jsonb,   -- Geographic/network location for ABAC
  expires_at       TIMESTAMPTZ NOT NULL,
  risk_score       INT         DEFAULT 0,
  -- IAM session pre-computation columns (populated at login, read-only thereafter)
  user_type        VARCHAR(20),                        -- copied from users.user_type for pool routing without a DB join
  configuration    JSONB       NOT NULL DEFAULT '{"flags":{},"settings":{},"prefs":{}}'::jsonb,
  entity_scope     JSONB       NOT NULL DEFAULT '{"type":"entity"}'::jsonb,
  created_at       TIMESTAMPTZ DEFAULT NOW(),
  last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
  is_active        BOOLEAN     DEFAULT TRUE
);

COMMENT ON TABLE user_sessions IS 'Active user sessions with security context including device, location, and access patterns for ABAC evaluation and security monitoring.';

COMMENT ON COLUMN user_sessions.id              IS 'UUID primary key for the session record';
COMMENT ON COLUMN user_sessions.tenant_id       IS 'Foreign key to tenants table for multi-tenant isolation';
COMMENT ON COLUMN user_sessions.user_id         IS 'Foreign key to users table identifying the session owner';
COMMENT ON COLUMN user_sessions.session_token   IS 'Unique session token for authentication';
COMMENT ON COLUMN user_sessions.refresh_token   IS 'Token used for session renewal';
COMMENT ON COLUMN user_sessions.ip_address      IS 'IP address of the client';
COMMENT ON COLUMN user_sessions.user_agent      IS 'Browser/client user agent string';
COMMENT ON COLUMN user_sessions.device_info     IS 'JSONB containing device fingerprinting data for security analysis';
COMMENT ON COLUMN user_sessions.location_info   IS 'JSONB containing geographic and network location data for location-based access control';
COMMENT ON COLUMN user_sessions.risk_score      IS 'Calculated risk score (0-100) based on action, context, and user behavior';
COMMENT ON COLUMN user_sessions.expires_at      IS 'Session expiration timestamp';
COMMENT ON COLUMN user_sessions.created_at      IS 'Session creation timestamp';
COMMENT ON COLUMN user_sessions.last_accessed_at IS 'Last activity timestamp for session timeout tracking';
COMMENT ON COLUMN user_sessions.is_active       IS 'Whether the session is currently active';
COMMENT ON COLUMN user_sessions.user_type       IS 'Copied from users.user_type at login. Used by SetDBPool middleware to select the right connection pool without a DB roundtrip.';
COMMENT ON COLUMN user_sessions.configuration   IS 'Pre-computed session config: {"flags":{"finance":true,...},"settings":{"finance.approval_threshold":"100000",...},"prefs":{"finance.entry_mode":"spreadsheet",...}}. Built at login from 5 concurrent queries. Read-only — refresh requires re-login.';
COMMENT ON COLUMN user_sessions.entity_scope    IS 'Pre-computed entity access scope: {"type":"all"|"subtree"|"entity","entity_id":"uuid","path_prefix":"/root/parent/"}. Built at login. Used by repos for subtree WHERE clauses without extra joins.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_user_sessions_tenant         ON user_sessions(tenant_id);                    -- tenant-based queries
CREATE INDEX idx_user_sessions_user           ON user_sessions(user_id);                      -- user association
CREATE UNIQUE INDEX idx_user_sessions_token   ON user_sessions(session_token);               -- primary authentication lookup
CREATE INDEX idx_user_sessions_refresh_token  ON user_sessions(refresh_token)
    WHERE refresh_token IS NOT NULL;                                                           -- refresh token lookups
CREATE INDEX idx_user_sessions_active         ON user_sessions(is_active, expires_at)
    WHERE is_active = TRUE;                                                                    -- most common active-session query
CREATE INDEX idx_user_sessions_expired        ON user_sessions(expires_at);                   -- expired sessions cleanup
CREATE INDEX idx_user_sessions_last_accessed  ON user_sessions(last_accessed_at);             -- session timeout tracking
CREATE INDEX idx_user_sessions_ip             ON user_sessions(ip_address)
    WHERE ip_address IS NOT NULL;                                                              -- IP-based security queries
CREATE INDEX idx_user_sessions_device_info_gin    ON user_sessions USING gin(device_info);   -- JSONB device data search
CREATE INDEX idx_user_sessions_location_info_gin  ON user_sessions USING gin(location_info); -- JSONB location data search
CREATE INDEX idx_user_sessions_user_type      ON user_sessions(user_type)
    WHERE user_type IS NOT NULL;                                                               -- SetDBPool middleware pool routing

-- ------------------------------------------------------------------------------------------------
-- CONSTRAINTS
-- ------------------------------------------------------------------------------------------------
ALTER TABLE user_sessions
  ADD CONSTRAINT valid_expiration_time CHECK (expires_at > created_at); -- expires_at must be after creation

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_sessions_tenant_isolation ON user_sessions FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL
        AND tenant_id = current_tenant_id()
    );

CREATE POLICY user_sessions_admin_access ON user_sessions FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON user_sessions TO application_role;
