-- ------------------------------------------------------------------------------------------------
-- UNIVERSAL AUDIT TRIGGER SYSTEM
-- ------------------------------------------------------------------------------------------------
-- Comprehensive audit logging system that can be applied to any table.
-- Automatically logs INSERT, UPDATE, DELETE operations with intelligent context extraction,
-- risk scoring, and compliance flag detection.
--
-- NOTE: Depends on audit_log table (000450) and current_tenant_id() helper from 000062.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- GET_CURRENT_USER_CONTEXT
-- ------------------------------------------------------------------------------------------------
-- Reads user, session, IP, and user-agent from PostgreSQL session variables for audit records.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION get_current_user_context()
RETURNS JSONB AS $$
DECLARE
  user_context JSONB;
BEGIN
  user_context := jsonb_build_object(
    'user_id',    COALESCE(current_setting('app.current_user_id',    true), NULL)::UUID,
    'session_id', COALESCE(current_setting('app.current_session_id', true), NULL)::UUID,
    'ip_address', COALESCE(current_setting('app.client_ip',          true), NULL)::INET,
    'user_agent', COALESCE(current_setting('app.user_agent',         true), NULL)
  );
  RETURN user_context;
EXCEPTION
  WHEN OTHERS THEN
    RETURN '{}'::JSONB;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION get_current_user_context() IS 'Extracts current user context from session variables for audit logging';

-- ------------------------------------------------------------------------------------------------
-- CALCULATE_AUDIT_RISK_SCORE
-- ------------------------------------------------------------------------------------------------
-- Returns a 0-100 risk score based on operation type, table sensitivity, and field changes.
-- Base scores: DELETE=30, UPDATE=20, INSERT=10. Sensitive tables add 20; each sensitive
-- field changed adds 15. Score is capped at 100.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION calculate_audit_risk_score(
  p_operation  TEXT,
  p_table_name TEXT,
  p_old_data   JSONB,
  p_new_data   JSONB
)
RETURNS INTEGER AS $$
DECLARE
  risk_score       INTEGER := 0;
  sensitive_tables TEXT[]  := ARRAY['users', 'roles', 'permissions', 'tenants', 'finance_transactions'];
  sensitive_fields TEXT[]  := ARRAY['password', 'email', 'phone', 'ssn', 'credit_card', 'salary'];
  field_name       TEXT;
BEGIN
  -- Base risk by operation
  CASE p_operation
    WHEN 'DELETE' THEN risk_score := 30;
    WHEN 'UPDATE' THEN risk_score := 20;
    WHEN 'INSERT' THEN risk_score := 10;
  END CASE;

  -- Increase risk for sensitive tables
  IF p_table_name = ANY(sensitive_tables) THEN
    risk_score := risk_score + 20;
  END IF;

  -- Check for sensitive field modifications
  IF p_operation = 'UPDATE' THEN
    FOR field_name IN SELECT jsonb_object_keys(p_new_data) LOOP
      IF field_name = ANY(sensitive_fields) THEN
        IF (p_old_data->field_name) IS DISTINCT FROM (p_new_data->field_name) THEN
          risk_score := risk_score + 15;
        END IF;
      END IF;
    END LOOP;
  END IF;

  -- Cap risk score at 100
  RETURN LEAST(risk_score, 100);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION calculate_audit_risk_score IS 'Calculates risk score (0-100) based on operation type, table sensitivity, and field changes';

-- ------------------------------------------------------------------------------------------------
-- EXTRACT_COMPLIANCE_FLAGS
-- ------------------------------------------------------------------------------------------------
-- Detects applicable regulatory frameworks from the data being modified and returns a JSONB
-- flags object. Checks GDPR (PII fields), SOX (financial tables), HIPAA (health tables),
-- and PCI-DSS (card data fields).
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION extract_compliance_flags(
  p_table_name TEXT,
  p_old_data   JSONB,
  p_new_data   JSONB
)
RETURNS JSONB AS $$
DECLARE
  flags            JSONB  := '{}'::JSONB;
  pii_fields       TEXT[] := ARRAY['email', 'phone', 'address', 'ssn', 'date_of_birth'];
  financial_tables TEXT[] := ARRAY['finance_transactions', 'finance_accounts', 'payments'];
  health_tables    TEXT[] := ARRAY['patient_records', 'medical_data'];
