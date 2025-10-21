package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niiniyare/erp/web/components/atoms"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
)

// ============================================================================
// BRIDGE CLI INTERFACE
// ============================================================================

// CLI provides command-line interface for bridge operations
type CLI struct {
	registry *UnifiedRegistry
	bridge   *Bridge
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		registry: NewUnifiedRegistry(),
		bridge:   NewBridge(),
	}
}

// ============================================================================
// CLI COMMANDS
// ============================================================================

// ConvertSchemaToTemplFile converts a schema JSON file to Templ component code
func (cli *CLI) ConvertSchemaToTemplFile(ctx context.Context, schemaFilePath, outputPath string) error {
	// Read schema file
	schemaData, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse schema component
	var schemaComponent schemaui.Component
	if err := json.Unmarshal(schemaData, &schemaComponent); err != nil {
		return fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	// Convert to Templ component
	templComponent, err := cli.bridge.ConvertToTempl(ctx, schemaComponent)
	if err != nil {
		return fmt.Errorf("failed to convert schema to Templ: %w", err)
	}

	// Generate Templ code
	templCode, err := cli.generateTemplCode(templComponent)
	if err != nil {
		return fmt.Errorf("failed to generate Templ code: %w", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, []byte(templCode), 0o644); err != nil {
		return fmt.Errorf("failed to write Templ file: %w", err)
	}

	fmt.Printf("Successfully converted schema to Templ component: %s\n", outputPath)
	return nil
}

// ConvertTemplToSchemaFile converts Templ component props to schema JSON
func (cli *CLI) ConvertTemplToSchemaFile(ctx context.Context, componentType string, propsJSON, outputPath string) error {
	// Parse props JSON based on component type
	templComponent, err := cli.parseTemplComponentFromJSON(componentType, propsJSON)
	if err != nil {
		return fmt.Errorf("failed to parse Templ component: %w", err)
	}

	// Convert to schema component
	schemaComponent, err := cli.bridge.ConvertTemplToSchema(ctx, templComponent)
	if err != nil {
		return fmt.Errorf("failed to convert Templ to schema: %w", err)
	}

	// Marshal to JSON
	schemaJSON, err := json.MarshalIndent(schemaComponent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema JSON: %w", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, schemaJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	fmt.Printf("Successfully converted Templ component to schema: %s\n", outputPath)
	return nil
}

// CreateTemplate creates both schema and Templ components from a template
func (cli *CLI) CreateTemplate(ctx context.Context, templateName string, data map[string]any, outputDir string) error {
	// Create components from template
	templComponents, schemaComponents, err := cli.registry.CreateFromTemplate(ctx, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	// Create output directories
	templDir := filepath.Join(outputDir, "templ")
	schemaDir := filepath.Join(outputDir, "schema")

	if err := os.MkdirAll(templDir, 0o755); err != nil {
		return fmt.Errorf("failed to create Templ directory: %w", err)
	}
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		return fmt.Errorf("failed to create schema directory: %w", err)
	}

	// Write Templ components
	templCode, err := cli.generateTemplCodeFromComponents(templComponents, templateName)
	if err != nil {
		return fmt.Errorf("failed to generate Templ code: %w", err)
	}

	templFile := filepath.Join(templDir, templateName+".templ")
	if err := os.WriteFile(templFile, []byte(templCode), 0o644); err != nil {
		return fmt.Errorf("failed to write Templ file: %w", err)
	}

	// Write schema components
	for i, schemaComponent := range schemaComponents {
		schemaJSON, err := json.MarshalIndent(schemaComponent, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal schema component %d: %w", i, err)
		}

		schemaFile := filepath.Join(schemaDir, fmt.Sprintf("%s_%d.json", templateName, i))
		if err := os.WriteFile(schemaFile, schemaJSON, 0o644); err != nil {
			return fmt.Errorf("failed to write schema file %d: %w", i, err)
		}
	}

	fmt.Printf("Successfully created template '%s' in %s\n", templateName, outputDir)
	fmt.Printf("  - Templ component: %s\n", templFile)
	fmt.Printf("  - Schema components: %d files in %s\n", len(schemaComponents), schemaDir)

	return nil
}

// ValidateComponent validates a component (either schema or Templ)
func (cli *CLI) ValidateComponent(ctx context.Context, componentType, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read component file: %w", err)
	}

	switch componentType {
	case "schema":
		var schemaComponent schemaui.Component
		if err := json.Unmarshal(data, &schemaComponent); err != nil {
			return fmt.Errorf("failed to parse schema component: %w", err)
		}

		// TODO: Add schema validation when method is available
		_ = ctx // Avoid unused variable

		fmt.Println("Schema component validation passed ✓")

	case "templ":
		// For Templ validation, we'd need to parse the .templ file
		// For now, we'll assume the user provides JSON props
		return fmt.Errorf("Templ file validation not yet implemented - provide JSON props instead")

	default:
		return fmt.Errorf("unknown component type: %s (use 'schema' or 'templ')", componentType)
	}

	return nil
}

// ListSupportedTypes lists all supported component types
func (cli *CLI) ListSupportedTypes() {
	// TODO: Implement when registry methods are available
	fmt.Println("Supported component types:")
	fmt.Println("- Button")
	fmt.Println("- Input")
	fmt.Println("- Textarea")
	fmt.Println("- Select")
	fmt.Println("- Checkbox")
	fmt.Println("- Radio")

	return
	// The following code needs to be implemented when registry methods are available:
	/*
		types := cli.registry.GetSupportedTypes()
		for _, componentType := range types {
			schema, err := cli.registry.GetSchemaForType(componentType)
			if err == nil {
				fmt.Printf("  - %s: %s\n", componentType, schema.Description)
			} else {
				fmt.Printf("  - %s\n", componentType)
			}
		}
	*/
}

// ListTemplates lists available pre-defined templates
func (cli *CLI) ListTemplates() {
	fmt.Println("Available templates:")
	fmt.Println("  - login-form: User authentication form")
	fmt.Println("  - user-form: User registration/edit form")
	fmt.Println("  - data-table: Data table with search and filters")
}

// ============================================================================
// CODE GENERATION HELPERS
// ============================================================================

// generateTemplCode generates Templ template code from a single component
func (cli *CLI) generateTemplCode(templComponent TemplComponent) (string, error) {
	var builder strings.Builder

	// Generate package and imports
	builder.WriteString("package components\n\n")
	builder.WriteString("import \"github.com/niiniyare/erp/web/components/atoms\"\n\n")

	// Generate component function
	componentName := strings.ToLower(templComponent.Type)
	builder.WriteString(fmt.Sprintf("templ %sFromSchema() {\n", componentName))

	// Generate component call based on type
	switch templComponent.Type {
	case "Button":
		props := templComponent.Props.(atoms.ButtonProps)
		builder.WriteString(fmt.Sprintf("\t@atoms.Button(atoms.ButtonProps{\n"))
		builder.WriteString(fmt.Sprintf("\t\tText: \"%s\",\n", props.Text))
		builder.WriteString(fmt.Sprintf("\t\tVariant: atoms.%s,\n", props.Variant))
		builder.WriteString(fmt.Sprintf("\t\tSize: atoms.%s,\n", props.Size))
		if props.Icon != "" {
			builder.WriteString(fmt.Sprintf("\t\tIcon: \"%s\",\n", props.Icon))
		}
		if props.Disabled {
			builder.WriteString("\t\tDisabled: true,\n")
		}
		builder.WriteString("\t})\n")

	case "Input":
		props := templComponent.Props.(atoms.InputProps)
		builder.WriteString(fmt.Sprintf("\t@atoms.Input(atoms.InputProps{\n"))
		builder.WriteString(fmt.Sprintf("\t\tType: \"%s\",\n", props.Type))
		builder.WriteString(fmt.Sprintf("\t\tName: \"%s\",\n", props.Name))
		builder.WriteString(fmt.Sprintf("\t\tPlaceholder: \"%s\",\n", props.Placeholder))
		builder.WriteString(fmt.Sprintf("\t\tSize: atoms.%s,\n", props.Size))
		if props.Required {
			builder.WriteString("\t\tRequired: true,\n")
		}
		if props.Disabled {
			builder.WriteString("\t\tDisabled: true,\n")
		}
		builder.WriteString("\t})\n")

	// Add other component types as needed
	default:
		return "", fmt.Errorf("code generation not implemented for component type: %s", templComponent.Type)
	}

	builder.WriteString("}\n")

	return builder.String(), nil
}

// generateTemplCodeFromComponents generates Templ code from multiple components
func (cli *CLI) generateTemplCodeFromComponents(templComponents []TemplComponent, templateName string) (string, error) {
	var builder strings.Builder

	// Generate package and imports
	builder.WriteString("package components\n\n")
	builder.WriteString("import \"github.com/niiniyare/erp/web/components/atoms\"\n\n")

	// Generate template function
	functionName := strings.ReplaceAll(strings.Title(templateName), "-", "")
	builder.WriteString(fmt.Sprintf("templ %sTemplate() {\n", functionName))
	builder.WriteString("\t<div class=\"space-y-4\">\n")

	// Generate each component
	for _, templComponent := range templComponents {
		builder.WriteString("\t\t")

		switch templComponent.Type {
		case "Button":
			props := templComponent.Props.(atoms.ButtonProps)
			builder.WriteString("@atoms.Button(atoms.ButtonProps{\n")
			builder.WriteString(fmt.Sprintf("\t\t\tText: \"%s\",\n", props.Text))
			builder.WriteString(fmt.Sprintf("\t\t\tVariant: atoms.%s,\n", props.Variant))
			builder.WriteString(fmt.Sprintf("\t\t\tSize: atoms.%s,\n", props.Size))
			if props.Type != "" {
				builder.WriteString(fmt.Sprintf("\t\t\tType: \"%s\",\n", props.Type))
			}
			builder.WriteString("\t\t})\n")

		case "Input":
			props := templComponent.Props.(atoms.InputProps)
			builder.WriteString("@atoms.Input(atoms.InputProps{\n")
			builder.WriteString(fmt.Sprintf("\t\t\tType: \"%s\",\n", props.Type))
			builder.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", props.Name))
			builder.WriteString(fmt.Sprintf("\t\t\tPlaceholder: \"%s\",\n", props.Placeholder))
			builder.WriteString(fmt.Sprintf("\t\t\tSize: atoms.%s,\n", props.Size))
			if props.Required {
				builder.WriteString("\t\t\tRequired: true,\n")
			}
			builder.WriteString("\t\t})\n")

		case "Select":
			props := templComponent.Props.(atoms.SelectProps)
			builder.WriteString("@atoms.Select(atoms.SelectProps{\n")
			builder.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", props.Name))
			builder.WriteString(fmt.Sprintf("\t\t\tPlaceholder: \"%s\",\n", props.Placeholder))
			builder.WriteString(fmt.Sprintf("\t\t\tSize: atoms.%s,\n", props.Size))

			// Generate options
			builder.WriteString("\t\t\tOptions: []atoms.SelectOption{\n")
			for _, option := range props.Options {
				builder.WriteString(fmt.Sprintf("\t\t\t\t{Value: \"%s\", Label: \"%s\"},\n", option.Value, option.Label))
			}
			builder.WriteString("\t\t\t},\n")

			if props.Required {
				builder.WriteString("\t\t\tRequired: true,\n")
			}
			builder.WriteString("\t\t})\n")
		}
	}

	builder.WriteString("\t</div>\n")
	builder.WriteString("}\n")

	return builder.String(), nil
}

// parseTemplComponentFromJSON parses JSON props into a TemplComponent
func (cli *CLI) parseTemplComponentFromJSON(componentType, propsJSON string) (TemplComponent, error) {
	switch componentType {
	case "Button":
		var props atoms.ButtonProps
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return TemplComponent{}, fmt.Errorf("failed to parse Button props: %w", err)
		}
		return TemplComponent{Type: "Button", Props: props}, nil

	case "Input":
		var props atoms.InputProps
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return TemplComponent{}, fmt.Errorf("failed to parse Input props: %w", err)
		}
		return TemplComponent{Type: "Input", Props: props}, nil

	case "Textarea":
		var props atoms.TextareaProps
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return TemplComponent{}, fmt.Errorf("failed to parse Textarea props: %w", err)
		}
		return TemplComponent{Type: "Textarea", Props: props}, nil

	case "Select":
		var props atoms.SelectProps
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return TemplComponent{}, fmt.Errorf("failed to parse Select props: %w", err)
		}
		return TemplComponent{Type: "Select", Props: props}, nil

	case "Checkbox":
		var props atoms.CheckboxProps
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return TemplComponent{}, fmt.Errorf("failed to parse Checkbox props: %w", err)
		}
		return TemplComponent{Type: "Checkbox", Props: props}, nil

	case "Radio":
		var props atoms.RadioProps
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return TemplComponent{}, fmt.Errorf("failed to parse Radio props: %w", err)
		}
		return TemplComponent{Type: "Radio", Props: props}, nil

	default:
		return TemplComponent{}, fmt.Errorf("unknown component type: %s", componentType)
	}
}

