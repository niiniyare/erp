-- ------------------------------------------------------------------------------------------------
-- AUDIT_LOG
-- ------------------------------------------------------------------------------------------------
-- Immutable audit log with compliance flags and risk scoring for security monitoring and
-- regulatory compliance (GDPR, SOX, HIPAA, PCI-DSS).
-- event_category IN ('ACCESS','ADMIN','DATA','AUTH','SYSTEM','COMPLIANCE').
-- severity IN ('LOW','INFO','WARN','HIGH','CRITICAL').
--
-- NOTE: FKs reference tenants (000001), users (000303), entities (000010), resources, actions,
--       roles (000405), permissions, and user_sessions from platform-IAM migrations.
--       Audit functions and triggers are in 000451_audit_funcs.up.sql.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
  id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  event_type       VARCHAR(50)   NOT NULL,                          -- e.g. 'users_UPDATE', 'SESSION_TERMINATED'
  event_category   VARCHAR(50)   DEFAULT 'ACCESS' CHECK (
    event_category IN (
      'ACCESS',
      'ADMIN',
      'DATA',
      'AUTH',
      'SYSTEM',
      'COMPLIANCE'
    )
  ),
  severity         VARCHAR(20)   DEFAULT 'INFO' CHECK (
    severity IN ('LOW', 'INFO', 'WARN', 'HIGH', 'CRITICAL')
  ),
  user_id          UUID          REFERENCES users(id),
  target_user_id   UUID          REFERENCES users(id),              -- For admin actions performed on another user
  entity_id        UUID          REFERENCES entities(uuid),
  resource_id      UUID          REFERENCES resources(id),
  action_id        UUID          REFERENCES actions(id),
  role_id          UUID          REFERENCES roles(id),
  permission_id    UUID          REFERENCES permissions(id),
  decision         VARCHAR(20),                                      -- ALLOW or DENY for access attempts
  reason           TEXT,                                             -- Human-readable explanation
  risk_score       INTEGER       DEFAULT 0,                         -- Calculated risk score 0-100
  context          JSONB         DEFAULT '{}'::jsonb,               -- Additional structured event context
  ip_address       INET,
  user_agent       TEXT,
  session_id       UUID          REFERENCES user_sessions(id),
  compliance_flags JSONB         DEFAULT '{}'::jsonb,               -- GDPR, SOX, HIPAA, PCI-DSS flags
  created_at       TIMESTAMPTZ   DEFAULT NOW()
);

COMMENT ON TABLE audit_log IS 'Immutable audit log with compliance tracking, risk scoring, and detailed context for security monitoring and regulatory compliance.';

COMMENT ON COLUMN audit_log.event_category IS 'Event category: ACCESS (authorization), ADMIN (administrative), DATA (data access), AUTH (authentication), SYSTEM (system events), COMPLIANCE (regulatory)';
COMMENT ON COLUMN audit_log.severity IS 'Event severity level: LOW, INFO, WARN, HIGH, CRITICAL';
COMMENT ON COLUMN audit_log.target_user_id IS 'Target user for administrative actions (e.g., admin modifying another user)';
COMMENT ON COLUMN audit_log.risk_score IS 'Calculated risk score from 0-100 based on action, context, and user behavior';
COMMENT ON COLUMN audit_log.compliance_flags IS 'JSONB containing compliance-related flags (GDPR, SOX, HIPAA, PCI, etc.)';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE  ROW LEVEL SECURITY;

CREATE POLICY audit_log_tenant_isolation ON audit_log FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY audit_log_admin_access ON audit_log FOR ALL TO admin_role
    USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY audit_log_ro_select ON audit_log
    FOR SELECT TO readonly_role
    USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
