package css

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/niiniyare/erp/pkg/schema/ui/css/datatypes"
)

// SchemaBasedValidator provides true schema-based validation using generated DataType Go types
type SchemaBasedValidator struct {
	generatedTypes map[string]bool
}

// NewSchemaBasedValidator creates a new schema-based validator
func NewSchemaBasedValidator() *SchemaBasedValidator {
	validator := &SchemaBasedValidator{
		generatedTypes: make(map[string]bool),
	}

	// Register all generated DataType names for quick lookup
	allTypes := datatypes.AllDataTypes()
	for _, typeName := range allTypes {
		validator.generatedTypes[strings.ToLower(typeName)] = true
	}

	return validator
}

// ValidateDataType validates a value against a specific CSS DataType using generated types
func (v *SchemaBasedValidator) ValidateDataType(dataTypeName, value string) error {
	// Normalize the datatype name
	normalizedName := strings.ToLower(dataTypeName)
	normalizedName = strings.TrimPrefix(normalizedName, "datatype.")

	// Check if we have a generated type for this DataType
	if !v.generatedTypes[normalizedName] {
		// No generated type available, allow value (graceful degradation)
		return nil
	}

	// Perform type-specific validation based on known DataTypes
	return v.validateSpecificDataType(normalizedName, value)
}

// validateSpecificDataType performs validation for specific DataTypes
func (v *SchemaBasedValidator) validateSpecificDataType(typeName, value string) error {
	switch typeName {
	case "blendmode":
		return v.validateBlendMode(value)
	case "color":
		return v.validateColor(value)
	case "bgposition":
		return v.validateBgPosition(value)
	case "bgsize":
		return v.validateBgSize(value)
	case "linewidth":
		return v.validateLineWidth(value)
	case "fontweightabsolute":
		return v.validateFontWeightAbsolute(value)
	case "genericfamily":
		return v.validateGenericFamily(value)
	case "easingfunction":
		return v.validateEasingFunction(value)
	case "position":
		return v.validatePosition(value)
	case "displayinside":
		return v.validateDisplayInside(value)
	case "displayoutside":
		return v.validateDisplayOutside(value)
	default:
		// For DataTypes without specific validation, validate as non-empty string
		return v.validateGenericString(value)
	}
}

// Specific validation methods for each major CSS DataType

func (v *SchemaBasedValidator) validateBlendMode(value string) error {
	validValues := []string{
		"normal", "multiply", "screen", "overlay", "darken", "lighten",
		"color-dodge", "color-burn", "hard-light", "soft-light",
		"difference", "exclusion", "hue", "saturation", "color", "luminosity",
	}
	return v.validateEnum(value, validValues, "BlendMode")
}

func (v *SchemaBasedValidator) validateColor(value string) error {
	// CSS Color validation - accept common formats
	value = strings.TrimSpace(strings.ToLower(value))
	
	// Named colors
	namedColors := []string{
		"currentcolor", "transparent", "initial", "inherit", "unset",
		"black", "white", "red", "green", "blue", "yellow", "cyan", "magenta",
		"gray", "grey", "silver", "maroon", "navy", "olive", "lime", "aqua",
		"teal", "purple", "fuchsia", "orange", "brown", "pink", "gold",
	}
	
	for _, color := range namedColors {
		if value == color {
			return nil
		}
	}
	
	// Hex colors
	if strings.HasPrefix(value, "#") {
		hex := value[1:]
		if len(hex) == 3 || len(hex) == 6 || len(hex) == 8 {
			// Validate hex characters
			for _, c := range hex {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					return fmt.Errorf("invalid hex color: %s", value)
				}
			}
			return nil
		}
	}
	
	// RGB/RGBA function
	if strings.HasPrefix(value, "rgb(") || strings.HasPrefix(value, "rgba(") {
		return nil // Assume valid for now - full RGB validation would be complex
	}
	
	// HSL/HSLA function
	if strings.HasPrefix(value, "hsl(") || strings.HasPrefix(value, "hsla(") {
		return nil // Assume valid for now - full HSL validation would be complex
	}
	
	return fmt.Errorf("invalid color value: %s", value)
}

func (v *SchemaBasedValidator) validateBgPosition(value string) error {
	validKeywords := []string{"top", "right", "bottom", "left", "center"}
	
	// Check for keyword values
	for _, keyword := range validKeywords {
		if value == keyword {
			return nil
		}
	}
	
	// Check for percentage or length values
	if strings.Contains(value, "%") || strings.Contains(value, "px") || 
	   strings.Contains(value, "em") || strings.Contains(value, "rem") {
		return nil
	}
	
	// Check for combined values like "top left", "50% 25%"
	parts := strings.Fields(value)
	if len(parts) == 2 {
		// Assume valid two-part position
		return nil
	}
	
	return fmt.Errorf("invalid background position: %s", value)
}

func (v *SchemaBasedValidator) validateBgSize(value string) error {
	validValues := []string{"auto", "cover", "contain"}
	return v.validateEnumOrLength(value, validValues, "background-size")
}

func (v *SchemaBasedValidator) validateLineWidth(value string) error {
	validValues := []string{"thin", "medium", "thick"}
	return v.validateEnumOrLength(value, validValues, "line width")
}

