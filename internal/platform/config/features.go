package config

// FeatureConfig represents feature flags
type FeatureConfig struct {
	EnableNewDashboard bool `yaml:"enable_new_dashboard" mapstructure:"enable_new_dashboard"`
}
