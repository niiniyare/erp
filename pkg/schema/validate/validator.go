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

// Validator provides server-side validation for schema data
type Validator struct {
	db Database // Optional database for uniqueness checks
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

// ValidateData validates form data against a schema
func (v *Validator) ValidateData(ctx context.Context, schema *schema.Schema, data map[string]any) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make(map[string][]string),
		Data:   make(map[string]any),
	}

	// Validate each field
	for _, field := range schema.Fields {
		value, exists := data[field.Name]

		// Validate field
		fieldErrors := v.ValidateField(ctx, &field, value, exists)
		if len(fieldErrors) > 0 {
			result.Valid = false
			result.Errors[field.Name] = fieldErrors
		} else {
			// Store validated/cleaned value
			result.Data[field.Name] = v.cleanValue(&field, value)
		}
	}

	// TODO: Run business rules validation when BusinessRules field is implemented
	// This would integrate with the business rules engine

	return result, nil
}

// ValidateField validates a single field value
func (v *Validator) ValidateField(ctx context.Context, field *schema.Field, value any, exists bool) []string {
	var errors []string

	// Check required
	if field.Required && (!exists || v.isEmpty(value)) {
		errors = append(errors, "This field is required")
		return errors // Don't continue validation if required field is missing
	}

	// Skip validation if field is empty and not required
	if !exists || v.isEmpty(value) {
		return errors
	}

	// Type-specific validation
	switch field.Type {
	case schema.FieldText, schema.FieldTextarea:
		errors = append(errors, v.validateString(field, value)...)
	case schema.FieldEmail:
		errors = append(errors, v.validateEmail(field, value)...)
	case schema.FieldPassword:
		errors = append(errors, v.validatePassword(field, value)...)
	case schema.FieldNumber, schema.FieldCurrency:
		errors = append(errors, v.validateNumber(field, value)...)
	case schema.FieldPhone:
		errors = append(errors, v.validatePhone(field, value)...)
	case schema.FieldURL:
		errors = append(errors, v.validateURL(field, value)...)
	case schema.FieldDate, schema.FieldDateTime, schema.FieldTime:
		errors = append(errors, v.validateDate(field, value)...)
	case schema.FieldSelect, schema.FieldRadio:
		errors = append(errors, v.validateSelect(field, value)...)
	case schema.FieldMultiSelect, schema.FieldCheckboxes:
		errors = append(errors, v.validateMultiSelect(field, value)...)
	case schema.FieldFile, schema.FieldImage:
		errors = append(errors, v.validateFile(field, value)...)
	}

	// Custom validation if specified
	if field.Validation != nil && field.Validation.Custom != "" {
		customErrors := v.validateCustom(field, value)
		errors = append(errors, customErrors...)
	}

	// TODO: Database uniqueness check when Unique field is implemented
	// This would check database constraints for unique values

	return errors
}

// ValidateBusinessRules runs business rules validation
func (v *Validator) ValidateBusinessRules(ctx context.Context, schema *schema.Schema, data map[string]any) map[string][]string {
	errors := make(map[string][]string)

	// TODO: Integrate with business rules engine when implemented
	// This would evaluate conditions and apply validation rules
	// For now, return empty errors as business rules engine would handle this

	return errors
}

// String validation
func (v *Validator) validateString(field *schema.Field, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	validation := field.Validation
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
			// TODO: Use custom pattern message when PatternMessage field is implemented
			errors = append(errors, "Invalid format")
		}
	}

	return errors
}

// Email validation
func (v *Validator) validateEmail(field *schema.Field, value any) []string {
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
func (v *Validator) validatePassword(field *schema.Field, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	validation := field.Validation
	if validation != nil {
		// Minimum length (common for passwords)
		if validation.MinLength != nil && len(str) < *validation.MinLength {
			errors = append(errors, fmt.Sprintf("Password must be at least %d characters", *validation.MinLength))
		}

		// Pattern validation (for complexity requirements)
		if validation.Pattern != "" {
			matched, _ := regexp.MatchString(validation.Pattern, str)
			if !matched {
				// TODO: Use custom pattern message when PatternMessage field is implemented
				errors = append(errors, "Password does not meet complexity requirements")
			}
		}
	}

	return errors
}

// Number validation
func (v *Validator) validateNumber(field *schema.Field, value any) []string {
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

	validation := field.Validation
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
func (v *Validator) validatePhone(_ *schema.Field, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Basic phone validation - digits, spaces, dashes, parentheses, plus
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`) // E.164 format
	matched, err := regexp.MatchString(phoneRegex.String(), str)
	if err != nil || !matched {
		errors = append(errors, "Must be a valid phone number")
	}

	// Length check (reasonable phone number length)
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(str, "")
	if len(cleaned) < 7 || len(cleaned) > 15 {
		errors = append(errors, "Phone number must be between 7 and 15 digits")
	}

	return errors
}

// URL validation
func (v *Validator) validateURL(_ *schema.Field, value any) []string {
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
func (v *Validator) validateDate(field *schema.Field, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	var parsedTime time.Time
	var err error

	// Try different date formats based on field type
	switch field.Type {
	case schema.FieldDate:
		parsedTime, err = time.Parse("2006-01-02", str)
	case schema.FieldTime:
		parsedTime, err = time.Parse("15:04", str)
	case schema.FieldDateTime:
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

	validation := field.Validation
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
func (v *Validator) validateSelect(field *schema.Field, value any) []string {
	var errors []string

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Check if value is in options
	if len(field.Options) > 0 {
		found := false
		for _, option := range field.Options {
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
func (v *Validator) validateMultiSelect(field *schema.Field, value any) []string {
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
	if len(field.Options) > 0 {
		validValues := make(map[string]bool)
		for _, option := range field.Options {
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
func (v *Validator) validateFile(field *schema.Field, value any) []string {
	var errors []string

	// File validation would typically happen at upload time
	// This is a placeholder for file metadata validation

	str, ok := value.(string)
	if !ok {
		errors = append(errors, "Must be a string")
		return errors
	}

	// Basic file path/name validation
	if str == "" {
		errors = append(errors, "File path cannot be empty")
	}

	return errors
}

// Custom validation
func (v *Validator) validateCustom(field *schema.Field, value any) []string {
	var errors []string

	// TODO: Implement custom validation expression evaluation
	// This would parse and evaluate the custom validation expression
	// For now, this is a placeholder

	return errors
}

// Check uniqueness in database
func (v *Validator) checkUniqueness(ctx context.Context, field *schema.Field, value any) (bool, error) {
	if v.db == nil {
		return true, nil // No database, assume unique
	}

	// Extract table name from field config or use schema context
	tableName := "unknown_table" // Would come from schema context
	if field.Config != nil {
		if table, ok := field.Config["table"].(string); ok {
			tableName = table
		}
	}

	exists, err := v.db.Exists(ctx, tableName, field.Name, value)
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
func (v *Validator) cleanValue(field *schema.Field, value any) any {
	if value == nil {
		return nil
	}

	switch field.Type {
	case schema.FieldText, schema.FieldTextarea, schema.FieldEmail, schema.FieldPassword:
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str)
		}
	case schema.FieldNumber, schema.FieldCurrency:
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

// Utility function to create a validation error
func NewValidationError(field, message, code string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}
