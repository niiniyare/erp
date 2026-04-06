-- ------------------------------------------------------------------------------------------------
-- NOTIFICATION PREFERENCES
-- ------------------------------------------------------------------------------------------------
-- Stores per-user notification delivery preferences scoped to a tenant.
-- Tracks channel toggles (email, in-app, Slack), per-type overrides, and quiet hours.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notification_preferences (
  id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id            UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id              UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  email_notifications  BOOLEAN     NOT NULL DEFAULT TRUE,
  in_app_notifications BOOLEAN     NOT NULL DEFAULT TRUE,
  slack_notifications  BOOLEAN     NOT NULL DEFAULT false,
  notification_types   JSONB       NOT NULL DEFAULT '{}'::jsonb,   -- per-type boolean overrides
  preferred_channels   JSONB       NOT NULL DEFAULT '[]'::jsonb,   -- ordered channel list
  quiet_hours          JSONB,                                       -- start/end time config
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT notification_preferences_user_id_unique UNIQUE (tenant_id, user_id)
);

COMMENT ON TABLE notification_preferences IS 'Stores user notification preferences.';

COMMENT ON COLUMN notification_preferences.notification_types IS 'JSONB object with notification types as keys and booleans as values.';
COMMENT ON COLUMN notification_preferences.preferred_channels IS 'JSONB array of preferred notification channels.';
COMMENT ON COLUMN notification_preferences.quiet_hours        IS 'JSONB object with quiet hours settings.';

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_preferences_tenant_isolation ON notification_preferences
  FOR ALL
  USING (tenant_id = current_tenant_id())
  WITH CHECK (tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- TRIGGERS
-- ------------------------------------------------------------------------------------------------
CREATE TRIGGER update_notification_preferences_updated_at
  BEFORE UPDATE ON notification_preferences
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
