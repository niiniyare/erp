package schema

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ValidationRegistryTestSuite struct {
	suite.Suite
	registry           *ValidationRegistry
	crossFieldRegistry *CrossFieldValidationRegistry
	ctx                context.Context
}

func (suite *ValidationRegistryTestSuite) SetupTest() {
	suite.registry = NewValidationRegistry()
	suite.crossFieldRegistry = NewCrossFieldValidationRegistry()
	suite.ctx = context.Background()
}

func TestValidationRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationRegistryTestSuite))
}

// Test ValidationRegistry creation and built-in validators
func (suite *ValidationRegistryTestSuite) TestNewValidationRegistry() {
	require.NotNil(suite.T(), suite.registry)
	require.NotNil(suite.T(), suite.registry.validators)
	require.NotNil(suite.T(), suite.registry.asyncValidators)

	// Test that built-in validators are registered
	builtInValidators := []string{"email", "phone", "url", "date_format", "time_range", "luhn", "business_day"}
	for _, name := range builtInValidators {
		err := suite.registry.Validate(suite.ctx, name, "", nil)
		// Should not return "validator not found" error
		if err != nil {
			require.NotContains(suite.T(), err.Error(), "validator "+name+" not found")
		}
	}

	// Test that built-in async validator is registered
	err := suite.registry.ValidateAsync(suite.ctx, "unique", "test", map[string]any{"table": "users", "field": "email"})
	require.NoError(suite.T(), err) // Should not return "async validator not found"
}

// Test custom validator registration
func (suite *ValidationRegistryTestSuite) TestRegisterValidator() {
	// Test successful registration
	customValidator := func(ctx context.Context, value any, params map[string]any) error {
		if value == "invalid" {
			return NewValidationError("custom_error", "custom validation failed")
		}
		return nil
	}

	err := suite.registry.Register("custom", customValidator)
	require.NoError(suite.T(), err)

	// Test validation works
	err = suite.registry.Validate(suite.ctx, "custom", "valid", nil)
	require.NoError(suite.T(), err)

	err = suite.registry.Validate(suite.ctx, "custom", "invalid", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "custom validation failed")

	// Test error cases
	err = suite.registry.Register("", customValidator)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "validator name is required")

	err = suite.registry.Register("test", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "validator function is required")
}

// Test async validator registration
func (suite *ValidationRegistryTestSuite) TestRegisterAsyncValidator() {
	asyncValidator := &AsyncValidator{
		Name:     "async_test",
		Debounce: 100 * time.Millisecond,
		Cache:    true,
		CacheTTL: 1 * time.Minute,
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			if value == "async_invalid" {
				return NewValidationError("async_error", "async validation failed")
			}
			return nil
		},
	}

	err := suite.registry.RegisterAsync(asyncValidator)
	require.NoError(suite.T(), err)

	// Test validation works
	err = suite.registry.ValidateAsync(suite.ctx, "async_test", "valid", nil)
	require.NoError(suite.T(), err)

	err = suite.registry.ValidateAsync(suite.ctx, "async_test", "async_invalid", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "async validation failed")

	// Test error cases
	err = suite.registry.RegisterAsync(nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "async validator is required")

	err = suite.registry.RegisterAsync(&AsyncValidator{Name: ""})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "async validator name is required")

	err = suite.registry.RegisterAsync(&AsyncValidator{Name: "test", Validate: nil})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "async validator function is required")
}

// Test built-in email validator
func (suite *ValidationRegistryTestSuite) TestEmailValidator() {
	testCases := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"user+tag@example.org", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
	}

	for _, tc := range testCases {
		err := suite.registry.Validate(suite.ctx, "email", tc.email, nil)
		if tc.valid {
			require.NoError(suite.T(), err, "Email %s should be valid", tc.email)
		} else {
			require.Error(suite.T(), err, "Email %s should be invalid", tc.email)
		}
	}

	// Test non-string value
	err := suite.registry.Validate(suite.ctx, "email", 123, nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "email validator requires string value")
}

