package config

import "fmt"

// AppStage represents application deployment stage
type AppStage string

const (
	DevelopmentStage AppStage = "dev"
	StagingStage     AppStage = "staging"
	ProductionStage  AppStage = "production"
	TestingStage     AppStage = "testing"
)

// String returns the string representation of AppStage
func (a AppStage) String() string {
	return string(a)
}

// IsProduction returns true if the stage is production
func (a AppStage) IsProduction() bool {
	return a == ProductionStage
}

// IsDevelopment returns true if the stage is development
func (a AppStage) IsDevelopment() bool {
	return a == DevelopmentStage
}

// IsStaging returns true if the stage is staging
func (a AppStage) IsStaging() bool {
	return a == StagingStage
}

// IsTesting returns true if the stage is testing
func (a AppStage) IsTesting() bool {
	return a == TestingStage
}

// AppConfig represents application-level configuration
type AppConfig struct {
	Name        string   `yaml:"name" mapstructure:"name"`               // Application name
	Version     string   `yaml:"version" mapstructure:"version"`         // Application version
	Stage       AppStage `yaml:"stage" mapstructure:"stage"`             // Application stage (development, staging, production, testing)
	Debug       bool     `yaml:"debug" mapstructure:"debug"`             // Enable debug mode
	Environment string   `yaml:"environment" mapstructure:"environment"` // Custom environment identifier
	Namespace   string   `yaml:"namespace" mapstructure:"namespace"`     // Kubernetes namespace or deployment namespace
}

// IsProduction returns true if the application stage is production.
func (a *AppConfig) IsProduction() bool {
	return a.Stage.IsProduction()
}

// Validate validates the app configuration
func (a *AppConfig) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	if a.Version == "" {
		return fmt.Errorf("app version cannot be empty")
	}

	validStages := map[AppStage]bool{
		DevelopmentStage: true,
		StagingStage:     true,
		ProductionStage:  true,
		TestingStage:     true,
	}

	if !validStages[a.Stage] {
		return fmt.Errorf("invalid app stage: %s, must be one of: development, staging, production, testing", a.Stage)
	}

	return nil
}

// GetLogLevel returns the appropriate log level based on app stage and debug settings
func (a *AppConfig) GetLogLevel() string {
	// If debug is explicitly enabled, use debug level
	if a.Debug {
		return "debug"
	}

	// Stage-based log level defaults
	switch a.Stage {
	case DevelopmentStage, TestingStage:
		return "debug"
	case StagingStage:
		return "info"
	case ProductionStage:
		return "warn"
	default:
		return "info"
	}
}

// ShouldEnableDevelopmentMode returns true if development features should be enabled
func (a *AppConfig) ShouldEnableDevelopmentMode() bool {
	return a.Debug || a.Stage == DevelopmentStage || a.Stage == TestingStage
}
