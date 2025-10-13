package main

import (
	"fmt"
	"os"

	"github.com/niiniyare/erp/internal/generator"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	dryRun    bool
	withTests bool
	withDocs  bool
	verbose   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awoctl",
	Short: "AWO ERP scaffolding tool",
	Long: `awoctl is a scaffolding tool for the AWO ERP system.
It generates boilerplate code following established patterns for:
- Modules (complete domain, service, repository, API layers)
- Features (additions to existing modules)
- Tests (unit and integration test scaffolds)

Examples:
  awoctl new module inventory
  awoctl new module finance --with-docs
  awoctl new module hr --with-tests --with-docs
  awoctl new module projects --dry-run`,
	Version: version,
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate new components",
	Long:  `Generate new modules, features, or other components with proper layered architecture.`,
}

// newModuleCmd represents the new module command
var newModuleCmd = &cobra.Command{
	Use:   "module [name]",
	Short: "Generate a new module with complete layered architecture",
	Long: `Generate a new module with:
- Domain layer (entities, validation, types)
- Service layer (business logic, interfaces)
- Repository layer (data access, SQLC integration)
- API layer (Goa DSL, handlers)
- Test scaffolds (if --with-tests flag is used)
- Complete documentation (if --with-docs flag is used)

The generated code follows ERP best practices including:
- Clean Architecture patterns
- ABAC security integration
- Feature flag support
- Observability (metrics, tracing, logging)
- Proper error handling

Documentation includes:
- Module README with architecture overview
- Complete API reference with examples
- Comprehensive testing strategy
- Detailed architecture guide`,
	Args: cobra.ExactArgs(1),
	RunE: runNewModule,
}

// newFeatureCmd represents the new feature command
var newFeatureCmd = &cobra.Command{
	Use:   "feature [module/feature]",
	Short: "Add a new feature to an existing module",
	Long: `Add a new feature to an existing module.
Format: module_name/feature_name

Example:
  awoctl new feature finance/reconciliation
  awoctl new feature inventory/adjustments`,
	Args: cobra.ExactArgs(1),
	RunE: runNewFeature,
}

// newComponentCmd represents the new component command
var newComponentCmd = &cobra.Command{
	Use:   "component [module] [component_name] [type]",
	Short: "Generate a specific component within an existing module",
	Long: `Generate a specific component within an existing module.

Available component types:
- entity: Domain entity with validation
- service: Business logic service
- repository: Data access layer
- handler: API handler
- dto: Data transfer object
- middleware: HTTP middleware
- validator: Input validator
- mapper: Data mapper

Examples:
  awoctl new component finance payment entity
  awoctl new component inventory item_adjustment service
  awoctl new component hr employee_profile repository`,
	Args: cobra.ExactArgs(3),
	RunE: runNewComponent,
}

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation for modules",
	Long:  `Generate documentation for modules including README, API reference, testing guide, and architecture documentation.`,
}

