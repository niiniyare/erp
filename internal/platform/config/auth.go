package config

// AuthConfig represents auth configuration loaded from config.yaml / environment.
//
// TODO(auth): Expand with session and brute-force settings so that
// identity.Config can be populated from this struct at startup rather than
// using hardcoded defaults in identity.DefaultConfig():
//
//	MaxFailedAttempts  int           `yaml:"max_failed_attempts"`   // default 5
//	LockoutDuration    time.Duration `yaml:"lockout_duration"`       // default 15m
//	SessionTTL         time.Duration `yaml:"session_ttl"`            // default 8h
//	CookieName         string        `yaml:"cookie_name"`            // default "session"
//	RequireHTTPS       bool          `yaml:"require_https"`          // default true
//
// NOTE(settings): Per-tenant overrides for the above come from the Settings
// module (module="iam", keys: max_failed_attempts, lockout_duration_minutes,
// session_ttl_hours). The values in this struct are the process-wide fallbacks.
// See docs/reference/modules/settings/15-service-integration.md §Pattern 1.
//
// FIXME: Remove JWTSecret once the session-token authn migration (tasks.md S6)
// is complete — the system uses opaque session tokens stored in user_sessions,
// not JWTs. Keeping it here only for backward compatibility during transition.
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret" mapstructure:"jwt_secret"`
}
