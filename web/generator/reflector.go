package generator

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// TypeRegistry maintains a registry of types for reflection-based analysis
type TypeRegistry struct {
	types    map[string]reflect.Type
	packages map[string][]reflect.Type
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types:    make(map[string]reflect.Type),
		packages: make(map[string][]reflect.Type),
	}
}

// RegisterType adds a type to the registry
func (tr *TypeRegistry) RegisterType(t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	name := t.Name()
	pkgPath := t.PkgPath()

	tr.types[name] = t
	tr.packages[pkgPath] = append(tr.packages[pkgPath], t)
}

// GetType retrieves a type by name
func (tr *TypeRegistry) GetType(name string) (reflect.Type, bool) {
	t, exists := tr.types[name]
	return t, exists
}

// GetPackageTypes returns all types in a package
func (tr *TypeRegistry) GetPackageTypes(packagePath string) []reflect.Type {
	return tr.packages[packagePath]
}

// Reflector provides runtime type inspection capabilities
type Reflector struct {
	analyzer *Analyzer
	registry *TypeRegistry
}

// NewReflector creates a new reflector with type registry
func NewReflector() *Reflector {
	return &Reflector{
		analyzer: NewAnalyzer(),
		registry: NewTypeRegistry(),
	}
}

// ReflectStruct performs comprehensive reflection analysis on a struct type
func (r *Reflector) ReflectStruct(v interface{}) (StructInfo, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return StructInfo{}, fmt.Errorf("cannot reflect nil interface")
	}

	// Register the type for future reference
	r.registry.RegisterType(t)

	// Use analyzer to get detailed struct information
	structInfo := r.analyzer.AnalyzeStruct(t)

	// Enhance with runtime reflection data
	r.enhanceWithReflection(&structInfo, t)

	return structInfo, nil
}

// ReflectValue analyzes a struct value and extracts current field values
func (r *Reflector) ReflectValue(v interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value is not a struct")
	}

	typ := val.Type()
	result := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle different field types
		value := r.extractFieldValue(fieldValue)
		result[field.Name] = value
	}

	return result, nil
}

// GetMethodInfo extracts method information from a type
func (r *Reflector) GetMethodInfo(t reflect.Type) []MethodInfo {
	var methods []MethodInfo

	// Get methods on the type
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodInfo := MethodInfo{
			Name:       method.Name,
			IsExported: method.IsExported(),
			Type:       method.Type.String(),
		}

		// Analyze method signature
		r.analyzeMethodSignature(&methodInfo, method.Type)

		methods = append(methods, methodInfo)
	}

	// Get methods on pointer to the type
	if t.Kind() != reflect.Ptr {
		ptrType := reflect.PtrTo(t)
		for i := 0; i < ptrType.NumMethod(); i++ {
			method := ptrType.Method(i)
			if !r.containsMethod(methods, method.Name) {
				methodInfo := MethodInfo{
					Name:        method.Name,
					IsExported:  method.IsExported(),
					Type:        method.Type.String(),
					RequiresPtr: true,
				}

				r.analyzeMethodSignature(&methodInfo, method.Type)
				methods = append(methods, methodInfo)
			}
		}
	}

	return methods
}

// GetInterfaceInfo analyzes if a type implements specific interfaces
func (r *Reflector) GetInterfaceInfo(t reflect.Type) []InterfaceInfo {
	var interfaces []InterfaceInfo

	// Common interfaces to check
	commonInterfaces := map[string]reflect.Type{
		"fmt.Stringer":     reflect.TypeOf((*fmt.Stringer)(nil)).Elem(),
		"error":            reflect.TypeOf((*error)(nil)).Elem(),
		"json.Marshaler":   getJSONMarshalerType(),
		"json.Unmarshaler": getJSONUnmarshalerType(),
		"sql.Scanner":      getSQLScannerType(),
		"driver.Valuer":    getDriverValuerType(),
	}

	for name, interfaceType := range commonInterfaces {
		if interfaceType != nil && t.Implements(interfaceType) {
			interfaces = append(interfaces, InterfaceInfo{
				Name:        name,
				Implemented: true,
			})
		}
	}

	return interfaces
}

