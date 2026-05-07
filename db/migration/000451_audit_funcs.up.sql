-- ------------------------------------------------------------------------------------------------
-- UNIVERSAL AUDIT TRIGGER SYSTEM — FUNCTIONS
-- ------------------------------------------------------------------------------------------------
-- All functions in this file are domain-agnostic. Sensitivity rules, compliance flags,
-- event categories, and delete severity are driven entirely by the configuration tables in
-- 000449_audit_config.up.sql. No industry, jurisdiction, or module is referenced here.
--
-- FUNCTION VOLATILITY:
--   calculate_audit_risk_score, extract_compliance_flags, determine_event_category, and
--   determine_severity are STABLE (not IMMUTABLE) because they read from the configuration
--   tables. PostgreSQL re-evaluates STABLE functions between statements but not within a
--   single statement, so repeated calls within the same trigger invocation are efficient.
--   The configuration tables are small and will remain hot in the shared buffer cache.
--
-- BACKWARD COMPATIBILITY:
--   All function signatures are unchanged from v1/v2. Existing call sites require no
--   modification. The only observable behavioral change is that rules previously hardcoded
--   in function bodies are now read from audit_sensitive_tables and audit_sensitive_fields.
--
-- DEPENDENCIES: 000449_audit_config.up.sql, 000450_audit_log.up.sql, 000062 (current_tenant_id).
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- GET_CURRENT_USER_CONTEXT
-- ------------------------------------------------------------------------------------------------
-- Reads user, session, IP, and user-agent from PostgreSQL session variables.
-- Each value is validated before casting. A malformed value emits a WARNING and yields NULL
-- for that field only — partial attribution is preserved rather than discarding the whole
-- context object.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION get_current_user_context()
RETURNS JSONB AS $$
DECLARE
  v_raw_user_id    TEXT;
  v_raw_session_id TEXT;
  v_raw_ip         TEXT;
  v_user_id        UUID;
  v_session_id     UUID;
  v_ip_address     INET;
  UUID_PATTERN     CONSTANT TEXT :=
    '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$';
BEGIN
  v_raw_user_id    := current_setting('app.current_user_id',    true);
  v_raw_session_id := current_setting('app.current_session_id', true);
  v_raw_ip         := current_setting('app.client_ip',          true);

  IF v_raw_user_id IS NOT NULL AND v_raw_user_id <> '' THEN
    IF v_raw_user_id ~ UUID_PATTERN THEN
      v_user_id := v_raw_user_id::UUID;
    ELSE
      RAISE WARNING 'get_current_user_context: invalid UUID in app.current_user_id: %', v_raw_user_id;
    END IF;
  END IF;

  IF v_raw_session_id IS NOT NULL AND v_raw_session_id <> '' THEN
    IF v_raw_session_id ~ UUID_PATTERN THEN
      v_session_id := v_raw_session_id::UUID;
    ELSE
      RAISE WARNING 'get_current_user_context: invalid UUID in app.current_session_id: %', v_raw_session_id;
    END IF;
  END IF;

  IF v_raw_ip IS NOT NULL AND v_raw_ip <> '' THEN
    BEGIN
      v_ip_address := v_raw_ip::INET;
    EXCEPTION WHEN OTHERS THEN
      RAISE WARNING 'get_current_user_context: invalid INET in app.client_ip: %', v_raw_ip;
    END;
  END IF;

  RETURN jsonb_build_object(
    'user_id',    v_user_id,
    'session_id', v_session_id,
    'ip_address', v_ip_address,
    'user_agent', NULLIF(current_setting('app.user_agent', true), '')
  );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION get_current_user_context() IS
  'Extracts current user context from PostgreSQL session variables for audit logging. '
  'Each field is validated individually — a bad value emits a WARNING and yields NULL '
  'for that field without discarding the rest of the context.';

-- ------------------------------------------------------------------------------------------------
-- CALCULATE_AUDIT_RISK_SCORE
-- ------------------------------------------------------------------------------------------------
-- Returns a 0–100 risk score.
--
-- Scoring model:
--   Base score by operation:  DELETE=30, UPDATE=20, INSERT=10
--   + risk_weight from audit_sensitive_tables  if the table is registered
--   + risk_weight from audit_sensitive_fields  for each sensitive field that changed (UPDATE)
--     or is present in the row (INSERT/DELETE)
--   Capped at 100.
--
-- Changed from IMMUTABLE → STABLE to enable config table reads.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION calculate_audit_risk_score(
  p_operation  TEXT,
  p_table_name TEXT,
  p_old_data   JSONB,
  p_new_data   JSONB
)
RETURNS INTEGER AS $$
DECLARE
  v_score         INTEGER := 0;
  v_table_weight  INTEGER;
  v_field         RECORD;
  v_combined_data JSONB;
