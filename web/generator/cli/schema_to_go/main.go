// Code generated CLI tool for converting JSON Schema to Go types.
// This tool parses JSON Schema files and generates corresponding Go structs
// with appropriate UI tags for the ERP schema-driven architecture.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// CLI configuration
type Config struct {
	Input   string // Input file or directory
	Output  string // Output directory
	Package string // Go package name
	Single  bool   // Generate single combined file
	Verbose bool   // Verbose logging
	DryRun  bool   // Preview only, don't write files
}

// JSONSchema represents a parsed JSON Schema
type JSONSchema struct {
	ID                   string                     `json:"$id"`
	Schema               string                     `json:"$schema"`
	Type                 interface{}                `json:"type"` // Can be string or []string
	Title                string                     `json:"title"`
	Description          string                     `json:"description"`
	Properties           map[string]*JSONSchema     `json:"properties"`
	Items                *JSONSchema                `json:"items"`
	AdditionalProperties interface{}                `json:"additionalProperties"` // Can be bool or schema
	Required             []string                   `json:"required"`
	Enum                 []interface{}              `json:"enum"`
	Const                interface{}                `json:"const"`
	Ref                  string                     `json:"$ref"`
	Definitions          map[string]*JSONSchema     `json:"definitions"`
	AnyOf                []*JSONSchema              `json:"anyOf"`
	OneOf                []*JSONSchema              `json:"oneOf"`
	AllOf                []*JSONSchema              `json:"allOf"`
	// Custom UI-related properties
	UIComponent          string                     `json:"ui:component"`
	UILabel              string                     `json:"ui:label"`
	UIRequired           bool                       `json:"ui:required"`
	UIHidden             bool                       `json:"ui:hidden"`
}

// GoType represents a generated Go type
type GoType struct {
	Name        string
	Comment     string
	Fields      []GoField
	SourceFile  string
	SourcePath  string
	IsEnum      bool
	EnumValues  []string
	Imports     []string
}

// GoField represents a Go struct field
type GoField struct {
	Name        string
	Type        string
	JSONTag     string
	UITag       string
	ValidateTag string
	Comment     string
	Optional    bool
}

// Generator handles the conversion from JSON Schema to Go types
type Generator struct {
	config      Config
	schemas     map[string]*JSONSchema
	types       map[string]*GoType
	processed   map[string]bool // Track processed schemas to avoid infinite recursion
	refCache    map[string]*JSONSchema // Cache for resolved references
}

func main() {
	config := parseFlags()
	
	if config.Verbose {
		log.Printf("Starting schema-to-go generator with config: %+v", config)
	}
	
	generator := NewGenerator(config)
	
	if err := generator.Run(); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}
	
	if config.Verbose {
		log.Println("Schema-to-go generation completed successfully")
	}
}

func parseFlags() Config {
	var config Config
	
	flag.StringVar(&config.Input, "input", "", "Input file or directory (required)")
	flag.StringVar(&config.Input, "i", "", "Input file or directory (short)")
	flag.StringVar(&config.Output, "output", "./internal/generated", "Output directory")
	flag.StringVar(&config.Output, "o", "./internal/generated", "Output directory (short)")
	flag.StringVar(&config.Package, "package", "generated", "Go package name")
	flag.StringVar(&config.Package, "p", "generated", "Go package name (short)")
	flag.BoolVar(&config.Single, "single", false, "Generate single combined file")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose logging (short)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Preview only, don't write files")
	
	flag.Parse()
	
	if config.Input == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -input <file_or_directory> [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	return config
}

func NewGenerator(config Config) *Generator {
	return &Generator{
		config:    config,
		schemas:   make(map[string]*JSONSchema),
		types:     make(map[string]*GoType),
		processed: make(map[string]bool),
		refCache:  make(map[string]*JSONSchema),
	}
}

func (g *Generator) Run() error {
	// Step 1: Load all JSON Schema files
	if err := g.loadSchemas(); err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}
	
	if g.config.Verbose {
		log.Printf("Loaded %d schema files", len(g.schemas))
	}
	
	// Step 2: Generate Go types from schemas
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}
	
	if g.config.Verbose {
		log.Printf("Generated %d Go types", len(g.types))
	}
	
	// Step 3: Write Go files
	if !g.config.DryRun {
		if err := g.writeGoFiles(); err != nil {
			return fmt.Errorf("failed to write Go files: %w", err)
		}
	} else {
		if err := g.previewGoFiles(); err != nil {
			return fmt.Errorf("failed to preview Go files: %w", err)
		}
	}
	
	return nil
}