// GetEmbeddedTypes finds all embedded types in a struct
func (r *Reflector) GetEmbeddedTypes(t reflect.Type) []EmbeddedTypeInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	var embedded []EmbeddedTypeInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous {
			embeddedInfo := EmbeddedTypeInfo{
				Type:      field.Type,
				Name:      field.Type.Name(),
				Package:   field.Type.PkgPath(),
				IsPointer: field.Type.Kind() == reflect.Ptr,
			}

			embedded = append(embedded, embeddedInfo)
		}
	}

	return embedded
}

// InspectMemoryLayout provides information about struct memory layout
func (r *Reflector) InspectMemoryLayout(t reflect.Type) MemoryLayoutInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	layout := MemoryLayoutInfo{
		Size:      int(t.Size()),
		Alignment: int(t.Align()),
		Fields:    make([]FieldLayoutInfo, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		fieldLayout := FieldLayoutInfo{
			Name:   field.Name,
			Offset: int(field.Offset),
			Size:   int(field.Type.Size()),
			Type:   field.Type.String(),
		}

		layout.Fields = append(layout.Fields, fieldLayout)
	}

	return layout
}

// enhanceWithReflection adds runtime reflection data to struct info
func (r *Reflector) enhanceWithReflection(structInfo *StructInfo, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Add method information
	// TODO: Integrate method analysis into schema generation
	// This would help determine if the struct has CRUD methods,
	// validation methods, or other business logic

	// Add interface implementation info
	// TODO: Use interface information to determine UI capabilities
	// For example, if a struct implements json.Marshaler,
	// it might need special handling in forms

	// Add embedded type information
	// TODO: Handle embedded fields in UI generation
	// Embedded types might need to be flattened or grouped in the UI

	// Memory layout analysis
	// NOTE: Memory layout is primarily for debugging and optimization
	// but could be useful for detecting struct packing issues
}

// extractFieldValue safely extracts a field value
func (r *Reflector) extractFieldValue(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Handle interfaces
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Slice, reflect.Array:
		return r.extractSliceValue(v)
	case reflect.Map:
		return r.extractMapValue(v)
	case reflect.Struct:
		// Special handling for common types
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time)
		}
		// For other structs, recursively extract
		return r.extractStructValue(v)
	default:
		// For complex types, return the interface
		if v.CanInterface() {
			return v.Interface()
		}
		return v.String() // Fallback
	}
}

// extractSliceValue extracts slice/array values
func (r *Reflector) extractSliceValue(v reflect.Value) []interface{} {
	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = r.extractFieldValue(v.Index(i))
	}
	return result
}

// extractMapValue extracts map values
func (r *Reflector) extractMapValue(v reflect.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range v.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		result[keyStr] = r.extractFieldValue(v.MapIndex(key))
	}
	return result
}

// extractStructValue recursively extracts struct field values
func (r *Reflector) extractStructValue(v reflect.Value) map[string]interface{} {
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldValue := v.Field(i)
			result[field.Name] = r.extractFieldValue(fieldValue)
		}
	}

	return result
}

// analyzeMethodSignature analyzes a method's signature
func (r *Reflector) analyzeMethodSignature(methodInfo *MethodInfo, methodType reflect.Type) {
	// Analyze input parameters
	for i := 1; i < methodType.NumIn(); i++ { // Skip receiver
		param := methodType.In(i)
		methodInfo.InputTypes = append(methodInfo.InputTypes, param.String())
	}

	// Analyze return types
	for i := 0; i < methodType.NumOut(); i++ {
		returnType := methodType.Out(i)
		methodInfo.OutputTypes = append(methodInfo.OutputTypes, returnType.String())

		// Check if returns error
		if returnType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			methodInfo.ReturnsError = true
		}
	}

	// Determine method purpose based on naming conventions
	methodName := strings.ToLower(methodInfo.Name)
	switch {
	case strings.HasPrefix(methodName, "get") || strings.HasPrefix(methodName, "find"):
		methodInfo.Purpose = "getter"
	case strings.HasPrefix(methodName, "set") || strings.HasPrefix(methodName, "update"):
		methodInfo.Purpose = "setter"
	case strings.HasPrefix(methodName, "create") || strings.HasPrefix(methodName, "new"):
		methodInfo.Purpose = "creator"
	case strings.HasPrefix(methodName, "delete") || strings.HasPrefix(methodName, "remove"):
		methodInfo.Purpose = "deleter"
	case strings.HasPrefix(methodName, "validate"):
		methodInfo.Purpose = "validator"
	case strings.HasPrefix(methodName, "is") || strings.HasPrefix(methodName, "has"):
		methodInfo.Purpose = "predicate"
	default:
		methodInfo.Purpose = "business_logic"
	}
}

