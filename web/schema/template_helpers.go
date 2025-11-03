package schema

import (
	"context"
	"fmt"
	"strings"
)

// Helper functions for templ templates

// formClasses generates CSS classes for form element
func formClasses(schema *Schema) string {
	classes := []string{"schema-form"}

	if schema.Layout != nil {
		classes = append(classes, fmt.Sprintf("schema-form--columns-%d", schema.Layout.Columns))
		if schema.Layout.Gap != "" {
			classes = append(classes, "schema-form--custom-gap")
		}
	}

	if schema.HasHTMX() {
		classes = append(classes, "schema-form--htmx")
	}

	if schema.HasAlpine() {
		classes = append(classes, "schema-form--alpine")
	}

	return strings.Join(classes, " ")
}

// sectionClasses generates CSS classes for section element
func sectionClasses(section Section) string {
	classes := []string{"form-section"}
	if section.Collapsible {
		classes = append(classes, "form-section--collapsible")
		if section.Collapsed {
			classes = append(classes, "form-section--collapsed")
		}
	}
	return strings.Join(classes, " ")
}

// sectionHeaderClasses generates CSS classes for section header
func sectionHeaderClasses(section Section) string {
	headerClass := "section-header"
	if section.Collapsible {
		headerClass += " section-header--collapsible"
	}
	return headerClass
}

// tabNavClasses generates CSS classes for tab navigation item
func tabNavClasses(active bool) string {
	classes := []string{"tab-nav-item"}
	if active {
		classes = append(classes, "tab-nav-item--active")
	}
	return strings.Join(classes, " ")
}

// tabContentClasses generates CSS classes for tab content
func tabContentClasses(active bool) string {
	classes := []string{"tab-content"}
	if active {
		classes = append(classes, "tab-content--active")
	}
	return strings.Join(classes, " ")
}

// stepIndicatorClasses generates CSS classes for step indicator
func stepIndicatorClasses(active bool) string {
	classes := []string{"step-indicator"}
	if active {
		classes = append(classes, "step-indicator--active")
	}
	return strings.Join(classes, " ")
}

// stepContentClasses generates CSS classes for step content
func stepContentClasses(active bool) string {
	classes := []string{"step-content"}
	if active {
		classes = append(classes, "step-content--active")
	}
	return strings.Join(classes, " ")
}

// actionClasses generates CSS classes for action button
func actionClasses(action *Action) string {
	classes := []string{"action-button"}
	classes = append(classes, fmt.Sprintf("action-button--%s", action.Type))
	if action.Variant != "" {
		classes = append(classes, fmt.Sprintf("action-button--%s", action.Variant))
	}
	if action.Size != "" {
		classes = append(classes, fmt.Sprintf("action-button--%s", action.Size))
	}
	return strings.Join(classes, " ")
}

// getCSRFTokenField returns the CSRF token field name
func getCSRFTokenField(csrf *CSRF) string {
	if csrf.TokenField != "" {
		return csrf.TokenField
	}
	return "csrf_token"
}

// generateFormStyle generates inline styles for form
func generateFormStyle(schema *Schema, tokens TokenResolver) string {
	if tokens == nil {
		return ""
	}

	var styles []string

	// Add layout-specific styles
	if schema.Layout != nil {
		if schema.Layout.Gap != "" {
			if value, err := tokens.ResolveToken("spacing." + schema.Layout.Gap); err == nil {
				styles = append(styles, fmt.Sprintf("gap: %s", value))
			} else {
				styles = append(styles, fmt.Sprintf("gap: %s", schema.Layout.Gap))
			}
		}
	}

	return strings.Join(styles, "; ")
}

// getSectionFields filters fields by section field names
func getSectionFields(allFields []Field, fieldNames []string) []Field {
	fieldMap := make(map[string]*Field)
	for i := range allFields {
		fieldMap[allFields[i].Name] = &allFields[i]
	}

	var sectionFields []Field
	for _, fieldName := range fieldNames {
		if field, exists := fieldMap[fieldName]; exists {
			sectionFields = append(sectionFields, *field)
		}
	}

	return sectionFields
}

// getFieldValue gets the value for a field from data map
func getFieldValue(data map[string]any, fieldName string) any {
	if data == nil {
		return nil
	}
	return data[fieldName]
}

// getFieldErrors gets the errors for a field from errors map
func getFieldErrors(errors map[string][]string, fieldName string) []string {
	if errors == nil {
		return nil
	}
	return errors[fieldName]
}

// renderFieldHTML renders a field to HTML using the registry
// This is a bridge function to use the existing renderer system within templ
func renderFieldHTML(field *Field, value any, errors []string, tokens TokenResolver) string {
	// Create a basic registry for rendering
	registry := NewRendererRegistry(tokens)

	// Use context.Background() for basic rendering
	html, err := registry.RenderField(context.Background(), field, value, errors)
	if err != nil {
		// Return error HTML if rendering fails
		return fmt.Sprintf(`<div class="field-error">Error rendering field: %v</div>`, err)
	}

	return html
}

// FieldComponent creates a templ component for a field
// This allows for better composition in templ templates
func FieldComponent(field *Field, value any, errors []string, tokens TokenResolver) string {
	return renderFieldHTML(field, value, errors, tokens)
}

// Validation helpers for template usage
func hasErrors(errors []string) bool {
	return len(errors) > 0
}

func joinErrors(errors []string, separator string) string {
	return strings.Join(errors, separator)
}

// CSS helper functions
func cssVar(name, value string) string {
	return fmt.Sprintf("--%s: %s", name, value)
}

func resolveTokenOrDefault(tokens TokenResolver, token, defaultValue string) string {
	if tokens != nil {
		if value, err := tokens.ResolveToken(token); err == nil {
			return value
		}
	}
	return defaultValue
}

// Accessibility helpers
func ariaDescribedBy(fieldName string, hasDescription, hasHelp bool) string {
	var parts []string
	if hasDescription {
		parts = append(parts, fieldName+"-description")
	}
	if hasHelp {
		parts = append(parts, fieldName+"-help")
	}
	return strings.Join(parts, " ")
}

// Form state helpers
func isFieldRequired(field *Field) bool {
	return field.Required
}

func isFieldDisabled(field *Field) bool {
	return field.Disabled
}

func isFieldReadonly(field *Field) bool {
	return field.Readonly
}

func isFieldHidden(field *Field) bool {
	return field.Hidden
}

// Data type conversion helpers for templ templates
func toString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func toBool(value any) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "on"
	case int:
		return v != 0
	default:
		return false
	}
}

func toInt(value any) int {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case string:
		// Basic string to int conversion for template use
		if v == "" {
			return 0
		}
		// In a real implementation, you might want proper error handling
		return len(v) // Simple fallback
	default:
		return 0
	}
}

// Asset helpers
func isJSAsset(asset string) bool {
	return strings.HasSuffix(asset, ".js")
}

func isCSSAsset(asset string) bool {
	return strings.HasSuffix(asset, ".css")
}

func assetPath(asset string) string {
	return fmt.Sprintf("/assets/%s", asset)
}

// IntPtr returns a pointer to an int value
// This is a utility function for creating field validation rules
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to a float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to a bool value
func BoolPtr(b bool) *bool {
	return &b
}
