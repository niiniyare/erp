-- ================================================================================================
-- USERS TABLE - System access accounts with authentication and RBAC integration
-- ================================================================================================
--
-- System access accounts with authentication data and RBAC integration.
-- Can be linked to persons/employees or exist independently for service accounts.
--
-- Prerequisites:
-- - tenants table with UUID primary key
-- - entities table with UUID primary key
-- - persons table with UUID primary key
-- - employees table with UUID primary key
-- ================================================================================================
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_id UUID NOT NULL REFERENCES entities(uuid) ON DELETE RESTRICT,
  person_id UUID REFERENCES persons(id) ON DELETE
  SET
    NULL,
    employee_id UUID REFERENCES employees(id) ON DELETE
  SET
    NULL,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255),
    user_type VARCHAR(20) NOT NULL DEFAULT 'INTERNAL' CHECK (
      user_type IN (
        'INTERNAL',
        'CUSTOMER',
        'VENDOR',
        'PARTNER',
        'API',
        'SERVICE',
        'ADMIN'
      )
    ),
    account_status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (
      account_status IN (
        'ACTIVE',
        'INACTIVE',
        'LOCKED',
        'SUSPENDED',
        'PENDING_VERIFICATION'
      )
    ),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    failed_login_attempts INTEGER DEFAULT 0,
    lockout_until TIMESTAMPTZ,
    session_timeout_minutes INTEGER DEFAULT 480,  -- 8 hours default
    mfa_enabled BOOLEAN DEFAULT false,
    mfa_secret VARCHAR(255),
    user_attributes JSONB DEFAULT '{}'::jsonb,  -- ABAC user attributes
    settings JSONB DEFAULT '{}'::jsonb,  -- User preferences and settings
    password_strength INT DEFAULT 0,
    compromised BOOLEAN DEFAULT false,
    rotation_required BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    -- Ensure email uniqueness per tenant
    CONSTRAINT users_email_unique_active EXCLUDE (tenant_id WITH =, email WITH =)
  WHERE
    (deleted_at IS NULL),
    -- Ensure username uniqueness per tenant when provided
    CONSTRAINT users_username_unique_active EXCLUDE (tenant_id WITH =, username WITH =)
  WHERE
    (
      username IS NOT NULL
      AND deleted_at IS NULL
    )
);

-- Add table and column comments
COMMENT ON TABLE users IS 'System user accounts with authentication, authorization, and session management. Can be linked to persons/employees or exist independently for service accounts.';

COMMENT ON COLUMN users.id IS 'UUID primary key for the user record';

COMMENT ON COLUMN users.tenant_id IS 'Foreign key to tenants table for multi-tenant isolation';

COMMENT ON COLUMN users.entity_id IS 'Foreign key to entities table for organizational assignment';

COMMENT ON COLUMN users.person_id IS 'Optional foreign key to persons table (NULL for service accounts)';

COMMENT ON COLUMN users.employee_id IS 'Optional foreign key to employees table (NULL for non-employee users)';

COMMENT ON COLUMN users.username IS 'Unique username for login (optional, email can be used instead)';

COMMENT ON COLUMN users.email IS 'Email address for login and communication (must be unique per tenant)';

COMMENT ON COLUMN users.password_hash IS 'Hashed password for authentication';

COMMENT ON COLUMN users.user_type IS 'Classification of user account: INTERNAL, CUSTOMER, VENDOR, PARTNER, API, SERVICE, ADMIN';

COMMENT ON COLUMN users.account_status IS 'Current account status affecting login ability';

COMMENT ON COLUMN users.is_active IS 'Whether the user account is currently active';

COMMENT ON COLUMN users.last_login_at IS 'Timestamp of last successful login';

COMMENT ON COLUMN users.password_changed_at IS 'Timestamp of last password change';

COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for failed login attempts for security monitoring';

COMMENT ON COLUMN users.lockout_until IS 'Timestamp until which account is locked due to failed attempts';

COMMENT ON COLUMN users.session_timeout_minutes IS 'Session timeout in minutes (default 480 = 8 hours)';