// containsMethod checks if a method with the given name already exists
func (r *Reflector) containsMethod(methods []MethodInfo, name string) bool {
	for _, method := range methods {
		if method.Name == name {
			return true
		}
	}
	return false
}

// Helper functions for getting interface types safely

// getJSONMarshalerType returns the json.Marshaler interface type
func getJSONMarshalerType() reflect.Type {
	// TODO: This requires importing encoding/json
	// For now, return nil to avoid import issues
	return nil
	// In a complete implementation:
	// return reflect.TypeOf((*json.Marshaler)(nil)).Elem()
}

// getJSONUnmarshalerType returns the json.Unmarshaler interface type
func getJSONUnmarshalerType() reflect.Type {
	// TODO: Similar to above
	return nil
}

// getSQLScannerType returns the sql.Scanner interface type
func getSQLScannerType() reflect.Type {
	// TODO: This requires importing database/sql
	return nil
}

// getDriverValuerType returns the driver.Valuer interface type
func getDriverValuerType() reflect.Type {
	// TODO: This requires importing database/sql/driver
	return nil
}

// Supporting types

// MethodInfo contains information about a struct method
type MethodInfo struct {
	Name         string
	IsExported   bool
	Type         string
	RequiresPtr  bool
	InputTypes   []string
	OutputTypes  []string
	ReturnsError bool
	Purpose      string // getter, setter, validator, etc.
}

// InterfaceInfo contains information about interface implementations
type InterfaceInfo struct {
	Name        string
	Implemented bool
}

// EmbeddedTypeInfo contains information about embedded types
type EmbeddedTypeInfo struct {
	Type      reflect.Type
	Name      string
	Package   string
	IsPointer bool
}

// MemoryLayoutInfo contains struct memory layout information
type MemoryLayoutInfo struct {
	Size      int
	Alignment int
	Fields    []FieldLayoutInfo
}

// FieldLayoutInfo contains field memory layout information
type FieldLayoutInfo struct {
	Name   string
	Offset int
	Size   int
	Type   string
}

// Advanced reflection utilities

// GetTypeByName attempts to find a type by its full name
func (r *Reflector) GetTypeByName(fullName string) (reflect.Type, error) {
	// TODO: Implement type lookup by name
	// This would require maintaining a global registry of types
	// or using runtime symbol table inspection
	return nil, fmt.Errorf("type lookup by name not implemented")
}

// GetAllTypes returns all registered types
func (r *Reflector) GetAllTypes() map[string]reflect.Type {
	return r.registry.types
}

// GetDependencies analyzes type dependencies
func (r *Reflector) GetDependencies(t reflect.Type) []reflect.Type {
	// TODO: Implement dependency analysis
	// This would traverse all field types and find dependencies
	return nil
}

// FIXME: Known limitations and issues:
// 1. Interface type detection requires proper imports
// 2. Type lookup by name needs global registry
// 3. Circular dependency detection not implemented
// 4. Performance may degrade with very large structs
// 5. Memory layout inspection may not work on all platforms

// TODO: Future enhancements:
// 1. Add support for generic types (Go 1.18+)
// 2. Implement type constraint analysis
// 3. Add support for function types and closures
// 4. Integrate with AST analysis for complete type information
// 5. Add caching mechanisms for expensive reflection operations
// 6. Support for custom type annotations via build tags

// NOTE: Design considerations:
// 1. The reflector complements the analyzer by providing runtime type information
// 2. Memory layout inspection is primarily for debugging and optimization
// 3. Method analysis helps determine entity capabilities for UI generation
// 4. Interface detection enables polymorphic UI behavior
// 5. Type registry enables cross-reference resolution
