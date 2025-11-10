package validate

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// Database interface for uniqueness checks
type Database interface {
	// Exists checks if a value exists in the specified table/column
	Exists(ctx context.Context, table, column string, value any) (bool, error)
}

// BusinessRulesEngine interface for business rules validation
type BusinessRulesEngine interface {
	ApplyRules(ctx context.Context, schema *Schema, data map[string]any) (*Schema, error)
}

// Import schema types
type Schema = schema.Schema

// Field types
type FieldType string

const (
	// Basic text inputs
	FieldText     FieldType = "text"
	FieldEmail    FieldType = "email"
	FieldPassword FieldType = "password"
	FieldNumber   FieldType = "number"
	FieldHidden   FieldType = "hidden"
	FieldPhone    FieldType = "phone"
	FieldURL      FieldType = "url"

	// Date and time
	FieldDate     FieldType = "date"
	FieldTime     FieldType = "time"
	FieldDateTime FieldType = "datetime"

	// Text content
	FieldTextarea FieldType = "textarea"

	// Selection
	FieldSelect      FieldType = "select"
	FieldMultiSelect FieldType = "multiselect"
	FieldRadio       FieldType = "radio"
	FieldCheckboxes  FieldType = "checkboxes"

	// Specialized
	FieldCurrency FieldType = "currency"
	FieldFile     FieldType = "file"
	FieldImage    FieldType = "image"
)

// Field validation configuration
type FieldValidation struct {
	MinLength *int     `json:"minLength,omitempty"`
	MaxLength *int     `json:"maxLength,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Step      *float64 `json:"step,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Format    string   `json:"format,omitempty"`
	Custom    string   `json:"custom,omitempty"`
}

// Option for select fields
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FieldRuntime represents runtime state of a field (populated by enricher)
type FieldRuntime interface {
	IsVisible() bool
	IsEditable() bool
	GetReason() string
}

// Field represents a form field (interface to avoid circular import)
type FieldInterface interface {
	GetName() string
	GetType() FieldType
	GetRequired() bool
	GetValidation() *FieldValidation
	GetOptions() []Option
	GetConfig() map[string]any
}

// EnrichedFieldInterface extends FieldInterface with runtime state
type EnrichedFieldInterface interface {
	FieldInterface
	GetRuntime() FieldRuntime // Returns nil if not enriched
}

// Schema represents a form schema (interface to avoid circular import)
type SchemaInterface interface {
	GetID() string
	GetType() string
	GetTitle() string
	GetFields() []FieldInterface
	GetValidation() any
}

// EnrichedSchemaInterface extends SchemaInterface with enriched field support
type EnrichedSchemaInterface interface {
	SchemaInterface
	GetEnrichedFields() []EnrichedFieldInterface
}

// Validator provides server-side validation for schema data
type Validator struct {
	db                  Database // Optional database for uniqueness checks
	businessRulesEngine BusinessRulesEngine // Optional business rules engine
}

// ValidationResult contains the results of validation
type ValidationResult struct {
	Valid  bool                `json:"valid"`
	Errors map[string][]string `json:"errors"` // field -> error messages
	Data   map[string]any      `json:"data"`   // validated/cleaned data
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidator creates a new validator
func NewValidator(db Database) *Validator {
	return &Validator{
		db: db,
	}
}

// NewValidatorWithBusinessRules creates a new validator with business rules engine
func NewValidatorWithBusinessRules(db Database, bre BusinessRulesEngine) *Validator {
	return &Validator{
		db:                  db,
		businessRulesEngine: bre,
	}
}

// ValidateSchema validates schema structure (implements schema.Validator interface)
func (v *Validator) ValidateSchema(ctx context.Context, schema *schema.Schema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Use the existing schema validation method
	return schema.Validate(ctx)
}

// ValidateData validates form data against a schema (implements schema.Validator interface)
func (v *Validator) ValidateData(ctx context.Context, schema *schema.Schema, data map[string]any) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Use the existing schema data validation method
	return schema.ValidateData(ctx, data)
}

