-- ================================================================================================
-- AUDIT LOG QUERIES - ADVANCED WITH SQLC.NARG/SQLC.ARG AND TENANT ISOLATION
-- ================================================================================================
-- name: CreateAuditEvent :one
INSERT INTO
  audit_log (
    tenant_id,
    user_id,
    event_type,
    event_category,
    severity,
    entity_id,
    decision,
    reason,
    context
  )
VALUES
  (
    current_tenant_id(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
  )
RETURNING
  *;

-- name: GetAuditEvents :many
-- Get audit events with optional filters and pagination
SELECT
  al.id,
  al.tenant_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.user_id,
  al.target_user_id,
  al.entity_id,
  al.resource_id,
  al.action_id,
  al.role_id,
  al.permission_id,
  al.decision,
  al.reason,
  al.risk_score,
  al.context,
  al.ip_address,
  al.user_agent,
  al.session_id,
  al.compliance_flags,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND (
    sqlc.narg('user_id')::UUID IS NULL
    OR al.user_id = sqlc.narg('user_id')
  )
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR al.event_category = sqlc.narg('event_category')
  )
  AND (
    sqlc.narg('severity')::VARCHAR IS NULL
    OR al.severity = sqlc.narg('severity')
  )
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditEventByID :one
-- Get a specific audit event by ID
SELECT
  *
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND id = sqlc.arg('id');

-- name: GetUserAuditHistory :many
-- Get audit history for a specific user
SELECT
  al.id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.reason,
  al.risk_score,
  al.context,
  al.ip_address,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.user_id = sqlc.arg('user_id')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR al.event_category = sqlc.narg('event_category')
  )
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetHighRiskEvents :many
-- Get high-risk audit events (risk_score >= threshold)
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.reason,
  al.risk_score,
  al.context,
  al.ip_address,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.risk_score >= sqlc.arg('min_risk_score')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR al.event_category = sqlc.narg('event_category')
  )
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.risk_score DESC,
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetFailedAccessAttempts :many
-- Get failed access attempts within a time range
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.entity_id,
  al.resource_id,
  al.reason,
  al.risk_score,
  al.ip_address,
  al.user_agent,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.decision = 'DENY'
  AND al.event_category = 'ACCESS'
  AND al.created_at >= sqlc.arg('start_time')
  AND al.created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('user_id')::UUID IS NULL
    OR al.user_id = sqlc.narg('user_id')
  )
  AND (
    sqlc.narg('ip_address')::INET IS NULL
    OR al.ip_address = sqlc.narg('ip_address')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetComplianceEvents :many
-- Get events with specific compliance flags
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.context,
  al.compliance_flags,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.compliance_flags ? sqlc.arg('compliance_flag')
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditEventsByEntity :many
-- Get audit events for a specific entity
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.reason,
  al.risk_score,
  al.context,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.entity_id = sqlc.arg('entity_id')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR al.event_category = sqlc.narg('event_category')
  )
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAuditEventsByResource :many
-- Get audit events for a specific resource
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.reason,
  al.context,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.resource_id = sqlc.arg('resource_id')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR al.event_category = sqlc.narg('event_category')
  )
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetAdminActions :many
-- Get administrative actions (events with target_user_id)
SELECT
  al.id,
  al.user_id,
  al.target_user_id,
  al.event_type,
  al.event_category,
  al.decision,
  al.reason,
  al.context,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.target_user_id IS NOT NULL
  AND al.event_category = 'ADMIN'
  AND (
    sqlc.narg('admin_user_id')::UUID IS NULL
    OR al.user_id = sqlc.narg('admin_user_id')
  )
  AND (
    sqlc.narg('target_user_id')::UUID IS NULL
    OR al.target_user_id = sqlc.narg('target_user_id')
  )
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
  AND (
    sqlc.narg('end_time')::TIMESTAMPTZ IS NULL
    OR al.created_at <= sqlc.narg('end_time')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetUserSessionEvents :many
-- Get audit events for a specific session
SELECT
  id,
  event_type,
  event_category,
  severity,
  decision,
  reason,
  context,
  created_at
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND session_id = sqlc.arg('session_id')
ORDER BY
  created_at ASC;

-- ================================================================================================
-- ADVANCED ANALYTICS QUERIES
-- ================================================================================================
-- name: GetAuditStatsByCategory :many
-- Count events by category within a time range with optional filters
SELECT
  event_category,
  COUNT(*) AS event_count,
  COUNT(*) FILTER (
    WHERE
      decision = 'DENY'
  ) AS denied_count,
  AVG(risk_score) AS avg_risk_score,
  MAX(risk_score) AS max_risk_score,
  COUNT(DISTINCT user_id) AS unique_users
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND created_at >= sqlc.arg('start_time')
  AND created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('severity')::VARCHAR IS NULL
    OR severity = sqlc.narg('severity')
  )
