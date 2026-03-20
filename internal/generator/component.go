package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// ComponentGenerator generates individual components within existing modules

// ComponentType represents the type of component to generate
type ComponentType string

const (
	ComponentTypeEntity     ComponentType = "entity"
	ComponentTypeService    ComponentType = "service"
	ComponentTypeRepository ComponentType = "repository"
	ComponentTypeHandler    ComponentType = "handler"
	ComponentTypeDTO        ComponentType = "dto"
	ComponentTypeMiddleware ComponentType = "middleware"
	ComponentTypeValidator  ComponentType = "validator"
	ComponentTypeMapper     ComponentType = "mapper"
)

// ComponentGenerator generates individual components
type ComponentGenerator struct {
	fs   *FileSystemOperations
	tmpl *template.Template
}

// ComponentConfig represents configuration for component generation
type ComponentConfig struct {
	ModuleName    string        // Existing module name
	ComponentName string        // Component name (e.g., "user_profile")
	ComponentType ComponentType // Type of component to generate
	WithTests     bool          // Generate test scaffolds
	DryRun        bool          // Preview mode
	Verbose       bool          // Verbose output
	Overwrite     bool          // Overwrite existing files
	SkipExisting  bool          // Skip if file already exists
}

// NewComponentGenerator creates a new component generator
func NewComponentGenerator() (*ComponentGenerator, error) {
	// Load templates for components
	tmpl, err := template.New("component").Funcs(TemplateFunctions()).ParseFS(templateFS,
		"templates/component/*",
		"templates/domain/*",
		"templates/service/*",
		"templates/repository/*",
		"templates/api/*",
		"templates/test/*")
	if err != nil {
		return nil, fmt.Errorf("failed to parse component templates: %w", err)
	}

	return &ComponentGenerator{
		tmpl: tmpl,
	}, nil
}

// Generate generates a specific component within an existing module
func (cg *ComponentGenerator) Generate(config ComponentConfig) error {
	// Initialize file system operations
	cg.fs = NewFileSystemOperations(config.DryRun, config.Verbose)
	cg.fs.SetOverwriteMode(config.Overwrite)
	cg.fs.SetSkipExistingMode(config.SkipExisting)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid component configuration: %w", err)
	}

	// Validate project structure
	if err := cg.fs.ValidateProjectStructure(); err != nil {
		return fmt.Errorf("invalid project structure: %w", err)
	}

	// Check if module exists
	if !cg.fs.CheckModuleExists(config.ModuleName) {
		return fmt.Errorf("module '%s' does not exist", config.ModuleName)
	}

	// Convert config to template data
	data := config.ToTemplateData()

	if config.Verbose {
		fmt.Printf("Generating %s component '%s' in module '%s'\n",
			config.ComponentType, config.ComponentName, config.ModuleName)
	}

	// Generate component based on type
	switch config.ComponentType {
	case ComponentTypeEntity:
		return cg.generateEntity(data)
	case ComponentTypeService:
		return cg.generateService(data)
	case ComponentTypeRepository:
		return cg.generateRepository(data)
	case ComponentTypeHandler:
		return cg.generateHandler(data)
	case ComponentTypeDTO:
		return cg.generateDTO(data)
	case ComponentTypeMiddleware:
		return cg.generateMiddleware(data)
	case ComponentTypeValidator:
		return cg.generateValidator(data)
	case ComponentTypeMapper:
		return cg.generateMapper(data)
	default:
		return fmt.Errorf("unsupported component type: %s", config.ComponentType)
	}
}

// =============================================================================
// COMPONENT GENERATION METHODS
// =============================================================================