func (g *Generator) loadSchemas() error {
	info, err := os.Stat(g.config.Input)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}
	
	var files []string
	
	if info.IsDir() {
		// Recursively find all .json files
		err = filepath.WalkDir(g.config.Input, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			
			if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
				files = append(files, path)
			}
			
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		files = []string{g.config.Input}
	}
	
	// Load each schema file
	for _, file := range files {
		if err := g.loadSchemaFile(file); err != nil {
			if g.config.Verbose {
				log.Printf("Warning: Failed to load schema file %s: %v", file, err)
			}
			continue
		}
	}
	
	if len(g.schemas) == 0 {
		return fmt.Errorf("no valid JSON Schema files found")
	}
	
	return nil
}

func (g *Generator) loadSchemaFile(filePath string) error {
	if g.config.Verbose {
		log.Printf("Loading schema file: %s", filePath)
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	var schema JSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Generate schema ID if not present
	schemaID := schema.ID
	if schemaID == "" {
		schemaID = g.generateSchemaID(filePath)
	}
	
	// Store schema with metadata
	g.schemas[schemaID] = &schema
	
	return nil
}

func (g *Generator) generateSchemaID(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	
	// Convert from snake_case or kebab-case to PascalCase
	return g.toPascalCase(name)
}

func (g *Generator) generateTypes() error {
	// Process each schema
	for schemaID, schema := range g.schemas {
		if g.processed[schemaID] {
			continue
		}
		
		if g.config.Verbose {
			log.Printf("Processing schema: %s", schemaID)
		}
		
		goType, err := g.processSchema(schemaID, schema)
		if err != nil {
			if g.config.Verbose {
				log.Printf("Warning: Failed to process schema %s: %v", schemaID, err)
			}
			continue
		}
		
		if goType != nil {
			g.types[schemaID] = goType
		}
		
		g.processed[schemaID] = true
	}
	
	return nil
}

func (g *Generator) processSchema(name string, schema *JSONSchema) (*GoType, error) {
	// Handle schema references
	if schema.Ref != "" {
		resolvedSchema, err := g.resolveRef(schema.Ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve reference %s: %w", schema.Ref, err)
		}
		return g.processSchema(name, resolvedSchema)
	}
	
	// Determine schema type
	schemaType := g.getSchemaType(schema)
	
	switch schemaType {
	case "object":
		return g.processObjectSchema(name, schema)
	case "array":
		// Arrays don't generate standalone types, they're handled as field types
		return nil, nil
	case "string":
		if len(schema.Enum) > 0 {
			return g.processEnumSchema(name, schema)
		}
		return nil, nil
	default:
		// Primitive types don't generate standalone types
		return nil, nil
	}
}

func (g *Generator) processObjectSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.sanitizeTypeName(name),
		Comment:    schema.Description,
		SourceFile: name,
		Fields:     []GoField{},
		Imports:    []string{},
	}
	
	// Add standard imports that might be needed
	importsSet := make(map[string]bool)
	
	// Process properties
	if schema.Properties != nil {
		// Sort properties for consistent output
		var propNames []string
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)
		
		for _, propName := range propNames {
			propSchema := schema.Properties[propName]
			
			field, err := g.processProperty(propName, propSchema, schema.Required)
			if err != nil {
				if g.config.Verbose {
					log.Printf("Warning: Failed to process property %s.%s: %v", name, propName, err)
				}
				continue
			}
			
			goType.Fields = append(goType.Fields, *field)
			
			// Collect imports needed for this field
			for _, imp := range g.getFieldImports(field.Type) {
				importsSet[imp] = true
			}
		}
	}
	
	// Convert imports set to slice
	for imp := range importsSet {
		goType.Imports = append(goType.Imports, imp)
	}
	sort.Strings(goType.Imports)
	
	return goType, nil
}