GROUP BY
  event_category
ORDER BY
  event_count DESC;

-- name: GetAuditStatsBySeverity :many
-- Count events by severity within a time range
SELECT
  severity,
  COUNT(*) AS event_count,
  COUNT(DISTINCT user_id) AS unique_users,
  AVG(risk_score) AS avg_risk_score
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND created_at >= sqlc.arg('start_time')
  AND created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR event_category = sqlc.narg('event_category')
  )
GROUP BY
  severity
ORDER BY
  CASE
    severity
    WHEN 'CRITICAL' THEN 5
    WHEN 'HIGH' THEN 4
    WHEN 'WARN' THEN 3
    WHEN 'INFO' THEN 2
    WHEN 'LOW' THEN 1
  END DESC;

-- name: GetSuspiciousActivityByIP :many
-- Get suspicious activity from specific IP addresses with risk analysis
SELECT
  al.ip_address,
  al.user_id,
  al.event_type,
  al.decision,
  al.risk_score,
  al.created_at,
  COUNT(*) OVER (PARTITION BY al.ip_address) AS ip_event_count,
  COUNT(*) FILTER (
    WHERE
      al.decision = 'DENY'
  ) OVER (PARTITION BY al.ip_address) AS ip_denied_count,
  COUNT(DISTINCT al.user_id) OVER (PARTITION BY al.ip_address) AS unique_users_per_ip
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND (
    sqlc.narg('ip_address')::INET IS NULL
    OR al.ip_address = sqlc.narg('ip_address')
  )
  AND (
    al.decision = 'DENY'
    OR al.risk_score >= sqlc.arg('min_risk_score')
  )
  AND al.created_at >= sqlc.arg('start_time')
ORDER BY
  al.risk_score DESC,
  al.created_at DESC
LIMIT
  sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetUserRiskProfile :one
-- Get comprehensive risk profile for a user
SELECT
  al.user_id,
  COUNT(*) AS total_events,
  AVG(al.risk_score) AS avg_risk_score,
  MAX(al.risk_score) AS max_risk_score,
  COUNT(*) FILTER (
    WHERE
      al.decision = 'DENY'
  ) AS failed_attempts,
  COUNT(*) FILTER (
    WHERE
      al.severity IN ('HIGH', 'CRITICAL')
  ) AS high_severity_events,
  COUNT(DISTINCT al.ip_address) AS unique_ips,
  COUNT(DISTINCT al.event_category) AS unique_categories,
  MIN(al.created_at) AS first_event,
  MAX(al.created_at) AS last_event,
  COUNT(*) FILTER (
    WHERE
      al.created_at >= NOW() - INTERVAL '24 hours'
  ) AS events_last_24h,
  COUNT(*) FILTER (
    WHERE
      al.created_at >= NOW() - INTERVAL '7 days'
  ) AS events_last_7d
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.user_id = sqlc.arg('user_id')
  AND (
    sqlc.narg('start_time')::TIMESTAMPTZ IS NULL
    OR al.created_at >= sqlc.narg('start_time')
  )
GROUP BY
  al.user_id;

