package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/niiniyare/erp/web/engine"
	"github.com/niiniyare/erp/web/schemas"
)

// Schema CLI configuration structure
type SchemaCLIConfig struct {
	// Input options
	File    string
	Package string
	Model   string

	// Output options
	Output string
	Format string
	Layout string

	// Generation options
	APIPrefix string
	PageSize  int
	Patterns  string
	Templates string
	Config    string

	// Control flags
	Permissions bool
	Audit       bool
	Validate    bool
	Force       bool
	Verbose     bool
	Preview     bool
	Watch       bool
}

// Generator state
type SchemaGenerator struct {
	config   SchemaCLIConfig
	registry *engine.SchemaComponentRegistry
	fileSet  *token.FileSet
	cache    map[string]time.Time // File modification cache
}

func schemaMain() {
	config := parseFlags()

	if config.Verbose {
		log.Printf("Starting schema generation with config: %+v", config)
	}

	generator := NewSchemaGenerator(config)

	if config.Watch {
		err := generator.StartWatchMode()
		if err != nil {
			log.Fatalf("Watch mode failed: %v", err)
		}
	} else {
		err := generator.Generate()
		if err != nil {
			log.Fatalf("Generation failed: %v", err)
		}
	}
}

func parseFlags() SchemaCLIConfig {
	var config SchemaCLIConfig

	// Input flags
	flag.StringVar(&config.File, "file", "", "Input Go file to analyze")
	flag.StringVar(&config.Package, "package", "", "Input Go package directory")
	flag.StringVar(&config.Model, "model", "", "Specific struct name to generate")

	// Output flags
	flag.StringVar(&config.Output, "output", "./generated", "Output directory for schemas")
	flag.StringVar(&config.Format, "format", "json", "Output format (json, yaml, go)")
	flag.StringVar(&config.Layout, "layout", "app", "Default layout template")

	// Generation flags
	flag.StringVar(&config.APIPrefix, "api-prefix", "/api", "API endpoint prefix")
	flag.IntVar(&config.PageSize, "page-size", 20, "Default table page size")
	flag.StringVar(&config.Patterns, "patterns", "", "Specific UI patterns to generate")
	flag.StringVar(&config.Templates, "templates", "", "Custom template directory")
	flag.StringVar(&config.Config, "config", "", "Configuration file path")

	// Control flags
	flag.BoolVar(&config.Permissions, "permissions", true, "Enable permission controls")
	flag.BoolVar(&config.Audit, "audit", true, "Enable audit trail features")
	flag.BoolVar(&config.Validate, "validate", true, "Validate generated schemas")
	flag.BoolVar(&config.Force, "force", false, "Force overwrite existing files")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.Preview, "preview", false, "Preview without writing files")
	flag.BoolVar(&config.Watch, "watch", false, "Watch for file changes")

	flag.Parse()

	// Validation
	if config.File == "" && config.Package == "" {
		log.Fatal("Either -file or -package must be specified")
	}

	return config
}

func NewSchemaGenerator(config SchemaCLIConfig) *SchemaGenerator {
	return &SchemaGenerator{
		config:   config,
		registry: engine.NewSchemaComponentRegistry(),
		fileSet:  token.NewFileSet(),
		cache:    make(map[string]time.Time),
	}
}

func (g *SchemaGenerator) Generate() error {
	if g.config.Verbose {
		log.Println("Starting schema generation...")
	}

	var files []string
	var err error

	if g.config.File != "" {
		// Single file mode
		files = []string{g.config.File}
	} else {
		// Package mode
		files, err = g.findGoFiles(g.config.Package)
		if err != nil {
			return fmt.Errorf("failed to find Go files: %w", err)
		}
	}

	if g.config.Verbose {
		log.Printf("Found %d Go files to process", len(files))
	}

	// Process each file
	var allSchemas []schemas.ComponentDefinition
	for _, file := range files {
		schemas, err := g.processFile(file)
		if err != nil {
			if g.config.Verbose {
				log.Printf("Warning: Failed to process file %s: %v", file, err)
			}
			continue
		}
		allSchemas = append(allSchemas, schemas...)
	}

	if len(allSchemas) == 0 {
		log.Println("No schemas generated")
		return nil
	}

	if g.config.Verbose {
		log.Printf("Generated %d component schemas", len(allSchemas))
	}

	// Validate schemas if requested
	if g.config.Validate {
		if err := g.validateSchemas(allSchemas); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
	}

	// Preview mode - just print schemas
	if g.config.Preview {
		return g.printSchemas(allSchemas)
	}

	// Write schemas to files
	return g.writeSchemas(allSchemas)
}

