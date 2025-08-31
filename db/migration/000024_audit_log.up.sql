-- ------------------------------------------------------------------------------------------------
-- AUDIT LOG
-- ------------------------------------------------------------------------------------------------
-- Comprehensive audit logging with compliance flags and risk scoring.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  event_type VARCHAR(50) NOT NULL,
  event_category VARCHAR(50) DEFAULT 'ACCESS' CHECK (
    event_category IN (
      'ACCESS',
      'ADMIN',
      'DATA',
      'AUTH',
      'SYSTEM',
      'COMPLIANCE'
    )
  ),
  severity VARCHAR(20) DEFAULT 'INFO' CHECK (
    severity IN ('LOW', 'INFO', 'WARN', 'HIGH', 'CRITICAL')
  ),
  user_id UUID REFERENCES users(id),
  target_user_id UUID REFERENCES users(id),  -- For admin actions on other users
  entity_id UUID REFERENCES entities(uuid),
  resource_id UUID REFERENCES resources(id),
  action_id UUID REFERENCES actions(id),
  role_id UUID REFERENCES roles(id),
  permission_id UUID REFERENCES permissions(id),
  decision VARCHAR(20),  -- ALLOW/DENY for access attempts
  reason TEXT,  -- Human-readable reason
  risk_score INTEGER DEFAULT 0,  -- Calculated risk score (0-100)
  context JSONB DEFAULT '{}'::jsonb,  -- Additional event context
  ip_address INET,
  user_agent TEXT,
  session_id UUID REFERENCES user_sessions(id),
  compliance_flags JSONB DEFAULT '{}'::jsonb,  -- GDPR, SOX, HIPAA, etc.
  created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE audit_log IS 'Comprehensive audit log with compliance tracking, risk scoring, and detailed context for security monitoring and regulatory compliance.';

COMMENT ON COLUMN audit_log.event_category IS 'Event category: ACCESS (authorization), ADMIN (administrative), DATA (data access), AUTH (authentication), SYSTEM (system events), COMPLIANCE (regulatory)';

COMMENT ON COLUMN audit_log.severity IS 'Event severity level: LOW, INFO, WARN, HIGH, CRITICAL';

COMMENT ON COLUMN audit_log.target_user_id IS 'Target user for administrative actions (e.g., admin modifying another user)';

COMMENT ON COLUMN audit_log.risk_score IS 'Calculated risk score from 0-100 based on action, context, and user behavior';

COMMENT ON COLUMN audit_log.compliance_flags IS 'JSONB containing compliance-related flags (GDPR, SOX, HIPAA, PCI, etc.)';

-- Enable RLS and create policies
ALTER TABLE
  audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_log_tenant_isolation ON audit_log FOR ALL TO public USING (
  tenant_id = current_setting('app.current_tenant_id')::UUID
);
