// Package components - Advanced validation system for component schemas
package components

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ValidationContext provides context for validation operations
type ValidationContext struct {
	ComponentType   string
	ComponentName   string
	ValidationMode  ValidationMode
	CustomRules     map[string]ComponentValidationRule
	ErrorCollection []ComponentValidationError
}

// ValidationMode defines different validation strictness levels
type ValidationMode int

const (
	ValidationModeStrict ValidationMode = iota // Strict validation - all rules enforced
	ValidationModeRelaxed                      // Relaxed validation - warnings instead of errors
	ValidationModeProduction                   // Production validation - performance optimized
)

// ComponentValidationError represents a validation error with detailed context
type ComponentValidationError struct {
	Field       string            `json:"field"`
	Value       interface{}       `json:"value,omitempty"`
	Rule        string            `json:"rule"`
	Message     string            `json:"message"`
	Severity    ValidationSeverity `json:"severity"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string          `json:"suggestions,omitempty"`
}

// ValidationSeverity defines the severity of validation issues
type ValidationSeverity int

const (
	SeverityError ValidationSeverity = iota
	SeverityWarning
	SeverityInfo
)

func (s ValidationSeverity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// ValidationResult contains the complete validation outcome
type ValidationResult struct {
	Valid       bool                         `json:"valid"`
	Errors      []ComponentValidationError   `json:"errors,omitempty"`
	Warnings    []ComponentValidationError   `json:"warnings,omitempty"`
	Info        []ComponentValidationError   `json:"info,omitempty"`
	Summary     ValidationSummary            `json:"summary"`
	Performance ValidationMetrics           `json:"performance,omitempty"`
}

// ValidationSummary provides a high-level overview of validation results
type ValidationSummary struct {
	TotalIssues   int `json:"totalIssues"`
	ErrorCount    int `json:"errorCount"`
	WarningCount  int `json:"warningCount"`
	InfoCount     int `json:"infoCount"`
	ComponentType string `json:"componentType"`
	ValidationMode string `json:"validationMode"`
}

// ValidationMetrics tracks performance of validation operations
type ValidationMetrics struct {
	ValidationTimeMs int64 `json:"validationTimeMs"`
	RulesExecuted   int   `json:"rulesExecuted"`
	FieldsValidated int   `json:"fieldsValidated"`
}

// SchemaValidator provides advanced schema validation capabilities
type SchemaValidator struct {
	context       *ValidationContext
	rules         map[string][]ComponentValidationRule
	customRules   map[string]ValidationFunc
	cache         map[string]*ValidationResult
	performance   ValidationMetrics
}

// ValidationFunc represents a custom validation function
type ValidationFunc func(value interface{}, context *ValidationContext) []ComponentValidationError

// ComponentValidationRule represents a single validation rule with metadata
type ComponentValidationRule struct {
	Name        string            `json:"name"`
	Field       string            `json:"field"`
	Type        string            `json:"type"`
	Required    bool              `json:"required,omitempty"`
	Pattern     string            `json:"pattern,omitempty"`
	MinLength   *int              `json:"minLength,omitempty"`
	MaxLength   *int              `json:"maxLength,omitempty"`
	MinValue    *float64          `json:"minValue,omitempty"`
	MaxValue    *float64          `json:"maxValue,omitempty"`
	Enum        []interface{}     `json:"enum,omitempty"`
	Custom      ValidationFunc    `json:"-"`
	Message     string            `json:"message,omitempty"`
	Severity    ValidationSeverity `json:"severity,omitempty"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Description string            `json:"description,omitempty"`
}

// NewSchemaValidator creates a new advanced schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		rules:       make(map[string][]ComponentValidationRule),
		customRules: make(map[string]ValidationFunc),
		cache:       make(map[string]*ValidationResult),
		context: &ValidationContext{
			ValidationMode: ValidationModeStrict,
			CustomRules:    make(map[string]ComponentValidationRule),
		},
	}
}

