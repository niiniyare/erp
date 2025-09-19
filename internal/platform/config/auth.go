package config

// AuthConfig represents auth configuration
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret" mapstructure:"jwt_secret"`
}