-- name: GetRecentSecurityEvents :many
-- Get recent security-related events (high risk, denials, critical severity)
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.reason,
  al.risk_score,
  al.ip_address,
  al.context,
  al.created_at
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND (
    al.risk_score >= sqlc.arg('min_risk_score')
    OR al.decision = 'DENY'
    OR al.severity IN ('HIGH', 'CRITICAL')
  )
  AND al.created_at >= sqlc.arg('start_time')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR al.event_category = sqlc.narg('event_category')
  )
ORDER BY
  CASE
    al.severity
    WHEN 'CRITICAL' THEN 5
    WHEN 'HIGH' THEN 4
    WHEN 'WARN' THEN 3
    WHEN 'INFO' THEN 2
    WHEN 'LOW' THEN 1
  END DESC,
  al.risk_score DESC,
  al.created_at DESC
LIMIT
  sqlc.arg('limit');

-- name: GetAnomalousUserBehavior :many
-- Detect anomalous user behavior patterns
WITH user_stats AS (
  SELECT
    user_id,
    AVG(risk_score) AS avg_risk_score,
    STDDEV(risk_score) AS stddev_risk_score,
    COUNT(*) AS event_count
  FROM
    audit_log
  WHERE
    tenant_id = current_tenant_id()
    AND audit_log.created_at >= sqlc.arg('baseline_start')
    AND audit_log.created_at <= sqlc.arg('baseline_end')
  GROUP BY
    user_id
  HAVING
    COUNT(*) >= sqlc.arg('min_baseline_events')
)
SELECT
  al.user_id,
  al.event_type,
  al.event_category,
  al.risk_score,
  al.created_at,
  u.email AS user_email,
  e.name AS entity_name,
  r.name AS resource_name,
  CASE
    WHEN al.context ? 'personal_data' THEN 'PERSONAL_DATA'
    WHEN al.context ? 'sensitive_data' THEN 'SENSITIVE_DATA'
    WHEN al.context ? 'financial_data' THEN 'FINANCIAL_DATA'
    WHEN al.context ? 'health_data' THEN 'HEALTH_DATA'
    ELSE 'GENERAL_DATA'
  END AS data_classification
FROM
  audit_log al
  LEFT JOIN users u ON al.user_id = u.id
  LEFT JOIN entities e ON al.entity_id = e.uuid
  AND u.tenant_id = current_tenant_id()
  AND e.tenant_id = current_tenant_id()
  LEFT JOIN resources r ON al.resource_id = r.id
  AND r.tenant_id = current_tenant_id()
WHERE
  al.tenant_id = current_tenant_id()
  AND al.event_category = 'DATA'
  AND al.created_at >= sqlc.arg('start_time')
  AND al.created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('data_classification')::VARCHAR IS NULL
    OR CASE
      WHEN al.context ? 'personal_data' THEN 'PERSONAL_DATA'
      WHEN al.context ? 'sensitive_data' THEN 'SENSITIVE_DATA'
      WHEN al.context ? 'financial_data' THEN 'FINANCIAL_DATA'
      WHEN al.context ? 'health_data' THEN 'HEALTH_DATA'
      ELSE 'GENERAL_DATA'
    END = sqlc.narg('data_classification')
  )
ORDER BY
  al.created_at DESC
LIMIT
  sqlc.arg('limit');

