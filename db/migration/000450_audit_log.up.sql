-- ------------------------------------------------------------------------------------------------
-- AUDIT_LOG
-- ------------------------------------------------------------------------------------------------
-- Immutable, append-only audit log for security monitoring and regulatory compliance.
-- Designed to serve any ERP module. No domain-specific columns or assumptions.
--
-- event_category IN ('ACCESS','ADMIN','DATA','AUTH','SYSTEM','COMPLIANCE').
-- severity       IN ('LOW','INFO','WARN','HIGH','CRITICAL').
--
-- WHAT MAKES IT GENERIC:
--   All sensitivity, compliance, and categorisation rules live in audit_sensitive_tables and
--   audit_sensitive_fields (000449). The audit_log table itself stores outcomes only;
--   it has no opinion about which industry or jurisdiction produced them.
--
-- COMPLIANCE_FLAGS COLUMN:
--   A JSONB map of regulatory framework flags (e.g. {"GDPR": true, "SOX": true}).
--   Which flags appear is determined entirely by the configuration tables, not this schema.
--   Any compliance framework can be represented without a DDL change.
--
-- IMMUTABILITY:
--   Rows must never be updated or deleted outside a controlled retention process.
--   The admin_role policy is SELECT-only. The audit_retention_role is the only principal
--   permitted to delete rows (e.g. GDPR right-to-erasure, data TTL archival jobs).
--
-- DEPENDENCIES:
--   tenants (000001), users (000303), entities (000010), resources, actions,
--   roles (000405), permissions, user_sessions (platform-IAM migrations).
--   Audit configuration: 000449_audit_config.up.sql.
--   Audit functions and triggers: 000451_audit_funcs.up.sql.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
  id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

  -- Event identification
  event_type       VARCHAR(100)  NOT NULL,                          -- '<table>_<OP>' for DML triggers; named slug for manual events
  operation        VARCHAR(10)   GENERATED ALWAYS AS (             -- Extracted from event_type for reliable querying
    CASE
      WHEN event_type LIKE '%_INSERT' THEN 'INSERT'
      WHEN event_type LIKE '%_UPDATE' THEN 'UPDATE'
      WHEN event_type LIKE '%_DELETE' THEN 'DELETE'
      ELSE NULL
    END
  ) STORED,
  event_category   VARCHAR(50)   NOT NULL DEFAULT 'DATA' CHECK (
    event_category IN ('ACCESS','ADMIN','DATA','AUTH','SYSTEM','COMPLIANCE')
  ),
  severity         VARCHAR(20)   NOT NULL DEFAULT 'INFO' CHECK (
    severity IN ('LOW','INFO','WARN','HIGH','CRITICAL')
  ),

  -- Actor
  user_id          UUID          REFERENCES users(id),
  session_id       UUID          REFERENCES user_sessions(id),
  ip_address       INET,
  user_agent       TEXT,

  -- Subject (optional FK columns; populate whichever applies for the event)
  target_user_id   UUID          REFERENCES users(id),             -- Admin acting on another user
  entity_id        UUID          REFERENCES entities(uuid),
  resource_id      UUID          REFERENCES resources(id),
  action_id        UUID          REFERENCES actions(id),
  role_id          UUID          REFERENCES roles(id),
  permission_id    UUID          REFERENCES permissions(id),

  -- Decision and explanation
  decision         VARCHAR(20),                                      -- 'ALLOW' or 'DENY' for access-check events
  reason           TEXT,                                             -- Human-readable explanation of why the event occurred

  -- Risk and compliance
  risk_score       INTEGER       NOT NULL DEFAULT 0 CHECK (risk_score BETWEEN 0 AND 100),
  compliance_flags JSONB         NOT NULL DEFAULT '{}',             -- e.g. {"GDPR": true, "SOX": true}

  -- Full event payload
  context          JSONB         NOT NULL DEFAULT '{}',             -- table_name, schema_name, record_id, changes, etc.

  created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------------------------------------------
