package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRules represents validation rules for configuration values
type ValidationRules map[string]any

// NewValidationRules creates new validation rules
func NewValidationRules() ValidationRules {
	return make(ValidationRules)
}

// SetRequired sets whether the configuration value is required
func (vr ValidationRules) SetRequired(required bool) {
	vr["required"] = required
}

// IsRequired returns true if the configuration value is required
func (vr ValidationRules) IsRequired() bool {
	if required, exists := vr["required"]; exists {
		if req, ok := required.(bool); ok {
			return req
		}
	}
	return false
}

// SetMinValue sets the minimum value for numeric configurations
func (vr ValidationRules) SetMinValue(min float64) {
	vr["min_value"] = min
}

// GetMinValue returns the minimum value constraint
func (vr ValidationRules) GetMinValue() (float64, bool) {
	if min, exists := vr["min_value"]; exists {
		switch v := min.(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case string:
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				return val, true
			}
		}
	}
	return 0, false
}

// SetMaxValue sets the maximum value for numeric configurations
func (vr ValidationRules) SetMaxValue(max float64) {
	vr["max_value"] = max
}

// GetMaxValue returns the maximum value constraint
func (vr ValidationRules) GetMaxValue() (float64, bool) {
	if max, exists := vr["max_value"]; exists {
		switch v := max.(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case string:
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				return val, true
			}
		}
	}
	return 0, false
}

// SetMinLength sets the minimum length for string configurations
func (vr ValidationRules) SetMinLength(minLen int) {
	vr["min_length"] = minLen
}

// GetMinLength returns the minimum length constraint
func (vr ValidationRules) GetMinLength() (int, bool) {
	if minLen, exists := vr["min_length"]; exists {
		switch v := minLen.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case string:
			if val, err := strconv.Atoi(v); err == nil {
				return val, true
			}
		}
	}
	return 0, false
}

// SetMaxLength sets the maximum length for string configurations
func (vr ValidationRules) SetMaxLength(maxLen int) {
	vr["max_length"] = maxLen
}

// GetMaxLength returns the maximum length constraint
func (vr ValidationRules) GetMaxLength() (int, bool) {
	if maxLen, exists := vr["max_length"]; exists {
		switch v := maxLen.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case string:
			if val, err := strconv.Atoi(v); err == nil {
				return val, true
			}
		}
	}
	return 0, false
}

// SetPattern sets a regex pattern for string configurations
func (vr ValidationRules) SetPattern(pattern string) {
	vr["pattern"] = pattern
}

// GetPattern returns the regex pattern constraint
func (vr ValidationRules) GetPattern() (string, bool) {
	if pattern, exists := vr["pattern"]; exists {
		if pat, ok := pattern.(string); ok {
			return pat, true
		}
	}
	return "", false
}

// SetAllowedValues sets the list of allowed values
func (vr ValidationRules) SetAllowedValues(values []any) {
	vr["allowed_values"] = values
}

// GetAllowedValues returns the list of allowed values
func (vr ValidationRules) GetAllowedValues() ([]any, bool) {
	if values, exists := vr["allowed_values"]; exists {
		if vals, ok := values.([]any); ok {
			return vals, true
		}
	}
	return nil, false
}

// SetCustomRule sets a custom validation rule
func (vr ValidationRules) SetCustomRule(name string, value any) {
	vr[name] = value
}

// GetCustomRule returns a custom validation rule
func (vr ValidationRules) GetCustomRule(name string) (any, bool) {
	value, exists := vr[name]
	return value, exists
}

// Validate validates a configuration value against the rules
func (vr ValidationRules) Validate(value ConfigValue) error {
	// Check required constraint
	if vr.IsRequired() && value.IsEmpty() {
		return fmt.Errorf("configuration value is required")
	}

	// Skip other validations if value is empty and not required
	if value.IsEmpty() {
		return nil
	}

	// Validate based on data type
	switch value.DataType {
	case DataTypeString:
		return vr.validateString(value)
	case DataTypeInteger:
		return vr.validateInteger(value)
	case DataTypeDecimal:
		return vr.validateDecimal(value)
	case DataTypeBoolean:
		return vr.validateBoolean(value)
	case DataTypeJSON:
		return vr.validateJSON(value)
	}

	return nil
}

