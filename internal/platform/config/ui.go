package config

import (
	"crypto/rsa"
	"time"
)

// UIConfig contains all UI-related configuration
type UIConfig struct {
	// Server Configuration
	Server UIServerConfig `yaml:"server" json:"server"`

	// Service-specific configurations
	Console   UIServiceConfig `yaml:"console" json:"console"`
	Workspace UIServiceConfig `yaml:"workspace" json:"workspace"`
	Portal    UIServiceConfig `yaml:"portal" json:"portal"`

	// Security Configuration
	Security UISecurityConfig `yaml:"security" json:"security"`

	// Session Management
	Session UISessionConfig `yaml:"session" json:"session"`

	// Asset Management
	Assets UIAssetsConfig `yaml:"assets" json:"assets"`

	// Development Settings
	Development UIDevelopmentConfig `yaml:"development" json:"development"`
}

// UIServerConfig contains server-specific configuration
type UIServerConfig struct {
	// Single port strategy
	Port         string        `yaml:"port" json:"port" env:"UI_PORT" default:"8080"`
	Host         string        `yaml:"host" json:"host" env:"UI_HOST" default:"0.0.0.0"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" default:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout" default:"120s"`

	// Multi-port strategy (alternative)
	MultiPort UIMultiPortConfig `yaml:"multi_port" json:"multi_port"`

	// TLS Configuration
	TLS UITLSConfig `yaml:"tls" json:"tls"`
}

// UIMultiPortConfig for multi-port deployment strategy
type UIMultiPortConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled" env:"UI_MULTI_PORT" default:"false"`
	ConsolePort   string `yaml:"console_port" json:"console_port" env:"UI_CONSOLE_PORT" default:"8081"`
	WorkspacePort string `yaml:"workspace_port" json:"workspace_port" env:"UI_WORKSPACE_PORT" default:"8082"`
	PortalPort    string `yaml:"portal_port" json:"portal_port" env:"UI_PORTAL_PORT" default:"8083"`
}

// UITLSConfig contains TLS/SSL configuration
type UITLSConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled" env:"UI_TLS_ENABLED" default:"false"`
	CertFile   string `yaml:"cert_file" json:"cert_file" env:"UI_TLS_CERT_FILE"`
	KeyFile    string `yaml:"key_file" json:"key_file" env:"UI_TLS_KEY_FILE"`
	MinVersion string `yaml:"min_version" json:"min_version" default:"1.2"`
}

// UIServiceConfig contains configuration for individual UI services
type UIServiceConfig struct {
	// Service Identity
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Version     string `yaml:"version" json:"version" default:"1.0.0"`

	// Access Configuration
	Enabled     bool     `yaml:"enabled" json:"enabled" default:"true"`
	BaseURL     string   `yaml:"base_url" json:"base_url"`
	ExternalURL string   `yaml:"external_url" json:"external_url"`
	AllowedIPs  []string `yaml:"allowed_ips" json:"allowed_ips"`

	// Feature Flags
	Features UIServiceFeatures `yaml:"features" json:"features"`

	// Branding
	Branding UIBrandingConfig `yaml:"branding" json:"branding"`

	// Rate Limiting
	RateLimit UIRateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

// UIServiceFeatures contains feature flags for UI services
type UIServiceFeatures struct {
	DarkMode       bool `yaml:"dark_mode" json:"dark_mode" default:"true"`
	Notifications  bool `yaml:"notifications" json:"notifications" default:"true"`
	RealTimeUpdate bool `yaml:"real_time_update" json:"real_time_update" default:"false"`
	BulkOperations bool `yaml:"bulk_operations" json:"bulk_operations" default:"false"`
	Export         bool `yaml:"export" json:"export" default:"true"`
	Search         bool `yaml:"search" json:"search" default:"true"`
	MultiLanguage  bool `yaml:"multi_language" json:"multi_language" default:"false"`
}

// UIBrandingConfig contains branding configuration
type UIBrandingConfig struct {
	CompanyName    string `yaml:"company_name" json:"company_name" default:"Awo ERP"`
	LogoURL        string `yaml:"logo_url" json:"logo_url"`
	FaviconURL     string `yaml:"favicon_url" json:"favicon_url"`
	PrimaryColor   string `yaml:"primary_color" json:"primary_color" default:"#1e40af"`
	SecondaryColor string `yaml:"secondary_color" json:"secondary_color" default:"#6b7280"`
	CustomCSS      string `yaml:"custom_css" json:"custom_css"`
}

// UIRateLimitConfig contains rate limiting configuration
type UIRateLimitConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled" default:"true"`
	Requests   int           `yaml:"requests" json:"requests" default:"100"`
	Window     time.Duration `yaml:"window" json:"window" default:"1m"`
	BurstLimit int           `yaml:"burst_limit" json:"burst_limit" default:"10"`
	SkipPaths  []string      `yaml:"skip_paths" json:"skip_paths"`
}