BEGIN
  -- Base score by operation
  v_score := CASE p_operation
    WHEN 'DELETE' THEN 30
    WHEN 'UPDATE' THEN 20
    WHEN 'INSERT' THEN 10
    ELSE 5
  END;

  -- Table-level weight from configuration
  SELECT risk_weight INTO v_table_weight
  FROM audit_sensitive_tables
  WHERE table_name = p_table_name;

  IF FOUND THEN
    v_score := v_score + v_table_weight;
  END IF;

  -- Field-level weights from configuration
  -- For UPDATE: only fields that actually changed contribute.
  -- For INSERT/DELETE: any sensitive field present in the row contributes once.
  v_combined_data := COALESCE(p_new_data, '{}'::JSONB) || COALESCE(p_old_data, '{}'::JSONB);

  FOR v_field IN
    SELECT asf.field_name, asf.risk_weight
    FROM audit_sensitive_fields asf
    WHERE v_combined_data ? asf.field_name
  LOOP
    IF p_operation = 'UPDATE'
       AND p_old_data IS NOT NULL
       AND p_new_data IS NOT NULL
    THEN
      -- Only score the field if its value actually changed
      IF (p_old_data->v_field.field_name) IS DISTINCT FROM (p_new_data->v_field.field_name) THEN
        v_score := v_score + v_field.risk_weight;
      END IF;
    ELSE
      v_score := v_score + v_field.risk_weight;
    END IF;
  END LOOP;

  RETURN LEAST(v_score, 100);
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION calculate_audit_risk_score IS
  'Calculates a 0–100 risk score from operation type, table weight (audit_sensitive_tables), '
  'and per-field weights (audit_sensitive_fields). No domain knowledge is embedded here — '
  'all weights are configuration data.';

-- ------------------------------------------------------------------------------------------------
-- EXTRACT_COMPLIANCE_FLAGS
-- ------------------------------------------------------------------------------------------------
-- Builds a JSONB map of applicable regulatory framework flags by merging:
--   1. compliance_flags from the matching audit_sensitive_tables row (if any)
--   2. compliance_flags from any audit_sensitive_fields rows whose field_name is present
--      in either the old or new row data
--
-- The result is the union of all matched flags. Which frameworks appear (GDPR, SOX, HIPAA,
-- PCI-DSS, or any custom flag a module registers) is entirely determined by configuration.
--
-- Changed from IMMUTABLE → STABLE to enable config table reads.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION extract_compliance_flags(
  p_table_name TEXT,
  p_old_data   JSONB,
  p_new_data   JSONB
)
RETURNS JSONB AS $$
DECLARE
  v_flags         JSONB := '{}'::JSONB;
  v_combined_data JSONB;
  v_row           RECORD;
BEGIN
  v_combined_data := COALESCE(p_new_data, '{}'::JSONB) || COALESCE(p_old_data, '{}'::JSONB);

  -- Merge table-level compliance flags
  SELECT compliance_flags INTO v_row
  FROM audit_sensitive_tables
  WHERE table_name = p_table_name
    AND compliance_flags != '{}';

  IF FOUND THEN
    v_flags := v_flags || v_row.compliance_flags;
  END IF;

  -- Merge field-level compliance flags for every sensitive field present in the row
  FOR v_row IN
    SELECT asf.compliance_flags
    FROM audit_sensitive_fields asf
    WHERE v_combined_data ? asf.field_name
      AND asf.compliance_flags != '{}'
  LOOP
    v_flags := v_flags || v_row.compliance_flags;
  END LOOP;

  RETURN v_flags;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION extract_compliance_flags IS
  'Returns the union of all compliance flags applicable to this table and row. '
  'Flags come entirely from audit_sensitive_tables and audit_sensitive_fields — '
  'no framework (GDPR, SOX, HIPAA, PCI-DSS, etc.) is hardcoded in this function.';

