package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObservabilityConfig_Default(t *testing.T) {
	config := DefaultObservabilityConfig()
	
	assert.Equal(t, "erp-api", config.ServiceName)
	assert.Equal(t, false, config.DetailedLogging)
	assert.NotEmpty(t, config.CustomLabels)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/metrics")
}

func TestObservabilityConfig_Custom(t *testing.T) {
	config := ObservabilityConfig{
		ServiceName:     "test-service",
		DetailedLogging: false,
		SkipPaths:       []string{"/test"},
		CustomLabels:    map[string]string{"env": "test"},
	}
	
	assert.Equal(t, "test-service", config.ServiceName)
	assert.Equal(t, false, config.DetailedLogging)
	assert.Equal(t, []string{"/test"}, config.SkipPaths)
	assert.Equal(t, "test", config.CustomLabels["env"])
}