// UISecurityConfig contains security-related configuration
type UISecurityConfig struct {
	// JWT Configuration
	JWT UIJWTConfig `yaml:"jwt" json:"jwt"`

	// CORS Configuration
	CORS UICORSConfig `yaml:"cors" json:"cors"`

	// CSP Configuration
	CSP UICSPConfig `yaml:"csp" json:"csp"`

	// Authentication
	Auth UIAuthConfig `yaml:"auth" json:"auth"`

	// Encryption
	Encryption UIEncryptionConfig `yaml:"encryption" json:"encryption"`
}

// UIJWTConfig contains JWT-related configuration
type UIJWTConfig struct {
	// Signing Configuration
	Algorithm      string          `yaml:"algorithm" json:"algorithm" default:"RS256"`
	PublicKeyPath  string          `yaml:"public_key_path" json:"public_key_path" env:"JWT_PUBLIC_KEY_PATH"`
	PrivateKeyPath string          `yaml:"private_key_path" json:"private_key_path" env:"JWT_PRIVATE_KEY_PATH"`
	PublicKey      *rsa.PublicKey  `yaml:"-" json:"-"` // Loaded at runtime
	PrivateKey     *rsa.PrivateKey `yaml:"-" json:"-"` // Loaded at runtime

	// Token Configuration
	Issuer          string        `yaml:"issuer" json:"issuer" env:"JWT_ISSUER" default:"awo-erp"`
	Audience        string        `yaml:"audience" json:"audience" env:"JWT_AUDIENCE" default:"awo-erp-ui"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" json:"access_token_ttl" default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" json:"refresh_token_ttl" default:"24h"`

	// Validation Configuration
	ClockSkew        time.Duration `yaml:"clock_skew" json:"clock_skew" default:"5m"`
	ValidateAudience bool          `yaml:"validate_audience" json:"validate_audience" default:"true"`
	ValidateIssuer   bool          `yaml:"validate_issuer" json:"validate_issuer" default:"true"`
}

// UICORSConfig contains CORS configuration
type UICORSConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled" default:"true"`
	AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" json:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials" default:"true"`
	MaxAge           int      `yaml:"max_age" json:"max_age" default:"86400"`
}

// UICSPConfig contains Content Security Policy configuration
type UICSPConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled" default:"true"`
	DefaultSrc string `yaml:"default_src" json:"default_src" default:"'self'"`
	ScriptSrc  string `yaml:"script_src" json:"script_src" default:"'self' 'unsafe-inline' https://cdn.tailwindcss.com"`
	StyleSrc   string `yaml:"style_src" json:"style_src" default:"'self' 'unsafe-inline' https://cdn.tailwindcss.com"`
	ImgSrc     string `yaml:"img_src" json:"img_src" default:"'self' data: https:"`
	FontSrc    string `yaml:"font_src" json:"font_src" default:"'self' https://fonts.gstatic.com"`
	ConnectSrc string `yaml:"connect_src" json:"connect_src" default:"'self'"`
	ReportURI  string `yaml:"report_uri" json:"report_uri"`
	ReportOnly bool   `yaml:"report_only" json:"report_only" default:"false"`
}

// UIAuthConfig contains authentication configuration
type UIAuthConfig struct {
	// OAuth Configuration
	OAuth UIOAuthConfig `yaml:"oauth" json:"oauth"`

	// SAML Configuration
	SAML UISAMLConfig `yaml:"saml" json:"saml"`

	// LDAP Configuration
	LDAP UILDAPConfig `yaml:"ldap" json:"ldap"`

	// Multi-factor Authentication
	MFA UIMFAConfig `yaml:"mfa" json:"mfa"`

	// Password Policy
	PasswordPolicy UIPasswordPolicyConfig `yaml:"password_policy" json:"password_policy"`
}

// UIOAuthConfig contains OAuth configuration
type UIOAuthConfig struct {
	Enabled     bool                  `yaml:"enabled" json:"enabled" default:"false"`
	Providers   map[string]UIProvider `yaml:"providers" json:"providers"`
	RedirectURL string                `yaml:"redirect_url" json:"redirect_url"`
}