// ValidateDataDetailed validates form data against a schema and returns detailed results
func (v *Validator) ValidateDataDetailed(ctx context.Context, schema SchemaInterface, data map[string]any) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string][]string),
		Data:   make(map[string]any),
	}

	// Validate each field
	for _, field := range schema.GetFields() {
		value, exists := data[field.GetName()]

		// Validate field
		fieldErrors := v.ValidateField(ctx, field, value, exists)
		if len(fieldErrors) > 0 {
			result.Valid = false
			result.Errors[field.GetName()] = fieldErrors
		} else {
			// Store validated/cleaned value
			result.Data[field.GetName()] = v.cleanValue(field, value)
		}
	}

	// Apply business rules validation if engine is available
	if v.businessRulesEngine != nil && len(data) > 0 {
		if modifiedSchema, err := v.businessRulesEngine.ApplyRules(ctx, schema, data); err == nil {
			// Use the modified schema for validation context
			_ = modifiedSchema // Business rules applied successfully
		}
	}

	return result, nil
}

// ValidateField validates a single field value
func (v *Validator) ValidateField(ctx context.Context, field FieldInterface, value any, exists bool) []string {
	var errors []string

	// Check required
	if field.GetRequired() && (!exists || v.isEmpty(value)) {
		errors = append(errors, "This field is required")
		return errors // Don't continue validation if required field is missing
	}

	// Skip validation if field is empty and not required
	if !exists || v.isEmpty(value) {
		return errors
	}

	// Type-specific validation
	switch field.GetType() {
	case FieldText, FieldTextarea:
		errors = append(errors, v.validateString(field, value)...)
	case FieldEmail:
		errors = append(errors, v.validateEmail(field, value)...)
	case FieldPassword:
		errors = append(errors, v.validatePassword(field, value)...)
	case FieldNumber, FieldCurrency:
		errors = append(errors, v.validateNumber(field, value)...)
	case FieldPhone:
		errors = append(errors, v.validatePhone(field, value)...)
	case FieldURL:
		errors = append(errors, v.validateURL(field, value)...)
	case FieldDate, FieldDateTime, FieldTime:
		errors = append(errors, v.validateDate(field, value)...)
	case FieldSelect, FieldRadio:
		errors = append(errors, v.validateSelect(field, value)...)
	case FieldMultiSelect, FieldCheckboxes:
		errors = append(errors, v.validateMultiSelect(field, value)...)
	case FieldFile, FieldImage:
		errors = append(errors, v.validateFile(field, value)...)
	}

	// Custom validation if specified
	validation := field.GetValidation()
	if validation != nil && validation.Custom != "" {
		customErrors := v.validateCustom(field, value)
		errors = append(errors, customErrors...)
	}

	// Database uniqueness check (requires database connection)
	if validation != nil && validation.Unique && v.db != nil {
		if err := v.checkUniqueness(field, value); err != nil {
			errors = append(errors, "This value must be unique")
		}
	}

	return errors
}

// ValidateBusinessRules runs business rules validation
func (v *Validator) ValidateBusinessRules(ctx context.Context, schema SchemaInterface, data map[string]any) map[string][]string {
	errors := make(map[string][]string)

	// Integrate with business rules engine
	if v.businessRulesEngine != nil {
		if _, err := v.businessRulesEngine.ApplyRules(ctx, schema.(*Schema), data); err != nil {
			if schemaErr, ok := err.(SchemaError); ok {
				for field, fieldErrs := range schemaErr.ValidationErrors {
					errors[field] = fieldErrs
				}
			}
		}
	}
	// For now, return empty errors as business rules engine would handle this

	return errors
}

// ValidateEnrichedSchema validates form data against an enriched schema
// This method respects field visibility and editability from the enricher
func (v *Validator) ValidateEnrichedSchema(ctx context.Context, schema EnrichedSchemaInterface, data map[string]any) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string][]string),
		Data:   make(map[string]any),
	}

	// Validate each enriched field
	for _, field := range schema.GetEnrichedFields() {
		// Skip validation for invisible fields
		runtime := field.GetRuntime()
		if runtime != nil && !runtime.IsVisible() {
			continue
		}

		value, exists := data[field.GetName()]

		// Validate field with enriched context
		fieldErrors := v.ValidateWithFieldState(ctx, field, value, exists)
		if len(fieldErrors) > 0 {
			result.Valid = false
			result.Errors[field.GetName()] = fieldErrors
		} else {
			// Store validated/cleaned value
			result.Data[field.GetName()] = v.cleanValue(field, value)
		}
	}

	// Run business rules validation
	businessRuleErrors := v.ValidateBusinessRules(ctx, schema, data)
	for field, errors := range businessRuleErrors {
		if len(errors) > 0 {
			result.Valid = false
			if result.Errors[field] == nil {
				result.Errors[field] = errors
			} else {
				result.Errors[field] = append(result.Errors[field], errors...)
			}
		}
	}

	return result, nil
}