BEGIN
  -- GDPR: Check for PII data access/modification
  IF EXISTS (
    SELECT 1 FROM jsonb_object_keys(COALESCE(p_new_data, p_old_data)) AS key
    WHERE key = ANY(pii_fields)
  ) THEN
    flags := flags || jsonb_build_object('GDPR', true);
  END IF;

  -- SOX: Check for financial data
  IF p_table_name = ANY(financial_tables) THEN
    flags := flags || jsonb_build_object('SOX', true);
  END IF;

  -- HIPAA: Check for health-related data
  IF p_table_name = ANY(health_tables) THEN
    flags := flags || jsonb_build_object('HIPAA', true);
  END IF;

  -- PCI-DSS: Check for payment card data
  IF p_new_data ? 'credit_card' OR p_old_data ? 'credit_card' OR
     p_new_data ? 'card_number' OR p_old_data ? 'card_number' THEN
    flags := flags || jsonb_build_object('PCI_DSS', true);
  END IF;

  RETURN flags;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION extract_compliance_flags IS 'Extracts applicable compliance flags (GDPR, SOX, HIPAA, PCI) based on data being accessed/modified';

-- ------------------------------------------------------------------------------------------------
-- DETERMINE_EVENT_CATEGORY
-- ------------------------------------------------------------------------------------------------
-- Maps table names to event categories used in the audit_log.event_category column.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION determine_event_category(p_table_name TEXT)
RETURNS VARCHAR(50) AS $$
BEGIN
  CASE
    WHEN p_table_name IN ('users', 'user_sessions', 'auth_tokens') THEN
      RETURN 'AUTH';
    WHEN p_table_name IN ('roles', 'permissions', 'policies', 'tenants') THEN
      RETURN 'ADMIN';
    WHEN p_table_name LIKE 'finance_%' THEN
      RETURN 'DATA';
    WHEN p_table_name IN ('audit_log', 'system_logs') THEN
      RETURN 'SYSTEM';
    WHEN p_table_name IN ('compliance_reports', 'audit_reports') THEN
      RETURN 'COMPLIANCE';
    ELSE
      RETURN 'ACCESS';
  END CASE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION determine_event_category IS 'Automatically determines event category based on table name';

-- ------------------------------------------------------------------------------------------------
-- DETERMINE_SEVERITY
-- ------------------------------------------------------------------------------------------------
-- Maps operation, table, and risk score to a severity level for the audit_log.severity column.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION determine_severity(
  p_operation  TEXT,
  p_table_name TEXT,
  p_risk_score INTEGER
)
RETURNS VARCHAR(20) AS $$
BEGIN
  IF p_risk_score >= 70 THEN
    RETURN 'CRITICAL';
  END IF;

  IF p_operation = 'DELETE' AND p_table_name IN ('users', 'tenants', 'finance_transactions') THEN
    RETURN 'HIGH';
  END IF;

  IF p_risk_score >= 50 THEN
    RETURN 'HIGH';
  END IF;

  IF p_risk_score >= 30 THEN
    RETURN 'WARN';
  END IF;

  IF p_risk_score >= 10 THEN
    RETURN 'INFO';
  END IF;

  RETURN 'LOW';
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION determine_severity IS 'Determines severity level based on operation, table, and risk score';

