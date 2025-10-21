package css

import (
	"testing"
)

func TestSchemaBasedValidator(t *testing.T) {
	validator := NewSchemaBasedValidator()

	tests := []struct {
		dataType string
		value    string
		valid    bool
		desc     string
	}{
		// Color validation
		{"color", "red", true, "named color should be valid"},
		{"color", "#ff0000", true, "hex color should be valid"},
		{"color", "#xyz", false, "invalid hex color should be invalid"},
		{"color", "currentcolor", true, "currentcolor should be valid"},
		{"color", "", false, "empty color should be invalid"},

		// BlendMode validation
		{"blendmode", "normal", true, "normal blend mode should be valid"},
		{"blendmode", "multiply", true, "multiply blend mode should be valid"},
		{"blendmode", "invalid", false, "invalid blend mode should be invalid"},

		// Position validation
		{"bgposition", "center", true, "center position should be valid"},
		{"bgposition", "top", true, "top position should be valid"},
		{"bgposition", "50%", true, "percentage position should be valid"},
		{"bgposition", "invalid", false, "invalid position should be invalid"},

		// LineWidth validation
		{"linewidth", "thin", true, "thin line width should be valid"},
		{"linewidth", "2px", true, "pixel line width should be valid"},
		{"linewidth", "invalid", false, "invalid line width should be invalid"},

		// FontWeightAbsolute validation
		{"fontweightabsolute", "400", true, "400 font weight should be valid"},
		{"fontweightabsolute", "700", true, "700 font weight should be valid"},
		{"fontweightabsolute", "150", false, "150 font weight should be invalid"},
		{"fontweightabsolute", "abc", false, "non-numeric font weight should be invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := validator.ValidateDataType(tt.dataType, tt.value)
			isValid := (err == nil)

			if isValid != tt.valid {
				if tt.valid {
					t.Errorf("Expected %s='%s' to be valid, but got error: %v", tt.dataType, tt.value, err)
				} else {
					t.Errorf("Expected %s='%s' to be invalid, but validation passed", tt.dataType, tt.value)
				}
			}
		})
	}
}

func TestPropertyRegistrySchemaValidation(t *testing.T) {
	// Create a temporary registry for testing
	registry := NewPropertyRegistry("") // Empty schema dir for this test

	// Test that the registry has schema-based validation available
	if !registry.IsSchemaBasedValidationAvailable() {
		t.Error("Expected schema-based validation to be available")
	}

	// Test that we have DataTypes available
	supportedTypes := registry.GetSupportedDataTypes()
	if len(supportedTypes) == 0 {
		t.Error("Expected some supported DataTypes")
	}

	// Test DataType info retrieval
	info := registry.GetDataTypeInfo("color")
	if info["supported"] != true {
		t.Error("Expected color DataType to be supported")
	}
}

func TestFactorySchemaValidation(t *testing.T) {
	factory := NewFactory("") // Empty schema dir for this test

	// Test that the factory has schema-based validation available
	if !factory.IsSchemaBasedValidationAvailable() {
		t.Error("Expected schema-based validation to be available in factory")
	}

	// Test direct DataType validation
	tests := []struct {
		dataType string
		value    string
		valid    bool
	}{
		{"color", "red", true},
		{"color", "invalid", false},
		{"blendmode", "normal", true},
		{"blendmode", "invalid", false},
	}

	for _, tt := range tests {
		err := factory.ValidateDataType(tt.dataType, tt.value)
		isValid := (err == nil)

		if isValid != tt.valid {
			t.Errorf("Factory validation: expected %s='%s' valid=%v, got valid=%v",
				tt.dataType, tt.value, tt.valid, isValid)
		}
	}
}

func TestDataTypeIntegration(t *testing.T) {
	// Test that all generated DataTypes are properly integrated
	validator := NewSchemaBasedValidator()
	supportedTypes := validator.GetSupportedDataTypes()

	// We should have 49 DataTypes from the generated types
	if len(supportedTypes) != 49 {
		t.Errorf("Expected 49 supported DataTypes, got %d", len(supportedTypes))
	}

	// Test a sample of different DataType categories
	categoryTests := map[string][]string{
		"color":              {"red", "#ff0000", "currentcolor"},
		"blendmode":          {"normal", "multiply", "screen"},
		"fontweightabsolute": {"400", "700", "900"},
		"genericfamily":      {"serif", "sans-serif", "monospace"},
		"easingfunction":     {"ease", "linear", "ease-in-out"},
	}

	for dataType, validValues := range categoryTests {
		if !validator.IsDataTypeSupported(dataType) {
			continue // Skip if not supported
		}

		for _, value := range validValues {
			if err := validator.ValidateDataType(dataType, value); err != nil {
				t.Errorf("Expected %s='%s' to be valid, got error: %v", dataType, value, err)
			}
		}
	}
}

func BenchmarkSchemaValidation(b *testing.B) {
	validator := NewSchemaBasedValidator()

	b.Run("Color validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.ValidateDataType("color", "red")
		}
	})

	b.Run("BlendMode validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.ValidateDataType("blendmode", "normal")
		}
	})

	b.Run("FontWeight validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.ValidateDataType("fontweightabsolute", "400")
		}
	})
}
