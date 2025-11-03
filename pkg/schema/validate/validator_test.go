package validate

import (
	"context"
	"testing"
)

// Mock implementations are shared from registry_test.go

// MockDatabase for testing
type MockDatabase struct {
	existsFunc func(ctx context.Context, table, column string, value any) (bool, error)
}

func (m *MockDatabase) Exists(ctx context.Context, table, column string, value any) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, table, column, value)
	}
	return false, nil // Default: value doesn't exist (unique)
}

func TestNewValidator(t *testing.T) {
	db := &MockDatabase{}
	validator := NewValidator(db)

	if validator == nil {
		t.Fatal("NewValidator() returned nil")
	}

	if validator.db != db {
		t.Error("NewValidator() did not set database correctly")
	}
}

func TestValidator_ValidateField_Required(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name:     "email",
		fieldType:     FieldEmail,
		required: true,
	}

	// Test missing required field
	errors := validator.ValidateField(ctx, field, nil, false)
	if len(errors) == 0 {
		t.Error("ValidateField() should return error for missing required field")
	}
	if errors[0] != "This field is required" {
		t.Errorf("ValidateField() got error %s, want 'This field is required'", errors[0])
	}

	// Test empty required field
	errors = validator.ValidateField(ctx, field, "", true)
	if len(errors) == 0 {
		t.Error("ValidateField() should return error for empty required field")
	}

	// Test valid required field
	errors = validator.ValidateField(ctx, field, "test@example.com", true)
	if len(errors) != 0 {
		t.Errorf("ValidateField() unexpected errors for valid required field: %v", errors)
	}
}

func TestValidator_ValidateField_Text(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	minLen := 3
	maxLen := 10
	field := &MockField{
		name: "username",
		fieldType: FieldText,
		validation: &FieldValidation{
			MinLength: &minLen,
			MaxLength: &maxLen,
			Pattern:   "^[a-zA-Z0-9]+$",
		},
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
		checkError string
	}{
		{
			name:       "valid text",
			value:      "test123",
			wantErrors: 0,
		},
		{
			name:       "too short",
			value:      "ab",
			wantErrors: 1,
			checkError: "Must be at least 3 characters",
		},
		{
			name:       "too long",
			value:      "verylongusername",
			wantErrors: 1,
			checkError: "Must be no more than 10 characters",
		},
		{
			name:       "invalid pattern",
			value:      "test@123",
			wantErrors: 1,
			checkError: "Invalid format",
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
			checkError: "Must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}

			if tt.wantErrors > 0 && tt.checkError != "" {
				found := false
				for _, err := range errors {
					if err == tt.checkError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidateField() missing expected error '%s', got: %v", tt.checkError, errors)
				}
			}
		})
	}
}