-- ------------------------------------------------------------------------------------------------
-- AUDIT_TRIGGER_FUNCTION
-- ------------------------------------------------------------------------------------------------
-- Universal AFTER trigger that logs every INSERT, UPDATE, DELETE into audit_log.
-- Tracks changed fields for UPDATE, extracts user context from session variables, calculates
-- risk score and compliance flags, and never fails the originating DML (exceptions are demoted
-- to warnings so the original operation always succeeds).
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
  v_tenant_id        UUID;
  v_user_context     JSONB;
  v_old_data         JSONB;
  v_new_data         JSONB;
  v_changed_fields   JSONB        := '{}'::JSONB;
  v_risk_score       INTEGER;
  v_compliance_flags JSONB;
  v_event_category   VARCHAR(50);
  v_severity         VARCHAR(20);
  v_context          JSONB        := '{}'::JSONB;
  v_field_name       TEXT;
BEGIN
  -- Extract tenant_id from the record
  v_tenant_id := COALESCE(
    CASE WHEN TG_OP = 'DELETE' THEN OLD.tenant_id ELSE NEW.tenant_id END,
    current_tenant_id()
  );

  -- Get user context from session
  v_user_context := get_current_user_context();

  -- Convert OLD and NEW to JSONB
  v_old_data := CASE WHEN TG_OP != 'INSERT' THEN to_jsonb(OLD) ELSE NULL END;
  v_new_data := CASE WHEN TG_OP != 'DELETE' THEN to_jsonb(NEW) ELSE NULL END;

  -- For UPDATE operations, track changed fields
  IF TG_OP = 'UPDATE' THEN
    FOR v_field_name IN SELECT jsonb_object_keys(v_new_data) LOOP
      IF (v_old_data->v_field_name) IS DISTINCT FROM (v_new_data->v_field_name) THEN
        v_changed_fields := v_changed_fields || jsonb_build_object(
          v_field_name,
          jsonb_build_object(
            'old', v_old_data->v_field_name,
            'new', v_new_data->v_field_name
          )
        );
      END IF;
    END LOOP;
  END IF;

  -- Calculate risk score
  v_risk_score := calculate_audit_risk_score(TG_OP, TG_TABLE_NAME, v_old_data, v_new_data);

  -- Extract compliance flags
  v_compliance_flags := extract_compliance_flags(TG_TABLE_NAME, v_old_data, v_new_data);

  -- Determine event category
  v_event_category := determine_event_category(TG_TABLE_NAME);

  -- Determine severity
  v_severity := determine_severity(TG_OP, TG_TABLE_NAME, v_risk_score);

  -- Build context object
  v_context := jsonb_build_object(
    'table_name',   TG_TABLE_NAME,
    'schema_name',  TG_TABLE_SCHEMA,
    'operation',    TG_OP,
    'changed_fields', v_changed_fields,
    'record_id', COALESCE(
      CASE WHEN TG_OP = 'DELETE' THEN OLD.id  ELSE NEW.id  END,
      CASE WHEN TG_OP = 'DELETE' THEN OLD.uuid ELSE NEW.uuid END
    ),
    'trigger_timing', TG_WHEN,
    'trigger_level',  TG_LEVEL
  );

  -- Add operation-specific payload to context
  IF TG_OP = 'UPDATE' THEN
    v_context := v_context || jsonb_build_object('changes', v_changed_fields);
  ELSIF TG_OP = 'INSERT' THEN
    v_context := v_context || jsonb_build_object('new_record', v_new_data);
  ELSIF TG_OP = 'DELETE' THEN
    v_context := v_context || jsonb_build_object('deleted_record', v_old_data);
  END IF;

  -- Insert audit log entry
  INSERT INTO audit_log (
    tenant_id,
    event_type,
    event_category,
    severity,
    user_id,
    decision,
    reason,
    risk_score,
    context,
    ip_address,
    user_agent,
    session_id,
    compliance_flags,
    created_at
  ) VALUES (
    v_tenant_id,
    format('%s_%s', TG_TABLE_NAME, TG_OP),  -- e.g. 'users_UPDATE'
    v_event_category,
    v_severity,
    (v_user_context->>'user_id')::UUID,
    'ALLOW',  -- Operation was allowed since trigger fired
    format('%s operation on %s table', TG_OP, TG_TABLE_NAME),
    v_risk_score,
    v_context,
    (v_user_context->>'ip_address')::INET,
    v_user_context->>'user_agent',
    (v_user_context->>'session_id')::UUID,
    v_compliance_flags,
    NOW()
  );

  -- Return appropriate value based on operation
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  ELSE
    RETURN NEW;
  END IF;

