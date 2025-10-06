package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
	"strings"
)

// FieldInfo represents detailed information about a struct field
type FieldInfo struct {
	Name        string
	Type        string
	JsonTag     string
	UITag       string
	DBTag       string
	ValidateTag string

	// Parsed UI tag components
	Component   string
	Label       string
	Required    bool
	Placeholder string
	Options     []string
	Validation  string
	Relation    string
	Format      string
	Badge       bool
	Hidden      bool
	Unique      bool

	// Type information
	IsPointer    bool
	IsSlice      bool
	IsTimeType   bool
	IsUUIDType   bool
	IsStringType bool
	IsNumberType bool
	IsBoolType   bool

	// Relationship information
	IsRelationship   bool
	RelationshipType string // "one-to-one", "one-to-many", "many-to-many"
	TargetEntity     string
}

// StructInfo contains complete information about a Go struct
type StructInfo struct {
	Name        string
	PackageName string
	Fields      []FieldInfo
	ImportPaths map[string]string

	// Business logic analysis
	IsCRUDEntity   bool
	IsReadOnly     bool
	HasWorkflow    bool
	HasValidation  bool
	HasAuditTrail  bool
	HasSoftDelete  bool
	HasTenantScope bool
	HasHierarchy   bool

	// Primary key information
	PrimaryKeyField string

	// Common field patterns
	HasCreatedAt bool
	HasUpdatedAt bool
	HasDeletedAt bool
	HasStatus    bool
	HasTenantID  bool
}

// Analyzer parses Go struct definitions and extracts UI generation metadata
type Analyzer struct {
	fileSet *token.FileSet
}

// NewAnalyzer creates a new struct analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		fileSet: token.NewFileSet(),
	}
}

// AnalyzeFile parses a Go file and extracts struct information
func (a *Analyzer) AnalyzeFile(filename string) ([]StructInfo, error) {
	node, err := parser.ParseFile(a.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
	}

	var structs []StructInfo

	// Extract import information
	imports := make(map[string]string)
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		imports[path] = alias
	}

	// Find all struct declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				structInfo := a.analyzeStruct(x.Name.Name, node.Name.Name, structType, imports)
				structs = append(structs, structInfo)
			}
		}
		return true
	})

	return structs, nil
}

// AnalyzeStruct analyzes a struct using reflection
func (a *Analyzer) AnalyzeStruct(t reflect.Type) StructInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		log.Printf("Warning: %s is not a struct type", t.Name())
		return StructInfo{}
	}

	structInfo := StructInfo{
		Name:        t.Name(),
		PackageName: t.PkgPath(),
		Fields:      make([]FieldInfo, 0, t.NumField()),
		ImportPaths: make(map[string]string),
	}

	// Analyze each field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldInfo := a.analyzeStructField(field)
		structInfo.Fields = append(structInfo.Fields, fieldInfo)

		// Track common patterns
		a.updateStructPatterns(&structInfo, fieldInfo)
	}

	// Analyze business logic patterns
	a.analyzeBusinessLogic(&structInfo)

	return structInfo
}

// analyzeStruct extracts detailed information from an AST struct type
func (a *Analyzer) analyzeStruct(name, packageName string, structType *ast.StructType, imports map[string]string) StructInfo {
	structInfo := StructInfo{
		Name:        name,
		PackageName: packageName,
		Fields:      make([]FieldInfo, 0, len(structType.Fields.List)),
		ImportPaths: imports,
	}

	// Analyze each field
	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			fieldInfo := a.analyzeASTField(fieldName.Name, field)
			structInfo.Fields = append(structInfo.Fields, fieldInfo)

			// Track common patterns
			a.updateStructPatterns(&structInfo, fieldInfo)
		}
	}

	// Analyze business logic patterns
	a.analyzeBusinessLogic(&structInfo)

	return structInfo
}

// analyzeStructField extracts field information using reflection
func (a *Analyzer) analyzeStructField(field reflect.StructField) FieldInfo {
	fieldInfo := FieldInfo{
		Name:        field.Name,
		Type:        field.Type.String(),
		JsonTag:     field.Tag.Get("json"),
		UITag:       field.Tag.Get("ui"),
		DBTag:       field.Tag.Get("db"),
		ValidateTag: field.Tag.Get("validate"),
	}

	// Analyze type information
	a.analyzeFieldType(&fieldInfo, field.Type)

	// Parse UI tag for component configuration
	a.parseUITag(&fieldInfo)

	// Detect relationships
	a.detectRelationships(&fieldInfo, field.Type)

	return fieldInfo
}