func (v *SchemaBasedValidator) validateFontWeightAbsolute(value string) error {
	validValues := []string{"100", "200", "300", "400", "500", "600", "700", "800", "900"}
	
	// Check for numeric string values
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}
	
	// Check for valid integer
	if weight, err := strconv.Atoi(value); err == nil {
		if weight >= 100 && weight <= 900 && weight%100 == 0 {
			return nil
		}
	}
	
	return fmt.Errorf("invalid font weight: %s (must be 100-900 in increments of 100)", value)
}

func (v *SchemaBasedValidator) validateGenericFamily(value string) error {
	validValues := []string{"serif", "sans-serif", "monospace", "cursive", "fantasy"}
	return v.validateEnum(value, validValues, "generic font family")
}

func (v *SchemaBasedValidator) validateEasingFunction(value string) error {
	validValues := []string{
		"linear", "ease", "ease-in", "ease-out", "ease-in-out", 
		"step-start", "step-end",
	}
	
	// Check for predefined easing functions
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}
	
	// Check for cubic-bezier function
	if strings.HasPrefix(value, "cubic-bezier(") && strings.HasSuffix(value, ")") {
		return nil // Assume valid for now
	}
	
	// Check for steps function
	if strings.HasPrefix(value, "steps(") && strings.HasSuffix(value, ")") {
		return nil // Assume valid for now
	}
	
	return fmt.Errorf("invalid easing function: %s", value)
}

func (v *SchemaBasedValidator) validatePosition(value string) error {
	validValues := []string{"static", "relative", "absolute", "fixed", "sticky"}
	return v.validateEnum(value, validValues, "position")
}

func (v *SchemaBasedValidator) validateDisplayInside(value string) error {
	validValues := []string{"flow", "flow-root", "table", "flex", "grid", "ruby"}
	return v.validateEnum(value, validValues, "display inside")
}

func (v *SchemaBasedValidator) validateDisplayOutside(value string) error {
	validValues := []string{"block", "inline", "run-in"}
	return v.validateEnum(value, validValues, "display outside")
}

// Helper validation methods

func (v *SchemaBasedValidator) validateEnum(value string, validValues []string, typeName string) error {
	value = strings.TrimSpace(strings.ToLower(value))
	
	for _, validValue := range validValues {
		if value == strings.ToLower(validValue) {
			return nil
		}
	}
	
	return fmt.Errorf("invalid %s value: %s (valid values: %s)", 
		typeName, value, strings.Join(validValues, ", "))
}

func (v *SchemaBasedValidator) validateEnumOrLength(value string, validValues []string, typeName string) error {
	value = strings.TrimSpace(strings.ToLower(value))
	
	// Check enum values first
	for _, validValue := range validValues {
		if value == strings.ToLower(validValue) {
			return nil
		}
	}
	
	// Check for length/percentage values
	if strings.Contains(value, "px") || strings.Contains(value, "em") || 
	   strings.Contains(value, "rem") || strings.Contains(value, "%") ||
	   strings.Contains(value, "vw") || strings.Contains(value, "vh") {
		return nil
	}
	
	// Check for numeric value (assume pixels)
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return nil
	}
	
	return fmt.Errorf("invalid %s value: %s", typeName, value)
}

func (v *SchemaBasedValidator) validateGenericString(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// GetSupportedDataTypes returns all DataTypes that have schema-based validation
func (v *SchemaBasedValidator) GetSupportedDataTypes() []string {
	types := make([]string, 0, len(v.generatedTypes))
	for typeName := range v.generatedTypes {
		types = append(types, typeName)
	}
	return types
}

// IsDataTypeSupported checks if a DataType is supported for schema validation
func (v *SchemaBasedValidator) IsDataTypeSupported(dataTypeName string) bool {
	normalizedName := strings.ToLower(dataTypeName)
	normalizedName = strings.TrimPrefix(normalizedName, "datatype.")
	return v.generatedTypes[normalizedName]
}

// GetDataTypeInfo returns information about a specific DataType
func (v *SchemaBasedValidator) GetDataTypeInfo(dataTypeName string) map[string]any {
	normalizedName := strings.ToLower(dataTypeName)
	normalizedName = strings.TrimPrefix(normalizedName, "datatype.")

	info := map[string]any{
		"name":      dataTypeName,
		"supported": v.generatedTypes[normalizedName],
		"generated": true,
	}

	// Add specific information for known types
	switch normalizedName {
	case "color":
		info["description"] = "CSS color values (hex, rgb, hsl, named colors)"
		info["examples"] = []string{"#ff0000", "rgb(255,0,0)", "red", "currentcolor"}
	case "blendmode":
		info["description"] = "CSS blend mode values"
		info["examples"] = []string{"normal", "multiply", "screen", "overlay"}
	case "position":
		info["description"] = "CSS position values"
		info["examples"] = []string{"static", "relative", "absolute", "fixed"}
	default:
		info["description"] = fmt.Sprintf("CSS DataType: %s", dataTypeName)
		info["examples"] = []string{}
	}

	return info
}