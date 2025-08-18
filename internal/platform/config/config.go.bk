package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Temporal TemporalConfig `yaml:"temporal"`
	Auth     AuthConfig     `yaml:"auth"`
	Features FeatureConfig  `yaml:"features"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	GRPCPort     string        `yaml:"grpc_port"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
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
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// TemporalConfig represents Temporal configuration
type TemporalConfig struct {
	HostPort string `yaml:"host_port"`
}

// AuthConfig represents auth configuration
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

// FeatureConfig represents feature flags
type FeatureConfig struct {
	EnableNewDashboard bool `yaml:"enable_new_dashboard"`
}

// Load loads configuration from environment variables and files
func Load() *Config {
	config := &Config{}

	// Load from environment variables with defaults
	config.Server.Port = getEnv("SERVER_PORT", "8080")
	config.Server.GRPCPort = getEnv("GRPC_PORT", "9090")
	config.Server.ReadTimeout = getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second)
	config.Server.WriteTimeout = getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second)

	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getIntEnv("DB_PORT", 5432)
	config.Database.User = getEnv("DB_USER", "admin")
	config.Database.Password = getEnv("DB_PASSWORD", "admin")
	config.Database.Database = getEnv("DB_NAME", "ledger")
	config.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")
	config.Database.MaxOpenConns = getIntEnv("DB_MAX_OPEN_CONNS", 25)
	config.Database.MaxIdleConns = getIntEnv("DB_MAX_IDLE_CONNS", 5)
	config.Database.ConnMaxLifetime = getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	config.Redis.Host = getEnv("REDIS_HOST", "localhost")
	config.Redis.Port = getIntEnv("REDIS_PORT", 6379)
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")
	config.Redis.DB = getIntEnv("REDIS_DB", 0)

	config.Temporal.HostPort = getEnv("TEMPORAL_HOST_PORT", "localhost:7233")

	// Validate configuration
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here
	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