// Test built-in phone validator
func (suite *ValidationRegistryTestSuite) TestPhoneValidator() {
	testCases := []struct {
		phone string
		valid bool
	}{
		{"+1234567890", true},
		{"(555) 123-4567", true},
		{"555.123.4567", true},
		{"5551234567", true},
		{"123", false},
		{"abc", false},
		{"", false},
	}

	for _, tc := range testCases {
		err := suite.registry.Validate(suite.ctx, "phone", tc.phone, nil)
		if tc.valid {
			require.NoError(suite.T(), err, "Phone %s should be valid", tc.phone)
		} else {
			require.Error(suite.T(), err, "Phone %s should be invalid", tc.phone)
		}
	}
}

// Test built-in URL validator
func (suite *ValidationRegistryTestSuite) TestURLValidator() {
	testCases := []struct {
		url   string
		valid bool
	}{
		{"https://example.com", true},
		{"http://subdomain.example.org/path", true},
		{"https://example.com/path?query=value", true},
		{"invalid-url", false},
		{"ftp://example.com", false},
		{"", false},
	}

	for _, tc := range testCases {
		err := suite.registry.Validate(suite.ctx, "url", tc.url, nil)
		if tc.valid {
			require.NoError(suite.T(), err, "URL %s should be valid", tc.url)
		} else {
			require.Error(suite.T(), err, "URL %s should be invalid", tc.url)
		}
	}
}

// Test date format validator
func (suite *ValidationRegistryTestSuite) TestDateFormatValidator() {
	// Test with default ISO8601 format
	err := suite.registry.Validate(suite.ctx, "date_format", "2023-12-25", nil)
	require.NoError(suite.T(), err)

	err = suite.registry.Validate(suite.ctx, "date_format", "25/12/2023", nil)
	require.Error(suite.T(), err)

	// Test with custom format
	params := map[string]any{"layout": "01/02/2006"}
	err = suite.registry.Validate(suite.ctx, "date_format", "12/25/2023", params)
	require.NoError(suite.T(), err)

	err = suite.registry.Validate(suite.ctx, "date_format", "2023-12-25", params)
	require.Error(suite.T(), err)
}

// Test Luhn algorithm validator
func (suite *ValidationRegistryTestSuite) TestLuhnValidator() {
	testCases := []struct {
		number string
		valid  bool
	}{
		{"4532015112830366", true},    // Valid Visa
		{"5555555555554444", true},    // Valid MasterCard
		{"4111 1111 1111 1111", true}, // Valid with spaces
		{"4111-1111-1111-1111", true}, // Valid with dashes
		{"1234567890123456", false},   // Invalid
		{"", false},
		{"abc", false},
	}

	for _, tc := range testCases {
		err := suite.registry.Validate(suite.ctx, "luhn", tc.number, nil)
		if tc.valid {
			require.NoError(suite.T(), err, "Card number %s should be valid", tc.number)
		} else {
			require.Error(suite.T(), err, "Card number %s should be invalid", tc.number)
		}
	}
}

// Test business day validator
func (suite *ValidationRegistryTestSuite) TestBusinessDayValidator() {
	// Monday (business day) - 2023-12-25 is a Monday
	monday := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	err := suite.registry.Validate(suite.ctx, "business_day", monday, nil)
	// Check that it doesn't return invalid type error
	if err != nil {
		require.NotContains(suite.T(), err.Error(), "invalid_type")
	}

	// Test string version
	err = suite.registry.Validate(suite.ctx, "business_day", "2023-12-25", nil)
	// Just ensure it doesn't crash and doesn't return invalid type
	if err != nil {
		require.NotContains(suite.T(), err.Error(), "invalid_type")
	}

	// Test invalid type
	err = suite.registry.Validate(suite.ctx, "business_day", 123, nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "invalid_type")
}

// Test async validator caching
func (suite *ValidationRegistryTestSuite) TestAsyncValidatorCaching() {
	callCount := 0
	asyncValidator := &AsyncValidator{
		Name:     "cache_test",
		Cache:    true,
		CacheTTL: 1 * time.Second,
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			callCount++
			return nil
		},
	}

	err := suite.registry.RegisterAsync(asyncValidator)
	require.NoError(suite.T(), err)

	// First call should execute the function
	err = suite.registry.ValidateAsync(suite.ctx, "cache_test", "test_value", nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, callCount)

	// Second call should use cache
	err = suite.registry.ValidateAsync(suite.ctx, "cache_test", "test_value", nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, callCount, "Should not have called validator again due to caching")

	// Wait for cache to expire
	time.Sleep(1100 * time.Millisecond)

	// Third call should execute the function again
	err = suite.registry.ValidateAsync(suite.ctx, "cache_test", "test_value", nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 2, callCount, "Should have called validator again after cache expiry")
}

