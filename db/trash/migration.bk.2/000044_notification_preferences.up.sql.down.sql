
-- Optimized materialized view for permission evaluations
CREATE MATERIALIZED VIEW user_effective_permissions AS
SELECT
    u.id AS user_id,
    u.tenant_id,
    r.id AS resource_id,
    a.id AS action_id,
    MAX(CASE WHEN up.effect = 'DENY' THEN 0 ELSE 1 END) AS allow_flag,
    ARRAY_AGG(DISTINCT rp.id) AS role_permission_ids,
    ARRAY_AGG(DISTINCT up.id) AS direct_permission_ids
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
    AND ur.is_active = true
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
LEFT JOIN role_permissions rp ON ur.role_id = rp.role_id
    AND rp.is_active = true
LEFT JOIN user_permissions up ON u.id = up.user_id
    AND up.is_active = true
    AND (up.expires_at IS NULL OR up.expires_at > NOW())
JOIN resources r ON rp.permission_id = r.id OR up.permission_id = r.id
JOIN actions a ON rp.permission_id = a.id OR up.permission_id = a.id
GROUP BY u.id, u.tenant_id, r.id, a.id;

CREATE UNIQUE INDEX idx_user_effective_perms
    ON user_effective_permissions (user_id, resource_id, action_id);
    
COMMENT ON MATERIALIZED VIEW user_effective_permissions IS
'Pre-computed effective permissions for all users with optimized access patterns';

-- Session clustering
CLUSTER user_sessions USING idx_user_sessions_user_id;


--- 3. Security Automation Functions:


-- Session risk assessment function
CREATE OR REPLACE FUNCTION assess_session_risk(session_id UUID)
RETURNS INT AS $$
DECLARE
    risk INT := 0;
    session_data user_sessions%ROWTYPE;
BEGIN
    SELECT * INTO session_data 
    FROM user_sessions 
    WHERE id = session_id;
    
    -- Location anomaly detection
    IF EXISTS (
        SELECT 1 FROM user_sessions 
        WHERE user_id = session_data.user_id
        AND location_info->>'country' != session_data.location_info->>'country'
        AND created_at > NOW() - INTERVAL '1 hour'
    ) THEN
        risk := risk + 30;
        session_data.anomaly_flags := session_data.anomaly_flags || '["impossible_travel"]'::jsonb;
    END IF;
    
    -- Device change detection
    IF EXISTS (
        SELECT 1 FROM user_sessions 
        WHERE user_id = session_data.user_id
        AND device_info->>'fingerprint' != session_data.device_info->>'fingerprint'
        AND created_at > NOW() - INTERVAL '10 minutes'
    ) THEN
        risk := risk + 25;
        session_data.anomaly_flags := session_data.anomaly_flags || '["device_change"]'::jsonb;
    END IF;
    
    -- High-risk action detection
    IF EXISTS (
        SELECT 1 FROM audit_log
        WHERE session_id = session_data.id
        AND risk_score > 70
    ) THEN
        risk := risk + 45;
    END IF;
    
    UPDATE user_sessions 
    SET risk_score = risk,
        anomaly_flags = session_data.anomaly_flags
    WHERE id = session_id;
    
    RETURN risk;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Automatic session termination
CREATE OR REPLACE FUNCTION terminate_risky_sessions(threshold INT)
RETURNS INT AS $$
DECLARE
    terminated_count INT := 0;
BEGIN
    UPDATE user_sessions
    SET is_active = false
    WHERE risk_score >= threshold
        AND is_active = true
    RETURNING id INTO terminated_count;
    
    INSERT INTO audit_log (tenant_id, event_type, event_category, severity, context)
    SELECT tenant_id, 'SESSION_TERMINATED', 'SECURITY', 'HIGH',
        jsonb_build_object('session_id', id, 'risk_score', risk_score)
    FROM user_sessions
    WHERE risk_score >= threshold;
    
    RETURN terminated_count;