// UIProvider contains OAuth provider configuration
type UIProvider struct {
	ClientID     string   `yaml:"client_id" json:"client_id" env:"OAUTH_CLIENT_ID"`
	ClientSecret string   `yaml:"client_secret" json:"client_secret" env:"OAUTH_CLIENT_SECRET"`
	AuthURL      string   `yaml:"auth_url" json:"auth_url"`
	TokenURL     string   `yaml:"token_url" json:"token_url"`
	UserInfoURL  string   `yaml:"user_info_url" json:"user_info_url"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
}

// UISAMLConfig contains SAML configuration
type UISAMLConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled" default:"false"`
	MetadataURL string `yaml:"metadata_url" json:"metadata_url"`
	EntityID    string `yaml:"entity_id" json:"entity_id"`
	ACSURL      string `yaml:"acs_url" json:"acs_url"`
	CertPath    string `yaml:"cert_path" json:"cert_path"`
	KeyPath     string `yaml:"key_path" json:"key_path"`
}

// UILDAPConfig contains LDAP configuration
type UILDAPConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled" default:"false"`
	Host       string `yaml:"host" json:"host" env:"LDAP_HOST"`
	Port       int    `yaml:"port" json:"port" env:"LDAP_PORT" default:"389"`
	UseTLS     bool   `yaml:"use_tls" json:"use_tls" default:"false"`
	BaseDN     string `yaml:"base_dn" json:"base_dn" env:"LDAP_BASE_DN"`
	BindDN     string `yaml:"bind_dn" json:"bind_dn" env:"LDAP_BIND_DN"`
	BindPasswd string `yaml:"bind_passwd" json:"bind_passwd" env:"LDAP_BIND_PASSWD"`
	UserFilter string `yaml:"user_filter" json:"user_filter" default:"(uid=%s)"`
}

// UIMFAConfig contains multi-factor authentication configuration
type UIMFAConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled" default:"false"`
	Required    bool     `yaml:"required" json:"required" default:"false"`
	Methods     []string `yaml:"methods" json:"methods"` // "totp", "sms", "email"
	TOTPIssuer  string   `yaml:"totp_issuer" json:"totp_issuer" default:"Awo ERP"`
	SMSProvider string   `yaml:"sms_provider" json:"sms_provider"`
}

// UIPasswordPolicyConfig contains password policy configuration
type UIPasswordPolicyConfig struct {
	MinLength      int  `yaml:"min_length" json:"min_length" default:"8"`
	RequireUpper   bool `yaml:"require_upper" json:"require_upper" default:"true"`
	RequireLower   bool `yaml:"require_lower" json:"require_lower" default:"true"`
	RequireDigit   bool `yaml:"require_digit" json:"require_digit" default:"true"`
	RequireSpecial bool `yaml:"require_special" json:"require_special" default:"true"`
	MaxAge         int  `yaml:"max_age" json:"max_age" default:"90"` // days
	HistoryCount   int  `yaml:"history_count" json:"history_count" default:"5"`
}

// UIEncryptionConfig contains encryption configuration
type UIEncryptionConfig struct {
	// Data Encryption
	DataKey       string `yaml:"data_key" json:"data_key" env:"UI_DATA_KEY"`
	KeyDerivation string `yaml:"key_derivation" json:"key_derivation" default:"pbkdf2"`

	// Transport Encryption
	ForceHTTPS bool `yaml:"force_https" json:"force_https" default:"true"`
	HSTSMaxAge int  `yaml:"hsts_max_age" json:"hsts_max_age" default:"31536000"`
}

// UISessionConfig contains session management configuration
type UISessionConfig struct {
	// Redis Configuration
	Store     string `yaml:"store" json:"store" default:"redis"`
	RedisURL  string `yaml:"redis_url" json:"redis_url" env:"REDIS_URL"`
	RedisDB   int    `yaml:"redis_db" json:"redis_db" default:"0"`
	KeyPrefix string `yaml:"key_prefix" json:"key_prefix" default:"ui:session:"`

	// Cookie Configuration
	Cookie UISessionCookieConfig `yaml:"cookie" json:"cookie"`

	// Session Lifecycle
	TTL               time.Duration `yaml:"ttl" json:"ttl" default:"8h"`
	RefreshTTL        time.Duration `yaml:"refresh_ttl" json:"refresh_ttl" default:"30d"`
	InactivityTimeout time.Duration `yaml:"inactivity_timeout" json:"inactivity_timeout" default:"2h"`
	MaxConcurrent     int           `yaml:"max_concurrent" json:"max_concurrent" default:"3"`

	// Security
	Regenerate bool `yaml:"regenerate" json:"regenerate" default:"true"`
	CSRFToken  bool `yaml:"csrf_token" json:"csrf_token" default:"true"`
}

// UISessionCookieConfig contains session cookie configuration
type UISessionCookieConfig struct {
	Name     string `yaml:"name" json:"name" default:"awo_session"`
	Domain   string `yaml:"domain" json:"domain"`
	Path     string `yaml:"path" json:"path" default:"/"`
	Secure   bool   `yaml:"secure" json:"secure" default:"true"`
	HttpOnly bool   `yaml:"http_only" json:"http_only" default:"true"`
	SameSite string `yaml:"same_site" json:"same_site" default:"Lax"`
}