// SetValidationMode configures the validation strictness
func (v *SchemaValidator) SetValidationMode(mode ValidationMode) {
	v.context.ValidationMode = mode
}

// AddCustomRule registers a custom validation rule
func (v *SchemaValidator) AddCustomRule(name string, rule ValidationFunc) {
	v.customRules[name] = rule
}

// ValidateComponent performs comprehensive component validation
func (v *SchemaValidator) ValidateComponent(component interface{}) *ValidationResult {
	startTime := getCurrentTimeMs()
	
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ComponentValidationError, 0),
		Warnings: make([]ComponentValidationError, 0),
		Info:     make([]ComponentValidationError, 0),
	}

	// Determine component type
	componentType := getComponentType(component)
	v.context.ComponentType = componentType
	
	// Get validation rules for this component type
	rules := v.getValidationRules(componentType)
	
	// Add custom rules if any
	for name, customRule := range v.customRules {
		customValidationRule := ComponentValidationRule{
			Name:    name,
			Type:    "custom",
			Custom:  customRule,
			Message: "Custom validation rule: " + name,
		}
		rules = append(rules, customValidationRule)
	}
	
	// Perform validation
	errors := v.validateAgainstRules(component, rules)
	
	// Categorize errors by severity
	for _, err := range errors {
		switch err.Severity {
		case SeverityError:
			result.Errors = append(result.Errors, err)
			result.Valid = false
		case SeverityWarning:
			result.Warnings = append(result.Warnings, err)
		case SeverityInfo:
			result.Info = append(result.Info, err)
		}
	}
	
	// Generate summary
	result.Summary = ValidationSummary{
		TotalIssues:    len(result.Errors) + len(result.Warnings) + len(result.Info),
		ErrorCount:     len(result.Errors),
		WarningCount:   len(result.Warnings),
		InfoCount:      len(result.Info),
		ComponentType:  componentType,
		ValidationMode: v.context.ValidationMode.String(),
	}
	
	// Performance metrics
	endTime := getCurrentTimeMs()
	result.Performance = ValidationMetrics{
		ValidationTimeMs: endTime - startTime,
		RulesExecuted:    len(rules),
		FieldsValidated:  countFields(component),
	}
	
	return result
}

// ValidateComponentBatch validates multiple components efficiently
func (v *SchemaValidator) ValidateComponentBatch(components map[string]interface{}) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)
	
	for name, component := range components {
		v.context.ComponentName = name
		results[name] = v.ValidateComponent(component)
	}
	
	return results
}

// getValidationRules returns validation rules for a specific component type
func (v *SchemaValidator) getValidationRules(componentType string) []ComponentValidationRule {
	// Base rules that apply to all components
	baseRules := []ComponentValidationRule{
		{
			Name:     "type_required",
			Field:    "type",
			Type:     "string",
			Required: true,
			Message:  "Component type is required",
			Severity: SeverityError,
		},
		{
			Name:     "id_format",
			Field:    "id",
			Type:     "string",
			Pattern:  "^[a-zA-Z][a-zA-Z0-9_-]*$",
			Message:  "ID must start with a letter and contain only alphanumeric characters, underscores, or hyphens",
			Severity: SeverityWarning,
		},
		{
			Name:     "className_format",
			Field:    "className",
			Type:     "string",
			Pattern:  "^[a-zA-Z][a-zA-Z0-9\\s_-]*$",
			Message:  "Class name should follow CSS class naming conventions",
			Severity: SeverityWarning,
		},
	}
	
	// Component-specific rules
	specificRules := v.getComponentSpecificRules(componentType)
	
	// Combine base and specific rules
	allRules := append(baseRules, specificRules...)
	
	return allRules
}

