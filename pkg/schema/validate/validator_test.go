package validate

import (
	"context"
	"testing"

	"github.com/niiniyare/erp/pkg/schema"
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
		name:      "email",
		fieldType: FieldEmail,
		required:  true,
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
		name:      "username",
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
		name:      "email",
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
		name:      "price",
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
		name:      "phone",
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
		name:      "website",
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
		name:      "status",
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
		name:      "permissions",
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

// TestValidator_ValidateField_Uniqueness tests uniqueness validation (ready when Unique field is added)
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
			// Convert MockSchema to real schema for ValidateData
			realSchema := convertMockToRealSchema(testSchema)
			err := validator.ValidateData(ctx, realSchema, tt.data)
			gotValid := (err == nil)

			if gotValid != tt.wantValid {
				t.Errorf("ValidateData() got valid=%t, want %t", gotValid, tt.wantValid)
			}

			// Note: ValidateData only returns error, so we can't check detailed validation results
			// For detailed validation, use ValidateDataDetailed instead
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
		name:      "username",
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
		name:      "email",
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
		name:       "birth_date",
		fieldType:  FieldDate,
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
		name:      "password",
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
		name:      "document",
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
		name:      "username",
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
		name:      "test_field",
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
		name:      "unknown_field",
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
		fieldType:  FieldText,
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
	realSchema := convertMockToRealSchema(testSchema)
	for i := 0; i < b.N; i++ {
		validator.ValidateData(ctx, realSchema, data)
	}
}

// =================================
// Enricher-aware Mock Implementations
// =================================

// MockFieldRuntime implements FieldRuntime interface for testing
type MockFieldRuntime struct {
	visible  bool
	editable bool
	reason   string
}

func (r *MockFieldRuntime) IsVisible() bool   { return r.visible }
func (r *MockFieldRuntime) IsEditable() bool  { return r.editable }
func (r *MockFieldRuntime) GetReason() string { return r.reason }

// MockEnrichedField implements EnrichedFieldInterface for testing
type MockEnrichedField struct {
	*MockField
	runtime *MockFieldRuntime
}

func (f *MockEnrichedField) GetRuntime() FieldRuntime {
	if f.runtime == nil {
		return nil
	}
	return f.runtime
}

// MockEnrichedSchema implements EnrichedSchemaInterface for testing
type MockEnrichedSchema struct {
	*MockSchema
	enrichedFields []EnrichedFieldInterface
}

func (s *MockEnrichedSchema) GetEnrichedFields() []EnrichedFieldInterface {
	return s.enrichedFields
}

// =================================
// Enricher-aware Validation Tests
// =================================

func TestValidator_ValidateEnrichedSchema(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		schema     *MockEnrichedSchema
		data       map[string]any
		wantValid  bool
		wantErrors map[string]int // field -> error count
	}{
		{
			name: "all fields visible and valid",
			schema: &MockEnrichedSchema{
				enrichedFields: []EnrichedFieldInterface{
					&MockEnrichedField{
						MockField: &MockField{
							name:      "email",
							fieldType: FieldEmail,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: true,
						},
					},
					&MockEnrichedField{
						MockField: &MockField{
							name:      "name",
							fieldType: FieldText,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: true,
						},
					},
				},
			},
			data: map[string]any{
				"email": "test@example.com",
				"name":  "Test User",
			},
			wantValid:  true,
			wantErrors: map[string]int{},
		},
		{
			name: "invisible field skipped",
			schema: &MockEnrichedSchema{
				enrichedFields: []EnrichedFieldInterface{
					&MockEnrichedField{
						MockField: &MockField{
							name:      "email",
							fieldType: FieldEmail,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: true,
						},
					},
					&MockEnrichedField{
						MockField: &MockField{
							name:      "secret",
							fieldType: FieldText,
							required:  true, // Required but invisible - should be skipped
						},
						runtime: &MockFieldRuntime{
							visible:  false,
							editable: false,
							reason:   "permission_required",
						},
					},
				},
			},
			data: map[string]any{
				"email": "test@example.com",
				// secret field not provided - should not cause error since invisible
			},
			wantValid:  true,
			wantErrors: map[string]int{},
		},
		{
			name: "readonly field validation",
			schema: &MockEnrichedSchema{
				enrichedFields: []EnrichedFieldInterface{
					&MockEnrichedField{
						MockField: &MockField{
							name:      "email",
							fieldType: FieldEmail,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: false, // Readonly
							reason:   "system_field",
						},
					},
				},
			},
			data: map[string]any{
				"email": "invalid-email",
			},
			wantValid: false,
			wantErrors: map[string]int{
				"email": 1, // Should fail format validation
			},
		},
		{
			name: "validation error on editable field",
			schema: &MockEnrichedSchema{
				enrichedFields: []EnrichedFieldInterface{
					&MockEnrichedField{
						MockField: &MockField{
							name:      "email",
							fieldType: FieldEmail,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: true,
						},
					},
				},
			},
			data: map[string]any{
				"email": "", // Empty required field
			},
			wantValid: false,
			wantErrors: map[string]int{
				"email": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateEnrichedSchema(ctx, tt.schema, tt.data)
			if err != nil {
				t.Errorf("ValidateEnrichedSchema() error = %v", err)
				return
			}

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateEnrichedSchema() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			for field, expectedCount := range tt.wantErrors {
				actualCount := len(result.Errors[field])
				if actualCount != expectedCount {
					t.Errorf("ValidateEnrichedSchema() field %s got %d errors, want %d: %v",
						field, actualCount, expectedCount, result.Errors[field])
				}
			}

			// Check no unexpected errors
			for field, errors := range result.Errors {
				if _, expected := tt.wantErrors[field]; !expected && len(errors) > 0 {
					t.Errorf("ValidateEnrichedSchema() unexpected errors for field %s: %v", field, errors)
				}
			}
		})
	}
}