func (g *SchemaGenerator) StartWatchMode() error {
	log.Println("Starting watch mode...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Add paths to watch
	watchPath := g.config.Package
	if watchPath == "" {
		watchPath = filepath.Dir(g.config.File)
	}

	err = filepath.WalkDir(watchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add watch paths: %w", err)
	}

	log.Printf("Watching directory: %s", watchPath)

	// Initial generation
	if err := g.Generate(); err != nil {
		log.Printf("Initial generation failed: %v", err)
	}

	// Watch for changes
	debounceTimer := time.NewTimer(0)
	debounceTimer.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to .go file changes
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			if g.config.Verbose {
				log.Printf("File event: %s %s", event.Op, event.Name)
			}

			// Debounce rapid file changes
			debounceTimer.Reset(500 * time.Millisecond)

		case <-debounceTimer.C:
			log.Println("Regenerating schemas due to file changes...")
			if err := g.Generate(); err != nil {
				log.Printf("Regeneration failed: %v", err)
			} else {
				log.Println("Regeneration completed successfully")
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watch error: %v", err)
		}
	}
}

func (g *SchemaGenerator) findGoFiles(packagePath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(packagePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (g *SchemaGenerator) processFile(filePath string) ([]schemas.ComponentDefinition, error) {
	if g.config.Verbose {
		log.Printf("Processing file: %s", filePath)
	}

	// Check if file has changed since last processing
	if !g.config.Force {
		if lastMod, exists := g.cache[filePath]; exists {
			if stat, err := os.Stat(filePath); err == nil {
				if !stat.ModTime().After(lastMod) {
					if g.config.Verbose {
						log.Printf("Skipping unchanged file: %s", filePath)
					}
					return nil, nil
				}
			}
		}
	}

	// Parse the Go file
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, err := parser.ParseFile(g.fileSet, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var schemas []schemas.ComponentDefinition

	// Walk through AST to find struct declarations
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := node.Type.(*ast.StructType); ok {
				// Skip if specific model is requested and this isn't it
				if g.config.Model != "" && node.Name.Name != g.config.Model {
					return true
				}

				schema, err := g.processStruct(node.Name.Name, structType, filePath)
				if err != nil {
					if g.config.Verbose {
						log.Printf("Warning: Failed to process struct %s: %v", node.Name.Name, err)
					}
					return true
				}

				if schema != nil {
					schemas = append(schemas, *schema)
				}
			}
		}
		return true
	})

	// Update cache
	if stat, err := os.Stat(filePath); err == nil {
		g.cache[filePath] = stat.ModTime()
	}

	return schemas, nil
}

func (g *SchemaGenerator) processStruct(structName string, structType *ast.StructType, filePath string) (*schemas.ComponentDefinition, error) {
	if g.config.Verbose {
		log.Printf("Processing struct: %s", structName)
	}

	// Extract field information
	var fieldSchemas []schemas.ComponentDefinition

	for _, field := range structType.Fields.List {
		// Skip fields without names (embedded types)
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name

		// Skip unexported fields
		if !ast.IsExported(fieldName) {
			continue
		}

		// Parse struct tags
		tags := g.parseFieldTags(field.Tag)

		// Determine component type from tags or field type
		componentType := g.determineComponentType(fieldName, field.Type, tags)
		if componentType == "" {
			continue // Skip fields without UI components
		}

		// Create component schema using registry
		fieldType := reflect.TypeOf("") // Simplified - would need proper type resolution
		_, err := g.registry.ResolveFromTags(componentType, fieldName, fieldType, tags, context.Background())
		if err != nil {
			if g.config.Verbose {
				log.Printf("Warning: Failed to resolve component for field %s: %v", fieldName, err)
			}
			continue
		}

		// Convert component to schema definition
		// This is a placeholder - actual implementation would extract schema from component
		fieldSchema := schemas.ComponentDefinition{
			Type:  componentType,
			ID:    strings.ToLower(fieldName),
			Props: make(map[string]any),
		}

		// Apply tags to schema props
		for tagName, tagValue := range tags {
			fieldSchema.Props[tagName] = tagValue
		}

		fieldSchemas = append(fieldSchemas, fieldSchema)
	}

	// If no UI fields found, skip this struct
	if len(fieldSchemas) == 0 {
		return nil, nil
	}

	// Create container schema based on patterns
	patterns := g.detectPatterns(structName, fieldSchemas)

	// Generate appropriate UI pattern
	var schema *schemas.ComponentDefinition
	if len(patterns) > 0 {
		schema = g.generatePatternSchema(structName, patterns[0], fieldSchemas)
	} else {
		// Default to simple form
		schema = g.generateFormSchema(structName, fieldSchemas)
	}

	return schema, nil
}

