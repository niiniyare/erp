package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// Schema command configuration
var (
	schemaInput   string
	schemaOutput  string
	schemaPackage string
	schemaSingle  bool
)

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Schema utilities for UI generation",
	Long: `Schema utilities for working with JSON Schema files and Go type generation.
	
Supports:
- Converting JSON Schema files to Go structs with UI tags
- Generating complete type definitions for schema-driven components
- Integrating with the ERP UI schema architecture`,
}

// schemaToGoCmd converts JSON Schema files to Go types
var schemaToGoCmd = &cobra.Command{
	Use:   "to-go",
	Short: "Convert JSON Schema files to Go types with UI tags",
	Long: `Convert JSON Schema files to Go structs with appropriate UI tags.

This command parses JSON Schema files and generates corresponding Go structs
with proper struct definitions, type mapping, and UI tags for the ERP 
schema-driven architecture.

Examples:
  awoctl schema to-go -i ./docs/ui/Schema/definitions -o ./internal/generated -p ui
  awoctl schema to-go -i ./schema.json -o ./types --single
  awoctl schema to-go -i ./schemas/ -o ./generated --dry-run -v

Features:
- Recursive directory scanning for *.json files
- Intelligent type mapping (string, integer, boolean, object, array)
- Automatic UI tag generation from schema metadata
- Support for nested structs and references
- Enum type generation with constants
- Single file or multi-file output options`,
	RunE: runSchemaToGo,
}

func init() {
	// Add schema command to root
	rootCmd.AddCommand(schemaCmd)
	
	// Add to-go subcommand
	schemaCmd.AddCommand(schemaToGoCmd)
	
	// Schema to-go flags
	schemaToGoCmd.Flags().StringVarP(&schemaInput, "input", "i", "", "Input file or directory (required)")
	schemaToGoCmd.Flags().StringVarP(&schemaOutput, "output", "o", "./internal/generated", "Output directory")
	schemaToGoCmd.Flags().StringVarP(&schemaPackage, "package", "p", "generated", "Go package name")
	schemaToGoCmd.Flags().BoolVar(&schemaSingle, "single", false, "Generate single combined file")
	
	// Mark input as required
	schemaToGoCmd.MarkFlagRequired("input")
}

func runSchemaToGo(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Converting JSON Schema to Go types\n")
		fmt.Printf("Input: %s\n", schemaInput)
		fmt.Printf("Output: %s\n", schemaOutput)
		fmt.Printf("Package: %s\n", schemaPackage)
		fmt.Printf("Single file: %v\n", schemaSingle)
		fmt.Printf("Dry run: %v\n", dryRun)
	}

	// Create schema generator
	generator := &SchemaToGoGenerator{
		input:   schemaInput,
		output:  schemaOutput,
		pkg:     schemaPackage,
		single:  schemaSingle,
		dryRun:  dryRun,
		verbose: verbose,
		schemas: make(map[string]*JSONSchema),
		types:   make(map[string]*GoType),
		processed: make(map[string]bool),
		refCache:  make(map[string]*JSONSchema),
	}

	if err := generator.Run(); err != nil {
		return fmt.Errorf("schema to-go conversion failed: %w", err)
	}

	return nil
}