-- ------------------------------------------------------------------------------------------------
-- DETERMINE_EVENT_CATEGORY
-- ------------------------------------------------------------------------------------------------
-- Resolves an event_category string for audit_log.event_category.
--
-- Resolution order:
--   1. Explicit override in audit_sensitive_tables.event_category (module-controlled)
--   2. Built-in rules for the platform IAM tables that every deployment has
--   3. Default: 'DATA' (correct for any trigger-generated DML event)
--
-- 'ACCESS' is intentionally NOT the default. ACCESS should only be set on authorisation-check
-- events (ALLOW/DENY decisions), not on DML mutations. Modules that need ACCESS events should
-- call audit_log_event() directly with the correct category rather than relying on a trigger.
--
-- Changed from IMMUTABLE → STABLE to enable config table reads.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION determine_event_category(p_table_name TEXT)
RETURNS VARCHAR(50) AS $$
DECLARE
  v_override VARCHAR(50);
BEGIN
  -- 1. Module-registered override takes highest precedence
  SELECT event_category INTO v_override
  FROM audit_sensitive_tables
  WHERE table_name = p_table_name
    AND event_category IS NOT NULL;

  IF FOUND THEN
    RETURN v_override;
  END IF;

  -- 2. Built-in platform rules (these tables exist in every deployment)
  RETURN CASE
    WHEN p_table_name IN ('users', 'user_sessions', 'auth_tokens')
      THEN 'AUTH'
    WHEN p_table_name IN ('roles', 'permissions', 'policies', 'tenants')
      THEN 'ADMIN'
    WHEN p_table_name IN ('audit_log', 'system_logs', 'audit_sensitive_tables', 'audit_sensitive_fields')
      THEN 'SYSTEM'
    WHEN p_table_name IN ('compliance_reports', 'audit_reports')
      THEN 'COMPLIANCE'
    -- 3. Default: DATA is correct for any DML trigger event
    ELSE 'DATA'
  END;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION determine_event_category IS
  'Resolves event_category for an audit_log row. Checks audit_sensitive_tables for a module '
  'override first, then applies built-in platform IAM rules, then defaults to DATA. '
  'ACCESS is never the default — it must be set explicitly for authorisation-check events.';

-- ------------------------------------------------------------------------------------------------
-- DETERMINE_SEVERITY
-- ------------------------------------------------------------------------------------------------
-- Resolves a severity level for audit_log.severity.
--
-- Resolution order:
--   1. Risk score >= 70                              → CRITICAL (always)
--   2. DELETE on a table with severity_on_delete set → use that value (module-controlled)
--   3. Risk score thresholds: >=50 HIGH, >=30 WARN, >=10 INFO, else LOW
--
-- Changed from IMMUTABLE → STABLE to enable config table reads.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION determine_severity(
  p_operation  TEXT,
  p_table_name TEXT,
  p_risk_score INTEGER
)
RETURNS VARCHAR(20) AS $$
DECLARE
  v_delete_severity VARCHAR(20);
BEGIN
  -- High risk score always overrides to CRITICAL
  IF p_risk_score >= 70 THEN
    RETURN 'CRITICAL';
  END IF;

  -- Module-registered DELETE severity override
  IF p_operation = 'DELETE' THEN
    SELECT severity_on_delete INTO v_delete_severity
    FROM audit_sensitive_tables
    WHERE table_name = p_table_name
      AND severity_on_delete IS NOT NULL;

    IF FOUND THEN
      RETURN v_delete_severity;
    END IF;
  END IF;

  -- Risk score thresholds
  RETURN CASE
    WHEN p_risk_score >= 50 THEN 'HIGH'
    WHEN p_risk_score >= 30 THEN 'WARN'
    WHEN p_risk_score >= 10 THEN 'INFO'
    ELSE 'LOW'
  END;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION determine_severity IS
  'Resolves severity level for an audit_log row. Risk score >= 70 always yields CRITICAL. '
  'DELETE severity on sensitive tables is controlled by audit_sensitive_tables.severity_on_delete. '
  'Falls back to risk score thresholds. No hardcoded table names.';

