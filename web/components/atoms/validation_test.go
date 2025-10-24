package atoms

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ValidationTestSuite tests the validation functions in validation.go
type ValidationTestSuite struct {
	suite.Suite
}

// SetupTest runs before each test method
func (suite *ValidationTestSuite) SetupTest() {
	ResetIDCounter()
	ClearRegexCache() // Clear cache for consistent testing
}

// TestValidationTestSuite runs the validation test suite
func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

// ============================================================================
// REGEX CACHE TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestGetCompiledRegex() {
	// Test valid regex
	regex, err := getCompiledRegex(`^\d+$`)
	suite.NoError(err)
	suite.NotNil(regex)
	suite.True(regex.MatchString("123"))
	suite.False(regex.MatchString("abc"))

	// Test invalid regex
	_, err = getCompiledRegex(`[`)
	suite.Error(err)
}

func (suite *ValidationTestSuite) TestRegexCaching() {
	pattern := `^\d+$`
	
	// First call should compile and cache
	regex1, err1 := getCompiledRegex(pattern)
	suite.NoError(err1)
	
	// Second call should use cache
	regex2, err2 := getCompiledRegex(pattern)
	suite.NoError(err2)
	
	// Should be the same instance (cached)
	suite.Equal(regex1, regex2)
}

func (suite *ValidationTestSuite) TestClearRegexCache() {
	pattern := `^\d+$`
	
	// Compile and cache regex
	_, err := getCompiledRegex(pattern)
	suite.NoError(err)
	
	// Clear cache
	ClearRegexCache()
	
	// Should work after clearing
	regex, err := getCompiledRegex(pattern)
	suite.NoError(err)
	suite.NotNil(regex)
}

// ============================================================================
// REQUIRED VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestRequiredValidator() {
	validator := RequiredValidator{}

	// Test valid values
	suite.NoError(validator.Validate("valid string"))
	suite.NoError(validator.Validate("  valid  "))
	suite.NoError(validator.Validate([]any{1, 2, 3}))
	suite.NoError(validator.Validate(map[string]any{"key": "value"}))

	// Test invalid values
	suite.Error(validator.Validate(""))
	suite.Error(validator.Validate("   "))
	suite.Error(validator.Validate(nil))
	suite.Error(validator.Validate([]any{}))
	suite.Error(validator.Validate(map[string]any{}))
}

func (suite *ValidationTestSuite) TestRequiredValidatorErrorMessage() {
	// Default message
	validator := RequiredValidator{}
	suite.Equal("This field is required", validator.ErrorMessage())

	// Custom message
	customValidator := RequiredValidator{Message: "Custom required message"}
	suite.Equal("Custom required message", customValidator.ErrorMessage())
}

// ============================================================================
// MIN LENGTH VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestMinLengthValidator() {
	validator := MinLengthValidator{MinLength: 5}

	// Test valid values
	suite.NoError(validator.Validate("12345"))
	suite.NoError(validator.Validate("123456"))

	// Test invalid values
	suite.Error(validator.Validate("1234"))
	suite.Error(validator.Validate(""))

	// Test non-string value
	suite.Error(validator.Validate(123))
}

func (suite *ValidationTestSuite) TestMinLengthValidatorErrorMessage() {
	// Default message
	validator := MinLengthValidator{MinLength: 5}
	suite.Equal("Must be at least 5 characters", validator.ErrorMessage())

	// Custom message
	customValidator := MinLengthValidator{MinLength: 5, Message: "Too short"}
	suite.Equal("Too short", customValidator.ErrorMessage())
}

// ============================================================================
// MAX LENGTH VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestMaxLengthValidator() {
	validator := MaxLengthValidator{MaxLength: 10}

	// Test valid values
	suite.NoError(validator.Validate("12345"))
	suite.NoError(validator.Validate("1234567890"))

	// Test invalid values
	suite.Error(validator.Validate("12345678901"))

	// Test non-string value
	suite.Error(validator.Validate(123))
}

func (suite *ValidationTestSuite) TestMaxLengthValidatorErrorMessage() {
	// Default message
	validator := MaxLengthValidator{MaxLength: 10}
	suite.Equal("Must be no more than 10 characters", validator.ErrorMessage())

	// Custom message
	customValidator := MaxLengthValidator{MaxLength: 10, Message: "Too long"}
	suite.Equal("Too long", customValidator.ErrorMessage())
}

