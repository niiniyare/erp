package abac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// NoOpSpan is a no-op implementation of the tracing.Span interface for testing
type NoOpSpan struct{}

func (n *NoOpSpan) End(opts ...tracing.SpanEndOption)                 {}
func (n *NoOpSpan) AddEvent(name string, attrs ...attribute.KeyValue) {}
func (n *NoOpSpan) SetAttributes(attrs ...attribute.KeyValue)         {}
func (n *NoOpSpan) SetStatus(code codes.Code, description string)     {}
func (n *NoOpSpan) SetName(name string)                               {}
func (n *NoOpSpan) RecordError(err error, opts ...trace.EventOption)  {}
func (n *NoOpSpan) IsRecording() bool                                 { return true }
func (n *NoOpSpan) SpanContext() trace.SpanContext                    { return trace.SpanContext{} }

// Test Suite for Attribute Service
func TestAttributeService(t *testing.T) {
	t.Run("TestCreateAttributeDefinition_Success", func(t *testing.T) {
		// Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockAttributeRepository(ctrl)
		mockLogger := logger.WithFields(logger.Fields{})
		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

		mockTracer.EXPECT().StartSpan(gomock.Any(), "abac.attribute_service.CreateAttributeDefinition").Return(context.Background(), &NoOpSpan{}).AnyTimes()

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
			AllowedValues: []any{
				"engineering", "sales", "marketing", "hr", "finance",
			},
			ValidationRules: []AttributeValidationRule{
				{
					RuleType:   "length",
					Parameters: map[string]any{"min": 1, "max": 50},
				},
			},
			Constraints: AttributeConstraints{
				MinLength: int32Ptr(1),
				MaxLength: int32Ptr(50),
			},
			SecuritySettings: AttributeSecuritySettings{
				EncryptionRequired: false,
				AuditingRequired:   false,
				AccessLevel:        "standard",
			},
			Metadata: map[string]any{
				"source": "hr_system",
			},
			Tags:      []string{"user", "department", "organizational"},
			CreatedBy: &userID,
		}

		expectedResult := &AttributeDefinitionResult{
			AttributeDefinition: &models.AttributeDefinition{
				ID:          uuid.New(),
				Name:        request.Name,
				DisplayName: request.DisplayName,
				Description: request.Description,
				DataType:    request.DataType,
				Category:    request.Category,
				IsRequired:  request.IsRequired,
			},
		}

		// Mock expectations
		mockRepo.EXPECT().CreateAttributeDefinition(gomock.Any(), gomock.Any()).Return(expectedResult.AttributeDefinition, nil).AnyTimes()

		// Execute
		result, err := service.CreateAttributeDefinition(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedResult.AttributeDefinition.Name, result.AttributeDefinition.Name)
		assert.Equal(t, expectedResult.AttributeDefinition.DataType, result.AttributeDefinition.DataType)
		assert.Equal(t, expectedResult.AttributeDefinition.Category, result.AttributeDefinition.Category)
		assert.Equal(t, expectedResult.AttributeDefinition.IsRequired, result.AttributeDefinition.IsRequired)

		// Verify mock calls
		// With gomock, expectations are verified by ctrl.Finish() implicitly at the end of the test
	})

	t.Run("TestCreateAttributeDefinition_ValidationError", func(t *testing.T) {
		// Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockAttributeRepository(ctrl)
		mockLogger := logger.WithFields(logger.Fields{})
		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

		mockTracer.EXPECT().StartSpan(gomock.Any(), "abac.attribute_service.CreateAttributeDefinition").Return(context.Background(), &NoOpSpan{}).AnyTimes()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data with validation errors
		request := &CreateAttributeDefinitionRequest{
			Name:     "", // Invalid: empty name
			DataType: types.AttributeDataTypeString,
			Category: types.AttributeCategoryUser,
		}

		// Mock expectations
		mockRepo.EXPECT().CreateAttributeDefinition(gomock.Any(), gomock.Any()).Times(0)

		// Execute
		result, err := service.CreateAttributeDefinition(context.Background(), request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation failed")

		// Verify no repository calls were made
		// With gomock, expectations are verified by ctrl.Finish() implicitly at the end of the test
	})

	t.Run("TestValidateAttributeValue_Success", func(t *testing.T) {
		// Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockAttributeRepository(ctrl)
		mockLogger := logger.WithFields(logger.Fields{})
		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

		mockTracer.EXPECT().StartSpan(gomock.Any(), "abac.attribute_service.ValidateAttributeValue").Return(context.Background(), &NoOpSpan{}).AnyTimes()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &ValidateAttributeValueRequest{
			AttributeID:     uuid.New(),
			Value:           "engineering",
			ValidationLevel: "strict",
		}

		// Mock expectations
		// Need to mock GetAttributeDefinitionByID and BuildAttributeContext
		mockRepo.EXPECT().GetAttributeDefinitionByID(gomock.Any(), gomock.Any()).Return(&models.AttributeDefinition{ /* ... populate with necessary fields ... */ }, nil).AnyTimes()
		mockRepo.EXPECT().BuildAttributeContext(gomock.Any(), gomock.Any()).Return(&models.AttributeContext{ /* ... populate with necessary fields ... */ }, nil).AnyTimes()

		// Execute
		result, err := service.ValidateAttributeValue(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.ValidationResults)
	})

	t.Run("TestValidateAttributeValue_FailedValidation", func(t *testing.T) {
		// Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockAttributeRepository(ctrl)
		mockLogger := logger.WithFields(logger.Fields{})
		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

		mockTracer.EXPECT().StartSpan(gomock.Any(), "abac.attribute_service.ValidateAttributeValue").Return(context.Background(), &NoOpSpan{}).AnyTimes()

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &ValidateAttributeValueRequest{
			AttributeID:     uuid.New(),
			Value:           "invalid_department",
			ValidationLevel: "strict",
		}

		// Mock expectations
		// Need to mock GetAttributeDefinitionByID and BuildAttributeContext
		mockRepo.EXPECT().GetAttributeDefinitionByID(gomock.Any(), gomock.Any()).Return(&models.AttributeDefinition{ /* ... populate with necessary fields ... */ }, nil).AnyTimes()
		mockRepo.EXPECT().BuildAttributeContext(gomock.Any(), gomock.Any()).Return(&models.AttributeContext{ /* ... populate with necessary fields ... */ }, nil).AnyTimes()

		// Execute
		result, err := service.ValidateAttributeValue(context.Background(), request)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.ValidationResults)
		assert.False(t, result.ValidationResults[0].Passed)
	})
}

// Helper functions (boolPtr only - using shared stringPtr and int32Ptr)
func boolPtr(b bool) *bool {
	return &b
}
