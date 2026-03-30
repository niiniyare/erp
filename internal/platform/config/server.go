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
}