// ============================================================================
// PATTERN VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestPatternValidator() {
	validator := PatternValidator{Pattern: `^\d+$`}

	// Test valid values
	suite.NoError(validator.Validate("123"))
	suite.NoError(validator.Validate("0"))

	// Test invalid values
	suite.Error(validator.Validate("abc"))
	suite.Error(validator.Validate("12a"))

	// Test non-string value
	suite.Error(validator.Validate(123))
}

func (suite *ValidationTestSuite) TestPatternValidatorInvalidRegex() {
	validator := PatternValidator{Pattern: `[`}
	err := validator.Validate("test")
	suite.Error(err)
	suite.Contains(err.Error(), "invalid regex pattern")
}

func (suite *ValidationTestSuite) TestPatternValidatorErrorMessage() {
	// Default message
	validator := PatternValidator{Pattern: `^\d+$`}
	suite.Equal("Invalid format", validator.ErrorMessage())

	// Custom message
	customValidator := PatternValidator{Pattern: `^\d+$`, Message: "Numbers only"}
	suite.Equal("Numbers only", customValidator.ErrorMessage())
}

// ============================================================================
// EMAIL VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestEmailValidator() {
	validator := EmailValidator{}

	// Test valid emails
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
		"first.last@subdomain.example.com",
	}

	for _, email := range validEmails {
		suite.NoError(validator.Validate(email), "Expected %s to be valid", email)
	}

	// Test invalid emails
	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"test@",
		"",
		"test@example", // Missing TLD
	}

	for _, email := range invalidEmails {
		suite.Error(validator.Validate(email), "Expected %s to be invalid", email)
	}

	// Test non-string value
	suite.Error(validator.Validate(123))
}

func (suite *ValidationTestSuite) TestEmailValidatorErrorMessage() {
	// Default message
	validator := EmailValidator{}
	suite.Equal("Please enter a valid email address", validator.ErrorMessage())

	// Custom message
	customValidator := EmailValidator{Message: "Invalid email"}
	suite.Equal("Invalid email", customValidator.ErrorMessage())
}

// ============================================================================
// URL VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestURLValidator() {
	validator := URLValidator{}

	// Test valid URLs
	validURLs := []string{
		"https://example.com",
		"http://www.example.org",
		"https://sub.domain.com/path",
		"http://localhost:8080",
	}

	for _, url := range validURLs {
		suite.NoError(validator.Validate(url), "Expected %s to be valid", url)
	}

	// Test invalid URLs
	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com",
		"",
		"example.com",
		"http://",
	}

	for _, url := range invalidURLs {
		suite.Error(validator.Validate(url), "Expected %s to be invalid", url)
	}

	// Test non-string value
	suite.Error(validator.Validate(123))
}

func (suite *ValidationTestSuite) TestURLValidatorErrorMessage() {
	// Default message
	validator := URLValidator{}
	suite.Equal("Please enter a valid URL", validator.ErrorMessage())

	// Custom message
	customValidator := URLValidator{Message: "Invalid URL"}
	suite.Equal("Invalid URL", customValidator.ErrorMessage())
}

// ============================================================================
// RANGE VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestRangeValidator() {
	validator := RangeValidator{Min: 1.0, Max: 10.0}

	// Test valid numeric values
	validValues := []any{
		1.0, 5.5, 10.0, int(5), int64(7), int32(3), float32(2.5), "5.0", "1", "10",
	}

	for _, value := range validValues {
		suite.NoError(validator.Validate(value), "Expected %v to be valid", value)
	}

	// Test invalid numeric values
	invalidValues := []any{
		0.5, 10.1, int(-1), int64(15), "0", "11", "-5",
	}

	for _, value := range invalidValues {
		suite.Error(validator.Validate(value), "Expected %v to be invalid", value)
	}

	// Test non-numeric values
	nonNumericValues := []any{
		"abc", []int{1, 2, 3}, map[string]string{"key": "value"},
	}

	for _, value := range nonNumericValues {
		suite.Error(validator.Validate(value), "Expected %v to error", value)
	}
}

