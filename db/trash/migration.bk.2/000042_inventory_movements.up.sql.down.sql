CREATE TABLE security_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    notification_type VARCHAR(50) NOT NULL 
        CHECK (notification_type IN ('SUSPICIOUS_LOGIN', 'PASSWORD_COMPROMISED', 'ROLE_CHANGE', 'PERMISSION_GRANT')),
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    acknowledged BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '7 days'
);
ALTER TABLE security_notifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON security_notifications
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY admin_full_access_policy ON security_notifications
    FOR ALL TO admin_role
    USING (true);

-- Notification trigger function
CREATE OR REPLACE FUNCTION trigger_security_notification()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' AND TG_TABLE_NAME = 'user_sessions') THEN
        IF NEW.risk_score > 60 THEN
            INSERT INTO security_notifications (tenant_id, user_id, notification_type, title, message, metadata)
            VALUES (
                NEW.tenant_id,
                NEW.user_id,
                'SUSPICIOUS_LOGIN',
                'New login from unusual location',
                'We detected a login from ' || (NEW.location_info->>'city') || ', ' || (NEW.location_info->>'country'),
                jsonb_build_object('session_id', NEW.id, 'device', NEW.device_info)
            );
        END IF;
    ELSIF (TG_OP = 'INSERT' AND TG_TABLE_NAME = 'user_roles') THEN
        INSERT INTO security_notifications (tenant_id, user_id, notification_type, title, message, metadata)
        VALUES (
            (SELECT tenant_id FROM users WHERE id = NEW.user_id),
            NEW.user_id,
            'ROLE_CHANGE',
            'Role assignment: ' || (SELECT name FROM roles WHERE id = NEW.role_id),
            'You have been assigned a new role',
            jsonb_build_object('role_id', NEW.role_id, 'assigned_by', NEW.assigned_by)
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
CREATE TRIGGER notify_suspicious_login
    AFTER INSERT ON user_sessions
    FOR EACH ROW
    WHEN (NEW.risk_score > 60)
    EXECUTE FUNCTION trigger_security_notification();
    
CREATE TRIGGER notify_role_changes
    AFTER INSERT ON user_roles
    FOR EACH ROW
    EXECUTE FUNCTION trigger_security_notification();