// SchemaToGoGenerator handles the conversion from JSON Schema to Go types
type SchemaToGoGenerator struct {
	input     string
	output    string
	pkg       string
	single    bool
	dryRun    bool
	verbose   bool
	schemas   map[string]*JSONSchema
	types     map[string]*GoType
	processed map[string]bool
	refCache  map[string]*JSONSchema
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
	Name         string
	Comment      string
	Fields       []GoField
	SourceFile   string
	SourcePath   string
	IsEnum       bool
	EnumValues   []string
	Imports      []string
	Dependencies []string // Types this type depends on
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

func (g *SchemaToGoGenerator) Run() error {
	// Step 1: Load all JSON Schema files
	if err := g.loadSchemas(); err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	if g.verbose {
		fmt.Printf("✅ Loaded %d schema files\n", len(g.schemas))
	}

	// Step 2: Generate Go types from schemas
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	if g.verbose {
		fmt.Printf("✅ Generated %d Go types\n", len(g.types))
	}

	// Step 3: Write Go files
	if !g.dryRun {
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

func (g *SchemaToGoGenerator) loadSchemas() error {
	info, err := os.Stat(g.input)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	var files []string

	if info.IsDir() {
		// Recursively find all .json files
		err = filepath.WalkDir(g.input, func(path string, d fs.DirEntry, err error) error {
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
		files = []string{g.input}
	}

	// Load each schema file
	for _, file := range files {
		if err := g.loadSchemaFile(file); err != nil {
			if g.verbose {
				fmt.Printf("⚠️  Warning: Failed to load schema file %s: %v\n", file, err)
			}
			continue
		}
	}

	if len(g.schemas) == 0 {
		return fmt.Errorf("no valid JSON Schema files found")
	}

	return nil
}

func (g *SchemaToGoGenerator) loadSchemaFile(filePath string) error {
	if g.verbose {
		fmt.Printf("📄 Loading schema file: %s\n", filePath)
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

func (g *SchemaToGoGenerator) generateSchemaID(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Convert from snake_case or kebab-case to PascalCase
	return g.toPascalCase(name)
}

func (g *SchemaToGoGenerator) generateTypes() error {
	// Process each schema
	for schemaID, schema := range g.schemas {
		if g.processed[schemaID] {
			continue
		}

		if g.verbose {
			fmt.Printf("🔄 Processing schema: %s\n", schemaID)
		}

		goType, err := g.processSchema(schemaID, schema)
		if err != nil {
			if g.verbose {
				fmt.Printf("⚠️  Warning: Failed to process schema %s: %v\n", schemaID, err)
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

func (g *SchemaToGoGenerator) processSchema(name string, schema *JSONSchema) (*GoType, error) {
	// Handle schema references
	if schema.Ref != "" {
		resolvedSchema, err := g.resolveRef(schema.Ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve reference %s: %w", schema.Ref, err)
		}
		return g.processSchema(name, resolvedSchema)
	}

	// Check for anyOf patterns
	if len(schema.AnyOf) > 0 {
		enumValues := g.extractEnumFromAnyOf(schema.AnyOf)
		if len(enumValues) > 0 {
			return g.processEnumSchemaFromValues(name, schema, enumValues)
		}
		// For complex anyOf patterns (mixed types), don't generate a type - use interface{}
		return nil, nil
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

func (g *SchemaToGoGenerator) processObjectSchema(name string, schema *JSONSchema) (*GoType, error) {
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

			field, err := g.processPropertyWithContext(propName, propSchema, schema.Required, goType.Name)
			if err != nil {
				if g.verbose {
					fmt.Printf("⚠️  Warning: Failed to process property %s.%s: %v\n", name, propName, err)
				}
				continue
			}

			goType.Fields = append(goType.Fields, *field)

			// Collect imports needed for this field
			for _, imp := range g.getFieldImports(field.Type) {
				importsSet[imp] = true
			}
			
			// Collect dependencies (custom types used by this field)
			if dep := g.extractTypeDependency(field.Type); dep != "" {
				// Only add dependency if we can actually generate this type
				if g.canGenerateType(dep) {
					goType.Dependencies = append(goType.Dependencies, dep)
				}
			}
		}
	}

	// Convert imports set to slice
	for imp := range importsSet {
		goType.Imports = append(goType.Imports, imp)
	}
	sort.Strings(goType.Imports)

	// Skip empty types (no fields and no enum values)
	if len(goType.Fields) == 0 && len(goType.EnumValues) == 0 {
		return nil, nil
	}

	return goType, nil
}

// extractEnumFromAnyOf extracts string constant values from anyOf schemas
func (g *SchemaToGoGenerator) extractEnumFromAnyOf(anyOf []*JSONSchema) []interface{} {
	var enumValues []interface{}
	
	for _, subSchema := range anyOf {
		// Check if this is a string const pattern
		if subSchema.Type != nil {
			if typeStr, ok := subSchema.Type.(string); ok && typeStr == "string" {
				if subSchema.Const != nil {
					enumValues = append(enumValues, subSchema.Const)
				}
			}
		}
		// Also handle array type definitions
		if subSchema.Type != nil {
			if typeArray, ok := subSchema.Type.([]interface{}); ok {
				for _, t := range typeArray {
					if typeStr, ok := t.(string); ok && typeStr == "string" {
						if subSchema.Const != nil {
							enumValues = append(enumValues, subSchema.Const)
						}
					}
				}
			}
		}
	}
	
	return enumValues
}

// processEnumSchemaFromValues creates an enum type from extracted values
func (g *SchemaToGoGenerator) processEnumSchemaFromValues(name string, schema *JSONSchema, enumValues []interface{}) (*GoType, error) {
	goType := &GoType{
		Name:       g.sanitizeTypeName(name),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}

	// Convert enum values to strings and handle duplicates
	seen := make(map[string]bool)
	for _, value := range enumValues {
		if str, ok := value.(string); ok {
			if !seen[str] {
				goType.EnumValues = append(goType.EnumValues, str)
				seen[str] = true
			}
		}
	}

	// Skip if no valid enum values found
	if len(goType.EnumValues) == 0 {
		return nil, nil
	}

	return goType, nil
}


func (g *SchemaToGoGenerator) processProperty(propName string, propSchema *JSONSchema, required []string) (*GoField, error) {
	return g.processPropertyWithContext(propName, propSchema, required, "")
}

func (g *SchemaToGoGenerator) processPropertyWithContext(propName string, propSchema *JSONSchema, required []string, currentTypeName string) (*GoField, error) {
	field := &GoField{
		Name:     g.sanitizeFieldName(propName),
		JSONTag:  propName,
		Comment:  propSchema.Description,
		Optional: !g.isRequired(propName, required),
	}

	// Special handling for schema meta fields - use interface{} for flexibility
	if propName == "$$id" || propName == "$ref" || propName == "$schema" {
		field.Type = "interface{}"
	} else {
		// Generate Go type for this property with context for recursion detection
		goType, err := g.generateGoTypeWithContext(propSchema, currentTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Go type: %w", err)
		}
		
		// If the type is a custom type that can't be generated, use interface{}
		if g.isCustomType(goType) && !g.canGenerateType(goType) {
			field.Type = "interface{}"
		} else {
			field.Type = goType
		}
	}

	// Generate UI tag from schema metadata
	field.UITag = g.generateUITag(propSchema, propName)

	// Generate validation tag
	field.ValidateTag = g.generateValidateTag(propSchema, !field.Optional)

	return field, nil
}

func (g *SchemaToGoGenerator) processEnumSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.sanitizeTypeName(name),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}

	// Convert enum values to strings and handle duplicates
	seen := make(map[string]bool)
	for _, value := range schema.Enum {
		if str, ok := value.(string); ok {
			constantName := g.toPascalCase(str)
			
			// Handle empty string values
			if str == "" {
				constantName = "Empty"
			}
			
			// Prevent conflicts with the type name itself
			typeName := g.sanitizeTypeName(name)
			if constantName == typeName {
				constantName = typeName + "Default"
			}
			
			if !seen[constantName] {
				goType.EnumValues = append(goType.EnumValues, str)
				seen[constantName] = true
			}
		}
	}

	return goType, nil
}

func (g *SchemaToGoGenerator) generateGoType(schema *JSONSchema) (string, error) {
	return g.generateGoTypeWithContext(schema, "")
}

func (g *SchemaToGoGenerator) generateGoTypeWithContext(schema *JSONSchema, currentTypeName string) (string, error) {
	// Handle references
	if schema.Ref != "" {
		refName := g.extractRefName(schema.Ref)
		refTypeName := g.sanitizeTypeName(refName)
		
		// Check for self-reference (recursion) - use pointer to break cycle
		if refTypeName == currentTypeName {
			return "*" + refTypeName, nil
		}
		
		// Check if the referenced schema exists and what type it is
		if refSchema, err := g.resolveRef(schema.Ref); err == nil {
			// Handle transitive references (references to references)
			resolvedType, err := g.resolveToFinalType(refSchema)
			if err == nil {
				return resolvedType, nil
			}
			
			// If the referenced schema is a simple type (string, etc.) without properties,
			// use the underlying type instead of creating a custom type
			refType := g.getSchemaType(refSchema)
			if refType == "string" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
				return "string", nil
			}
			if refType == "integer" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
				return "int", nil
			}
			if refType == "number" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
				return "float64", nil
			}
			if refType == "boolean" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
				return "bool", nil
			}
			
			// Check if this schema would be processed to create a type
			// If not (e.g., anyOf patterns), use interface{}
			if !g.wouldGenerateType(refSchema) {
				return "interface{}", nil
			}
			
			return refTypeName, nil
		} else {
			// Referenced schema not found - use interface{} instead of failing
			if g.verbose {
				fmt.Printf("⚠️  Warning: Referenced schema not found: %s, using interface{}\n", schema.Ref)
			}
			return "interface{}", nil
		}
		
	}

	// Handle anyOf/oneOf by using interface{}
	if len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
		return "interface{}", nil
	}

	schemaType := g.getSchemaType(schema)

	switch schemaType {
	case "string":
		if len(schema.Enum) > 0 {
			return "string", nil
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
			return "map[string]interface{}", nil // Fallback
		}

		return "map[string]interface{}", nil

	default:
		return "interface{}", nil
	}
}

// Helper methods (implementation continues with all the utility functions)

func (g *SchemaToGoGenerator) generateUITag(schema *JSONSchema, propName string) string {
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

func (g *SchemaToGoGenerator) generateValidateTag(schema *JSONSchema, required bool) string {
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
	}

	if len(parts) > 0 {
		return strings.Join(parts, ",")
	}

	return ""
}

// Additional helper methods for the generator (resolveRef, extractRefName, getSchemaType, etc.)
// [Implementation continues with all utility functions from the original main.go]

func (g *SchemaToGoGenerator) resolveToFinalType(schema *JSONSchema) (string, error) {
	// Track visited schemas to prevent infinite loops
	visited := make(map[string]bool)
	return g.resolveToFinalTypeInternal(schema, visited)
}

func (g *SchemaToGoGenerator) resolveToFinalTypeInternal(schema *JSONSchema, visited map[string]bool) (string, error) {
	// Check for reference
	if schema.Ref != "" {
		// Prevent infinite loops
		if visited[schema.Ref] {
			return "", fmt.Errorf("circular reference detected: %s", schema.Ref)
		}
		visited[schema.Ref] = true
		
		// Resolve the reference
		refSchema, err := g.resolveRef(schema.Ref)
		if err != nil {
			return "", err
		}
		
		// Continue resolving
		return g.resolveToFinalTypeInternal(refSchema, visited)
	}
	
	// Check if it's a simple type
	schemaType := g.getSchemaType(schema)
	if len(schema.Properties) == 0 && len(schema.Enum) == 0 {
		switch schemaType {
		case "string":
			return "string", nil
		case "integer":
			return "int", nil
		case "number":
			return "float64", nil
		case "boolean":
			return "bool", nil
		case "array":
			// Handle array types
			if schema.Items != nil {
				itemType, err := g.generateGoType(schema.Items)
				if err != nil {
					return "", err
				}
				return "[]" + itemType, nil
			}
			return "[]interface{}", nil
		}
	}
	
	// Not a simple type
	return "", fmt.Errorf("not a simple type")
}

func (g *SchemaToGoGenerator) resolveRef(ref string) (*JSONSchema, error) {
	if cached, exists := g.refCache[ref]; exists {
		return cached, nil
	}

	if strings.HasPrefix(ref, "#/definitions/") {
		defName := strings.TrimPrefix(ref, "#/definitions/")
		
		// URL decode the reference name
		defName = strings.ReplaceAll(defName, "%3C", "<")
		defName = strings.ReplaceAll(defName, "%3E", ">")
		defName = strings.ReplaceAll(defName, "%7C", "|")
		defName = strings.ReplaceAll(defName, "%28", "(")
		defName = strings.ReplaceAll(defName, "%29", ")")
		defName = strings.ReplaceAll(defName, "%20", " ")
		
		// First, look in local definitions within schemas
		for _, schema := range g.schemas {
			if schema.Definitions != nil {
				if def, exists := schema.Definitions[defName]; exists {
					g.refCache[ref] = def
					return def, nil
				}
			}
		}
		
		// If not found in definitions, look for a standalone schema with that name
		for schemaID, schema := range g.schemas {
			// Check if this schema ID matches the definition name (case-insensitive)
			if strings.EqualFold(strings.TrimSuffix(schemaID, ".json"), defName) ||
				strings.EqualFold(schemaID, "./"+defName+".json") ||
				strings.EqualFold(schemaID, defName) {
				g.refCache[ref] = schema
				return schema, nil
			}
		}
	}

	return nil, fmt.Errorf("reference not found: %s", ref)
}

func (g *SchemaToGoGenerator) extractRefName(ref string) string {
	if strings.HasPrefix(ref, "#/definitions/") {
		refName := strings.TrimPrefix(ref, "#/definitions/")
		// URL decode the reference name (e.g., %3C becomes <, %3E becomes >)
		refName = strings.ReplaceAll(refName, "%3C", "<")
		refName = strings.ReplaceAll(refName, "%3E", ">")
		refName = strings.ReplaceAll(refName, "%7C", "|")
		refName = strings.ReplaceAll(refName, "%28", "(")
		refName = strings.ReplaceAll(refName, "%29", ")")
		refName = strings.ReplaceAll(refName, "%20", " ")
		return refName
	}
	if strings.HasPrefix(ref, "./") && strings.HasSuffix(ref, ".json") {
		base := filepath.Base(ref)
		return strings.TrimSuffix(base, ".json")
	}
	return ref
}

func (g *SchemaToGoGenerator) getSchemaType(schema *JSONSchema) string {
	if schema.Type == nil {
		return "object"
	}

	switch t := schema.Type.(type) {
	case string:
		return t
	case []interface{}:
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

func (g *SchemaToGoGenerator) isRequired(propName string, required []string) bool {
	for _, req := range required {
		if req == propName {
			return true
		}
	}
	return false
}

func (g *SchemaToGoGenerator) getFieldImports(fieldType string) []string {
	var imports []string

	if strings.Contains(fieldType, "time.Time") {
		imports = append(imports, "time")
	}
	if strings.Contains(fieldType, "uuid.UUID") {
		imports = append(imports, "github.com/google/uuid")
	}

	return imports
}

func (g *SchemaToGoGenerator) extractTypeDependency(fieldType string) string {
	// Remove pointer prefix if present
	fieldType = strings.TrimPrefix(fieldType, "*")
	
	// Skip basic Go types
	basicTypes := map[string]bool{
		"string": true, "int": true, "int32": true, "int64": true,
		"float32": true, "float64": true, "bool": true,
		"interface{}": true, "map[string]interface{}": true,
		"time.Time": true, "uuid.UUID": true,
	}
	
	// Skip array/slice types by extracting the element type
	if strings.HasPrefix(fieldType, "[]") {
		fieldType = strings.TrimPrefix(fieldType, "[]")
	}
	if strings.HasPrefix(fieldType, "map[") {
		return "" // Maps with custom value types are complex, skip for now
	}
	
	if basicTypes[fieldType] {
		return ""
	}
	
	// This is likely a custom type
	return fieldType
}

// canGenerateType checks if we can generate a type with the given name
func (g *SchemaToGoGenerator) canGenerateType(typeName string) bool {
	// Check if we have a schema for this type and if it would actually generate
	for schemaID, schema := range g.schemas {
		if g.sanitizeTypeName(schemaID) == typeName {
			return g.wouldGenerateType(schema)
		}
	}
	return false
}

// isCustomType checks if a type name represents a custom type (not a built-in Go type)
func (g *SchemaToGoGenerator) isCustomType(typeName string) bool {
	// Remove pointer prefix if present
	typeName = strings.TrimPrefix(typeName, "*")
	
	// Skip array/slice types by extracting the element type
	if strings.HasPrefix(typeName, "[]") {
		typeName = strings.TrimPrefix(typeName, "[]")
	}
	
	// Basic Go types
	basicTypes := map[string]bool{
		"string": true, "int": true, "int32": true, "int64": true,
		"float32": true, "float64": true, "bool": true,
		"interface{}": true, "map[string]interface{}": true,
		"time.Time": true, "uuid.UUID": true,
	}
	
	return !basicTypes[typeName] && !strings.HasPrefix(typeName, "map[")
}

// wouldGenerateType checks if a schema would result in a generated type
func (g *SchemaToGoGenerator) wouldGenerateType(schema *JSONSchema) bool {
	// Check for anyOf patterns that aren't enum-like
	if len(schema.AnyOf) > 0 {
		enumValues := g.extractEnumFromAnyOf(schema.AnyOf)
		if len(enumValues) == 0 {
			return false // Complex anyOf patterns don't generate types
		}
	}
	
	// Check schema type
	schemaType := g.getSchemaType(schema)
	switch schemaType {
	case "object":
		// Only generate object types if they have properties
		return len(schema.Properties) > 0
	case "string":
		return len(schema.Enum) > 0 // Only enums generate types
	default:
		return false
	}
}

// topologicalSort sorts types by dependencies so dependencies come first
func (g *SchemaToGoGenerator) topologicalSort(types []*GoType) []*GoType {
	// Create a map for quick lookup
	typeMap := make(map[string]*GoType)
	for _, t := range types {
		typeMap[t.Name] = t
	}
	
	// Track visited and current path for cycle detection
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]*GoType, 0, len(types))
	
	var visit func(typeName string) bool
	visit = func(typeName string) bool {
		if visiting[typeName] {
			// Cycle detected, skip this dependency
			return true
		}
		if visited[typeName] {
			return true
		}
		
		goType, exists := typeMap[typeName]
		if !exists {
			// External dependency or doesn't exist in our set
			return true
		}
		
		visiting[typeName] = true
		
		// Visit all dependencies first
		for _, dep := range goType.Dependencies {
			if !visit(dep) {
				visiting[typeName] = false
				return false
			}
		}
		
		visiting[typeName] = false
		visited[typeName] = true
		result = append(result, goType)
		return true
	}
	
	// Visit all types
	for _, t := range types {
		visit(t.Name)
	}
	
	return result
}

func (g *SchemaToGoGenerator) sanitizeTypeName(name string) string {
	name = strings.TrimSuffix(name, ".json")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, ".", "")
	
	// Convert common type patterns to readable names
	name = regexp.MustCompile(`<\(string\|number\)>`).ReplaceAllString(name, "StringOrNumber")
	name = regexp.MustCompile(`<string>`).ReplaceAllString(name, "String")
	name = regexp.MustCompile(`<number>`).ReplaceAllString(name, "Number")
	name = regexp.MustCompile(`<\(.*?\)>`).ReplaceAllString(name, "Union")
	
	// Remove remaining invalid characters for Go type names
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, "|", "Or")
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "&", "And")
	name = strings.ReplaceAll(name, "*", "")
	name = strings.ReplaceAll(name, "+", "")
	name = strings.ReplaceAll(name, "=", "")
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")
	name = strings.ReplaceAll(name, ":", "")
	name = strings.ReplaceAll(name, ";", "")
	name = strings.ReplaceAll(name, ",", "")
	name = strings.ReplaceAll(name, "?", "")
	name = strings.ReplaceAll(name, "!", "")
	name = strings.ReplaceAll(name, "@", "")
	name = strings.ReplaceAll(name, "#", "")
	name = strings.ReplaceAll(name, "$", "")
	name = strings.ReplaceAll(name, "%", "")
	name = strings.ReplaceAll(name, "^", "")
	name = strings.ReplaceAll(name, "`", "")
	name = strings.ReplaceAll(name, "~", "")
	
	return g.toPascalCase(name)
}

func (g *SchemaToGoGenerator) sanitizeFieldName(name string) string {
	if name == "$$id" {
		return "SchemaID"
	}
	if name == "$ref" {
		return "Ref"
	}
	if name == "$schema" {
		return "Schema"
	}
	
	// Handle case variations that would create conflicts
	switch name {
	case "readOnly":
		return "ReadOnly"
	case "readonly":
		return "ReadonlyLowerCase"
	case "onClick":
		return "OnClick"
	case "onclick":
		return "OnclickLowerCase"
	case "testIdBuilder":
		return "TestIdBuilder"
	case "testidBuilder":
		return "TestidBuilderLowerCase"
	}
	
	return g.toPascalCase(name)
}

func (g *SchemaToGoGenerator) toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	words := regexp.MustCompile(`[-_\s]+`).Split(s, -1)

	var result strings.Builder
	for _, word := range words {
		if word == "" {
			continue
		}

		if len(word) == 1 {
			result.WriteString(strings.ToUpper(word))
		} else {
			result.WriteString(strings.ToUpper(string(word[0])))
			result.WriteString(strings.ToLower(word[1:]))
		}
	}

	pascalCase := result.String()

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

func (g *SchemaToGoGenerator) humanizeFieldName(fieldName string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	humanized := re.ReplaceAllString(fieldName, `$1 $2`)

	if len(humanized) > 0 {
		humanized = strings.ToUpper(string(humanized[0])) + humanized[1:]
	}

	return humanized
}

func (g *SchemaToGoGenerator) writeGoFiles() error {
	if err := os.MkdirAll(g.output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if g.single {
		return g.writeSingleFile()
	} else {
		return g.writeMultipleFiles()
	}
}

func (g *SchemaToGoGenerator) writeSingleFile() error {
	filePath := filepath.Join(g.output, "schemas.go")

	content, err := g.generateSingleFileContent()
	if err != nil {
		return fmt.Errorf("failed to generate file content: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Generated: %s\n", filePath)
	return nil
}

func (g *SchemaToGoGenerator) writeMultipleFiles() error {
	for _, goType := range g.types {
		content, err := g.generateTypeFileContent(goType)
		if err != nil {
			return fmt.Errorf("failed to generate content for type %s: %w", goType.Name, err)
		}

		fileName := g.toSnakeCase(goType.Name) + ".go"
		filePath := filepath.Join(g.output, fileName)

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		if g.verbose {
			fmt.Printf("✅ Generated: %s\n", filePath)
		}
	}

	fmt.Printf("✅ Successfully generated %d Go type files\n", len(g.types))
	return nil
}

func (g *SchemaToGoGenerator) previewGoFiles() error {
	if g.single {
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

func (g *SchemaToGoGenerator) generateSingleFileContent() ([]byte, error) {
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
		Package   string
		Imports   []string
		Types     []*GoType
		Generated string
	}{
		Package:   g.pkg,
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
		if g.verbose {
			fmt.Printf("⚠️  Warning: Failed to format generated code: %v\n", err)
		}
		return []byte(buf.String()), nil
	}

	return formatted, nil
}

func (g *SchemaToGoGenerator) generateTypeFileContent(goType *GoType) ([]byte, error) {
	tmpl := template.Must(template.New("type_file").Funcs(g.getTemplateFuncs()).Parse(typeFileTemplate))

	data := struct {
		Package   string
		Imports   []string
		Type      *GoType
		Generated string
	}{
		Package:   g.pkg,
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
		if g.verbose {
			fmt.Printf("⚠️  Warning: Failed to format generated code for type %s: %v\n", goType.Name, err)
		}
		return []byte(buf.String()), nil
	}

	return formatted, nil
}

func (g *SchemaToGoGenerator) getSortedTypes() []*GoType {
	var types []*GoType
	for _, goType := range g.types {
		types = append(types, goType)
	}

	// First, sort by dependencies using topological sort
	types = g.topologicalSort(types)
	
	// Then alphabetically within dependency groups for consistent output
	// (Topological sort preserves relative ordering where no dependencies exist)
	return types
}

func (g *SchemaToGoGenerator) toSnakeCase(s string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	snake := re.ReplaceAllString(s, `${1}_${2}`)
	return strings.ToLower(snake)
}

func (g *SchemaToGoGenerator) getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"join": strings.Join,
		"quote": func(s string) string {
			return strconv.Quote(s)
		},
		"hasPrefix": strings.HasPrefix,
		"trimSpace": strings.TrimSpace,
		"title": func(s string) string {
			// Handle enum constant names properly
			if s == "" {
				return "Empty"
			}
			return g.toPascalCase(s)
		},
		"enumConstName": func(enumTypeName, value string) string {
			constantName := g.toPascalCase(value)
			
			// Handle empty string values
			if value == "" {
				constantName = "Empty"
			}
			
			// Prevent conflicts with the type name itself
			if constantName == enumTypeName {
				constantName = enumTypeName + "Default"
			}
			
			return enumTypeName + constantName
		},
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
		"comment": func(text string) string {
			if text == "" {
				return ""
			}
			// Replace newlines and normalize whitespace for inline comments
			cleaned := strings.ReplaceAll(text, "\n", " ")
			cleaned = strings.ReplaceAll(cleaned, "\r", " ")
			cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
			cleaned = strings.TrimSpace(cleaned)
			return cleaned
		},
		"multiComment": func(text string) string {
			if text == "" {
				return ""
			}
			// Format as proper multi-line comment block
			lines := strings.Split(text, "\n")
			var result strings.Builder
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if i > 0 {
					result.WriteString("\n// ")
				} else {
					result.WriteString("// ")
				}
				result.WriteString(line)
			}
			return result.String()
		},
	}
}

// Template definitions
const singleFileTemplate = `// Code generated by awoctl schema to-go. DO NOT EDIT.
// Generated: {{.Generated}}

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}	"{{.}}"
{{end}})
{{end}}

{{range .Types}}
{{if .Comment}}{{.Comment | multiComment}}{{end}}
{{if .IsEnum -}}
type {{.Name}} string

const (
{{range $i, $value := .EnumValues}}	{{enumConstName $.Name $value}} {{$.Name}} = {{$value | quote}}
{{end}})
{{else -}}
type {{.Name}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} {{. | tags}}{{if .Comment}} // {{.Comment | comment}}{{end}}
{{end}}}
{{end}}

{{end}}`

const typeFileTemplate = `// Code generated by awoctl schema to-go. DO NOT EDIT.
// Source: {{.Type.SourceFile}}
// Generated: {{.Generated}}

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}	"{{.}}"
{{end}})
{{end}}

{{if .Type.Comment}}// {{.Type.Comment | comment}}{{end}}
{{if .Type.IsEnum -}}
type {{.Type.Name}} string

const (
{{range $i, $value := .Type.EnumValues}}	{{enumConstName $.Type.Name $value}} {{$.Type.Name}} = {{$value | quote}}
{{end}})
{{else -}}
type {{.Type.Name}} struct {
{{range .Type.Fields}}{{if .Comment}}	// {{.Comment | comment}}
{{end}}	{{.Name}} {{.Type}} {{. | tags}}
{{end}}}
{{end}}`