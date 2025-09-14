package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// AppStage represents application deployment stage
type AppStage string

const (
	DevelopmentStage AppStage = "dev"
	StagingStage     AppStage = "staging"
	ProductionStage  AppStage = "production"
	TestingStage     AppStage = "testing"
)

// String returns the string representation of AppStage
func (a AppStage) String() string {
	return string(a)
}

// IsProduction returns true if the stage is production
func (a AppStage) IsProduction() bool {
	return a == ProductionStage
}

// IsDevelopment returns true if the stage is development
func (a AppStage) IsDevelopment() bool {
	return a == DevelopmentStage
}

// IsStaging returns true if the stage is staging
func (a AppStage) IsStaging() bool {
	return a == StagingStage
}

// IsTesting returns true if the stage is testing
func (a AppStage) IsTesting() bool {
	return a == TestingStage
}

// AppConfig represents application-level configuration
type AppConfig struct {
	Name        string   `yaml:"name" mapstructure:"name"`               // Application name
	Version     string   `yaml:"version" mapstructure:"version"`         // Application version
	Stage       AppStage `yaml:"stage" mapstructure:"stage"`             // Application stage (development, staging, production, testing)
	Debug       bool     `yaml:"debug" mapstructure:"debug"`             // Enable debug mode
	Environment string   `yaml:"environment" mapstructure:"environment"` // Custom environment identifier
	Namespace   string   `yaml:"namespace" mapstructure:"namespace"`     // Kubernetes namespace or deployment namespace
}

// Config represents application configuration
type Config struct {
	App      AppConfig      `yaml:"app" mapstructure:"app"`
	Server   ServerConfig   `yaml:"server" mapstructure:"server"`
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
	Redis    RedisConfig    `yaml:"redis" mapstructure:"redis"`
	Temporal TemporalConfig `yaml:"temporal" mapstructure:"temporal"`
	Auth     AuthConfig     `yaml:"auth" mapstructure:"auth"`
	Features FeatureConfig  `yaml:"features" mapstructure:"features"`
	Logger   LoggerConfig   `yaml:"logger" mapstructure:"logger"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port         string        `yaml:"port" mapstructure:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	GRPCPort     string        `yaml:"grpc_port" mapstructure:"grpc_port"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" mapstructure:"host"`
	Port            int           `yaml:"port" mapstructure:"port"`
	User            string        `yaml:"user" mapstructure:"user"`
	Password        string        `yaml:"password" mapstructure:"password"`
	Database        string        `yaml:"database" mapstructure:"database"`
	SSLMode         string        `yaml:"ssl_mode" mapstructure:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
}

// GetDatabaseURL returns the PostgreSQL connection URL
func (d *DatabaseConfig) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

// TemporalConfig is now defined in temporal.go for system-wide use

// AuthConfig represents auth configuration
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret" mapstructure:"jwt_secret"`
}