// validateString validates string values
func (vr ValidationRules) validateString(value ConfigValue) error {
	strVal, err := value.AsString()
	if err != nil {
		return err
	}

	// Check minimum length
	if minLen, exists := vr.GetMinLength(); exists {
		if len(strVal) < minLen {
			return fmt.Errorf("string length %d is less than minimum %d", len(strVal), minLen)
		}
	}

	// Check maximum length
	if maxLen, exists := vr.GetMaxLength(); exists {
		if len(strVal) > maxLen {
			return fmt.Errorf("string length %d exceeds maximum %d", len(strVal), maxLen)
		}
	}

	// Check pattern
	if pattern, exists := vr.GetPattern(); exists {
		matched, err := regexp.MatchString(pattern, strVal)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("string does not match required pattern: %s", pattern)
		}
	}

	// Check allowed values
	if allowedValues, exists := vr.GetAllowedValues(); exists {
		allowed := false
		for _, allowedVal := range allowedValues {
			if strVal == fmt.Sprintf("%v", allowedVal) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("value '%s' is not in allowed values list", strVal)
		}
	}

	return nil
}

// validateInteger validates integer values
func (vr ValidationRules) validateInteger(value ConfigValue) error {
	intVal, err := value.AsInt()
	if err != nil {
		return err
	}

	floatVal := float64(intVal)

	// Check minimum value
	if minVal, exists := vr.GetMinValue(); exists {
		if floatVal < minVal {
			return fmt.Errorf("value %d is less than minimum %.0f", intVal, minVal)
		}
	}

	// Check maximum value
	if maxVal, exists := vr.GetMaxValue(); exists {
		if floatVal > maxVal {
			return fmt.Errorf("value %d exceeds maximum %.0f", intVal, maxVal)
		}
	}

	// Check allowed values
	if allowedValues, exists := vr.GetAllowedValues(); exists {
		allowed := false
		for _, allowedVal := range allowedValues {
			// Convert allowed value to string first, then parse to int
			strVal := fmt.Sprintf("%v", allowedVal)
			intAllowedVal, err := strconv.Atoi(strVal)
			if err != nil {
				// Handle error if the value can't be converted to int
				continue // or return error depending on your requirements
			}
			if intVal == intAllowedVal {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("value %d is not in allowed values list", intVal)
		}
	}

	return nil
}

// validateDecimal validates decimal values
func (vr ValidationRules) validateDecimal(value ConfigValue) error {
	decimalVal, err := value.AsDecimal()
	if err != nil {
		return err
	}

	// Check minimum value
	if minVal, exists := vr.GetMinValue(); exists {
		if decimalVal < minVal {
			return fmt.Errorf("value %.2f is less than minimum %.2f", decimalVal, minVal)
		}
	}

	// Check maximum value
	if maxVal, exists := vr.GetMaxValue(); exists {
		if decimalVal > maxVal {
			return fmt.Errorf("value %.2f exceeds maximum %.2f", decimalVal, maxVal)
		}
	}

	// Check allowed values
	if allowedValues, exists := vr.GetAllowedValues(); exists {
		allowed := false
		for _, allowedVal := range allowedValues {
			if decimalVal == allowedVal {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("value %.2f is not in allowed values list", decimalVal)
		}
	}

	return nil
}

// validateBoolean validates boolean values
func (vr ValidationRules) validateBoolean(value ConfigValue) error {
	_, err := value.AsBool()
	if err != nil {
		return err
	}

	// Booleans don't have complex validation rules beyond type checking
	return nil
}

// validateJSON validates JSON values
func (vr ValidationRules) validateJSON(value ConfigValue) error {
	jsonVal, err := value.AsJSON()
	if err != nil {
		return err
	}

	// Check for required keys in JSON schema validation
	if requiredKeys, exists := vr.GetCustomRule("required_keys"); exists {
		if keys, ok := requiredKeys.([]string); ok {
			for _, key := range keys {
				if _, exists := jsonVal[key]; !exists {
					return fmt.Errorf("required JSON key '%s' is missing", key)
				}
			}
		}
	}

	// Additional JSON schema validation could be implemented here
	// For now, basic structure validation is sufficient

	return nil
}

// ToJSON converts validation rules to JSON
func (vr ValidationRules) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]any(vr))
}