func (g *SchemaGenerator) parseFieldTags(tagLit *ast.BasicLit) map[string]string {
	tags := make(map[string]string)

	if tagLit == nil {
		return tags
	}

	// Remove quotes from tag literal
	tagString := strings.Trim(tagLit.Value, "`")

	// Parse struct tag format: `key:"value" key2:"value2"`
	// Simplified parser - a full implementation would be more robust
	parts := strings.Fields(tagString)
	for _, part := range parts {
		if colonIndex := strings.Index(part, ":"); colonIndex > 0 {
			key := part[:colonIndex]
			value := strings.Trim(part[colonIndex+1:], `"`)
			tags[key] = value
		}
	}

	return tags
}

func (g *SchemaGenerator) determineComponentType(fieldName string, fieldType ast.Expr, tags map[string]string) string {
	// Check ui: tag first
	if uiTag, exists := tags["ui"]; exists {
		if strings.Contains(uiTag, "component=") {
			// Extract component type from ui tag
			pairs := strings.Split(uiTag, ";")
			for _, pair := range pairs {
				if strings.HasPrefix(pair, "component=") {
					componentName := strings.TrimPrefix(pair, "component=")
					return "atoms." + componentName
				}
			}
		}
	}

	// Fallback to field type inference
	switch t := fieldType.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "atoms.input"
		case "bool":
			return "atoms.checkbox"
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64":
			return "atoms.input"
		}
	}

	// Check field name patterns
	fieldNameLower := strings.ToLower(fieldName)
	if strings.Contains(fieldNameLower, "button") ||
		strings.Contains(fieldNameLower, "action") ||
		strings.Contains(fieldNameLower, "submit") {
		return "atoms.button"
	}

	return ""
}

func (g *SchemaGenerator) detectPatterns(structName string, fieldSchemas []schemas.ComponentDefinition) []string {
	var patterns []string

	// Basic pattern detection based on struct name and fields
	structNameLower := strings.ToLower(structName)

	// Check for CRUD patterns
	if strings.Contains(structNameLower, "user") ||
		strings.Contains(structNameLower, "product") ||
		strings.Contains(structNameLower, "order") ||
		strings.Contains(structNameLower, "customer") {
		patterns = append(patterns, "crud_table", "create_form", "edit_form", "detail_view")
	}

	// Check for dashboard patterns
	if strings.Contains(structNameLower, "dashboard") ||
		strings.Contains(structNameLower, "metric") ||
		strings.Contains(structNameLower, "analytics") {
		patterns = append(patterns, "dashboard")
	}

	// Check for workflow patterns
	if strings.Contains(structNameLower, "task") ||
		strings.Contains(structNameLower, "workflow") ||
		strings.Contains(structNameLower, "process") {
		patterns = append(patterns, "kanban_board")
	}

	// Default to form if no specific pattern detected
	if len(patterns) == 0 {
		patterns = append(patterns, "create_form")
	}

	return patterns
}

func (g *SchemaGenerator) generatePatternSchema(structName string, pattern string, fieldSchemas []schemas.ComponentDefinition) *schemas.ComponentDefinition {
	switch pattern {
	case "crud_table":
		return g.generateTableSchema(structName, fieldSchemas)
	case "create_form", "edit_form":
		return g.generateFormSchema(structName, fieldSchemas)
	case "dashboard":
		return g.generateDashboardSchema(structName, fieldSchemas)
	case "kanban_board":
		return g.generateKanbanSchema(structName, fieldSchemas)
	default:
		return g.generateFormSchema(structName, fieldSchemas)
	}
}

