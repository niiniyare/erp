package config

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Type        string `yaml:"type" mapstructure:"type"`                 // "zap", "zerolog", "slog"
	Level       string `yaml:"level" mapstructure:"level"`               // "debug", "info", "warn", "error", "fatal"
	Format      string `yaml:"format" mapstructure:"format"`             // "json", "text", "console"
	Development bool   `yaml:"development" mapstructure:"development"`   // Enable development mode
	ServiceName string `yaml:"service_name" mapstructure:"service_name"` // Service name for structured logging
	Version     string `yaml:"version" mapstructure:"version"`           // Service version
	Output      string `yaml:"output" mapstructure:"output"`             // "stdout", "stderr", or file path
}

// LoggerPackageConfig represents the config structure expected by your logger package
// This matches the structure in your logger package
type LoggerPackageConfig struct {
	Type        string
	Level       int
	Output      io.Writer
	Format      string
	Development bool
	ServiceName string
	Version     string
}

// Validate validates the logger configuration
func (l *LoggerConfig) Validate() error {
	validTypes := map[string]bool{
		"zap":     true,
		"zerolog": true,
		"slog":    true,
	}
	if !validTypes[l.Type] {
		return fmt.Errorf("invalid logger type: %s, must be one of: zap, zerolog, slog", l.Type)
	}

	if l.Level != "" {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
			"fatal": true,
		}
		if !validLevels[strings.ToLower(l.Level)] {
			return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error, fatal", l.Level)
		}
	}

	validFormats := map[string]bool{
		"json":    true,
		"text":    true,
		"console": true,
	}
	if !validFormats[l.Format] {
		return fmt.Errorf("invalid log format: %s, must be one of: json, text, console", l.Format)
	}

	return nil
}

// ToLoggerConfig converts LoggerConfig to logger package Config
// This bridges the gap between your config and the logger package
func (l *LoggerConfig) ToLoggerConfig(appConfig *AppConfig) LoggerPackageConfig {
	config := LoggerPackageConfig{
		ServiceName: l.ServiceName,
		Version:     l.Version,
		Development: l.Development,
		Format:      l.Format,
		Output:      os.Stdout, // Default to stdout
	}

	// Use app config values if logger config values are empty
	if config.ServiceName == "" {
		config.ServiceName = appConfig.Name
	}
	if config.Version == "" {
		config.Version = appConfig.Version
	}

	// Override development mode based on app config
	if appConfig.ShouldEnableDevelopmentMode() {
		config.Development = true
	}

	// Convert logger type
	switch strings.ToLower(l.Type) {
	case "zap":
		config.Type = "zap"
	case "zerolog":
		config.Type = "zerolog"
	case "slog":
		config.Type = "slog"
	default:
		config.Type = "zerolog" // Default fallback
	}

	// Convert log level - prioritize app config's intelligent level detection
	logLevel := l.Level
	if logLevel == "" || (appConfig.Debug && logLevel != "debug") {
		logLevel = appConfig.GetLogLevel()
	}

	switch strings.ToLower(logLevel) {
	case "debug":
		config.Level = 0 // DebugLevel
	case "info":
		config.Level = 1 // InfoLevel
	case "warn", "warning":
		config.Level = 2 // WarnLevel
	case "error":
		config.Level = 3 // ErrorLevel
	case "fatal":
		config.Level = 4 // FatalLevel
	default:
		config.Level = 1 // Default to InfoLevel
	}

	// Handle output destination
	switch strings.ToLower(l.Output) {
	case "stderr":
		config.Output = os.Stderr
	case "stdout", "":
		config.Output = os.Stdout
	default:
		// If it's not stdout/stderr, assume it's a file path
		if file, err := os.OpenFile(l.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666); err == nil {
			config.Output = file
		} else {
			config.Output = os.Stdout // Fallback to stdout if file can't be opened
		}
	}

	return config
}