// Test async validator debouncing
func (suite *ValidationRegistryTestSuite) TestAsyncValidatorDebouncing() {
	asyncValidator := &AsyncValidator{
		Name:     "debounce_test",
		Debounce: 100 * time.Millisecond,
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			return nil
		},
	}

	err := suite.registry.RegisterAsync(asyncValidator)
	require.NoError(suite.T(), err)

	start := time.Now()
	err = suite.registry.ValidateAsync(suite.ctx, "debounce_test", "test", nil)
	require.NoError(suite.T(), err)
	elapsed := time.Since(start)

	require.GreaterOrEqual(suite.T(), elapsed, 100*time.Millisecond, "Should have debounced for at least 100ms")
}

// Test cross-field validation registry
func (suite *ValidationRegistryTestSuite) TestCrossFieldValidationRegistry() {
	require.NotNil(suite.T(), suite.crossFieldRegistry)
	require.NotNil(suite.T(), suite.crossFieldRegistry.validators)
	require.NotNil(suite.T(), suite.crossFieldRegistry.evaluator)

	// Test built-in cross-field validators are registered
	builtInValidators := []string{"date_range", "password_confirmation", "business_hours"}
	for _, name := range builtInValidators {
		err := suite.crossFieldRegistry.Validate(suite.ctx, name, map[string]any{})
		// Should not return "cross-field validator not found" error
		if err != nil {
			require.NotContains(suite.T(), err.Error(), "cross-field validator "+name+" not found")
		}
	}
}

// Test cross-field validator registration
func (suite *ValidationRegistryTestSuite) TestRegisterCrossFieldValidator() {
	validator := &CrossFieldValidator{
		Name:   "test_cross",
		Fields: []string{"field1", "field2"},
		Validate: func(ctx context.Context, values map[string]any) error {
			val1, ok1 := values["field1"]
			val2, ok2 := values["field2"]
			if ok1 && ok2 && val1 == val2 {
				return NewValidationError("fields_equal", "fields should not be equal")
			}
			return nil
		},
	}

	err := suite.crossFieldRegistry.Register(validator)
	require.NoError(suite.T(), err)

	// Test validation
	data := map[string]any{"field1": "value", "field2": "different"}
	err = suite.crossFieldRegistry.Validate(suite.ctx, "test_cross", data)
	require.NoError(suite.T(), err)

	data = map[string]any{"field1": "same", "field2": "same"}
	err = suite.crossFieldRegistry.Validate(suite.ctx, "test_cross", data)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "fields should not be equal")

	// Test error cases
	err = suite.crossFieldRegistry.Register(nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "cross-field validator is required")

	err = suite.crossFieldRegistry.Register(&CrossFieldValidator{Name: ""})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "cross-field validator name is required")

	err = suite.crossFieldRegistry.Register(&CrossFieldValidator{Name: "test", Fields: []string{}})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "cross-field validator must specify fields")
}

// Test built-in date range validator
func (suite *ValidationRegistryTestSuite) TestDateRangeValidator() {
	// Valid date range
	data := map[string]any{
		"start_date": "2023-12-01",
		"end_date":   "2023-12-31",
	}
	err := suite.crossFieldRegistry.Validate(suite.ctx, "date_range", data)
	require.NoError(suite.T(), err)

	// Invalid date range (start after end)
	data = map[string]any{
		"start_date": "2023-12-31",
		"end_date":   "2023-12-01",
	}
	err = suite.crossFieldRegistry.Validate(suite.ctx, "date_range", data)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "start date must be before end date")

	// Missing fields (should not error)
	data = map[string]any{"start_date": "2023-12-01"}
	err = suite.crossFieldRegistry.Validate(suite.ctx, "date_range", data)
	require.NoError(suite.T(), err)
}