func (g *SchemaGenerator) generateTableSchema(structName string, fieldSchemas []schemas.ComponentDefinition) *schemas.ComponentDefinition {
	schema := &schemas.ComponentDefinition{
		Type: "organisms.table",
		ID:   strings.ToLower(structName) + "-table",
		Props: map[string]any{
			"title":     structName + " Table",
			"columns":   fieldSchemas,
			"apiPrefix": g.config.APIPrefix,
			"pageSize":  g.config.PageSize,
		},
	}

	return schema
}

func (g *SchemaGenerator) generateFormSchema(structName string, fieldSchemas []schemas.ComponentDefinition) *schemas.ComponentDefinition {
	schema := &schemas.ComponentDefinition{
		Type: "organisms.sectioned-form",
		ID:   strings.ToLower(structName) + "-form",
		Props: map[string]any{
			"title":     "Create " + structName,
			"fields":    fieldSchemas,
			"apiPrefix": g.config.APIPrefix,
		},
	}

	return schema
}

func (g *SchemaGenerator) generateDashboardSchema(structName string, fieldSchemas []schemas.ComponentDefinition) *schemas.ComponentDefinition {
	schema := &schemas.ComponentDefinition{
		Type: "organisms.chart-dashboard",
		ID:   strings.ToLower(structName) + "-dashboard",
		Props: map[string]any{
			"title":   structName + " Dashboard",
			"widgets": fieldSchemas,
		},
	}

	return schema
}

func (g *SchemaGenerator) generateKanbanSchema(structName string, fieldSchemas []schemas.ComponentDefinition) *schemas.ComponentDefinition {
	schema := &schemas.ComponentDefinition{
		Type: "organisms.kanban-board",
		ID:   strings.ToLower(structName) + "-kanban",
		Props: map[string]any{
			"title":  structName + " Board",
			"fields": fieldSchemas,
		},
	}

	return schema
}

func (g *SchemaGenerator) validateSchemas(schemas []schemas.ComponentDefinition) error {
	for i, schema := range schemas {
		if schema.Type == "" {
			return fmt.Errorf("schema %d: missing component type", i)
		}

		if schema.ID == "" {
			return fmt.Errorf("schema %d: missing component ID", i)
		}

		// Validate using registry if available
		if !g.registry.Exists(schema.Type) {
			log.Printf("Warning: Unknown component type: %s", schema.Type)
		}
	}

	return nil
}

func (g *SchemaGenerator) printSchemas(schemas []schemas.ComponentDefinition) error {
	for _, schema := range schemas {
		output, err := g.formatSchema(schema)
		if err != nil {
			return fmt.Errorf("failed to format schema: %w", err)
		}

		fmt.Printf("=== %s ===\n", schema.ID)
		fmt.Println(output)
		fmt.Println()
	}

	return nil
}

func (g *SchemaGenerator) writeSchemas(schemas []schemas.ComponentDefinition) error {
	// Create output directory
	if err := os.MkdirAll(g.config.Output, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, schema := range schemas {
		output, err := g.formatSchema(schema)
		if err != nil {
			return fmt.Errorf("failed to format schema %s: %w", schema.ID, err)
		}

		fileName := g.generateFileName(schema)
		filePath := filepath.Join(g.config.Output, fileName)

		// Check if file exists and force flag
		if !g.config.Force {
			if _, err := os.Stat(filePath); err == nil {
				if g.config.Verbose {
					log.Printf("Skipping existing file: %s (use -force to overwrite)", filePath)
				}
				continue
			}
		}

		if err := os.WriteFile(filePath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		if g.config.Verbose {
			log.Printf("Generated: %s", filePath)
		}
	}

	return nil
}

func (g *SchemaGenerator) formatSchema(schema schemas.ComponentDefinition) (string, error) {
	switch g.config.Format {
	case "json":
		data, err := json.MarshalIndent(schema, "", "  ")
		return string(data), err
	case "yaml":
		return g.registry.GenerateCode("yaml", schema)
	case "go":
		return g.registry.GenerateCode("go", schema)
	default:
		return "", fmt.Errorf("unsupported format: %s", g.config.Format)
	}
}

func (g *SchemaGenerator) generateFileName(schema schemas.ComponentDefinition) string {
	baseName := strings.ReplaceAll(schema.ID, "-", "_")

	switch g.config.Format {
	case "json":
		return baseName + ".json"
	case "yaml":
		return baseName + ".yaml"
	case "go":
		return baseName + ".templ"
	default:
		return baseName + ".json"
	}
}
