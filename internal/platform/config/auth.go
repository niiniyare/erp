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

	// MFA settings
	// MFAEncryptionKey must be exactly 32 characters — used as AES-256 key to encrypt TOTP secrets at rest.
	// Set via MFA_ENCRYPTION_KEY env var.  Dev default is a fixed insecure string; override in production.
	MFAEncryptionKey string `yaml:"mfa_encryption_key" mapstructure:"mfa_encryption_key"`
	// MFAIssuer is the issuer name shown in authenticator apps (e.g. "AWO ERP").
	MFAIssuer string `yaml:"mfa_issuer" mapstructure:"mfa_issuer"`

	// SSOEncryptionKey is the 32-byte AES-256 key used to encrypt OAuth client secrets at rest.
	// Set via SSO_ENCRYPTION_KEY env var. Dev default is insecure; override in production.
	SSOEncryptionKey string `yaml:"sso_encryption_key" mapstructure:"sso_encryption_key"`
}