func (suite *ValidationTestSuite) TestRangeValidatorErrorMessage() {
	// Default message
	validator := RangeValidator{Min: 1.0, Max: 10.0}
	suite.Equal("Must be between 1.00 and 10.00", validator.ErrorMessage())

	// Custom message
	customValidator := RangeValidator{Min: 1.0, Max: 10.0, Message: "Out of range"}
	suite.Equal("Out of range", customValidator.ErrorMessage())
}

// ============================================================================
// CUSTOM VALIDATOR TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestCustomValidator() {
	// Custom validator that checks if string starts with "test"
	validator := CustomValidator{
		ValidateFunc: func(value any) error {
			str, ok := value.(string)
			if !ok {
				return errors.New("must be string")
			}
			if !strings.HasPrefix(str, "test") {
				return errors.New("must start with 'test'")
			}
			return nil
		},
	}

	// Test valid values
	suite.NoError(validator.Validate("test123"))
	suite.NoError(validator.Validate("testing"))

	// Test invalid values
	suite.Error(validator.Validate("abc"))
	suite.Error(validator.Validate(123))

	// Test nil function
	nilValidator := CustomValidator{}
	suite.NoError(nilValidator.Validate("anything"))
}

func (suite *ValidationTestSuite) TestCustomValidatorErrorMessage() {
	// Default message
	validator := CustomValidator{}
	suite.Equal("Validation failed", validator.ErrorMessage())

	// Custom message
	customValidator := CustomValidator{Message: "Custom error"}
	suite.Equal("Custom error", customValidator.ErrorMessage())
}

// ============================================================================
// VALIDATION CHAIN TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestValidationChain() {
	chain := ValidationChain{
		Validators: []Validator{
			RequiredValidator{},
			MinLengthValidator{MinLength: 3},
			MaxLengthValidator{MaxLength: 10},
		},
	}

	// Test valid value
	suite.NoError(chain.Validate("test"))

	// Test fails required
	suite.Error(chain.Validate(""))

	// Test fails min length
	suite.Error(chain.Validate("ab"))

	// Test fails max length
	suite.Error(chain.Validate("this is too long"))
}

func (suite *ValidationTestSuite) TestValidationChainValidateAll() {
	chain := ValidationChain{
		Validators: []Validator{
			RequiredValidator{},
			MinLengthValidator{MinLength: 5},
			PatternValidator{Pattern: `^\d+$`},
		},
	}

	// Test value that fails multiple validators
	errors := chain.ValidateAll("ab") // fails min length and pattern
	suite.Len(errors, 2)

	// Test valid value
	errors = chain.ValidateAll("12345")
	suite.Empty(errors)
}

func (suite *ValidationTestSuite) TestValidationChainErrorMessage() {
	chain := ValidationChain{
		Validators: []Validator{
			RequiredValidator{Message: "First error"},
			MinLengthValidator{MinLength: 5, Message: "Second error"},
		},
	}

	suite.Equal("First error", chain.ErrorMessage())

	// Empty chain
	emptyChain := ValidationChain{}
	suite.Equal("Validation failed", emptyChain.ErrorMessage())
}

// ============================================================================
// FIELD STATE TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestNewFieldState() {
	validator := RequiredValidator{}
	fieldState := NewFieldState("username", validator)

	suite.Equal("username", fieldState.Name)
	suite.True(fieldState.IsValid)
	suite.False(fieldState.IsDirty)
	suite.False(fieldState.IsTouched)
	suite.Empty(fieldState.Errors)
	suite.Len(fieldState.Validators, 1)
}

func (suite *ValidationTestSuite) TestFieldStateValidate() {
	validator := RequiredValidator{}
	fieldState := NewFieldState("username", validator)

	// Test with empty value (should fail)
	fieldState.Value = ""
	fieldState.Validate()
	suite.False(fieldState.IsValid)
	suite.NotEmpty(fieldState.Errors)

	// Test with valid value
	fieldState.Value = "testuser"
	fieldState.Validate()
	suite.True(fieldState.IsValid)
	suite.Empty(fieldState.Errors)
}

func (suite *ValidationTestSuite) TestFieldStateGetFirstError() {
	validator := RequiredValidator{}
	fieldState := NewFieldState("username", validator)

	// No errors
	suite.Equal("", fieldState.GetFirstError())

	// With errors
	fieldState.Errors = []string{"First error", "Second error"}
	suite.Equal("First error", fieldState.GetFirstError())
}

