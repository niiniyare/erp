package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// ModuleGenerator generates complete module structures
type ModuleGenerator struct {
	fs   *FileSystemOperations
	tmpl *template.Template
}

// FeatureGenerator generates features within existing modules
type FeatureGenerator struct {
	fs   *FileSystemOperations
	tmpl *template.Template
}

// NewModuleGenerator creates a new module generator
func NewModuleGenerator() (*ModuleGenerator, error) {
	// Load templates
	tmpl, err := template.New("module").Funcs(TemplateFunctions()).ParseFS(templateFS, "templates/module/*", "templates/domain/*", "templates/service/*", "templates/repository/*", "templates/api/*", "templates/database/*", "templates/test/*")
	if err != nil {
		return nil, fmt.Errorf("failed to parse module templates: %w", err)
	}

	return &ModuleGenerator{
		tmpl: tmpl,
	}, nil
}

// NewFeatureGenerator creates a new feature generator
func NewFeatureGenerator() (*FeatureGenerator, error) {
	// Load feature templates
	tmpl, err := template.New("feature").Funcs(TemplateFunctions()).ParseFS(templateFS, "templates/feature/*")
	if err != nil {
		return nil, fmt.Errorf("failed to parse feature templates: %w", err)
	}

	return &FeatureGenerator{
		tmpl: tmpl,
	}, nil
}

// Generate generates a complete module structure
func (mg *ModuleGenerator) Generate(config ModuleConfig) error {
	// Initialize file system operations
	mg.fs = NewFileSystemOperations(config.DryRun, config.Verbose)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid module configuration: %w", err)
	}

	// Validate project structure
	if err := mg.fs.ValidateProjectStructure(); err != nil {
		return fmt.Errorf("invalid project structure: %w", err)
	}

	// Validate module path
	if err := mg.fs.ValidateModulePath(config.Name); err != nil {
		return fmt.Errorf("invalid module name: %w", err)
	}

	// Check if module already exists
	if mg.fs.CheckModuleExists(config.Name) {
		return fmt.Errorf("module '%s' already exists", config.Name)
	}

	// Convert config to template data
	data := config.ToTemplateData()

	if config.Verbose {
		fmt.Printf("Generating module '%s' with entity '%s'\n", data.ModuleName, data.EntityName)
	}

	// Generate module structure
	if err := mg.generateModuleStructure(data); err != nil {
		return fmt.Errorf("failed to generate module structure: %w", err)
	}

	return nil
}

// generateModuleStructure generates the complete module directory structure and files
func (mg *ModuleGenerator) generateModuleStructure(data TemplateData) error {
	// Create module directories
	if err := mg.createModuleDirectories(data); err != nil {
		return fmt.Errorf("failed to create module directories: %w", err)
	}

	// Generate domain layer
	if err := mg.generateDomainLayer(data); err != nil {
		return fmt.Errorf("failed to generate domain layer: %w", err)
	}

	// Generate service layer
	if err := mg.generateServiceLayer(data); err != nil {
		return fmt.Errorf("failed to generate service layer: %w", err)
	}

	// Generate repository layer
	if err := mg.generateRepositoryLayer(data); err != nil {
		return fmt.Errorf("failed to generate repository layer: %w", err)
	}

	// Generate API layer
	if err := mg.generateAPILayer(data); err != nil {
		return fmt.Errorf("failed to generate API layer: %w", err)
	}

	// Generate database queries
	if err := mg.generateDatabaseQueries(data); err != nil {
		return fmt.Errorf("failed to generate database queries: %w", err)
	}

	// Generate tests if requested
	if data.WithTests {
		if err := mg.generateTestStructure(data); err != nil {
			return fmt.Errorf("failed to generate test structure: %w", err)
		}
	}

	return nil
}

