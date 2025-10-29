package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents application configuration
type Config struct {
	App       AppConfig       `yaml:"app" mapstructure:"app"`
	Server    ServerConfig    `yaml:"server" mapstructure:"server"`
	Database  DatabaseConfig  `yaml:"database" mapstructure:"database"`
	Migration MigrationConfig `yaml:"migration" mapstructure:"migration"`
	Redis     RedisConfig     `yaml:"redis" mapstructure:"redis"`
	Temporal  TemporalConfig  `yaml:"temporal" mapstructure:"temporal"`
	Auth      AuthConfig      `yaml:"auth" mapstructure:"auth"`
	Features  FeatureConfig   `yaml:"features" mapstructure:"features"`
	Logger    LoggerConfig    `yaml:"logger" mapstructure:"logger"`
	UI        UIConfig        `yaml:"ui" mapstructure:"ui"`
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

	// Try to read .env file (for backward compatibility)
	// This must come AFTER bindEnvVars to not override exported env vars
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

	// Migration defaults
	v.SetDefault("migration.url", "file://db/migration")
	v.SetDefault("migration.timeout", 5*time.Minute)
	v.SetDefault("migration.lock_timeout", 15*time.Minute)
	v.SetDefault("migration.verbose", true)
	v.SetDefault("migration.no_verify", false)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Temporal defaults - system-wide configuration
	SetTemporalDefaults(v)

	// UI defaults
	SetUIDefaults(v)

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

	// Temporal - system-wide environment bindings
	BindTemporalEnvVars(v)

	// UI environment bindings
	BindUIEnvVars(v)

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

	// Validate UI configuration
	if err := c.UI.Validate(); err != nil {
		return fmt.Errorf("ui config validation failed: %w", err)
	}

	return nil
}

// loadDotEnvFile loads .env file if it exists (for backward compatibility)
func loadDotEnvFile(_ *viper.Viper) {
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
		lines := bytes.Split(content, []byte("\n"))
		for _, line := range lines {
			lineStr := strings.TrimSpace(string(line))
			if lineStr == "" || strings.HasPrefix(lineStr, "#") {
				continue
			}

			parts := strings.SplitN(lineStr, "=", 2)
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

// Deprecated: Use Load() instead. This function is kept for backward compatibility.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