-- ------------------------------------------------------------------------------------------------
-- AUDIT_TRIGGER_FUNCTION
-- ------------------------------------------------------------------------------------------------
-- Universal AFTER trigger for INSERT, UPDATE, DELETE on any table.
-- Writes one row to audit_log per DML operation.
--
-- Key behaviours:
--   - Skips noise columns (updated_at, created_at, updated_by) from UPDATE diff.
--   - Extracts tenant_id from the row, falling back to current_tenant_id().
--   - Extracts user/session context from session variables via get_current_user_context().
--   - Calls the four helper functions above for scoring and classification.
--   - NEVER fails the originating DML: all exceptions are demoted to RAISE WARNING.
--   - Supports tables with either `id` or `uuid` as the primary key alias.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
  NOISE_COLUMNS CONSTANT TEXT[] := ARRAY['updated_at', 'created_at', 'updated_by'];

  v_tenant_id        UUID;
  v_user_context     JSONB;
  v_old_data         JSONB;
  v_new_data         JSONB;
  v_changed_fields   JSONB   := '{}'::JSONB;
  v_risk_score       INTEGER;
  v_compliance_flags JSONB;
  v_event_category   VARCHAR(50);
  v_severity         VARCHAR(20);
  v_context          JSONB   := '{}'::JSONB;
  v_record_id        TEXT;
  v_field_name       TEXT;
BEGIN
  -- ── Tenant resolution ──────────────────────────────────────────────────────────────────────
  v_tenant_id := COALESCE(
    CASE WHEN TG_OP = 'DELETE' THEN OLD.tenant_id ELSE NEW.tenant_id END,
    current_tenant_id()
  );

  -- ── User context ───────────────────────────────────────────────────────────────────────────
  v_user_context := get_current_user_context();

  -- ── Row serialisation ──────────────────────────────────────────────────────────────────────
  v_old_data := CASE WHEN TG_OP != 'INSERT' THEN to_jsonb(OLD) ELSE NULL END;
  v_new_data := CASE WHEN TG_OP != 'DELETE' THEN to_jsonb(NEW) ELSE NULL END;

  -- ── Changed-fields diff (UPDATE only; skip noise columns) ─────────────────────────────────
  IF TG_OP = 'UPDATE' THEN
    FOR v_field_name IN SELECT jsonb_object_keys(v_new_data) LOOP
      CONTINUE WHEN v_field_name = ANY(NOISE_COLUMNS);
      IF (v_old_data->v_field_name) IS DISTINCT FROM (v_new_data->v_field_name) THEN
        v_changed_fields := v_changed_fields || jsonb_build_object(
          v_field_name,
          jsonb_build_object('old', v_old_data->v_field_name, 'new', v_new_data->v_field_name)
        );
      END IF;
    END LOOP;
  END IF;

  -- ── Scoring and classification ─────────────────────────────────────────────────────────────
  v_risk_score       := calculate_audit_risk_score(TG_OP, TG_TABLE_NAME, v_old_data, v_new_data);
  v_compliance_flags := extract_compliance_flags(TG_TABLE_NAME, v_old_data, v_new_data);
  v_event_category   := determine_event_category(TG_TABLE_NAME);
  v_severity         := determine_severity(TG_OP, TG_TABLE_NAME, v_risk_score);

  -- ── Record ID (supports `id` or `uuid` PK aliases) ────────────────────────────────────────
  v_record_id := COALESCE(
    CASE WHEN TG_OP = 'DELETE' THEN (v_old_data->>'id')   ELSE (v_new_data->>'id')   END,
    CASE WHEN TG_OP = 'DELETE' THEN (v_old_data->>'uuid') ELSE (v_new_data->>'uuid') END
  );

  -- ── Context assembly ───────────────────────────────────────────────────────────────────────
  -- Top-level keys are stable across all versions for backward-compatible JSONB path queries.
  -- Operation-specific data is nested under `payload`.
  v_context := jsonb_build_object(
    'table_name',     TG_TABLE_NAME,
    'schema_name',    TG_TABLE_SCHEMA,
    'operation',      TG_OP,
    'record_id',      v_record_id,
    'changed_fields', v_changed_fields,
    'trigger_timing', TG_WHEN,
    'trigger_level',  TG_LEVEL,
    'payload',
      CASE TG_OP
        WHEN 'UPDATE' THEN jsonb_build_object('changes',        v_changed_fields)
        WHEN 'INSERT' THEN jsonb_build_object('new_record',     v_new_data)
        WHEN 'DELETE' THEN jsonb_build_object('deleted_record', v_old_data)
        ELSE '{}'::JSONB
      END
  );

  -- ── Write audit row ────────────────────────────────────────────────────────────────────────
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
    format('%s_%s', TG_TABLE_NAME, TG_OP),
    v_event_category,
    v_severity,
    (v_user_context->>'user_id')::UUID,
    'ALLOW',
    format('%s on %s by %s', TG_OP, TG_TABLE_NAME,
      COALESCE(v_user_context->>'user_id', 'unknown')),
    v_risk_score,
    v_context,
    (v_user_context->>'ip_address')::INET,
    v_user_context->>'user_agent',
    (v_user_context->>'session_id')::UUID,
    v_compliance_flags,
    NOW()
  );

  RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;

