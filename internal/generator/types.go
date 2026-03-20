package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ModuleConfig represents configuration for module generation
type ModuleConfig struct {
	Name      string // Module name (e.g., "inventory", "payroll")
	WithTests bool   // Generate test scaffolds
	DryRun    bool   // Preview mode
	Verbose   bool   // Verbose output
}

// FeatureConfig represents configuration for feature generation
type FeatureConfig struct {
	Path      string // Module/feature path (e.g., "finance/budgets")
	WithTests bool   // Generate test scaffolds
	DryRun    bool   // Preview mode
	Verbose   bool   // Verbose output
}

// DocsConfig represents configuration for documentation generation
type DocsConfig struct {
	ModuleName string // Module name (e.g., "inventory", "payroll")
	DryRun     bool   // Preview mode
	Verbose    bool   // Verbose output
}

// TemplateData contains all data needed for template rendering
type TemplateData struct {
	// Module information
	ModuleName        string // Original name (e.g., "inventory")
	ModuleNamePascal  string // PascalCase (e.g., "Inventory")
	ModuleNameCamel   string // camelCase (e.g., "inventory")
	ModuleNameSnake   string // snake_case (e.g., "inventory")
	ModuleNameKebab   string // kebab-case (e.g., "inventory")
	ModuleNamePlural  string // Plural form (e.g., "inventories")
	ModuleDescription string // Module description

	// Feature information (for feature generation)
	FeatureName         string // Feature name (e.g., "budgets")
	FeatureNamePascal   string // PascalCase (e.g., "Budgets")
	FeatureNameCamel    string // camelCase (e.g., "budgets")
	FeatureNameSnake    string // snake_case (e.g., "budgets")
	FeatureNameSingular string // Singular form (e.g., "budget")

	// Component information (for component generation)
	ComponentName       string // Component name (e.g., "user_profile")
	ComponentNamePascal string // PascalCase (e.g., "UserProfile")
	ComponentNameCamel  string // camelCase (e.g., "userProfile")
	ComponentNameSnake  string // snake_case (e.g., "user_profile")

	// Entity information (derived from module/feature)
	EntityName       string // Main entity name (e.g., "Item", "Budget")
	EntityNamePascal string // PascalCase (e.g., "Item")
	EntityNameCamel  string // camelCase (e.g., "item")
	EntityNameSnake  string // snake_case (e.g., "item")
	EntityNamePlural string // Plural form (e.g., "Items")

	// Package and import information
	PackageName string // Go package name
	ImportPath  string // Import path for the module

	// Generation metadata
	GeneratedAt      time.Time // When generated
	GeneratorVersion string    // awoctl version
	WithTests        bool      // Include tests

	// ERP-specific data
	RootType           string   // Default root type for financial modules
	DefaultPermissions []string // Default ABAC permissions

	// File paths
	DomainPath     string // Path to domain layer
	ServicePath    string // Path to service layer
	RepositoryPath string // Path to repository layer
	APIPath        string // Path to API layer
}

// FileTemplate represents a template file to be generated
type FileTemplate struct {
	TemplatePath string // Path to template file
	OutputPath   string // Output file path
	Data         TemplateData
}

// GenerationResult represents the result of a generation operation
type GenerationResult struct {
	FilesGenerated []string // List of generated files
	Errors         []error  // Any errors encountered
	DryRun         bool     // Whether this was a dry run
}

// Validate validates the module configuration
func (c *ModuleConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Validate module name format
	if !isValidIdentifier(c.Name) {
		return fmt.Errorf("module name '%s' is not a valid Go identifier", c.Name)
	}

	// Check for reserved Go keywords
	if isReservedKeyword(c.Name) {
		return fmt.Errorf("module name '%s' is a reserved Go keyword", c.Name)
	}

	return nil
}

// Validate validates the feature configuration
func (c *FeatureConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("feature path cannot be empty")
	}

	parts := strings.Split(c.Path, "/")
	if len(parts) != 2 {
		return fmt.Errorf("feature path must be in format 'module/feature', got '%s'", c.Path)
	}

	moduleName, featureName := parts[0], parts[1]

	if !isValidIdentifier(moduleName) {
		return fmt.Errorf("module name '%s' is not a valid Go identifier", moduleName)
	}

	if !isValidIdentifier(featureName) {
		return fmt.Errorf("feature name '%s' is not a valid Go identifier", featureName)
	}

	if isReservedKeyword(moduleName) || isReservedKeyword(featureName) {
		return fmt.Errorf("module or feature name uses reserved Go keyword")
	}

	return nil
}

