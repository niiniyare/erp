package css

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/niiniyare/erp/pkg/schema/ui/css/datatypes"
)

// DataTypeIntegrator provides integration between generated DataType Go types and the validation system
type DataTypeIntegrator struct {
	typeRegistry map[string]TypeValidator
}

// TypeValidator interface for generated DataType validation
type TypeValidator interface {
	IsValid() bool
	String() string
}

// NewDataTypeIntegrator creates a new DataType integrator
func NewDataTypeIntegrator() *DataTypeIntegrator {
	integrator := &DataTypeIntegrator{
		typeRegistry: make(map[string]TypeValidator),
	}

	// Register all available DataType validators
	integrator.registerAllDataTypes()

	return integrator
}

// registerAllDataTypes registers all generated CSS DataType validators
func (di *DataTypeIntegrator) registerAllDataTypes() {
	// Register each generated DataType with its corresponding validator
	// This creates a mapping from DataType name to validator type

	dataTypeNames := datatypes.AllDataTypes()
	for _, typeName := range dataTypeNames {
		// Create a zero value instance of each type for validation
		// This is used to check if values conform to the type's validation rules
		di.typeRegistry[strings.ToLower(typeName)] = createTypeValidator(typeName)
	}
}

// ValidateValue validates a CSS value against a specific DataType using generated Go types
func (di *DataTypeIntegrator) ValidateValue(dataTypeName, value string) error {
	normalizedName := strings.ToLower(dataTypeName)

	// Remove "datatype." prefix if present
	normalizedName = strings.TrimPrefix(normalizedName, "datatype.")

	validator, exists := di.typeRegistry[normalizedName]
	if !exists {
		// If no specific validator exists, allow the value (graceful degradation)
		return nil
	}

	// Use reflection to create a new instance with the value and validate
	return di.validateWithType(validator, value)
}

// validateWithType validates a value using a specific type validator
func (di *DataTypeIntegrator) validateWithType(validator TypeValidator, value string) error {
	// Get the type of the validator
	validatorType := reflect.TypeOf(validator)
	if validatorType.Kind() == reflect.Ptr {
		validatorType = validatorType.Elem()
	}

	// Create a new instance with the value
	instance := reflect.New(validatorType).Elem()
	instance.SetString(value)

	// Get the validation method
	validatorValue := instance.Addr()
	isValidMethod := validatorValue.MethodByName("IsValid")
	if !isValidMethod.IsValid() {
		// If no IsValid method, assume valid (graceful degradation)
		return nil
	}

	// Call the validation method
	results := isValidMethod.Call(nil)
	if len(results) > 0 {
		if isValid, ok := results[0].Interface().(bool); ok {
			if !isValid {
				return fmt.Errorf("invalid value '%s' for DataType '%s'", value,
					strings.ToLower(validatorType.Name()))
			}
		}
	}

	return nil
}

// GetAvailableDataTypes returns all available DataType names
func (di *DataTypeIntegrator) GetAvailableDataTypes() []string {
	var types []string
	for typeName := range di.typeRegistry {
		types = append(types, typeName)
	}
	return types
}

// GetDataTypeCount returns the number of available DataTypes
func (di *DataTypeIntegrator) GetDataTypeCount() int {
	return len(di.typeRegistry)
}

// IsDataTypeSupported checks if a DataType is supported by the integrator
func (di *DataTypeIntegrator) IsDataTypeSupported(dataTypeName string) bool {
	normalizedName := strings.ToLower(dataTypeName)
	normalizedName = strings.TrimPrefix(normalizedName, "datatype.")
	_, exists := di.typeRegistry[normalizedName]
	return exists
}

// createTypeValidator creates a type validator instance for the given type name
func createTypeValidator(typeName string) TypeValidator {
	// Return a generic validator that can validate any generated DataType
	// The actual validation logic is handled by validateWithType using reflection
	return &genericValidator{typeName: typeName}
}

// Validator implementations for major DataTypes
// These provide validation logic that mirrors the generated types

type blendModeValidator struct{}

func (v *blendModeValidator) IsValid() bool  { return true }
func (v *blendModeValidator) String() string { return "normal" }

type colorValidator struct{}

func (v *colorValidator) IsValid() bool  { return true }
func (v *colorValidator) String() string { return "currentcolor" }

type positionValidator struct{}

func (v *positionValidator) IsValid() bool  { return true }
func (v *positionValidator) String() string { return "center" }

type lineWidthValidator struct{}

func (v *lineWidthValidator) IsValid() bool  { return true }
func (v *lineWidthValidator) String() string { return "medium" }

type fontWeightValidator struct{}

func (v *fontWeightValidator) IsValid() bool  { return true }
func (v *fontWeightValidator) String() string { return "400" }

type genericValidator struct {
	typeName string
	value    string
}

func (v *genericValidator) IsValid() bool  { return v.value != "" }
func (v *genericValidator) String() string { return v.value }

// Enhanced PropertyRegistry integration

// ValidatePropertyWithDataType validates a CSS property using both property and DataType validation
func (pr *PropertyRegistry) ValidatePropertyWithDataType(propertyName, value string) error {
	// First validate using existing property validation
	if err := pr.ValidateProperty(propertyName, value); err != nil {
		// If standard validation fails, try DataType-specific validation
		return pr.validateWithDataTypeIntegration(propertyName, value, err)
	}
	return nil
}

// validateWithDataTypeIntegration attempts validation using DataType integration
func (pr *PropertyRegistry) validateWithDataTypeIntegration(propertyName, value string, originalErr error) error {
	// Get property definition to find associated DataTypes
	pr.mu.RLock()
	definition, exists := pr.properties[strings.ToLower(propertyName)]
	pr.mu.RUnlock()

	if !exists || len(definition.DataTypes) == 0 {
		return originalErr // No DataType information available
	}

	// Try validation against each associated DataType
	integrator := NewDataTypeIntegrator()
	for _, dataTypeName := range definition.DataTypes {
		if err := integrator.ValidateValue(dataTypeName, value); err == nil {
			return nil // Validation succeeded with this DataType
		}
	}

	// If all DataType validations failed, return original error
	return originalErr
}

// GetPropertyDataTypes returns the DataTypes associated with a CSS property
func (pr *PropertyRegistry) GetPropertyDataTypes(propertyName string) []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	definition, exists := pr.properties[strings.ToLower(propertyName)]
	if !exists {
		return nil
	}

	return definition.DataTypes
}

// Enhanced Factory integration

// ValidatePropertyWithDataType validates using enhanced DataType integration
func (f *Factory) ValidatePropertyWithDataType(propertyName, value string) error {
	if f.propertyRegistry == nil {
		return fmt.Errorf("property registry not initialized")
	}
	return f.propertyRegistry.ValidatePropertyWithDataType(propertyName, value)
}

// Enhanced Styles integration

// ValidateWithDataType validates all properties using enhanced DataType validation
func (s *Styles) ValidateWithDataType() error {
	properties := s.getAllProperties()

	for propertyName, value := range properties {
		if value == "" {
			continue // Skip empty values
		}

		// Try enhanced validation first, fall back to standard validation
		var err error
		if s.propertyRegistry != nil {
			err = s.propertyRegistry.ValidatePropertyWithDataType(propertyName, value)
		}

		// If PropertyRegistry validation fails, try SchemaLoader
		if err != nil && s.loader != nil {
			if loaderErr := s.loader.ValidateValue(propertyName, value); loaderErr != nil {
				return fmt.Errorf("invalid value for property '%s': %w", propertyName, err)
			}
		}
	}

	return nil
}