func (g *Generator) processProperty(propName string, propSchema *JSONSchema, required []string) (*GoField, error) {
	field := &GoField{
		Name:     g.sanitizeFieldName(propName),
		JSONTag:  propName,
		Comment:  propSchema.Description,
		Optional: !g.isRequired(propName, required),
	}
	
	// Generate Go type for this property
	goType, err := g.generateGoType(propSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Go type: %w", err)
	}
	
	field.Type = goType
	
	// Generate UI tag from schema metadata
	field.UITag = g.generateUITag(propSchema, propName)
	
	// Generate validation tag
	field.ValidateTag = g.generateValidateTag(propSchema, !field.Optional)
	
	return field, nil
}

func (g *Generator) processEnumSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.sanitizeTypeName(name),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}
	
	// Convert enum values to strings
	for _, value := range schema.Enum {
		if str, ok := value.(string); ok {
			goType.EnumValues = append(goType.EnumValues, str)
		}
	}
	
	return goType, nil
}

func (g *Generator) generateGoType(schema *JSONSchema) (string, error) {
	// Handle references
	if schema.Ref != "" {
		refName := g.extractRefName(schema.Ref)
		return g.sanitizeTypeName(refName), nil
	}
	
	// Handle anyOf/oneOf by using interface{}
	if len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
		return "interface{}", nil
	}
	
	schemaType := g.getSchemaType(schema)
	
	switch schemaType {
	case "string":
		if len(schema.Enum) > 0 {
			// This is an enum type - we'll need to generate a custom type
			return "string", nil // For now, use string, but could be improved
		}
		return "string", nil
		
	case "integer":
		return "int", nil
		
	case "number":
		return "float64", nil
		
	case "boolean":
		return "bool", nil
		
	case "array":
		if schema.Items != nil {
			itemType, err := g.generateGoType(schema.Items)
			if err != nil {
				return "", fmt.Errorf("failed to generate array item type: %w", err)
			}
			return "[]" + itemType, nil
		}
		return "[]interface{}", nil
		
	case "object":
		// Handle additionalProperties
		if schema.AdditionalProperties != nil {
			if addProps, ok := schema.AdditionalProperties.(bool); ok && addProps {
				return "map[string]interface{}", nil
			}
			if addPropsSchema, ok := schema.AdditionalProperties.(*JSONSchema); ok {
				valueType, err := g.generateGoType(addPropsSchema)
				if err != nil {
					return "", fmt.Errorf("failed to generate additionalProperties type: %w", err)
				}
				return "map[string]" + valueType, nil
			}
		}
		
		// If it has properties, it should be a separate struct
		if len(schema.Properties) > 0 {
			// This will be handled as a separate type
			return "map[string]interface{}", nil // Fallback
		}
		
		return "map[string]interface{}", nil
		
	default:
		return "interface{}", nil
	}
}

func (g *Generator) generateUITag(schema *JSONSchema, propName string) string {
	var parts []string
	
	// Component type
	if schema.UIComponent != "" {
		parts = append(parts, "component="+schema.UIComponent)
	} else {
		// Infer component type from schema type
		schemaType := g.getSchemaType(schema)
		switch schemaType {
		case "string":
			if len(schema.Enum) > 0 {
				parts = append(parts, "component=select")
			} else {
				parts = append(parts, "component=text")
			}
		case "boolean":
			parts = append(parts, "component=toggle")
		case "integer", "number":
			parts = append(parts, "component=number")
		case "array":
			parts = append(parts, "component=multi-select")
		}
	}
	
	// Label
	label := schema.UILabel
	if label == "" && schema.Title != "" {
		label = schema.Title
	}
	if label == "" {
		label = g.humanizeFieldName(propName)
	}
	if label != "" {
		parts = append(parts, "label="+label)
	}
	
	// Required
	if schema.UIRequired {
		parts = append(parts, "required=true")
	}
	
	// Hidden
	if schema.UIHidden {
		parts = append(parts, "hidden=true")
	}
	
	// Options for enum fields
	if len(schema.Enum) > 0 {
		var options []string
		for _, val := range schema.Enum {
			if str, ok := val.(string); ok {
				options = append(options, str)
			}
		}
		if len(options) > 0 {
			parts = append(parts, "options="+strings.Join(options, ","))
		}
	}
	
	if len(parts) > 0 {
		return strings.Join(parts, ";")
	}
	
	return ""
}