// ValidateWithFieldState validates a field considering its runtime state
// This method integrates with enricher's permission system
func (v *Validator) ValidateWithFieldState(ctx context.Context, field EnrichedFieldInterface, value any, exists bool) []string {
	var errors []string

	runtime := field.GetRuntime()
	if runtime != nil {
		// Skip validation for invisible fields
		if !runtime.IsVisible() {
			return errors
		}

		// For readonly/non-editable fields, validate but allow existing values
		if !runtime.IsEditable() {
			// Only validate type and format, not required or constraints
			// since users can't modify these fields
			return v.validateReadOnlyField(field, value, exists)
		}
	}

	// For editable fields or fields without runtime state, run full validation
	return v.ValidateField(ctx, field, value, exists)
}

// ValidateConditionalFields validates fields based on their visibility/editability state
// This method supports conditional validation based on enriched permissions
func (v *Validator) ValidateConditionalFields(ctx context.Context, schema EnrichedSchemaInterface, data map[string]any) map[string][]string {
	errors := make(map[string][]string)

	for _, field := range schema.GetEnrichedFields() {
		runtime := field.GetRuntime()
		if runtime == nil {
			continue
		}

		value, exists := data[field.GetName()]

		// Apply conditional validation logic based on runtime state
		var fieldErrors []string

		if runtime.IsVisible() {
			if runtime.IsEditable() {
				// Full validation for editable fields
				fieldErrors = v.ValidateField(ctx, field, value, exists)
			} else {
				// Limited validation for readonly fields
				fieldErrors = v.validateReadOnlyField(field, value, exists)
			}
		} else {
			// For invisible fields, ensure no data is submitted
			if exists && !v.isEmpty(value) {
				fieldErrors = append(fieldErrors, "Field is not accessible")
			}
		}

		if len(fieldErrors) > 0 {
			errors[field.GetName()] = fieldErrors
		}
	}

	return errors
}

// validateReadOnlyField validates readonly fields with limited rules
func (v *Validator) validateReadOnlyField(field FieldInterface, value any, exists bool) []string {
	var errors []string

	// Skip validation if field is empty (readonly fields can be empty)
	if !exists || v.isEmpty(value) {
		return errors
	}

	// Only validate type and format for readonly fields
	switch field.GetType() {
	case FieldText, FieldTextarea:
		if _, ok := value.(string); !ok {
			errors = append(errors, "Must be a string")
		}
	case FieldEmail:
		if str, ok := value.(string); ok {
			emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			if matched, _ := regexp.MatchString(emailRegex, str); !matched {
				errors = append(errors, "Must be a valid email address")
			}
		} else {
			errors = append(errors, "Must be a string")
		}
	case FieldNumber, FieldCurrency:
		// Validate that it's a valid number
		switch value.(type) {
		case float64, float32, int, int64:
			// Valid number types
		case string:
			if _, err := strconv.ParseFloat(value.(string), 64); err != nil {
				errors = append(errors, "Must be a valid number")
			}
		default:
			errors = append(errors, "Must be a valid number")
		}
	case FieldDate, FieldDateTime, FieldTime:
		// Basic date format validation for readonly fields
		if str, ok := value.(string); ok {
			var err error
			switch field.GetType() {
			case FieldDate:
				_, err = time.Parse("2006-01-02", str)
			case FieldTime:
				_, err = time.Parse("15:04", str)
			case FieldDateTime:
				formats := []string{"2006-01-02T15:04:05Z", "2006-01-02T15:04:05", "2006-01-02 15:04:05"}
				for _, format := range formats {
					_, err = time.Parse(format, str)
					if err == nil {
						break
					}
				}
			}
			if err != nil {
				errors = append(errors, "Invalid date format")
			}
		} else {
			errors = append(errors, "Must be a string")
		}
	}

	return errors
}