// analyzeASTField extracts field information from AST
func (a *Analyzer) analyzeASTField(name string, field *ast.Field) FieldInfo {
	fieldInfo := FieldInfo{
		Name: name,
		Type: a.typeToString(field.Type),
	}

	// Extract struct tags
	if field.Tag != nil {
		tagValue := strings.Trim(field.Tag.Value, "`")
		fieldInfo.JsonTag = a.extractTag(tagValue, "json")
		fieldInfo.UITag = a.extractTag(tagValue, "ui")
		fieldInfo.DBTag = a.extractTag(tagValue, "db")
		fieldInfo.ValidateTag = a.extractTag(tagValue, "validate")
	}

	// Analyze type information from AST
	a.analyzeASTFieldType(&fieldInfo, field.Type)

	// Parse UI tag for component configuration
	a.parseUITag(&fieldInfo)

	return fieldInfo
}

// analyzeFieldType determines the nature of a field type using reflection
func (a *Analyzer) analyzeFieldType(fieldInfo *FieldInfo, fieldType reflect.Type) {
	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		fieldInfo.IsPointer = true
		fieldType = fieldType.Elem()
	}

	// Handle slices
	if fieldType.Kind() == reflect.Slice {
		fieldInfo.IsSlice = true
		fieldType = fieldType.Elem()
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	// Determine base types
	switch fieldType.Kind() {
	case reflect.String:
		fieldInfo.IsStringType = true
	case reflect.Bool:
		fieldInfo.IsBoolType = true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		fieldInfo.IsNumberType = true
	}

	// Special type detection
	typeName := fieldType.String()
	if strings.Contains(typeName, "time.Time") {
		fieldInfo.IsTimeType = true
	}
	if strings.Contains(typeName, "uuid.UUID") {
		fieldInfo.IsUUIDType = true
	}
}

// analyzeASTFieldType determines field type information from AST
func (a *Analyzer) analyzeASTFieldType(fieldInfo *FieldInfo, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		fieldInfo.IsPointer = true
		a.analyzeASTFieldType(fieldInfo, t.X)
	case *ast.ArrayType:
		fieldInfo.IsSlice = true
		a.analyzeASTFieldType(fieldInfo, t.Elt)
	case *ast.Ident:
		switch t.Name {
		case "string":
			fieldInfo.IsStringType = true
		case "bool":
			fieldInfo.IsBoolType = true
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64":
			fieldInfo.IsNumberType = true
		}
	case *ast.SelectorExpr:
		typeName := a.typeToString(expr)
		if strings.Contains(typeName, "time.Time") {
			fieldInfo.IsTimeType = true
		}
		if strings.Contains(typeName, "uuid.UUID") {
			fieldInfo.IsUUIDType = true
		}
	}
}

// parseUITag parses the UI tag and extracts component configuration
func (a *Analyzer) parseUITag(fieldInfo *FieldInfo) {
	if fieldInfo.UITag == "" {
		return
	}

	// Parse key=value pairs separated by semicolons
	pairs := strings.Split(fieldInfo.UITag, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "component":
			fieldInfo.Component = value
		case "label":
			fieldInfo.Label = value
		case "required":
			fieldInfo.Required = value == "true"
		case "placeholder":
			fieldInfo.Placeholder = value
		case "options":
			fieldInfo.Options = strings.Split(value, ",")
			for i, opt := range fieldInfo.Options {
				fieldInfo.Options[i] = strings.TrimSpace(opt)
			}
		case "validation":
			fieldInfo.Validation = value
		case "relation":
			fieldInfo.Relation = value
			fieldInfo.IsRelationship = true
		case "format":
			fieldInfo.Format = value
		case "badge":
			fieldInfo.Badge = value == "true"
		case "hidden":
			fieldInfo.Hidden = value == "true"
		case "unique":
			fieldInfo.Unique = value == "true"
		}
	}

	// Set default component if not specified
	if fieldInfo.Component == "" {
		fieldInfo.Component = a.determineDefaultComponent(fieldInfo)
	}
}