func (g *Generator) generateValidateTag(schema *JSONSchema, required bool) string {
	var parts []string
	
	if required {
		parts = append(parts, "required")
	}
	
	schemaType := g.getSchemaType(schema)
	switch schemaType {
	case "string":
		if len(schema.Enum) > 0 {
			var options []string
			for _, val := range schema.Enum {
				if str, ok := val.(string); ok {
					options = append(options, str)
				}
			}
			if len(options) > 0 {
				parts = append(parts, "oneof="+strings.Join(options, " "))
			}
		}
	case "integer", "number":
		// Could add min/max validation based on schema properties
	}
	
	if len(parts) > 0 {
		return strings.Join(parts, ",")
	}
	
	return ""
}

// Helper functions

func (g *Generator) resolveRef(ref string) (*JSONSchema, error) {
	// Cache resolved references
	if cached, exists := g.refCache[ref]; exists {
		return cached, nil
	}
	
	// Simple reference resolution - assumes internal references only
	if strings.HasPrefix(ref, "#/definitions/") {
		defName := strings.TrimPrefix(ref, "#/definitions/")
		
		// Look for definition in any loaded schema
		for _, schema := range g.schemas {
			if schema.Definitions != nil {
				if def, exists := schema.Definitions[defName]; exists {
					g.refCache[ref] = def
					return def, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("reference not found: %s", ref)
}

func (g *Generator) extractRefName(ref string) string {
	if strings.HasPrefix(ref, "#/definitions/") {
		return strings.TrimPrefix(ref, "#/definitions/")
	}
	if strings.HasPrefix(ref, "./") && strings.HasSuffix(ref, ".json") {
		base := filepath.Base(ref)
		return strings.TrimSuffix(base, ".json")
	}
	return ref
}

func (g *Generator) getSchemaType(schema *JSONSchema) string {
	if schema.Type == nil {
		return "object" // Default to object if no type specified
	}
	
	switch t := schema.Type.(type) {
	case string:
		return t
	case []interface{}:
		// Multiple types - return the first one or prefer non-null
		for _, typ := range t {
			if str, ok := typ.(string); ok && str != "null" {
				return str
			}
		}
		if len(t) > 0 {
			if str, ok := t[0].(string); ok {
				return str
			}
		}
	}
	
	return "object"
}

func (g *Generator) isRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}

func (g *Generator) getFieldImports(fieldType string) []string {
	var imports []string
	
	// Check for common imports needed
	if strings.Contains(fieldType, "time.Time") {
		imports = append(imports, "time")
	}
	if strings.Contains(fieldType, "uuid.UUID") {
		imports = append(imports, "github.com/google/uuid")
	}
	
	return imports
}

func (g *Generator) sanitizeTypeName(name string) string {
	// Remove file extensions and path separators
	name = strings.TrimSuffix(name, ".json")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, ".", "")
	
	// Convert to PascalCase
	return g.toPascalCase(name)
}

func (g *Generator) sanitizeFieldName(name string) string {
	// Handle special JSON schema fields
	if name == "$$id" {
		return "SchemaID"
	}
	if name == "$ref" {
		return "Ref"
	}
	if name == "$schema" {
		return "Schema"
	}
	
	// Convert to PascalCase for exported fields
	return g.toPascalCase(name)
}

func (g *Generator) toPascalCase(s string) string {
	// Handle empty strings
	if s == "" {
		return ""
	}
	
	// Split on common separators
	words := regexp.MustCompile(`[-_\s]+`).Split(s, -1)
	
	var result strings.Builder
	for _, word := range words {
		if word == "" {
			continue
		}
		
		// Capitalize first letter, lowercase the rest
		if len(word) == 1 {
			result.WriteString(strings.ToUpper(word))
		} else {
			result.WriteString(strings.ToUpper(string(word[0])))
			result.WriteString(strings.ToLower(word[1:]))
		}
	}
	
	pascalCase := result.String()
	
	// Ensure it starts with uppercase letter
	if len(pascalCase) > 0 && pascalCase[0] >= 'a' && pascalCase[0] <= 'z' {
		pascalCase = strings.ToUpper(string(pascalCase[0])) + pascalCase[1:]
	}
	
	// Handle edge cases for Go naming
	pascalCase = strings.ReplaceAll(pascalCase, "Id", "ID")
	pascalCase = strings.ReplaceAll(pascalCase, "Url", "URL")
	pascalCase = strings.ReplaceAll(pascalCase, "Api", "API")
	pascalCase = strings.ReplaceAll(pascalCase, "Http", "HTTP")
	pascalCase = strings.ReplaceAll(pascalCase, "Json", "JSON")
	pascalCase = strings.ReplaceAll(pascalCase, "Xml", "XML")
	
	return pascalCase
}

func (g *Generator) humanizeFieldName(fieldName string) string {
	// Convert PascalCase to human readable
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	humanized := re.ReplaceAllString(fieldName, `$1 $2`)
	
	// Capitalize first letter
	if len(humanized) > 0 {
		humanized = strings.ToUpper(string(humanized[0])) + humanized[1:]
	}
	
	return humanized
}

func (g *Generator) writeGoFiles() error {
	// Create output directory
	if err := os.MkdirAll(g.config.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	if g.config.Single {
		return g.writeSingleFile()
	} else {
		return g.writeMultipleFiles()
	}
}

func (g *Generator) writeSingleFile() error {
	filePath := filepath.Join(g.config.Output, "schemas.go")
	
	content, err := g.generateSingleFileContent()
	if err != nil {
		return fmt.Errorf("failed to generate file content: %w", err)
	}
	
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	if g.config.Verbose {
		log.Printf("Generated: %s", filePath)
	}
	
	return nil
}

func (g *Generator) writeMultipleFiles() error {
	for _, goType := range g.types {
		content, err := g.generateTypeFileContent(goType)
		if err != nil {
			return fmt.Errorf("failed to generate content for type %s: %w", goType.Name, err)
		}
		
		fileName := g.toSnakeCase(goType.Name) + ".go"
		filePath := filepath.Join(g.config.Output, fileName)
		
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		
		if g.config.Verbose {
			log.Printf("Generated: %s", filePath)
		}
	}
	
	return nil
}

func (g *Generator) previewGoFiles() error {
	if g.config.Single {
		content, err := g.generateSingleFileContent()
		if err != nil {
			return err
		}
		
		fmt.Printf("=== schemas.go ===\n")
		fmt.Printf("%s\n", content)
	} else {
		for _, goType := range g.types {
			content, err := g.generateTypeFileContent(goType)
			if err != nil {
				return err
			}
			
			fileName := g.toSnakeCase(goType.Name) + ".go"
			fmt.Printf("=== %s ===\n", fileName)
			fmt.Printf("%s\n", content)
		}
	}
	
	return nil
}

func (g *Generator) generateSingleFileContent() ([]byte, error) {
	var allImports []string
	importsSet := make(map[string]bool)
	
	// Collect all imports
	for _, goType := range g.types {
		for _, imp := range goType.Imports {
			if !importsSet[imp] {
				allImports = append(allImports, imp)
				importsSet[imp] = true
			}
		}
	}
	sort.Strings(allImports)
	
	tmpl := template.Must(template.New("single_file").Funcs(g.getTemplateFuncs()).Parse(singleFileTemplate))
	
	data := struct {
		Package string
		Imports []string
		Types   []*GoType
		Generated string
	}{
		Package:   g.config.Package,
		Imports:   allImports,
		Types:     g.getSortedTypes(),
		Generated: time.Now().Format(time.RFC3339),
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	
	// Format the generated Go code
	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		// If formatting fails, return unformatted code with a warning
		if g.config.Verbose {
			log.Printf("Warning: Failed to format generated code: %v", err)
		}
		return []byte(buf.String()), nil
	}
	
	return formatted, nil
}

func (g *Generator) generateTypeFileContent(goType *GoType) ([]byte, error) {
	tmpl := template.Must(template.New("type_file").Funcs(g.getTemplateFuncs()).Parse(typeFileTemplate))
	
	data := struct {
		Package   string
		Imports   []string
		Type      *GoType
		Generated string
	}{
		Package:   g.config.Package,
		Imports:   goType.Imports,
		Type:      goType,
		Generated: time.Now().Format(time.RFC3339),
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	
	// Format the generated Go code
	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		// If formatting fails, return unformatted code with a warning
		if g.config.Verbose {
			log.Printf("Warning: Failed to format generated code for type %s: %v", goType.Name, err)
		}
		return []byte(buf.String()), nil
	}
	
	return formatted, nil
}

func (g *Generator) getSortedTypes() []*GoType {
	var types []*GoType
	for _, goType := range g.types {
		types = append(types, goType)
	}
	
	// Sort by name for consistent output
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})
	
	return types
}

func (g *Generator) toSnakeCase(s string) string {
	// Convert PascalCase to snake_case
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	snake := re.ReplaceAllString(s, `${1}_${2}`)
	return strings.ToLower(snake)
}

func (g *Generator) getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"join": strings.Join,
		"quote": func(s string) string {
			return strconv.Quote(s)
		},
		"hasPrefix": strings.HasPrefix,
		"trimSpace": strings.TrimSpace,
		"title": strings.Title,
		"tags": func(field GoField) string {
			var tags []string
			
			// JSON tag
			if field.JSONTag != "" {
				jsonTag := field.JSONTag
				if field.Optional {
					jsonTag += ",omitempty"
				}
				tags = append(tags, fmt.Sprintf("json:%q", jsonTag))
			}
			
			// UI tag
			if field.UITag != "" {
				tags = append(tags, fmt.Sprintf("ui:%q", field.UITag))
			}
			
			// Validate tag
			if field.ValidateTag != "" {
				tags = append(tags, fmt.Sprintf("validate:%q", field.ValidateTag))
			}
			
			if len(tags) > 0 {
				return "`" + strings.Join(tags, " ") + "`"
			}
			return ""
		},
	}
}