// String validation
func (v *Validator) validateString(field FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	validation := field.GetValidation()
	if validation == nil {
		return errors
	}

	// Length validation
	if validation.MinLength != nil && len(str) < *validation.MinLength {
		errors = append(errors, fmt.Sprintf("Must be at least %d characters", *validation.MinLength))
	}
	if validation.MaxLength != nil && len(str) > *validation.MaxLength {
		errors = append(errors, fmt.Sprintf("Must be no more than %d characters", *validation.MaxLength))
	}

	// Pattern validation
	if validation.Pattern != "" {
		matched, err := regexp.MatchString(validation.Pattern, str)
		if err != nil {
			errors = append(errors, "Invalid pattern validation")
		} else if !matched {
			message := "Invalid format"
			if validation.PatternMessage != "" {
				message = validation.PatternMessage
			}
			errors = append(errors, message)
		}
	}

	return errors
}

// Email validation
func (v *Validator) validateEmail(field FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Basic email regex
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, str)
	if err != nil || !matched {
		errors = append(errors, "Must be a valid email address")
	}

	// Additional string validation
	errors = append(errors, v.validateString(field, value)...)

	return errors
}

// Password validation
func (v *Validator) validatePassword(field FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	validation := field.GetValidation()
	if validation != nil {
		// Minimum length (common for passwords)
		if validation.MinLength != nil && len(str) < *validation.MinLength {
			errors = append(errors, fmt.Sprintf("Password must be at least %d characters", *validation.MinLength))
		}

		// Pattern validation (for complexity requirements)
		if validation.Pattern != "" {
			matched, _ := regexp.MatchString(validation.Pattern, str)
			if !matched {
				message := "Password does not meet complexity requirements"
				if validation.PatternMessage != "" {
					message = validation.PatternMessage
				}
				errors = append(errors, message)
			}
		}
	}

	return errors
}

// Number validation
func (v *Validator) validateNumber(field FieldInterface, value any) []string {
	var errors []string

	var num float64
	var ok bool

	// Convert to float64
	switch v := value.(type) {
	case float64:
		num = v
		ok = true
	case float32:
		num = float64(v)
		ok = true
	case int:
		num = float64(v)
		ok = true
	case int64:
		num = float64(v)
		ok = true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			num = parsed
			ok = true
		}
	}

	if !ok {
		errors = append(errors, "Must be a valid number")
		return errors
	}

	validation := field.GetValidation()
	if validation == nil {
		return errors
	}

	// Range validation
	if validation.Min != nil && num < *validation.Min {
		errors = append(errors, fmt.Sprintf("Must be at least %g", *validation.Min))
	}
	if validation.Max != nil && num > *validation.Max {
		errors = append(errors, fmt.Sprintf("Must be no more than %g", *validation.Max))
	}

	// Step validation
	if validation.Step != nil && *validation.Step > 0 {
		remainder := num - (float64(int(num / *validation.Step)) * *validation.Step)
		if remainder != 0 {
			errors = append(errors, fmt.Sprintf("Must be a multiple of %g", *validation.Step))
		}
	}

	return errors
}

// Phone validation
func (v *Validator) validatePhone(_ FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Clean the input - remove all non-digit characters except leading +
	cleaned := str
	hasPlus := strings.HasPrefix(str, "+")
	cleaned = regexp.MustCompile(`[^\d]`).ReplaceAllString(str, "")

	// Length check (reasonable phone number length)
	if len(cleaned) < 7 || len(cleaned) > 15 {
		errors = append(errors, "Phone number must be between 7 and 15 digits")
		return errors // Return early to avoid multiple errors
	}

	// Basic E.164 format validation for international numbers with + prefix
	if hasPlus {
		phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(str) {
			errors = append(errors, "Must be a valid phone number")
		}
	} else {
		// For numbers without +, require proper format (this will trigger error for simple numbers)
		if len(cleaned) == 10 {
			// Allow 10-digit numbers as they could be local numbers
			return errors
		}
		errors = append(errors, "Must be a valid phone number")
	}

	return errors
}

// URL validation
func (v *Validator) validateURL(_ FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Basic URL validation
	urlRegex := `^https?://[^\s/$.?#].[^\s]*$`
	matched, err := regexp.MatchString(urlRegex, str)
	if err != nil || !matched {
		errors = append(errors, "Must be a valid URL")
	}

	return errors
}