// Validate validates the docs configuration
func (c *DocsConfig) Validate() error {
	if c.ModuleName == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Validate module name format
	if !isValidIdentifier(c.ModuleName) {
		return fmt.Errorf("module name '%s' is not a valid Go identifier", c.ModuleName)
	}

	// Check for reserved Go keywords
	if isReservedKeyword(c.ModuleName) {
		return fmt.Errorf("module name '%s' is a reserved Go keyword", c.ModuleName)
	}

	return nil
}

// ToTemplateData converts DocsConfig to TemplateData
func (c *DocsConfig) ToTemplateData() TemplateData {
	return TemplateData{
		ModuleName:        c.ModuleName,
		ModuleNamePascal:  ToPascalCase(c.ModuleName),
		ModuleNameCamel:   ToCamelCase(c.ModuleName),
		ModuleNameSnake:   ToSnakeCase(c.ModuleName),
		ModuleNameKebab:   ToKebabCase(c.ModuleName),
		ModuleNamePlural:  Pluralize(c.ModuleName),
		ModuleDescription: fmt.Sprintf("%s management module", ToPascalCase(c.ModuleName)),

		EntityName:       ToPascalCase(Singularize(c.ModuleName)),
		EntityNamePascal: ToPascalCase(Singularize(c.ModuleName)),
		EntityNameCamel:  ToCamelCase(Singularize(c.ModuleName)),
		EntityNameSnake:  ToSnakeCase(Singularize(c.ModuleName)),
		EntityNamePlural: ToPascalCase(c.ModuleName),

		PackageName: ToSnakeCase(c.ModuleName),
		ImportPath:  fmt.Sprintf("awo/internal/core/%s", ToSnakeCase(c.ModuleName)),

		GeneratedAt:      time.Now(),
		GeneratorVersion: "0.1.0",
		WithTests:        false, // Docs don't have tests

		RootType:           inferRootType(c.ModuleName),
		DefaultPermissions: generateDefaultPermissions(c.ModuleName),

		DomainPath:     filepath.Join("internal", "core", ToSnakeCase(c.ModuleName), "domain"),
		ServicePath:    filepath.Join("internal", "core", ToSnakeCase(c.ModuleName)),
		RepositoryPath: filepath.Join("internal", "core", ToSnakeCase(c.ModuleName), "repository"),
		APIPath:        filepath.Join("internal", "api", "design", "services", ToSnakeCase(c.ModuleName)),
	}
}

// ToTemplateData converts ModuleConfig to TemplateData
func (c *ModuleConfig) ToTemplateData() TemplateData {
	return TemplateData{
		ModuleName:        c.Name,
		ModuleNamePascal:  ToPascalCase(c.Name),
		ModuleNameCamel:   ToCamelCase(c.Name),
		ModuleNameSnake:   ToSnakeCase(c.Name),
		ModuleNameKebab:   ToKebabCase(c.Name),
		ModuleNamePlural:  Pluralize(c.Name),
		ModuleDescription: fmt.Sprintf("%s management module", ToPascalCase(c.Name)),

		EntityName:       ToPascalCase(Singularize(c.Name)),
		EntityNamePascal: ToPascalCase(Singularize(c.Name)),
		EntityNameCamel:  ToCamelCase(Singularize(c.Name)),
		EntityNameSnake:  ToSnakeCase(Singularize(c.Name)),
		EntityNamePlural: ToPascalCase(c.Name),

		PackageName: ToSnakeCase(c.Name),
		ImportPath:  fmt.Sprintf("awo/internal/core/%s", ToSnakeCase(c.Name)),

		GeneratedAt:      time.Now(),
		GeneratorVersion: "0.1.0",
		WithTests:        c.WithTests,

		RootType:           inferRootType(c.Name),
		DefaultPermissions: generateDefaultPermissions(c.Name),

		DomainPath:     filepath.Join("internal", "core", ToSnakeCase(c.Name), "domain"),
		ServicePath:    filepath.Join("internal", "core", ToSnakeCase(c.Name)),
		RepositoryPath: filepath.Join("internal", "core", ToSnakeCase(c.Name), "repository"),
		APIPath:        filepath.Join("internal", "api", "design", "services", ToSnakeCase(c.Name)),
	}
}

