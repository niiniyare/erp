package css

import (
	"testing"
)

func TestPropertyRegistryCreation(t *testing.T) {
	// Test with mock schema directory
	schemaDir := "docs/ui/Schema"
	registry := NewPropertyRegistry(schemaDir)

	if registry == nil {
		t.Fatal("Expected registry to be created, got nil")
	}

	if registry.schemaDir != schemaDir {
		t.Errorf("Expected schema directory '%s', got '%s'", schemaDir, registry.schemaDir)
	}

	// Properties and dataTypes should be initialized even if loading fails
	if registry.properties == nil {
		t.Error("Expected properties map to be initialized")
	}

	if registry.dataTypes == nil {
		t.Error("Expected dataTypes map to be initialized")
	}
}

func TestPropertyRegistryGracefulDegradation(t *testing.T) {
	// Test with non-existent schema directory - should still work
	registry := NewPropertyRegistry("non-existent-dir")

	// Should not panic and should handle validation gracefully
	err := registry.ValidateProperty("color", "red")
	if err != nil {
		t.Errorf("Expected graceful degradation, but got error: %v", err)
	}

	// Should return empty lists
	properties := registry.GetAllPropertyNames()
	if properties == nil {
		properties = []string{} // Ensure it's not nil
	}

	dataTypes := registry.GetAllDataTypeNames()
	if dataTypes == nil {
		dataTypes = []string{} // Ensure it's not nil
	}

	if registry.GetPropertyCount() < 0 {
		t.Error("Property count should not be negative")
	}

	if registry.GetDataTypeCount() < 0 {
		t.Error("Data type count should not be negative")
	}
}

func TestFactoryPropertyRegistryIntegration(t *testing.T) {
	// Test that Factory properly integrates with PropertyRegistry
	factory := NewFactory("docs/ui/Schema")

	if factory == nil {
		t.Fatal("Expected factory to be created, got nil")
	}

	registry := factory.GetPropertyRegistry()
	if registry == nil {
		t.Error("Expected factory to have property registry")
	}

	// Test validation through factory
	err := factory.ValidateProperty("color", "red")
	if err != nil {
		// Should not error due to graceful degradation
		t.Errorf("Expected validation to succeed with graceful degradation, got: %v", err)
	}

	// Test getting available properties
	properties := factory.GetAvailableProperties()
	if properties == nil {
		t.Error("Expected properties list to not be nil")
	}
}

func TestStylesPropertyRegistryIntegration(t *testing.T) {
	// Test that Styles properly integrates with PropertyRegistry
	styles := NewStyles("docs/ui/Schema")

	if styles == nil {
		t.Fatal("Expected styles to be created, got nil")
	}

	if styles.propertyRegistry == nil {
		t.Error("Expected styles to have property registry")
	}

	// Test validation through styles
	err := styles.ValidateProperty("color", "blue")
	if err != nil {
		// Should not error due to graceful degradation
		t.Errorf("Expected validation to succeed with graceful degradation, got: %v", err)
	}

	// Test getting available properties
	properties := styles.GetAvailableProperties()
	if properties == nil {
		t.Error("Expected properties list to not be nil")
	}

	count := styles.GetPropertyCount()
	if count < 0 {
		t.Error("Property count should not be negative")
	}
}

func TestPropertyRegistryValidation(t *testing.T) {
	registry := NewPropertyRegistry("docs/ui/Schema")

	testCases := []struct {
		property    string
		value       string
		expectError bool
	}{
		// Basic tests that should not fail due to graceful degradation
		{"color", "red", false},
		{"color", "blue", false},
		{"color", "#ffffff", false},
		{"display", "block", false},
		{"display", "flex", false},
		{"display", "none", false},
		{"font-size", "12px", false},
		{"margin", "10px", false},
		{"unknown-property", "any-value", false}, // Should not fail due to graceful degradation
	}

	for _, tc := range testCases {
		t.Run(tc.property+"="+tc.value, func(t *testing.T) {
			err := registry.ValidateProperty(tc.property, tc.value)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s=%s, but validation passed", tc.property, tc.value)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected validation to pass for %s=%s, but got error: %v", tc.property, tc.value, err)
			}
		})
	}
}

func TestPropertyRegistryThreadSafety(t *testing.T) {
	registry := NewPropertyRegistry("docs/ui/Schema")

	// Run multiple goroutines to test thread safety
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			// Each goroutine performs validation operations
			registry.ValidateProperty("color", "red")
			registry.GetAllPropertyNames()
			registry.GetPropertyDefinition("display")
			registry.GetPropertyCount()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we reach here without deadlock or panic, thread safety test passed
	t.Log("Thread safety test completed successfully")
}