// Date validation
func (v *Validator) validateDate(field FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	var parsedTime time.Time
	var err error

	// Try different date formats based on field type
	switch field.GetType() {
	case FieldDate:
		parsedTime, err = time.Parse("2006-01-02", str)
	case FieldTime:
		parsedTime, err = time.Parse("15:04", str)
	case FieldDateTime:
		// Try multiple datetime formats
		formats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		for _, format := range formats {
			parsedTime, err = time.Parse(format, str)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		errors = append(errors, "Invalid date format")
		return errors
	}

	validation := field.GetValidation()
	if validation == nil {
		return errors
	}

	// Date range validation
	if validation.Min != nil {
		if minTime, err := time.Parse("2006-01-02", fmt.Sprintf("%v", validation.Min)); err == nil {
			if parsedTime.Before(minTime) {
				errors = append(errors, fmt.Sprintf("Date must be after %s", minTime.Format("2006-01-02")))
			}
		}
	}
	if validation.Max != nil {
		if maxTime, err := time.Parse("2006-01-02", fmt.Sprintf("%v", validation.Max)); err == nil {
			if parsedTime.After(maxTime) {
				errors = append(errors, fmt.Sprintf("Date must be before %s", maxTime.Format("2006-01-02")))
			}
		}
	}

	return errors
}

// Select validation
func (v *Validator) validateSelect(field FieldInterface, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Check if value is in options
	options := field.GetOptions()
	if len(options) > 0 {
		found := false
		for _, option := range options {
			if option.Value == str {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, "Invalid selection")
		}
	}

	return errors
}

// Multi-select validation
func (v *Validator) validateMultiSelect(field FieldInterface, value any) []string {
	var errors []string

	// Handle different input formats
	var values []string
	switch v := value.(type) {
	case []string:
		values = v
	case []any:
		for _, item := range v {
			if str, ok := item.(string); ok {
				values = append(values, str)
			} else {
				errors = append(errors, "All values must be strings")
				return errors
			}
		}
	case string:
		// Handle comma-separated values
		if v != "" {
			values = strings.Split(v, ",")
			for i, val := range values {
				values[i] = strings.TrimSpace(val)
			}
		}
	default:
		errors = append(errors, "Must be an array of strings")
		return errors
	}

	// Check if all values are in options
	options := field.GetOptions()
	if len(options) > 0 {
		validValues := make(map[string]bool)
		for _, option := range options {
			validValues[option.Value] = true
		}

		for _, val := range values {
			if !validValues[val] {
				errors = append(errors, fmt.Sprintf("Invalid selection: %s", val))
			}
		}
	}

	return errors
}

// File validation
func (v *Validator) validateFile(field FieldInterface, value any) []string {
	var errors []string

	// File validation would typically happen at upload time
	// This is a placeholder for file metadata validation

	if _, ok := value.(string); !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Basic file path/name validation
	// Note: Empty string is allowed for optional file fields
	// Required validation is handled separately

	return errors
}

// Custom validation
func (v *Validator) validateCustom(field FieldInterface, value any) []string {
	var errors []string
	
	validation := field.GetValidation()
	if validation == nil || validation.Custom == "" {
		return errors
	}

	// Use condition package for custom validation expression evaluation
	// Create evaluation context with field value
	data := map[string]any{
		"value": value,
		"field": field.GetName(),
	}
	
	// Try to evaluate the custom expression as a formula
	// If it returns false, the validation fails
	if v.businessRulesEngine != nil {
		// Custom expressions should return boolean true for valid values
		// Example: "value > 0 && value < 100" or "len(value) >= 3"
		// We'd need an expression evaluator here, but for now return empty
		// The actual implementation would use the condition package's evaluator
	}

	return errors
}

// Check uniqueness in database
func (v *Validator) checkUniqueness(ctx context.Context, field FieldInterface, value any) (bool, error) {
	if v.db == nil {
		return true, nil // No database, assume unique
	}

	// Extract table name from field config or use schema context
	tableName := "unknown_table" // Would come from schema context
	config := field.GetConfig()
	if config != nil {
		if table, ok := config["table"].(string); ok {
			tableName = table
		}
	}

	exists, err := v.db.Exists(ctx, tableName, field.GetName(), value)
	if err != nil {
		return false, err
	}

	return !exists, nil // Unique if it doesn't exist
}

// Check if value is empty
func (v *Validator) isEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

// Clean and normalize value
func (v *Validator) cleanValue(field FieldInterface, value any) any {
	if value == nil {
		return nil
	}

	switch field.GetType() {
	case FieldText, FieldTextarea, FieldEmail, FieldPassword:
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str)
		}
	case FieldNumber, FieldCurrency:
		// Ensure consistent number format
		switch v := value.(type) {
		case string:
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				return num
			}
		case float64, float32, int, int64:
			return v
		}
	}

	return value
}