EXCEPTION
  WHEN OTHERS THEN
    -- Log the error but don't fail the original operation
    RAISE WARNING 'Audit logging failed for %.%: %', TG_TABLE_SCHEMA, TG_TABLE_NAME, SQLERRM;
    IF TG_OP = 'DELETE' THEN
      RETURN OLD;
    ELSE
      RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION audit_trigger_function() IS 'Universal audit trigger function that logs all table operations with intelligent risk scoring, compliance tracking, and detailed context';

-- ------------------------------------------------------------------------------------------------
-- ENABLE_AUDIT_ON_TABLE
-- ------------------------------------------------------------------------------------------------
-- Creates AFTER triggers for the specified operations on a single table.
-- Drops and recreates existing audit triggers to allow idempotent re-runs.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enable_audit_on_table(
  p_schema_name TEXT,
  p_table_name  TEXT,
  p_operations  TEXT[] DEFAULT ARRAY['INSERT', 'UPDATE', 'DELETE']
)
RETURNS TEXT AS $$
DECLARE
  v_trigger_name TEXT;
  v_operation    TEXT;
  v_result       TEXT := '';
BEGIN
  -- Validate table exists
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = p_schema_name
      AND table_name   = p_table_name
  ) THEN
    RAISE EXCEPTION 'Table %.% does not exist', p_schema_name, p_table_name;
  END IF;

  -- Create trigger for each operation
  FOREACH v_operation IN ARRAY p_operations LOOP
    v_trigger_name := format('audit_%s_%s_trigger', p_table_name, lower(v_operation));

    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I.%I',
      v_trigger_name, p_schema_name, p_table_name);

    EXECUTE format('
      CREATE TRIGGER %I
      AFTER %s ON %I.%I
      FOR EACH ROW
      EXECUTE FUNCTION audit_trigger_function()',
      v_trigger_name, v_operation, p_schema_name, p_table_name
    );

    v_result := v_result || format('Created trigger %s for %s on %s.%s' || E'\n',
      v_trigger_name, v_operation, p_schema_name, p_table_name);
  END LOOP;

  RETURN v_result;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enable_audit_on_table IS 'Enables audit logging on a specific table for specified operations (INSERT, UPDATE, DELETE)';

-- ------------------------------------------------------------------------------------------------
-- DISABLE_AUDIT_ON_TABLE
-- ------------------------------------------------------------------------------------------------
-- Drops all audit_* triggers from the specified table.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION disable_audit_on_table(
  p_schema_name TEXT,
  p_table_name  TEXT
)
RETURNS TEXT AS $$
DECLARE
  v_trigger RECORD;
  v_result  TEXT := '';
BEGIN
  FOR v_trigger IN
    SELECT trigger_name
    FROM information_schema.triggers
    WHERE event_object_schema = p_schema_name
      AND event_object_table  = p_table_name
      AND trigger_name LIKE 'audit_%'
  LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I.%I',
      v_trigger.trigger_name, p_schema_name, p_table_name);

    v_result := v_result || format('Dropped trigger %s from %s.%s' || E'\n',
      v_trigger.trigger_name, p_schema_name, p_table_name);
  END LOOP;

  IF v_result = '' THEN
    v_result := format('No audit triggers found on %s.%s', p_schema_name, p_table_name);
  END IF;

  RETURN v_result;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION disable_audit_on_table IS 'Disables all audit logging on a specific table by dropping audit triggers';

