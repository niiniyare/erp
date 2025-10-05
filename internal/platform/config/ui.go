package config

import (
	"time"

	"github.com/spf13/viper"
)

// UIConfig contains all UI-related configuration
type UIConfig struct {
	// Server Configuration
	Server UIServerConfig `yaml:"server" json:"server" mapstructure:"server"`

	// Service-specific configurations
	Console   UIServiceConfig `yaml:"console" json:"console" mapstructure:"console"`
	Workspace UIServiceConfig `yaml:"workspace" json:"workspace" mapstructure:"workspace"`
	Portal    UIServiceConfig `yaml:"portal" json:"portal" mapstructure:"portal"`

	// Security Configuration
	Security UISecurityConfig `yaml:"security" json:"security" mapstructure:"security"`

	// Session Management
	Session UISessionConfig `yaml:"session" json:"session" mapstructure:"session"`

	// Asset Management
	Assets UIAssetsConfig `yaml:"assets" json:"assets" mapstructure:"assets"`

	// Development Settings
	Development UIDevelopmentConfig `yaml:"development" json:"development" mapstructure:"development"`
}

// UIServerConfig contains UI-specific server configuration
type UIServerConfig struct {
	// Multi-port strategy for UI services
	MultiPort UIMultiPortConfig `yaml:"multi_port" json:"multi_port" mapstructure:"multi_port"`

	// UI-specific timeouts (inherits base timeouts from main ServerConfig)
	IdleTimeout time.Duration `yaml:"idle_timeout" json:"idle_timeout" mapstructure:"idle_timeout"`
}

// UIMultiPortConfig for multi-port deployment strategy
type UIMultiPortConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	ConsolePort   string `yaml:"console_port" json:"console_port" mapstructure:"console_port"`
	WorkspacePort string `yaml:"workspace_port" json:"workspace_port" mapstructure:"workspace_port"`
	PortalPort    string `yaml:"portal_port" json:"portal_port" mapstructure:"portal_port"`
}

// Note: TLS configuration is handled by main ServerConfig - no need to duplicate

// UIServiceConfig contains configuration for individual UI services
type UIServiceConfig struct {
	// Service Identity
	Name        string `yaml:"name" json:"name" mapstructure:"name"`
	Description string `yaml:"description" json:"description" mapstructure:"description"`
	Version     string `yaml:"version" json:"version" mapstructure:"version"`

	// Access Configuration
	Enabled     bool     `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	BaseURL     string   `yaml:"base_url" json:"base_url" mapstructure:"base_url"`
	ExternalURL string   `yaml:"external_url" json:"external_url" mapstructure:"external_url"`
	AllowedIPs  []string `yaml:"allowed_ips" json:"allowed_ips" mapstructure:"allowed_ips"`

	// Feature Flags
	Features UIServiceFeatures `yaml:"features" json:"features" mapstructure:"features"`

	// Branding
	Branding UIBrandingConfig `yaml:"branding" json:"branding" mapstructure:"branding"`

	// Rate Limiting
	RateLimit UIRateLimitConfig `yaml:"rate_limit" json:"rate_limit" mapstructure:"rate_limit"`
}

// UIServiceFeatures contains feature flags for UI services
type UIServiceFeatures struct {
	DarkMode       bool `yaml:"dark_mode" json:"dark_mode" mapstructure:"dark_mode"`
	Notifications  bool `yaml:"notifications" json:"notifications" mapstructure:"notifications"`
	RealTimeUpdate bool `yaml:"real_time_update" json:"real_time_update" mapstructure:"real_time_update"`
	BulkOperations bool `yaml:"bulk_operations" json:"bulk_operations" mapstructure:"bulk_operations"`
	Export         bool `yaml:"export" json:"export" mapstructure:"export"`
	Search         bool `yaml:"search" json:"search" mapstructure:"search"`
	MultiLanguage  bool `yaml:"multi_language" json:"multi_language" mapstructure:"multi_language"`
}

// UIBrandingConfig contains branding configuration
type UIBrandingConfig struct {
	CompanyName    string `yaml:"company_name" json:"company_name" mapstructure:"company_name"`
	LogoURL        string `yaml:"logo_url" json:"logo_url" mapstructure:"logo_url"`
	FaviconURL     string `yaml:"favicon_url" json:"favicon_url" mapstructure:"favicon_url"`
	PrimaryColor   string `yaml:"primary_color" json:"primary_color" mapstructure:"primary_color"`
	SecondaryColor string `yaml:"secondary_color" json:"secondary_color" mapstructure:"secondary_color"`
	CustomCSS      string `yaml:"custom_css" json:"custom_css" mapstructure:"custom_css"`
}

// UIRateLimitConfig contains rate limiting configuration
type UIRateLimitConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Requests   int           `yaml:"requests" json:"requests" mapstructure:"requests"`
	Window     time.Duration `yaml:"window" json:"window" mapstructure:"window"`
	BurstLimit int           `yaml:"burst_limit" json:"burst_limit" mapstructure:"burst_limit"`
	SkipPaths  []string      `yaml:"skip_paths" json:"skip_paths" mapstructure:"skip_paths"`
}

// UISecurityConfig contains security-related configuration
type UISecurityConfig struct {
	// JWT Configuration
	JWT UIJWTConfig `yaml:"jwt" json:"jwt" mapstructure:"jwt"`

	// CORS Configuration
	CORS UICORSConfig `yaml:"cors" json:"cors" mapstructure:"cors"`

	// CSP Configuration
	CSP UICSPConfig `yaml:"csp" json:"csp" mapstructure:"csp"`

	// Authentication
	Auth UIAuthConfig `yaml:"auth" json:"auth" mapstructure:"auth"`

	// Encryption
	Encryption UIEncryptionConfig `yaml:"encryption" json:"encryption" mapstructure:"encryption"`
}

// Note: JWT configuration should extend main AuthConfig - UI-specific token settings only
type UIJWTConfig struct {
	// UI-specific token settings (inherits from main AuthConfig)
	Audience        string        `yaml:"audience" json:"audience" mapstructure:"audience"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" json:"access_token_ttl" mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" json:"refresh_token_ttl" mapstructure:"refresh_token_ttl"`
}

// UICORSConfig contains CORS configuration
type UICORSConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers" mapstructure:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" json:"exposed_headers" mapstructure:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" json:"max_age" mapstructure:"max_age"`
}