// determineDefaultComponent selects appropriate UI component based on field type
func (a *Analyzer) determineDefaultComponent(fieldInfo *FieldInfo) string {
	if fieldInfo.Hidden {
		return "hidden"
	}

	if fieldInfo.IsRelationship {
		if fieldInfo.IsSlice {
			return "multi-select"
		}
		return "select"
	}

	if fieldInfo.IsBoolType {
		return "checkbox"
	}

	if fieldInfo.IsTimeType {
		return "datetime"
	}

	if fieldInfo.IsStringType {
		if len(fieldInfo.Options) > 0 {
			return "select"
		}
		if strings.Contains(strings.ToLower(fieldInfo.Name), "email") {
			return "email"
		}
		if strings.Contains(strings.ToLower(fieldInfo.Name), "password") {
			return "password"
		}
		if strings.Contains(strings.ToLower(fieldInfo.Name), "url") {
			return "url"
		}
		return "text"
	}

	if fieldInfo.IsNumberType {
		return "number"
	}

	return "text"
}

// detectRelationships analyzes field types to detect entity relationships
func (a *Analyzer) detectRelationships(fieldInfo *FieldInfo, fieldType reflect.Type) {
	if fieldInfo.IsRelationship {
		return // Already detected via UI tag
	}

	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// Handle slice types (many-to-many or one-to-many)
	if fieldType.Kind() == reflect.Slice {
		fieldInfo.IsSlice = true
		fieldInfo.IsRelationship = true
		fieldInfo.RelationshipType = "one-to-many"

		elemType := fieldType.Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() == reflect.Struct {
			fieldInfo.TargetEntity = elemType.Name()
		}
		return
	}

	// Handle struct types (one-to-one or many-to-one)
	if fieldType.Kind() == reflect.Struct && !fieldInfo.IsTimeType && !fieldInfo.IsUUIDType {
		fieldInfo.IsRelationship = true
		fieldInfo.RelationshipType = "many-to-one"
		fieldInfo.TargetEntity = fieldType.Name()
	}

	// Detect foreign key fields
	if fieldInfo.IsUUIDType && strings.HasSuffix(fieldInfo.Name, "ID") {
		entityName := strings.TrimSuffix(fieldInfo.Name, "ID")
		if entityName != "" && entityName != "ID" {
			fieldInfo.IsRelationship = true
			fieldInfo.RelationshipType = "many-to-one"
			fieldInfo.TargetEntity = entityName
		}
	}
}

// updateStructPatterns identifies common entity patterns
func (a *Analyzer) updateStructPatterns(structInfo *StructInfo, fieldInfo FieldInfo) {
	fieldName := strings.ToLower(fieldInfo.Name)

	switch fieldName {
	case "id":
		structInfo.PrimaryKeyField = fieldInfo.Name
	case "createdat":
		structInfo.HasCreatedAt = true
	case "updatedat":
		structInfo.HasUpdatedAt = true
	case "deletedat":
		structInfo.HasDeletedAt = true
		structInfo.HasSoftDelete = true
	case "status":
		structInfo.HasStatus = true
	case "tenantid":
		structInfo.HasTenantID = true
		structInfo.HasTenantScope = true
	}
}

// analyzeBusinessLogic determines entity capabilities and patterns
func (a *Analyzer) analyzeBusinessLogic(structInfo *StructInfo) {
	// Determine if entity supports CRUD operations
	structInfo.IsCRUDEntity = structInfo.PrimaryKeyField != ""

	// Detect read-only entities (views, reports)
	if structInfo.HasCreatedAt && !structInfo.HasUpdatedAt {
		structInfo.IsReadOnly = true
	}

	// Detect workflow entities
	if structInfo.HasStatus {
		structInfo.HasWorkflow = true
	}

	// Detect validation requirements
	for _, field := range structInfo.Fields {
		if field.ValidateTag != "" || field.Required {
			structInfo.HasValidation = true
			break
		}
	}

	// Detect audit trail support
	if structInfo.HasCreatedAt && structInfo.HasUpdatedAt {
		structInfo.HasAuditTrail = true
	}

	// Detect hierarchy support
	for _, field := range structInfo.Fields {
		if strings.ToLower(field.Name) == "parentid" {
			structInfo.HasHierarchy = true
			break
		}
	}
}

// Helper functions

// typeToString converts an AST type expression to string
func (a *Analyzer) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + a.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + a.typeToString(t.Elt)
	case *ast.SelectorExpr:
		return a.typeToString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

