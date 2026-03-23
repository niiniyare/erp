package config

import "time"

// AuthConfig represents auth configuration loaded from config.yaml / environment.
//
// NOTE(settings): Per-tenant overrides for MaxFailedAttempts, LockoutDuration,
// and SessionTTL come from the Settings module (module="iam"). The values here
// are the process-wide fallbacks.
//
// FIXME: Remove JWTSecret once the session-token authn migration is complete —
// the system uses opaque session tokens stored in user_sessions, not JWTs.
type AuthConfig struct {
	// FIXME: remove after session-token migration is complete
	JWTSecret string `yaml:"jwt_secret" mapstructure:"jwt_secret"`

	// Brute-force protection thresholds
	MaxFailedAttempts int           `yaml:"max_failed_attempts" mapstructure:"max_failed_attempts"`
	LockoutDuration   time.Duration `yaml:"lockout_duration" mapstructure:"lockout_duration"`

	// Session settings
	SessionTTL   time.Duration `yaml:"session_ttl" mapstructure:"session_ttl"`
	CookieName   string        `yaml:"cookie_name" mapstructure:"cookie_name"`
	RequireHTTPS bool          `yaml:"require_https" mapstructure:"require_https"`
}