// FromJSON creates validation rules from JSON
func (vr ValidationRules) FromJSON(data []byte) error {
	var rules map[string]any
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	for key, value := range rules {
		vr[key] = value
	}

	return nil
}

// Clone creates a deep copy of the validation rules
func (vr ValidationRules) Clone() ValidationRules {
	clone := make(ValidationRules)
	for key, value := range vr {
		clone[key] = value
	}
	return clone
}

// IsEmpty returns true if there are no validation rules
func (vr ValidationRules) IsEmpty() bool {
	return len(vr) == 0
}

// String returns a string representation of the validation rules
func (vr ValidationRules) String() string {
	if len(vr) == 0 {
		return "ValidationRules{}"
	}

	var rules []string
	for key, value := range vr {
		rules = append(rules, fmt.Sprintf("%s: %v", key, value))
	}

	return fmt.Sprintf("ValidationRules{%s}", strings.Join(rules, ", "))
}

// Predefined validation rule builders for common scenarios

// RequiredString creates validation rules for a required string with optional length constraints
func RequiredString(minLength, maxLength int) ValidationRules {
	rules := NewValidationRules()
	rules.SetRequired(true)
	if minLength > 0 {
		rules.SetMinLength(minLength)
	}
	if maxLength > 0 {
		rules.SetMaxLength(maxLength)
	}
	return rules
}

// OptionalString creates validation rules for an optional string with optional length constraints
func OptionalString(minLength, maxLength int) ValidationRules {
	rules := NewValidationRules()
	rules.SetRequired(false)
	if minLength > 0 {
		rules.SetMinLength(minLength)
	}
	if maxLength > 0 {
		rules.SetMaxLength(maxLength)
	}
	return rules
}

// RequiredInteger creates validation rules for a required integer with optional range constraints
func RequiredInteger(minValue, maxValue *int) ValidationRules {
	rules := NewValidationRules()
	rules.SetRequired(true)
	if minValue != nil {
		rules.SetMinValue(float64(*minValue))
	}
	if maxValue != nil {
		rules.SetMaxValue(float64(*maxValue))
	}
	return rules
}

// OptionalInteger creates validation rules for an optional integer with optional range constraints
func OptionalInteger(minValue, maxValue *int) ValidationRules {
	rules := NewValidationRules()
	rules.SetRequired(false)
	if minValue != nil {
		rules.SetMinValue(float64(*minValue))
	}
	if maxValue != nil {
		rules.SetMaxValue(float64(*maxValue))
	}
	return rules
}

// RequiredEnum creates validation rules for a required enumeration
func RequiredEnum(allowedValues ...string) ValidationRules {
	rules := NewValidationRules()
	rules.SetRequired(true)

	values := make([]any, len(allowedValues))
	for i, val := range allowedValues {
		values[i] = val
	}
	rules.SetAllowedValues(values)

	return rules
}

// OptionalEnum creates validation rules for an optional enumeration
func OptionalEnum(allowedValues ...string) ValidationRules {
	rules := NewValidationRules()
	rules.SetRequired(false)

	values := make([]any, len(allowedValues))
	for i, val := range allowedValues {
		values[i] = val
	}
	rules.SetAllowedValues(values)

	return rules
}