// UIAssetsConfig contains asset management configuration
type UIAssetsConfig struct {
	// Asset Serving
	StaticURL        string        `yaml:"static_url" json:"static_url" default:"/static"`
	CacheMaxAge      time.Duration `yaml:"cache_max_age" json:"cache_max_age" default:"24h"`
	CompressionLevel int           `yaml:"compression_level" json:"compression_level" default:"6"`

	// CDN Configuration
	CDN UIAssetsCDNConfig `yaml:"cdn" json:"cdn"`

	// Build Configuration
	Build UIAssetsBuildConfig `yaml:"build" json:"build"`
}

// UIAssetsCDNConfig contains CDN configuration
type UIAssetsCDNConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled" default:"false"`
	BaseURL   string `yaml:"base_url" json:"base_url"`
	KeyID     string `yaml:"key_id" json:"key_id" env:"CDN_KEY_ID"`
	SecretKey string `yaml:"secret_key" json:"secret_key" env:"CDN_SECRET_KEY"`
}

// UIAssetsBuildConfig contains build configuration
type UIAssetsBuildConfig struct {
	SourcePath string `yaml:"source_path" json:"source_path" default:"./web/src"`
	BuildPath  string `yaml:"build_path" json:"build_path" default:"./internal/ui/assets"`
	Minify     bool   `yaml:"minify" json:"minify" default:"true"`
	Sourcemaps bool   `yaml:"sourcemaps" json:"sourcemaps" default:"false"`
}

// UIDevelopmentConfig contains development-specific configuration
type UIDevelopmentConfig struct {
	// Development Features
	Enabled   bool `yaml:"enabled" json:"enabled" env:"DEV_MODE" default:"false"`
	HotReload bool `yaml:"hot_reload" json:"hot_reload" default:"false"`
	MockData  bool `yaml:"mock_data" json:"mock_data" default:"false"`
	DebugMode bool `yaml:"debug_mode" json:"debug_mode" default:"false"`

	// Logging
	LogLevel     string `yaml:"log_level" json:"log_level" default:"info"`
	LogRequests  bool   `yaml:"log_requests" json:"log_requests" default:"true"`
	LogResponses bool   `yaml:"log_responses" json:"log_responses" default:"false"`

	// Development Tools
	Profiling    bool   `yaml:"profiling" json:"profiling" default:"false"`
	Metrics      bool   `yaml:"metrics" json:"metrics" default:"true"`
	DevToolsPort string `yaml:"dev_tools_port" json:"dev_tools_port" default:"9090"`
}

// DefaultUIConfig returns a default UI configuration
func DefaultUIConfig() *UIConfig {
	return &UIConfig{
		Server: UIServerConfig{
			Port:         "8080",
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Console: UIServiceConfig{
			Name:        "Admin Console",
			Description: "Administrative interface for system management",
			Enabled:     true,
			Features: UIServiceFeatures{
				DarkMode:       true,
				Notifications:  true,
				BulkOperations: true,
				Export:         true,
				Search:         true,
			},
		},
		Workspace: UIServiceConfig{
			Name:        "Tenant Workspace",
			Description: "Collaborative workspace for tenant users",
			Enabled:     true,
			Features: UIServiceFeatures{
				DarkMode:       true,
				Notifications:  true,
				RealTimeUpdate: true,
				Export:         true,
				Search:         true,
			},
		},
		Portal: UIServiceConfig{
			Name:        "Client Portal",
			Description: "Self-service portal for end customers",
			Enabled:     true,
			Features: UIServiceFeatures{
				DarkMode:      true,
				Notifications: true,
				Export:        false,
				Search:        true,
			},
		},
		Security: UISecurityConfig{
			JWT: UIJWTConfig{
				Algorithm:       "RS256",
				Issuer:          "awo-erp",
				Audience:        "awo-erp-ui",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 24 * time.Hour,
				ClockSkew:       5 * time.Minute,
			},
		},
		Session: UISessionConfig{
			Store:             "redis",
			TTL:               8 * time.Hour,
			RefreshTTL:        30 * 24 * time.Hour,
			InactivityTimeout: 2 * time.Hour,
			MaxConcurrent:     3,
		},
	}
}

// Validate validates the UI configuration
func (c *UIConfig) Validate() error {
	// Add validation logic here
	return nil
}

// GetEffectiveBaseURL returns the effective base URL for a service
func (c *UIServiceConfig) GetEffectiveBaseURL() string {
	if c.ExternalURL != "" {
		return c.ExternalURL
	}
	return c.BaseURL
}