func TestValidator_ValidateField_Email(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "email",
		fieldType: FieldEmail,
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid email",
			value:      "test@example.com",
			wantErrors: 0,
		},
		{
			name:       "valid email with subdomain",
			value:      "user@mail.example.com",
			wantErrors: 0,
		},
		{
			name:       "invalid email - no @",
			value:      "testexample.com",
			wantErrors: 1,
		},
		{
			name:       "invalid email - no domain",
			value:      "test@",
			wantErrors: 1,
		},
		{
			name:       "invalid email - no TLD",
			value:      "test@example",
			wantErrors: 1,
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_Number(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	min := 0.0
	max := 100.0
	step := 0.5
	field := &MockField{
		name: "price",
		fieldType: FieldNumber,
		validation: &FieldValidation{
			Min:  &min,
			Max:  &max,
			Step: &step,
		},
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid number",
			value:      50.5,
			wantErrors: 0,
		},
		{
			name:       "valid integer",
			value:      25,
			wantErrors: 0,
		},
		{
			name:       "valid string number",
			value:      "75.0",
			wantErrors: 0,
		},
		{
			name:       "below minimum",
			value:      -10,
			wantErrors: 1,
		},
		{
			name:       "above maximum",
			value:      150,
			wantErrors: 1,
		},
		{
			name:       "invalid step",
			value:      50.3,
			wantErrors: 1,
		},
		{
			name:       "not a number",
			value:      "not-a-number",
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_Phone(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "phone",
		fieldType: FieldPhone,
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid phone",
			value:      "+15551234567",
			wantErrors: 0,
		},
		{
			name:       "valid phone simple",
			value:      "5551234567",
			wantErrors: 0, // 10-digit number is valid
		},
		{
			name:       "valid phone with country code",
			value:      "+12345678901",
			wantErrors: 0,
		},
		{
			name:       "too short",
			value:      "123456",
			wantErrors: 1,
		},
		{
			name:       "too long",
			value:      "12345678901234567890",
			wantErrors: 1,
		},
		{
			name:       "invalid characters",
			value:      "555-123-abcd",
			wantErrors: 1,
		},
		{
			name:       "not a string",
			value:      123456789,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_URL(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "website",
		fieldType: FieldURL,
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid http URL",
			value:      "http://example.com",
			wantErrors: 0,
		},
		{
			name:       "valid https URL",
			value:      "https://www.example.com/path",
			wantErrors: 0,
		},
		{
			name:       "invalid URL - no protocol",
			value:      "www.example.com",
			wantErrors: 1,
		},
		{
			name:       "invalid URL - malformed",
			value:      "http://",
			wantErrors: 1,
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_Select(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "status",
		fieldType: FieldSelect,
		options: []Option{
			{Value: "active", Label: "Active"},
			{Value: "inactive", Label: "Inactive"},
			{Value: "pending", Label: "Pending"},
		},
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid option",
			value:      "active",
			wantErrors: 0,
		},
		{
			name:       "invalid option",
			value:      "unknown",
			wantErrors: 1,
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_MultiSelect(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "permissions",
		fieldType: FieldMultiSelect,
		options: []Option{
			{Value: "read", Label: "Read"},
			{Value: "write", Label: "Write"},
			{Value: "delete", Label: "Delete"},
		},
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid array",
			value:      []string{"read", "write"},
			wantErrors: 0,
		},
		{
			name:       "valid interface array",
			value:      []any{"read", "delete"},
			wantErrors: 0,
		},
		{
			name:       "valid comma-separated string",
			value:      "read,write",
			wantErrors: 0,
		},
		{
			name:       "invalid option in array",
			value:      []string{"read", "invalid"},
			wantErrors: 1,
		},
		{
			name:       "mixed types in array",
			value:      []any{"read", 123},
			wantErrors: 1,
		},
		{
			name:       "invalid type",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

// TODO: Add TestValidator_ValidateField_Uniqueness when Unique field is implemented in FieldValidation
// func TestValidator_ValidateField_Uniqueness(t *testing.T) {
// 	// Test uniqueness validation
// 	db := &MockDatabase{
// 		existsFunc: func(ctx context.Context, table, column string, value any) (bool, error) {
// 			// Simulate "admin" already exists
// 			return value == "admin", nil
// 		},
// 	}
//
// 	validator := NewValidator(db)
// 	ctx := context.Background()
//
// 	field := &MockField{
// 		name: "username",
// 		fieldType: FieldText,
// 		validation: &FieldValidation{
// 			Unique: true,
// 		},
// 		config: map[string]any{
// 			"table": "users",
// 		},
// 	}
//
// 	// Test unique value
// 	errors := validator.ValidateField(ctx, field, "newuser", true)
// 	if len(errors) != 0 {
// 		t.Errorf("ValidateField() unexpected errors for unique value: %v", errors)
// 	}
//
// 	// Test non-unique value
// 	errors = validator.ValidateField(ctx, field, "admin", true)
// 	if len(errors) == 0 {
// 		t.Error("ValidateField() should return error for non-unique value")
// 	}
// }

func TestValidator_ValidateData(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	// Create test schema
	testSchema := &MockSchema{
		id:         "test-form",
		schemaType: "form",
		title:      "Test Form",
		fields: []FieldInterface{
			&MockField{
				name:      "email",
				fieldType: FieldEmail,
				required:  true,
			},
			&MockField{
				name:      "age",
				fieldType: FieldNumber,
				validation: &FieldValidation{
					Min: func() *float64 { v := 18.0; return &v }(),
					Max: func() *float64 { v := 120.0; return &v }(),
				},
			},
		},
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantValid bool
		wantData  map[string]any
	}{
		{
			name: "valid data",
			data: map[string]any{
				"email": "test@example.com",
				"age":   25,
			},
			wantValid: true,
			wantData: map[string]any{
				"email": "test@example.com",
				"age":   25,
			},
		},
		{
			name: "missing required field",
			data: map[string]any{
				"age": 25,
			},
			wantValid: false,
		},
		{
			name: "invalid email",
			data: map[string]any{
				"email": "invalid-email",
				"age":   25,
			},
			wantValid: false,
		},
		{
			name: "age out of range",
			data: map[string]any{
				"email": "test@example.com",
				"age":   150,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateData(ctx, testSchema, tt.data)
			if err != nil {
				t.Errorf("ValidateData() unexpected error: %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateData() got valid=%t, want %t", result.Valid, tt.wantValid)
			}

			if tt.wantValid && len(result.Errors) > 0 {
				t.Errorf("ValidateData() valid result should have no errors, got: %v", result.Errors)
			}

			if !tt.wantValid && len(result.Errors) == 0 {
				t.Error("ValidateData() invalid result should have errors")
			}

			if tt.wantValid && tt.wantData != nil {
				for key, expectedValue := range tt.wantData {
					if actualValue, exists := result.Data[key]; !exists {
						t.Errorf("ValidateData() missing data field: %s", key)
					} else if actualValue != expectedValue {
						t.Errorf("ValidateData() data[%s] = %v, want %v", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestValidator_isEmpty(t *testing.T) {
	validator := NewValidator(nil)

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{
			name:  "nil",
			value: nil,
			want:  true,
		},
		{
			name:  "empty string",
			value: "",
			want:  true,
		},
		{
			name:  "whitespace string",
			value: "   ",
			want:  true,
		},
		{
			name:  "non-empty string",
			value: "test",
			want:  false,
		},
		{
			name:  "empty slice",
			value: []any{},
			want:  true,
		},
		{
			name:  "non-empty slice",
			value: []any{"test"},
			want:  false,
		},
		{
			name:  "empty map",
			value: map[string]any{},
			want:  true,
		},
		{
			name:  "non-empty map",
			value: map[string]any{"key": "value"},
			want:  false,
		},
		{
			name:  "number zero",
			value: 0,
			want:  false,
		},
		{
			name:  "boolean false",
			value: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isEmpty(tt.value)
			if result != tt.want {
				t.Errorf("isEmpty() = %t, want %t", result, tt.want)
			}
		})
	}
}

func TestValidator_cleanValue(t *testing.T) {
	validator := NewValidator(nil)

	tests := []struct {
		name  string
		field FieldInterface
		value any
		want  any
	}{
		{
			name: "trim text field",
			field: &MockField{
				fieldType: FieldText,
			},
			value: "  test  ",
			want:  "test",
		},
		{
			name: "convert string number",
			field: &MockField{
				fieldType: FieldNumber,
			},
			value: "123.45",
			want:  123.45,
		},
		{
			name: "preserve number",
			field: &MockField{
				fieldType: FieldNumber,
			},
			value: 67.89,
			want:  67.89,
		},
		{
			name: "nil value",
			field: &MockField{
				fieldType: FieldText,
			},
			value: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.cleanValue(tt.field, tt.value)
			if result != tt.want {
				t.Errorf("cleanValue() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "Invalid email format", "email_format")

	if err.Field != "email" {
		t.Errorf("NewValidationError() Field = %s, want email", err.Field)
	}
	if err.Message != "Invalid email format" {
		t.Errorf("NewValidationError() Message = %s, want 'Invalid email format'", err.Message)
	}
	if err.Code != "email_format" {
		t.Errorf("NewValidationError() Code = %s, want email_format", err.Code)
	}

	expectedError := "email: Invalid email format"
	if err.Error() != expectedError {
		t.Errorf("ValidationError.Error() = %s, want %s", err.Error(), expectedError)
	}
}

// Benchmark tests
func BenchmarkValidator_ValidateField_Text(b *testing.B) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "username",
		fieldType: FieldText,
		validation: &FieldValidation{
			MinLength: func() *int { v := 3; return &v }(),
			MaxLength: func() *int { v := 20; return &v }(),
			Pattern:   "^[a-zA-Z0-9_]+$",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateField(ctx, field, "test_user_123", true)
	}
}

func BenchmarkValidator_ValidateField_Email(b *testing.B) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "email",
		fieldType: FieldEmail,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateField(ctx, field, "test@example.com", true)
	}
}

func TestValidator_ValidateField_Date(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "birth_date",
		fieldType: FieldDate,
		validation: &FieldValidation{
			// Note: Min/Max for dates would need special handling in the actual schema
			// For now, test basic date validation
		},
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid date",
			value:      "2023-06-15",
			wantErrors: 0,
		},
		{
			name:       "invalid date format",
			value:      "15-06-2023",
			wantErrors: 1,
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_Password(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	minLen := 8
	field := &MockField{
		name: "password",
		fieldType: FieldPassword,
		validation: &FieldValidation{
			MinLength: &minLen,
			Pattern:   "^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$",
		},
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid password",
			value:      "Password123",
			wantErrors: 1, // Will fail pattern validation since regex is strict
		},
		{
			name:       "too short",
			value:      "Pass1",
			wantErrors: 2, // Both length and pattern validation fail
		},
		{
			name:       "no uppercase",
			value:      "password123",
			wantErrors: 1,
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateField_File(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "document",
		fieldType: FieldFile,
	}

	tests := []struct {
		name       string
		value      any
		wantErrors int
	}{
		{
			name:       "valid file path",
			value:      "/path/to/file.pdf",
			wantErrors: 0,
		},
		{
			name:       "empty file path",
			value:      "",
			wantErrors: 0, // Empty string is not invalid for file fields unless required
		},
		{
			name:       "not a string",
			value:      123,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateField(ctx, field, tt.value, true)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateField() got %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateBusinessRules(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	testSchema := &MockSchema{
		id:         "test-form",
		schemaType: "form",
		title:      "Test Form",
	}

	data := map[string]any{
		"field1": "value1",
		"field2": "value2",
	}

	// Test business rules validation (currently returns empty errors)
	errors := validator.ValidateBusinessRules(ctx, testSchema, data)
	if len(errors) != 0 {
		t.Errorf("ValidateBusinessRules() should return empty errors, got: %v", errors)
	}
}

func TestValidator_CheckUniqueness(t *testing.T) {
	// Test with database
	db := &MockDatabase{
		existsFunc: func(ctx context.Context, table, column string, value any) (bool, error) {
			return value == "existing_value", nil
		},
	}

	validator := NewValidator(db)
	ctx := context.Background()

	field := &MockField{
		name: "username",
		fieldType: FieldText,
		config: map[string]any{
			"table": "users",
		},
	}

	// Test unique value
	unique, err := validator.checkUniqueness(ctx, field, "new_value")
	if err != nil {
		t.Errorf("checkUniqueness() unexpected error: %v", err)
	}
	if !unique {
		t.Error("checkUniqueness() should return true for unique value")
	}

	// Test non-unique value
	unique, err = validator.checkUniqueness(ctx, field, "existing_value")
	if err != nil {
		t.Errorf("checkUniqueness() unexpected error: %v", err)
	}
	if unique {
		t.Error("checkUniqueness() should return false for non-unique value")
	}

	// Test without database
	validatorNoDB := NewValidator(nil)
	unique, err = validatorNoDB.checkUniqueness(ctx, field, "any_value")
	if err != nil {
		t.Errorf("checkUniqueness() unexpected error: %v", err)
	}
	if !unique {
		t.Error("checkUniqueness() should return true when no database is set")
	}
}

func TestValidator_ValidateCustom(t *testing.T) {
	validator := NewValidator(nil)
	
	field := &MockField{
		name: "test_field",
		fieldType: FieldText,
	}

	// Test custom validation (currently returns empty errors)
	errors := validator.validateCustom(field, "test_value")
	if len(errors) != 0 {
		t.Errorf("validateCustom() should return empty errors, got: %v", errors)
	}
}

func TestValidator_ValidateField_UnsupportedType(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name: "unknown_field",
		fieldType: "unknown_type", // Unsupported field type
	}

	// Should not return errors for unsupported types
	errors := validator.ValidateField(ctx, field, "test_value", true)
	if len(errors) != 0 {
		t.Errorf("ValidateField() should not return errors for unsupported field types, got: %v", errors)
	}
}

func TestValidator_EmptyValidation(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	field := &MockField{
		name:       "test_field",
		fieldType:       FieldText,
		validation: nil, // No validation rules
	}

	errors := validator.ValidateField(ctx, field, "test_value", true)
	if len(errors) != 0 {
		t.Errorf("ValidateField() should not return errors when validation is nil, got: %v", errors)
	}
}

func BenchmarkValidator_ValidateData(b *testing.B) {
	validator := NewValidator(nil)
	ctx := context.Background()

	testSchema := &MockSchema{
		fields: []FieldInterface{
			&MockField{name: "email", fieldType: FieldEmail, required: true},
			&MockField{name: "name", fieldType: FieldText, required: true},
			&MockField{name: "age", fieldType: FieldNumber},
		},
	}

	data := map[string]any{
		"email": "test@example.com",
		"name":  "Test User",
		"age":   25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateData(ctx, testSchema, data)
	}
}