// getComponentSpecificRules returns rules specific to each component type
func (v *SchemaValidator) getComponentSpecificRules(componentType string) []ComponentValidationRule {
	switch componentType {
	case "crud":
		return []ComponentValidationRule{
			{
				Name:     "crud_mode_valid",
				Field:    "mode",
				Type:     "string",
				Enum:     []interface{}{"table", "cards", "list"},
				Message:  "CRUD mode must be one of: table, cards, list",
				Severity: SeverityError,
			},
			{
				Name:     "crud_title_recommended",
				Field:    "title",
				Type:     "string",
				Required: false,
				Message:  "CRUD title is recommended for better user experience",
				Severity: SeverityInfo,
			},
		}
		
	case "form":
		return []ComponentValidationRule{
			{
				Name:     "form_mode_valid",
				Field:    "mode",
				Type:     "string",
				Enum:     []interface{}{"normal", "inline", "horizontal"},
				Message:  "Form mode must be one of: normal, inline, horizontal",
				Severity: SeverityError,
			},
			{
				Name:     "form_body_not_empty",
				Field:    "body",
				Type:     "array",
				Required: false,
				Message:  "Form should have at least one form control in body",
				Severity: SeverityWarning,
			},
		}
		
	case "table":
		return []ComponentValidationRule{
			{
				Name:     "table_columns_not_empty",
				Field:    "columns",
				Type:     "array",
				Required: false,
				Message:  "Table should have at least one column defined",
				Severity: SeverityWarning,
			},
		}
		
	case "input-text":
		return []ComponentValidationRule{
			{
				Name:     "text_input_name_required",
				Field:    "name",
				Type:     "string",
				Required: true,
				Message:  "Text input name is required for form submission",
				Severity: SeverityError,
			},
			{
				Name:     "text_input_label_recommended",
				Field:    "label",
				Type:     "string",
				Required: false,
				Message:  "Text input label is recommended for accessibility",
				Severity: SeverityWarning,
			},
		}
		
	case "select":
		return []ComponentValidationRule{
			{
				Name:     "select_name_required",
				Field:    "name",
				Type:     "string",
				Required: true,
				Message:  "Select control name is required for form submission",
				Severity: SeverityError,
			},
			{
				Name:     "select_options_or_source",
				Field:    "",
				Type:     "custom",
				Custom:   v.validateSelectOptionsOrSource,
				Message:  "Select control must have either options or source defined",
				Severity: SeverityError,
			},
		}
		
	case "input-date":
		return []ComponentValidationRule{
			{
				Name:     "date_name_required",
				Field:    "name",
				Type:     "string",
				Required: true,
				Message:  "Date control name is required for form submission",
				Severity: SeverityError,
			},
			{
				Name:     "date_format_valid",
				Field:    "format",
				Type:     "string",
				Pattern:  "^[YMDHmsS\\-\\s/:.,]+$",
				Message:  "Date format should use valid date format tokens",
				Severity: SeverityWarning,
			},
		}
		
	case "action":
		return []ComponentValidationRule{
			{
				Name:     "action_label_or_icon",
				Field:    "",
				Type:     "custom",
				Custom:   v.validateActionLabelOrIcon,
				Message:  "Action should have either label or icon for user understanding",
				Severity: SeverityWarning,
			},
			{
				Name:     "action_type_valid",
				Field:    "actionType",
				Type:     "string",
				Enum:     []interface{}{"ajax", "link", "dialog", "drawer", "submit", "reset", "button"},
				Message:  "Action type should be one of the supported types",
				Severity: SeverityError,
			},
		}
		
	case "dialog":
		return []ComponentValidationRule{
			{
				Name:     "dialog_title_or_body",
				Field:    "",
				Type:     "custom",
				Custom:   v.validateDialogContent,
				Message:  "Dialog should have either title or body content",
				Severity: SeverityWarning,
			},
			{
				Name:     "dialog_size_valid",
				Field:    "size",
				Type:     "string",
				Enum:     []interface{}{"xs", "sm", "md", "lg", "xl", "full"},
				Message:  "Dialog size should be one of the predefined sizes",
				Severity: SeverityWarning,
			},
		}
		
	default:
		return []ComponentValidationRule{}
	}
}

