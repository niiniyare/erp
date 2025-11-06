package validate

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Mock implementations for testing
type MockField struct {
	name       string
	fieldType  FieldType
	required   bool
	validation *FieldValidation
	options    []Option
	config     map[string]any
}

func (f *MockField) GetName() string                 { return f.name }
func (f *MockField) GetType() FieldType              { return f.fieldType }
func (f *MockField) GetRequired() bool               { return f.required }
func (f *MockField) GetValidation() *FieldValidation { return f.validation }
func (f *MockField) GetOptions() []Option            { return f.options }
func (f *MockField) GetConfig() map[string]any       { return f.config }

type MockSchema struct {
	id         string
	schemaType string
	title      string
	fields     []FieldInterface
	validation any
}

func (s *MockSchema) GetID() string               { return s.id }
func (s *MockSchema) GetType() string             { return s.schemaType }
func (s *MockSchema) GetTitle() string            { return s.title }
func (s *MockSchema) GetFields() []FieldInterface { return s.fields }
func (s *MockSchema) GetValidation() any          { return s.validation }

func TestNewValidationRegistry(t *testing.T) {
	registry := NewValidationRegistry()

	if registry == nil {
		t.Fatal("NewValidationRegistry() returned nil")
	}

	if registry.validators == nil {
		t.Error("NewValidationRegistry() validators map is nil")
	}

	if registry.asyncValidators == nil {
		t.Error("NewValidationRegistry() asyncValidators map is nil")
	}
}

func TestValidationRegistry_Register(t *testing.T) {
	registry := NewValidationRegistry()
	ctx := context.Background()

	// Test successful registration
	err := registry.Register("test", func(ctx context.Context, value any, params map[string]any) error {
		return nil
	})
	if err != nil {
		t.Errorf("Register() unexpected error: %v", err)
	}

	// Test validation works
	err = registry.Validate(ctx, "test", "value", map[string]any{})
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Test invalid registration - empty name
	err = registry.Register("", func(ctx context.Context, value any, params map[string]any) error {
		return nil
	})
	if err == nil {
		t.Error("Register() should return error for empty name")
	}

	// Test invalid registration - nil function
	err = registry.Register("test2", nil)
	if err == nil {
		t.Error("Register() should return error for nil function")
	}
}

