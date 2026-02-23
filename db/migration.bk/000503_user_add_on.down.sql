-- 7. Revert Security Notification System
DROP TRIGGER IF EXISTS notify_role_changes ON user_roles;

DROP TRIGGER IF EXISTS notify_suspicious_login ON user_sessions;

DROP FUNCTION IF EXISTS trigger_security_notification;

DROP TABLE IF EXISTS security_notifications;

-- 6. Revert Index Optimizations
DROP INDEX IF EXISTS idx_active_roles;

DROP INDEX IF EXISTS idx_active_users;

DROP INDEX IF EXISTS idx_policies_rule_gin;

-- DROP INDEX IF EXISTS idx_users_attributes_gin;
DROP INDEX IF EXISTS idx_user_sessions_time_brin;

DROP INDEX IF EXISTS idx_audit_log_time_brin;

-- 5. Revert Advanced Threat Detection View
DROP VIEW IF EXISTS v_security_threat_dashboard;

-- 4. Revert Compliance Enhancements
DROP FUNCTION IF EXISTS enforce_data_retention;

DROP FUNCTION IF EXISTS gdpr_user_deletion;

-- 3. Revert Security Automation Functions
DROP FUNCTION IF EXISTS terminate_risky_sessions;

DROP FUNCTION IF EXISTS assess_session_risk;

-- 2. Revert Performance Optimizations
DROP MATERIALIZED VIEW IF EXISTS mv_user_effective_permissions;

-- 1. Revert Security Hardening Enhancements
ALTER TABLE
  user_sessions DROP COLUMN IF EXISTS anomaly_flags,
  DROP COLUMN IF EXISTS mfa_verified_at;