-- -- name: GetAccessControlEffectiveness :many
-- -- Analyze access control effectiveness
-- -- NOTE:
-- SELECT
--   stats.user_id,
--   stats.resource_id,
--   r.name AS resource_name,
--   stats.role_id,
--   ro.name AS role_name,
--   stats.permission_id,
--   p.name AS permission_name,
--   stats.total_attempts,
--   stats.allowed_attempts,
--   stats.denied_attempts,
--   ROUND(
--     stats.denied_attempts::NUMERIC / stats.total_attempts * 100,
--     2
--   ) AS denial_rate_pct,
--   ROUND(stats.avg_risk_score, 2) AS avg_risk_score,
--   stats.first_attempt,
--   stats.last_attempt
-- FROM (
--   SELECT
--     user_id,
--     resource_id,
--     role_id,
--     permission_id,
--     COUNT(*) AS total_attempts,
--     COUNT(*) FILTER (WHERE decision = 'ALLOW') AS allowed_attempts,
--     COUNT(*) FILTER (WHERE decision = 'DENY') AS denied_attempts,
--     AVG(risk_score) AS avg_risk_score,
--     MIN(created_at) AS first_attempt,
--     MAX(created_at) AS last_attempt
--   FROM audit_log al
--   WHERE
--     tenant_id = current_tenant_id()
--     AND event_category = 'ACCESS'
--     AND decision IS NOT NULL
--     AND al.created_at >= sqlc.arg('start_time')
--     AND al.created_at <= sqlc.arg('end_time')
--   GROUP BY user_id, resource_id, role_id, permission_id
-- ) stats
-- LEFT JOIN resources r ON stats.resource_id = r.id AND r.tenant_id = current_tenant_id()
-- LEFT JOIN roles ro ON stats.role_id = ro.id AND ro.tenant_id = current_tenant_id()
-- LEFT JOIN permissions p ON stats.permission_id = p.id AND p.tenant_id = current_tenant_id()
-- WHERE
--   stats.total_attempts >= sqlc.arg('min_attempts')
--   AND (
--     sqlc.narg('min_denial_rate')::NUMERIC IS NULL
--     OR (stats.denied_attempts::NUMERIC / stats.total_attempts) >= sqlc.narg('min_denial_rate')
--   )
-- ORDER BY denial_rate_pct DESC, stats.avg_risk_score DESC
-- LIMIT sqlc.arg('limit');
-- ================================================================================================
-- FORENSIC INVESTIGATION QUERIES
-- ================================================================================================
-- name: GetEventTimelineForUser :many
-- Get detailed event timeline for forensic investigation
SELECT
  al.id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.reason,
  al.risk_score,
  al.context,
  al.ip_address,
  al.user_agent,
  al.session_id,
  al.created_at,
  LAG(al.created_at) OVER (
    ORDER BY
      al.created_at
  ) AS prev_event_time,
  LEAD(al.created_at) OVER (
    ORDER BY
      al.created_at
  ) AS next_event_time,
  EXTRACT(
    EPOCH
    FROM
      (
        al.created_at - LAG(al.created_at) OVER (
          ORDER BY
            al.created_at
        )
      )
  ) AS seconds_since_prev
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.user_id = sqlc.arg('user_id')
  AND al.created_at >= sqlc.arg('start_time')
  AND al.created_at <= sqlc.arg('end_time')
ORDER BY
  al.created_at ASC;

-- name: GetRelatedEventsByContext :many
-- Find related events by context similarity
SELECT
  al.id,
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.risk_score,
  al.context,
  al.created_at,
  CASE
    WHEN al.context ? sqlc.arg('context_key') THEN 'EXACT_MATCH'
    WHEN al.context::text ILIKE '%' || sqlc.arg('context_search') || '%' THEN 'PARTIAL_MATCH'
    ELSE 'NO_MATCH'
  END AS context_match_type
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND (
    al.context ? sqlc.arg('context_key')
    OR al.context::text ILIKE '%' || sqlc.arg('context_search') || '%'
  )
  AND al.created_at >= sqlc.arg('start_time')
  AND al.created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('exclude_event_id')::UUID IS NULL
    OR al.id != sqlc.narg('exclude_event_id')
  )
ORDER BY
  CASE
    context_match_type
    WHEN 'EXACT_MATCH' THEN 2
    WHEN 'PARTIAL_MATCH' THEN 1
    ELSE 0
  END DESC,
  al.risk_score DESC,
  al.created_at DESC
LIMIT
  sqlc.arg('limit');