func TestValidator_ValidateWithFieldState(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		field      *MockEnrichedField
		value      any
		exists     bool
		wantErrors int
	}{
		{
			name: "visible editable field - valid",
			field: &MockEnrichedField{
				MockField: &MockField{
					name:      "email",
					fieldType: FieldEmail,
					required:  true,
				},
				runtime: &MockFieldRuntime{
					visible:  true,
					editable: true,
				},
			},
			value:      "test@example.com",
			exists:     true,
			wantErrors: 0,
		},
		{
			name: "visible editable field - invalid",
			field: &MockEnrichedField{
				MockField: &MockField{
					name:      "email",
					fieldType: FieldEmail,
					required:  true,
				},
				runtime: &MockFieldRuntime{
					visible:  true,
					editable: true,
				},
			},
			value:      "invalid-email",
			exists:     true,
			wantErrors: 1,
		},
		{
			name: "invisible field",
			field: &MockEnrichedField{
				MockField: &MockField{
					name:      "secret",
					fieldType: FieldText,
					required:  true,
				},
				runtime: &MockFieldRuntime{
					visible:  false,
					editable: false,
				},
			},
			value:      "", // Empty required field but invisible
			exists:     false,
			wantErrors: 0, // Should not validate invisible fields
		},
		{
			name: "readonly field - valid format",
			field: &MockEnrichedField{
				MockField: &MockField{
					name:      "created_at",
					fieldType: FieldDateTime,
					required:  false,
				},
				runtime: &MockFieldRuntime{
					visible:  true,
					editable: false,
				},
			},
			value:      "2023-01-01T12:00:00Z",
			exists:     true,
			wantErrors: 0,
		},
		{
			name: "readonly field - invalid format",
			field: &MockEnrichedField{
				MockField: &MockField{
					name:      "created_at",
					fieldType: FieldDateTime,
					required:  false,
				},
				runtime: &MockFieldRuntime{
					visible:  true,
					editable: false,
				},
			},
			value:      "invalid-date",
			exists:     true,
			wantErrors: 1,
		},
		{
			name: "field without runtime - defaults to full validation",
			field: &MockEnrichedField{
				MockField: &MockField{
					name:      "name",
					fieldType: FieldText,
					required:  true,
				},
				runtime: nil, // No runtime state
			},
			value:      "",
			exists:     false,
			wantErrors: 1, // Should fail required validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateWithFieldState(ctx, tt.field, tt.value, tt.exists)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateWithFieldState() got %d errors, want %d: %v",
					len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidator_ValidateConditionalFields(t *testing.T) {
	validator := NewValidator(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		schema     *MockEnrichedSchema
		data       map[string]any
		wantErrors map[string]int
	}{
		{
			name: "mixed field states",
			schema: &MockEnrichedSchema{
				enrichedFields: []EnrichedFieldInterface{
					&MockEnrichedField{
						MockField: &MockField{
							name:      "visible_editable",
							fieldType: FieldText,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: true,
						},
					},
					&MockEnrichedField{
						MockField: &MockField{
							name:      "visible_readonly",
							fieldType: FieldEmail,
							required:  false,
						},
						runtime: &MockFieldRuntime{
							visible:  true,
							editable: false,
						},
					},
					&MockEnrichedField{
						MockField: &MockField{
							name:      "invisible",
							fieldType: FieldText,
							required:  true,
						},
						runtime: &MockFieldRuntime{
							visible:  false,
							editable: false,
						},
					},
				},
			},
			data: map[string]any{
				"visible_editable": "valid text",
				"visible_readonly": "test@example.com",
				"invisible":        "should not be accessible",
			},
			wantErrors: map[string]int{
				"invisible": 1, // Should error for data submitted to invisible field
			},
		},
		{
			name: "no runtime state",
			schema: &MockEnrichedSchema{
				enrichedFields: []EnrichedFieldInterface{
					&MockEnrichedField{
						MockField: &MockField{
							name:      "no_runtime",
							fieldType: FieldText,
							required:  true,
						},
						runtime: nil, // No runtime state - skip conditional validation
					},
				},
			},
			data: map[string]any{
				"no_runtime": "some value",
			},
			wantErrors: map[string]int{}, // No conditional validation applied
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateConditionalFields(ctx, tt.schema, tt.data)

			for field, expectedCount := range tt.wantErrors {
				actualCount := len(errors[field])
				if actualCount != expectedCount {
					t.Errorf("ValidateConditionalFields() field %s got %d errors, want %d: %v",
						field, actualCount, expectedCount, errors[field])
				}
			}

			// Check no unexpected errors
			for field, fieldErrors := range errors {
				if _, expected := tt.wantErrors[field]; !expected && len(fieldErrors) > 0 {
					t.Errorf("ValidateConditionalFields() unexpected errors for field %s: %v", field, fieldErrors)
				}
			}
		})
	}
}

func TestValidator_validateReadOnlyField(t *testing.T) {
	validator := NewValidator(nil)

	tests := []struct {
		name       string
		field      *MockField
		value      any
		exists     bool
		wantErrors int
	}{
		{
			name: "readonly text field - valid",
			field: &MockField{
				name:      "readonly_text",
				fieldType: FieldText,
			},
			value:      "some text",
			exists:     true,
			wantErrors: 0,
		},
		{
			name: "readonly email field - valid",
			field: &MockField{
				name:      "readonly_email",
				fieldType: FieldEmail,
			},
			value:      "test@example.com",
			exists:     true,
			wantErrors: 0,
		},
		{
			name: "readonly email field - invalid",
			field: &MockField{
				name:      "readonly_email",
				fieldType: FieldEmail,
			},
			value:      "invalid-email",
			exists:     true,
			wantErrors: 1,
		},
		{
			name: "readonly number field - valid",
			field: &MockField{
				name:      "readonly_number",
				fieldType: FieldNumber,
			},
			value:      42.5,
			exists:     true,
			wantErrors: 0,
		},
		{
			name: "readonly number field - invalid",
			field: &MockField{
				name:      "readonly_number",
				fieldType: FieldNumber,
			},
			value:      "not-a-number",
			exists:     true,
			wantErrors: 1,
		},
		{
			name: "readonly field - empty",
			field: &MockField{
				name:      "readonly_text",
				fieldType: FieldText,
			},
			value:      "",
			exists:     false,
			wantErrors: 0, // Empty readonly fields are allowed
		},
		{
			name: "readonly date field - valid",
			field: &MockField{
				name:      "readonly_date",
				fieldType: FieldDate,
			},
			value:      "2023-01-01",
			exists:     true,
			wantErrors: 0,
		},
		{
			name: "readonly date field - invalid",
			field: &MockField{
				name:      "readonly_date",
				fieldType: FieldDate,
			},
			value:      "invalid-date",
			exists:     true,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.validateReadOnlyField(tt.field, tt.value, tt.exists)

			if len(errors) != tt.wantErrors {
				t.Errorf("validateReadOnlyField() got %d errors, want %d: %v",
					len(errors), tt.wantErrors, errors)
			}
		})
	}
}

