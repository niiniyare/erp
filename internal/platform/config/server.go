package config

import "time"

// ServerConfig represents server configuration
type ServerConfig struct {
	Port         string        `yaml:"port" mapstructure:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	GRPCPort     string        `yaml:"grpc_port" mapstructure:"grpc_port"`
}