-- ------------------------------------------------------------------------------------------------
-- ENABLE_AUDIT_ON_SCHEMA
-- ------------------------------------------------------------------------------------------------
-- Enables audit logging on every table in the schema that has a tenant_id column,
-- excluding audit_log itself and schema_migrations.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enable_audit_on_schema(
  p_schema_name    TEXT    DEFAULT 'public',
  p_exclude_tables TEXT[]  DEFAULT ARRAY['audit_log', 'schema_migrations']
)
RETURNS TEXT AS $$
DECLARE
  v_table  RECORD;
  v_result TEXT := '';
BEGIN
  FOR v_table IN
    SELECT table_name
    FROM information_schema.tables
    WHERE table_schema = p_schema_name
      AND table_type   = 'BASE TABLE'
      AND table_name  != ALL(p_exclude_tables)
      -- Only tables with tenant_id support multi-tenant audit
      AND EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = p_schema_name
          AND table_name   = v_table.table_name
          AND column_name  = 'tenant_id'
      )
  LOOP
    v_result := v_result || enable_audit_on_table(p_schema_name, v_table.table_name);
  END LOOP;

  RETURN v_result;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enable_audit_on_schema IS 'Enables audit logging on all tables in a schema (with tenant_id column), excluding specified tables';

-- ------------------------------------------------------------------------------------------------
-- GET_AUDIT_STATISTICS
-- ------------------------------------------------------------------------------------------------
-- Returns per-category, per-operation counts with severity breakdown and average risk score
-- over the last p_hours_back hours. Pass NULL for p_tenant_id to query all tenants.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION get_audit_statistics(
  p_tenant_id  UUID    DEFAULT NULL,
  p_hours_back INTEGER DEFAULT 24
)
RETURNS TABLE (
  event_category VARCHAR(50),
  operation      TEXT,
  total_count    BIGINT,
  critical_count BIGINT,
  high_count     BIGINT,
  avg_risk_score NUMERIC,
  unique_users   BIGINT
) AS $$
BEGIN
  RETURN QUERY
  SELECT
    al.event_category,
    split_part(al.event_type, '_', 2) AS operation,
    COUNT(*)                                                    AS total_count,
    COUNT(*) FILTER (WHERE al.severity = 'CRITICAL')            AS critical_count,
    COUNT(*) FILTER (WHERE al.severity = 'HIGH')                AS high_count,
    ROUND(AVG(al.risk_score)::NUMERIC, 2)                       AS avg_risk_score,
    COUNT(DISTINCT al.user_id)                                  AS unique_users
  FROM audit_log al
  WHERE (p_tenant_id IS NULL OR al.tenant_id = p_tenant_id)
    AND al.created_at >= NOW() - (p_hours_back || ' hours')::INTERVAL
  GROUP BY al.event_category, split_part(al.event_type, '_', 2)
  ORDER BY total_count DESC;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_audit_statistics IS 'Returns audit statistics for a tenant over a specified time period';

-- ------------------------------------------------------------------------------------------------
-- EXAMPLE USAGE
-- ------------------------------------------------------------------------------------------------
-- NOTE: These are reference examples only — not executed during migration.

-- Enable audit on specific table:
-- SELECT enable_audit_on_table('public', 'users');

-- Enable audit on all eligible tables in the public schema:
-- SELECT enable_audit_on_schema('public');

-- Disable audit on a specific table:
-- SELECT disable_audit_on_table('public', 'users');

-- Get audit statistics for current tenant over last 24 hours:
-- SELECT * FROM get_audit_statistics(current_tenant_id(), 24);

-- Query recent high-risk events:
-- SELECT event_type, severity, risk_score, reason,
--        context->>'table_name' AS table_name, created_at
-- FROM audit_log
-- WHERE tenant_id = current_tenant_id()
--   AND risk_score >= 50
-- ORDER BY created_at DESC
-- LIMIT 100;