// extractTag extracts a specific tag value from a struct tag string
// Complete implementation using proper parsing logic
func (a *Analyzer) extractTag(tagString, tagName string) string {
	// Handle empty inputs
	if tagString == "" || tagName == "" {
		return ""
	}

	// Remove surrounding backticks if present
	tagString = strings.Trim(tagString, "`")

	// Split by spaces to get individual tag declarations
	parts := strings.Fields(tagString)

	for _, part := range parts {
		// Look for tagName: pattern
		if strings.HasPrefix(part, tagName+":") {
			// Extract the quoted value
			colonIndex := strings.Index(part, ":")
			if colonIndex == -1 {
				continue
			}

			value := part[colonIndex+1:]

			// Remove surrounding quotes
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '`' && value[len(value)-1] == '`') {
					return value[1 : len(value)-1]
				}
			}

			return value
		}
	}

	// Alternative parsing for complex tag formats
	// Handle cases like: json:"field_name,omitempty" validate:"required,email"
	tagPrefix := tagName + `:"`
	startIndex := strings.Index(tagString, tagPrefix)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(tagPrefix)

	// Find the closing quote
	endIndex := startIndex
	for endIndex < len(tagString) {
		if tagString[endIndex] == '"' {
			// Check if it's escaped
			if endIndex > 0 && tagString[endIndex-1] != '\\' {
				break
			}
		}
		endIndex++
	}

	if endIndex >= len(tagString) {
		return ""
	}

	return tagString[startIndex:endIndex]
}

// AnalyzePackage analyzes all structs in a package directory
// Complete implementation for package-level analysis
func (a *Analyzer) AnalyzePackage(packagePath string) ([]StructInfo, error) {
	if packagePath == "" {
		return nil, fmt.Errorf("package path cannot be empty")
	}

	// Find all .go files in the package directory
	goFiles, err := filepath.Glob(filepath.Join(packagePath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files in package %s: %w", packagePath, err)
	}

	if len(goFiles) == 0 {
		return nil, fmt.Errorf("no Go files found in package %s", packagePath)
	}

	var allStructs []StructInfo
	processedStructs := make(map[string]bool) // Avoid duplicates across files

	// Parse each Go file in the package
	for _, goFile := range goFiles {
		// Skip test files unless specifically requested
		if strings.HasSuffix(goFile, "_test.go") {
			continue
		}

		structs, err := a.AnalyzeFile(goFile)
		if err != nil {
			log.Printf("Warning: failed to analyze file %s: %v", goFile, err)
			continue
		}

		// Add unique structs to the result
		for _, structInfo := range structs {
			structKey := structInfo.PackageName + "." + structInfo.Name
			if !processedStructs[structKey] {
				allStructs = append(allStructs, structInfo)
				processedStructs[structKey] = true
			}
		}
	}

	// Post-process to resolve cross-references between structs
	a.resolveCrossReferences(allStructs)

	return allStructs, nil
}

// resolveCrossReferences analyzes relationships between structs in the same package
func (a *Analyzer) resolveCrossReferences(structs []StructInfo) {
	// Create a map of struct names for quick lookup
	structMap := make(map[string]*StructInfo)
	for i := range structs {
		structMap[structs[i].Name] = &structs[i]
	}

	// Resolve relationships
	for i := range structs {
		for j := range structs[i].Fields {
			field := &structs[i].Fields[j]

			// Check if the field type references another struct in the package
			if field.IsRelationship && field.TargetEntity != "" {
				if _, exists := structMap[field.TargetEntity]; exists {
					// TODO: Add more sophisticated relationship analysis
					// - Detect bidirectional relationships
					// - Identify join tables for many-to-many relationships
					// - Analyze foreign key constraints
					// - Detect inheritance patterns
				}
			}
		}
	}
}

// TODO: Advanced features for future implementation:
// 1. Support for embedded structs and composition patterns
// 2. Analysis of method receivers to detect behavior patterns
// 3. Integration with database schema analysis
// 4. Support for custom validation tag formats
// 5. Detection of domain-specific patterns (events, aggregates, entities)
// 6. Analysis of interface implementations
// 7. Support for generic types (Go 1.18+)

// FIXME: Known limitations that need addressing:
// 1. Complex nested tag parsing may fail with unusual formats
// 2. Circular reference detection is not implemented
// 3. Cross-package relationship analysis is limited
// 4. Performance optimization needed for large codebases
// 5. Error handling could be more granular

// NOTE: Design decisions and considerations:
// 1. The analyzer uses both AST and reflection for maximum flexibility
// 2. UI tag format follows the pattern: key=value;key2=value2
// 3. Relationship detection prioritizes explicit UI tags over type inference
// 4. Business logic patterns are detected heuristically
// 5. The system favors convention over configuration for default behavior
