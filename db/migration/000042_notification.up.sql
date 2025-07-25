CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    in_app_notifications BOOLEAN NOT NULL DEFAULT true,
    slack_notifications BOOLEAN NOT NULL DEFAULT false,
    notification_types JSONB NOT NULL DEFAULT '{}'::jsonb,
    preferred_channels JSONB NOT NULL DEFAULT '[]'::jsonb,
    quiet_hours JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notification_preferences_user_id_unique UNIQUE (tenant_id, user_id)
);

COMMENT ON TABLE notification_preferences IS 'Stores user notification preferences.';
COMMENT ON COLUMN notification_preferences.notification_types IS 'JSONB object with notification types as keys and booleans as values.';
COMMENT ON COLUMN notification_preferences.preferred_channels IS 'JSONB array of preferred notification channels.';
COMMENT ON COLUMN notification_preferences.quiet_hours IS 'JSONB object with quiet hours settings.';

ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_preferences_tenant_isolation ON notification_preferences
    FOR ALL
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

CREATE TRIGGER update_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