func (suite *ValidationTestSuite) TestFieldStateMarkAsTouched() {
	fieldState := NewFieldState("username")
	suite.False(fieldState.IsTouched)

	fieldState.MarkAsTouched()
	suite.True(fieldState.IsTouched)
}

func (suite *ValidationTestSuite) TestFieldStateMarkAsDirty() {
	fieldState := NewFieldState("username")
	suite.False(fieldState.IsDirty)

	fieldState.MarkAsDirty()
	suite.True(fieldState.IsDirty)
}

func (suite *ValidationTestSuite) TestFieldStateSetValue() {
	fieldState := NewFieldState("username")
	suite.False(fieldState.IsDirty)

	fieldState.SetValue("newvalue")
	suite.Equal("newvalue", fieldState.Value)
	suite.True(fieldState.IsDirty)
}

func (suite *ValidationTestSuite) TestFieldStateReset() {
	fieldState := NewFieldState("username")
	fieldState.Value = "test"
	fieldState.IsDirty = true
	fieldState.IsTouched = true
	fieldState.IsValid = false
	fieldState.Errors = []string{"error"}

	fieldState.Reset()

	suite.Nil(fieldState.Value)
	suite.False(fieldState.IsDirty)
	suite.False(fieldState.IsTouched)
	suite.True(fieldState.IsValid)
	suite.Nil(fieldState.Errors)
}

// ============================================================================
// FORM STATE TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestNewFormState() {
	formState := NewFormState()

	suite.NotNil(formState.Fields)
	suite.True(formState.IsValid)
	suite.False(formState.IsSubmitting)
	suite.Equal(0, formState.SubmitCount)
}

func (suite *ValidationTestSuite) TestFormStateRegisterField() {
	formState := NewFormState()
	validator := RequiredValidator{}

	formState.RegisterField("username", validator)

	suite.Contains(formState.Fields, "username")
	field := formState.Fields["username"]
	suite.Equal("username", field.Name)
	suite.Len(field.Validators, 1)
}

func (suite *ValidationTestSuite) TestFormStateSetFieldValue() {
	formState := NewFormState()
	formState.RegisterField("username")

	formState.SetFieldValue("username", "testuser")

	field := formState.Fields["username"]
	suite.Equal("testuser", field.Value)
	suite.True(field.IsDirty)
}

func (suite *ValidationTestSuite) TestFormStateValidateField() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})

	// Test with invalid value
	formState.SetFieldValue("username", "")
	isValid := formState.ValidateField("username")
	suite.False(isValid)

	// Test with valid value
	formState.SetFieldValue("username", "testuser")
	isValid = formState.ValidateField("username")
	suite.True(isValid)

	// Test non-existent field
	isValid = formState.ValidateField("nonexistent")
	suite.False(isValid)
}

func (suite *ValidationTestSuite) TestFormStateValidateAll() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})
	formState.RegisterField("email", RequiredValidator{}, EmailValidator{})

	// Test with all invalid
	formState.SetFieldValue("username", "")
	formState.SetFieldValue("email", "")
	isValid := formState.ValidateAll()
	suite.False(isValid)
	suite.False(formState.IsValid)

	// Test with all valid
	formState.SetFieldValue("username", "testuser")
	formState.SetFieldValue("email", "test@example.com")
	isValid = formState.ValidateAll()
	suite.True(isValid)
	suite.True(formState.IsValid)
}

func (suite *ValidationTestSuite) TestFormStateValidateTouchedFields() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})
	formState.RegisterField("email", RequiredValidator{})

	// Set values but only touch username
	formState.SetFieldValue("username", "")
	formState.SetFieldValue("email", "") // Not touched
	formState.Fields["username"].MarkAsTouched()

	// Should only validate touched field
	isValid := formState.ValidateTouchedFields()
	suite.False(isValid) // username is invalid
	suite.False(formState.Fields["username"].IsValid)
	suite.True(formState.Fields["email"].IsValid) // Not validated because not touched
}

