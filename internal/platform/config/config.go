package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration object.
//
// Load priority (highest → lowest):
//  1. Environment variables (see bindEnvVars for the full mapping)
//  2. config.yaml (or the file found via AddConfigPath)
//  3. Built-in defaults (see setDefaults)
type Config struct {
	App       AppConfig       `yaml:"app"       mapstructure:"app"`
	Server    ServerConfig    `yaml:"server"    mapstructure:"server"`
	Database  DatabaseConfig  `yaml:"database"  mapstructure:"database"`
	Migration MigrationConfig `yaml:"migration" mapstructure:"migration"`
	Redis     RedisConfig     `yaml:"redis"     mapstructure:"redis"`
	Temporal  TemporalConfig  `yaml:"temporal"  mapstructure:"temporal"`
	Auth      AuthConfig      `yaml:"auth"      mapstructure:"auth"`
	Features  FeatureConfig   `yaml:"features"  mapstructure:"features"`
	Logger    LoggerConfig    `yaml:"logger"    mapstructure:"logger"`
	Tracing   TracingConfig   `yaml:"tracing"   mapstructure:"tracing"`
	UI        UIConfig        `yaml:"ui"        mapstructure:"ui"`
}

// Load loads configuration from config.yaml and environment variables.
func Load() *Config {
	cfg, _ := load()
	return cfg
}

// LoadWithViper is like Load but also returns the underlying Viper instance
// for callers that need raw key access or dynamic reconfiguration.
func LoadWithViper() (*Config, *viper.Viper) {
	return load()
}

// load is the shared implementation used by both Load and LoadWithViper.
func load() (*Config, *viper.Viper) {
	v := viper.New()

	// Load .env first so its values are available as real env vars before
	// viper bindings are evaluated. Exported shell vars always win because
	// loadDotEnvFile only calls os.Setenv when the var is not already set.
	loadDotEnvFile()

	setDefaults(v)

	// Explicit env-var bindings (see bindEnvVars).
	// We do NOT call AutomaticEnv because it can shadow config-file values:
	// viper treats every key as if it has an env binding, and if the derived
	// var name happens to exist in the environment (even from an unrelated
	// tool) the config-file value would be silently ignored.
	bindEnvVars(v)

	configureConfigFile(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("config: no config file found (searched %v), using defaults + env vars\n", configSearchPaths(v))
		} else {
			fmt.Printf("config: error reading config file: %v\n", err)
		}
	} else {
		fmt.Printf("config: loaded %s\n", v.ConfigFileUsed())
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("config: unable to decode: %v", err))
	}
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("config: invalid: %v", err))
	}
	return &cfg, v
}

// configureConfigFile sets up config file search paths.
// If CONFIG_FILE env var is set it is used directly; otherwise viper searches
// several conventional locations relative to the working directory.
func configureConfigFile(v *viper.Viper) {
	if explicit := os.Getenv("CONFIG_FILE"); explicit != "" {
		v.SetConfigFile(explicit)
		return
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Search in working directory and common parent paths so the binary works
	// regardless of whether it is invoked as `go run ./cmd/server/` or
	// `./bin/server` from the project root.
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Walk up a few levels to support running from a sub-directory.
	if abs, err := filepath.Abs("."); err == nil {
		v.AddConfigPath(filepath.Join(abs, ".."))
		v.AddConfigPath(filepath.Join(abs, "../.."))
		v.AddConfigPath(filepath.Join(abs, "../../.."))
	}

	v.AddConfigPath("/etc/myapp")
}

// configSearchPaths returns a human-readable list of paths viper searched.
func configSearchPaths(v *viper.Viper) []string {
	return []string{".", "./config", "..", "../..", "../../..", "/etc/myapp"}
}

// ─── Defaults ────────────────────────────────────────────────────────────────

func setDefaults(v *viper.Viper) {
	// App
	v.SetDefault("app.name", "ledger")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.stage", string(DevelopmentStage))
	v.SetDefault("app.debug", false)
	v.SetDefault("app.environment", "local")
	v.SetDefault("app.namespace", "default")

	// Server
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.grpc_port", "9090")
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)

	// Database
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "admin")
	v.SetDefault("database.password", "admin")
	v.SetDefault("database.database", "ledger")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)

	// Migration
	v.SetDefault("migration.url", "file://db/migration")
	v.SetDefault("migration.timeout", 5*time.Minute)
	v.SetDefault("migration.lock_timeout", 15*time.Minute)
	v.SetDefault("migration.verbose", true)
	v.SetDefault("migration.no_verify", false)

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Auth
	v.SetDefault("auth.jwt_secret", "")
	v.SetDefault("auth.max_failed_attempts", 5)
	v.SetDefault("auth.lockout_duration", 15*time.Minute)
	v.SetDefault("auth.session_ttl", 8*time.Hour)
	v.SetDefault("auth.cookie_name", "session")
	v.SetDefault("auth.require_https", true)
	v.SetDefault("auth.mfa_encryption_key", "dev-mfa-key-change-me-in-prod!!!")
	v.SetDefault("auth.mfa_issuer", "AWO ERP")

	// Logger
	v.SetDefault("logger.type", "zerolog")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.development", false)
	v.SetDefault("logger.service_name", "AwoERP")
	v.SetDefault("logger.version", "1.0.0")
	v.SetDefault("logger.output", "stdout")

	// Tracing
	v.SetDefault("tracing.enabled", false)
	v.SetDefault("tracing.exporter", "stdout")
	v.SetDefault("tracing.protocol", "stdout")
	v.SetDefault("tracing.endpoint", "localhost:4317")
	v.SetDefault("tracing.insecure", true)
	v.SetDefault("tracing.sampling_ratio", 1.0)

	// Features
	v.SetDefault("features.enable_new_dashboard", false)

	// Temporal and UI have their own default functions to keep this file manageable.
	SetTemporalDefaults(v)
	SetUIDefaults(v)
}

