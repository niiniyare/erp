package config

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	Provider  string // "prometheus" or "otel"
	Namespace string
	Subsystem string
	Enabled   bool
}
