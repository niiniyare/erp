package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *ErrorsTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}

// Test SchemaError interface methods
func (suite *ErrorsTestSuite) TestSchemaErrorMethods() {
	err := NewValidationError("test_field", "test error message")

	// Test Code method
	require.Equal(suite.T(), "test_field", err.Code())

	// Test Type method
	require.Equal(suite.T(), ErrorTypeValidation, err.Type())

	// Test Field method
	require.Equal(suite.T(), "", err.Field())

	// Test Details method
	require.NotNil(suite.T(), err.Details())
}

// Test error creation with field
func (suite *ErrorsTestSuite) TestErrorWithField() {
	err := NewValidationError("test_code", "test message").WithField("field_name")

	require.Equal(suite.T(), "field_name", err.Field())
	require.Contains(suite.T(), err.Error(), "field_name")
}

// Test error creation with details
func (suite *ErrorsTestSuite) TestErrorWithDetail() {
	err := NewValidationError("test_code", "test message").WithDetail("custom_key", "custom_value")

	require.NotNil(suite.T(), err.Details())
	require.Equal(suite.T(), "custom_value", err.Details()["custom_key"])
}

// Test ValidationErrorCollection methods
func (suite *ErrorsTestSuite) TestValidationErrorCollection() {
	collection := NewValidationErrorCollection()

	// Test initial state
	require.False(suite.T(), collection.HasErrors())
	require.Equal(suite.T(), 0, collection.Count())
	require.Equal(suite.T(), "no validation errors", collection.Error())

	// Add errors
	err1 := NewValidationError("code1", "message1")
	err2 := NewValidationError("code2", "message2").WithField("field1")

	collection.Add(err1)
	collection.AddWithField("field2", err2)

	// Test after adding errors
	require.True(suite.T(), collection.HasErrors())
	require.Equal(suite.T(), 2, collection.Count())
	require.NotEqual(suite.T(), "", collection.Error())

	// Test Errors method
	errors := collection.Errors()
	require.Len(suite.T(), errors, 2)

	// Test ErrorsByField method
	fieldErrors := collection.ErrorsByField()
	require.Contains(suite.T(), fieldErrors, "general") // err1 has no field
	require.Contains(suite.T(), fieldErrors, "field2")  // err2 was added with field2
}

// Test error type checking functions
func (suite *ErrorsTestSuite) TestErrorTypeChecking() {
	validationErr := NewValidationError("test", "message")
	notFoundErr := NewNotFoundError("resource", "id123")
	permissionErr := NewPermissionError("action", "unauthorized")

	// Test IsErrorType
	require.True(suite.T(), IsErrorType(validationErr, ErrorTypeValidation))
	require.False(suite.T(), IsErrorType(validationErr, ErrorTypeNotFound))

	// Test specific type checkers
	require.True(suite.T(), IsValidationError(validationErr))
	require.False(suite.T(), IsValidationError(notFoundErr))

	require.True(suite.T(), IsNotFoundError(notFoundErr))
	require.False(suite.T(), IsNotFoundError(validationErr))

	require.True(suite.T(), IsPermissionError(permissionErr))
	require.False(suite.T(), IsPermissionError(validationErr))
}

// Test error code and details extraction
func (suite *ErrorsTestSuite) TestErrorExtraction() {
	err := NewValidationError("test_code", "message").WithDetail("key", "value")

	// Test GetErrorCode
	code := GetErrorCode(err)
	require.Equal(suite.T(), "test_code", code)

	// Test GetErrorDetails
	details := GetErrorDetails(err)
	require.Equal(suite.T(), "value", details["key"])

	// Test with non-schema error
	regularErr := fmt.Errorf("regular error")
	code = GetErrorCode(regularErr)
	require.Equal(suite.T(), "unknown_error", code)

	details = GetErrorDetails(regularErr)
	require.Contains(suite.T(), details, "error")
}

// Test HTTP status code mapping
func (suite *ErrorsTestSuite) TestHTTPStatusCode() {
	testCases := []struct {
		err        SchemaError
		statusCode int
	}{
		{NewValidationError("test", "message"), 400},
		{NewNotFoundError("resource", "id"), 404},
		{NewConflictError("resource", "conflict"), 409},
		{NewPermissionError("action", "denied"), 403},
		{NewInternalError("internal", "error"), 500},
	}

	for _, tc := range testCases {
		code := GetHTTPStatusCode(tc.err)
		require.Equal(suite.T(), tc.statusCode, code)
	}

	// Test with non-schema error
	regularErr := fmt.Errorf("regular error")
	code := GetHTTPStatusCode(regularErr)
	require.Equal(suite.T(), 500, code)
}

// Test error response conversion
func (suite *ErrorsTestSuite) TestErrorResponse() {
	err := NewValidationError("test_code", "test message").WithField("test_field")

	response := ToErrorResponse(err)
	require.Equal(suite.T(), "test_code", response.Code)
	require.Contains(suite.T(), response.Error, "test message")
	require.Equal(suite.T(), "validation", response.Type)
	require.Equal(suite.T(), "test_field", response.Field)

	// Test with non-schema error
	regularErr := fmt.Errorf("regular error")
	response = ToErrorResponse(regularErr)
	require.Equal(suite.T(), "unknown_error", response.Code)
	require.Contains(suite.T(), response.Error, "regular error")
}

// Test multi-error response conversion
func (suite *ErrorsTestSuite) TestMultiErrorResponse() {
	collection := NewValidationErrorCollection()
	collection.Add(NewValidationError("code1", "message1"))
	collection.Add(NewValidationError("code2", "message2").WithField("field1"))

	response := ToMultiErrorResponse(collection)
	require.Len(suite.T(), response.Errors, 2)
	require.Equal(suite.T(), "code1", response.Errors[0].Code)
	require.Equal(suite.T(), "code2", response.Errors[1].Code)
	require.Equal(suite.T(), "field1", response.Errors[1].Field)
}

// Test WrapError function
func (suite *ErrorsTestSuite) TestWrapErrorFunction() {
	// Test with schema error
	schemaErr := NewValidationError("test", "message")
	wrapped := WrapError(schemaErr, "test_code", "wrapped")
	require.Equal(suite.T(), schemaErr, wrapped)

	// Test with regular error
	regularErr := fmt.Errorf("regular error")
	wrapped = WrapError(regularErr, "wrap_code", "wrapped error")
	require.IsType(suite.T(), &BaseError{}, wrapped)
	require.Contains(suite.T(), wrapped.Error(), "regular error")
}

// Test error handler
func (suite *ErrorsTestSuite) TestErrorHandler() {
	handler := NewErrorHandler(nil)
	require.NotNil(suite.T(), handler)

	// Test handling schema error
	schemaErr := NewValidationError("test", "message")
	result := handler.HandleError(schemaErr)
	require.Equal(suite.T(), schemaErr, result)

	// Test handling regular error
	regularErr := fmt.Errorf("regular error")
	result = handler.HandleError(regularErr)
	require.IsType(suite.T(), &BaseError{}, result)
}

// Test WrapError function
func (suite *ErrorsTestSuite) TestWrapError() {
	originalErr := fmt.Errorf("original error")
	wrappedErr := WrapError(originalErr, "wrapped_code", "wrapped message")

	require.IsType(suite.T(), &BaseError{}, wrappedErr)
	require.Equal(suite.T(), "wrapped_code", wrappedErr.Code())
	require.Contains(suite.T(), wrappedErr.Error(), "wrapped message")
	require.Contains(suite.T(), wrappedErr.Error(), "original error")
}