// Template definitions
const singleFileTemplate = `// Code generated by schema-to-go. DO NOT EDIT.
// Generated: {{.Generated}}

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}	"{{.}}"
{{end}})
{{end}}

{{range .Types}}
{{if .Comment}}// {{.Comment}}{{end}}
{{if .IsEnum -}}
type {{.Name}} string

const (
{{range $i, $value := .EnumValues}}	{{$.Name}}{{$value | title}} {{$.Name}} = {{$value | quote}}
{{end}})
{{else -}}
type {{.Name}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} {{. | tags}}{{if .Comment}} // {{.Comment}}{{end}}
{{end}}}
{{end}}

{{end}}`

const typeFileTemplate = `// Code generated by schema-to-go. DO NOT EDIT.
// Source: {{.Type.SourceFile}}
// Generated: {{.Generated}}

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}	"{{.}}"
{{end}})
{{end}}

{{if .Type.Comment}}// {{.Type.Comment}}{{end}}
{{if .Type.IsEnum -}}
type {{.Type.Name}} string

const (
{{range $i, $value := .Type.EnumValues}}	{{$.Type.Name}}{{$value | title}} {{$.Type.Name}} = {{$value | quote}}
{{end}})
{{else -}}
type {{.Type.Name}} struct {
{{range .Type.Fields}}	{{.Name}} {{.Type}} {{. | tags}}{{if .Comment}} // {{.Comment}}{{end}}
{{end}}}
{{end}}`