// ToTemplateData converts FeatureConfig to TemplateData
func (c *FeatureConfig) ToTemplateData() TemplateData {
	parts := strings.Split(c.Path, "/")
	moduleName, featureName := parts[0], parts[1]

	return TemplateData{
		ModuleName:        moduleName,
		ModuleNamePascal:  ToPascalCase(moduleName),
		ModuleNameCamel:   ToCamelCase(moduleName),
		ModuleNameSnake:   ToSnakeCase(moduleName),
		ModuleNameKebab:   ToKebabCase(moduleName),
		ModuleNamePlural:  Pluralize(moduleName),
		ModuleDescription: fmt.Sprintf("%s management module", ToPascalCase(moduleName)),

		FeatureName:         featureName,
		FeatureNamePascal:   ToPascalCase(featureName),
		FeatureNameCamel:    ToCamelCase(featureName),
		FeatureNameSnake:    ToSnakeCase(featureName),
		FeatureNameSingular: Singularize(featureName),

		EntityName:       ToPascalCase(Singularize(featureName)),
		EntityNamePascal: ToPascalCase(Singularize(featureName)),
		EntityNameCamel:  ToCamelCase(Singularize(featureName)),
		EntityNameSnake:  ToSnakeCase(Singularize(featureName)),
		EntityNamePlural: ToPascalCase(featureName),

		PackageName: ToSnakeCase(moduleName),
		ImportPath:  fmt.Sprintf("awo/internal/core/%s", ToSnakeCase(moduleName)),

		GeneratedAt:      time.Now(),
		GeneratorVersion: "0.1.0",
		WithTests:        c.WithTests,

		RootType:           inferRootType(moduleName),
		DefaultPermissions: generateDefaultPermissions(featureName),

		DomainPath:     filepath.Join("internal", "core", ToSnakeCase(moduleName), "domain"),
		ServicePath:    filepath.Join("internal", "core", ToSnakeCase(moduleName)),
		RepositoryPath: filepath.Join("internal", "core", ToSnakeCase(moduleName), "repository"),
		APIPath:        filepath.Join("internal", "api", "design", "services", ToSnakeCase(moduleName)),
	}
}

// Helper functions

// isValidIdentifier checks if a string is a valid Go identifier
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Go identifier regex: letter followed by letters, digits, underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, s)
	return matched
}

// isReservedKeyword checks if a string is a reserved Go keyword
func isReservedKeyword(s string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
		// Built-in types
		"bool": true, "byte": true, "complex64": true, "complex128": true, "error": true,
		"float32": true, "float64": true, "int": true, "int8": true, "int16": true,
		"int32": true, "int64": true, "rune": true, "string": true, "uint": true,
		"uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		// Built-in functions
		"append": true, "cap": true, "close": true, "complex": true, "copy": true,
		"delete": true, "imag": true, "len": true, "make": true, "new": true,
		"panic": true, "print": true, "println": true, "real": true, "recover": true,
	}

	return keywords[s]
}

// inferRootType infers the financial root type based on module name
func inferRootType(moduleName string) string {
	lowerName := strings.ToLower(moduleName)

	switch {
	case strings.Contains(lowerName, "asset") || strings.Contains(lowerName, "inventory"):
		return "RootTypeAsset"
	case strings.Contains(lowerName, "liability") || strings.Contains(lowerName, "payable"):
		return "RootTypeLiability"
	case strings.Contains(lowerName, "equity") || strings.Contains(lowerName, "capital"):
		return "RootTypeEquity"
	case strings.Contains(lowerName, "revenue") || strings.Contains(lowerName, "income"):
		return "RootTypeRevenue"
	case strings.Contains(lowerName, "expense") || strings.Contains(lowerName, "cost"):
		return "RootTypeExpense"
	default:
		return "RootTypeAsset" // Default to asset
	}
}

// generateDefaultPermissions generates default ABAC permissions for a module
func generateDefaultPermissions(name string) []string {
	entity := ToSnakeCase(Singularize(name))
	return []string{
		fmt.Sprintf("%s.create", entity),
		fmt.Sprintf("%s.read", entity),
		fmt.Sprintf("%s.update", entity),
		fmt.Sprintf("%s.delete", entity),
		fmt.Sprintf("%s.list", entity),
		fmt.Sprintf("%s.search", entity),
	}
}