EXCEPTION
  WHEN OTHERS THEN
    -- Audit failure must never abort the originating transaction.
    RAISE WARNING 'Audit logging failed for %.% (%): %',
      TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_OP, SQLERRM;
    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION audit_trigger_function() IS
  'Universal AFTER trigger. Logs INSERT/UPDATE/DELETE into audit_log with risk scoring, '
  'compliance detection, and full context. All rules are config-table driven. '
  'Audit failures emit WARNING and never abort the originating DML.';

-- ------------------------------------------------------------------------------------------------
-- ENABLE_AUDIT_ON_TABLE
-- ------------------------------------------------------------------------------------------------
-- Creates AFTER triggers for the specified operations on a single table. Idempotent.
-- Signature unchanged from v1.
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
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = p_schema_name
      AND table_name   = p_table_name
  ) THEN
    RAISE EXCEPTION 'Table %.% does not exist', p_schema_name, p_table_name;
  END IF;

  FOREACH v_operation IN ARRAY p_operations LOOP
    v_operation    := upper(v_operation);
    v_trigger_name := format('audit_%s_%s_trigger', p_table_name, lower(v_operation));

    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I.%I',
      v_trigger_name, p_schema_name, p_table_name);

    EXECUTE format(
      'CREATE TRIGGER %I AFTER %s ON %I.%I FOR EACH ROW EXECUTE FUNCTION audit_trigger_function()',
      v_trigger_name, v_operation, p_schema_name, p_table_name
    );

    v_result := v_result || format('Created %s for %s on %s.%s%s',
      v_trigger_name, v_operation, p_schema_name, p_table_name, E'\n');
  END LOOP;

  RETURN v_result;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enable_audit_on_table IS
  'Enables audit logging on a specific table for the specified operations. '
  'Idempotent — drops and recreates triggers on each call. Operation names are case-insensitive.';

-- ------------------------------------------------------------------------------------------------
-- DISABLE_AUDIT_ON_TABLE
-- ------------------------------------------------------------------------------------------------
-- Drops all audit_* triggers from the specified table. Signature unchanged from v1.
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

    v_result := v_result || format('Dropped %s from %s.%s%s',
      v_trigger.trigger_name, p_schema_name, p_table_name, E'\n');
  END LOOP;

  RETURN COALESCE(
    NULLIF(v_result, ''),
    format('No audit triggers found on %s.%s', p_schema_name, p_table_name)
  );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION disable_audit_on_table IS
  'Drops all audit_* triggers from a specific table.';

-- ------------------------------------------------------------------------------------------------
-- ENABLE_AUDIT_ON_SCHEMA
-- ------------------------------------------------------------------------------------------------
-- Enables audit logging on every BASE TABLE in the schema that has a tenant_id column,
-- excluding specified tables. Idempotent. Signature unchanged from v1.
--
-- v2 fix: outer query is aliased `t` so the EXISTS subquery references t.table_name
-- explicitly, eliminating the scoping ambiguity present in v1.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION enable_audit_on_schema(
  p_schema_name    TEXT    DEFAULT 'public',
  p_exclude_tables TEXT[]  DEFAULT ARRAY['audit_log', 'schema_migrations',
                                         'audit_sensitive_tables', 'audit_sensitive_fields']
)
RETURNS TEXT AS $$
DECLARE
  v_table  RECORD;
  v_result TEXT := '';
BEGIN
  FOR v_table IN
    SELECT t.table_name
    FROM information_schema.tables t
    WHERE t.table_schema = p_schema_name
      AND t.table_type   = 'BASE TABLE'
      AND t.table_name  != ALL(p_exclude_tables)
      AND EXISTS (
        SELECT 1
        FROM information_schema.columns c
        WHERE c.table_schema = p_schema_name
          AND c.table_name   = t.table_name
          AND c.column_name  = 'tenant_id'
      )
  LOOP
    v_result := v_result || enable_audit_on_table(p_schema_name, v_table.table_name);
  END LOOP;

  RETURN COALESCE(
    NULLIF(v_result, ''),
    format('No eligible tables found in schema %s', p_schema_name)
  );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION enable_audit_on_schema IS
  'Enables audit logging on all tenant_id-bearing BASE TABLEs in a schema, excluding the '
  'specified list. Idempotent. The audit config tables themselves are excluded by default.';

