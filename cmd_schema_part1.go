package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// Schema command configuration
var (
	schemaInput      string
	schemaOutput     string
	schemaPackage    string
	schemaSingle     bool
	schemaValidate   bool
	schemaMaxWorkers int
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
- Single file or multi-file output options
- Format-aware type generation (date-time, uuid, email, etc.)
- Validation tag generation from schema constraints`,
	RunE: runSchemaToGo,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.AddCommand(schemaToGoCmd)

	schemaToGoCmd.Flags().StringVarP(&schemaInput, "input", "i", "", "Input file or directory (required)")
	schemaToGoCmd.Flags().StringVarP(&schemaOutput, "output", "o", "./internal/generated", "Output directory")
	schemaToGoCmd.Flags().StringVarP(&schemaPackage, "package", "p", "generated", "Go package name")
	schemaToGoCmd.Flags().BoolVar(&schemaSingle, "single", false, "Generate single combined file")
	schemaToGoCmd.Flags().BoolVar(&schemaValidate, "validate", true, "Validate schemas before generation")
	schemaToGoCmd.Flags().IntVar(&schemaMaxWorkers, "workers", 10, "Max concurrent workers for file loading")

	schemaToGoCmd.MarkFlagRequired("input")
}

func runSchemaToGo(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Converting JSON Schema to Go types\n")
		fmt.Printf("Input: %s\n", schemaInput)
		fmt.Printf("Output: %s\n", schemaOutput)
		fmt.Printf("Package: %s\n", schemaPackage)
		fmt.Printf("Single file: %v\n", schemaSingle)
		fmt.Printf("Validate: %v\n", schemaValidate)
		fmt.Printf("Dry run: %v\n", dryRun)
	}

	// Validate package name
	if !isValidPackageName(schemaPackage) {
		return fmt.Errorf("invalid package name: %s (must be valid Go identifier)", schemaPackage)
	}

	generator := &SchemaToGoGenerator{
		input:      schemaInput,
		output:     schemaOutput,
		pkg:        schemaPackage,
		single:     schemaSingle,
		validate:   schemaValidate,
		dryRun:     dryRun,
		verbose:    verbose,
		maxWorkers: schemaMaxWorkers,
		schemas:    make(map[string]*JSONSchema),
		types:      make(map[string]*GoType),
		processed:  make(map[string]bool),
		refCache:   make(map[string]*JSONSchema),
		nameCache:  make(map[string]string),
		usedNames:  make(map[string]int),
		progress:   &Progress{},
		errors:     []error{},
		warnings:   []string{},
	}

	if err := generator.Run(); err != nil {
		return fmt.Errorf("schema to-go conversion failed: %w", err)
	}

	return nil
}

// Progress tracks generation progress
type Progress struct {
	mu        sync.Mutex
	Total     int
	Processed int
	Failed    int
	Warnings  int
}

func (p *Progress) IncrementProcessed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Processed++
}

func (p *Progress) IncrementFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Failed++
}

func (p *Progress) IncrementWarnings() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Warnings++
}

func (p *Progress) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return fmt.Sprintf("Processed: %d/%d, Failed: %d, Warnings: %d",
		p.Processed, p.Total, p.Failed, p.Warnings)
}

// SchemaToGoGenerator handles the conversion from JSON Schema to Go types
type SchemaToGoGenerator struct {
	input      string
	output     string
	pkg        string
	single     bool
	validate   bool
	dryRun     bool
	verbose    bool
	maxWorkers int
	schemas    map[string]*JSONSchema
	types      map[string]*GoType
	processed  map[string]bool
	refCache   map[string]*JSONSchema
	nameCache  map[string]string
	usedNames  map[string]int
	progress   *Progress
	errors     []error
	warnings   []string
	mu         sync.RWMutex
}

// JSONSchema represents a parsed JSON Schema
type JSONSchema struct {
	ID                   string                 `json:"$id"`
	Schema               string                 `json:"$schema"`
	Type                 any            `json:"type"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Properties           map[string]*JSONSchema `json:"properties"`
	Items                *JSONSchema            `json:"items"`
	AdditionalProperties any            `json:"additionalProperties"`
	Required             []string               `json:"required"`
	Enum                 []any          `json:"enum"`
	Const                any            `json:"const"`
	Ref                  string                 `json:"$ref"`
	Definitions          map[string]*JSONSchema `json:"definitions"`
	AnyOf                []*JSONSchema          `json:"anyOf"`
	OneOf                []*JSONSchema          `json:"oneOf"`
	AllOf                []*JSONSchema          `json:"allOf"`
	Format               string                 `json:"format"`
	Pattern              string                 `json:"pattern"`
	MinLength            *int                   `json:"minLength"`
	MaxLength            *int                   `json:"maxLength"`
	Minimum              *float64               `json:"minimum"`
	Maximum              *float64               `json:"maximum"`
	ExclusiveMinimum     *float64               `json:"exclusiveMinimum"`
	ExclusiveMaximum     *float64               `json:"exclusiveMaximum"`
	MinItems             *int                   `json:"minItems"`
	MaxItems             *int                   `json:"maxItems"`
	UniqueItems          bool                   `json:"uniqueItems"`
	// Custom UI-related properties
	UIComponent string `json:"ui:component"`
	UILabel     string `json:"ui:label"`
	UIRequired  bool   `json:"ui:required"`
	UIHidden    bool   `json:"ui:hidden"`
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
	Dependencies []string
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
	startTime := time.Now()

	// Step 1: Load all JSON Schema files
	if err := g.loadSchemas(); err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	if g.verbose {
		fmt.Printf("✅ Loaded %d schema files in %v\n", len(g.schemas), time.Since(startTime))
	}

	// Step 2: Validate schemas if requested
	if g.validate {
		if err := g.validateSchemas(); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
		if g.verbose {
			fmt.Printf("✅ Validated %d schemas\n", len(g.schemas))
		}
	}

	// Step 3: Generate Go types from schemas
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	if g.verbose {
		fmt.Printf("✅ Generated %d Go types\n", len(g.types))
	}

	// Step 4: Check for name conflicts
	if err := g.checkNameConflicts(); err != nil {
		return fmt.Errorf("name conflict detected: %w", err)
	}

	// Step 5: Write Go files
	if !g.dryRun {
		if err := g.writeGoFiles(); err != nil {
			return fmt.Errorf("failed to write Go files: %w", err)
		}
	} else {
		if err := g.previewGoFiles(); err != nil {
			return fmt.Errorf("failed to preview Go files: %w", err)
		}
	}

	// Report summary
	g.printSummary(time.Since(startTime))

	return nil
}

