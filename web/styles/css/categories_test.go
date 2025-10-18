package css

import (
	"strings"
	"testing"
)

func TestCategoryManager(t *testing.T) {
	cm := NewCategoryManager()

	// Test basic functionality
	t.Run("Initialization", func(t *testing.T) {
		categories := cm.GetAllCategories()
		if len(categories) == 0 {
			t.Error("Expected some categories to be initialized")
		}

		categoryNames := cm.GetCategoryNames()
		if len(categoryNames) != len(categories) {
			t.Error("Category names count doesn't match categories count")
		}

		// Check that categories are sorted by priority
		for i := 1; i < len(categories); i++ {
			if categories[i-1].Priority > categories[i].Priority {
				t.Error("Categories are not sorted by priority")
				break
			}
		}
	})

	t.Run("Property Categorization", func(t *testing.T) {
		tests := map[string]string{
			"color":              "Colors",
			"background-color":   "Background",
			"margin":             "Spacing",
			"padding":            "Spacing",
			"display":            "Layout",
			"flex-direction":     "Flexbox",
			"grid-template":      "Grid",
			"font-size":          "Typography",
			"border":             "Borders",
			"transform":          "Transforms",
			"animation":          "Animations",
			"transition":         "Transitions",
			"box-shadow":         "Visual Effects",
			"--custom-property":  "Custom Properties",
			"-webkit-transform":  "Experimental",
			"unknown-property":   "Other",
		}

		for property, expectedCategory := range tests {
			category := cm.GetPropertyCategory(property)
			if category != expectedCategory {
				t.Errorf("Property %s: expected category %s, got %s", property, expectedCategory, category)
			}
		}
	})

	t.Run("Category Retrieval", func(t *testing.T) {
		layoutCategory, exists := cm.GetCategory("Layout")
		if !exists {
			t.Error("Layout category should exist")
		}

		if layoutCategory.Name != "Layout" {
			t.Error("Layout category name mismatch")
		}

		if len(layoutCategory.Properties) == 0 {
			t.Error("Layout category should have properties")
		}
	})

	t.Run("Properties by Category", func(t *testing.T) {
		layoutProps := cm.GetPropertiesByCategory("Layout")
		if len(layoutProps) == 0 {
			t.Error("Layout category should have properties")
		}

		// Check that properties are sorted
		for i := 1; i < len(layoutProps); i++ {
			if layoutProps[i-1] > layoutProps[i] {
				t.Error("Properties in category are not sorted")
				break
			}
		}

		// Test non-existent category
		nonExistent := cm.GetPropertiesByCategory("NonExistent")
		if nonExistent != nil {
			t.Error("Non-existent category should return nil")
		}
	})

	t.Run("Property Search", func(t *testing.T) {
		results := cm.SearchProperties("margin")
		if len(results) == 0 {
			t.Error("Search for 'margin' should return results")
		}

		// Should find margin in Spacing category
		spacingResults, exists := results["Spacing"]
		if !exists {
			t.Error("Search should find margin properties in Spacing category")
		}

		found := false
		for _, prop := range spacingResults {
			if strings.Contains(prop, "margin") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Should find margin-related properties")
		}
	})

	t.Run("Related Properties", func(t *testing.T) {
		related := cm.GetRelatedProperties("margin")
		if len(related) == 0 {
			t.Error("margin should have related properties")
		}

		// Should not include the original property
		for _, prop := range related {
			if prop == "margin" {
				t.Error("Related properties should not include the original property")
			}
		}
	})

	t.Run("Categorized Properties", func(t *testing.T) {
		categorized := cm.GetCategorizedProperties()
		if len(categorized) == 0 {
			t.Error("Should return categorized properties")
		}

		// Check that each category has properties
		for categoryName, properties := range categorized {
			if len(properties) == 0 {
				t.Errorf("Category %s should have properties", categoryName)
			}
		}
	})

	t.Run("Category Stats", func(t *testing.T) {
		stats := cm.GetCategoryStats()
		if len(stats) == 0 {
			t.Error("Should return category statistics")
		}

		for name, stat := range stats {
			if stat.Name != name {
				t.Errorf("Stat name mismatch for category %s", name)
			}
			if stat.PropertyCount <= 0 {
				t.Errorf("Category %s should have positive property count", name)
			}
		}
	})
}

