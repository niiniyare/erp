-- ================================================================================================
-- USER SESSIONS TABLE - Tracks active user sessions with security context
-- ================================================================================================
--
-- Extracted from existing user migration to maintain consistency with implemented application code.
-- Tracks active user sessions with security context for ABAC evaluation.
-- Includes device fingerprinting, location data, and access patterns for security monitoring.
--
-- Prerequisites:
-- - tenants table with UUID primary key
-- - users table with UUID primary key
-- ================================================================================================
CREATE TABLE user_sessions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_token VARCHAR(255) UNIQUE NOT NULL,
  refresh_token VARCHAR(255),
  ip_address INET,
  user_agent TEXT,
  device_info JSONB DEFAULT '{}'::jsonb,  -- Device fingerprinting data
  location_info JSONB DEFAULT '{}'::jsonb,  -- Geographic/network location for ABAC
  expires_at TIMESTAMPTZ NOT NULL,
  risk_score INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
  is_active BOOLEAN DEFAULT TRUE
);

-- Add table and column comments
COMMENT ON TABLE user_sessions IS 'Active user sessions with security context including device, location, and access patterns for ABAC evaluation and security monitoring.';

COMMENT ON COLUMN user_sessions.id IS 'UUID primary key for the session record';

COMMENT ON COLUMN user_sessions.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';

COMMENT ON COLUMN user_sessions.user_id IS 'Foreign key to users table identifying the session owner';

COMMENT ON COLUMN user_sessions.session_token IS 'Unique session token for authentication';

COMMENT ON COLUMN user_sessions.refresh_token IS 'Token used for session renewal';

COMMENT ON COLUMN user_sessions.ip_address IS 'IP address of the client';

COMMENT ON COLUMN user_sessions.user_agent IS 'Browser/client user agent string';

COMMENT ON COLUMN user_sessions.device_info IS 'JSONB containing device fingerprinting data for security analysis';

COMMENT ON COLUMN user_sessions.location_info IS 'JSONB containing geographic and network location data for location-based access control';

COMMENT ON COLUMN user_sessions.risk_score IS 'Calculated risk score (0-100) based on action, context, and user behavior';

COMMENT ON COLUMN user_sessions.expires_at IS 'Session expiration timestamp';

COMMENT ON COLUMN user_sessions.created_at IS 'Session creation timestamp';

COMMENT ON COLUMN user_sessions.last_accessed_at IS 'Last activity timestamp for session timeout tracking';

COMMENT ON COLUMN user_sessions.is_active IS 'Whether the session is currently active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================
-- Index for tenant-based queries
CREATE INDEX idx_user_sessions_tenant ON user_sessions(tenant_id);

-- Index for user association
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);

-- Index for session token lookups (primary authentication query)
CREATE UNIQUE INDEX idx_user_sessions_token ON user_sessions(session_token);

-- Index for refresh token lookups
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token)
WHERE
  refresh_token IS NOT NULL;

-- Index for active sessions (most common query)
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active, expires_at)
WHERE
  is_active = TRUE;

-- Index for expired sessions cleanup
CREATE INDEX idx_user_sessions_expired ON user_sessions(expires_at);

-- Index for session timeout tracking
CREATE INDEX idx_user_sessions_last_accessed ON user_sessions(last_accessed_at);

-- Index for IP-based security queries
CREATE INDEX idx_user_sessions_ip ON user_sessions(ip_address)
WHERE
  ip_address IS NOT NULL;

-- GIN indexes for JSONB columns
CREATE INDEX idx_user_sessions_device_info_gin ON user_sessions USING gin(device_info);

CREATE INDEX idx_user_sessions_location_info_gin ON user_sessions USING gin(location_info);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================
-- Ensure expires_at is in the future for new sessions
ALTER TABLE
  user_sessions
ADD
  CONSTRAINT valid_expiration_time CHECK (expires_at > created_at);

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================
-- Enable RLS on user_sessions table
ALTER TABLE
  user_sessions ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY user_sessions_tenant_isolation ON user_sessions FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

-- Admin bypass policy
CREATE POLICY user_sessions_admin_access ON user_sessions FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

-- =====================================================================
-- PERMISSIONS AND GRANTS
-- =====================================================================
-- Grant necessary permissions to application role
GRANT
SELECT
,
INSERT
,
UPDATE
,
  DELETE ON user_sessions TO application_role;