-- name: GetSimilarIncidentPatterns :many
-- Find similar incident patterns for threat intelligence
WITH incident_features AS (
  SELECT
    al_inner.event_type,
    al_inner.event_category,
    COUNT(DISTINCT al_inner.user_id) AS unique_users,
    COUNT(DISTINCT al_inner.ip_address) AS unique_ips,
    AVG(al_inner.risk_score) AS avg_risk_score,
    COUNT(*) FILTER (
      WHERE
        al_inner.decision = 'DENY'
    ) AS denied_count,
    COUNT(*) AS total_events,
    array_agg(DISTINCT al_inner.severity) AS severity_levels
  FROM
    audit_log al_inner
  WHERE
    al_inner.tenant_id = current_tenant_id()
    AND al_inner.event_type = sqlc.arg('incident_event_type')
    AND al_inner.event_category = sqlc.arg('incident_event_category')
    AND al_inner.created_at >= sqlc.arg('pattern_start')
    AND al_inner.created_at <= sqlc.arg('pattern_end')
  GROUP BY
    al_inner.event_type,
    al_inner.event_category
)
SELECT
  al.user_id,
  al.event_type,
  al.event_category,
  al.severity,
  al.decision,
  al.risk_score,
  al.ip_address,
  al.created_at,
  iff.avg_risk_score AS pattern_avg_risk,
  iff.denied_count AS pattern_denied_count,
  ABS(al.risk_score - iff.avg_risk_score) AS risk_deviation
FROM
  audit_log al
  CROSS JOIN incident_features iff
WHERE
  al.tenant_id = current_tenant_id()
  AND al.event_type = iff.event_type
  AND al.event_category = iff.event_category
  AND al.created_at >= sqlc.arg('search_start')
  AND al.created_at <= sqlc.arg('search_end')
  AND ABS(al.risk_score - iff.avg_risk_score) <= sqlc.arg('max_risk_deviation')
ORDER BY
  risk_deviation ASC,
  al.created_at DESC
LIMIT
  sqlc.arg('limit');

-- ================================================================================================
-- PERFORMANCE AND MONITORING QUERIES
-- ================================================================================================
-- name: GetAuditLogHealth :one
-- Get audit log health metrics and performance indicators
WITH event_volume AS (
  SELECT
    COUNT(*) AS total_events,
    COUNT(*) FILTER (
      WHERE
        created_at >= NOW() - INTERVAL '1 hour'
    ) AS events_last_hour,
    COUNT(*) FILTER (
      WHERE
        created_at >= NOW() - INTERVAL '24 hours'
    ) AS events_last_24h,
    COUNT(*) FILTER (
      WHERE
        created_at >= NOW() - INTERVAL '7 days'
    ) AS events_last_7d
  FROM
    audit_log
  WHERE
    tenant_id = current_tenant_id()
),
risk_metrics AS (
  SELECT
    AVG(risk_score) AS avg_risk_score,
    STDDEV(risk_score) AS stddev_risk_score,
    PERCENTILE_CONT(0.95) WITHIN GROUP (
      ORDER BY
        risk_score
    ) AS p95_risk_score,
    COUNT(*) FILTER (
      WHERE
        risk_score >= 80
    ) AS high_risk_events
  FROM
    audit_log
  WHERE
    tenant_id = current_tenant_id()
    AND created_at >= NOW() - INTERVAL '24 hours'
),
decision_metrics AS (
  SELECT
    COUNT(*) FILTER (
      WHERE
        decision = 'ALLOW'
    ) AS allowed_events,
    COUNT(*) FILTER (
      WHERE
        decision = 'DENY'
    ) AS denied_events,
    ROUND(
      COUNT(*) FILTER (
        WHERE
          decision = 'DENY'
      )::NUMERIC / COUNT(*) * 100,
      2
    ) AS denial_rate_pct
  FROM
    audit_log
  WHERE
    tenant_id = current_tenant_id()
    AND decision IS NOT NULL
    AND created_at >= NOW() - INTERVAL '24 hours'
)
SELECT
  ev.*,
  rm.avg_risk_score,
  rm.stddev_risk_score,
  rm.p95_risk_score,
  rm.high_risk_events,
  dm.allowed_events,
  dm.denied_events,
  dm.denial_rate_pct
FROM
  event_volume ev
  CROSS JOIN risk_metrics rm
  CROSS JOIN decision_metrics dm;

-- name: GetHourlyEventRates :many
-- Get hourly event rates for capacity planning
SELECT
  DATE_TRUNC('hour', al.created_at) AS hour_bucket,
  COUNT(*) AS event_count,
  COUNT(DISTINCT al.user_id) AS unique_users,
  AVG(al.risk_score) AS avg_risk_score,
  COUNT(*) FILTER (
    WHERE
      al.decision = 'DENY'
  ) AS denied_count
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.created_at >= sqlc.arg('start_time')
  AND al.created_at <= sqlc.arg('end_time')
