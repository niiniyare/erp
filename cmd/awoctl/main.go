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
  awoctl new feature finance/budgets
  awoctl new module hr --with-tests
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

The generated code follows ERP best practices including:
- Clean Architecture patterns
- ABAC security integration
- Feature flag support
- Observability (metrics, tracing, logging)
- Proper error handling`,
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

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview what would be generated without creating files")
	rootCmd.PersistentFlags().BoolVar(&withTests, "with-tests", false, "Generate test scaffolds alongside implementation")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(newModuleCmd)
	newCmd.AddCommand(newFeatureCmd)
}

func runNewModule(cmd *cobra.Command, args []string) error {
	moduleName := args[0]

	if verbose {
		fmt.Printf("Generating module: %s\n", moduleName)
		fmt.Printf("Dry run: %v\n", dryRun)
		fmt.Printf("With tests: %v\n", withTests)
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

	if !dryRun {
		fmt.Printf("✅ Successfully generated module: %s\n", moduleName)
		fmt.Printf("📂 Module created at: internal/core/%s/\n", moduleName)
		if withTests {
			fmt.Printf("🧪 Test scaffolds created\n")
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}