func (suite *ValidationTestSuite) TestFormStateGetFieldErrors() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})

	// No errors initially
	errors := formState.GetFieldErrors("username")
	suite.Empty(errors)

	// Validate with invalid value
	formState.SetFieldValue("username", "")
	formState.ValidateField("username")
	errors = formState.GetFieldErrors("username")
	suite.NotEmpty(errors)

	// Non-existent field
	errors = formState.GetFieldErrors("nonexistent")
	suite.Nil(errors)
}

func (suite *ValidationTestSuite) TestFormStateGetAllErrors() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})
	formState.RegisterField("email", RequiredValidator{})

	// Set invalid values and validate
	formState.SetFieldValue("username", "")
	formState.SetFieldValue("email", "")
	formState.ValidateAll()

	allErrors := formState.GetAllErrors()
	suite.Len(allErrors, 2)
	suite.Contains(allErrors, "username")
	suite.Contains(allErrors, "email")
}

func (suite *ValidationTestSuite) TestFormStateReset() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})

	// Set some state
	formState.SetFieldValue("username", "test")
	formState.Fields["username"].MarkAsTouched()
	formState.IsSubmitting = true
	formState.SubmitCount = 5

	formState.Reset()

	// Check everything is reset
	field := formState.Fields["username"]
	suite.Nil(field.Value)
	suite.False(field.IsDirty)
	suite.False(field.IsTouched)
	suite.True(field.IsValid)
	suite.False(formState.IsSubmitting)
	suite.True(formState.IsValid)
	suite.Equal(0, formState.SubmitCount)
}

func (suite *ValidationTestSuite) TestFormStateJSONSerialization() {
	formState := NewFormState()
	formState.RegisterField("username", RequiredValidator{})
	formState.SetFieldValue("username", "testuser")
	formState.Fields["username"].MarkAsTouched()

	// Test ToJSON
	jsonData, err := formState.ToJSON()
	suite.NoError(err)
	suite.NotEmpty(jsonData)

	// Test FromJSON
	newFormState := NewFormState()
	newFormState.RegisterField("username", RequiredValidator{})
	err = newFormState.FromJSON(jsonData)
	suite.NoError(err)

	field := newFormState.Fields["username"]
	suite.Equal("testuser", field.Value)
	suite.True(field.IsTouched)
}

// ============================================================================
// VALIDATION PROPS BUILDER TESTS
// ============================================================================

func (suite *ValidationTestSuite) TestNewValidationProps() {
	builder := NewValidationProps()
	props := builder.Build()

	suite.Equal(StateDefault, props.State)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderWithState() {
	props := NewValidationProps().WithState(StateError).Build()
	suite.Equal(StateError, props.State)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderWithError() {
	props := NewValidationProps().WithError("Error message").Build()
	suite.Equal(StateError, props.State)
	suite.Equal("Error message", props.ErrorText)
	suite.True(props.ShowValidation)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderWithSuccess() {
	props := NewValidationProps().WithSuccess("Success message").Build()
	suite.Equal(StateSuccess, props.State)
	suite.Equal("Success message", props.SuccessText)
	suite.True(props.ShowValidation)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderWithWarning() {
	props := NewValidationProps().WithWarning("Warning message").Build()
	suite.Equal(StateWarning, props.State)
	suite.Equal("Warning message", props.WarningText)
	suite.True(props.ShowValidation)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderWithInfo() {
	props := NewValidationProps().WithInfo("Info message").Build()
	suite.Equal(StateInfo, props.State)
	suite.Equal("Info message", props.InfoText)
	suite.True(props.ShowValidation)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderWithHelpText() {
	props := NewValidationProps().WithHelpText("Help text").Build()
	suite.Equal("Help text", props.HelpText)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderShowValidationUI() {
	props := NewValidationProps().ShowValidationUI(true).Build()
	suite.True(props.ShowValidation)

	props = NewValidationProps().ShowValidationUI(false).Build()
	suite.False(props.ShowValidation)
}

func (suite *ValidationTestSuite) TestValidationPropsBuilderChaining() {
	props := NewValidationProps().
		WithError("Error message").
		WithHelpText("Help text").
		ShowValidationUI(true).
		Build()

	suite.Equal(StateError, props.State)
	suite.Equal("Error message", props.ErrorText)
	suite.Equal("Help text", props.HelpText)
	suite.True(props.ShowValidation)
}