// UICSPConfig contains Content Security Policy configuration
type UICSPConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	DefaultSrc string `yaml:"default_src" json:"default_src" mapstructure:"default_src"`
	ScriptSrc  string `yaml:"script_src" json:"script_src" mapstructure:"script_src"`
	StyleSrc   string `yaml:"style_src" json:"style_src" mapstructure:"style_src"`
	ImgSrc     string `yaml:"img_src" json:"img_src" mapstructure:"img_src"`
	FontSrc    string `yaml:"font_src" json:"font_src" mapstructure:"font_src"`
	ConnectSrc string `yaml:"connect_src" json:"connect_src" mapstructure:"connect_src"`
	ReportURI  string `yaml:"report_uri" json:"report_uri" mapstructure:"report_uri"`
	ReportOnly bool   `yaml:"report_only" json:"report_only" mapstructure:"report_only"`
}

// UIAuthConfig contains authentication configuration
type UIAuthConfig struct {
	// OAuth Configuration
	OAuth UIOAuthConfig `yaml:"oauth" json:"oauth" mapstructure:"oauth"`

	// SAML Configuration
	SAML UISAMLConfig `yaml:"saml" json:"saml" mapstructure:"saml"`

	// LDAP Configuration
	LDAP UILDAPConfig `yaml:"ldap" json:"ldap" mapstructure:"ldap"`

	// Multi-factor Authentication
	MFA UIMFAConfig `yaml:"mfa" json:"mfa" mapstructure:"mfa"`

	// Password Policy
	PasswordPolicy UIPasswordPolicyConfig `yaml:"password_policy" json:"password_policy" mapstructure:"password_policy"`
}

