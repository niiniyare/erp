DROP POLICY IF EXISTS notification_preferences_tenant_isolation ON notification_preferences;
DROP POLICY IF EXISTS notification_preferences_admin_access ON notification_preferences;
DROP POLICY IF EXISTS notification_preferences_ro_select ON notification_preferences;
ALTER TABLE IF EXISTS notification_preferences NO FORCE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS notification_preferences;
