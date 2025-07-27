package abac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Mock AttributeRepository
type MockAttributeRepository struct {
	mock.Mock
}

func (m *MockAttributeRepository) CreateAttributeDefinition(ctx context.Context, req *CreateAttributeDefinitionRequest) (*AttributeDefinitionResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*AttributeDefinitionResult), args.Error(1)
}

func (m *MockAttributeRepository) UpdateAttributeDefinition(ctx context.Context, req *UpdateAttributeDefinitionRequest) (*AttributeDefinitionResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*AttributeDefinitionResult), args.Error(1)
}

func (m *MockAttributeRepository) DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAttributeRepository) GetAttributeDefinition(ctx context.Context, id uuid.UUID) (*AttributeDefinitionDetails, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*AttributeDefinitionDetails), args.Error(1)
}

func (m *MockAttributeRepository) ListAttributeDefinitions(ctx context.Context, req *ListAttributeDefinitionsRequest) (*AttributeDefinitionListResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*AttributeDefinitionListResult), args.Error(1)
}

// Test Suite for Attribute Service
func TestAttributeService(t *testing.T) {
	t.Run("TestCreateAttributeDefinition_Success", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		userID := uuid.New()
		request := &CreateAttributeDefinitionRequest{
			Name:         "user_department",
			DisplayName:  stringPtr("User Department"),
			Description:  stringPtr("The department the user belongs to"),
			DataType:     types.AttributeDataTypeString,
			Category:     types.AttributeCategoryUser,
			IsRequired:   true,
			IsMultiValue: false,
			DefaultValue: "engineering",
			AllowedValues: []interface{}{
				"engineering", "sales", "marketing", "hr", "finance",
			},
			ValidationRules: []AttributeValidationRule{
				{
					Type:       "length",
					Parameters: map[string]interface{}{"min": 1, "max": 50},
				},
			},
			Constraints: AttributeConstraints{
				MinLength: intPtr(1),
				MaxLength: intPtr(50),
			},
			SecuritySettings: AttributeSecuritySettings{
				IsEncrypted: false,
				IsSensitive: false,
				AccessLevel: "standard",
			},
			Metadata: map[string]interface{}{
				"source": "hr_system",
			},
			Tags:      []string{"user", "department", "organizational"},
			CreatedBy: &userID,
		}

		expectedResult := &AttributeDefinitionResult{
			ID:           uuid.New(),
			Name:         request.Name,
			DisplayName:  request.DisplayName,
			Description:  request.Description,
			DataType:     request.DataType,
			Category:     request.Category,
			IsRequired:   request.IsRequired,
			IsMultiValue: request.IsMultiValue,
			DefaultValue: request.DefaultValue,
			CreatedBy:    request.CreatedBy,
		}

		// Mock expectations
		mockRepo.On("CreateAttributeDefinition", mock.Anything, request).Return(expectedResult, nil)

		// Execute
		result, err := service.CreateAttributeDefinition(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedResult.Name, result.Name)
		assert.Equal(t, expectedResult.DataType, result.DataType)
		assert.Equal(t, expectedResult.Category, result.Category)
		assert.Equal(t, expectedResult.IsRequired, result.IsRequired)

		// Verify mock calls
		mockRepo.AssertExpectations(t)
	})

	t.Run("TestCreateAttributeDefinition_ValidationError", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data with validation errors
		request := &CreateAttributeDefinitionRequest{
			Name:     "", // Invalid: empty name
			DataType: types.AttributeDataTypeString,
			Category: types.AttributeCategoryUser,
		}

		// Execute
		result, err := service.CreateAttributeDefinition(context.Background(), request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation failed")

		// Verify no repository calls were made
		mockRepo.AssertNotCalled(t, "CreateAttributeDefinition")
	})

	t.Run("TestValidateAttributeValue_Success", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &ValidateAttributeValueRequest{
			AttributeName: "user_department",
			Value:         "engineering",
			DataType:      types.AttributeDataTypeString,
			ValidationRules: []AttributeValidationRule{
				{
					Type: "allowed_values",
					Parameters: map[string]interface{}{
						"values": []string{"engineering", "sales", "marketing"},
					},
				},
			},
		}

		// Execute
		result, err := service.ValidateAttributeValue(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.ValidationErrors)
	})

	t.Run("TestValidateAttributeValue_FailedValidation", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &ValidateAttributeValueRequest{
			AttributeName: "user_department",
			Value:         "invalid_department",
			DataType:      types.AttributeDataTypeString,
			ValidationRules: []AttributeValidationRule{
				{
					Type: "allowed_values",
					Parameters: map[string]interface{}{
						"values": []string{"engineering", "sales", "marketing"},
					},
				},
			},
		}

		// Execute
		result, err := service.ValidateAttributeValue(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.ValidationErrors)
		assert.Contains(t, result.ValidationErrors[0].Message, "not in allowed values")
	})

	t.Run("TestEncryptAttributeValue_Success", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &EncryptAttributeValueRequest{
			AttributeName: "social_security_number",
			Value:         "123-45-6789",
			EncryptionKey: "user-specific-key",
		}

		// Execute
		result, err := service.EncryptAttributeValue(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.EncryptedValue)
		assert.NotEqual(t, request.Value, result.EncryptedValue)
		assert.Equal(t, "AES-256-GCM", result.EncryptionMethod)
	})

	t.Run("TestDecryptAttributeValue_Success", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// First encrypt a value
		originalValue := "sensitive-data-123"
		encryptRequest := &EncryptAttributeValueRequest{
			AttributeName: "sensitive_field",
			Value:         originalValue,
			EncryptionKey: "test-key",
		}

		encryptResult, err := service.EncryptAttributeValue(context.Background(), encryptRequest)
		require.NoError(t, err)

		// Then decrypt it
		decryptRequest := &DecryptAttributeValueRequest{
			AttributeName:  "sensitive_field",
			EncryptedValue: encryptResult.EncryptedValue,
			EncryptionKey:  "test-key",
		}

		// Execute
		result, err := service.DecryptAttributeValue(context.Background(), decryptRequest)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, originalValue, result.DecryptedValue)
	})

	t.Run("TestGetSupportedDataTypes", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Execute
		result, err := service.GetSupportedDataTypes(context.Background())

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.DataTypes)

		// Check for expected data types
		dataTypeNames := make([]string, len(result.DataTypes))
		for i, dt := range result.DataTypes {
			dataTypeNames[i] = dt.Name
		}

		assert.Contains(t, dataTypeNames, "string")
		assert.Contains(t, dataTypeNames, "number")
		assert.Contains(t, dataTypeNames, "boolean")
		assert.Contains(t, dataTypeNames, "date")
		assert.Contains(t, dataTypeNames, "json")
		assert.Contains(t, dataTypeNames, "array")
		assert.Contains(t, dataTypeNames, "enum")
	})

	t.Run("TestGetAttributeCategories", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.NewMockLogger()
		mockMetrics := metrics.NewMockMetricsProvider()
		mockTracer := tracing.NewMockTracingService()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Execute
		result, err := service.GetAttributeCategories(context.Background())

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Categories)

		// Check for expected categories
		categoryNames := make([]string, len(result.Categories))
		for i, cat := range result.Categories {
			categoryNames[i] = cat.Name
		}

		assert.Contains(t, categoryNames, "user")
		assert.Contains(t, categoryNames, "resource")
		assert.Contains(t, categoryNames, "environment")
		assert.Contains(t, categoryNames, "action")
		assert.Contains(t, categoryNames, "entity")
		assert.Contains(t, categoryNames, "session")
	})
}