// FeatureConfig represents feature flags
type FeatureConfig struct {
	EnableNewDashboard bool `yaml:"enable_new_dashboard" mapstructure:"enable_new_dashboard"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Type        string `yaml:"type" mapstructure:"type"`                 // "zap", "zerolog", "slog"
	Level       string `yaml:"level" mapstructure:"level"`               // "debug", "info", "warn", "error", "fatal"
	Format      string `yaml:"format" mapstructure:"format"`             // "json", "text", "console"
	Development bool   `yaml:"development" mapstructure:"development"`   // Enable development mode
	ServiceName string `yaml:"service_name" mapstructure:"service_name"` // Service name for structured logging
	Version     string `yaml:"version" mapstructure:"version"`           // Service version
	Output      string `yaml:"output" mapstructure:"output"`             // "stdout", "stderr", or file path
}

// LoggerPackageConfig represents the config structure expected by your logger package
// This matches the structure in your logger package
type LoggerPackageConfig struct {
	Type        string
	Level       int
	Output      io.Writer
	Format      string
	Development bool
	ServiceName string
	Version     string
}

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	Provider  string // "prometheus" or "otel"
	Namespace string
	Subsystem string
	Enabled   bool
}

// Load loads configuration from environment variables and files using Viper
func Load() *Config {
	v := viper.New()

	// Set configuration file details - support multiple formats including .env
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../../../")
	v.AddConfigPath("/etc/myapp")

	// Enable reading from environment variables
	v.AutomaticEnv()

	// Set environment variable replacer for nested keys
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults(v)

	// Bind environment variables BEFORE reading config files
	// This ensures env vars take precedence over config files
	bindEnvVars(v)

	// Try to read .env file first (for backward compatibility)
	loadDotEnvFile(v)

	// Try to read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		// Config file not found or error reading - continue with env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Warning: Error reading config file: %v\n", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		panic(fmt.Sprintf("Unable to decode config: %v", err))
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	return &config
}

// LoadWithViper loads configuration and returns both config and viper instance
// This is useful for advanced usage where you need access to viper directly
func LoadWithViper() (*Config, *viper.Viper) {
	v := viper.New()

	// Set configuration file details
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.config/myapp")
	v.AddConfigPath("/etc/myapp")

	// Enable reading from environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults(v)

	// Bind environment variables BEFORE reading config files
	// This ensures env vars take precedence over config files
	bindEnvVars(v)

	// Try to read .env file first
	loadDotEnvFile(v)

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Warning: Error reading config file: %v\n", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		panic(fmt.Sprintf("Unable to decode config: %v", err))
	}

	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	return &config, v
}

// setDefaults sets all default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "ledger")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.stage", string(DevelopmentStage))
	v.SetDefault("app.debug", false)
	v.SetDefault("app.environment", "local")
	v.SetDefault("app.namespace", "default")

	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.grpc_port", "9090")
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "admin")
	v.SetDefault("database.password", "admin")
	v.SetDefault("database.database", "ledger")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Temporal defaults - comprehensive system-wide configuration
	SetTemporalDefaults(v)

	// Auth defaults
	v.SetDefault("auth.jwt_secret", "")

	// Feature defaults
	v.SetDefault("features.enable_new_dashboard", false)

	// Logger defaults
	v.SetDefault("logger.type", "zerolog")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.dev", false)
	v.SetDefault("logger.service_name", "AwoERP")
	v.SetDefault("logger.version", "1.0.0")
	v.SetDefault("logger.output", "stdout")
}

// bindEnvVars binds environment variables to maintain backward compatibility
func bindEnvVars(v *viper.Viper) {
	// App
	v.BindEnv("app.name", "APP_NAME")
	v.BindEnv("app.version", "APP_VERSION")
	v.BindEnv("app.stage", "APP_STAGE")
	v.BindEnv("app.debug", "DEBUG", "APP_DEBUG")
	v.BindEnv("app.environment", "ENVIRONMENT", "APP_ENV")
	v.BindEnv("app.namespace", "NAMESPACE", "APP_NAMESPACE")

	// Server
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.grpc_port", "GRPC_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	// Database
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.database", "DB_NAME")
	v.BindEnv("database.ssl_mode", "DB_SSL_MODE")
	v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	// Redis
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")

	// Temporal - comprehensive system-wide environment bindings
	BindTemporalEnvVars(v)

	// Auth
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")

	// Features
	v.BindEnv("features.enable_new_dashboard", "ENABLE_NEW_DASHBOARD")

	// Logger
	v.BindEnv("logger.type", "LOG_TYPE")
	v.BindEnv("logger.level", "LOG_LEVEL")
	v.BindEnv("logger.format", "LOG_FORMAT")
	v.BindEnv("logger.dev", "LOG_DEV")
	v.BindEnv("logger.service_name", "SERVICE_NAME")
	v.BindEnv("logger.version", "SERVICE_VERSION")
	v.BindEnv("logger.output", "LOG_OUTPUT")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate app configuration
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app config validation failed: %w", err)
	}

	// Add validation logic here
	if c.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return fmt.Errorf("redis port must be between 1 and 65535")
	}

	// Validate logger configuration
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger config validation failed: %w", err)
	}

	return nil
}

// Validate validates the app configuration
func (a *AppConfig) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	if a.Version == "" {
		return fmt.Errorf("app version cannot be empty")
	}

	validStages := map[AppStage]bool{
		DevelopmentStage: true,
		StagingStage:     true,
		ProductionStage:  true,
		TestingStage:     true,
	}

	if !validStages[a.Stage] {
		return fmt.Errorf("invalid app stage: %s, must be one of: development, staging, production, testing", a.Stage)
	}

	return nil
}

// GetLogLevel returns the appropriate log level based on app stage and debug settings
func (a *AppConfig) GetLogLevel() string {
	// If debug is explicitly enabled, use debug level
	if a.Debug {
		return "debug"
	}

	// Stage-based log level defaults
	switch a.Stage {
	case DevelopmentStage, TestingStage:
		return "debug"
	case StagingStage:
		return "info"
	case ProductionStage:
		return "warn"
	default:
		return "info"
	}
}

// ShouldEnableDevelopmentMode returns true if development features should be enabled
func (a *AppConfig) ShouldEnableDevelopmentMode() bool {
	return a.Debug || a.Stage == DevelopmentStage || a.Stage == TestingStage
}

// Validate validates the logger configuration
func (l *LoggerConfig) Validate() error {
	validTypes := map[string]bool{
		"zap":     true,
		"zerolog": true,
		"slog":    true,
	}
	if !validTypes[l.Type] {
		return fmt.Errorf("invalid logger type: %s, must be one of: zap, zerolog, slog", l.Type)
	}

	if l.Level != "" {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
			"fatal": true,
		}
		if !validLevels[strings.ToLower(l.Level)] {
			return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error, fatal", l.Level)
		}
	}

	validFormats := map[string]bool{
		"json":    true,
		"text":    true,
		"console": true,
	}
	if !validFormats[l.Format] {
		return fmt.Errorf("invalid log format: %s, must be one of: json, text, console", l.Format)
	}

	return nil
}

// ToLoggerConfig converts LoggerConfig to logger package Config
// This bridges the gap between your config and the logger package
func (l *LoggerConfig) ToLoggerConfig(appConfig *AppConfig) LoggerPackageConfig {
	config := LoggerPackageConfig{
		ServiceName: l.ServiceName,
		Version:     l.Version,
		Development: l.Development,
		Format:      l.Format,
		Output:      os.Stdout, // Default to stdout
	}

	// Use app config values if logger config values are empty
	if config.ServiceName == "" {
		config.ServiceName = appConfig.Name
	}
	if config.Version == "" {
		config.Version = appConfig.Version
	}

	// Override development mode based on app config
	if appConfig.ShouldEnableDevelopmentMode() {
		config.Development = true
	}

	// Convert logger type
	switch strings.ToLower(l.Type) {
	case "zap":
		config.Type = "zap"
	case "zerolog":
		config.Type = "zerolog"
	case "slog":
		config.Type = "slog"
	default:
		config.Type = "zerolog" // Default fallback
	}

	// Convert log level - prioritize app config's intelligent level detection
	logLevel := l.Level
	if logLevel == "" || (appConfig.Debug && logLevel != "debug") {
		logLevel = appConfig.GetLogLevel()
	}

	switch strings.ToLower(logLevel) {
	case "debug":
		config.Level = 0 // DebugLevel
	case "info":
		config.Level = 1 // InfoLevel
	case "warn", "warning":
		config.Level = 2 // WarnLevel
	case "error":
		config.Level = 3 // ErrorLevel
	case "fatal":
		config.Level = 4 // FatalLevel
	default:
		config.Level = 1 // Default to InfoLevel
	}

	// Handle output destination
	switch strings.ToLower(l.Output) {
	case "stderr":
		config.Output = os.Stderr
	case "stdout", "":
		config.Output = os.Stdout
	default:
		// If it's not stdout/stderr, assume it's a file path
		if file, err := os.OpenFile(l.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666); err == nil {
			config.Output = file
		} else {
			config.Output = os.Stdout // Fallback to stdout if file can't be opened
		}
	}

	return config
}

// loadDotEnvFile loads .env file if it exists (for backward compatibility)
func loadDotEnvFile(v *viper.Viper) {
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		file, err := os.Open(envFile)
		if err != nil {
			fmt.Printf("Warning: Could not open .env file: %v\n", err)
			return
		}
		defer file.Close()

		// Read .env file line by line
		content := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				content = append(content, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Parse .env content
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove quotes if present
				if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
					value = value[1 : len(value)-1]
				}
				// Only set the environment variable if it's not already set
				// This allows command-line env vars to override .env file values
				if os.Getenv(key) == "" {
					os.Setenv(key, value)
				}
			}
		}
	}
}

// Legacy helper functions - kept for backward compatibility
// These are deprecated but maintained to avoid breaking existing code

// Deprecated: Use Load() instead. This function is kept for backward compatibility.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Deprecated: Use Load() instead. This function is kept for backward compatibility.
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Deprecated: Use Load() instead. This function is kept for backward compatibility.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
