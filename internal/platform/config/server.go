package config

import "time"

// ServerConfig represents HTTP/gRPC server configuration.
type ServerConfig struct {
	Host         string        `yaml:"host"          mapstructure:"host"`
	Port         string        `yaml:"port"          mapstructure:"port"`
	GRPCPort     string        `yaml:"grpc_port"     mapstructure:"grpc_port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"  mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"  mapstructure:"idle_timeout"`
	// AllowedOrigins is the comma-separated list of origins permitted by CORS.
	// Set via CORS_ALLOWED_ORIGINS env var.
	AllowedOrigins string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
}
