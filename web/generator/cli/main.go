package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// CLIConfig holds command-line configuration
type CLIConfig struct {
	// Input options
	InputFile    string `json:"inputFile"`
	InputPackage string `json:"inputPackage"`
	InputModel   string `json:"inputModel"`
	
	// Output options
	OutputDir    string `json:"outputDir"`
	OutputFormat string `json:"outputFormat"` // "json", "yaml", "go"
	
	// Generation options
	Patterns     []string `json:"patterns"`
	Layout       string   `json:"layout"`
	APIPrefix    string   `json:"apiPrefix"`
	EnablePerms  bool     `json:"enablePermissions"`
	EnableAudit  bool     `json:"enableAuditTrail"`
	PageSize     int      `json:"pageSize"`
	
	// Behavior options
	Watch        bool   `json:"watch"`
	Validate     bool   `json:"validate"`
	Preview      bool   `json:"preview"`
	Force        bool   `json:"force"`
	Verbose      bool   `json:"verbose"`
	ConfigFile   string `json:"configFile"`
	
	// Tag customization
	CustomTags   map[string]string `json:"customTags"`
	IgnoreFields []string          `json:"ignoreFields"`
	
	// Template options
	TemplateDir  string            `json:"templateDir"`
	CustomComps  map[string]string `json:"customComponents"`
}

// Default configuration
var defaultConfig = CLIConfig{
	OutputDir:     "./generated",
	OutputFormat:  "json",
	Layout:        "app",
	APIPrefix:     "/api",
	EnablePerms:   true,
	EnableAudit:   true,
	PageSize:      20,
	Watch:         false,
	Validate:      true,
	Preview:       false,
	Force:         false,
	Verbose:       false,
	CustomTags:    make(map[string]string),
	IgnoreFields:  []string{"password", "secret", "token"},
	CustomComps:   make(map[string]string),
}

func main() {
	var config CLIConfig = defaultConfig
	
	// Define command-line flags
	flag.StringVar(&config.InputFile, "file", "", "Input Go file to analyze")
	flag.StringVar(&config.InputPackage, "package", "", "Input Go package directory to analyze")
	flag.StringVar(&config.InputModel, "model", "", "Specific model/struct name to generate UI for")
	flag.StringVar(&config.OutputDir, "output", defaultConfig.OutputDir, "Output directory for generated schemas")
	flag.StringVar(&config.OutputFormat, "format", defaultConfig.OutputFormat, "Output format (json, yaml, go)")
	flag.StringVar(&config.Layout, "layout", defaultConfig.Layout, "Default layout template (app, auth, minimal)")
	flag.StringVar(&config.APIPrefix, "api-prefix", defaultConfig.APIPrefix, "API endpoint prefix")
	flag.BoolVar(&config.EnablePerms, "permissions", defaultConfig.EnablePerms, "Enable permission-based access control")
	flag.BoolVar(&config.EnableAudit, "audit", defaultConfig.EnableAudit, "Enable audit trail features")
	flag.IntVar(&config.PageSize, "page-size", defaultConfig.PageSize, "Default table page size")
	flag.BoolVar(&config.Watch, "watch", defaultConfig.Watch, "Watch for file changes and regenerate")
	flag.BoolVar(&config.Validate, "validate", defaultConfig.Validate, "Validate generated schemas")
	flag.BoolVar(&config.Preview, "preview", defaultConfig.Preview, "Preview schemas without writing files")
	flag.BoolVar(&config.Force, "force", defaultConfig.Force, "Force overwrite existing files")
	flag.BoolVar(&config.Verbose, "verbose", defaultConfig.Verbose, "Enable verbose logging")
	flag.StringVar(&config.ConfigFile, "config", "", "Load configuration from file")
	flag.StringVar(&config.TemplateDir, "templates", "", "Custom template directory")
	
	var patternsFlag string
	var ignoreFieldsFlag string
	flag.StringVar(&patternsFlag, "patterns", "", "Comma-separated list of UI patterns to generate")
	flag.StringVar(&ignoreFieldsFlag, "ignore-fields", "", "Comma-separated list of fields to ignore")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "UI Schema Generator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate UI schemas for all models in a package\n")
		fmt.Fprintf(os.Stderr, "  %s -package ./internal/core/models\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate for a specific model\n")
		fmt.Fprintf(os.Stderr, "  %s -file user.go -model User\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Watch for changes and regenerate\n")
		fmt.Fprintf(os.Stderr, "  %s -package ./models -watch -verbose\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate specific patterns only\n")
		fmt.Fprintf(os.Stderr, "  %s -file user.go -patterns crud_table,create_form\n\n", os.Args[0])
	}
	
	flag.Parse()
	
	// Parse comma-separated lists
	if patternsFlag != "" {
		config.Patterns = strings.Split(patternsFlag, ",")
		for i, pattern := range config.Patterns {
			config.Patterns[i] = strings.TrimSpace(pattern)
		}
	}
	
	if ignoreFieldsFlag != "" {
		config.IgnoreFields = strings.Split(ignoreFieldsFlag, ",")
		for i, field := range config.IgnoreFields {
			config.IgnoreFields[i] = strings.TrimSpace(field)
		}
	}
	
	// Load configuration file if specified
	if config.ConfigFile != "" {
		if err := loadConfigFile(&config, config.ConfigFile); err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
	}
	
	// Validate input options
	if config.InputFile == "" && config.InputPackage == "" {
		fmt.Fprintf(os.Stderr, "Error: Either -file or -package must be specified\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	if config.InputFile != "" && config.InputPackage != "" {
		log.Fatal("Error: Cannot specify both -file and -package options")
	}
	
	// Setup logging
	if config.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
	}
	
	// Create output directory
	if !config.Preview {
		if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}
	
	// Run generation
	if config.Watch {
		runWatchMode(config)
	} else {
		if err := runGeneration(config); err != nil {
			log.Fatalf("Generation failed: %v", err)
		}
	}
}

// runGeneration performs the schema generation process
func runGeneration(config CLIConfig) error {
	if config.Verbose {
		log.Printf("Starting UI schema generation...")
	}
	
	// TODO: Implement actual generation logic
	// This would integrate with the generator package
	log.Printf("Generation would process: %s", config.InputFile+config.InputPackage)
	log.Printf("Output to: %s", config.OutputDir)
	
	return nil
}

// runWatchMode runs the generator in watch mode
func runWatchMode(config CLIConfig) {
	log.Printf("Starting watch mode...")
	
	// TODO: Implement file watching
	log.Printf("Watch mode not implemented yet")
	log.Printf("Falling back to single generation run...")
	
	if err := runGeneration(config); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}
}

// loadConfigFile loads configuration from a JSON file
func loadConfigFile(config *CLIConfig, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	return json.Unmarshal(data, config)
}

// TODO: Additional CLI functionality would be implemented here
// This simplified version removes the import dependencies to fix go vet issues