func TestValidationRegistry_Validate(t *testing.T) {
	registry := NewValidationRegistry()
	ctx := context.Background()

	// Register a test validator
	registry.Register("length_check", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "must be string", "VALIDATION_ERROR")
		}

		minLength, hasMin := params["min"].(int)
		if hasMin && len(str) < minLength {
			return NewValidationError("too_short", fmt.Sprintf("must be at least %d characters", minLength), "VALIDATION_ERROR")
		}

		return nil
	})

	tests := []struct {
		name      string
		validator string
		value     any
		params    map[string]any
		wantError bool
	}{
		{
			name:      "valid string",
			validator: "length_check",
			value:     "hello",
			params:    map[string]any{"min": 3},
			wantError: false,
		},
		{
			name:      "too short string",
			validator: "length_check",
			value:     "hi",
			params:    map[string]any{"min": 5},
			wantError: true,
		},
		{
			name:      "invalid type",
			validator: "length_check",
			value:     123,
			params:    map[string]any{},
			wantError: true,
		},
		{
			name:      "unknown validator",
			validator: "unknown",
			value:     "test",
			params:    map[string]any{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Validate(ctx, tt.validator, tt.value, tt.params)

			if tt.wantError && err == nil {
				t.Error("Validate() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestValidationRegistry_BuiltInValidators(t *testing.T) {
	registry := NewValidationRegistry()
	ctx := context.Background()

	tests := []struct {
		name      string
		validator string
		value     any
		params    map[string]any
		wantError bool
	}{
		{
			name:      "valid email",
			validator: "email",
			value:     "test@example.com",
			params:    map[string]any{},
			wantError: false,
		},
		{
			name:      "invalid email",
			validator: "email",
			value:     "not-an-email",
			params:    map[string]any{},
			wantError: true,
		},
		{
			name:      "valid phone",
			validator: "phone",
			value:     "+1234567890",
			params:    map[string]any{},
			wantError: false,
		},
		{
			name:      "invalid phone",
			validator: "phone",
			value:     "123",
			params:    map[string]any{},
			wantError: true,
		},
		{
			name:      "valid URL",
			validator: "url",
			value:     "https://example.com",
			params:    map[string]any{},
			wantError: false,
		},
		{
			name:      "invalid URL",
			validator: "url",
			value:     "not-a-url",
			params:    map[string]any{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Validate(ctx, tt.validator, tt.value, tt.params)

			if tt.wantError && err == nil {
				t.Errorf("Validate(%s) expected error but got none", tt.validator)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate(%s) unexpected error: %v", tt.validator, err)
			}
		})
	}
}

func TestValidationRegistry_AsyncValidator(t *testing.T) {
	registry := NewValidationRegistry()
	ctx := context.Background()

	// Register an async validator
	asyncValidator := &AsyncValidator{
		Name:     "async_test",
		Debounce: 10 * time.Millisecond,
		Cache:    true,
		CacheTTL: 1 * time.Second,
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			str, ok := value.(string)
			if !ok {
				return NewValidationError("invalid_type", "must be string", "VALIDATION_ERROR")
			}
			if str == "error" {
				return NewValidationError("test_error", "test error", "VALIDATION_ERROR")
			}
			return nil
		},
	}

	err := registry.RegisterAsync(asyncValidator)
	if err != nil {
		t.Errorf("RegisterAsync() unexpected error: %v", err)
	}

	// Test async validation
	err = registry.ValidateAsync(ctx, "async_test", "valid", map[string]any{})
	if err != nil {
		t.Errorf("ValidateAsync() unexpected error: %v", err)
	}

	// Test async validation with error
	err = registry.ValidateAsync(ctx, "async_test", "error", map[string]any{})
	if err == nil {
		t.Error("ValidateAsync() expected error but got none")
	}

	// Test caching (second call should be faster)
	start := time.Now()
	err = registry.ValidateAsync(ctx, "async_test", "valid", map[string]any{})
	duration := time.Since(start)

	if err != nil {
		t.Errorf("ValidateAsync() cached call unexpected error: %v", err)
	}
	if duration > 5*time.Millisecond { // Should be much faster due to caching
		t.Errorf("ValidateAsync() cached call took too long: %v", duration)
	}

	// Clean up
	registry.UnregisterAsync("async_test")
}

func TestNewCrossFieldValidationRegistry(t *testing.T) {
	registry := NewCrossFieldValidationRegistry()

	if registry == nil {
		t.Fatal("NewCrossFieldValidationRegistry() returned nil")
	}

	if registry.validators == nil {
		t.Error("NewCrossFieldValidationRegistry() validators map is nil")
	}

	if registry.evaluator == nil {
		t.Error("NewCrossFieldValidationRegistry() evaluator is nil")
	}
}

func TestCrossFieldValidationRegistry_Register(t *testing.T) {
	registry := NewCrossFieldValidationRegistry()

	validator := &CrossFieldValidator{
		Name:   "test_cross",
		Fields: []string{"field1", "field2"},
		Validate: func(ctx context.Context, values map[string]any) error {
			return nil
		},
	}

	err := registry.Register(validator)
	if err != nil {
		t.Errorf("Register() unexpected error: %v", err)
	}

	// Test invalid registration - nil validator
	err = registry.Register(nil)
	if err == nil {
		t.Error("Register() should return error for nil validator")
	}

	// Test invalid registration - empty name
	invalidValidator := &CrossFieldValidator{
		Name:   "",
		Fields: []string{"field1"},
		Validate: func(ctx context.Context, values map[string]any) error {
			return nil
		},
	}
	err = registry.Register(invalidValidator)
	if err == nil {
		t.Error("Register() should return error for empty name")
	}

	// Test invalid registration - no fields
	invalidValidator2 := &CrossFieldValidator{
		Name:   "test",
		Fields: []string{},
		Validate: func(ctx context.Context, values map[string]any) error {
			return nil
		},
	}
	err = registry.Register(invalidValidator2)
	if err == nil {
		t.Error("Register() should return error for empty fields")
	}
}

func TestCrossFieldValidationRegistry_Validate(t *testing.T) {
	registry := NewCrossFieldValidationRegistry()
	ctx := context.Background()

	// Register a test validator
	validator := &CrossFieldValidator{
		Name:   "match_fields",
		Fields: []string{"password", "confirm_password"},
		Validate: func(ctx context.Context, values map[string]any) error {
			password, hasPassword := values["password"]
			confirm, hasConfirm := values["confirm_password"]

			if !hasPassword || !hasConfirm {
				return nil // Skip if fields missing
			}

			if password != confirm {
				return NewValidationError("password_mismatch", "passwords do not match", "VALIDATION_ERROR")
			}

			return nil
		},
	}

	err := registry.Register(validator)
	if err != nil {
		t.Errorf("Register() unexpected error: %v", err)
	}

	tests := []struct {
		name      string
		validator string
		values    map[string]any
		wantError bool
	}{
		{
			name:      "matching passwords",
			validator: "match_fields",
			values: map[string]any{
				"password":         "secret123",
				"confirm_password": "secret123",
			},
			wantError: false,
		},
		{
			name:      "non-matching passwords",
			validator: "match_fields",
			values: map[string]any{
				"password":         "secret123",
				"confirm_password": "different",
			},
			wantError: true,
		},
		{
			name:      "missing fields",
			validator: "match_fields",
			values: map[string]any{
				"password": "secret123",
			},
			wantError: false, // Should skip validation
		},
		{
			name:      "unknown validator",
			validator: "unknown",
			values:    map[string]any{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Validate(ctx, tt.validator, tt.values)

			if tt.wantError && err == nil {
				t.Error("Validate() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestCrossFieldValidationRegistry_BuiltInValidators(t *testing.T) {
	registry := NewCrossFieldValidationRegistry()
	ctx := context.Background()

	tests := []struct {
		name      string
		validator string
		values    map[string]any
		wantError bool
	}{
		{
			name:      "valid date range",
			validator: "date_range",
			values: map[string]any{
				"start_date": "2023-01-01",
				"end_date":   "2023-12-31",
			},
			wantError: false,
		},
		{
			name:      "invalid date range",
			validator: "date_range",
			values: map[string]any{
				"start_date": "2023-12-31",
				"end_date":   "2023-01-01",
			},
			wantError: true,
		},
		{
			name:      "matching passwords",
			validator: "password_confirmation",
			values: map[string]any{
				"password":              "secret123",
				"password_confirmation": "secret123",
			},
			wantError: false,
		},
		{
			name:      "non-matching passwords",
			validator: "password_confirmation",
			values: map[string]any{
				"password":              "secret123",
				"password_confirmation": "different",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Validate(ctx, tt.validator, tt.values)

			if tt.wantError && err == nil {
				t.Errorf("Validate(%s) expected error but got none", tt.validator)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate(%s) unexpected error: %v", tt.validator, err)
			}
		})
	}
}

func TestValidatorChain(t *testing.T) {
	chain := NewValidatorChain()
	ctx := context.Background()

	// Add validators to chain
	chain.Add(func(ctx context.Context, value any, params map[string]any) error {
		if value == nil {
			return NewValidationError("nil_value", "value cannot be nil", "VALIDATION_ERROR")
		}
		return nil
	})

	chain.Add(func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "must be string", "VALIDATION_ERROR")
		}
		if len(str) < 3 {
			return NewValidationError("too_short", "must be at least 3 characters", "VALIDATION_ERROR")
		}
		return nil
	})

	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{
			name:      "valid string",
			value:     "hello",
			wantError: false,
		},
		{
			name:      "nil value",
			value:     nil,
			wantError: true,
		},
		{
			name:      "invalid type",
			value:     123,
			wantError: true,
		},
		{
			name:      "too short",
			value:     "hi",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chain.Validate(ctx, tt.value, map[string]any{})

			if tt.wantError && err == nil {
				t.Error("Validate() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateWithRegistry(t *testing.T) {
	validator := NewValidator(nil)
	registry := NewValidationRegistry()
	ctx := context.Background()

	// Register a custom validator
	registry.Register("custom_test", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "must be string", "VALIDATION_ERROR")
		}
		if str == "forbidden" {
			return NewValidationError("forbidden_value", "this value is not allowed", "VALIDATION_ERROR")
		}
		return nil
	})

	field := &MockField{
		name:      "test_field",
		fieldType: FieldText,
		required:  true,
		validation: &FieldValidation{
			Custom: "custom_test",
		},
	}

	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{
			name:      "valid value",
			value:     "allowed",
			wantError: false,
		},
		{
			name:      "forbidden value",
			value:     "forbidden",
			wantError: true,
		},
		{
			name:      "invalid type",
			value:     123,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateWithRegistry(ctx, field, tt.value, registry)

			if tt.wantError && len(errors) == 0 {
				t.Error("ValidateWithRegistry() expected errors but got none")
			}
			if !tt.wantError && len(errors) > 0 {
				t.Errorf("ValidateWithRegistry() unexpected errors: %v", errors)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidationRegistry_Validate(b *testing.B) {
	registry := NewValidationRegistry()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Validate(ctx, "email", "test@example.com", map[string]any{})
	}
}

func BenchmarkValidationRegistry_ValidateAsync(b *testing.B) {
	registry := NewValidationRegistry()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.ValidateAsync(ctx, "unique", "test_value", map[string]any{})
	}
}