func (g *SchemaToGoGenerator) loadSchemas() error {
	info, err := os.Stat(g.input)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	var files []string

	if info.IsDir() {
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

	if len(files) == 0 {
		return fmt.Errorf("no JSON files found in %s", g.input)
	}

	g.progress.Total = len(files)

	// Load files concurrently
	return g.loadSchemasConcurrent(files)
}

func (g *SchemaToGoGenerator) loadSchemasConcurrent(files []string) error {
	type result struct {
		id     string
		schema *JSONSchema
		err    error
		file   string
	}

	workers := g.maxWorkers
	if workers > len(files) {
		workers = len(files)
	}

	fileChan := make(chan string, len(files))
	resultChan := make(chan result, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				id, schema, err := g.loadSchemaFileWorker(file)
				resultChan <- result{id: id, schema: schema, err: err, file: file}
			}
		}()
	}

	// Send files to workers
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		g.progress.IncrementProcessed()

		if res.err != nil {
			g.progress.IncrementFailed()
			g.addWarning(fmt.Sprintf("Failed to load %s: %v", res.file, res.err))
			continue
		}

		g.mu.Lock()
		g.schemas[res.id] = res.schema
		g.mu.Unlock()

		if g.verbose {
			fmt.Printf("📄 Loaded: %s\n", res.file)
		}
	}

	if len(g.schemas) == 0 {
		return fmt.Errorf("no valid JSON Schema files found")
	}

	return nil
}

func (g *SchemaToGoGenerator) loadSchemaFileWorker(filePath string) (string, *JSONSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file: %w", err)
	}

	var schema JSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return "", nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	schemaID := schema.ID
	if schemaID == "" {
		schemaID = g.generateSchemaID(filePath)
	}

	return schemaID, &schema, nil
}

func (g *SchemaToGoGenerator) validateSchemas() error {
	for id, schema := range g.schemas {
		if err := g.validateSchema(schema); err != nil {
			g.addWarning(fmt.Sprintf("Schema %s validation warning: %v", id, err))
		}
	}
	return nil
}

func (g *SchemaToGoGenerator) validateSchema(schema *JSONSchema) error {
	// Basic validation checks
	if schema.Type == nil && len(schema.Properties) == 0 && schema.Ref == "" &&
		len(schema.AnyOf) == 0 && len(schema.OneOf) == 0 && len(schema.AllOf) == 0 {
		return fmt.Errorf("schema has no type, properties, or composition keywords")
	}

	// Validate references
	if schema.Ref != "" {
		if _, err := g.resolveRef(schema.Ref); err != nil {
			return fmt.Errorf("invalid reference: %w", err)
		}
	}

	// Validate nested schemas
	for propName, propSchema := range schema.Properties {
		if err := g.validateSchema(propSchema); err != nil {
			return fmt.Errorf("property %s: %w", propName, err)
		}
	}

	return nil
}

func (g *SchemaToGoGenerator) checkNameConflicts() error {
	conflicts := make(map[string][]string)

	for id, goType := range g.types {
		if existing, exists := conflicts[goType.Name]; exists {
			conflicts[goType.Name] = append(existing, id)
		} else {
			conflicts[goType.Name] = []string{id}
		}
	}

	for name, ids := range conflicts {
		if len(ids) > 1 {
			g.addWarning(fmt.Sprintf("Name conflict for type %s from schemas: %v", name, ids))
		}
	}

	return nil
}

func (g *SchemaToGoGenerator) generateTypes() error {
	for schemaID, schema := range g.schemas {
		if g.processed[schemaID] {
			continue
		}

		if g.verbose {
			fmt.Printf("🔄 Processing schema: %s\n", schemaID)
		}

		goType, err := g.processSchema(schemaID, schema)
		if err != nil {
			g.addWarning(fmt.Sprintf("Failed to process schema %s: %v", schemaID, err))
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
	if schema.Ref != "" {
		resolvedSchema, err := g.resolveRef(schema.Ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve reference %s: %w", schema.Ref, err)
		}
		return g.processSchema(name, resolvedSchema)
	}

	if len(schema.AnyOf) > 0 {
		enumValues := g.extractEnumFromAnyOf(schema.AnyOf)
		if len(enumValues) > 0 {
			return g.processEnumSchemaFromValues(name, schema, enumValues)
		}
		return nil, nil
	}

	schemaType := g.getSchemaType(schema)

	switch schemaType {
	case "object":
		return g.processObjectSchema(name, schema)
	case "string":
		if len(schema.Enum) > 0 {
			return g.processEnumSchema(name, schema)
		}
		return nil, nil
	default:
		return nil, nil
	}
}

func (g *SchemaToGoGenerator) processObjectSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
		Comment:    schema.Description,
		SourceFile: name,
		Fields:     []GoField{},
		Imports:    []string{},
	}

	importsSet := make(map[string]bool)

	if schema.Properties != nil {
		propNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propSchema := schema.Properties[propName]

			field, err := g.processPropertyWithContext(propName, propSchema, schema.Required, goType.Name)
			if err != nil {
				g.addWarning(fmt.Sprintf("Failed to process property %s.%s: %v", name, propName, err))
				continue
			}

			goType.Fields = append(goType.Fields, *field)

			for _, imp := range g.getFieldImports(field.Type) {
				importsSet[imp] = true
			}

			if dep := g.extractTypeDependency(field.Type); dep != "" {
				if g.canGenerateType(dep) {
					goType.Dependencies = append(goType.Dependencies, dep)
				}
			}
		}
	}

	for imp := range importsSet {
		goType.Imports = append(goType.Imports, imp)
	}
	sort.Strings(goType.Imports)

	if len(goType.Fields) == 0 && len(goType.EnumValues) == 0 {
		return nil, nil
	}

	return goType, nil
}