// validateAgainstRules executes validation rules against a component
func (v *SchemaValidator) validateAgainstRules(component interface{}, rules []ComponentValidationRule) []ComponentValidationError {
	var errors []ComponentValidationError
	
	componentValue := reflect.ValueOf(component)
	if componentValue.Kind() == reflect.Ptr {
		componentValue = componentValue.Elem()
	}
	
	for _, rule := range rules {
		ruleErrors := v.executeValidationRule(component, componentValue, rule)
		errors = append(errors, ruleErrors...)
	}
	
	return errors
}

// executeValidationRule executes a single validation rule
func (v *SchemaValidator) executeValidationRule(component interface{}, componentValue reflect.Value, rule ComponentValidationRule) []ComponentValidationError {
	var errors []ComponentValidationError
	
	// Handle custom validation functions
	if rule.Custom != nil {
		customErrors := rule.Custom(component, v.context)
		return customErrors
	}
	
	// Handle standard field validation
	if rule.Field != "" {
		fieldValue := getFieldValue(componentValue, rule.Field)
		fieldErrors := v.validateFieldValue(fieldValue, rule)
		errors = append(errors, fieldErrors...)
	}
	
	return errors
}

// validateFieldValue validates a single field value against a rule
func (v *SchemaValidator) validateFieldValue(fieldValue interface{}, rule ComponentValidationRule) []ComponentValidationError {
	var errors []ComponentValidationError
	
	// Required field validation
	if rule.Required && isEmptyValue(fieldValue) {
		errors = append(errors, ComponentValidationError{
			Field:    rule.Field,
			Value:    fieldValue,
			Rule:     rule.Name,
			Message:  rule.Message,
			Severity: rule.Severity,
		})
		return errors
	}
	
	// For warning-level validations, check empty values even if not required
	if rule.Severity == SeverityWarning && isEmptyValue(fieldValue) {
		errors = append(errors, ComponentValidationError{
			Field:    rule.Field,
			Value:    fieldValue,
			Rule:     rule.Name,
			Message:  rule.Message,
			Severity: rule.Severity,
		})
		return errors
	}
	
	// Skip further validation if field is empty and not required (for non-warnings)
	if isEmptyValue(fieldValue) {
		return errors
	}
	
	// Type validation
	if rule.Type != "" && rule.Type != "custom" {
		if !v.validateType(fieldValue, rule.Type) {
			errors = append(errors, ComponentValidationError{
				Field:    rule.Field,
				Value:    fieldValue,
				Rule:     rule.Name,
				Message:  fmt.Sprintf("Expected type %s but got %T", rule.Type, fieldValue),
				Severity: SeverityError,
			})
		}
	}
	
	// Pattern validation
	if rule.Pattern != "" {
		if str, ok := fieldValue.(string); ok {
			matched, err := regexp.MatchString(rule.Pattern, str)
			if err != nil || !matched {
				errors = append(errors, ComponentValidationError{
					Field:    rule.Field,
					Value:    fieldValue,
					Rule:     rule.Name,
					Message:  rule.Message,
					Severity: rule.Severity,
				})
			}
		}
	}
	
	// Length validation
	if rule.MinLength != nil || rule.MaxLength != nil {
		length := getValueLength(fieldValue)
		if rule.MinLength != nil && length < *rule.MinLength {
			errors = append(errors, ComponentValidationError{
				Field:    rule.Field,
				Value:    fieldValue,
				Rule:     rule.Name,
				Message:  fmt.Sprintf("Minimum length is %d, got %d", *rule.MinLength, length),
				Severity: rule.Severity,
			})
		}
		if rule.MaxLength != nil && length > *rule.MaxLength {
			errors = append(errors, ComponentValidationError{
				Field:    rule.Field,
				Value:    fieldValue,
				Rule:     rule.Name,
				Message:  fmt.Sprintf("Maximum length is %d, got %d", *rule.MaxLength, length),
				Severity: rule.Severity,
			})
		}
	}
	
	// Enum validation
	if len(rule.Enum) > 0 {
		valid := false
		for _, enumValue := range rule.Enum {
			if fieldValue == enumValue {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ComponentValidationError{
				Field:    rule.Field,
				Value:    fieldValue,
				Rule:     rule.Name,
				Message:  rule.Message,
				Severity: rule.Severity,
				Suggestions: convertEnumToStrings(rule.Enum),
			})
		}
	}
	
	return errors
}