GROUP BY
  DATE_TRUNC('hour', al.created_at)
ORDER BY
  hour_bucket DESC
LIMIT
  sqlc.arg('limit');

-- name: GetEventTypeDistribution :many
-- Get event type distribution for analysis
SELECT
  event_type,
  event_category,
  COUNT(*) AS event_count,
  COUNT(DISTINCT user_id) AS unique_users,
  AVG(risk_score) AS avg_risk_score,
  COUNT(*) FILTER (
    WHERE
      decision = 'DENY'
  ) AS denied_count,
  COUNT(*) FILTER (
    WHERE
      severity IN ('HIGH', 'CRITICAL')
  ) AS high_severity_count,
  MIN(created_at) AS first_occurrence,
  MAX(created_at) AS last_occurrence
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND created_at >= sqlc.arg('start_time')
  AND created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR event_category = sqlc.narg('event_category')
  )
GROUP BY
  event_type,
  event_category
HAVING
  COUNT(*) >= sqlc.arg('min_occurrences')
ORDER BY
  event_count DESC
LIMIT
  sqlc.arg('limit');

-- name: GetUserAgentAnalysis :many
-- Analyze user agent patterns for security insights
SELECT
  user_agent,
  COUNT(*) AS event_count,
  COUNT(DISTINCT user_id) AS unique_users,
  COUNT(DISTINCT ip_address) AS unique_ips,
  AVG(risk_score) AS avg_risk_score,
  COUNT(*) FILTER (
    WHERE
      decision = 'DENY'
  ) AS denied_count,
  MIN(created_at) AS first_seen,
  MAX(created_at) AS last_seen,
  CASE
    WHEN user_agent IS NULL THEN 'NO_USER_AGENT'
    WHEN user_agent ILIKE '%bot%'
    OR user_agent ILIKE '%crawler%' THEN 'BOT'
    WHEN user_agent ILIKE '%curl%'
    OR user_agent ILIKE '%wget%' THEN 'AUTOMATED_TOOL'
    WHEN LENGTH(user_agent) < 20 THEN 'SUSPICIOUS_SHORT'
    ELSE 'NORMAL'
  END AS agent_category
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND created_at >= sqlc.arg('start_time')
  AND created_at <= sqlc.arg('end_time')
GROUP BY
  user_agent
HAVING
  COUNT(*) >= sqlc.arg('min_events')
ORDER BY
  CASE
    agent_category
    WHEN 'SUSPICIOUS_SHORT' THEN 4
    WHEN 'AUTOMATED_TOOL' THEN 3
    WHEN 'BOT' THEN 2
    WHEN 'NO_USER_AGENT' THEN 1
    ELSE 0
  END DESC,
  avg_risk_score DESC,
  event_count DESC
LIMIT
  sqlc.arg('limit');

-- ================================================================================================
-- MAINTENANCE AND UTILITY QUERIES
-- ================================================================================================
-- name: UpdateEventRiskScore :exec
-- Update risk score for an audit event
UPDATE
  audit_log
SET
  risk_score = sqlc.arg('risk_score')
WHERE
  tenant_id = current_tenant_id()
  AND id = sqlc.arg('id');

-- name: UpdateEventComplianceFlags :exec
-- Update compliance flags for an audit event
UPDATE
  audit_log
SET
  compliance_flags = sqlc.arg('compliance_flags')
WHERE
  tenant_id = current_tenant_id()
  AND id = sqlc.arg('id');

-- name: BulkUpdateEventRiskScores :exec
-- Bulk update risk scores based on criteria
UPDATE
  audit_log
SET
  risk_score = sqlc.arg('new_risk_score')
WHERE
  tenant_id = current_tenant_id()
  AND event_type = sqlc.arg('event_type')
  AND created_at >= sqlc.arg('start_time')
  AND created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('current_risk_score')::INTEGER IS NULL
    OR risk_score = sqlc.narg('current_risk_score')
  )
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR event_category = sqlc.narg('event_category')
  );