// Test built-in password confirmation validator
func (suite *ValidationRegistryTestSuite) TestPasswordConfirmationValidator() {
	// Matching passwords
	data := map[string]any{
		"password":              "secret123",
		"password_confirmation": "secret123",
	}
	err := suite.crossFieldRegistry.Validate(suite.ctx, "password_confirmation", data)
	require.NoError(suite.T(), err)

	// Non-matching passwords
	data = map[string]any{
		"password":              "secret123",
		"password_confirmation": "different",
	}
	err = suite.crossFieldRegistry.Validate(suite.ctx, "password_confirmation", data)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "password confirmation does not match")

	// Missing fields (should not error)
	data = map[string]any{"password": "secret123"}
	err = suite.crossFieldRegistry.Validate(suite.ctx, "password_confirmation", data)
	require.NoError(suite.T(), err)
}

// Test validator not found errors
func (suite *ValidationRegistryTestSuite) TestValidatorNotFound() {
	err := suite.registry.Validate(suite.ctx, "nonexistent", "value", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "validator nonexistent not found")

	err = suite.registry.ValidateAsync(suite.ctx, "nonexistent_async", "value", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "async validator nonexistent_async not found")

	err = suite.crossFieldRegistry.Validate(suite.ctx, "nonexistent_cross", map[string]any{})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "cross-field validator nonexistent_cross not found")
}

// Test field validation with registry
func (suite *ValidationRegistryTestSuite) TestFieldValidateWithRegistry() {
	// Create a field with custom validation
	field := &Field{
		Name:     "email_field",
		Type:     FieldEmail,
		Required: true,
		Validation: &FieldValidation{
			Custom: "email",
		},
	}

	// Valid email
	err := field.ValidateWithRegistry(suite.ctx, "test@example.com", suite.registry)
	require.NoError(suite.T(), err)

	// Invalid email
	err = field.ValidateWithRegistry(suite.ctx, "invalid-email", suite.registry)
	require.Error(suite.T(), err)

	// Field without custom validation
	field2 := &Field{
		Name:     "text_field",
		Type:     FieldText,
		Required: true,
	}

	err = field2.ValidateWithRegistry(suite.ctx, "any text", suite.registry)
	require.NoError(suite.T(), err)
}

// Test Luhn algorithm implementation
func (suite *ValidationRegistryTestSuite) TestLuhnAlgorithm() {
	// Test the isValidLuhn function directly
	testCases := []struct {
		number string
		valid  bool
	}{
		{"4532015112830366", true},
		{"5555555555554444", true},
		{"378282246310005", true},
		{"6011111111111117", true},
		{"1234567890123456", false},
		{"4111111111111112", false},
		{"", false},
		{"abc", false},
	}

	for _, tc := range testCases {
		result := isValidLuhn(tc.number)
		require.Equal(suite.T(), tc.valid, result, "Luhn validation for %s should be %t", tc.number, tc.valid)
	}
}

// Test async validator context cancellation
func (suite *ValidationRegistryTestSuite) TestAsyncValidatorContextCancellation() {
	asyncValidator := &AsyncValidator{
		Name:     "cancel_test",
		Debounce: 1 * time.Second, // Long debounce to test cancellation
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			return nil
		},
	}

	err := suite.registry.RegisterAsync(asyncValidator)
	require.NoError(suite.T(), err)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(suite.ctx)

	// Start validation in a goroutine
	done := make(chan error, 1)
	go func() {
		err := suite.registry.ValidateAsync(ctx, "cancel_test", "test", nil)
		done <- err
	}()

	// Cancel the context after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the validation to complete
	select {
	case err := <-done:
		require.Error(suite.T(), err)
		require.Equal(suite.T(), context.Canceled, err)
	case <-time.After(2 * time.Second):
		suite.T().Fatal("Validation did not complete within timeout")
	}
}

// Test time range validator
func (suite *ValidationRegistryTestSuite) TestTimeRangeValidator() {
	// Valid duration
	err := suite.registry.Validate(suite.ctx, "time_range", "2h30m", nil)
	require.NoError(suite.T(), err)

	// Valid duration with min/max
	params := map[string]any{
		"min": "1h",
		"max": "4h",
	}
	err = suite.registry.Validate(suite.ctx, "time_range", "2h", params)
	require.NoError(suite.T(), err)

	// Duration too short
	err = suite.registry.Validate(suite.ctx, "time_range", "30m", params)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "duration must be at least")

	// Duration too long
	err = suite.registry.Validate(suite.ctx, "time_range", "5h", params)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "duration must be at most")

	// Invalid duration format
	err = suite.registry.Validate(suite.ctx, "time_range", "invalid", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "invalid duration format")
}