// Custom validation functions for specific components

func (v *SchemaValidator) validateSelectOptionsOrSource(value interface{}, context *ValidationContext) []ComponentValidationError {
	var errors []ComponentValidationError
	
	selectControl, ok := value.(*SelectControlSchema)
	if !ok {
		return errors
	}
	
	hasOptions := len(selectControl.Options) > 0
	hasSource := selectControl.Source != "" || selectControl.API != nil
	
	if !hasOptions && !hasSource {
		errors = append(errors, ComponentValidationError{
			Field:    "options/source",
			Rule:     "select_options_or_source",
			Message:  "Select control must have either options array or data source defined",
			Severity: SeverityError,
			Suggestions: []string{
				"Add options array with at least one option",
				"Set source property to data endpoint",
				"Configure api property for dynamic options",
			},
		})
	}
	
	return errors
}

func (v *SchemaValidator) validateActionLabelOrIcon(value interface{}, context *ValidationContext) []ComponentValidationError {
	var errors []ComponentValidationError
	
	action, ok := value.(*ActionSchema)
	if !ok {
		return errors
	}
	
	hasLabel := action.Label != ""
	hasIcon := action.Icon != ""
	
	if !hasLabel && !hasIcon {
		errors = append(errors, ComponentValidationError{
			Field:    "label/icon",
			Rule:     "action_label_or_icon",
			Message:  "Action should have either label or icon for better user experience",
			Severity: SeverityWarning,
			Suggestions: []string{
				"Add label property with descriptive text",
				"Add icon property with icon name",
				"Consider adding both for better accessibility",
			},
		})
	}
	
	return errors
}

func (v *SchemaValidator) validateDialogContent(value interface{}, context *ValidationContext) []ComponentValidationError {
	var errors []ComponentValidationError
	
	dialog, ok := value.(*DialogSchema)
	if !ok {
		return errors
	}
	
	hasTitle := dialog.Title != ""
	hasBody := len(dialog.Body) > 0
	
	if !hasTitle && !hasBody {
		errors = append(errors, ComponentValidationError{
			Field:    "title/body",
			Rule:     "dialog_title_or_body",
			Message:  "Dialog should have either title or body content",
			Severity: SeverityWarning,
			Suggestions: []string{
				"Add title property with dialog title",
				"Add body array with dialog content",
				"Consider adding both for better user experience",
			},
		})
	}
	
	return errors
}

// Utility functions

func getComponentType(component interface{}) string {
	switch c := component.(type) {
	case *CRUDSchema:
		return c.Type
	case *TableSchema:
		return c.Type
	case *FormSchema:
		return c.Type
	case *ChartSchema:
		return c.Type
	case *ActionSchema:
		return c.Type
	case *DialogSchema:
		return c.Type
	case *TextControlSchema:
		return c.Type
	case *SelectControlSchema:
		return c.Type
	case *DateControlSchema:
		return c.Type
	default:
		return "unknown"
	}
}