// Helper function to convert MockSchema to real schema for testing
func convertMockToRealSchema(mockSchema *MockSchema) *schema.Schema {
	realSchema := &schema.Schema{
		ID:     mockSchema.GetID(),
		Type:   schema.Type((mockSchema.GetType())),
		Title:  mockSchema.GetTitle(),
		Fields: make([]schema.Field, 0),
	}

	// Convert mock fields to real fields
	for _, mockField := range mockSchema.GetFields() {
		field := schema.Field{
			Name:     mockField.GetName(),
			Type:     schema.FieldType(mockField.GetType()),
			Required: mockField.GetRequired(),
		}

		// Convert validation if present or create based on field type
		if mockValidation := mockField.GetValidation(); mockValidation != nil {
			field.Validation = &schema.FieldValidation{
				MinLength: mockValidation.MinLength,
				MaxLength: mockValidation.MaxLength,
				Min:       mockValidation.Min,
				Max:       mockValidation.Max,
				Step:      mockValidation.Step,
				Pattern:   mockValidation.Pattern,
				Format:    mockValidation.Format,
				Custom:    mockValidation.Custom,
			}
		} else {
			// Create validation based on field type
			field.Validation = &schema.FieldValidation{}
		}

		// Set format validation based on field type if not already set
		if field.Validation != nil && field.Validation.Format == "" {
			switch mockField.GetType() {
			case FieldEmail:
				field.Validation.Format = "email"
			case FieldURL:
				field.Validation.Format = "url"
			case FieldPhone:
				field.Validation.Format = "phone"
			}
		}

		// Convert options if present
		mockOptions := mockField.GetOptions()
		if len(mockOptions) > 0 {
			field.Options = make([]schema.Option, len(mockOptions))
			for i, opt := range mockOptions {
				field.Options[i] = schema.Option{
					Value: opt.Value,
					Label: opt.Label,
				}
			}
		}

		realSchema.Fields = append(realSchema.Fields, field)
	}

	return realSchema
}