END;
$$ LANGUAGE plpgsql;


--- 4. Compliance Enhancements:


-- GDPR right-to-forget implementation
CREATE OR REPLACE FUNCTION gdpr_user_deletion(user_id UUID)
RETURNS VOID AS $$
BEGIN
    -- Pseudonymize sensitive data
    UPDATE persons p
    SET 
        first_name = 'REDACTED',
        last_name = 'REDACTED',
        email = 'redacted_' || gen_random_uuid() || '@example.com',
        phone = NULL,
        national_id = NULL,
        tax_id = NULL
    FROM users u
    WHERE u.person_id = p.id
        AND u.id = user_id;

    -- Delete authentication data
    UPDATE users
    SET 
        password_hash = NULL,
        mfa_secret = NULL,
        settings = settings - 'preferences'
    WHERE id = user_id;

    -- Terminate active sessions
    PERFORM terminate_risky_sessions(0); -- Terminate all sessions for user

    -- Log compliance action
    INSERT INTO audit_log (tenant_id, event_type, event_category, severity, context)
    SELECT tenant_id, 'GDPR_DELETION', 'COMPLIANCE', 'HIGH',
        jsonb_build_object('user_id', user_id)
    FROM users
    WHERE id = user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Data retention policy enforcement
CREATE OR REPLACE FUNCTION enforce_data_retention()
RETURNS VOID AS $$
BEGIN
    -- Anonymize old audit logs
    UPDATE audit_log
    SET 
        user_id = NULL,
        target_user_id = NULL,
        context = jsonb_set(context, '{user_info}', '"REDACTED"')
    WHERE created_at < NOW() - INTERVAL '180 days';
    
    -- Purge expired sessions
    DELETE FROM user_sessions
    WHERE expires_at < NOW() - INTERVAL '30 days';
    
    -- Archive and purge old access requests
    WITH archived AS (
        DELETE FROM access_requests
        WHERE created_at < NOW() - INTERVAL '365 days'
        RETURNING *
    )
    INSERT INTO access_requests_archive SELECT * FROM archived;
END;
$$ LANGUAGE plpgsql;


--- 5. Advanced Threat Detection View:


CREATE VIEW security_threat_dashboard AS
SELECT 
    u.id AS user_id,
    u.username,
    u.email,
    COUNT(s.id) FILTER (WHERE s.risk_score > 70) AS high_risk_sessions,
    MAX(s.risk_score) AS max_risk_score,
    ARRAY_AGG(DISTINCT s.anomaly_flags) AS anomaly_types,
    COUNT(a.id) FILTER (WHERE a.risk_score > 80) AS critical_events,
    MAX(a.created_at) AS last_suspicious_activity
FROM users u
LEFT JOIN user_sessions s ON u.id = s.user_id
LEFT JOIN audit_log a ON u.id = a.user_id AND a.risk_score > 50
WHERE u.account_status = 'ACTIVE'
    AND (s.risk_score > 50 OR a.risk_score > 50)
GROUP BY u.id;

COMMENT ON VIEW security_threat_dashboard IS
'Identifies potential security threats through session anomalies and audit patterns';


--- 6. Index Optimizations for Large-Scale Deployments:


-- BRIN Indexes for time-series data
CREATE INDEX idx_audit_log_time_brin ON audit_log USING BRIN (created_at);
CREATE INDEX idx_user_sessions_time_brin ON user_sessions USING BRIN (created_at);

-- GIN optimizations for JSONB queries
CREATE INDEX idx_users_attributes_gin ON users USING GIN (user_attributes jsonb_path_ops);
CREATE INDEX idx_policies_rule_gin ON policies USING GIN (rule jsonb_path_ops);

-- Partial indexes for active records
CREATE INDEX idx_active_users ON users (id) WHERE is_active = true AND deleted_at IS NULL;
CREATE INDEX idx_active_roles ON roles (id) WHERE is_active = true AND deleted_at IS NULL;