COMMENT ON COLUMN users.mfa_enabled IS 'Whether multi-factor authentication is enabled';

COMMENT ON COLUMN users.mfa_secret IS 'Secret key for MFA token generation';

COMMENT ON COLUMN users.password_strength IS 'Password strength score (0-100) based on complexity';

COMMENT ON COLUMN users.compromised IS 'Flag if password found in breach databases';

COMMENT ON COLUMN users.rotation_required IS 'Forces password change on next login';

COMMENT ON COLUMN users.user_attributes IS 'JSONB containing ABAC attributes for fine-grained access control';

COMMENT ON COLUMN users.settings IS 'JSONB containing user preferences and application settings';

COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp - NULL means record is active';

-- =====================================================================
-- PERFORMANCE OPTIMIZATION INDEXES
-- =====================================================================
-- Index for tenant-based queries
CREATE INDEX idx_users_tenant ON users(tenant_id);

-- Index for entity association
CREATE INDEX idx_users_entity ON users(entity_id);

-- Index for person/employee associations
CREATE INDEX idx_users_person ON users(person_id)
WHERE
  person_id IS NOT NULL;

CREATE INDEX idx_users_employee ON users(employee_id)
WHERE
  employee_id IS NOT NULL;

-- Index for authentication queries
CREATE INDEX idx_users_email_lower ON users(lower(email));

CREATE INDEX idx_users_username_lower ON users(lower(username))
WHERE
  username IS NOT NULL;

-- Index for user type filtering
CREATE INDEX idx_users_type ON users(user_type);

-- Index for account status
CREATE INDEX idx_users_account_status ON users(account_status);

-- Index for active users (most common query)
CREATE INDEX idx_users_active ON users(is_active, account_status)
WHERE
  is_active = TRUE
  AND account_status = 'ACTIVE';

-- Index for security monitoring
CREATE INDEX idx_users_failed_attempts ON users(failed_login_attempts)
WHERE
  failed_login_attempts > 0;

CREATE INDEX idx_users_lockout ON users(lockout_until)
WHERE
  lockout_until IS NOT NULL;

-- Index for MFA users
CREATE INDEX idx_users_mfa ON users(mfa_enabled)
WHERE
  mfa_enabled = TRUE;

-- Index for soft delete queries
CREATE INDEX idx_users_deleted_at ON users(deleted_at)
WHERE
  deleted_at IS NOT NULL;

-- GIN indexes for JSONB columns
CREATE INDEX idx_users_attributes_gin ON users USING gin(user_attributes);

CREATE INDEX idx_users_settings_gin ON users USING gin(settings);

-- =====================================================================
-- DATA INTEGRITY CONSTRAINTS
-- =====================================================================
-- Ensure failed login attempts is non-negative
ALTER TABLE
  users
ADD
  CONSTRAINT valid_failed_attempts CHECK (failed_login_attempts >= 0);

-- Ensure session timeout is positive
ALTER TABLE
  users
ADD
  CONSTRAINT valid_session_timeout CHECK (session_timeout_minutes > 0);

-- Ensure lockout_until is in the future when set
ALTER TABLE
  users
ADD
  CONSTRAINT valid_lockout_time CHECK (
    lockout_until IS NULL
    OR lockout_until > NOW()
  );

-- =====================================================================
-- ROW LEVEL SECURITY (RLS)
-- =====================================================================
-- Enable RLS on users table
ALTER TABLE
  users ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY users_tenant_isolation ON users FOR ALL TO application_role USING (
  current_tenant_id() IS NOT NULL
  AND deleted_at IS NOT NULL
  AND tenant_id = current_tenant_id()
) WITH CHECK (
  current_tenant_id() IS NOT NULL
  AND tenant_id = current_tenant_id()
);

-- Admin bypass policy
CREATE POLICY users_admin_access ON users FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

-- =====================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================================
-- Apply the existing update_updated_at_column trigger to users
CREATE TRIGGER update_users_updated_at BEFORE
UPDATE
  ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

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
  DELETE ON users TO application_role;