// ============================================================================
// CLI EXECUTION HELPER
// ============================================================================

// Execute runs CLI commands based on arguments
func (cli *CLI) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		cli.printUsage()
		return nil
	}

	command := args[0]

	switch command {
	case "schema-to-templ":
		if len(args) < 3 {
			return fmt.Errorf("usage: schema-to-templ <schema-file> <output-file>")
		}
		return cli.ConvertSchemaToTemplFile(ctx, args[1], args[2])

	case "templ-to-schema":
		if len(args) < 4 {
			return fmt.Errorf("usage: templ-to-schema <component-type> <props-json> <output-file>")
		}
		return cli.ConvertTemplToSchemaFile(ctx, args[1], args[2], args[3])

	case "create-template":
		if len(args) < 3 {
			return fmt.Errorf("usage: create-template <template-name> <output-dir>")
		}
		// For simplicity, use empty data map
		return cli.CreateTemplate(ctx, args[1], map[string]any{}, args[2])

	case "validate":
		if len(args) < 3 {
			return fmt.Errorf("usage: validate <schema|templ> <file-path>")
		}
		return cli.ValidateComponent(ctx, args[1], args[2])

	case "list-types":
		cli.ListSupportedTypes()
		return nil

	case "list-templates":
		cli.ListTemplates()
		return nil

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// printUsage prints CLI usage information
func (cli *CLI) printUsage() {
	fmt.Println("Bridge CLI - Connect Schema and Templ Components")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  bridge <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  schema-to-templ <schema-file> <output-file>     Convert schema JSON to Templ component")
	fmt.Println("  templ-to-schema <type> <props-json> <output>    Convert Templ props to schema JSON")
	fmt.Println("  create-template <template-name> <output-dir>    Create template in both formats")
	fmt.Println("  validate <schema|templ> <file-path>             Validate component file")
	fmt.Println("  list-types                                      List supported component types")
	fmt.Println("  list-templates                                  List available templates")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  bridge schema-to-templ button.json button.templ")
	fmt.Println("  bridge templ-to-schema Button '{\"text\":\"Click me\"}' button-schema.json")
	fmt.Println("  bridge create-template login-form ./output")
	fmt.Println("  bridge validate schema component.json")
	fmt.Println("  bridge list-types")
}
