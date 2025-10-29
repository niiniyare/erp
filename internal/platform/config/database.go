package config

import (
	"fmt"
	"time"
)

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
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

// MigrationConfig represents database migration configuration
type MigrationConfig struct {
	URL         string        `yaml:"url" mapstructure:"url"`
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`
	LockTimeout time.Duration `yaml:"lock_timeout" mapstructure:"lock_timeout"`
	Verbose     bool          `yaml:"verbose" mapstructure:"verbose"`
	NoVerify    bool          `yaml:"no_verify" mapstructure:"no_verify"`
}