func (g *SchemaToGoGenerator) extractEnumFromAnyOf(anyOf []*JSONSchema) []any {
	var enumValues []any

	for _, subSchema := range anyOf {
		if subSchema.Type != nil {
			if typeStr, ok := subSchema.Type.(string); ok && typeStr == "string" {
				if subSchema.Const != nil {
					enumValues = append(enumValues, subSchema.Const)
				}
			}
		}
		if subSchema.Type != nil {
			if typeArray, ok := subSchema.Type.([]any); ok {
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

func (g *SchemaToGoGenerator) processEnumSchemaFromValues(name string, schema *JSONSchema, enumValues []any) (*GoType, error) {
	goType := &GoType{
		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}

	seen := make(map[string]bool)
	for _, value := range enumValues {
		if str, ok := value.(string); ok {
			if !seen[str] {
				goType.EnumValues = append(goType.EnumValues, str)
				seen[str] = true
			}
		}
	}

	if len(goType.EnumValues) == 0 {
		return nil, nil
	}

	return goType, nil
}

func (g *SchemaToGoGenerator) processPropertyWithContext(propName string, propSchema *JSONSchema, required []string, currentTypeName string) (*GoField, error) {
	field := &GoField{
		Name:     g.sanitizeFieldName(propName),
		JSONTag:  propName,
		Comment:  propSchema.Description,
		Optional: !g.isRequired(propName, required),
	}

	if propName == "$$id" || propName == "$ref" || propName == "$schema" {
		field.Type = "any"
	} else {
		goType, err := g.generateGoTypeWithContext(propSchema, currentTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Go type: %w", err)
		}

		if g.isCustomType(goType) && !g.canGenerateType(goType) {
			field.Type = "any"
		} else {
			field.Type = goType
		}
	}

	field.UITag = g.generateUITag(propSchema, propName)
	field.ValidateTag = g.generateValidateTag(propSchema, !field.Optional)

	return field, nil
}

func (g *SchemaToGoGenerator) processEnumSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}

	seen := make(map[string]bool)
	for _, value := range schema.Enum {
		if str, ok := value.(string); ok {
			constantName := g.toPascalCase(str)

			if str == "" {
				constantName = "Empty"
			}

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

func (g *SchemaToGoGenerator) generateGoTypeWithContext(schema *JSONSchema, currentTypeName string) (string, error) {
	if schema.Ref != "" {
		refName := g.extractRefName(schema.Ref)
		refTypeName := g.sanitizeTypeName(refName)

		if refTypeName == currentTypeName {
			return "*" + refTypeName, nil
		}

		if refSchema, err := g.resolveRef(schema.Ref); err == nil {
			if resolvedType, err := g.resolveToFinalType(refSchema); err == nil {
				return resolvedType, nil
			}

			refType := g.getSchemaType(refSchema)
			if refType == "string" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
				return g.getGoTypeForFormat(refSchema.Format, "string"), nil
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

			if !g.wouldGenerateType(refSchema) {
				return "any", nil
			}

			return refTypeName, nil
		} else {
			g.addWarning(fmt.Sprintf("Referenced schema not found: %s, using any", schema.Ref))
			return "any", nil
		}
	}

	if len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
		return "any", nil
	}

	schemaType := g.getSchemaType(schema)

	switch schemaType {
	case "string":
		return g.getGoTypeForFormat(schema.Format, "string"), nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "boolean":
		return "bool", nil
	case "array":
		if schema.Items != nil {
			itemType, err := g.generateGoTypeWithContext(schema.Items, currentTypeName)
			if err != nil {
				return "", fmt.Errorf("failed to generate array item type: %w", err)
			}
			return "[]" + itemType, nil
		}
		return "[]any", nil
	case "object":
		if schema.AdditionalProperties != nil {
			if addProps, ok := schema.AdditionalProperties.(bool); ok && addProps {
				return "map[string]any", nil
			}
			if addPropsSchema, ok := schema.AdditionalProperties.(*JSONSchema); ok {
				valueType, err := g.generateGoTypeWithContext(addPropsSchema, currentTypeName)
				if err != nil {
					return "", fmt.Errorf("failed to generate additionalProperties type: %w", err)
				}
				return "map[string]" + valueType, nil
			}
		}
		if len(schema.Properties) > 0 {
			return "map[string]any", nil
		}
		return "map[string]any", nil
	default:
		return "any", nil
	}
}

func (g *SchemaToGoGenerator) getGoTypeForFormat(format, baseType string) string {
	switch format {
	case "date-time":
		return "time.Time"
	case "date":
		return "time.Time"
	case "time":
		return "time.Time"
	case "uuid":
		return "uuid.UUID"
	case "email", "hostname", "ipv4", "ipv6", "uri", "uri-reference":
		return "string"
	default:
		return baseType
	}
}

func (g *SchemaToGoGenerator) generateUITag(schema *JSONSchema, propName string) string {
	var parts []string

	if schema.UIComponent != "" {
		parts = append(parts, "component="+schema.UIComponent)
	} else {
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

	if schema.UIRequired {
		parts = append(parts, "required=true")
	}

	if schema.UIHidden {
		parts = append(parts, "hidden=true")
	}

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
		if schema.MinLength != nil {
			parts = append(parts, fmt.Sprintf("min=%d", *schema.MinLength))
		}
		if schema.MaxLength != nil {
			parts = append(parts, fmt.Sprintf("max=%d", *schema.MaxLength))
		}
		if schema.Pattern != "" {
			parts = append(parts, fmt.Sprintf("regexp=%s", schema.Pattern))
		}
		switch schema.Format {
		case "email":
			parts = append(parts, "email")
		case "uri", "uri-reference":
			parts = append(parts, "uri")
		case "uuid":
			parts = append(parts, "uuid")
		case "ipv4":
			parts = append(parts, "ipv4")
		case "ipv6":
			parts = append(parts, "ipv6")
		case "hostname":
			parts = append(parts, "hostname")
		}
	case "integer", "number":
		if schema.Minimum != nil {
			parts = append(parts, fmt.Sprintf("min=%v", *schema.Minimum))
		}
		if schema.Maximum != nil {
			parts = append(parts, fmt.Sprintf("max=%v", *schema.Maximum))
		}
		if schema.ExclusiveMinimum != nil {
			parts = append(parts, fmt.Sprintf("gt=%v", *schema.ExclusiveMinimum))
		}
		if schema.ExclusiveMaximum != nil {
			parts = append(parts, fmt.Sprintf("lt=%v", *schema.ExclusiveMaximum))
		}
	case "array":
		if schema.MinItems != nil {
			parts = append(parts, fmt.Sprintf("min=%d", *schema.MinItems))
		}
		if schema.MaxItems != nil {
			parts = append(parts, fmt.Sprintf("max=%d", *schema.MaxItems))
		}
		if schema.UniqueItems {
			parts = append(parts, "unique")
		}
	}

	if len(parts) > 0 {
		return strings.Join(parts, ",")
	}

	return ""
}

func (g *SchemaToGoGenerator) resolveToFinalType(schema *JSONSchema) (string, error) {
	visited := make(map[string]bool)
	return g.resolveToFinalTypeInternal(schema, visited)
}

func (g *SchemaToGoGenerator) resolveToFinalTypeInternal(schema *JSONSchema, visited map[string]bool) (string, error) {
	if schema.Ref != "" {
		if visited[schema.Ref] {
			return "", fmt.Errorf("circular reference detected: %s", schema.Ref)
		}
		visited[schema.Ref] = true

		refSchema, err := g.resolveRef(schema.Ref)
		if err != nil {
			return "", err
		}

		return g.resolveToFinalTypeInternal(refSchema, visited)
	}

	schemaType := g.getSchemaType(schema)
	if len(schema.Properties) == 0 && len(schema.Enum) == 0 {
		switch schemaType {
		case "string":
			return g.getGoTypeForFormat(schema.Format, "string"), nil
		case "integer":
			return "int", nil
		case "number":
			return "float64", nil
		case "boolean":
			return "bool", nil
		case "array":
			if schema.Items != nil {
				itemType, err := g.generateGoTypeWithContext(schema.Items, "")
				if err != nil {
					return "", err
				}
				return "[]" + itemType, nil
			}
			return "[]any", nil
		}
	}

	return "", fmt.Errorf("not a simple type")
}

func (g *SchemaToGoGenerator) resolveRef(ref string) (*JSONSchema, error) {
	if cached, exists := g.refCache[ref]; exists {
		return cached, nil
	}

	if strings.HasPrefix(ref, "#/definitions/") {
		defName := strings.TrimPrefix(ref, "#/definitions/")

		// URL decode the reference name
		decoded, err := url.QueryUnescape(defName)
		if err != nil {
			decoded = defName
		}
		defName = decoded

		for _, schema := range g.schemas {
			if schema.Definitions != nil {
				if def, exists := schema.Definitions[defName]; exists {
					g.refCache[ref] = def
					return def, nil
				}
			}
		}

		for schemaID, schema := range g.schemas {
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
		decoded, err := url.QueryUnescape(refName)
		if err != nil {
			return refName
		}
		return decoded
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
	case []any:
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
	fieldType = strings.TrimPrefix(fieldType, "*")

	basicTypes := map[string]bool{
		"string": true, "int": true, "int32": true, "int64": true,
		"float32": true, "float64": true, "bool": true,
		"any": true, "map[string]any": true,
		"time.Time": true, "uuid.UUID": true,
	}

	if strings.HasPrefix(fieldType, "[]") {
		fieldType = strings.TrimPrefix(fieldType, "[]")
	}
	if strings.HasPrefix(fieldType, "map[") {
		return ""
	}

	if basicTypes[fieldType] {
		return ""
	}

	return fieldType
}

func (g *SchemaToGoGenerator) canGenerateType(typeName string) bool {
	for schemaID, schema := range g.schemas {
		if g.sanitizeTypeName(schemaID) == typeName {
			return g.wouldGenerateType(schema)
		}
	}
	return false
}

func (g *SchemaToGoGenerator) isCustomType(typeName string) bool {
	typeName = strings.TrimPrefix(typeName, "*")

	if strings.HasPrefix(typeName, "[]") {
		typeName = strings.TrimPrefix(typeName, "[]")
	}

	basicTypes := map[string]bool{
		"string": true, "int": true, "int32": true, "int64": true,
		"float32": true, "float64": true, "bool": true,
		"any": true, "map[string]any": true,
		"time.Time": true, "uuid.UUID": true,
	}

	return !basicTypes[typeName] && !strings.HasPrefix(typeName, "map[")
}

func (g *SchemaToGoGenerator) wouldGenerateType(schema *JSONSchema) bool {
	if len(schema.AnyOf) > 0 {
		enumValues := g.extractEnumFromAnyOf(schema.AnyOf)
		if len(enumValues) == 0 {
			return false
		}
	}

	schemaType := g.getSchemaType(schema)
	switch schemaType {
	case "object":
		return len(schema.Properties) > 0
	case "string":
		return len(schema.Enum) > 0
	default:
		return false
	}
}

func (g *SchemaToGoGenerator) topologicalSort(types []*GoType) []*GoType {
	typeMap := make(map[string]*GoType)
	for _, t := range types {
		typeMap[t.Name] = t
	}

	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]*GoType, 0, len(types))

	var visit func(typeName string) bool
	visit = func(typeName string) bool {
		if visiting[typeName] {
			return true
		}
		if visited[typeName] {
			return true
		}

		goType, exists := typeMap[typeName]
		if !exists {
			return true
		}

		visiting[typeName] = true

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

	for _, t := range types {
		visit(t.Name)
	}

	return result
}

func (g *SchemaToGoGenerator) sanitizeTypeName(name string) string {
	// Check cache first
	g.mu.RLock()
	if cached, exists := g.nameCache[name]; exists {
		g.mu.RUnlock()
		return cached
	}
	g.mu.RUnlock()

	original := name
	name = strings.TrimSuffix(name, ".json")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, ".", "")

	name = regexp.MustCompile(`<\(string\|number\)>`).ReplaceAllString(name, "StringOrNumber")
	name = regexp.MustCompile(`<string>`).ReplaceAllString(name, "String")
	name = regexp.MustCompile(`<number>`).ReplaceAllString(name, "Number")
	name = regexp.MustCompile(`<\(.*?\)>`).ReplaceAllString(name, "Union")

	invalidChars := []string{
		"<", ">", "(", ")", " ", "&", "*", "+", "=",
		"[", "]", "{", "}", ":", ";", ",", "?", "!",
		"@", "#", "$", "%", "^", "`", "~",
	}
	for _, char := range invalidChars {
		if char == "|" {
			name = strings.ReplaceAll(name, char, "Or")
		} else {
			name = strings.ReplaceAll(name, char, "")
		}
	}
	name = strings.ReplaceAll(name, "|", "Or")

	result := g.toPascalCase(name)

	// Cache the result
	g.mu.Lock()
	g.nameCache[original] = result
	g.mu.Unlock()

	return result
}

func (g *SchemaToGoGenerator) getUniqueName(name string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if count, exists := g.usedNames[name]; exists {
		g.usedNames[name] = count + 1
		return fmt.Sprintf("%s%d", name, count+1)
	}

	g.usedNames[name] = 0
	return name
}

func (g *SchemaToGoGenerator) sanitizeFieldName(name string) string {
	specialCases := map[string]string{
		"$id":           "SchemaID",
		"$ref":          "Ref",
		"$schema":       "Schema",
		"readOnly":      "ReadOnly",
		"readonly":      "ReadonlyLowerCase",
		"onClick":       "OnClick",
		"onclick":       "OnclickLowerCase",
		"testIdBuilder": "TestIdBuilder",
		"testidBuilder": "TestidBuilderLowerCase",
	}

	if special, exists := specialCases[name]; exists {
		return special
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

	// Handle common Go naming conventions
	replacements := map[string]string{
		"Id":   "ID",
		"Url":  "URL",
		"Api":  "API",
		"Http": "HTTP",
		"Json": "JSON",
		"Xml":  "XML",
		"Sql":  "SQL",
		"Uuid": "UUID",
	}

	for old, new := range replacements {
		pascalCase = strings.ReplaceAll(pascalCase, old, new)
	}

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

func (g *SchemaToGoGenerator) generateSchemaID(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return g.toPascalCase(name)
}

func (g *SchemaToGoGenerator) addWarning(msg string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.warnings = append(g.warnings, msg)
	g.progress.IncrementWarnings()
}

func (g *SchemaToGoGenerator) addError(err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.errors = append(g.errors, err)
}

func (g *SchemaToGoGenerator) printSummary(duration time.Duration) {
	fmt.Printf("\n=== Generation Summary ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Schemas loaded: %d\n", len(g.schemas))
	fmt.Printf("Types generated: %d\n", len(g.types))
	fmt.Printf("Warnings: %d\n", len(g.warnings))
	fmt.Printf("Errors: %d\n", len(g.errors))

	if g.verbose && len(g.warnings) > 0 {
		fmt.Printf("\n=== Warnings ===\n")
		for i, warning := range g.warnings {
			if i >= 10 {
				fmt.Printf("... and %d more warnings\n", len(g.warnings)-10)
				break
			}
			fmt.Printf("  • %s\n", warning)
		}
	}

	if len(g.errors) > 0 {
		fmt.Printf("\n=== Errors ===\n")
		for _, err := range g.errors {
			fmt.Printf("  • %s\n", err)
		}
	}
}

func (g *SchemaToGoGenerator) writeGoFiles() error {
	if err := os.MkdirAll(g.output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if g.single {
		return g.writeSingleFile()
	}
	return g.writeMultipleFiles()
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

	fmt.Printf("✅ Generated: %s (%d bytes)\n", filePath, len(content))
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
			fmt.Printf("✅ Generated: %s (%d bytes)\n", filePath, len(content))
		}
	}

	fmt.Printf("✅ Successfully generated %d Go type files\n", len(g.types))
	return nil
}

func (g *SchemaToGoGenerator) previewGoFiles() error {
	fmt.Println("\n=== DRY RUN: Preview of Generated Files ===\n")

	if g.single {
		content, err := g.generateSingleFileContent()
		if err != nil {
			return err
		}

		fmt.Printf("=== schemas.go (%d bytes) ===\n", len(content))
		fmt.Printf("%s\n", content)
	} else {
		for _, goType := range g.types {
			content, err := g.generateTypeFileContent(goType)
			if err != nil {
				return err
			}

			fileName := g.toSnakeCase(goType.Name) + ".go"
			fmt.Printf("=== %s (%d bytes) ===\n", fileName, len(content))
			fmt.Printf("%s\n\n", content)
		}
	}

	return nil
}

func (g *SchemaToGoGenerator) generateSingleFileContent() ([]byte, error) {
	var allImports []string
	importsSet := make(map[string]bool)

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

	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		g.addWarning(fmt.Sprintf("Failed to format generated code: %v", err))
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
		return nil, fmt.Errorf("failed to execute template for type %s: %w", goType.Name, err)
	}

	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		g.addWarning(fmt.Sprintf("Failed to format generated code for type %s: %v", goType.Name, err))
		return []byte(buf.String()), nil
	}

	return formatted, nil
}

func (g *SchemaToGoGenerator) getSortedTypes() []*GoType {
	var types []*GoType
	for _, goType := range g.types {
		types = append(types, goType)
	}

	types = g.topologicalSort(types)
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
			if s == "" {
				return "Empty"
			}
			return g.toPascalCase(s)
		},
		"enumConstName": func(enumTypeName, value string) string {
			constantName := g.toPascalCase(value)

			if value == "" {
				constantName = "Empty"
			}

			if constantName == enumTypeName {
				constantName = enumTypeName + "Default"
			}

			return enumTypeName + constantName
		},
		"tags": func(field GoField) string {
			var tags []string

			if field.JSONTag != "" {
				jsonTag := field.JSONTag
				if field.Optional {
					jsonTag += ",omitempty"
				}
				tags = append(tags, fmt.Sprintf("json:%q", jsonTag))
			}

			if field.UITag != "" {
				tags = append(tags, fmt.Sprintf("ui:%q", field.UITag))
			}

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

func isValidPackageName(name string) bool {
	if name == "" {
		return false
	}
	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	return validName.MatchString(name)
}

type SchemaToGoGenerator struct {
	input      string
	output     string
	pkg        string
	single     bool
	validate   bool
	dryRun     bool
	verbose    bool
	maxWorkers int
	schemas    map[string]*JSONSchema
	types      map[string]*GoType
	processed  map[string]bool
	refCache   map[string]*JSONSchema
	nameCache  map[string]string
	usedNames  map[string]int
	progress   *Progress
	errors     []error
	warnings   []string
	mu         sync.RWMutex
}

// JSONSchema represents a parsed JSON Schema
type JSONSchema struct {
	ID                   string                 `json:"$id"`
	Schema               string                 `json:"$schema"`
	Type                 any            `json:"type"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Properties           map[string]*JSONSchema `json:"properties"`
	Items                *JSONSchema            `json:"items"`
	AdditionalProperties any            `json:"additionalProperties"`
	Required             []string               `json:"required"`
	Enum                 []any          `json:"enum"`
	Const                any            `json:"const"`
	Ref                  string                 `json:"$ref"`
	Definitions          map[string]*JSONSchema `json:"definitions"`
	AnyOf                []*JSONSchema          `json:"anyOf"`
	OneOf                []*JSONSchema          `json:"oneOf"`
	AllOf                []*JSONSchema          `json:"allOf"`
	Format               string                 `json:"format"`
	Pattern              string                 `json:"pattern"`
	MinLength            *int                   `json:"minLength"`
	MaxLength            *int                   `json:"maxLength"`
	Minimum              *float64               `json:"minimum"`
	Maximum              *float64               `json:"maximum"`
	ExclusiveMinimum     *float64               `json:"exclusiveMinimum"`
	ExclusiveMaximum     *float64               `json:"exclusiveMaximum"`
	MinItems             *int                   `json:"minItems"`
	MaxItems             *int                   `json:"maxItems"`
	UniqueItems          bool                   `json:"uniqueItems"`
	// Custom UI-related properties
	UIComponent string `json:"ui:component"`
	UILabel     string `json:"ui:label"`
	UIRequired  bool   `json:"ui:required"`
	UIHidden    bool   `json:"ui:hidden"`
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
	Dependencies []string
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
	startTime := time.Now()

	// Step 1: Load all JSON Schema files
	if err := g.loadSchemas(); err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	if g.verbose {
		fmt.Printf("✅ Loaded %d schema files in %v\n", len(g.schemas), time.Since(startTime))
	}

	// Step 2: Validate schemas if requested
	if g.validate {
		if err := g.validateSchemas(); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
		if g.verbose {
			fmt.Printf("✅ Validated %d schemas\n", len(g.schemas))
		}
	}

	// Step 3: Generate Go types from schemas
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	if g.verbose {
		fmt.Printf("✅ Generated %d Go types\n", len(g.types))
	}

	// Step 4: Check for name conflicts
	if err := g.checkNameConflicts(); err != nil {
		return fmt.Errorf("name conflict detected: %w", err)
	}

	// Step 5: Write Go files
	if !g.dryRun {
		if err := g.writeGoFiles(); err != nil {
			return fmt.Errorf("failed to write Go files: %w", err)
		}
	} else {
		if err := g.previewGoFiles(); err != nil {
			return fmt.Errorf("failed to preview Go files: %w", err)
		}
	}

	// Report summary
	g.printSummary(time.Since(startTime))

	return nil
}

func (g *SchemaToGoGenerator) loadSchemas() error {
	info, err := os.Stat(g.input)
	if err != nil {
		return fmt.Errorf("input path error: %w", err)
	}

	var files []string

	if info.IsDir() {
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

	if len(files) == 0 {
		return fmt.Errorf("no JSON files found in %s", g.input)
	}

	g.progress.Total = len(files)

	// Load files concurrently
	return g.loadSchemasConcurrent(files)
}

func (g *SchemaToGoGenerator) loadSchemasConcurrent(files []string) error {
	type result struct {
		id     string
		schema *JSONSchema
		err    error
		file   string
	}

	workers := g.maxWorkers
	if workers > len(files) {
		workers = len(files)
	}

	fileChan := make(chan string, len(files))
	resultChan := make(chan result, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				id, schema, err := g.loadSchemaFileWorker(file)
				resultChan <- result{id: id, schema: schema, err: err, file: file}
			}
		}()
	}

	// Send files to workers
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		g.progress.IncrementProcessed()

		if res.err != nil {
			g.progress.IncrementFailed()
			g.addWarning(fmt.Sprintf("Failed to load %s: %v", res.file, res.err))
			continue
		}

		g.mu.Lock()
		g.schemas[res.id] = res.schema
		g.mu.Unlock()

		if g.verbose {
			fmt.Printf("📄 Loaded: %s\n", res.file)
		}
	}

	if len(g.schemas) == 0 {
		return fmt.Errorf("no valid JSON Schema files found")
	}

	return nil
}

func (g *SchemaToGoGenerator) loadSchemaFileWorker(filePath string) (string, *JSONSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file: %w", err)
	}

	var schema JSONSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return "", nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	schemaID := schema.ID
	if schemaID == "" {
		schemaID = g.generateSchemaID(filePath)
	}

	return schemaID, &schema, nil
}

func (g *SchemaToGoGenerator) validateSchemas() error {
	for id, schema := range g.schemas {
		if err := g.validateSchema(schema); err != nil {
			g.addWarning(fmt.Sprintf("Schema %s validation warning: %v", id, err))
		}
	}
	return nil
}

func (g *SchemaToGoGenerator) validateSchema(schema *JSONSchema) error {
	// Basic validation checks
	if schema.Type == nil && len(schema.Properties) == 0 && schema.Ref == "" &&
		len(schema.AnyOf) == 0 && len(schema.OneOf) == 0 && len(schema.AllOf) == 0 {
		return fmt.Errorf("schema has no type, properties, or composition keywords")
	}

	// Validate references
	if schema.Ref != "" {
		if _, err := g.resolveRef(schema.Ref); err != nil {
			return fmt.Errorf("invalid reference: %w", err)
		}
	}

	// Validate nested schemas
	for propName, propSchema := range schema.Properties {
		if err := g.validateSchema(propSchema); err != nil {
			return fmt.Errorf("property %s: %w", propName, err)
		}
	}

	return nil
}

func (g *SchemaToGoGenerator) checkNameConflicts() error {
	conflicts := make(map[string][]string)

	for id, goType := range g.types {
		if existing, exists := conflicts[goType.Name]; exists {
			conflicts[goType.Name] = append(existing, id)
		} else {
			conflicts[goType.Name] = []string{id}
		}
	}

	for name, ids := range conflicts {
		if len(ids) > 1 {
			g.addWarning(fmt.Sprintf("Name conflict for type %s from schemas: %v", name, ids))
		}
	}

	return nil
}

func (g *SchemaToGoGenerator) generateTypes() error {
	for schemaID, schema := range g.schemas {
		if g.processed[schemaID] {
			continue
		}

		if g.verbose {
			fmt.Printf("🔄 Processing schema: %s\n", schemaID)
		}

		goType, err := g.processSchema(schemaID, schema)
		if err != nil {
			g.addWarning(fmt.Sprintf("Failed to process schema %s: %v", schemaID, err))
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
	if schema.Ref != "" {
		resolvedSchema, err := g.resolveRef(schema.Ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve reference %s: %w", schema.Ref, err)
		}
		return g.processSchema(name, resolvedSchema)
	}

	if len(schema.AnyOf) > 0 {
		enumValues := g.extractEnumFromAnyOf(schema.AnyOf)
		if len(enumValues) > 0 {
			return g.processEnumSchemaFromValues(name, schema, enumValues)
		}
		return nil, nil
	}

	schemaType := g.getSchemaType(schema)

	sswitch schemaType {
	case "object":
		return g.processObjectSchema(name, schema)
	case "string":
		if len(schema.Enum) > 0 {
			return g.processEnumSchema(name, schema)
		}
		return nil, nil
	default:
		return nil, nil
	}
}

func (g *SchemaToGoGenerator) processObjectSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
		Comment:    schema.Description,
		SourceFile: name,
		Fields:     []GoField{},
		Imports:    []string{},
	}

	importsSet := make(map[string]bool)

	if schema.Properties != nil {
		propNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propSchema := schema.Properties[propName]

			field, err := g.processPropertyWithContext(propName, propSchema, schema.Required, goType.Name)
			if err != nil {
				g.addWarning(fmt.Sprintf("Failed to process property %s.%s: %v", name, propName, err))
				continue
			}

			goType.Fields = append(goType.Fields, *field)

			for _, imp := range g.getFieldImports(field.Type) {
				importsSet[imp] = true
			}

			if dep := g.extractTypeDependency(field.Type); dep != "" {
				if g.canGenerateType(dep) {
					goType.Dependencies = append(goType.Dependencies, dep)
				}
			}
		}
	}

	for imp := range importsSet {
		goType.Imports = append(goType.Imports, imp)
	}
	sort.Strings(goType.Imports)

	if len(goType.Fields) == 0 && len(goType.EnumValues) == 0 {
		return nil, nil
	}

	return goType, nil
}

func (g *SchemaToGoGenerator) extractEnumFromAnyOf(anyOf []*JSONSchema) []any {
	var enumValues []any

	for _, subSchema := range anyOf {
		if subSchema.Type != nil {
			if typeStr, ok := subSchema.Type.(string); ok && typeStr == "string" {
				if subSchema.Const != nil {
					enumValues = append(enumValues, subSchema.Const)
				}
			}
		}
		if subSchema.Type != nil {
			if typeArray, ok := subSchema.Type.([]any); ok {
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

func (g *SchemaToGoGenerator) processEnumSchemaFromValues(name string, schema *JSONSchema, enumValues []any) (*GoType, error) {
	goType := &GoType{
		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}

	seen := make(map[string]bool)
	for _, value := range enumValues {
		if str, ok := value.(string); ok {
			if !seen[str] {
				goType.EnumValues = append(goType.EnumValues, str)
				seen[str] = true
			}
		}
	}

	if len(goType.EnumValues) == 0 {
		return nil, nil
	}

	return goType, nil
}

func (g *SchemaToGoGenerator) processPropertyWithContext(propName string, propSchema *JSONSchema, required []string, currentTypeName string) (*GoField, error) {
	field := &GoField{
		Name:     g.sanitizeFieldName(propName),
		JSONTag:  propName,
		Comment:  propSchema.Description,
		Optional: !g.isRequired(propName, required),
	}

	if propName == "$$id" || propName == "$ref" || propName == "$schema" {
		field.Type = "any"
	} else {
		goType, err := g.generateGoTypeWithContext(propSchema, currentTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Go type: %w", err)
		}

		if g.isCustomType(goType) && !g.canGenerateType(goType) {
			field.Type = "any"
		} else {
			field.Type = goType
		}
	}

	field.UITag = g.generateUITag(propSchema, propName)
	field.ValidateTag = g.generateValidateTag(propSchema, !field.Optional)

	return field, nil
}

func (g *SchemaToGoGenerator) processEnumSchema(name string, schema *JSONSchema) (*GoType, error) {
	goType := &GoType{
		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
		Comment:    schema.Description,
		SourceFile: name,
		IsEnum:     true,
		EnumValues: []string{},
	}

	seen := make(map[string]bool)
	for _, value := range schema.Enum {
		if str, ok := value.(string); ok {
			constantName := g.toPascalCase(str)

			if str == "" {
				constantName = "Empty"
			}

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

func (g *SchemaToGoGenerator) generateGoTypeWithContext(schema *JSONSchema, currentTypeName string) (string, error) {
	if schema.Ref != "" {
		refName := g.extractRefName(schema.Ref)
		refTypeName := g.sanitizeTypeName(refName)

		if refTypeName == currentTypeName {
			return "*" + refTypeName, nil
		}

		if refSchema, err := g.resolveRef(schema.Ref); err == nil {
			if resolvedType, err := g.resolveToFinalType(refSchema); err == nil {
				return resolvedType, nil
			}

			refType := g.getSchemaType(refSchema)
			if refType == "string" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
				return g.getGoTypeForFormat(refSchema.Format, "string"), nil
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

			if !g.wouldGenerateType(refSchema) {
				return "any", nil
			}

			return refTypeName, nil
		} else {
			g.addWarning(fmt.Sprintf("Referenced schema not found: %s, using any", schema.Ref))
			return "any", nil
		}
	}

	if len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
		return "any", nil
	}

	schemaType := g.getSchemaType(schema)

	sswitch schemaType {
	case "string":
		return g.getGoTypeForFormat(schema.Format, "string"), nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "boolean":
		return "bool", nil
	case "array":
		if schema.Items != nil {
			itemType, err := g.generateGoTypeWithContext(schema.Items, currentTypeName)
			if err != nil {
				return "", fmt.Errorf("failed to generate array item type: %w", err)
			}
			return "[]" + itemType, nil
		}
		return "[]any", nil
	case "object":
		if schema.AdditionalProperties != nil {
			if addProps, ok := schema.AdditionalProperties.(bool); ok && addProps {
				return "map[string]any", nil
			}
			if addPropsSchema, ok := schema.AdditionalProperties.(*JSONSchema); ok {
				valueType, err := g.generateGoTypeWithContext(addPropsSchema, currentTypeName)
				if err != nil {
					return "", fmt.Errorf("failed to generate additionalProperties type: %w", err)
				}
				return "map[string]" + valueType, nil
			}
		}
		if len(schema.Properties) > 0 {
			return "map[string]any", nil
		}
		return "map[string]any", nil
	default:
		return "any", nil
	}
}

func (g *SchemaToGoGenerator) getGoTypeForFormat(format, baseType string) string {
	switch format {
	case "date-time":
		return "time.Time"
	case "date":
		return "time.Time"
	case "time":
		return "time.Time"
	case "uuid":
		return "uuid.UUID"
	case "email", "hostname", "ipv4", "ipv6", "uri", "uri-reference":
		return "string"
	default:
		return baseType
	}
}

func (g *SchemaToGoGenerator) generateUITag(schema *JSONSchema, propName string) string {
	var parts []string

	if schema.UIComponent != "" {
		parts = append(parts, "component="+schema.UIComponent)
	} else {
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

	if schema.UIRequired {
		parts = append(parts, "required=true")
	}

	if schema.UIHidden {
		parts = append(parts, "hidden=true")
	}

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
	sswitch schemaType {
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
		if schema.MinLength != nil {
			parts = append(parts, fmt.Sprintf("min=%d", *schema.MinLength))
		}
		if schema.MaxLength != nil {
			parts = append(parts, fmt.Sprintf("max=%d", *schema.MaxLength))
		}
		if schema.Pattern != "" {
			parts = append(parts, fmt.Sprintf("regexp=%s", schema.Pattern))
		}
		switch schema.Format {
		case "email":
			parts = append(parts, "email")
		case "uri", "uri-reference":
			parts = append(parts, "uri")
		case "uuid":
			parts = append(parts, "uuid")
		case "ipv4":
			parts = append(parts, "ipv4")
		}
	case "number", "integer":
		if schema.Minimum != nil {
			parts = append(parts, fmt.Sprintf("min=%v", *schema.Minimum))
		}
		if schema.Maximum != nil {
			parts = append(parts, fmt.Sprintf("max=%v", *schema.Maximum))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ",")
}

func isValidPackageName(name string) bool {
	if name == "" {
		return false
	}
	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`) // This line was missing in the original potentially_problematic_new_string
	return validName.MatchString(name)
}

const singleFileTemplate = `// Code generated by awoctl schema to-go. DO NOT EDIT.
// Generated: {{.Generated}}

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}