// ─── Environment variable bindings ───────────────────────────────────────────
//
// Each binding maps a conventional env var name to the exact viper key used in
// the YAML file. To add a new env-var override, add a BindEnv line here and a
// matching "# Env: VAR_NAME" comment in config.yaml.

func bindEnvVars(v *viper.Viper) {
	// App
	v.BindEnv("app.name", "APP_NAME")
	v.BindEnv("app.version", "APP_VERSION")
	v.BindEnv("app.stage", "APP_STAGE")
	v.BindEnv("app.debug", "DEBUG", "APP_DEBUG")
	v.BindEnv("app.environment", "ENVIRONMENT", "APP_ENV")
	v.BindEnv("app.namespace", "NAMESPACE", "APP_NAMESPACE")

	// Server
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.grpc_port", "GRPC_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

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

	// Auth
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")
	v.BindEnv("auth.max_failed_attempts", "AUTH_MAX_FAILED_ATTEMPTS")
	v.BindEnv("auth.lockout_duration", "AUTH_LOCKOUT_DURATION")
	v.BindEnv("auth.session_ttl", "AUTH_SESSION_TTL")
	v.BindEnv("auth.cookie_name", "AUTH_COOKIE_NAME")
	v.BindEnv("auth.require_https", "AUTH_REQUIRE_HTTPS")
	v.BindEnv("auth.mfa_encryption_key", "MFA_ENCRYPTION_KEY")
	v.BindEnv("auth.mfa_issuer", "MFA_ISSUER")

	// Logger
	v.BindEnv("logger.type", "LOG_TYPE")
	v.BindEnv("logger.level", "LOG_LEVEL")
	v.BindEnv("logger.format", "LOG_FORMAT")
	v.BindEnv("logger.development", "LOG_DEV")
	v.BindEnv("logger.service_name", "SERVICE_NAME")
	v.BindEnv("logger.version", "SERVICE_VERSION")
	v.BindEnv("logger.output", "LOG_OUTPUT")

	// Tracing
	v.BindEnv("tracing.enabled", "TRACING_ENABLED")
	v.BindEnv("tracing.exporter", "TRACING_EXPORTER")
	v.BindEnv("tracing.protocol", "TRACING_PROTOCOL")
	v.BindEnv("tracing.endpoint", "OTEL_ENDPOINT")
	v.BindEnv("tracing.insecure", "OTEL_INSECURE")
	v.BindEnv("tracing.sampling_ratio", "TRACING_SAMPLING_RATIO")

	// Features
	v.BindEnv("features.enable_new_dashboard", "ENABLE_NEW_DASHBOARD")

	// Temporal and UI have their own binding functions.
	BindTemporalEnvVars(v)
	BindUIEnvVars(v)
}

// ─── Validation ──────────────────────────────────────────────────────────────

func (c *Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host cannot be empty")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port must be between 1 and 65535")
	}
	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return fmt.Errorf("redis.port must be between 1 and 65535")
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	if err := c.UI.Validate(); err != nil {
		return fmt.Errorf("ui: %w", err)
	}
	return nil
}

// ─── .env file loader ────────────────────────────────────────────────────────

// loadDotEnvFile reads a .env file from the current directory and sets any
// variables not already present in the environment. This gives .env lower
// priority than exported shell variables.
func loadDotEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return // .env is optional; absence is not an error
	}
	defer file.Close()

	content, _ := readAll(file)
	for _, line := range bytes.Split(content, []byte("\n")) {
		s := strings.TrimSpace(string(line))
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func readAll(f *os.File) ([]byte, error) {
	var out []byte
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return out, nil
}