func getFieldValue(componentValue reflect.Value, fieldPath string) interface{} {
	if componentValue.Kind() != reflect.Struct {
		return nil
	}
	
	// Handle nested field paths like "config.type"
	parts := strings.Split(fieldPath, ".")
	currentValue := componentValue
	
	for _, part := range parts {
		// Convert field name to proper Go struct field case
		fieldName := strings.Title(strings.ToLower(part))
		if part == "id" {
			fieldName = "ID"
		} else if part == "className" {
			fieldName = "ClassName"
		} else if part == "api" {
			fieldName = "API"
		} else if part == "body" {
			fieldName = "Body"
		} else if part == "label" {
			fieldName = "Label"
		} else if part == "actionType" {
			fieldName = "ActionType"
		}
		
		field := currentValue.FieldByName(fieldName)
		if !field.IsValid() {
			return nil
		}
		currentValue = field
	}
	
	if !currentValue.CanInterface() {
		return nil
	}
	
	return currentValue.Interface()
}

func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return true
		}
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return rv.Len() == 0
		}
		if rv.Kind() == reflect.String {
			return rv.String() == ""
		}
		return false
	}
}

func (v *SchemaValidator) validateType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			return true
		default:
			return false
		}
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
	case "object":
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Map || rv.Kind() == reflect.Struct
	default:
		return true // Unknown types pass validation
	}
}

func getValueLength(value interface{}) int {
	switch v := value.(type) {
	case string:
		return len(v)
	case []interface{}:
		return len(v)
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return rv.Len()
		}
		return 0
	}
}

func convertEnumToStrings(enum []interface{}) []string {
	suggestions := make([]string, len(enum))
	for i, value := range enum {
		suggestions[i] = fmt.Sprintf("%v", value)
	}
	return suggestions
}

func countFields(component interface{}) int {
	rv := reflect.ValueOf(component)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	return rv.NumField()
}

func getCurrentTimeMs() int64 {
	// Simple timestamp - in a real implementation, use time.Now().UnixMilli()
	return 1234567890 // Placeholder
}

func (vm ValidationMode) String() string {
	switch vm {
	case ValidationModeStrict:
		return "strict"
	case ValidationModeRelaxed:
		return "relaxed"
	case ValidationModeProduction:
		return "production"
	default:
		return "unknown"
	}
}

// GetValidationReport generates a comprehensive validation report
func (result *ValidationResult) GetValidationReport() string {
	var report strings.Builder
	
	report.WriteString(fmt.Sprintf("=== Validation Report ===\n"))
	report.WriteString(fmt.Sprintf("Component Type: %s\n", result.Summary.ComponentType))
	report.WriteString(fmt.Sprintf("Validation Mode: %s\n", result.Summary.ValidationMode))
	report.WriteString(fmt.Sprintf("Valid: %t\n", result.Valid))
	report.WriteString(fmt.Sprintf("Total Issues: %d\n", result.Summary.TotalIssues))
	report.WriteString(fmt.Sprintf("Errors: %d, Warnings: %d, Info: %d\n\n", 
		result.Summary.ErrorCount, result.Summary.WarningCount, result.Summary.InfoCount))
	
	if len(result.Errors) > 0 {
		report.WriteString("ERRORS:\n")
		for i, err := range result.Errors {
			report.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, err.Field, err.Message))
		}
		report.WriteString("\n")
	}
	
	if len(result.Warnings) > 0 {
		report.WriteString("WARNINGS:\n")
		for i, warn := range result.Warnings {
			report.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, warn.Field, warn.Message))
		}
		report.WriteString("\n")
	}
	
	return report.String()
}

// HasErrors returns true if the validation result contains errors
func (result *ValidationResult) HasErrors() bool {
	return len(result.Errors) > 0
}

// HasWarnings returns true if the validation result contains warnings
func (result *ValidationResult) HasWarnings() bool {
	return len(result.Warnings) > 0
}

// ToJSON serializes the validation result to JSON
func (result *ValidationResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize validation result: %w", err)
	}
	return string(data), nil
}