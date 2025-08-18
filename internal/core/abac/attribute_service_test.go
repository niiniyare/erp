package abac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// Mock AttributeRepository
type MockAttributeRepository struct {
	mock.Mock
}

func (m *MockAttributeRepository) CreateAttributeDefinition(ctx context.Context, req *repository.CreateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.AttributeDefinition), args.Error(1)
}

func (m *MockAttributeRepository) UpdateAttributeDefinition(ctx context.Context, req *repository.UpdateAttributeDefinitionRequest) (*models.AttributeDefinition, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.AttributeDefinition), args.Error(1)
}

func (m *MockAttributeRepository) DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAttributeRepository) GetAttributeDefinitionByID(ctx context.Context, id uuid.UUID) (*models.AttributeDefinition, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.AttributeDefinition), args.Error(1)
}

func (m *MockAttributeRepository) GetAttributeDefinitionByName(ctx context.Context, name string) (*models.AttributeDefinition, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*models.AttributeDefinition), args.Error(1)
}

func (m *MockAttributeRepository) ListAttributeDefinitions(ctx context.Context, req *repository.ListAttributeDefinitionsRequest) ([]*models.AttributeDefinition, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*models.AttributeDefinition), args.Error(1)
}

func (m *MockAttributeRepository) BuildAttributeContext(ctx context.Context, req *repository.BuildAttributeContextRequest) (*models.AttributeContext, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.AttributeContext), args.Error(1)
}

func (m *MockAttributeRepository) CreateAttributeValue(ctx context.Context, req *repository.CreateAttributeValueRequest) (*models.AttributeValue, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.AttributeValue), args.Error(1)
}

func (m *MockAttributeRepository) GetAttributeValue(ctx context.Context, definitionID uuid.UUID, entityID uuid.UUID) (*models.AttributeValue, error) {
	args := m.Called(ctx, definitionID, entityID)
	return args.Get(0).(*models.AttributeValue), args.Error(1)
}

func (m *MockAttributeRepository) GetAttributeValuesByEntity(ctx context.Context, entityID uuid.UUID, category types.AttributeCategory) ([]*models.AttributeValue, error) {
	args := m.Called(ctx, entityID, category)
	return args.Get(0).([]*models.AttributeValue), args.Error(1)
}

func (m *MockAttributeRepository) UpdateAttributeValue(ctx context.Context, definitionID, entityID uuid.UUID, value any) (*models.AttributeValue, error) {
	args := m.Called(ctx, definitionID, entityID, value)
	return args.Get(0).(*models.AttributeValue), args.Error(1)
}

func (m *MockAttributeRepository) DeleteAttributeValue(ctx context.Context, definitionID, entityID uuid.UUID) error {
	args := m.Called(ctx, definitionID, entityID)
	return args.Error(0)
}

func (m *MockAttributeRepository) StoreAttributeValues(ctx context.Context, values []*repository.StoreAttributeValueRequest) ([]*models.AttributeValue, error) {
	args := m.Called(ctx, values)
	return args.Get(0).([]*models.AttributeValue), args.Error(1)
}

func (m *MockAttributeRepository) GetAttributeValuesByDefinitions(ctx context.Context, definitionIDs []uuid.UUID, entityID uuid.UUID) ([]*models.AttributeValue, error) {
	args := m.Called(ctx, definitionIDs, entityID)
	return args.Get(0).([]*models.AttributeValue), args.Error(1)
}

func (m *MockAttributeRepository) GetUserAttributeContext(ctx context.Context, userID uuid.UUID) (*models.AttributeContext, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.AttributeContext), args.Error(1)
}

func (m *MockAttributeRepository) GetResourceAttributeContext(ctx context.Context, resourceType string, resourceID uuid.UUID) (*models.AttributeContext, error) {
	args := m.Called(ctx, resourceType, resourceID)
	return args.Get(0).(*models.AttributeContext), args.Error(1)
}

func (m *MockAttributeRepository) CleanupExpiredAttributes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAttributeRepository) GetAttributeStats(ctx context.Context) (*repository.AttributeStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*repository.AttributeStats), args.Error(1)
}

// Test Suite for Attribute Service
func TestAttributeService(t *testing.T) {
	t.Run("TestCreateAttributeDefinition_Success", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.WithFields(logger.Fields{})
		ctrl := gomock.NewController(t)
		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

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
		mockRepo.On("CreateAttributeDefinition", mock.Anything, request).Return(expectedResult, nil)

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
		mockRepo.AssertExpectations(t)
	})

	t.Run("TestCreateAttributeDefinition_ValidationError", func(t *testing.T) {
		// Setup
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.WithFields(logger.Fields{})
		ctrl := gomock.NewController(t)

		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

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
		ctrl := gomock.NewController(t)

		mockLogger := logger.WithFields(logger.Fields{})
		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &ValidateAttributeValueRequest{
			AttributeID:     uuid.New(),
			Value:           "engineering",
			ValidationLevel: "strict",
		}

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
		mockRepo := &MockAttributeRepository{}
		mockLogger := logger.WithFields(logger.Fields{})
		ctrl := gomock.NewController(t)

		mockMetrics := &metrics.MetricsService{}
		mockTracer := tracing.NewMockTracingService(ctrl)

		encryptionKey := []byte("test-key-for-encryption-32-byte")
		service := NewAttributeService(mockRepo, encryptionKey, mockLogger, mockMetrics, mockTracer)

		// Test data
		request := &ValidateAttributeValueRequest{
			AttributeID:     uuid.New(),
			Value:           "invalid_department",
			ValidationLevel: "strict",
		}

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

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