// UIOAuthConfig contains OAuth configuration
type UIOAuthConfig struct {
	Enabled     bool                  `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Providers   map[string]UIProvider `yaml:"providers" json:"providers" mapstructure:"providers"`
	RedirectURL string                `yaml:"redirect_url" json:"redirect_url" mapstructure:"redirect_url"`
}

// UIProvider contains OAuth provider configuration
type UIProvider struct {
	ClientID     string   `yaml:"client_id" json:"client_id" mapstructure:"client_id"`
	ClientSecret string   `yaml:"client_secret" json:"client_secret" mapstructure:"client_secret"`
	AuthURL      string   `yaml:"auth_url" json:"auth_url" mapstructure:"auth_url"`
	TokenURL     string   `yaml:"token_url" json:"token_url" mapstructure:"token_url"`
	UserInfoURL  string   `yaml:"user_info_url" json:"user_info_url" mapstructure:"user_info_url"`
	Scopes       []string `yaml:"scopes" json:"scopes" mapstructure:"scopes"`
}

// UISAMLConfig contains SAML configuration
type UISAMLConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	MetadataURL string `yaml:"metadata_url" json:"metadata_url" mapstructure:"metadata_url"`
	EntityID    string `yaml:"entity_id" json:"entity_id" mapstructure:"entity_id"`
	ACSURL      string `yaml:"acs_url" json:"acs_url" mapstructure:"acs_url"`
	CertPath    string `yaml:"cert_path" json:"cert_path" mapstructure:"cert_path"`
	KeyPath     string `yaml:"key_path" json:"key_path" mapstructure:"key_path"`
}

// UILDAPConfig contains LDAP configuration
type UILDAPConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Host       string `yaml:"host" json:"host" mapstructure:"host"`
	Port       int    `yaml:"port" json:"port" mapstructure:"port"`
	UseTLS     bool   `yaml:"use_tls" json:"use_tls" mapstructure:"use_tls"`
	BaseDN     string `yaml:"base_dn" json:"base_dn" mapstructure:"base_dn"`
	BindDN     string `yaml:"bind_dn" json:"bind_dn" mapstructure:"bind_dn"`
	BindPasswd string `yaml:"bind_passwd" json:"bind_passwd" mapstructure:"bind_passwd"`
	UserFilter string `yaml:"user_filter" json:"user_filter" mapstructure:"user_filter"`
}

// UIMFAConfig contains multi-factor authentication configuration
type UIMFAConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Required    bool     `yaml:"required" json:"required" mapstructure:"required"`
	Methods     []string `yaml:"methods" json:"methods" mapstructure:"methods"` // "totp", "sms", "email"
	TOTPIssuer  string   `yaml:"totp_issuer" json:"totp_issuer" mapstructure:"totp_issuer"`
	SMSProvider string   `yaml:"sms_provider" json:"sms_provider" mapstructure:"sms_provider"`
}

// UIPasswordPolicyConfig contains password policy configuration
type UIPasswordPolicyConfig struct {
	MinLength      int  `yaml:"min_length" json:"min_length" mapstructure:"min_length"`
	RequireUpper   bool `yaml:"require_upper" json:"require_upper" mapstructure:"require_upper"`
	RequireLower   bool `yaml:"require_lower" json:"require_lower" mapstructure:"require_lower"`
	RequireDigit   bool `yaml:"require_digit" json:"require_digit" mapstructure:"require_digit"`
	RequireSpecial bool `yaml:"require_special" json:"require_special" mapstructure:"require_special"`
	MaxAge         int  `yaml:"max_age" json:"max_age" mapstructure:"max_age"` // days
	HistoryCount   int  `yaml:"history_count" json:"history_count" mapstructure:"history_count"`
}

// UIEncryptionConfig contains encryption configuration
type UIEncryptionConfig struct {
	// Data Encryption
	DataKey       string `yaml:"data_key" json:"data_key" mapstructure:"data_key"`
	KeyDerivation string `yaml:"key_derivation" json:"key_derivation" mapstructure:"key_derivation"`

	// Transport Encryption
	ForceHTTPS bool `yaml:"force_https" json:"force_https" mapstructure:"force_https"`
	HSTSMaxAge int  `yaml:"hsts_max_age" json:"hsts_max_age" mapstructure:"hsts_max_age"`
}

// UISessionConfig contains UI-specific session management configuration
// Note: Redis configuration is inherited from main RedisConfig
type UISessionConfig struct {
	// Cookie Configuration
	Cookie UISessionCookieConfig `yaml:"cookie" json:"cookie" mapstructure:"cookie"`

	// UI-specific Session Lifecycle
	TTL               time.Duration `yaml:"ttl" json:"ttl" mapstructure:"ttl"`
	RefreshTTL        time.Duration `yaml:"refresh_ttl" json:"refresh_ttl" mapstructure:"refresh_ttl"`
	InactivityTimeout time.Duration `yaml:"inactivity_timeout" json:"inactivity_timeout" mapstructure:"inactivity_timeout"`
	MaxConcurrent     int           `yaml:"max_concurrent" json:"max_concurrent" mapstructure:"max_concurrent"`

	// UI Security
	Regenerate bool `yaml:"regenerate" json:"regenerate" mapstructure:"regenerate"`
	CSRFToken  bool `yaml:"csrf_token" json:"csrf_token" mapstructure:"csrf_token"`
}

// UISessionCookieConfig contains session cookie configuration
type UISessionCookieConfig struct {
	Name     string `yaml:"name" json:"name" mapstructure:"name"`
	Domain   string `yaml:"domain" json:"domain" mapstructure:"domain"`
	Path     string `yaml:"path" json:"path" mapstructure:"path"`
	Secure   bool   `yaml:"secure" json:"secure" mapstructure:"secure"`
	HttpOnly bool   `yaml:"http_only" json:"http_only" mapstructure:"http_only"`
	SameSite string `yaml:"same_site" json:"same_site" mapstructure:"same_site"`
}

// UIAssetsConfig contains asset management configuration
type UIAssetsConfig struct {
	// Asset Serving
	StaticURL        string        `yaml:"static_url" json:"static_url" mapstructure:"static_url"`
	CacheMaxAge      time.Duration `yaml:"cache_max_age" json:"cache_max_age" mapstructure:"cache_max_age"`
	CompressionLevel int           `yaml:"compression_level" json:"compression_level" mapstructure:"compression_level"`

	// CDN Configuration
	CDN UIAssetsCDNConfig `yaml:"cdn" json:"cdn" mapstructure:"cdn"`

	// Build Configuration
	Build UIAssetsBuildConfig `yaml:"build" json:"build" mapstructure:"build"`
}

// UIAssetsCDNConfig contains CDN configuration
type UIAssetsCDNConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	BaseURL   string `yaml:"base_url" json:"base_url" mapstructure:"base_url"`
	KeyID     string `yaml:"key_id" json:"key_id" mapstructure:"key_id"`
	SecretKey string `yaml:"secret_key" json:"secret_key" mapstructure:"secret_key"`
}

// UIAssetsBuildConfig contains build configuration
type UIAssetsBuildConfig struct {
	SourcePath string `yaml:"source_path" json:"source_path" mapstructure:"source_path"`
	BuildPath  string `yaml:"build_path" json:"build_path" mapstructure:"build_path"`
	Minify     bool   `yaml:"minify" json:"minify" mapstructure:"minify"`
	Sourcemaps bool   `yaml:"sourcemaps" json:"sourcemaps" mapstructure:"sourcemaps"`
}

// UIDevelopmentConfig contains development-specific configuration
type UIDevelopmentConfig struct {
	// Development Features
	Enabled   bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	HotReload bool `yaml:"hot_reload" json:"hot_reload" mapstructure:"hot_reload"`
	MockData  bool `yaml:"mock_data" json:"mock_data" mapstructure:"mock_data"`
	DebugMode bool `yaml:"debug_mode" json:"debug_mode" mapstructure:"debug_mode"`

	// Logging
	LogLevel     string `yaml:"log_level" json:"log_level" mapstructure:"log_level"`
	LogRequests  bool   `yaml:"log_requests" json:"log_requests" mapstructure:"log_requests"`
	LogResponses bool   `yaml:"log_responses" json:"log_responses" mapstructure:"log_responses"`

	// Development Tools
	Profiling    bool   `yaml:"profiling" json:"profiling" mapstructure:"profiling"`
	Metrics      bool   `yaml:"metrics" json:"metrics" mapstructure:"metrics"`
	DevToolsPort string `yaml:"dev_tools_port" json:"dev_tools_port" mapstructure:"dev_tools_port"`
}

// SetUIDefaults sets default values for UI configuration
func SetUIDefaults(v *viper.Viper) {
	// UI Server - only UI-specific settings (inherits main server config)
	v.SetDefault("ui.server.idle_timeout", 120*time.Second)
	v.SetDefault("ui.server.multi_port.enabled", false)
	v.SetDefault("ui.server.multi_port.console_port", "8081")
	v.SetDefault("ui.server.multi_port.workspace_port", "8082")
	v.SetDefault("ui.server.multi_port.portal_port", "8083")

	v.SetDefault("ui.console.name", "Admin Console")
	v.SetDefault("ui.console.description", "Administrative interface for system management")
	v.SetDefault("ui.console.version", "1.0.0")
	v.SetDefault("ui.console.enabled", true)
	v.SetDefault("ui.console.features.dark_mode", true)
	v.SetDefault("ui.console.features.notifications", true)
	v.SetDefault("ui.console.features.real_time_update", false)
	v.SetDefault("ui.console.features.bulk_operations", true)
	v.SetDefault("ui.console.features.export", true)
	v.SetDefault("ui.console.features.search", true)
	v.SetDefault("ui.console.features.multi_language", false)

	v.SetDefault("ui.workspace.name", "Tenant Workspace")
	v.SetDefault("ui.workspace.description", "Collaborative workspace for tenant users")
	v.SetDefault("ui.workspace.version", "1.0.0")
	v.SetDefault("ui.workspace.enabled", true)
	v.SetDefault("ui.workspace.features.dark_mode", true)
	v.SetDefault("ui.workspace.features.notifications", true)
	v.SetDefault("ui.workspace.features.real_time_update", true)
	v.SetDefault("ui.workspace.features.bulk_operations", false)
	v.SetDefault("ui.workspace.features.export", true)
	v.SetDefault("ui.workspace.features.search", true)
	v.SetDefault("ui.workspace.features.multi_language", false)

	v.SetDefault("ui.portal.name", "Client Portal")
	v.SetDefault("ui.portal.description", "Self-service portal for end customers")
	v.SetDefault("ui.portal.version", "1.0.0")
	v.SetDefault("ui.portal.enabled", true)
	v.SetDefault("ui.portal.features.dark_mode", true)
	v.SetDefault("ui.portal.features.notifications", true)
	v.SetDefault("ui.portal.features.real_time_update", false)
	v.SetDefault("ui.portal.features.bulk_operations", false)
	v.SetDefault("ui.portal.features.export", false)
	v.SetDefault("ui.portal.features.search", true)
	v.SetDefault("ui.portal.features.multi_language", false)

	// Security - only UI-specific JWT settings (inherits main auth config)
	v.SetDefault("ui.security.jwt.audience", "awo-erp-ui")
	v.SetDefault("ui.security.jwt.access_token_ttl", 15*time.Minute)
	v.SetDefault("ui.security.jwt.refresh_token_ttl", 24*time.Hour)

	v.SetDefault("ui.security.cors.enabled", true)
	v.SetDefault("ui.security.cors.allow_credentials", true)
	v.SetDefault("ui.security.cors.max_age", 86400)

	v.SetDefault("ui.security.csp.enabled", true)
	v.SetDefault("ui.security.csp.default_src", "'self'")
	v.SetDefault("ui.security.csp.script_src", "'self' 'unsafe-inline' https://cdn.tailwindcss.com")
	v.SetDefault("ui.security.csp.style_src", "'self' 'unsafe-inline' https://cdn.tailwindcss.com")
	v.SetDefault("ui.security.csp.img_src", "'self' data: https:")
	v.SetDefault("ui.security.csp.font_src", "'self' https://fonts.gstatic.com")
	v.SetDefault("ui.security.csp.connect_src", "'self'")
	v.SetDefault("ui.security.csp.report_only", false)

	v.SetDefault("ui.security.auth.oauth.enabled", false)
	v.SetDefault("ui.security.auth.saml.enabled", false)
	v.SetDefault("ui.security.auth.ldap.enabled", false)
	v.SetDefault("ui.security.auth.ldap.port", 389)
	v.SetDefault("ui.security.auth.ldap.use_tls", false)
	v.SetDefault("ui.security.auth.ldap.user_filter", "(uid=%s)")
	v.SetDefault("ui.security.auth.mfa.enabled", false)
	v.SetDefault("ui.security.auth.mfa.required", false)
	v.SetDefault("ui.security.auth.mfa.totp_issuer", "Awo ERP")
	v.SetDefault("ui.security.auth.password_policy.min_length", 8)
	v.SetDefault("ui.security.auth.password_policy.require_upper", true)
	v.SetDefault("ui.security.auth.password_policy.require_lower", true)
	v.SetDefault("ui.security.auth.password_policy.require_digit", true)
	v.SetDefault("ui.security.auth.password_policy.require_special", true)
	v.SetDefault("ui.security.auth.password_policy.max_age", 90)
	v.SetDefault("ui.security.auth.password_policy.history_count", 5)

	v.SetDefault("ui.security.encryption.force_https", true)
	v.SetDefault("ui.security.encryption.hsts_max_age", 31536000)

	// Session - UI-specific settings (Redis config inherited)
	v.SetDefault("ui.session.cookie.name", "awo_session")
	v.SetDefault("ui.session.cookie.path", "/")
	v.SetDefault("ui.session.cookie.secure", true)
	v.SetDefault("ui.session.cookie.http_only", true)
	v.SetDefault("ui.session.cookie.same_site", "Lax")
	v.SetDefault("ui.session.ttl", 8*time.Hour)
	v.SetDefault("ui.session.refresh_ttl", 30*24*time.Hour)
	v.SetDefault("ui.session.inactivity_timeout", 2*time.Hour)
	v.SetDefault("ui.session.max_concurrent", 3)
	v.SetDefault("ui.session.regenerate", true)
	v.SetDefault("ui.session.csrf_token", true)

	v.SetDefault("ui.assets.static_url", "/static")
	v.SetDefault("ui.assets.cache_max_age", 24*time.Hour)
	v.SetDefault("ui.assets.compression_level", 6)
	v.SetDefault("ui.assets.cdn.enabled", false)
	v.SetDefault("ui.assets.build.source_path", "./web/src")
	v.SetDefault("ui.assets.build.build_path", "./internal/ui/assets")
	v.SetDefault("ui.assets.build.minify", true)
	v.SetDefault("ui.assets.build.sourcemaps", false)

	v.SetDefault("ui.development.enabled", false)
	v.SetDefault("ui.development.hot_reload", false)
	v.SetDefault("ui.development.mock_data", false)
	v.SetDefault("ui.development.debug_mode", false)
	v.SetDefault("ui.development.log_level", "info")
	v.SetDefault("ui.development.log_requests", true)
	v.SetDefault("ui.development.log_responses", false)
	v.SetDefault("ui.development.profiling", false)
	v.SetDefault("ui.development.metrics", true)
	v.SetDefault("ui.development.dev_tools_port", "9090")
}

// BindUIEnvVars binds environment variables for UI configuration
func BindUIEnvVars(v *viper.Viper) {
	// UI-specific environment variables only (avoid duplicating main config vars)
	v.BindEnv("ui.server.multi_port.enabled", "UI_MULTI_PORT")
	v.BindEnv("ui.server.multi_port.console_port", "UI_CONSOLE_PORT")
	v.BindEnv("ui.server.multi_port.workspace_port", "UI_WORKSPACE_PORT")
	v.BindEnv("ui.server.multi_port.portal_port", "UI_PORTAL_PORT")
	v.BindEnv("ui.security.jwt.audience", "UI_JWT_AUDIENCE")
	v.BindEnv("ui.security.auth.oauth.providers.google.client_id", "OAUTH_CLIENT_ID") // Example for one provider
	v.BindEnv("ui.security.auth.oauth.providers.google.client_secret", "OAUTH_CLIENT_SECRET")
	v.BindEnv("ui.security.auth.ldap.host", "LDAP_HOST")
	v.BindEnv("ui.security.auth.ldap.port", "LDAP_PORT")
	v.BindEnv("ui.security.auth.ldap.base_dn", "LDAP_BASE_DN")
	v.BindEnv("ui.security.auth.ldap.bind_dn", "LDAP_BIND_DN")
	v.BindEnv("ui.security.auth.ldap.bind_passwd", "LDAP_BIND_PASSWD")
	v.BindEnv("ui.security.encryption.data_key", "UI_DATA_KEY")
	v.BindEnv("ui.assets.cdn.key_id", "CDN_KEY_ID")
	v.BindEnv("ui.assets.cdn.secret_key", "CDN_SECRET_KEY")
	v.BindEnv("ui.development.enabled", "UI_DEV_MODE")
}

// Validate validates the UI configuration
func (c *UIConfig) Validate() error {
	// TODO:Add validation logic here
	return nil
}

// GetEffectiveBaseURL returns the effective base URL for a service
func (c *UIServiceConfig) GetEffectiveBaseURL() string {
	if c.ExternalURL != "" {
		return c.ExternalURL
	}
	return c.BaseURL
}

// GetRedisConfig returns the Redis configuration from main config for UI usage
// This demonstrates how UI config leverages shared configurations
func (c *UIConfig) GetRedisConfig(mainConfig *Config) *RedisConfig {
	return &mainConfig.Redis
}

// GetServerConfig returns the main server configuration for UI inheritance
func (c *UIConfig) GetServerConfig(mainConfig *Config) *ServerConfig {
	return &mainConfig.Server
}

// GetAuthConfig returns the main auth configuration for UI inheritance
func (c *UIConfig) GetAuthConfig(mainConfig *Config) *AuthConfig {
	return &mainConfig.Auth
}