// newDocsCmd represents the new docs command
var newDocsCmd = &cobra.Command{
	Use:   "module [module_name]",
	Short: "Generate complete documentation set for a module",
	Long: `Generate a complete documentation set for a module including:
- README.md: Module overview and quick start guide
- api-reference.md: Complete API documentation with examples
- testing.md: Testing strategy and test cases
- architecture-guide.md: Detailed architecture documentation

Examples:
  awoctl docs module finance
  awoctl docs module inventory --dry-run
  awoctl docs module hr --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runNewDocs,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview what would be generated without creating files")
	rootCmd.PersistentFlags().BoolVar(&withTests, "with-tests", false, "Generate test scaffolds alongside implementation")
	rootCmd.PersistentFlags().BoolVar(&withDocs, "with-docs", false, "Generate comprehensive documentation alongside implementation")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(newModuleCmd)
	newCmd.AddCommand(newFeatureCmd)
	newCmd.AddCommand(newComponentCmd)

	rootCmd.AddCommand(docsCmd)
	docsCmd.AddCommand(newDocsCmd)
}

func runNewModule(cmd *cobra.Command, args []string) error {
	moduleName := args[0]

	if verbose {
		fmt.Printf("Generating module: %s\n", moduleName)
		fmt.Printf("Dry run: %v\n", dryRun)
		fmt.Printf("With tests: %v\n", withTests)
		fmt.Printf("With docs: %v\n", withDocs)
	}

	config := generator.ModuleConfig{
		Name:      moduleName,
		WithTests: withTests,
		DryRun:    dryRun,
		Verbose:   verbose,
	}

	gen, err := generator.NewModuleGenerator()
	if err != nil {
		return fmt.Errorf("failed to create module generator: %w", err)
	}

	if err := gen.Generate(config); err != nil {
		return fmt.Errorf("failed to generate module %s: %w", moduleName, err)
	}

	// Generate documentation if requested
	if withDocs {
		if verbose {
			fmt.Printf("Generating documentation for module: %s\n", moduleName)
		}

		docsConfig := generator.DocsConfig{
			ModuleName: moduleName,
			DryRun:     dryRun,
			Verbose:    verbose,
		}

		docsGen, err := generator.NewDocsGenerator()
		if err != nil {
			return fmt.Errorf("failed to create docs generator: %w", err)
		}

		if err := docsGen.Generate(docsConfig); err != nil {
			return fmt.Errorf("failed to generate documentation for module %s: %w", moduleName, err)
		}
	}

	if !dryRun {
		fmt.Printf("✅ Successfully generated module: %s\n", moduleName)
		fmt.Printf("📂 Module created at: internal/core/%s/\n", moduleName)
		if withTests {
			fmt.Printf("🧪 Test scaffolds created\n")
		}
		if withDocs {
			fmt.Printf("📚 Documentation generated at: docs/reference/modules/%s/\n", moduleName)
		}
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Update domain models in internal/core/%s/domain/\n", moduleName)
		fmt.Printf("2. Implement business logic in internal/core/%s/service/\n", moduleName)
		fmt.Printf("3. Create database queries in db/queries/%s.sql\n", moduleName)
		fmt.Printf("4. Run 'make sqlc' to generate repository code\n")
		fmt.Printf("5. Run 'make goa' to generate API types\n")
	} else {
		fmt.Printf("🔍 Dry run completed for module: %s\n", moduleName)
	}

	return nil
}

func runNewFeature(cmd *cobra.Command, args []string) error {
	featurePath := args[0]

	if verbose {
		fmt.Printf("Generating feature: %s\n", featurePath)
		fmt.Printf("Dry run: %v\n", dryRun)
		fmt.Printf("With tests: %v\n", withTests)
	}

	config := generator.FeatureConfig{
		Path:      featurePath,
		WithTests: withTests,
		DryRun:    dryRun,
		Verbose:   verbose,
	}

	gen, err := generator.NewFeatureGenerator()
	if err != nil {
		return fmt.Errorf("failed to create feature generator: %w", err)
	}

	if err := gen.Generate(config); err != nil {
		return fmt.Errorf("failed to generate feature %s: %w", featurePath, err)
	}

	if !dryRun {
		fmt.Printf("✅ Successfully generated feature: %s\n", featurePath)
		if withTests {
			fmt.Printf("🧪 Test scaffolds created\n")
		}
	} else {
		fmt.Printf("🔍 Dry run completed for feature: %s\n", featurePath)
	}

	return nil
}

func runNewComponent(cmd *cobra.Command, args []string) error {
	moduleName := args[0]
	componentName := args[1]
	componentTypeStr := args[2]

	// Parse component type
	var componentType generator.ComponentType
	switch componentTypeStr {
	case "entity":
		componentType = generator.ComponentTypeEntity
	case "service":
		componentType = generator.ComponentTypeService
	case "repository":
		componentType = generator.ComponentTypeRepository
	case "handler":
		componentType = generator.ComponentTypeHandler
	case "dto":
		componentType = generator.ComponentTypeDTO
	case "middleware":
		componentType = generator.ComponentTypeMiddleware
	case "validator":
		componentType = generator.ComponentTypeValidator
	case "mapper":
		componentType = generator.ComponentTypeMapper
	default:
		return fmt.Errorf("invalid component type '%s'. Valid types: entity, service, repository, handler, dto, middleware, validator, mapper", componentTypeStr)
	}

	if verbose {
		fmt.Printf("Generating %s component '%s' in module '%s'\n", componentType, componentName, moduleName)
		fmt.Printf("Dry run: %v\n", dryRun)
		fmt.Printf("With tests: %v\n", withTests)
	}

	config := generator.ComponentConfig{
		ModuleName:    moduleName,
		ComponentName: componentName,
		ComponentType: componentType,
		WithTests:     withTests,
		DryRun:        dryRun,
		Verbose:       verbose,
	}

	gen, err := generator.NewComponentGenerator()
	if err != nil {
		return fmt.Errorf("failed to create component generator: %w", err)
	}

	if err := gen.Generate(config); err != nil {
		return fmt.Errorf("failed to generate component %s: %w", componentName, err)
	}

	if !dryRun {
		fmt.Printf("✅ Successfully generated %s component: %s\n", componentType, componentName)
		if withTests {
			fmt.Printf("🧪 Test scaffolds created\n")
		}
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Review generated code in internal/core/%s/\n", moduleName)
		fmt.Printf("2. Implement business logic as needed\n")
		fmt.Printf("3. Run 'make goa sqlc' to update generated code\n")
	} else {
		fmt.Printf("🔍 Dry run completed for %s component: %s\n", componentType, componentName)
	}

	return nil
}

func runNewDocs(cmd *cobra.Command, args []string) error {
	moduleName := args[0]

	if verbose {
		fmt.Printf("Generating documentation for module: %s\n", moduleName)
		fmt.Printf("Dry run: %v\n", dryRun)
	}

	config := generator.DocsConfig{
		ModuleName: moduleName,
		DryRun:     dryRun,
		Verbose:    verbose,
	}

	gen, err := generator.NewDocsGenerator()
	if err != nil {
		return fmt.Errorf("failed to create docs generator: %w", err)
	}

	if err := gen.Generate(config); err != nil {
		return fmt.Errorf("failed to generate documentation for module %s: %w", moduleName, err)
	}

	if !dryRun {
		fmt.Printf("✅ Successfully generated documentation for module: %s\n", moduleName)
		fmt.Printf("📚 Documentation created at: docs/reference/modules/%s/\n", moduleName)
		fmt.Printf("\nGenerated files:\n")
		fmt.Printf("- README.md: Module overview and quick start\n")
		fmt.Printf("- api-reference.md: Complete API documentation\n")
		fmt.Printf("- testing.md: Testing strategy and test cases\n")
		fmt.Printf("- architecture-guide.md: Detailed architecture documentation\n")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Review and customize the generated documentation\n")
		fmt.Printf("2. Update API examples with actual endpoint data\n")
		fmt.Printf("3. Add module-specific business rules and constraints\n")
		fmt.Printf("4. Update integration points and dependencies\n")
	} else {
		fmt.Printf("🔍 Dry run completed for module documentation: %s\n", moduleName)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
