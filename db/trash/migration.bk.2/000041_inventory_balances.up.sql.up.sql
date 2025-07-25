-- Drop the security_notifications table
DROP TABLE IF EXISTS security_notifications;

-- Drop the triggers
DROP TRIGGER IF EXISTS notify_suspicious_login ON user_sessions;
DROP TRIGGER IF EXISTS notify_role_changes ON user_roles;

-- Drop the notification trigger function
DROP FUNCTION IF EXISTS trigger_security_notification();

