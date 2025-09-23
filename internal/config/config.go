package config

import (
	"fmt"
	"os"
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

// AppConfig represents the entire application configuration
// This is our single source of truth for ALL configuration
type AppConfig struct {
	App      AppSettings      `yaml:"app" mapstructure:"app"`
	Server   ServerSettings   `yaml:"server" mapstructure:"server"`
	Database DatabaseSettings `yaml:"database" mapstructure:"database"`
	Redis    RedisSettings    `yaml:"redis" mapstructure:"redis"`
	Temporal TemporalSettings `yaml:"temporal" mapstructure:"temporal"`
	Auth     AuthSettings     `yaml:"auth" mapstructure:"auth"`
	Features FeatureSettings  `yaml:"features" mapstructure:"features"`
	Logger   LoggerSettings   `yaml:"logger" mapstructure:"logger"`
	UI       UIConfig         `yaml:"ui" mapstructure:"ui"`
}

type AppSettings struct {
	Name        string   `yaml:"name" mapstructure:"name"`
	Version     string   `yaml:"version" mapstructure:"version"`
	Stage       AppStage `yaml:"stage" mapstructure:"stage"`
	Debug       bool     `yaml:"debug" mapstructure:"debug"`
	Environment string   `yaml:"environment" mapstructure:"environment"`
	Namespace   string   `yaml:"namespace" mapstructure:"namespace"`
}

type ServerSettings struct {
	Port         string        `yaml:"port" mapstructure:"port"`
	Host         string        `yaml:"host" mapstructure:"host"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	GRPCPort     string        `yaml:"grpc_port" mapstructure:"grpc_port"`
}

type DatabaseSettings struct {
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
func (d *DatabaseSettings) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
		d.SSLMode,
	)
}

type RedisSettings struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

type TemporalSettings struct {
	HostPort  string         `yaml:"host_port" mapstructure:"host_port"`
	Namespace string         `yaml:"namespace" mapstructure:"namespace"`
	Workers   WorkerSettings `yaml:"workers" mapstructure:"workers"`
	Client    ClientSettings `yaml:"client" mapstructure:"client"`
}

type WorkerSettings struct {
	MaxConcurrentActivities int           `yaml:"max_concurrent_activities" mapstructure:"max_concurrent_activities"`
	MaxConcurrentWorkflows  int           `yaml:"max_concurrent_workflows" mapstructure:"max_concurrent_workflows"`
	WorkerStopTimeout       time.Duration `yaml:"worker_stop_timeout" mapstructure:"worker_stop_timeout"`
}

type ClientSettings struct {
	Identity          string        `yaml:"identity" mapstructure:"identity"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" mapstructure:"connection_timeout"`
}

type AuthSettings struct {
	JWTSecret string `yaml:"jwt_secret" mapstructure:"jwt_secret"`
}

type FeatureSettings struct {
	EnableNewDashboard bool `yaml:"enable_new_dashboard" mapstructure:"enable_new_dashboard"`
}

type LoggerSettings struct {
	Type        string `yaml:"type" mapstructure:"type"`
	Level       string `yaml:"level" mapstructure:"level"`
	Format      string `yaml:"format" mapstructure:"format"`
	Development bool   `yaml:"development" mapstructure:"development"`
	ServiceName string `yaml:"service_name" mapstructure:"service_name"`
	Version     string `yaml:"version" mapstructure:"version"`
	Output      string `yaml:"output" mapstructure:"output"`
}

// Load loads configuration from all sources with proper precedence
func Load() (*AppConfig, error) {
	v := viper.New()

	// Set configuration sources
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/myapp")

	// Enable reading from environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults(v)

	// Bind environment variables BEFORE reading config files
	bindEnvVars(v)

	// Try to read .env file (for backward compatibility)
	loadDotEnvFile(v)

	// Try to read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// Validate validates all configuration settings
func (c *AppConfig) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}

	if err := c.UI.Validate(); err != nil {
		return fmt.Errorf("ui config validation failed: %w", err)
	}

	return nil
}

// Helper methods
func (c *AppConfig) IsProduction() bool {
	return c.App.Stage.IsProduction()
}

func (c *AppConfig) IsDevelopment() bool {
	return c.App.Stage.IsDevelopment()
}

func (c *AppConfig) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// setDefaults sets sensible defaults for all configurations
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "erp-server")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.stage", string(DevelopmentStage))
	v.SetDefault("app.debug", false)
	v.SetDefault("app.environment", "local")
	v.SetDefault("app.namespace", "default")

	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.grpc_port", "9090")

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

	// Temporal defaults
	v.SetDefault("temporal.host_port", "localhost:7233")
	v.SetDefault("temporal.namespace", "default")
	v.SetDefault("temporal.workers.max_concurrent_activities", 100)
	v.SetDefault("temporal.workers.max_concurrent_workflows", 100)
	v.SetDefault("temporal.workers.worker_stop_timeout", 30*time.Second)
	v.SetDefault("temporal.client.identity", "awo-erp")
	v.SetDefault("temporal.client.connection_timeout", 30*time.Second)

	// Auth defaults
	v.SetDefault("auth.jwt_secret", "")

	// Feature defaults
	v.SetDefault("features.enable_new_dashboard", false)

	// Logger defaults
	v.SetDefault("logger.type", "zerolog")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.development", false)
	v.SetDefault("logger.service_name", "AwoERP")
	v.SetDefault("logger.version", "1.0.0")
	v.SetDefault("logger.output", "stdout")

	// UI Defaults
	SetUIDefaults(v)
}

// bindEnvVars binds environment variables
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
	v.BindEnv("server.host", "SERVER_HOST")
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

	// Temporal
	v.BindEnv("temporal.host_port", "TEMPORAL_HOST_PORT")

	// Auth
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")

	// Features
	v.BindEnv("features.enable_new_dashboard", "ENABLE_NEW_DASHBOARD")

	// Logger
	v.BindEnv("logger.type", "LOG_TYPE")
	v.BindEnv("logger.level", "LOG_LEVEL")
	v.BindEnv("logger.format", "LOG_FORMAT")
	v.BindEnv("logger.development", "LOG_DEV")
	v.BindEnv("logger.service_name", "SERVICE_NAME")
	v.BindEnv("logger.version", "SERVICE_VERSION")
	v.BindEnv("logger.output", "LOG_OUTPUT")

	// UI
	BindUIEnvVars(v)
}

// loadDotEnvFile loads .env file if it exists (for backward compatibility)
func loadDotEnvFile(v *viper.Viper) {
	envFile := ".env"
	if _, err := os.Stat(envFile); err == nil {
		file, err := os.Open(envFile)
		if err != nil {
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