-- name: BulkAddComplianceFlags :exec
-- Bulk add compliance flags to events matching criteria
UPDATE
  audit_log
SET
  compliance_flags = compliance_flags || sqlc.arg('new_compliance_flags')::jsonb
WHERE
  tenant_id = current_tenant_id()
  AND event_type = sqlc.arg('event_type')
  AND created_at >= sqlc.arg('start_time')
  AND created_at <= sqlc.arg('end_time')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR event_category = sqlc.narg('event_category')
  );

-- name: DeleteOldAuditEvents :exec
-- Delete audit events older than specified date (for retention policies)
DELETE FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND created_at < sqlc.arg('cutoff_date');

-- name: ArchiveOldAuditEvents :many
-- Archive old audit events before deletion (returns what will be deleted)
SELECT
  id,
  event_type,
  event_category,
  user_id,
  risk_score,
  created_at
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND created_at < sqlc.arg('cutoff_date')
  AND (
    sqlc.narg('event_category')::VARCHAR IS NULL
    OR event_category = sqlc.narg('event_category')
  )
ORDER BY
  created_at DESC
LIMIT
  sqlc.arg('limit');

-- name: GetAuditStorageStats :one
-- Get audit log storage statistics and metrics
SELECT
  COUNT(*) AS total_events,
  COUNT(DISTINCT user_id) AS unique_users,
  COUNT(DISTINCT ip_address) AS unique_ips,
  COUNT(DISTINCT event_type) AS unique_event_types,
  pg_size_pretty(pg_total_relation_size('audit_log')) AS table_size,
  MIN(created_at) AS oldest_event,
  MAX(created_at) AS newest_event,
  COUNT(*) FILTER (
    WHERE
      created_at >= NOW() - INTERVAL '24 hours'
  ) AS events_last_24h,
  COUNT(*) FILTER (
    WHERE
      created_at >= NOW() - INTERVAL '7 days'
  ) AS events_last_7d,
  COUNT(*) FILTER (
    WHERE
      created_at >= NOW() - INTERVAL '30 days'
  ) AS events_last_30d
FROM
  audit_log
WHERE
  tenant_id = current_tenant_id();

-- name: GetDuplicateEventAnalysis :many
-- Identify potential duplicate events for cleanup
SELECT
  al.user_id,
  al.event_type,
  al.event_category,
  al.ip_address,
  DATE_TRUNC('minute', al.created_at) AS time_bucket,
  COUNT(*) AS duplicate_count,
  array_agg(
    al.id
    ORDER BY
      al.created_at
  ) AS event_ids,
  MIN(al.created_at) AS first_occurrence,
  MAX(al.created_at) AS last_occurrence
FROM
  audit_log al
WHERE
  al.tenant_id = current_tenant_id()
  AND al.created_at >= sqlc.arg('start_time')
  AND al.created_at <= sqlc.arg('end_time')
GROUP BY
  al.user_id,
  al.event_type,
  al.event_category,
  al.ip_address,
  DATE_TRUNC('minute', al.created_at)
HAVING
  COUNT(*) >= sqlc.arg('min_duplicate_count')
ORDER BY
  duplicate_count DESC,
  first_occurrence DESC
LIMIT
  sqlc.arg('limit');

-- name: CleanupDuplicateEvents :exec
-- Remove duplicate events keeping only the first occurrence
WITH duplicates AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY user_id,
      event_type,
      event_category,
      ip_address,
      DATE_TRUNC('minute', created_at)
      ORDER BY
        created_at ASC
    ) AS rn
  FROM
    audit_log
  WHERE
    tenant_id = current_tenant_id()
    AND audit_log.created_at >= sqlc.arg('start_time')
    AND audit_log.created_at <= sqlc.arg('end_time')
)
DELETE FROM
  audit_log
WHERE
  tenant_id = current_tenant_id()
  AND id IN (
    SELECT
      id
    FROM
      duplicates
    WHERE
      rn > 1
  );