-- ------------------------------------------------------------------------------------------------
-- GET_AUDIT_STATISTICS
-- ------------------------------------------------------------------------------------------------
-- Per-category, per-operation counts with severity breakdown and average risk score
-- over a rolling window. Signature unchanged from v1; warn_count added in v2.
--
-- Uses the generated `operation` column (000450 v2) with a fallback to array extraction
-- for pre-v2 rows where operation IS NULL.
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
  warn_count     BIGINT,
  avg_risk_score NUMERIC,
  unique_users   BIGINT
) AS $$
BEGIN
  RETURN QUERY
  SELECT
    al.event_category,
    COALESCE(
      al.operation,
      -- Fallback for pre-v2 rows: last segment of event_type (handles compound table names)
      NULLIF(
        (string_to_array(al.event_type, '_'))[cardinality(string_to_array(al.event_type, '_'))],
        al.event_type
      )
    )                                                           AS operation,
    COUNT(*)                                                    AS total_count,
    COUNT(*) FILTER (WHERE al.severity = 'CRITICAL')           AS critical_count,
    COUNT(*) FILTER (WHERE al.severity = 'HIGH')               AS high_count,
    COUNT(*) FILTER (WHERE al.severity = 'WARN')               AS warn_count,
    ROUND(AVG(al.risk_score)::NUMERIC, 2)                      AS avg_risk_score,
    COUNT(DISTINCT al.user_id)                                 AS unique_users
  FROM audit_log al
  WHERE (p_tenant_id IS NULL OR al.tenant_id = p_tenant_id)
    AND al.created_at >= NOW() - make_interval(hours => p_hours_back)
  GROUP BY
    al.event_category,
    COALESCE(
      al.operation,
      NULLIF(
        (string_to_array(al.event_type, '_'))[cardinality(string_to_array(al.event_type, '_'))],
        al.event_type
      )
    )
  ORDER BY total_count DESC;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_audit_statistics IS
  'Per-category, per-operation audit statistics over a rolling time window. '
  'Uses the generated operation column with a safe fallback for pre-v2 rows.';

-- ------------------------------------------------------------------------------------------------
-- EXAMPLE USAGE
-- ------------------------------------------------------------------------------------------------
-- NOTE: Reference examples only — not executed during migration.

-- Register a module's sensitive tables (put this in the module's own migration file):
-- SELECT register_audit_sensitive_table(
--   'finance_accounts', 20, 'DATA', 'HIGH', '{"SOX": true}', 'finance'
-- );
-- SELECT register_audit_sensitive_field('balance',        15, '{"SOX": true}', 'finance');
-- SELECT register_audit_sensitive_field('account_number', 15, '{}',            'finance');

-- Enable audit on a specific table:
-- SELECT enable_audit_on_table('public', 'finance_accounts');
-- SELECT enable_audit_on_table('public', 'documents', ARRAY['INSERT', 'DELETE']);

-- Enable audit on all eligible tables in a schema:
-- SELECT enable_audit_on_schema('public');

-- Disable audit on a table:
-- SELECT disable_audit_on_table('public', 'finance_accounts');

-- Inspect current configuration:
-- SELECT * FROM audit_sensitive_tables ORDER BY module_name, table_name;
-- SELECT * FROM audit_sensitive_fields ORDER BY module_name, field_name;

-- Audit statistics for the current tenant over the last 24 hours:
-- SELECT * FROM get_audit_statistics(current_tenant_id(), 24);

-- Recent high-risk events:
-- SELECT event_type, severity, risk_score, context->>'record_id', created_at
-- FROM audit_log
-- WHERE tenant_id  = current_tenant_id()
--   AND risk_score >= 50
-- ORDER BY created_at DESC
-- LIMIT 100;

-- Events matching a specific compliance flag (any framework):
-- SELECT event_type, severity, compliance_flags, created_at
-- FROM audit_log
-- WHERE tenant_id       = current_tenant_id()
--   AND compliance_flags ? 'SOX'
-- ORDER BY created_at DESC;

-- Changed fields on a specific record:
-- SELECT context->'payload'->'changes', created_at
-- FROM audit_log
-- WHERE tenant_id             = current_tenant_id()
--   AND context->>'record_id' = '<uuid>'
-- ORDER BY created_at DESC;
