package main

//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"go/format"
// 	"io/fs"
// 	"net/url"
// 	"os"
// 	"path/filepath"
// 	"regexp"
// 	"sort"
// 	"strings"
// 	"sync"
// 	"text/template"
// 	"time"
//
// 	"github.com/spf13/cobra"
// )
//
// // Schema command configuration
// var (
// 	schemaInput      string
// 	schemaOutput     string
// 	schemaPackage    string
// 	schemaSingle     bool
// 	schemaValidate   bool
// 	schemaMaxWorkers int
// )
//
// // schemaCmd represents the schema command
// var schemaCmd = &cobra.Command{
// 	Use:   "schema",
// 	Short: "Schema utilities for UI generation",
// 	Long: `Schema utilities for working with JSON Schema files and Go type generation.
//
// Supports:
// - Converting JSON Schema files to Go structs with UI tags
// - Generating complete type definitions for schema-driven components
// - Integrating with the ERP UI schema architecture`,
// }
//
// // schemaToGoCmd converts JSON Schema files to Go types
// var schemaToGoCmd = &cobra.Command{
// 	Use:   "to-go",
// 	Short: "Convert JSON Schema files to Go types with UI tags",
// 	Long: `Convert JSON Schema files to Go structs with appropriate UI tags.
//
// This command parses JSON Schema files and generates corresponding Go structs
// with proper struct definitions, type mapping, and UI tags for the ERP
// schema-driven architecture.
//
// Examples:
//   awoctl schema to-go -i ./docs/ui/Schema/definitions -o ./internal/generated -p ui
//   awoctl schema to-go -i ./schema.json -o ./types --single
//   awoctl schema to-go -i ./schemas/ -o ./generated --dry-run -v
//
// Features:
// - Recursive directory scanning for *.json files
// - Intelligent type mapping (string, integer, boolean, object, array)
// - Automatic UI tag generation from schema metadata
// - Support for nested structs and references
// - Enum type generation with constants
// - Single file or multi-file output options
// - Format-aware type generation (date-time, uuid, email, etc.)
// - Validation tag generation from schema constraints`,
// 	RunE: runSchemaToGo,
// }
//
// func init() {
// 	rootCmd.AddCommand(schemaCmd)
// 	schemaCmd.AddCommand(schemaToGoCmd)
//
// 	schemaToGoCmd.Flags().StringVarP(&schemaInput, "input", "i", "", "Input file or directory (required)")
// 	schemaToGoCmd.Flags().StringVarP(&schemaOutput, "output", "o", "./internal/generated", "Output directory")
// 	schemaToGoCmd.Flags().StringVarP(&schemaPackage, "package", "p", "generated", "Go package name")
// 	schemaToGoCmd.Flags().BoolVar(&schemaSingle, "single", false, "Generate single combined file")
// 	schemaToGoCmd.Flags().BoolVar(&schemaValidate, "validate", true, "Validate schemas before generation")
// 	schemaToGoCmd.Flags().IntVar(&schemaMaxWorkers, "workers", 10, "Max concurrent workers for file loading")
//
// 	schemaToGoCmd.MarkFlagRequired("input")
// }
//
// func runSchemaToGo(cmd *cobra.Command, args []string) error {
// 	if verbose {
// 		fmt.Printf("Converting JSON Schema to Go types\n")
// 		fmt.Printf("Input: %s\n", schemaInput)
// 		fmt.Printf("Output: %s\n", schemaOutput)
// 		fmt.Printf("Package: %s\n", schemaPackage)
// 		fmt.Printf("Single file: %v\n", schemaSingle)
// 		fmt.Printf("Validate: %v\n", schemaValidate)
// 		fmt.Printf("Dry run: %v\n", dryRun)
// 	}
//
// 	// Validate package name
// 	if !isValidPackageName(schemaPackage) {
// 		return fmt.Errorf("invalid package name: %s (must be valid Go identifier)", schemaPackage)
// 	}
//
// 	generator := &SchemaToGoGenerator{
// 		input:      schemaInput,
// 		output:     schemaOutput,
// 		pkg:        schemaPackage,
// 		single:     schemaSingle,
// 		validate:   schemaValidate,
// 		dryRun:     dryRun,
// 		verbose:    verbose,
// 		maxWorkers: schemaMaxWorkers,
// 		schemas:    make(map[string]*JSONSchema),
// 		types:      make(map[string]*GoType),
// 		processed:  make(map[string]bool),
// 		refCache:   make(map[string]*JSONSchema),
// 		nameCache:  make(map[string]string),
// 		usedNames:  make(map[string]int),
// 		progress:   &Progress{},
// 		errors:     []error{},
// 		warnings:   []string{},
// 	}
//
// 	if err := generator.Run(); err != nil {
// 		return fmt.Errorf("schema to-go conversion failed: %w", err)
// 	}
//
// 	return nil
// }
//
// // Progress tracks generation progress
// type Progress struct {
// 	mu        sync.Mutex
// 	Total     int
// 	Processed int
// 	Failed    int
// 	Warnings  int
// }
//
// func (p *Progress) IncrementProcessed() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	p.Processed++
// }
//
// func (p *Progress) IncrementFailed() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	p.Failed++
// }
//
// func (p *Progress) IncrementWarnings() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	p.Warnings++
// }
//
// func (p *Progress) String() string {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	return fmt.Sprintf("Processed: %d/%d, Failed: %d, Warnings: %d",
// 		p.Processed, p.Total, p.Failed, p.Warnings)
// }
//
// // SchemaToGoGenerator handles the conversion from JSON Schema to Go types
// type SchemaToGoGenerator struct {
// 	input      string
// 	output     string
// 	pkg        string
// 	single     bool
// 	validate   bool
// 	dryRun     bool
// 	verbose    bool
// 	maxWorkers int
// 	schemas    map[string]*JSONSchema
// 	types      map[string]*GoType
// 	processed  map[string]bool
// 	refCache   map[string]*JSONSchema
// 	nameCache  map[string]string
// 	usedNames  map[string]int
// 	progress   *Progress
// 	errors     []error
// 	warnings   []string
// 	mu         sync.RWMutex
// }
//
// // JSONSchema represents a parsed JSON Schema
// type JSONSchema struct {
// 	ID                   string                 `json:"id"`
// 	Schema               string                 `json:"schema"`
// 	Type                 any                    `json:"type"`
// 	Title                string                 `json:"title"`
// 	Description          string                 `json:"description"`
// 	Properties           map[string]*JSONSchema `json:"properties"`
// 	Items                *JSONSchema            `json:"items"`
// 	AdditionalProperties any                    `json:"additionalProperties"`
// 	Required             []string               `json:"required"`
// 	Enum                 []any                  `json:"enum"`
// 	Const                any                    `json:"const"`
// 	Ref                  string                 `json:"ref"`
// 	Definitions          map[string]*JSONSchema `json:"definitions"`
// 	AnyOf                []*JSONSchema          `json:"anyOf"`
// 	OneOf                []*JSONSchema          `json:"oneOf"`
// 	AllOf                []*JSONSchema          `json:"allOf"`
// 	Format               string                 `json:"format"`
// 	Pattern              string                 `json:"pattern"`
// 	MinLength            *int                   `json:"minLength"`
// 	MaxLength            *int                   `json:"maxLength"`
// 	Minimum              *float64               `json:"minimum"`
// 	Maximum              *float64               `json:"maximum"`
// 	ExclusiveMinimum     *float64               `json:"exclusiveMinimum"`
// 	ExclusiveMaximum     *float64               `json:"exclusiveMaximum"`
// 	MinItems             *int                   `json:"minItems"`
// 	MaxItems             *int                   `json:"maxItems"`
// 	UniqueItems          bool                   `json:"uniqueItems"`
// 	// Custom UI-related properties
// 	UIComponent string `json:"ui:component"`
// 	UILabel     string `json:"ui:label"`
// 	UIRequired  bool   `json:"ui:required"`
// 	UIHidden    bool   `json:"ui:hidden"`
// }
//
// // GoType represents a generated Go type
// type GoType struct {
// 	Name         string
// 	Comment      string
// 	Fields       []GoField
// 	SourceFile   string
// 	SourcePath   string
// 	IsEnum       bool
// 	EnumValues   []string
// 	Imports      []string
// 	Dependencies []string
// }
//
// // GoField represents a Go struct field
// type GoField struct {
// 	Name        string
// 	Type        string
// 	JSONTag     string
// 	UITag       string
// 	ValidateTag string
// 	Comment     string
// 	Optional    bool
// }
//
// func (g *SchemaToGoGenerator) Run() error {
// 	startTime := time.Now()
//
// 	// Step 1: Load all JSON Schema files
// 	if err := g.loadSchemas(); err != nil {
// 		return fmt.Errorf("failed to load schemas: %w", err)
// 	}
//
// 	if g.verbose {
// 		fmt.Printf("✅ Loaded %d schema files in %v\n", len(g.schemas), time.Since(startTime))
// 	}
//
// 	// Step 2: Validate schemas if requested
// 	if g.validate {
// 		if err := g.validateSchemas(); err != nil {
// 			return fmt.Errorf("schema validation failed: %w", err)
// 		}
// 		if g.verbose {
// 			fmt.Printf("✅ Validated %d schemas\n", len(g.schemas))
// 		}
// 	}
//
// 	// Step 3: Generate Go types from schemas
// 	if err := g.generateTypes(); err != nil {
// 		return fmt.Errorf("failed to generate types: %w", err)
// 	}
//
// 	if g.verbose {
// 		fmt.Printf("✅ Generated %d Go types\n", len(g.types))
// 	}
//
// 	// Step 4: Check for name conflicts
// 	if err := g.checkNameConflicts(); err != nil {
// 		return fmt.Errorf("name conflict detected: %w", err)
// 	}
//
// 	// Step 5: Write Go files
// 	if !g.dryRun {
// 		if err := g.writeGoFiles(); err != nil {
// 			return fmt.Errorf("failed to write Go files: %w", err)
// 		}
// 	} else {
// 		if err := g.previewGoFiles(); err != nil {
// 			return fmt.Errorf("failed to preview Go files: %w", err)
// 		}
// 	}
//
// 	// Report summary
// 	g.printSummary(time.Since(startTime))
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) loadSchemas() error {
// 	info, err := os.Stat(g.input)
// 	if err != nil {
// 		return fmt.Errorf("input path error: %w", err)
// 	}
//
// 	var files []string
//
// 	if info.IsDir() {
// 		err = filepath.WalkDir(g.input, func(path string, d fs.DirEntry, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
// 				files = append(files, path)
// 			}
// 			return nil
// 		})
// 		if err != nil {
// 			return fmt.Errorf("failed to walk directory: %w", err)
// 		}
// 	} else {
// 		files = []string{g.input}
// 	}
//
// 	if len(files) == 0 {
// 		return fmt.Errorf("no JSON files found in %s", g.input)
// 	}
//
// 	g.progress.Total = len(files)
//
// 	// Load files concurrently
// 	return g.loadSchemasConcurrent(files)
// }
//
// func (g *SchemaToGoGenerator) loadSchemasConcurrent(files []string) error {
// 	type result struct {
// 		id     string
// 		schema *JSONSchema
// 		err    error
// 		file   string
// 	}
//
// 	workers := g.maxWorkers
// 	if workers > len(files) {
// 		workers = len(files)
// 	}
//
// 	fileChan := make(chan string, len(files))
// 	resultChan := make(chan result, len(files))
//
// 	// Start workers
// 	var wg sync.WaitGroup
// 	for i := 0; i < workers; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for file := range fileChan {
// 				id, schema, err := g.loadSchemaFileWorker(file)
// 				resultChan <- result{id: id, schema: schema, err: err, file: file}
// 			}
// 		}()
// 	}
//
// 	// Send files to workers
// 	for _, file := range files {
// 		fileChan <- file
// 	}
// 	close(fileChan)
//
// 	// Wait for all workers to finish
// 	go func() {
// 		wg.Wait()
// 		close(resultChan)
// 	}()
//
// 	// Collect results
// 	for res := range resultChan {
// 		g.progress.IncrementProcessed()
//
// 		if res.err != nil {
// 			g.progress.IncrementFailed()
// 			g.addWarning(fmt.Sprintf("Failed to load %s: %v", res.file, res.err))
// 			continue
// 		}
//
// 		g.mu.Lock()
// 		g.schemas[res.id] = res.schema
// 		g.mu.Unlock()
//
// 		if g.verbose {
// 			fmt.Printf("📄 Loaded: %s\n", res.file)
// 		}
// 	}
//
// 	if len(g.schemas) == 0 {
// 		return fmt.Errorf("no valid JSON Schema files found")
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) loadSchemaFileWorker(filePath string) (string, *JSONSchema, error) {
// 	data, err := os.ReadFile(filePath)
// 	if err != nil {
// 		return "", nil, fmt.Errorf("failed to read file: %w", err)
// 	}
//
// 	var schema JSONSchema
// 	if err := json.Unmarshal(data, &schema); err != nil {
// 		return "", nil, fmt.Errorf("failed to parse JSON: %w", err)
// 	}
//
// 	schemaID := schema.ID
// 	if schemaID == "" {
// 		schemaID = g.generateSchemaID(filePath)
// 	}
//
// 	return schemaID, &schema, nil
// }
//
// func (g *SchemaToGoGenerator) validateSchemas() error {
// 	for id, schema := range g.schemas {
// 		if err := g.validateSchema(schema); err != nil {
// 			g.addWarning(fmt.Sprintf("Schema %s validation warning: %v", id, err))
// 		}
// 	}
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) validateSchema(schema *JSONSchema) error {
// 	// Basic validation checks
// 	if schema.Type == nil && len(schema.Properties) == 0 && schema.Ref == "" &&
// 		len(schema.AnyOf) == 0 && len(schema.OneOf) == 0 && len(schema.AllOf) == 0 {
// 		return fmt.Errorf("schema has no type, properties, or composition keywords")
// 	}
//
// 	// Validate references
// 	if schema.Ref != "" {
// 		if _, err := g.resolveRef(schema.Ref); err != nil {
// 			return fmt.Errorf("invalid reference: %w", err)
// 		}
// 	}
//
// 	// Validate nested schemas
// 	for propName, propSchema := range schema.Properties {
// 		if err := g.validateSchema(propSchema); err != nil {
// 			return fmt.Errorf("property %s: %w", propName, err)
// 		}
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) checkNameConflicts() error {
// 	conflicts := make(map[string][]string)
//
// 	for id, goType := range g.types {
// 		if existing, exists := conflicts[goType.Name]; exists {
// 			conflicts[goType.Name] = append(existing, id)
// 		} else {
// 			conflicts[goType.Name] = []string{id}
// 		}
// 	}
//
// 	for name, ids := range conflicts {
// 		if len(ids) > 1 {
// 			g.addWarning(fmt.Sprintf("Name conflict for type %s from schemas: %v", name, ids))
// 		}
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) generateTypes() error {
// 	for schemaID, schema := range g.schemas {
// 		if g.processed[schemaID] {
// 			continue
// 		}
//
// 		if g.verbose {
// 			fmt.Printf("🔄 Processing schema: %s\n", schemaID)
// 		}
//
// 		goType, err := g.processSchema(schemaID, schema)
// 		if err != nil {
// 			g.addWarning(fmt.Sprintf("Failed to process schema %s: %v", schemaID, err))
// 			continue
// 		}
//
// 		if goType != nil {
// 			g.types[schemaID] = goType
// 		}
//
// 		g.processed[schemaID] = true
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) processSchema(name string, schema *JSONSchema) (*GoType, error) {
// 	if schema.Ref != "" {
// 		resolvedSchema, err := g.resolveRef(schema.Ref)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to resolve reference %s: %w", schema.Ref, err)
// 		}
// 		return g.processSchema(name, resolvedSchema)
// 	}
//
// 	if len(schema.AnyOf) > 0 {
// 		enumValues := g.extractEnumFromAnyOf(schema.AnyOf)
// 		if len(enumValues) > 0 {
// 			return g.processEnumSchemaFromValues(name, schema, enumValues)
// 		}
// 		return nil, nil
// 	}
//
// 	schemaType := g.getSchemaType(schema)
//
// 	switch schemaType {
// 	case "object":
// 		return g.processObjectSchema(name, schema)
// 	case "string":
// 		if len(schema.Enum) > 0 {
// 			return g.processEnumSchema(name, schema)
// 		}
// 		return nil, nil
// 	default:
// 		return nil, nil
// 	}
// }
//
// func (g *SchemaToGoGenerator) processObjectSchema(name string, schema *JSONSchema) (*GoType, error) {
// 	goType := &GoType{
// 		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
// 		Comment:    schema.Description,
// 		SourceFile: name,
// 		Fields:     []GoField{},
// 		Imports:    []string{},
// 	}
//
// 	importsSet := make(map[string]bool)
//
// 	if schema.Properties != nil {
// 		propNames := make([]string, 0, len(schema.Properties))
// 		for propName := range schema.Properties {
// 			propNames = append(propNames, propName)
// 		}
// 		sort.Strings(propNames)
//
// 		// Track field names to ensure uniqueness within the struct
// 		usedFieldNames := make(map[string]int)
//
// 		for _, propName := range propNames {
// 			propSchema := schema.Properties[propName]
//
// 			field, err := g.processPropertyWithContext(propName, propSchema, schema.Required, goType.Name)
// 			if err != nil {
// 				g.addWarning(fmt.Sprintf("Failed to process property %s.%s: %v", name, propName, err))
// 				continue
// 			}
//
// 			// Ensure field name is unique within this struct
// 			originalName := field.Name
// 			if count, exists := usedFieldNames[originalName]; exists {
// 				field.Name = fmt.Sprintf("%s%d", originalName, count+1)
// 				usedFieldNames[originalName] = count + 1
// 			} else {
// 				usedFieldNames[originalName] = 0
// 			}
//
// 			goType.Fields = append(goType.Fields, *field)
//
// 			for _, imp := range g.getFieldImports(field.Type) {
// 				importsSet[imp] = true
// 			}
//
// 			if dep := g.extractTypeDependency(field.Type); dep != "" {
// 				if g.canGenerateType(dep) {
// 					goType.Dependencies = append(goType.Dependencies, dep)
// 				}
// 			}
// 		}
// 	}
//
// 	for imp := range importsSet {
// 		goType.Imports = append(goType.Imports, imp)
// 	}
// 	sort.Strings(goType.Imports)
//
// 	if len(goType.Fields) == 0 && len(goType.EnumValues) == 0 {
// 		return nil, nil
// 	}
//
// 	return goType, nil
// }
//
// func (g *SchemaToGoGenerator) extractEnumFromAnyOf(anyOf []*JSONSchema) []any {
// 	var enumValues []any
//
// 	for _, subSchema := range anyOf {
// 		if subSchema.Type != nil {
// 			if typeStr, ok := subSchema.Type.(string); ok && typeStr == "string" {
// 				if subSchema.Const != nil {
// 					enumValues = append(enumValues, subSchema.Const)
// 				}
// 			}
// 		}
// 		if subSchema.Type != nil {
// 			if typeArray, ok := subSchema.Type.([]any); ok {
// 				for _, t := range typeArray {
// 					if typeStr, ok := t.(string); ok && typeStr == "string" {
// 						if subSchema.Const != nil {
// 							enumValues = append(enumValues, subSchema.Const)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
//
// 	return enumValues
// }
//
// func (g *SchemaToGoGenerator) processEnumSchemaFromValues(name string, schema *JSONSchema, enumValues []any) (*GoType, error) {
// 	goType := &GoType{
// 		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
// 		Comment:    schema.Description,
// 		SourceFile: name,
// 		IsEnum:     true,
// 		EnumValues: []string{},
// 	}
//
// 	seen := make(map[string]bool)
// 	for _, value := range enumValues {
// 		if str, ok := value.(string); ok {
// 			if !seen[str] {
// 				goType.EnumValues = append(goType.EnumValues, str)
// 				seen[str] = true
// 			}
// 		}
// 	}
//
// 	if len(goType.EnumValues) == 0 {
// 		return nil, nil
// 	}
//
// 	return goType, nil
// }
//
// func (g *SchemaToGoGenerator) processPropertyWithContext(propName string, propSchema *JSONSchema, required []string, currentTypeName string) (*GoField, error) {
// 	field := &GoField{
// 		Name:     g.sanitizeFieldName(propName),
// 		JSONTag:  propName,
// 		Comment:  propSchema.Description,
// 		Optional: !g.isRequired(propName, required),
// 	}
//
// 	if propName == "$$id" || propName == "$ref" || propName == "$schema" {
// 		field.Type = "any"
// 	} else {
// 		goType, err := g.generateGoTypeWithContext(propSchema, currentTypeName)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to generate Go type: %w", err)
// 		}
//
// 		if g.isCustomType(goType) && !g.canGenerateType(goType) {
// 			field.Type = "any"
// 		} else {
// 			field.Type = goType
// 		}
// 	}
//
// 	field.UITag = g.generateUITag(propSchema, propName)
// 	field.ValidateTag = g.generateValidateTag(propSchema, !field.Optional)
//
// 	return field, nil
// }
//
// func (g *SchemaToGoGenerator) processEnumSchema(name string, schema *JSONSchema) (*GoType, error) {
// 	goType := &GoType{
// 		Name:       g.getUniqueName(g.sanitizeTypeName(name)),
// 		Comment:    schema.Description,
// 		SourceFile: name,
// 		IsEnum:     true,
// 		EnumValues: []string{},
// 	}
//
// 	seen := make(map[string]bool)
// 	for _, value := range schema.Enum {
// 		if str, ok := value.(string); ok {
// 			constantName := g.toPascalCase(str)
//
// 			if str == "" {
// 				constantName = "Empty"
// 			}
//
// 			typeName := g.sanitizeTypeName(name)
// 			if constantName == typeName {
// 				constantName = typeName + "Default"
// 			}
//
// 			if !seen[constantName] {
// 				goType.EnumValues = append(goType.EnumValues, str)
// 				seen[constantName] = true
// 			}
// 		}
// 	}
//
// 	return goType, nil
// }
//
// func (g *SchemaToGoGenerator) generateGoTypeWithContext(schema *JSONSchema, currentTypeName string) (string, error) {
// 	if schema.Ref != "" {
// 		refName := g.extractRefName(schema.Ref)
// 		refTypeName := g.sanitizeTypeName(refName)
//
// 		if refTypeName == currentTypeName {
// 			return "*" + refTypeName, nil
// 		}
//
// 		if refSchema, err := g.resolveRef(schema.Ref); err == nil {
// 			if resolvedType, err := g.resolveToFinalType(refSchema); err == nil {
// 				return resolvedType, nil
// 			}
//
// 			refType := g.getSchemaType(refSchema)
// 			if refType == "string" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
// 				return g.getGoTypeForFormat(refSchema.Format, "string"), nil
// 			}
// 			if refType == "integer" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
// 				return "int", nil
// 			}
// 			if refType == "number" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
// 				return "float64", nil
// 			}
// 			if refType == "boolean" && len(refSchema.Properties) == 0 && len(refSchema.Enum) == 0 {
// 				return "bool", nil
// 			}
//
// 			if !g.wouldGenerateType(refSchema) {
// 				return "any", nil
// 			}
//
// 			return refTypeName, nil
// 		} else {
// 			g.addWarning(fmt.Sprintf("Referenced schema not found: %s, using any", schema.Ref))
// 			return "any", nil
// 		}
// 	}
//
// 	if len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
// 		return "any", nil
// 	}
//
// 	schemaType := g.getSchemaType(schema)
//
// 	switch schemaType {
// 	case "string":
// 		return g.getGoTypeForFormat(schema.Format, "string"), nil
// 	case "integer":
// 		return "int", nil
// 	case "number":
// 		return "float64", nil
// 	case "boolean":
// 		return "bool", nil
// 	case "array":
// 		if schema.Items != nil {
// 			itemType, err := g.generateGoTypeWithContext(schema.Items, currentTypeName)
// 			if err != nil {
// 				return "", fmt.Errorf("failed to generate array item type: %w", err)
// 			}
// 			return "[]" + itemType, nil
// 		}
// 		return "[]any", nil
// 	case "object":
// 		if schema.AdditionalProperties != nil {
// 			if addProps, ok := schema.AdditionalProperties.(bool); ok && addProps {
// 				return "map[string]any", nil
// 			}
// 			if addPropsSchema, ok := schema.AdditionalProperties.(*JSONSchema); ok {
// 				valueType, err := g.generateGoTypeWithContext(addPropsSchema, currentTypeName)
// 				if err != nil {
// 					return "", fmt.Errorf("failed to generate additionalProperties type: %w", err)
// 				}
// 				return "map[string]" + valueType, nil
// 			}
// 		}
// 		if len(schema.Properties) > 0 {
// 			return "map[string]any", nil
// 		}
// 		return "map[string]any", nil
// 	default:
// 		return "any", nil
// 	}
// }
//
// func (g *SchemaToGoGenerator) getGoTypeForFormat(format, baseType string) string {
// 	switch format {
// 	case "date-time":
// 		return "time.Time"
// 	case "date":
// 		return "time.Time"
// 	case "time":
// 		return "time.Time"
// 	case "uuid":
// 		return "uuid.UUID"
// 	case "email", "hostname", "ipv4", "ipv6", "uri", "uri-reference":
// 		return "string"
// 	default:
// 		return baseType
// 	}
// }
//
// func (g *SchemaToGoGenerator) generateUITag(schema *JSONSchema, propName string) string {
// 	var parts []string
//
// 	if schema.UIComponent != "" {
// 		parts = append(parts, "component="+schema.UIComponent)
// 	} else {
// 		schemaType := g.getSchemaType(schema)
// 		switch schemaType {
// 		case "string":
// 			if len(schema.Enum) > 0 {
// 				parts = append(parts, "component=select")
// 			} else {
// 				parts = append(parts, "component=text")
// 			}
// 		case "boolean":
// 			parts = append(parts, "component=toggle")
// 		case "integer", "number":
// 			parts = append(parts, "component=number")
// 		case "array":
// 			parts = append(parts, "component=multi-select")
// 		}
// 	}
//
// 	label := schema.UILabel
// 	if label == "" && schema.Title != "" {
// 		label = schema.Title
// 	}
// 	if label == "" {
// 		label = g.humanizeFieldName(propName)
// 	}
// 	if label != "" {
// 		parts = append(parts, "label="+label)
// 	}
//
// 	if schema.UIRequired {
// 		parts = append(parts, "required=true")
// 	}
//
// 	if schema.UIHidden {
// 		parts = append(parts, "hidden=true")
// 	}
//
// 	if len(schema.Enum) > 0 {
// 		var options []string
// 		for _, val := range schema.Enum {
// 			if str, ok := val.(string); ok {
// 				options = append(options, str)
// 			}
// 		}
// 		if len(options) > 0 {
// 			parts = append(parts, "options="+strings.Join(options, ","))
// 		}
// 	}
//
// 	if len(parts) > 0 {
// 		return strings.Join(parts, ";")
// 	}
//
// 	return ""
// }
//
// func (g *SchemaToGoGenerator) generateValidateTag(schema *JSONSchema, required bool) string {
// 	var parts []string
//
// 	if required {
// 		parts = append(parts, "required")
// 	}
//
// 	schemaType := g.getSchemaType(schema)
// 	switch schemaType {
// 	case "string":
// 		if len(schema.Enum) > 0 {
// 			var options []string
// 			for _, val := range schema.Enum {
// 				if str, ok := val.(string); ok {
// 					options = append(options, str)
// 				}
// 			}
// 			if len(options) > 0 {
// 				parts = append(parts, "oneof="+strings.Join(options, " "))
// 			}
// 		}
// 		if schema.MinLength != nil {
// 			parts = append(parts, fmt.Sprintf("min=%d", *schema.MinLength))
// 		}
// 		if schema.MaxLength != nil {
// 			parts = append(parts, fmt.Sprintf("max=%d", *schema.MaxLength))
// 		}
// 		if schema.Pattern != "" {
// 			parts = append(parts, fmt.Sprintf("regexp=%s", schema.Pattern))
// 		}
// 		switch schema.Format {
// 		case "email":
// 			parts = append(parts, "email")
// 		case "uri", "uri-reference":
// 			parts = append(parts, "uri")
// 		case "uuid":
// 			parts = append(parts, "uuid")
// 		case "ipv4":
// 			parts = append(parts, "ipv4")
// 		}
// 	case "number", "integer":
// 		if schema.Minimum != nil {
// 			parts = append(parts, fmt.Sprintf("min=%v", *schema.Minimum))
// 		}
// 		if schema.Maximum != nil {
// 			parts = append(parts, fmt.Sprintf("max=%v", *schema.Maximum))
// 		}
// 	}
//
// 	if len(parts) == 0 {
// 		return ""
// 	}
// 	return strings.Join(parts, ",")
// }
//
// // Helper functions
//
// func isValidPackageName(name string) bool {
// 	if name == "" {
// 		return false
// 	}
// 	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
// 	return validName.MatchString(name)
// }
//
// func (g *SchemaToGoGenerator) addWarning(warning string) {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()
// 	g.warnings = append(g.warnings, warning)
// 	g.progress.IncrementWarnings()
// }
//
// func (g *SchemaToGoGenerator) getSchemaType(schema *JSONSchema) string {
// 	if schema.Type == nil {
// 		if len(schema.Properties) > 0 {
// 			return "object"
// 		}
// 		return "any"
// 	}
//
// 	switch t := schema.Type.(type) {
// 	case string:
// 		return t
// 	case []any:
// 		if len(t) > 0 {
// 			if firstType, ok := t[0].(string); ok {
// 				return firstType
// 			}
// 		}
// 		return "any"
// 	default:
// 		return "any"
// 	}
// }
//
// func (g *SchemaToGoGenerator) getUniqueName(name string) string {
// 	g.mu.Lock()
// 	defer g.mu.Unlock()
//
// 	if existing, exists := g.usedNames[name]; exists {
// 		g.usedNames[name] = existing + 1
// 		return fmt.Sprintf("%s%d", name, existing+1)
// 	}
//
// 	g.usedNames[name] = 0
// 	return name
// }
//
// func (g *SchemaToGoGenerator) sanitizeTypeName(name string) string {
// 	name = strings.TrimSpace(name)
// 	if name == "" {
// 		return "UnknownType"
// 	}
//
// 	// Extract the base name from file path or URL
// 	if strings.Contains(name, "/") {
// 		parts := strings.Split(name, "/")
// 		name = parts[len(parts)-1]
// 	}
//
// 	// Remove file extension
// 	if strings.HasSuffix(name, ".json") {
// 		name = strings.TrimSuffix(name, ".json")
// 	}
//
// 	// Convert to PascalCase
// 	return g.toPascalCase(name)
// }
//
// func (g *SchemaToGoGenerator) sanitizeFieldName(name string) string {
// 	if name == "" {
// 		return "Field"
// 	}
//
// 	// Handle special JSON Schema fields
// 	switch name {
// 	case "$$id":
// 		return "SchemaID"
// 	case "$ref":
// 		return "Ref"
// 	case "$schema":
// 		return "Schema"
// 	}
//
// 	// Remove invalid characters and convert to PascalCase
// 	sanitized := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(name, "")
// 	return g.toPascalCase(sanitized)
// }
//
// func (g *SchemaToGoGenerator) toPascalCase(s string) string {
// 	if s == "" {
// 		return ""
// 	}
//
// 	// Replace common separators including dots
// 	s = regexp.MustCompile(`[-_\s.]+`).ReplaceAllString(s, " ")
//
// 	// Remove angle bracket expressions like <(string|number)>
// 	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
//
// 	words := strings.Fields(s)
// 	var result strings.Builder
//
// 	for _, word := range words {
// 		if len(word) > 0 {
// 			result.WriteString(strings.ToUpper(word[:1]))
// 			if len(word) > 1 {
// 				result.WriteString(strings.ToLower(word[1:]))
// 			}
// 		}
// 	}
//
// 	pascalCase := result.String()
//
// 	// Ensure first character is uppercase
// 	if len(pascalCase) > 0 && (pascalCase[0] < 'A' || pascalCase[0] > 'Z') {
// 		return "Field" + pascalCase
// 	}
//
// 	return pascalCase
// }
//
// func (g *SchemaToGoGenerator) humanizeFieldName(fieldName string) string {
// 	// Convert camelCase/PascalCase to human readable
// 	re := regexp.MustCompile(`([a-z])([A-Z])`)
// 	humanized := re.ReplaceAllString(fieldName, "$1 $2")
//
// 	// Capitalize first letter
// 	if len(humanized) > 0 {
// 		humanized = strings.ToUpper(humanized[:1]) + humanized[1:]
// 	}
//
// 	return humanized
// }
//
// func (g *SchemaToGoGenerator) isRequired(propName string, required []string) bool {
// 	for _, req := range required {
// 		if req == propName {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func (g *SchemaToGoGenerator) generateSchemaID(filePath string) string {
// 	base := filepath.Base(filePath)
// 	ext := filepath.Ext(base)
// 	return strings.TrimSuffix(base, ext)
// }
//
// func (g *SchemaToGoGenerator) resolveRef(ref string) (*JSONSchema, error) {
// 	// Cache resolved references
// 	if cached, exists := g.refCache[ref]; exists {
// 		return cached, nil
// 	}
//
// 	// Parse reference
// 	if strings.HasPrefix(ref, "#/") {
// 		// Internal reference - not fully implemented for this version
// 		return nil, fmt.Errorf("internal references not supported: %s", ref)
// 	}
//
// 	// External reference - would need to be loaded from file/URL
// 	return nil, fmt.Errorf("external references not supported: %s", ref)
// }
//
// func (g *SchemaToGoGenerator) resolveToFinalType(schema *JSONSchema) (string, error) {
// 	// Basic type resolution for simple schemas
// 	schemaType := g.getSchemaType(schema)
// 	switch schemaType {
// 	case "string":
// 		return g.getGoTypeForFormat(schema.Format, "string"), nil
// 	case "integer":
// 		return "int", nil
// 	case "number":
// 		return "float64", nil
// 	case "boolean":
// 		return "bool", nil
// 	default:
// 		return "", fmt.Errorf("cannot resolve complex type")
// 	}
// }
//
// func (g *SchemaToGoGenerator) wouldGenerateType(schema *JSONSchema) bool {
// 	schemaType := g.getSchemaType(schema)
// 	return schemaType == "object" || (schemaType == "string" && len(schema.Enum) > 0)
// }
//
// func (g *SchemaToGoGenerator) isCustomType(goType string) bool {
// 	builtinTypes := map[string]bool{
// 		"string": true, "int": true, "int32": true, "int64": true,
// 		"float32": true, "float64": true, "bool": true, "any": true,
// 		"time.Time": true, "uuid.UUID": true,
// 	}
//
// 	// Check for array/slice types
// 	if strings.HasPrefix(goType, "[]") {
// 		return g.isCustomType(goType[2:])
// 	}
//
// 	// Check for map types
// 	if strings.HasPrefix(goType, "map[") {
// 		return false // maps are considered built-in for this purpose
// 	}
//
// 	// Check for pointer types
// 	if strings.HasPrefix(goType, "*") {
// 		return g.isCustomType(goType[1:])
// 	}
//
// 	return !builtinTypes[goType]
// }
//
// func (g *SchemaToGoGenerator) canGenerateType(typeName string) bool {
// 	// For this version, assume we can generate types that are referenced
// 	return true
// }
//
// func (g *SchemaToGoGenerator) extractTypeDependency(goType string) string {
// 	// Extract type name from complex types
// 	if strings.HasPrefix(goType, "[]") {
// 		return g.extractTypeDependency(goType[2:])
// 	}
// 	if strings.HasPrefix(goType, "*") {
// 		return g.extractTypeDependency(goType[1:])
// 	}
// 	if strings.HasPrefix(goType, "map[") {
// 		return ""
// 	}
//
// 	// Return the type if it's a custom type
// 	if g.isCustomType(goType) {
// 		return goType
// 	}
//
// 	return ""
// }
//
// func (g *SchemaToGoGenerator) getFieldImports(goType string) []string {
// 	var imports []string
//
// 	if strings.Contains(goType, "time.Time") {
// 		imports = append(imports, "time")
// 	}
// 	if strings.Contains(goType, "uuid.UUID") {
// 		imports = append(imports, "github.com/google/uuid")
// 	}
//
// 	return imports
// }
//
// func (g *SchemaToGoGenerator) extractRefName(ref string) string {
// 	// Extract name from reference
// 	if strings.HasPrefix(ref, "#/") {
// 		parts := strings.Split(ref, "/")
// 		if len(parts) > 0 {
// 			return parts[len(parts)-1]
// 		}
// 	}
//
// 	// For external refs, parse URL
// 	if u, err := url.Parse(ref); err == nil {
// 		return filepath.Base(u.Path)
// 	}
//
// 	return ref
// }
//
// func (g *SchemaToGoGenerator) writeGoFiles() error {
// 	if err := os.MkdirAll(g.output, 0o755); err != nil {
// 		return fmt.Errorf("failed to create output directory: %w", err)
// 	}
//
// 	if g.single {
// 		return g.writeSingleFile()
// 	}
//
// 	return g.writeMultipleFiles()
// }
//
// func (g *SchemaToGoGenerator) previewGoFiles() error {
// 	if g.single {
// 		return g.previewSingleFile()
// 	}
//
// 	return g.previewMultipleFiles()
// }
//
// func (g *SchemaToGoGenerator) writeSingleFile() error {
// 	fileName := "types.go"
// 	filePath := filepath.Join(g.output, fileName)
//
// 	content, err := g.generateSingleFileContent()
// 	if err != nil {
// 		return fmt.Errorf("failed to generate single file content: %w", err)
// 	}
//
// 	formatted, err := format.Source([]byte(content))
// 	if err != nil {
// 		g.addWarning(fmt.Sprintf("Failed to format generated code: %v", err))
// 		formatted = []byte(content)
// 	}
//
// 	if err := os.WriteFile(filePath, formatted, 0o644); err != nil {
// 		return fmt.Errorf("failed to write file %s: %w", filePath, err)
// 	}
//
// 	if g.verbose {
// 		fmt.Printf("✅ Generated: %s\n", filePath)
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) writeMultipleFiles() error {
// 	for _, goType := range g.types {
// 		fileName := strings.ToLower(goType.Name) + ".go"
// 		filePath := filepath.Join(g.output, fileName)
//
// 		content, err := g.generateFileContent(goType)
// 		if err != nil {
// 			g.addWarning(fmt.Sprintf("Failed to generate content for %s: %v", goType.Name, err))
// 			continue
// 		}
//
// 		formatted, err := format.Source([]byte(content))
// 		if err != nil {
// 			g.addWarning(fmt.Sprintf("Failed to format %s: %v", goType.Name, err))
// 			formatted = []byte(content)
// 		}
//
// 		if err := os.WriteFile(filePath, formatted, 0o644); err != nil {
// 			g.addWarning(fmt.Sprintf("Failed to write %s: %v", filePath, err))
// 			continue
// 		}
//
// 		if g.verbose {
// 			fmt.Printf("✅ Generated: %s\n", filePath)
// 		}
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) previewSingleFile() error {
// 	content, err := g.generateSingleFileContent()
// 	if err != nil {
// 		return fmt.Errorf("failed to generate preview content: %w", err)
// 	}
//
// 	fmt.Printf("\n=== PREVIEW: types.go ===\n")
// 	fmt.Println(content)
// 	fmt.Printf("=== END PREVIEW ===\n\n")
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) previewMultipleFiles() error {
// 	for _, goType := range g.types {
// 		fileName := strings.ToLower(goType.Name) + ".go"
//
// 		content, err := g.generateFileContent(goType)
// 		if err != nil {
// 			g.addWarning(fmt.Sprintf("Failed to generate preview for %s: %v", goType.Name, err))
// 			continue
// 		}
//
// 		fmt.Printf("\n=== PREVIEW: %s ===\n", fileName)
// 		fmt.Println(content)
// 		fmt.Printf("=== END PREVIEW ===\n\n")
// 	}
//
// 	return nil
// }
//
// func (g *SchemaToGoGenerator) generateSingleFileContent() (string, error) {
// 	tmpl := template.Must(template.New("single").Funcs(template.FuncMap{
// 		"title": strings.Title,
// 		"enumConstName": func(typeName, value string) string {
// 			// Handle special cases for empty values
// 			if value == "" {
// 				return typeName + "Empty"
// 			}
// 			// Clean the value for Go identifier
// 			cleaned := g.toPascalCase(value)
// 			// Ensure it doesn't conflict with type name
// 			if cleaned == typeName {
// 				cleaned = typeName + "Default"
// 			}
// 			return typeName + cleaned
// 		},
// 		"comment": func(text string) string {
// 			if text == "" {
// 				return ""
// 			}
// 			// Replace newlines and normalize whitespace for inline comments
// 			cleaned := strings.ReplaceAll(text, "\n", " ")
// 			cleaned = strings.ReplaceAll(cleaned, "\r", " ")
// 			cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
// 			cleaned = strings.TrimSpace(cleaned)
// 			return cleaned
// 		},
// 	}).Parse(singleFileTemplate))
//
// 	data := struct {
// 		Package   string
// 		Generated string
// 		Imports   []string
// 		Types     []*GoType
// 	}{
// 		Package:   g.pkg,
// 		Generated: time.Now().Format(time.RFC3339),
// 		Imports:   g.collectAllImports(),
// 		Types:     g.getSortedTypes(),
// 	}
//
// 	var buf strings.Builder
// 	if err := tmpl.Execute(&buf, data); err != nil {
// 		return "", fmt.Errorf("template execution failed: %w", err)
// 	}
//
// 	return buf.String(), nil
// }
//
// func (g *SchemaToGoGenerator) generateFileContent(goType *GoType) (string, error) {
// 	tmpl := template.Must(template.New("file").Funcs(template.FuncMap{
// 		"title": strings.Title,
// 		"enumConstName": func(typeName, value string) string {
// 			// Handle special cases for empty values
// 			if value == "" {
// 				return typeName + "Empty"
// 			}
// 			// Clean the value for Go identifier
// 			cleaned := g.toPascalCase(value)
// 			// Ensure it doesn't conflict with type name
// 			if cleaned == typeName {
// 				cleaned = typeName + "Default"
// 			}
// 			return typeName + cleaned
// 		},
// 		"comment": func(text string) string {
// 			if text == "" {
// 				return ""
// 			}
// 			// Replace newlines and normalize whitespace for inline comments
// 			cleaned := strings.ReplaceAll(text, "\n", " ")
// 			cleaned = strings.ReplaceAll(cleaned, "\r", " ")
// 			cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
// 			cleaned = strings.TrimSpace(cleaned)
// 			return cleaned
// 		},
// 	}).Parse(fileTemplate))
//
// 	data := struct {
// 		Package   string
// 		Generated string
// 		Imports   []string
// 		Type      *GoType
// 	}{
// 		Package:   g.pkg,
// 		Generated: time.Now().Format(time.RFC3339),
// 		Imports:   goType.Imports,
// 		Type:      goType,
// 	}
//
// 	var buf strings.Builder
// 	if err := tmpl.Execute(&buf, data); err != nil {
// 		return "", fmt.Errorf("template execution failed: %w", err)
// 	}
//
// 	return buf.String(), nil
// }
//
// func (g *SchemaToGoGenerator) collectAllImports() []string {
// 	importSet := make(map[string]bool)
//
// 	for _, goType := range g.types {
// 		for _, imp := range goType.Imports {
// 			importSet[imp] = true
// 		}
// 	}
//
// 	var imports []string
// 	for imp := range importSet {
// 		imports = append(imports, imp)
// 	}
//
// 	sort.Strings(imports)
// 	return imports
// }
//
// func (g *SchemaToGoGenerator) getSortedTypes() []*GoType {
// 	var types []*GoType
// 	for _, goType := range g.types {
// 		types = append(types, goType)
// 	}
//
// 	sort.Slice(types, func(i, j int) bool {
// 		return types[i].Name < types[j].Name
// 	})
//
// 	return types
// }
//
// func (g *SchemaToGoGenerator) printSummary(duration time.Duration) {
// 	fmt.Printf("\n🎉 Schema to Go conversion completed in %v\n", duration)
// 	fmt.Printf("📊 Summary:\n")
// 	fmt.Printf("   • Schemas processed: %d\n", len(g.schemas))
// 	fmt.Printf("   • Types generated: %d\n", len(g.types))
// 	fmt.Printf("   • Warnings: %d\n", len(g.warnings))
//
// 	if len(g.warnings) > 0 && g.verbose {
// 		fmt.Printf("\n⚠️  Warnings:\n")
// 		for _, warning := range g.warnings {
// 			fmt.Printf("   • %s\n", warning)
// 		}
// 	}
//
// 	if !g.dryRun {
// 		fmt.Printf("   • Output directory: %s\n", g.output)
// 	} else {
// 		fmt.Printf("   • Dry run mode - no files written\n")
// 	}
// }
//
// // Templates
//
// const singleFileTemplate = `// Code generated by awoctl schema to-go. DO NOT EDIT.
// // Generated: {{.Generated}}
//
// package {{.Package}}
//
// {{if .Imports}}
// import (
// {{range .Imports}}	"{{.}}"
// {{end}})
//
// {{end}}{{range .Types}}{{if .IsEnum}}// {{.Name}} represents the {{.Name}} enum type
// {{if .Comment}}// {{.Comment | comment}}{{end}}
// type {{.Name}} string
//
// const (
// {{range $i, $value := .EnumValues}}	{{enumConstName $.Name $value}} = "{{$value}}"
// {{end}})
//
// {{else}}// {{.Name}} represents the {{.Name}} struct
// {{if .Comment}}// {{.Comment | comment}}{{end}}
// type {{.Name}} struct {
// {{range .Fields}}	{{.Name}} {{.Type}} ` + "`" + `json:"{{.JSONTag}}"{{if .UITag}} ui:"{{.UITag}}"{{end}}{{if .ValidateTag}} validate:"{{.ValidateTag}}"{{end}}` + "`" + `{{if .Comment}} // {{.Comment | comment}}{{end}}
// {{end}}}
//
// {{end}}{{end}}`
//
// const fileTemplate = `// Code generated by awoctl schema to-go. DO NOT EDIT.
// // Generated: {{.Generated}}
//
// package {{.Package}}
//
// {{if .Imports}}
// import (
// {{range .Imports}}	"{{.}}"
// {{end}})
//
// {{end}}{{if .Type.IsEnum}}// {{.Type.Name}} represents the {{.Type.Name}} enum type
// {{if .Type.Comment}}// {{.Type.Comment | comment}}{{end}}
// type {{.Type.Name}} string
//
// const (
// {{range $i, $value := .Type.EnumValues}}	{{enumConstName $.Type.Name $value}} = "{{$value}}"
// {{end}})
// {{else}}// {{.Type.Name}} represents the {{.Type.Name}} struct
// {{if .Type.Comment}}// {{.Type.Comment | comment}}{{end}}
// type {{.Type.Name}} struct {
// {{range .Type.Fields}}	{{.Name}} {{.Type}} ` + "`" + `json:"{{.JSONTag}}"{{if .UITag}} ui:"{{.UITag}}"{{end}}{{if .ValidateTag}} validate:"{{.ValidateTag}}"{{end}}` + "`" + `{{if .Comment}} // {{.Comment | comment}}{{end}}
// {{end}}}
// {{end}}`