-- COMMENTS
-- ------------------------------------------------------------------------------------------------
COMMENT ON TABLE  audit_log IS
  'Immutable append-only audit log. Serves all ERP modules. Sensitivity and compliance rules '
  'are driven by audit_sensitive_tables and audit_sensitive_fields (000449) — no domain '
  'knowledge is embedded in this table or the trigger functions.';

COMMENT ON COLUMN audit_log.event_type       IS 'Composite slug: <table>_INSERT|UPDATE|DELETE for DML triggers; arbitrary slug for manual events (e.g. SESSION_TERMINATED, EXPORT_REQUESTED)';
COMMENT ON COLUMN audit_log.operation        IS 'Generated: INSERT, UPDATE, DELETE, or NULL for non-DML events';
COMMENT ON COLUMN audit_log.event_category   IS 'ACCESS=authz check, ADMIN=config/IAM change, DATA=record mutation, AUTH=authentication, SYSTEM=internal, COMPLIANCE=regulatory';
COMMENT ON COLUMN audit_log.target_user_id   IS 'Populated when an admin performs an action on or behalf of another user';
COMMENT ON COLUMN audit_log.decision         IS 'ALLOW or DENY — relevant for access-check events, NULL for trigger-generated DML events';
COMMENT ON COLUMN audit_log.risk_score       IS '0–100 composite score calculated from operation type, table sensitivity, and field-level changes';
COMMENT ON COLUMN audit_log.compliance_flags IS 'JSONB map of applicable regulatory flags. Which flags appear is determined by audit_sensitive_tables and audit_sensitive_fields — not hardcoded here';
COMMENT ON COLUMN audit_log.context          IS 'Structured event payload: table_name, schema_name, record_id, changed_fields, trigger metadata, and operation-specific nested payload';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------

-- Primary access pattern: tenant-scoped time-series queries
CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_created
  ON audit_log (tenant_id, created_at DESC);

-- Alert / SIEM pattern: high-signal events only
CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_severity
  ON audit_log (tenant_id, severity, created_at DESC)
  WHERE severity IN ('HIGH', 'CRITICAL');

-- User activity pattern: "what did this user do?"
CREATE INDEX IF NOT EXISTS idx_audit_log_user_created
  ON audit_log (user_id, created_at DESC)
  WHERE user_id IS NOT NULL;

-- Risk threshold queries: "show me everything above score X"
CREATE INDEX IF NOT EXISTS idx_audit_log_risk_score
  ON audit_log (tenant_id, risk_score DESC, created_at DESC)
  WHERE risk_score >= 30;

-- Compliance queries: GIN for containment operators (@>, ?)
CREATE INDEX IF NOT EXISTS idx_audit_log_compliance_flags
  ON audit_log USING GIN (compliance_flags)
  WHERE compliance_flags != '{}';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE  ROW LEVEL SECURITY;

-- Application role: INSERT within own tenant only.
-- Triggers write directly as the application role; no UPDATE or DELETE.
CREATE POLICY audit_log_tenant_insert ON audit_log
  FOR INSERT TO application_role
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- Application role: SELECT within own tenant.
CREATE POLICY audit_log_tenant_select ON audit_log
  FOR SELECT TO application_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- Admin role: cross-tenant SELECT only.
-- Admins must not be able to modify or delete audit rows from the application layer.
CREATE POLICY audit_log_admin_read ON audit_log
  FOR SELECT TO admin_role
  USING (TRUE);

-- Read-only role: scoped SELECT.
CREATE POLICY audit_log_ro_select ON audit_log
  FOR SELECT TO readonly_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());





-- Retention role: the only principal that may delete rows.
-- Assign only to the archival service account (GDPR erasure, TTL jobs).
-- Never assign to interactive user roles.
CREATE POLICY audit_log_retention_delete ON audit_log
  FOR DELETE TO audit_retention_role
  USING (TRUE);

COMMENT ON POLICY audit_log_retention_delete ON audit_log IS
  'Allows the audit_retention_role service account to delete any row from audit_log. '
  'This policy enables GDPR erasure requests, TTL-based cleanup jobs, and archival processes. '
  'The role itself is restricted to service accounts only - never interactive users.';


GRANT DELETE ON audit_log TO audit_retention_role;


