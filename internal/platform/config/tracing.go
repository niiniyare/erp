package config

// TracingConfig holds OpenTelemetry tracing configuration
type TracingConfig struct {
	Enabled       bool    `yaml:"enabled" mapstructure:"enabled"`
	Exporter      string  `yaml:"exporter" mapstructure:"exporter"`
	Protocol      string  `yaml:"protocol" mapstructure:"protocol"`
	Endpoint      string  `yaml:"endpoint" mapstructure:"endpoint"`
	Insecure      bool    `yaml:"insecure" mapstructure:"insecure"`
	SamplingRatio float64 `yaml:"sampling_ratio" mapstructure:"sampling_ratio"`
}