func TestCategoryCustomization(t *testing.T) {
	cm := NewCategoryManager()

	t.Run("Add Custom Category", func(t *testing.T) {
		customCategory := &PropertyCategory{
			Name:        "Custom",
			Description: "Custom properties for testing",
			Properties:  []string{"custom-prop-1", "custom-prop-2"},
			Priority:    99,
		}

		cm.AddCustomCategory(customCategory)

		// Check that the category was added
		category, exists := cm.GetCategory("Custom")
		if !exists {
			t.Error("Custom category should be added")
		}

		if category.Name != "Custom" {
			t.Error("Custom category name should match")
		}

		// Check that properties are mapped
		if cm.GetPropertyCategory("custom-prop-1") != "Custom" {
			t.Error("Custom property should be in Custom category")
		}
	})

	t.Run("Add Property to Category", func(t *testing.T) {
		success := cm.AddPropertyToCategory("new-layout-prop", "Layout")
		if !success {
			t.Error("Should be able to add property to existing category")
		}

		if cm.GetPropertyCategory("new-layout-prop") != "Layout" {
			t.Error("New property should be in Layout category")
		}

		// Test adding to non-existent category
		success = cm.AddPropertyToCategory("some-prop", "NonExistent")
		if success {
			t.Error("Should not be able to add property to non-existent category")
		}
	})
}

func TestPropertyRegistryCategorization(t *testing.T) {
	registry := NewPropertyRegistry("") // Empty schema dir for testing

	t.Run("Category Integration", func(t *testing.T) {
		categories := registry.GetAllCategories()
		if len(categories) == 0 {
			t.Error("PropertyRegistry should have categories")
		}

		categoryNames := registry.GetCategoryNames()
		if len(categoryNames) == 0 {
			t.Error("PropertyRegistry should have category names")
		}
	})

	t.Run("Property Category Lookup", func(t *testing.T) {
		category := registry.GetPropertyCategory("display")
		if category != "Layout" {
			t.Errorf("Expected display to be in Layout category, got %s", category)
		}

		category = registry.GetPropertyCategory("color")
		if category != "Colors" {
			t.Errorf("Expected color to be in Colors category, got %s", category)
		}
	})

	t.Run("Category Properties", func(t *testing.T) {
		layoutProps := registry.GetPropertiesByCategory("Layout")
		if len(layoutProps) == 0 {
			t.Error("Layout category should have properties")
		}

		// Should include common layout properties
		expectedProps := []string{"display", "position", "width", "height"}
		for _, expected := range expectedProps {
			found := false
			for _, prop := range layoutProps {
				if prop == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Layout category should include property: %s", expected)
			}
		}
	})

	t.Run("Property Search", func(t *testing.T) {
		results := registry.SearchProperties("margin")
		if len(results) == 0 {
			t.Error("Should find margin properties")
		}
	})

	t.Run("Related Properties", func(t *testing.T) {
		related := registry.GetRelatedProperties("margin")
		if len(related) == 0 {
			t.Error("margin should have related properties")
		}
	})

	t.Run("Category Statistics", func(t *testing.T) {
		stats := registry.GetCategoryStats()
		if len(stats) == 0 {
			t.Error("Should have category statistics")
		}
	})
}

func TestFactoryCategorization(t *testing.T) {
	factory := NewFactory("") // Empty schema dir for testing

	t.Run("Factory Category Integration", func(t *testing.T) {
		categories := factory.GetAllCategories()
		if len(categories) == 0 {
			t.Error("Factory should have categories")
		}

		categoryNames := factory.GetCategoryNames()
		if len(categoryNames) == 0 {
			t.Error("Factory should have category names")
		}
	})

	t.Run("Factory Property Categorization", func(t *testing.T) {
		category := factory.GetPropertyCategory("background-color")
		if category != "Background" {
			t.Errorf("Expected background-color to be in Background category, got %s", category)
		}
	})

	t.Run("Factory Category Methods", func(t *testing.T) {
		backgroundProps := factory.GetPropertiesByCategory("Background")
		if len(backgroundProps) == 0 {
			t.Error("Background category should have properties")
		}

		searchResults := factory.SearchProperties("font")
		if len(searchResults) == 0 {
			t.Error("Should find font properties")
		}

		stats := factory.GetCategoryStats()
		if len(stats) == 0 {
			t.Error("Should have category statistics")
		}
	})
}

func TestCategoryValidation(t *testing.T) {
	cm := NewCategoryManager()

	t.Run("Category Structure Validation", func(t *testing.T) {
		issues := cm.ValidateCategoryStructure()
		if len(issues) > 0 {
			for _, issue := range issues {
				t.Errorf("Category structure issue: %s", issue)
			}
		}
	})

	t.Run("Export Categories JSON", func(t *testing.T) {
		exported := cm.ExportCategoriesJSON()
		if exported == nil {
			t.Error("Should export categories as JSON")
		}

		if _, exists := exported["categories"]; !exists {
			t.Error("Exported data should include categories")
		}

		if _, exists := exported["stats"]; !exists {
			t.Error("Exported data should include stats")
		}
	})
}

func BenchmarkCategorization(b *testing.B) {
	cm := NewCategoryManager()

	b.Run("Property Categorization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cm.GetPropertyCategory("margin")
		}
	})

	b.Run("Category Lookup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cm.GetCategory("Layout")
		}
	})

	b.Run("Property Search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cm.SearchProperties("margin")
		}
	})
}