func (cg *ComponentGenerator) generateEntity(data TemplateData) error {
	files := []FileTemplate{
		{
			TemplatePath: "templates/component/entity.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+".go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/entity_test.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+"_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateService(data TemplateData) error {
	files := []FileTemplate{
		{
			TemplatePath: "templates/component/service.go.tmpl",
			OutputPath:   filepath.Join(data.ServicePath, ToSnakeCase(data.ComponentName)+"_service.go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/service_test.go.tmpl",
			OutputPath:   filepath.Join(data.ServicePath, ToSnakeCase(data.ComponentName)+"_service_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateRepository(data TemplateData) error {
	files := []FileTemplate{
		{
			TemplatePath: "templates/component/repository.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, ToSnakeCase(data.ComponentName)+"_repository.go"),
			Data:         data,
		},
		{
			TemplatePath: "templates/component/repository_interface.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+"_repository.go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/repository_test.go.tmpl",
			OutputPath:   filepath.Join(data.RepositoryPath, ToSnakeCase(data.ComponentName)+"_repository_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateHandler(data TemplateData) error {
	handlerPath := filepath.Join("internal", "api", "handlers", ToSnakeCase(data.ModuleName))

	// Ensure handler directory exists
	if err := cg.fs.EnsureDir(handlerPath); err != nil {
		return fmt.Errorf("failed to create handler directory: %w", err)
	}

	files := []FileTemplate{
		{
			TemplatePath: "templates/component/handler.go.tmpl",
			OutputPath:   filepath.Join(handlerPath, ToSnakeCase(data.ComponentName)+"_handler.go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/handler_test.go.tmpl",
			OutputPath:   filepath.Join(handlerPath, ToSnakeCase(data.ComponentName)+"_handler_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateDTO(data TemplateData) error {
	files := []FileTemplate{
		{
			TemplatePath: "templates/component/dto.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+"_dto.go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/dto_test.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+"_dto_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateMiddleware(data TemplateData) error {
	middlewarePath := filepath.Join("internal", "api", "middleware")

	// Ensure middleware directory exists
	if err := cg.fs.EnsureDir(middlewarePath); err != nil {
		return fmt.Errorf("failed to create middleware directory: %w", err)
	}

	files := []FileTemplate{
		{
			TemplatePath: "templates/component/middleware.go.tmpl",
			OutputPath:   filepath.Join(middlewarePath, ToSnakeCase(data.ComponentName)+".go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/middleware_test.go.tmpl",
			OutputPath:   filepath.Join(middlewarePath, ToSnakeCase(data.ComponentName)+"_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateValidator(data TemplateData) error {
	files := []FileTemplate{
		{
			TemplatePath: "templates/component/validator.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+"_validator.go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/validator_test.go.tmpl",
			OutputPath:   filepath.Join(data.DomainPath, ToSnakeCase(data.ComponentName)+"_validator_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

func (cg *ComponentGenerator) generateMapper(data TemplateData) error {
	mapperPath := filepath.Join(data.RepositoryPath, "mappers")

	// Ensure mapper directory exists
	if err := cg.fs.EnsureDir(mapperPath); err != nil {
		return fmt.Errorf("failed to create mapper directory: %w", err)
	}

	files := []FileTemplate{
		{
			TemplatePath: "templates/component/mapper.go.tmpl",
			OutputPath:   filepath.Join(mapperPath, ToSnakeCase(data.ComponentName)+"_mapper.go"),
			Data:         data,
		},
	}

	if data.WithTests {
		files = append(files, FileTemplate{
			TemplatePath: "templates/component/mapper_test.go.tmpl",
			OutputPath:   filepath.Join(mapperPath, ToSnakeCase(data.ComponentName)+"_mapper_test.go"),
			Data:         data,
		})
	}

	return cg.generateFiles(files)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

func (cg *ComponentGenerator) generateFiles(files []FileTemplate) error {
	for _, file := range files {
		if err := cg.generateFile(file); err != nil {
			return fmt.Errorf("failed to generate file %s: %w", file.OutputPath, err)
		}
	}
	return nil
}

func (cg *ComponentGenerator) generateFile(file FileTemplate) error {
	// Check if file exists and handle accordingly
	if cg.fs.FileExists(file.OutputPath) {
		if cg.fs.ShouldSkipExisting() {
			if cg.fs.IsVerbose() {
				fmt.Printf("Skipping existing file: %s\n", file.OutputPath)
			}
			return nil
		}

		if !cg.fs.ShouldOverwrite() {
			return fmt.Errorf("file already exists (use --overwrite to replace): %s", file.OutputPath)
		}
	}

	// Execute template
	var content strings.Builder
	if err := cg.tmpl.ExecuteTemplate(&content, filepath.Base(file.TemplatePath), file.Data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", file.TemplatePath, err)
	}

	// Create the file
	if err := cg.fs.CreateFile(file.OutputPath, content.String()); err != nil {
		return fmt.Errorf("failed to create file %s: %w", file.OutputPath, err)
	}

	return nil
}

// =============================================================================
// CONFIGURATION VALIDATION
// =============================================================================

// Validate validates the component configuration
func (c *ComponentConfig) Validate() error {
	if c.ModuleName == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	if c.ComponentName == "" {
		return fmt.Errorf("component name cannot be empty")
	}

	// Validate module name format
	if !isValidIdentifier(c.ModuleName) {
		return fmt.Errorf("module name '%s' is not a valid Go identifier", c.ModuleName)
	}

	// Validate component name format
	if !isValidIdentifier(c.ComponentName) {
		return fmt.Errorf("component name '%s' is not a valid Go identifier", c.ComponentName)
	}

	// Validate component type
	validTypes := []ComponentType{
		ComponentTypeEntity,
		ComponentTypeService,
		ComponentTypeRepository,
		ComponentTypeHandler,
		ComponentTypeDTO,
		ComponentTypeMiddleware,
		ComponentTypeValidator,
		ComponentTypeMapper,
	}

	isValidType := false
	for _, validType := range validTypes {
		if c.ComponentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return fmt.Errorf("invalid component type '%s'", c.ComponentType)
	}

	// Validate mutually exclusive flags
	if c.Overwrite && c.SkipExisting {
		return fmt.Errorf("--overwrite and --skip-existing flags are mutually exclusive")
	}

	return nil
}

// ToTemplateData converts ComponentConfig to TemplateData
func (c *ComponentConfig) ToTemplateData() TemplateData {
	return TemplateData{
		ModuleName:        c.ModuleName,
		ModuleNamePascal:  ToPascalCase(c.ModuleName),
		ModuleNameCamel:   ToCamelCase(c.ModuleName),
		ModuleNameSnake:   ToSnakeCase(c.ModuleName),
		ModuleNameKebab:   ToKebabCase(c.ModuleName),
		ModuleNamePlural:  Pluralize(c.ModuleName),
		ModuleDescription: fmt.Sprintf("%s management module", ToPascalCase(c.ModuleName)),

		ComponentName:       c.ComponentName,
		ComponentNamePascal: ToPascalCase(c.ComponentName),
		ComponentNameCamel:  ToCamelCase(c.ComponentName),
		ComponentNameSnake:  ToSnakeCase(c.ComponentName),

		EntityName:       ToPascalCase(Singularize(c.ComponentName)),
		EntityNamePascal: ToPascalCase(Singularize(c.ComponentName)),
		EntityNameCamel:  ToCamelCase(Singularize(c.ComponentName)),
		EntityNameSnake:  ToSnakeCase(Singularize(c.ComponentName)),
		EntityNamePlural: ToPascalCase(Pluralize(c.ComponentName)),

		PackageName: ToSnakeCase(c.ModuleName),
		ImportPath:  fmt.Sprintf("awo/internal/core/%s", ToSnakeCase(c.ModuleName)),

		GeneratedAt:      time.Now(),
		GeneratorVersion: "0.1.0",
		WithTests:        c.WithTests,

		RootType:           inferRootType(c.ModuleName),
		DefaultPermissions: generateDefaultPermissions(c.ComponentName),

		DomainPath:     filepath.Join("internal", "core", ToSnakeCase(c.ModuleName), "domain"),
		ServicePath:    filepath.Join("internal", "core", ToSnakeCase(c.ModuleName)),
		RepositoryPath: filepath.Join("internal", "core", ToSnakeCase(c.ModuleName), "repository"),
		APIPath:        filepath.Join("internal", "api", "design", "services", ToSnakeCase(c.ModuleName)),
	}
}

// TODO: Add support for generating multiple related components at once
// TODO: Add dependency analysis for component generation
// TODO: Add component template validation
// TODO: Add support for custom component templates