// Test Suite for Attribute Validation
func TestAttributeValidation(t *testing.T) {
	service := &attributeService{}

	t.Run("TestValidateStringAttribute", func(t *testing.T) {
		t.Run("ValidString", func(t *testing.T) {
			errors := service.validateAttributeValue("test", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "length",
					Parameters: map[string]interface{}{
						"min": 1,
						"max": 10,
					},
				},
			})
			assert.Empty(t, errors)
		})

		t.Run("StringTooShort", func(t *testing.T) {
			errors := service.validateAttributeValue("", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "length",
					Parameters: map[string]interface{}{
						"min": 1,
						"max": 10,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "too short")
		})

		t.Run("StringTooLong", func(t *testing.T) {
			errors := service.validateAttributeValue("this string is way too long for validation", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "length",
					Parameters: map[string]interface{}{
						"min": 1,
						"max": 10,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "too long")
		})

		t.Run("AllowedValues", func(t *testing.T) {
			// Valid value
			errors := service.validateAttributeValue("admin", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "allowed_values",
					Parameters: map[string]interface{}{
						"values": []string{"admin", "user", "guest"},
					},
				},
			})
			assert.Empty(t, errors)

			// Invalid value
			errors = service.validateAttributeValue("superuser", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "allowed_values",
					Parameters: map[string]interface{}{
						"values": []string{"admin", "user", "guest"},
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "not in allowed values")
		})

		t.Run("RegexPattern", func(t *testing.T) {
			// Valid email pattern
			errors := service.validateAttributeValue("test@example.com", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "regex",
					Parameters: map[string]interface{}{
						"pattern": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
					},
				},
			})
			assert.Empty(t, errors)

			// Invalid email pattern
			errors = service.validateAttributeValue("invalid-email", types.AttributeDataTypeString, []AttributeValidationRule{
				{
					Type: "regex",
					Parameters: map[string]interface{}{
						"pattern": `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "does not match pattern")
		})
	})

	t.Run("TestValidateNumberAttribute", func(t *testing.T) {
		t.Run("ValidNumber", func(t *testing.T) {
			errors := service.validateAttributeValue(42, types.AttributeDataTypeNumber, []AttributeValidationRule{
				{
					Type: "range",
					Parameters: map[string]interface{}{
						"min": 0,
						"max": 100,
					},
				},
			})
			assert.Empty(t, errors)
		})

		t.Run("NumberTooSmall", func(t *testing.T) {
			errors := service.validateAttributeValue(-5, types.AttributeDataTypeNumber, []AttributeValidationRule{
				{
					Type: "range",
					Parameters: map[string]interface{}{
						"min": 0,
						"max": 100,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "below minimum")
		})

		t.Run("NumberTooLarge", func(t *testing.T) {
			errors := service.validateAttributeValue(150, types.AttributeDataTypeNumber, []AttributeValidationRule{
				{
					Type: "range",
					Parameters: map[string]interface{}{
						"min": 0,
						"max": 100,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "above maximum")
		})
	})

	t.Run("TestValidateBooleanAttribute", func(t *testing.T) {
		t.Run("ValidBoolean", func(t *testing.T) {
			errors := service.validateAttributeValue(true, types.AttributeDataTypeBoolean, []AttributeValidationRule{})
			assert.Empty(t, errors)

			errors = service.validateAttributeValue(false, types.AttributeDataTypeBoolean, []AttributeValidationRule{})
			assert.Empty(t, errors)
		})

		t.Run("InvalidBooleanType", func(t *testing.T) {
			errors := service.validateAttributeValue("true", types.AttributeDataTypeBoolean, []AttributeValidationRule{})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "must be a boolean")
		})
	})

	t.Run("TestValidateArrayAttribute", func(t *testing.T) {
		t.Run("ValidArray", func(t *testing.T) {
			errors := service.validateAttributeValue([]string{"admin", "user"}, types.AttributeDataTypeArray, []AttributeValidationRule{
				{
					Type: "array_length",
					Parameters: map[string]interface{}{
						"min": 1,
						"max": 5,
					},
				},
			})
			assert.Empty(t, errors)
		})

		t.Run("ArrayTooShort", func(t *testing.T) {
			errors := service.validateAttributeValue([]string{}, types.AttributeDataTypeArray, []AttributeValidationRule{
				{
					Type: "array_length",
					Parameters: map[string]interface{}{
						"min": 1,
						"max": 5,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "too few elements")
		})

		t.Run("ArrayTooLong", func(t *testing.T) {
			errors := service.validateAttributeValue([]string{"a", "b", "c", "d", "e", "f"}, types.AttributeDataTypeArray, []AttributeValidationRule{
				{
					Type: "array_length",
					Parameters: map[string]interface{}{
						"min": 1,
						"max": 5,
					},
				},
			})
			assert.NotEmpty(t, errors)
			assert.Contains(t, errors[0].Message, "too many elements")
		})
	})
}

// Test Suite for Encryption/Decryption
func TestAttributeEncryption(t *testing.T) {
	encryptionKey := []byte("test-key-for-encryption-32-byte")
	service := &attributeService{
		encryptionKey: encryptionKey,
	}

	t.Run("TestEncryptDecryptRoundTrip", func(t *testing.T) {
		originalValue := "sensitive-personal-data"

		// Encrypt
		encryptedValue, err := service.encryptValue(originalValue)
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedValue)
		assert.NotEqual(t, originalValue, encryptedValue)

		// Decrypt
		decryptedValue, err := service.decryptValue(encryptedValue)
		require.NoError(t, err)
		assert.Equal(t, originalValue, decryptedValue)
	})

	t.Run("TestEncryptEmptyValue", func(t *testing.T) {
		encryptedValue, err := service.encryptValue("")
		require.NoError(t, err)
		assert.NotEmpty(t, encryptedValue)

		decryptedValue, err := service.decryptValue(encryptedValue)
		require.NoError(t, err)
		assert.Equal(t, "", decryptedValue)
	})

	t.Run("TestDecryptInvalidValue", func(t *testing.T) {
		_, err := service.decryptValue("invalid-encrypted-value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode base64")
	})

	t.Run("TestDecryptTooShortValue", func(t *testing.T) {
		// Create a base64 encoded value that's too short
		shortValue := "YWJj" // "abc" in base64, which is too short for nonce
		_, err := service.decryptValue(shortValue)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext too short")
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

// Benchmark tests
func BenchmarkAttributeValidation(b *testing.B) {
	service := &attributeService{}
	rules := []AttributeValidationRule{
		{
			Type: "length",
			Parameters: map[string]interface{}{
				"min": 1,
				"max": 100,
			},
		},
		{
			Type: "regex",
			Parameters: map[string]interface{}{
				"pattern": `^[a-zA-Z0-9_]+$`,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.validateAttributeValue("test_value_123", types.AttributeDataTypeString, rules)
	}
}

func BenchmarkAttributeEncryption(b *testing.B) {
	encryptionKey := []byte("test-key-for-encryption-32-byte")
	service := &attributeService{
		encryptionKey: encryptionKey,
	}
	value := "test-value-for-encryption"

	b.Run("Encrypt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = service.encryptValue(value)
		}
	})

	// Pre-encrypt for decrypt benchmark
	encryptedValue, _ := service.encryptValue(value)

	b.Run("Decrypt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = service.decryptValue(encryptedValue)
		}
	})
}