// ValidateWithRegistry validates a field using the validation registry
func (v *Validator) ValidateWithRegistry(ctx context.Context, field FieldInterface, value any, registry *ValidationRegistry) []string {
	// First run standard validation
	errors := v.ValidateField(ctx, field, value, value != nil)

	// Then run custom validators if configured
	validation := field.GetValidation()
	if validation != nil && validation.Custom != "" {
		params := make(map[string]any)

		// Add validation parameters
		if validation.Min != nil {
			params["min"] = *validation.Min
		}
		if validation.Max != nil {
			params["max"] = *validation.Max
		}
		if validation.MinLength != nil {
			params["minLength"] = *validation.MinLength
		}
		if validation.MaxLength != nil {
			params["maxLength"] = *validation.MaxLength
		}
		if validation.Pattern != "" {
			params["pattern"] = validation.Pattern
		}
		if validation.Format != "" {
			params["format"] = validation.Format
		}

		if err := registry.Validate(ctx, validation.Custom, value, params); err != nil {
			if validationErr, ok := err.(ValidationError); ok {
				errors = append(errors, validationErr.Message)
			} else {
				errors = append(errors, err.Error())
			}
		}
	}

	return errors
}

// ValidateSchemaWithRegistries validates a schema using both validation registries
func ValidateSchemaWithRegistries(ctx context.Context, schema SchemaInterface, data map[string]any,
	fieldRegistry *ValidationRegistry, crossFieldRegistry *CrossFieldValidationRegistry,
) error {
	validator := NewValidator(nil)
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string][]string),
		Data:   make(map[string]any),
	}

	// Validate individual fields
	for _, field := range schema.GetFields() {
		if value, exists := data[field.GetName()]; exists || field.GetRequired() {
			fieldErrors := validator.ValidateWithRegistry(ctx, field, value, fieldRegistry)
			if len(fieldErrors) > 0 {
				result.Valid = false
				result.Errors[field.GetName()] = fieldErrors
			} else {
				// Store validated/cleaned value
				result.Data[field.GetName()] = validator.cleanValue(field, value)
			}
		}
	}

	// Validate cross-field rules if configured
	if schema.GetValidation() != nil {
		// Use built-in validators
		for _, validatorName := range []string{"date_range", "password_confirmation", "business_hours"} {
			if err := crossFieldRegistry.Validate(ctx, validatorName, data); err != nil {
				result.Valid = false
				if validationErr, ok := err.(ValidationError); ok {
					if result.Errors[""] == nil {
						result.Errors[""] = []string{}
					}
					result.Errors[""] = append(result.Errors[""], validationErr.Message)
				} else {
					if result.Errors[""] == nil {
						result.Errors[""] = []string{}
					}
					result.Errors[""] = append(result.Errors[""], err.Error())
				}
			}
		}
	}

	if !result.Valid {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// checkUniqueness checks if a field value is unique in the database
func (v *Validator) checkUniqueness(field FieldInterface, value any) error {
	if v.db == nil {
		return nil // No database configured, skip uniqueness check
	}
	
	// Determine table and column for uniqueness check
	// This could be enhanced to support custom table/column mapping
	tableName := "entities" // Default table
	columnName := field.GetName()
	
	// Check if value already exists
	exists, err := v.db.Exists(context.Background(), tableName, columnName, value)
	if err != nil {
		return fmt.Errorf("uniqueness check failed: %w", err)
	}
	
	if exists {
		return fmt.Errorf("value already exists")
	}
	
	return nil
}

// Utility function to create a validation error
func NewValidationError(field, message, code string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}