// createModuleDirectories creates the module directory structure
func (mg *ModuleGenerator) createModuleDirectories(data TemplateData) error {
	directories := []string{
		data.DomainPath,     // Domain subdirectory
		data.ServicePath,    // Service files in module root
		data.RepositoryPath, // Repository subdirectory
		data.APIPath,        // API design directory
		filepath.Join(data.ServicePath, "activities"), // Temporal activities
		filepath.Join(data.ServicePath, "workflows"),  // Temporal workflows
		filepath.Join(data.RepositoryPath, "mappers"), // SQLC mappers
	}

	if data.WithTests {
		directories = append(directories, []string{
			filepath.Join(data.ServicePath, "test"),      // Test directory in module root
			filepath.Join(data.DomainPath, "..", "test"), // Alternative test location
		}...)
	}

	for _, dir := range directories {
		if err := mg.fs.EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateDomainLayer generates domain layer files
func (mg *ModuleGenerator) generateDomainLayer(data TemplateData) error {
	domainFiles := []FileTemplate{
		{
			TemplatePath: "templates/domain/entity.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.EntityName)+".go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/domain/types.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, "types.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/domain/errors.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, "errors.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/domain/validation.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, "validation.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/domain/repository.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, "repository.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/domain/requests.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, "requests.go"),
			Data:         data,
		},
	}

	for _, file := range domainFiles {
		if err := mg.generateFile(file); err != nil {
			return fmt.Errorf("failed to generate domain file %s: %w", file.OutputPath, err)
		}
	}

	return nil
}

// generateServiceLayer generates service layer files
func (mg *ModuleGenerator) generateServiceLayer(data TemplateData) error {
	serviceFiles := []FileTemplate{
		{
			TemplatePath: "templates/service/service.go.tmpl",
			OutputPath:   filepath.Join(data.ServicePath, ToSnakeCase(data.EntityName)+"_service.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/service/settings_helper.go.tmpl",
			OutputPath:   filepath.Join(data.ServicePath, "settings_helper.go"),
			Data:         data,
		},
	}

	for _, file := range serviceFiles {
		if err := mg.generateFile(file); err != nil {
			return fmt.Errorf("failed to generate service file %s: %w", file.OutputPath, err)
		}
	}

	return nil
}

// generateRepositoryLayer generates repository layer files
func (mg *ModuleGenerator) generateRepositoryLayer(data TemplateData) error {
	repositoryFiles := []FileTemplate{
		{
			TemplatePath: "templates/repository/interface.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, "interface.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/repository/implementation.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, ToSnakeCase(data.EntityName)+"_repository.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/repository/mappers.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, "mappers", ToSnakeCase(data.EntityName)+"_mappers.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/repository/filters.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, "filters.go"),
			Data:         data,
		},
	}

	for _, file := range repositoryFiles {
		if err := mg.generateFile(file); err != nil {
			return fmt.Errorf("failed to generate repository file %s: %w", file.OutputPath, err)
		}
	}

	return nil
}

// generateAPILayer generates API layer files
func (mg *ModuleGenerator) generateAPILayer(data TemplateData) error {
	apiFiles := []FileTemplate{
		{
			TemplatePath: "templates/api/service.go.tmpl",
			OutputPath:   filepath.Join(data.APIPath, ToSnakeCase(data.EntityName)+".go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/api/types.go.tmpl",
			OutputPath:   filepath.Join(data.APIPath, "types.go"),
			Data:         data,
		},
	}

	// Create API handler directory path
	handlerPath := filepath.Join("internal", "api", "handlers", ToSnakeCase(data.ModuleName))
	if err := mg.fs.EnsureDir(handlerPath); err != nil {
		return fmt.Errorf("failed to create handler directory: %w", err)
	}

	// Add handler file
	apiFiles = append(apiFiles, FileTemplate{
		TemplatePath: "templates/api/handler.go.tmpl",
		OutputPath:   filepath.Join(handlerPath, ToSnakeCase(data.EntityName)+"_handler.go"),
		Data:         data,
	})

	for _, file := range apiFiles {
		if err := mg.generateFile(file); err != nil {
			return fmt.Errorf("failed to generate API file %s: %w", file.OutputPath, err)
		}
	}

	return nil
}

// generateDatabaseQueries generates SQLC query files
func (mg *ModuleGenerator) generateDatabaseQueries(data TemplateData) error {
	queryPath := filepath.Join("db", "queries", ToSnakeCase(data.ModuleName)+".sql")

	queryFile := FileTemplate{
		TemplatePath: "templates/database/queries.sql.tmpl",
		OutputPath:   queryPath,
		Data:         data,
	}

	if err := mg.generateFile(queryFile); err != nil {
		return fmt.Errorf("failed to generate database queries: %w", err)
	}

	return nil
}

// generateTestStructure generates test scaffolds
func (mg *ModuleGenerator) generateTestStructure(data TemplateData) error {
	testFiles := []FileTemplate{
		{
			TemplatePath: "templates/test/domain_test.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.EntityName)+"_test.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/test/service_test.go.tmpl",
			OutputPath:   filepath.Join(data.ServicePath, ToSnakeCase(data.EntityName)+"_test.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/test/repository_test.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, ToSnakeCase(data.EntityName)+"_test.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/test/integration_test.go.tmpl",
			OutputPath:   filepath.Join("internal", "core", ToSnakeCase(data.ModuleName), "test", "integration_test.go"),
			Data:         data,
		},
	}

	for _, file := range testFiles {
		if err := mg.generateFile(file); err != nil {
			return fmt.Errorf("failed to generate test file %s: %w", file.OutputPath, err)
		}
	}

	return nil
}

// generateFile generates a single file from template
func (mg *ModuleGenerator) generateFile(file FileTemplate) error {
	// Execute template
	var content strings.Builder
	if err := mg.tmpl.ExecuteTemplate(&content, filepath.Base(file.TemplatePath), file.Data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", file.TemplatePath, err)
	}

	// Create the file
	if err := mg.fs.CreateFile(file.OutputPath, content.String()); err != nil {
		return fmt.Errorf("failed to create file %s: %w", file.OutputPath, err)
	}

	return nil
}

// Generate generates a feature within an existing module
func (fg *FeatureGenerator) Generate(config FeatureConfig) error {
	// Initialize file system operations
	fg.fs = NewFileSystemOperations(config.DryRun, config.Verbose)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid feature configuration: %w", err)
	}

	// Parse module and feature from path
	parts := strings.Split(config.Path, "/")
	moduleName, featureName := parts[0], parts[1]

	// Check if module exists
	if !fg.fs.CheckModuleExists(moduleName) {
		return fmt.Errorf("module '%s' does not exist", moduleName)
	}

	// Convert config to template data
	data := config.ToTemplateData()

	if config.Verbose {
		fmt.Printf("Generating feature '%s' in module '%s'\n", featureName, moduleName)
	}

	// Generate feature files
	if err := fg.generateFeatureFiles(data); err != nil {
		return fmt.Errorf("failed to generate feature files: %w", err)
	}

	return nil
}

// generateFeatureFiles generates files for a new feature
func (fg *FeatureGenerator) generateFeatureFiles(data TemplateData) error {
	featureFiles := []FileTemplate{
		{
			TemplatePath: "templates/feature/entity.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.FeatureName)+".go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/feature/service.go.tmpl",
			OutputPath:   filepath.Join(data.ServicePath, ToSnakeCase(data.FeatureName)+"_service.go"),
			Data:         data,
		},
	}

	// Add test files if requested
	if data.WithTests {
		featureFiles = append(featureFiles, []FileTemplate{
			{
				TemplatePath: "templates/feature/entity_test.go.tmpl",
				OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.FeatureName)+"_test.go"),
				Data:         data,
			},
			{
				TemplatePath: "templates/feature/service_test.go.tmpl",
				OutputPath:   filepath.Join(data.ServicePath, ToSnakeCase(data.FeatureName)+"_service_test.go"),
				Data:         data,
			},
		}...)
	}

	for _, file := range featureFiles {
		if err := fg.generateFeatureFile(file); err != nil {
			return fmt.Errorf("failed to generate feature file %s: %w", file.OutputPath, err)
		}
	}

	return nil
}

// generateFeatureFile generates a single feature file from template
func (fg *FeatureGenerator) generateFeatureFile(file FileTemplate) error {
	// Execute template
	var content strings.Builder
	if err := fg.tmpl.ExecuteTemplate(&content, filepath.Base(file.TemplatePath), file.Data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", file.TemplatePath, err)
	}

	// Create the file
	if err := fg.fs.CreateFile(file.OutputPath, content.String()); err != nil {
		return fmt.Errorf("failed to create file %s: %w", file.OutputPath, err)
	}

	return nil
}
