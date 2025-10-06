package core

import (
	"fmt"
	"strings"
)

// SchemaValidator validates schema compositions
type SchemaValidator struct {
	rules []ValidationRule
}

// ValidationRule defines a validation rule
type ValidationRule interface {
	Validate(schema *CompositionSchema) []ValidationIssue
}

// ValidationIssue represents a validation problem
type ValidationIssue struct {
	Level       string `json:"level"` // "error", "warning", "info"
	Code        string `json:"code"`
	Message     string `json:"message"`
	ComponentID string `json:"componentId,omitempty"`
	Field       string `json:"field,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	validator := &SchemaValidator{
		rules: []ValidationRule{},
	}

	// Register default validation rules
	validator.registerDefaultRules()
	return validator
}

// ValidateSchema validates the complete schema
func (v *SchemaValidator) ValidateSchema(schema *CompositionSchema) error {
	var errors []string
	var warnings []string

	for _, rule := range v.rules {
		issues := rule.Validate(schema)
		for _, issue := range issues {
			switch issue.Level {
			case "error":
				errors = append(errors, issue.Message)
			case "warning":
				warnings = append(warnings, issue.Message)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetValidationIssues returns all validation issues without failing
func (v *SchemaValidator) GetValidationIssues(schema *CompositionSchema) []ValidationIssue {
	var allIssues []ValidationIssue

	for _, rule := range v.rules {
		issues := rule.Validate(schema)
		allIssues = append(allIssues, issues...)
	}

	return allIssues
}

// registerDefaultRules registers built-in validation rules
func (v *SchemaValidator) registerDefaultRules() {
	v.rules = append(v.rules,
		&RequiredFieldsRule{},
		&ComponentValidationRule{},
		&LayoutValidationRule{},
		&DataSourceValidationRule{},
		&PermissionValidationRule{},
		&AccessibilityRule{},
		&PerformanceRule{},
	)
}

// RequiredFieldsRule validates required schema fields
type RequiredFieldsRule struct{}

func (r *RequiredFieldsRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	if schema.ID == "" {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Code:    "MISSING_ID",
			Message: "Schema must have an ID",
		})
	}

	if schema.Name == "" {
		issues = append(issues, ValidationIssue{
			Level:   "warning",
			Code:    "MISSING_NAME",
			Message: "Schema should have a name",
		})
	}

	if schema.Title == "" {
		issues = append(issues, ValidationIssue{
			Level:   "warning",
			Code:    "MISSING_TITLE",
			Message: "Schema should have a title",
		})
	}

	return issues
}

// ComponentValidationRule validates component configurations
type ComponentValidationRule struct{}

func (r *ComponentValidationRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	for _, component := range schema.Components {
		// Check required component fields
		if component.ID == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Code:    "MISSING_COMPONENT_ID",
				Message: "Component must have an ID",
			})
		}

		if component.Type == "" {
			issues = append(issues, ValidationIssue{
				Level:       "error",
				Code:        "MISSING_COMPONENT_TYPE",
				Message:     "Component must have a type",
				ComponentID: component.ID,
			})
		}

		// Validate component props based on type
		issues = append(issues, r.validateComponentProps(component)...)

		// Check for overlapping components
		issues = append(issues, r.checkOverlaps(component, schema.Components)...)
	}

	return issues
}

func (r *ComponentValidationRule) validateComponentProps(component ComponentInstance) []ValidationIssue {
	var issues []ValidationIssue

	// Type-specific validation
	switch component.Type {
	case "organisms.table":
		if component.Props["columns"] == nil {
			issues = append(issues, ValidationIssue{
				Level:       "error",
				Code:        "MISSING_TABLE_COLUMNS",
				Message:     "Table component must have columns defined",
				ComponentID: component.ID,
				Suggestion:  "Add columns property to table configuration",
			})
		}

	case "organisms.form":
		if component.Props["fields"] == nil {
			issues = append(issues, ValidationIssue{
				Level:       "error",
				Code:        "MISSING_FORM_FIELDS",
				Message:     "Form component must have fields defined",
				ComponentID: component.ID,
				Suggestion:  "Add fields property to form configuration",
			})
		}

	case "molecules.button":
		if component.Props["label"] == nil && component.Props["icon"] == nil {
			issues = append(issues, ValidationIssue{
				Level:       "warning",
				Code:        "BUTTON_NO_CONTENT",
				Message:     "Button should have either a label or icon",
				ComponentID: component.ID,
				Suggestion:  "Add label or icon property to button",
			})
		}
	}

	return issues
}

func (r *ComponentValidationRule) checkOverlaps(component ComponentInstance, allComponents []ComponentInstance) []ValidationIssue {
	var issues []ValidationIssue

	for _, other := range allComponents {
		if other.ID == component.ID {
			continue
		}

		// Check if components overlap significantly
		if r.componentsOverlap(component, other) {
			issues = append(issues, ValidationIssue{
				Level:       "warning",
				Code:        "COMPONENT_OVERLAP",
				Message:     fmt.Sprintf("Component overlaps with %s", other.Type),
				ComponentID: component.ID,
				Suggestion:  "Adjust component positions to avoid overlap",
			})
		}
	}

	return issues
}

func (r *ComponentValidationRule) componentsOverlap(a, b ComponentInstance) bool {
	// Simple overlap detection
	aRight := a.Position.X + a.Size.Width
	aBottom := a.Position.Y + a.Size.Height
	bRight := b.Position.X + b.Size.Width
	bBottom := b.Position.Y + b.Size.Height

	return !(aRight <= b.Position.X || bRight <= a.Position.X ||
		aBottom <= b.Position.Y || bBottom <= a.Position.Y)
}

// LayoutValidationRule validates layout configurations
type LayoutValidationRule struct{}

func (r *LayoutValidationRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	validLayouts := []string{"app", "auth", "minimal", "base"}
	isValidLayout := false
	for _, layout := range validLayouts {
		if schema.Layout == layout {
			isValidLayout = true
			break
		}
	}

	if !isValidLayout {
		issues = append(issues, ValidationIssue{
			Level:      "warning",
			Code:       "INVALID_LAYOUT",
			Message:    fmt.Sprintf("Layout '%s' may not be supported", schema.Layout),
			Suggestion: "Use one of: " + strings.Join(validLayouts, ", "),
		})
	}

	// Validate component layouts
	for _, component := range schema.Components {
		if component.Layout != nil {
			issues = append(issues, r.validateComponentLayout(component)...)
		}
	}

	return issues
}

func (r *LayoutValidationRule) validateComponentLayout(component ComponentInstance) []ValidationIssue {
	var issues []ValidationIssue

	if component.Layout.Grid != nil {
		grid := component.Layout.Grid
		if grid.Columns == "" {
			issues = append(issues, ValidationIssue{
				Level:       "warning",
				Code:        "MISSING_GRID_COLUMNS",
				Message:     "Grid layout should specify columns",
				ComponentID: component.ID,
			})
		}
	}

	return issues
}

// DataSourceValidationRule validates data source configurations
type DataSourceValidationRule struct{}

func (r *DataSourceValidationRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	for _, dataSource := range schema.DataSources {
		if dataSource.ID == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Code:    "MISSING_DATASOURCE_ID",
				Message: "Data source must have an ID",
			})
		}

		if dataSource.Type == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Code:    "MISSING_DATASOURCE_TYPE",
				Message: "Data source must have a type",
			})
		}

		if dataSource.Type == "api" && dataSource.Endpoint == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Code:    "MISSING_API_ENDPOINT",
				Message: "API data source must have an endpoint",
			})
		}
	}

	return issues
}

// PermissionValidationRule validates permission configurations
type PermissionValidationRule struct{}

func (r *PermissionValidationRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	for _, component := range schema.Components {
		if component.Permissions != nil {
			if len(component.Permissions.RequiredRoles) == 0 &&
				len(component.Permissions.RequiredPermissions) == 0 &&
				len(component.Permissions.Conditions) == 0 {
				issues = append(issues, ValidationIssue{
					Level:       "warning",
					Code:        "EMPTY_PERMISSIONS",
					Message:     "Permission rules are defined but empty",
					ComponentID: component.ID,
				})
			}
		}
	}

	return issues
}

// AccessibilityRule validates accessibility compliance
type AccessibilityRule struct{}

func (r *AccessibilityRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	for _, component := range schema.Components {
		// Check for missing accessibility labels
		switch component.Type {
		case "molecules.button":
			if component.Props["label"] == nil && component.Props["aria-label"] == nil {
				issues = append(issues, ValidationIssue{
					Level:       "warning",
					Code:        "MISSING_BUTTON_LABEL",
					Message:     "Button should have an accessible label",
					ComponentID: component.ID,
					Suggestion:  "Add label or aria-label property",
				})
			}

		case "atoms.input":
			if component.Props["label"] == nil && component.Props["aria-label"] == nil {
				issues = append(issues, ValidationIssue{
					Level:       "warning",
					Code:        "MISSING_INPUT_LABEL",
					Message:     "Input should have an accessible label",
					ComponentID: component.ID,
					Suggestion:  "Add label or aria-label property",
				})
			}
		}

		// Check color contrast (basic check)
		if bgColor, exists := component.Props["backgroundColor"]; exists {
			if textColor, exists2 := component.Props["color"]; exists2 {
				if !r.hasGoodContrast(bgColor.(string), textColor.(string)) {
					issues = append(issues, ValidationIssue{
						Level:       "warning",
						Code:        "LOW_COLOR_CONTRAST",
						Message:     "Color combination may have poor contrast",
						ComponentID: component.ID,
						Suggestion:  "Choose colors with better contrast ratio",
					})
				}
			}
		}
	}

	return issues
}

func (r *AccessibilityRule) hasGoodContrast(bgColor, textColor string) bool {
	// Basic contrast check - in production, use proper color contrast calculation
	lightColors := []string{"white", "#ffffff", "#fff", "yellow", "lightgray"}
	_ = []string{"black", "#000000", "#000", "darkblue", "darkgreen"}

	bgIsLight := r.isLightColor(bgColor, lightColors)
	textIsLight := r.isLightColor(textColor, lightColors)

	return bgIsLight != textIsLight // Different lightness = good contrast
}

func (r *AccessibilityRule) isLightColor(color string, lightColors []string) bool {
	color = strings.ToLower(color)
	for _, light := range lightColors {
		if color == light {
			return true
		}
	}
	return false
}

// PerformanceRule validates performance considerations
type PerformanceRule struct{}

func (r *PerformanceRule) Validate(schema *CompositionSchema) []ValidationIssue {
	var issues []ValidationIssue

	// Check for too many components
	if len(schema.Components) > 50 {
		issues = append(issues, ValidationIssue{
			Level:      "warning",
			Code:       "TOO_MANY_COMPONENTS",
			Message:    fmt.Sprintf("Page has %d components, consider splitting into multiple pages", len(schema.Components)),
			Suggestion: "Break complex pages into smaller, focused pages",
		})
	}

	// Check for nested components depth
	maxDepth := r.calculateMaxDepth(schema.Components)
	if maxDepth > 5 {
		issues = append(issues, ValidationIssue{
			Level:      "warning",
			Code:       "DEEP_NESTING",
			Message:    fmt.Sprintf("Component nesting is %d levels deep", maxDepth),
			Suggestion: "Reduce nesting depth for better performance",
		})
	}

	// Check for large data sources
	for _, dataSource := range schema.DataSources {
		if dataSource.Type == "api" {
			// Could check endpoint patterns for potentially large datasets
			endpoint := strings.ToLower(dataSource.Endpoint)
			if strings.Contains(endpoint, "all") || strings.Contains(endpoint, "list") {
				issues = append(issues, ValidationIssue{
					Level:      "info",
					Code:       "POTENTIAL_LARGE_DATASET",
					Message:    "Data source may return large datasets",
					Suggestion: "Consider pagination or filtering",
				})
			}
		}
	}

	return issues
}

func (r *PerformanceRule) calculateMaxDepth(components []ComponentInstance) int {
	maxDepth := 0
	for _, component := range components {
		depth := 1 + r.calculateComponentDepth(component.Children)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

func (r *PerformanceRule) calculateComponentDepth(children []ComponentInstance) int {
	if len(children) == 0 {
		return 0
	}

	maxChildDepth := 0
	for _, child := range children {
		childDepth := 1 + r.calculateComponentDepth(child.Children)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}
	return maxChildDepth
}
